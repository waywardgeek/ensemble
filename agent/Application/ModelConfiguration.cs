namespace Ensemble;

public sealed partial class Ensemble
{
    public partial void LoadModelConfiguration(ReadOnlyMemory<byte> utf8)
    {
        var definitions = ModelConfigurationDocument.Read(utf8);
        lock (ModelGraphGate)
        {
            ObjectDisposedException.ThrowIf(destroyed, this);
            var created = new List<ModelVendor>();
            try
            {
                foreach (ModelVendorDefinition definition in definitions)
                {
                    var vendor = new ModelVendor(this, definition.ID, definition.Label);
                    created.Add(vendor);
                    foreach (ModelEndpointDefinition endpointDefinition in definition.Endpoints)
                    {
                        // Borrow the actual root-owned binding. The document cannot
                        // create credentials, read their files, or discover replacements.
                        ModelAuthentication binding = modelAuthentications.FirstOrDefault(
                            item => item.ID == endpointDefinition.Authentication)
                            ?? throw new InvalidDataException("Configured authentication binding is unavailable.");
                        var endpoint = new ModelEndpoint(vendor, endpointDefinition.ID,
                            endpointDefinition.Address, endpointDefinition.Surface, binding);
                        foreach (ModelDefinition model in endpointDefinition.Models)
                        {
                            _ = new Model(endpoint, model.ID, model.Name, model.Label,
                                model.Features, model.ServingContextTokens, model.RequestPolicy);
                        }
                    }
                }
            }
            catch
            {
                // No observer can acquire the graph gate until publication or
                // rollback completes. Vendors own only this newly built subtree;
                // endpoint destruction detaches but never destroys borrowed bindings.
                foreach (ModelVendor vendor in created) vendor.Destroy();
                throw;
            }
        }
    }
}
