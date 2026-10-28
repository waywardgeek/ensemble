using System.Collections.Immutable;

namespace Ensemble;

public sealed partial class Context
{
    public partial ImmutableHashSet<string> EnabledTools()
    {
        return Dialogue.SelectMany(entry => entry.Parts).OfType<SkillPart>()
            .Where(skill => !skill.Pending).SelectMany(skill => skill.Tools)
            .ToImmutableHashSet(StringComparer.Ordinal);
    }

    public partial ImmutableHashSet<string> RevokedTools()
    {
        // A tool can have several granting parents. Unloading one manual must not
        // revoke the authority of another manual that still grants the same tool.
        var pending = Dialogue.SelectMany(entry => entry.Parts).OfType<SkillPart>()
            .Where(skill => skill.Pending).SelectMany(skill => skill.Tools);
        return pending.ToImmutableHashSet(StringComparer.Ordinal).Except(EnabledTools());
    }
}
