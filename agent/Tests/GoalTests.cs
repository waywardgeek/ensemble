using System.Net;
using System.Text.Json;
using System.Text.Json.Nodes;
using Ensemble;
using Application = Ensemble.Ensemble;

internal static partial class MeterTests
{
    private static ToolCallPart GoalCall(string name, object args)
    {
        return new ToolCallPart(Guid.NewGuid().ToString("N"),
            new Provenance(Vendor.OpenAI, "goal-test", Surface.ChatCompletions),
            name, JsonSerializer.SerializeToElement(args));
    }

    private static string GoalSkills(string temp)
    {
        string directory = Path.Combine(temp, "goal-skills-" + Guid.NewGuid().ToString("N"));
        Directory.CreateDirectory(Path.Combine(directory, "base"));
        Directory.CreateDirectory(Path.Combine(directory, "goals"));
        File.WriteAllText(Path.Combine(directory, "base", "SKILL.md"),
            "---\nname: base\ntype: primary\ndescription: Goal fixture\ntools: read_file micro_handoff\nloadable-skills: goals\n---\nInspect before acting.\n$TOOLS\n$SKILLS");
        string source = Path.Combine(Directory.GetCurrentDirectory(), "agent/Skills/goals/SKILL.md");
        File.Copy(source, Path.Combine(directory, "goals", "SKILL.md"));
        return directory;
    }

    private static Agent GoalAgent(Application app, string skills, string url = "http://127.0.0.1:1",
        Vendor vendor = Vendor.OpenAI, Surface surface = Surface.ChatCompletions)
    {
        return new Agent(app, new Configuration(vendor, "goal-test", url,
            surface: surface, disableStreaming: true), skillSpec: new SkillSpec(skills, "base"));
    }

    private static async Task LoadGoals(Agent agent)
    {
        var result = await agent.DispatchAsync(GoalCall("load_skill", new { name = "goals" }));
        Check(!result.IsError && agent.ToolEnabled("goal_control"), "Goals skill did not grant its tool");
    }

    private static async Task GoalRefuses(Func<Task<GoalStatus>> operation)
    {
        try { await operation(); }
        catch (Exception error) when (error is ArgumentException or IOException or InvalidOperationException)
        {
            return;
        }
        throw new Exception("Goal update unexpectedly succeeded");
    }

    public static async Task GoalsConfiguration(string temp)
    {
        var app = new Application(TextWriter.Null);
        try
        {
            var legacy = new Agent(app, new Configuration(Vendor.OpenAI, "goal-test", "http://127.0.0.1:1"));
            legacy.InstallTools();
            Check(!legacy.ToolEnabled("goal_control"), "Legacy fallback granted Goals without a skill");
            Check(!legacy.Engine().Render(new Context()).ToJsonString().Contains("goal_control"),
                "Legacy fallback declared Goals without a skill");
            await GoalRefuses(() => legacy.GoalControlAsync());
            Check((await legacy.DispatchAsync(GoalCall("goal_control", new { }))).IsError,
                "Legacy actor dispatched Goals without a skill");
            var agent = GoalAgent(app, GoalSkills(temp));
            Check(!agent.ToolEnabled("goal_control"), "Default actor has goal authority");
            await GoalRefuses(() => agent.GoalControlAsync());
            await LoadGoals(agent);
            GoalStatus initial = await agent.GoalControlAsync();
            Check(!initial.Enabled && initial.File == "" && initial.IdleSeconds == 480
                && initial.Phase == "unconfigured", "Loading configured reminders");
            string file = Path.Combine(temp, "goal-config.md");
            File.WriteAllText(file, "Keep the objective but do not deploy.");
            var selected = await agent.GoalControlAsync(new GoalUpdate(File: file));
            Check(!selected.Enabled && selected.File == file, "Selecting a file enabled reminders");
            foreach (double invalid in new[] { 0d, -1d, double.NaN, double.PositiveInfinity, .00001, 1e10 })
            {
                await GoalRefuses(() => agent.GoalControlAsync(new GoalUpdate(IdleSeconds: invalid)));
                Check(await agent.GoalControlAsync() == selected, "Invalid delay changed configuration");
            }
            await GoalRefuses(() => agent.GoalControlAsync(new GoalUpdate(File: file + ".missing")));
            Check(await agent.GoalControlAsync() == selected, "Invalid file changed configuration");
            var queried = await agent.DispatchAsync(GoalCall("goal_control", new { }));
            Check(!queried.IsError && queried.Parts.OfType<TextPart>().Single().Text.Contains("idle_seconds"),
                "Tool query failed or used CLR property names");
            var nullFile = await agent.DispatchAsync(GoalCall("goal_control", new { file = (string?)null }));
            Check(nullFile.IsError, "Explicit null file was treated as omission");
            Check(await agent.GoalControlAsync() == selected, "Malformed tool call changed state");
            File.Delete(file);
            Check(!(await agent.GoalControlAsync(new GoalUpdate(Enabled: false))).Enabled,
                "Off depended on a readable file");
            var cleared = await agent.GoalControlAsync(new GoalUpdate(File: ""));
            Check(cleared.File == "" && !cleared.Enabled, "Clear did not remove goal state");
        }
        finally { app.Destroy(); }
    }

