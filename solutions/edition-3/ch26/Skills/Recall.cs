using System.Collections.Immutable;
using Microsoft.Win32.SafeHandles;

namespace Ensemble;

public sealed partial class Skills
{
    internal partial ImmutableArray<(string File, string Text, string Hash)> RecallFiles(CancellationToken cancellation)
    {
        lock (gate)
        {
            ObjectDisposedException.ThrowIf(destroyed, this);
            cancellation.ThrowIfCancellationRequested();
            // An explicitly selected skill tree is trusted input even outside
            // the workspace. This capability never becomes a Sandbox exception.
            recallRoot ??= Sandbox.OpenReadRoot(sourceDirectory);
            var files = new List<(string File, string Text, string Hash)>();
            ReadRecallDirectory(recallRoot, "skills", files, Parent().Configuration(), cancellation);
            return files.OrderBy(file => file.File, StringComparer.Ordinal).ToImmutableArray();
        }
    }

    private partial void ReadRecallDirectory(SafeFileHandle directory, string relative,
        List<(string File, string Text, string Hash)> files, Configuration policy, CancellationToken cancellation)
    {
        foreach (string name in Sandbox.ReadRootNames(directory))
        {
            cancellation.ThrowIfCancellationRequested();
            if (name.StartsWith('.') || (name == "web-search" && !policy.AllowsSkill(name)))
            {
                continue;
            }
            string path = relative + "/" + name;
            if (name == "SKILL.md")
            {
                // Open without following links or blocking on a FIFO. Only
                // discovered unreadable entries are skipped; malformed readable
                // content refuses the complete refresh, never publishes a subset.
                FileStream stream;
                try
                {
                    stream = Sandbox.OpenReadFile(directory, name);
                }
                catch (IOException) { continue; }
                catch (UnauthorizedAccessException) { continue; }
                using (stream)
                {
                    files.Add(RecallText.ReadCorpusStream(stream, path, cancellation));
                }
                continue;
            }
            SafeFileHandle child;
            try
            {
                // Non-SKILL files and directory symlinks fail this directory-only
                // open. A renamed ancestor cannot redirect a later descent.
                child = Sandbox.OpenReadDirectory(directory, name);
            }
            catch (IOException) { continue; }
            catch (UnauthorizedAccessException) { continue; }
            using (child)
            {
                ReadRecallDirectory(child, path, files, policy, cancellation);
            }
        }
    }
}
