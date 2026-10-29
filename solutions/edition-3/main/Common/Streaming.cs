using System.Collections.Immutable;
using System.Text.Json.Nodes;

namespace Ensemble;

// Value-like display data. IDs are local to one response, and zero is valid.
// Chunks are transient approximations; a finalized Part remains authoritative.
public enum ChunkKind { Reply, Thinking, ToolCall, ReasoningSummary }
public sealed record Chunk(int PartID, ChunkKind Kind, string Text = "",
    string ToolName = "", string ArgsText = "");
public sealed record ParsedResponse(ImmutableArray<Event> Events,
    ImmutableArray<int> PartIDs, bool ToolBeforeText = false, UsageIdentity? Billing = null,
    bool IncompleteOutput = false);

public sealed partial class Engine
{
    // One HTTP parsing seam for JSON and SSE. The callback is a transient output
    // sink for this invocation, never a stored service or a History writer.
    public partial Task<ParsedResponse> ParseAsync(HttpResponseMessage response,
        Action<Chunk>? onChunk = null, CancellationToken cancellation = default);
    // Authenticated sends supply only invocation-local diagnostic secrets. They
    // are never persisted on Engine or used to rewrite canonical provider data.
    private partial Task<ParsedResponse> ParseCoreAsync(Configuration config, HttpResponseMessage response,
        IReadOnlyList<string> exchangeSecrets, Action<Chunk>? onChunk, CancellationToken cancellation,
        bool requireComplete);

    // Value-like Responses parse state, never shared between HTTP requests.
    private sealed class ResponsesAssembly
    {
        public readonly SortedDictionary<int, JsonObject> Completed = new();
        public readonly HashSet<int> SeenItems = [];
        public readonly Dictionary<(int Item, int Part), int> IDs = new();
        public readonly Dictionary<int, string> Displayed = new();
        public JsonObject? Terminal;
        public bool SeenText;
        public bool ToolBeforeText;
    }

    // Value-like, per-parse scratch state. Native blocks stay native until the
    // same completed-body decoder used by plain JSON produces canonical Parts.
    private sealed class StreamAssembly
    {
        public readonly JsonObject Root = new();
        public readonly SortedDictionary<int, JsonObject> Blocks = new();
        public readonly Dictionary<int, string> Arguments = new();
        public readonly List<int> Order = [];
        public readonly HashSet<int> OpenBlocks = [];
        public readonly Dictionary<int, int> ToolIDs = new();
        public int NextID;
        public bool Terminal;
        public bool FinishSeen;
        public bool SeenText;
        public bool ToolBeforeText;
        public int? GeminiCurrent;
    }
}
