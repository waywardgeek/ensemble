namespace Ensemble;

// Value-like mailbox result. Failure before constructing a child has no meter;
// an admitted child returns its final immutable measurement even on failure.
internal sealed record MemoryCompletion(string TaskID, MemoryWorkerResult? Result, string? Error);

public sealed partial class Agent
{
    // One actor-owned attempt. The immutable capture is not a second transcript
    // store; it is the selected input whose identity must survive worker latency.
    private MemoryCapture? memoryCapture;
    private MemorySettings? memoryLaunchSettings;
    private CancellationTokenSource? memoryCancellation;
    private Task<MemoryCompletion>? memoryTask;
    private MemoryCompletion? memoryCompletion;
    // Process-only receipt for the retained attempt. History replay must never
    // import a saved outcome into a fresh process meter.
    private bool memoryUsageImported;

    private partial void StartMemoryCompression(bool includeConversation);
    private partial Task<MemoryCompletion> MemoryWorker(MemoryCapture capture,
        MemoryBandSettings destination, CancellationToken cancellation);
    private partial void CompleteMemoryCompression();
    private partial void AcceptMemoryCompletion(MemoryCompletion completion);
    private partial Task InvalidateMemoryPolicy(MemorySettings desired);
    private partial Task StopMemoryCompression(CompressionState state, string reason, bool record = true);
    private partial void EndMemoryCompression(CompressionState state, string reason);
    private partial void ReleaseMemoryCompression();
}
