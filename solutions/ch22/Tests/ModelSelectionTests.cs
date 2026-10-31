using System.Text.Json;
using Ensemble;
using Application = Ensemble.Ensemble;

internal static partial class MeterTests
{
    private static async Task BusyModel(Agent agent, string model)
    {
        Configuration before = agent.Configuration();
        try
        {
            await agent.SetModelAsync(model).WaitAsync(TimeSpan.FromSeconds(10));
        }
        catch (InvalidOperationException)
        {
            Check(ReferenceEquals(before, agent.Configuration()), "busy selection changes nothing");
            return;
        }
        throw new Exception("Expected a busy model-selection refusal");
    }

    public static async Task ModelSelection(string directory)
    {
        using var listener = Listen();
        var app = new Application(TextWriter.Null);
        try
        {
            var agent = new Agent(app, new Configuration(Vendor.OpenAI, "gpt-4o-mini", listener.Prefixes.Single(),
                "dummy", resolvedModel: "old-resolved-identity", disableStreaming: true));
            Engine engine = agent.Engine();
            var lens = new CacheLens(engine, Path.Combine(directory, "selection-lens"));
            async Task Turn(string prompt, string model, string resolved, string? prior = null)
            {
                var turn = agent.SubmitAsync([new TextPart(prompt)]);
                var request = await listener.GetContextAsync().WaitAsync(TimeSpan.FromSeconds(10));
                using var reader = new StreamReader(request.Request.InputStream);
                string wire = await reader.ReadToEndAsync();
                using var json = JsonDocument.Parse(wire);
                Check(json.RootElement.GetProperty("model").GetString() == model, "new selection reaches actual wire");
                Check(prior is null || wire.Contains(prior), "switch preserves previous conversation");
                await BusyModel(agent, "must-not-change");
                await Reply(request, JsonSerializer.Serialize(new
                {
                    model = resolved, service_tier = "default",
                    choices = new[] { new { message = new { content = "answer-" + prompt } } },
                    usage = new { prompt_tokens = 120, prompt_tokens_details = new { cached_tokens = 20 }, completion_tokens = 10 }
                }));
                await turn.WaitAsync(TimeSpan.FromSeconds(10));
            }
            await Turn("first", "gpt-4o-mini", "gpt-4o-mini-2024-07-18");
            MeterSnapshot first = engine.Meter();
            await agent.SetModelAsync("gpt-4o");
            Configuration second = agent.Configuration();
            Check(second.Model == "gpt-4o" && second.ResolvedModel == "gpt-4o"
                && second.ApiKey == "dummy" && second.BaseUrl == listener.Prefixes.Single().TrimEnd('/')
                && second.Vendor == Vendor.OpenAI && second.Surface == Surface.ChatCompletions,
                "replace selection, not route or credentials; clear stale resolved identity");
            MeterSnapshot selected = engine.Meter();
            Check(ReferenceEquals(engine, agent.Engine()) && ReferenceEquals(lens, engine.Lens()),
                "selection preserves Engine and lens ownership");
            Check(selected.Totals.Input == first.Totals.Input && selected.Totals.Output == first.Totals.Output
                && selected.Totals.CacheRead == first.Totals.CacheRead && selected.Totals.CacheWrite == first.Totals.CacheWrite
                && selected.EstimatedUsd() == first.EstimatedUsd(),
                "selection alone preserves counters and accumulated price");
            await Turn("second", "gpt-4o", "gpt-4o-2024-08-06", "answer-first");
            Check(engine.Meter().Rows.Length == 2 && engine.Meter().EstimatedUsd() == .0003975m,
                "mixed known models retain their own pricing rows");
            await agent.SetModelAsync("unpriced-local-fixture");
            await Turn("third", "unpriced-local-fixture", "unpriced-local-fixture", "answer-second");
            await agent.SetModelAsync("gpt-4o-mini");
            Check(engine.Meter().EstimatedUsd() is null && engine.Meter().Totals.Input == 300,
                "returning to priced selection never prices earlier unknown work");
            int events = agent.History().Events().Length;
            Configuration current = agent.Configuration();
            foreach (string invalid in new[] { "", "   ", "gpt-6-astra" })
            {
                try
                {
                    await agent.SetModelAsync(invalid);
                    throw new Exception("Expected invalid/incompatible selection refusal");
                }
                catch (ArgumentException) { }
                Check(ReferenceEquals(current, agent.Configuration()) && events == agent.History().Events().Length,
                    "invalid model never changes configuration or conversation");
            }
            await agent.PostAsync(new AgentInput(InputKind.Pause));
            var queued = agent.SubmitAsync([new TextPart("queued")]);
            await BusyModel(agent, "gpt-4o");
            await agent.PostAsync(new AgentInput(InputKind.Resume));
            var resumed = await listener.GetContextAsync().WaitAsync(TimeSpan.FromSeconds(10));
            await Reply(resumed, """{"choices":[{"message":{"content":"resumed"}}],"usage":{"prompt_tokens":10,"completion_tokens":1}}""");
            await queued.WaitAsync(TimeSpan.FromSeconds(10));
            await agent.SetModelAsync("gpt-4o");
            await agent.ShutdownAsync();
            try
            {
                await agent.SetModelAsync("gpt-4o-mini");
                throw new Exception("Expected closed Agent refusal");
            }
            catch (ObjectDisposedException) { }
        }
        finally
        {
            app.Destroy();
        }
    }

