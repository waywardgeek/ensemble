using System.Net;
using System.Net.Sockets;
using System.Text;
using System.Text.Json;
using Ensemble;
using EnsembleGui;

internal static class JobWindowTests
{
    private static void Check(bool value, string message)
    {
        if (!value) throw new InvalidOperationException(message);
    }

    public static async Task RunAsync()
    {
        string work = Path.Combine(Path.GetTempPath(), "job-window-" + Guid.NewGuid().ToString("N"));
        Directory.CreateDirectory(work);
        var root = new Ensemble.Ensemble(TextWriter.Null);
        var config = new Configuration(Vendor.Anthropic, "offline", "http://127.0.0.1:1", sandboxRoot: work);
        var owner = new Agent(root, config, name: "owner");
        _ = new Agent(root, config, name: "other");
        owner.InstallTools();
        var release = new TaskCompletionSource(TaskCreationOptions.RunContinuationsAsynchronously);
        var published = new TaskCompletionSource(TaskCreationOptions.RunContinuationsAsynchronously);
        string text = "prefix" + new string('é', 40000) + "\u001b[31mTAIL";
        byte[] expected = Encoding.UTF8.GetBytes(text);
        _ = new Tool(owner, "held_output", "Offline window fixture",
            JsonSerializer.SerializeToElement(new { type = "object", properties = new { } }),
            async (job, _, token) =>
            {
                job.AppendOutput(text);
                published.SetResult();
                await release.Task.WaitAsync(token);
                return "END";
            });
        using var deadline = new CancellationTokenSource(TimeSpan.FromSeconds(45));
        using var stop = new CancellationTokenSource();
        using var listener = new TcpListener(IPAddress.Loopback, 0);
        listener.Start();
        int port = ((IPEndPoint)listener.LocalEndpoint).Port;
        listener.Stop();
        string local = $"http://127.0.0.1:{port}";
        var builder = WebApplication.CreateBuilder();
        builder.Logging.ClearProviders();
        builder.WebHost.UseUrls(local);
        await using var host = builder.Build();
        Task serving = Gui.ServeAsync(host, owner, new GuiOptions(port,
            SettingsPath: Path.Combine(work, "settings.json")), stop.Token);
        using var http = new HttpClient(new SocketsHttpHandler { UseProxy = false });
        async Task<HttpResponseMessage> Get(string path, bool guard = true, string? origin = null)
        {
            using var request = new HttpRequestMessage(HttpMethod.Get, local + path);
            if (guard) request.Headers.Add("X-Ensemble-Output", "1");
            if (origin is not null) request.Headers.Add("Origin", origin);
            return await http.SendAsync(request, deadline.Token);
        }
        try
        {
            await host.StartAsync(deadline.Token);
            while (true)
            {
                if (serving.IsCompleted) await serving;
                try { using var ready = await Get("/"); if (ready.IsSuccessStatusCode) break; }
                catch (HttpRequestException) { }
                await Task.Delay(20, deadline.Token);
            }
            var initial = await owner.DispatchAsync(new ToolCallPart("origin", owner.Configuration().Target(),
                "held_output", JsonSerializer.SerializeToElement(new { ai_callback_delay = 0.01, max_output_bytes = 32 })), deadline.Token);
            await published.Task.WaitAsync(deadline.Token);
            int handle = initial.Job!.Handle;
            // Publish after the model observation, before the human window read.
            // The next model wait must still receive these already-existing bytes.
            owner.Jobs().Find(handle).AppendOutput("UNOBSERVED");
            expected = Encoding.UTF8.GetBytes(text + "UNOBSERVED");
            string query = $"?agent=owner&handle={handle}";
            using (var response = await Get("/jobs/window" + query))
            {
                Check(response.StatusCode == HttpStatusCode.OK, "running window available");
                using var json = JsonDocument.Parse(await response.Content.ReadAsStringAsync(deadline.Token));
                var data = json.RootElement;
                byte[] bytes = Convert.FromBase64String(data.GetProperty("data").GetString()!);
                Check(bytes.Length == 65536 && bytes.AsSpan().SequenceEqual(expected.AsSpan(expected.Length - 65536)), "exact bounded byte tail");
                Check(data.GetProperty("start").GetInt64() == expected.Length - 65536
                    && data.GetProperty("end").GetInt64() == expected.Length && data.GetProperty("rolled_out").GetBoolean(), "explicit offsets and rolled-out notice");
                Check(data.GetProperty("origin_call_id").GetString() == "origin", "original call identity");
            }
            foreach (var refusal in new[]
            {
                ("/jobs/window?agent=other&handle=" + handle, true, (string?)null, HttpStatusCode.NotFound),
                ("/jobs/window" + query, false, (string?)null, HttpStatusCode.Forbidden),
                ("/jobs/window" + query, true, "https://foreign.invalid", HttpStatusCode.Forbidden),
                ("/jobs/window" + query + "&download=1", true, (string?)null, HttpStatusCode.BadRequest),
                ("/jobs/output" + query, true, (string?)null, HttpStatusCode.Conflict)
            })
            {
                using var response = await Get(refusal.Item1, refusal.Item2, refusal.Item3);
                Check(response.StatusCode == refusal.Item4, "window ownership/guard and download distinction");
            }
            release.SetResult();
            var final = await owner.DispatchAsync(new ToolCallPart("wait", owner.Configuration().Target(), "wait_for_job",
                JsonSerializer.SerializeToElement(new { handle, ai_callback_delay = 2, ai_callback_pattern = "END", max_output_bytes = 32 })), deadline.Token);
            Check(JsonCodec.Text(final.Parts).Contains("UNOBSERVED") && JsonCodec.Text(final.Parts).Contains("END"),
                "human window did not consume model cursor");
            Console.WriteLine("PASS Job window: bounded bytes, origin, ownership, guards, running download refusal, independent cursor");
        }
        finally
        {
            release.TrySetResult();
            stop.Cancel();
            try { await serving; }
            finally
            {
                try { await host.StopAsync(CancellationToken.None); }
                finally { root.Destroy(); }
            }
        }
    }
}
