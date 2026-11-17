namespace Ensemble;

public sealed partial class Agent
{
    // Actor-owned transient state: the admitted task and its work generation stay
    // alive. Timer callbacks only wake the mailbox; they never send or replay work.
    private Timer? providerRetryTimer;
    private DateTimeOffset? providerRetryDue;
    private int providerFailureCount;
    private partial bool ScheduleProviderRetry(ParsedResponse response);
    private partial bool AwaitProviderRetry();
    private partial void ClearProviderRetry();
    private static partial TimeSpan FailureRetryDelay(int failures);
}
