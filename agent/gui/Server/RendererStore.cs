using System.Collections.Immutable;
using System.Security.Cryptography;
using System.Text;
using System.Text.Json;
using System.Text.Json.Serialization;

namespace EnsembleGui;

public sealed partial class RendererStore
{
    private const int Limit = 1024 * 1024;
    private static readonly JsonSerializerOptions Json = new()
    {
        UnmappedMemberHandling = JsonUnmappedMemberHandling.Disallow
    };

    private static void Name(string name)
    {
        if (string.IsNullOrEmpty(name) || name.Length > 128 ||
            name.Any(c => !char.IsAsciiLetterOrDigit(c) && c is not '-' and not '_'))
            throw new ArgumentException("Invalid renderer name.");
    }

    private static void Version(string version)
    {
        if (version.Length != 64 || version.Any(c => c is not (>= '0' and <= '9') and not (>= 'a' and <= 'f')))
            throw new ArgumentException("Invalid renderer version.");
    }

    private static byte[] Encode(RendererPackage package)
    {
        Name(package.Renderer);
        if (package.Files is null || package.Files.Count is < 1 or > 64 ||
            package.Entry is null || !package.Files.ContainsKey(package.Entry))
            throw new ArgumentException("Renderer package needs an entry and at most 64 files.");
        long sourceBytes = 0;
        foreach (var (path, source) in package.Files)
        {
            if (path.Length > 256 || path.Split('/').Any(segment => segment.Length == 0 ||
                segment is "." or ".." || segment.Any(c => !char.IsAsciiLetterOrDigit(c) && c is not '-' and not '_' and not '.')) ||
                !path.EndsWith(".js", StringComparison.Ordinal) || source is null)
                throw new ArgumentException("Renderer packages contain named JavaScript files only.");
            // Reject invalid UTF-16 instead of hashing replacement characters.
            sourceBytes += new UTF8Encoding(false, true).GetByteCount(source);
            if (sourceBytes > Limit) throw new ArgumentException("Renderer source exceeds 1 MiB.");
        }
        var canonical = package with { Files = package.Files.ToImmutableSortedDictionary(StringComparer.Ordinal) };
        byte[] bytes = JsonSerializer.SerializeToUtf8Bytes(canonical, Json);
        if (bytes.Length > Limit) throw new ArgumentException("Renderer package exceeds 1 MiB.");
        return bytes;
    }

    private static string Hash(byte[] bytes) => Convert.ToHexString(SHA256.HashData(bytes)).ToLowerInvariant();

    private static byte[] BoundedRead(string path)
    {
        using var input = new FileStream(path, FileMode.Open, FileAccess.Read, FileShare.Read);
        using var output = new MemoryStream();
        var buffer = new byte[8192];
        while (true)
        {
            int count = input.Read(buffer, 0, Math.Min(buffer.Length, Limit + 1 - (int)output.Length));
            if (count == 0) break;
            output.Write(buffer, 0, count);
            if (output.Length > Limit) throw new InvalidDataException("Renderer record exceeds 1 MiB.");
        }
        return output.ToArray();
    }

    private static T Decode<T>(byte[] bytes)
    {
        using var document = JsonDocument.Parse(bytes, new JsonDocumentOptions { MaxDepth = 16 });
        static void Unique(JsonElement value)
        {
            if (value.ValueKind == JsonValueKind.Object)
            {
                var names = new HashSet<string>(StringComparer.Ordinal);
                foreach (var field in value.EnumerateObject())
                {
                    if (!names.Add(field.Name)) throw new InvalidDataException("Duplicate renderer field.");
                    Unique(field.Value);
                }
            }
            else if (value.ValueKind == JsonValueKind.Array)
                foreach (var item in value.EnumerateArray()) Unique(item);
        }
        Unique(document.RootElement);
        return JsonSerializer.Deserialize<T>(bytes, Json) ?? throw new InvalidDataException("Missing renderer record.");
    }

    private static void Save(string path, byte[] bytes, bool replace)
    {
        string temporary = path + "." + Guid.NewGuid().ToString("N") + ".tmp";
        try
        {
            using (var output = new FileStream(temporary, FileMode.CreateNew, FileAccess.Write, FileShare.None))
            {
                output.Write(bytes);
                output.Flush(true);
            }
            File.Move(temporary, path, replace);
        }
        finally { if (File.Exists(temporary)) File.Delete(temporary); }
    }

    public partial string Stage(RendererPackage package)
    {
        byte[] bytes = Encode(package);
        string version = Hash(bytes);
        lock (gate)
        {
            string folder = Path.Combine(directory, "versions");
            Directory.CreateDirectory(folder);
            string path = Path.Combine(folder, version + ".json");
            if (File.Exists(path))
            {
                if (!BoundedRead(path).SequenceEqual(bytes)) throw new InvalidDataException("Renderer snapshot changed.");
            }
            else Save(path, bytes, false);
        }
        return version;
    }

    public partial RendererPackage Read(string version)
    {
        Version(version);
        byte[] bytes = BoundedRead(Path.Combine(directory, "versions", version + ".json"));
        if (Hash(bytes) != version) throw new InvalidDataException("Renderer snapshot hash mismatch.");
        var package = Decode<RendererPackage>(bytes);
        if (!Encode(package).SequenceEqual(bytes)) throw new InvalidDataException("Noncanonical renderer snapshot.");
        return package;
    }

    public partial RendererSelection Selection(string renderer)
    {
        Name(renderer);
        string path = Path.Combine(directory, renderer + ".json");
        RendererSelection selection;
        try { selection = Decode<RendererSelection>(BoundedRead(path)); }
        catch (FileNotFoundException) { return new(); }
        catch (DirectoryNotFoundException) { return new(); }
        if (selection.Accepted is { } accepted) Version(accepted);
        if (selection.Previous is { } previous) Version(previous);
        return selection;
    }

    private RendererSelection Change(string renderer, string? version, string? expected, bool rollback)
    {
        Name(renderer);
        if (expected is not null) Version(expected);
        lock (gate)
        {
            Directory.CreateDirectory(directory);
            // Cooperating hosts refuse concurrent promotion rather than lose an
            // acceptance. This is not protection against arbitrary disk writers.
            using var lease = new FileStream(Path.Combine(directory, "selection.lock"),
                FileMode.OpenOrCreate, FileAccess.ReadWrite, FileShare.None);
            var current = Selection(renderer);
            if (current.Accepted != expected) throw new InvalidOperationException("Renderer selection changed; refresh before accepting.");
            string? target = rollback ? current.Previous : version;
            if (target is not null && Read(target).Renderer != renderer)
                throw new ArgumentException("Renderer version belongs to another renderer.");
            if (target == current.Accepted) return current;
            var next = new RendererSelection(target, current.Accepted);
            Save(Path.Combine(directory, renderer + ".json"), JsonSerializer.SerializeToUtf8Bytes(next, Json), true);
            return next;
        }
    }

    public partial RendererSelection Accept(string renderer, string version, string? expected)
    {
        Version(version);
        return Change(renderer, version, expected, false);
    }

    public partial RendererSelection Rollback(string renderer, string? expected) => Change(renderer, null, expected, true);
}
