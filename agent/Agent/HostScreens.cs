using System.Diagnostics;
using System.Runtime.ExceptionServices;

namespace Ensemble;

internal static partial class HostScreens
{
    internal static partial IReadOnlyList<HostDisplay> ListDisplays()
    {
        if (!OperatingSystem.IsMacOS())
            throw new PlatformNotSupportedException("Host screenshots currently support macOS only.");
        // Fixed capacity bounds allocation and refuses a changing/oversized inventory.
        var ids = new uint[64];
        if (CGGetActiveDisplayList((uint)ids.Length, ids, out uint count) != 0 || count == 0 || count >= ids.Length)
            throw new IOException("Cannot obtain a bounded active-display inventory.");
        uint primary = CGMainDisplayID();
        var ordered = ids.Take((int)count).OrderBy(id => id == primary ? 0 : 1).ToArray();
        var result = new HostDisplay[ordered.Length];
        static int Coordinate(double value)
        {
            if (!double.IsFinite(value) || value != Math.Truncate(value) || value < int.MinValue || value > int.MaxValue)
                throw new IOException("Display coordinates are not representable as integer host units.");
            return (int)value;
        }
        for (int index = 0; index < ordered.Length; index++)
        {
            uint id = ordered[index];
            var bounds = CGDisplayBounds(id);
            // CGDisplayPixelsWide/High can report logical dimensions on HiDPI
            // modes. The copied mode supplies the backing raster used by capture.
            nint mode = CGDisplayCopyDisplayMode(id);
            if (mode == 0) throw new IOException("Cannot obtain the active display mode.");
            try
            {
                result[index] = new HostDisplay(index, id,
                    new ScreenRectangle(Coordinate(bounds.X), Coordinate(bounds.Y),
                        Coordinate(bounds.Width), Coordinate(bounds.Height)),
                    checked((int)CGDisplayModeGetPixelWidth(mode)),
                    checked((int)CGDisplayModeGetPixelHeight(mode)));
            }
            finally
            {
                CGDisplayModeRelease(mode);
            }
        }
        return Array.AsReadOnly(result);
    }

    internal static partial ScreenSelection Select(IReadOnlyList<HostDisplay> displays,
        int? displayIndex, ScreenRectangle? rectangle)
    {
        if (displays.Count == 0 || displays.Count > 64)
            throw new ArgumentException("No bounded display inventory is available.");
        if (displayIndex.HasValue && rectangle.HasValue)
            throw new ArgumentException("Choose a display or a rectangle, not both.");
        HostDisplay selected;
        ScreenRectangle area;
        if (rectangle is { } region)
        {
            area = region;
            // Do not expand the requested rectangle or capture another display to
            // fill an off-screen gap. Cross-display compositing is not supported.
            var matches = displays.Where(display => Contains(display.Bounds, area)).ToArray();
            if (matches.Length != 1)
                throw new ArgumentException("The rectangle must fit exactly one active display.");
            selected = matches[0];
        }
        else
        {
            int index = displayIndex ?? 0;
            if (index < 0 || index >= displays.Count || displays[index].Index != index)
                throw new ArgumentOutOfRangeException(nameof(displayIndex));
            selected = displays[index];
            area = selected.Bounds;
        }
        if (selected.Bounds.Width <= 0 || selected.Bounds.Height <= 0 || selected.PixelWidth <= 0 || selected.PixelHeight <= 0)
            throw new ArgumentException("Display dimensions must be positive.");
        double scale = Math.Max(1, Math.Max((double)selected.PixelWidth / selected.Bounds.Width,
            (double)selected.PixelHeight / selected.Bounds.Height));
        if (!double.IsFinite(scale) || scale > 16)
            throw new ArgumentException("Display scaling is unsupported.");
        var selection = new ScreenSelection(area, (int)Math.Ceiling(scale));
        ValidateSelection(selection);
        return selection;
    }

    private static bool Contains(ScreenRectangle outer, ScreenRectangle inner) =>
        inner.Width > 0 && inner.Height > 0 && inner.X >= outer.X && inner.Y >= outer.Y
        && (long)inner.X + inner.Width <= (long)outer.X + outer.Width
        && (long)inner.Y + inner.Height <= (long)outer.Y + outer.Height;

    private static partial void ValidateSelection(ScreenSelection selection)
    {
        var area = selection.Bounds;
        if (area.Width <= 0 || area.Height <= 0 || area.Width > 32768 || area.Height > 32768
            || selection.ScaleCeiling < 1 || selection.ScaleCeiling > 16
            || (long)area.X + area.Width > int.MaxValue || (long)area.Y + area.Height > int.MaxValue
            || (long)area.Width * area.Height * selection.ScaleCeiling * selection.ScaleCeiling > Agent.MaximumImagePixels)
            throw new ArgumentOutOfRangeException(nameof(selection), "Capture exceeds the bounded source-pixel limit.");
    }

