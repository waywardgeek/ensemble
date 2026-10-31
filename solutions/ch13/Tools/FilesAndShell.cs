using System.Diagnostics;
using System.Text;
using System.Text.Json;
using System.Text.RegularExpressions;

namespace Ensemble;

public sealed partial class Tool
{
    private static string ReadFile(JsonElement args)
    {
        var start = Integer(args, "start_line", 1);
        var end = Integer(args, "end_line", int.MaxValue);
        var maximum = Integer(args, "max_bytes", int.MaxValue);
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
            maximum = int.MaxValue;
        }
        if (maximum < 0)
        {
            throw new ArgumentException("max_bytes must be nonnegative.");
        }
        if (start < 1 || end < start)
        {
            throw new ArgumentException("Line numbers must be positive, with end_line >= start_line.");
        }
        var text = File.ReadAllText(String(args, "path"));
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
            return result.ToString();
        }
        // Back up to a UTF-8 boundary: a byte cap must not manufacture a broken
        // character. The report is separate from the capped file payload.
        var boundary = maximum;
        while (boundary > 0 && (bytes[boundary] & 0xc0) == 0x80)
        {
            boundary--;
        }
        return Encoding.UTF8.GetString(bytes, 0, boundary) + "\n[truncated]\n";
    }

    private static string ListDirectory(JsonElement args)
    {
        var directory = String(args, "path", ".");
        var entries = Directory.EnumerateFileSystemEntries(directory)
            .OrderBy(Path.GetFileName, StringComparer.Ordinal);
        var result = new StringBuilder();
        foreach (var entry in entries)
        {
            result.Append(Path.GetFileName(entry));
            if (Directory.Exists(entry))
            {
                result.Append('/');
            }
            result.Append('\n');
        }
        return result.ToString();
    }

    private static IEnumerable<string> SearchPaths(string path)
    {
        if (!Directory.Exists(path))
        {
            yield return path;
            yield break;
        }
        foreach (var entry in Directory.EnumerateFileSystemEntries(path).Order(StringComparer.Ordinal))
        {
            if (Directory.Exists(entry))
            {
                // Directory links are not traversed: recursive searches must not
                // loop forever through a link to an ancestor.
                if ((File.GetAttributes(entry) & FileAttributes.ReparsePoint) == 0)
                {
                    foreach (var file in SearchPaths(entry))
                    {
                        yield return file;
                    }
                }
            }
            else
            {
                yield return entry;
            }
        }
    }

    private static string SearchFiles(JsonElement args)
    {
        var radius = Integer(args, "context_lines", 0);
        if (radius < 0)
        {
            throw new ArgumentException("context_lines must be nonnegative.");
        }
        var pattern = new Regex(String(args, "pattern"));
        var result = new StringBuilder();
        foreach (var file in SearchPaths(String(args, "path")))
        {
            var lines = File.ReadAllLines(file);
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
                    var separator = matches[i] ? ':' : '-';
                    result.Append(file).Append(separator).Append(i + 1)
                        .Append(separator).Append(lines[i]).Append('\n');
                }
            }
        }
        return result.ToString();
    }

    private static string WriteFile(JsonElement args)
    {
        var path = String(args, "path");
        var append = Boolean(args, "append");
        var overwrite = Boolean(args, "overwrite");
        if (!append && !overwrite && File.Exists(path))
        {
            throw new IOException($"File exists: {path} ({new FileInfo(path).Length} bytes); use overwrite:true or append:true.");
        }
        var directory = Path.GetDirectoryName(Path.GetFullPath(path))!;
        Directory.CreateDirectory(directory);
        // CreateNew also refuses a file created between our check and the open.
        var mode = append ? FileMode.Append : overwrite ? FileMode.Create : FileMode.CreateNew;
        using var stream = new FileStream(path, mode, FileAccess.Write);
        using var writer = new StreamWriter(stream, new UTF8Encoding(false));
        writer.Write(String(args, "content"));
        return "File written.";
    }

    private static string EditFile(JsonElement args)
    {
        var path = String(args, "path");
        var oldText = String(args, "old_text");
        var replacement = String(args, "new_text");
        if (oldText.Length == 0)
        {
            throw new ArgumentException("old_text is empty for " + path + "; read the file and choose a nonempty unique anchor.");
        }
        var text = File.ReadAllText(path);
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
        File.WriteAllText(path, edited, new UTF8Encoding(false));
        return "File edited.";
    }

}
