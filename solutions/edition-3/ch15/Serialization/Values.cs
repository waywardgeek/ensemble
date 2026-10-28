using System.Collections.Immutable;
using System.Text.Json;

namespace Ensemble;

// Value constructors validate at the boundary and clone borrowed JSON documents.
// No parent bookkeeping is appropriate for these immutable values.
public sealed partial class Configuration
{
    public partial Configuration(Vendor vendor, string model, string baseUrl, string apiKey, string resolvedModel, bool requireToolExplanation, int toolRoundLimit, string primaryInstructions, bool disableStreaming, bool deferredTools, bool stubsToolResults, bool cacheBreakpoints, string promptCacheKey)
    {
        if (!Enum.IsDefined(vendor))
        {
            throw new ArgumentException("Unknown vendor");
        }

        if (string.IsNullOrWhiteSpace(model))
        {
            throw new ArgumentException("Model must not be empty");
        }

        if (!Uri.TryCreate(baseUrl, UriKind.Absolute, out var uri) || (uri.Scheme != "http" && uri.Scheme != "https"))
        {
            throw new ArgumentException("Base URL must be absolute HTTP(S)");
        }

        if (deferredTools && vendor != Vendor.Anthropic)
        {
            throw new ArgumentException("Deferred tools require the Anthropic adapter");
        }
        if (cacheBreakpoints && vendor != Vendor.Anthropic)
        {
            throw new ArgumentException("Block cache breakpoints require the Anthropic adapter");
        }
        ArgumentNullException.ThrowIfNull(promptCacheKey);
        if (promptCacheKey.Length > 0 && vendor != Vendor.OpenAI)
        {
            throw new ArgumentException("Prompt cache keys require the OpenAI adapter");
        }
        CacheBreakpoints = cacheBreakpoints;
        PromptCacheKey = promptCacheKey;
        DeferredTools = deferredTools;
        StubsToolResults = stubsToolResults;
        Vendor = vendor;
        Model = model;
        BaseUrl = baseUrl.TrimEnd('/');
        ApiKey = apiKey;
        ArgumentNullException.ThrowIfNull(primaryInstructions);
        PrimaryInstructions = primaryInstructions;
        ToolRoundLimit = toolRoundLimit == 0 ? 200 : toolRoundLimit;
        RequireToolExplanation = requireToolExplanation;
        DisableStreaming = disableStreaming;
        ResolvedModel = string.IsNullOrEmpty(resolvedModel) ? model : resolvedModel;
    }

    public partial Provenance Target() => new(Vendor, ResolvedModel, Vendor switch
    {
        Vendor.Anthropic => Surface.Messages,
        Vendor.OpenAI => Surface.ChatCompletions,
        _ => Surface.GenerateContent
    });
}
public sealed partial class Provenance
{
    public partial Provenance(Vendor vendor, string model, Surface surface)
    {
        if (!Enum.IsDefined(vendor) || !Enum.IsDefined(surface) || string.IsNullOrEmpty(model))
        {
            throw new ArgumentException("Invalid provenance");
        }

        Vendor = vendor;
        Model = model;
        Surface = surface;
    }

    public partial bool Matches(Provenance target) => Vendor == target.Vendor && Model == target.Model && Surface == target.Surface;
}
public sealed partial class BlobRef
{
    public partial BlobRef(RefKind kind, string locator)
    {
        if (!Enum.IsDefined(kind) || string.IsNullOrEmpty(locator))
        {
            throw new ArgumentException("Invalid blob reference");
        }

        Kind = kind;
        Locator = locator;
    }
}
public sealed partial class Usage
{
    public partial Usage(long? input, long? cacheWrite, long? cacheRead, long? output)
    {
        if (input < 0 || cacheWrite < 0 || cacheRead < 0 || output < 0)
        {
            throw new ArgumentException("Negative usage");
        }

        Input = input;
        CacheWrite = cacheWrite;
        CacheRead = cacheRead;
        Output = output;
    }

