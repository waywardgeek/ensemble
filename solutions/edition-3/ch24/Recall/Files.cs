using System.Collections.Immutable;
using System.Security.Cryptography;
using System.Text;

namespace Ensemble;

internal static partial class RecallText
{
    internal static partial ImmutableArray<(string File, string Text, string Hash)> ReadCorpus(
        Sandbox sandbox, string home, RecallSource source, CancellationToken cancellation)
    {
        var files = ImmutableArray.CreateBuilder<(string File, string Text, string Hash)>();
        var visited = new HashSet<string>(StringComparer.Ordinal);
        string[] directories = source switch
        {
            RecallSource.Memory => ["memory", "bucket-0", "bucket-1"],
            RecallSource.Docs => ["docs"],
            RecallSource.Handoffs => ["handoffs"],
            RecallSource.Learnings => ["learnings"],
            _ => throw new ArgumentOutOfRangeException(nameof(source))
        };
        if (source == RecallSource.Memory)
        {
            foreach (string name in new[] { "SOUL.md", "MEMORY.md" })
            {
                ReadCorpusFile(sandbox, home, Path.Combine(home, name), files, cancellation);
            }
        }
        foreach (string directory in directories)
        {
            ReadCorpusDirectory(sandbox, home, Path.Combine(home, directory), source,
                visited, files, cancellation);
        }
        return files.OrderBy(file => file.File, StringComparer.Ordinal).ToImmutableArray();
    }

    private static partial string? ResolveCorpusPath(Sandbox sandbox, string home, string path)
    {
        // Unconfined Resolve is lexical, so do not follow discovered links there.
        // A confined Resolve checks the real target; hidden staging cannot become
        // eligible through a visible alias. Restrict targets to this Agent's home
        // even when its broader Sandbox also permits other workspace directories.
        if (sandbox.Root().Length == 0
            && (File.GetAttributes(path) & FileAttributes.ReparsePoint) != 0)
        {
            return null;
        }
        string resolved = sandbox.Resolve(path);
        string relative = Path.GetRelativePath(home, resolved);
        if (Path.IsPathRooted(relative)
            || relative.Split(Path.DirectorySeparatorChar).Any(part => part.StartsWith('.')))
        {
            return null;
        }
        return resolved;
    }

    private static partial void ReadCorpusDirectory(Sandbox sandbox, string home, string directory,
        RecallSource source, HashSet<string> visited,
        ImmutableArray<(string File, string Text, string Hash)>.Builder files,
        CancellationToken cancellation)
    {
        cancellation.ThrowIfCancellationRequested();
        string[] names;
        try
        {
            string? resolved = ResolveCorpusPath(sandbox, home, directory);
            if (resolved is null || !visited.Add(resolved)) return;
            names = sandbox.ReadDirectory(resolved);
        }
        catch (IOException) { return; }
        catch (UnauthorizedAccessException) { return; }

        foreach (string name in names)
        {
            cancellation.ThrowIfCancellationRequested();
            // Hidden staging trees, dotfiles and temporary candidates never enter
            // an index, even when a nested name happens to end in .md.
            if (name.StartsWith('.')) continue;
            string path = Path.Combine(directory, name);
            bool isDirectory;
            try { isDirectory = sandbox.IsDirectory(path); }
            catch (IOException) { continue; }
            catch (UnauthorizedAccessException) { continue; }
            if (isDirectory)
            {
                ReadCorpusDirectory(sandbox, home, path, source, visited, files, cancellation);
            }
            else if (name.EndsWith(".md", StringComparison.OrdinalIgnoreCase)
                || (source == RecallSource.Memory
                    && name.EndsWith(".md.retired", StringComparison.OrdinalIgnoreCase)))
            {
                ReadCorpusFile(sandbox, home, path, files, cancellation);
            }
        }
    }

    private static partial void ReadCorpusFile(Sandbox sandbox, string home, string path,
        ImmutableArray<(string File, string Text, string Hash)>.Builder files,
        CancellationToken cancellation)
    {
        cancellation.ThrowIfCancellationRequested();
        try
        {
            string? resolved = ResolveCorpusPath(sandbox, home, path);
            if (resolved is null) return;
            using var input = sandbox.OpenRegularRead(resolved);
            string relative = Path.GetRelativePath(home, path).Replace('\\', '/');
            files.Add(ReadCorpusStream(input, relative, cancellation));
        }
        catch (InvalidDataException) { throw; }
        // Discovered inaccessible/outward/non-file entries are not corpus data.
        // Decode and limit failures refuse refresh rather than publish a partial index.
        catch (IOException) { return; }
        catch (UnauthorizedAccessException) { return; }
    }

    internal static partial (string File, string Text, string Hash) ReadCorpusStream(
        Stream stream, string relative, CancellationToken cancellation)
    {
        cancellation.ThrowIfCancellationRequested();
        if (stream.Length > Memory.MaximumFileBytes)
        {
            throw new InvalidDataException("Recall source exceeds the file safety limit");
        }
        using var output = new MemoryStream();
        byte[] buffer = new byte[16 * 1024];
        int count;
        while ((count = stream.Read(buffer, 0, buffer.Length)) != 0)
        {
            cancellation.ThrowIfCancellationRequested();
            if (output.Length + count > Memory.MaximumFileBytes)
            {
                throw new InvalidDataException("Recall source grew beyond the file safety limit");
            }
            output.Write(buffer, 0, count);
        }
        byte[] bytes = output.ToArray();
        string text = new UTF8Encoding(false, true).GetString(bytes);
        string hash = Convert.ToHexStringLower(SHA256.HashData(bytes));
        return (relative, text, hash);
    }

    internal static partial string CorpusFingerprint(
        ImmutableArray<(string File, string Text, string Hash)> files)
    {
        using var digest = IncrementalHash.CreateHash(HashAlgorithmName.SHA256);
        foreach (var file in files)
        {
            // NUL cannot occur in a filename. The fixed-length content hash and
            // delimiter keep distinct path/content pairs unambiguous.
            digest.AppendData(Encoding.UTF8.GetBytes(file.File + "\0" + file.Hash));
        }
        return Convert.ToHexStringLower(digest.GetHashAndReset());
    }
}
