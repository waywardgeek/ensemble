using System.Net.Http.Headers;
using System.Text;
using System.Text.Json;
using System.Text.Json.Nodes;
using System.Threading.Channels;

namespace Ensemble;

public sealed partial class McpHttpTransport
{
    public partial McpHttpTransport(string url, string? bearer)
    {
        if (!Uri.TryCreate(url, UriKind.Absolute, out var parsed)
            || parsed.Scheme is not ("http" or "https") || parsed.UserInfo != "" || parsed.Fragment != "")
        {
            throw new ArgumentException("MCP requires an absolute HTTP(S) URL without userinfo or fragment.");
        }
        endpoint = parsed;
        if (bearer is not null && bearer.Any(value => !char.IsAsciiLetterOrDigit(value)
            && value is not ('-' or '.' or '_' or '~' or '+' or '/' or '=')))
        {
            throw new ArgumentException("Invalid MCP bearer header value.");
        }
        this.bearer = string.IsNullOrEmpty(bearer) ? null : bearer;
        // Bearers must never follow redirects to another endpoint. Timeout is
        // caller-owned, just as it is for stdio and model transport.
        http = new HttpClient(new HttpClientHandler { AllowAutoRedirect = false, UseCookies = false })
        {
            Timeout = Timeout.InfiniteTimeSpan
        };
    }

    public partial Task SendAsync(JsonElement message, CancellationToken cancellation)
    {
        lock (gate)
        {
            ObjectDisposedException.ThrowIf(closed, this);
            cancellation.ThrowIfCancellationRequested();
            bool request = message.TryGetProperty("method", out _) && message.TryGetProperty("id", out _);
            posts.RemoveWhere(task => task.IsCompletedSuccessfully);
            Task post = PostAsync(message.Clone(), request, cancellation);
            posts.Add(post);
            // Request replies arrive through ReceiveAsync. Returning now lets
            // the client's ordinary pending-reply cancellation remain local to
            // this call, rather than treating it as a torn framed-stream write.
            return request ? Task.CompletedTask : post;
        }
    }

    private partial async Task PostAsync(JsonElement message, bool request, CancellationToken cancellation)
    {
        using var linked = CancellationTokenSource.CreateLinkedTokenSource(cancellation, closing.Token);
        try
        {
            using var post = new HttpRequestMessage(HttpMethod.Post, endpoint);
            post.Content = new ByteArrayContent(JsonSerializer.SerializeToUtf8Bytes(message));
            post.Content.Headers.ContentType = new MediaTypeHeaderValue("application/json");
            post.Headers.Accept.ParseAdd("application/json, text/event-stream");
            if (bearer is not null)
            {
                post.Headers.Authorization = new AuthenticationHeaderValue("Bearer", bearer);
            }
            lock (gate)
            {
                if (session is not null)
                {
                    post.Headers.Add("Mcp-Session-Id", session);
                }
            }
            using var response = await http.SendAsync(post, HttpCompletionOption.ResponseHeadersRead, linked.Token);
            if (!response.IsSuccessStatusCode)
            {
                string detail = await ReadErrorAsync(response, linked.Token);
                throw new HttpRequestException("MCP HTTP " + (int)response.StatusCode + detail);
            }
            // Publish session state before exposing initialize's result. The
            // next initialized notification must carry exactly this identifier.
            if (message.TryGetProperty("method", out var method) && method.GetString() == "initialize"
                && response.Headers.TryGetValues("Mcp-Session-Id", out var values))
            {
                var ids = values.ToArray();
                if (ids.Length != 1 || string.IsNullOrWhiteSpace(ids[0]) || ids[0].Length > 4096
                    || ids[0].Any(value => value < 0x21 || value > 0x7e))
                {
                    throw new InvalidDataException("Invalid MCP session header.");
                }
                lock (gate)
                {
                    session = ids[0];
                }
            }
            if (!request)
            {
                return; // Notifications and reverse replies require no RPC reply.
            }
            if (response.StatusCode is System.Net.HttpStatusCode.Accepted or System.Net.HttpStatusCode.NoContent)
            {
                throw new InvalidDataException("MCP request received no reply; GET-stream delivery is unsupported.");
            }
            await ReadReplyAsync(response, message.GetProperty("id"), linked.Token);
        }
        catch (OperationCanceledException) when (linked.IsCancellationRequested)
        {
            if (!request)
            {
                throw;
            }
            // Its caller has cancelled, or disposal is ending the reader. No
            // synthetic result is needed for a now-absent pending request.
        }
        catch (Exception error) when (error is HttpRequestException or IOException or InvalidDataException
            or JsonException or DecoderFallbackException)
        {
            if (!request)
            {
                throw new IOException("MCP HTTP notification failed.");
            }
            string detail = error is HttpRequestException && error.Message.StartsWith("MCP HTTP ", StringComparison.Ordinal)
                ? error.Message : "MCP HTTP reply was unavailable or invalid.";
            await PublishAsync(JsonSerializer.SerializeToElement(new
            {
                jsonrpc = "2.0", id = message.GetProperty("id"),
                error = new { code = -32000, message = detail }
            }), linked.Token);
        }
    }

