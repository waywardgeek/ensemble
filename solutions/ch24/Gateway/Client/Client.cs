using System.Text;
using System.Text.Json;
using System.Net.WebSockets;

namespace Ensemble.Gateway;

public partial class NopHandler
{
    public virtual partial void OnConnected(WelcomePayload welcome) { }
    public virtual partial void OnDeliver(Envelope envelope) { }
    public virtual partial void OnDisconnected(string reason) { }
    public virtual partial void OnError(Exception error) { }
}

public sealed partial class Client
{
    private partial Client(ClientConfig config, IHandler handler, CancellationToken cancellation)
    {
        this.config = config;
        this.handler = handler;
        lastAck = config.LastAckedSeq;
        lifetime = CancellationTokenSource.CreateLinkedTokenSource(cancellation);
        // Cancellation interrupts an actual outstanding socket read/write. It is
        // not merely cancellation of the caller's wait for a background operation.
        abortRegistration = lifetime.Token.Register(socket.Abort);
    }

    public static partial async Task<Client> ConnectAsync(ClientConfig config, IHandler handler,
        CancellationToken cancellation)
    {
        ArgumentNullException.ThrowIfNull(config);
        ArgumentNullException.ThrowIfNull(handler);
        ArgumentNullException.ThrowIfNull(config.StaticKey);
        if (!Uri.TryCreate(config.GatewayURL, UriKind.Absolute, out Uri? address)
            || address.Scheme is not ("ws" or "wss") || address.UserInfo.Length != 0
            || address.Fragment.Length != 0 || string.IsNullOrEmpty(config.AgentID)
            || config.GatewayPubKey.Length != 32 || config.HandshakeTimeout < TimeSpan.Zero)
        {
            throw new ArgumentException("Invalid Gateway connection configuration.", nameof(config));
        }
        Client client = new(config, handler, cancellation);
        try
        {
            await client.HandshakeAsync(cancellation);
            // Run independently: a synchronous application callback must not hold
            // ConnectAsync's caller after authenticated welcome has been reported.
            _ = Task.Run(client.RunAsync);
            return client;
        }
        catch
        {
            client.Close();
            await client.ReleaseAsync();
            client.done.TrySetResult();
            throw;
        }
    }

    private partial async Task HandshakeAsync(CancellationToken cancellation)
    {
        TimeSpan timeout = config.HandshakeTimeout == TimeSpan.Zero
            ? TimeSpan.FromSeconds(10) : config.HandshakeTimeout;
        using CancellationTokenSource startup = CancellationTokenSource.CreateLinkedTokenSource(lifetime.Token, cancellation);
        startup.CancelAfter(timeout);
        UriBuilder address = new(config.GatewayURL);
        // Replace any existing selector, rather than sending ambiguous duplicate
        // identities. Other query fields are preserved as URL data.
        IEnumerable<string> fields = address.Query.TrimStart('?').Split('&', StringSplitOptions.RemoveEmptyEntries)
            .Where(field => Uri.UnescapeDataString(field.Split('=')[0]) != "agent_id");
        address.Query = string.Join('&', fields.Append("agent_id=" + Uri.EscapeDataString(config.AgentID)));
        await socket.ConnectAsync(address.Uri, startup.Token);
        _ = new NoiseKK(this, NoiseRole.Initiator, config.StaticKey, config.GatewayPubKey.AsSpan(),
            Encoding.UTF8.GetBytes("ensemble-gateway-v1:" + config.AgentID));
        byte[] first = noise!.WriteHandshakeMessage([]);
        await socket.SendAsync(first.AsMemory(), WebSocketMessageType.Binary, true, startup.Token);
        noise.ReadHandshakeMessage(await Wire.ReadBinaryAsync(socket, startup.Token));
        await WriteAsync(new Frame("hello", Version: 1, AgentID: config.AgentID, Seq: lastAck), startup.Token);
        Frame frame = await ReadAsync(startup.Token);
        if (frame.Type != "welcome" || frame.Welcome is not { } welcome
            || welcome.AgentID != config.AgentID || welcome.QueueDepth < 0
            || string.IsNullOrEmpty(welcome.ServerVersion))
        {
            throw new InvalidDataException("Invalid authenticated Gateway welcome.");
        }
        handler.OnConnected(welcome);
    }

    internal partial void AddNoise(NoiseKK child)
    {
        lock (gate)
        {
            if (closed || noise is not null || child.parent is not null || child.sessionParent is not null)
            {
                throw new InvalidOperationException("Gateway Noise ownership is unavailable.");
            }
            noise = child;
            child.parent = this;
        }
    }

    internal partial void RemoveNoise(NoiseKK child)
    {
        lock (gate)
        {
            if (!ReferenceEquals(noise, child) || !ReferenceEquals(child.parent, this))
            {
                throw new InvalidOperationException("Gateway Noise owner mismatch.");
            }
            noise = null;
            child.parent = null;
        }
    }

    private partial async Task WriteAsync(Frame frame, CancellationToken cancellation)
    {
        // Encode/size-check before touching the nonce. An oversized local request
        // is a refused call, not a reason to destroy a healthy connection.
        byte[] plain = Wire.Encode(frame);
        using CancellationTokenSource write = CancellationTokenSource.CreateLinkedTokenSource(lifetime.Token, cancellation);
        write.CancelAfter(TimeSpan.FromSeconds(10));
        await writeGate.WaitAsync(write.Token);
        try
        {
            write.Token.ThrowIfCancellationRequested();
            byte[] encrypted = noise!.Encrypt(plain, []);
            try
            {
                await socket.SendAsync(encrypted.AsMemory(), WebSocketMessageType.Binary, true, write.Token);
            }
            catch
            {
                // Encryption consumed a nonce. Neither cancellation nor a partial
                // write permits continuing this session with the next request.
                Close();
                throw;
            }
        }
        finally
        {
            writeGate.Release();
        }
    }

