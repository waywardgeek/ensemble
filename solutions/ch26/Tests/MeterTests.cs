using System.Net;
using System.Net.Sockets;
using System.Text;
using Ensemble;
using Application = Ensemble.Ensemble;

internal static partial class MeterTests
{
    private static void Check(bool condition, string reason)
    {
        if (!condition)
        {
            throw new InvalidOperationException(reason);
        }
    }

    private static HttpListener Listen()
    {
        var reservation = new TcpListener(IPAddress.Loopback, 0);
        reservation.Start();
        int port = ((IPEndPoint)reservation.LocalEndpoint).Port;
        reservation.Stop();
        var listener = new HttpListener();
        listener.Prefixes.Add($"http://127.0.0.1:{port}/");
        listener.Start();
        return listener;
    }

    private static async Task Reply(HttpListenerContext request, string body,
        int status = 200, string type = "application/json")
    {
        using var reader = new StreamReader(request.Request.InputStream);
        await reader.ReadToEndAsync();
        byte[] bytes = Encoding.UTF8.GetBytes(body);
        request.Response.StatusCode = status;
        request.Response.ContentType = type;
        request.Response.ContentLength64 = bytes.Length;
        await request.Response.OutputStream.WriteAsync(bytes);
        request.Response.Close();
    }

    private static string Response(int input, int read, int output) =>
        "{\"content\":[],\"usage\":{\"input_tokens\":" + input
        + ",\"cache_read_input_tokens\":" + read + ",\"output_tokens\":" + output + "}}";

    private static MeterSnapshot Snapshot(Usage usage) =>
        new(null, [new MeterEntry(new UsageIdentity(Vendor.OpenAI, "unknown", null), usage)]);

    public static void Arithmetic()
    {
        Check(new Usage(10, 20, 30, null).AllInput() == 60, "ALL includes each disjoint input bucket once");
        Check(new Usage(1, null, 2, 3).AllInput() is null, "unknown write is not zero");
        Check(new Usage(long.MaxValue, 1).AllInput() is null, "ALL overflow is unknown");
        var snapshot = Snapshot(new Usage(10, 20, 30, null));
        Check(snapshot.HitRate() == 0.5, "output uncertainty does not erase input hit rate");
        Check(Snapshot(new Usage()).HitRate() is null,
            "empty denominator does not fabricate a cache rate");
        Check(Snapshot(new Usage(null, 0, 3)).HitRate() is null,
            "unknown denominator remains unknown");
    }

