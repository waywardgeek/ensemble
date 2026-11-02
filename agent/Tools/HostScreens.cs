using System.Text.Json;

namespace Ensemble;

public sealed partial class Tool
{
    internal static partial (int? Display, ScreenRectangle? Rectangle) ParseScreenRequest(JsonElement args)
    {
        if (args.ValueKind != JsonValueKind.Object
            || args.EnumerateObject().Select(property => property.Name).Distinct(StringComparer.Ordinal).Count() != args.EnumerateObject().Count())
            throw new ArgumentException("Screen request must be an object without duplicate fields.");
        int? Read(string name)
        {
            if (!args.TryGetProperty(name, out var value)) return null;
            if (value.ValueKind != JsonValueKind.Number || !value.TryGetInt32(out int number))
                throw new ArgumentException(name + " must be a 32-bit integer.");
            return number;
        }
        int? display = Read("display");
        int? x = Read("x"), y = Read("y"), width = Read("width"), height = Read("height");
        bool rectangle = x.HasValue || y.HasValue || width.HasValue || height.HasValue;
        if (display < 0) throw new ArgumentException("Display index must not be negative.");
        if (!rectangle) return (display, null);
        if (display.HasValue || !x.HasValue || !y.HasValue || !width.HasValue || !height.HasValue)
            throw new ArgumentException("Supply either display or all four rectangle fields, never both.");
        if (width <= 0 || height <= 0 || (long)width * height > Agent.MaximumImagePixels
            || (long)x + width > int.MaxValue || (long)y + height > int.MaxValue)
            throw new ArgumentException("Screen rectangle has invalid or excessive bounds.");
        return (null, new ScreenRectangle(x.Value, y.Value, width.Value, height.Value));
    }

    private partial async Task<ToolResultPart> HostScreenAsync(ToolCallPart call, CancellationToken cancellation)
    {
        try
        {
            var agent = Parent();
            if (!agent.ToolEnabled(Name())) throw new UnauthorizedAccessException("Host-screen authority is not active.");
            cancellation.ThrowIfCancellationRequested();
            if (kind == ToolKind.ListDisplays)
            {
                var displays = HostScreens.ListDisplays();
                cancellation.ThrowIfCancellationRequested();
                if (!agent.ToolEnabled(Name())) throw new UnauthorizedAccessException("Host-screen authority was revoked.");
                string report = JsonSerializer.Serialize(new
                {
                    coordinates = "Logical host coordinates; primary display upper-left is (0,0).",
                    displays = displays.Select(display => new
                    {
                        display = display.Index, x = display.Bounds.X, y = display.Bounds.Y,
                        width = display.Bounds.Width, height = display.Bounds.Height,
                        pixel_width = display.PixelWidth, pixel_height = display.PixelHeight
                    })
                });
                return new ToolResultPart(call.CallID, [new TextPart(report)]);
            }
            var request = ParseScreenRequest(call.Args);
            // This validates placement/capability only; it never opens a file.
            // No enumeration, permission check or capture precedes this guard.
            var probe = new BlobPart("image/png", new BlobRef(RefKind.Path, "host-screen-capability-probe"));
            agent.Engine().ValidateMedia([new ToolResultPart(call.CallID, [probe])]);
            cancellation.ThrowIfCancellationRequested();
            var selection = HostScreens.Select(HostScreens.ListDisplays(), request.Display, request.Rectangle);
            byte[] pixels = await HostScreens.CaptureAsync(selection, cancellation).ConfigureAwait(false);
            return PrepareScreenshotResult(call, selection, pixels, cancellation);
        }
        catch (Exception error) when (error is InvalidDataException or NotSupportedException)
        {
            return new ToolResultPart(call.CallID, [new TextPart(error.Message)], isError: true);
        }
        catch (OperationCanceledException) when (!cancellation.IsCancellationRequested)
        {
            return new ToolResultPart(call.CallID, [new TextPart("Host screenshot deadline expired; capture cleanup has joined.")], isError: true);
        }
    }

    internal partial ToolResultPart PrepareScreenshotResult(ToolCallPart call, ScreenSelection selection,
        ReadOnlyMemory<byte> bytes, CancellationToken cancellation)
    {
        cancellation.ThrowIfCancellationRequested();
        var agent = Parent();
        if (!agent.ToolEnabled("screenshot")) throw new UnauthorizedAccessException("Host-screen authority was revoked.");
        var probe = new BlobPart("image/png", new BlobRef(RefKind.Path, "host-screen-capability-probe"));
        agent.Engine().ValidateMedia([new ToolResultPart(call.CallID, [probe])]);
        var prepared = agent.PrepareImage(bytes);
        var bounds = selection.Bounds;
        // A successful helper exit is not proof it captured the requested region.
        // Refuse unexpectedly sized output rather than label it as that region.
        if (bounds.Width <= 0 || bounds.Height <= 0
            || prepared.OriginalWidth % bounds.Width != 0 || prepared.OriginalHeight % bounds.Height != 0)
            throw new InvalidDataException("Native capture dimensions do not match the requested region.");
        int scale = prepared.OriginalWidth / bounds.Width;
        if (scale < 1 || scale > selection.ScaleCeiling || prepared.OriginalHeight / bounds.Height != scale)
            throw new InvalidDataException("Native capture scale does not match the selected display.");
        cancellation.ThrowIfCancellationRequested();
        if (!agent.ToolEnabled("screenshot")) throw new UnauthorizedAccessException("Host-screen authority was revoked.");
        string report = $"Captured host rectangle ({bounds.X}, {bounds.Y}, {bounds.Width}, {bounds.Height}). "
            + $"Prepared {prepared.OriginalWidth} × {prepared.OriginalHeight} to {prepared.Width} × {prepared.Height}; "
            + $"{prepared.OutputBytes} PNG bytes. Actual image pixels are attached; temporary capture removed.";
        return new ToolResultPart(call.CallID, [new TextPart(report), prepared.Image]);
    }
}
