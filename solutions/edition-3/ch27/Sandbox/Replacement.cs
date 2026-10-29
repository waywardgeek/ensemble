namespace Ensemble;

public sealed partial class Sandbox
{
    public partial void ReplaceFile(string path, ReadOnlySpan<byte> content)
    {
        string fullPath = Resolve(path);
        string directory = System.IO.Path.GetDirectoryName(fullPath)!;
        CreateDirectory(directory);
        string temporary = System.IO.Path.Combine(directory, "." + System.IO.Path.GetFileName(path) + "." + Guid.NewGuid().ToString("N") + ".tmp");
        try
        {
            using (var stream = OpenFile(temporary, FileMode.CreateNew, FileAccess.Write))
            {
                stream.Write(content);
                stream.Flush(flushToDisk: true);
            }
            // Same-directory replacement avoids exposing a partly written JSON file.
            MoveFile(temporary, fullPath);
        }
        catch
        {
            // A failed write or rename may leave an owned temporary. Cleanup
            // uses the same boundary; a concurrent path swap must neither send
            // deletion outside the root nor mask the original operation error.
            try
            {
                DeleteFile(temporary);
            }
            catch (Exception cleanup) when (cleanup is IOException or UnauthorizedAccessException)
            {
                // Preserve the first failure. An unreachable temporary can be
                // left behind; it is never a valid recovery candidate.
            }
            throw;
        }
    }
}