    public static async Task GoalsProjection(string temp)
    {
        string skills = GoalSkills(temp);
        foreach (var (vendor, surface) in new[] { (Vendor.OpenAI, Surface.ChatCompletions),
            (Vendor.OpenAI, Surface.Responses), (Vendor.Anthropic, Surface.Messages),
            (Vendor.Gemini, Surface.GenerateContent) })
        {
            var app = new Application(TextWriter.Null);
            try
            {
                var agent = GoalAgent(app, skills, vendor: vendor, surface: surface);
                await LoadGoals(agent);
                string file = Path.Combine(temp, "goal-projection.md");
                File.WriteAllText(file, "OBJECTIVE-ONE retain uncertainty");
                await agent.GoalControlAsync(new GoalUpdate(File: file));
                long before = agent.History().Snapshot().Sequence;
                string first = agent.Engine().Render(agent.History().View()).ToJsonString();
                var root = agent.Engine().Render(agent.History().View());
                string system = surface == Surface.Responses ? root["instructions"]!.ToJsonString()
                    : vendor == Vendor.Anthropic ? root["system"]!.ToJsonString()
                    : vendor == Vendor.Gemini ? root["systemInstruction"]!.ToJsonString()
                    : root["messages"]![0]!["content"]!.ToJsonString();
                Check(system.Contains("OBJECTIVE-ONE"), "Goal was not in provider instruction field: " + surface);
                File.WriteAllText(file, "OBJECTIVE-TWO new evidence");
                string second = agent.Engine().Render(agent.History().View()).ToJsonString();
                Check(second.Contains("OBJECTIVE-TWO") && !second.Contains("OBJECTIVE-ONE"), "Goal text cached");
                Check(agent.History().Snapshot().Sequence == before, "Goal render mutated History");
                var manager = agent.SkillManager()!;
                agent.RemoveSkills(manager);
                try
                {
                    Check(!agent.Engine().Render(agent.History().View()).ToJsonString().Contains("OBJECTIVE-TWO"),
                        "Detached Skills left goal injection active");
                }
                finally { agent.AddSkills(manager); }
                Check((await agent.GoalControlAsync()).File == "", "Detachment retained goal configuration");
                await agent.GoalControlAsync(new GoalUpdate(File: file));
                File.WriteAllBytes(file, [0xff, 0xfe, 0xff]);
                string warning = agent.Engine().Render(agent.History().View()).ToJsonString();
                Check(warning.Contains("Reminders are off"), "Human request lost goal read error");
                var failed = await agent.GoalControlAsync();
                Check(!failed.Enabled && failed.LastError.Length > 0, "Read failure did not stop reminders");
                File.WriteAllText(file, "OBJECTIVE-THREE restored");
                Check(agent.Engine().Render(agent.History().View()).ToJsonString().Contains("OBJECTIVE-THREE"),
                    "Off goal was not read after repair");
                var unloaded = await agent.DispatchAsync(GoalCall("unload_skill", new { name = "goals" }));
                Check(!unloaded.IsError && !agent.ToolEnabled("goal_control"), "Unload left goal authority");
                await agent.WorkStatusAsync();
                Check(!agent.Engine().Render(agent.History().View()).ToJsonString().Contains("OBJECTIVE-THREE"),
                    "Pending manual retained goal injection");
                Check(File.Exists(file), "Unload deleted the goal file");
            }
            finally { app.Destroy(); }
        }
    }