    public static async Task Completion(string root)
    {
        using var listener = Listen();
        var app = new Application(TextWriter.Null);
        try
        {
            var agent = new Agent(app, new Configuration(Vendor.Anthropic, "unpriced", listener.Prefixes.Single()));
            var engine = agent.Engine();
            Check(engine.Meter().Latest is null && engine.Meter().Totals.AllInput() == 0, "new meter is empty");
            engine.Parse(200, Response(900, 800, 700));
            engine.Render(agent.History().View());
            Check(engine.Meter().Latest is null, "offline parse and render do not charge usage");

            var first = engine.SendAsync("{}");
            var firstRequest = await listener.GetContextAsync().WaitAsync(TimeSpan.FromSeconds(10));
            var second = engine.SendAsync("{}");
            var secondRequest = await listener.GetContextAsync().WaitAsync(TimeSpan.FromSeconds(10));
            await Reply(secondRequest, Response(20, 30, 4));
            await second;
            var earlier = engine.Meter();
            Check(earlier.Latest!.Counts.AllInput() == 50 && earlier.Totals.Output == 4, "completed second request counted first");
            await Reply(firstRequest, Response(10, 40, 6));
            await first;
            var meter = engine.Meter();
            Check(meter.Latest!.Counts.Input == 10 && meter.Totals.Input == 30 && meter.Totals.CacheRead == 70
                && meter.Totals.Output == 10 && meter.HitRate() == 0.7, "concurrent completions count once each");
            Check(earlier.Totals.Input == 20 && earlier.Latest.Counts.Input == 20, "older snapshot remains immutable");

            foreach (var failure in new[] { (503, "unavailable"), (200, "not-json") })
            {
                var send = engine.SendAsync("{}");
                await Reply(await listener.GetContextAsync().WaitAsync(TimeSpan.FromSeconds(10)), failure.Item2, failure.Item1);
                var result = await send;
                Check(result.Events.Any(fact => fact.Kind == EventKind.ErrorOccurred), "failure is real parsed failure");
                Check(ReferenceEquals(engine.Meter().Latest, meter.Latest), "failure does not fabricate usage");
            }
            using (var cancellation = new CancellationTokenSource())
            {
                var send = engine.SendAsync("{}", cancellation.Token);
                var held = await listener.GetContextAsync().WaitAsync(TimeSpan.FromSeconds(10));
                cancellation.Cancel();
                bool cancelled = false;
                try { await send; }
                catch (OperationCanceledException) { cancelled = true; }
                held.Response.Abort();
                Check(cancelled && ReferenceEquals(engine.Meter().Latest, meter.Latest), "cancelled transport leaves meter unchanged");
            }

            // A missing required usage field must poison only its own total.
            var missing = engine.SendAsync("{}");
            await Reply(await listener.GetContextAsync().WaitAsync(TimeSpan.FromSeconds(10)), "{\"content\":[],\"usage\":{\"output_tokens\":2}}");
            await missing;
            meter = engine.Meter();
            Check(meter.Latest!.Counts.Input is null && meter.Latest.Counts.CacheRead == 0 && meter.Latest.Counts.CacheWrite == 0,
                "missing ordinary usage differs from omitted optional cache buckets");
            Check(meter.Totals.Input is null && meter.Totals.Output == 12 && meter.Totals.CacheRead == 70,
                "unknown usage propagates by category without losing known counts");

            string save = Path.Combine(root, "meter-save.json");
            agent.History().Append(new Event(EventKind.ResponseEnded, usage: new Usage(999, 0, 100, 9)));
            agent.History().Save(new PersistenceOptions(save, RetainLog: 0));
            var restored = new Agent(app, new Configuration(Vendor.Anthropic, "unpriced", listener.Prefixes.Single()));
            Check(restored.History().LoadSave(save), "saved history restored");
            Check(restored.Engine().Meter().Latest is null && restored.Engine().Meter().Totals.AllInput() == 0,
                "restored usage never seeds a new process meter");
        }
        finally
        {
            app.Destroy();
        }
    }

