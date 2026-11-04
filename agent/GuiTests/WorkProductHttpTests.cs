using System.Net;
using System.Net.Sockets;
using Ensemble;
using EnsembleGui;

internal static class WorkProductHttpTests
{
    private static void Check(bool value, string message)
    {
        if (!value) throw new InvalidOperationException(message);
    }

    private sealed class SlowCopy
    {
        public readonly TaskCompletionSource Entered = new(TaskCreationOptions.RunContinuationsAsynchronously);
        public readonly TaskCompletionSource Cancelled = new(TaskCreationOptions.RunContinuationsAsynchronously);
        public readonly TaskCompletionSource Release = new(TaskCreationOptions.RunContinuationsAsynchronously);
        public readonly TaskCompletionSource Done = new(TaskCreationOptions.RunContinuationsAsynchronously);
    }
    // Only the external response sink is slowed. Production routing, linked
    // cancellation, file access and request draining all remain in the test.
    private sealed class SlowBody(Stream inner, SlowCopy gate) : Stream
    {
        public override async ValueTask WriteAsync(ReadOnlyMemory<byte> bytes, CancellationToken cancellationToken = default)
        {
            await inner.WriteAsync(bytes, cancellationToken);
            await inner.FlushAsync(cancellationToken);
            gate.Entered.TrySetResult();
            try { await Task.Delay(Timeout.Infinite, cancellationToken); }
            catch (OperationCanceledException)
            {
                gate.Cancelled.TrySetResult();
                await gate.Release.Task;
                throw;
            }
        }
        public override bool CanRead => false;
        public override bool CanSeek => false;
        public override bool CanWrite => true;
        public override long Length => throw new NotSupportedException();
        public override long Position { get => throw new NotSupportedException(); set => throw new NotSupportedException(); }
        public override void Flush() => inner.Flush();
        public override int Read(byte[] buffer, int offset, int count) => throw new NotSupportedException();
        public override long Seek(long offset, SeekOrigin origin) => throw new NotSupportedException();
        public override void SetLength(long value) => throw new NotSupportedException();
        public override void Write(byte[] buffer, int offset, int count) => throw new NotSupportedException();
    }

