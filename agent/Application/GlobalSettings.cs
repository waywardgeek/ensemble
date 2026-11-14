using System.Collections.Immutable;
using System.Text;
using System.Text.Json;
using System.Text.Json.Nodes;

namespace Ensemble;

public sealed partial class Ensemble
{
    public partial GlobalSettings? GlobalSettings() { lock (agents) return globalSettings; }

    internal partial void AddGlobalSettings(GlobalSettings child)
    {
        lock (agents)
        {
            ObjectDisposedException.ThrowIf(destroyed, this);
            if (globalSettings is not null || child.parent is not null || child.destroyed)
                throw new InvalidOperationException("Global settings are already owned or unavailable.");
            globalSettings = child;
            child.parent = this;
        }
    }

    internal partial void RemoveGlobalSettings(GlobalSettings child)
    {
        lock (agents)
        {
            if (!ReferenceEquals(globalSettings, child) || !ReferenceEquals(child.parent, this))
                throw new InvalidOperationException("Global settings are not owned here.");
            globalSettings = null;
            child.parent = null;
        }
    }
}

public sealed partial class GlobalSettings
{
    public partial GlobalSettings(Ensemble parent, string path)
    {
        ArgumentNullException.ThrowIfNull(parent);
        ArgumentException.ThrowIfNullOrWhiteSpace(path);
        this.path = Path.GetFullPath(path);
        graphGate = parent.ModelGraphGate;
        lock (graphGate)
        {
            ObjectDisposedException.ThrowIf(parent.ModelGraphClosed, parent);
            _ = ReadDocument();
            parent.AddGlobalSettings(this);
        }
    }

    public partial Ensemble Parent() { lock (graphGate) return parent ?? throw new ObjectDisposedException(nameof(GlobalSettings)); }

    private partial Ensemble RequireOwner()
    {
        var root = parent ?? throw new ObjectDisposedException(nameof(GlobalSettings));
        ObjectDisposedException.ThrowIf(destroyed || root.ModelGraphClosed, this);
        return root;
    }

    public partial ModelAuthentication? BindApiKey(Vendor vendor, string bindingID)
    {
        string field = vendor switch
        {
            Vendor.OpenAI => "openai_api_key",
            Vendor.Anthropic => "anthropic_api_key",
            Vendor.Gemini => "gemini_api_key",
            _ => throw new ArgumentException("No global API-key field for this vendor.", nameof(vendor))
        };
        return BindApiKeyField(field, bindingID, ApiKeyHeader.ProtocolDefault);
    }

    public partial ModelAuthentication? BindOpenRouterApiKey(string bindingID)
    {
        return BindApiKeyField("openrouter_api_key", bindingID, ApiKeyHeader.Bearer);
    }

    private partial ModelAuthentication? BindApiKeyField(string field, string bindingID,
        ApiKeyHeader keyHeader)
    {
        ModelVendor.ValidateID(bindingID);
        lock (graphGate)
        {
            var root = RequireOwner();
            var document = ReadDocument();
            string? key = document[field]?.GetValue<string>();
            // The binding owns the key. No key is exposed in the returned
            // preference map or copied into a Model or Agent settings file.
            if (string.IsNullOrWhiteSpace(key))
            {
                return null;
            }
            return new ModelAuthentication(root, bindingID, key, keyHeader);
        }
    }

    public partial ImmutableDictionary<string, string> FundingDefaults()
    {
        lock (graphGate)
        {
            _ = RequireOwner();
            return ReadFunding(ReadDocument());
        }
    }

    public partial void RememberFunding(Model offering)
    {
        ArgumentNullException.ThrowIfNull(offering);
        // Reject foreign ownership before taking another root's graph gate.
        if (!ReferenceEquals(offering.graphGate, graphGate))
            throw new ArgumentException("The offering belongs to another application.", nameof(offering));
        lock (graphGate)
        {
            var root = RequireOwner();
            if (offering.destroyed || !ReferenceEquals(offering.Vendor().Parent(), root))
                throw new InvalidOperationException("The offering is unavailable.");
            var endpoint = offering.Parent();
            var authentication = endpoint.Authentication();
            authentication.ValidateEndpoint(endpoint.Address, endpoint.Surface);
            Directory.CreateDirectory(Path.GetDirectoryName(path)!);
            // Cooperating processes serialize read/modify/replace using this
            // persistent lock-file name. Contention refuses; it is not a retry.
            var options = new FileStreamOptions
            {
                Mode = FileMode.OpenOrCreate, Access = FileAccess.ReadWrite, Share = FileShare.None
            };
            if (!OperatingSystem.IsWindows()) options.UnixCreateMode = UnixFileMode.UserRead | UnixFileMode.UserWrite;
            using var held = new FileStream(path + ".lock", options);
            var document = ReadDocument();
            var defaults = document[FundingField] as JsonObject ?? new JsonObject();
            defaults[offering.Vendor().ID] = authentication.ID;
            document[FundingField] = defaults;
            ValidateDocument(document);
            WriteDocument(document);
        }
    }

