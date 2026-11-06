using System.Collections.Immutable;
using System.Net;

namespace Ensemble;

public sealed partial class Ensemble
{
    public partial ImmutableArray<ModelAuthentication> ModelAuthentications()
    {
        lock (ModelGraphGate)
        {
            return modelAuthentications.ToImmutableArray();
        }
    }

    internal partial void AddModelAuthentication(ModelAuthentication child)
    {
        lock (ModelGraphGate)
        {
            ObjectDisposedException.ThrowIf(destroyed, this);
            if (child.destroyed || child.parent is not null || modelAuthentications.Contains(child)
                || modelAuthentications.Any(binding => binding.ID == child.ID)
                || !ReferenceEquals(child.graphGate, ModelGraphGate))
            {
                throw new InvalidOperationException("Invalid or duplicate model authentication registration.");
            }
            if (child.planCredentials is { } plan && (plan.destroyed || !ReferenceEquals(plan.parent, this)))
            {
                throw new InvalidOperationException("Plan credentials belong to a different or closed root.");
            }
            if (child.codexCredentials is { } codex && (codex.destroyed || !ReferenceEquals(codex.parent, this)))
            {
                throw new InvalidOperationException("Codex credentials belong to a different or closed root.");
            }
            modelAuthentications.Add(child);
            child.parent = this;
            child.planCredentials?.modelBindings.Add(child);
            child.codexCredentials?.modelBindings.Add(child);
        }
    }

    internal partial void RemoveModelAuthentication(ModelAuthentication child)
    {
        lock (ModelGraphGate)
        {
            if (!ReferenceEquals(child.parent, this) || !modelAuthentications.Contains(child))
            {
                throw new InvalidOperationException("Model authentication is not owned here.");
            }
            child.planCredentials?.modelBindings.Remove(child);
            child.codexCredentials?.modelBindings.Remove(child);
            child.planCredentials = null;
            child.codexCredentials = null;
            modelAuthentications.Remove(child);
            child.parent = null;
        }
    }

    internal partial void DetachModelBindings(Credentials source)
    {
        lock (ModelGraphGate)
        {
            foreach (ModelAuthentication binding in source.modelBindings)
            {
                binding.planCredentials = null;
            }
            source.modelBindings.Clear();
        }
    }

    internal partial void DetachModelBindings(CodexCredentials source)
    {
        lock (ModelGraphGate)
        {
            foreach (ModelAuthentication binding in source.modelBindings)
            {
                binding.codexCredentials = null;
            }
            source.modelBindings.Clear();
        }
    }
}

public sealed partial class ModelAuthentication
{
    public partial ModelAuthentication(Ensemble parent, string id)
        : this(parent, id, ModelAuthenticationKind.None, "", null, null, "") { }

    public partial ModelAuthentication(Ensemble parent, string id, string apiKey)
        : this(parent, id, ModelAuthenticationKind.ApiKey, apiKey, null, null, "") { }

    public partial ModelAuthentication(Ensemble parent, string id, Credentials credentials, string account)
        : this(parent, id, ModelAuthenticationKind.ChatGptPlan, "",
            credentials ?? throw new ArgumentNullException(nameof(credentials)), null, account) { }

    public partial ModelAuthentication(Ensemble parent, string id, CodexCredentials credentials)
        : this(parent, id, ModelAuthenticationKind.Codex, "", null,
            credentials ?? throw new ArgumentNullException(nameof(credentials)), "") { }

    private partial ModelAuthentication(Ensemble parent, string id, ModelAuthenticationKind kind,
        string apiKey, Credentials? planCredentials, CodexCredentials? codexCredentials, string account)
    {
        ArgumentNullException.ThrowIfNull(parent);
        ModelVendor.ValidateID(id);
        if (kind == ModelAuthenticationKind.ApiKey)
        {
            ArgumentException.ThrowIfNullOrWhiteSpace(apiKey);
            if (apiKey.Any(char.IsControl)) throw new ArgumentException("API key contains invalid control characters.");
        }
        if (kind == ModelAuthenticationKind.ChatGptPlan)
        {
            ModelVendor.ValidateLabel(account);
        }
        ID = id;
        Kind = kind;
        Account = account;
        this.apiKey = apiKey;
        this.planCredentials = planCredentials;
        this.codexCredentials = codexCredentials;
        graphGate = parent.ModelGraphGate;
        parent.AddModelAuthentication(this);
    }

