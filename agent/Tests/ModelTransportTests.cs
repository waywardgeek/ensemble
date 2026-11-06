using System.Net;
using System.Text.Json;
using Ensemble;
using Application = Ensemble.Ensemble;

internal static partial class MeterTests
{
    public static async Task ModelEndpointRedirects()
    {
        using var source = Listen();
        using var destination = Listen();
        using var cancellation = new CancellationTokenSource(TimeSpan.FromSeconds(15));
        var app = new Application(TextWriter.Null);
        Task? serving = null;
        Task? redirected = null;
        Task<ParsedResponse>? exchange = null;
        int destinationRequests = 0;
        try
        {
            var agent = new Agent(app, new Configuration(Vendor.OpenAI, "fixture-model", source.Prefixes.Single(),
                "synthetic-key", disableStreaming: true));
            serving = Task.Run(async () =>
            {
                var incoming = await source.GetContextAsync().WaitAsync(cancellation.Token);
                Check(incoming.Request.Headers["Authorization"] == "Bearer synthetic-key", "Initial request reaches configured source.");
                incoming.Response.RedirectLocation = destination.Prefixes.Single() + "redirect-target";
                await Reply(incoming, "", 307);
            });
            redirected = Task.Run(async () =>
            {
                try
                {
                    var incoming = await destination.GetContextAsync().WaitAsync(cancellation.Token);
                    Interlocked.Increment(ref destinationRequests);
                    await Reply(incoming, JsonSerializer.Serialize(new
                    {
                        model = "fixture-model",
                        choices = new[] { new { message = new { content = "wrong destination" } } },
                        usage = new { prompt_tokens = 1, completion_tokens = 1 }
                    }));
                }
                catch (Exception error) when (cancellation.IsCancellationRequested
                    && error is OperationCanceledException or HttpListenerException or ObjectDisposedException)
                {
                    // The owned listener is released only after the exchange has
                    // completed or failed. Cancellation is cleanup, not the oracle.
                }
            });
            exchange = agent.Engine().SendAsync(agent.Engine().Render(new Context()).ToJsonString(), cancellation.Token);
            ParsedResponse result = await exchange;
            await serving;
            Check(Volatile.Read(ref destinationRequests) == 0,
                "Configured endpoint must not automatically follow an inference redirect.");
            Check(result.Events.Any(fact => fact.Kind == EventKind.ErrorOccurred)
                && !result.Events.Any(fact => fact.Kind == EventKind.ResponseEnded),
                "Redirect is a terminal error, not a successful response from another route.");
        }
        finally
        {
            cancellation.Cancel();
            source.Stop();
            destination.Stop();
            try
            {
                try { await Task.WhenAll(new Task?[] { serving, redirected, exchange }.OfType<Task>()); }
                catch (Exception error) when (error is OperationCanceledException or HttpListenerException
                    or ObjectDisposedException or HttpRequestException) { }
            }
            finally { app.Destroy(); }
        }
    }

    public static async Task ModelEndpointCookies()
    {
        using var source = Listen();
        using var cancellation = new CancellationTokenSource(TimeSpan.FromSeconds(15));
        var app = new Application(TextWriter.Null);
        Task? serving = null;
        Task<ParsedResponse>? exchange = null;
        string? secondCookie = null;
        int received = 0;
        try
        {
            var agent = new Agent(app, new Configuration(Vendor.OpenAI, "fixture-model", source.Prefixes.Single(),
                "synthetic-key", disableStreaming: true));
            serving = Task.Run(async () =>
            {
                for (int index = 0; index < 2; index++)
                {
                    var incoming = await source.GetContextAsync().WaitAsync(cancellation.Token);
                    if (index == 0)
                    {
                        incoming.Response.Headers.Add("Set-Cookie", "model_route_cookie=must-not-become-authority; Path=/");
                    }
                    else
                    {
                        secondCookie = incoming.Request.Headers["Cookie"];
                    }
                    await Reply(incoming, JsonSerializer.Serialize(new
                    {
                        model = "fixture-model",
                        choices = new[] { new { message = new { content = "answer" } } },
                        usage = new { prompt_tokens = 1, completion_tokens = 1 }
                    }));
                    Interlocked.Increment(ref received);
                }
            });
            for (int index = 0; index < 2; index++)
            {
                exchange = agent.Engine().SendAsync(agent.Engine().Render(new Context()).ToJsonString(), cancellation.Token);
                ParsedResponse result = await exchange;
                Check(result.Events.Any(fact => fact.Kind == EventKind.ResponseEnded), "Cookie fixture produces a real successful exchange.");
            }
            await serving;
            Check(received == 2 && string.IsNullOrEmpty(secondCookie),
                "Inference transport must not retain response cookies as hidden authentication state.");
        }
        finally
        {
            cancellation.Cancel();
            source.Stop();
            try
            {
                try { await Task.WhenAll(new Task?[] { serving, exchange }.OfType<Task>()); }
                catch (Exception error) when (error is OperationCanceledException or HttpListenerException
                    or ObjectDisposedException or HttpRequestException) { }
            }
            finally { app.Destroy(); }
        }
    }
}
