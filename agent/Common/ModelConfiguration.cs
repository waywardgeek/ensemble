using System.Collections.Immutable;

namespace Ensemble;

public sealed partial class Ensemble
{
    /// <summary>Load a strict, credential-free host document atomically. Bindings
    /// must already exist in this root. Existing graph objects and selections are
    /// never replaced; a refused load removes every object it created.</summary>
    public partial void LoadModelConfiguration(ReadOnlyMemory<byte> utf8);
}

internal sealed record ModelVendorDefinition(string ID, string Label,
    ImmutableArray<ModelEndpointDefinition> Endpoints);
internal sealed record ModelEndpointDefinition(string ID, string Address, Surface Surface,
    string Authentication, ImmutableArray<ModelDefinition> Models);
internal sealed record ModelDefinition(string ID, string Name, string Label,
    ModelFeatures Features, long? ServingContextTokens, ModelRequestPolicy? RequestPolicy);

internal static partial class ModelConfigurationDocument
{
    internal const int MaximumBytes = 1024 * 1024;
    internal static partial ImmutableArray<ModelVendorDefinition> Read(ReadOnlyMemory<byte> utf8);
}
