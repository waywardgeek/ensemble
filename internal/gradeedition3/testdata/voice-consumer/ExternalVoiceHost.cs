using System.Net;
using System.Buffers.Binary;
using System.Text.Json;
using Ensemble;
using EnsembleGui;

// Real browser -> socket -> lease -> transcriber -> actor -> model request.
// Adapted from the internal fixture, now a separately compiled public-API consumer.
// The grader owns this oracle; candidate code supplies all runtime behavior.
// Only the two external HTTP provider boundaries are fake. No real credentials.
internal static class ExternalVoiceHost
{
    private static readonly string Transcript = Environment.GetEnvironmentVariable("VOICE_GRADE_TRANSCRIPT")
        ?? throw new InvalidOperationException("Grader transcript challenge is required");
    private static readonly int Port = int.Parse(Environment.GetEnvironmentVariable("VOICE_GRADE_PORT")
        ?? throw new InvalidOperationException("Grader loopback port is required"));
    private static readonly string Mode = Environment.GetEnvironmentVariable("VOICE_GRADE_MODE") ?? "lifecycle";
    private static readonly bool FundingOff = Mode == "funding-off";
    private static readonly bool NoAdmission = Mode != "lifecycle";
    internal static async Task RunAsync()
    {
        Check(Mode is "lifecycle" or "funding-off" or "permission-delay" or "disconnect-recording", "Unknown voice grading mode");
        Check(Port is > 1024 and < 65536 && Port is not (8088 or 8089 or 8090), "Refuse live-service ports");
        string home = Directory.GetCurrentDirectory();
        using var stop = new CancellationTokenSource(TimeSpan.FromMinutes(3));
        var root = new Ensemble.Ensemble(TextWriter.Null, dataDirectory: Path.Combine(home, "data"));
        var audio = new AudioHandler();
        if (!FundingOff) _ = new VoiceTranscriber(root, "synthetic-browser-payer", audio);
        var builder = WebApplication.CreateBuilder(); builder.Logging.ClearProviders();
        await using var host = builder.Build(); host.Urls.Add($"http://127.0.0.1:{Port}");
        var config = new Configuration(Vendor.Anthropic, "claude-3-haiku-20240307",
            $"http://127.0.0.1:{Port}", "synthetic-browser-model", disableStreaming: true, sandboxRoot: home);
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
                int transcriptInputs = 0;
                foreach (var message in request.RootElement.GetProperty("messages").EnumerateArray())
                {
                    if (message.GetProperty("role").GetString() != "user") continue;
                    var content = message.GetProperty("content");
                    if (content.ValueKind == JsonValueKind.String)
                    {
                        if (content.GetString() == Transcript) transcriptInputs++;
                    }
                    else if (content.ValueKind == JsonValueKind.Array)
                    {
                        foreach (var part in content.EnumerateArray())
                        {
                            if (part.GetProperty("type").GetString() == "text"
                                && part.GetProperty("text").GetString() == Transcript) transcriptInputs++;
                        }
                    }
                }
                Check(transcriptInputs == 1, "Expected exactly one human transcript in model input");
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
            serving = Gui.ServeAsync(host, agent, new GuiOptions(Port, BasePath: "/assistant",
                SettingsPath: Path.Combine(home, "settings.json"), TracePath: Path.Combine(home, "gui.log")), stop.Token);
            await host.StartAsync(stop.Token);
            Console.WriteLine($"READY http://127.0.0.1:{Port}/assistant/");
            while (!File.Exists(Path.Combine(home, "stop"))) await Task.Delay(50, stop.Token);
            Check(failure is null && audio.Failure is null, failure ?? audio.Failure ?? "Provider failure");
            Check(audio.Calls == (NoAdmission ? 0 : 2) && audio.Completed == (NoAdmission ? 0 : 2)
                && modelRequests == (NoAdmission ? 0 : 1) && Admitted() == (NoAdmission ? 0 : 1),
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
            finally
            {
                try { root.Destroy(); }
                finally { if (FundingOff) audio.Dispose(); } // Never transferred in disabled mode.
            }
            File.WriteAllText(Path.Combine(home, "result.json"), JsonSerializer.Serialize(new {
                audioCalls = audio.Calls, audioCompleted = audio.Completed, modelRequests, admitted = joinedAdmissions, failure = failure ?? audio.Failure }));
            Console.WriteLine("STOPPED host and Agents; transcription work joined");
        }
        Check(audio.Disposed, "Owned transcription handler was not disposed");
        Check(audio.Calls == (NoAdmission ? 0 : 2) && audio.Completed == (NoAdmission ? 0 : 2)
            && modelRequests == (NoAdmission ? 0 : 1) && joinedAdmissions == (NoAdmission ? 0 : 1),
            "Final joined counts changed after cancellation");
        Console.WriteLine(NoAdmission ? $"PASS {Mode}: no audio/model request or admission"
            : "PASS two fake audio requests, one model request, one canonical transcript after joined teardown");
    }
    private static void Check(bool valid, string reason) { if (!valid) throw new InvalidOperationException(reason); }
    private sealed class AudioHandler : HttpMessageHandler
    {
        internal int Calls, Completed, WaveBytes;
        internal bool Disposed;
        protected override void Dispose(bool disposing)
        {
            Disposed = true;
            base.Dispose(disposing);
        }
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
                Check(wave.Length > 44 && wave.Length <= 1_920_044, "Empty or oversized browser PCM");
                Check(wave.AsSpan(0, 4).SequenceEqual("RIFF"u8)
                    && wave.AsSpan(8, 8).SequenceEqual("WAVEfmt "u8)
                    && wave.AsSpan(36, 4).SequenceEqual("data"u8), "Noncanonical WAV identifiers");
                Check(BinaryPrimitives.ReadInt32LittleEndian(wave.AsSpan(4)) == wave.Length - 8
                    && BinaryPrimitives.ReadInt32LittleEndian(wave.AsSpan(16)) == 16
                    && BinaryPrimitives.ReadInt16LittleEndian(wave.AsSpan(20)) == 1
                    && BinaryPrimitives.ReadInt16LittleEndian(wave.AsSpan(22)) == 1
                    && BinaryPrimitives.ReadInt32LittleEndian(wave.AsSpan(24)) == 16000
                    && BinaryPrimitives.ReadInt32LittleEndian(wave.AsSpan(28)) == 32000
                    && BinaryPrimitives.ReadInt16LittleEndian(wave.AsSpan(32)) == 2
                    && BinaryPrimitives.ReadInt16LittleEndian(wave.AsSpan(34)) == 16
                    && BinaryPrimitives.ReadInt32LittleEndian(wave.AsSpan(40)) == wave.Length - 44
                    && (wave.Length - 44) % 2 == 0, "Noncanonical WAV format/length");
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
