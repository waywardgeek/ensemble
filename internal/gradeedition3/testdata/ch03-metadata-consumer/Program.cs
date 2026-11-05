// Public Chapter 3 supplement: retained metadata, not live download authority.
using System.Text.Json;
using System.Text.Json.Nodes;
using Ensemble;

if (args.Length == 1 && args[0] == "library") return LibraryChecks.Run();
if (args.Length == 1 && args[0] is "log-write" or "log-read") return LogChecks.Run(args[0]);
if (args.Length != 0)
{
    Console.Error.WriteLine("Expected no arguments (metadata), library, log-write, or log-read.");
    return 2;
}

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
const long limit = 64L * 1024 * 1024;
string id = Guid.NewGuid().ToString("N");
string filename = "report-" + Guid.NewGuid().ToString("N") + ".txt";
string payload = "ordinary-result-" + Guid.NewGuid().ToString("N");
bool Same(WorkProductInfo? actual, WorkProductInfo expected) => actual is not null
    && actual.ID == expected.ID && actual.Name == expected.Name && actual.OfferedBytes == expected.OfferedBytes;
Part Decode(JsonObject node) => JsonCodec.ReadPart(JsonSerializer.SerializeToElement(node));

try
{
    var info = new WorkProductInfo(id, filename, 12345);
    Require(info.ID == id && info.Name == filename && info.OfferedBytes == 12345, "value-fields");
    Require(new WorkProductInfo(id, filename, 0).OfferedBytes == 0, "zero-size");
    Require(new WorkProductInfo(id, filename, limit).OfferedBytes == limit, "exact-64mib-size");
    Refuse(() => new WorkProductInfo(id, filename, -1), "negative-size-refused");
    Refuse(() => new WorkProductInfo(id, filename, limit + 1), "oversize-refused");
    Refuse(() => new WorkProductInfo("x" + id[1..], filename, 1), "nonhex-id-refused");
    Refuse(() => new WorkProductInfo(id[..31], filename, 1), "short-id-refused");
    Refuse(() => new WorkProductInfo(id + "0", filename, 1), "long-id-refused");
    string[] badNames = ["", " ", ".", "..", "parent/file", "parent\\file", "line\nfile"];
    for (int i = 0; i < badNames.Length; i++)
        Refuse(() => new WorkProductInfo(id, badNames[i], 1), "invalid-name-" + i);

    const string callID = "metadata-call";
    var result = new ToolResultPart(callID, [new TextPart(payload)], workProduct: info);
    Require(Same(result.WorkProduct, info), "result-preserves-description");
    Refuse(() => new ToolResultPart(callID, [new TextPart(payload)], isError: true, workProduct: info), "error-result-refused");
    var wire = JsonCodec.WritePart(result);
    var metadata = wire["work_product"]!.AsObject();
    Require(metadata["id"]!.GetValue<string>() == id && metadata["name"]!.GetValue<string>() == filename
        && metadata["offered_bytes"]!.GetValue<long>() == 12345, "part-wire-vocabulary");
    Require(Decode(wire) is ToolResultPart restored && Same(restored.WorkProduct, info)
        && JsonCodec.Text(restored.Parts) == payload, "part-roundtrip");
    wire.Remove("work_product");
    Require(Decode(wire) is ToolResultPart absent && absent.WorkProduct is null, "legacy-part-absent");
    wire["work_product"] = null;
    Require(Decode(wire) is ToolResultPart nullable && nullable.WorkProduct is null, "legacy-part-null");

    var fact = new Event(EventKind.ToolReturned, seq: 1, callID: callID,
        parts: [new TextPart(payload)], workProduct: info);
    Refuse(() => new Event(EventKind.ToolReturned, callID: callID, parts: [new TextPart(payload)],
        isError: true, workProduct: info), "error-event-refused");
    Refuse(() => new Event(EventKind.MessageReceived, actor: Actor.Human,
        parts: [new TextPart(payload)], workProduct: info), "nonresult-event-refused");
    var eventWire = JsonNode.Parse(JsonCodec.WriteEvent(fact))!.AsObject();
    Require(eventWire["tool"]?["work_product"]?["id"]?.GetValue<string>() == id, "event-wire-vocabulary");
    var restoredFact = JsonCodec.ReadEvent(eventWire.ToJsonString());
    Require(Same(restoredFact.WorkProduct, info) && restoredFact.CallID == callID
        && JsonCodec.Text(restoredFact.Parts) == payload, "event-roundtrip");
    eventWire["tool"]!.AsObject().Remove("work_product");
    Require(JsonCodec.ReadEvent(eventWire.ToJsonString()).WorkProduct is null, "legacy-event-absent");
    eventWire["tool"]!["work_product"] = null;
    Require(JsonCodec.ReadEvent(eventWire.ToJsonString()).WorkProduct is null, "legacy-event-null");
    var numbered = fact.Numbered(7, DateTimeOffset.UtcNow);
    Require(Same(numbered.WorkProduct, info) && numbered.Seq == 7 && fact.Seq == 1, "numbering-preserves-description");
    var context = History.Reduce(new Context(), restoredFact);
    var retained = context.Dialogue.Single().Parts.OfType<ToolResultPart>().Single();
    Require(Same(retained.WorkProduct, info) && JsonCodec.Text(retained.Parts) == payload, "reduction-preserves-description");
    var filtered = History.Reduce(context, new Event(EventKind.Redacted, seq: 2,
        redactFrom: 1, redactTo: 1, level: Redaction.Result, reason: "public fixture"));
    Require(filtered.Dialogue.Single().Parts.OfType<ToolResultPart>().Single().WorkProduct is null,
        "result-redaction-removes-description");
    Require(Same(retained.WorkProduct, info) && JsonCodec.Text(retained.Parts) == payload,
        "redaction-preserves-original");
    var prose = new ToolResultPart(callID, [new TextPart("/tmp/" + filename + " https://example.invalid/" + id)]);
    Require(((ToolResultPart)Decode(JsonCodec.WritePart(prose))).WorkProduct is null, "pathname-prose-is-not-metadata");

    using var diagnostics = new StringWriter();
    var app = new Ensemble.Ensemble(diagnostics);
    try
    {
        foreach (var (vendor, model, surface) in new[] {
            (Vendor.Anthropic, "claude-3-haiku-20240307", Surface.Messages),
            (Vendor.OpenAI, "gpt-4o-mini", Surface.ChatCompletions),
            (Vendor.Gemini, "gemini-2.0-flash", Surface.GenerateContent) })
        {
            var configuration = new Configuration(vendor, model, "http://127.0.0.1:1", "synthetic-unused");
            Require(configuration.Target().Surface == surface, "early-route-" + surface);
            var owner = new Agent(app, configuration);
            try
            {
                var call = new ToolCallPart(callID, configuration.Target(), "metadata_probe",
                    JsonSerializer.SerializeToElement(new { argument = "argument-" + Guid.NewGuid().ToString("N") }));
                var plain = new Context([new Entry(1, Actor.Agent, [call]),
                    new Entry(2, Actor.Tool, [new ToolResultPart(callID, [new TextPart(payload)])])]);
                var described = new Context([new Entry(1, Actor.Agent, [call]), new Entry(2, Actor.Tool, [result])]);
                var baseline = owner.Engine().Render(plain);
                var actual = owner.Engine().Render(described);
                string json = actual.ToJsonString();
                Require(json.Contains(payload, StringComparison.Ordinal) && json.Contains("metadata_probe", StringComparison.Ordinal),
                    "render-payload-" + surface);
                Require(JsonNode.DeepEquals(baseline, actual), "render-independent-" + surface);
            }
            finally { owner.Destroy(); }
        }
    }
    finally { app.Destroy(); }
    Console.WriteLine(JsonSerializer.Serialize(new { contract = "chapter3-metadata-v1", passed = true, checks,
        scope = "Supplement only; no chapter score, live offers, Jobs, download or student certification." }));
    return 0;
}
catch (Exception error)
{
    Console.WriteLine(JsonSerializer.Serialize(new { contract = "chapter3-metadata-v1", passed = false, checks, failure = error.Message }));
    return 1;
}
