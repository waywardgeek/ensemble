using System.Collections.Immutable;

namespace Ensemble;

public sealed partial class Ensemble
{
    public partial ImmutableArray<ModelVendor> ModelVendors()
    {
        lock (ModelGraphGate)
        {
            return modelVendors.ToImmutableArray();
        }
    }

    internal partial void AddModelVendor(ModelVendor child)
    {
        lock (ModelGraphGate)
        {
            ObjectDisposedException.ThrowIf(destroyed, this);
            if (child.destroyed || child.parent is not null || modelVendors.Contains(child)
                || modelVendors.Any(vendor => vendor.ID == child.ID)
                || !ReferenceEquals(child.graphGate, ModelGraphGate))
            {
                throw new InvalidOperationException("Invalid or duplicate model vendor registration.");
            }
            modelVendors.Add(child);
            child.parent = this;
        }
    }

    internal partial void RemoveModelVendor(ModelVendor child)
    {
        lock (ModelGraphGate)
        {
            if (!ReferenceEquals(child.parent, this) || !modelVendors.Contains(child))
            {
                throw new InvalidOperationException("Model vendor is not owned here.");
            }
            modelVendors.Remove(child);
            child.parent = null;
        }
    }
}

public sealed partial class ModelVendor
{
    public partial ModelVendor(Ensemble parent, string id, string label)
    {
        ArgumentNullException.ThrowIfNull(parent);
        ValidateID(id);
        ValidateLabel(label);
        ID = id;
        Label = label;
        graphGate = parent.ModelGraphGate;
        parent.AddModelVendor(this);
    }

    internal static partial void ValidateID(string id)
    {
        // Path-independent segments compose a future persisted selection key.
        // Reject ambiguous separators now rather than normalize saved identities.
        ArgumentException.ThrowIfNullOrWhiteSpace(id);
        if (id.Length > 128 || !id.All(c => char.IsAsciiLetterOrDigit(c) || c is '-' or '_'))
        {
            throw new ArgumentException("Model configuration IDs require ASCII letters, digits, hyphen or underscore.");
        }
    }

    internal static partial void ValidateLabel(string label)
    {
        ArgumentException.ThrowIfNullOrWhiteSpace(label);
        if (label.Length > 256 || label.Any(char.IsControl))
        {
            throw new ArgumentException("Model labels must be bounded display text without control characters.");
        }
    }

    public partial Ensemble Parent()
    {
        lock (graphGate)
        {
            return parent ?? throw new ObjectDisposedException(nameof(ModelVendor));
        }
    }

    public partial ImmutableArray<ModelEndpoint> Endpoints()
    {
        lock (graphGate)
        {
            return endpoints.ToImmutableArray();
        }
    }

    internal partial void AddEndpoint(ModelEndpoint child)
    {
        lock (graphGate)
        {
            ObjectDisposedException.ThrowIf(destroyed || parent is null || parent.ModelGraphClosed, this);
            if (child.destroyed || child.parent is not null || endpoints.Contains(child)
                || endpoints.Any(endpoint => endpoint.ID == child.ID)
                || !ReferenceEquals(child.graphGate, graphGate))
            {
                throw new InvalidOperationException("Invalid or duplicate model endpoint registration.");
            }
            endpoints.Add(child);
            child.parent = this;
        }
    }

    internal partial void RemoveEndpoint(ModelEndpoint child)
    {
        lock (graphGate)
        {
            if (!ReferenceEquals(child.parent, this) || !endpoints.Contains(child))
            {
                throw new InvalidOperationException("Model endpoint is not owned here.");
            }
            endpoints.Remove(child);
            child.parent = null;
        }
    }

    public partial void Destroy()
    {
        lock (graphGate)
        {
            if (destroyed) return;
            destroyed = true;
            foreach (ModelEndpoint endpoint in endpoints.ToArray())
            {
                endpoint.Destroy();
            }
            parent?.RemoveModelVendor(this);
        }
    }
}

