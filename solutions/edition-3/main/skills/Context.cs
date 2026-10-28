using System.Collections.Immutable;

namespace Ensemble;

public sealed partial class Context
{
    public partial ImmutableHashSet<string> EnabledTools(Configuration? policy)
    {
        return Dialogue.SelectMany(entry => entry.Parts).OfType<SkillPart>()
            .Where(skill => !skill.Pending && (policy is null || policy.AllowsSkill(skill.Name)))
            .SelectMany(skill => skill.Tools)
            .Where(name => policy is null || policy.AllowsTool(name))
            .ToImmutableHashSet(StringComparer.Ordinal);
    }

    public partial ImmutableHashSet<string> RevokedTools(Configuration? policy)
    {
        // A tool can have several granting parents. Unloading one manual must not
        // revoke the authority of another manual that still grants the same tool.
        var restricted = Dialogue.SelectMany(entry => entry.Parts).OfType<SkillPart>()
            .SelectMany(skill => skill.Tools.Where(name => skill.Pending
                || (policy is not null && (!policy.AllowsSkill(skill.Name) || !policy.AllowsTool(name)))));
        return restricted.ToImmutableHashSet(StringComparer.Ordinal).Except(EnabledTools(policy));
    }
}
