namespace Ensemble;

public sealed partial class Agent
{
    internal partial Task ShutdownForPeerDeletionAsync()
    {
        lock (ingress)
        {
            ObjectDisposedException.ThrowIf(destroyed || destroying, this);
            // Match submission's bottom-up rule. Idle, failed and stopped are
            // not synonyms for a valid submission. Holding parent ingress also
            // excludes new child registration until ShutdownAsync closes it.
            string[] unfinished = children.Where(child => child.Submission() is null)
                .Select(child => child.childID!).ToArray();
            if (unfinished.Length != 0)
            {
                throw new InvalidOperationException("Join or remove unfinished children before deleting the peer: "
                    + string.Join(", ", unfinished));
            }
            return ShutdownAsync();
        }
    }
}
