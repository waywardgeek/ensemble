using System.Security.Cryptography;
using System.Text.Json;
using System.Text.Json.Nodes;
using Ensemble;

internal static class DiagnosticMediaTests
{
    public static Task Projection()
    {
        byte[] bytes = Enumerable.Range(0, 9000).Select(index => (byte)(index % 251)).ToArray();
        string encoded = Convert.ToBase64String(bytes);
        string hash = Convert.ToHexStringLower(SHA256.HashData(bytes));
        string source = JsonSerializer.Serialize(new
        {
            text = "preserve ordinary caption", data = "ordinary data field",
            image_url = "data:image/png;base64," + encoded,
            source = new { type = "base64", media_type = "image/png", data = encoded },
            inlineData = new { mimeType = "audio/wav", data = encoded },
            input_audio = new { format = "wav", data = encoded },
            output = new { type = "response.audio.delta", delta = encoded }
        });
        string projected = DiagnosticMedia.Project(source);
        Check(!projected.Contains(encoded), "Diagnostic retained inline payload");
        Check(projected.Contains(hash) && projected.Contains("decoded_bytes=9000"), "Diagnostic media metadata is incorrect");
        var parsed = JsonNode.Parse(projected)!;
        Check(parsed["text"]!.GetValue<string>() == "preserve ordinary caption", "Projection changed caption");
        Check(parsed["data"]!.GetValue<string>() == "ordinary data field", "Projection removed unrelated data");
        Check(JsonNode.Parse(source)!["source"]!["data"]!.GetValue<string>() == encoded, "Projection changed source request");
        string malformed = DiagnosticMedia.Project("{\"image_url\":\"data:image/png;base64,%%%private-payload\"}");
        Check(!malformed.Contains("private-payload") && malformed.Contains("unavailable"), "Malformed media leaked");
        Check(!DiagnosticMedia.Project("{\"source\":{\"media_type\":\"image/png\",\"data\":{\"private\":\"bytes\"}}}").Contains("private"),
            "Non-string media leaked");
        Check(DiagnosticMedia.Project("{broken") == "[diagnostic structured payload omitted: invalid JSON]", "Malformed structured trace was dumped raw");
        string duplicate = "{\"source\":{},\"source\":{}}";
        Check(DiagnosticMedia.Project(duplicate).Contains("payload omitted"), "Ambiguous structured trace was not safely omitted");
        string explicitMime = DiagnosticMedia.Project("{\"audio\":{\"mime_type\":\"audio/mpeg\",\"data\":\"AA==\"}}");
        Check(explicitMime.Contains("mime=audio/mpeg"), "Explicit audio MIME was replaced");
        string rootUrl = JsonSerializer.Serialize("data:image/png;base64," + encoded);
        Check(JsonSerializer.Deserialize<string>(DiagnosticMedia.Project(rootUrl))!.Contains("media omitted"), "Root JSON media string leaked");
        Check(DiagnosticMedia.Project("[DONE]") == "[DONE]", "SSE completion marker changed");
        return Task.CompletedTask;
    }

    public static Task TraceCopies()
    {
        string directory = Path.Combine(Path.GetTempPath(), "diagnostic-media-" + Guid.NewGuid().ToString("N"));
        Directory.CreateDirectory(directory);
        using var agentTrace = new StringWriter();
        using var hostTrace = new StringWriter();
        var root = new Ensemble.Ensemble(TextWriter.Null, hostTrace, dataDirectory: directory);
        try
        {
            var agent = new Agent(root, new Configuration(Vendor.OpenAI, "gpt-4.1", "http://127.0.0.1:1",
                "synthetic-media-key", sandboxRoot: directory));
            agent.AttachDiagnostics(TextWriter.Null, agentTrace);
            string encoded = Convert.ToBase64String(Enumerable.Range(0, 100).Select(index => (byte)index).ToArray());
            string payload = JsonSerializer.Serialize(new { text = "synthetic-media-key", image_url = "data:image/png;base64," + encoded });
            string originalPayload = payload;
            int history = agent.History().View().Dialogue.Count();
            agent.Trace("request", payload);
            root.Trace(agent, "request", payload);
            foreach (string log in new[] { agentTrace.ToString(), hostTrace.ToString() })
            {
                string diagnosticPayload = JsonNode.Parse(log.Trim())!["payload"]!.GetValue<string>();
                string imageUrl = JsonNode.Parse(diagnosticPayload)!["image_url"]!.GetValue<string>();
                Check(!imageUrl.Contains(encoded), "Trace entry point retained media bytes");
                Check(!log.Contains("synthetic-media-key"), "Media projection bypassed credential masking");
                Check(log.Contains("media omitted"), "Trace lost explicit media marker");
            }
            Check(payload == originalPayload && JsonNode.Parse(payload)!["image_url"]!.GetValue<string>() == "data:image/png;base64," + encoded,
                "Trace altered provider input");
            Check(agent.History().View().Dialogue.Count() == history, "Diagnostic projection modified History");
        }
        finally { root.Destroy(); Directory.Delete(directory, true); }
        agentTrace.WriteLine("still borrowed");
        return Task.CompletedTask;
    }

    private static void Check(bool valid, string message)
    {
        if (!valid) throw new Exception(message);
    }
}
