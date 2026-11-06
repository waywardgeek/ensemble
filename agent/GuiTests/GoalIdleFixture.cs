using System.Net;
using System.Net.Sockets;
using System.Text.Json;
using Ensemble;
using EnsembleGui;

// Real GUI/Agent with an owned provider. No live credentials or private APIs.
internal static class GoalIdleFixture
{
    internal static async Task RunAsync()
    {
        string work = Directory.GetCurrentDirectory();
        string skills = Path.Combine(work, "skills");
        Directory.CreateDirectory(Path.Combine(skills, "base"));
        Directory.CreateDirectory(Path.Combine(skills, "goals"));
        File.WriteAllText(Path.Combine(skills, "base", "SKILL.md"),
            "---\nname: base\ntype: primary\ndescription: Owned goal fixture\ntools: read_file\nloadable-skills: goals\n---\n$TOOLS\n$SKILLS");
        File.Copy(Environment.GetEnvironmentVariable("GOAL_SKILL_SOURCE")
            ?? throw new InvalidOperationException("GOAL_SKILL_SOURCE required"),
            Path.Combine(skills, "goals", "SKILL.md"));
        string goal = Path.Combine(work, "GOAL.md");
        File.WriteAllText(goal, "Inspect recorded work. This owned fixture grants no external actions.");
        var builder = WebApplication.CreateBuilder();
        builder.Logging.ClearProviders();
        var provider = builder.Build();
        provider.Urls.Add("http://127.0.0.1:0");
        int requests = 0;
        provider.MapPost("/v1/chat/completions", async (HttpContext context) =>
        {
            using var body = await JsonDocument.ParseAsync(context.Request.Body,
                cancellationToken: context.RequestAborted);
            int count = Interlocked.Increment(ref requests);
            if (count <= 2)
            {
                context.Response.StatusCode = 503;
                await context.Response.WriteAsJsonAsync(new { error = new { message = "owned transient failure" } }, context.RequestAborted);
                return;
            }
            // Give the browser time to turn continuation off while real work is active.
            await Task.Delay(1000, context.RequestAborted);
            await context.Response.WriteAsJsonAsync(new { choices = new[] {
                new { message = new { role = "assistant", content = "Owned retry completed." }, finish_reason = "stop" }
            } }, context.RequestAborted);
        });
        using var stop = new CancellationTokenSource(TimeSpan.FromMinutes(2));
        var root = new Ensemble.Ensemble(TextWriter.Null, dataDirectory: Path.Combine(work, "data"));
        Task? serving = null;
        try
        {
            await provider.StartAsync(stop.Token);
            var config = new Configuration(Vendor.OpenAI, "goal-fixture", provider.Urls.Single(), "owned-fake",
                surface: Surface.ChatCompletions, disableStreaming: true);
            var agent = new Agent(root, config, name: "goal-fixture", skillSpec: new SkillSpec(skills, "base"));
            var loaded = await agent.DispatchAsync(new ToolCallPart("load-goals", config.Target(),
                "load_skill", JsonSerializer.SerializeToElement(new { name = "goals" })), stop.Token);
            if (loaded.IsError) throw new InvalidOperationException("Fixture skill load failed");
            using var reservation = new TcpListener(IPAddress.Loopback, 0);
            reservation.Start();
            int port = ((IPEndPoint)reservation.LocalEndpoint).Port;
            reservation.Stop();
            serving = Gui.RunAsync(agent, new GuiOptions(port,
                SettingsPath: Path.Combine(work, "appearance.json")), stop.Token);
            File.WriteAllText(Path.Combine(work, "fixture.json"), JsonSerializer.Serialize(new {
                url = $"http://127.0.0.1:{port}", goal
            }));
            Console.WriteLine("READY");
            while (!File.Exists(Path.Combine(work, "stop")))
            {
                if (serving.IsCompleted) await serving;
                await Task.Delay(25, stop.Token);
            }
        }
        finally
        {
            stop.Cancel();
            try
            {
                if (serving is not null) await serving;
            }
            finally
            {
                try { root.Destroy(); }
                finally
                {
                    await provider.StopAsync();
                    await provider.DisposeAsync();
                }
            }
        }
        File.WriteAllText(Path.Combine(work, "result.json"), JsonSerializer.Serialize(new { requests, joined = true }));
        Console.WriteLine("STOPPED");
    }
}
