using Ensemble;
using System.Collections.Immutable;
using System.Net;
using System.Net.WebSockets;
using System.Text;
using System.Text.Json;
using System.Text.Json.Nodes;
using System.Threading.Channels;

namespace EnsembleGui;

public static partial class Gui
{
    public static partial async Task RunAsync(Agent agent, GuiOptions options,
        CancellationToken cancellation)
    {
        cancellation.ThrowIfCancellationRequested();
        var builder = WebApplication.CreateBuilder(Array.Empty<string>());
        builder.Logging.ClearProviders();
        builder.WebHost.ConfigureKestrel(server => server.Listen(IPAddress.Loopback, options.Port));
        await using var host = builder.Build();
        using var stop = CancellationTokenSource.CreateLinkedTokenSource(cancellation);
        using var registration = cancellation.Register(host.Lifetime.StopApplication);
        Task serving = ServeAsync(host, agent, options, stop.Token);
        bool started = false;
        var stopped = new TaskCompletionSource(TaskCreationOptions.RunContinuationsAsynchronously);
        using var stoppedRegistration = host.Lifetime.ApplicationStopping.Register(() => stopped.TrySetResult());
        try
        {
            // ServeAsync registers routes synchronously before its first await.
            // Surface validation failures before attempting to listen.
            if (serving.IsCompleted)
            {
                await serving;
            }
            await host.StartAsync(cancellation);
            started = true;
            await Task.WhenAny(serving, stopped.Task);
            if (serving.IsCompleted)
            {
                await serving;
            }
        }
        catch (OperationCanceledException) when (cancellation.IsCancellationRequested)
        {
        }
        finally
        {
            stop.Cancel();
            try
            {
                await serving;
            }
            finally
            {
                try
                {
                    // Preserve standalone hosting's account cleanup. The
                    // borrowed ServeAsync operation never cancels shared login.
                    if (started && agent.Parent().Credentials() is { } credentials)
                    {
                        await credentials.CancelLoginAsync();
                    }
                }
                finally
                {
                    await host.StopAsync(CancellationToken.None);
                }
            }
        }
    }

