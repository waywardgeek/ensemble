using System.Net.WebSockets;
using System.Text;
using System.Text.Json;

namespace SpeechListener;

internal sealed partial class Session
{
    private partial async Task<JsonElement> CommandAsync(string method, object parameters, CancellationToken cancellation)
    {
        // One request at a time: unsolicited DevTools events are ignored, never
        // exposed to the listener. No pending-reply registry or extra reader task.
        await protocol.WaitAsync(cancellation);
        try
        {
            int id = ++rpcID;
            byte[] request = JsonSerializer.SerializeToUtf8Bytes(new { id, method, @params = parameters });
            await socket.SendAsync(request, WebSocketMessageType.Text, true, cancellation);
            var buffer = new byte[8192];
            while (true)
            {
                using var message = new MemoryStream();
                WebSocketReceiveResult frame;
                do
                {
                    frame = await socket.ReceiveAsync(buffer, cancellation);
                    if (frame.MessageType == WebSocketMessageType.Close) throw new IOException("Owned browser disconnected");
                    message.Write(buffer, 0, frame.Count);
                    if (message.Length > 16 * 1024 * 1024) throw new IOException("Browser reply exceeds 16 MiB");
                } while (!frame.EndOfMessage);
                using var document = JsonDocument.Parse(message.ToArray());
                var response = document.RootElement;
                if (!response.TryGetProperty("id", out var replyID) || replyID.GetInt32() != id) continue;
                if (response.TryGetProperty("error", out _)) throw new IOException("Browser command failed: " + method);
                return response.GetProperty("result").Clone();
            }
        }
        finally
        {
            protocol.Release();
        }
    }

    private partial async Task<JsonElement> EvaluateAsync(string expression, CancellationToken cancellation)
    {
        var result = await CommandAsync("Runtime.evaluate", new { expression, returnByValue = true, awaitPromise = true }, cancellation);
        if (result.TryGetProperty("exceptionDetails", out _)) throw new IOException("Listener channel operation failed");
        return result.GetProperty("result").GetProperty("value").Clone();
    }
}
