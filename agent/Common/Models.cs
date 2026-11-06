using System.Collections.Immutable;

namespace Ensemble;

public sealed partial class Ensemble
{
    // The configured inference graph shares the root membership gate. Admission
    // and root shutdown therefore cannot race to leave an unowned vendor alive.
    private readonly List<ModelVendor> modelVendors = [];
    internal object ModelGraphGate => agents;
    internal bool ModelGraphClosed => destroyed;
    public partial ImmutableArray<ModelVendor> ModelVendors();
    internal partial void AddModelVendor(ModelVendor child);
    internal partial void RemoveModelVendor(ModelVendor child);
}

/// <summary>A configured vendor, strongly owned by Ensemble. This identity is
/// independent of the historical Vendor enum used by wire codecs: an Ollama
/// endpoint may speak Chat Completions without becoming the OpenAI vendor.</summary>
public sealed partial class ModelVendor
{
    internal Ensemble? parent;
    internal bool destroyed;
    internal readonly object graphGate;
    private readonly List<ModelEndpoint> endpoints = [];
    public readonly string ID;
    public readonly string Label;

    public partial ModelVendor(Ensemble parent, string id, string label);
    public partial Ensemble Parent();
    public partial ImmutableArray<ModelEndpoint> Endpoints();
    internal partial void AddEndpoint(ModelEndpoint child);
    internal partial void RemoveEndpoint(ModelEndpoint child);
    public partial void Destroy();
    internal static partial void ValidateID(string id);
    internal static partial void ValidateLabel(string label);
}

/// <summary>A configured service and wire surface, strongly owned by its vendor.
/// It owns configured model offerings, not requests or credential renewal work.
/// This graph alone grants no transport access.</summary>
public sealed partial class ModelEndpoint
{
    internal ModelVendor? parent;
    internal bool destroyed;
    internal readonly object graphGate;
    private readonly List<Model> models = [];
    public readonly string ID;
    public readonly Uri Address;
    public readonly Surface Surface;

    public partial ModelEndpoint(ModelVendor parent, string id, string address, Surface surface,
        ModelAuthentication authentication);
    public partial ModelVendor Parent();
    public partial ImmutableArray<Model> Models();
    internal partial void AddModel(Model child);
    internal partial void RemoveModel(Model child);
    public partial void Destroy();
}

/// <summary>A configured offering, strongly owned by one endpoint. Name is the
/// provider-facing model name; ID is its stable local configuration identity.
/// Two endpoints may offer the same Name without sharing identity or funding.</summary>
public sealed partial class Model
{
    internal ModelEndpoint? parent;
    internal bool destroyed;
    internal readonly object graphGate;
    public readonly string ID;
    public readonly string Name;
    public readonly string Label;
    public readonly ModelFeatures Features;
    // Configured serving capacity, not inferred from a model's advertised maximum
    // and not a provider-reported next-request token count. Null means unknown.
    public readonly long? ServingContextTokens;

    public partial Model(ModelEndpoint parent, string id, string name, string label,
        ModelFeatures features, long? servingContextTokens = null);
    public partial ModelEndpoint Parent();
    public partial ModelVendor Vendor();
    public partial void Destroy();
}
