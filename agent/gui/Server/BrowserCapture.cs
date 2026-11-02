using Ensemble;
using System.Text.Json;
using System.Threading.Channels;

namespace EnsembleGui;

public sealed partial class McpTunnel
{
    public partial McpTunnel(ChannelWriter<string> output, string agentID, Agent captureTarget)
        : this(output, agentID)
    {
        this.captureTarget = captureTarget ?? throw new ArgumentNullException(nameof(captureTarget));
        captureSelection = captureTarget;
    }

    public partial void SetCaptureSelection(Agent selected)
    {
        ArgumentNullException.ThrowIfNull(selected);
        lock (captureGate)
        {
            captureSelection = selected;
            long? previous = captureID;
            captureID = null;
            // Even selecting away and back cannot revive a previous request.
            // The host has already queued its selection replay; the browser
            // uses that same event to cancel pending consent and native capture.
            if (previous is not null && Volatile.Read(ref disposed) == 0
                && !input.Writer.TryWrite(CaptureError(previous.Value, "Page capture cancelled by selection change.")))
            {
                input.Writer.TryComplete(new IOException("MCP browser input is full."));
            }
        }
    }

    private static partial long? ReplyID(JsonElement message)
    {
        return message.ValueKind == JsonValueKind.Object
            && message.TryGetProperty("id", out var id) && id.ValueKind == JsonValueKind.Number
            && id.TryGetInt64(out long value)
            ? value : null;
    }

    public partial bool IsCaptureReply(JsonElement message)
    {
        long? id = ReplyID(message);
        lock (captureGate)
            return id is not null && captureID == id && !captureReplyStarted
                && Volatile.Read(ref disposed) == 0;
    }

    private partial void RequireCaptureCapability()
    {
        if (captureTarget is null)
            throw new InvalidOperationException("Capture requires an explicitly bound Agent.");
        // This read-free check tests tool placement as well as model capability.
        // The placeholder is never decoded or returned as an image.
        var image = new BlobPart("image/png", new BlobRef(RefKind.Uri, "data:image/png;base64,AA=="));
        captureTarget.Engine().ValidateMedia([new ToolResultPart("capture-policy", [image])]);
    }

    private partial bool AdmitCaptureRequest(JsonElement message)
    {
        if (message.ValueKind != JsonValueKind.Object
            || !message.TryGetProperty("method", out var method)) return true;
        string? name = method.ValueKind == JsonValueKind.String ? method.GetString() : null;
        if (name == "$/cancelRequest" && message.TryGetProperty("params", out var cancellation))
        {
            long? cancelled = ReplyID(cancellation);
            lock (captureGate)
                if (cancelled is not null && captureID == cancelled) captureID = null;
            return true; // The browser must also stop its pending consent/capture.
        }
        if (name != "tools/call" || !message.TryGetProperty("params", out var parameters)
            || parameters.ValueKind != JsonValueKind.Object
            || !parameters.TryGetProperty("name", out var tool)
            || tool.ValueKind != JsonValueKind.String || tool.GetString() != "gui_capture") return true;

        long id = ReplyID(message) ?? throw new InvalidDataException("Capture request needs an integer RPC ID.");
        string? refusal = null;
        try { RequireCaptureCapability(); }
        catch (Exception error) when (error is InvalidOperationException or NotSupportedException)
        { refusal = "The selected model or connection does not support screenshot tool images."; }
        lock (captureGate)
        {
            if (refusal is null && !ReferenceEquals(captureSelection, captureTarget))
                refusal = "Select the requesting Agent before requesting a page capture.";
            if (refusal is null && captureID is not null)
                refusal = "A page capture is already pending on this connection.";
            if (refusal is null)
            {
                captureID = id;
                captureReplyStarted = false;
                return true;
            }
        }
        // Refusal is a normal paired tool result, not a broken MCP connection.
        if (!input.Writer.TryWrite(CaptureError(id, refusal)))
            throw new IOException("MCP browser channel is closed or full.");
        return false;
    }

