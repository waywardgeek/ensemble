using System.Collections.Immutable;
using System.Net.Http.Headers;
using System.Text;
namespace Ensemble;

public sealed partial class Engine
{
    public partial Engine(Agent parent)
    {
        // A model may spend more than a minute before headers or body arrive.
        // Caller cancellation, not a blanket generation timer, controls this wait.
        http = new HttpClient { Timeout = Timeout.InfiniteTimeSpan };
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
        // Accept either a service base or the complete documented endpoint.
        var model = config.Model;
        if (model.StartsWith("models/", StringComparison.Ordinal))
        {
            model = model[7..];
        }
        string path = config.Vendor switch
        {
            Vendor.Anthropic => "/v1/messages",
            Vendor.OpenAI => "/v1/chat/completions",
            _ => "/v1beta/models/" + Uri.EscapeDataString(model) + ":generateContent"
        };
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

    public partial async Task<ImmutableArray<Event>> SendAsync(string requestBody, CancellationToken cancellation)
    {
        // Configuration remains authoritative on the parent. The body is an
        // immutable value prepared before this worker starts; hints arriving
        // during HTTP cannot change the bytes of the request already admitted.
        var config = Parent().Configuration();
        using var request = new HttpRequestMessage(HttpMethod.Post, Endpoint(config));
        request.Content = new StringContent(requestBody, Encoding.UTF8, "application/json");
        request.Headers.Accept.Add(new MediaTypeWithQualityHeaderValue("application/json"));
        switch (config.Vendor)
        {
            case Vendor.Anthropic:
                request.Headers.Add("x-api-key", config.ApiKey);
                request.Headers.Add("anthropic-version", "2023-06-01");
                break;
            case Vendor.OpenAI:
                request.Headers.Authorization = new AuthenticationHeaderValue("Bearer", config.ApiKey);
                break;
            case Vendor.Gemini:
                request.Headers.Add("x-goog-api-key", config.ApiKey);
                break;
        }
        try
        {
            using var response = await http.SendAsync(request, HttpCompletionOption.ResponseHeadersRead, cancellation);
            var responseBody = await response.Content.ReadAsStringAsync(cancellation);
            return Parse((int)response.StatusCode, responseBody);
        }
        catch (Exception ex) when (ex is HttpRequestException or OperationCanceledException or FormatException)
        {
            // Only our parser's fixed diagnostic is safe to retain. Transport
            // exception messages can include URLs or other sensitive details.
            var message = ex switch
            {
                FormatException => ex.Message,
                OperationCanceledException => "Request cancelled",
                _ => "HTTP request failed"
            };
            return [new Event(EventKind.ErrorOccurred, error: message)];
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
            http.Dispose();
        }
        finally
        {
            parent?.RemoveEngine(this);
        }
    }
}
