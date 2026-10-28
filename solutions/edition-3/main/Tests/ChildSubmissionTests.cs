using System.Net;
using System.Net.Sockets;
using System.Text.Json;
using Ensemble;
using Application = Ensemble.Ensemble;

internal static class ChildSubmissionTests
{
    private static JsonElement Json(string text) => JsonDocument.Parse(text).RootElement.Clone();
    private static void Check(bool condition, string message)
    {
        if (!condition) throw new Exception(message);
    }
    private static void Refuses(Action action)
    {
        try { action(); }
        catch (Exception error) when (error is ArgumentException or InvalidOperationException or IOException or UnauthorizedAccessException) { return; }
        throw new Exception("Expected refusal.");
    }

    public static void Durable(string temp)
    {
        string work = Path.Combine(temp, "child-submit-durable");
        Directory.CreateDirectory(work);
        var app = new Application(TextWriter.Null);
        try
        {
            var config = new Configuration(Vendor.Anthropic, "fixture", "http://127.0.0.1:1", "local-only", sandboxRoot: work);
            var parent = new Agent(app, config);
            Refuses(() => parent.PublishSubmission("no parent"));
            int before = app.Agents().Length;
            Refuses(() => new Agent(app, parent, "bad", config, submitSchema: Json("{\"type\":\"array\"}")));
            Check(app.Agents().Length == before && parent.Children().IsEmpty,
                "Invalid schema registered an Agent.");
            JsonElement schema = Json("""{"type":"object","required":["summary","items"],"additionalProperties":false,"properties":{"summary":{"type":"string"},"items":{"type":"array","items":{"type":["number","null","boolean"]}}}}""");
            var child = new Agent(app, parent, "reviewer", config, submitSchema: schema);
            Refuses(() => child.PublishSubmission(Json("""{"summary":"ok","items":[]}""")));
            Check(child.Submission() is null, "Unconfigured child published completion.");
            child.Destroy();
            string childHome = Path.Combine(work, "children", "reviewer");
            child = new Agent(app, parent, "reviewer", config, submitSchema: schema, dataDirectory: childHome);
            string save = Path.Combine(childHome, "session", "save.json");
            Check(!child.StartPersistence(new PersistenceOptions(save)), "Unexpected restore.");
            Check(child.DataDirectory() == childHome, "History persistence changed the actor home.");
            child.History().Append(new Event(EventKind.MessageReceived, actor: Actor.Human,
                parts: [new TextPart("reviewer task survives reconstruction")]));
            JsonElement payload = Json("""{"summary":"done","items":[9007199254740993,null,false]}""");
            JsonElement published = child.PublishSubmission(payload);
            Check(published.GetProperty("items")[0].GetRawText() == "9007199254740993", "Typed number changed.");
            string path = Path.Combine(child.DataDirectory(), "submit.md");
            string saved = File.ReadAllText(path);
            using (var row = JsonDocument.Parse(saved))
            {
                Check(row.RootElement.GetProperty("version").GetInt32() == 1, "Wrong row version.");
                Check(row.RootElement.GetProperty("agent_id").GetString() == "reviewer", "Wrong owner.");
                Check(JsonElement.DeepEquals(row.RootElement.GetProperty("submit_schema"), schema), "Schema was not retained.");
                Check(JsonElement.DeepEquals(row.RootElement.GetProperty("payload"), payload), "Payload was not retained.");
            }
            Refuses(() => child.PublishSubmission("cannot bypass custom schema"));
            Refuses(() => child.PublishSubmission(Json("""{"summary":"invalid"}""")));
            Check(File.ReadAllText(path) == saved && JsonElement.DeepEquals(child.Submission()!.Value, payload),
                "Invalid submission changed durable or visible completion.");
            child.Destroy();
            Check(File.Exists(save) && File.ReadAllText(path) == saved, "Destruction erased saved work.");
            var restored = new Agent(app, parent, "reviewer", config, submitSchema: schema, dataDirectory: childHome);
            Check(restored.StartPersistence(new PersistenceOptions(save)), "Own History did not restore.");
            Check(restored.History().Events().SelectMany(e => e.Parts).OfType<TextPart>()
                .Any(p => p.Text == "reviewer task survives reconstruction"), "Restored wrong conversation.");
            Check(restored.Submission() is null, "Old artifact became fresh completion.");
            var sibling = new Agent(app, parent, "ordinary", config,
                dataDirectory: Path.Combine(work, "children", "ordinary"));
            sibling.StartPersistence(new PersistenceOptions(Path.Combine(work, "children", "ordinary", "save.json")));
            JsonElement legacy = sibling.PublishSubmission(Json("""{"result":"legacy","ignored":false}"""));
            Check(legacy.EnumerateObject().Count() == 1 && legacy.GetProperty("result").GetString() == "legacy",
                "Default sibling inherited custom schema or failed normalization.");
            Check(restored.Submission() is null, "Sibling submission leaked across Agents.");
        }
        finally { app.Destroy(); }
    }

