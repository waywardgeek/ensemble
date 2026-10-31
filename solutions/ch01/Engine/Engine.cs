using System.Collections.Immutable;
using System.Net.Http.Headers;
using System.Text;
using System.Text.Json;

namespace Ensemble;

public sealed partial class Engine
{
    public partial Engine(Agent parent)
    {
        ArgumentNullException.ThrowIfNull(parent);
        // Connection setup can be bounded without timing out a valid generation.
        // HttpClient's default 100-second whole-request timeout is unsuitable.
        var transport = new SocketsHttpHandler
        {
            ConnectTimeout = TimeSpan.FromSeconds(30),
            AllowAutoRedirect = false,
            UseProxy = false
        };
        http = new HttpClient(transport);
        http.Timeout = Timeout.InfiniteTimeSpan;

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

    public partial Agent Parent()
    {
        return parent ?? throw new InvalidOperationException("Engine is detached.");
    }

    public partial Usage Usage()
    {
        return usage;
    }

    public partial async Task<string> ExchangeAsync(
        ImmutableArray<Message> requestMessages,
        CancellationToken cancellation)
    {
        ObjectDisposedException.ThrowIf(destroyed, this);
        Configuration config = Parent().Configuration();
        try
        {
            string endpoint = config.BaseUrl.TrimEnd('/') + "/v1/messages";
            var body = new
            {
                model = config.Model,
                max_tokens = 1024,
                system = SystemInstructions,
                stream = false,
                messages = requestMessages.Select(message => new
                {
                    role = message.Role,
                    content = message.Content
                })
            };

            using var request = new HttpRequestMessage(HttpMethod.Post, endpoint);
            request.Headers.Add("x-api-key", config.ApiKey);
            request.Headers.Add("anthropic-version", "2023-06-01");
            string json = JsonSerializer.Serialize(body);
            request.Content = new StringContent(json, Encoding.UTF8);
            request.Content.Headers.ContentType = new MediaTypeHeaderValue("application/json");

            // One send, no retry. ResponseHeadersRead plus the same cancellation
            // token keeps both the header and body waits under caller control.
            using HttpResponseMessage response = await http.SendAsync(
                request, HttpCompletionOption.ResponseHeadersRead, cancellation);
            if (!response.IsSuccessStatusCode)
            {
                throw new InvalidDataException("Provider refused the request.");
            }

            await using Stream stream = await response.Content.ReadAsStreamAsync(cancellation);
            using JsonDocument document = await JsonDocument.ParseAsync(
                stream, cancellationToken: cancellation);
            JsonElement root = document.RootElement;
            JsonElement measured = root.GetProperty("usage");
            long input = measured.GetProperty("input_tokens").GetInt64();
            long output = measured.GetProperty("output_tokens").GetInt64();
            if (input < 0 || output < 0)
            {
                throw new InvalidDataException("Usage cannot be negative.");
            }

            // Usage is a receipt, even when the reply has no usable text.
            // Calculate both before publishing so overflow cannot update one half.
            long totalInput = checked(usage.Input + input);
            long totalOutput = checked(usage.Output + output);
            usage = new Usage(totalInput, totalOutput);

            var answer = new StringBuilder();
            foreach (JsonElement block in root.GetProperty("content").EnumerateArray())
            {
                if (block.GetProperty("type").GetString() == "text")
                {
                    answer.Append(block.GetProperty("text").GetString());
                }
            }

            string text = answer.ToString();
            if (string.IsNullOrWhiteSpace(text))
            {
                throw new InvalidDataException("Provider returned no usable text.");
            }

            return text;
        }
        catch (OperationCanceledException) when (cancellation.IsCancellationRequested)
        {
            Parent().Parent().Log("Request cancelled by caller.");
            throw;
        }
        catch (Exception error) when (
            error is HttpRequestException or IOException or InvalidDataException or JsonException or
            KeyNotFoundException or InvalidOperationException or FormatException or
            OverflowException or ArgumentException or OperationCanceledException)
        {
            // Exceptions may contain URLs, credentials or provider-controlled
            // bodies. Do not pass their text or inner exception across this boundary.
            const string safeMessage = "Provider exchange failed; no conversation pair was saved.";
            Parent().Parent().Log(safeMessage);
            throw new InvalidOperationException(safeMessage);
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
