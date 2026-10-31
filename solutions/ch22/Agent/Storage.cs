namespace Ensemble;

public sealed partial class Agent
{
    public partial string DataDirectory()
    {
        lock (ingress)
        {
            ObjectDisposedException.ThrowIf(destroyed, this);
            return dataDirectory ?? throw new InvalidOperationException(
                "Construct Agent with a data directory before using child storage.");
        }
    }
}
