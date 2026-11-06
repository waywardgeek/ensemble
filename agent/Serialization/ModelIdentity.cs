namespace Ensemble;

public sealed partial record ModelIdentity
{
    public partial ModelIdentity(string vendorID, string endpointID, string modelID)
    {
        ModelVendor.ValidateID(vendorID);
        ModelVendor.ValidateID(endpointID);
        ModelVendor.ValidateID(modelID);
        VendorID = vendorID;
        EndpointID = endpointID;
        ModelID = modelID;
    }
}
