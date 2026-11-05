// Chapter 3 public library boundaries. No tool execution or later owner APIs.
using System.Text.Json.Nodes;
using Ensemble;

static class LibraryChecks
{
    public static int Run()
    {
        var checks = new List<string>();
        void Require(bool condition, string name)
        {
            if (!condition) throw new InvalidOperationException(name);
            checks.Add(name);
        }
        bool Detached(Tool tool)
        {
            try { _ = tool.Parent(); return false; }
            catch (InvalidOperationException) { return true; }
        }
        using var diagnostics = new StringWriter();
        var app = new Ensemble.Ensemble(diagnostics);
        try
        {
            foreach (var (vendor, model, surface) in new[] {
                (Vendor.Anthropic, "claude-3-haiku-20240307", Surface.Messages),
                (Vendor.OpenAI, "gpt-4o-mini", Surface.ChatCompletions),
                (Vendor.Gemini, "gemini-2.0-flash", Surface.GenerateContent) })
            {
                void Check(bool condition, string name) => Require(condition, surface + "-" + name);
                var configuration = new Configuration(vendor, model, "http://127.0.0.1:1", "synthetic-unused");
                Check(configuration.Target().Surface == surface, "early-route");
                var owner = new Agent(app, configuration);
                var peer = new Agent(app, configuration);
                try
                {
                    var context = new Context([new Entry(1, Actor.Human, [new TextPart("library-" + Guid.NewGuid().ToString("N"))])]);
                    Check(owner.Tools().IsEmpty && peer.Tools().IsEmpty, "initial-registry-empty");
                    var empty = owner.Engine().Render(context);
                    Check(!empty.ContainsKey("tools") && !peer.Engine().Render(context).ContainsKey("tools"), "empty-declarations-omitted");
                    var tool = new Tool(owner, ToolKind.ReadFile);
                    Check(ReferenceEquals(tool.Parent(), owner) && owner.Tools().Length == 1
                        && ReferenceEquals(owner.Tools()[0], tool) && tool.Name() == "read_file", "constructor-links-both-ends");
                    Check(peer.Tools().IsEmpty && !peer.Engine().Render(context).ContainsKey("tools"), "peer-remains-empty");
                    var declaration = tool.Declaration(vendor);
                    var expected = declaration.DeepClone();
                    var request = owner.Engine().Render(context);
                    var declarations = vendor == Vendor.Gemini
                        ? request["tools"]?[0]?["functionDeclarations"]?.AsArray()
                        : request["tools"]?.AsArray();
                    Check(declarations is { Count: 1 } && JsonNode.DeepEquals(declarations[0], expected), "owned-declaration-present");
                    var body = vendor == Vendor.OpenAI ? declaration["function"]!.AsObject() : declaration;
                    Check(body["name"]?.GetValue<string>() == "read_file"
                        && !string.IsNullOrWhiteSpace(body["description"]?.GetValue<string>()), "definition-present");
                    string schemaKey = vendor == Vendor.Anthropic ? "input_schema" : "parameters";
                    Check(body[schemaKey]?["properties"]?["path"] is not null, "path-schema-present");
                    // Mutate only caller-owned projections. Later projections must be fresh.
                    body["name"] = "external-corruption";
                    body[schemaKey] = new JsonObject();
                    declarations![0] = new JsonObject { ["name"] = "request-corruption" };
                    Check(JsonNode.DeepEquals(tool.Declaration(vendor), expected), "declaration-mutation-isolated");
                    var rerendered = owner.Engine().Render(context);
                    var fresh = vendor == Vendor.Gemini
                        ? rerendered["tools"]?[0]?["functionDeclarations"]?.AsArray()
                        : rerendered["tools"]?.AsArray();
                    Check(fresh is { Count: 1 } && JsonNode.DeepEquals(fresh[0], expected), "request-mutation-isolated");
                    foreach (var other in new[] { Vendor.Anthropic, Vendor.OpenAI, Vendor.Gemini })
                        _ = tool.Declaration(other);
                    Check(JsonNode.DeepEquals(tool.Declaration(vendor), expected), "cross-vendor-projection-stable");
                    tool.Destroy();
                    Check(owner.Tools().IsEmpty && Detached(tool), "tool-destruction-detaches");
                    Check(JsonNode.DeepEquals(owner.Engine().Render(context), empty), "last-tool-removal-restores-request");
                    var owned = new Tool(owner, ToolKind.ReadFile);
                    var independent = new Tool(peer, ToolKind.ListDirectory);
                    owner.Destroy();
                    Check(Detached(owned), "owner-destruction-detaches-child");
                    Check(ReferenceEquals(independent.Parent(), peer) && peer.Tools().Length == 1
                        && peer.Engine().Render(context).ContainsKey("tools"), "peer-survives-owner-destruction");
                    independent.Destroy();
                    Check(peer.Tools().IsEmpty && !peer.Engine().Render(context).ContainsKey("tools"), "peer-final-removal-omits");
                }
                finally { owner.Destroy(); peer.Destroy(); }
            }
            Console.WriteLine(System.Text.Json.JsonSerializer.Serialize(new {
                contract = "chapter3-library-v1", passed = true, checks,
                scope = "Public ownership and rendering only; no effectful dispatch or chapter score." }));
            return 0;
        }
        catch (Exception error)
        {
            Console.WriteLine(System.Text.Json.JsonSerializer.Serialize(new {
                contract = "chapter3-library-v1", passed = false, checks, failure = error.Message }));
            return 1;
        }
        finally { app.Destroy(); }
    }
}
