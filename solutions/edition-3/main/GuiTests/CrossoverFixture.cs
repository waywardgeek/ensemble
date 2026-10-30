using System.Net;
using System.Text.Json;
using Ensemble;
using EnsembleGui;

// A public-API consumer used by crossover-browser.cjs. Only the external model
// boundary is fake. The History, actors, Jobs, host, sockets and assets are real.
internal static class CrossoverFixture
{
    internal static async Task RunAsync()
    {
        string originalDirectory = Directory.GetCurrentDirectory();
        string directory = Path.Combine(originalDirectory, ".crossover", "runtime");
        Directory.CreateDirectory(directory);
        Directory.SetCurrentDirectory(directory);
        string stopFile = Path.Combine(directory, "stop");
        File.Delete(stopFile);
        using var deadline = new CancellationTokenSource(TimeSpan.FromMinutes(15));
        var token = deadline.Token;
        var root = new Ensemble.Ensemble(TextWriter.Null, dataDirectory: Path.Combine(directory, "data"));
        var providerBuilder = WebApplication.CreateBuilder();
        providerBuilder.Logging.ClearProviders();
        var provider = providerBuilder.Build();
        provider.Urls.Add("http://127.0.0.1:18203");
        string key = Guid.NewGuid().ToString("N");
        provider.MapPost("/v1/messages", async (HttpContext context) =>
        {
            // Validate the boundary instead of accepting an arbitrary scripted
            // internal call. The generated key has no external authority.
            using var request = await JsonDocument.ParseAsync(context.Request.Body,
                cancellationToken: context.RequestAborted);
            if (context.Request.Headers["x-api-key"] != key
                || !request.RootElement.GetProperty("stream").GetBoolean()
                || request.RootElement.GetProperty("messages").GetArrayLength() == 0)
            {
                context.Response.StatusCode = 400;
                return;
            }
            context.Response.ContentType = "text/event-stream";
            async Task Frame(object frame)
            {
                await context.Response.WriteAsync("data: " + JsonSerializer.Serialize(frame) + "\n\n", context.RequestAborted);
                await context.Response.Body.FlushAsync(context.RequestAborted);
            }
            try
            {
                await Frame(new { type = "message_start", message = new { model = "claude-3-haiku-20240307", usage = new { input_tokens = 31, output_tokens = 0 } } });
                await Frame(new { type = "content_block_start", index = 0, content_block = new { type = "text", text = "" } });
                await Frame(new { type = "content_block_delta", index = 0, delta = new { type = "text_delta", text = "STREAM-IDENTITY first " } });
                await Task.Delay(2200, context.RequestAborted);
                await Frame(new { type = "content_block_delta", index = 0, delta = new { type = "text_delta", text = "and final." } });
                await Frame(new { type = "content_block_stop", index = 0 });
                await Frame(new { type = "message_delta", delta = new { stop_reason = "end_turn" }, usage = new { output_tokens = 7 } });
                await Frame(new { type = "message_stop" });
            }
            catch (OperationCanceledException)
            {
                // Interrupt closes the real HTTP request. There is no successful
                // final response to record for an aborted stream.
            }
        });
        var hostBuilder = WebApplication.CreateBuilder();
        hostBuilder.Logging.ClearProviders();
        var host = hostBuilder.Build();
        host.Urls.Add("http://127.0.0.1:18202");
        try
        {
            await provider.StartAsync(token);
            // The browser first switches this launch model to the catalogued
            // Haiku model, exercising an actual selection before inference.
            var configuration = new Configuration(Vendor.Anthropic, "crossover-initial",
                "http://127.0.0.1:18203", key);
            var selected = new Agent(root, configuration, name: "captain", outputDirectory: Path.Combine(directory, "jobs"));
            selected.InstallTools();
            var child = new Agent(root, selected, "worker", configuration);
            _ = new Agent(root, child, "nested", configuration);
            var retained = new Agent(root, selected, "retained", configuration);
            retained.PublishSubmission("Completed fixture work; retain this child for inspection.");
            var retainedState = await retained.WorkStatusAsync(token);
            if (!retainedState.Submitted || retainedState.Activity != AgentActivity.Idle
                || !ReferenceEquals(selected.FindChild("retained"), retained))
                throw new InvalidOperationException("Submission must retain an idle, queryable child.");
            var peer = new Agent(root, configuration, name: "peer");
            peer.History().Append(new Event(EventKind.MessageReceived, actor: Actor.Human,
                parts: [new TextPart("PEER-ONLY history") ]));
            string code = string.Join('\n', Enumerable.Range(1, 65).Select(i => $"const value{i} = {i}; // teaching code"));
            selected.History().Append(new Event(EventKind.MessageReceived, actor: Actor.Human,
                parts: [new TextPart("Please examine the fixture output.")]));
            selected.History().Append(new Event(EventKind.ResponseEnded, actor: Actor.Agent,
                parts: [new TextPart("## Offline crossover fixture\nMarkdown **text** and `inline code`.\n```javascript\n" + code + "\n```"),
                    new OpaquePart(configuration.Target(), JsonSerializer.SerializeToElement(new { type = "thinking", thinking = "Labelled reasoning, not another answer.", signature = "fixture" }))]));
            // Dispatch a real shell Job through the registered built-in tool. It
            // writes a large file; only the normal bounded preview enters History.
            var call = new ToolCallPart("large-output", configuration.Target(), "run_command",
                JsonSerializer.SerializeToElement(new { command = "i=0; while [ $i -lt 2000 ]; do printf 'fixture-line-%04d complete-output\\n' \"$i\"; i=$((i+1)); done", max_output_bytes = 1000 }));
            selected.History().Append(new Event(EventKind.ResponseEnded, actor: Actor.Agent, parts: [call]));
            var result = await selected.DispatchAsync(call, token);
            if (result.IsError)
            {
                throw new InvalidOperationException(JsonCodec.Text(result.Parts));
            }
            int handle = selected.Jobs().Children().Single().Handle;
            // Real running output and a genuinely nested owner exercise both
            // routes without granting the parent access by handle alone.
            await selected.DispatchAsync(new ToolCallPart("running-job", configuration.Target(), "run_command",
                JsonSerializer.SerializeToElement(new { command = "printf running-start; sleep 300; printf running-end", ai_callback_delay = 0.1 })), token);
            int runningHandle = selected.Jobs().Children().Max(job => job.Handle);
            child.InstallTools();
            await child.DispatchAsync(new ToolCallPart("nested-job", configuration.Target(), "run_command",
                JsonSerializer.SerializeToElement(new { command = "printf nested-owned-output" })), token);
            int nestedHandle = child.Jobs().Children().Single().Handle;
            await selected.DispatchAsync(new ToolCallPart("oversized-job", configuration.Target(), "run_command",
                JsonSerializer.SerializeToElement(new { command = "head -c 17825792 /dev/zero", max_output_bytes = 128 })), token);
            int oversizedHandle = selected.Jobs().Children().Max(job => job.Handle);
            await selected.DispatchAsync(new ToolCallPart("cursor-job", configuration.Target(), "run_command",
                JsonSerializer.SerializeToElement(new { command = "printf observed-prefix; sleep 1; printf unobserved-suffix", ai_callback_delay = 0.1 })), token);
            int cursorHandle = selected.Jobs().Children().Max(job => job.Handle);
            while (selected.Jobs().Find(cursorHandle).Snapshot().Status == "running")
            {
                await Task.Delay(20, token);
            }
            using var completeStream = selected.Sandbox().OpenRegularRead(selected.Jobs().Find(handle).Snapshot().OutputPath);
            string outputHash = Convert.ToHexString(System.Security.Cryptography.SHA256.HashData(completeStream));
            const string pixel = "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mP8/x8AAwMCAO+aZ1cAAAAASUVORK5CYII=";
            selected.History().Append(new Event(EventKind.MessageReceived, actor: Actor.Tool,
                parts: [new BlobPart("image/png", new BlobRef(RefKind.Uri, "data:image/png;base64," + pixel))]));
            // A dead owner's handle must not remain retrievable just because its
            // durable output file still exists.
            // Separate fixture Agents keep failure files out of the primary
            // transcript. Their Jobs are real; only the backing filesystem is
            // deliberately damaged after completion.
            var fileFailures = new Agent(root, configuration, name: "files");
            fileFailures.InstallTools();
            foreach (string kind in new[] { "unreadable", "fifo", "missing" })
            {
                await fileFailures.DispatchAsync(new ToolCallPart(kind, configuration.Target(), "run_command",
                    JsonSerializer.SerializeToElement(new { command = "printf fixture-output" })), token);
                Job owned = fileFailures.Jobs().Children().MaxBy(job => job.Handle)!;
                string outputPath = owned.Snapshot().OutputPath;
                if (kind == "unreadable")
                {
                    if (OperatingSystem.IsWindows())
                    {
                        throw new PlatformNotSupportedException("The crossover fixture requires Unix permissions and FIFOs.");
                    }
                    File.SetUnixFileMode(outputPath, UnixFileMode.None);
                }
                else
                {
                    File.Delete(outputPath);
                    if (kind == "fifo")
                    {
                        var start = new System.Diagnostics.ProcessStartInfo("/usr/bin/mkfifo");
                        start.ArgumentList.Add(outputPath);
                        using var process = System.Diagnostics.Process.Start(start)!;
                        await process.WaitForExitAsync(token);
                        if (process.ExitCode != 0)
                        {
                            throw new InvalidOperationException("Fixture FIFO creation failed");
                        }
                    }
                }
            }
            var bulk = new Agent(root, configuration, name: "bulk");
            bulk.InstallTools();
            await bulk.DispatchAsync(new ToolCallPart("bulk-job", configuration.Target(), "run_command",
                JsonSerializer.SerializeToElement(new { command = "head -c 8388608 /dev/zero", max_output_bytes = 128, ai_callback_delay = 1 })), token);
            int bulkHandle = bulk.Jobs().Children().Single().Handle;
            string confinedRoot = Path.Combine(root.DataDirectory(), "agents", "confined");
            Directory.CreateDirectory(confinedRoot);
            var confinedConfiguration = new Configuration(Vendor.Anthropic, "claude-3-haiku-20240307",
                "http://127.0.0.1:18203", key, sandboxRoot: confinedRoot);
            var confined = new Agent(root, confinedConfiguration, Path.Combine(confinedRoot, "jobs"), name: "confined");
            confined.InstallTools();
            await confined.DispatchAsync(new ToolCallPart("confined-job", confinedConfiguration.Target(), "run_command",
                JsonSerializer.SerializeToElement(new { command = "printf confined-output" })), token);
            Job confinedJob = confined.Jobs().Children().Single();
            string outside = Path.Combine(directory, "outside-canary");
            File.WriteAllText(outside, "must-not-cross-root");
            File.Delete(confinedJob.Snapshot().OutputPath);
            File.CreateSymbolicLink(confinedJob.Snapshot().OutputPath, outside);
            var removed = new Agent(root, configuration, name: "removed");
            removed.InstallTools();
            await removed.DispatchAsync(new ToolCallPart("removed-job", configuration.Target(), "run_command",
                JsonSerializer.SerializeToElement(new { command = "printf removed-output" })), token);
            int removedHandle = removed.Jobs().Children().Single().Handle;
            removed.Destroy();
            host.MapGet("/fixture/observe", async () =>
            {
                var observed = await selected.Jobs().Find(cursorHandle).ObserveAsync(
                    "after-download", delay: 0, maxBytes: 128000, cancellation: token);
                string text = string.Concat(observed.Parts.OfType<TextPart>().Select(part => part.Text));
                return Results.Text(text);
            });
            Task mounted = Gui.ServeAsync(host, selected, new GuiOptions(18202,
                SettingsPath: Path.Combine(directory, "appearance.json"),
                TracePath: Path.Combine(directory, "gui.log"), BasePath: "/assistant"), token);
            await host.StartAsync(token);
            var statusResult = await selected.DispatchAsync(new ToolCallPart("retained-status",
                configuration.Target(), "agents_status", JsonSerializer.SerializeToElement(new { })), token);
            if (statusResult.IsError) throw new InvalidOperationException("Child status tool refused.");
            using (var statusJson = JsonDocument.Parse(string.Concat(statusResult.Parts.OfType<TextPart>().Select(part => part.Text))))
            {
                if (!statusJson.RootElement.GetProperty("agents").EnumerateArray().Any(row =>
                    row.GetProperty("agent_id").GetString() == "retained"
                    && row.GetProperty("submitted").GetBoolean()
                    && row.GetProperty("activity").GetString() == "idle"))
                    throw new InvalidOperationException("Submitted child missing from agents_status.");
            }
            Console.WriteLine("PASS submitted child remains idle, queryable, owned, and listed in agents_status");
            File.WriteAllText(Path.Combine(directory, "fixture.json"), JsonSerializer.Serialize(new { handle, runningHandle, nestedHandle, oversizedHandle, cursorHandle, bulkHandle, removedHandle, outputHash, retainedArtifact = Path.Combine(retained.DataDirectory(), "submit.md") }));
            Console.WriteLine("READY http://127.0.0.1:18202/assistant/ (embedded assets; loopback provider :18203)");
            while (!File.Exists(stopFile))
            {
                if (mounted.IsCompleted)
                {
                    await mounted;
                }
                await Task.Delay(100, token);
            }
            deadline.Cancel();
            await mounted;
        }
        finally
        {
            deadline.Cancel();
            await host.StopAsync();
            await provider.StopAsync();
            await host.DisposeAsync();
            await provider.DisposeAsync();
            root.Destroy();
            Directory.SetCurrentDirectory(originalDirectory);
        }
        Console.WriteLine("STOPPED crossover fixture, provider and owned Agents");
    }
}
