using System.Diagnostics;
using System.Text;
using System.Text.Json;
using System.Text.RegularExpressions;

namespace Ensemble;

public sealed partial class Tool
{
    private ToolResultPart ReadFile(ToolCallPart call)
    {
        var args = call.Args;
        var start = Integer(args, "start_line", 1);
        var end = Integer(args, "end_line", int.MaxValue);
        var maximum = Integer(args, "max_bytes", 32768);
        if (start == 0)
        {
            start = 1;
        }
        if (end == 0)
        {
            end = int.MaxValue;
        }
        if (maximum == 0)
        {
            maximum = 32768;
        }
        if (maximum < 0)
        {
            throw new ArgumentException("max_bytes must be nonnegative.");
        }
        if (start < 1 || end < start)
        {
            throw new ArgumentException("Line numbers must be positive, with end_line >= start_line.");
        }
        using var input = Parent().Sandbox().OpenFile(String(args, "path"), FileMode.Open, FileAccess.Read);
        using var reader = new StreamReader(input);
        var text = reader.ReadToEnd();
        // Include each line's original terminator. Splitting on '\n' and joining
        // would silently normalize CRLF or add a newline to an unterminated file.
        var result = new StringBuilder();
        int offset = 0;
        int line = 1;
        while (offset < text.Length && line <= end)
        {
            var newline = text.IndexOf('\n', offset);
            var limit = newline < 0 ? text.Length : newline + 1;
            if (line >= start)
            {
                result.Append(text, offset, limit - offset);
            }
            offset = limit;
            line++;
        }
        var bytes = Encoding.UTF8.GetBytes(result.ToString());
        if (bytes.Length <= maximum)
        {
            return new ToolResultPart(call.CallID, [new TextPart(NumberReadLines(result.ToString(), start))],
                fileEvidence: new FileEvidence(FileOperation.Read, String(args, "path"), start, 1,
                    FileExcerpt.Capture(result.ToString())));
        }
        // Back up to a UTF-8 boundary: a byte cap must not manufacture a broken
        // character. The report is separate from the capped file payload.
        var boundary = maximum;
        while (boundary > 0 && (bytes[boundary] & 0xc0) == 0x80)
        {
            boundary--;
        }
        var captured = Encoding.UTF8.GetString(bytes, 0, boundary);
        var excerpt = FileExcerpt.Capture(captured);
        return new ToolResultPart(call.CallID, [new TextPart(NumberReadLines(captured, start) + "\n[truncated]\n")],
            fileEvidence: new FileEvidence(FileOperation.Read, String(args, "path"), start, 1,
                new FileExcerpt(excerpt.Text, bytes.LongLength)));
    }

    // Number the model-facing presentation, not the captured source evidence.
    // Preserve indentation, terminators and an unterminated final line exactly.
    private static string NumberReadLines(string text, int startLine)
    {
        var numbered = new StringBuilder();
        int offset = 0;
        long line = startLine;
        while (offset < text.Length)
        {
            var newline = text.IndexOf('\n', offset);
            var limit = newline < 0 ? text.Length : newline + 1;
            numbered.Append(line.ToString(System.Globalization.CultureInfo.InvariantCulture)).Append(": ");
            numbered.Append(text, offset, limit - offset);
            offset = limit;
            line++;
        }
        return numbered.ToString();
    }

    private ToolResultPart ListDirectory(ToolCallPart call)
    {
        var directory = String(call.Args, "path", ".");
        var files = Parent().Sandbox();
        var result = new StringBuilder();
        var entries = new List<DirectoryEntry>();
        int total = 0, nameBytes = 0;
        bool displayFull = false;
        foreach (string name in files.ReadDirectory(directory))
        {
            total++;
            var entryKind = DirectoryEntryKind.Unknown;
            result.Append(name);
            try
            {
                if (files.IsDirectory(Path.Combine(directory, name)))
                {
                    result.Append('/');
                    entryKind = DirectoryEntryKind.Directory;
                }
            }
            catch (IOException)
            {
                // A vanished entry or an inaccessible/outward symlink remains
                // a visible name. Decoration must not bypass the boundary.
            }
            result.Append('\n');
            int size = Encoding.UTF8.GetByteCount(name);
            if (!displayFull && entries.Count < DirectoryListing.MaximumEntries
                && size <= DirectoryListing.MaximumNameBytes - nameBytes)
            {
                entries.Add(new DirectoryEntry(name, entryKind));
                nameBytes += size;
            }
            else displayFull = true;
        }
        return new ToolResultPart(call.CallID, [new TextPart(result.ToString())],
            directoryListing: new DirectoryListing(entries, total));
    }

