using System.Net;
using System.Text.Json;
using Ensemble;
using EnsembleGui;

// Real browser -> socket -> lease -> transcriber -> actor -> model request.
// Only the two external HTTP provider boundaries are fake. No real credentials.
internal static class VoiceBrowserFixture
{
    private const string Transcript = "VOICE-FIXTURE admitted exactly once";
    internal static async Task RunAsync()
    {
        string home = Directory.GetCurrentDirectory();
        using var stop = new CancellationTokenSource(TimeSpan.FromMinutes(3));
        var root = new Ensemble.Ensemble(TextWriter.Null, dataDirectory: Path.Combine(home, "data"));
        using var audio = new AudioHandler();
        _ = new VoiceTranscriber(root, "synthetic-browser-payer", audio, TimeSpan.FromSeconds(30));
        var builder = WebApplication.CreateBuilder(); builder.Logging.ClearProviders();
        await using var host = builder.Build(); host.Urls.Add("http://127.0.0.1:18202");
        var config = new Configuration(Vendor.Anthropic, "claude-3-haiku-20240307",
            "http://127.0.0.1:18202", "synthetic-browser-model", disableStreaming: true, sandboxRoot: home);
        var agent = new Agent(root, config, Path.Combine(home, "jobs"));
        var history = agent.History();
        int modelRequests = 0;
        string? failure = null;
        int Admitted() => history.View().Dialogue.SelectMany(m => m.Parts)
            .OfType<TextPart>().Count(p => p.Text == Transcript);
        object Status() => new { audioCalls = Volatile.Read(ref audio.Calls), audioCompleted = Volatile.Read(ref audio.Completed),
            waveBytes = Volatile.Read(ref audio.WaveBytes), modelRequests = Volatile.Read(ref modelRequests),
            admitted = Admitted(), failure = Volatile.Read(ref failure) ?? Volatile.Read(ref audio.Failure) };
        host.MapGet("/fixture/status", () => Results.Json(Status()));
        host.MapPost("/fixture/release", () => { audio.Release.TrySetResult(); return Results.Ok(); });
        host.MapPost("/v1/messages", async (HttpContext context) =>
        {
            try
            {
                Check(Interlocked.Increment(ref modelRequests) == 1, "Unexpected extra model request");
                using var request = await JsonDocument.ParseAsync(context.Request.Body, cancellationToken: context.RequestAborted);
                Check(context.Request.Headers["x-api-key"] == "synthetic-browser-model", "Wrong model payer");
                Check((!request.RootElement.TryGetProperty("stream", out var stream) || !stream.GetBoolean()), "Unexpected streaming request");
                Check(request.RootElement.GetProperty("messages").GetRawText().Contains(Transcript, StringComparison.Ordinal), "Transcript missing from model request");
                await context.Response.WriteAsJsonAsync(new { id = "fixture-reply", type = "message", role = "assistant",
                    model = "claude-3-haiku-20240307", content = new[] { new { type = "text", text = "VOICE-FIXTURE model received transcript" } },
                    stop_reason = "end_turn", usage = new { input_tokens = 10, output_tokens = 5 } }, context.RequestAborted);
            }
            catch (Exception error) { Volatile.Write(ref failure, error.Message); context.Response.StatusCode = 500; }
        });
        Task? serving = null;
        int joinedAdmissions = -1;
        try
        {
            serving = Gui.ServeAsync(host, agent, new GuiOptions(18202, BasePath: "/assistant",
                SettingsPath: Path.Combine(home, "settings.json"), TracePath: Path.Combine(home, "gui.log")), stop.Token);
            await host.StartAsync(stop.Token);
            Console.WriteLine("READY http://127.0.0.1:18202/assistant/");
            while (!File.Exists(Path.Combine(home, "stop"))) await Task.Delay(50, stop.Token);
            Check(failure is null && audio.Failure is null, failure ?? audio.Failure ?? "Provider failure");
            Check(audio.Calls == 2 && audio.Completed == 2 && modelRequests == 1 && Admitted() == 1,
                "Browser sequence did not establish exactly one admission and late-cancel refusal");
        }
        finally
        {
            audio.Release.TrySetResult(); stop.Cancel();
            try
            {
                await host.StopAsync(); if (serving is not null) await serving;
                // ServeAsync has joined socket leases; History is still alive here.
                joinedAdmissions = Admitted();
            }
            finally { root.Destroy(); }
            File.WriteAllText(Path.Combine(home, "result.json"), JsonSerializer.Serialize(new {
                audioCalls = audio.Calls, audioCompleted = audio.Completed, modelRequests, admitted = joinedAdmissions, failure = failure ?? audio.Failure }));
            Console.WriteLine("STOPPED host and Agents; transcription work joined");
        }
        Check(audio.Calls == 2 && audio.Completed == 2 && modelRequests == 1 && joinedAdmissions == 1,
            "Final joined counts changed after cancellation");
        Console.WriteLine("PASS two fake audio requests, one model request, one canonical transcript after joined teardown");
    }
    private static void Check(bool valid, string reason) { if (!valid) throw new InvalidOperationException(reason); }
    private sealed class AudioHandler : HttpMessageHandler
    {
        internal int Calls, Completed, WaveBytes;
        internal string? Failure;
        internal readonly TaskCompletionSource Release = new(TaskCreationOptions.RunContinuationsAsynchronously);
        protected override async Task<HttpResponseMessage> SendAsync(HttpRequestMessage request, CancellationToken cancellation)
        {
            int call = Interlocked.Increment(ref Calls);
            try
            {
                Check(call <= 2, "Unexpected extra transcription request");
                Check(request.RequestUri?.AbsoluteUri == "https://api.openai.com/v1/audio/transcriptions"
                    && request.Method == HttpMethod.Post, "Wrong transcription endpoint");
                Check(request.Headers.Authorization?.Parameter == "synthetic-browser-payer", "Wrong transcription payer");
                Check(request.Content is MultipartFormDataContent, "Missing multipart audio");
                var fields = ((MultipartFormDataContent)request.Content!).ToDictionary(
                    p => p.Headers.ContentDisposition!.Name!.Trim('"'), p => p);
                Check(fields.Count == 4 && await fields["model"].ReadAsStringAsync(cancellation) == "gpt-transcribe"
                    && await fields["stream"].ReadAsStringAsync(cancellation) == "false"
                    && await fields["response_format"].ReadAsStringAsync(cancellation) == "json", "Wrong transcription contract");
                Check(fields["file"].Headers.ContentType?.MediaType == "audio/wav", "Wrong audio MIME");
                byte[] wave = await fields["file"].ReadAsByteArrayAsync(cancellation);
                var pcm = PcmVoiceAudio.FromWave(wave);
                Check(pcm.WaveBytes.Length > 44, "Empty browser PCM");
                Volatile.Write(ref WaveBytes, wave.Length);
                // Deliberately ignore cancellation here: late success must not admit.
                if (call == 2) await Release.Task;
                return new HttpResponseMessage(HttpStatusCode.OK) { Content = new StringContent(
                    JsonSerializer.Serialize(new { text = Transcript })) };
            }
            catch (Exception error) { Volatile.Write(ref Failure, error.Message); throw; }
            finally { Interlocked.Increment(ref Completed); }
        }
    }
}
