using System.Collections.Immutable;

namespace Ensemble;

public enum ModelAuthenticationKind { None = 1, ApiKey, ChatGptPlan, Codex }

public sealed partial class Ensemble
{
    private readonly List<ModelAuthentication> modelAuthentications = [];
    public partial ImmutableArray<ModelAuthentication> ModelAuthentications();
    internal partial void AddModelAuthentication(ModelAuthentication child);
    internal partial void RemoveModelAuthentication(ModelAuthentication child);
    internal partial void DetachModelBindings(Credentials source);
    internal partial void DetachModelBindings(CodexCredentials source);
}

public sealed partial class Credentials
{
    // Weak reverse relationships, protected by the owning root's model graph gate.
    internal readonly List<ModelAuthentication> modelBindings = [];
}

public sealed partial class CodexCredentials
{
    internal readonly List<ModelAuthentication> modelBindings = [];
}

/// <summary>Explicit authentication/funding choice, strongly owned by Ensemble.
/// It owns an API key only in API-key mode. OAuth/Codex stores remain independent
/// owners with weak, bidirectional bindings. No token is copied into a Model.
/// Losing a source invalidates the binding; it never changes its Kind.</summary>
public sealed partial class ModelAuthentication
{
    internal Ensemble? parent;
    internal bool destroyed;
    internal readonly object graphGate;
    internal Credentials? planCredentials;
    internal CodexCredentials? codexCredentials;
    private readonly List<ModelEndpoint> endpoints = [];
    private readonly string apiKey;
    public readonly string ID;
    public readonly ModelAuthenticationKind Kind;
    public readonly string Account;

    public partial ModelAuthentication(Ensemble parent, string id);
    public partial ModelAuthentication(Ensemble parent, string id, string apiKey);
    public partial ModelAuthentication(Ensemble parent, string id, Credentials credentials, string account);
    public partial ModelAuthentication(Ensemble parent, string id, CodexCredentials credentials);
    private partial ModelAuthentication(Ensemble parent, string id, ModelAuthenticationKind kind,
        string apiKey, Credentials? planCredentials, CodexCredentials? codexCredentials, string account);
    public partial Ensemble Parent();
    public partial Credentials? PlanCredentials();
    public partial CodexCredentials? CodexCredentials();
    public partial ImmutableArray<ModelEndpoint> Endpoints();
    internal partial void AddEndpoint(ModelEndpoint endpoint);
    internal partial void RemoveEndpoint(ModelEndpoint endpoint);
    internal partial void ValidateEndpoint(Uri address, Surface surface);
    internal partial string DiagnosticKey();
    public partial void Destroy();
}

public sealed partial class ModelEndpoint
{
    internal ModelAuthentication? authentication;
    public partial ModelAuthentication Authentication();
}
