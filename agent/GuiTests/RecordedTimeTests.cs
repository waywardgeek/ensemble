using System.Collections.Immutable;
using System.Text.Json;
using Ensemble;
using EnsembleGui;

internal static class RecordedTimeTests
{
    public static void Run()
    {
        var time = new DateTimeOffset(2026, 11, 5, 9, 10, 11, TimeSpan.FromHours(-8));
        var call = new ToolCallPart("call", new Configuration(Vendor.Anthropic, "model", "http://127.0.0.1:1").Target(),
            "read_file", JsonSerializer.SerializeToElement(new { path = "synthetic" }));
        var context = new Context([
            new Entry(1, Actor.Human, [new TextPart("question")], time: time),
            new Entry(2, Actor.Agent, [call], time: time.AddSeconds(1)),
            new Entry(4, Actor.Tool, [new ToolResultPart("call", [new TextPart("answer")])], time: time.AddSeconds(3), observationDuration: TimeSpan.FromTicks(1234567))
        ]);
        var started = new Event(EventKind.ToolStarted, seq: 3, time: time.AddSeconds(2),
            callID: "call", name: "read_file", args: call.Args);
        var projection = Gui.Project(new HistorySnapshot([started], context, 4), "agent");
        foreach (var frame in projection.Frames.Where(frame => frame["type"]!.GetValue<string>() is
            "part_final" or "tool_dispatched" or "tool_finished"))
        {
            long seq = frame["seq"]!.GetValue<long>();
            string expected = time.AddSeconds(seq - 1).ToString("O");
            if (frame["recorded_at"]?.GetValue<string>() != expected)
                throw new Exception("Projection lost original time at sequence " + seq);
        }
        var snapshotOnly = Gui.Project(new HistorySnapshot([], context, 4), "agent");
        if (snapshotOnly.Frames.Single(frame => frame["type"]!.GetValue<string>() == "tool_finished")
            ["recorded_at"]?.GetValue<string>() != time.AddSeconds(3).ToString("O"))
            throw new Exception("Result time depended on diagnostic tail");
        if (snapshotOnly.Frames.Single(frame => frame["type"]!.GetValue<string>() == "tool_finished")
            ["observation_ms"]?.GetValue<double>() != 123.4567)
            throw new Exception("Snapshot-only result lost duration");
        var legacy = Gui.Project(new HistorySnapshot([], new Context([
            new Entry(1, Actor.Human, [new TextPart("legacy")])]), 1), "agent");
        if (legacy.Frames.Single(frame => frame["type"]!.GetValue<string>() == "part_final")["recorded_at"] is not null)
            throw new Exception("Legacy projection invented current time");
        Console.WriteLine("PASS original recorded times in finalized, dispatch, result and snapshot-only GUI projection; legacy unknown");
    }
}
