using System.Text.Json;
using Ensemble;
using Application = Ensemble.Ensemble;

namespace SpeechListener;

internal sealed partial class Session
{
    public partial Session(Configuration configuration, string directory)
    {
        this.directory = Path.GetFullPath(directory);
        if (Directory.Exists(this.directory) && Directory.EnumerateFileSystemEntries(this.directory).Any())
        {
            throw new ArgumentException("Listener requires a new or empty owned session directory");
        }
        Directory.CreateDirectory(this.directory);
        application = new Application(Console.Error);
        try
        {
            _ = new Agent(application, configuration, Path.Combine(this.directory, "jobs"), "listener");
            // Library Agents start empty. Do not install built-ins or a Skills
            // manager and then try to hide them in a provider projection.
            Install();
        }
        catch
        {
            application.Destroy();
            throw;
        }
    }

    public partial Agent Listener() => application.FindAgent("listener")
        ?? throw new InvalidOperationException("Listener has been destroyed");

    private partial void Install()
    {
        var definitions = new[]
        {
            ("say", "Submit literal text to the target. Refused while its previous turn is pending.", "text", "string"),
            ("listen", "Return only new admitted speech, with source labels. Wait until speech, turn completion or timeout; status distinguishes pending from idle.", "timeout_seconds", "number"),
            ("sleep", "Wait for the supplied nonnegative duration; shutdown cancels the wait.", "seconds", "number"),
            ("set_speech_speed", "Set the target's acknowledged speech rate, from 0.25 to 4. Transcript mode remains silent.", "rate", "number"),
            ("save_history", "Save typed submissions and heard speech as JSONL, not this Agent's conversation.", "path", "string")
        };
        foreach (var (name, description, parameter, type) in definitions)
        {
            var schema = JsonSerializer.SerializeToElement(new
            {
                type = "object", properties = new Dictionary<string, object>
                {
                    [parameter] = new { type }
                },
                required = new[] { parameter }, additionalProperties = false
            });
            _ = new Tool(Listener(), name, description, schema,
                (job, args, cancellation) => ExecuteAsync(job, name, args, cancellation),
                waitForCompletion: true);
        }
    }

    private static partial double Seconds(JsonElement arguments, string name)
    {
        double value = arguments.GetProperty(name).GetDouble();
        if (!double.IsFinite(value) || value < 0 || value > TimeSpan.MaxValue.TotalSeconds / 2)
        {
            throw new ArgumentException("Duration must be finite, nonnegative and representable");
        }
        return value;
    }

    private partial async Task<string> ExecuteAsync(Job job, string name, JsonElement arguments, CancellationToken cancellation)
    {
        if (job.Parent().Parent() != Listener())
        {
            throw new InvalidOperationException("Foreign listener job");
        }
        if (name == "sleep")
        {
            // Slice long waits only to respect Task.Delay's platform range. This
            // is not a policy cap; every slice still observes Job cancellation.
            double remaining = Seconds(arguments, "seconds");
            while (remaining > 0)
            {
                double slice = Math.Min(remaining, 86400);
                await Task.Delay(TimeSpan.FromSeconds(slice), cancellation);
                remaining -= slice;
            }
            return "Sleep completed";
        }
        await operations.WaitAsync(cancellation);
        try
        {
            bool pending = await CollectAsync(cancellation);
            switch (name)
            {
                case "say":
                    string text = arguments.GetProperty("text").GetString()!;
                    if (string.IsNullOrWhiteSpace(text)) throw new ArgumentException("Say requires nonempty text");
                    if (pending) throw new ArgumentException("Target turn is still pending; listen first");
                    // Set pending before sending, so a fast response cannot race
                    // the acknowledgement. Only the page's real turn_ended clears it.
                    string submit = $$"""
                        (() => {
                            if (socket.readyState !== 1) throw Error('Disconnected');
                            window._listenerPending = true;
                            const input = document.querySelector('#prompt');
                            input.value = {{JsonSerializer.Serialize(text)}};
                            input.dispatchEvent(new CustomEvent('input', {
                                bubbles: true, detail: { mcp: true }
                            }));
                            document.querySelector('#compose').requestSubmit();
                            return true;
                        })()
                        """;
                    await EvaluateAsync(submit, cancellation);
                    history.Add(new Entry(history.Count + 1, "typed", "listener", text));
                    return "Submitted; use listen for speech";
                case "listen":
                    var timer = System.Diagnostics.Stopwatch.StartNew();
                    double timeout = Seconds(arguments, "timeout_seconds");
                    while (true)
                    {
                        var entries = history.Skip(heardCursor).Where(entry => entry.Kind == "heard").ToArray();
                        if (entries.Length > 0 || !pending || timer.Elapsed.TotalSeconds >= timeout)
                        {
                            heardCursor = history.Count;
                            return JsonSerializer.Serialize(new
                            {
                                status = pending ? (entries.Length > 0 ? "speaking" : "timeout") : "idle",
                                pending, entries
                            });
                        }
                        await Task.Delay(TimeSpan.FromSeconds(Math.Max(0, Math.Min(0.05, timeout - timer.Elapsed.TotalSeconds))), cancellation);
                        pending = await CollectAsync(cancellation);
                    }
                case "set_speech_speed":
                    double rate = arguments.GetProperty("rate").GetDouble();
                    if (!double.IsFinite(rate) || rate < 0.25 || rate > 4) throw new ArgumentException("Rate must be between 0.25 and 4");
                    await EvaluateAsync("(() => { patchSettings({tts_speed:" + JsonSerializer.Serialize(rate) + "}); return true; })()", cancellation);
                    while (!(await EvaluateAsync("settings.tts_speed === " + JsonSerializer.Serialize(rate), cancellation)).GetBoolean())
                    {
                        await Task.Delay(50, cancellation);
                    }
                    return "Speech rate acknowledged";
                case "save_history":
                    string path = Path.GetFullPath(arguments.GetProperty("path").GetString()!, directory);
                    // Saving does not consume the hearing cursor. The transcript
                    // is a semantic projection, never Agent.History.Dump().
                    await File.WriteAllLinesAsync(path, history.Select(entry => JsonSerializer.Serialize(entry)), cancellation);
                    return "Saved semantic history";
                default:
                    throw new ArgumentException("Undeclared listener tool");
            }
        }
        finally
        {
            operations.Release();
        }
    }

    private partial async Task<bool> CollectAsync(CancellationToken cancellation)
    {
        string observation = $$"""
            (() => {
                if (socket.readyState !== 1) throw Error('Disconnected');
                return {
                    pending: window._listenerPending,
                    entries: window._ttsLog
                        .filter(e => e.kind === 'utterance' && e.seq > {{speechSequence}})
                        .map(e => ({ seq: e.seq, source: e.source, text: e.text }))
                };
            })()
            """;
        var value = await EvaluateAsync(observation, cancellation);
        foreach (var entry in value.GetProperty("entries").EnumerateArray())
        {
            speechSequence = entry.GetProperty("seq").GetInt64();
            history.Add(new Entry(history.Count + 1, "heard", entry.GetProperty("source").GetString()!, entry.GetProperty("text").GetString()!));
        }
        return value.GetProperty("pending").GetBoolean();
    }
}
