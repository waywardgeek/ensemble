using System.Net.WebSockets;
using System.Runtime.ExceptionServices;
using System.Text;

namespace Ensemble.Gateway;

internal sealed partial class Session
{
    internal partial Session(RouterIdentity parent, WebSocket socket, CancellationToken cancellation)
    {
        this.socket = socket;
        lifetime = CancellationTokenSource.CreateLinkedTokenSource(cancellation);
        abortRegistration = lifetime.Token.Register(socket.Abort);
        try { parent.AddSession(this); }
        catch
        {
            abortRegistration.Dispose();
            lifetime.Dispose();
            writeGate.Dispose();
            throw;
        }
    }

    internal partial RouterIdentity Parent() => parent ?? throw new ObjectDisposedException(nameof(Session));
    internal partial bool HasNoise() => noise is not null;

    internal partial void AddNoise(NoiseKK child)
    {
        RouterIdentity identity = Parent();
        Router root = identity.Parent();
        lock (root.gate)
        {
            root.RequireLive();
            if (parent != identity || noise is not null || child.parent is not null || child.sessionParent is not null)
                throw new InvalidOperationException("Session Noise ownership is unavailable.");
            noise = child;
            child.sessionParent = this;
        }
    }

    internal partial void RemoveNoise(NoiseKK child)
    {
        Router root = Parent().Parent();
        lock (root.gate)
        {
            if (noise != child || child.sessionParent != this)
                throw new InvalidOperationException("Session Noise ownership mismatch.");
            noise = null;
            child.sessionParent = null;
        }
    }

    // Called under the root gate, together with the policy action it authorizes.
    private partial void RequireCurrent()
    {
        RouterIdentity identity = Parent();
        Router root = identity.Parent();
        root.RequireLive();
        root.RequireIdentity(identity);
        lifetime.Token.ThrowIfCancellationRequested();
        if (!identity.IsCurrent(this)) throw new InvalidOperationException("Session has been replaced.");
    }

    private partial async Task HandshakeAsync()
    {
        RouterIdentity identity = Parent();
        Router root = identity.Parent();
        using CancellationTokenSource startup = CancellationTokenSource.CreateLinkedTokenSource(lifetime.Token);
        startup.CancelAfter(TimeSpan.FromSeconds(10));
        _ = new NoiseKK(this, NoiseRole.Responder, root.StaticKey(), identity.publicKey.AsSpan(),
            Encoding.UTF8.GetBytes("ensemble-gateway-v1:" + identity.id));
        noise!.ReadHandshakeMessage(await Wire.ReadBinaryAsync(socket, startup.Token));
        byte[] second = noise.WriteHandshakeMessage([]);
        await socket.SendAsync(second.AsMemory(), WebSocketMessageType.Binary, true, startup.Token);
        Frame hello = await ReadAsync(startup.Token);
        if (hello.Type != "hello" || hello.Version != 1 || hello.AgentID != identity.id)
            throw new InvalidDataException("Invalid authenticated Gateway hello.");

        Session? previous;
        WelcomePayload welcome;
        lock (root.gate)
        {
            startup.Token.ThrowIfCancellationRequested();
            (previous, welcome, sent) = identity.Activate(this, hello.Seq);
        }
        // Authority changed atomically above; closing a previous socket must not
        // hold the policy lock. Its cleanup cannot remove the new current link.
        previous?.Close();
        await WriteAsync(new Frame("welcome", Welcome: welcome), startup.Token);
    }

    private partial async Task<Frame> ReadAsync(CancellationToken cancellation)
    {
        byte[] encrypted = await Wire.ReadBinaryAsync(socket, cancellation);
        return Wire.Decode(noise!.Decrypt(encrypted, []));
    }

    private partial async Task WriteAsync(Frame frame, CancellationToken cancellation)
    {
        byte[] plain = Wire.Encode(frame);
        using CancellationTokenSource write = CancellationTokenSource.CreateLinkedTokenSource(lifetime.Token, cancellation);
        write.CancelAfter(TimeSpan.FromSeconds(10));
        await writeGate.WaitAsync(write.Token);
        try
        {
            write.Token.ThrowIfCancellationRequested();
            Router root = Parent().Parent();
            lock (root.gate)
            {
                RequireCurrent();
                // Publish the frontier before the send can reach a fast peer.
                // A send failure is terminal and leaves unacked data queued.
                if (frame.Type == "deliver") sent = frame.Envelope!.Seq;
            }
            await SendLockedAsync(plain, write.Token);
        }
        finally { writeGate.Release(); }
    }

    private partial async Task SendLockedAsync(byte[] plain, CancellationToken cancellation)
    {
        try
        {
            byte[] encrypted = noise!.Encrypt(plain, []);
            await socket.SendAsync(encrypted.AsMemory(), WebSocketMessageType.Binary, true, cancellation);
        }
        catch
        {
            // Advancing a send nonce is not reversible. Neither a cancelled send
            // nor a partial WebSocket write permits reuse of this connection.
            Close();
            throw;
        }
    }

