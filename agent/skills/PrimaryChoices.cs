using System.Collections.Immutable;

namespace Ensemble;

public sealed partial class Skills
{
    public partial ImmutableArray<PrimarySkillChoice> PrimaryChoices()
    {
        lock (gate)
        {
            ObjectDisposedException.ThrowIf(destroyed, this);
            return definitions.Values.Where(value => value.Type == SkillType.Primary)
                .OrderBy(value => value.Name, StringComparer.Ordinal)
                .Select(value => new PrimarySkillChoice(value.Name, value.Description))
                .ToImmutableArray();
        }
    }
}