    private partial async Task ReadReplyAsync(HttpResponseMessage response, JsonElement id, CancellationToken cancellation)
    {
        using var stream = await response.Content.ReadAsStreamAsync(cancellation);
        using var reader = new StreamReader(stream, new UTF8Encoding(false, true));
        string? media = response.Content.Headers.ContentType?.MediaType;
        if (string.Equals(media, "application/json", StringComparison.OrdinalIgnoreCase))
        {
            // Bound the response before parsing rather than trusting Content-Length.
            var text = new StringBuilder();
            var buffer = new char[4096];
            int count;
            while ((count = await reader.ReadAsync(buffer.AsMemory(), cancellation)) != 0)
            {
                text.Append(buffer, 0, count);
                if (text.Length > 4 * 1024 * 1024)
                {
                    throw new InvalidDataException("MCP JSON reply is too large.");
                }
            }
            using var document = JsonDocument.Parse(text.ToString());
            var message = document.RootElement;
            if (message.ValueKind != JsonValueKind.Object
                || !message.TryGetProperty("id", out var replyId) || replyId.GetRawText() != id.GetRawText())
            {
                throw new InvalidDataException("MCP HTTP reply identity does not match its POST.");
            }
            await PublishAsync(message, cancellation);
            return;
        }
        if (!string.Equals(media, "text/event-stream", StringComparison.OrdinalIgnoreCase))
        {
            throw new InvalidDataException("Unsupported MCP HTTP reply content type.");
        }
        var data = new StringBuilder();
        // Read incrementally, with a bounded line as well as a bounded event.
        // An open SSE response completes when this POST's matching reply arrives;
        // comments, progress and unrelated IDs are not its completion condition.
        var line = new StringBuilder();
        var chars = new char[4096];
        int used;
        bool previousCarriageReturn = false;
        while ((used = await reader.ReadAsync(chars.AsMemory(), cancellation)) != 0)
        {
            for (int index = 0; index < used; index++)
            {
                char value = chars[index];
                if (value == '\n' && previousCarriageReturn)
                {
                    previousCarriageReturn = false;
                    continue;
                }
                previousCarriageReturn = value == '\r';
                if (value is not ('\r' or '\n'))
                {
                    line.Append(value);
                    if (line.Length > 4 * 1024 * 1024)
                    {
                        throw new InvalidDataException("MCP SSE line is too large.");
                    }
                    continue;
                }
                string field = line.ToString();
                line.Clear();
                if (field.Length == 0 && data.Length != 0)
                {
                    using var document = JsonDocument.Parse(data.ToString());
                    var message = document.RootElement;
                    if (message.ValueKind != JsonValueKind.Object)
                    {
                        throw new InvalidDataException("MCP SSE requires JSON-RPC objects.");
                    }
                    bool reply = !message.TryGetProperty("method", out _)
                        && message.TryGetProperty("id", out var replyId)
                        && replyId.GetRawText() == id.GetRawText();
                    if (reply || message.TryGetProperty("method", out _))
                    {
                        await PublishAsync(message, cancellation);
                    }
                    data.Clear();
                    if (reply)
                    {
                        return;
                    }
                }
                else if (field.StartsWith("data:", StringComparison.Ordinal))
                {
                    string fragment = field[5..];
                    if (fragment.StartsWith(' '))
                    {
                        fragment = fragment[1..];
                    }
                    data.Append(fragment).Append('\n');
                    if (data.Length > 4 * 1024 * 1024)
                    {
                        throw new InvalidDataException("MCP SSE event is too large.");
                    }
                }
            }
        }
        throw new InvalidDataException("MCP SSE ended without its request reply.");
    }

