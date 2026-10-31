using System.Collections.Immutable;

namespace Ensemble;

// A value, not a device or service owner. Only validated, canonical browser WAV
// bytes can construct it. Duration is derived from those bytes, never supplied
// by a browser clock or an untrusted container-duration field.
public sealed partial class PcmVoiceAudio
{
    public const int SampleRate = 16_000;
    public const int Channels = 1;
    public const int BitsPerSample = 16;
    public const int BlockAlign = Channels * BitsPerSample / 8;
    public const int HeaderBytes = 44;
    public const int MaximumSeconds = 60;
    public const int MaximumSamples = SampleRate * MaximumSeconds;
    public const int MaximumWaveBytes = HeaderBytes + MaximumSamples * BlockAlign;
    public const string MimeType = "audio/wav";

    public ImmutableArray<byte> WaveBytes { get; }
    public int SampleCount => (WaveBytes.Length - HeaderBytes) / BlockAlign;
    public TimeSpan Duration => TimeSpan.FromTicks(
        (long)SampleCount * TimeSpan.TicksPerSecond / SampleRate);

    private PcmVoiceAudio(ImmutableArray<byte> validatedBytes)
    {
        WaveBytes = validatedBytes;
    }

    // Refuses empty, oversized, compressed, extended, or inconsistent WAVs.
    // This proves format and duration, not the presence or accuracy of speech.
    public static partial PcmVoiceAudio FromWave(ReadOnlySpan<byte> source);
}
