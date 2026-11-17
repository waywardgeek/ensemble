namespace Ensemble;

public sealed partial class Agent
{
    // Shared with goal reminder failures: unlimited attempts, capped delay.
    private static partial TimeSpan FailureRetryDelay(int failures) => TimeSpan.FromSeconds(
        failures <= 1 ? 1 : failures == 2 ? 10 : failures == 3 ? 100 : 1000);

    private partial bool ScheduleProviderRetry(ParsedResponse response)
    {
        // Goal reminders retain their existing turn-finish/backoff lifecycle.
        // Compressors have a separate bounded lifetime and must not retry forever.
        if (!response.ProviderFailure || active?.Kind != MailKind.Submit
            || active.AutomaticGoal || memoryCompressor || interruptPending || stopping)
        {
            return false;
        }
        if (providerFailureCount < int.MaxValue)
        {
            providerFailureCount++;
        }
        TimeSpan delay = FailureRetryDelay(providerFailureCount);
        // ParsedResponse is also constructible locally. Do not let invalid advice
        // overflow the timestamp or exceed System.Threading.Timer's range.
        if (response.RetryAfter is TimeSpan advice && advice > delay
            && advice.TotalMilliseconds <= uint.MaxValue - 1)
        {
            delay = advice;
        }
        providerRetryDue = DateTimeOffset.UtcNow + delay;
        calls = [];
        nextCall = 0;
        responseText = "";
        return true;
    }

    private partial bool AwaitProviderRetry()
    {
        if (providerRetryDue is not DateTimeOffset due)
        {
            return false;
        }
        TimeSpan remaining = due - DateTimeOffset.UtcNow;
        if (remaining <= TimeSpan.Zero)
        {
            providerRetryTimer?.Dispose();
            providerRetryTimer = null;
            providerRetryDue = null;
            return false;
        }
        providerRetryTimer ??= new Timer(_ => mailbox.Writer.TryWrite(new Mail(MailKind.ProviderRetryWake)),
            null, remaining, Timeout.InfiniteTimeSpan);
        return true;
    }

    private partial void ClearProviderRetry()
    {
        providerRetryTimer?.Dispose();
        providerRetryTimer = null;
        providerRetryDue = null;
        providerFailureCount = 0;
    }
}
