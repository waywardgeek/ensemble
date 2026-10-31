using System.Collections.Immutable;
using System.Text;

namespace Ensemble;

public sealed partial class Agent
{
    internal partial ImmutableArray<Event> SelectImmediateCuts(Context context, long requestSequence, int threshold)
    {
        // Every successful response has a canonical assistant entry, including
        // an empty response. Only the results after the latest such entry can
        // be newly consumed. Recording the next response closes this boundary
        // permanently: a later text-only turn or snapshot replay cannot revisit
        // a batch that was kept. No second batch registry or saved cursor exists.
        var response = context.Dialogue.LastOrDefault(entry => entry.Seq < requestSequence
            && entry.Actor == Actor.Agent);
        if (response is null)
        {
            return [];
        }
        var calls = new Dictionary<string, ToolCallPart>(StringComparer.Ordinal);
        foreach (var call in response.Parts.OfType<ToolCallPart>())
        {
            if (!calls.TryAdd(call.CallID, call))
            {
                return [];
            }
        }
        if (calls.Count == 0)
        {
            return [];
        }

        var cuts = ImmutableArray.CreateBuilder<Event>();
        foreach (var entry in context.Dialogue.Where(entry => entry.Seq > response.Seq && entry.Seq < requestSequence))
        {
            bool stubEntry = false;
            foreach (var part in entry.Parts)
            {
                if (part is ToolCallPart)
                {
                    return []; // Do not infer a new batch from an orphaned call.
                }
                if (part is not ToolResultPart result)
                {
                    continue;
                }
                if (!calls.Remove(result.CallID, out var call))
                {
                    return []; // Duplicate/orphan results have no safe batch identity.
                }
                bool keepControl = Tools().Any(tool => tool.Name() == call.Name && tool.IsKeepResults());
                if (keepControl || result.Parts.All(value => value is RedactedPart))
                {
                    continue;
                }
                int bytes = Encoding.UTF8.GetByteCount(JsonCodec.WritePart(result).ToJsonString());
                if (bytes > threshold)
                {
                    stubEntry = true;
                }
            }
            if (stubEntry)
            {
                cuts.Add(new Event(EventKind.Redacted, redactFrom: entry.Seq, redactTo: entry.Seq,
                    level: Redaction.Result, reason: "once-only tool result stubbing"));
            }
        }
        return calls.Count == 0 ? cuts.ToImmutable() : [];
    }

    private partial void StageKeepResults(ToolCallPart call, ToolResultPart result)
    {
        // This is called only after the model-dispatched result is durable.
        // A same-named custom tool, bad arguments or an undispatched call cannot
        // veto policy. Host controls are not part of the model's response.
        if (!result.IsError && Tools().Any(tool => tool.Name() == call.Name && tool.IsKeepResults()))
        {
            pendingResultStubs = [];
        }
    }

    private partial void ApplyResultStubs()
    {
        var cuts = pendingResultStubs;
        pendingResultStubs = [];
        foreach (var cut in cuts)
        {
            Record(cut);
        }
    }
}