    private partial async Task<Frame> ReadAsync(CancellationToken cancellation)
    {
        byte[] encrypted = await Wire.ReadBinaryAsync(socket, cancellation);
        return Wire.Decode(noise!.Decrypt(encrypted, []));
    }

    private partial async Task DeliverAsync(Envelope envelope)
    {
        if (envelope.Seq == 0 || envelope.ToAgent != config.AgentID || string.IsNullOrEmpty(envelope.ID)
            || envelope.Content is null || envelope.Timestamp == default || envelope.Source is not { } source
            || string.IsNullOrEmpty(source.UserID) || string.IsNullOrEmpty(source.ChannelID)
            || source.Type != "agent" || source.Trust is not ("owner" or "trusted" or "external"))
        {
            throw new InvalidDataException("Invalid Gateway delivery envelope.");
        }
        ulong acknowledged = LastAckedSeq();
        if (envelope.Seq > acknowledged)
        {
            handler.OnDeliver(envelope);
            await WriteAsync(new Frame("ack", Seq: envelope.Seq), lifetime.Token);
            lock (gate)
            {
                lastAck = envelope.Seq;
            }
        }
        else
        {
            // Resume/duplicate suppression is receipt-based, not exactly-once
            // execution. Applications persist their own accepted-work boundary.
            await WriteAsync(new Frame("ack", Seq: acknowledged), lifetime.Token);
        }
    }

    private partial void ReceiveReply(RpcMessage reply)
    {
        if (reply.Jsonrpc != "2.0" || reply.ID is not { } id || id.ValueKind != JsonValueKind.Number
            || !id.TryGetUInt64(out ulong number) || (reply.Result is null) == (reply.Error is null))
        {
            throw new InvalidDataException("Invalid Gateway RPC response.");
        }
        TaskCompletionSource<RpcMessage>? waiter;
        lock (gate)
        {
            pending.Remove(number, out waiter);
        }
        // A caller may have cancelled after sending. Its late response is not
        // another request's result and does not require a connection restart.
        waiter?.TrySetResult(reply);
    }

    private partial async Task RunAsync()
    {
        Exception? failure = null;
        string reason = "closed";
        try
        {
            while (true)
            {
                Frame frame = await ReadAsync(lifetime.Token);
                switch (frame.Type)
                {
                    case "deliver" when frame.Envelope is not null:
                        await DeliverAsync(frame.Envelope);
                        break;
                    case "mcp_response" when frame.Rpc is not null:
                        ReceiveReply(frame.Rpc);
                        break;
                    case "ping":
                        await WriteAsync(new Frame("pong"), lifetime.Token);
                        break;
                    case "pong":
                        break;
                    case "error" when frame.Error is not null:
                        handler.OnError(new InvalidOperationException(frame.Error));
                        break;
                    default:
                        throw new InvalidDataException("Unexpected Gateway frame.");
                }
            }
        }
        catch (Exception error)
        {
            if (!lifetime.IsCancellationRequested)
            {
                reason = error.Message;
                if (error is not EndOfStreamException)
                {
                    failure = error;
                }
            }
        }
        finally
        {
            Close();
            Exception? releaseFailure = null;
            try
            {
                await ReleaseAsync();
            }
            catch (Exception error)
            {
                releaseFailure = error;
            }
            try
            {
                try
                {
                    if (failure is not null)
                    {
                        handler.OnError(failure);
                    }
                }
                finally
                {
                    handler.OnDisconnected(reason);
                }
                if (releaseFailure is null)
                {
                    done.TrySetResult();
                }
                else
                {
                    done.TrySetException(releaseFailure);
                }
            }
            catch (Exception callbackFailure)
            {
                // Callback failure cannot prevent resource release or leave Done
                // hanging; joined callers can observe this application error.
                done.TrySetException(releaseFailure is null ? callbackFailure
                    : new AggregateException(releaseFailure, callbackFailure));
            }
        }
    }

    public partial ulong LastAckedSeq()
    {
        lock (gate)
        {
            return lastAck;
        }
    }

    public partial void Close()
    {
        TaskCompletionSource<RpcMessage>[] waiting;
        lock (gate)
        {
            if (closed)
            {
                return;
            }
            closed = true;
            waiting = pending.Values.ToArray();
            pending.Clear();
        }
        foreach (TaskCompletionSource<RpcMessage> waiter in waiting)
        {
            waiter.TrySetException(new IOException("Gateway connection closed."));
        }
        try
        {
            lifetime.Cancel();
        }
        catch (ObjectDisposedException)
        {
            // The receive loop may have finished releasing the connection after
            // our closed-state publication but before this cancellation call.
        }
    }

    private partial async Task ReleaseAsync()
    {
        // Close interrupts writes; join any write already in its nonce critical
        // section before destroying the child. Future writes see cancellation.
        await writeGate.WaitAsync();
        try
        {
            List<Exception> failures = [];
            void Release(Action operation)
            {
                try
                {
                    operation();
                }
                catch (Exception error)
                {
                    failures.Add(error);
                }
            }
            // Attempt every owned resource even if an earlier disposal fails.
            Release(() => noise?.Dispose());
            Release(socket.Dispose);
            Release(abortRegistration.Dispose);
            Release(lifetime.Dispose);
            if (failures.Count != 0)
            {
                throw new AggregateException("Gateway connection cleanup failed.", failures);
            }
        }
        finally
        {
            writeGate.Release();
        }
    }

    public partial Task Done() => done.Task;

    public partial async ValueTask DisposeAsync()
    {
        Close();
        await Done();
    }
}
