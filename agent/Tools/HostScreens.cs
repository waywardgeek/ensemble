using System.Text.Json;

namespace Ensemble;

public sealed partial class Tool
{
    internal static partial int ParseScreenRequest(JsonElement args)
    {
        if (args.ValueKind != JsonValueKind.Object
            || args.EnumerateObject().Select(property => property.Name).Distinct(StringComparer.Ordinal).Count() != args.EnumerateObject().Count())
            throw new ArgumentException("Screen request must be an object without duplicate fields.");
        // Refuse stale region requests rather than silently expanding them to
        // a full display. Observation controls remain separate from selection.
        foreach (var property in args.EnumerateObject())
        {
            if (property.Name is not ("display" or "ai_callback_delay" or "ai_callback_pattern" or "max_output_bytes"))
                throw new ArgumentException("Screenshot accepts only an optional display index and observation controls; rectangle selection is not supported.");
        }
        if (!args.TryGetProperty("display", out var value))
            return 0;
        if (value.ValueKind != JsonValueKind.Number || !value.TryGetInt32(out int display) || display < 0)
            throw new ArgumentException("Display must be a nonnegative 32-bit integer, or omitted for display 0.");
        return display;
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
            int display = ParseScreenRequest(call.Args);
            // This validates placement/capability only; it never opens a file.
            // No enumeration, permission check or capture precedes this guard.
            var probe = new BlobPart("image/png", new BlobRef(RefKind.Path, "host-screen-capability-probe"));
            agent.Engine().ValidateMedia([new ToolResultPart(call.CallID, [probe])]);
            cancellation.ThrowIfCancellationRequested();
            var selection = HostScreens.Select(HostScreens.ListDisplays(), display, null);
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
        string report = $"Captured host display bounds ({bounds.X}, {bounds.Y}, {bounds.Width}, {bounds.Height}). "
            + $"Prepared {prepared.OriginalWidth} × {prepared.OriginalHeight} to {prepared.Width} × {prepared.Height}; "
            + $"{prepared.OutputBytes} PNG bytes. Actual image pixels are attached; temporary capture removed.";
        return new ToolResultPart(call.CallID, [new TextPart(report), prepared.Image]);
    }
}
