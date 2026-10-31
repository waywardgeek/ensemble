using System.Collections.Immutable;
namespace Ensemble;

public sealed partial class Ensemble
{
    public partial Ensemble(TextWriter diagnostics, TextWriter? apiTrace)
    {
        this.diagnostics = diagnostics;
        this.apiTrace = apiTrace;
        agents = [];
    }

    internal partial void Trace(Agent origin, string direction, string payload)
    {
        if (apiTrace is null)
        {
            return;
        }
        string key = origin.Configuration().ApiKey;
        if (key.Length > 0)
        {
            // Both literal text and JSON-escaped text can occur in a payload.
            string escaped = System.Text.Json.JsonSerializer.Serialize(key)[1..^1];
            foreach (string token in new[] { key, escaped }.Distinct().OrderByDescending(value => value.Length))
            {
                payload = payload.Replace(token, "[REDACTED]", StringComparison.Ordinal);
            }
        }
        lock (traceGate)
        {
            apiTrace.WriteLine(new System.Text.Json.Nodes.JsonObject
            {
                ["direction"] = direction, ["agent"] = origin.Name, ["payload"] = payload
            }.ToJsonString());
            apiTrace.Flush();
        }
    }

    public partial ImmutableArray<Agent> Agents() => agents.ToImmutableArray();
    public partial Agent? FindAgent(string name)
    {
        ObjectDisposedException.ThrowIf(destroyed, this);
        ArgumentException.ThrowIfNullOrWhiteSpace(name);
        return agents.FirstOrDefault(agent => string.Equals(agent.Name, name, StringComparison.Ordinal));
    }

    internal partial void AddAgent(Agent child)
    {
        ObjectDisposedException.ThrowIf(destroyed, this);
        if (child.destroyed || child.parent is not null || agents.Contains(child))
        {
            throw new InvalidOperationException("Invalid Agent registration");
        }

        if (child.Name is not null && FindAgent(child.Name) is not null)
        {
            throw new InvalidOperationException("An Agent with that name is already owned here.");
        }
        agents.Add(child);
        child.parent = this;
    }

    internal partial void RemoveAgent(Agent child)
    {
        if (!ReferenceEquals(child.parent, this) || !agents.Contains(child))
        {
            throw new InvalidOperationException("Agent is not registered here");
        }

        agents.Remove(child);
        child.parent = null;
    }

    internal partial int AllocateJobHandle()
    {
        lock (handleGate)
        {
            return nextJobHandle++;
        }
    }

    public partial void Log(string message)
    {
        lock (diagnostics)
        {
            diagnostics.WriteLine($"{DateTimeOffset.UtcNow:O} {message}");
        }
    }
    public partial void Destroy()
    {
        if (destroyed)
        {
            return;
        }

        destroyed = true;
        // Children remove themselves from this collection. Enumerate a snapshot.
        List<Exception>? failures = null;
        foreach (var child in agents.ToArray())
        {
            try
            {
                child.Destroy();
            }
            catch (Exception ex)
            {
                failures ??= [];
                failures.Add(ex);
            }
        }
        // Complete every ownership cascade even if one log cannot close cleanly.
        observations.Writer.TryComplete();
        if (failures is not null)
        {
            throw new AggregateException("Agent cleanup failed", failures);
        }
    }
}
