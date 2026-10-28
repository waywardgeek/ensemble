using System.Text.Json;
using Ensemble;
using Application = Ensemble.Ensemble;

internal static class DiagnosticTests
{
    public static async Task KnownKeys()
    {
        string directory = Path.Combine(Path.GetTempPath(), "ensemble-diagnostics-" + Guid.NewGuid().ToString("N"));
        Directory.CreateDirectory(directory);
        string shortKey = "fixture-secret";
        string longKey = shortKey + "-longer";
        string escapedKey = "fixture-quoted-\"\\key";
        using var diagnostics = new StreamWriter(Path.Combine(directory, "diagnostics.log")) { AutoFlush = true };
        using var trace = new StreamWriter(Path.Combine(directory, "api.log")) { AutoFlush = true };
        var app = new Application(diagnostics, trace);
        try
        {
            var agent = new Agent(app, new Configuration(Vendor.OpenAI, "gpt-4.1", "http://127.0.0.1:1",
                shortKey, sandboxRoot: directory), name: shortKey);
            _ = new Agent(app, new Configuration(Vendor.OpenAI, "gpt-4.1", "http://127.0.0.1:1",
                longKey, sandboxRoot: directory));
            _ = new Agent(app, new Configuration(Vendor.OpenAI, "gpt-4.1", "http://127.0.0.1:1",
                escapedKey, sandboxRoot: directory));
            string message = "request-id=local-17 Authorization: Bearer " + longKey + " other=" + shortKey
                + " quoted=" + escapedKey + " json=" + JsonSerializer.Serialize(escapedKey)
                + " standard-json=" + JsonEncodedText.Encode(escapedKey, System.Text.Encodings.Web.JavaScriptEncoder.UnsafeRelaxedJsonEscaping);
            agent.InstallTools();
            app.Log(message);
            app.Trace(agent, "response " + longKey, message);
            Check(agent.Name == shortKey, "Diagnostic filtering changed actual Agent identity");
            await Task.WhenAll(Enumerable.Range(0, 8).Select(index => Task.Run(() =>
            {
                var transient = new Agent(app, new Configuration(Vendor.OpenAI, "gpt-4.1",
                    "http://127.0.0.1:1", "transient-secret-" + index, sandboxRoot: directory));
                try
                {
                    app.Log("concurrent " + transient.Configuration().ApiKey);
                    app.Trace(transient, "response", message);
                }
                finally
                {
                    transient.Destroy();
                }
            })));

            // A root snapshot can outlive detach. Its diagnostic read must stay
            // valid without making the public disposed-object API permissive.
            var retiring = new Agent(app, new Configuration(Vendor.OpenAI, "gpt-4.1",
                "http://127.0.0.1:1", "retiring-key", sandboxRoot: directory));
            var captured = app.Agents().Single(value => ReferenceEquals(value, retiring));
            retiring.Destroy();
            Check(captured.DiagnosticKey() == "retiring-key", "Snapshot lost authoritative diagnostic key during detach");
            bool disposed = false;
            try
            {
                captured.Configuration();
            }
            catch (ObjectDisposedException)
            {
                disposed = true;
            }
            Check(disposed, "Diagnostic read weakened public destruction contract");

            foreach (string path in new[] { "diagnostics.log", "api.log" })
            {
                // Read the actual artifact through the ordinary model-facing tool.
                var result = await agent.DispatchAsync(new ToolCallPart(Guid.NewGuid().ToString("N"),
                    agent.Configuration().Target(), "read_file", JsonSerializer.SerializeToElement(new { path })));
                string text = JsonCodec.Text(result.Parts);
                Check(!result.IsError && text.Contains("request-id=local-17"), "Readable diagnostic context lost");
                Check(text.Contains("[REDACTED]"), "Whole-value replacement absent");
                Check(!text.Contains(shortKey) && !text.Contains("-longer") && !text.Contains("fixture-quoted") && !text.Contains("transient-secret"),
                    "Credential or overlapping credential suffix reached readable log");
            }
        }
        finally
        {
            app.Destroy();
            diagnostics.Dispose();
            trace.Dispose();
            Directory.Delete(directory, true);
        }
    }

    private static void Check(bool condition, string message)
    {
        if (!condition)
        {
            throw new Exception(message);
        }
    }
}
