using System.Buffers.Binary;
using System.Net;
using System.Text;
using Ensemble;
using EnsembleGui;

if (args is ["--browser-host"])
{
    await VoiceBrowserFixture.RunAsync();
    return;
}

string workspace = Directory.CreateTempSubdirectory("ensemble-voice-lease-").FullName;
Directory.SetCurrentDirectory(workspace);
var root = new Ensemble.Ensemble(TextWriter.Null);
var handler = new Handler();
var client = new VoiceTranscriber(root, "synthetic-lease-key", handler, TimeSpan.FromSeconds(5));
var target = new Agent(root, new Configuration(Vendor.Anthropic, "offline", apiKey: "", baseUrl: "http://127.0.0.1:1"), outputDirectory: "first-jobs", name: "first");
var other = new Agent(root, new Configuration(Vendor.Anthropic, "offline", apiKey: "", baseUrl: "http://127.0.0.1:1"), outputDirectory: "other-jobs", name: "other");
var audio = PcmVoiceAudio.FromWave(Wave());
using var deadline = new CancellationTokenSource(TimeSpan.FromSeconds(20));
try
{
    await target.PostAsync(new AgentInput(InputKind.Pause)).WaitAsync(deadline.Token);
    await other.PostAsync(new AgentInput(InputKind.Pause)).WaitAsync(deadline.Token);
    await using (var lease = new VoiceLease(target, 1, deadline.Token))
    {
        Refuses(() => lease.CommitAsync(target, 1), "unready commit");
        await lease.Start(audio).WaitAsync(deadline.Token);
        Refuses(() => lease.Start(audio), "duplicate audio");
        Refuses(() => lease.CommitAsync(other, 1), "wrong Agent");
        Refuses(() => lease.CommitAsync(target, 2), "wrong sequence");
        await lease.CommitAsync(target, 1).WaitAsync(deadline.Token);
        Refuses(() => lease.CommitAsync(target, 1), "duplicate commit");
        Check(ReferenceEquals(lease.Target, target), "Target was reconstructed");
    }
    Console.WriteLine("PASS real paused-Agent commit, readiness, identity and send-once refusal");
    await using (var lease = new VoiceLease(target, 2, deadline.Token))
    {
        await lease.Start(audio).WaitAsync(deadline.Token);
        lease.Cancel();
        Refuses(() => lease.CommitAsync(target, 2), "cancelled ready transcript");
    }
    Console.WriteLine("PASS cancellation after ready prevents admission");
    using (var disconnected = new CancellationTokenSource())
    {
        await using var lease = new VoiceLease(target, 3, disconnected.Token);
        int before = handler.Calls;
        disconnected.Cancel();
        Refuses(() => lease.Start(audio), "disconnected upload");
        Check(handler.Calls == before, "Disconnected upload reached provider");
    }
    Console.WriteLine("PASS connection cancellation prevents paid work");
    var entered = new TaskCompletionSource(TaskCreationOptions.RunContinuationsAsynchronously);
    var release = new TaskCompletionSource(TaskCreationOptions.RunContinuationsAsynchronously);
    handler.Respond = async _ => { entered.TrySetResult(); await release.Task; return Response(); };
    var held = new VoiceLease(target, 4, deadline.Token);
    Task<VoiceTranscript> request = held.Start(audio);
    await entered.Task.WaitAsync(deadline.Token);
    Task destruction = held.DisposeAsync().AsTask();
    try
    {
        Check(!destruction.IsCompleted, "Disposal abandoned live request");
        Refuses(() => held.CommitAsync(target, 4), "disposing commit");
    }
    finally { release.TrySetResult(); }
    try { await request.WaitAsync(deadline.Token); throw new Exception("Late transcript accepted"); }
    catch (OperationCanceledException) { }
    await destruction.WaitAsync(deadline.Token);
    await held.DisposeAsync();
    Refuses(() => held.Start(audio), "disposed upload");
    Console.WriteLine("PASS uncooperative request joined; late result inert; repeated disposal");
    handler.Respond = _ => Task.FromResult(new HttpResponseMessage(HttpStatusCode.Unauthorized));
    await using (var failed = new VoiceLease(target, 5, deadline.Token))
    {
        int before = handler.Calls;
        try { await failed.Start(audio).WaitAsync(deadline.Token); throw new Exception("Provider refusal accepted"); }
        catch (HttpRequestException) { }
        Refuses(() => failed.CommitAsync(target, 5), "failed transcript");
        Check(handler.Calls == before + 1, "Provider request retried");
    }
    Console.WriteLine("PASS failed transcription cannot commit or retry");
}
finally { root.Destroy(); }
Check(handler.Disposed, "Root failed to dispose transcriber transport");
Console.WriteLine("PASS root transport cleanup; no live providers or microphone");
await VoiceSocketTests.RunAsync(audio);

static void Check(bool condition, string message) { if (!condition) throw new Exception(message); }
static void Refuses(Func<Task> operation, string name)
{
    try { _ = operation(); }
    catch (InvalidOperationException) { return; }
    throw new Exception("Accepted " + name);
}
static HttpResponseMessage Response() => new(HttpStatusCode.OK)
{ Content = new StringContent("{\"text\":\"synthetic lease transcript\"}", Encoding.UTF8, "application/json") };
static byte[] Wave()
{
    byte[] bytes = new byte[46];
    "RIFF"u8.CopyTo(bytes); BinaryPrimitives.WriteUInt32LittleEndian(bytes.AsSpan(4), 38);
    "WAVEfmt "u8.CopyTo(bytes.AsSpan(8)); BinaryPrimitives.WriteUInt32LittleEndian(bytes.AsSpan(16), 16);
    BinaryPrimitives.WriteUInt16LittleEndian(bytes.AsSpan(20), 1); BinaryPrimitives.WriteUInt16LittleEndian(bytes.AsSpan(22), 1);
    BinaryPrimitives.WriteUInt32LittleEndian(bytes.AsSpan(24), 16000); BinaryPrimitives.WriteUInt32LittleEndian(bytes.AsSpan(28), 32000);
    BinaryPrimitives.WriteUInt16LittleEndian(bytes.AsSpan(32), 2); BinaryPrimitives.WriteUInt16LittleEndian(bytes.AsSpan(34), 16);
    "data"u8.CopyTo(bytes.AsSpan(36)); BinaryPrimitives.WriteUInt32LittleEndian(bytes.AsSpan(40), 2);
    return bytes;
}
sealed class Handler : HttpMessageHandler
{
    internal int Calls;
    internal bool Disposed;
    internal Func<CancellationToken, Task<HttpResponseMessage>> Respond = _ => Task.FromResult(
        new HttpResponseMessage(HttpStatusCode.OK) { Content = new StringContent("{\"text\":\"synthetic lease transcript\"}") });
    protected override Task<HttpResponseMessage> SendAsync(HttpRequestMessage request, CancellationToken cancellation)
    { Calls++; return Respond(cancellation); }
    protected override void Dispose(bool disposing) { Disposed = true; base.Dispose(disposing); }
}
