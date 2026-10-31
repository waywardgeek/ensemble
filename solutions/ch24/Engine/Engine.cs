using System.Collections.Immutable;
using System.Net.Http.Headers;
using System.Text;
namespace Ensemble;

public sealed partial class PlanAuthorizationException
{
    public partial PlanAuthorizationException()
        : base("Plan authorization unavailable. Check the selected account, sign-in, plan permission, and credential store. No API-key fallback was used.")
    {
    }
}

public sealed partial class Engine
{
    public partial Engine(Agent parent)
        : this(parent, new HttpClientHandler
        {
            // Bearer transports must not redirect or retain cookies.
            AllowAutoRedirect = parent.Configuration().Billing == BillingMode.ApiKey,
            UseCookies = parent.Configuration().Billing == BillingMode.ApiKey
        })
    {
    }

    internal partial Engine(Agent parent, HttpMessageHandler transport)
    {
        // Caller cancellation, not a blanket generation timer, bounds inference.
        http = new HttpClient(transport) { Timeout = Timeout.InfiniteTimeSpan };
        codexSessionID = Guid.NewGuid().ToString("N");
        try
        {
            parent.AddEngine(this);
        }
        catch
        {
            http.Dispose();
            throw;
        }
    }

    public partial Agent Parent() => parent ?? throw new ObjectDisposedException(nameof(Engine));
    private static string Endpoint(Configuration config)
    {
        if (config.Billing == BillingMode.Codex)
        {
            return CodexCredentials.ResponsesEndpoint;
        }
        // Accept either a service base or the complete documented endpoint.
        bool streaming = !config.DisableStreaming
            && LookupModel(config.Vendor, config.Model, out var features) && features.Stream;
        var model = config.Model;
        if (model.StartsWith("models/", StringComparison.Ordinal))
        {
            model = model[7..];
        }
        string path = config.Vendor switch
        {
            Vendor.Anthropic => "/v1/messages",
            Vendor.OpenAI => config.Surface == Surface.Responses ? "/v1/responses" : "/v1/chat/completions",
            _ => "/v1beta/models/" + Uri.EscapeDataString(model)
                + (streaming ? ":streamGenerateContent?alt=sse" : ":generateContent")
        };
        if (config.Vendor == Vendor.Gemini && streaming
            && config.BaseUrl.EndsWith(":generateContent", StringComparison.Ordinal))
        {
            return config.BaseUrl[..^":generateContent".Length] + ":streamGenerateContent?alt=sse";
        }
        if (config.BaseUrl.EndsWith(path, StringComparison.Ordinal))
        {
            return config.BaseUrl;
        }

        var version = config.Vendor == Vendor.Gemini ? "/v1beta" : "/v1";
        if (config.BaseUrl.EndsWith(version, StringComparison.Ordinal))
        {
            return config.BaseUrl + path[version.Length..];
        }
        return config.BaseUrl + path;
    }

    public partial async Task<ParsedResponse> SendAsync(string requestBody, CancellationToken cancellation, Action<Chunk>? onChunk)
    {
        Agent owner = Parent();
        Configuration config = owner.BeginEngineSend();
        try
        {
            return await SendCoreAsync(config, requestBody, cancellation, onChunk);
        }
        finally
        {
            owner.EndEngineSend();
        }
    }

