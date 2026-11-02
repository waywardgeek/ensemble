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
        string? publicOrigin = ValidatePublicOrigin(options.PublicOrigin);
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
        static string AddressOf(Agent target) => target.Address() ?? "";
        object gate = new();
        // Serialize memory read/patch/save/apply across tabs without holding
        // the projection lock while the actor joins an invalidated compressor.
        using var agentSettingsGate = new SemaphoreSlim(1, 1);
        // Notification deduplication only; Engine remains the accounting authority.
        var lastMeterWire = new Dictionary<Agent, string>();
        // An explicit path is an embedded-host trace override. Ordinary hosting
        // uses the actual Agent's writer, never a shared root conversation log.
        using var trace = options.TracePath is null ? null : new StreamWriter(agent.Sandbox().OpenFile(
            options.TracePath, FileMode.Create, FileAccess.Write)) { AutoFlush = true };
        void Trace(Agent? owner, string direction, string json)
        {
            string projection = TraceProjection(json);
            if (trace is not null) agent.Parent().Log($"{direction} {projection}", trace);
            else if (owner is not null) owner.TraceGui(direction, projection);
            else agent.Parent().Log($"GUI {direction} {projection}");
        }
        var peers = new Dictionary<WebSocket, VoicePeer>();
        // One reader and preview buffer per viewed conversation. A tab switch
        // never makes two readers compete for the Agent's lossy observation queue.
        var previews = new Dictionary<Agent, Dictionary<string, JsonObject>>();
        var observers = new Dictionary<Agent, Task>();
        JsonObject? latestTree = null;
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
                    if (RequestOrigin(context, port, publicOrigin) is null)
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
        MapAuth(routes, callbackRoutes, agent, port, publicOrigin, stopping.Token);
        MapPeers(routes, agent, port, publicOrigin, stopping.Token);
        MapJobOutput(routes, agent, port, publicOrigin, stopping.Token);
        MapAssets(routes, options.AssetsDirectory);
        var initialMcp = agent.EnsureMcp();

        JsonObject State(Agent selected, Projection projection) => new()
        {
            ["type"] = "agent_state", ["agent_id"] = AddressOf(selected),
            ["name"] = selected.ChildID() ?? selected.Name ?? "(unnamed)", ["state"] = projection.State
        };

        // Must be called under gate. TryWrite never waits on a browser.
        bool Queue(WebSocket socket, Channel<string> queue, JsonObject frame)
        {
            string json = frame.ToJsonString();
            if (!queue.Writer.TryWrite(json))
            {
                queue.Writer.TryComplete();
                socket.Abort();
                return false;
            }
            if (trace is null)
            {
                // These are server-created addresses, captured while enqueueing.
                // Later display selection cannot reattribute an older frame.
                string? address = frame["agent_id"]?.GetValue<string>();
                Agent? owner = address is null ? null : agent.Parent().FindAgentAddress(address);
                Trace(owner, "<queued", json);
            }
            return true;
        }

        // Called under gate. Readers are operation-scoped and joined on exit;
        // retaining a projection never grants ownership of its Agent.
        void ObserveOnce(Agent selected)
        {
            if (observers.ContainsKey(selected)) return;
            previews.Add(selected, new Dictionary<string, JsonObject>(StringComparer.Ordinal));
            observers.Add(selected, Task.Run(() => Observe(selected)));
        }

        (HistorySnapshot Snapshot, Projection Projection) Capture(Agent selected, long? through = null)
        {
            var partials = previews[selected];
            var snapshot = selected.History().Snapshot();
            if (through is long sequence)
            {
                snapshot = snapshot with
                {
                    Events = snapshot.Events.Where(e => e.Seq <= sequence).ToImmutableArray(),
                    Context = new Context(snapshot.Context.Dialogue.Where(entry => entry.Seq <= sequence)),
                    Sequence = Math.Min(snapshot.Sequence, sequence)
                };
            }
            var projection = Project(snapshot, AddressOf(selected));
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

        JsonObject Replay(Agent selected, Projection projection) => new()
        {
            ["type"] = "replay", ["agent_id"] = AddressOf(selected), ["data"] = new JsonObject
            {
                ["events"] = new JsonArray(projection.Frames.Select(frame => (JsonNode)frame.DeepClone()).ToArray()),
                ["partials"] = new JsonArray(previews[selected].Values.Select(frame => (JsonNode)frame.DeepClone()).ToArray()),
                ["agents"] = new JsonArray(State(selected, projection))
            }
        };

        void Broadcast(Agent selected, Observation? observation)
        {
            lock (gate)
            {
                var partials = previews[selected];
                string agentID = AddressOf(selected);
                var meter = MeterFrame(selected);
                string meterWire = meter.ToJsonString();
                if (!lastMeterWire.TryGetValue(selected, out string? previousMeter) || meterWire != previousMeter)
                {
                    foreach (var (socket, peer) in peers.ToArray())
                    {
                        if (peer.Selected == selected) Queue(socket, peer.Queue, meter);
                    }
                    lastMeterWire[selected] = meterWire;
                }
                long? through = observation?.Fact?.Seq;
                if (observation?.Kind == ObservationKind.PartDelta)
                {
                    through = observation.ResponseID;
                }
                var (snapshot, projection) = Capture(selected, through);
                long latest = snapshot.Sequence;
                foreach (var (socket, peer) in peers.ToArray())
                {
                    if (peer.Selected != selected) continue;
                    bool reduced = snapshot.Events.Any(fact => fact.Seq > peer.Cursor
                        && fact.Kind is EventKind.Redacted or EventKind.HandoffRecorded);
                    bool missedTail = latest > peer.Cursor && (snapshot.Events.IsEmpty
                        || snapshot.Events[0].Seq > peer.Cursor + 1);
                    if (reduced || missedTail)
                    {
                        // A cut changes older card contents/identities, not just
                        // the newest sequence. Reuse silent replay rather than
                        // leave stale cards or invent a second transcript store.
                        Queue(socket, peer.Queue, Replay(selected, projection));
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
                        Queue(socket, peer.Queue, State(selected, projection));
                        peer.Cursor = latest;
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
                long finalizedAt = selected.History().Snapshot().Events
                    .FirstOrDefault(e => e.Kind == EventKind.ResponseEnded && e.ResponseID == observation.ResponseID)?.Seq
                    ?? long.MaxValue;
                foreach (var (socket, peer) in peers)
                {
                    if (peer.Selected == selected && peer.Cursor < finalizedAt)
                    {
                        Queue(socket, peer.Queue, delta);
                    }
                }
            }
        }

        async Task Observe(Agent selected)
        {
            while (!stopping.IsCancellationRequested)
            {
                using var wait = CancellationTokenSource.CreateLinkedTokenSource(stopping.Token);
                wait.CancelAfter(TimeSpan.FromMilliseconds(200));
                Observation? observation = null;
                try
                {
                    observation = await selected.WaitAsync(_ => true, wait.Token);
                }
                catch (OperationCanceledException) when (!stopping.IsCancellationRequested)
                {
                    // Recover durable facts even if every progress wake-up was dropped.
                }
                catch (ChannelClosedException) { return; }
                catch (ObjectDisposedException) { return; }
                if (!stopping.IsCancellationRequested &&
                    (observation is null || observation.Kind is ObservationKind.EventRecorded or ObservationKind.PartDelta))
                {
                    try { Broadcast(selected, observation); }
                    catch (ObjectDisposedException) { return; }
                }
            }
        }

        bool Select(WebSocket socket, Channel<string> queue, Agent target)
        {
            lock (gate)
            {
                ObserveOnce(target);
                var (snapshot, projection) = Capture(target);
                // A single bounded queue item contains the entire replay.
                // Publish selection only after nonblocking admission succeeds.
                if (!Queue(socket, queue, Replay(target, projection))) return false;
                if (peers.TryGetValue(socket, out var current))
                {
                    current.Voice?.Cancel();
                    current.BrowserTunnel?.SetCaptureSelection(target);
                    current.Selected = target;
                    current.Cursor = snapshot.Sequence;
                }
                else peers[socket] = new VoicePeer(queue, target, snapshot.Sequence);
                Queue(socket, queue, MeterFrame(target));
                Queue(socket, queue, State(target, projection));
                Queue(socket, queue, new JsonObject
                {
                    ["type"] = "current_settings", ["agent_id"] = AddressOf(target),
                    ["settings"] = AgentSettings(preferences, target)
                });
                if (latestTree is not null) Queue(socket, queue, latestTree);
                return true;
            }
        }

        async Task RefreshTree()
        {
            while (!stopping.IsCancellationRequested)
            {
                var rows = new JsonArray();
                var present = new HashSet<Agent>();
                async Task Visit(AgentTreeNode node, string? parent)
                {
                    try
                    {
                        var status = await node.Agent.WorkStatusAsync(stopping.Token);
                        var goal = await node.Agent.GoalStatusAsync(stopping.Token);
                        present.Add(node.Agent);
                        rows.Add(new JsonObject
                        {
                            // Only this mount's initial Agent has a local empty
                            // address. Other unnamed roots remain unaddressable.
                            ["key"] = node.Address ?? (ReferenceEquals(node.Agent, agent) ? "" : null),
                            ["id"] = node.Agent.ChildID() ?? node.Agent.Name ?? "(unnamed)",
                            ["parent"] = parent,
                            ["state"] = status.Activity.ToString().ToLowerInvariant(),
                            ["paused"] = status.Paused,
                            ["submitted"] = status.Submitted,
                            ["can_delete"] = node.Address is not null && CanDeleteAgent(agent, node.Agent),
                            ["goal"] = GoalValue(goal),
                            ["phase"] = status.Work?.Phase.ToString().ToLowerInvariant() ?? "idle",
                            ["tool"] = status.Work?.Tool ?? "",
                            ["detail"] = status.Work?.Detail ?? ""
                        });
                    }
                    catch (ObjectDisposedException) { return; }
                    foreach (var child in node.Children) await Visit(child, node.Address);
                }
                // Never hold the socket gate while awaiting an actor query.
                foreach (var root in agent.Parent().AgentTree()) await Visit(root, null);
                var frame = new JsonObject { ["type"] = "agent_tree", ["agents"] = rows };
                lock (gate)
                {
                    // Deletion removes an owned conversation, not a socket's
                    // fixed browser relay. Admit a fresh default replay before
                    // changing displays whose selected Agent has disappeared.
                    foreach (var (socket, peer) in peers.ToArray())
                    {
                        if (present.Contains(peer.Selected)) continue;
                        // The tree was captured before its asynchronous status
                        // queries. A newly created/selected Agent may be newer
                        // than that snapshot; recheck current ownership first.
                        string? address = peer.Selected.Address();
                        if (address is not null && ReferenceEquals(agent.Parent().FindAgentAddress(address), peer.Selected)) continue;
                        Select(socket, peer.Queue, agent);
                    }
                    if (latestTree is null || frame.ToJsonString() != latestTree.ToJsonString())
                    {
                        latestTree = frame;
                        foreach (var (socket, peer) in peers) Queue(socket, peer.Queue, frame);
                    }
                }
                // Coalesce activity bursts; this also finds membership-only changes.
                await Task.Delay(200, stopping.Token);
            }
        }

        routes.Map("/ws", async (HttpContext http) =>
        {
            // Loopback is not an origin check. A hostile web page must not gain
            // local tool authority merely by opening this WebSocket endpoint.
            string? origin = RequestOrigin(http, port, publicOrigin);
            bool publicRequest = origin is not null && origin == publicOrigin;
            if (!http.WebSockets.IsWebSocketRequest || origin is null
                || !MatchesOrigin(http.Request, origin, required: publicRequest))
            {
                http.Response.StatusCode = StatusCodes.Status403Forbidden;
                return;
            }
            // Relay identity is chosen at connection admission and never follows
            // display selection. A separate tab may explicitly serve another Agent.
            Agent relayAgent = agent;
            if (http.Request.Query.TryGetValue("agent_id", out var requested))
            {
                Agent? found = requested.Count == 1
                    ? agent.Parent().FindAgentAddress(requested[0] ?? "") : null;
                if (found is null)
                {
                    http.Response.StatusCode = StatusCodes.Status400BadRequest;
                    return;
                }
                relayAgent = found;
            }
            var manager = relayAgent.EnsureMcp();
            string relayID = AddressOf(relayAgent);
            using var socket = await http.WebSockets.AcceptWebSocketAsync();
            using var lifetime = CancellationTokenSource.CreateLinkedTokenSource(
                stopping.Token, http.RequestAborted, host.Lifetime.ApplicationStopping);
            var queue = Channel.CreateBounded<string>(new BoundedChannelOptions(256)
            {
                FullMode = BoundedChannelFullMode.Wait, SingleReader = true, SingleWriter = false
            });
            await using var tunnel = new McpTunnel(queue.Writer, relayID, relayAgent);
            McpBrowser? browser = null;
            async Task Send()
            {
                await foreach (string frame in queue.Reader.ReadAllAsync(lifetime.Token))
                {
                    await socket.SendAsync(Encoding.UTF8.GetBytes(frame), WebSocketMessageType.Text, true, lifetime.Token);
                    if (trace is not null) Trace(null, "<", frame);
                    else
                    {
                        using var sent = JsonDocument.Parse(frame);
                        if (sent.RootElement.GetProperty("type").GetString() == "mcp")
                            Trace(relayAgent, "<sent", frame);
                    }
                }
            }
            var sender = Send();
            VoicePeer? connection = null;
            async Task NotifyVoice(VoiceLease lease, Task<VoiceTranscript> transcription)
            {
                bool ready;
                try { await transcription.ConfigureAwait(false); ready = true; }
                catch { ready = false; }
                lock (gate)
                {
                    if (!peers.TryGetValue(socket, out var peer)
                        || !ReferenceEquals(peer.Voice, lease)
                        || !ReferenceEquals(peer.Selected, lease.Target) || !lease.IsActive()) return;
                    Queue(socket, queue, new JsonObject
                    {
                        ["type"] = ready ? "voice_ready" : "voice_failed",
                        ["agent_id"] = AddressOf(lease.Target), ["sequence"] = lease.Sequence
                    });
                }
            }
            try
            {
                lock (gate)
                {
                    Queue(socket, queue, new JsonObject
                    {
                        ["type"] = "connection", ["relay_agent_id"] = relayID,
                        ["transcription"] = new JsonObject
                        {
                            ["enabled"] = relayAgent.Parent().Transcriber() is not null,
                            ["payer"] = relayAgent.Parent().Transcriber() is null ? "off" : "openai-api-key",
                            ["model"] = relayAgent.Parent().Transcriber() is null ? null : VoiceTranscriber.Model
                        }
                    });
                    // First tab for this relay Agent is its active browser. Viewing
                    // another conversation never destroys that attachment or its clients.
                    if (ReferenceEquals(relayAgent.Mcp(), manager) && manager.Browser() is null)
                    {
                        try
                        {
                            browser = new McpBrowser(manager, tunnel);
                        }
                        catch (InvalidOperationException)
                        {
                            // Agent shutdown may close admission concurrently.
                            // This socket can still observe the finalized replay.
                        }
                    }
                }
                await relayAgent.WorkStatusAsync(lifetime.Token);
                if (!Select(socket, queue, relayAgent)) return;
                lock (gate)
                {
                    connection = peers[socket];
                    connection.BrowserTunnel = tunnel;
                    tunnel.SetCaptureSelection(connection.Selected);
                }
                var buffer = new byte[8192];
                long lastImageSequence = 0;
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
                        if (received.MessageType != WebSocketMessageType.Text || content.Length + received.Count > MaximumImageControlBytes)
                        {
                            throw new InvalidDataException("Control exceeds the bounded image envelope or is not text.");
                        }
                        content.Write(buffer, 0, received.Count);
                    }
                    while (!received.EndOfMessage);
                    long? imageSequence = null;
                    long? voiceSequence = null;
                    try
                    {
                        using var json = JsonDocument.Parse(content.ToArray());
                        var command = json.RootElement;
                        string type = command.GetProperty("type").GetString() ?? "";
                        Agent selected;
                        lock (gate) selected = peers[socket].Selected;
                        Trace(type == "mcp" ? relayAgent : selected, trace is null ? ">received" : ">", command.GetRawText());
                        if (type != "image_prompt" && command.TryGetProperty("images", out _))
                            throw new InvalidDataException("Images require an image_prompt envelope.");
                        bool voiceControl = type is "voice_begin" or "voice_audio" or "voice_cancel" or "voice_commit";
                        if (type != "voice_audio" && command.TryGetProperty("audio", out _))
                            throw new InvalidDataException("Audio requires a voice_audio envelope.");
                        if (type == "voice_audio" && content.Length > MaximumVoiceControlBytes)
                            throw new InvalidDataException("Voice envelope exceeds its bound.");
                        if (voiceControl)
                        {
                            voiceSequence = command.GetProperty("sequence").GetInt64();
                            if (voiceSequence <= 0 || voiceSequence > 9007199254740991L
                                || !command.TryGetProperty("agent_id", out _))
                                throw new InvalidDataException("Voice requires an explicit Agent and sequence.");
                        }
                        // Only the original pending capture RPC may use the
                        // existing bounded image envelope instead of 1 MiB.
                        bool captureReply = type == "mcp" && command.TryGetProperty("data", out var rpc)
                            && tunnel.IsCaptureReply(rpc);
                        if (type is not ("image_prompt" or "voice_audio") && !captureReply
                            && content.Length > MaximumOrdinaryControlBytes)
                            throw new InvalidDataException("Ordinary controls must be under 1 MiB.");
                        if (type == "image_prompt")
                        {
                            imageSequence = command.GetProperty("sequence").GetInt64();
                            if (imageSequence <= lastImageSequence || imageSequence > 9007199254740991L)
                                throw new InvalidDataException("Repeated or invalid image sequence.");
                            // Reservation includes refusals: no automatic retry can
                            // turn an uncertain outcome into another admission.
                            lastImageSequence = imageSequence.Value;
                            if (!command.TryGetProperty("agent_id", out _))
                                throw new InvalidDataException("Image messages require an explicit Agent.");
                        }
                        // Capture the acknowledged display target for this command.
                        // Tree recovery may change future admissions, never this one.
                        string expectedID = type == "mcp" ? relayID : AddressOf(selected);
                        if (type != "select_agent" && command.TryGetProperty("agent_id", out var target)
                            && (target.ValueKind != JsonValueKind.String || target.GetString() != expectedID))
                        {
                            throw new InvalidDataException("Command targets a different Agent.");
                        }
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
                            case "select_agent":
                                string address = command.GetProperty("agent_id").GetString() ?? "";
                                Agent? choice = address.Length == 0 && agent.Address() is null
                                    ? agent : agent.Parent().FindAgentAddress(address);
                                if (choice is null) throw new InvalidDataException("Unknown Agent address.");
                                await choice.WorkStatusAsync(lifetime.Token);
                                await connection!.ReleasePauseAsync();
                                Select(socket, queue, choice);
                                break;
                            case "subscribe":
                                // The initial snapshot is already queued atomically
                                // with registration. Subscription is idempotent.
                                break;
                            case "get_goal":
                            case "update_goal":
                                GoalStatus goalStatus;
                                try
                                {
                                    goalStatus = type == "get_goal"
                                        ? await selected.GoalStatusAsync(lifetime.Token)
                                        : await selected.GoalControlAsync(ReadGoalUpdate(command.GetProperty("goal")), lifetime.Token);
                                }
                                catch (Exception error) when (error is IOException or UnauthorizedAccessException)
                                {
                                    // A refused file selection must not close the socket or
                                    // reveal an operating-system path in an error frame.
                                    throw new InvalidDataException("Goal control refused");
                                }
                                lock (gate)
                                {
                                    // Broadcast only to views of this Agent. The frame carries
                                    // its identity so a queued old reply cannot change a new view.
                                    foreach (var (goalSocket, goalPeer) in peers)
                                    {
                                        if (!ReferenceEquals(goalPeer.Selected, selected)) continue;
                                        if (type == "get_goal" && goalSocket != socket) continue;
                                        Queue(goalSocket, goalPeer.Queue, new JsonObject
                                        {
                                            ["type"] = type == "get_goal" ? "goal_status" : "goal_updated",
                                            ["agent_id"] = AddressOf(selected), ["goal"] = GoalValue(goalStatus)
                                        });
                                    }
                                }
                                break;
                            case "get_settings":
                                await selected.WorkStatusAsync(lifetime.Token);
                                lock (gate)
                                {
                                    Queue(socket, queue, new JsonObject
                                    {
                                        ["type"] = "current_settings", ["agent_id"] = AddressOf(selected),
                                        ["settings"] = AgentSettings(preferences, selected)
                                    });
                                }
                                break;
                            case "update_settings":
                                var patch = command.GetProperty("settings");
                                bool memoryUpdate = patch.TryGetProperty("memory", out var memoryPatch);
                                bool recallUpdate = patch.TryGetProperty("recall", out var recallPatch);
                                if (memoryUpdate || recallUpdate)
                                {
                                    // Policy and appearance have different owners/stores.
                                    // A patch updates one policy section; never imply an
                                    // atomic operation spanning appearance and Agent state.
                                    if (patch.EnumerateObject().Count() != 1)
                                    {
                                        throw new InvalidDataException("Agent policy sections must be updated separately");
                                    }
                                    await agentSettingsGate.WaitAsync(lifetime.Token);
                                    try
                                    {
                                        await selected.WorkStatusAsync(lifetime.Token);
                                        // Once admitted, await the definite actor outcome even
                                        // if this tab disconnects. Observer cancellation cannot
                                        // roll back a durable settings replacement.
                                        if (memoryUpdate)
                                        {
                                            var desired = MergeMemorySettings(memoryPatch, selected.MemorySettings());
                                            await selected.SaveMemorySettingsAsync(desired);
                                        }
                                        else
                                        {
                                            var desired = MergeRecallSettings(recallPatch, selected.RecallSettings());
                                            await selected.SaveRecallSettingsAsync(desired);
                                        }
                                        lock (gate)
                                        {
                                            foreach (var (peerSocket, peer) in peers.ToArray())
                                            {
                                                if (!ReferenceEquals(peer.Selected, selected))
                                                {
                                                    continue;
                                                }
                                                Queue(peerSocket, peer.Queue, new JsonObject
                                                {
                                                    ["type"] = "settings_updated", ["agent_id"] = AddressOf(selected),
                                                    ["settings"] = AgentSettings(preferences, selected)
                                                });
                                            }
                                        }
                                    }
                                    catch (Exception error) when (error is IOException or InvalidDataException
                                        or UnauthorizedAccessException or ArgumentException or InvalidOperationException
                                        or FormatException or JsonException)
                                    {
                                        lock (gate)
                                        {
                                            Queue(socket, queue, new JsonObject
                                            {
                                                ["type"] = "error",
                                                ["message"] = (memoryUpdate ? "Memory" : "Recall") + " settings were not acknowledged. Reload settings to inspect the current policy; a saved change may apply on restart."
                                            });
                                        }
                                    }
                                    finally
                                    {
                                        agentSettingsGate.Release();
                                    }
                                    break;
                                }
                                lock (gate)
                                {
                                    var candidate = MergeSettings(command.GetProperty("settings"), preferences, selected.MaxToolRounds(), selected.ContextByteBudget());
                                    var saved = Settings(candidate.Preferences, candidate.Budget, candidate.ContextTarget);
                                    try
                                    {
                                        // Serialize save/publication across clients. Only after
                                        // replacement succeeds may new turns see the new policy.
                                        SaveSettings(agent, settingsPath, saved);
                                        selected.SetMaxToolRounds(candidate.Budget);
                                        selected.SetContextByteBudget(candidate.ContextTarget);
                                        preferences = candidate.Preferences;
                                        foreach (var (peerSocket, peer) in peers.ToArray())
                                        {
                                            Queue(peerSocket, peer.Queue, new JsonObject
                                            {
                                                ["type"] = "settings_updated", ["agent_id"] = AddressOf(peer.Selected),
                                                ["settings"] = AgentSettings(preferences, peer.Selected)
                                            });
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
                            case "image_prompt":
                                var imageAdmission = PrepareImageInput(selected, command, lifetime.Token);
                                await selected.PostAsync(imageAdmission.Input);
                                lock (gate)
                                {
                                    Queue(socket, queue, new JsonObject
                                    {
                                        ["type"] = "image_accepted", ["agent_id"] = AddressOf(selected),
                                        ["sequence"] = imageSequence,
                                        ["images"] = JsonSerializer.SerializeToNode(imageAdmission.Images.Select(image => new
                                        {
                                            original_width = image.OriginalWidth, original_height = image.OriginalHeight,
                                            width = image.Width, height = image.Height, resized = image.Resized,
                                            source_bytes = image.SourceBytes, output_bytes = image.OutputBytes
                                        }))
                                    });
                                }
                                break;
                            case "voice_begin":
                                if (voiceSequence <= connection.VoiceSequence)
                                    throw new InvalidDataException("Repeated voice sequence.");
                                connection.VoiceSequence = voiceSequence!.Value;
                                if (!connection.VoiceNotification.IsCompleted)
                                    throw new InvalidOperationException("Previous voice work is still closing.");
                                await connection.DestroyVoiceAsync();
                                if (selected.Parent().Transcriber() is null)
                                    throw new InvalidOperationException("Separate transcription funding is disabled.");
                                lock (gate)
                                {
                                    // Recheck after the awaited cleanup: tree recovery may
                                    // have changed selection while this command was waiting.
                                    if (!ReferenceEquals(connection.Selected, selected))
                                        throw new InvalidOperationException("Voice selection changed.");
                                    connection.Voice = new VoiceLease(selected, voiceSequence.Value, lifetime.Token);
                                    Queue(socket, queue, new JsonObject { ["type"] = "voice_begun",
                                        ["agent_id"] = AddressOf(selected), ["sequence"] = voiceSequence });
                                }
                                break;
                            case "voice_audio":
                            case "voice_cancel":
                            case "voice_commit":
                                VoiceLease lease;
                                lock (gate)
                                {
                                    lease = connection.Voice ?? throw new InvalidOperationException("No voice capture.");
                                    if (!ReferenceEquals(lease.Target, selected) || lease.Sequence != voiceSequence
                                        || !ReferenceEquals(connection.Selected, selected))
                                        throw new InvalidOperationException("Stale voice capture.");
                                }
                                if (type == "voice_audio")
                                {
                                    var audio = PrepareVoiceAudio(command);
                                    // Keep receive available for cancel/select/disconnect.
                                    connection.VoiceNotification = NotifyVoice(lease, lease.Start(audio));
                                }
                                else if (type == "voice_cancel")
                                {
                                    lease.Cancel();
                                    lock (gate) Queue(socket, queue, new JsonObject { ["type"] = "voice_cancelled",
                                        ["agent_id"] = AddressOf(selected), ["sequence"] = voiceSequence });
                                }
                                else
                                {
                                    Task<string> admission;
                                    lock (gate)
                                    {
                                        if (!ReferenceEquals(connection.Selected, selected))
                                            throw new InvalidOperationException("Voice selection changed.");
                                        admission = lease.CommitAsync(selected, voiceSequence!.Value);
                                    }
                                    await admission;
                                    lock (gate) Queue(socket, queue, new JsonObject { ["type"] = "voice_accepted",
                                        ["agent_id"] = AddressOf(selected), ["sequence"] = voiceSequence });
                                }
                                break;
                            case "prompt":
                                await selected.PostAsync(new AgentInput(InputKind.Message, text));
                                break;
                            case "hint":
                                await selected.PostAsync(new AgentInput(InputKind.Hint, text));
                                break;
                            case "automatic_pause":
                                if (text is not ("held" or "released"))
                                    throw new InvalidDataException("Invalid automatic pause state.");
                                await connection!.SetPauseAsync(selected, text == "held");
                                break;
                            case "pause":
                                await selected.PostAsync(new AgentInput(InputKind.Pause));
                                break;
                            case "unpause":
                                await selected.PostAsync(new AgentInput(InputKind.Resume));
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
                                await selected.PostAsync(new AgentInput(interrupt));
                                break;
                            case "set_model":
                                try
                                {
                                    await selected.SetModelAsync(text);
                                    Broadcast(selected, null);
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
                                    Queue(socket, queue, MeterFrame(selected));
                                }
                                break;
                            case "clear_tools":
                                await selected.PostAsync(new AgentInput(InputKind.ClearTools));
                                break;
                            case "shutdown":
                                await selected.ShutdownAsync();
                                break;
                            default:
                                throw new ArgumentException("Unknown GUI control");
                        }
                    }
                    catch (Exception error) when (error is JsonException or ArgumentException or KeyNotFoundException or InvalidOperationException or InvalidDataException or IOException or UnauthorizedAccessException or FormatException or NotSupportedException)
                    {
                        lock (gate)
                        {
                            // This sealed exception carries fixed application guidance,
                            // never provider text, file paths or credentials.
                            string refusal = error is ImageBudgetExceededException
                                ? error.Message : "Invalid or unavailable control";
                            Queue(socket, queue, new JsonObject { ["type"] = "error", ["message"] = refusal, ["image_sequence"] = imageSequence, ["voice_sequence"] = voiceSequence });
                        }
                    }
                }
            }
            catch (Exception error) when (error is OperationCanceledException or WebSocketException or IOException or ObjectDisposedException)
            {
                // A selected peer can close after lookup; selection is not a lease.
                // Disconnect ends this tab's browser tools, not the Agent turn.
                // An in-flight call reports its real connection failure.
            }
            finally
            {
                lock (gate)
                {
                    if (connection is not null) connection.BrowserTunnel = null;
                    peers.Remove(socket);
                    queue.Writer.TryComplete();
                }
                lifetime.Cancel();
                socket.Abort();
                try
                {
                    if (browser is not null) await browser.DestroyAsync();
                }
                finally
                {
                    try
                    {
                        try { await sender; }
                        catch (Exception error) when (error is OperationCanceledException or WebSocketException or IOException)
                        {
                            // The single sender is joined before disposing its socket.
                        }
                    }
                    finally
                    {
                        if (connection is not null)
                        {
                            try { await connection.ReleasePauseAsync(); }
                            finally { await connection.DestroyVoiceAsync(); }
                        }
                    }
                }
            }
        });
        lock (gate) { ObserveOnce(agent); }
        var treeRefresh = RefreshTree();
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
                    await initialMcp.WaitForBrowserAsync(stopping.Token);
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
                Task[] readers;
                lock (gate) { readers = observers.Values.ToArray(); }
                await Task.WhenAll(readers.Append(treeRefresh));
            }
            catch (OperationCanceledException) when (stopping.IsCancellationRequested)
            {
            }
            // Borrowed hosting must not cancel the caller's shared credential
            // attempts. The caller owns that lifecycle as well as its host.
        }
    }
}
