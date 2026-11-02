using System.Runtime.InteropServices;

namespace Ensemble;

// Host coordinates are logical display units, with the primary display's upper
// left at (0,0). Backing pixels may be larger on a scaled display.
internal readonly record struct ScreenRectangle(int X, int Y, int Width, int Height);
internal readonly record struct HostDisplay(int Index, uint NativeId,
    ScreenRectangle Bounds, int PixelWidth, int PixelHeight);
internal readonly record struct ScreenSelection(ScreenRectangle Bounds, int ScaleCeiling);

// Internal backend only. Model authority and image placement checks belong at
// tool dispatch; no caller-supplied executable or output path is advertised.
internal static partial class HostScreens
{
    internal static partial IReadOnlyList<HostDisplay> ListDisplays();
    internal static partial ScreenSelection Select(IReadOnlyList<HostDisplay> displays,
        int? displayIndex, ScreenRectangle? rectangle);
    internal static partial Task<byte[]> CaptureAsync(ScreenSelection selection,
        CancellationToken cancellation);
    internal static partial Task<byte[]> CaptureFileAsync(ScreenSelection selection,
        string executable, TimeSpan deadline, CancellationToken cancellation);
    private static partial void ValidateSelection(ScreenSelection selection);

    private const string CoreGraphics = "/System/Library/Frameworks/CoreGraphics.framework/CoreGraphics";
    [StructLayout(LayoutKind.Sequential)]
    private struct NativeRectangle
    {
        internal double X, Y, Width, Height;
    }
    [DllImport(CoreGraphics)]
    private static extern int CGGetActiveDisplayList(uint maximum, [Out] uint[]? ids, out uint count);
    [DllImport(CoreGraphics)]
    private static extern uint CGMainDisplayID();
    [DllImport(CoreGraphics)]
    private static extern NativeRectangle CGDisplayBounds(uint display);
    [DllImport(CoreGraphics)]
    private static extern nuint CGDisplayPixelsWide(uint display);
    [DllImport(CoreGraphics)]
    private static extern nuint CGDisplayPixelsHigh(uint display);
    [DllImport(CoreGraphics)]
    [return: MarshalAs(UnmanagedType.I1)]
    private static extern bool CGPreflightScreenCaptureAccess();
}

public sealed partial class Tool
{
    private partial Task<ToolResultPart> HostScreenAsync(ToolCallPart call, CancellationToken cancellation);
    internal static partial (int? Display, ScreenRectangle? Rectangle) ParseScreenRequest(System.Text.Json.JsonElement args);
    internal partial ToolResultPart PrepareScreenshotResult(ToolCallPart call, ScreenSelection selection,
        ReadOnlyMemory<byte> bytes, CancellationToken cancellation);
}
