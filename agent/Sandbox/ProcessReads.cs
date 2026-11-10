using System.Collections.Immutable;

namespace Ensemble;

public sealed partial class Sandbox
{
    private static partial bool ContainsPath(string directory, string path)
    {
        string relative = Path.GetRelativePath(directory, path);
        return relative != ".."
            && !relative.StartsWith("../", StringComparison.Ordinal)
            && !Path.IsPathRooted(relative);
    }

    private static partial ImmutableArray<ProcessReadPath> ResolveProcessReads(
        ImmutableArray<string> paths, string writableRoot)
    {
        if (paths.IsDefault || paths.Length > 64)
        {
            throw new ArgumentException("Invalid process read path list.");
        }
        if (paths.Length != 0 && writableRoot.Length == 0)
        {
            throw new ArgumentException("Read-only process grants require a writable sandbox root.");
        }
        var result = ImmutableArray.CreateBuilder<ProcessReadPath>();
        var seen = new HashSet<string>(StringComparer.Ordinal);
        foreach (string requested in paths)
        {
            if (string.IsNullOrWhiteSpace(requested) || requested.Length > 4096
                || !Path.IsPathFullyQualified(requested) || requested.Any(char.IsControl))
            {
                throw new ArgumentException("Process read grants require absolute paths.");
            }
            string physical = CanonicalPath(requested);
            bool directory = Directory.Exists(physical);
            if (!directory)
            {
                // Reuse the no-follow regular-file check. This refuses FIFOs and
                // other special handles without waiting for a peer to open them.
                string parentPath = Path.GetDirectoryName(physical)
                    ?? throw new IOException("Invalid process read file.");
                using var parentHandle = OpenReadRoot(parentPath);
                using var file = OpenReadFile(parentHandle, Path.GetFileName(physical));
            }
            if (ContainsPath(writableRoot, physical)
                || (directory && ContainsPath(physical, writableRoot)))
            {
                throw new ArgumentException("Read-only process paths must not overlap the writable root.");
            }
            if (!seen.Add(physical))
            {
                throw new ArgumentException("Duplicate process read path.");
            }
            result.Add(new ProcessReadPath(physical, directory));
        }
        return result.ToImmutable();
    }
}
