using System.Text.Json;
using System.Text.Json.Nodes;
using Ensemble;
using Application = Ensemble.Ensemble;

internal static partial class MeterTests
{
    private static async Task RefuseConfiguredSelection(Agent agent, Model model)
    {
        Model? previous = agent.SelectedModel();
        Configuration configuration = agent.Configuration();
        try
        {
            await agent.SetModelAsync(model).WaitAsync(TimeSpan.FromSeconds(10));
        }
        catch (Exception error) when (error is ArgumentException or InvalidOperationException or IOException
            or UnauthorizedAccessException or System.Text.Json.JsonException or FormatException or InvalidDataException)
        {
            Check(ReferenceEquals(previous, agent.SelectedModel())
                && ReferenceEquals(configuration, agent.Configuration()),
                "Refused configured selection preserves the relationship and configuration snapshot.");
            return;
        }
        throw new Exception("Expected configured selection refusal.");
    }

    public static async Task ConfiguredModelSelection()
    {
        using var first = Listen();
        using var second = Listen();
        using var deadline = new CancellationTokenSource(TimeSpan.FromSeconds(30));
        var app = new Application(TextWriter.Null);
        var foreign = new Application(TextWriter.Null);
        try
        {
            var vendor = new ModelVendor(app, "local", "Local");
            var none = new ModelAuthentication(app, "none");
            var key = new ModelAuthentication(app, "key", "selected-synthetic-key");
            var one = new Model(new ModelEndpoint(vendor, "one", first.Prefixes.Single(),
                Surface.ChatCompletions, none), "small", "same-wire", "One", new ModelFeatures(Media.None), 4096);
            var two = new Model(new ModelEndpoint(vendor, "two", second.Prefixes.Single(),
                Surface.ChatCompletions, key), "small", "same-wire", "Two", new ModelFeatures(Media.None), 8192);
            var agent = new Agent(app, new Configuration(Vendor.Anthropic, "legacy",
                "http://127.0.0.1:1", "old-synthetic-key", disableStreaming: true));
            Engine engine = agent.Engine();
            await agent.SetModelAsync(one);
            Check(ReferenceEquals(agent.SelectedModel(), one) && one.Agents().Contains(agent)
                && agent.Configuration().Billing == BillingMode.None,
                "Public selector attaches the actual Model and replaces legacy funding.");

            async Task Turn(System.Net.HttpListener listener, Model model, Model rejected,
                string prompt, string? authorization, string? previousAnswer)
            {
                Task<string> turn = agent.SubmitAsync([new TextPart(prompt)]);
                var request = await listener.GetContextAsync().WaitAsync(deadline.Token);
                try
                {
                    using var reader = new StreamReader(request.Request.InputStream);
                    string body = await reader.ReadToEndAsync(deadline.Token);
                    using var json = JsonDocument.Parse(body);
                    Check(json.RootElement.GetProperty("model").GetString() == model.Name
                        && request.Request.Headers["Authorization"] == authorization
                        && !body.Contains("old-synthetic-key"),
                        "Actual request uses the selected wire name and authentication only.");
                    Check(previousAnswer is null || body.Contains(previousAnswer),
                        "Offering switch preserves ordinary conversation.");
                    await RefuseConfiguredSelection(agent, rejected);
                    await Reply(request, JsonSerializer.Serialize(new
                    {
                        model = model.Name,
                        choices = new[] { new { message = new { content = "answer-" + prompt } } },
                        usage = new { prompt_tokens = 100, completion_tokens = 1 }
                    }));
                    await turn.WaitAsync(deadline.Token);
                }
                finally { request.Response.Abort(); }
                Check(agent.WorkingContextStatus().InputTokens == 100,
                    "Completed main request updates the selected offering's context measurement.");
            }

            await Turn(first, one, two, "first", null, null);
            await agent.SetModelAsync(two);
            Check(!one.Agents().Contains(agent) && two.Agents().Contains(agent)
                && ReferenceEquals(agent.Engine(), engine) && agent.WorkingContextStatus().InputTokens is null
                && agent.WorkingContextStatus().WorkingTokens == 8192,
                "Switch moves both membership ends, preserves Engine, and resets last-request context.");
            await Turn(second, two, one, "second", "Bearer selected-synthetic-key", "answer-first");

            // Direct sends bypass the turn loop but still hold the same selection guard.
            var send = agent.Engine().SendAsync("{}", deadline.Token);
            var held = await second.GetContextAsync().WaitAsync(deadline.Token);
            try
            {
                await RefuseConfiguredSelection(agent, one);
                await Reply(held, "{\"choices\":[{\"message\":{\"content\":\"direct\"}}]}");
                await send.WaitAsync(deadline.Token);
            }
            finally { held.Response.Abort(); }

            var otherVendor = new ModelVendor(foreign, "local", "Local");
            var other = new Model(new ModelEndpoint(otherVendor, "two", second.Prefixes.Single(),
                Surface.ChatCompletions, new ModelAuthentication(foreign, "none")),
                "small", "same-wire", "Foreign", new ModelFeatures(Media.None));
            await RefuseConfiguredSelection(agent, other);
            await agent.SetModelAsync(two);
            Check(two.Agents().Count(member => ReferenceEquals(member, agent)) == 1,
                "Selecting the same Model does not duplicate reverse membership.");
            two.Destroy();
            Check(agent.SelectedModel() is null, "Model destruction removes selection without destroying Agent.");
            await agent.SetModelAsync(one);
            Check(ReferenceEquals(agent.SelectedModel(), one) && agent.Configuration().Offering == one.Identity(),
                "Explicit configured selection recovers without an automatic fallback.");
            await RefuseConfiguredSelection(agent, two);
            var unavailableBinding = new ModelAuthentication(app, "unavailable");
            var unavailable = new Model(new ModelEndpoint(vendor, "unavailable", first.Prefixes.Single(),
                Surface.ChatCompletions, unavailableBinding), "small", "same-wire", "Unavailable",
                new ModelFeatures(Media.None));
            unavailableBinding.Destroy();
            await RefuseConfiguredSelection(agent, unavailable);
            Check(one.Agents().Contains(agent) && unavailable.Agents().IsEmpty,
                "Unavailable targets never detach the current selection or gain membership.");
        }
        finally
        {
            try { app.Destroy(); }
            finally { foreign.Destroy(); }
        }
    }

