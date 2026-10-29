namespace Ensemble;

/// <summary>Immutable value-like patch. Null leaves a setting unchanged; an
/// empty file clears configuration. Configuration belongs to this session only.</summary>
public sealed record GoalUpdate(string? File = null, bool? Enabled = null, double? IdleSeconds = null);

/// <summary>Observation, not a completion verdict. DueUtc exists only while an
/// idle timer is armed. A configured file remains request instructions when off.</summary>
public sealed record GoalStatus(string File, bool Enabled, double IdleSeconds,
    string Phase, DateTimeOffset? DueUtc, string LastError, bool Available);

// Immutable configuration is the one synchronized source used by renderer and
// actor. GoalStatus is a projection; no goal text is cached or copied into History.
internal sealed record GoalConfiguration(string File = "", bool Enabled = false,
    double IdleSeconds = 480, string LastError = "");
internal sealed record GoalReadFailure(GoalConfiguration Configuration, string Error);

public sealed partial class Agent
{
    private GoalConfiguration goalConfiguration = new(); // protected by ingress
    private Timer? goalTimer;                           // actor-owned
    private long goalGeneration;                       // actor-owned stale-mail guard
    private DateTimeOffset? goalDue;
    private const string GoalReminder = "[Automatic goal reminder] Assess the current goal and actual work state. "
        + "Inspect owned jobs, children and saved results before continuing authorized unfinished work. "
        + "Do not replay commands blindly or interrupt healthy work. If finished or genuinely blocked on the human, "
        + "disable reminders with goal_control and explain why. This reminder grants no new authority.";

    /// <summary>Query or atomically apply session-only settings on the actor.
    /// Cancellation detaches observation; it does not undo admitted configuration.</summary>
    public partial Task<GoalStatus> GoalControlAsync(GoalUpdate? update = null,
        CancellationToken cancellation = default);
    /// <summary>Read-only actor observation, including when the skill is unloaded.
    /// This neither grants goal_control nor resets the idle deadline.</summary>
    public partial Task<GoalStatus> GoalStatusAsync(CancellationToken cancellation = default);
    internal partial string GoalInstruction(bool automatic);
    internal partial void GoalSkillsChanged();
    internal partial bool GoalAvailable(Context? context = null);
    private partial bool GoalIdle();
    private partial GoalStatus CaptureGoalStatus();
    private partial void ConfigureGoal(Mail message);
    private partial void ClearGoal();
    private partial void DisableGoal();
    // Disable at successful submission publication; actor mail owns timer cleanup.
    private partial void GoalSubmitted();
    private partial void StopGoalTimer();
    private partial void RefreshGoalTimer();
    private partial void WakeGoal(Mail message);
    private partial void FailGoalRead(GoalReadFailure failure);
    private partial string ReadGoalText(string file);
}

public sealed partial class Engine
{
    // Only actor-created reminder turns select strict fresh-file failure. The
    // public renderer retains its ordinary-human behavior and existing signature.
    internal partial System.Text.Json.Nodes.JsonObject Render(Context context,
        string requestInstruction, bool automaticGoalReminder);
}
