namespace Ensemble;

public sealed partial class BlobRef
{
    public partial byte[] Data()
    {
        if (Kind == RefKind.Path)
        {
            return File.ReadAllBytes(Locator);
        }
        if (Kind == RefKind.Uri && Locator.StartsWith("data:", StringComparison.Ordinal))
        {
            int separator = Locator.IndexOf(";base64,", StringComparison.Ordinal);
            if (separator > "data:".Length)
            {
                // This is inline content, not a URL fetch. Keep the locator in
                // History and decode only for a request or explicit consumer.
                return Convert.FromBase64String(Locator[(separator + ";base64,".Length)..]);
            }
        }
        throw new NotSupportedException("Only local paths and base64 data URIs can be resolved.");
    }
}
