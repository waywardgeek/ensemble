using Ensemble;
using System.Threading.Channels;

namespace EnsembleGui;

// The host peer collection owns this connection state; its socket scope joins
// voice work on removal. Selected is borrowed, not a second Agent owner.
internal sealed partial class VoicePeer
{
    internal Channel<string> Queue { get; }
    internal Agent Selected { get; set; }
    // Borrowed from the socket scope, which owns and disposes the transport.
    internal McpTunnel? BrowserTunnel { get; set; }
    // Borrowed target of this connection's transient automatic hold.
    internal long Cursor { get; set; }
    internal long VoiceSequence { get; set; }
    internal VoiceLease? Voice { get; set; }
    internal Task VoiceNotification { get; set; } = Task.CompletedTask;
    internal partial VoicePeer(Channel<string> queue, Agent selected, long cursor);
    internal partial Task DestroyVoiceAsync();
}

public static partial class Gui
{
    private const int MaximumVoiceControlBytes = ((PcmVoiceAudio.MaximumWaveBytes + 2) / 3) * 4 + 4096;
    private static partial PcmVoiceAudio PrepareVoiceAudio(System.Text.Json.JsonElement command);
}
