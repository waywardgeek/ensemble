using System.Collections.Immutable;
using System.Text;

namespace Ensemble;

public sealed partial class DirectoryListing
{
    public partial DirectoryListing(IEnumerable<DirectoryEntry> entries, int totalEntries)
    {
        ArgumentNullException.ThrowIfNull(entries);
        Entries = entries.Take(MaximumEntries + 1).ToImmutableArray();
        if (Entries.Length > MaximumEntries || totalEntries < Entries.Length)
            throw new ArgumentException("Invalid directory display count");
        int bytes = 0;
        foreach (var entry in Entries)
        {
            if (entry is null || string.IsNullOrEmpty(entry.Name) || entry.Name.Contains('/')
                || entry.Name.Contains('\0') || !Enum.IsDefined(entry.Kind))
                throw new ArgumentException("Invalid directory display entry");
            bytes = checked(bytes + Encoding.UTF8.GetByteCount(entry.Name));
            if (bytes > MaximumNameBytes) throw new ArgumentException("Directory display exceeds byte bound");
        }
        TotalEntries = totalEntries;
    }
}
