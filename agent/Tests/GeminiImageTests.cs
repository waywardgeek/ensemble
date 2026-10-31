using System.Text.Json;
using System.Text.Json.Nodes;
using Ensemble;
using Application = Ensemble.Ensemble;

internal static class GeminiImageTests
{
    private static void Check(bool condition, string message)
    {
        if (!condition) throw new Exception(message);
    }

    private static void Unsupported(Action action)
    {
        try { action(); }
        catch (NotSupportedException) { return; }
        throw new Exception("Expected explicit unsupported-image refusal.");
    }

    public static void FunctionResponses(string temp)
    {
        string home = Path.Combine(temp, "gemini-image-results-" + Guid.NewGuid().ToString("N"));
        Directory.CreateDirectory(home);
        var app = new Application(TextWriter.Null);
        try
        {
            var config = new Configuration(Vendor.Gemini, "gemini-3-flash-preview",
                "http://127.0.0.1:1", surface: Surface.GenerateContent, sandboxRoot: home);
            var agent = new Agent(app, config, Path.Combine(home, "jobs"));
            var image = new BlobPart("image/png", new BlobRef(RefKind.Uri,
                "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mP8/x8AAwMCAO+jRZkAAAAASUVORK5CYII="));
            var call = new ToolCallPart("paired-call", config.Target(), "inspect_image",
                JsonSerializer.SerializeToElement(new { path = "original.png" }),
                JsonSerializer.SerializeToElement("original-thought-signature"));
            var marker = new RedactedPart("tool_result", 12, "budget",
                new BlobRef(RefKind.Path, "/not-opened/private-job-output"));
            Part[] content = [new TextPart("before"), image, marker, new TextPart("between"), image];
            Context ContextFor(IEnumerable<Part> parts, bool error = false) => new([
                new Entry(1, Actor.Agent, [call]),
                new Entry(2, Actor.Tool, [new ToolResultPart("paired-call", parts, error)])]);
            JsonObject Response(JsonNode wire) => (JsonObject)wire["contents"]![1]!["parts"]![0]!["functionResponse"]!;

            foreach (bool error in new[] { false, true })
            {
                var context = ContextFor(content, error);
                var wire = agent.Engine().Render(context);
                var result = Response(wire);
                Check(result["id"]!.GetValue<string>() == "paired-call"
                    && result["name"]!.GetValue<string>() == "inspect_image", "Gemini call pairing changed.");
                Check(wire["contents"]![0]!["parts"]![0]!["thoughtSignature"]!.GetValue<string>()
                    == "original-thought-signature", "Gemini call signature changed.");
                var response = error ? result["response"]!["error"]! : result["response"]!;
                var ordered = response["content"]!.AsArray();
                Check(ordered.Count == 5 && ordered[0]!["text"]!.GetValue<string>() == "before"
                    && ordered[2]!["text"]!.GetValue<string>().Contains("[redacted tool_result:")
                    && ordered[3]!["text"]!.GetValue<string>() == "between", "Gemini mixed result order changed.");
                var media = result["parts"]!.AsArray();
                Check(media.Count == 2, "Gemini result discarded or duplicated image parts.");
                for (int index = 0; index < media.Count; index++)
                {
                    var blob = media[index]!["inlineData"]!;
                    string name = blob["displayName"]!.GetValue<string>();
                    int referenceIndex = index == 0 ? 1 : 4;
                    Check(ordered[referenceIndex]!["$ref"]!.GetValue<string>() == name,
                        "Gemini image reference does not match its payload.");
                    Check(blob["mimeType"]!.GetValue<string>() == "image/png"
                        && Convert.FromBase64String(blob["data"]!.GetValue<string>()).SequenceEqual(image.Ref.Data()),
                        "Gemini image result contains metadata instead of bytes.");
                }
                Check(media[0]!["inlineData"]!["displayName"]!.GetValue<string>()
                    != media[1]!["inlineData"]!["displayName"]!.GetValue<string>(), "Gemini reused an image name.");
                Check(!wire.ToJsonString().Contains("private-job-output"), "Gemini leaked the redaction locator.");
                Check(((ToolResultPart)context.Dialogue[1].Parts[0]).Parts.SequenceEqual(content),
                    "Gemini rendering changed canonical parts.");
            }

            var plain = Response(agent.Engine().Render(ContextFor([new TextPart("{\"answer\":7}")])));
            Check(plain["response"]!["answer"]!.GetValue<int>() == 7 && plain["parts"] is null,
                "Gemini text-only object compatibility changed.");
            var missing = new BlobPart("image/png", new BlobRef(RefKind.Path, Path.Combine(home, "missing.png")));
            Unsupported(() => agent.Engine().Render(ContextFor([
                new ToolResultPart("nested", [missing])])));
            Unsupported(() => agent.Engine().Render(ContextFor([
                image, new ToolCallPart("nested-call", config.Target(), "other", JsonSerializer.SerializeToElement(new { }))])));
            Unsupported(() => agent.Engine().Render(ContextFor([
                new BlobPart("audio/wav", new BlobRef(RefKind.Path, Path.Combine(home, "missing.wav")))])));

            foreach (string model in new[] { "gemini-2.5-flash", "gemini-3-newer-not-verified" })
            {
                var unsupported = new Configuration(Vendor.Gemini, model, "http://127.0.0.1:1",
                    surface: Surface.GenerateContent, sandboxRoot: home);
                var other = new Agent(app, unsupported, Path.Combine(home, model));
                try
                {
                    var ownCall = new ToolCallPart("paired-call", unsupported.Target(), "inspect_image",
                        JsonSerializer.SerializeToElement(new { }));
                    Unsupported(() => other.Engine().Render(new Context([
                        new Entry(1, Actor.Agent, [ownCall]),
                        new Entry(2, Actor.Tool, [new ToolResultPart("paired-call", [missing])])])));
                }
                finally { other.Destroy(); }
            }
        }
        finally
        {
            app.Destroy();
            Directory.Delete(home, recursive: true);
        }
    }
}
