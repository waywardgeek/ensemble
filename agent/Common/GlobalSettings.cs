using System.Collections.Immutable;
using System.Text.Json.Nodes;

namespace Ensemble;

public sealed partial class Ensemble
{
    private GlobalSettings? globalSettings;
    public partial GlobalSettings? GlobalSettings();
    internal partial void AddGlobalSettings(GlobalSettings child);
    internal partial void RemoveGlobalSettings(GlobalSettings child);
}

/// <summary>Application-owned global settings file, explicitly supplied by the
/// host. Agent policy remains in each Agent's settings.json. Reads never create
/// files; writes preserve unrelated fields and never retarget an Agent.</summary>
public sealed partial class GlobalSettings
{
    internal Ensemble? parent;
    internal bool destroyed;
    private readonly object graphGate;
    private readonly string path;
    private const int MaximumBytes = 1024 * 1024;
    private const string FundingField = "model_funding_defaults";

    public partial GlobalSettings(Ensemble parent, string path);
    public partial Ensemble Parent();
    /// <summary>Create an ordinary root-owned API-key binding from the named
    /// vendor field. Missing/empty keys return null, never another vendor's key.</summary>
    public partial ModelAuthentication? BindApiKey(Vendor vendor, string bindingID);
    public partial ImmutableDictionary<string, string> FundingDefaults();
    /// <summary>Persist the explicitly selected offering's vendor and funding
    /// binding ID. This does not select a Model or change any Agent.</summary>
    public partial void RememberFunding(Model offering);
    /// <summary>Resolve a wire model within a configured vendor using its saved
    /// funding. With no saved choice, only one configured offering is unambiguous.
    /// An unavailable/incompatible saved choice never falls back.</summary>
    public partial Model ResolveModel(ModelVendor vendor, string name);
    public partial void Destroy();
    private partial Ensemble RequireOwner();
    private partial JsonObject ReadDocument();
    private partial void WriteDocument(JsonObject document);
    private static partial ImmutableDictionary<string, string> ReadFunding(JsonObject document);
    private static partial void ValidateDocument(JsonObject document);
}
