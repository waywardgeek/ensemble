using System.Text.Json;
using System.Net;
using Ensemble;
using Application = Ensemble.Ensemble;

internal static partial class MeterTests
{
    public static async Task ConfiguredProjection()
    {
        using var source = Listen();
        using var cancellation = new CancellationTokenSource(TimeSpan.FromSeconds(15));
        var app = new Application(TextWriter.Null);
        Task? serving = null;
        Task<ParsedResponse>? exchange = null;
        try
        {
            var vendor = new ModelVendor(app, "fixture", "Owned fixture");
            var anonymous = new ModelAuthentication(app, "anonymous");
            var keyed = new ModelAuthentication(app, "keyed", "selected-synthetic-key");
            var firstEndpoint = new ModelEndpoint(vendor, "anonymous", source.Prefixes.Single(),
                Surface.ChatCompletions, anonymous);
            var secondEndpoint = new ModelEndpoint(vendor, "keyed", source.Prefixes.Single(),
                Surface.ChatCompletions, keyed);
            var features = new ModelFeatures(Media.None, Stream: true, ContextWindow: 262144);
            var first = new Model(firstEndpoint, "model", "uncatalogued-fixture", "Anonymous", features, 4096);
            var second = new Model(secondEndpoint, "model", first.Name, "Keyed", features, 8192);
            var policy = new Configuration(Vendor.Anthropic, "old-model", "https://unused.invalid",
                "must-not-leak-old-key", primaryInstructions: "Keep my instructions.", safeMode: true);
            var firstConfig = new Configuration(policy, first);
            var secondConfig = new Configuration(policy, second);
            Check(firstConfig.ApiKey == "" && secondConfig.ApiKey == ""
                && firstConfig.PrimaryInstructions == policy.PrimaryInstructions && firstConfig.SafeMode,
                "Projection preserves Agent policy without copying API secrets.");
            Check(firstConfig.Billing == BillingMode.None && firstConfig.Features.Stream
                && firstConfig.FundingDescription().Contains("Unauthenticated"),
                "Configured unknown model has explicit capabilities and no-auth funding.");
            Check(!firstConfig.Target().Matches(secondConfig.Target()),
                "Identical wire names at distinct offerings cannot share opaque replay.");
            foreach (Configuration derived in new[] { firstConfig.ForMemoryCompressor(),
                firstConfig.ForRecallJudge("", "Judge only."), new Configuration(firstConfig, true, false, "") })
            {
                Check(ReferenceEquals(derived.Authentication, anonymous)
                    && derived.Offering == first.Identity() && derived.Features == features,
                    "Auxiliary/security derivation preserves the exact selected binding and capabilities.");
            }
            bool refused = false;
            try { firstConfig.ForRecallJudge("different", "Judge only."); }
            catch (ArgumentException) { refused = true; }
            Check(refused, "Unconfigured judge override must refuse instead of changing wire names.");
            refused = false;
            try { firstConfig.WithModel("different"); }
            catch (ArgumentException) { refused = true; }
            Check(refused && ReferenceEquals(firstConfig.WithModel(first.Name), firstConfig),
                "String-only model changes cannot drop a configured binding.");
            app.RemoveModelVendor(vendor);
            refused = false;
            try { _ = new Configuration(policy, first); }
            catch (ObjectDisposedException) { refused = true; }
            finally { app.AddModelVendor(vendor); }
            Check(refused, "Detached vendor cannot yield an authorized request projection.");
            var firstAgent = new Agent(app, firstConfig);
            var secondAgent = new Agent(app, secondConfig);
            Check(firstAgent.WorkingContextStatus().WorkingTokens == 4096
                && secondAgent.WorkingContextStatus().WorkingTokens == 8192,
                "Default working context uses configured serving allocation, not advertised maximum.");
            serving = Task.Run(async () =>
            {
                for (int index = 0; index < 2; index++)
                {
                    var incoming = await source.GetContextAsync().WaitAsync(cancellation.Token);
                    Check(incoming.Request.Headers["Authorization"] == (index == 0 ? null : "Bearer selected-synthetic-key")
                        && incoming.Request.Headers["x-api-key"] is null
                        && incoming.Request.Headers["x-goog-api-key"] is null,
                        "Request authentication must come only from the selected binding.");
                    using var reader = new StreamReader(incoming.Request.InputStream);
                    using var body = JsonDocument.Parse(await reader.ReadToEndAsync(cancellation.Token));
                    Check(body.RootElement.GetProperty("model").GetString() == first.Name
                        && body.RootElement.GetProperty("stream").GetBoolean(),
                        "Provider receives exact wire name and configured streaming policy.");
                    await Reply(incoming, JsonSerializer.Serialize(new
                    {
                        model = first.Name,
                        choices = new[] { new { message = new { content = "answer" } } },
                        usage = new { prompt_tokens = 1, completion_tokens = 1 }
                    }));
                }
            });
            foreach (Agent agent in new[] { firstAgent, secondAgent })
            {
                exchange = agent.Engine().SendAsync(agent.Engine().Render(new Context()).ToJsonString(), cancellation.Token);
                ParsedResponse result = await exchange;
                Check(result.Events.Single(fact => fact.Kind == EventKind.ResponseEnded).From!
                    .Matches(agent.Configuration().Target()), "Response preserves the configured offering provenance.");
            }
            await serving;
            anonymous.Destroy();
            refused = false;
            try
            {
                exchange = firstAgent.Engine().SendAsync(firstAgent.Engine().Render(new Context()).ToJsonString(), cancellation.Token);
                await exchange;
            }
            catch (InvalidOperationException) { refused = true; }
            Check(refused, "Removed authentication refuses instead of falling back to an empty or old key.");
            exchange = null; // The expected refusal above was observed, not a cleanup failure.
        }
        finally
        {
            cancellation.Cancel();
            source.Stop();
            try
            {
                try { await Task.WhenAll(new Task?[] { serving, exchange }.OfType<Task>()); }
                catch (Exception error) when (error is OperationCanceledException or HttpListenerException
                    or ObjectDisposedException or HttpRequestException) { }
            }
            finally { app.Destroy(); }
        }
    }
}
