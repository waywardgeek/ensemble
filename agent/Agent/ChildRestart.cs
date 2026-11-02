using System.Text.Json;
using System.Security.Cryptography;
using System.Text.Json.Serialization;

namespace Ensemble;

public sealed partial class Agent
{
    private partial void SaveChildRestart(GoalConfiguration? goal, bool? pause,
        long? generation, bool? submitted, string? model, JsonElement? completion)
    {
        lock (ingress)
        {
            if (!childRestartEnabled || childRestoring)
            {
                return;
            }
            bool isSubmitted = submitted ?? submittedPayload.HasValue;
            string? submissionHash = isSubmitted
                ? Convert.ToHexString(SHA256.HashData(JsonSerializer.SerializeToUtf8Bytes(
                    completion ?? submittedPayload!.Value))) : null;
            var state = new ChildRestartState(1, childID!, model ?? configuration.Model,
                submissionContract.Schema(), generation ?? admittedWorkGeneration,
                isSubmitted, submissionHash, pause ?? explicitPause, goal ?? goalConfiguration);
            Sandbox().ReplaceFile(Path.Combine(DataDirectory(), "child.json"),
                JsonSerializer.SerializeToUtf8Bytes(state));
        }
    }

    internal partial void RetireChildRestart()
    {
        lock (ingress)
        {
            if (!childRestartEnabled) { return; }
            Sandbox().DeleteFile(Path.Combine(DataDirectory(), "child.json"));
            childRestartEnabled = false;
        }
    }

    private partial ChildRestartState? ReadChildRestart(string id)
    {
        ValidateChildID(id);
        string path = Path.Combine(DataDirectory(), "children", id, "child.json");
        byte[] bytes;
        try
        {
            using var stream = Sandbox().OpenRegularRead(path);
            using var contents = new MemoryStream();
            byte[] buffer = new byte[8192];
            while (true)
            {
                int count = stream.Read(buffer, 0, buffer.Length);
                if (count == 0) { break; }
                if (contents.Length + count > SubmissionArtifactBytes)
                {
                    throw new InvalidDataException("Child restart record exceeds its byte limit.");
                }
                contents.Write(buffer, 0, count);
            }
            bytes = contents.ToArray();
        }
        catch (FileNotFoundException) { return null; }
        catch (DirectoryNotFoundException) { return null; }
        try
        {
            using var document = JsonDocument.Parse(bytes);
            JsonElement root = document.RootElement;
            foreach (JsonElement item in new[] { root, root.GetProperty("Goal") })
            {
                var names = new HashSet<string>(StringComparer.Ordinal);
                foreach (JsonProperty property in item.EnumerateObject())
                {
                    if (!names.Add(property.Name))
                    {
                        throw new InvalidDataException("Duplicate child restart field.");
                    }
                }
            }
            var state = JsonSerializer.Deserialize<ChildRestartState>(bytes,
                new JsonSerializerOptions { UnmappedMemberHandling = JsonUnmappedMemberHandling.Disallow });
            if (state is null || state.Version != 1 || state.ID != id
                || string.IsNullOrWhiteSpace(state.Model) || state.Generation < 0 || state.Goal is null
                || state.Goal.File is null || state.Goal.LastError is null
                || !double.IsFinite(state.Goal.IdleSeconds) || state.Goal.IdleSeconds < .001
                || state.Goal.IdleSeconds > (uint.MaxValue - 1) / 1000d
                || (state.Goal.Enabled && state.Goal.File.Length == 0)
                || (state.Submitted ? state.SubmissionHash is null || state.SubmissionHash.Length != 64
                    || !state.SubmissionHash.All(char.IsAsciiHexDigit) : state.SubmissionHash is not null)
                || root.EnumerateObject().Count() != 9
                || root.GetProperty("Goal").EnumerateObject().Count() != 4)
            {
                throw new InvalidDataException("Invalid child restart identity or settings.");
            }
            _ = new SubmissionContract(state.Schema);
            return state;
        }
        catch (Exception error) when (error is JsonException or ArgumentException
            or InvalidOperationException or KeyNotFoundException)
        {
            throw new InvalidDataException("Malformed child restart record.", error);
        }
    }

    public partial async Task RestoreChildrenAsync(CancellationToken cancellation, bool waitForBrowser)
    {
        string directory = Path.Combine(DataDirectory(), "children");
        string[] names;
        try { names = Sandbox().ReadDirectory(directory); }
        catch (DirectoryNotFoundException) { return; }
        catch (FileNotFoundException) { return; }
        // Validate the complete immediate set before constructing any child.
        // Registration is local to each child, not inferred from arbitrary History.
        var candidates = new List<ChildRestartState>();
        foreach (string name in names)
        {
            cancellation.ThrowIfCancellationRequested();
            if (!Sandbox().IsDirectory(Path.Combine(directory, name))) { continue; }
            ChildRestartState? state = ReadChildRestart(name);
            if (state is not null && FindChild(name) is null) { candidates.Add(state); }
        }
        var restored = new List<Agent>();
        try
        {
            foreach (ChildRestartState state in candidates)
            {
                Agent child = await CreateChildCoreAsync(state.ID, state.Model,
                    state.Schema, state, cancellation, waitForBrowser).ConfigureAwait(false);
                restored.Add(child);
                await child.RestoreChildrenAsync(cancellation, waitForBrowser).ConfigureAwait(false);
            }
            // No restored timer may start work before its descendant tree exists.
            for (int index = 0; index < restored.Count; index++)
            {
                await restored[index].RestoreChildGoalAsync(candidates[index], cancellation).ConfigureAwait(false);
            }
        }
        catch (Exception failure)
        {
            var errors = new List<Exception> { failure };
            foreach (Agent child in restored)
            {
                try { child.Destroy(); }
                catch (Exception cleanup) { errors.Add(cleanup); }
            }
            if (errors.Count > 1) { throw new AggregateException("Child restoration and cleanup failed.", errors); }
            throw;
        }
    }

    private partial async Task RestoreChildGoalAsync(ChildRestartState restart, CancellationToken cancellation)
    {
        GoalConfiguration goal = restart.Goal;
        if (goal.File.Length == 0) { return; }
        try
        {
            if (!GoalAvailable())
            {
                if (SkillManager() is null || !ToolEnabled("load_skill"))
                {
                    throw new InvalidOperationException("Current runtime has no skill-loading authority.");
                }
                ToolResultPart loaded = await DispatchAsync(new ToolCallPart(
                    Guid.NewGuid().ToString("N"), Configuration().Target(), "load_skill",
                    JsonSerializer.SerializeToElement(new { name = "goals" })), cancellation).ConfigureAwait(false);
                if (loaded.IsError || !GoalAvailable())
                {
                    throw new InvalidOperationException("Current runtime does not grant goals.");
                }
            }
            await GoalControlAsync(new GoalUpdate(goal.File,
                goal.Enabled && !restart.Submitted, goal.IdleSeconds), cancellation).ConfigureAwait(false);
        }
        catch (Exception error) when (error is IOException or UnauthorizedAccessException or ArgumentException or InvalidOperationException)
        {
            const string diagnostic = "Restored child goal is unavailable under current runtime authority. Reminders are off; inspect and explicitly re-enable.";
            lock (ingress)
            {
                var disabled = goal with { Enabled = false, LastError = diagnostic };
                SaveChildRestart(goal: disabled);
                goalConfiguration = disabled;
            }
            History().Append(new Event(EventKind.ErrorOccurred, error: diagnostic));
        }
    }
}
