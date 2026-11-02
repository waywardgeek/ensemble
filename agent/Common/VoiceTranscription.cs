namespace Ensemble;

// Provider usage is separate from the admitted audio's physical duration.
// Null usage means unreported, not free. These records contain no credentials.
public abstract record VoiceTranscriptionUsage;
public sealed record VoiceTokenUsage(long InputTokens, long OutputTokens, long TotalTokens,
    long? AudioInputTokens, long? TextInputTokens) : VoiceTranscriptionUsage;
public sealed record VoiceDurationUsage(double Seconds) : VoiceTranscriptionUsage;
public sealed record VoiceTranscript(string Text, TimeSpan InputDuration,
    VoiceTranscriptionUsage? Usage);

public sealed partial class Ensemble
{
    // Protected by the existing root membership lock, also used to close admission.
    private VoiceTranscriber? voiceTranscriber;
    public partial VoiceTranscriber? Transcriber();
    internal partial void AddVoiceTranscriber(VoiceTranscriber child);
    internal partial void RemoveVoiceTranscriber(VoiceTranscriber child);
}

/// <summary>Optional, separately funded OpenAI API-key transcription owner.
/// Never consults an Agent's provider, plan selection, environment or credentials.
/// One request at a time; no automatic retry, persistence, tracing or chat send.</summary>
public sealed partial class VoiceTranscriber
{
    public const string Model = "gpt-transcribe";
    public const int MaximumResponseBytes = 65_536;
    public const int MaximumTranscriptBytes = 32_768;
    internal Ensemble? parent;
    private readonly object gate = new();
    private string apiKey;
    // This owner owns both. HttpClient borrows its handler so failed attachment
    // does not dispose a caller's test transport before ownership transfers.
    private readonly HttpClient http;
    private readonly HttpMessageHandler transport;
    private readonly CancellationTokenSource lifetime = new();
    private readonly TimeSpan requestTimeout;
    private bool stopping;
    // Retain at most the last request until replacement/destruction. A completed
    // Task may be replaced; running work must always remain joinable.
    private Task<VoiceTranscript>? active;
    private Task? destruction;

    public partial VoiceTranscriber(Ensemble parent, string apiKey);
    internal partial VoiceTranscriber(Ensemble parent, string apiKey,
        HttpMessageHandler? transport, TimeSpan requestTimeout);
    public partial Ensemble Parent();
    public partial Task<VoiceTranscript> TranscribeAsync(PcmVoiceAudio audio,
        CancellationToken cancellation = default);
    public partial Task DestroyAsync();
    public partial void Destroy();
    internal partial string[] DiagnosticKeys();
    private partial Task<VoiceTranscript> RequestAsync(PcmVoiceAudio audio,
        CancellationToken cancellation, Task admitted);
    private partial Task DestroyOwnedAsync(Task<VoiceTranscript>? work, Task admitted);
    private static partial VoiceTranscript ParseTranscript(ReadOnlyMemory<byte> body,
        TimeSpan inputDuration, string apiKey);
}
