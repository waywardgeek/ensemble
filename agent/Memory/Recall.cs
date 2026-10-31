using System.Collections.Immutable;

namespace Ensemble;

public sealed partial class Memory
{
    internal partial ImmutableArray<(string File, string Text, string Hash)> RecallFiles(
        CancellationToken cancellation)
    {
        lock (gate)
        {
            ObjectDisposedException.ThrowIf(destroyed, this);
            var owner = Parent();
            // A live, unstaged worker has not changed the source corpus. A
            // committed replay or any receipt may represent partially finalized
            // disk state; only actor-owned recovery can make that searchable.
            bool receiptPresent;
            try
            {
                using var receipt = owner.Sandbox().OpenRegularRead(HomePath(".memory-stage.json"));
                receiptPresent = true;
            }
            catch (FileNotFoundException) { receiptPresent = false; }
            catch (DirectoryNotFoundException) { receiptPresent = false; }
            if (owner.History().View().Compression?.Committed is not null || receiptPresent)
            {
                throw new InvalidOperationException("Recover memory before searching its corpus");
            }
            return RecallText.ReadCorpus(owner.Sandbox(), owner.DataDirectory(),
                RecallSource.Memory, cancellation);
        }
    }
}
