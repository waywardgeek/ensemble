using System.Collections.Immutable;

namespace EnsembleGui;

// Immutable executable snapshot. Files are reviewed together; their source is
// data until a human explicitly approves a trial in the current browser page.
public sealed record RendererPackage(string Renderer, string Entry,
    ImmutableSortedDictionary<string, string> Files);
public sealed record RendererSelection(string? Accepted = null, string? Previous = null);

// The GUI hosting operation owns this store, shared by its Agent views. It is
// independent of Agent settings and contains no persisted trial selection.
public sealed partial class RendererStore
{
    private readonly string directory;
    private readonly object gate = new();
    public RendererStore(string directory)
    {
        this.directory = Path.GetFullPath(directory);
    }

    public partial string Stage(RendererPackage package);
    public partial RendererPackage Read(string version);
    public partial RendererSelection Selection(string renderer);
    // Compare the accepted identity the human saw before publishing a new one.
    // Null means the shipped default. Persistence failure changes no selection.
    public partial RendererSelection Accept(string renderer, string version, string? expected);
    public partial RendererSelection Rollback(string renderer, string? expected);
}
