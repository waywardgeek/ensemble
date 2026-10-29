namespace Ensemble;

public sealed partial class Agent
{
    public partial string? Address()
    {
        Ensemble? application;
        Agent? owner;
        lock (ingress)
        {
            // Destruction joins children before detaching ownership. Keep the
            // path available during that join so a finishing child cannot lose
            // its parent's provenance merely because cleanup has begun.
            if (destroyed)
            {
                return null;
            }
            application = parent;
            owner = agentParent;
        }
        if (application is null)
        {
            return null;
        }
        if (childID is not null)
        {
            // Release this child's lock before following its parent. Creation
            // locks parent then child; taking them in reverse would deadlock.
            string? parentAddress = owner?.Address();
            if (parentAddress is null || !ReferenceEquals(owner!.FindChild(childID), this))
            {
                return null;
            }
            return parentAddress + "/" + Uri.EscapeDataString(childID);
        }
        if (Name is null)
        {
            return null;
        }
        return Uri.EscapeDataString(Name);
    }
}
