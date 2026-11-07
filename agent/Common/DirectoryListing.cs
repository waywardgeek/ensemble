using System.Collections.Immutable;

namespace Ensemble;

// Operation evidence for presentation, never a filesystem lookup or tool grant.
public enum DirectoryEntryKind { Unknown = 1, File, Directory }
public sealed record DirectoryEntry(string Name, DirectoryEntryKind Kind);
public sealed partial class DirectoryListing
{
    public const int MaximumEntries = 256;
    public const int MaximumNameBytes = 32768;
    public readonly ImmutableArray<DirectoryEntry> Entries;
    public readonly int TotalEntries;
    public bool Complete => Entries.Length == TotalEntries;
    public partial DirectoryListing(IEnumerable<DirectoryEntry> entries, int totalEntries);
}