    private partial async Task<string> ReadErrorAsync(HttpResponseMessage response, CancellationToken cancellation)
    {
        using var stream = await response.Content.ReadAsStreamAsync(cancellation);
        using var reader = new StreamReader(stream, new UTF8Encoding(false, true));
        var buffer = new char[8193];
        int length = await reader.ReadBlockAsync(buffer.AsMemory(), cancellation);
        if (length > 8192)
        {
            return ": oversized error body omitted.";
        }
        string text = new string(buffer, 0, length);
        // Error bodies are untrusted tool data. A complete bounded body can be
        // redacted safely; an oversized body is omitted, never cut mid-secret.
        if (bearer is not null)
        {
            text = text.Replace(bearer, "[redacted]", StringComparison.Ordinal);
            string escaped = JsonSerializer.Serialize(bearer)[1..^1];
            text = text.Replace(escaped, "[redacted]", StringComparison.Ordinal);
        }
        return text.Length == 0 ? "" : ": " + text;
    }

    private partial async Task PublishAsync(JsonElement message, CancellationToken cancellation)
    {
        // A successful HTTP status can still carry an RPC error. Its decoded
        // diagnostic may echo our bearer, just like a non-2xx HTTP body. Redact
        // that diagnostic only: result content and reply identity are not ours
        // to rewrite. Decode first so JSON escaping cannot hide the token.
        if (bearer is not null && message.TryGetProperty("error", out var error)
            && error.ValueKind == JsonValueKind.Object
            && error.TryGetProperty("message", out var detail) && detail.ValueKind == JsonValueKind.String
            && detail.GetString()!.Contains(bearer, StringComparison.Ordinal))
        {
            var envelope = JsonNode.Parse(message.GetRawText())!.AsObject();
            envelope["error"]!["message"] = detail.GetString()!.Replace(bearer, "[redacted]", StringComparison.Ordinal);
            message = JsonSerializer.SerializeToElement(envelope);
        }
        await incoming.Writer.WriteAsync(message.Clone(), cancellation);
    }

    public partial async Task<JsonElement?> ReceiveAsync(CancellationToken cancellation)
    {
        try
        {
            return await incoming.Reader.ReadAsync(cancellation);
        }
        catch (ChannelClosedException)
        {
            return null;
        }
    }

    public partial ValueTask DisposeAsync()
    {
        lock (gate)
        {
            closed = true;
            destruction ??= Task.Run(DestroyAsync);
            return new ValueTask(destruction);
        }
    }

    private partial async Task DestroyAsync()
    {
        closing.Cancel();
        http.Dispose();
        Task[] pending;
        lock (gate)
        {
            pending = posts.ToArray();
        }
        try
        {
            await Task.WhenAll(pending);
        }
        catch (Exception error) when (error is OperationCanceledException or IOException)
        {
            // Notification senders retain their own failure. All posts are
            // nevertheless joined before the owner detaches this transport.
        }
        finally
        {
            incoming.Writer.TryComplete();
        }
    }
}
