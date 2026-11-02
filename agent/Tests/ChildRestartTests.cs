using System.Text.Json;
using System.Text.Json.Nodes;
using Ensemble;
using Application = Ensemble.Ensemble;

internal static partial class ActorTests
{
    public static async Task ChildRestart(string temp)
    {
        string home = Path.Combine(temp, "child-restart");
        Directory.CreateDirectory(home);
        string skillDirectory = Path.Combine(home, "skills");
        Directory.CreateDirectory(Path.Combine(skillDirectory, "base"));
        Directory.CreateDirectory(Path.Combine(skillDirectory, "goals"));
        File.WriteAllText(Path.Combine(skillDirectory, "base", "SKILL.md"),
            "---\nname: base\ntype: primary\ndescription: Restart fixture\ntools: read_file kill_agent\nloadable-skills: goals\n---\nInspect before acting.\n$TOOLS\n$SKILLS");
        File.Copy(Path.Combine(Directory.GetCurrentDirectory(), "agent/skills/goals/SKILL.md"),
            Path.Combine(skillDirectory, "goals", "SKILL.md"));
        string goalFile = Path.Combine(home, "goal.md");
        File.WriteAllText(goalFile, "Inspect saved results before continuing. Do not replay calls.");
        using var deadline = new CancellationTokenSource(TimeSpan.FromSeconds(30));
        var configuration = new Configuration(Vendor.Anthropic, "restart-fixture",
            "http://127.0.0.1:1", "local", sandboxRoot: home);
        var app = new Application(TextWriter.Null);
        try
        {
            var parent = new Agent(app, configuration, skillSpec: new SkillSpec(skillDirectory, "base"), dataDirectory: home);
            parent.StartPersistence(new PersistenceOptions(Path.Combine(home, "save.json")));
            Agent paused = await parent.CreateChildAsync("paused", cancellation: deadline.Token);
            await paused.PostAsync(new AgentInput(InputKind.Pause));
            _ = await paused.CreateChildAsync("nested", cancellation: deadline.Token);
            var load = await paused.DispatchAsync(new ToolCallPart("load-goal", configuration.Target(),
                "load_skill", JsonSerializer.SerializeToElement(new { name = "goals" })), deadline.Token);
            Check(!load.IsError, "Fixture could not load current goals skill.");
            await paused.GoalControlAsync(new GoalUpdate(goalFile, true, 180), deadline.Token);
            paused.History().Append(new Event(EventKind.MessageReceived, actor: Actor.Human,
                parts: [new TextPart("Retain this conversation, but do not replay its tools.")]));
            Agent complete = await parent.CreateChildAsync("complete", cancellation: deadline.Token);
            complete.PublishSubmission("Finished before restart.");
            Agent renewed = await parent.CreateChildAsync("renewed", cancellation: deadline.Token);
            renewed.PublishSubmission("An old answer, not completion of the next task.");
            await renewed.PostAsync(new AgentInput(InputKind.Pause));
            Task<string> pending = renewed.SubmitAsync([new TextPart("New work remains unfinished.")]);
            Check(!(await renewed.WorkStatusAsync(deadline.Token)).Submitted,
                "New work did not invalidate old completion.");
            Agent removed = await parent.CreateChildAsync("removed", cancellation: deadline.Token);
            removed.PublishSubmission("Retain this historical answer after removal.");
            ToolResultPart removal = await parent.DispatchAsync(new ToolCallPart("remove", configuration.Target(),
                "kill_agent", JsonSerializer.SerializeToElement(new { agent_id = "removed" })), deadline.Token);
            if (removal.Job is { Status: "running" } running)
            {
                removal = await parent.Jobs().Find(running.Handle).CompletionAsync(deadline.Token);
            }
            Check(!removal.IsError && parent.FindChild("removed") is null,
                "Explicit removal did not finish.");
            // Temporary helpers have History but no ordinary-child registration.
            var helper = new Agent(app, parent, "compress-fixture", configuration);
            helper.StartPersistence(new PersistenceOptions(Path.Combine(helper.DataDirectory(), "save.json")));
            await parent.ShutdownAsync();
            try { await pending; }
            catch (ObjectDisposedException) { }
        }
        finally { app.Destroy(); }

        var restarted = new Application(TextWriter.Null);
        try
        {
            var parent = new Agent(restarted, configuration, skillSpec: new SkillSpec(skillDirectory, "base"), dataDirectory: home);
            parent.StartPersistence(new PersistenceOptions(Path.Combine(home, "save.json")));
            await parent.RestoreChildrenAsync(deadline.Token);
            Check(parent.Children().Length == 3 && parent.FindChild("compress-fixture") is null
                && parent.FindChild("removed") is null,
                "Restart restored temporary workers or lost ordinary children.");
            Agent paused = parent.FindChild("paused")!;
            Check((await paused.WorkStatusAsync(deadline.Token)).Paused,
                "Restart discarded an explicit pause.");
            Check(JsonCodec.Text(paused.History().View().Dialogue.SelectMany(item => item.Parts))
                .Contains("Retain this conversation"), "Restart lost child History.");
            Check(paused.FindChild("nested")?.AgentParent() == paused,
                "Restart lost nested ownership.");
            GoalStatus goal = await paused.GoalStatusAsync(deadline.Token);
            Check(goal.Enabled && goal.IdleSeconds == 180 && goal.File == goalFile
                && goal.Phase == "paused" && goal.DueUtc is null,
                "Restart lost enabled goal configuration or woke a paused child.");
            string statePath = Path.Combine(paused.DataDirectory(), "child.json");
            byte[] priorState = File.ReadAllBytes(statePath);
            File.Delete(statePath);
            Directory.CreateDirectory(statePath);
            try
            {
                bool refused = false;
                try { await paused.GoalControlAsync(new GoalUpdate(IdleSeconds: 181), deadline.Token); }
                catch (Exception error) when (error is IOException or UnauthorizedAccessException) { refused = true; }
                Check(refused && (await paused.GoalStatusAsync(deadline.Token)).IdleSeconds == 180,
                    "Failed durable update changed the live goal settings.");
            }
            finally
            {
                Directory.Delete(statePath);
                File.WriteAllBytes(statePath, priorState);
            }
            Agent complete = parent.FindChild("complete")!;
            Check(complete.Submission() is not null && (await complete.WorkStatusAsync(deadline.Token)).Submitted,
                "Restart lost the current submitted state.");
            Check((await complete.GoalStatusAsync(deadline.Token)).Enabled == false,
                "Restart resumed completed work.");
            Agent renewed = parent.FindChild("renewed")!;
            Check(renewed.Submission() is null && (await renewed.WorkStatusAsync(deadline.Token)).Paused,
                "Restart turned a historical answer into current completion or resumed paused work.");
            Check(parent.RetrieveSubmission("renewed").Payload.GetProperty("result").GetString()!
                .Contains("old answer"), "Historical retrieval lost the old artifact.");
            await parent.RestoreChildrenAsync(deadline.Token);
            Check(parent.Children().Length == 3, "Repeated restoration duplicated children.");
        }
        finally { restarted.Destroy(); }

        var restricted = new Application(TextWriter.Null);
        try
        {
            var parent = new Agent(restricted, configuration, dataDirectory: home);
            await parent.RestoreChildrenAsync(deadline.Token);
            GoalStatus unavailable = await parent.FindChild("paused")!.GoalStatusAsync(deadline.Token);
            Check(!unavailable.Enabled && !unavailable.Available && unavailable.LastError.Length != 0,
                "Saved goal configuration restored stale skill authority.");
        }
        finally { restricted.Destroy(); }

        string registration = Path.Combine(home, "children", "paused", "child.json");
        byte[] original = File.ReadAllBytes(registration);
        foreach (string malformed in new[] { "{", "{}",
            System.Text.Encoding.UTF8.GetString(original).Replace("\"ID\":\"paused\"", "\"ID\":\"foreign\""),
            System.Text.Encoding.UTF8.GetString(original).Replace("\"Version\":1", "\"Version\":1,\"Version\":1") })
        {
            File.WriteAllText(registration, malformed);
            var refusedApp = new Application(TextWriter.Null);
            try
            {
                var parent = new Agent(refusedApp, configuration, dataDirectory: home);
                try
                {
                    await parent.RestoreChildrenAsync(deadline.Token);
                    throw new Exception("Malformed child registration was accepted.");
                }
                catch (InvalidDataException) { }
                Check(parent.Children().Length == 0, "Refused restore left a partial child tree.");
            }
            finally { refusedApp.Destroy(); }
        }
        File.WriteAllBytes(registration, original);
        string answerPath = Path.Combine(home, "children", "complete", "submit.md");
        string acceptedAnswer = File.ReadAllText(answerPath);
        File.WriteAllText(answerPath, acceptedAnswer.Replace("Finished before restart.", "Unacknowledged replacement."));
        var mismatch = new Application(TextWriter.Null);
        try
        {
            var parent = new Agent(mismatch, configuration, dataDirectory: home);
            bool refused = false;
            try { await parent.RestoreChildrenAsync(deadline.Token); }
            catch (InvalidDataException) { refused = true; }
            Check(refused && parent.Children().Length == 0,
                "Restart accepted an artifact not acknowledged by its completion record.");
        }
        finally
        {
            mismatch.Destroy();
            File.WriteAllText(answerPath, acceptedAnswer);
        }
    }
}
