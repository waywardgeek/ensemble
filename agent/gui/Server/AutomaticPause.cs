using Ensemble;

namespace EnsembleGui;

internal sealed partial class VoicePeer
{
    internal partial async Task SetPauseAsync(Agent target, bool held)
    {
        if (!held || !ReferenceEquals(pauseTarget, target)) await ReleasePauseAsync();
        if (!held || ReferenceEquals(pauseTarget, target)) return;
        await target.SetAutomaticPauseAsync(this, true);
        pauseTarget = target;
    }

    internal partial async Task ReleasePauseAsync()
    {
        Agent? target = pauseTarget;
        if (target is null) return;
        try { await target.SetAutomaticPauseAsync(this, false); }
        catch (ObjectDisposedException) { /* Agent destruction also clears transient owners. */ }
        finally { pauseTarget = null; }
    }
}
