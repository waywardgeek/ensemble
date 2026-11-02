using Ensemble;
using SkiaSharp;
using Application = Ensemble.Ensemble;

internal static partial class HostScreenTests
{
    private static void Check(bool condition, string message)
    {
        if (!condition) throw new Exception(message);
    }
    private static void Refuses(Action action)
    {
        try { action(); }
        catch (ArgumentException) { return; }
        throw new Exception("Invalid screen selection was accepted.");
    }
    public static void Selection()
    {
        HostDisplay[] displays = [
            new(0, 10, new(0, 0, 1920, 1080), 3840, 2160),
            new(1, 11, new(-1280, -200, 1280, 1024), 1280, 1024)];
        Check(HostScreens.Select(displays, null, null).Bounds == displays[0].Bounds, "Default primary display changed.");
        Check(HostScreens.Select(displays, 1, null).Bounds == displays[1].Bounds, "Secondary display lost negative origin.");
        var selected = HostScreens.Select(displays, null, new(-1200, -100, 64, 48));
        Check(selected.Bounds.X == -1200 && selected.ScaleCeiling == 1, "Rectangle was expanded or transformed.");
        Check(HostScreens.Select(displays, 0, null).ScaleCeiling == 2, "Backing-pixel bound ignored scaling.");
        Refuses(() => HostScreens.Select(displays, 0, new(0, 0, 10, 10)));
        Refuses(() => HostScreens.Select(displays, -1, null));
        Refuses(() => HostScreens.Select(displays, 2, null));
        foreach (var rectangle in new ScreenRectangle[] {
            new(0, 0, 0, 1), new(0, 0, 1, -1), new(-1, 0, 2, 2),
            new(1900, 1000, 21, 81), new(int.MaxValue, 0, 10, 10) })
            Refuses(() => HostScreens.Select(displays, null, rectangle));
        Refuses(() => HostScreens.Select([displays[0], displays[0]], null, new(0, 0, 1, 1)));
        Refuses(() => HostScreens.Select([new(0, 1, new(0, 0, 4000, 4000), 8000, 8000)], 0, null));
        var bounded = HostScreens.Select([new(0, 1, new(0, 0, 4000, 4000), 8000, 8000)], null, new(0, 0, 100, 100));
        Check(bounded.ScaleCeiling == 2, "Small region on large display was refused.");
    }

    public static async Task ProcessLifecycle(string temp)
    {
        if (OperatingSystem.IsWindows()) throw new PlatformNotSupportedException("This fake-process fixture uses POSIX executables.");
        string root = Path.Combine(temp, "host-screen-process");
        Directory.CreateDirectory(root);
        string fixture = Path.Combine(root, "synthetic.png");
        using (var bitmap = new SKBitmap(2, 2))
        {
            bitmap.Erase(SKColors.Red);
            using var image = SKImage.FromBitmap(bitmap);
            using var png = image.Encode(SKEncodedImageFormat.Png, 100);
            File.WriteAllBytes(fixture, png.ToArray());
        }
        string receipt = Path.Combine(root, "receipt");
        string executable = Path.Combine(root, "fake-capture");
        static string Quote(string value) => "'" + value.Replace("'", "'\\''") + "'";
        void Script(string body)
        {
            if (File.Exists(receipt)) File.Delete(receipt);
            File.WriteAllText(executable, "#!/bin/sh\nset -eu\nprintf '%s\\n' \"$PWD\" \"$5\" > " + Quote(receipt)
                + "\n[ -z \"${ENSEMBLE_SCREEN_TEST_SECRET+x}\" ] || exit 91\n" + body + "\n");
            File.SetUnixFileMode(executable, UnixFileMode.UserRead | UnixFileMode.UserWrite | UnixFileMode.UserExecute);
        }
        void Cleaned()
        {
            var lines = File.ReadAllLines(receipt);
            Check(lines[1] == "-100,20,20,30", "Native rectangle arguments changed.");
            Check(!Directory.Exists(lines[0]), "Capture scratch survived completion or failure.");
        }
        var selection = new ScreenSelection(new(-100, 20, 20, 30), 1);
        string? old = Environment.GetEnvironmentVariable("ENSEMBLE_SCREEN_TEST_SECRET");
        Environment.SetEnvironmentVariable("ENSEMBLE_SCREEN_TEST_SECRET", "synthetic-not-a-key");
        try
        {
            Script("/bin/cp " + Quote(fixture) + " \"$6\"");
            byte[] bytes = await HostScreens.CaptureFileAsync(selection, executable, TimeSpan.FromSeconds(5), default);
            Check(bytes.SequenceEqual(File.ReadAllBytes(fixture)), "Native result lost pixels.");
            Cleaned();
            var app = new Application(TextWriter.Null);
            try
            {
                var agent = new Agent(app, new Configuration(Vendor.OpenAI, "gpt-6-astra", "http://127.0.0.1:1"), Path.Combine(root, "jobs"));
                var prepared = agent.PrepareImage(bytes);
                Check(prepared.Width == 2 && prepared.Height == 2, "Capture bytes did not enter image preparation.");
            }
            finally { app.Destroy(); }

            foreach (string body in new[] { "exit 7", "printf invalid! > \"$6\"",
                "/bin/dd if=/dev/zero of=\"$6\" bs=1 count=0 seek=33554433 2>/dev/null" })
            {
                Script(body);
                bool refused = false;
                try { await HostScreens.CaptureFileAsync(selection, executable, TimeSpan.FromSeconds(5), default); }
                catch (Exception error) when (error is IOException or InvalidDataException) { refused = true; }
                Check(refused, "Failed or oversized capture delivered bytes.");
                Cleaned();
            }
            foreach (bool deadline in new[] { false, true })
            {
                Script("/bin/sleep 30\n/bin/cp " + Quote(fixture) + " \"$6\"");
                using var cancel = new CancellationTokenSource();
                Task<byte[]> pending = HostScreens.CaptureFileAsync(selection, executable,
                    deadline ? TimeSpan.FromMilliseconds(300) : TimeSpan.FromSeconds(5), cancel.Token);
                if (!deadline)
                {
                    for (int attempt = 0; !File.Exists(receipt) && attempt < 100; attempt++) await Task.Delay(10);
                    Check(File.Exists(receipt), "Fake capture did not start.");
                    cancel.Cancel();
                }
                bool cancelled = false;
                try { await pending; }
                catch (OperationCanceledException) { cancelled = true; }
                Check(cancelled, "Cancelled capture delivered late bytes.");
                Cleaned();
            }
            Script("exit 99");
            using var alreadyCancelled = new CancellationTokenSource();
            alreadyCancelled.Cancel();
            try { await HostScreens.CaptureFileAsync(selection, executable, TimeSpan.FromSeconds(5), alreadyCancelled.Token); }
            catch (OperationCanceledException) { }
            Check(!File.Exists(receipt), "Pre-cancelled capture started a process.");
        }
        finally { Environment.SetEnvironmentVariable("ENSEMBLE_SCREEN_TEST_SECRET", old); }
    }
}