    private IEnumerable<string> SearchPaths(string path, HashSet<string> visited)
    {
        var files = Parent().Sandbox();
        if (!files.IsDirectory(path))
        {
            yield return path;
            yield break;
        }
        // Canonical identity prevents inward links from recursively revisiting
        // a directory. This is only a cycle check: each enumeration and open
        // still enforces the rooted boundary independently.
        if (!visited.Add(files.Resolve(path)))
        {
            yield break;
        }
        foreach (string name in files.ReadDirectory(path))
        {
            string entry = Path.Combine(path, name);
            bool directory;
            try
            {
                directory = files.IsDirectory(entry);
            }
            catch (Exception error) when (error is IOException or UnauthorizedAccessException)
            {
                // A recursive scan may encounter an outward link or a vanished
                // entry. Skip it without opening outside or losing other matches.
                // The explicitly requested root was checked above and still
                // refuses; this catch applies only to discovered children.
                continue;
            }
            if (directory)
            {
                // Preserve full-access search's existing no-directory-links
                // behavior. Host metadata is consulted only in full access;
                // confined traversal never uses path-based host operations.
                if (files.Root().Length == 0 && (File.GetAttributes(entry) & FileAttributes.ReparsePoint) != 0)
                {
                    continue;
                }
                foreach (string file in SearchPaths(entry, visited))
                {
                    yield return file;
                }
            }
            else
            {
                yield return entry;
            }
        }
    }

    private ToolResultPart SearchFiles(ToolCallPart call)
    {
        var args = call.Args;
        var captured = new List<SearchLine>();
        long totalLines = 0;
        int displayBytes = 0;
        bool displayFull = false;
        var radius = Integer(args, "context_lines", 0);
        if (radius < 0)
        {
            throw new ArgumentException("context_lines must be nonnegative.");
        }
        var pattern = new Regex(String(args, "pattern"));
        var result = new StringBuilder();
        var visited = new HashSet<string>(StringComparer.Ordinal);
        string path = String(args, "path");
        bool recursive = Parent().Sandbox().IsDirectory(path);
        foreach (var file in SearchPaths(path, visited))
        {
            var lineList = new List<string>();
            try
            {
                using var input = Parent().Sandbox().OpenFile(file, FileMode.Open, FileAccess.Read);
                using var reader = new StreamReader(input);
                while (reader.ReadLine() is { } line)
                {
                    lineList.Add(line);
                }
            }
            catch (Exception error) when (recursive && (error is IOException or UnauthorizedAccessException))
            {
                // Recursive discovery is best effort for inaccessible and
                // non-file entries; a direct file request remains a refusal.
                continue;
            }
            var lines = lineList.ToArray();
            var matches = lines.Select(line => pattern.IsMatch(line)).ToArray();
            var windows = new List<(int Start, int End)>();
            for (int i = 0; i < lines.Length; i++)
            {
                if (!matches[i])
                {
                    continue;
                }
                // Use long arithmetic so a legal large radius cannot overflow.
                var start = (int)Math.Max(0L, (long)i - radius);
                var end = (int)Math.Min(lines.Length - 1L, (long)i + radius);
                if (windows.Count > 0 && start <= windows[^1].End + 1)
                {
                    windows[^1] = (windows[^1].Start, end);
                }
                else
                {
                    windows.Add((start, end));
                }
            }
            for (int w = 0; w < windows.Count; w++)
            {
                if (w > 0 && radius > 0)
                {
                    result.Append("--\n");
                }
                for (int i = windows[w].Start; i <= windows[w].End; i++)
                {
                    totalLines++;
                    if (!displayFull)
                    {
                        long bytes = Encoding.UTF8.GetByteCount(file) + (long)Encoding.UTF8.GetByteCount(lines[i]);
                        if (captured.Count == SearchEvidence.MaximumLines || bytes + displayBytes > SearchEvidence.MaximumBytes)
                            displayFull = true;
                        else
                        {
                            var spans = pattern.Matches(lines[i]).Cast<Match>()
                                .Take(SearchEvidence.MaximumSpansPerLine + 1)
                                .Select(match => new SearchSpan(match.Index, match.Length)).ToArray();
                            if (spans.Length > SearchEvidence.MaximumSpansPerLine) displayFull = true;
                            else
                            {
                                captured.Add(new SearchLine(file, i + 1, lines[i], spans));
                                displayBytes += (int)bytes;
                            }
                        }
                    }
                    var separator = matches[i] ? ':' : '-';
                    result.Append(file).Append(separator).Append(i + 1)
                        .Append(separator).Append(lines[i]).Append('\n');
                }
            }
        }
        return new ToolResultPart(call.CallID, [new TextPart(result.ToString())],
            searchEvidence: new SearchEvidence(captured, totalLines));
    }

