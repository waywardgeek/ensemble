namespace Ensemble;

public sealed partial class MemoryBandSettings
{
    public partial MemoryBandSettings(long budget, bool disabled)
    {
        ArgumentOutOfRangeException.ThrowIfNegative(budget);
        ArgumentOutOfRangeException.ThrowIfGreaterThan(budget, long.MaxValue / 2);
        Budget = budget;
        Disabled = disabled;
    }

    internal partial MemoryBandSettings Effective(long defaultBudget)
    {
        return new MemoryBandSettings(Budget == 0 ? defaultBudget : Budget, Disabled);
    }

    internal partial bool Same(MemoryBandSettings other)
    {
        return Budget == other.Budget && Disabled == other.Disabled;
    }
}

public sealed partial class MemorySettings
{
    public partial MemorySettings(MemoryBandSettings? conversation, MemoryBandSettings? soul,
        MemoryBandSettings? memory, MemoryBandSettings? archive64, MemoryBandSettings? archive8,
        MemoryBandSettings? session)
    {
        Conversation = (conversation ?? new()).Effective(3000);
        Soul = (soul ?? new()).Effective(1500);
        Memory = (memory ?? new()).Effective(1500);
        Archive64 = (archive64 ?? new()).Effective(1500);
        Archive8 = (archive8 ?? new()).Effective(1500);
        Session = (session ?? new()).Effective(1500);
    }

    internal partial MemoryBandSettings For(MemoryBand? band)
    {
        return band switch
        {
            null => Conversation,
            MemoryBand.Soul => Soul,
            MemoryBand.Memory => Memory,
            MemoryBand.Archive64 => Archive64,
            MemoryBand.Archive8 => Archive8,
            MemoryBand.Session => Session,
            _ => throw new ArgumentOutOfRangeException(nameof(band))
        };
    }
}
