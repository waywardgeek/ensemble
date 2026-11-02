using Ensemble;
using System.Text.Json;
using System.Threading.Channels;

namespace EnsembleGui;

internal sealed partial class VoicePeer
{
    internal partial VoicePeer(Channel<string> queue, Agent selected, long cursor)
    {
        Queue = queue;
        Selected = selected;
        Cursor = cursor;
    }

    internal partial async Task DestroyVoiceAsync()
    {
        var lease = Voice;
        if (lease is null) return;
        // Invalidate before joining notification, which may still be waiting for
        // a transport that has not acknowledged cancellation.
        var destruction = lease.DisposeAsync();
        try { await VoiceNotification.ConfigureAwait(false); }
        finally { await destruction.ConfigureAwait(false); }
        Voice = null;
        VoiceNotification = Task.CompletedTask;
    }
}

public static partial class Gui
{
    private static partial PcmVoiceAudio PrepareVoiceAudio(JsonElement command)
    {
        var audio = command.GetProperty("audio");
        if (audio.GetProperty("mime").GetString() != PcmVoiceAudio.MimeType)
            throw new InvalidDataException("Voice requires canonical PCM WAV.");
        string data = audio.GetProperty("data").GetString() ?? "";
        int maximum = ((PcmVoiceAudio.MaximumWaveBytes + 2) / 3) * 4;
        if (data.Length == 0 || data.Length > maximum || data.Length % 4 != 0
            || data.Any(char.IsWhiteSpace))
            throw new InvalidDataException("Voice payload exceeds its bounded base64 format.");
        var bytes = new byte[data.Length / 4 * 3];
        if (!Convert.TryFromBase64String(data, bytes, out int count))
            throw new InvalidDataException("Voice payload is not base64.");
        return PcmVoiceAudio.FromWave(bytes.AsSpan(0, count));
    }
}
