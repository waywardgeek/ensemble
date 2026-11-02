using System.Net;
using System.Text;
using Ensemble;

internal static class VoiceTranscriptionTests
{
    private const string Key = "synthetic-voice-test-key";
    private static void Check(bool condition, string reason)
    {
        if (!condition) throw new Exception(reason);
    }

    private sealed class Handler : HttpMessageHandler
    {
        internal Func<HttpRequestMessage, CancellationToken, Task<HttpResponseMessage>> Reply =
            (_, _) => Task.FromResult(Json("{\"text\":\"hello\"}"));
        internal int Calls;
        internal bool Disposed;
        internal bool FailDisposal;
        protected override Task<HttpResponseMessage> SendAsync(HttpRequestMessage request, CancellationToken token)
        {
            Calls++;
            return Reply(request, token);
        }
        protected override void Dispose(bool disposing)
        {
            Disposed = true;
            base.Dispose(disposing);
            if (FailDisposal) throw new IOException("Synthetic transport disposal failure: " + Key);
        }
    }

    // StreamContent cannot infer a length from this forward-only source.
    private sealed class ResponseStream(byte[] bytes) : MemoryStream(bytes, writable: false)
    {
        internal int BytesRead;
        internal bool Disposed;
        public override bool CanSeek => false;
        public override ValueTask<int> ReadAsync(Memory<byte> buffer, CancellationToken token = default)
        {
            token.ThrowIfCancellationRequested();
            int count = base.Read(buffer.Span[..Math.Min(buffer.Length, 997)]);
            BytesRead += count;
            return ValueTask.FromResult(count);
        }
        protected override void Dispose(bool disposing)
        {
            Disposed = true;
            base.Dispose(disposing);
        }
    }

    private static HttpResponseMessage Json(string body, HttpStatusCode status = HttpStatusCode.OK) =>
        new(status) { Content = new StringContent(body, Encoding.UTF8, "application/json") };

    private static async Task<Exception> Failure(Func<Task> action)
    {
        try { await action(); }
        catch (Exception error) { return error; }
        throw new Exception("Expected refusal, but operation succeeded.");
    }

    public static async Task Contract()
    {
        var root = new Ensemble.Ensemble(TextWriter.Null);
        var handler = new Handler();
        var audio = PcmVoiceAudio.FromWave(VoiceAudioTests.Wave(160));
        handler.Reply = async (request, token) =>
        {
            Check(request.Method == HttpMethod.Post, "Wrong HTTP method.");
            Check(request.RequestUri!.AbsoluteUri == "https://api.openai.com/v1/audio/transcriptions", "Wrong endpoint.");
            Check(request.Headers.Authorization?.Scheme == "Bearer"
                && request.Headers.Authorization.Parameter == Key, "Wrong payer.");
            Check(request.Content is MultipartFormDataContent, "Missing multipart body.");
            var parts = ((MultipartFormDataContent)request.Content!).ToDictionary(
                part => part.Headers.ContentDisposition!.Name!.Trim('"'));
            Check(parts.Count == 4, "Unexpected multipart fields.");
            Check(await parts["model"].ReadAsStringAsync(token) == "gpt-transcribe", "Wrong model.");
            Check(await parts["response_format"].ReadAsStringAsync(token) == "json", "Wrong response format.");
            Check(await parts["stream"].ReadAsStringAsync(token) == "false", "Unexpected streaming.");
            Check(parts["file"].Headers.ContentDisposition!.FileName!.Trim('"') == "voice.wav", "Wrong filename.");
            Check(parts["file"].Headers.ContentType!.MediaType == "audio/wav", "Wrong MIME type.");
            byte[] sent = await parts["file"].ReadAsByteArrayAsync(token);
            Check(sent.AsSpan().SequenceEqual(audio.WaveBytes.AsSpan()), "Audio changed in transport.");
            return Json("{\"text\":\" hello \"}");
        };
        try
        {
            var client = new VoiceTranscriber(root, Key, handler, TimeSpan.FromSeconds(5));
            Check(ReferenceEquals(root.Transcriber(), client) && ReferenceEquals(client.Parent(), root), "Missing ownership links.");
            var transcript = await client.TranscribeAsync(audio);
            Check(transcript.Text == "hello" && transcript.Usage is null, "Text or unknown usage changed.");
            Check(transcript.InputDuration == audio.Duration && handler.Calls == 1, "Duration or request count incorrect.");
            var log = new StringWriter();
            root.Log(Key, log);
            Check(!log.ToString().Contains(Key), "Diagnostic key leaked.");
        }
        finally { root.Destroy(); }
        Check(handler.Disposed && root.Transcriber() is null, "Root did not dispose/detach client.");
    }

