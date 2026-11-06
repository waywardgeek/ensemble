using System.Collections.Immutable;

namespace Ensemble;

public sealed partial class Agent
{
    public partial ModelSelectionStatus ModelSelectionStatus()
    {
        lock (ingress)
        {
            ObjectDisposedException.ThrowIf(destroyed, this);
            Ensemble root = Parent();
            lock (root.ModelGraphGate)
            {
                bool available = !root.ModelGraphClosed;
                if (available && configuration.Offering is not null)
                {
                    try
                    {
                        _ = ValidatedConfiguration();
                        configuration.Authentication!.ValidateEndpoint(
                            selectedModel!.Parent().Address, configuration.Surface);
                    }
                    catch (Exception error) when (error is InvalidOperationException or ArgumentException)
                    {
                        available = false;
                    }
                }

                var choices = ImmutableArray.CreateBuilder<ModelSelectionChoice>();
                foreach (ModelVendor vendor in root.ModelVendors())
                {
                    foreach (ModelEndpoint endpoint in vendor.Endpoints())
                    {
                        foreach (Model model in endpoint.Models())
                        {
                            bool selectable = !root.ModelGraphClosed;
                            string funding = "Unavailable authentication";
                            if (selectable)
                            {
                                try
                                {
                                    // Reuse selection validation without publishing a relationship,
                                    // reading credentials, renewing tokens, or sending a request.
                                    var candidate = new Configuration(configuration, model);
                                    funding = candidate.FundingDescription();
                                }
                                catch (Exception error) when (error is InvalidOperationException or ArgumentException)
                                {
                                    selectable = false;
                                }
                            }
                            choices.Add(new ModelSelectionChoice(model.Identity(), vendor.Label,
                                endpoint.ID, model.Name, model.Label, funding,
                                model.ServingContextTokens, model.Features.ContextWindow, selectable));
                        }
                    }
                }
                // A configured Agent cannot use a wire-name-only choice to escape
                // its offering. Explicit configured choices remain available for recovery.
                var legacy = configuration.Offering is null && available
                    ? global::Ensemble.Engine.ModelChoices(configuration) : ImmutableArray<string>.Empty;
                return new ModelSelectionStatus(configuration.Vendor, configuration.Model,
                    configuration.Offering, configuration.FundingDescription(), available,
                    legacy, choices.ToImmutable());
            }
        }
    }
}