    private ToolResultPart WriteFile(ToolCallPart call)
    {
        var args = call.Args;
        var path = String(args, "path");
        var append = Boolean(args, "append");
        var overwrite = Boolean(args, "overwrite");
        var files = Parent().Sandbox();
        var directory = Path.GetDirectoryName(path);
        files.CreateDirectory(string.IsNullOrEmpty(directory) ? "." : directory);
        // CreateNew enforces no-overwrite at the actual open. A separate
        // existence/size probe would race, and must not reopen a host path.
        var content = String(args, "content");
        var excerpt = FileExcerpt.Capture(content);
        var evidence = new FileEvidence(append ? FileOperation.Append : FileOperation.Write,
            path, 1, 1, excerpt);
        var mode = append ? FileMode.OpenOrCreate : overwrite ? FileMode.Create : FileMode.CreateNew;
        using var stream = files.OpenFile(path, mode, append ? FileAccess.ReadWrite : FileAccess.Write);
        long startLine = 1, startColumn = 1;
        if (append)
        {
            // Count and write through the same opened file, not a second path
            // lookup. This is not a lock against unrelated concurrent writers.
            using (var reader = new StreamReader(stream, leaveOpen: true))
            {
                var buffer = new char[4096];
                int count;
                while ((count = reader.Read(buffer, 0, buffer.Length)) != 0)
                    foreach (char character in buffer.AsSpan(0, count))
                        if (character == '\n') { startLine++; startColumn = 1; }
                        else startColumn++;
            }
            stream.Seek(0, SeekOrigin.End);
        }
        if (append) evidence = new FileEvidence(FileOperation.Append, path, startLine, startColumn, excerpt);
        // Dispose before returning success: flushing can fail too.
        using (var writer = new StreamWriter(stream, new UTF8Encoding(false), leaveOpen: true))
            writer.Write(content);
        if (content.Length == 0) return new ToolResultPart(call.CallID,
            [new TextPart("File written; no content written (empty input).")], fileEvidence: evidence);
        long endLine = startLine;
        for (int index = 0; index < content.Length - 1; index++)
            if (content[index] == '\n') endLine++;
        return new ToolResultPart(call.CallID,
            [new TextPart($"File written. Lines {startLine}–{endLine}; starting column {startColumn} (UTF-16).")],
            fileEvidence: evidence);
    }

    private ToolResultPart EditFile(ToolCallPart call)
    {
        var args = call.Args;
        var path = String(args, "path");
        var oldText = String(args, "old_text");
        var replacement = String(args, "new_text");
        if (oldText.Length == 0)
        {
            throw new ArgumentException("old_text is empty for " + path + "; read the file and choose a nonempty unique anchor.");
        }
        // Read and replace through one handle. A swapped directory entry must
        // not redirect the write after the anchor was checked in another file.
        using var stream = Parent().Sandbox().OpenFile(path, FileMode.Open, FileAccess.ReadWrite);
        string text;
        using (var reader = new StreamReader(stream, leaveOpen: true))
        {
            text = reader.ReadToEnd();
        }
        var first = text.IndexOf(oldText, StringComparison.Ordinal);
        if (first < 0)
        {
            throw new ArgumentException("old_text has no match in " + path + "; read the file and choose an exact anchor.");
        }
        // Advance by one, not by anchor length: "aa" occurs twice in "aaa".
        if (text.IndexOf(oldText, first + 1, StringComparison.Ordinal) >= 0)
        {
            throw new ArgumentException("old_text has multiple matches in " + path + "; read the file and narrow the anchor.");
        }
        var edited = text[..first] + replacement + text[(first + oldText.Length)..];
        long startLine = 1, startColumn = 1;
        foreach (char character in text.AsSpan(0, first))
            if (character == '\n') { startLine++; startColumn = 1; }
            else startColumn++;
        var evidence = new FileEvidence(FileOperation.Edit, path, startLine, startColumn,
            FileExcerpt.Capture(replacement), FileExcerpt.Capture(oldText));
        stream.Position = 0;
        stream.SetLength(0);
        using (var writer = new StreamWriter(stream, new UTF8Encoding(false), leaveOpen: true))
            writer.Write(edited);
        return new ToolResultPart(call.CallID, [new TextPart("File edited.")], fileEvidence: evidence);
    }

}
