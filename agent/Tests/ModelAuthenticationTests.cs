using Ensemble;
using Application = Ensemble.Ensemble;

internal static class ModelAuthenticationTests
{
    private static void Check(bool condition, string reason)
    {
        if (!condition) throw new Exception(reason);
    }

    private static void Refuses(Action action)
    {
        try { action(); }
        catch (Exception error) when (error is ArgumentException or InvalidOperationException) { return; }
        throw new Exception("Expected authentication-binding refusal.");
    }

    public static void Bindings()
    {
        using var diagnostics = new StringWriter();
        var app = new Application(diagnostics);
        var foreign = new Application(TextWriter.Null);
        string unusedPath = Path.Combine(Path.GetTempPath(), "model-auth-unused-" + Guid.NewGuid().ToString("N"));
        try
        {
            var vendor = new ModelVendor(app, "vendor", "Vendor");
            var none = new ModelAuthentication(app, "local");
            var local = new ModelEndpoint(vendor, "local", "http://192.0.2.1:11434", Surface.ChatCompletions, none);
            Check(none.Kind == ModelAuthenticationKind.None && local.Authentication() == none
                && none.Endpoints().Single() == local, "Explicit unauthenticated binding establishes both ends.");

            const string key = "synthetic-model-key-never-print";
            var api = new ModelAuthentication(app, "api", key);
            var apiEndpoint = new ModelEndpoint(vendor, "api", "https://example.invalid/v1", Surface.ChatCompletions, api);
            Refuses(() => new ModelAuthentication(app, "api", "another-synthetic-key"));
            Refuses(() => new ModelAuthentication(app, "empty", ""));
            Refuses(() => new ModelEndpoint(vendor, "unsafe", "http://192.0.2.1/v1", Surface.ChatCompletions, api));
            Check(api.Endpoints().Single() == apiEndpoint, "Unsafe endpoint and duplicate binding change nothing.");
            app.Log("Synthetic diagnostic containing " + key);
            Check(!diagnostics.ToString().Contains(key, StringComparison.Ordinal)
                && diagnostics.ToString().Contains("[REDACTED]", StringComparison.Ordinal),
                "Owned API key is redacted even before any Agent selects it.");
            Check(!api.ToString()!.Contains(key, StringComparison.Ordinal), "Authentication is not a secret-printing record.");

            var foreignNone = new ModelAuthentication(foreign, "foreign-local");
            int endpointCount = vendor.Endpoints().Length;
            Refuses(() => new ModelEndpoint(vendor, "foreign", "http://127.0.0.1", Surface.ChatCompletions, foreignNone));
            Refuses(() => new ModelEndpoint(vendor, "local", "http://127.0.0.1", Surface.ChatCompletions, none));
            Check(vendor.Endpoints().Length == endpointCount && foreignNone.Endpoints().Length == 0
                && none.Endpoints().Single() == local, "Foreign and duplicate endpoint admission roll back both relationships.");

            var planStore = new Credentials(app, unusedPath);
            var plan = new ModelAuthentication(app, "plan", planStore, "synthetic-account");
            var planEndpoint = new ModelEndpoint(vendor, "plan", "https://api.openai.com/v1", Surface.Responses, plan);
            Check(plan.PlanCredentials() == planStore && planStore.modelBindings.Single() == plan,
                "Plan binding references the existing owner and reverse membership.");
            Refuses(() => new ModelAuthentication(foreign, "wrong-owner", planStore, "synthetic-account"));
            Refuses(() => new ModelEndpoint(vendor, "wrong-surface", "https://api.openai.com", Surface.ChatCompletions, plan));
            Check(planStore.modelBindings.Single() == plan, "Foreign binding refusal preserves existing owner memberships.");
            planStore.Destroy();
            Check(plan.PlanCredentials() is null && planStore.modelBindings.Count == 0
                && plan.Kind == ModelAuthenticationKind.ChatGptPlan && planEndpoint.Authentication() == plan,
                "Credential destruction detaches weak source links without changing binding kind or destroying endpoints.");
            Refuses(() => plan.ValidateEndpoint(planEndpoint.Address, planEndpoint.Surface));
            var replacement = new Credentials(app, unusedPath);
            Check(plan.PlanCredentials() is null && replacement.modelBindings.Count == 0,
                "A replacement store does not silently authorize an old binding.");
            Refuses(() => new ModelEndpoint(vendor, "stale-plan", "https://api.openai.com", Surface.Responses, plan));

            var codexStore = new CodexCredentials(app, unusedPath + ".json");
            var codex = new ModelAuthentication(app, "codex", codexStore);
            var codexEndpoint = new ModelEndpoint(vendor, "codex", CodexCredentials.ServiceBase, Surface.Responses, codex);
            Check(codex.CodexCredentials() == codexStore && codexStore.modelBindings.Single() == codex,
                "Codex source is distinct from public OAuth.");
            Refuses(() => new ModelEndpoint(vendor, "codex-other", "https://example.invalid", Surface.Responses, codex));
            codexStore.Destroy();
            Check(codex.CodexCredentials() is null && codexStore.modelBindings.Count == 0,
                "Codex destruction detaches both weak relationship ends.");
            Refuses(() => codex.ValidateEndpoint(codexEndpoint.Address, codexEndpoint.Surface));
            Check(!Directory.Exists(unusedPath) && !File.Exists(unusedPath + ".json"),
                "Graph construction/destruction creates no credential files; no credential acquisition was requested.");

            var model = new Model(local, "small", "wire-model", "Small", new ModelFeatures(Media.None));
            none.Destroy();
            Refuses(() => local.Authentication());
            Check(local.Parent() == vendor && model.Parent() == local && none.Endpoints().Length == 0,
                "Binding destruction invalidates endpoints without cascading into model ownership.");
            Check(api.Parent() == app && apiEndpoint.Authentication() == api,
                "Independent authentication remains live.");
            app.Destroy();
            Check(app.ModelVendors().Length == 0 && app.ModelAuthentications().Length == 0
                && api.Endpoints().Length == 0 && plan.Endpoints().Length == 0,
                "Root destruction removes model and authentication graphs.");
        }
        finally
        {
            app.Destroy();
            foreign.Destroy();
        }
    }
}