    public static async Task Responses()
    {
        var root = new Ensemble.Ensemble(TextWriter.Null);
        var handler = new Handler();
        var audio = PcmVoiceAudio.FromWave(VoiceAudioTests.Wave(1));
        var client = new VoiceTranscriber(root, Key, handler, TimeSpan.FromSeconds(5));
        try
        {
            string[] invalid = ["{}", "{\"text\":\" \"}", "{\"text\":\"a\",\"text\":\"b\"}",
                "{\"text\":\"" + Key + "\"}", "{\"text\":\"ok\",\"usage\":{\"type\":\"unknown\"}}",
                "{\"text\":\"ok\",\"usage\":{\"type\":\"tokens\",\"input_tokens\":1,\"output_tokens\":2,\"total_tokens\":4}}",
                "{\"text\":\"" + new string('x', 32769) + "\"}", new string(' ', 65537)];
            foreach (string body in invalid)
            {
                handler.Reply = (_, _) => Task.FromResult(Json(body));
                var error = await Failure(async () => await client.TranscribeAsync(audio));
                Check(!error.ToString().Contains(Key), "Refusal retained credential material.");
            }
            foreach (var status in new[] { HttpStatusCode.Unauthorized, HttpStatusCode.ServiceUnavailable })
            {
                int before = handler.Calls;
                handler.Reply = (_, _) => Task.FromResult(Json(Key, status));
                var error = await Failure(async () => await client.TranscribeAsync(audio));
                Check(handler.Calls == before + 1 && !error.ToString().Contains(Key), "HTTP error retried or leaked body.");
            }
            handler.Reply = (_, _) => Task.FromResult(Json("{\"text\":\"recovered\",\"usage\":{\"type\":\"duration\",\"seconds\":2}}"));
            var recovered = await client.TranscribeAsync(audio);
            Check(recovered.Usage is VoiceDurationUsage { Seconds: 2 } && recovered.InputDuration == audio.Duration,
                "Reported duration replaced physical duration or recovery failed.");
            handler.Reply = (_, _) => Task.FromResult(Json("{\"text\":\"ok\",\"usage\":{\"type\":\"tokens\",\"input_tokens\":3,\"output_tokens\":2,\"total_tokens\":5,\"input_token_details\":{\"audio_tokens\":2,\"text_tokens\":1}}}"));
            Check((await client.TranscribeAsync(audio)).Usage is VoiceTokenUsage { TotalTokens: 5, AudioInputTokens: 2 }, "Token usage lost.");
        }
        finally { root.Destroy(); }
    }

    public static async Task ResponseBounds()
    {
        var root = new Ensemble.Ensemble(TextWriter.Null);
        var handler = new Handler();
        var client = new VoiceTranscriber(root, Key, handler, TimeSpan.FromSeconds(5));
        var audio = PcmVoiceAudio.FromWave(VoiceAudioTests.Wave(1));
        try
        {
            foreach (bool declared in new[] { false, true })
            {
                var stream = new ResponseStream(new byte[VoiceTranscriber.MaximumResponseBytes + 4096]);
                var content = new StreamContent(stream);
                if (declared)
                {
                    content.Headers.ContentLength = VoiceTranscriber.MaximumResponseBytes + 4096;
                }
                else
                {
                    Check(content.Headers.ContentLength is null, "Fixture unexpectedly has a known length.");
                }
                handler.Reply = (_, _) => Task.FromResult(new HttpResponseMessage(HttpStatusCode.OK) { Content = content });
                var error = await Failure(async () => await client.TranscribeAsync(audio));
                Check(error is InvalidDataException, "Oversized response did not explicitly refuse.");
                Check(stream.BytesRead == (declared ? 0 : VoiceTranscriber.MaximumResponseBytes + 1),
                    "Response read exceeded its bound or failed to detect streamed excess.");
                Check(stream.Disposed, "Refused response leaked its stream.");
            }
            string body = "{\"text\":\"boundary\"}";
            body = body.PadRight(VoiceTranscriber.MaximumResponseBytes, ' ');
            var boundary = new ResponseStream(Encoding.UTF8.GetBytes(body));
            handler.Reply = (_, _) => Task.FromResult(new HttpResponseMessage(HttpStatusCode.OK)
            {
                Content = new StreamContent(boundary)
            });
            Check((await client.TranscribeAsync(audio)).Text == "boundary", "Exact response limit was refused.");
            Check(boundary.BytesRead == VoiceTranscriber.MaximumResponseBytes && boundary.Disposed,
                "Successful response was not bounded and disposed.");
            Check(handler.Calls == 3, "Response-limit refusal retried.");
        }
        finally
        {
            root.Destroy();
        }
    }

    public static async Task CleanupFailure()
    {
        var root = new Ensemble.Ensemble(TextWriter.Null);
        var handler = new Handler { FailDisposal = true };
        var client = new VoiceTranscriber(root, Key, handler, TimeSpan.FromSeconds(5));
        var error = await Failure(() => Task.Run(() => root.Destroy()));
        Check(error is AggregateException, "Cleanup failure was swallowed.");
        Check(!error.ToString().Contains(Key), "Cleanup failure exposed its former credential.");
        Check(handler.Disposed && root.Transcriber() is null, "Cleanup failure retained the child.");
        Check(client.DestroyAsync().IsFaulted, "Cleanup failure was not retained on the joined task.");
        var audio = PcmVoiceAudio.FromWave(VoiceAudioTests.Wave(1));
        await Failure(async () => await client.TranscribeAsync(audio));
        root.Destroy();
    }