    public partial Model ResolveModel(ModelVendor vendor, string name)
    {
        ArgumentNullException.ThrowIfNull(vendor);
        ArgumentException.ThrowIfNullOrWhiteSpace(name);
        if (!ReferenceEquals(vendor.graphGate, graphGate))
            throw new ArgumentException("The vendor belongs to another application.", nameof(vendor));
        lock (graphGate)
        {
            var root = RequireOwner();
            if (vendor.destroyed || !ReferenceEquals(vendor.Parent(), root))
                throw new InvalidOperationException("The vendor is unavailable.");
            var defaults = ReadFunding(ReadDocument());
            defaults.TryGetValue(vendor.ID, out string? funding);
            var candidates = vendor.Endpoints().SelectMany(endpoint => endpoint.Models())
                .Where(model => model.Name == name
                    && (funding is null || model.Parent().authentication?.ID == funding)).ToArray();
            if (candidates.Length != 1)
                throw new InvalidOperationException(funding is null
                    ? "Choose an explicit offering: vendor funding is missing or ambiguous."
                    : "Saved vendor funding does not identify one compatible offering. No fallback was used.");
            var selected = candidates[0];
            var endpoint = selected.Parent();
            endpoint.Authentication().ValidateEndpoint(endpoint.Address, endpoint.Surface);
            return selected;
        }
    }

    public partial void Destroy()
    {
        lock (graphGate)
        {
            if (destroyed) return;
            destroyed = true;
            parent?.RemoveGlobalSettings(this);
        }
    }

    private partial JsonObject ReadDocument()
    {
        Stream stream;
        try { stream = new FileStream(path, FileMode.Open, FileAccess.Read, FileShare.Read | FileShare.Delete); }
        catch (FileNotFoundException) { return new JsonObject(); }
        catch (DirectoryNotFoundException) { return new JsonObject(); }
        using (stream)
        {
            byte[] bytes = new byte[MaximumBytes + 1];
            int count = stream.ReadAtLeast(bytes, bytes.Length, throwOnEndOfStream: false);
            if (count > MaximumBytes) throw new InvalidDataException("Global settings exceed 1 MiB.");
            _ = new UTF8Encoding(false, true).GetCharCount(bytes, 0, count);
            using var parsed = JsonDocument.Parse(bytes.AsMemory(0, count), new JsonDocumentOptions { MaxDepth = 16 });
            CheckDuplicates(parsed.RootElement);
            var document = JsonNode.Parse(parsed.RootElement.GetRawText()) as JsonObject
                ?? throw new InvalidDataException("Global settings must be an object.");
            ValidateDocument(document);
            return document;
        }

        static void CheckDuplicates(JsonElement element)
        {
            if (element.ValueKind == JsonValueKind.Object)
            {
                var names = new HashSet<string>(StringComparer.Ordinal);
                foreach (var property in element.EnumerateObject())
                {
                    if (!names.Add(property.Name)) throw new InvalidDataException("Duplicate global settings field.");
                    CheckDuplicates(property.Value);
                }
            }
            else if (element.ValueKind == JsonValueKind.Array)
                foreach (var item in element.EnumerateArray()) CheckDuplicates(item);
        }
    }

    private static partial ImmutableDictionary<string, string> ReadFunding(JsonObject document)
    {
        var result = ImmutableDictionary.CreateBuilder<string, string>(StringComparer.Ordinal);
        if (!document.TryGetPropertyValue(FundingField, out var node)) return result.ToImmutable();
        if (node is not JsonObject defaults || defaults.Count > 64)
            throw new InvalidDataException("Invalid vendor funding defaults.");
        foreach (var field in defaults)
        {
            ModelVendor.ValidateID(field.Key);
            if (field.Value is not JsonValue value || !value.TryGetValue<string>(out string? binding)
                || string.IsNullOrWhiteSpace(binding)) throw new InvalidDataException("Invalid saved funding reference.");
            ModelVendor.ValidateID(binding);
            result.Add(field.Key, binding);
        }
        return result.ToImmutable();
    }

    private static partial void ValidateDocument(JsonObject document)
    {
        foreach (string field in new[] { "openai_api_key", "anthropic_api_key", "gemini_api_key", "openrouter_api_key" })
            if (document.TryGetPropertyValue(field, out var node)
                && (node is not JsonValue value || !value.TryGetValue<string>(out _)))
                throw new InvalidDataException("Global API-key fields must be strings.");
        _ = ReadFunding(document);
    }

    private partial void WriteDocument(JsonObject document)
    {
        byte[] bytes = Encoding.UTF8.GetBytes(document.ToJsonString());
        if (bytes.Length > MaximumBytes) throw new InvalidDataException("Global settings exceed 1 MiB.");
        string temporary = path + "." + Guid.NewGuid().ToString("N") + ".tmp";
        try
        {
            var options = new FileStreamOptions { Mode = FileMode.CreateNew, Access = FileAccess.Write, Share = FileShare.None };
            if (!OperatingSystem.IsWindows()) options.UnixCreateMode = UnixFileMode.UserRead | UnixFileMode.UserWrite;
            using (var stream = new FileStream(temporary, options))
            {
                stream.Write(bytes);
                stream.Flush(flushToDisk: true);
            }
            File.Move(temporary, path, overwrite: true);
        }
        finally { if (File.Exists(temporary)) File.Delete(temporary); }
    }
}
