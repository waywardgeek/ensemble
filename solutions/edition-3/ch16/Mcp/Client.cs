using System.Collections.Immutable;
using System.Text.Json;
using System.Text;

namespace Ensemble;

public sealed partial class McpClient
{
    public partial McpClient(Mcp parent, string name, IMcpTransport transport,
        Skills? skillOwner, string? skillName, McpBrowser? browserOwner)
    {
        ArgumentNullException.ThrowIfNull(parent);
        ArgumentException.ThrowIfNullOrWhiteSpace(name);
        ArgumentNullException.ThrowIfNull(transport);
        Name = name;
        this.transport = transport;
        if ((skillOwner is null) != (skillName is null))
        {
            throw new ArgumentException("A skill connection requires both owner and skill name.");
        }
        SkillName = skillName;
        parent.AddClient(this);
        try
        {
            skillOwner?.AddClient(this);
            browserOwner?.AddClient(this);
        }
        catch
        {
            this.skillOwner?.RemoveClient(this);
            parent.RemoveClient(this);
            throw;
        }
    }

    public partial Skills? SkillOwner() => skillOwner;

    public partial McpBrowser? BrowserOwner() => browserOwner;

    public partial Mcp Parent() => parent ?? throw new ObjectDisposedException(nameof(McpClient));

    public partial ImmutableArray<McpToolDefinition> Catalog()
    {
        lock (gate)
        {
            return catalog;
        }
    }

    public partial Task ConnectAsync(CancellationToken cancellation)
    {
        lock (gate)
        {
            ObjectDisposedException.ThrowIf(destroyed || ended, this);
            if (serving)
            {
                throw new InvalidOperationException("An MCP service cannot initiate discovery.");
            }
            connection ??= ConnectCoreAsync(cancellation);
            return connection;
        }
    }

    private partial async Task ConnectCoreAsync(CancellationToken cancellation)
    {
        try
        {
            // This chapter uses its published historical MCP version. This is
            // not a claim of support for every newer MCP transport/version.
            var initialized = await RequestAsync("initialize", JsonSerializer.SerializeToElement(new
            {
                protocolVersion = "2024-11-05",
                capabilities = new { },
                clientInfo = new { name = "ensemble-csharp", version = "1" }
            }), cancellation);
            if (!initialized.TryGetProperty("protocolVersion", out var version)
                || version.GetString() != "2024-11-05")
            {
                throw new InvalidDataException("Unsupported MCP protocol version.");
            }
            await NotifyAsync("notifications/initialized", JsonSerializer.SerializeToElement(new { }), cancellation);
            var listed = await RequestAsync("tools/list", JsonSerializer.SerializeToElement(new { }), cancellation);
            if (!listed.TryGetProperty("tools", out var tools) || tools.ValueKind != JsonValueKind.Array)
            {
                throw new InvalidDataException("MCP tools/list requires a tools array.");
            }
            var names = new HashSet<string>(StringComparer.Ordinal);
            var discovered = ImmutableArray.CreateBuilder<McpToolDefinition>();
            foreach (var tool in tools.EnumerateArray())
            {
                string? name = tool.TryGetProperty("name", out var named) ? named.GetString() : null;
                if (string.IsNullOrWhiteSpace(name) || !names.Add(name)
                    || !tool.TryGetProperty("inputSchema", out var schema) || schema.ValueKind != JsonValueKind.Object)
                {
                    throw new InvalidDataException("Invalid or duplicate MCP tool definition.");
                }
                string description = tool.TryGetProperty("description", out var text) ? text.GetString() ?? "" : "";
                string mode = tool.TryGetProperty("ephemeral", out var ephemeral) ? ephemeral.GetString() ?? "" : "";
                if (mode is not ("" or "round" or "turn"))
                {
                    throw new InvalidDataException("Unsupported MCP ephemeral mode.");
                }
                string prefixMode = Mcp.AutomaticFrequency(name);
                if (prefixMode != "")
                {
                    if (mode != "" && mode != prefixMode)
                    {
                        throw new InvalidDataException("Conflicting MCP automatic frequency.");
                    }
                    mode = prefixMode;
                }
                discovered.Add(new McpToolDefinition(name, description, schema.Clone(), mode));
            }
            lock (gate)
            {
                ObjectDisposedException.ThrowIf(destroyed || ended, this);
                catalog = discovered.ToImmutable();
            }
        }
        catch
        {
            // Failed discovery cannot leave a half-connected live registration.
            await DestroyAsync();
            throw;
        }
    }

