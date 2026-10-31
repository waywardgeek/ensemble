using System.Net.WebSockets;
using System.Text.Json;

namespace Ensemble.Gateway;

public static partial class Protocol
{
    public static partial string ProvenanceHeader(Source source)
    {
        ArgumentNullException.ThrowIfNull(source);
        return $"[GATEWAY source={source.ChannelID} user={source.UserID} trust={source.Trust}]";
    }
}

internal static partial class Wire
{
    internal static partial byte[] Encode(Frame frame)
    {
        byte[] bytes = JsonSerializer.SerializeToUtf8Bytes(frame, Options);
        if (bytes.Length > MaxPlaintext)
        {
            throw new ArgumentException("Gateway frame exceeds the Noise message limit.", nameof(frame));
        }
        return bytes;
    }

    internal static partial Envelope NormalizeEnvelope(Envelope value) => value with
    {
        // Empty optional strings are values, not missing required provenance or
        // content. The same normalization governs network and disk reads.
        ConversationID = value.ConversationID ?? "",
        ResponseMode = value.ResponseMode ?? "",
        Source = value.Source is { } source ? source with { DisplayName = source.DisplayName ?? "" } : null!
    };

    internal static partial Frame Decode(ReadOnlyMemory<byte> bytes)
    {
        using JsonDocument document = JsonDocument.Parse(bytes, new JsonDocumentOptions { MaxDepth = 32 });
        UniqueNames(document.RootElement);
        Frame frame = document.RootElement.Deserialize<Frame>(Options)
            ?? throw new JsonException("Missing Gateway frame.");
        if (string.IsNullOrEmpty(frame.Type))
        {
            throw new JsonException("Missing Gateway frame type.");
        }
        if (frame.Envelope is { } envelope)
        {
            frame = frame with { Envelope = NormalizeEnvelope(envelope) };
        }
        if (frame.Type == "mcp_response" && frame.Rpc is { } reply)
        {
            JsonElement rpc = document.RootElement.GetProperty("rpc");
            bool hasResult = rpc.TryGetProperty("result", out JsonElement result);
            bool hasError = rpc.TryGetProperty("error", out _);
            if (hasResult == hasError)
            {
                throw new JsonException("RPC response requires exactly one result or error.");
            }
            if (hasResult)
            {
                // Nullable<JsonElement> deserialization maps explicit JSON null
                // to absence. Preserve the supplied result (including null), not
                // a borrowed element whose document is about to be disposed.
                frame = frame with { Rpc = reply with { Result = result.Clone() } };
            }
        }
        return frame;
    }

    internal static partial void UniqueNames(JsonElement value)
    {
        if (value.ValueKind == JsonValueKind.Object)
        {
            HashSet<string> names = new(StringComparer.Ordinal);
            foreach (JsonProperty property in value.EnumerateObject())
            {
                if (!names.Add(property.Name))
                {
                    throw new JsonException("Duplicate Gateway JSON property.");
                }
                UniqueNames(property.Value);
            }
        }
        else if (value.ValueKind == JsonValueKind.Array)
        {
            foreach (JsonElement item in value.EnumerateArray())
            {
                UniqueNames(item);
            }
        }
    }

    internal static partial async Task<byte[]> ReadBinaryAsync(WebSocket socket, CancellationToken cancellation)
    {
        // One extra byte distinguishes a legal full frame from an oversized one,
        // including fragmented messages. No allocation follows a peer's length.
        byte[] buffer = new byte[MaxCiphertext + 1];
        int length = 0;
        while (true)
        {
            ValueWebSocketReceiveResult received = await socket.ReceiveAsync(buffer.AsMemory(length), cancellation);
            if (received.MessageType == WebSocketMessageType.Close)
            {
                throw new EndOfStreamException("Gateway peer closed the connection.");
            }
            if (received.MessageType != WebSocketMessageType.Binary)
            {
                throw new InvalidDataException("Gateway requires binary WebSocket messages.");
            }
            length += received.Count;
            if (length > MaxCiphertext)
            {
                throw new InvalidDataException("Gateway binary message exceeds the Noise limit.");
            }
            if (received.EndOfMessage)
            {
                return buffer.AsSpan(0, length).ToArray();
            }
        }
    }
}
