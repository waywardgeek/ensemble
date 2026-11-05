using System.Globalization;
using System.Text.Json;
using Ensemble;
using EnsembleGui;

// Public persistence and projection boundaries, with grader-owned historical facts.
internal static class TimingChecks
{
    internal static int Run()
    {
        var checks = new List<string>();
        void Require(bool condition, string name)
        {
            if (!condition) throw new InvalidOperationException(name);
            checks.Add(name);
        }
        var app = new Ensemble.Ensemble(TextWriter.Null);
        try
        {
            string workspace = Path.Combine(Directory.GetCurrentDirectory(), "workspace");
            Directory.CreateDirectory(workspace);
            var time = DateTimeOffset.Parse("2004-03-02T01:02:03.4567890+05:30", CultureInfo.InvariantCulture);
            var duration = TimeSpan.FromTicks(12345678);
            var entries = new[]
            {
                new Entry(1, Actor.Tool, [new ToolResultPart("known", [new TextPart("historical result")])],
                    time: time, observationDuration: duration),
                new Entry(2, Actor.Tool, [new ToolResultPart("zero", [new TextPart("measured zero")])],
                    time: time.AddMinutes(1), observationDuration: TimeSpan.Zero),
                new Entry(3, Actor.Tool, [new ToolResultPart("unknown", [new TextPart("legacy result")])])
            };
            var fact = new Event(EventKind.ToolReturned, seq: 1, time: time, callID: "known",
                parts: [new TextPart("historical result")], observationDuration: duration);
            var decoded = JsonCodec.ReadEvent(JsonCodec.WriteEvent(fact));
            Require(decoded.Time?.EqualsExact(time) == true && decoded.ObservationDuration == duration,
                "timing-event-codec-preserves-exact-values");
            var savedConfig = JsonSerializer.SerializeToElement(new
            {
                vendor = "anthropic", model = "offline", system_prompt = "", tools = Array.Empty<object>()
            });
            string path = Path.Combine(workspace, "timing-snapshot.json");
            File.WriteAllText(path, JsonCodec.WriteSave(new SaveDocument(new Context(entries), 3, savedConfig, [])));
            var config = new Configuration(Vendor.Anthropic, "claude-3-haiku-20240307", "http://127.0.0.1:1",
                "grader-unused", disableStreaming: true, sandboxRoot: workspace);
            var owner = new Agent(app, config, name: "timing");
            Require(owner.History().LoadSave(path), "timing-snapshot-actually-loaded");
            var context = owner.History().View();
            Require(context.Dialogue.Length == 3, "timing-snapshot-retains-three-results");
            Require(context.Dialogue[0].Time?.EqualsExact(time) == true &&
                context.Dialogue[0].ObservationDuration == duration, "timing-known-retained-after-load");
            Require(context.Dialogue[1].Time?.EqualsExact(time.AddMinutes(1)) == true &&
                context.Dialogue[1].ObservationDuration == TimeSpan.Zero, "timing-zero-retained-after-load");
            Require(context.Dialogue[2].Time is null && context.Dialogue[2].ObservationDuration is null,
                "timing-unknown-remains-unknown-after-load");
            var frames = Gui.Project(new HistorySnapshot([], context, 3), "timing").Frames
                .Where(frame => frame["type"]?.GetValue<string>() == "tool_finished").ToArray();
            Require(frames.Length == 3, "timing-snapshot-only-projects-three-results");
            Require(frames[0]["recorded_at"]?.GetValue<string>() == time.ToString("O") &&
                frames[0]["observation_ms"]?.GetValue<double>() == duration.TotalMilliseconds,
                "timing-known-snapshot-only-projection");
            Require(frames[1]["recorded_at"]?.GetValue<string>() == time.AddMinutes(1).ToString("O") &&
                frames[1]["observation_ms"]?.GetValue<double>() == 0,
                "timing-zero-snapshot-only-projection");
            Require(frames[2]["recorded_at"] is null && frames[2]["observation_ms"] is null,
                "timing-unknown-snapshot-only-projection");
            Console.WriteLine(JsonSerializer.Serialize(new { contract = "artifact-timing-core-v1", passed = true, checks }));
            return 0;
        }
        catch (Exception error)
        {
            Console.WriteLine(JsonSerializer.Serialize(new { contract = "artifact-timing-core-v1", passed = false,
                checks, failure = error.Message }));
            return 1;
        }
        finally { app.Destroy(); }
    }
}
