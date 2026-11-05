using Ensemble;

// The external sink is controlled; candidate ownership, file access and shutdown
// remain real. No private fields or candidate test classes are consulted.
internal static class TransferChecks
{
    private sealed class HeldSink : Stream
    {
        internal readonly TaskCompletionSource Entered = new(TaskCreationOptions.RunContinuationsAsynchronously);
        internal readonly TaskCompletionSource Cancelled = new(TaskCreationOptions.RunContinuationsAsynchronously);
        internal readonly TaskCompletionSource Release = new(TaskCreationOptions.RunContinuationsAsynchronously);
        internal bool Disposed;
        public override async ValueTask WriteAsync(ReadOnlyMemory<byte> buffer, CancellationToken cancellationToken = default)
        {
            Entered.TrySetResult();
            try { await Task.Delay(Timeout.Infinite, cancellationToken); }
            catch (OperationCanceledException)
            {
                Cancelled.TrySetResult();
                await Release.Task;
                throw;
            }
        }
        protected override void Dispose(bool disposing) { Disposed = true; base.Dispose(disposing); }
        public override bool CanRead => false;
        public override bool CanSeek => false;
        public override bool CanWrite => true;
        public override long Length => throw new NotSupportedException();
        public override long Position { get => throw new NotSupportedException(); set => throw new NotSupportedException(); }
        public override void Flush() { }
        public override int Read(byte[] buffer, int offset, int count) => throw new NotSupportedException();
        public override long Seek(long offset, SeekOrigin origin) => throw new NotSupportedException();
        public override void SetLength(long value) => throw new NotSupportedException();
        public override void Write(byte[] buffer, int offset, int count) => throw new NotSupportedException();
    }
    internal static async Task RunAsync(Agent owner, WorkProductInfo offer, byte[] expected,
        Action<bool, string> require, CancellationToken deadline)
    {
        using var caller = new CancellationTokenSource();
        using var first = new HeldSink();
        using var second = new HeldSink();
        try
        {
            var copy = owner.CopyWorkProductAsync(offer.ID, first, caller.Token);
            await first.Entered.Task.WaitAsync(deadline);
            using var independent = new MemoryStream();
            await Task.Run(() => owner.CopyWorkProductAsync(offer.ID, independent, deadline)).WaitAsync(deadline);
            require(independent.ToArray().SequenceEqual(expected), "blocked-copy-does-not-block-another");
            caller.Cancel();
            await first.Cancelled.Task.WaitAsync(deadline);
            require(!copy.IsCompleted, "caller-cancel-joins-copy-cleanup");
            first.Release.TrySetResult();
            bool cancelled = false;
            try { await copy.WaitAsync(deadline); }
            catch (OperationCanceledException) when (!deadline.IsCancellationRequested) { cancelled = true; }
            require(cancelled && !first.Disposed, "caller-cancel-preserves-borrowed-sink");
            var active = owner.CopyWorkProductAsync(offer.ID, second);
            await second.Entered.Task.WaitAsync(deadline);
            var destroy = Task.Run(owner.Destroy);
            await second.Cancelled.Task.WaitAsync(deadline);
            require(!destroy.IsCompleted, "agent-destruction-joins-transfer-cleanup");
            second.Release.TrySetResult();
            cancelled = false;
            try { await active.WaitAsync(deadline); }
            catch (OperationCanceledException) when (!deadline.IsCancellationRequested) { cancelled = true; }
            await destroy.WaitAsync(deadline);
            require(cancelled && !second.Disposed, "agent-destruction-cancels-without-owning-sink");
        }
        finally
        {
            first.Release.TrySetResult();
            second.Release.TrySetResult();
            caller.Cancel();
            owner.Destroy();
        }
    }
}
