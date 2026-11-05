// Grader-owned external I/O boundary. Candidate routing and lifetime code stay real.
internal sealed class ResponseBarrier
{
    internal readonly TaskCompletionSource Entered = new(TaskCreationOptions.RunContinuationsAsynchronously);
    internal readonly TaskCompletionSource Cancelled = new(TaskCreationOptions.RunContinuationsAsynchronously);
    internal readonly TaskCompletionSource Release = new(TaskCreationOptions.RunContinuationsAsynchronously);
    internal readonly TaskCompletionSource Done = new(TaskCreationOptions.RunContinuationsAsynchronously);
    private int claimed;
    internal bool Claim() => Interlocked.Exchange(ref claimed, 1) == 0;
    internal object Status() => new { entered = Entered.Task.IsCompleted, cancelled = Cancelled.Task.IsCompleted,
        done = Done.Task.IsCompleted };
    internal Stream Wrap(Stream borrowed) => new HeldResponse(borrowed, this);

    private sealed class HeldResponse(Stream borrowed, ResponseBarrier barrier) : Stream
    {
        public override async ValueTask WriteAsync(ReadOnlyMemory<byte> bytes, CancellationToken cancellationToken = default)
        {
            await borrowed.WriteAsync(bytes, cancellationToken);
            await borrowed.FlushAsync(cancellationToken);
            barrier.Entered.TrySetResult();
            try { await Task.Delay(Timeout.Infinite, cancellationToken); }
            catch (OperationCanceledException)
            {
                barrier.Cancelled.TrySetResult();
                await barrier.Release.Task;
                throw;
            }
        }
        public override bool CanRead => false;
        public override bool CanSeek => false;
        public override bool CanWrite => true;
        public override long Length => throw new NotSupportedException();
        public override long Position { get => throw new NotSupportedException(); set => throw new NotSupportedException(); }
        public override void Flush() => borrowed.Flush();
        public override int Read(byte[] buffer, int offset, int count) => throw new NotSupportedException();
        public override long Seek(long offset, SeekOrigin origin) => throw new NotSupportedException();
        public override void SetLength(long value) => throw new NotSupportedException();
        public override void Write(byte[] buffer, int offset, int count) => throw new NotSupportedException();
    }
}
