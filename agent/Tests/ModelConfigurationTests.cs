using System.Text;
using System.Text.Json.Nodes;
using Ensemble;
using Application = Ensemble.Ensemble;

internal static class ModelConfigurationTests
{
    private const string Document = """
        {"version":1,"vendors":[{"id":"local","label":"Local inference","endpoints":[
          {"id":"ollama","address":"http://127.0.0.1:1","surface":"chat_completions",
           "authentication":"local-none","models":[
             {"id":"small","name":"same-wire","label":"Small","features":
              {"stream":true,"context_window":262144},"serving_context_tokens":4096}]}]}]}
        """;

    private static void Check(bool value, string message)
    {
        if (!value) throw new Exception(message);
    }

    private static void Refuses(Application root, string document)
    {
        try { root.LoadModelConfiguration(Encoding.UTF8.GetBytes(document)); }
        catch (Exception error) when (error is InvalidDataException or ArgumentException or InvalidOperationException)
        {
            return;
        }
        throw new Exception("Expected model configuration refusal.");
    }

    public static void Loading()
    {
        var root = new Application(TextWriter.Null);
        try
        {
            var none = new ModelAuthentication(root, "local-none");
            var agent = new Agent(root, new Configuration(Vendor.Anthropic, "unchanged", "http://127.0.0.1:2", "synthetic"));
            Configuration original = agent.Configuration();
            root.LoadModelConfiguration(Encoding.UTF8.GetBytes(Document));
            var identity = new ModelIdentity("local", "ollama", "small");
            Model model = root.ResolveModel(identity);
            Check(model.Name == "same-wire" && model.Features.Stream && model.Features.Media == Media.None
                && model.Features.ContextWindow == 262144 && model.ServingContextTokens == 4096,
                "Document preserves configured capabilities and distinct advertised/serving capacities.");
            Check(ReferenceEquals(model.Parent().Authentication(), none) && none.Endpoints().Length == 1,
                "Document references the actual existing binding with both relationship ends.");
            Check(ReferenceEquals(agent.Configuration(), original) && agent.SelectedModel() is null,
                "Loading offerings does not replace the existing launch route or funding.");
            Refuses(root, Document);
            Check(ReferenceEquals(root.ResolveModel(identity), model) && root.ModelVendors().Length == 1,
                "Repeated vendor identity refuses without replacing existing objects.");

            var badDocuments = new List<string>
            {
                Document.Replace("\"version\":1", "\"version\":2"),
                Document.Replace("\"version\":1", "\"version\":1,\"version\":1"),
                Document.Replace("\"stream\":true", "\"stream\":null"),
                Document.Replace("\"stream\":true", "\"stream\":true,\"unknown\":false"),
                Document.Replace("\"stream\":true", "\"stream\":true,\"stream\":false"),
                Document.Replace("\"stream\":true", "\"stream\":true,\"media\":[\"image\",\"image\"]"),
                Document.Replace("\"stream\":true", "\"stream\":true,\"media\":[\"magic\"]"),
                Document.Replace("4096", "0"),
                Document.Replace("\"serving_context_tokens\":4096", "\"request_policy\":null"),
                Document.Replace("\"serving_context_tokens\":4096", "\"request_policy\":{\"unknown\":true}"),
                Document.Replace("\"serving_context_tokens\":4096", "\"request_policy\":{\"disable_streaming\":1}"),
                Document.Replace("\"serving_context_tokens\":4096", "\"request_policy\":{\"reasoning_effort\":null}"),
                Document.Replace("\"serving_context_tokens\":4096", "\"request_policy\":{\"disable_streaming\":true,\"disable_streaming\":false}"),
                Document.Replace("\"version\":1", "\"version\":null"),
                "{", new string(' ', 1024 * 1024 + 1)
            };
            foreach (string bad in badDocuments)
            {
                // A fresh vendor ID prevents duplicate-existing-vendor refusal
                // from masking an incorrectly accepted malformed document.
                Refuses(root, bad.Replace("\"id\":\"local\"", "\"id\":\"malformed\""));
                Check(root.ModelVendors().Length == 1 && none.Endpoints().Length == 1,
                    "Malformed document cannot add a graph or binding membership.");
            }
            byte[] invalidUtf8 = Encoding.UTF8.GetBytes(Document.Replace("\"id\":\"local\"", "\"id\":\"invalid-utf8\""));
            int label = Array.IndexOf(invalidUtf8, (byte)'L');
            invalidUtf8[label] = 0xff;
            try
            {
                root.LoadModelConfiguration(invalidUtf8);
                throw new Exception("Invalid UTF-8 was accepted.");
            }
            catch (InvalidDataException) { }
            Check(root.ModelVendors().Length == 1, "Invalid UTF-8 leaves the graph unchanged.");

            // These documents parse successfully and fail only after some new graph
            // objects have been constructed. No previously owned object may change.
            foreach (string fault in new[] { "binding", "address", "capacity", "model-duplicate", "vendor-duplicate", "policy" })
            {
                JsonObject document = JsonNode.Parse(Document)!.AsObject();
                JsonArray vendors = document["vendors"]!.AsArray();
                var first = vendors[0]!.AsObject();
                first["id"] = "candidate";
                var later = first.DeepClone().AsObject();
                later["id"] = "later";
                var endpoint = later["endpoints"]![0]!;
                switch (fault)
                {
                    case "policy": endpoint["models"]![0]!["request_policy"] = new JsonObject { ["deferred_tools"] = true }; break;
                    case "binding": endpoint["authentication"] = "missing"; break;
                    case "address": endpoint["address"] = "http://user:secret@127.0.0.1:1"; break;
                    case "capacity": endpoint["models"]![0]!["serving_context_tokens"] = 300000; break;
                    case "model-duplicate": endpoint["models"]!.AsArray().Add(endpoint["models"]![0]!.DeepClone()); break;
                    case "vendor-duplicate": later["id"] = "candidate"; break;
                }
                vendors.Add(later);
                Refuses(root, document.ToJsonString());
                Check(root.ModelVendors().Length == 1 && ReferenceEquals(root.ResolveModel(identity), model)
                    && none.Endpoints().Length == 1 && root.ModelAuthentications().Length == 1
                    && ReferenceEquals(agent.Configuration(), original),
                    "Refused load rolls back only new subtrees and retains binding memberships: " + fault);
            }
            JsonObject configured = JsonNode.Parse(Document)!.AsObject();
            configured["vendors"]![0]!["id"] = "policy";
            configured["vendors"]![0]!["endpoints"]![0]!["models"]![0]!["request_policy"] = new JsonObject
            {
                ["disable_streaming"] = true, ["require_tool_explanation"] = true,
                ["stub_tool_results"] = true, ["prompt_cache_key"] = "local-cache"
            };
            root.LoadModelConfiguration(Encoding.UTF8.GetBytes(configured.ToJsonString()));
            Model withPolicy = root.ResolveModel(new ModelIdentity("policy", "ollama", "small"));
            var projected = new Configuration(original, withPolicy);
            Check(projected.DisableStreaming && projected.RequireToolExplanation && projected.StubsToolResults
                && projected.PromptCacheKey == "local-cache",
                "Host document policy reaches the configured request projection.");
            Check(agent.History().Events().All(item => item.Kind != EventKind.RequestSent),
                "Configuration loading must not initiate inference.");
        }
        finally { root.Destroy(); }
    }
}