    private partial async Task<ParsedResponse> SendCoreAsync(Configuration config, string requestBody, CancellationToken cancellation, Action<Chunk>? onChunk, bool auxiliary)
    {
        // Configuration remains authoritative on the parent. The body is an
        // immutable value prepared before this worker starts; hints arriving
        // during HTTP cannot change the bytes of the request already admitted.
        cancellation.ThrowIfCancellationRequested();
        using var request = new HttpRequestMessage(HttpMethod.Post, Endpoint(config));
        request.Content = new StringContent(requestBody, Encoding.UTF8, "application/json");
        request.Headers.Accept.Add(new MediaTypeWithQualityHeaderValue("application/json"));
        request.Headers.Accept.Add(new MediaTypeWithQualityHeaderValue("text/event-stream"));
        // Diagnostic secrets belong to this exchange, not persistent Engine
        // state. A retried response can echo the rejected or rotated bearer.
        var traceSecrets = new List<string> { config.ApiKey };
        switch (config.Vendor)
        {
            case Vendor.Anthropic:
                request.Headers.Add("x-api-key", config.ApiKey);
                request.Headers.Add("anthropic-version", "2023-06-01");
                break;
            case Vendor.OpenAI:
                // Resolve secret authority only here, never in projection or History.
                string bearer = config.ApiKey;
                if (config.Billing == BillingMode.ChatGptPlan)
                {
                    bearer = await PlanBearerAsync(null, cancellation);
                }
                else if (config.Billing == BillingMode.Codex)
                {
                    var source = Parent().Parent().CodexCredentials() ?? throw new CodexAuthorizationException();
                    var grant = await source.ReadAsync(cancellation);
                    bearer = grant.Bearer;
                    traceSecrets.Add(grant.Account);
                    request.Headers.Add("ChatGPT-Account-ID", grant.Account);
                    using var body = System.Text.Json.JsonDocument.Parse(requestBody);
                    string session = codexSessionID;
                    if (body.RootElement.TryGetProperty("prompt_cache_key", out var key)
                        && key.ValueKind == System.Text.Json.JsonValueKind.String
                        && !string.IsNullOrEmpty(key.GetString()))
                    {
                        session = key.GetString()!;
                    }
                    request.Headers.Add("session-id", session);
                }
                traceSecrets.Add(bearer);
                request.Headers.Authorization = new AuthenticationHeaderValue("Bearer", bearer);
                break;
            case Vendor.Gemini:
                request.Headers.Add("x-goog-api-key", config.ApiKey);
                break;
        }
        cancellation.ThrowIfCancellationRequested();
        CacheLens? lens = auxiliary ? null : Lens();
        CachePrediction? prediction = lens?.Begin(requestBody);
        Usage? completedUsage = null;
        try
        {
            Parent().Parent().Trace(Parent(), "request", requestBody, traceSecrets);
            var response = await http.SendAsync(request, HttpCompletionOption.ResponseHeadersRead, cancellation);
            ParsedResponse parsed;
            try
            {
                if (config.Billing == BillingMode.ChatGptPlan
                    && response.StatusCode == System.Net.HttpStatusCode.Unauthorized)
                {
                    // Only HTTP 401 grants this one retry. Do not parse its body,
                    // publish deltas, or count usage. Other errors and interrupted
                    // successful streams are not safe grounds to replay inference.
                    response.Dispose();
                    string bearer = await PlanBearerAsync(request.Headers.Authorization!.Parameter, cancellation);
                    traceSecrets.Add(bearer);
                    using var retry = new HttpRequestMessage(request.Method, request.RequestUri);
                    retry.Content = new StringContent(requestBody, Encoding.UTF8, "application/json");
                    foreach (var header in request.Headers)
                    {
                        retry.Headers.TryAddWithoutValidation(header.Key, header.Value);
                    }
                    retry.Headers.Authorization = new AuthenticationHeaderValue("Bearer", bearer);
                    response = await http.SendAsync(retry, HttpCompletionOption.ResponseHeadersRead, cancellation);
                }
                if (config.Billing == BillingMode.Codex
                    && response.StatusCode == System.Net.HttpStatusCode.Unauthorized)
                {
                    // Codex owns renewal; do not replay inference or touch its file.
                    throw new CodexAuthorizationException();
                }
                // A second public-plan 401 reaches the normal safe error parser. No loop,
                // payment transition, or successful fact from the first attempt.
                parsed = await ParseCoreAsync(config, response, traceSecrets, onChunk, cancellation, auxiliary || Parent().IsMemoryCompressor());
            }
            finally
            {
                response.Dispose();
            }
            var completed = parsed.Events.FirstOrDefault(fact => fact.Kind == EventKind.ResponseEnded);
            completedUsage = completed?.Usage;
            if (completed is not null)
            {
                // Count the completed transport once, independently of whether
                // its facts are subsequently recorded or replayed by History.
                var identity = parsed.Billing ?? new UsageIdentity(config.Vendor, "", null);
                RecordUsage(identity, completedUsage ?? new Usage(null, null, null, null), auxiliary);
            }
            // Meter completed transport even when its text cannot authorize a
            // memory write or recall verdict. Incomplete specialized replies do
            // not reach their caller as success; ordinary acceptance is unchanged.
            if (parsed.IncompleteOutput)
            {
                return new ParsedResponse([new Event(EventKind.ErrorOccurred,
                    error: auxiliary ? "Recall judge response did not finish successfully."
                        : "Compressor response did not finish successfully.")], []);
            }
            return parsed;
        }
        catch (Exception ex) when (ex is HttpRequestException or IOException)
        {
            // Remote diagnostic strings and URLs are never conversation facts.
            return new ParsedResponse([new Event(EventKind.ErrorOccurred,
                error: "HTTP exchange failed (" + ex.GetType().Name + ")")], []);
        }
        finally
        {
            if (prediction is not null)
            {
                lens!.Complete(prediction, completedUsage);
            }
        }
    }

    private partial async Task<string> PlanBearerAsync(string? rejectedBearer, CancellationToken cancellation)
    {
        try
        {
            var credentials = Parent().Parent().Credentials() ?? throw new PlanAuthorizationException();
            return await credentials.GetBearerAfterRejectionAsync(Parent().Configuration().Account,
                rejectedBearer, cancellation);
        }
        catch (Exception error) when (error is not ObjectDisposedException
            && error is (InvalidOperationException or InvalidDataException
                or IOException or UnauthorizedAccessException or HttpRequestException))
        {
            // Store paths, issuer diagnostics and tokens are private. Cancellation
            // and disposed-owner faults retain their ordinary behavior. This catch
            // covers credential acquisition, not rendering or actor bookkeeping.
            throw new PlanAuthorizationException();
        }
    }

    public partial void Destroy()
    {
        if (destroyed)
        {
            return;
        }

        destroyed = true;
        try
        {
            Lens()?.Destroy();
            http.Dispose();
        }
        finally
        {
            parent?.RemoveEngine(this);
        }
    }
}
