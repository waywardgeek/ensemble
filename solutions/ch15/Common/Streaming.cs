using System.Collections.Immutable;
using System.Text.Json.Nodes;

namespace Ensemble;

// Value-like display data. IDs are local to one response, and zero is valid.
// Chunks are transient approximations; a finalized Part remains authoritative.
public enum ChunkKind { Reply, Thinking, ToolCall }
public sealed record Chunk(int PartID, ChunkKind Kind, string Text = "",
    string ToolName = "", string ArgsText = "");
public sealed record ParsedResponse(ImmutableArray<Event> Events,
    ImmutableArray<int> PartIDs, bool ToolBeforeText = false, UsageIdentity? Billing = null);

public sealed partial class Engine
{
    // One HTTP parsing seam for JSON and SSE. The callback is a transient output
    // sink for this invocation, never a stored service or a History writer.
    public partial Task<ParsedResponse> ParseAsync(HttpResponseMessage response,
        Action<Chunk>? onChunk = null, CancellationToken cancellation = default);

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