    public partial async Task<JsonElement> RequestAsync(string method, JsonElement parameters,
        CancellationToken cancellation)
    {
        ArgumentException.ThrowIfNullOrWhiteSpace(method);
        cancellation.ThrowIfCancellationRequested();
        var reply = new TaskCompletionSource<JsonElement>(TaskCreationOptions.RunContinuationsAsynchronously);
        long id;
        lock (gate)
        {
            ObjectDisposedException.ThrowIf(destroyed || ended, this);
            if (serving)
            {
                throw new InvalidOperationException("An MCP service cannot initiate outgoing RPC.");
            }
            id = checked(++nextID);
            pending.Add(id, reply);
            readerTask ??= Task.Run(ReadAsync);
        }
        try
        {
            var message = JsonSerializer.SerializeToElement(new { jsonrpc = "2.0", id, method, @params = parameters });
            try
            {
                await transport.SendAsync(message, cancellation);
            }
            catch (Exception error) when (error is IOException or OperationCanceledException or ObjectDisposedException)
            {
                // A failed write might have emitted only part of a frame. It
                // is unsafe to send another request on that byte stream.
                lock (gate)
                {
                    pending.Remove(id);
                }
                End(error);
                await DestroyAsync();
                throw;
            }
            try
            {
                return await reply.Task.WaitAsync(cancellation);
            }
            catch (OperationCanceledException) when (cancellation.IsCancellationRequested)
            {
                // Only explicit Job cancellation reaches this token. Cancelling
                // a foreground observation leaves the remote Job running.
                if (method == "tools/call" && !lifetime.IsCancellationRequested)
                {
                    await NotifyAsync("$/cancelRequest", JsonSerializer.SerializeToElement(new { id }),
                        lifetime.Token);
                }
                throw;
            }
        }
        finally
        {
            lock (gate)
            {
                pending.Remove(id);
            }
        }
    }

    public partial async Task NotifyAsync(string method, JsonElement parameters, CancellationToken cancellation)
    {
        lock (gate)
        {
            ObjectDisposedException.ThrowIf(destroyed || ended, this);
        }
        var message = JsonSerializer.SerializeToElement(new { jsonrpc = "2.0", method, @params = parameters });
        try
        {
            await transport.SendAsync(message, cancellation);
        }
        catch (Exception error) when (error is IOException or OperationCanceledException or ObjectDisposedException)
        {
            End(error);
            await DestroyAsync();
            throw;
        }
    }

    private partial async Task ReadAsync()
    {
        Exception terminal = new EndOfStreamException("MCP peer closed the connection.");
        try
        {
            while (await transport.ReceiveAsync(lifetime.Token) is { } message)
            {
                if (message.ValueKind != JsonValueKind.Object
                    || !message.TryGetProperty("jsonrpc", out var version) || version.GetString() != "2.0")
                {
                    throw new InvalidDataException("Invalid JSON-RPC envelope.");
                }
                if (message.TryGetProperty("method", out _))
                {
                    if (!message.TryGetProperty("id", out _))
                    {
                        continue; // Notifications never get a fabricated reply.
                    }
                    lock (gate)
                    {
                        reverseCalls.RemoveWhere(task => task.IsCompletedSuccessfully);
                        if (reverseCalls.Count >= 32)
                        {
                            throw new InvalidDataException("Too many concurrent MCP reverse requests.");
                        }
                        // Running separately is essential: a reverse request can
                        // await a local tool while this reader routes other replies.
                        reverseCalls.Add(Task.Run(() => ReverseAsync(message)));
                    }
                    continue;
                }
                if (!message.TryGetProperty("id", out var identity) || !identity.TryGetInt64(out long id))
                {
                    throw new InvalidDataException("MCP reply requires a numeric request identity.");
                }
                TaskCompletionSource<JsonElement>? reply;
                lock (gate)
                {
                    pending.Remove(id, out reply);
                }
                if (reply is null)
                {
                    continue; // A cancelled caller may have left a late reply.
                }
                if (message.TryGetProperty("error", out _))
                {
                    reply.TrySetException(new IOException("MCP peer returned an RPC error."));
                }
                else if (message.TryGetProperty("result", out var result))
                {
                    reply.TrySetResult(result.Clone());
                }
                else
                {
                    reply.TrySetException(new InvalidDataException("MCP reply has neither result nor error."));
                }
            }
        }
        catch (Exception error) when (error is IOException or InvalidDataException or JsonException or OperationCanceledException
            or ObjectDisposedException or InvalidOperationException or DecoderFallbackException)
        {
            terminal = error;
        }
        finally
        {
            End(terminal);
        }
    }

