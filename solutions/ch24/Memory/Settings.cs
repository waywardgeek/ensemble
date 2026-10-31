namespace Ensemble;

public sealed partial class MemoryBandSettings
{
    public partial MemoryBandSettings(long budget, long highWatermark, long lowWatermark, bool disabled)
    {
        ArgumentOutOfRangeException.ThrowIfNegative(budget);
        ArgumentOutOfRangeException.ThrowIfNegative(highWatermark);
        ArgumentOutOfRangeException.ThrowIfNegative(lowWatermark);
        Budget = budget;
        HighWatermark = highWatermark;
        LowWatermark = lowWatermark;
        Disabled = disabled;
    }

    internal partial MemoryBandSettings Effective(long defaultBudget)
    {
        long budget = Budget == 0 ? defaultBudget : Budget;
        long high = HighWatermark == 0 ? checked(2 * budget) : HighWatermark;
        long low = LowWatermark == 0 ? budget : LowWatermark;
        if (low > high)
        {
            throw new ArgumentException("Memory low watermark exceeds high watermark");
        }
        return new MemoryBandSettings(budget, high, low, Disabled);
    }

    internal partial bool Same(MemoryBandSettings other)
    {
        return Budget == other.Budget && HighWatermark == other.HighWatermark
            && LowWatermark == other.LowWatermark && Disabled == other.Disabled;
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
