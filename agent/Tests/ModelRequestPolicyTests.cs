using Ensemble;
using Application = Ensemble.Ensemble;

internal static class ModelRequestPolicyTests
{
    private static void Check(bool value, string reason)
    {
        if (!value) throw new Exception(reason);
    }

    public static void Admission()
    {
        var app = new Application(TextWriter.Null);
        try
        {
            var auth = new ModelAuthentication(app, "none");
            var vendor = new ModelVendor(app, "fixture", "Fixture");
            var responses = new ModelEndpoint(vendor, "responses", "http://127.0.0.1:1",
                Surface.Responses, auth);
            var chat = new ModelEndpoint(vendor, "chat", "http://127.0.0.1:2",
                Surface.ChatCompletions, auth);
            var features = new ModelFeatures(Media.None, Stream: true,
                ReasoningEfforts: ["medium", "high"], ReasoningSummaries: ["auto"],
                RequireToolExplanation: true);
            var requested = new ModelRequestPolicy(disableStreaming: true,
                stubsToolResults: true, promptCacheKey: "fixture-cache",
                reasoningEffort: "high", reasoningSummary: "off");
            var model = new Model(responses, "explicit", "wire", "Explicit", features,
                requestPolicy: requested);
            Check(model.RequestPolicy.ReasoningEffort == "high"
                && model.RequestPolicy.ReasoningSummary == "off"
                && model.RequestPolicy.DisableStreaming && model.RequestPolicy.StubsToolResults
                && model.RequestPolicy.PromptCacheKey == "fixture-cache",
                "Offering retains explicit credential-free request options.");
            Check(model.RequestPolicy.RequireToolExplanation && !requested.RequireToolExplanation,
                "Capability enforcement is normalized without mutating the supplied value.");
            var defaults = new Model(responses, "defaults", "wire", "Defaults", features);
            Check(defaults.RequestPolicy.ReasoningEffort == "medium"
                && defaults.RequestPolicy.ReasoningSummary == "auto",
                "Omitted reasoning options use validated defaults.");
            var local = new Model(chat, "local", "wire", "Local", new ModelFeatures(Media.None));
            Check(local.RequestPolicy.ReasoningEffort == "" && local.RequestPolicy.ReasoningSummary == ""
                && !local.RequestPolicy.RequireToolExplanation && !local.RequestPolicy.DisableStreaming,
                "Same wire name does not inherit another offering's policy.");
            foreach (ModelRequestPolicy invalid in new[] {
                new ModelRequestPolicy(deferredTools: true),
                new ModelRequestPolicy(cacheBreakpoints: true),
                new ModelRequestPolicy(reasoningEffort: "high"),
                new ModelRequestPolicy(reasoningSummary: "off") })
            {
                bool refused = false;
                try { _ = new Model(chat, "bad", "wire", "Invalid", new ModelFeatures(Media.None), requestPolicy: invalid); }
                catch (ArgumentException) { refused = true; }
                Check(refused && chat.Models().Length == 1 && chat.Models()[0] == local,
                    "Invalid policy refuses before endpoint membership changes.");
            }
            var agentPolicy = new Configuration(Vendor.Anthropic, "legacy", "https://unused.invalid",
                primaryInstructions: "Retain instructions", toolRoundLimit: 7, safeMode: true,
                requireToolExplanation: true, disableStreaming: true, deferredTools: true);
            var projected = new Configuration(agentPolicy, model);
            var switched = new Configuration(projected, local);
            var returned = new Configuration(switched, model);
            Check(projected.ReasoningEffort == "high" && projected.ReasoningSummary == ""
                && projected.DisableStreaming && projected.StubsToolResults
                && projected.PromptCacheKey == "fixture-cache" && !projected.DeferredTools,
                "Projection uses the offering policy rather than the previous route's options.");
            Check(!switched.DisableStreaming && !switched.RequireToolExplanation
                && switched.ReasoningEffort == "" && switched.PromptCacheKey == ""
                && !switched.StubsToolResults,
                "Switching to local policy does not retain cloud request options.");
            Check(returned.ReasoningEffort == projected.ReasoningEffort
                && returned.ReasoningSummary == projected.ReasoningSummary
                && returned.DisableStreaming == projected.DisableStreaming
                && returned.PromptCacheKey == projected.PromptCacheKey
                && returned.RequireToolExplanation == projected.RequireToolExplanation,
                "Returning to an offering restores its own explicit request policy.");
            Check(switched.PrimaryInstructions == "Retain instructions"
                && switched.ToolRoundLimit == 7 && switched.SafeMode,
                "Route changes preserve Agent instructions, limits and security.");
            bool unsupported = false;
            try
            {
                _ = new Model(responses, "bad", "wire", "Invalid", features,
                    requestPolicy: new ModelRequestPolicy(reasoningEffort: "max"));
            }
            catch (ArgumentException) { unsupported = true; }
            Check(unsupported && responses.Models().Length == 2,
                "Reasoning validation uses this offering's capabilities.");
        }
        finally { app.Destroy(); }
    }
}
