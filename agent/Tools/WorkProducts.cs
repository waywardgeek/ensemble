namespace Ensemble;

public sealed partial class Tool
{
    private partial ToolResultPart OfferWorkProduct(ToolCallPart call, CancellationToken cancellation)
    {
        cancellation.ThrowIfCancellationRequested();
        var agent = Parent();
        // Dispatch already checked the grant; recheck at the actual file effect.
        if (!agent.ToolEnabled(Name()))
            throw new UnauthorizedAccessException("Work-product authority is not active.");
        try
        {
            var offer = agent.OfferWorkProduct(String(call.Args, "path"));
            string text = $"Offered {offer.Name} ({offer.OfferedBytes} bytes when offered). "
                + "The human can click Download to read current file contents. "
                + "The offer expires when this Agent is destroyed or the application restarts.";
            return new ToolResultPart(call.CallID, [new TextPart(text)], workProduct: offer);
        }
        catch (InvalidOperationException)
        {
            return new ToolResultPart(call.CallID,
                [new TextPart("Work-product offers are unavailable or this Agent has reached its 64-offer limit.")], isError: true);
        }
    }
}
