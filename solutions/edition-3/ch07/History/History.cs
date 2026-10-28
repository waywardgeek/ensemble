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
            return new HistorySnapshot(events.ToImmutableArray(), context);
        }
    }
    public partial void StartLog(string path)
    {
        lock (gate)
        {
            ObjectDisposedException.ThrowIf(destroyed, this);
            if (writer is not null || events.Count != 0)
            {
                throw new InvalidOperationException("Start the log before recording events");
            }

            var opened = new StreamWriter(path, false, new UTF8Encoding(false));
            try
            {
                opened.WriteLine("{\"log_version\":1}");
                opened.Flush();
                writer = opened;
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
        lock (gate)
        {
            ObjectDisposedException.ThrowIf(destroyed, this);
            var sequence = events.Count == 0 ? 1 : checked(events[^1].Seq + 1);
            var next = value.Numbered(sequence, DateTimeOffset.UtcNow);
            var reduced = Reduce(context, next);
            // Serialize/write/flush before publication. Disk durability beyond Flush is
            // not promised here; later persistence chapters add stronger recovery rules.
            var line = JsonCodec.WriteEvent(next);
            if (writer is not null)
            {
                writer.WriteLine(line);
                writer.Flush();
            }
            events.Add(next);
            context = reduced;
            // A notification is a lossy wake-up, not a second log. Readers can
            // recover all finalized facts from Snapshot after missed wake-ups.
            parent?.Publish(new Observation(ObservationKind.EventRecorded, next.Seq, Fact: next));
            return next;
        }
    }

    public partial void Load(string path)
    {
        lock (gate)
        {
            ObjectDisposedException.ThrowIf(destroyed, this);
            if (writer is not null)
            {
                throw new InvalidOperationException("Cannot replace an active output log");
            }

            // Stage both the facts and their reduction locally. A malformed late line
            // must not replace a valid current conversation with a partial one.
            using var reader = new StreamReader(path);
            var loaded = new List<Event>();
            var view = new Context();
            long previous = 0;
            int lineNumber = 0;
            while (reader.ReadLine() is string line)
            {
                lineNumber++;
                // An explicit unknown version is refused. Historical headerless logs
                // are current format, so their first event must not be discarded.
                if (lineNumber == 1)
                {
                    using var first = JsonDocument.Parse(line);
                    if (first.RootElement.TryGetProperty("log_version", out var version))
                    {
                        if (!version.TryGetInt32(out var n) || n != 1)
                        {
                            throw new FormatException("Unsupported log version");
                        }

                        continue;
                    }
                }
                if (string.IsNullOrWhiteSpace(line))
                {
                    continue;
                }

                try
                {
                    var value = JsonCodec.ReadEvent(line);
                    if (value.Seq <= previous)
                    {
                        throw new FormatException("Sequence must increase strictly");
                    }

                    previous = value.Seq;
                    view = Reduce(view, value);
                    loaded.Add(value);
                }
                catch (Exception ex) when (ex is JsonException or FormatException or ArgumentException or InvalidOperationException or KeyNotFoundException)
                {
                    throw new FormatException($"Invalid log at line {lineNumber}: {ex.Message}", ex);
                }
            }
            events.Clear();
            events.AddRange(loaded);
            context = view;
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
                parent?.RemoveHistory(this);
                events.Clear();
                context = new Context();
            }
        }
    }

    public static partial Context Reduce(Context before, Event value)
    {
        // Work on a fresh collection. Existing Context snapshots remain usable.
        var dialogue = before.Dialogue.ToList();
        var ephemeral = before.Ephemeral;
        var usage = before.Usage;
        var phase = before.Phase;
        switch (value.Kind)
        {
            case EventKind.HintReceived:
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
                dialogue.Add(new Entry(value.Seq, Actor.Agent, value.Parts));
                usage = usage.Add(value.Usage ?? new Usage());
                if (phase == Phase.InFlight)
                {
                    phase = value.Parts.Any(p => p is ToolCallPart) ? Phase.ToolsPending : Phase.Idle;
                }

                break;
            case EventKind.ToolReturned:
                dialogue.Add(new Entry(value.Seq, Actor.Tool, [new ToolResultPart(value.CallID, value.Parts, value.IsError, value.Job)]));
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
        return new Context(dialogue, ephemeral, phase, usage);
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
    private static RedactedPart Stub(Part p, string reason) => new(p switch
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
            if (!Selected(entry))
            {
                result.Add(entry);
                continue;
            }
            if (value.Level == Redaction.Summary)
            {
                if (!inserted)
                {
                    var count = dialogue.Where(Selected).SelectMany(e => e.Parts).Sum(Size);
                    var replacement = value.Parts;
                    if (replacement.IsEmpty)
                    {
                        replacement = [new RedactedPart("summary", count, value.Reason)];
                    }
                    result.Add(new Entry(value.RedactFrom, Actor.System, replacement));
                    inserted = true;
                }
                continue;
            }
            // Filters keep the entry's position but remove the selected category.
            // Only result redaction substitutes stubs for the removed payload.
            var parts = entry.Parts
                .Where(p => value.Level != Redaction.Tool || p is not ToolCallPart and not ToolResultPart)
                .Where(p => value.Level != Redaction.Dialogue || p is ToolCallPart or ToolResultPart)
                .Select(p => value.Level == Redaction.Result && p is ToolResultPart t
                    ? (Part)new ToolResultPart(t.CallID,
                        t.Job is null
                            ? t.Parts.Select(x => Stub(x, value.Reason))
                            : [new RedactedPart("tool_result", t.Job.Bytes, value.Reason,
                                new BlobRef(RefKind.Path, t.Job.OutputPath))],
                        t.IsError, t.Job)
                    : p);
            result.Add(new Entry(entry.Seq, entry.Actor, parts));
        }
        return result;
    }
}
