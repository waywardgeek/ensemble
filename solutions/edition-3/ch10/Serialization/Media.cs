namespace Ensemble;

public sealed partial class BlobRef
{
    public partial byte[] Data()
    {
        if (Kind != RefKind.Path)
        {
            throw new NotSupportedException("Only local-path media can be resolved in Chapter 5.");
        }
        return File.ReadAllBytes(Locator);
    }
}
