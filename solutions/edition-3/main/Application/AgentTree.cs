using System.Collections.Immutable;

namespace Ensemble;

public sealed partial class Ensemble
{
    public partial ImmutableArray<AgentTreeNode> AgentTree()
    {
        var roots = ImmutableArray.CreateBuilder<AgentTreeNode>();
        foreach (Agent agent in Agents())
        {
            // Ensemble owns descendants as well as standalone Agents. A child
            // remains a child through teardown: testing its mutable parent link
            // here could briefly promote a detaching child into a false root.
            if (agent.ChildID() is null)
            {
                roots.Add(CaptureTree(agent));
            }
        }
        return roots.ToImmutable();
    }

    private static partial AgentTreeNode CaptureTree(Agent agent)
    {
        var descendants = ImmutableArray.CreateBuilder<AgentTreeNode>();
        // Children returns a locked membership snapshot. No owner lock spans
        // recursion, status queries, I/O, or an asynchronous continuation.
        foreach (Agent child in agent.Children())
        {
            descendants.Add(CaptureTree(child));
        }
        return new AgentTreeNode(agent, descendants.ToImmutable());
    }
}
