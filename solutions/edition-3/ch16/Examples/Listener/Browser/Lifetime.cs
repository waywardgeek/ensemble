using System.Diagnostics;
using System.Net;
using System.Net.Sockets;
using System.Text.Json;

namespace SpeechListener;

internal sealed partial class Session
{
    private partial Process Launch(string executable, IEnumerable<string> arguments, IReadOnlyDictionary<string, string> environment)
    {
        var start = new ProcessStartInfo(executable)
        {
            WorkingDirectory = directory, UseShellExecute = false,
            RedirectStandardOutput = true, RedirectStandardError = true
        };
        foreach (string argument in arguments) start.ArgumentList.Add(argument);
        start.Environment.Clear();
        // These are host runtime inputs, not provider or skill configuration.
        foreach (string name in new[] { "PATH", "HOME", "TMPDIR", "DOTNET_ROOT", "DOTNET_CLI_HOME" })
        {
            string? value = Environment.GetEnvironmentVariable(name);
            if (value is not null) start.Environment[name] = value;
        }
        start.Environment["DOTNET_NOLOGO"] = "1";
        start.Environment["DOTNET_CLI_TELEMETRY_OPTOUT"] = "1";
        foreach (var pair in environment) start.Environment[pair.Key] = pair.Value;
        return Process.Start(start) ?? throw new IOException("Could not start owned process");
    }

    public partial async Task StartAsync(string dotnet, string gui, string chrome,
        IReadOnlyDictionary<string, string> targetEnvironment, CancellationToken cancellation)
    {
        if (target is not null || browser is not null) throw new InvalidOperationException("Session already started");
        // Reserve an unused loopback port, never the human's 8088. Closing the
        // reservation leaves a bind race; a failed owned host is a startup error.
        int port;
        do
        {
            using var reservation = new TcpListener(IPAddress.Loopback, 0);
            reservation.Start();
            port = ((IPEndPoint)reservation.LocalEndpoint).Port;
        } while (port == 8088);
        string origin = $"http://127.0.0.1:{port}";
        target = Launch(dotnet, [Path.GetFullPath(gui), "--port", port.ToString(), "--no-save", "--settings", Path.Combine(directory, "settings.json")], targetEnvironment);
        targetOutput = target.StandardOutput.ReadToEndAsync();
        targetError = target.StandardError.ReadToEndAsync();
        while (true)
        {
            cancellation.ThrowIfCancellationRequested();
            if (target.HasExited) throw new IOException("Owned GUI exited during startup");
            try
            {
                using var response = await http.GetAsync(origin, cancellation);
                if (response.IsSuccessStatusCode) break;
            }
            catch (HttpRequestException) { }
            await Task.Delay(50, cancellation);
        }
        Console.Error.WriteLine("Owned GUI is serving; starting browser");
        string profile = Path.Combine(directory, "chrome-profile");
        if (Directory.Exists(profile)) throw new IOException("Use a fresh listener directory, not an existing browser profile");
        browser = Launch(chrome, ["--headless=new", "--no-first-run", "--no-default-browser-check", "--disable-background-networking",
            "--disable-extensions", "--disable-component-extensions-with-background-pages",
            "--use-mock-keychain", "--password-store=basic", "--mute-audio", "--enable-automation",
            "--remote-debugging-port=0", "--user-data-dir=" + profile, "about:blank"], new Dictionary<string, string>());
        browserOutput = browser.StandardOutput.ReadToEndAsync();
        browserError = browser.StandardError.ReadToEndAsync();
        string discovery = Path.Combine(profile, "DevToolsActivePort");
        string[] lines = [];
        while (lines.Length == 0)
        {
            cancellation.ThrowIfCancellationRequested();
            if (browser.HasExited) throw new IOException("Owned browser exited during startup");
            if (File.Exists(discovery)) lines = await File.ReadAllLinesAsync(discovery, cancellation);
            if (lines.Length == 0) await Task.Delay(50, cancellation);
        }
        Console.Error.WriteLine("Owned browser debugging endpoint is ready");
        int debugPort = int.Parse(lines[0], System.Globalization.CultureInfo.InvariantCulture);
        string debugger = $"http://127.0.0.1:{debugPort}";
        using var creation = await http.PutAsync(debugger + "/json/new?about:blank", null, cancellation);
        creation.EnsureSuccessStatusCode();
        using var page = JsonDocument.Parse(await creation.Content.ReadAsStringAsync(cancellation));
        await socket.ConnectAsync(new Uri(page.RootElement.GetProperty("webSocketDebuggerUrl").GetString()!), cancellation);
        var navigation = await CommandAsync("Page.navigate", new { url = origin + "?speech=transcript" }, cancellation);
        if (navigation.TryGetProperty("errorText", out _)) throw new IOException("Owned page navigation failed");
        Console.Error.WriteLine("Owned page navigation submitted; waiting for GUI connection");
        while (!(await EvaluateAsync("typeof settings !== 'undefined' && settings !== null && typeof socket !== 'undefined' && socket.readyState===1 && typeof agentID !== 'undefined' && agentID!==null", cancellation)).GetBoolean())
        {
            await Task.Delay(50, cancellation);
        }
        Console.Error.WriteLine("Owned GUI page connected");
        // This private observer retains only completion, never frame.text or a
        // DOM snapshot. The model can hear solely through _ttsLog admission.
        const string observeChannel = """
            (() => {
                window._listenerPending = false;
                socket.addEventListener('message', event => {
                    const frame = JSON.parse(event.data);
                    if (frame.type === 'turn_ended') window._listenerPending = false;
                });
                patchSettings({ tts_enabled: true });
                return true;
            })()
            """;
        await EvaluateAsync(observeChannel, cancellation);
        while (!(await EvaluateAsync("settings.tts_enabled", cancellation)).GetBoolean())
        {
            await Task.Delay(50, cancellation);
        }
    }

