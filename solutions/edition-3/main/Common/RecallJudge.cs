using System.Collections.Immutable;
using System.Text.Json.Nodes;

namespace Ensemble;

public sealed partial class Configuration
{
    // A disposable same-provider configuration, never a running-model change.
    // Drop skills, cache markers, and main-model reasoning settings. Constructor
    // validation applies the selected judge model's own documented defaults.
    internal partial Configuration ForRecallJudge(string model, string instructions);
}

public sealed partial class Engine
{
    // One capability rule for ordinary and judge request streaming options.
    private static partial void AddStreamingOptions(JsonObject request, Configuration config);


    // One stateless tool-free exchange. Session totals include its actual usage;
    // Latest and CacheLens remain observations of ordinary sends. No History is
    // read or published here, and cancellation reaches the real HTTP operation.
    internal partial Task<ParsedResponse> JudgeRecallAsync(string model,
        string instructions, string input, CancellationToken cancellation);
}

public sealed partial class Recaller
{
    private static readonly TimeSpan JudgeTimeout = TimeSpan.FromSeconds(15);

    // Input hits have already been refreshed and filtered for retained context.
    // Judge failure falls back over the whole pool, not its candidate shortlist.
    // A successful empty verdict stays empty. Caller cancellation never becomes
    // fallback, while the private judge deadline does. No History publication.
    internal partial Task<RecallAttachment?> SelectAsync(IEnumerable<RecallHit> hits,
        string query, Context context, RecallSettings settings,
        CancellationToken cancellation = default);
}

internal static partial class RecallText
{
    // JSON framing keeps conversation and source snippets explicitly data.
    internal static partial string JudgeInput(string query, Context context,
        int contextMessages, ImmutableArray<RecallHit> candidates);
}