    public static async Task ConcurrentPolicy(string temp)
    {
        string work = Path.Combine(temp, "child-submit-policy");
        Directory.CreateDirectory(work);
        var app = new Application(TextWriter.Null);
        try
        {
            var config = new Configuration(Vendor.Anthropic, "fixture", "http://127.0.0.1:1", "local-only",
                safeMode: true, sandboxRoot: work);
            var parent = new Agent(app, config);
            var child = new Agent(app, parent, "child", config, dataDirectory: Path.Combine(work, "child"));
            child.StartPersistence(new PersistenceOptions(Path.Combine(work, "child", "save.json")));
            var start = new TaskCompletionSource(TaskCreationOptions.RunContinuationsAsynchronously);
            Task publishing = Task.Run(async () =>
            {
                await start.Task;
                for (int i = 0; i < 100; i++) child.PublishSubmission("revision " + i);
            });
            Task deriving = Task.Run(async () =>
            {
                await start.Task;
                for (int i = 0; i < 1000; i++)
                {
                    var policy = child.Sandbox().DeriveChild(null);
                    Check(policy.Refusal is null && policy.Effective.SafeMode && policy.Effective.SandboxRoot == work,
                        "Concurrent derivation lost inherited restrictions.");
                }
            });
            start.SetResult();
            await Task.WhenAll(publishing, deriving).WaitAsync(TimeSpan.FromSeconds(15));
            Check(child.Submission()!.Value.GetProperty("result").GetString() == "revision 99",
                "Concurrent policy derivation interrupted publication.");
        }
        finally { app.Destroy(); }
    }

    public static void FailedPublication(string temp)
    {
        string work = Path.Combine(temp, "child-submit-failure");
        Directory.CreateDirectory(work);
        string outside = Path.Combine(temp, "submission-sentinel");
        File.WriteAllText(outside, "outside unchanged");
        var app = new Application(TextWriter.Null);
        try
        {
            var config = new Configuration(Vendor.Anthropic, "fixture", "http://127.0.0.1:1", "local-only", sandboxRoot: work);
            var parent = new Agent(app, config);
            var child = new Agent(app, parent, "child", config, dataDirectory: Path.Combine(work, "child"));
            child.StartPersistence(new PersistenceOptions(Path.Combine(work, "child", "save.json")));
            child.PublishSubmission("first");
            string path = Path.Combine(child.DataDirectory(), "submit.md");
            string saved = File.ReadAllText(path);
            File.Move(path, path + ".retained");
            Directory.CreateDirectory(path);
            Refuses(() => child.PublishSubmission("rename must fail"));
            Check(child.Submission()!.Value.GetProperty("result").GetString() == "first",
                "Failed replacement announced success.");
            Check(File.ReadAllText(path + ".retained") == saved, "Failed replacement damaged earlier artifact.");
            Check(!Directory.EnumerateFiles(child.DataDirectory(), ".submit.md.*.tmp").Any(), "Owned temporary leaked.");
            Directory.Delete(path);
            File.CreateSymbolicLink(path, outside);
            Refuses(() => child.PublishSubmission("outside must fail"));
            Check(File.ReadAllText(outside) == "outside unchanged", "Submission escaped confinement.");
            File.Delete(path);
            File.Move(path + ".retained", path);
            child.ShutdownAsync().GetAwaiter().GetResult();
            Refuses(() => child.PublishSubmission("too late"));
            Check(File.ReadAllText(path) == saved, "Stopped child replaced durable completion.");
        }
        finally { app.Destroy(); }
    }
}

