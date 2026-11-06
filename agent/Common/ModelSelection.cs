using System.Collections.Immutable;

namespace Ensemble;

/// <summary>Documentary choice information, never permission to send. Availability
/// checks local relationships and request policy, not remote account entitlement
/// or server health. No endpoint address or credential is exposed here.</summary>
public sealed record ModelSelectionChoice(ModelIdentity Offering, string VendorLabel,
    string EndpointID, string Model, string Label, string Funding,
    long? ServingContextTokens, long? AdvertisedContextTokens, bool Available);

/// <summary>A coherent display snapshot. A lost selection retains its last identity
/// so a user can explicitly choose a replacement without a routing fallback.</summary>
public sealed record ModelSelectionStatus(Vendor Vendor, string Model, ModelIdentity? Offering,
    string Funding, bool Available, ImmutableArray<string> LegacyChoices,
    ImmutableArray<ModelSelectionChoice> ConfiguredChoices);

public sealed partial class Agent
{
    public partial ModelSelectionStatus ModelSelectionStatus();
}