    private partial async Task DeliverAsync()
    {
        try
        {
            while (await wake.Reader.WaitToReadAsync(lifetime.Token))
            {
                while (wake.Reader.TryRead(out _)) { }
                while (true)
                {
                    Envelope? message;
                    RouterIdentity identity = Parent();
                    Router root = identity.Parent();
                    lock (root.gate)
                    {
                        RequireCurrent();
                        message = identity.Queue().After(sent);
                    }
                    if (message is null) break;
                    await WriteAsync(new Frame("deliver", Envelope: message), lifetime.Token);
                }
            }
        }
        finally { Close(); }
    }

    private partial async Task ReceiveAsync()
    {
        while (true)
        {
            Frame frame = await ReadAsync(lifetime.Token);
            Frame? response = null;
            RouterIdentity identity = Parent();
            Router root = identity.Parent();
            lock (root.gate)
            {
                RequireCurrent();
                switch (frame.Type)
                {
                    case "ack":
                        if (frame.Seq > sent) throw new InvalidDataException("Receipt exceeds this session's delivery frontier.");
                        identity.Queue().Acknowledge(frame.Seq);
                        break;
                    case "ping":
                        response = new Frame("pong");
                        break;
                    case "mcp_request":
                        if (frame.Rpc is null) throw new InvalidDataException("Missing MCP request.");
                        RpcMessage? reply = root.Request(identity, frame.Rpc);
                        if (reply is not null) response = new Frame("mcp_response", Rpc: reply);
                        break;
                    case "response":
                        if (frame.Message is null) throw new InvalidDataException("Missing response message.");
                        try { root.Route(identity, frame.Message); }
                        catch (Exception error) when (error is ArgumentException or InvalidOperationException)
                        {
                            response = new Frame("error", Error: error.Message);
                        }
                        break;
                    default:
                        response = new Frame("error", Error: "unknown frame type");
                        break;
                }
            }
            if (response is not null) await WriteAsync(response, lifetime.Token);
        }
    }

    internal partial void Notify() => wake.Writer.TryWrite(0);
    internal partial Task Done() => done.Task;

    internal partial void Close()
    {
        wake.Writer.TryComplete();
        try { lifetime.Cancel(); }
        catch (ObjectDisposedException) { /* Cleanup already joined this connection. */ }
    }

    internal partial async Task RunAsync()
    {
        Exception? failure = null;
        List<Exception> cleanupFailures = [];
        try
        {
            await HandshakeAsync();
            delivery = DeliverAsync();
            Notify();
            await ReceiveAsync();
        }
        catch (OperationCanceledException) when (lifetime.IsCancellationRequested) { }
        catch (EndOfStreamException) { /* A peer's close frame ends its session. */ }
        catch (Exception error) { failure = error; }
        finally
        {
            try { Close(); }
            catch (Exception error) { cleanupFailures.Add(error); }
            if (delivery is not null)
            {
                try { await delivery; }
                catch (OperationCanceledException) when (lifetime.IsCancellationRequested) { }
                catch (Exception error) { failure ??= error; }
            }
            // No loop can use Noise or the write gate after this join. Attempt
            // every cleanup and retain failures instead of stranding shutdown.
            try { noise?.Dispose(); }
            catch (Exception error) { cleanupFailures.Add(error); }
            try { abortRegistration.Dispose(); }
            catch (Exception error) { cleanupFailures.Add(error); }
            try { socket.Dispose(); }
            catch (Exception error) { cleanupFailures.Add(error); }
            try { lifetime.Dispose(); }
            catch (Exception error) { cleanupFailures.Add(error); }
            try { writeGate.Dispose(); }
            catch (Exception error) { cleanupFailures.Add(error); }
            try { Parent().RemoveSession(this); }
            catch (Exception error) { cleanupFailures.Add(error); }
            if (cleanupFailures.Count == 0) done.TrySetResult();
            else
            {
                done.TrySetException(new AggregateException(cleanupFailures));
                // A normally ended request may already have left the identity's
                // child set. Its HTTP caller receives this failure below even if
                // there is no concurrent Router disposer awaiting the private task.
                _ = done.Task.Exception;
            }
        }
        // The HTTP caller observes peer/protocol errors. Router shutdown joins
        // Done, whose failures are cleanup failures, not malicious peer input.
        if (cleanupFailures.Count != 0)
        {
            if (failure is not null) cleanupFailures.Insert(0, failure);
            throw new AggregateException(cleanupFailures);
        }
        if (failure is not null) ExceptionDispatchInfo.Capture(failure).Throw();
    }
}