    private partial void End(Exception error)
    {
        lock (gate)
        {
            ended = true;
            foreach (var reply in pending.Values)
            {
                reply.TrySetException(error);
            }
            pending.Clear();
        }
    }

    private partial async Task ReverseAsync(JsonElement request)
    {
        JsonElement id = request.GetProperty("id").Clone();
        JsonElement response;
        try
        {
            string? method = request.GetProperty("method").GetString();
            var agent = Parent().Parent();
            if (method == "tools/list")
            {
                var tools = agent.Tools().Where(tool => agent.ToolEnabled(tool.Name()))
                    .Select(tool => tool.Declaration(Vendor.Anthropic)).ToArray();
                // Provider declaration and MCP discovery share schema data, but
                // their field names differ. Never mutate the stored definition.
                var definitions = tools.Select(tool => new
                {
                    name = tool["name"]!.GetValue<string>(),
                    description = tool["description"]!.GetValue<string>(),
                    inputSchema = tool["input_schema"]
                }).ToArray();
                response = JsonSerializer.SerializeToElement(new { jsonrpc = "2.0", id, result = new { tools = definitions } });
            }
            else if (method == "tools/call")
            {
                var parameters = request.GetProperty("params");
                string name = parameters.GetProperty("name").GetString()!;
                var args = parameters.GetProperty("arguments");
                var call = new ToolCallPart("mcp:" + id.GetRawText(), agent.Configuration().Target(), name, args);
                var result = await agent.DispatchReverseAsync(call, lifetime.Token);
                response = JsonSerializer.SerializeToElement(new
                {
                    jsonrpc = "2.0", id, result = McpContent.Encode(result)
                });
            }
            else
            {
                response = JsonSerializer.SerializeToElement(new
                {
                    jsonrpc = "2.0", id, error = new { code = -32601, message = "Method not found" }
                });
            }
        }
        catch (Exception error) when (error is ArgumentException or JsonException or KeyNotFoundException
            or InvalidOperationException or IOException or InvalidDataException or OperationCanceledException)
        {
            response = JsonSerializer.SerializeToElement(new
            {
                jsonrpc = "2.0", id, error = new { code = -32602, message = "Reverse request failed" }
            });
        }
        try
        {
            await transport.SendAsync(response, lifetime.Token);
        }
        catch (Exception error) when (error is IOException or OperationCanceledException or ObjectDisposedException)
        {
            // Disconnect is already terminating all pending RPC. There is no
            // remaining peer to receive this reverse reply.
        }
    }

    public partial Task DestroyAsync()
    {
        lock (gate)
        {
            // Close admission before returning the join task. Cleanup may not
            // have been scheduled yet, but no new RPC or parent may adopt us.
            destroyed = true;
            destruction ??= Task.Run(DestroyCoreAsync);
            return destruction;
        }
    }

    private partial async Task DestroyCoreAsync()
    {
        End(new ObjectDisposedException(nameof(McpClient)));
        lifetime.Cancel();
        try
        {
            try
            {
                await transport.DisposeAsync();
            }
            finally
            {
                try
                {
                    if (readerTask is not null)
                    {
                        try
                        {
                            await readerTask;
                        }
                        catch (Exception error) when (serving && error is IOException or InvalidDataException
                            or JsonException or OperationCanceledException or ObjectDisposedException or DecoderFallbackException)
                        {
                            // ServeAsync retains this expected wire/cancellation
                            // outcome for its caller. Joining it during cleanup
                            // must not turn it into a second shutdown failure.
                        }
                    }
                }
                finally
                {
                    Task[] reverse;
                    lock (gate)
                    {
                        reverse = reverseCalls.ToArray();
                    }
                    await Task.WhenAll(reverse);
                }
            }
        }
        finally
        {
            // Tool destruction completes both ownership relationships. The
            // connection cannot leave callable names behind after teardown.
            foreach (var tool in Tools())
            {
                tool.Destroy();
            }
            browserOwner?.RemoveClient(this);
            skillOwner?.RemoveClient(this);
            parent?.RemoveClient(this);
            lifetime.Dispose();
        }
    }
}
