namespace Ensemble;

public sealed partial class Ensemble
{
    public partial string DataDirectory()
    {
        return dataDirectory ?? throw new InvalidOperationException(
            "Construct Ensemble with a data directory before using persistent actors.");
    }

    internal partial string? AgentDirectory(string? name, string sandboxRoot)
    {
        if (dataDirectory is null || name is null)
        {
            return null;
        }
        // Names are one local address component, never a relative filesystem path.
        if (string.IsNullOrWhiteSpace(name) || name is "." or ".."
            || name.Contains('/') || name.Contains('\\') || name.IndexOfAny(Path.GetInvalidFileNameChars()) >= 0)
        {
            throw new ArgumentException(
                "Persistent Agent names must be nonblank local names, not paths.", nameof(name));
        }
        string home = Path.Combine(dataDirectory, "agents", name);
        if (sandboxRoot.Length == 0) return home;
        string workspace = Sandbox.CanonicalPath(Path.GetFullPath(sandboxRoot));
        string physicalHome = Sandbox.CanonicalPath(Path.GetFullPath(home));
        string relative = Path.GetRelativePath(workspace, physicalHome);
        if (relative != ".." && !relative.StartsWith("../", StringComparison.Ordinal)
            && !Path.IsPathRooted(relative)) return physicalHome;
        // The host still assigns named homes; callers cannot supply an arbitrary
        // override. Keep this Agent's writable state inside its workspace.
        return Path.Combine(workspace, ".ensemble", "agents", name);
    }
}
