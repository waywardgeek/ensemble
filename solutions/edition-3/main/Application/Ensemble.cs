using System.Collections.Immutable;
namespace Ensemble;

public sealed partial class Ensemble
{
    public partial Ensemble(TextWriter diagnostics, TextWriter? apiTrace, string? dataDirectory)
    {
        ArgumentNullException.ThrowIfNull(diagnostics);
        if (dataDirectory is not null)
        {
            ArgumentException.ThrowIfNullOrWhiteSpace(dataDirectory);
            this.dataDirectory = Path.GetFullPath(dataDirectory);
        }
        this.diagnostics = diagnostics;
        this.apiTrace = apiTrace;
        agents = [];
    }

    public partial void AttachDiagnostics(TextWriter diagnostics, TextWriter? apiTrace)
    {
        ArgumentNullException.ThrowIfNull(diagnostics);
        lock (traceGate)
        {
            ObjectDisposedException.ThrowIf(destroyed, this);
            this.diagnostics = diagnostics;
            this.apiTrace = apiTrace;
        }
    }

    internal partial void Trace(Agent origin, string direction, string payload, IReadOnlyList<string>? exchangeSecrets)
    {
        payload = RedactDiagnostic(payload, exchangeSecrets);
        direction = RedactDiagnostic(direction, exchangeSecrets);
        string? diagnosticName = origin.Name is null ? null : RedactDiagnostic(origin.Name, exchangeSecrets);
        lock (traceGate)
        {
            if (apiTrace is null)
            {
                return;
            }
            apiTrace.WriteLine(new System.Text.Json.Nodes.JsonObject
            {
                ["direction"] = direction, ["agent"] = diagnosticName, ["payload"] = payload
            }.ToJsonString());
            apiTrace.Flush();
        }
    }

    private partial string RedactDiagnostic(string message, IReadOnlyList<string>? exchangeSecrets)
    {
        // A diagnostic may mention more than its originating Agent. Resolve the
        // current configured keys through ownership, not a copied secret cache.
        // Include ordinary and HTML-safe JSON spellings in raw wire traces.
        // This is exact known-value masking, not an arbitrary encoding detector.
        var tokens = Agents()
            .Select(agent => agent.DiagnosticKey())
            .Concat(exchangeSecrets ?? Array.Empty<string>())
            .Where(key => key.Length > 0)
            .SelectMany(key => new[]
            {
                key,
                System.Text.Json.JsonEncodedText.Encode(key).ToString(),
                System.Text.Json.JsonEncodedText.Encode(key,
                    System.Text.Encodings.Web.JavaScriptEncoder.UnsafeRelaxedJsonEscaping).ToString()
            })
            .Distinct(StringComparer.Ordinal)
            .OrderByDescending(key => key.Length)
            .ToArray();
        if (tokens.Length == 0)
        {
            return message;
        }

        // Scan original text only. Longest-first breaks ties at the same offset,
        // preventing a shorter key from leaving a longer key's suffix exposed.
        // Ordinary ordinal searches need no pattern language or size-limited
        // regex compilation, and replacement markers are never scanned again.
        var redacted = new System.Text.StringBuilder();
        int offset = 0;
        while (offset < message.Length)
        {
            int next = message.Length;
            string? matched = null;
            foreach (string token in tokens)
            {
                int index = message.IndexOf(token, offset, StringComparison.Ordinal);
                if (index >= 0 && index < next)
                {
                    next = index;
                    matched = token;
                }
            }
            if (matched is null)
            {
                redacted.Append(message, offset, message.Length - offset);
                break;
            }
            redacted.Append(message, offset, next - offset);
            redacted.Append("[REDACTED]");
            offset = next + matched.Length;
        }
        return redacted.ToString();
    }

    public partial ImmutableArray<Agent> Agents()
    {
        lock (agents)
        {
            return agents.ToImmutableArray();
        }
    }
    public partial Agent? FindAgent(string name)
    {
        lock (agents)
        {
            ObjectDisposedException.ThrowIf(destroyed, this);
            ArgumentException.ThrowIfNullOrWhiteSpace(name);
            return agents.FirstOrDefault(agent => string.Equals(agent.Name, name, StringComparison.Ordinal));
        }
    }

    internal partial void AddAgent(Agent child)
    {
        lock (agents)
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
    }

    internal partial void RemoveAgent(Agent child)
    {
        lock (agents)
        {
            if (!ReferenceEquals(child.parent, this) || !agents.Contains(child))
            {
                throw new InvalidOperationException("Agent is not registered here");
            }

            agents.Remove(child);
            child.parent = null;
        }
    }

    internal partial int AllocateJobHandle()
    {
        lock (handleGate)
        {
            return nextJobHandle++;
        }
    }

    public partial void Log(string message, TextWriter? destination)
    {
        message = RedactDiagnostic(message);
        lock (traceGate)
        {
            (destination ?? diagnostics).WriteLine($"{DateTimeOffset.UtcNow:O} {message}");
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
        foreach (var child in Agents())
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
        // Agents finish first so their final transport/cleanup still has access
        // to the shared credential owner. Closing the app is not signing out.
        try
        {
            Credentials()?.Destroy();
        }
        catch (Exception ex)
        {
            failures ??= [];
            failures.Add(ex);
        }
        // Complete every ownership cascade even if one log cannot close cleanly.
        observations.Writer.TryComplete();
        if (failures is not null)
        {
            throw new AggregateException("Agent cleanup failed", failures);
        }
    }
}