internal static partial class ActorTests
{
    public static async Task ChildSubmissionGenerations(string temp)
    {
        string work = Path.Combine(temp, "child-submit-generations");
        Directory.CreateDirectory(work);
        var listener = new TcpListener(IPAddress.Loopback, 0);
        listener.Start();
        using var deadline = new CancellationTokenSource(TimeSpan.FromSeconds(15));
        var first = new TaskCompletionSource(TaskCreationOptions.RunContinuationsAsynchronously);
        var release = new TaskCompletionSource(TaskCreationOptions.RunContinuationsAsynchronously);
        var second = new TaskCompletionSource(TaskCreationOptions.RunContinuationsAsynchronously);
        var finish = new TaskCompletionSource(TaskCreationOptions.RunContinuationsAsynchronously);
        var server = Task.Run(async () =>
        {
            for (int index = 0; index < 3; index++)
            {
                using var socket = await listener.AcceptTcpClientAsync(deadline.Token);
                using var stream = socket.GetStream();
                await ReadRequest(stream, deadline.Token);
                if (index == 0)
                {
                    first.TrySetResult();
                    await release.Task.WaitAsync(deadline.Token);
                }
                else if (index == 2)
                {
                    second.TrySetResult();
                    await finish.Task.WaitAsync(deadline.Token);
                }
                // Index 1 completes the original turn's required live-hint
                // continuation; index 2 belongs to the separately queued task.
                await Reply(stream, "ordinary final text is not submission", deadline.Token);
            }
        });
        var app = new Application(TextWriter.Null);
        try
        {
            var config = new Configuration(Vendor.Anthropic, "fixture",
                "http://127.0.0.1:" + ((IPEndPoint)listener.LocalEndpoint).Port, "local-only", sandboxRoot: work);
            var parent = new Agent(app, config);
            var child = new Agent(app, parent, "child", config, dataDirectory: Path.Combine(work, "child"));
            child.StartPersistence(new PersistenceOptions(Path.Combine(work, "child", "save.json")));
            Task<string> turn = child.SubmitAsync([new TextPart("first task")]);
            await first.Task.WaitAsync(deadline.Token);
            child.PublishSubmission("early");
            // Drain the first acceptance, then wait for exactly the second. This
            // is actor admission evidence, not a delay or a model-call count.
            await child.WaitAsync(o => o.Kind == ObservationKind.Input && o.Accepted == true, deadline.Token);
            Task<string> next = child.SubmitAsync([new TextPart("second task")]);
            await child.WaitAsync(o => o.Kind == ObservationKind.Input && o.Accepted == true, deadline.Token);
            Check(child.Submission() is null, "New admitted work retained early completion.");
            Task<JsonElement> currentSubmission = child.WaitForSubmissionAsync(deadline.Token);
            Check(!currentSubmission.IsCompleted, "Wait accepted an earlier generation submission.");
            await using var joined = parent.JoinChildrenAsync(cancellation: deadline.Token).GetAsyncEnumerator();
            Task<bool> joinedResult = joined.MoveNextAsync().AsTask();
            Check(!joinedResult.IsCompleted, "Join accepted an earlier generation submission.");
            bool refused = false;
            try { child.PublishSubmission("old turn cannot finish new task"); }
            catch (InvalidOperationException) { refused = true; }
            Check(refused, "Old turn published for queued work.");
            // A hint joins the live conversation; it must not orphan the already
            // queued task behind a generation that no request will ever start.
            await child.PostAsync(new AgentInput(InputKind.Hint, "detail for pending work"));
            release.TrySetResult();
            await turn.WaitAsync(deadline.Token);
            await second.Task.WaitAsync(deadline.Token);
            Check(child.Submission() is null, "Plain final text marked submitted.");
            child.PublishSubmission("second done");
            Check((await currentSubmission).GetProperty("result").GetString() == "second done",
                "Current-generation publication did not wake the waiting observer.");
            Check(await joinedResult
                && joined.Current.Submissions!["child"].GetProperty("result").GetString() == "second done",
                "Join did not return the current accepted generation.");
            finish.TrySetResult();
            await next.WaitAsync(deadline.Token);
            Check(child.Submission()!.Value.GetProperty("result").GetString() == "second done", "Current submission lost.");
            await server;
        }
        finally
        {
            release.TrySetResult();
            finish.TrySetResult();
            app.Destroy();
            deadline.Cancel();
            listener.Stop();
            try { await server; }
            catch (Exception error) when (error is OperationCanceledException or IOException or SocketException) { }
        }
    }
}