    public static async Task GoalsTimer(string temp)
    {
        using var listener = Listen();
        var app = new Application(TextWriter.Null);
        try
        {
            var agent = GoalAgent(app, GoalSkills(temp), listener.Prefixes.Single());
            await LoadGoals(agent);
            string file = Path.Combine(temp, "goal-timer.md");
            File.WriteAllText(file, "GOAL-BEFORE-IDLE");
            var pendingHttp = listener.GetContextAsync();
            await agent.PostAsync(new AgentInput(InputKind.Pause));
            await agent.GoalControlAsync(new GoalUpdate(File: file, Enabled: true, IdleSeconds: .15));
            Check((await agent.GoalControlAsync()).Phase == "paused", "Pause did not suspend timer");
            await Task.Delay(300);
            Check(!pendingHttp.IsCompleted, "Paused timer sent a request");
            File.WriteAllText(file, "GOAL-AT-WAKE do not repeat commands blindly");
            await agent.PostAsync(new AgentInput(InputKind.Resume));
            var scheduled = await agent.GoalControlAsync();
            Check(scheduled.Phase == "scheduled" && scheduled.DueUtc.HasValue, "Resume did not arm timer");
            Check((await agent.GoalControlAsync()).DueUtc == scheduled.DueUtc, "Status reset idle delay");
            var request = await pendingHttp.WaitAsync(TimeSpan.FromSeconds(5));
            string body = (await ReadRecallRequest(request)).ToJsonString();
            Check(body.Contains("GOAL-AT-WAKE") && !body.Contains("GOAL-BEFORE-IDLE")
                && body.Contains("Automatic goal reminder"), "Reminder used stale goal or omitted its origin");
            var secondHttp = listener.GetContextAsync();
            await Task.Delay(300);
            Check(!secondHttp.IsCompleted, "Reminder overlapped a busy request");
            await agent.GoalControlAsync(new GoalUpdate(Enabled: false));
            var workingOff = await agent.GoalStatusAsync();
            Check(!workingOff.Enabled && workingOff.Phase == "busy", "Off hid an active request");
            await Reply(request, "{\"choices\":[{\"message\":{\"role\":\"assistant\",\"content\":\"checked\"},\"finish_reason\":\"stop\"}]}");
            await agent.FlushAsync().WaitAsync(TimeSpan.FromSeconds(5));
            await Task.Delay(300);
            Check(!secondHttp.IsCompleted, "Off rearmed after the current reminder completed");
            Check((await agent.GoalStatusAsync()).Phase == "off", "Completed request still reported busy");
            await agent.GoalControlAsync(new GoalUpdate(Enabled: true, IdleSeconds: 30));
            await agent.PostAsync(new AgentInput(InputKind.InterruptNow));
            Check(!(await agent.GoalControlAsync()).Enabled, "Idle interrupt did not disable reminders");
            await agent.ShutdownAsync();
        }
        finally { app.Destroy(); }
    }

    public static async Task GoalsSubmission(string temp)
    {
        using var listener = Listen();
        var app = new Application(TextWriter.Null);
        try
        {
            string skills = GoalSkills(temp);
            var parent = GoalAgent(app, skills, listener.Prefixes.Single());
            await LoadGoals(parent);
            string file = Path.Combine(temp, "goal-submit.md");
            File.WriteAllText(file, "Finish the delegated task without repeating completed work.");
            await parent.GoalControlAsync(new GoalUpdate(file, true, 600));
            string home = Directory.CreateDirectory(Path.Combine(temp, "goal-submit-child")).FullName;
            var child = new Agent(app, parent, "goal-child", parent.Configuration(),
                skillSpec: new SkillSpec(skills, "base"), dataDirectory: home);
            await LoadGoals(child);
            Check((await child.GoalStatusAsync()).File == "", "Child inherited parent's goal");
            await child.GoalControlAsync(new GoalUpdate(file, true, .15));
            try
            {
                child.PublishSubmission(JsonSerializer.SerializeToElement(17));
                throw new Exception("Invalid submission accepted");
            }
            catch (ArgumentException) { }
            Check((await child.GoalStatusAsync()).Enabled, "Refused submission disabled goal");
            var request = listener.GetContextAsync();
            child.PublishSubmission("Completed the task.");
            var submitted = await child.GoalStatusAsync();
            Check(!submitted.Enabled && submitted.File == file && submitted.DueUtc is null,
                "Successful submission retained an automatic wake");
            Check((await parent.GoalStatusAsync()).Enabled, "Child submission changed parent's goal");
            await Task.Delay(350);
            Check(!request.IsCompleted && child.Submission().HasValue,
                "Reminder invalidated a successful submission");
            Check(File.Exists(Path.Combine(home, "submit.md")), "Submission was not durable");
        }
        finally { app.Destroy(); }
    }

