using System.Net;
using System.Text;
using Ensemble;

// External consumer: no friend-assembly access or real provider credentials.
static class VoiceTransportTests
{
    private sealed class Transport : HttpMessageHandler
    {
        public required Func<HttpRequestMessage, CancellationToken, Task<HttpResponseMessage>> Reply;
        public int Calls;
        public bool Disposed;
        protected override Task<HttpResponseMessage> SendAsync(HttpRequestMessage request, CancellationToken cancellation)
        {
            Calls++;
            return Reply(request, cancellation);
        }
        protected override void Dispose(bool disposing)
        {
            Disposed = true;
            base.Dispose(disposing);
        }
    }
    private static void Check(bool condition, string message)
    {
        if (!condition) throw new InvalidOperationException(message);
    }
    private static byte[] Wave()
    {
        using var stream = new MemoryStream();
        using var writer = new BinaryWriter(stream, Encoding.ASCII, leaveOpen: true);
        writer.Write(Encoding.ASCII.GetBytes("RIFF")); writer.Write(38);
        writer.Write(Encoding.ASCII.GetBytes("WAVEfmt ")); writer.Write(16);
        writer.Write((short)1); writer.Write((short)1); writer.Write(16000);
        writer.Write(32000); writer.Write((short)2); writer.Write((short)16);
        writer.Write(Encoding.ASCII.GetBytes("data")); writer.Write(2); writer.Write((short)42);
        return stream.ToArray();
    }
    public static async Task RunAsync()
    {
        var root = new Ensemble.Ensemble(TextWriter.Null);
        byte[] wave = Wave();
        var audio = PcmVoiceAudio.FromWave(wave);
        var transport = new Transport { Reply = async (request, cancellation) =>
        {
            Check(request.Method == HttpMethod.Post && request.RequestUri?.AbsoluteUri ==
                "https://api.openai.com/v1/audio/transcriptions", "Production route changed");
            Check(request.Headers.Authorization?.Parameter == "synthetic-test-key", "Wrong explicit payer");
            Check(request.Content is MultipartFormDataContent, "Expected multipart transcription request");
            var parts = ((MultipartFormDataContent)request.Content!).ToDictionary(
                part => part.Headers.ContentDisposition!.Name!.Trim('"'));
            Check(parts.Count == 4 && await parts["model"].ReadAsStringAsync(cancellation) == "gpt-transcribe", "Wrong model/fields");
            Check((await parts["file"].ReadAsByteArrayAsync(cancellation)).SequenceEqual(wave), "WAV bytes changed");
            return new HttpResponseMessage(HttpStatusCode.OK) { Content = new StringContent("{\"text\":\"Synthetic transcript\"}") };
        }};
        try
        {
            var transcriber = new VoiceTranscriber(root, "synthetic-test-key", transport);
            var result = await transcriber.TranscribeAsync(audio);
            Check(result.Text == "Synthetic transcript" && transport.Calls == 1, "Result or exactly-once request wrong");
            using var refused = new Transport { Reply = (_, _) => throw new InvalidOperationException("Must not send") };
            bool rejected = false;
            try { _ = new VoiceTranscriber(root, "synthetic-test-key", refused); }
            catch (InvalidOperationException) { rejected = true; }
            Check(rejected && root.Transcriber() == transcriber, "Duplicate owner accepted");
            Check(!refused.Disposed && refused.Calls == 0, "Refusal took caller's transport");
            try { _ = new VoiceTranscriber(root, "synthetic-test-key", (HttpMessageHandler)null!); throw new Exception("Null accepted"); }
            catch (ArgumentNullException) { }
            await transcriber.DestroyAsync();
            Check(transport.Disposed && root.Transcriber() is null, "Owned handler not disposed/detached");
        }
        finally { root.Destroy(); }
        Console.WriteLine("PASS public voice transport: fixed route, explicit synthetic payer, exact WAV, no retry, refused/owned lifetime");

        var entered = new TaskCompletionSource(TaskCreationOptions.RunContinuationsAsynchronously);
        var cancelled = new TaskCompletionSource(TaskCreationOptions.RunContinuationsAsynchronously);
        var release = new TaskCompletionSource(TaskCreationOptions.RunContinuationsAsynchronously);
        var blocking = new Transport { Reply = async (_, cancellation) =>
        {
            entered.SetResult();
            try { await Task.Delay(Timeout.Infinite, cancellation); }
            finally { cancelled.SetResult(); await release.Task; }
            throw new InvalidOperationException("Cancelled request completed normally");
        }};
        var second = new Ensemble.Ensemble(TextWriter.Null);
        Task? requestTask = null;
        Task? destruction = null;
        try
        {
            var transcriber = new VoiceTranscriber(second, "synthetic-test-key", blocking);
            requestTask = transcriber.TranscribeAsync(audio);
            await entered.Task.WaitAsync(TimeSpan.FromSeconds(5));
            destruction = transcriber.DestroyAsync();
            await cancelled.Task.WaitAsync(TimeSpan.FromSeconds(5));
            Check(!destruction.IsCompleted && !blocking.Disposed, "Destruction did not join pending cleanup");
            release.SetResult();
            await destruction.WaitAsync(TimeSpan.FromSeconds(5));
            try { await requestTask; throw new InvalidOperationException("Request was not cancelled"); }
            catch (OperationCanceledException) { }
            Check(blocking.Disposed && blocking.Calls == 1, "Handler leaked or request retried");
        }
        finally
        {
            release.TrySetResult();
            if (destruction is not null) await destruction;
            second.Destroy();
        }
        Console.WriteLine("PASS public voice transport: cancellation joins handler work before disposal");
    }
}
