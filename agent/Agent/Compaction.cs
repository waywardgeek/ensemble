namespace Ensemble;

public sealed partial class Agent
{
    public partial int ContextByteBudget()
    {
        lock (ingress)
        {
            return contextByteBudget;
        }
    }

    public partial void SetContextByteBudget(int bytes)
    {
        ArgumentOutOfRangeException.ThrowIfNegative(bytes);
        lock (ingress)
        {
            ObjectDisposedException.ThrowIf(destroyed, this);
            contextByteBudget = bytes;
        }
    }

}
