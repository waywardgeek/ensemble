using System.Collections.Immutable;
using System.Text;
using System.Text.Json;
namespace Ensemble;

public sealed partial class History
{
    public partial History(Agent parent)
    {
        events = [];
        context = new Context();
        parent.AddHistory(this);
    }

    public partial Agent Parent() => parent ?? throw new ObjectDisposedException(nameof(History));
    public partial Context View()
    {
        lock (gate)
        {
            ObjectDisposedException.ThrowIf(destroyed, this);
            return context;
        }
    }

    public partial ImmutableArray<Event> Events()
    {
        lock (gate)
        {
            return events.ToImmutableArray();
        }
    }

    public partial HistorySnapshot Snapshot()
    {
        lock (gate)
        {
            ObjectDisposedException.ThrowIf(destroyed, this);
            return new HistorySnapshot(events.ToImmutableArray(), context, foldedSequence);
        }
    }
    public partial void StartLog(string path)
    {
        lock (saveGate)
        lock (gate)
        {
            ObjectDisposedException.ThrowIf(destroyed, this);
            if (writer is not null)
            {
                throw new InvalidOperationException("Output log already open");
            }

            var files = Parent().Sandbox();
            string destination = files.Resolve(path);
            if (destination == journalPath)
            {
                throw new ArgumentException("Diagnostic log must not replace the recovery journal");
            }
            var opened = new StreamWriter(files.OpenFile(destination, FileMode.Create, FileAccess.Write), new UTF8Encoding(false));
            try
            {
                opened.WriteLine("{\"log_version\":1}");
                foreach (var fact in events)
                {
                    opened.WriteLine(JsonCodec.WriteEvent(fact));
                }
                opened.Flush();
                writer = opened;
                writerPath = destination;
            }
            catch
            {
                opened.Dispose();
                throw;
            }
        }
    }

    public partial Event Append(Event value)
    {
        lock (saveGate)
        {
            Event next;
            Context reduced;
            lock (gate)
            {
                ObjectDisposedException.ThrowIf(destroyed, this);
                if (appendFaulted)
                {
                    throw new IOException("History append previously failed; recover before writing again");
                }
                next = value.Numbered(checked(foldedSequence + 1), DateTimeOffset.UtcNow);
                reduced = Apply(context, next);
            }
            string line = JsonCodec.WriteEvent(next);
            try
            {
                // The journal is the recovery authority; the diagnostic writer is
                // secondary. Flush closes the process-crash window, not a promise
                // about loss of power. No reader sees this fact until both succeed.
                if (journalPath is not null)
                {
                    using var output = Parent().Sandbox().OpenFile(journalPath, FileMode.Append, FileAccess.Write);
                    using var journal = new StreamWriter(output, new UTF8Encoding(false));
                    journal.WriteLine(line);
                    journal.Flush();
                }
                if (writer is not null)
                {
                    writer.WriteLine(line);
                    writer.Flush();
                }
            }
            catch
            {
                // A failed write may have left a complete or partial record. Do
                // not reuse its sequence or append behind unknown file contents.
                appendFaulted = true;
                throw;
            }
            lock (gate)
            {
                events.Add(next);
                context = reduced;
                foldedSequence = next.Seq;
                parent?.Publish(new Observation(ObservationKind.EventRecorded, next.Seq, Fact: next));
            }
            return next;
        }
    }

    public partial void Load(string path)
    {
        lock (saveGate)
        lock (gate)
        {
            ObjectDisposedException.ThrowIf(destroyed, this);
            if (writer is not null || journalPath is not null)
            {
                throw new InvalidOperationException("Cannot replace active History writers");
            }

            // Stage both the facts and their reduction locally. A malformed late line
            // must not replace a valid current conversation with a partial one.
            var loaded = ReadLog(path);
            var view = new Context();
            long previous = 0;
            foreach (var value in loaded)
            {
                view = Apply(view, value);
                previous = value.Seq;
            }
            events.Clear();
            events.AddRange(loaded);
            context = view;
            foldedSequence = previous;
        }
    }

    public partial void Dump(TextWriter destination)
    {
        lock (gate)
        {
            ObjectDisposedException.ThrowIf(destroyed, this);
            destination.WriteLine("{\"log_version\":1}");
            foreach (var value in events)
            {
                destination.WriteLine(JsonCodec.WriteEvent(value));
            }
        }
    }

    public partial void Destroy()
    {
        lock (saveGate)
        lock (gate)
        {
            if (destroyed)
            {
                return;
            }

            destroyed = true;
            try
            {
                writer?.Dispose();
            }
            finally
            {
                writer = null;
                writerPath = null;
                parent?.RemoveHistory(this);
                events.Clear();
                context = new Context();
            }
        }
    }