    public partial Usage Add(Usage other)
    {
        var input = Sum(Input, other.Input);
        var cacheWrite = Sum(CacheWrite, other.CacheWrite);
        var cacheRead = Sum(CacheRead, other.CacheRead);
        var output = Sum(Output, other.Output);
        return new Usage(input, cacheWrite, cacheRead, output);
    }

    public partial long? AllInput() => Sum(Sum(Input, CacheWrite), CacheRead);

    // Overflow is unknown accounting, not wraparound into a valid-looking count.
    private static long? Sum(long? a, long? b)
    {
        if (a is not long x || b is not long y)
        {
            return null;
        }
        if (x > long.MaxValue - y)
        {
            return null;
        }
        return x + y;
    }
}
public abstract partial class Part
{
    protected partial Part()
    {
    }
}
public sealed partial class TextPart
{
    public partial TextPart(string text, Provenance? from, JsonElement? opaque)
    {
        if (opaque is not null && from is null)
        {
            throw new ArgumentException("Opaque text material requires provenance");
        }
        From = from;
        Opaque = opaque?.Clone();
        Text = text;
    }
}
public sealed partial class BlobPart
{
    public partial BlobPart(string mime, BlobRef reference)
    {
        Mime = mime;
        Ref = reference;
    }
}
public sealed partial class OpaquePart
{
    public partial OpaquePart(Provenance from, JsonElement data)
    {
        From = from;
        // The parser will dispose its JsonDocument; this value must outlive it.
        Data = data.Clone();
    }
}
public sealed partial class ToolCallPart
{
    public partial ToolCallPart(string callID, Provenance from, string name, JsonElement args, JsonElement? opaque)
    {
        if (string.IsNullOrEmpty(callID) || string.IsNullOrEmpty(name))
        {
            throw new ArgumentException("Tool call needs ID and name");
        }

        CallID = callID;
        From = from;
        Name = name;
        Args = args.Clone();
        Opaque = opaque?.Clone();
    }
}
public sealed partial class ToolResultPart
{
    public partial ToolResultPart(string callID, IEnumerable<Part> parts, bool isError, JobInfo? job, IEnumerable<string>? toolReferences)
    {
        if (string.IsNullOrEmpty(callID))
        {
            throw new ArgumentException("Tool result needs an ID");
        }

        CallID = callID;
        // Copy the sequence so callers cannot change an already recorded result.
        Parts = parts.ToImmutableArray();
        IsError = isError;
        Job = job;
        ToolReferences = (toolReferences ?? []).Distinct(StringComparer.Ordinal).ToImmutableArray();
        if (ToolReferences.Any(string.IsNullOrWhiteSpace) || (isError && !ToolReferences.IsEmpty))
        {
            throw new ArgumentException("Discovery references require successful named grants");
        }
    }
}
public sealed partial class RedactedPart
{
    public partial RedactedPart(string kind, long bytes, string reason, BlobRef? reference)
    {
        if (bytes < 0)
        {
            throw new ArgumentException("Negative redaction size");
        }

        Kind = kind;
        Bytes = bytes;
        Reason = reason;
        Ref = reference;
    }
}
public sealed partial class Event
{
    public partial Event(EventKind kind, long seq, DateTimeOffset? time, Actor? actor, IEnumerable<Part>? parts,
        Provenance? from, Usage? usage, string callID, string name, JsonElement? args, bool isError,
        string error, long redactFrom, long redactTo, Redaction? level, string reason, JobInfo? job,
        long responseID, IEnumerable<int>? partIDs, IEnumerable<string>? toolReferences)
    {
        if (!Enum.IsDefined(kind) || (actor.HasValue && !Enum.IsDefined(actor.Value)))
        {
            throw new ArgumentException("Unknown event or actor");
        }

        if (seq < 0)
        {
            throw new ArgumentException("Negative sequence");
        }

        Kind = kind;
        Seq = seq;
        Time = time;
        Actor = actor;
        Parts = parts?.ToImmutableArray() ?? [];
        From = from;
        Usage = usage;
        CallID = callID;
        Name = name;
        Args = args?.Clone();
        IsError = isError;
        Error = error;
        RedactFrom = redactFrom;
        RedactTo = redactTo;
        Level = level;
        Reason = reason;
        Job = job;
        ResponseID = responseID;
        PartIDs = partIDs?.ToImmutableArray() ?? [];
        ToolReferences = (toolReferences ?? []).Distinct(StringComparer.Ordinal).ToImmutableArray();
        if (!ToolReferences.IsEmpty && (kind is not (EventKind.ToolReturned or EventKind.ToolObserved)
            || isError || ToolReferences.Any(string.IsNullOrWhiteSpace)))
        {
            throw new ArgumentException("Discovery references require a successful tool result");
        }
        if (responseID < 0 || (responseID == 0 && PartIDs.Length != 0)
            || (responseID != 0 && (kind != EventKind.ResponseEnded
                || PartIDs.Length != Parts.Length || PartIDs.Any(id => id < 0)
                || PartIDs.Distinct().Count() != PartIDs.Length)))
        {
            throw new ArgumentException("Invalid response display identity");
        }
        // Applicability of recognized payloads belongs to History.Reduce. Keep
        // wire identities and part decoding strict, but allow a known event with
        // a missing body or stale range to be diagnosed and skipped during replay.
        if (level is not null && !Enum.IsDefined(level.Value))
        {
            throw new ArgumentException("Unknown redaction level");
        }

        if (kind is EventKind.ToolCalled or EventKind.ToolStarted or EventKind.ToolReturned && string.IsNullOrEmpty(callID))
        {
            throw new ArgumentException("Tool event needs call ID");
        }

        if (kind is EventKind.ToolCalled or EventKind.ToolStarted && (string.IsNullOrEmpty(name) || args is null))
        {
            throw new ArgumentException("Tool call needs name and args");
        }
    }

