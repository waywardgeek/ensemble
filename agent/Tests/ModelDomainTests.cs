using Ensemble;
using Application = Ensemble.Ensemble;

internal static class ModelDomainTests
{
    private static void Check(bool condition, string reason)
    {
        if (!condition) throw new Exception(reason);
    }

    private static void Refuses(Action action)
    {
        try { action(); }
        catch (Exception error) when (error is ArgumentException or InvalidOperationException) { return; }
        throw new Exception("Expected model-domain refusal.");
    }

    public static void Ownership()
    {
        var app = new Application(TextWriter.Null);
        try
        {
            var auth = new ModelAuthentication(app, "local");
            var vendor = new ModelVendor(app, "ollama", "Ollama");
            var endpoint = new ModelEndpoint(vendor, "laptop", "http://127.0.0.1:11434", Surface.ChatCompletions, auth);
            var model = new Model(endpoint, "mimo", "mimo-v2.6:9b", "MiMo on laptop",
                new ModelFeatures(Media.None, ContextWindow: 262144), servingContextTokens: 4096);
            Check(ReferenceEquals(vendor.Parent(), app) && app.ModelVendors().Single() == vendor,
                "Vendor constructor establishes both ownership ends.");
            Check(ReferenceEquals(endpoint.Parent(), vendor) && vendor.Endpoints().Single() == endpoint,
                "Endpoint constructor establishes both ownership ends.");
            Check(ReferenceEquals(model.Parent(), endpoint) && endpoint.Models().Single() == model
                && ReferenceEquals(model.Vendor(), vendor), "Model follows actual parents.");
            Check(model.Features.ContextWindow == 262144 && model.ServingContextTokens == 4096,
                "Advertised and serving context are independent facts.");
            Check(endpoint.Surface == Surface.ChatCompletions && model.Vendor().ID == "ollama",
                "OpenAI-compatible surface does not change vendor identity.");

            var otherEndpoint = new ModelEndpoint(vendor, "other", "http://127.0.0.1:11435/v1", Surface.ChatCompletions, auth);
            var other = new Model(otherEndpoint, "mimo", model.Name, "MiMo elsewhere", new ModelFeatures(Media.None));
            Check(other != model && other.Parent() != model.Parent() && other.ServingContextTokens is null,
                "Same provider model name through another endpoint is a distinct offering; unknown capacity stays unknown.");
            var snapshot = vendor.Endpoints();
            otherEndpoint.Destroy();
            Check(snapshot.Length == 2 && vendor.Endpoints().Length == 1,
                "Snapshots are values, not mutable ownership collections.");
            Refuses(() => other.Parent());
            Check(model.Parent() == endpoint, "Destroying one endpoint preserves siblings.");
        }
        finally { app.Destroy(); }
        Check(app.ModelVendors().Length == 0, "Root destruction removes all vendors.");
    }

