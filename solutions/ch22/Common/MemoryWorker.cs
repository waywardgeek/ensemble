namespace Ensemble;

/// <summary>Value-like outcome of one joined compressor attempt. Usage belongs
/// to this attempt even on failure; only a nonblank Summary can become a staged
/// candidate. This value never confers filesystem or History authority.</summary>
internal sealed record MemoryWorkerResult(string TaskID, string? Summary,
    string? Error, bool Cancelled, MeterSnapshot Meter);

public sealed partial class Agent
{
    // Fixed before the child actor starts. Its existing Agent parent owns the
    // relationship and teardown; this role is not another child registry.
    private bool memoryCompressor;
    private int compressorSends;
    internal partial bool IsMemoryCompressor();
    internal partial void ConfigureMemoryCompressor();
    /// <summary>Run a skill-free owned child, then join and detach it before
    /// returning immutable output. The parent actor alone may stage or commit
    /// that output. Cancellation stops this attempt, not the parent turn.</summary>
    internal partial Task<MemoryWorkerResult> RunMemoryCompressorAsync(
        MemoryCapture capture, MemoryBandSettings destinationPolicy,
        CancellationToken cancellation);
}

public sealed partial class Configuration
{
    // Reuse this actor's actual provider/funding/security choices, but never its
    // primary instructions, inherited skills, or ordinary tool allowance.
    internal partial Configuration ForMemoryCompressor();
}

public sealed partial class Engine
{
    // A terminal-success check scoped to the compressor. Ordinary parsing keeps
    // its established behavior. Send still meters the completed exchange before
    // refusing its incomplete output to the actor.
    private partial bool IncompleteCompression(string body);
}