public sealed partial class ModelEndpoint
{
    public partial ModelEndpoint(ModelVendor parent, string id, string address, Surface surface,
        ModelAuthentication authentication)
    {
        ArgumentNullException.ThrowIfNull(parent);
        ModelVendor.ValidateID(id);
        if (!Enum.IsDefined(surface))
        {
            throw new ArgumentException("Unknown model endpoint surface.");
        }
        if (!Uri.TryCreate(address, UriKind.Absolute, out Uri? uri)
            || uri.Scheme is not ("http" or "https") || uri.UserInfo.Length != 0
            || uri.Query.Length != 0 || uri.Fragment.Length != 0)
        {
            throw new ArgumentException("Model endpoints require an HTTP(S) service address without credentials, query or fragment.");
        }
        ID = id;
        Address = uri;
        Surface = surface;
        graphGate = parent.graphGate;
        ArgumentNullException.ThrowIfNull(authentication);
        if (!ReferenceEquals(authentication.graphGate, graphGate))
        {
            throw new InvalidOperationException("Endpoint authentication belongs to a different root.");
        }
        lock (graphGate)
        {
            // The binding and owning membership publish together. Any refused
            // vendor admission rolls back the weak binding before releasing the gate.
            authentication.AddEndpoint(this);
            try
            {
                parent.AddEndpoint(this);
            }
            catch
            {
                authentication.RemoveEndpoint(this);
                throw;
            }
        }
    }

    public partial ModelVendor Parent()
    {
        lock (graphGate)
        {
            return parent ?? throw new ObjectDisposedException(nameof(ModelEndpoint));
        }
    }

    public partial ImmutableArray<Model> Models()
    {
        lock (graphGate)
        {
            return models.ToImmutableArray();
        }
    }

    internal partial void AddModel(Model child)
    {
        lock (graphGate)
        {
            ObjectDisposedException.ThrowIf(destroyed || parent is null || parent.Parent().ModelGraphClosed, this);
            if (child.destroyed || child.parent is not null || models.Contains(child)
                || models.Any(model => model.ID == child.ID)
                || !ReferenceEquals(child.graphGate, graphGate))
            {
                throw new InvalidOperationException("Invalid or duplicate model registration.");
            }
            models.Add(child);
            child.parent = this;
        }
    }

    internal partial void RemoveModel(Model child)
    {
        lock (graphGate)
        {
            if (!ReferenceEquals(child.parent, this) || !models.Contains(child))
            {
                throw new InvalidOperationException("Model is not owned here.");
            }
            models.Remove(child);
            child.parent = null;
        }
    }

    public partial void Destroy()
    {
        lock (graphGate)
        {
            if (destroyed) return;
            destroyed = true;
            foreach (Model model in models.ToArray())
            {
                model.Destroy();
            }
            authentication?.RemoveEndpoint(this);
            parent?.RemoveEndpoint(this);
        }
    }
}

public sealed partial class Model
{
    public partial Model(ModelEndpoint parent, string id, string name, string label,
        ModelFeatures features, long? servingContextTokens)
    {
        ArgumentNullException.ThrowIfNull(parent);
        ModelVendor.ValidateID(id);
        ModelVendor.ValidateLabel(name);
        ModelVendor.ValidateLabel(label);
        if (features.ContextWindow is <= 0 || servingContextTokens is <= 0
            || (features.ContextWindow is { } maximum && servingContextTokens > maximum))
        {
            throw new ArgumentException("Serving context must be positive and cannot exceed an explicit model maximum.");
        }
        if (features.ResponsesOnly && parent.Surface != Surface.Responses)
        {
            throw new ArgumentException("Model requires the Responses surface.");
        }
        ID = id;
        Name = name;
        Label = label;
        Features = features;
        ServingContextTokens = servingContextTokens;
        graphGate = parent.graphGate;
        parent.AddModel(this);
    }

    public partial ModelEndpoint Parent()
    {
        lock (graphGate)
        {
            return parent ?? throw new ObjectDisposedException(nameof(Model));
        }
    }

    public partial ModelVendor Vendor()
    {
        lock (graphGate)
        {
            return Parent().Parent();
        }
    }

    public partial ModelIdentity Identity()
    {
        lock (graphGate)
        {
            ModelEndpoint endpoint = Parent();
            ModelVendor vendor = endpoint.Parent();
            return new ModelIdentity(vendor.ID, endpoint.ID, ID);
        }
    }

    public partial void Destroy()
    {
        lock (graphGate)
        {
            if (destroyed) return;
            destroyed = true;
            parent?.RemoveModel(this);
        }
    }
}
