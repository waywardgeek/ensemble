namespace Ensemble;

public sealed partial class Agent
{
    public partial string? DisplayName()
    {
        lock (ingress) { return displayName; }
    }

    internal static partial string? NormalizeDisplayName(string? name)
    {
        if (name is null) return null;
        if (name.Length > 120 || name.Any(character => char.IsControl(character)
            || character is '\u2028' or '\u2029'))
            throw new ArgumentException("Display name must be a single line of at most 120 characters");
        name = name.Trim();
        return name.Length == 0 ? null : name;
    }

    public partial async Task SetDisplayNameAsync(string? name, CancellationToken cancellation)
    {
        name = NormalizeDisplayName(name);
        cancellation.ThrowIfCancellationRequested();
        var reply = new TaskCompletionSource<string>(TaskCreationOptions.RunContinuationsAsynchronously);
        Enqueue(new Mail(MailKind.DisplayNameSettings, Reply: reply, DisplayName: name));
        await reply.Task.WaitAsync(cancellation);
    }

    private partial void ApplyDisplayName(Mail message)
    {
        try
        {
            var saved = Memory().ReadSettings();
            Memory().WriteSettings(new AgentSettings(saved?.Memory ?? MemorySettings(),
                saved?.Recall ?? RecallSettings(), saved?.WorkingContext ?? WorkingContextLimits(),
                saved?.SelectedOffering, message.DisplayName));
            lock (ingress) { displayName = message.DisplayName; }
            message.Reply!.TrySetResult("");
        }
        catch (Exception error)
        {
            message.Reply!.TrySetException(error);
        }
    }
}
