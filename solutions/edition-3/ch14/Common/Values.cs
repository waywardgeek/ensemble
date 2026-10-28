using System.Collections.Immutable;
using System.Text.Json;

namespace Ensemble;

// These are closed vocabularies. Zero is never a valid recorded enum value.
public enum Vendor
{
    Anthropic = 1, OpenAI, Gemini
}
public enum Surface
{
    Messages = 1, ChatCompletions, GenerateContent
}
public enum Actor
{
    Human = 1, Agent, System, Tool
}
public enum Phase
{
    Idle = 1, InputPending, InFlight, ToolsPending
}
public enum EventKind
{
    MessageReceived = 1, RequestSent, ResponseStarted, ResponseEnded, ToolCalled, ToolReturned, Redacted, ErrorOccurred, JobKilled, HintReceived, TurnInterrupted, TurnEnded, ToolObserved, ToolStarted, AgentPaused, AgentResumed, AgentShutdown, SkillLoaded, SkillUnloadPending, RuntimeConfigured, HandoffRecorded
}
public enum Redaction
{
    Result = 1, Tool, Dialogue, Summary
}
public enum RefKind
{
    Path = 1, Uri, Handle
}

/// <summary>Value-like configuration. The Agent holds the authoritative instance;
/// transports reach it through their parent, never copy it. Credentials are not logged.</summary>
public sealed partial class Configuration
{
    // Explicit provider capability assertion; default full-declaration mode is portable.
    public readonly bool DeferredTools;
    // Explicit selected-model policy, not guessed from a provider/name prefix.
    // Off by default: ordinary configurations retain whole results until the
    // byte ladder or an explicit clearing operation selects them.
    public readonly bool StubsToolResults;
    public readonly Vendor Vendor;
    // Model is the routing name. ResolvedModel identifies opaque replay material.
    public readonly string Model;
    public readonly string ResolvedModel;
    public readonly string BaseUrl;
    public readonly string ApiKey;
    // Caller-supplied role instructions live once in immutable configuration.
    public readonly string PrimaryInstructions;
    // Selected-model policy, not an inference from a provider name.
    public readonly bool RequireToolExplanation;
    // Delivery override; canonical response content must not depend on it.
    public readonly bool DisableStreaming;
    // Zero selects 200; a positive value caps tool batches; negative is unlimited.
    public readonly int ToolRoundLimit;
    public partial Configuration(Vendor vendor, string model, string baseUrl, string apiKey = "", string resolvedModel = "", bool requireToolExplanation = false, int toolRoundLimit = 200, string primaryInstructions = "You are a helpful assistant.", bool disableStreaming = false, bool deferredTools = false, bool stubsToolResults = false);
    // Derive the replay identity from the configured vendor, surface, and exact model.
    public partial Provenance Target();
}

/// <summary>Value-like origin, captured when content arrives. Routing aliases are not identities.</summary>
public sealed partial class Provenance
{
    public readonly Vendor Vendor;
    public readonly string Model;
    public readonly Surface Surface;
    public partial Provenance(Vendor vendor, string model, Surface surface);
    // All three fields must agree; same vendor alone never authorizes opaque replay.
    public partial bool Matches(Provenance target);
}

/// <summary>Value-like location only. Request projection reads local paths without changing the locator.</summary>
public sealed partial class BlobRef
{
    public readonly RefKind Kind;
    public readonly string Locator;
    public partial BlobRef(RefKind kind, string locator);
}

/// <summary>Value-like measured counts, not prices. Null is unknown, not zero.
/// Addition propagates unknown independently for each disjoint category.</summary>
public sealed partial class Usage
{
    public readonly long? Input;
    public readonly long? CacheWrite;
    public readonly long? CacheRead;
    public readonly long? Output;
    public partial Usage(long? input = 0, long? cacheWrite = 0, long? cacheRead = 0, long? output = 0);
    // Unknown or overflowing sums remain unknown in that category only.
    public partial Usage Add(Usage other);
}

// Parts are immutable values, not owned runtime children. Inheritance models the
// actual alternatives of this closed union, not containment or service wiring.
public abstract partial class Part
{
    protected partial Part();
}
// Value-like continuation note, explicitly marked so replacement never relies
// on matching prose. It has no permission or ownership of its own.
public sealed partial class HandoffPart : Part
{
    public readonly string Text;
    public partial HandoffPart(string text);
}