    internal static partial Task<byte[]> CaptureAsync(ScreenSelection selection, CancellationToken cancellation)
    {
        ValidateSelection(selection);
        cancellation.ThrowIfCancellationRequested();
        if (!OperatingSystem.IsMacOS())
            throw new PlatformNotSupportedException("Host screenshots currently support macOS only.");
        // Never request permissions or elevate privileges on the model's behalf.
        if (!CGPreflightScreenCaptureAccess())
            throw new UnauthorizedAccessException("macOS Screen Recording permission is required for the server process.");
        return CaptureFileAsync(selection, "/usr/sbin/screencapture", TimeSpan.FromSeconds(15), cancellation);
    }

    internal static partial async Task<byte[]> CaptureFileAsync(ScreenSelection selection,
        string executable, TimeSpan deadline, CancellationToken cancellation)
    {
        ValidateSelection(selection);
        if (deadline <= TimeSpan.Zero || deadline > TimeSpan.FromSeconds(15))
            throw new ArgumentOutOfRangeException(nameof(deadline));
        cancellation.ThrowIfCancellationRequested();
        var directory = Directory.CreateTempSubdirectory("ensemble-screen-");
        Exception? failure = null;
        byte[] result = [];
        try
        {
            if (!OperatingSystem.IsWindows())
                File.SetUnixFileMode(directory.FullName, UnixFileMode.UserRead | UnixFileMode.UserWrite | UnixFileMode.UserExecute);
            string output = Path.Combine(directory.FullName, "capture.png");
            var area = selection.Bounds;
            var start = new ProcessStartInfo(executable)
            {
                UseShellExecute = false, RedirectStandardOutput = true, RedirectStandardError = true,
                WorkingDirectory = directory.FullName
            };
            // The native helper receives no provider keys, inherited shell or
            // user-selected command. Its only output location is owned scratch.
            start.Environment.Clear();
            start.ArgumentList.Add("-x");
            start.ArgumentList.Add("-t");
            start.ArgumentList.Add("png");
            start.ArgumentList.Add("-R");
            start.ArgumentList.Add(FormattableString.Invariant($"{area.X},{area.Y},{area.Width},{area.Height}"));
            start.ArgumentList.Add(output);
            using var timeout = new CancellationTokenSource(deadline);
            using var linked = CancellationTokenSource.CreateLinkedTokenSource(cancellation, timeout.Token);
            using var process = new Process { StartInfo = start };
            linked.Token.ThrowIfCancellationRequested();
            if (!process.Start()) throw new IOException("Cannot start the host capture helper.");
            Task stdout = process.StandardOutput.BaseStream.CopyToAsync(Stream.Null);
            Task stderr = process.StandardError.BaseStream.CopyToAsync(Stream.Null);
            Exception? waitFailure = null;
            try
            {
                await process.WaitForExitAsync(linked.Token).ConfigureAwait(false);
            }
            catch (Exception error)
            {
                waitFailure = error;
            }
            // Cancellation invalidates delivery. Even a failed kill must not
            // skip the join or hide the original wait error. No scratch is
            // removed while the helper or its pipe readers remain owned work.
            var cleanupErrors = new List<Exception>();
            try
            {
                if (!process.HasExited) process.Kill(entireProcessTree: true);
            }
            catch (Exception error) { cleanupErrors.Add(error); }
            try { await process.WaitForExitAsync().ConfigureAwait(false); }
            catch (Exception error) { cleanupErrors.Add(error); }
            try { await Task.WhenAll(stdout, stderr).ConfigureAwait(false); }
            catch (Exception error) { cleanupErrors.Add(error); }
            if (cleanupErrors.Count != 0)
            {
                if (waitFailure is not null) cleanupErrors.Insert(0, waitFailure);
                throw new AggregateException(cleanupErrors);
            }
            if (waitFailure is not null) ExceptionDispatchInfo.Capture(waitFailure).Throw();
            linked.Token.ThrowIfCancellationRequested();
            if (process.ExitCode != 0)
                throw new IOException("Host capture failed. Check OS permission and the logged-in console session.");
            using var stream = new FileStream(output, FileMode.Open, FileAccess.Read, FileShare.Read);
            long length = stream.Length;
            if (length < 8 || length > Agent.MaximumImageSourceBytes)
                throw new InvalidDataException("Captured PNG exceeds the source byte limit or is empty.");
            result = new byte[(int)length];
            await stream.ReadExactlyAsync(result, linked.Token).ConfigureAwait(false);
            if (stream.ReadByte() != -1 || !result.AsSpan(0, 8).SequenceEqual(new byte[] { 137, 80, 78, 71, 13, 10, 26, 10 }))
                throw new InvalidDataException("Capture did not produce a bounded PNG.");
            linked.Token.ThrowIfCancellationRequested();
        }
        catch (Exception error)
        {
            failure = error;
        }
        try
        {
            directory.Delete(recursive: true);
        }
        catch (Exception cleanup)
        {
            if (failure is not null) throw new AggregateException(failure, cleanup);
            throw;
        }
        if (failure is not null) ExceptionDispatchInfo.Capture(failure).Throw();
        cancellation.ThrowIfCancellationRequested();
        return result;
    }
}