    public static async Task CancellationRecovery()
    {
        foreach (bool callerCancels in new[] { true, false })
        {
            var root = new Ensemble.Ensemble(TextWriter.Null);
            var handler = new Handler();
            var entered = new TaskCompletionSource(TaskCreationOptions.RunContinuationsAsynchronously);
            var observedCancellation = new TaskCompletionSource(TaskCreationOptions.RunContinuationsAsynchronously);
            using var caller = new CancellationTokenSource();
            handler.Reply = async (_, token) =>
            {
                using var registration = token.Register(() => observedCancellation.TrySetResult());
                entered.TrySetResult();
                await Task.Delay(Timeout.InfiniteTimeSpan, token);
                throw new Exception("Cancelled transport unexpectedly resumed.");
            };
            var deadline = callerCancels ? TimeSpan.FromSeconds(5) : TimeSpan.FromMilliseconds(100);
            var client = new VoiceTranscriber(root, Key, handler, deadline);
            var audio = PcmVoiceAudio.FromWave(VoiceAudioTests.Wave(1));
            try
            {
                var request = client.TranscribeAsync(audio, caller.Token);
                await entered.Task.WaitAsync(TimeSpan.FromSeconds(3));
                if (callerCancels)
                {
                    caller.Cancel();
                }
                await observedCancellation.Task.WaitAsync(TimeSpan.FromSeconds(3));
                var error = await Failure(async () => await request.WaitAsync(TimeSpan.FromSeconds(3)));
                Check(error is OperationCanceledException, "Cancellation was not preserved.");
                Check(handler.Calls == 1, "Cancelled request retried.");
                Check(!handler.Disposed && ReferenceEquals(root.Transcriber(), client),
                    "Request cancellation destroyed the reusable client.");
                handler.Reply = (_, _) => Task.FromResult(Json("{\"text\":\"after cancellation\"}"));
                var recovered = await client.TranscribeAsync(audio);
                Check(recovered.Text == "after cancellation" && handler.Calls == 2,
                    "Client did not recover after cancellation.");
            }
            finally
            {
                caller.Cancel();
                await client.DestroyAsync();
                root.Destroy();
            }
            Check(handler.Disposed && root.Transcriber() is null, "Recovery test leaked its transport.");
        }
    }

    public static async Task Lifetime()
    {
        var root = new Ensemble.Ensemble(TextWriter.Null);
        var handler = new Handler();
        var entered = new TaskCompletionSource(TaskCreationOptions.RunContinuationsAsynchronously);
        var cancelled = new TaskCompletionSource(TaskCreationOptions.RunContinuationsAsynchronously);
        var release = new TaskCompletionSource(TaskCreationOptions.RunContinuationsAsynchronously);
        handler.Reply = async (_, token) =>
        {
            using var registration = token.Register(() => cancelled.TrySetResult());
            entered.SetResult();
            // Deliberately ignore cancellation until the fixture releases us.
            await release.Task;
            return Json("{\"text\":\"late\"}");
        };
        var client = new VoiceTranscriber(root, Key, handler, TimeSpan.FromSeconds(5));
        Task<VoiceTranscript>? request = null;
        try
        {
            var refusedHandler = new Handler();
            await Failure(() => Task.FromResult(new VoiceTranscriber(root, Key, refusedHandler, TimeSpan.FromSeconds(5))));
            Check(!refusedHandler.Disposed && ReferenceEquals(root.Transcriber(), client), "Refused attachment changed ownership.");
            refusedHandler.Dispose();
            var audio = PcmVoiceAudio.FromWave(VoiceAudioTests.Wave(1));
            request = client.TranscribeAsync(audio);
            await entered.Task.WaitAsync(TimeSpan.FromSeconds(3));
            await Failure(async () => await client.TranscribeAsync(audio));
            Task destruction = client.DestroyAsync();
            await cancelled.Task.WaitAsync(TimeSpan.FromSeconds(3));
            Check(!destruction.IsCompleted && !handler.Disposed, "Destroy failed to join running transport.");
            await Failure(async () => await client.TranscribeAsync(audio));
            release.SetResult();
            Check(await Failure(async () => await request) is OperationCanceledException, "Late transcript escaped cancellation.");
            await destruction.WaitAsync(TimeSpan.FromSeconds(3));
            Check(handler.Disposed && root.Transcriber() is null, "Destruction did not detach/dispose.");
            Check(ReferenceEquals(destruction, client.DestroyAsync()), "Repeated destruction changed task.");
        }
        finally
        {
            release.TrySetResult();
            await client.DestroyAsync();
            root.Destroy();
        }
    }
}