    public partial Task DestroyAsync()
    {
        lock (destructionGate)
        {
            destruction ??= DestroyCoreAsync();
            return destruction;
        }
    }

    private partial async Task DestroyCoreAsync()
    {
        try
        {
            // Cancel/join custom Jobs (including pending sleeps) before disposing
            // their channel. Shutdown, not caller SendAsync cancellation, owns it.
            await Listener().ShutdownAsync();
        }
        finally
        {
            try
            {
                application.Destroy();
            }
            finally
            {
                socket.Dispose();
                http.Dispose();
                try
                {
                    await StopAsync(browser);
                }
                finally
                {
                    // A browser cleanup failure must not strand the target.
                    await StopAsync(target);
                }
                foreach (var (name, task) in new[] { ("target.stdout", targetOutput), ("target.stderr", targetError), ("browser.stdout", browserOutput), ("browser.stderr", browserError) })
                {
                    if (task is not null)
                    {
                        await File.WriteAllTextAsync(Path.Combine(directory, name), await task);
                    }
                }
            }
        }
    }

    private static partial async Task StopAsync(Process? process)
    {
        if (process is null) return;
        try
        {
            if (!process.HasExited)
            {
                // Signal only our process. Chrome owns its helpers; on timeout
                // Kill(true) ends its remaining owned descendant tree as well.
                using var signal = Process.Start(new ProcessStartInfo("/bin/kill")
                {
                    ArgumentList = { "-TERM", process.Id.ToString() }, UseShellExecute = false
                });
                if (signal is not null)
                {
                    await signal.WaitForExitAsync();
                }
                using var grace = new CancellationTokenSource(TimeSpan.FromSeconds(5));
                try
                {
                    await process.WaitForExitAsync(grace.Token);
                }
                catch (OperationCanceledException)
                {
                    process.Kill(entireProcessTree: true);
                }
            }
            await process.WaitForExitAsync();
        }
        finally
        {
            process.Dispose();
        }
    }
}
