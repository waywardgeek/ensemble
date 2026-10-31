namespace Ensemble;

public sealed partial class Ensemble
{
    internal partial void Observe(Agent child, Observation value)
    {
        observations.Writer.TryWrite(new AgentObservation(child, value));
    }

    public partial async Task<AgentObservation> WaitAnyAsync(Func<AgentObservation, bool> predicate,
        CancellationToken cancellation)
    {
        await foreach (var value in observations.Reader.ReadAllAsync(cancellation))
        {
            if (predicate(value))
            {
                return value;
            }
        }
        throw new ObjectDisposedException(nameof(Ensemble));
    }
}
