using System.Collections.Immutable;

namespace Ensemble;

// Value-like foreground result. Full selected source bytes allow the actor to
// recheck deduplication after hints or skill controls arrive during the judge.
internal sealed record RecallCompletion(ImmutableArray<RecallHit> Selected, int ByteLimit);

// Value-like disk document. This is a transfer value, not a second live policy
// owner: Agent remains authoritative for current settings.
internal sealed record AgentSettings(MemorySettings Memory, RecallSettings Recall);

public static partial class JsonCodec
{
    public static partial System.Text.Json.Nodes.JsonObject WriteRecallSettings(RecallSettings settings);
    public static partial RecallSettings ReadRecallSettings(System.Text.Json.JsonElement value);
}

public sealed partial class Agent
{
    // Like the tool-round policy, this immutable setting is ingress-protected
    // and captured at turn admission. This setter does not persist settings.
    private RecallSettings recallSettings = new();
    private Task? recallTask;
    public partial RecallSettings RecallSettings();
    public partial void SetRecallSettings(RecallSettings settings);
    // Save before publishing next-turn policy. Cancellation detaches the observer
    // after admission; it does not retract a durable update or cancel recall.
    public partial Task SaveRecallSettingsAsync(RecallSettings settings,
        CancellationToken cancellation = default);
    private partial void ApplyRecallSettings(Mail message);
    private partial bool BeginRecall(Mail turn);
    private partial Task RecallWorker(Mail turn, string query, Context context,
        string instructions, RecallSettings settings, CancellationToken cancellation);
    private partial void ReceiveRecall(Mail message);
}
