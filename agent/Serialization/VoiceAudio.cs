using System.Buffers.Binary;
using System.Collections.Immutable;

namespace Ensemble;

public sealed partial class PcmVoiceAudio
{
    public static partial PcmVoiceAudio FromWave(ReadOnlySpan<byte> source)
    {
        // Bound allocation first, then validate the exact snapshot we retain.
        // Validating caller memory before copying would permit header mutation
        // between the check and admission.
        if (source.Length < HeaderBytes || source.Length > MaximumWaveBytes)
            throw new InvalidDataException("Voice WAV must contain at most 60 seconds of mono 16 kHz PCM16 audio.");

        ImmutableArray<byte> snapshot = ImmutableArray.Create<byte>(source);
        ReadOnlySpan<byte> wave = snapshot.AsSpan();
        if (!wave[..4].SequenceEqual("RIFF"u8)
            || !wave.Slice(8, 4).SequenceEqual("WAVE"u8)
            || !wave.Slice(12, 4).SequenceEqual("fmt "u8)
            || !wave.Slice(36, 4).SequenceEqual("data"u8))
            throw new InvalidDataException("Voice audio requires the canonical PCM WAV layout.");

        uint riffBytes = BinaryPrimitives.ReadUInt32LittleEndian(wave.Slice(4, 4));
        uint formatBytes = BinaryPrimitives.ReadUInt32LittleEndian(wave.Slice(16, 4));
        ushort format = BinaryPrimitives.ReadUInt16LittleEndian(wave.Slice(20, 2));
        ushort channels = BinaryPrimitives.ReadUInt16LittleEndian(wave.Slice(22, 2));
        uint sampleRate = BinaryPrimitives.ReadUInt32LittleEndian(wave.Slice(24, 4));
        uint byteRate = BinaryPrimitives.ReadUInt32LittleEndian(wave.Slice(28, 4));
        ushort blockAlign = BinaryPrimitives.ReadUInt16LittleEndian(wave.Slice(32, 2));
        ushort bitsPerSample = BinaryPrimitives.ReadUInt16LittleEndian(wave.Slice(34, 2));
        uint declaredDataBytes = BinaryPrimitives.ReadUInt32LittleEndian(wave.Slice(40, 4));
        int actualDataBytes = wave.Length - HeaderBytes;

        // All rates and sizes must agree with our fixed wire contract. We do
        // not calculate allocation sizes or duration from claimed header rates.
        if (formatBytes != 16 || format != 1 || channels != Channels
            || sampleRate != SampleRate || byteRate != SampleRate * BlockAlign
            || blockAlign != BlockAlign || bitsPerSample != BitsPerSample)
            throw new InvalidDataException("Voice audio must be mono, 16 kHz, 16-bit little-endian PCM.");

        if (riffBytes != wave.Length - 8 || declaredDataBytes != actualDataBytes
            || actualDataBytes == 0 || actualDataBytes % BlockAlign != 0)
            throw new InvalidDataException("Voice WAV has empty, truncated, extra, or misaligned sample data.");

        return new PcmVoiceAudio(snapshot);
    }
}
