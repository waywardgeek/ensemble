using System.Collections.Immutable;
using System.Diagnostics;
using System.Text;
using System.Text.RegularExpressions;

namespace Ensemble;

public sealed partial class Agent
{
    public partial async Task<ChildOutput> ReadOutputAsync(long cursor, bool wait,
        CancellationToken cancellation)
    {
        if (cursor < 0)
        {
            throw new ArgumentOutOfRangeException(nameof(cursor), "Cursor must be nonnegative.");
        }
        // Hold the actual History reference, not a name that could resolve to a
        // newly created child after deletion. Keep the last immutable snapshot
        // locally because logical History destruction releases its event list.
        History history = History();
        ImmutableArray<Event> retained = history.Events();
        if (cursor > (retained.IsEmpty ? 0 : retained[^1].Seq))
        {
            throw new ArgumentException("Cursor is beyond this child's History.", nameof(cursor));
        }
        while (true)
        {
            cancellation.ThrowIfCancellationRequested();
            AgentWorkStatus status = await WorkStatusAsync(cancellation);
            ImmutableArray<Event> current = history.Events();
            if (!current.IsEmpty)
            {
                retained = current;
            }
            ChildOutput output = CaptureOutput(retained, cursor, status);
            if (!wait || output.Output.Length != 0 || status.Activity != AgentActivity.Busy)
            {
                return output;
            }
            await Task.Delay(100, cancellation);
        }
    }

    private partial ChildOutput CaptureOutput(ImmutableArray<Event> events, long cursor, AgentWorkStatus status)
    {
        long next = events.IsEmpty ? 0 : events[^1].Seq;
        if (cursor > next)
        {
            throw new ArgumentException("Cursor is beyond this child's History.", nameof(cursor));
        }
        var text = new StringBuilder();
        foreach (Event fact in events)
        {
            if (fact.Seq <= cursor || fact.Kind is not (EventKind.ResponseEnded
                or EventKind.ToolReturned or EventKind.ToolObserved))
            {
                continue;
            }
            foreach (TextPart part in fact.Parts.OfType<TextPart>())
            {
                text.Append(part.Text);
            }
            if (fact.Parts.OfType<TextPart>().Any())
            {
                text.Append('\n');
            }
        }
        return new ChildOutput(childID ?? "", next, text.ToString(), status);
    }

    public partial async Task<ChildChange> WaitForChildChangeAsync(IEnumerable<string>? childIDs,
        string? pattern, double timeoutSeconds, double stuckSeconds, CancellationToken cancellation)
    {
        if (!double.IsFinite(timeoutSeconds) || timeoutSeconds < 0
            || !double.IsFinite(stuckSeconds) || stuckSeconds < 0)
        {
            throw new ArgumentException("Timeout and stuck intervals must be finite and nonnegative.");
        }
        Regex? match = string.IsNullOrEmpty(pattern) ? null
            : new Regex(pattern, RegexOptions.CultureInvariant, TimeSpan.FromMilliseconds(100));
        string[] ids = childIDs?.ToArray() ?? [];
        Agent[] selected;
        lock (ingress)
        {
            ObjectDisposedException.ThrowIf(admissionClosed || destroyed, this);
            if (ids.Distinct(StringComparer.Ordinal).Count() != ids.Length)
            {
                throw new ArgumentException("Duplicate child ID in wait.");
            }
            selected = ids.Length == 0 ? children.ToArray() : ids.Select(id =>
                children.FirstOrDefault(child => child.childID == id)
                ?? throw new ArgumentException("No owned child has that ID.")).ToArray();
        }
        // Capture real owners once. No child is resurrected and no new sibling
        // joins this admitted wait. These local baselines are not a live registry.
        var histories = new History?[selected.Length];
        for (int index = 0; index < selected.Length; index++)
        {
            Agent child = selected[index];
            lock (child.ingress)
            {
                histories[index] = child.history;
            }
        }
        ImmutableArray<Event>[] retained = histories.Select(history => history?.Events() ?? []).ToArray();
        var baseline = new AgentWorkStatus?[selected.Length];
        long[] cursors = retained.Select(events => events.IsEmpty ? 0 : events[^1].Seq).ToArray();
        long[] streams = selected.Select(child => Interlocked.Read(ref child.streamProgress)).ToArray();
        var clock = Stopwatch.StartNew();
        bool initial = true;
        while (true)
        {
            cancellation.ThrowIfCancellationRequested();
            var rows = ImmutableArray.CreateBuilder<ChildOutput>();
            string? reason = null;
            for (int index = 0; index < selected.Length; index++)
            {
                Agent child = selected[index];
                AgentWorkStatus status = await child.WorkStatusAsync(cancellation);
                ImmutableArray<Event> current = histories[index]?.Events() ?? [];
                if (!current.IsEmpty)
                {
                    retained[index] = current;
                }
                ChildOutput row = child.CaptureOutput(retained[index], 0, status);
                rows.Add(row);
                long stream = Interlocked.Read(ref child.streamProgress);
                if (FindChild(child.childID!) != child)
                {
                    reason ??= "removed";
                }
                else if (status.Activity is AgentActivity.Stopping or AgentActivity.Stopped)
                {
                    reason ??= "stopped";
                }
                if (match is not null && match.IsMatch(row.Output))
                {
                    reason ??= "pattern";
                }
                if (row.Cursor != cursors[index] || stream != streams[index]
                    || (!initial && row.Status != baseline[index]))
                {
                    reason ??= "changed";
                }
                if (initial)
                {
                    baseline[index] = row.Status;
                }
                if (stuckSeconds > 0 && status.Activity == AgentActivity.Busy
                    && clock.Elapsed.TotalSeconds >= stuckSeconds)
                {
                    reason ??= "stuck";
                }
            }
            ImmutableArray<QuestionInfo> questions;
            // Match actual sender relationships, not a reused local ID. Never
            // take ingress while holding the Ensemble's question graph lock.
            lock (Parent().questionGate)
            {
                questions = incomingQuestions.Where(question => selected.Contains(question.sender))
                    .Select(question => question.Snapshot()).ToImmutableArray();
            }
            if (!questions.IsEmpty)
            {
                reason = "question";
            }
            if (reason is not null || selected.Length == 0)
            {
                return new ChildChange(reason ?? "empty", rows.ToImmutable(), questions);
            }
            if (timeoutSeconds > 0 && clock.Elapsed.TotalSeconds >= timeoutSeconds)
            {
                return new ChildChange("timeout", rows.ToImmutable(), questions);
            }
            initial = false;
            await Task.Delay(100, cancellation);
        }
    }
}
