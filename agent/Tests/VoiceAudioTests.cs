using System.Buffers.Binary;
using Ensemble;

internal static class VoiceAudioTests
{
    private static void Check(bool condition, string message)
    {
        if (!condition) throw new Exception(message);
    }

    private static void Refuse(byte[] wave, string reason)
    {
        try { PcmVoiceAudio.FromWave(wave); }
        catch (InvalidDataException) { return; }
        throw new Exception("Accepted invalid voice audio: " + reason);
    }

    // Build the published wire layout independently of the product's constants.
    // Changing the product's policy must not silently change the test fixtures.
    private static byte[] Wave(int samples)
    {
        var wave = new byte[44 + samples * 2];
        "RIFF"u8.CopyTo(wave);
        BinaryPrimitives.WriteUInt32LittleEndian(wave.AsSpan(4), (uint)wave.Length - 8);
        "WAVEfmt "u8.CopyTo(wave.AsSpan(8));
        BinaryPrimitives.WriteUInt32LittleEndian(wave.AsSpan(16), 16);
        BinaryPrimitives.WriteUInt16LittleEndian(wave.AsSpan(20), 1);
        BinaryPrimitives.WriteUInt16LittleEndian(wave.AsSpan(22), 1);
        BinaryPrimitives.WriteUInt32LittleEndian(wave.AsSpan(24), 16_000);
        BinaryPrimitives.WriteUInt32LittleEndian(wave.AsSpan(28), 32_000);
        BinaryPrimitives.WriteUInt16LittleEndian(wave.AsSpan(32), 2);
        BinaryPrimitives.WriteUInt16LittleEndian(wave.AsSpan(34), 16);
        "data"u8.CopyTo(wave.AsSpan(36));
        BinaryPrimitives.WriteUInt32LittleEndian(wave.AsSpan(40), (uint)samples * 2);
        return wave;
    }

    public static void Boundaries()
    {
        byte[] one = Wave(1);
        BinaryPrimitives.WriteInt16LittleEndian(one.AsSpan(44), short.MinValue);
        var smallest = PcmVoiceAudio.FromWave(one);
        Check(smallest.SampleCount == 1 && smallest.Duration.Ticks == 625,
            "One sample must have its actual nonzero duration.");
        Check(smallest.WaveBytes.AsSpan().SequenceEqual(one), "Sample bytes were transformed.");
        Check(PcmVoiceAudio.MimeType == "audio/wav", "Incorrect voice MIME type.");

        byte[] minute = Wave(960_000);
        var maximum = PcmVoiceAudio.FromWave(minute);
        Check(maximum.SampleCount == 960_000 && maximum.Duration == TimeSpan.FromSeconds(60),
            "The exact 60-second boundary was not admitted.");
        Check(maximum.WaveBytes.Length == 1_920_044, "Unexpected maximum byte count.");
        // Digital silence is structurally valid: this validator is not VAD.
        Check(maximum.WaveBytes.AsSpan().SequenceEqual(minute), "Silent samples changed.");
        Refuse(Wave(960_001), "one sample beyond 60 seconds");
        Refuse(new byte[1_920_045], "one byte beyond the allocation ceiling");
        Refuse(Wave(0), "empty audio");
    }

    public static void HeaderRefusals()
    {
        byte[] valid = Wave(4);
        for (int length = 0; length < valid.Length; length++)
            Refuse(valid.AsSpan(0, length).ToArray(), "truncation at byte " + length);

        // Every header bit participates in the canonical contract. This catches
        // omitted checks for rate, codec, channel count, alignment and lengths.
        for (int offset = 0; offset < 44; offset++)
        {
            for (int bit = 0; bit < 8; bit++)
            {
                byte[] changed = (byte[])valid.Clone();
                changed[offset] ^= (byte)(1 << bit);
                Refuse(changed, $"header byte {offset}, bit {bit}");
            }
        }

        byte[] trailing = new byte[valid.Length + 8];
        valid.CopyTo(trailing, 0);
        "JUNK"u8.CopyTo(trailing.AsSpan(valid.Length));
        Refuse(trailing, "undeclared trailing chunk");
        BinaryPrimitives.WriteUInt32LittleEndian(trailing.AsSpan(4), (uint)trailing.Length - 8);
        Refuse(trailing, "additional chunk inside RIFF");

        byte[] odd = valid[..^1];
        BinaryPrimitives.WriteUInt32LittleEndian(odd.AsSpan(4), (uint)odd.Length - 8);
        BinaryPrimitives.WriteUInt32LittleEndian(odd.AsSpan(40), (uint)odd.Length - 44);
        Refuse(odd, "consistent lengths but a partial PCM sample");

        foreach (int offset in new[] { 4, 16, 24, 28, 40 })
        {
            byte[] huge = (byte[])valid.Clone();
            BinaryPrimitives.WriteUInt32LittleEndian(huge.AsSpan(offset), uint.MaxValue);
            Refuse(huge, "oversized integer at " + offset);
        }
    }

    public static void SnapshotOwnership()
    {
        byte[] source = Wave(160);
        BinaryPrimitives.WriteInt16LittleEndian(source.AsSpan(44), 12_345);
        byte[] expected = (byte[])source.Clone();
        var first = PcmVoiceAudio.FromWave(source);
        var second = PcmVoiceAudio.FromWave(source);
        Array.Fill(source, (byte)0xff);
        Check(first.WaveBytes.AsSpan().SequenceEqual(expected), "Caller mutation changed admitted audio.");
        Check(second.WaveBytes.AsSpan().SequenceEqual(expected), "A second admission shared caller memory.");
        Check(first.SampleCount == 160 && first.Duration == TimeSpan.FromMilliseconds(10),
            "Duration changed with caller memory.");
        Check(!ReferenceEquals(first, second), "Admissions unexpectedly shared a mutable owner.");
        Check(typeof(PcmVoiceAudio).GetConstructors().Length == 0,
            "An unchecked public constructor bypasses audio admission.");
    }
}
