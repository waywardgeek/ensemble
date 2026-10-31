namespace Ensemble;

public sealed partial class Tool
{
    private partial ToolResultPart InspectImage(ToolCallPart call, CancellationToken cancellation)
    {
        try
        {
            var agent = Parent();
            string path = String(call.Args, "path");
            // Validate the real placement before touching the source. A model
            // capability does not imply its adapter supports image tool results.
            var source = new BlobPart("image/png", new BlobRef(RefKind.Path, path));
            agent.Engine().ValidateMedia([new ToolResultPart(call.CallID, [source])]);
            cancellation.ThrowIfCancellationRequested();
            var prepared = agent.PrepareImage(path);
            cancellation.ThrowIfCancellationRequested();
            string change = prepared.Resized ? "Resized" : "Prepared without resizing";
            string report = $"{change}: {prepared.OriginalWidth} × {prepared.OriginalHeight} "
                + $"to {prepared.Width} × {prepared.Height}; "
                + $"{prepared.SourceBytes} source bytes, {prepared.OutputBytes} PNG bytes. "
                + "Original file unchanged. Image pixels are attached.";
            return new ToolResultPart(call.CallID, [new TextPart(report), prepared.Image]);
        }
        catch (Exception error) when (error is InvalidDataException or NotSupportedException)
        {
            // Decode and capability failures are useful tool feedback, not an
            // empty successful image. Ordinary I/O errors use shared dispatch.
            return new ToolResultPart(call.CallID, [new TextPart(error.Message)], isError: true);
        }
    }
}