    public static partial async Task ServeAsync(WebApplication host, Agent agent, GuiOptions options,
        CancellationToken cancellation)
    {
        cancellation.ThrowIfCancellationRequested();
        int port = options.Port;
        if (port is < 1 or > 65535)
        {
            throw new ArgumentException("Port must be between 1 and 65535");
        }
        string basePath = options.BasePath;
        // A literal route prefix, not a route template or escaped URL. Keeping
        // it unambiguous also makes callback construction independent of routing.
        if (basePath.Length != 0 && (basePath[0] != '/' || basePath.EndsWith('/')
            || basePath.Split('/').Skip(1).Any(segment => segment.Length == 0
                || segment.Any(c => !char.IsAsciiLetterOrDigit(c) && c is not '-' and not '_'))))
        {
            throw new ArgumentException("GUI base path must be empty or slash-separated literal names.");
        }
        if (host.Lifetime.ApplicationStarted.IsCancellationRequested)
        {
            throw new InvalidOperationException("Mount the GUI before starting the application.");
        }
        string settingsPath = agent.Sandbox().Resolve(options.SettingsPath);
        var loaded = LoadSettings(agent, settingsPath);
        var preferences = loaded.Preferences;
        agent.SetMaxToolRounds(loaded.Budget);
        agent.SetContextByteBudget(loaded.ContextTarget);
        host.UseWebSockets();
        var routes = host.MapGroup(basePath);

        // Operation-scoped I/O state. The Agent remains owned by its Ensemble.
        // Queues and socket lifetimes end when ServeAsync ends; no socket is ever
        // stored on Agent or used to decide whether a model turn is complete.
        string agentID = agent.Name ?? "root";
        object gate = new();
        // Notification deduplication only; Engine remains the accounting authority.
        string? lastMeterWire = null;
        string tracePath = options.TracePath;
        using var trace = new StreamWriter(agent.Sandbox().OpenFile(
            tracePath, FileMode.Append, FileAccess.Write)) { AutoFlush = true };
        void Trace(string direction, string json)
        {
            // Keep the protocol trace separate, but use the root's one
            // diagnostic policy before bytes reach this caller-owned writer.
            agent.Parent().Log($"{direction} {json}", trace);
        }
        var peers = new Dictionary<WebSocket, (Channel<string> Queue, long Cursor)>();
        var partials = new Dictionary<string, JsonObject>(StringComparer.Ordinal);
        using var stopping = CancellationTokenSource.CreateLinkedTokenSource(
            cancellation, host.Lifetime.ApplicationStopping);
        int activeRequests = 0;
        var drained = new TaskCompletionSource(TaskCreationOptions.RunContinuationsAsynchronously);
        // Wrap only this route group. Admission closes before disposal; requests
        // already using the trace, sockets or credential endpoint are joined.
        // Unrelated caller endpoints retain their own authority and lifetime.
        void GuardEndpoint(EndpointBuilder endpoint)
        {
            RequestDelegate handler = endpoint.RequestDelegate
                ?? throw new InvalidOperationException("GUI endpoint has no handler.");
            endpoint.RequestDelegate = async context =>
            {
                lock (gate)
                {
                    if (stopping.IsCancellationRequested)
                    {
                        context.Response.StatusCode = StatusCodes.Status503ServiceUnavailable;
                        return;
                    }
                    activeRequests++;
                }
                try
                {
                    if (context.Request.Host.Host is not ("127.0.0.1" or "localhost" or "[::1]"))
                    {
                        context.Response.StatusCode = StatusCodes.Status400BadRequest;
                        return;
                    }
                    await handler(context);
                }
                finally
                {
                    lock (gate)
                    {
                        activeRequests--;
                        if (stopping.IsCancellationRequested && activeRequests == 0)
                        {
                            drained.TrySetResult();
                        }
                    }
                }
            };
        }
        ((IEndpointConventionBuilder)routes).Add(GuardEndpoint);
        var callbackRoutes = basePath.Length == 0 ? null : host.MapGroup("");
        if (callbackRoutes is not null)
        {
            ((IEndpointConventionBuilder)callbackRoutes).Add(GuardEndpoint);
        }
        MapAuth(routes, callbackRoutes, agent, port, stopping.Token);
        MapAssets(routes, options.AssetsDirectory);
        var manager = agent.Mcp() ?? new Mcp(agent);
        var browserConnected = new TaskCompletionSource(TaskCreationOptions.RunContinuationsAsynchronously);

        JsonObject State(Projection projection) => new()
        {
            ["type"] = "agent_state", ["agent_id"] = agentID,
            ["name"] = agentID, ["state"] = projection.State
        };

        // Must be called under gate. TryWrite never waits on a browser.
        void Queue(WebSocket socket, Channel<string> queue, JsonObject frame)
        {
            if (!queue.Writer.TryWrite(frame.ToJsonString()))
            {
                queue.Writer.TryComplete();
                socket.Abort();
            }
        }

        (HistorySnapshot Snapshot, Projection Projection) Capture(long? through = null)
        {
            var snapshot = agent.History().Snapshot();
            if (through is long sequence)
            {
                snapshot = snapshot with
                {
                    Events = snapshot.Events.Where(e => e.Seq <= sequence).ToImmutableArray(),
                    Context = new Context(snapshot.Context.Dialogue.Where(entry => entry.Seq <= sequence)),
                    Sequence = Math.Min(snapshot.Sequence, sequence)
                };
            }
            var projection = Project(snapshot, agentID);
            // A completed response retires its previews even when a mid-turn
            // cut removes the corresponding final tool cards from Context.
            long ended = snapshot.Events.LastOrDefault(e =>
                    e.Kind is EventKind.ResponseEnded or EventKind.TurnEnded)?.Seq ?? 0;
            foreach (var key in partials.Keys.ToArray())
            {
                var response = partials[key]["response_id"]!.GetValue<long>();
                if (projection.FinalIDs.Contains(key) || response <= ended)
                {
                    partials.Remove(key);
                }
            }
            return (snapshot, projection);
        }

        JsonObject Replay(Projection projection) => new()
        {
            ["type"] = "replay", ["data"] = new JsonObject
            {
                ["events"] = new JsonArray(projection.Frames.Select(frame => (JsonNode)frame.DeepClone()).ToArray()),
                ["partials"] = new JsonArray(partials.Values.Select(frame => (JsonNode)frame.DeepClone()).ToArray()),
                ["agents"] = new JsonArray(State(projection))
            }
        };

        void Broadcast(Observation? observation)
        {
            lock (gate)
            {
                var meter = MeterFrame(agent);
                string meterWire = meter.ToJsonString();
                if (meterWire != lastMeterWire)
                {
                    foreach (var (socket, peer) in peers.ToArray())
                    {
                        Queue(socket, peer.Queue, meter);
                    }
                    lastMeterWire = meterWire;
                }
                long? through = observation?.Fact?.Seq;
                if (observation?.Kind == ObservationKind.PartDelta)
                {
                    through = observation.ResponseID;
                }
                var (snapshot, projection) = Capture(through);
                long latest = snapshot.Sequence;
                foreach (var (socket, peer) in peers.ToArray())
                {
                    bool reduced = snapshot.Events.Any(fact => fact.Seq > peer.Cursor
                        && fact.Kind is EventKind.Redacted or EventKind.HandoffRecorded);
                    bool missedTail = latest > peer.Cursor && (snapshot.Events.IsEmpty
                        || snapshot.Events[0].Seq > peer.Cursor + 1);
                    if (reduced || missedTail)
                    {
                        // A cut changes older card contents/identities, not just
                        // the newest sequence. Reuse silent replay rather than
                        // leave stale cards or invent a second transcript store.
                        Queue(socket, peer.Queue, Replay(projection));
                    }
                    else
                    {
                        foreach (var frame in projection.Frames)
                        {
                            if (frame["seq"]!.GetValue<long>() > peer.Cursor)
                            {
                                Queue(socket, peer.Queue, frame);
                            }
                        }
                    }
                    if (latest > peer.Cursor)
                    {
                        Queue(socket, peer.Queue, State(projection));
                        peers[socket] = (peer.Queue, latest);
                    }
                }
                if (observation?.Kind != ObservationKind.PartDelta || observation.PartID is not int partID)
                {
                    return;
                }
                string id = PartID(agentID, observation.ResponseID, partID);
                long ended = snapshot.Events.LastOrDefault(e =>
                    e.Kind is EventKind.ResponseEnded or EventKind.TurnEnded)?.Seq ?? 0;
                if (projection.FinalIDs.Contains(id) || observation.ResponseID <= ended)
                {
                    return; // Late chunks cannot resurrect finalized/interrupted work.
                }
                string kind = observation.Delta!.Kind switch
                {
                    ChunkKind.Reply => "text",
                    ChunkKind.Thinking => "thinking",
                    ChunkKind.ReasoningSummary => "reasoning_summary",
                    ChunkKind.ToolCall => "tool_call",
                    _ => throw new InvalidOperationException("Unknown display chunk kind")
                };
                var delta = new JsonObject
                {
                    ["type"] = "part_delta", ["agent_id"] = agentID, ["part_id"] = id,
                    ["response_id"] = observation.ResponseID, ["kind"] = kind,
                    ["chunk"] = (observation.Delta!.Kind == ChunkKind.ToolCall ? observation.Delta.ArgsText : observation.Delta.Text), ["tool_name"] = observation.Delta!.ToolName
                };
                if (partials.TryGetValue(id, out var previous))
                {
                    previous["chunk"] = previous["chunk"]!.GetValue<string>() + (observation.Delta!.Kind == ChunkKind.ToolCall ? observation.Delta.ArgsText : observation.Delta.Text);
                    previous["tool_name"] = previous["tool_name"]!.GetValue<string>() + observation.Delta!.ToolName;
                }
                else
                {
                    partials[id] = (JsonObject)delta.DeepClone();
                }
                long finalizedAt = agent.History().Snapshot().Events
                    .FirstOrDefault(e => e.Kind == EventKind.ResponseEnded && e.ResponseID == observation.ResponseID)?.Seq
                    ?? long.MaxValue;
                foreach (var (socket, peer) in peers)
                {
                    if (peer.Cursor < finalizedAt)
                    {
                        Queue(socket, peer.Queue, delta);
                    }
                }
            }
        }

        async Task Observe()
        {
            while (!stopping.IsCancellationRequested)
            {
                using var wait = CancellationTokenSource.CreateLinkedTokenSource(stopping.Token);
                wait.CancelAfter(TimeSpan.FromMilliseconds(200));
                Observation? observation = null;
                try
                {
                    observation = await agent.WaitAsync(_ => true, wait.Token);
                }
                catch (OperationCanceledException) when (!stopping.IsCancellationRequested)
                {
                    // Recover durable facts even if every progress wake-up was dropped.
                }
                catch (ChannelClosedException)
                {
                    return;
                }
                if (!stopping.IsCancellationRequested &&
                    (observation is null || observation.Kind is ObservationKind.EventRecorded or ObservationKind.PartDelta))
                {
                    Broadcast(observation);
                }
            }
        }

        routes.Map("/ws", async (HttpContext http) =>
        {
            // Loopback is not an origin check. A hostile web page must not gain
            // local tool authority merely by opening this WebSocket endpoint.
            string origin = http.Request.Headers.Origin.ToString();
            if (!http.WebSockets.IsWebSocketRequest ||
                (origin.Length != 0 && origin != $"http://{http.Request.Host}"))
            {
                http.Response.StatusCode = StatusCodes.Status403Forbidden;
                return;
            }
            using var socket = await http.WebSockets.AcceptWebSocketAsync();
            using var lifetime = CancellationTokenSource.CreateLinkedTokenSource(
                stopping.Token, http.RequestAborted, host.Lifetime.ApplicationStopping);
            var queue = Channel.CreateBounded<string>(new BoundedChannelOptions(256)
            {
                FullMode = BoundedChannelFullMode.Wait, SingleReader = true, SingleWriter = false
            });
            await using var tunnel = new McpTunnel(queue.Writer, agentID);
            McpBrowser? browser = null;
            lock (gate)
            {
                // First connected tab is the one active browser target. Other
                // tabs still observe/control this Agent, but cannot inject RPC.
                if (ReferenceEquals(agent.Mcp(), manager) && manager.Browser() is null)
                {
                    try
                    {
                        browser = new McpBrowser(manager, tunnel);
                        browserConnected.TrySetResult();
                    }
                    catch (InvalidOperationException)
                    {
                        // Agent shutdown may close admission concurrently.
                        // This socket can still observe the finalized replay.
                    }
                }
                var (snapshot, projection) = Capture();
                peers.Add(socket, (queue, snapshot.Sequence));
                Queue(socket, queue, Replay(projection));
                Queue(socket, queue, MeterFrame(agent));
                Queue(socket, queue, State(projection));
                Queue(socket, queue, new JsonObject
                {
                    ["type"] = "current_settings",
                    ["settings"] = Settings(preferences, agent.MaxToolRounds(), agent.ContextByteBudget())
                });
            }
            async Task Send()
            {
                await foreach (string frame in queue.Reader.ReadAllAsync(lifetime.Token))
                {
                    await socket.SendAsync(Encoding.UTF8.GetBytes(frame), WebSocketMessageType.Text, true, lifetime.Token);
                    Trace("<", frame);
                }
            }
            var sender = Send();
            try
            {
                var buffer = new byte[8192];
                while (!lifetime.IsCancellationRequested)
                {
                    using var content = new MemoryStream();
                    WebSocketReceiveResult received;
                    do
                    {
                        received = await socket.ReceiveAsync(buffer, lifetime.Token);
                        if (received.MessageType == WebSocketMessageType.Close)
                        {
                            return;
                        }
                        if (received.MessageType != WebSocketMessageType.Text || content.Length + received.Count > 1024 * 1024)
                        {
                            throw new InvalidDataException("Expected a text control under 1 MiB");
                        }
                        content.Write(buffer, 0, received.Count);
                    }
                    while (!received.EndOfMessage);
                    try
                    {
                        using var json = JsonDocument.Parse(content.ToArray());
                        Trace(">", json.RootElement.GetRawText());
                        var command = json.RootElement;
                        if (command.TryGetProperty("agent_id", out var target)
                            && (target.ValueKind != JsonValueKind.String || target.GetString() != agentID))
                        {
                            throw new InvalidDataException("Command targets a different Agent.");
                        }
                        string type = command.GetProperty("type").GetString() ?? "";
                        string text = command.TryGetProperty("text", out var t) ? t.GetString() ?? "" : "";
                        switch (type)
                        {
                            case "mcp":
                                if (browser is null || !ReferenceEquals(manager.Browser(), browser))
                                {
                                    throw new InvalidDataException("This tab is not the active MCP browser.");
                                }
                                tunnel.Accept(command.GetProperty("data"));
                                break;
                            case "subscribe":
                                // The initial snapshot is already queued atomically
                                // with registration. Subscription is idempotent.
                                break;
                            case "get_settings":
                                lock (gate)
                                {
                                    Queue(socket, queue, new JsonObject
                                    {
                                        ["type"] = "current_settings",
                                        ["settings"] = Settings(preferences, agent.MaxToolRounds(), agent.ContextByteBudget())
                                    });
                                }
                                break;
                            case "update_settings":
                                lock (gate)
                                {
                                    var candidate = MergeSettings(command.GetProperty("settings"), preferences, agent.MaxToolRounds(), agent.ContextByteBudget());
                                    var saved = Settings(candidate.Preferences, candidate.Budget, candidate.ContextTarget);
                                    try
                                    {
                                        // Serialize save/publication across clients. Only after
                                        // replacement succeeds may new turns see the new policy.
                                        SaveSettings(agent, settingsPath, saved);
                                        agent.SetMaxToolRounds(candidate.Budget);
                                        agent.SetContextByteBudget(candidate.ContextTarget);
                                        preferences = candidate.Preferences;
                                        var frame = new JsonObject { ["type"] = "settings_updated", ["settings"] = saved };
                                        foreach (var (peerSocket, peer) in peers.ToArray())
                                        {
                                            Queue(peerSocket, peer.Queue, frame);
                                        }
                                    }
                                    catch (Exception error) when (error is IOException or UnauthorizedAccessException)
                                    {
                                        Queue(socket, queue, new JsonObject
                                        {
                                            ["type"] = "error", ["message"] = "Settings could not be saved; no changes applied."
                                        });
                                    }
                                }
                                break;
                            case "prompt":
                                await agent.PostAsync(new AgentInput(InputKind.Message, text));
                                break;
                            case "hint":
                                await agent.PostAsync(new AgentInput(InputKind.Hint, text));
                                break;
                            case "pause":
                                await agent.PostAsync(new AgentInput(InputKind.Pause));
                                break;
                            case "unpause":
                                await agent.PostAsync(new AgentInput(InputKind.Resume));
                                break;
                            case "interrupt":
                                string mode = command.TryGetProperty("mode", out var requestedMode)
                                    ? requestedMode.GetString() ?? "now" : "now";
                                InputKind interrupt = mode switch
                                {
                                    "now" => InputKind.InterruptNow,
                                    "after_tool" => InputKind.InterruptAfterTool,
                                    _ => throw new ArgumentException("Unknown interrupt mode")
                                };
                                await agent.PostAsync(new AgentInput(interrupt));
                                break;
                            case "set_model":
                                try
                                {
                                    await agent.SetModelAsync(text);
                                    Broadcast(null);
                                }
                                catch (Exception error) when (error is ArgumentException or InvalidOperationException)
                                {
                                    lock (gate)
                                    {
                                        Queue(socket, queue, new JsonObject { ["type"] = "error",
                                            ["message"] = "Model selection refused: Agent must be idle and the model must be compatible with this route." });
                                    }
                                }
                                // Also acknowledge unchanged selection and refusals.
                                // This restores the requesting tab's proposed picker
                                // without inventing a second model-setting store.
                                lock (gate)
                                {
                                    Queue(socket, queue, MeterFrame(agent));
                                }
                                break;
                            case "clear_tools":
                                await agent.PostAsync(new AgentInput(InputKind.ClearTools));
                                break;
                            case "shutdown":
                                await agent.ShutdownAsync();
                                break;
                            default:
                                throw new ArgumentException("Unknown GUI control");
                        }
                    }
                    catch (Exception error) when (error is JsonException or ArgumentException or KeyNotFoundException or InvalidOperationException or InvalidDataException)
                    {
                        lock (gate)
                        {
                            Queue(socket, queue, new JsonObject { ["type"] = "error", ["message"] = "Invalid or unavailable control" });
                        }
                    }
                }
            }
            catch (Exception error) when (error is OperationCanceledException or WebSocketException or IOException)
            {
                // Disconnect ends this tab's browser tools, not the Agent turn.
                // An in-flight call reports its real connection failure.
            }
            finally
            {
                lock (gate)
                {
                    peers.Remove(socket);
                    queue.Writer.TryComplete();
                }
                lifetime.Cancel();
                socket.Abort();
                if (browser is not null)
                {
                    await browser.DestroyAsync();
                }
                try
                {
                    await sender;
                }
                catch (Exception error) when (error is OperationCanceledException or WebSocketException or IOException)
                {
                    // The single sender is joined before disposing its socket.
                }
            }
        });
        var observer = Observe();
        try
        {
            var started = new TaskCompletionSource(TaskCreationOptions.RunContinuationsAsynchronously);
            using var startedRegistration = host.Lifetime.ApplicationStarted.Register(() => started.TrySetResult());
            await started.Task.WaitAsync(stopping.Token);
            if (agent.SkillManager() is { } skills && !skills.Ready())
            {
                // Server-backed startup needs a live event loop: the browser
                // must be able to answer initialize/tools/list before user work.
                if (skills.RequiresBrowserStartup())
                {
                    await browserConnected.Task.WaitAsync(stopping.Token);
                }
                await skills.InitializeAsync(stopping.Token);
            }
            await Task.Delay(Timeout.InfiniteTimeSpan, stopping.Token);
        }
        catch (OperationCanceledException) when (stopping.IsCancellationRequested)
        {
            // Normal host termination may arrive while waiting for a browser.
        }
        finally
        {
            stopping.Cancel();
            lock (gate)
            {
                foreach (var (socket, peer) in peers)
                {
                    peer.Queue.Writer.TryComplete();
                    socket.Abort();
                }
                if (activeRequests == 0)
                {
                    drained.TrySetResult();
                }
            }
            await drained.Task;
            try
            {
                await observer;
            }
            catch (OperationCanceledException) when (stopping.IsCancellationRequested)
            {
            }
            // Borrowed hosting must not cancel the caller's shared credential
            // attempts. The caller owns that lifecycle as well as its host.
        }
    }
}