    public static partial Context Reduce(Context before, Event value)
    {
        ValidateCompressionEvent(before, value);
        if (ReductionProblem(before, value) is not null)
        {
            return before;
        }
        // Work on a fresh collection. Existing Context snapshots remain usable.
        var compression = before.Compression;
        var compressionOutcome = before.CompressionOutcome;
        var dialogue = before.Dialogue.ToList();
        var memories = before.Memories.ToList();
        var ephemeral = before.Ephemeral;
        var usage = before.Usage;
        var phase = before.Phase;
        switch (value.Kind)
        {
            case EventKind.CompressionLaunched:
                compression = new PendingCompression(value.Compression!);
                break;
            case EventKind.CompressionFinished:
                compressionOutcome = value.Outcome;
                compression = null;
                break;
            case EventKind.MemoryPopulated:
                if (value.TaskID.Length != 0)
                {
                    var removal = compression!.Removal!;
                    if (removal.Kind == EventKind.Redacted)
                    {
                        // The band replaces the conversation, not instructions.
                        // No source leaves the projection before this commit.
                        dialogue.RemoveAll(entry => entry.Seq >= removal.RedactFrom && entry.Seq <= removal.RedactTo
                            && !entry.Parts.Any(part => part is SkillPart));
                    }
                    else
                    {
                        var retired = removal.Retirement!;
                        memories.RemoveAll(item => item.Band == retired.Band && retired.IDs.Contains(item.ID));
                    }
                    compression = new PendingCompression(compression.Launch, removal, value.Memory);
                }
                var memory = value.Memory!;
                // Human edits replace a stable identity, not append duplicates.
                memories.RemoveAll(item => item.Band == memory.Band && item.ID == memory.ID);
                memories.Add(memory);
                break;
            case EventKind.MemoryRetired:
                if (value.TaskID.Length != 0)
                {
                    compression = new PendingCompression(compression!.Launch, value);
                    break;
                }
                var retirement = value.Retirement!;
                memories.RemoveAll(item => item.Band == retirement.Band
                    && (retirement.IDs.IsEmpty || retirement.IDs.Contains(item.ID)));
                break;
            case EventKind.SkillUnloadPending:
                // Diagnostic retention may discard the source event. Preserve its
                // effect in the authoritative snapshot without editing prior values.
                var pendingEntries = new List<Entry>();
                foreach (var entry in dialogue)
                {
                    var updatedParts = new List<Part>();
                    foreach (var part in entry.Parts)
                    {
                        if (part is SkillPart skill && skill.Name == value.Name)
                        {
                            updatedParts.Add(new SkillPart(skill.Name, skill.Body,
                                skill.Tools, skill.Primary, pending: true));
                        }
                        else
                        {
                            updatedParts.Add(part);
                        }
                    }
                    pendingEntries.Add(new Entry(entry.Seq, entry.Actor,
                        updatedParts, entry.ResponseID, entry.PartIDs));
                }
                dialogue = pendingEntries;
                break;
            case EventKind.RuntimeConfigured:
                // A restart keeps conversation, not yesterday's capability grants.
                dialogue.RemoveAll(entry => entry.Parts.Any(part => part is SkillPart));
                if (!value.Parts.IsEmpty)
                {
                    dialogue.Add(new Entry(value.Seq, Actor.System, value.Parts));
                }
                phase = Phase.Idle;
                break;
            case EventKind.HintReceived:
                // Hints are human conversation, not primary or skill instructions.
                // Preserve an explicitly recorded system actor for ordinary
                // historical system entries; never infer role from event position.
                dialogue.Add(new Entry(value.Seq, value.Actor!.Value, value.Parts));
                break;
            case EventKind.RecallAdded:
                // Retrieved text has the user's data role but never changes the
                // user's entry, phase, usage, ephemera or tool authority.
                dialogue.Add(new Entry(value.Seq, Actor.Human, value.Parts));
                break;
            case EventKind.HandoffRecorded:
                dialogue = HandoffDialogue(dialogue, value);
                break;
            case EventKind.SkillLoaded:
                dialogue.Add(new Entry(value.Seq, Actor.System, value.Parts));
                break;
            case EventKind.MessageReceived:
                if (phase == Phase.Idle)
                {
                    phase = Phase.InputPending;
                }

                if (value.Actor == Actor.System)
                {
                    ephemeral = value.Parts.Where(p => p is not TextPart { Text.Length: 0 }).ToImmutableArray();
                }
                else
                {
                    dialogue.Add(new Entry(value.Seq, value.Actor!.Value, value.Parts));
                }
                break;
            case EventKind.RequestSent:
                // An attempted request consumes the slot, even if HTTP later fails.
                ephemeral = [];
                if (phase == Phase.InputPending)
                {
                    phase = Phase.InFlight;
                }

                break;
            case EventKind.ResponseEnded:
                dialogue.Add(new Entry(value.Seq, Actor.Agent, value.Parts, value.ResponseID, value.PartIDs));
                usage = usage.Add(value.Usage ?? new Usage());
                if (phase == Phase.InFlight)
                {
                    phase = value.Parts.Any(p => p is ToolCallPart) ? Phase.ToolsPending : Phase.Idle;
                }

                break;
            case EventKind.ToolReturned:
                dialogue.Add(new Entry(value.Seq, Actor.Tool, [new ToolResultPart(value.CallID, value.Parts, value.IsError, value.Job, value.ToolReferences)]));
                if (phase == Phase.ToolsPending && LastCallsComplete(dialogue))
                {
                    phase = Phase.InputPending;
                }

                break;
            case EventKind.TurnInterrupted:
            case EventKind.TurnEnded:
                phase = Phase.Idle;
                break;
            case EventKind.Redacted:
                if (value.TaskID.Length != 0)
                {
                    compression = new PendingCompression(compression!.Launch, value);
                    break;
                }
                dialogue = Redact(dialogue, value);
                break;
            case EventKind.ErrorOccurred:
                if (phase == Phase.InFlight)
                {
                    phase = Phase.Idle;
                }

                break;
                // Start and dispatch facts affect neither dialogue nor turn state.
                // Every other known event/state pair keeps the existing phase.
        }
        return new Context(dialogue, ephemeral, phase, usage, memories, compression, compressionOutcome);
    }

