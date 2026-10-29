using System.Net;
using System.Net.Sockets;
using Ensemble;
using Application = Ensemble.Ensemble;

internal static class ChildOwnershipTests
{
    public static async Task Relationships(string directory)
    {
        string root = Path.Combine(directory, "children");
        Directory.CreateDirectory(root);
        var app = new Application(TextWriter.Null);
        var other = new Application(TextWriter.Null);
        var config = new Configuration(Vendor.Anthropic, "model-a", "http://127.0.0.1:1",
            safeMode: true, sandboxRoot: root);
        try
        {
            var parent = new Agent(app, config, name: "main");
            var peer = new Agent(app, config, name: "peer");
            var child = new Agent(app, parent, "reviewer", config);
            var sameID = new Agent(app, peer, "reviewer", config);
            var grandchild = new Agent(app, child, "reviewer", config);
            Check(child.AgentParent() == parent && child.Parent() == app
                && child.ChildID() == "reviewer" && parent.Children().SequenceEqual([child])
                && parent.FindChild("reviewer") == child && peer.FindChild("reviewer") == sameID,
                "Child ownership or parent-local addressing is inconsistent");
            Check(app.Agents().Length == 5 && app.FindAgent("main") == parent
                && app.FindAgent("reviewer") is null, "Child ID entered the standalone name index");
            Check(child.Configuration().SafeMode && child.Sandbox().Root() == parent.Sandbox().Root(),
                "Child did not inherit effective confinement");
            var before = app.Agents();
            Refuses(() => new Agent(app, parent, "reviewer", config));
            Refuses(() => new Agent(other, parent, "wrong-root", config));
            Refuses(() => new Agent(app, parent, "../escape", config));
            Refuses(() => new Agent(app, parent, "wide", new Configuration(Vendor.Anthropic,
                "model-a", "http://127.0.0.1:1", enableWebSearch: true)));
            Refuses(() => peer.RemoveChild(child));
            Refuses(() => peer.AddChild(child));
            Refuses(() => new Agent(app, parent, "bad-output", config,
                Path.Combine(directory, "outside-output")));
            Check(app.Agents().SequenceEqual(before) && parent.FindChild("reviewer") == child
                && peer.Children().SequenceEqual([sameID]) && child.AgentParent() == parent,
                "Refusal partially changed an ownership relationship");

            // Targeted destruction removes the subtree, not its sibling or owner.
            child.Destroy();
            Check(parent.Children().IsEmpty && child.AgentParent() is null
                && grandchild.AgentParent() is null && app.Agents().Length == 3,
                "Targeted destruction left a descendant attached");
            Check(peer.FindChild("reviewer") == sameID && parent.Configuration() == config,
                "Targeted destruction damaged an unrelated Agent");
            var replacement = new Agent(app, parent, "reviewer", config);
            await parent.ShutdownAsync();
            Check(parent.FindChild("reviewer") == replacement,
                "Shutdown destroyed ownership instead of only joining work");
            Refuses(() => new Agent(app, parent, "late", config));
            Refuses(() => replacement.SubmitAsync([new TextPart("must not run")]));
            parent.Destroy();
            Check(replacement.AgentParent() is null && app.Agents().SequenceEqual([peer, sameID]),
                "Destruction did not remove both strong parents");
        }
        finally
        {
            app.Destroy();
            other.Destroy();
        }
        Check(app.Agents().IsEmpty, "Root cascade retained children");
    }

