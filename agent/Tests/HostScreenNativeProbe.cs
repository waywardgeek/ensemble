using System.Runtime.InteropServices;
using System.Text.Json;
using Ensemble;

// Explicit opt-in only: never part of the automated suite, and never captures
// or requests permission. The capture mode is for an explicitly approved synthetic window.
internal static class HostScreenNativeProbe
{
    [DllImport("/System/Library/Frameworks/CoreGraphics.framework/CoreGraphics")]
    [return: MarshalAs(UnmanagedType.I1)]
    private static extern bool CGPreflightScreenCaptureAccess();

    internal static async Task<int> Capture(string boundsFile, string screenshotFile)
    {
        using var document = JsonDocument.Parse(File.ReadAllText(boundsFile));
        var r = document.RootElement;
        var bounds = new ScreenRectangle(r.GetProperty("x").GetInt32(), r.GetProperty("y").GetInt32(),
            r.GetProperty("width").GetInt32(), r.GetProperty("height").GetInt32());
        if (bounds.Width != 640 || bounds.Height != 360)
            throw new InvalidDataException("This native fixture only accepts its 640 by 360 synthetic window.");
        var selection = HostScreens.Select(HostScreens.ListDisplays(), null, bounds);
        using var deadline = new CancellationTokenSource(TimeSpan.FromSeconds(20));
        byte[] png = await HostScreens.CaptureAsync(selection, deadline.Token);
        using var bitmap = SkiaSharp.SKBitmap.Decode(png);
        if (bitmap is null || bitmap.Width != 640 || bitmap.Height != 360)
            throw new InvalidDataException("Native capture returned unexpected dimensions.");
        var left = bitmap.GetPixel(160, 180);
        var right = bitmap.GetPixel(480, 180);
        if (left.Red < 230 || left.Blue > 25 || left.Green > 25 || right.Blue < 230 || right.Red > 25 || right.Green > 25)
            throw new InvalidDataException("Native pixels do not match the agreed synthetic window.");
        await File.WriteAllBytesAsync(screenshotFile, png);
        Console.WriteLine(JsonSerializer.Serialize(new { captured = true, width = bitmap.Width,
            height = bitmap.Height, bytes = png.Length, synthetic_pixels_verified = true }));
        return 0;
    }

    internal static int Permission()
    {
        if (!OperatingSystem.IsMacOS()) return 2;
        bool granted = CGPreflightScreenCaptureAccess();
        var displays = HostScreens.ListDisplays();
        Console.WriteLine(JsonSerializer.Serialize(new { permission_granted = granted,
            displays = displays.Select(d => new { index = d.Index, x = d.Bounds.X,
                y = d.Bounds.Y, width = d.Bounds.Width, height = d.Bounds.Height,
                pixel_width = d.PixelWidth, pixel_height = d.PixelHeight }) }));
        return 0;
    }
}