    private static partial bool HasImageResult(JsonElement message)
    {
        if (!message.TryGetProperty("result", out var result) || result.ValueKind != JsonValueKind.Object
            || !result.TryGetProperty("content", out var content) || content.ValueKind != JsonValueKind.Array)
            return false;
        return content.EnumerateArray().Any(item => item.ValueKind == JsonValueKind.Object
            && item.TryGetProperty("type", out var type) && type.ValueKind == JsonValueKind.String
            && type.GetString() is "image" or "audio");
    }

    private partial bool AcceptCaptureReply(JsonElement message)
    {
        long? id = ReplyID(message);
        lock (captureGate)
        {
            if (id is null || captureID != id)
            {
                if (captureTarget is not null && HasImageResult(message))
                    throw new InvalidDataException("Browser media requires a pending capture request.");
                return false;
            }
            if (captureReplyStarted) throw new InvalidDataException("Duplicate capture reply.");
            captureReplyStarted = true;
        }
        JsonElement prepared;
        try { prepared = PrepareCaptureReply(id.Value, message); }
        catch (Exception error) when (error is InvalidDataException or InvalidOperationException
            or NotSupportedException or FormatException or ArgumentException or KeyNotFoundException)
        { prepared = CaptureError(id.Value, "Page capture did not provide one valid, bounded PNG image."); }
        lock (captureGate)
        {
            // Explicit cancellation or disconnect may win during native decoding.
            if (captureID != id || Volatile.Read(ref disposed) != 0) return true;
            captureID = null;
            if (!input.Writer.TryWrite(prepared))
                throw new IOException("MCP browser channel is closed or full.");
        }
        return true;
    }

    private partial JsonElement PrepareCaptureReply(long id, JsonElement message)
    {
        RequireCaptureCapability();
        if (!message.TryGetProperty("result", out var result) || result.ValueKind != JsonValueKind.Object
            || !result.TryGetProperty("content", out var content) || content.ValueKind != JsonValueKind.Array
            || content.GetArrayLength() != 1)
            throw new InvalidDataException("Capture requires exactly one content block.");
        var item = content[0];
        if (result.TryGetProperty("isError", out var failed))
        {
            if (failed.ValueKind is not (JsonValueKind.True or JsonValueKind.False))
                throw new InvalidDataException("Invalid capture error marker.");
            if (failed.GetBoolean())
            {
                if (item.GetProperty("type").GetString() != "text"
                    || item.GetProperty("text").GetString() is not string text
                    || text.Length > 8192)
                    throw new InvalidDataException("Capture error requires bounded text.");
                return CaptureError(id, text);
            }
        }
        if (item.GetProperty("type").GetString() != "image"
            || item.GetProperty("mimeType").GetString() != "image/png")
            throw new InvalidDataException("Capture requires PNG pixels, not metadata.");
        string data = item.GetProperty("data").GetString() ?? "";
        int maximumBase64 = ((Agent.MaximumImageSourceBytes + 2) / 3) * 4;
        if (data.Length == 0 || data.Length > maximumBase64 || data.Any(char.IsWhiteSpace))
            throw new InvalidDataException("Capture source exceeds its bound or is malformed.");
        byte[] bytes = Convert.FromBase64String(data);
        ReadOnlySpan<byte> pngSignature = [137, 80, 78, 71, 13, 10, 26, 10];
        if (bytes.Length > Agent.MaximumImageSourceBytes || !bytes.AsSpan().StartsWith(pngSignature))
            throw new InvalidDataException("Capture source is not a bounded PNG.");
        PreparedImage prepared = captureTarget!.PrepareImage(bytes);
        string pixels = prepared.Image.Ref.Locator["data:image/png;base64,".Length..];
        string description = $"Browser-supplied screenshot PNG: {prepared.OriginalWidth}x{prepared.OriginalHeight}; "
            + $"prepared PNG {prepared.Width}x{prepared.Height}.";
        return JsonSerializer.SerializeToElement(new
        {
            jsonrpc = "2.0", id,
            result = new { content = new object[] {
                new { type = "text", text = description },
                new { type = "image", mimeType = "image/png", data = pixels }
            } }
        });
    }

    private static partial JsonElement CaptureError(long id, string message)
    {
        return JsonSerializer.SerializeToElement(new {
            jsonrpc = "2.0", id,
            result = new { isError = true, content = new[] { new { type = "text", text = message } } }
        });
    }
}