    public partial Ensemble Parent()
    {
        lock (graphGate)
        {
            return parent ?? throw new ObjectDisposedException(nameof(ModelAuthentication));
        }
    }

    public partial Credentials? PlanCredentials()
    {
        lock (graphGate) { return planCredentials; }
    }

    public partial CodexCredentials? CodexCredentials()
    {
        lock (graphGate) { return codexCredentials; }
    }

    public partial ImmutableArray<ModelEndpoint> Endpoints()
    {
        lock (graphGate) { return endpoints.ToImmutableArray(); }
    }

    internal partial void ValidateEndpoint(Uri address, Surface surface)
    {
        lock (graphGate)
        {
            ObjectDisposedException.ThrowIf(destroyed || parent is null || parent.ModelGraphClosed, this);
            bool loopback = IPAddress.TryParse(address.Host, out IPAddress? ip) && IPAddress.IsLoopback(ip);
            if (Kind != ModelAuthenticationKind.None && address.Scheme != "https" && !loopback)
            {
                throw new ArgumentException("Credential-bearing model endpoints require HTTPS or literal loopback.");
            }
            switch (Kind)
            {
                case ModelAuthenticationKind.ChatGptPlan:
                    if (planCredentials is null || planCredentials.destroyed)
                        throw new InvalidOperationException("Selected plan credential owner is unavailable; no fallback.");
                    if (surface != Surface.Responses)
                        throw new ArgumentException("Plan funding requires the Responses surface.");
                    Credentials.ValidatePlanEndpoint(address);
                    break;
                case ModelAuthenticationKind.Codex:
                    if (codexCredentials is null || codexCredentials.destroyed)
                        throw new InvalidOperationException("Selected Codex credential owner is unavailable; no fallback.");
                    if (surface != Surface.Responses)
                        throw new ArgumentException("Codex funding requires the Responses surface.");
                    global::Ensemble.CodexCredentials.ValidateEndpoint(address);
                    break;
            }
        }
    }

    internal partial void AddEndpoint(ModelEndpoint endpoint)
    {
        lock (graphGate)
        {
            ValidateEndpoint(endpoint.Address, endpoint.Surface);
            if (endpoint.destroyed || endpoint.authentication is not null || endpoints.Contains(endpoint)
                || !ReferenceEquals(endpoint.graphGate, graphGate))
            {
                throw new InvalidOperationException("Invalid model endpoint authentication relationship.");
            }
            endpoints.Add(endpoint);
            endpoint.authentication = this;
        }
    }

    internal partial void RemoveEndpoint(ModelEndpoint endpoint)
    {
        lock (graphGate)
        {
            if (!ReferenceEquals(endpoint.authentication, this) || !endpoints.Contains(endpoint))
            {
                throw new InvalidOperationException("Endpoint is not bound to this authentication owner.");
            }
            endpoints.Remove(endpoint);
            endpoint.authentication = null;
        }
    }

    internal partial string DiagnosticKey() => apiKey;

    public partial void Destroy()
    {
        lock (graphGate)
        {
            if (destroyed) return;
            destroyed = true;
            foreach (ModelEndpoint endpoint in endpoints.ToArray())
            {
                RemoveEndpoint(endpoint);
            }
            parent?.RemoveModelAuthentication(this);
        }
    }
}

public sealed partial class ModelEndpoint
{
    public partial ModelAuthentication Authentication()
    {
        lock (graphGate)
        {
            ObjectDisposedException.ThrowIf(destroyed, this);
            return authentication ?? throw new InvalidOperationException("Endpoint authentication is unavailable; no fallback.");
        }
    }
}
