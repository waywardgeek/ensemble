namespace Ensemble;

public sealed partial class Agent
{
    public partial Jobs Jobs() => jobs ?? throw new ObjectDisposedException(nameof(Jobs));

    internal partial void AddJobs(Jobs child)
    {
        if (destroyed || jobs is not null || child.parent is not null || child.destroyed)
        {
            throw new InvalidOperationException("Jobs already owned or destroyed");
        }
        jobs = child;
        child.parent = this;
    }

    internal partial void RemoveJobs(Jobs child)
    {
        if (jobs != child || child.parent != this)
        {
            throw new InvalidOperationException("Jobs belongs to another Agent");
        }
        jobs = null;
        child.parent = null;
    }
}