    public static async Task Pricing()
    {
        var standard = new UsageIdentity(Vendor.OpenAI, "gpt-4o-mini", "default");
        Check(TokenPrices.EstimateUsd(standard, new Usage(1_000_000, 0, 1_000_000, 1_000_000)) == 0.825m,
            "published standard input, read and output rates applied independently");
        Check(TokenPrices.EstimateUsd(standard, new Usage(0, 1, 0, 0)) is null,
            "unverified nonzero cache-write price cannot become free");
        Check(TokenPrices.EstimateUsd(standard, new Usage(0, 0, 0, null)) is null,
            "missing usage is not a zero cost");
        Check(TokenPrices.EstimateUsd(standard with { Vendor = Vendor.Anthropic }, new Usage()) is null,
            "vendor is part of exact identity");
        Check(TokenPrices.EstimateUsd(standard with { Model = "gpt-4o-mini-guessed" }, new Usage()) is null,
            "no model-prefix aliases");
        Check(new MeterSnapshot(null, []).EstimatedUsd() == 0, "empty session costs zero");
        using var listener = Listen();
        var app = new Application(TextWriter.Null);
        try
        {
            var agent = new Agent(app, new Configuration(Vendor.OpenAI, "gpt-4o", listener.Prefixes.Single()));
            var engine = agent.Engine();
            async Task Send(string? model, string? tier)
            {
                var send = engine.SendAsync("{}");
                string body = System.Text.Json.JsonSerializer.Serialize(new
                {
                    model,
                    service_tier = tier,
                    choices = new[] { new { message = new { content = "ok" } } },
                    usage = new { prompt_tokens = 120, prompt_tokens_details = new { cached_tokens = 20 }, completion_tokens = 10 }
                });
                await Reply(await listener.GetContextAsync().WaitAsync(TimeSpan.FromSeconds(10)), body);
                var parsed = await send;
                Check(parsed.Events.Any(fact => fact.Kind == EventKind.ResponseEnded), "priced response genuinely completed");
            }
            await Send("gpt-4o-mini-2024-07-18", "default");
            await Send("gpt-4o-mini-2024-07-18", "default");
            await Send("gpt-4o-2024-08-06", "default");
            var known = engine.Meter();
            Check(known.Rows.Length == 2 && known.Totals.AllInput() == 360
                && known.EstimatedUsd() == 0.00042m, "accumulate counts by actual model; derive each row at its own rate");
            await Send("gpt-4o-mini-2024-07-18", null);
            await Send("gpt-4o-mini-2024-07-18", "priority");
            var mixed = engine.Meter();
            Check(mixed.Rows.Length == 4 && mixed.EstimatedUsd() is null && mixed.Totals.AllInput() == 600
                && mixed.Latest!.Counts.AllInput() == 120, "unpriced tiers preserve latest usage and session counts");
            Check(known.EstimatedUsd() == 0.00042m, "older immutable snapshot remains priced independently");
            await Send(null, "default");
            Check(engine.Meter().Latest!.Identity.Model == "" && engine.Meter().EstimatedUsd() is null,
                "missing response model does not borrow configured alias for billing");
        }
        finally
        {
            app.Destroy();
        }
    }

    public static async Task Streaming()
    {
        using var listener = Listen();
        var app = new Application(TextWriter.Null);
        try
        {
            var agent = new Agent(app, new Configuration(Vendor.OpenAI, "unknown", listener.Prefixes.Single()));
            var send = agent.Engine().SendAsync("{}");
            var request = await listener.GetContextAsync().WaitAsync(TimeSpan.FromSeconds(10));
            string stream = "data: {\"model\":\"gpt-4o-mini-2024-07-18\",\"service_tier\":\"priority\",\"choices\":[{\"index\":0,\"delta\":{\"content\":\"ok\"},\"finish_reason\":null}],\"usage\":{\"prompt_tokens\":100,\"completion_tokens\":1}}\n\n"
                + "data: {\"choices\":[{\"index\":0,\"delta\":{},\"finish_reason\":\"stop\"}]}\n\n"
                + "data: {\"service_tier\":\"default\",\"choices\":[],\"usage\":{\"prompt_tokens\":100,\"prompt_tokens_details\":{\"cached_tokens\":80},\"completion_tokens\":5}}\n\n"
                + "data: [DONE]\n\n";
            await Reply(request, stream, type: "text/event-stream");
            var result = await send;
            Check(result.Events.Any(fact => fact.Kind == EventKind.ResponseEnded), "real SSE completes");
            var meter = agent.Engine().Meter();
            Check(meter.Latest!.Counts.Input == 20 && meter.Latest.Counts.CacheRead == 80 && meter.Latest.Counts.AllInput() == 100,
                "OpenAI inclusive prompt usage normalized before metering");
            Check(meter.Totals.Output == 5 && meter.Totals.AllInput() == 100 && meter.HitRate() == 0.8,
                "final cumulative stream usage counted once, not summed across frames");
            Check(meter.Latest.Identity.Model == "gpt-4o-mini-2024-07-18"
                && meter.Latest.Identity.ServiceTier == "default" && meter.EstimatedUsd() == 0.000012m,
                "final streaming tier/model preserved for standard token estimate");
        }
        finally
        {
            app.Destroy();
        }
    }
}
