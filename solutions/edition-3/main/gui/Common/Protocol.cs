using Ensemble;
using System.Collections.Immutable;
using System.Text.Json.Nodes;

namespace EnsembleGui;

// Value-like display projection. These disposable values are derived from the
// real History; they do not own Agents or become another transcript authority.
public sealed record Projection(ImmutableArray<JsonObject> Frames, string State,
    ImmutableHashSet<string> FinalIDs);

// This optional library uses the library's public API. RunAsync scopes all
// HTTP/WebSocket resources to one hosting operation; no GUI objects are added
// to the core ownership graph and no callback is used to schedule model work.
public static partial class Gui
{
    // The operation borrows the Agent. Cancellation joins HTTP, sockets and the
    // observation loop; the caller remains responsible for Agent destruction.
    public static partial Task RunAsync(Agent agent, GuiOptions options,
        CancellationToken cancellation = default);
    // Register the complete GUI before the caller starts its WebApplication.
    // The returned task is its lifetime, not a registration acknowledgment: it
    // joins admitted requests/observation on cancellation or application stop.
    // The caller owns the application and Agent; neither is stopped/destroyed.
    // Mount once per Agent/application, before startup; await this task before
    // destroying the Agent. Caller must bind loopback at options.Port. BasePath
    // namespaces controls/assets; /auth/callback remains reserved at the root.
    // Shared credential attempts remain caller-owned after this task finishes.
    public static partial Task ServeAsync(WebApplication host, Agent agent, GuiOptions options,
        CancellationToken cancellation = default);
    // Map only the shipped page/assets on a caller-owned endpoint collection.
    // A disk override is explicit host authority, never a model-provided path.
    public static partial void MapAssets(Microsoft.AspNetCore.Routing.IEndpointRouteBuilder routes,
        string? directory = null);
    // Host-scoped HTTP routes, not another credential owner. Private OAuth data
    // never enters the conversation, WebSocket projection, or GUI trace.
    private static partial void MapAuth(Microsoft.AspNetCore.Routing.IEndpointRouteBuilder routes,
        Microsoft.AspNetCore.Routing.IEndpointRouteBuilder? callbackRoutes,
        Agent agent, int port, CancellationToken stopping);
    // Trusted mounted peer controls use HTTP so browser-dependent construction
    // cannot block the WebSocket carrying that peer's initialization replies.
    private static partial void MapPeers(Microsoft.AspNetCore.Routing.IEndpointRouteBuilder routes,
        Agent source, int port, CancellationToken stopping);
    // Read-only complete Job output, resolved through a live Agent and its Jobs.
    // No caller path is accepted. Completed output is guarded and bounded; the
    // response copy observes both mount shutdown and request cancellation.
    private static partial void MapJobOutput(Microsoft.AspNetCore.Routing.IEndpointRouteBuilder routes,
        Agent source, int port, CancellationToken stopping);
    // Current-process accounting comes from Engine, never reconstructed from replay.
    public static partial JsonObject MeterFrame(Agent agent);
    // Replay and live durable delivery use this exact same projection.
    public static partial Projection Project(HistorySnapshot snapshot, string agentID);
    // Merge authoritative reduced entries with retained lifecycle diagnostics.
    // Synthesized display facts are local projections, never appended to History.
    private static partial ImmutableArray<Event> DisplayFacts(HistorySnapshot snapshot);
    public static partial string PartID(string agentID, long responseID, int partID);
    // Human-readable wire data retains the full canonical part under part.
    public static partial JsonObject PartFrame(string agentID, string id, Part part, Actor role, long sequence);
}
