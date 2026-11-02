using System.Net;
using System.Net.Sockets;
using System.Text.Json;
using Ensemble;
using Application = Ensemble.Ensemble;

internal static partial class ActorTests
{
    public static async Task ChildRestartReminder(string temp)
    {
        string home = Path.Combine(temp, "child-restart-reminder");
        string skills = Path.Combine(home, "skills");
        Directory.CreateDirectory(Path.Combine(skills, "base"));
        Directory.CreateDirectory(Path.Combine(skills, "goals"));
        File.WriteAllText(Path.Combine(skills, "base", "SKILL.md"),
            "---\nname: base\ntype: primary\ndescription: Restart reminder fixture\ntools: submit read_file\nloadable-skills: goals\n---\nInspect saved work before continuing.\n$TOOLS\n$SKILLS");
        File.Copy(Path.Combine(Directory.GetCurrentDirectory(), "agent/skills/goals/SKILL.md"),
            Path.Combine(skills, "goals", "SKILL.md"));
        string goal = Path.Combine(home, "goal.md");
        File.WriteAllText(goal, "RESTART_GOAL: inspect receipts; do not replay tools.");
        using var deadline = new CancellationTokenSource(TimeSpan.FromSeconds(20));
        var listener = new TcpListener(IPAddress.Loopback, 0);
        listener.Start();
        var configuration = new Configuration(Vendor.Anthropic, "restart-reminder",
            "http://127.0.0.1:" + ((IPEndPoint)listener.LocalEndpoint).Port, "local", sandboxRoot: home);
        int requests = 0;
        Task? server = null;
        var app = new Application(TextWriter.Null);
        try
        {
            var parent = new Agent(app, configuration, skillSpec: new SkillSpec(skills, "base"), dataDirectory: home);
            Agent child = await parent.CreateChildAsync("worker", cancellation: deadline.Token);
            child.History().Append(new Event(EventKind.MessageReceived, actor: Actor.Human,
                parts: [new TextPart("SAVED_WORK: use this existing conversation.")]));
            var loaded = await child.DispatchAsync(new ToolCallPart("goals", configuration.Target(),
                "load_skill", JsonSerializer.SerializeToElement(new { name = "goals" })), deadline.Token);
            Check(!loaded.IsError, "Could not activate reminder fixture.");
            await child.GoalControlAsync(new GoalUpdate(goal, true, .5), deadline.Token);
            await parent.ShutdownAsync();
            Check(!listener.Pending(), "Original process fired before the restart fixture shut down.");
        }
        finally { app.Destroy(); }

        var restarted = new Application(TextWriter.Null);
        try
        {
            var firstRequest = new TaskCompletionSource(TaskCreationOptions.RunContinuationsAsynchronously);
            server = Task.Run(async () =>
            {
                using (var socket = await listener.AcceptTcpClientAsync(deadline.Token))
                using (var stream = socket.GetStream())
                {
                    string request = await ReadRequest(stream, deadline.Token);
                    Interlocked.Increment(ref requests);
                    Check(request.Contains("RESTART_GOAL") && request.Contains("SAVED_WORK")
                        && request.Contains("Automatic goal reminder"),
                        "Restart wake lost its goal, conversation or explicit reminder identity.");
                    firstRequest.SetResult();
                    await ReplyJson(stream, new {
                        model = "restart-reminder", content = new object[] {
                            new { type = "tool_use", id = "restart-submit", name = "submit",
                                input = new { result = "Completed resumed work." } }
                        }, usage = new { input_tokens = 1, output_tokens = 1 }
                    }, deadline.Token);
                }
                using (var socket = await listener.AcceptTcpClientAsync(deadline.Token))
                using (var stream = socket.GetStream())
                {
                    string request = await ReadRequest(stream, deadline.Token);
                    Interlocked.Increment(ref requests);
                    Check(request.Contains("restart-submit"), "Resumed submit result lost pairing.");
                    await Reply(stream, "Submitted.", deadline.Token);
                }
            });
            var parent = new Agent(restarted, configuration, skillSpec: new SkillSpec(skills, "base"), dataDirectory: home);
            await parent.RestoreChildrenAsync(deadline.Token);
            Agent child = parent.FindChild("worker")!;
            GoalStatus status = await child.GoalStatusAsync(deadline.Token);
            Check(status.Enabled && status.DueUtc > DateTimeOffset.UtcNow,
                "Restoration did not schedule a fresh future interval.");
            await Task.Delay(100, deadline.Token);
            Check(!firstRequest.Task.IsCompleted, "Restart immediately replayed work instead of waiting for its interval.");
            await child.WaitForSubmissionAsync(deadline.Token);
            await server;
            await child.FlushAsync(deadline.Token);
            Check(requests == 2 && !(await child.GoalStatusAsync(deadline.Token)).Enabled,
                "Resumed work did not submit once and disable reminders.");
        }
        finally
        {
            deadline.Cancel();
            listener.Stop();
            restarted.Destroy();
            if (server is not null)
            {
                try { await server; }
                catch (OperationCanceledException) { }
                catch (SocketException) when (deadline.IsCancellationRequested) { }
            }
        }
    }
}
