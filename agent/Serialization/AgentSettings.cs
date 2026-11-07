using System.Text.Json;

namespace Ensemble;

internal sealed partial record AgentSettings
{
    internal partial void Write(Sandbox sandbox, string path)
    {
        var document = new System.Text.Json.Nodes.JsonObject
        {
            ["version"] = 1,
            ["memory"] = JsonCodec.WriteMemorySettings(Memory),
            ["recall"] = JsonCodec.WriteRecallSettings(Recall),
            ["working_context"] = JsonCodec.WriteWorkingContextLimits(WorkingContext.IsDefault ? [] : WorkingContext)
        };
        if (SelectedOffering is { } offering)
            document["selected_offering"] = JsonCodec.WriteModelIdentity(offering);
        byte[] bytes = System.Text.Encoding.UTF8.GetBytes(document.ToJsonString());
        if (bytes.Length > 16 * 1024)
            throw new InvalidDataException("Agent settings exceed 16 KiB");
        sandbox.ReplaceFile(path, bytes);
    }

    internal static partial AgentSettings? Read(Sandbox sandbox, string path)
    {
        // Only absence means defaults. Corruption, confinement refusal and
        // unreadable files must not silently enable a disabled band or recall.
        Stream stream;
        try
        {
            stream = sandbox.OpenFile(path, FileMode.Open, FileAccess.Read);
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
            if (root.ValueKind != JsonValueKind.Object)
            {
                throw new InvalidDataException("Invalid Agent settings document");
            }
            var fields = root.EnumerateObject().Select(field => field.Name).ToArray();
            if (fields.Distinct(StringComparer.Ordinal).Count() != fields.Length)
            {
                throw new InvalidDataException("Duplicate Agent settings field");
            }
            bool hasSelection = root.TryGetProperty("selected_offering", out var selected);
            var names = fields.Where(name => name != "selected_offering").Order(StringComparer.Ordinal).ToArray();
            // Chapter 22's complete memory-only document remains readable.
            // Supplied recall is strict; null/malformed never means defaults.
            bool legacy = names.SequenceEqual(new[] { "memory", "version" });
            bool complete = names.SequenceEqual(new[] { "memory", "recall", "version" });
            bool withContext = names.SequenceEqual(new[] { "memory", "recall", "version", "working_context" });
            if ((!legacy && !complete && !withContext) || root.GetProperty("version").GetInt32() != 1)
            {
                throw new InvalidDataException("Invalid Agent settings document");
            }
            var memory = JsonCodec.ReadMemorySettings(root.GetProperty("memory"));
            var recall = legacy ? new RecallSettings()
                : JsonCodec.ReadRecallSettings(root.GetProperty("recall"));
            var limits = withContext ? JsonCodec.ReadWorkingContextLimits(root.GetProperty("working_context")) : [];
            var offering = hasSelection ? JsonCodec.ReadModelIdentity(selected) : null;
            return new AgentSettings(memory, recall, limits, offering);
        }
    }
}