    public static async Task ConfiguredSelectionSaveFailure(string temp)
    {
        string home = Path.Combine(temp, "configured-selection-save");
        Directory.CreateDirectory(home);
        var app = new Application(TextWriter.Null);
        try
        {
            var vendor = new ModelVendor(app, "local", "Local");
            var endpoint = new ModelEndpoint(vendor, "test", "http://127.0.0.1:1",
                Surface.ChatCompletions, new ModelAuthentication(app, "none"));
            var one = new Model(endpoint, "one", "same-wire", "One", new ModelFeatures(Media.None));
            var two = new Model(endpoint, "two", "same-wire", "Two", new ModelFeatures(Media.None));
            var parent = new Agent(app, new Configuration(Vendor.OpenAI, "legacy", "http://127.0.0.1:1",
                sandboxRoot: home), dataDirectory: home);
            Agent child = await parent.CreateChildAsync("saved");
            await child.SetModelAsync(one);
            string path = Path.Combine(child.DataDirectory(), "settings.json");
            string backup = path + ".backup";
            string original = File.ReadAllText(path);
            File.Move(path, backup);
            Directory.CreateDirectory(path);
            try
            {
                await RefuseConfiguredSelection(child, two);
                Check(one.Agents().Contains(child) && !two.Agents().Contains(child)
                    && File.ReadAllText(backup) == original,
                    "Failed durable save leaves both membership collections and saved bytes unchanged.");
            }
            finally
            {
                Directory.Delete(path);
                File.Move(backup, path);
            }
            var malformed = JsonNode.Parse(original)!;
            malformed["selected_offering"] = null;
            foreach (string damaged in new[] { malformed.ToJsonString(), "{}" })
            {
                File.WriteAllText(path, damaged);
                try { await RefuseConfiguredSelection(child, two); }
                finally { File.WriteAllText(path, original); }
            }
            await child.SetModelAsync(two);
            Check(JsonNode.Parse(File.ReadAllText(path))!["selected_offering"]!["model"]!.GetValue<string>() == "two"
                && ReferenceEquals(child.SelectedModel(), two) && !one.Agents().Contains(child),
                "Successful selection saves the whole candidate and publishes both membership ends.");
        }
        finally { app.Destroy(); }
    }
}
