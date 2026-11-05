using System.Globalization;
using System.Text.Json;
using Ensemble;

internal static class LogReplay
{
    // Independent manifest types: do not serialize candidate Event/JobInfo values
    // as the oracle, or their codec could lose the same field on both sides.
    private sealed record ExpectedJob(int Handle, string Status, string OutputPath,
        long Bytes, bool Truncated, int? ExitCode, string Cwd);
    private sealed record Expected(string CallID, string Kind, string? Time,
        long? DurationTicks, bool IsError, string Text, ExpectedJob? Job);

    public static void Begin(History history, Action<bool, string> check)
    {
        check(new[] { "events.jsonl", "dump.jsonl", "expected.json" }.All(path => !File.Exists(path)),
            "fresh-observation-writer-workspace");
        history.StartLog(Path.GetFullPath("events.jsonl"));
    }

    public static void Write(History history, Event[] facts, Action<bool, string> check)
    {
        var expected = facts.Select(fact => new Expected(fact.CallID, fact.Kind.ToString(),
            fact.Time?.ToString("O", CultureInfo.InvariantCulture), fact.ObservationDuration?.Ticks,
            fact.IsError, JsonCodec.Text(fact.Parts), fact.Job is { } job
                ? new ExpectedJob(job.Handle, job.Status, job.OutputPath, job.Bytes,
                    job.Truncated, job.ExitCode, job.Cwd) : null)).ToArray();
        check(expected.Length == 4 && expected.All(value => value.Time is not null),
            "actual-observations-have-recorded-time");
        using (var dump = new StreamWriter("dump.jsonl")) history.Dump(dump);
        // Both files are compared before destruction adds any shutdown fact.
        check(File.ReadAllText("events.jsonl") == File.ReadAllText("dump.jsonl"),
            "observation-dump-matches-live-log");
        File.WriteAllText("expected.json", JsonSerializer.Serialize(expected));
        check(File.Exists("expected.json"), "observation-manifest-written");
    }

    private static bool Matches(Event fact, Expected expected)
    {
        bool timeMatches = expected.Time is null ? fact.Time is null : fact.Time is { } actual
            && actual.EqualsExact(DateTimeOffset.Parse(expected.Time, CultureInfo.InvariantCulture));
        bool jobMatches = expected.Job is null ? fact.Job is null : fact.Job is { } job
            && job.Handle == expected.Job.Handle && job.Status == expected.Job.Status
            && job.OutputPath == expected.Job.OutputPath && job.Bytes == expected.Job.Bytes
            && job.Truncated == expected.Job.Truncated && job.ExitCode == expected.Job.ExitCode
            && job.Cwd == expected.Job.Cwd;
        return fact.CallID == expected.CallID && fact.Kind.ToString() == expected.Kind
            && timeMatches && fact.ObservationDuration?.Ticks == expected.DurationTicks
            && fact.IsError == expected.IsError && JsonCodec.Text(fact.Parts) == expected.Text
            && jobMatches;
    }

    private static Event Find(IEnumerable<Event> facts, Expected expected) => facts.Single(fact =>
        fact.CallID == expected.CallID && (fact.Kind == EventKind.ToolReturned
            || fact.Kind.ToString() == "ToolObserved"));

    private static Event[] Decode(string text)
    {
        var facts = new List<Event>();
        foreach (var line in text.Split('\n', StringSplitOptions.RemoveEmptyEntries))
        {
            using var document = JsonDocument.Parse(line);
            // A log may contain a version header and unrelated lifecycle facts.
            // Preserve the real result events, not a hand-authored fixture stream.
            if (document.RootElement.TryGetProperty("tool", out _))
                facts.Add(JsonCodec.ReadEvent(line));
        }
        return facts.ToArray();
    }

    public static int Read()
    {
        var checks = new List<string>();
        void Check(bool condition, string name)
        {
            if (!condition) throw new InvalidOperationException(name);
            checks.Add(name);
        }
        using var diagnostics = new StringWriter();
        var application = new Ensemble.Ensemble(diagnostics);
        try
        {
            // There is no seed/repair fallback. The writer must have exited and
            // supplied these files before this independent process starts.
            var expected = JsonSerializer.Deserialize<Expected[]>(File.ReadAllText("expected.json"))
                ?? throw new InvalidDataException("Missing observation manifest");
            Check(expected.Length == 4 && expected.Select(value => value.CallID).Distinct().Count() == 4
                && expected.All(value => value.Time is not null)
                && expected.Take(3).All(value => value.DurationTicks is > 0)
                && expected[3].DurationTicks is null
                && expected[0].Job?.Status == "running" && expected[1].Job?.Status == "running"
                && expected[2].Job?.Status == "done"
                && expected[0].Job?.Handle == expected[1].Job?.Handle
                && expected[1].Job?.Handle == expected[2].Job?.Handle,
                "actual-observation-manifest-shape");
            foreach (string path in new[] { "events.jsonl", "dump.jsonl" })
            {
                var owner = new Agent(application, new Configuration(Vendor.Anthropic, "local",
                    "http://127.0.0.1:1", "synthetic-unused"));
                try
                {
                    var canonical = Decode(File.ReadAllText(path));
                    for (int index = 0; index < expected.Length; index++)
                        Check(Matches(Find(canonical, expected[index]), expected[index]),
                            path + "-canonical-observation-" + index);
                    owner.History().Load(Path.GetFullPath(path));
                    var loaded = owner.History().Events();
                    for (int index = 0; index < expected.Length; index++)
                        Check(Matches(Find(loaded, expected[index]), expected[index]),
                            path + "-loaded-observation-" + index);
                    using var dump = new StringWriter();
                    owner.History().Dump(dump);
                    var redumped = Decode(dump.ToString());
                    Check(expected.All(value => Matches(Find(redumped, value), value)),
                        path + "-redump-observations");
                    Check(owner.Tools().Length == 0, path + "-reader-no-tool-registration");
                }
                finally { owner.Destroy(); }
            }
            Console.WriteLine(JsonSerializer.Serialize(new { contract = "chapter4-observation-replay-v1", passed = true, checks }));
            return 0;
        }
        catch (Exception error)
        {
            Console.WriteLine(JsonSerializer.Serialize(new { contract = "chapter4-observation-replay-v1", passed = false, checks, failure = error.Message }));
            return 1;
        }
        finally { application.Destroy(); }
    }
}
