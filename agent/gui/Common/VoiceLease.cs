using Ensemble;

namespace EnsembleGui;

// Owned by one socket operation; Target is a borrowed real Agent, never an ID
// reconstructed after selection changes. The socket must dispose/join this lease.
internal sealed partial class VoiceLease : IAsyncDisposable
{
    private readonly object gate = new();
    private readonly CancellationTokenSource lifetime;
    private bool cancelled;
    private bool consumed;
    private Task<VoiceTranscript>? transcription;
    private Task<string>? commit;
    private Task? destruction;
    internal Agent Target { get; }
    internal long Sequence { get; }
    internal partial VoiceLease(Agent target, long sequence, CancellationToken connection);
    internal partial Task<VoiceTranscript> Start(PcmVoiceAudio audio);
    internal partial Task<string> CommitAsync(Agent selected, long sequence);
    internal partial bool IsActive();
    internal partial void Cancel();
    public partial ValueTask DisposeAsync();
    private partial Task DestroyOwnedAsync(Task<VoiceTranscript>? request,
        Task<string>? admission, TaskCompletionSource completed);
}
