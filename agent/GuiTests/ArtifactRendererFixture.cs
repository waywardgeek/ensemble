using System.Net;
using System.Net.Sockets;
using Ensemble;
using EnsembleGui;

// No provider calls: browser tests inject recorded facts into the real handler.
internal static class ArtifactRendererFixture
{
    internal static async Task RunAsync(string directory)
    {
        Directory.CreateDirectory(directory);
        using var reservation = new TcpListener(IPAddress.Loopback, 0);
        reservation.Start();
        int port = ((IPEndPoint)reservation.LocalEndpoint).Port;
        reservation.Stop();
        using var stop = new CancellationTokenSource(TimeSpan.FromMinutes(3));
        var root = new Ensemble.Ensemble(TextWriter.Null);
        Task? hosting = null;
        try
        {
            var agent = new Agent(root,
                new Configuration(Vendor.Anthropic, "offline", apiKey: "", baseUrl: "http://127.0.0.1:1"),
                outputDirectory: Path.Combine(directory, "jobs"));
            hosting = Gui.RunAsync(agent, new GuiOptions(port,
                SettingsPath: Path.Combine(directory, "settings.json"),
                TracePath: Path.Combine(directory, "gui.log"),
                RendererDirectory: Path.Combine(directory, "renderers")), stop.Token);
            string url = $"http://127.0.0.1:{port}/";
            using var client = new HttpClient();
            while (true)
            {
                if (hosting.IsCompleted) { await hosting; throw new InvalidOperationException("Host exited before readiness"); }
                try
                {
                    using var response = await client.GetAsync(url, stop.Token);
                    if (response.IsSuccessStatusCode) break;
                }
                catch (HttpRequestException) { }
                await Task.Delay(25, stop.Token);
            }
            await File.WriteAllTextAsync(Path.Combine(directory, "url"), url, stop.Token);
            while (!File.Exists(Path.Combine(directory, "stop")))
            {
                if (hosting.IsCompleted) { await hosting; throw new InvalidOperationException("Host exited before stop"); }
                await Task.Delay(25, stop.Token);
            }
        }
        finally
        {
            stop.Cancel();
            try { if (hosting is not null) await hosting; }
            catch (OperationCanceledException) when (stop.IsCancellationRequested) { }
            finally { root.Destroy(); }
        }
        Console.WriteLine("STOPPED artifact renderer fixture and owned Agent");
    }
}
