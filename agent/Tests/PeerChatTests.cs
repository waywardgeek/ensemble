using System.Net;
using System.Net.Sockets;
using Ensemble;
using Application = Ensemble.Ensemble;

internal static partial class ActorTests
{
    public static async Task PeerChat(string directory)
    {
        using var deadline = new CancellationTokenSource(TimeSpan.FromSeconds(15));
        var listener = new TcpListener(IPAddress.Loopback, 0);
        listener.Start();
        string container = Path.Combine(directory, "peer-chat");
        var configuration = new Configuration(Vendor.Anthropic, "peer-launch-model",
            "http://127.0.0.1:" + ((IPEndPoint)listener.LocalEndpoint).Port, "local-only");
        var app = new Application(TextWriter.Null, dataDirectory: container);
        int requests = 0;
        var server = Task.Run(async () =>
        {
            // One response for each lifetime; no submission or correction tool
            // is needed for an Ensemble-owned peer to finish ordinary chat.
            for (int turn = 0; turn < 2; turn++)
            {
                using var socket = await listener.AcceptTcpClientAsync(deadline.Token);
                using var stream = socket.GetStream();
                string request = await ReadRequest(stream, deadline.Token);
                requests++;
                Check(!request.Contains("Framework correction"), "Peer acquired delegated correction instructions.");
                if (turn == 1)
                {
                    Check(request.Contains("peer answer 0") && request.Contains("peer first turn"),
                        "Restarted peer lost its own conversation.");
                }
                await Reply(stream, "peer answer " + turn, deadline.Token);
            }
        });
        try
        {
            var root = new Agent(app, configuration, name: "root");
            var peer = await app.CreatePeerAsync(root, "writer", cancellation: deadline.Token);
            string first = await peer.SubmitAsync([new TextPart("peer first turn")]).WaitAsync(deadline.Token);
            Check(first.Contains("peer answer 0") && peer.Submission() is null && peer.ChildFailure() is null,
                "Ordinary peer chat required submission or reported failure.");
            app.Destroy();
            app = new Application(TextWriter.Null, dataDirectory: container);
            root = new Agent(app, configuration, name: "root");
            peer = (await app.RestorePeersAsync(root, deadline.Token)).Single();
            string second = await peer.SubmitAsync([new TextPart("peer second turn")]).WaitAsync(deadline.Token);
            await server.WaitAsync(deadline.Token);
            Check(second.Contains("peer answer 1") && requests == 2 && peer.Submission() is null
                && peer.ChildFailure() is null, "Restored peer did not complete a single ordinary model turn.");
        }
        finally
        {
            deadline.Cancel();
            listener.Stop();
            app.Destroy();
            try
            {
                await server;
            }
            catch (Exception error) when (error is OperationCanceledException or IOException or SocketException)
            {
                // Only expected fixture shutdown errors are suppressed.
            }
        }
    }
}