    public static void Refusals()
    {
        var app = new Application(TextWriter.Null);
        try
        {
            var auth = new ModelAuthentication(app, "local");
            Refuses(() => new ModelVendor(app, "a/b", "Invalid"));
            Refuses(() => new ModelVendor(app, "vendor", "bad\nlabel"));
            Check(app.ModelVendors().Length == 0, "Invalid vendor changes nothing.");
            var vendor = new ModelVendor(app, "vendor", "Vendor");
            Refuses(() => new ModelVendor(app, "vendor", "Duplicate"));
            Check(app.ModelVendors().Single() == vendor, "Duplicate vendor preserves original membership.");

            foreach (string address in new[] { "relative", "file:///tmp/model", "http://user:secret@localhost/",
                "http://localhost/?token=secret", "http://localhost/#fragment" })
            {
                Refuses(() => new ModelEndpoint(vendor, "invalid", address, Surface.ChatCompletions, auth));
            }
            Refuses(() => new ModelEndpoint(vendor, "invalid", "http://localhost", (Surface)0, auth));
            Check(vendor.Endpoints().Length == 0, "Invalid endpoint changes nothing.");
            var endpoint = new ModelEndpoint(vendor, "service", "https://example.invalid/v1", Surface.ChatCompletions, auth);
            Refuses(() => new ModelEndpoint(vendor, "service", "http://localhost", Surface.Messages, auth));
            Check(vendor.Endpoints().Single() == endpoint, "Duplicate endpoint cannot replace the destination.");

            Refuses(() => new Model(endpoint, "m", "model", "Model", new ModelFeatures(Media.None), 0));
            Refuses(() => new Model(endpoint, "m", "model", "Model", new ModelFeatures(Media.None, ContextWindow: 0)));
            Refuses(() => new Model(endpoint, "m", "model", "Model", new ModelFeatures(Media.None, ContextWindow: 4096), 4097));
            Refuses(() => new Model(endpoint, "m", "model", "Model", new ModelFeatures(Media.None, ResponsesOnly: true)));
            Refuses(() => new Model(endpoint, "m", "model\n", "Model", new ModelFeatures(Media.None)));
            Check(endpoint.Models().Length == 0, "Invalid model changes nothing.");
            var model = new Model(endpoint, "m", "model", "Model", new ModelFeatures(Media.None));
            Refuses(() => new Model(endpoint, "m", "another-model", "Duplicate", new ModelFeatures(Media.None)));
            Check(endpoint.Models().Single() == model, "Duplicate ID cannot replace the provider model name.");

            app.Destroy();
            Refuses(() => new ModelVendor(app, "late", "Late"));
            Refuses(() => new ModelEndpoint(vendor, "late", "http://localhost", Surface.ChatCompletions, auth));
            Refuses(() => new Model(endpoint, "late", "model", "Late", new ModelFeatures(Media.None)));
            Refuses(() => model.Parent());
            Refuses(() => endpoint.Parent());
            Refuses(() => vendor.Parent());
            Check(app.ModelVendors().Length == 0 && vendor.Endpoints().Length == 0 && endpoint.Models().Length == 0,
                "Destruction closes admission and recursively removes both relationship ends.");
        }
        finally { app.Destroy(); }
    }

    public static void Detach()
    {
        var app = new Application(TextWriter.Null);
        var foreign = new Application(TextWriter.Null);
        try
        {
            var auth = new ModelAuthentication(app, "local");
            var vendor = new ModelVendor(app, "vendor", "Vendor");
            var peer = new ModelVendor(app, "peer", "Peer");
            var endpoint = new ModelEndpoint(vendor, "service", "http://localhost", Surface.ChatCompletions, auth);
            var model = new Model(endpoint, "model", "wire-name", "Model", new ModelFeatures(Media.None));
            Refuses(() => peer.RemoveEndpoint(endpoint));
            Check(endpoint.Parent() == vendor && vendor.Endpoints().Single() == endpoint,
                "Wrong-parent removal changes neither end.");
            endpoint.RemoveModel(model);
            Check(endpoint.Models().Length == 0 && !model.destroyed, "Remove detaches without destroying.");
            Refuses(() => model.Parent());
            endpoint.AddModel(model);
            vendor.RemoveEndpoint(endpoint);
            Check(vendor.Endpoints().Length == 0 && !endpoint.destroyed && endpoint.Models().Single() == model,
                "Endpoint detach does not cascade through its children.");
            Refuses(() => endpoint.Parent());
            vendor.AddEndpoint(endpoint);
            app.RemoveModelVendor(vendor);
            Check(!vendor.destroyed && vendor.Endpoints().Single() == endpoint,
                "Vendor detach does not destroy its graph.");
            Refuses(() => foreign.AddModelVendor(vendor));
            Check(foreign.ModelVendors().Length == 0, "A graph cannot silently migrate its lifetime lock to another root.");
            app.AddModelVendor(vendor);
            vendor.Destroy();
            vendor.Destroy();
            Check(app.ModelVendors().Single() == peer && peer.Parent() == app,
                "Idempotent subtree destruction preserves an independent vendor.");
            Refuses(() => model.Parent());
            Refuses(() => endpoint.Parent());
        }
        finally
        {
            app.Destroy();
            foreign.Destroy();
        }
    }
}
