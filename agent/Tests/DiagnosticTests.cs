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

    public static async Task AgentFiles()
    {
        string directory = Path.Combine(Path.GetTempPath(), "ensemble-agent-logs-" + Guid.NewGuid().ToString("N"));
        Directory.CreateDirectory(directory);
        try
        {
            using var host = new StringWriter();
            var app = new Application(host, dataDirectory: directory, agentDiagnostics: true);
            var config = new Configuration(Vendor.OpenAI, "gpt-4.1", "http://127.0.0.1:1",
                "fixture-log-secret", sandboxRoot: directory);
            try
            {
                var first = new Agent(app, config, name: "first");
                var second = new Agent(app, config, name: "second");
                first.StartDiagnostics();
                second.StartDiagnostics();
                first.Log("first-debug fixture-log-secret");
                second.Log("second-debug");
                first.Trace("request", "first-api");
                second.Trace("request", "second-api");
                first.TraceGui("<", "first-gui");
                second.TraceGui("<", "second-gui");
                Agent child = await first.CreateChildAsync("child");
                Agent grandchild = await child.CreateChildAsync("grandchild");
                foreach (Agent descendant in new[] { child, grandchild })
                {
                    string marker = descendant.ChildID()!;
                    descendant.Log(marker + "-debug");
                    descendant.Trace("request", marker + "-api");
                    descendant.TraceGui("<", marker + "-gui");
                    foreach (string name in new[] { "debug", "api", "gui" })
                    {
                        string text = File.ReadAllText(Path.Combine(descendant.DataDirectory(), name + ".log"));
                        Check(text.Contains(marker + "-" + name), "Child did not receive owned diagnostic policy");
                        string parentText = File.ReadAllText(Path.Combine(first.DataDirectory(), name + ".log"));
                        Check(!parentText.Contains(marker + "-" + name), "Child diagnostics leaked into parent");
                    }
                }
                app.Log("host-only");
                foreach (var pair in new[] { (first, "first", "second"), (second, "second", "first") })
                {
                    foreach (string name in new[] { "debug", "api", "gui" })
                    {
                        string text = File.ReadAllText(Path.Combine(pair.Item1.DataDirectory(), name + ".log"));
                        Check(text.Contains(pair.Item2 + "-" + name) && !text.Contains(pair.Item3 + "-" + name),
                            "Agent diagnostic destination leaked");
                        Check(!text.Contains("fixture-log-secret"), "Agent diagnostics exposed credential");
                    }
                }
                Check(host.ToString().Contains("host-only") && !host.ToString().Contains("first-"), "Host logger mixed Agent work");
                string home = first.DataDirectory();
                File.WriteAllText(Path.Combine(home, "durable-sentinel"), "preserve");
                first.Destroy();
                var restored = new Agent(app, config, name: "first");
                restored.StartDiagnostics();
                foreach (string name in new[] { "debug", "api", "gui" })
                    Check(new FileInfo(Path.Combine(home, name + ".log")).Length == 0, "Restart appended diagnostics");
                Check(File.ReadAllText(Path.Combine(home, "durable-sentinel")) == "preserve", "Diagnostics changed durable data");
                var broken = new Agent(app, config, name: "broken-startup");
                string brokenHome = broken.DataDirectory();
                string previousApi = Path.Combine(brokenHome, "api.log");
                File.WriteAllText(previousApi, "previous-session-evidence");
                string brokenSave = Path.Combine(brokenHome, "save.json");
                File.WriteAllText(brokenSave, "{invalid");
                bool refused = false;
                try
                {
                    // The host order is intentional: never open fresh Agent
                    // files before this durable-state admission succeeds.
                    broken.StartPersistence(new PersistenceOptions(brokenSave));
                    broken.StartDiagnostics();
                }
                catch (Exception error) when (error is InvalidDataException or JsonException)
                {
                    refused = true;
                }
                Check(refused, "Invalid startup state was accepted");
                Check(File.ReadAllText(previousApi) == "previous-session-evidence",
                    "Failed restoration erased previous diagnostics");
            }
            finally { app.Destroy(); }
            // Borrowed host writers remain usable after the application closes.
            host.WriteLine("caller-still-owns-writer");
            string hostPath = Path.Combine(directory, "ensemble.log");
            var firstHost = new Application(TextWriter.Null, dataDirectory: directory);
            try
            {
                firstHost.StartDiagnostics();
                firstHost.Log("host-first-session");
                Check(File.ReadAllText(hostPath).Contains("host-first-session"), "Owned host log was not written");
            }
            finally { firstHost.Destroy(); }
            var nextHost = new Application(TextWriter.Null, dataDirectory: directory);
            try
            {
                nextHost.StartDiagnostics();
                Check(new FileInfo(hostPath).Length == 0, "Host restart appended diagnostics");
                nextHost.Log("host-next-session");
            }
            finally { nextHost.Destroy(); }
            Check(File.ReadAllText(hostPath).Contains("host-next-session"), "Host log did not close cleanly");
        }
        finally { Directory.Delete(directory, true); }
    }

    private static void Check(bool condition, string message)
    {
        if (!condition)
        {
            throw new Exception(message);
        }
    }
}
