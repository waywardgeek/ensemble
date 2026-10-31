using System.Text.Json;

namespace Ensemble;

public sealed partial class McpClient
{
    public partial Task ServeAsync(CancellationToken cancellation)
    {
        cancellation.ThrowIfCancellationRequested();
        lock (gate)
        {
            ObjectDisposedException.ThrowIf(destroyed || ended, this);
            if (serving)
            {
                return readerTask!;
            }
            if (connection is not null || readerTask is not null)
            {
                throw new InvalidOperationException("An outgoing MCP connection cannot become a service.");
            }

            // There is exactly one reader for either mode. The service does
            // not issue outgoing RPC, so sequential request dispatch cannot
            // starve a reply needed by this same connection.
            serving = true;
            readerTask = Task.Run(() => ServeCoreAsync(cancellation));
            return readerTask;
        }
    }

    private partial async Task ServeCoreAsync(CancellationToken cancellation)
    {
        // Cancelling service execution also releases a reverse-call observer.
        // Its independent Job retains its normal Agent-owned lifetime.
        using var registration = cancellation.Register(lifetime.Cancel);
        bool initialized = false;
        bool ready = false;
        try
        {
            while (await transport.ReceiveAsync(lifetime.Token) is { } request)
            {
                if (request.ValueKind != JsonValueKind.Object
                    || !request.TryGetProperty("jsonrpc", out var version)
                    || version.ValueKind != JsonValueKind.String || version.GetString() != "2.0"
                    || !request.TryGetProperty("method", out var methodValue)
                    || methodValue.ValueKind != JsonValueKind.String)
                {
                    throw new InvalidDataException("Invalid MCP service request envelope.");
                }
                string method = methodValue.GetString()!;
                if (!request.TryGetProperty("id", out var identity))
                {
                    // Notifications have no reply and cannot execute a tool.
                    if (method == "notifications/initialized")
                    {
                        if (!initialized)
                        {
                            throw new InvalidDataException("MCP initialized notification precedes initialize.");
                        }
                        ready = true;
                    }
                    continue;
                }
                if (identity.ValueKind != JsonValueKind.Number || !identity.TryGetInt64(out _))
                {
                    throw new InvalidDataException("MCP service requires integer request identities.");
                }

                JsonElement id = identity.Clone();
                if (method == "initialize")
                {
                    bool supported = request.TryGetProperty("params", out var parameters)
                        && parameters.ValueKind == JsonValueKind.Object
                        && parameters.TryGetProperty("protocolVersion", out var protocol)
                        && protocol.ValueKind == JsonValueKind.String
                        && protocol.GetString() == "2024-11-05";
                    if (initialized || !supported)
                    {
                        await transport.SendAsync(JsonSerializer.SerializeToElement(new
                        {
                            jsonrpc = "2.0", id,
                            error = new { code = -32602, message = "Invalid or repeated initialize" }
                        }), lifetime.Token);
                        continue;
                    }
                    await transport.SendAsync(JsonSerializer.SerializeToElement(new
                    {
                        jsonrpc = "2.0", id,
                        result = new
                        {
                            protocolVersion = "2024-11-05",
                            serverInfo = new { name = "ensemble-csharp", version = "1" },
                            capabilities = new { tools = new { } }
                        }
                    }), lifetime.Token);
                    initialized = true;
                }
                else if (!ready)
                {
                    await transport.SendAsync(JsonSerializer.SerializeToElement(new
                    {
                        jsonrpc = "2.0", id,
                        error = new { code = -32002, message = "MCP service is not initialized" }
                    }), lifetime.Token);
                }
                else
                {
                    // This handler already filters declarations through grants
                    // and routes real Jobs through the Agent's reverse lane.
                    await ReverseAsync(request);
                }
            }
        }
        finally
        {
            // Do not await DestroyAsync from the reader it would join. The
            // enclosing owner closes resources after this task has completed.
            End(new EndOfStreamException("MCP service stopped."));
        }
    }
}
