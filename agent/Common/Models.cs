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
    /// <summary>Resolve an exact configured offering in this root, or refuse.
    /// This does not substitute by wire name or grant authentication.</summary>
    public partial Model ResolveModel(ModelIdentity identity);
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
    private readonly List<Agent> agents = [];
    public readonly string Name;
    public readonly string Label;
    public readonly ModelFeatures Features;
    public readonly ModelRequestPolicy RequestPolicy;
    // Configured serving capacity, not inferred from a model's advertised maximum
    // and not a provider-reported next-request token count. Null means unknown.
    public readonly long? ServingContextTokens;

    public partial Model(ModelEndpoint parent, string id, string name, string label,
        ModelFeatures features, long? servingContextTokens = null, ModelRequestPolicy? requestPolicy = null);
    public partial ModelEndpoint Parent();
    public partial ModelVendor Vendor();
    public partial ModelIdentity Identity();
    /// <summary>Weak selections; destroying a Model does not destroy these Agents.</summary>
    public partial ImmutableArray<Agent> Agents();
    internal partial void AddAgent(Agent agent);
    internal partial void RemoveAgent(Agent agent);
    public partial void Destroy();
}

/// <summary>Credential-free configuration identity. IDs are scoped through the
/// actual vendor and endpoint, rather than confused with a provider model name.</summary>
public sealed partial record ModelIdentity
{
    public readonly string VendorID;
    public readonly string EndpointID;
    public readonly string ModelID;
    public partial ModelIdentity(string vendorID, string endpointID, string modelID);
}

public sealed partial class Agent
{
    // Only Model relationship operations write this weak link under the root
    // graph gate. A configured projection with no link is unavailable, not legacy.
    internal Model? selectedModel;
    public partial Model? SelectedModel();
    private partial void AttachConfiguredModel();
    private partial void DetachConfiguredModel();
    private partial Configuration ValidatedConfiguration();
    internal partial Configuration DocumentaryConfiguration();
}
