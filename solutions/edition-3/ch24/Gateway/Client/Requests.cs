using System.Text.Json;

namespace Ensemble.Gateway;

public sealed partial class Client
{
    public partial async Task<JsonElement> RequestAsync(string method, object? parameters,
        CancellationToken cancellation)
    {
        ArgumentException.ThrowIfNullOrEmpty(method);
        cancellation.ThrowIfCancellationRequested();
        // Freeze caller data before admission; a mutable object is not borrowed
        // by the receive loop or retained as the request's authoritative state.
        JsonElement body = JsonSerializer.SerializeToElement(parameters);
        TaskCompletionSource<RpcMessage> waiter = new(TaskCreationOptions.RunContinuationsAsynchronously);
        ulong id;
        lock (gate)
        {
            ObjectDisposedException.ThrowIf(closed, this);
            id = checked(++nextID);
            pending.Add(id, waiter);
        }
        try
        {
            RpcMessage request = new(ID: JsonSerializer.SerializeToElement(id), Method: method, Params: body);
            await WriteAsync(new Frame("mcp_request", Rpc: request), cancellation);
            RpcMessage reply = await waiter.Task.WaitAsync(cancellation);
            if (reply.Error is { } error)
            {
                throw new InvalidOperationException($"Gateway RPC {error.Code}: {error.Message}");
            }
            return reply.Result!.Value.Clone();
        }
        finally
        {
            lock (gate)
            {
                pending.Remove(id);
            }
            // A caller can cancel after writing, or encoding can refuse before
            // writing. Settle its private promise too; a concurrent Close may
            // already have faulted it, whose failure we explicitly observe.
            waiter.TrySetCanceled();
            _ = waiter.Task.Exception;
        }
    }

    public partial async Task<JsonElement> CallToolAsync(string name, object? arguments,
        CancellationToken cancellation)
    {
        ArgumentException.ThrowIfNullOrEmpty(name);
        JsonElement result = await RequestAsync("tools/call", new { name, arguments }, cancellation);
        if (result.ValueKind != JsonValueKind.Object)
        {
            throw new InvalidDataException("Invalid Gateway tool result.");
        }
        if (result.TryGetProperty("isError", out JsonElement refused))
        {
            if (refused.ValueKind is not (JsonValueKind.True or JsonValueKind.False))
            {
                throw new InvalidDataException("Invalid Gateway tool error flag.");
            }
            if (refused.GetBoolean())
            {
                string message = "Gateway tool refused the operation.";
                if (result.TryGetProperty("content", out JsonElement content) && content.ValueKind == JsonValueKind.Array)
                {
                    foreach (JsonElement item in content.EnumerateArray())
                    {
                        if (item.ValueKind == JsonValueKind.Object && item.TryGetProperty("text", out JsonElement text)
                            && text.ValueKind == JsonValueKind.String)
                        {
                            message = text.GetString()!;
                            break;
                        }
                    }
                }
                throw new InvalidOperationException(message);
            }
        }
        return result;
    }

    public partial async Task SendAsync(OutboundMessage message, CancellationToken cancellation)
    {
        ArgumentNullException.ThrowIfNull(message);
        _ = await CallToolAsync("send-message", message, cancellation);
    }

    public partial Task RespondAsync(Envelope prompt, string content, CancellationToken cancellation)
    {
        ArgumentNullException.ThrowIfNull(prompt);
        ArgumentNullException.ThrowIfNull(content);
        return SendAsync(new OutboundMessage(prompt.Source.UserID, content, prompt.ConversationID), cancellation);
    }

    public partial async Task DeclareContactableAsync(IEnumerable<string> groups, CancellationToken cancellation)
    {
        ArgumentNullException.ThrowIfNull(groups);
        _ = await CallToolAsync("declare-contactable", new { groups = groups.ToArray() }, cancellation);
    }

    public partial async Task RegisterChildAsync(ChildRegistration child, CancellationToken cancellation)
    {
        ArgumentNullException.ThrowIfNull(child);
        if (child.PublicKey.Length != 32 || child.Groups.IsDefault)
        {
            throw new ArgumentException("Invalid child registration.", nameof(child));
        }
        _ = await CallToolAsync("register-child", new
        {
            agent_id = child.AgentID,
            public_key = Convert.ToBase64String(child.PublicKey.AsSpan()),
            groups = child.Groups
        }, cancellation);
    }
}
