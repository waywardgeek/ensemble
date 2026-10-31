using System.Text;
using System.Text.Json;
using System.Text.Json.Nodes;

namespace Ensemble;

public sealed partial class Memory
{
    internal partial MemorySettings? ReadSettings()
    {
        lock (gate)
        {
            // Only absence means defaults. Corruption, confinement refusal and
            // unreadable files must not silently enable a formerly disabled band.
            Stream stream;
            try
            {
                stream = Parent().Sandbox().OpenFile(HomePath("settings.json"), FileMode.Open, FileAccess.Read);
            }
            catch (FileNotFoundException)
            {
                return null;
            }
            catch (DirectoryNotFoundException)
            {
                return null;
            }
            using (stream)
            {
                byte[] bytes = new byte[16 * 1024 + 1];
                int length = stream.ReadAtLeast(bytes, bytes.Length, throwOnEndOfStream: false);
                if (length == bytes.Length)
                {
                    throw new InvalidDataException("Agent settings exceed 16 KiB");
                }
                using var document = JsonDocument.Parse(bytes.AsMemory(0, length));
                var root = document.RootElement;
                if (root.ValueKind != JsonValueKind.Object
                    || !root.EnumerateObject().Select(field => field.Name).Order(StringComparer.Ordinal)
                        .SequenceEqual(new[] { "memory", "version" })
                    || root.GetProperty("version").GetInt32() != 1)
                {
                    throw new InvalidDataException("Invalid Agent settings document");
                }
                return JsonCodec.ReadMemorySettings(root.GetProperty("memory"));
            }
        }
    }

    internal partial void WriteSettings(MemorySettings settings)
    {
        lock (gate)
        {
            var document = new JsonObject
            {
                ["version"] = 1,
                ["memory"] = JsonCodec.WriteMemorySettings(settings)
            };
            Parent().Sandbox().ReplaceFile(HomePath("settings.json"),
                Encoding.UTF8.GetBytes(document.ToJsonString()));
        }
    }
}