    public static async Task DirectSendSelection()
    {
        using var listener = Listen();
        var app = new Application(TextWriter.Null);
        try
        {
            var agent = new Agent(app, new Configuration(Vendor.OpenAI, "gpt-4o", listener.Prefixes.Single()));
            var send = agent.Engine().SendAsync("{}");
            var request = await listener.GetContextAsync().WaitAsync(TimeSpan.FromSeconds(10));
            var overlapping = agent.Engine().SendAsync("{}");
            var secondRequest = await listener.GetContextAsync().WaitAsync(TimeSpan.FromSeconds(10));
            await BusyModel(agent, "gpt-4o-mini");
            await Reply(request, "{}", 503);
            await send;
            await BusyModel(agent, "gpt-4o-mini");
            await Reply(secondRequest, "{}", 503);
            await overlapping;
            await agent.SetModelAsync("gpt-4o-mini");
            using var cancellation = new CancellationTokenSource();
            send = agent.Engine().SendAsync("{}", cancellation.Token);
            var held = await listener.GetContextAsync().WaitAsync(TimeSpan.FromSeconds(10));
            await BusyModel(agent, "gpt-4o");
            cancellation.Cancel();
            try { await send; } catch (OperationCanceledException) { }
            held.Response.Abort();
            await agent.SetModelAsync("gpt-4o");
            Check(agent.Configuration().Model == "gpt-4o", "failed/cancelled direct sends release selection guard");
            Check(Engine.ModelChoices(agent.Configuration()).SequenceEqual(new[] { "gpt-4o", "gpt-4o-mini" }),
                "chat choices share provider and exclude Responses-only row without duplicates");
            var arbitrary = agent.Configuration().WithModel("unknown-current");
            Check(Engine.ModelChoices(arbitrary).SequenceEqual(new[] { "unknown-current", "gpt-4o", "gpt-4o-mini" }),
                "unknown current selection stays visible without inventing capabilities");
            Check(!Engine.LookupModel(Vendor.Anthropic, "gpt-4o", out var absent) && absent.Media == Media.None,
                "capability keys include vendor");
            foreach (string model in new[] { "gpt-4o", "gpt-4o-mini" })
            {
                Check(Engine.LookupModel(Vendor.OpenAI, model, out var features)
                    && features.Media == Media.Image && features.Stream && !features.ResponsesOnly,
                    "existing OpenAI media/stream row retained");
            }
            Check(Engine.LookupModel(Vendor.Anthropic, "claude-3-haiku-20240307", out var anthropic)
                && anthropic.Media == Media.Image && anthropic.Stream, "Anthropic row retained");
            Check(Engine.LookupModel(Vendor.Gemini, "gemini-2.5-flash", out var gemini)
                && gemini.Media == (Media.Image | Media.Audio | Media.Video | Media.Document)
                && gemini.Stream, "Gemini row retained");
            var reasoning = new Configuration(Vendor.OpenAI, "gpt-6-astra", listener.Prefixes.Single(),
                reasoningEffort: "high", reasoningSummary: "off");
            Check(Engine.ModelChoices(reasoning).SequenceEqual(new[] { "gpt-6-astra", "gpt-4o", "gpt-4o-mini" }),
                "Responses choices derive from the same selection validator");
            var unknown = reasoning.WithModel("unpriced-responses-model");
            Check(unknown.Surface == Surface.Responses && unknown.ReasoningEffort == ""
                && unknown.ReasoningSummary == "" && unknown.RequireToolExplanation,
                "new capabilities reset reasoning fields without loosening explanation policy");
            var back = unknown.WithModel("gpt-6-astra");
            Check(back.ReasoningEffort == "medium" && back.ReasoningSummary == "auto", "selected model receives valid fresh defaults");
            var plan = new Configuration(Vendor.OpenAI, "gpt-6-astra", listener.Prefixes.Single(),
                surface: Surface.Responses, billing: BillingMode.ChatGptPlan, account: "fixture").WithModel("unknown-plan-model");
            Check(plan.Billing == BillingMode.ChatGptPlan && plan.Account == "fixture" && plan.ApiKey == "",
                "model selection cannot switch funding or account");
        }
        finally
        {
            app.Destroy();
        }
    }
}
