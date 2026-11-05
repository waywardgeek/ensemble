// External Chapter 2 supplement: immutable facts, codec and pure reduction only.
// No candidate test assembly, reflection, Jobs, GUI, provider or private data.
using System.Globalization;
using System.Security.Cryptography;
using System.Text.Json;
using Ensemble;

if (args.Length != 0) return BoundaryChecks.Run(args);

var checks = new List<string>();
void Require(bool condition, string name)
{
    if (!condition) throw new InvalidOperationException(name);
    checks.Add(name);
}
void Refuse(Action action, string name)
{
    try { action(); }
    catch (ArgumentException) { checks.Add(name); return; }
    throw new InvalidOperationException(name);
}
bool Exact(DateTimeOffset? actual, DateTimeOffset? expected) =>
    actual is null ? expected is null : expected is { } value && actual.Value.EqualsExact(value);
try
{
    string nonce = Guid.NewGuid().ToString("N");
    DateTimeOffset recorded = DateTimeOffset.Parse("2004-03-02T01:02:03.4567890+05:30", CultureInfo.InvariantCulture);
    TimeSpan elapsed = TimeSpan.FromTicks(RandomNumberGenerator.GetInt32(1, 9_999_999));
    Event Result(long seq, DateTimeOffset? time, TimeSpan? duration) => new(
        EventKind.ToolReturned, seq: seq, time: time, callID: "call-" + seq + nonce,
        parts: [new TextPart("payload-" + seq + nonce)], observationDuration: duration);
    Event[] facts = [Result(1, recorded, elapsed), Result(2, recorded.AddMinutes(1), TimeSpan.Zero), Result(3, null, null)];
    string[] names = ["known", "zero", "unknown"];
    Context context = new();
    for (int index = 0; index < facts.Length; index++)
    {
        var fact = facts[index];
        string name = names[index];
        Require(Exact(fact.Time, index == 2 ? null : recorded.AddMinutes(index))
            && fact.ObservationDuration == (index == 0 ? elapsed : index == 1 ? TimeSpan.Zero : null), name + "-value");
        var decoded = JsonCodec.ReadEvent(JsonCodec.WriteEvent(fact));
        Require(Exact(decoded.Time, fact.Time) && decoded.ObservationDuration == fact.ObservationDuration
            && decoded.CallID == fact.CallID && JsonCodec.Text(decoded.Parts) == "payload-" + fact.Seq + nonce,
            name + "-codec");
        var before = context;
        context = History.Reduce(context, decoded);
        var entry = context.Dialogue.Last();
        Require(context.Dialogue.Length == index + 1 && before.Dialogue.Length == index
            && entry.Seq == fact.Seq && entry.Actor == Actor.Tool
            && entry.Parts.OfType<ToolResultPart>().Single().CallID == fact.CallID
            && Exact(entry.Time, fact.Time) && entry.ObservationDuration == fact.ObservationDuration,
            name + "-reduction");
    }
    var numbered = facts[0].Numbered(42, recorded.AddHours(1));
    Require(numbered.Seq == 42 && Exact(numbered.Time, recorded.AddHours(1))
        && numbered.ObservationDuration == elapsed && facts[0].Seq == 1
        && Exact(facts[0].Time, recorded), "numbering-retains-duration-without-mutation");
    Refuse(() => Result(4, recorded, TimeSpan.FromTicks(-1)), "negative-duration-refused");
    Refuse(() => new Event(EventKind.MessageReceived, actor: Actor.Human,
        parts: [new TextPart(nonce)], observationDuration: elapsed), "nonresult-duration-refused");
    foreach (var level in new[] { Redaction.Result, Redaction.Tool, Redaction.Dialogue })
    {
        var filtered = History.Reduce(context, new Event(EventKind.Redacted, seq: 4,
            time: recorded.AddDays(1), redactFrom: 1, redactTo: 1, level: level, reason: "fixture"));
        var original = filtered.Dialogue.Single(entry => entry.Seq == 1);
        Require(Exact(original.Time, recorded) && original.ObservationDuration == elapsed,
            "redaction-" + level + "-preserves-timing");
        Require(JsonCodec.Text(context.Dialogue[0].Parts).Contains(nonce, StringComparison.Ordinal),
            "redaction-" + level + "-preserves-original-context");
    }
    DateTimeOffset summaryTime = recorded.AddDays(2);
    var summarized = History.Reduce(context, new Event(EventKind.Redacted, seq: 5, time: summaryTime,
        redactFrom: 1, redactTo: 2, level: Redaction.Summary, parts: [new TextPart("summary-" + nonce)]));
    var summary = summarized.Dialogue[0];
    Require(summarized.Dialogue.Length == 2 && summary.Seq == 1 && summary.Actor == Actor.System
        && JsonCodec.Text(summary.Parts) == "summary-" + nonce && Exact(summary.Time, summaryTime)
        && summary.ObservationDuration is null, "summary-uses-replacement-fact-time");
    Require(summarized.Dialogue[1].Time is null && summarized.Dialogue[1].ObservationDuration is null,
        "summary-preserves-unselected-unknown");
    Console.WriteLine(JsonSerializer.Serialize(new { contract = "chapter2-timing-values-v1", passed = true, checks,
        scope = "Supplement only: no Chapter 2 score, persistence restart, provider or student certification." }));
    return 0;
}
catch (Exception error)
{
    Console.WriteLine(JsonSerializer.Serialize(new { contract = "chapter2-timing-values-v1", passed = false, checks, failure = error.Message }));
    return 1;
}
