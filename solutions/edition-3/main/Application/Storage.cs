namespace Ensemble;

public sealed partial class Ensemble
{
    public partial string DataDirectory()
    {
        return dataDirectory ?? throw new InvalidOperationException(
            "Construct Ensemble with a data directory before using persistent actors.");
    }

    internal partial string? AgentDirectory(string? name)
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
        return Path.Combine(dataDirectory, "agents", name);
    }
}
