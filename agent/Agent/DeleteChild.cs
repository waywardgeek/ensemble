namespace Ensemble;

public sealed partial class Agent
{
    public partial async Task DeleteChildAsync(string childID, CancellationToken cancellation)
    {
        Agent child = FindChild(childID)
            ?? throw new ArgumentException("No owned child has that ID.", nameof(childID));
        cancellation.ThrowIfCancellationRequested();
        // Retire before stopping: failure to update the durable registration must
        // not be acknowledged as deletion. Ordinary host shutdown keeps this file.
        child.RetireChildRestart();
        // From here removal is irreversible. Observe shutdown through Destroy even
        // when it failed, so all detach operations are attempted and failures surface.
        await child.ShutdownAsync().ConfigureAwait(ConfigureAwaitOptions.SuppressThrowing);
        child.Destroy();
    }
}
