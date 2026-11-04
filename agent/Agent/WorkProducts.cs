namespace Ensemble;

public sealed partial class Agent
{
    public partial WorkProductInfo OfferWorkProduct(string path)
    {
        ArgumentException.ThrowIfNullOrWhiteSpace(path);
        lock (workProductGate)
        {
            ObjectDisposedException.ThrowIf(workProductsClosed || destroyed, this);
            if (workProducts.Count >= MaximumWorkProductOffers)
            {
                throw new InvalidOperationException("This Agent has reached its 64 work-product offer limit.");
            }
            // Resolve is a value operation, never an open grant. Both this check
            // and every later download open repeat the Sandbox's file policy.
            string resolved = Sandbox().Resolve(path);
            using var input = Sandbox().OpenRegularRead(resolved);
            long offeredBytes = input.Length;
            if (offeredBytes > MaximumWorkProductBytes)
            {
                throw new IOException("Work product exceeds the 64 MiB transfer limit.");
            }
            string name = System.IO.Path.GetFileName(resolved);
            if (string.IsNullOrWhiteSpace(name) || name.Any(char.IsControl))
            {
                throw new ArgumentException("Work product needs a displayable filename.", nameof(path));
            }
            var info = new WorkProductInfo(Guid.NewGuid().ToString("N"), name, offeredBytes);
            workProducts.Add(info.ID, new WorkProductOffer(info, resolved));
            return info;
        }
    }

    public partial FileStream OpenWorkProduct(string id)
    {
        ArgumentException.ThrowIfNullOrWhiteSpace(id);
        lock (workProductGate)
        {
            ObjectDisposedException.ThrowIf(workProductsClosed || destroyed, this);
            if (!workProducts.TryGetValue(id, out var offer))
            {
                throw new KeyNotFoundException("Work-product offer is unavailable.");
            }
            var input = Sandbox().OpenRegularRead(offer.Path);
            try
            {
                if (input.Length > MaximumWorkProductBytes)
                {
                    throw new IOException("Work product exceeds the 64 MiB transfer limit.");
                }
                return input;
            }
            catch
            {
                input.Dispose();
                throw;
            }
        }
    }

    private sealed partial class WorkProductTransfer
    {
        internal partial WorkProductTransfer(Agent parent, FileStream input, CancellationToken cancellation)
        {
            // The caller holds the Agent's gate and owns input until registration
            // succeeds. Refusal leaves both membership directions unchanged.
            ObjectDisposedException.ThrowIf(parent.workProductsClosed || parent.destroyed, parent);
            stop = CancellationTokenSource.CreateLinkedTokenSource(parent.workProductStop.Token, cancellation);
            this.input = input;
            parent.workProductTransfers.Add(this);
            this.parent = parent;
        }
    }

    public partial async Task<long> CopyWorkProductAsync(string id, Stream destination, CancellationToken cancellation)
    {
        ArgumentNullException.ThrowIfNull(destination);
        cancellation.ThrowIfCancellationRequested();
        WorkProductTransfer transfer;
        lock (workProductGate)
        {
            var input = OpenWorkProduct(id);
            try { transfer = new WorkProductTransfer(this, input, cancellation); }
            catch { input.Dispose(); throw; }
        }
        try
        {
            var token = transfer.stop.Token;
            token.ThrowIfCancellationRequested();
            long length = transfer.input.Length;
            if (length > MaximumWorkProductBytes)
                throw new IOException("Work product exceeds the 64 MiB transfer limit.");
            long remaining = length;
            byte[] buffer = new byte[64 * 1024];
            while (remaining > 0)
            {
                int read = await transfer.input.ReadAsync(buffer.AsMemory(0,
                    (int)Math.Min(buffer.Length, remaining)), token).ConfigureAwait(false);
                if (read == 0) throw new EndOfStreamException("Work product changed during transfer.");
                await destination.WriteAsync(buffer.AsMemory(0, read), token).ConfigureAwait(false);
                remaining -= read;
            }
            // Refuse growth instead of silently offering a truncated successful
            // download. Same-size concurrent edits are not an immutable snapshot.
            if (await transfer.input.ReadAsync(buffer.AsMemory(0, 1), token).ConfigureAwait(false) != 0)
                throw new IOException("Work product changed during transfer.");
            return length;
        }
        finally
        {
            try { await transfer.input.DisposeAsync().ConfigureAwait(false); }
            finally
            {
                transfer.stop.Dispose();
                lock (workProductGate)
                {
                    workProductTransfers.Remove(transfer);
                    transfer.parent = null;
                }
                transfer.joined.TrySetResult();
            }
        }
    }

    private partial void CloseWorkProducts()
    {
        WorkProductTransfer[] transfers;
        lock (workProductGate)
        {
            if (workProductsClosed) return;
            workProductsClosed = true;
            workProducts.Clear();
            transfers = workProductTransfers.ToArray();
        }
        // Cancellation callbacks and joining can execute foreign stream code.
        // Neither runs while holding the ownership gate or Sandbox lifetime lock.
        try { workProductStop.Cancel(); }
        finally
        {
            try { Task.WhenAll(transfers.Select(item => item.joined.Task)).GetAwaiter().GetResult(); }
            finally { workProductStop.Dispose(); }
        }
    }
}