    public partial Event Numbered(long seq, DateTimeOffset time) => new(Kind, seq, time, Actor, Parts, From, Usage, CallID, Name, Args, IsError, Error, RedactFrom, RedactTo, Level, Reason, Job, ResponseID, PartIDs, ToolReferences);

    public partial Event WithDisplay(long responseID, IEnumerable<int> partIDs) =>
        new(Kind, Seq, Time, Actor, Parts, From, Usage, CallID, Name, Args,
            IsError, Error, RedactFrom, RedactTo, Level, Reason, Job, responseID, partIDs, ToolReferences);
}
public sealed partial class Entry
{
    public partial Entry(long seq, Actor actor, IEnumerable<Part> parts, long responseID, IEnumerable<int>? partIDs)
    {
        Seq = seq;
        Actor = actor;
        Parts = parts.ToImmutableArray();
        ResponseID = responseID;
        PartIDs = partIDs?.ToImmutableArray() ?? [];
        foreach (var part in Parts)
        {
            if (part is ToolCallPart or ToolResultPart)
            {
                ToolBytes += System.Text.Encoding.UTF8.GetByteCount(JsonCodec.WritePart(part).ToJsonString());
            }
            if (part is ToolResultPart result && result.Parts.Any(item => item is not RedactedPart))
            {
                HasWholeToolResults = true;
            }
        }
    }
}
public sealed partial class Context
{
    public partial Context(IEnumerable<Entry>? dialogue, IEnumerable<Part>? ephemeral, Phase phase, Usage? usage)
    {
        if (!Enum.IsDefined(phase))
        {
            throw new ArgumentException("Unknown phase");
        }

        Dialogue = dialogue?.ToImmutableArray() ?? [];
        Ephemeral = ephemeral?.ToImmutableArray() ?? [];
        Phase = phase;
        Usage = usage ?? new Usage();
    }
}

public sealed partial class HandoffPart
{
    public partial HandoffPart(string text)
    {
        ArgumentException.ThrowIfNullOrWhiteSpace(text);
        Text = text;
    }
}