    public static async Task GoalsMissingWake(string temp)
    {
        using var listener = Listen();
        var app = new Application(TextWriter.Null);
        try
        {
            var agent = GoalAgent(app, GoalSkills(temp), listener.Prefixes.Single());
            await LoadGoals(agent);
            string file = Path.Combine(temp, "goal-missing.md");
            File.WriteAllText(file, "Delete before wake");
            var pendingHttp = listener.GetContextAsync();
            await agent.GoalControlAsync(new GoalUpdate(File: file, Enabled: true, IdleSeconds: .15));
            File.Delete(file);
            using var deadline = new CancellationTokenSource(TimeSpan.FromSeconds(5));
            GoalStatus status;
            do
            {
                await Task.Delay(20, deadline.Token);
                status = await agent.GoalControlAsync(cancellation: deadline.Token);
            } while (status.Enabled);
            Check(status.LastError.Length > 0 && !pendingHttp.IsCompleted, "Unreadable goal reached provider");
            Check(!agent.History().Events().Any(e => e.Kind == EventKind.RequestSent),
                "Unreadable automatic goal claimed a request");
        }
        finally { app.Destroy(); }
    }

    public static async Task GoalsStaleTimer(string temp)
    {
        using var listener = Listen();
        var app = new Application(TextWriter.Null);
        try
        {
            string skills = GoalSkills(temp);
            var agent = GoalAgent(app, skills, listener.Prefixes.Single());
            await LoadGoals(agent);
            string file = Path.Combine(temp, "goal-stale.md");
            File.WriteAllText(file, "Wait for an authorized opportunity.");
            var http = listener.GetContextAsync();
            await agent.GoalControlAsync(new GoalUpdate(File: file, Enabled: true, IdleSeconds: .15));
            await agent.GoalControlAsync(new GoalUpdate(IdleSeconds: 30));
            await Task.Delay(300);
            Check(!http.IsCompleted && (await agent.GoalControlAsync()).Enabled,
                "Reconfiguration left an old wake active");
            await agent.DispatchAsync(GoalCall("unload_skill", new { name = "goals" }));
            await GoalRefuses(() => agent.GoalControlAsync());
            await Task.Delay(300);
            Check(!http.IsCompleted, "Unload left a live reminder");
            await agent.ShutdownAsync();
            var replacement = GoalAgent(app, skills, listener.Prefixes.Single());
            await LoadGoals(replacement);
            var fresh = await replacement.GoalControlAsync();
            Check(fresh.File == "" && !fresh.Enabled && fresh.IdleSeconds == 480,
                "New Agent inherited goal configuration");
            await replacement.GoalControlAsync(new GoalUpdate(File: file, Enabled: true, IdleSeconds: .15));
            await replacement.ShutdownAsync();
            await Task.Delay(300);
            Check(!http.IsCompleted, "Shutdown left a live reminder");
        }
        finally { app.Destroy(); }
    }

    public static async Task GoalsHandoff(string temp)
    {
        using var listener = Listen();
        var app = new Application(TextWriter.Null);
        try
        {
            var agent = GoalAgent(app, GoalSkills(temp), listener.Prefixes.Single());
            await LoadGoals(agent);
            string file = Path.Combine(temp, "goal-handoff.md");
            File.WriteAllText(file, "FRESH-GOAL-FIRST private objective marker");
            await agent.GoalControlAsync(new GoalUpdate(File: file));
            var turn = agent.SubmitAsync([new TextPart("Continue the authorized local work.")]);
            var first = await listener.GetContextAsync().WaitAsync(TimeSpan.FromSeconds(5));
            Check((await ReadRecallRequest(first)).ToJsonString().Contains("FRESH-GOAL-FIRST"),
                "Initial request omitted goal");
            File.WriteAllText(file, "FRESH-GOAL-SECOND keep reasons and uncertainty");
            string response = JsonSerializer.Serialize(new
            {
                choices = new[] { new { message = new { role = "assistant", content = "I will preserve the next step before continuing.",
                    tool_calls = new[] { new { id = "goal-handoff", type = "function", function = new { name = "micro_handoff",
                        arguments = JsonSerializer.Serialize(new { text = "Continue after checking the recorded results." }) } } } }, finish_reason = "tool_calls" } }
            });
            await Reply(first, response);
            var second = await listener.GetContextAsync().WaitAsync(TimeSpan.FromSeconds(5));
            string wire = (await ReadRecallRequest(second)).ToJsonString();
            Check(wire.Contains("FRESH-GOAL-SECOND") && !wire.Contains("FRESH-GOAL-FIRST"),
                "Same-turn request cached goal across tool/handoff");
            await Reply(second, "{\"choices\":[{\"message\":{\"role\":\"assistant\",\"content\":\"done\"},\"finish_reason\":\"stop\"}]}");
            Check(await turn.WaitAsync(TimeSpan.FromSeconds(5)) == "done", "Handoff turn failed");
            Check(!agent.History().Events().SelectMany(e => e.Parts).OfType<TextPart>()
                .Any(p => p.Text.Contains("FRESH-GOAL-")), "Goal text entered canonical dialogue");
        }
        finally { app.Destroy(); }
    }
}