// Ordinary readable content, including a genuinely empty answer.
public sealed partial class TextPart : Part
{
    public readonly string Text;
    // Optional provider material stays attached to this exact text part.
    public readonly Provenance? From;
    public readonly JsonElement? Opaque;
    public partial TextPart(string text, Provenance? from = null, JsonElement? opaque = null);
}
// A MIME type and locator, not loaded media bytes. Rendering checks model capabilities.
public sealed partial class BlobPart : Part
{
    public readonly string Mime;
    public readonly BlobRef Ref;
    public partial BlobPart(string mime, BlobRef reference);
}
// Provider-owned replay data. Clone the JSON and preserve its origin; never
// reinterpret this as readable reasoning for another model.
public sealed partial class OpaquePart : Part
{
    public readonly Provenance From;
    public readonly JsonElement Data;
    public partial OpaquePart(Provenance from, JsonElement data);
}
// A requested call, not proof of dispatch. Opaque material belongs to this call
// occurrence; Args retains exact JSON numbers rather than floating-point copies.
public sealed partial class ToolCallPart : Part
{
    public readonly string CallID;
    public readonly string Name;
    public readonly Provenance From;
    public readonly JsonElement Args;
    public readonly JsonElement? Opaque;
    public partial ToolCallPart(string callID, Provenance from, string name, JsonElement args, JsonElement? opaque = null);
}
// A recorded result is content even when IsError is true. Infrastructure failures
// instead use ErrorOccurred. Nested parts keep the original payload structure.
public sealed partial class ToolResultPart : Part
{
    public readonly string CallID;
    public readonly ImmutableArray<Part> Parts;
    public readonly bool IsError;
    public readonly JobInfo? Job;
    // Discovery evidence; only the deferred adapter emits provider references.
    public readonly ImmutableArray<string> ToolReferences;
    public partial ToolResultPart(string callID, IEnumerable<Part> parts, bool isError = false, JobInfo? job = null, IEnumerable<string>? toolReferences = null);
}
// A safe replacement in derived context. Ref preserves a blob locator internally;
// text projection must not disclose that locator to the model.
public sealed partial class RedactedPart : Part
{
    public readonly string Kind;
    public readonly string Reason;
    public readonly long Bytes;
    public readonly BlobRef? Ref;
    public partial RedactedPart(string kind, long bytes, string reason, BlobRef? reference = null);
}

/// <summary>Value-like canonical fact. Payload fields are discriminated by Kind;
/// the codec and History validate them before admission. Seq is assigned by History.</summary>
public sealed partial class Event
{
    public readonly long Seq;
    public readonly EventKind Kind;
    public readonly DateTimeOffset? Time;
    public readonly Actor? Actor;
    public readonly ImmutableArray<Part> Parts;
    public readonly Provenance? From;
    public readonly Usage? Usage;
    public readonly string CallID;
    public readonly string Name;
    public readonly string Error;
    public readonly string Reason;
    public readonly JsonElement? Args;
    public readonly bool IsError;
    public readonly long RedactFrom;
    public readonly long RedactTo;
    public readonly Redaction? Level;
    public readonly JobInfo? Job;
    public readonly ImmutableArray<string> ToolReferences;
    // Display identity is independent of provider call IDs and opaque payloads.
    // ResponseID is the durable RequestSent sequence; IDs are parser-local,
    // including zero. Empty metadata means a historical/non-response event.
    public readonly long ResponseID;
    public readonly ImmutableArray<int> PartIDs;
    public partial Event(EventKind kind, long seq = 0, DateTimeOffset? time = null,
        Actor? actor = null, IEnumerable<Part>? parts = null, Provenance? from = null,
        Usage? usage = null, string callID = "", string name = "", JsonElement? args = null,
        bool isError = false, string error = "", long redactFrom = 0, long redactTo = 0,
        Redaction? level = null, string reason = "", JobInfo? job = null,
        long responseID = 0, IEnumerable<int>? partIDs = null, IEnumerable<string>? toolReferences = null);
    // Attach finalized parser identities without altering canonical parts.
    public partial Event WithDisplay(long responseID, IEnumerable<int> partIDs);
    // Copy with History-assigned ordering; the original value is unchanged.
    public partial Event Numbered(long seq, DateTimeOffset time);
}

/// <summary>Value-like projection of one contribution, not another authoritative log.</summary>
public sealed partial class Entry
{
    // Derived once from immutable Parts, never serialized as another authority.
    // Count the actual retained wire representation, not a stub's original Bytes.
    public readonly long ToolBytes;
    public readonly bool HasWholeToolResults;
    public readonly long Seq;
    public readonly Actor Actor;
    public readonly ImmutableArray<Part> Parts;
    public readonly long ResponseID;
    public readonly ImmutableArray<int> PartIDs;
    public partial Entry(long seq, Actor actor, IEnumerable<Part> parts, long responseID = 0, IEnumerable<int>? partIDs = null);
}

/// <summary>Immutable replay result. History replaces this value after each event.
/// Pending ephemera is a replaceable one-request slot, never retained dialogue.</summary>
public sealed partial class Context
{
    public readonly ImmutableArray<Entry> Dialogue;
    public readonly ImmutableArray<Part> Ephemeral;
    public readonly Phase Phase;
    public readonly Usage Usage;
    // Derived permission, not another registry. Pending manuals retain their bodies
    // but grant nothing; a second nonpending manual may still grant the same name.
    public partial ImmutableHashSet<string> EnabledTools();
    public partial ImmutableHashSet<string> RevokedTools();
    public partial Context(IEnumerable<Entry>? dialogue = null, IEnumerable<Part>? ephemeral = null,
        Phase phase = Phase.Idle, Usage? usage = null);
}

// Value-like atomic read, not another authoritative transcript.
public sealed record HistorySnapshot(ImmutableArray<Event> Events, Context Context, long Sequence = 0);
