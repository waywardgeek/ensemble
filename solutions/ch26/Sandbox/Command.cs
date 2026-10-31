using System.Diagnostics;
using System.Text;
using System.Text.RegularExpressions;

namespace Ensemble;

public sealed partial class Sandbox
{
    internal partial ProcessStartInfo PrepareCommand(string? directory)
    {
        string working = Resolve(string.IsNullOrEmpty(directory) ? "." : directory);
        var start = new ProcessStartInfo("/bin/sh")
        {
            WorkingDirectory = working,
            UseShellExecute = false
        };
        if (root.Length == 0)
        {
            // Preserve full-access behavior. Confinement below deliberately uses
            // an allowlist instead: a suffix blacklist cannot identify arbitrary
            // credentials or executable shell startup hooks.
            foreach (string key in start.Environment.Keys.ToArray())
            {
                if (key.EndsWith("_API_KEY", StringComparison.OrdinalIgnoreCase)
                    || key is "ANTHROPIC_AUTH_TOKEN" or "GOOGLE_API_KEY" or "GEMINI_API_KEY" or "OPENAI_API_KEY" or "ANTHROPIC_API_KEY")
                {
                    start.Environment.Remove(key);
                }
            }
            start.Environment["TERM"] = "xterm-256color";
            return start;
        }
        if (!OperatingSystem.IsMacOS())
        {
            throw new PlatformNotSupportedException("Confined commands require the macOS process sandbox");
        }
        start.Environment.Clear();
        foreach (string name in new[] { "LANG", "LC_ALL", "LC_CTYPE", "TZ" })
        {
            string? value = Environment.GetEnvironmentVariable(name);
            if (value is not null)
            {
                start.Environment[name] = value;
            }
        }
        start.Environment["TERM"] = "xterm-256color";
        start.Environment["PATH"] = "/usr/bin:/bin:/usr/sbin:/sbin";
        start.Environment["HOME"] = root;
        start.Environment["TMPDIR"] = root;
        start.Environment["GIT_CONFIG_GLOBAL"] = Path.Combine(root, ".ensemble-gitconfig");
        start.Environment["GIT_CONFIG_SYSTEM"] = "/dev/null";
        start.Environment["GIT_PAGER"] = "cat";
        start.Environment["PAGER"] = "cat";
        // Concurrent jobs may launch Git while another launch refreshes this
        // file. Publish a complete identity by rename, never truncate it in place.
        string temporary = ".ensemble-gitconfig-" + Guid.NewGuid().ToString("N");
        try
        {
            using (var file = OpenFile(temporary, FileMode.CreateNew, FileAccess.Write))
            {
                file.Write(Encoding.UTF8.GetBytes("[user]\n name = Ensemble\n email = ensemble@localhost\n"));
            }
            MoveFile(temporary, ".ensemble-gitconfig");
        }
        catch
        {
            try
            {
                DeleteFile(temporary);
            }
            catch (Exception cleanup) when (cleanup is IOException or UnauthorizedAccessException)
            {
                // Preserve the launch failure if concurrent filesystem changes
                // also prevent cleanup. Never retry through unconfined I/O.
            }
            throw;
        }

        // The trusted host's provider connection is outside this profile. The
        // command and every descendant inherit these filesystem/network rules.
        // Toolchain/runtime reads are intentionally narrower than all of /Library
        // or /usr/local, which can hold unrelated configuration and credentials.
        var profile = new StringBuilder("""
            (version 1)
            (allow default)
            (deny file-read*)
            (deny file-write*)
            (deny network*)
            (allow file-read* (literal "/") (subpath "/bin") (subpath "/sbin")
              (subpath "/usr/bin") (subpath "/usr/sbin") (subpath "/usr/lib")
              (subpath "/usr/libexec") (subpath "/usr/share") (subpath "/System")
              (subpath "/Library/Developer") (subpath "/private/etc/ssl")
              (literal "/private/var/select/sh") (literal "/private/var/select/developer_dir")
              (literal "/private/var/db/xcode_select_link")
              (literal "/dev/null") (literal "/dev/zero") (literal "/dev/random")
              (literal "/dev/urandom") (literal "/dev/tty") (subpath "/dev/fd")
              (literal "/dev/stdin") (literal "/dev/stdout") (literal "/dev/stderr"))
            (allow file-write* (literal "/dev/null") (literal "/dev/tty")
              (literal "/dev/stdout") (literal "/dev/stderr"))
            (allow file-read-metadata (literal "/var") (literal "/tmp") (literal "/etc")
              (literal "/usr") (literal "/Library") (literal "/private/var/db")
              (literal "/private") (literal "/private/var") (literal "/private/var/select"))
            """);
        profile.Append("\n(allow file-read* file-write* (subpath ").Append(ProfileLiteral(root)).Append(") )\n");
        // Git resolves cwd component by component. Ancestor metadata permits
        // traversal without granting directory listings or sibling file reads.
        for (string? ancestor = Path.GetDirectoryName(root); ancestor is not null && ancestor != "/";
            ancestor = Path.GetDirectoryName(ancestor))
        {
            profile.Append("(allow file-read-metadata (literal ").Append(ProfileLiteral(ancestor)).Append("))\n");
        }
        // Xcode shims use this narrow cache even with a rooted TMPDIR. Only
        // immediate xcrun_db* files qualify, never similarly named nested files.
        string cache = "^" + Regex.Escape(Path.Combine(ToolchainTemporaryDirectory(), "xcrun_db")) + "[^/]*$";
        profile.Append("(allow file-read* file-write* (regex #").Append(ProfileLiteral(cache)).Append("))\n");
        start.FileName = "/usr/bin/sandbox-exec";
        start.ArgumentList.Add("-p");
        start.ArgumentList.Add(profile.ToString());
        start.ArgumentList.Add("/bin/sh");
        return start;
    }

    private static partial string ProfileLiteral(string value)
    {
        // SBPL is not JSON: JSON's optional Unicode escapes need not have the
        // same meaning here. Quote just its string syntax; refuse control bytes.
        if (value.Any(char.IsControl))
        {
            throw new IOException("Sandbox process paths cannot contain control characters");
        }
        return "\"" + value.Replace("\\", "\\\\", StringComparison.Ordinal)
            .Replace("\"", "\\\"", StringComparison.Ordinal) + "\"";
    }

    private static partial string ToolchainTemporaryDirectory()
    {
        // Darwin's _CS_DARWIN_USER_TEMP_DIR is 65537 (unistd.h). Query libc
        // directly rather than launching an unconfined helper or trusting TMPDIR.
        const int setting = 65537;
        nuint length = NativeConfstr(setting, null, 0);
        if (length < 2 || length > 65536)
        {
            throw new IOException("Sandbox cannot locate the toolchain cache directory");
        }
        var bytes = new byte[(int)length];
        nuint actual = NativeConfstr(setting, bytes, length);
        if (actual != length || bytes[^1] != 0)
        {
            throw new IOException("Sandbox cannot read the toolchain cache directory");
        }
        return CanonicalPath(Encoding.UTF8.GetString(bytes, 0, bytes.Length - 1));
    }
}
