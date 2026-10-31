using Ensemble;
using System.Collections.Immutable;
using System.Text.Json.Nodes;

namespace EnsembleGui;

// Value-like display projection. These disposable values are derived from the
// real History; they do not own Agents or become another transcript authority.
public sealed record Projection(ImmutableArray<JsonObject> Frames, string State,
    ImmutableHashSet<string> FinalIDs);

// This optional application uses the library's public API. RunAsync scopes all
// HTTP/WebSocket resources to one hosting operation; no GUI objects are added
// to the core ownership graph and no callback is used to schedule model work.
public static partial class Gui
{
    public static partial Task RunAsync(Agent agent, string[] args);
    // Replay and live durable delivery use this exact same projection.
    public static partial Projection Project(HistorySnapshot snapshot, string agentID);
    public static partial string PartID(string agentID, long responseID, int partID);
    // Human-readable wire data retains the full canonical part under part.
    public static partial JsonObject PartFrame(string agentID, string id, Part part, Actor role, long sequence);
}
