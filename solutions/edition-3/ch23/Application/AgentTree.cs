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

    public partial Agent? FindAgentAddress(string address)
    {
        ArgumentException.ThrowIfNullOrWhiteSpace(address);
        string[] components = address.Split('/');
        var names = new string[components.Length];
        for (int index = 0; index < components.Length; index++)
        {
            string component = components[index];
            string name = Uri.UnescapeDataString(component);
            // Reject malformed escapes and alternate spellings. In particular,
            // an encoded slash belongs to a name, not another ownership level.
            if (string.IsNullOrWhiteSpace(name)
                || Uri.EscapeDataString(name) != component)
            {
                return null;
            }
            names[index] = name;
        }
        Agent? selected = FindAgent(names[0]);
        if (selected is null || selected.ChildID() is not null)
        {
            return null;
        }
        for (int index = 1; index < names.Length; index++)
        {
            selected = selected.FindChild(names[index]);
            if (selected is null)
            {
                return null;
            }
        }
        return selected;
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
        return new AgentTreeNode(agent, descendants.ToImmutable(), agent.Address());
    }
}
