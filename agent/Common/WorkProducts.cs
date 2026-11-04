namespace Ensemble;

// Retained description, not live authority. The opaque ID is useful only while
// the owning Agent retains the corresponding live offer. No host path is exposed.
public sealed partial class WorkProductInfo
{
    public readonly string ID;
    public readonly string Name;
    public readonly long OfferedBytes;
    public partial WorkProductInfo(string id, string name, long offeredBytes);
}

public sealed partial class Agent
{
    public const long MaximumWorkProductBytes = 64L * 1024 * 1024;
    public const int MaximumWorkProductOffers = 64;
    private readonly object workProductGate = new();
    private bool workProductsClosed;
    // Immutable values owned by this Agent, not a second filesystem service.
    private sealed record WorkProductOffer(WorkProductInfo Info, string Path);
    private readonly Dictionary<string, WorkProductOffer> workProducts = new(StringComparer.Ordinal);
    private readonly CancellationTokenSource workProductStop = new();
    private readonly HashSet<WorkProductTransfer> workProductTransfers = [];
    private sealed partial class WorkProductTransfer
    {
        internal Agent? parent;
        internal readonly FileStream input;
        internal readonly CancellationTokenSource stop;
        internal readonly TaskCompletionSource joined = new(TaskCreationOptions.RunContinuationsAsynchronously);
        internal partial WorkProductTransfer(Agent parent, FileStream input, CancellationToken cancellation);
    }


    // Validate a regular file before registering a live offer. Failed creation
    // consumes no slot. The contents are not copied or frozen by this operation.
    public partial WorkProductInfo OfferWorkProduct(string path);
    // Admission reopens through the same Sandbox and checks the transfer bound.
    // The returned stream belongs to the caller, just like Sandbox.OpenRegularRead.
    public partial FileStream OpenWorkProduct(string id);
    // Own the file/copy lifetime until completion; destination remains borrowed.
    // Caller cancellation stops this transfer, not the offer or another transfer.
    public partial Task<long> CopyWorkProductAsync(string id, Stream destination, CancellationToken cancellation = default);
    private partial void CloseWorkProducts();
}