    public static void CleanupFailure(string directory)
    {
        string workspace = Path.Combine(directory, "children-failure");
        Directory.CreateDirectory(workspace);
        var app = new Application(TextWriter.Null);
        var config = new Configuration(Vendor.Anthropic, "model-a", "http://127.0.0.1:1",
            sandboxRoot: workspace);
        var parent = new Agent(app, config);
        var broken = new Agent(app, parent, "broken", config);
        var healthy = new Agent(app, parent, "healthy", config);
        try
        {
            broken.StartPersistence(new PersistenceOptions("broken/state.json"));
            healthy.StartPersistence(new PersistenceOptions("healthy/state.json"));
            string journal = Path.Combine(workspace, "broken", "state.json.journal.jsonl");
            File.Delete(journal);
            Directory.CreateDirectory(journal);
            Refuses(() => broken.History().Append(new Event(EventKind.MessageReceived,
                actor: Actor.Human, parts: [new TextPart("must not publish")])));
            bool failed = false;
            try
            {
                parent.Destroy();
            }
            catch (Exception)
            {
                failed = true;
            }
            Check(failed, "Poisoned child cleanup incorrectly reported success");
            Check(parent.Children().IsEmpty && app.Agents().IsEmpty
                && broken.AgentParent() is null && healthy.AgentParent() is null,
                "One cleanup failure prevented another ownership detach");
            Check(File.Exists(Path.Combine(workspace, "healthy", "state.json")),
                "Healthy sibling did not finish its save before destruction");
        }
        finally
        {
            app.Destroy();
        }
    }

    private static void Refuses(Action action)
    {
        try
        {
            action();
        }
        catch (Exception error) when (error is ArgumentException or InvalidOperationException
            or IOException or UnauthorizedAccessException)
        {
            return;
        }
        throw new Exception("Expected refusal");
    }

    private static void Check(bool condition, string message)
    {
        if (!condition)
        {
            throw new Exception(message);
        }
    }
}

internal static partial class ActorTests
{
    public static async Task ChildTreeShutdown()
    {
        using var deadline = new CancellationTokenSource(TimeSpan.FromSeconds(15));
        var listener = new TcpListener(IPAddress.Loopback, 0);
        listener.Start();
        var arrived = new TaskCompletionSource(TaskCreationOptions.RunContinuationsAsynchronously);
        var server = Task.Run(async () =>
        {
            using var client = await listener.AcceptTcpClientAsync(deadline.Token);
            using var stream = client.GetStream();
            await ReadRequest(stream, deadline.Token);
            arrived.TrySetResult();
            // Withhold the response. Shutdown must cancel the actual request,
            // not wait for a cooperative model completion or a polling timeout.
            byte[] buffer = new byte[1];
            Check(await stream.ReadAsync(buffer, deadline.Token) == 0,
                "Expected client disconnect during shutdown");
        });
        var app = new Application(TextWriter.Null);
        try
        {
            int port = ((IPEndPoint)listener.LocalEndpoint).Port;
            var config = new Configuration(Vendor.Anthropic, "model-a", "http://127.0.0.1:" + port);
            var parent = new Agent(app, config);
            var child = new Agent(app, parent, "child", config);
            var grandchild = new Agent(app, child, "grandchild", config);
            Task<string> request = grandchild.SubmitAsync([new TextPart("pending")]);
            await arrived.Task.WaitAsync(deadline.Token);
            Task stop = parent.ShutdownAsync();
            Check(ReferenceEquals(stop, parent.ShutdownAsync()), "Shutdown does not share its join");
            await stop.WaitAsync(deadline.Token);
            try
            {
                await request.WaitAsync(deadline.Token);
                throw new Exception("Stopped request reported success");
            }
            catch (Exception error) when (error is ObjectDisposedException or OperationCanceledException)
            {
                Check(!deadline.IsCancellationRequested, "Deadline, not shutdown, ended the request");
            }
            await server.WaitAsync(deadline.Token);
            Check(new[] { parent, child, grandchild }.All(agent => agent.History().Events()
                .Any(fact => fact.Kind == EventKind.AgentShutdown)), "A descendant did not complete shutdown");
            parent.Destroy();
            Check(app.Agents().IsEmpty, "Joined subtree retained a root membership");
        }
        finally
        {
            deadline.Cancel();
            listener.Stop();
            app.Destroy();
            try { await server; } catch (OperationCanceledException) { }
        }
    }
}
