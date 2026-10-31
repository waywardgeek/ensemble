using System.Collections.Immutable;

namespace Ensemble;

public sealed partial class Agent
{
    internal partial bool IsAutomaticTool(string name)
    {
        if (global::Ensemble.Mcp.AutomaticFrequency(name) != "")
        {
            return true;
        }
        // Compatibility annotations remain catalog facts, not a second registry.
        return Mcp()?.Clients().Any(client => client.Catalog()
            .Any(definition => definition.Name == name && definition.Ephemeral != "")) == true;
    }
}

public sealed partial class Mcp
{
    internal static partial string AutomaticFrequency(string name)
    {
        if (name.StartsWith("per_turn_", StringComparison.Ordinal))
        {
            return "turn";
        }
        if (name.StartsWith("per_round_", StringComparison.Ordinal))
        {
            return "round";
        }
        return "";
    }

    internal partial ImmutableArray<AutomaticTool> AutomaticTools(bool includeTurn)
    {
        var selected = ImmutableArray.CreateBuilder<AutomaticTool>();
        foreach (var client in Clients())
        {
            var owner = client.SkillOwner();
            // A detached granting owner never widens authority to the catalog.
            if (client.SkillName is not null && owner is null)
            {
                continue;
            }
            foreach (var definition in client.Catalog())
            {
                bool frequencyMatches = definition.Ephemeral == "round"
                    || (includeTurn && definition.Ephemeral == "turn");
                bool granted = owner is not null
                    ? owner.Active().Any(skill => skill.Name == client.SkillName
                        && !skill.Pending && Parent().Configuration().AllowsSkill(skill.Name)
                        && skill.Tools.Contains(definition.Name))
                    : Parent().SkillManager() is null;
                if (frequencyMatches && granted && Parent().Configuration().AllowsTool(definition.Name))
                {
                    selected.Add(new AutomaticTool(client, definition));
                }
            }
        }
        return selected.ToImmutable();
    }
}