    // Only the latest assistant call group determines whether this turn can
    // continue. Old completed calls elsewhere in history are not pending work.
    private static bool LastCallsComplete(List<Entry> dialogue)
    {
        var index = dialogue.FindLastIndex(e => e.Actor == Actor.Agent && e.Parts.Any(p => p is ToolCallPart));
        if (index < 0)
        {
            return false;
        }

        var calls = dialogue[index].Parts.OfType<ToolCallPart>().Select(p => p.CallID).ToHashSet(StringComparer.Ordinal);
        foreach (var entry in dialogue.Skip(index + 1))
        {
            foreach (var result in entry.Parts.OfType<ToolResultPart>())
            {
                calls.Remove(result.CallID);
            }
        }

        return calls.Count == 0;
    }

    private static long Size(Part p) => p switch
    {
        TextPart t => Encoding.UTF8.GetByteCount(t.Text),
        ToolResultPart t => t.Parts.Sum(Size),
        RedactedPart r => r.Bytes,
        BlobPart => 0, // The log contains a locator, not the blob's unknown bytes.
        _ => Encoding.UTF8.GetByteCount(JsonCodec.WritePart(p).ToJsonString())
    };
    private static RedactedPart Stub(Part p, string reason) => p is RedactedPart retained ? retained : new(p switch
    {
        TextPart => "text",
        BlobPart => "blob",
        OpaquePart => "opaque",
        ToolCallPart => "tool_call",
        ToolResultPart => "tool_result",
        _ => "redacted"
    }, Size(p), reason, (p as BlobPart)?.Ref);
    private static List<Entry> Redact(List<Entry> dialogue, Event value)
    {
        bool Selected(Entry e) => e.Seq >= value.RedactFrom && e.Seq <= value.RedactTo;
        var result = new List<Entry>();
        bool inserted = false;
        foreach (var entry in dialogue)
        {
            if (!Selected(entry) || entry.Parts.Any(part => part is SkillPart))
            {
                result.Add(entry);
                continue;
            }
            if (value.Level == Redaction.Summary)
            {
                if (!inserted)
                {
                    var count = dialogue.Where(Selected)
                        .Where(candidate => !candidate.Parts.Any(part => part is SkillPart))
                        .SelectMany(candidate => candidate.Parts)
                        .Sum(Size);
                    var replacement = value.Parts;
                    if (replacement.IsEmpty)
                    {
                        replacement = [new RedactedPart("summary", count, value.Reason)];
                    }
                    // A skill at the range start is preserved. Give the summary
                    // the first replaced entry's identity, not that skill's ID.
                    result.Add(new Entry(entry.Seq, Actor.System, replacement));
                    inserted = true;
                }
                continue;
            }
            // Filters keep the entry's position but remove the selected category.
            // Only result redaction substitutes stubs for the removed payload.
            var kept = Enumerable.Range(0, entry.Parts.Length)
                .Where(index => value.Level != Redaction.Tool || entry.Parts[index] is not ToolCallPart and not ToolResultPart)
                .Where(index => value.Level != Redaction.Dialogue || entry.Parts[index] is ToolCallPart or ToolResultPart)
                .ToArray();
            var parts = kept.Select(index => entry.Parts[index])
                .Select(p => value.Level == Redaction.Result && p is ToolResultPart t
                    ? (Part)new ToolResultPart(t.CallID,
                        t.Job is null
                            ? t.Parts.Select(x => Stub(x, value.Reason))
                            : [new RedactedPart("tool_result", t.Job.Bytes, value.Reason,
                                new BlobRef(RefKind.Path, t.Job.OutputPath))],
                        t.IsError, t.Job, t.ToolReferences)
                    : p);
            var ids = entry.PartIDs.IsEmpty ? [] : kept.Select(index => entry.PartIDs[index]).ToImmutableArray();
            result.Add(new Entry(entry.Seq, entry.Actor, parts, entry.ResponseID, ids));
        }
        return result;
    }
}