    public static async Task RunAsync()
    {
        string work = Path.Combine(Path.GetTempPath(), "work-product-http-" + Guid.NewGuid().ToString("N"));
        Directory.CreateDirectory(work);
        var root = new Ensemble.Ensemble(TextWriter.Null);
        var config = new Configuration(Vendor.Anthropic, "offline", "http://127.0.0.1:1", sandboxRoot: work);
        var source = new Agent(root, config, name: "source");
        var owner = new Agent(root, config, name: "products");
        string file = Path.Combine(work, "report.bin");
        byte[] original = [1, 2, 3];
        File.WriteAllBytes(file, original);
        var offer = owner.OfferWorkProduct(file);
        using var listener = new TcpListener(IPAddress.Loopback, 0);
        listener.Start();
        int port = ((IPEndPoint)listener.LocalEndpoint).Port;
        listener.Stop();
        string local = $"http://127.0.0.1:{port}";
        var builder = WebApplication.CreateBuilder();
        builder.Logging.ClearProviders();
        builder.WebHost.UseUrls(local);
        await using var host = builder.Build();
        SlowCopy? slow = null;
        host.Use(async (HttpContext context, RequestDelegate next) =>
        {
            var gate = slow;
            if (context.Request.Headers["X-Test-Slow"] != "1" || gate is null)
            {
                await next(context);
                return;
            }
            var originalBody = context.Response.Body;
            context.Response.Body = new SlowBody(originalBody, gate);
            try { await next(context); }
            finally
            {
                context.Response.Body = originalBody;
                gate.Done.TrySetResult();
            }
        });
        using var stop = new CancellationTokenSource();
        using var deadline = new CancellationTokenSource(TimeSpan.FromSeconds(60));
        using var http = new HttpClient(new SocketsHttpHandler { UseProxy = false, MaxResponseDrainSize = 0 });
        var serving = Gui.ServeAsync(host, source, new GuiOptions(port,
            SettingsPath: Path.Combine(work, "settings.json"), BasePath: "/assistant"), stop.Token);
        string path = "/assistant/work-products/download?agent=products&offer=" + offer.ID;
        async Task<HttpResponseMessage> Get(string suffix, bool guard = true, string? origin = null, string? site = null, bool slowResponse = false)
        {
            using var request = new HttpRequestMessage(HttpMethod.Get, local + suffix);
            if (guard) request.Headers.Add("X-Ensemble-Work-Product", "1");
            if (slowResponse) request.Headers.Add("X-Test-Slow", "1");
            if (origin is not null) request.Headers.Add("Origin", origin);
            if (site is not null) request.Headers.Add("Sec-Fetch-Site", site);
            return await http.SendAsync(request, HttpCompletionOption.ResponseHeadersRead, deadline.Token);
        }
        async Task Expect(string suffix, HttpStatusCode status, bool guard = true, string? origin = null, string? site = null)
        {
            using var response = await Get(suffix, guard, origin, site);
            Check(response.StatusCode == status, $"Expected {status}, got {response.StatusCode}");
        }
        try
        {
            await host.StartAsync(deadline.Token);
            while (true)
            {
                if (serving.IsCompleted) await serving;
                try
                {
                    using var ready = await http.GetAsync(local + "/assistant/", deadline.Token);
                    if (ready.IsSuccessStatusCode) break;
                }
                catch (HttpRequestException) { }
                await Task.Delay(20, deadline.Token);
            }
            await Expect(path, HttpStatusCode.Forbidden, guard: false);
            await Expect(path, HttpStatusCode.Forbidden, origin: "https://other.example");
            await Expect(path, HttpStatusCode.Forbidden, site: "cross-site");
            await Expect(path + "&path=report.bin", HttpStatusCode.BadRequest);
            await Expect(path.Replace("agent=products", "agent=source"), HttpStatusCode.NotFound);
            File.WriteAllBytes(file, [9, 8, 7, 6]);
            using (var response = await Get(path))
            {
                Check(response.StatusCode == HttpStatusCode.OK, "Owned offer refused");
                Check((await response.Content.ReadAsByteArrayAsync(deadline.Token)).SequenceEqual(new byte[] { 9, 8, 7, 6 }),
                    "Transfer did not use click-time bytes");
                Check(response.Headers.CacheControl?.NoStore == true, "Download was cacheable");
                Check(response.Content.Headers.ContentType?.MediaType == "application/octet-stream", "Download became inline content");
                Check(response.Content.Headers.ContentDisposition?.DispositionType == "attachment", "Missing download-only disposition");
                Check(response.Headers.GetValues("X-Content-Type-Options").Single() == "nosniff", "Missing nosniff");
            }
            using (var fileStream = File.OpenWrite(file)) fileStream.SetLength(Agent.MaximumWorkProductBytes);
            using (var response = await Get(path))
            {
                Check(response.StatusCode == HttpStatusCode.OK, "Exact 64 MiB HTTP transfer refused");
                await using var content = await response.Content.ReadAsStreamAsync(deadline.Token);
                long total = 0;
                byte[] buffer = new byte[64 * 1024];
                int count;
                while ((count = await content.ReadAsync(buffer, deadline.Token)) != 0) total += count;
                Check(total == Agent.MaximumWorkProductBytes, "64 MiB transfer was truncated");
            }
            using (var fileStream = File.OpenWrite(file)) fileStream.SetLength(Agent.MaximumWorkProductBytes + 1);
            await Expect(path, HttpStatusCode.NotFound);
            File.Delete(file);
            await Expect(path, HttpStatusCode.NotFound);
            File.WriteAllBytes(file, original);
            slow = new SlowCopy();
            using (var interrupted = await Get(path, slowResponse: true))
            {
                await slow.Entered.Task.WaitAsync(deadline.Token);
                interrupted.Dispose(); // Real HTTP client abort; no background drain.
                await slow.Cancelled.Task.WaitAsync(deadline.Token);
                Check(!slow.Done.Task.IsCompleted, "HTTP request finished before cancellation cleanup");
                slow.Release.TrySetResult();
                await slow.Done.Task.WaitAsync(deadline.Token);
            }
            Console.WriteLine("PASS real HTTP client abort reaches owned copy and joins cancellation cleanup");
            owner.Destroy();
            var replacement = new Agent(root, config, name: "products");
            await Expect(path, HttpStatusCode.NotFound);
            var nextOffer = replacement.OfferWorkProduct(file);
            slow = new SlowCopy();
            using (var interrupted = await Get(path.Replace(offer.ID, nextOffer.ID), slowResponse: true))
            {
                await slow.Entered.Task.WaitAsync(deadline.Token);
                stop.Cancel();
                await slow.Cancelled.Task.WaitAsync(deadline.Token);
                Check(!serving.IsCompleted && !slow.Done.Task.IsCompleted,
                    "Mount stopped before owned HTTP copy cleanup");
                slow.Release.TrySetResult();
                await slow.Done.Task.WaitAsync(deadline.Token);
                await serving.WaitAsync(deadline.Token);
            }
            Console.WriteLine("PASS mount stop cancels and joins actual HTTP transfer before completion");
            Console.WriteLine("PASS HTTP click-time bytes, guarded origin/header, no browser paths, live-owner expiry and exact 64 MiB bound");
        }
        finally
        {
            slow?.Release.TrySetResult();
            stop.Cancel();
            await serving;
            await host.StopAsync(CancellationToken.None);
            root.Destroy();
            Directory.Delete(work, true);
        }
    }
}
