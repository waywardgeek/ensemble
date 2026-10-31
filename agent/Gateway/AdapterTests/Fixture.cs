using System.Collections.Immutable;
using System.Net;
using System.Text.Json;
using System.Threading.Channels;
using Ensemble;
using Ensemble.Gateway;
using Ensemble.GatewayAdapter;
using Microsoft.AspNetCore.Hosting.Server;
using Microsoft.AspNetCore.Hosting.Server.Features;
using Application = Ensemble.Ensemble;

// A real HTTP provider boundary fake. Tests control complete responses, not
// internal Agent calls. Gateway, Noise, actor scheduling and History are real.
sealed class Pending(string body)
{
    public string Body { get; } = body;
    public TaskCompletionSource<(int Status, string Text)> Reply { get; } = new(TaskCreationOptions.RunContinuationsAsynchronously);
}
sealed class Inbox : IHandler
{
    public Channel<Envelope> Messages { get; } = Channel.CreateUnbounded<Envelope>();
    public void OnConnected(WelcomePayload welcome) { }
    public void OnDeliver(Envelope envelope) => Messages.Writer.TryWrite(envelope);
    public void OnDisconnected(string reason) => Messages.Writer.TryComplete();
    public void OnError(Exception error) { }
}
sealed class Fixture : IAsyncDisposable
{
    public readonly CancellationTokenSource Deadline = new(TimeSpan.FromSeconds(20));
    public readonly CancellationTokenSource Connection = new();
    public readonly Channel<Pending> Requests = Channel.CreateUnbounded<Pending>();
    public readonly StringWriter Diagnostics = new();
    public readonly Inbox Answers = new();
    public readonly Bridge Bridge = new();
    public readonly string DirectoryPath = Path.Combine(Path.GetTempPath(), "gateway-adapter-" + Guid.NewGuid().ToString("N"));
    public Application Application = null!;
    public Agent Agent = null!;
    public Router Router = null!;
    public Client Sender = null!;
    public Task Running = Task.CompletedTask;
    private WebApplication web = null!;

    public async Task Start(bool wrongPin = false)
    {
        Directory.CreateDirectory(DirectoryPath);
        var routerKey = Keys.GenerateKeypair();
        var subjectKey = Keys.GenerateKeypair();
        var senderKey = Keys.GenerateKeypair();
        var config = new RouterConfig(new RouterEndpoint("127.0.0.1:0", DirectoryPath),
            ImmutableDictionary<string, GroupDefinition>.Empty.Add("agents", new(["send-message", "declare-contactable"], "trusted")),
            ImmutableDictionary<string, IdentityDefinition>.Empty
                .Add("subject", new(Keys.EncodePublicKey(subjectKey.Public.AsSpan()), ["agents"]))
                .Add("sender", new(Keys.EncodePublicKey(senderKey.Public.AsSpan()), ["agents"])));
        Router = new Router(config, routerKey);
        var builder = WebApplication.CreateSlimBuilder();
        builder.Logging.ClearProviders();
        builder.WebHost.ConfigureKestrel(options => options.Listen(IPAddress.Loopback, 0));
        web = builder.Build();
        web.UseWebSockets();
        web.MapGet("/v1/ws", async context =>
        {
            string id = context.Request.Query["agent_id"].ToString();
            using var socket = await context.WebSockets.AcceptWebSocketAsync();
            try { await Router.AcceptAsync(id, socket, context.RequestAborted); }
            catch (Exception) when (socket.State != System.Net.WebSockets.WebSocketState.Open) { }
        });
        web.MapPost("/v1/messages", async context =>
        {
            using var reader = new StreamReader(context.Request.Body);
            var pending = new Pending(await reader.ReadToEndAsync(context.RequestAborted));
            await Requests.Writer.WriteAsync(pending, context.RequestAborted);
            var reply = await pending.Reply.Task.WaitAsync(context.RequestAborted);
            context.Response.StatusCode = reply.Status;
            await context.Response.WriteAsJsonAsync(new
            {
                id = "fake", type = "message", role = "assistant", model = "model-a",
                content = new[] { new { type = "text", text = reply.Text } },
                stop_reason = "end_turn", usage = new { input_tokens = 1, output_tokens = 1 }
            }, context.RequestAborted);
        });
        await web.StartAsync(Deadline.Token);
        string address = web.Services.GetRequiredService<IServer>().Features.Get<IServerAddressesFeature>()!.Addresses.Single();
        Application = new Application(Diagnostics);
        Agent = new Agent(Application, new Configuration(Vendor.Anthropic, "model-a", address,
            "fake-only", disableStreaming: true), name: "subject");
        string gateway = address.Replace("http:", "ws:") + "/v1/ws";
        Sender = await Client.ConnectAsync(new(gateway, "sender", senderKey, routerKey.Public), Answers, Deadline.Token);
        await Sender.DeclareContactableAsync(["agents"], Deadline.Token);
        Running = Bridge.RunAsync(Agent, new(gateway, "subject", subjectKey, wrongPin ? Keys.GenerateKeypair().Public : routerKey.Public), ["agents"], Connection.Token);
        await Bridge.Ready().WaitAsync(Deadline.Token);
    }

    public Task Send(string text, string conversation, string mode = Protocol.ResponseAsync) =>
        Sender.SendAsync(new("subject", text, conversation, mode), Deadline.Token);
    public async Task<Pending> Next() => await Requests.Reader.ReadAsync(Deadline.Token);
    public async Task<Envelope> Answer() => await Answers.Messages.Reader.ReadAsync(Deadline.Token);
    public ulong QueuedReplies()
    {
        Router.Save();
        using var json = JsonDocument.Parse(File.ReadAllText(Path.Combine(DirectoryPath, "queues", "sender.json")));
        return json.RootElement.GetProperty("next").GetUInt64();
    }
    public async Task Receipt(ulong expected)
    {
        // A stored receipt is an observed acceptance boundary, not elapsed time.
        while (true)
        {
            Deadline.Token.ThrowIfCancellationRequested();
            Router.Save();
            string path = Path.Combine(DirectoryPath, "queues", "subject.json");
            using var json = JsonDocument.Parse(await File.ReadAllTextAsync(path, Deadline.Token));
            if (json.RootElement.GetProperty("ack").GetUInt64() >= expected)
                return;
            await Task.Delay(10, Deadline.Token);
        }
    }
    public async ValueTask DisposeAsync()
    {
        Connection.Cancel();
        try { await Running.WaitAsync(Deadline.Token); }
        finally
        {
            if (Agent is not null) await Agent.ShutdownAsync();
            Application?.Destroy();
            if (Sender is not null) await Sender.DisposeAsync();
            if (Router is not null) await Router.DisposeAsync();
            if (web is not null) { await web.StopAsync(Deadline.Token); await web.DisposeAsync(); }
            Connection.Dispose();
            Deadline.Dispose();
            Diagnostics.Dispose();
        }
    }
}
