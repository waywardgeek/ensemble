using System.Collections.Immutable;

namespace Ensemble;

public sealed partial class Agent
{
    private partial bool BeginAutomaticCollection()
    {
        if (RequiresMemoryHandoff())
        {
            return false;
        }
        if (roundSnapshotsCollected)
        {
            return false;
        }
        var selected = Mcp()?.AutomaticTools(includeTurn: !turnSnapshotsCollected) ?? [];
        if (selected.IsEmpty)
        {
            turnSnapshotsCollected = true;
            roundSnapshotsCollected = true;
            return false;
        }
        var jobs = ImmutableArray.CreateBuilder<(McpToolDefinition Definition, Job Job)>();
        foreach (var automatic in selected)
        {
            // Admission and its recording happen on the actor; only the actual
            // external operation and completed-output observation run elsewhere.
            var job = Jobs().StartAutomatic(automatic);
            jobs.Add((automatic.Definition, job));
        }
        automaticPending = true;
        workerPending = true;
        _ = AutomaticWorker(jobs.ToImmutable(), observationCancellation!.Token);
        return true;
    }

    private partial async Task AutomaticWorker(ImmutableArray<(McpToolDefinition Definition, Job Job)> jobs,
        CancellationToken cancellation)
    {
        try
        {
            var results = ImmutableArray.CreateBuilder<AutomaticResult>();
            foreach (var item in jobs)
            {
                var result = await item.Job.CompletionAsync(cancellation);
                results.Add(new AutomaticResult(item.Definition, result));
            }
            mailbox.Writer.TryWrite(new Mail(MailKind.AutomaticDone, AutomaticResults: results.ToImmutable()));
        }
        catch (Exception error)
        {
            mailbox.Writer.TryWrite(new Mail(MailKind.AutomaticDone, Error: error));
        }
    }

    private partial void ReceiveAutomatic(Mail message)
    {
        automaticPending = false;
        if (interruptPending || stopping)
        {
            return; // Completed bytes remain on their Jobs, not in a later turn.
        }
        if (message.Error is not null)
        {
            Finish(message.Error);
            return;
        }
        var turn = ImmutableArray.CreateBuilder<Part>();
        var round = ImmutableArray.CreateBuilder<Part>();
        foreach (var snapshot in message.AutomaticResults)
        {
            var result = snapshot.Result;
            Record(new Event(EventKind.ToolObserved, callID: result.CallID, parts: result.Parts,
                isError: result.IsError, job: result.Job));
            var destination = snapshot.Definition.Ephemeral == "turn" ? turn : round;
            if (result.IsError)
            {
                // Failed bytes are diagnostics, not a usable snapshot. The
                // Job and observation above retain them in full for inspection.
                destination.Add(new TextPart($"Automatic snapshot {snapshot.Definition.Name} unavailable (error, job {result.Job!.Handle})."));
            }
            else
            {
                destination.Add(new TextPart($"Automatic snapshot {snapshot.Definition.Name} (complete, job {result.Job!.Handle}):"));
                destination.AddRange(result.Parts);
            }
        }
        if (!turnSnapshotsCollected)
        {
            turnSnapshots = turn.ToImmutable();
        }
        roundSnapshots = round.ToImmutable();
        turnSnapshotsCollected = true;
        roundSnapshotsCollected = true;
    }
}
