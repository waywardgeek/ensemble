using System.Collections.Immutable;
using System.Text;

namespace Ensemble;

public sealed partial class Agent
{
    public partial RecallSettings RecallSettings()
    {
        lock (ingress)
        {
            return recallSettings;
        }
    }

    public partial void SetRecallSettings(RecallSettings settings)
    {
        ArgumentNullException.ThrowIfNull(settings);
        lock (ingress)
        {
            ObjectDisposedException.ThrowIf(destroyed, this);
            recallSettings = settings;
        }
    }

    public partial async Task SaveRecallSettingsAsync(RecallSettings settings, CancellationToken cancellation)
    {
        ArgumentNullException.ThrowIfNull(settings);
        cancellation.ThrowIfCancellationRequested();
        var reply = new TaskCompletionSource<string>(TaskCreationOptions.RunContinuationsAsynchronously);
        Enqueue(new Mail(MailKind.RecallSettings, Reply: reply, RecallSettings: settings));
        await reply.Task.WaitAsync(cancellation);
    }

    private partial void ApplyRecallSettings(Mail message)
    {
        try
        {
            // Read the saved sibling, not an unrelated transient memory tweak.
            // The actor serializes durable writes; Memory owns rooted disk I/O.
            var saved = Memory().ReadSettings();
            Memory().WriteSettings(new AgentSettings(saved?.Memory ?? MemorySettings(), message.RecallSettings!));
            SetRecallSettings(message.RecallSettings!);
            message.Reply!.TrySetResult("");
        }
        catch (Exception error)
        {
            // No publication preceded replacement. A refusal leaves the current
            // policy and the captured in-flight recall task unchanged.
            message.Reply!.TrySetException(error);
        }
    }

    private partial bool BeginRecall(Mail turn)
    {
        // Only a new model turn runs retrieval. Busy hints remain immediately
        // recorded input to that turn; tools, wake mail and compressors do not
        // recursively retrieve. Recovery already ran at the admission gate.
        if (turn.Kind != MailKind.Submit || IsMemoryCompressor())
        {
            return false;
        }
        var settings = RecallSettings();
        string query = turn.RecallQuery ?? (turn.Parts.IsDefaultOrEmpty ? ""
            : string.Join("\n", turn.Parts.OfType<TextPart>().Select(part => part.Text)));
        if (settings.Disabled || query.EnumerateRunes().Count() < settings.MinQueryLength
            || RecallText.Tokens(query, query: true).Length == 0)
        {
            return false;
        }
        var context = History().View();
        string instructions = Engine().ProjectInstructions(context);
        observationCancellation = new CancellationTokenSource();
        var cancellation = observationCancellation.Token;
        workerPending = true;
        // The scan is synchronous rooted I/O. Start it outside the mailbox too,
        // not only the asynchronous judge, so controls remain responsive.
        recallTask = Task.Run(() => RecallWorker(turn, query, context, instructions,
            settings, cancellation));
        return true;
    }

    private partial async Task RecallWorker(Mail turn, string query, Context context,
        string instructions, RecallSettings settings, CancellationToken cancellation)
    {
        Mail completion;
        try
        {
            var recaller = Recaller();
            double floor = settings.DisableJudge ? settings.MinScore : 0.5;
            var hits = RecallText.Available(recaller.Search(query, floor, cancellation), context, instructions);
            var attachment = await recaller.SelectAsync(hits, query, context, settings, cancellation);
            // Keep judge order and only identities whose content actually fit.
            // A capped attachment still binds the full source for the actor's
            // final check, without reconstructing source from quoted output.
            var selected = attachment is null ? ImmutableArray<RecallHit>.Empty
                : attachment.Hashes.Select(hash => hits.First(hit => hit.Snippet.Hash() == hash)).ToImmutableArray();
            completion = new Mail(MailKind.RecallDone, RecallTurn: turn,
                RecallCompletion: new RecallCompletion(selected, settings.MaxRecallBytes));
        }
        catch (Exception error)
        {
            completion = new Mail(MailKind.RecallDone, RecallTurn: turn, Error: error);
        }
        mailbox.Writer.TryWrite(completion);
    }

    private partial void ReceiveRecall(Mail message)
    {
        if (!ReferenceEquals(active, message.RecallTurn) || interruptPending || stopping)
        {
            return; // Never attach canceled work to the next turn.
        }
        if (message.Error is not null)
        {
            // A corpus/recovery failure is not an empty successful search. The
            // stateless judge handles its own bounded mechanical fallback.
            Finish(message.Error);
            return;
        }
        var completion = message.RecallCompletion!;
        var current = History().View();
        var available = RecallText.Available(completion.Selected, current,
            Engine().ProjectInstructions(current));
        var attachment = RecallText.Format(available, completion.ByteLimit);
        if (attachment is not null)
        {
            Record(new Event(EventKind.RecallAdded, parts: [new RecallPart(attachment)]));
        }
    }
}
