using System.Collections.Immutable;
using System.Text.Json.Nodes;
using Ensemble;
using Application = Ensemble.Ensemble;

internal static class ReductionTests
{
    private static void Check(bool condition, string message)
    {
        if (!condition)
        {
            throw new Exception(message);
        }
    }

    public static void TotalReplay()
    {
        string directory = Path.Combine(Path.GetTempPath(), "total-reducer-" + Guid.NewGuid().ToString("N"));
        Directory.CreateDirectory(directory);
        var log = new StringWriter();
        var app = new Application(log);
        try
        {
            var agent = new Agent(app, new Configuration(Vendor.OpenAI, "model-a", "http://127.0.0.1:1"));
            var initial = new Context([new Entry(1, Actor.Human, [new TextPart("keep me")]),
                new Entry(2, Actor.System, [new SkillPart("manual", "SYSTEM BODY", ["read_file"])])]);
            Event[] bad = [
                new(EventKind.SkillLoaded),
                new(EventKind.SkillLoaded, parts: [new TextPart("not a manual")]),
                new(EventKind.RuntimeConfigured, parts: [new TextPart("not configuration")]),
                new(EventKind.MessageReceived),
                new(EventKind.HintReceived),
                new(EventKind.SkillUnloadPending, name: "absent"),
                new(EventKind.Redacted, redactFrom: 2, redactTo: 1, level: Redaction.Tool),
                new(EventKind.Redacted, redactFrom: 99, redactTo: 100, level: Redaction.Summary),
                new(EventKind.Redacted, redactFrom: 1, redactTo: 2, level: Redaction.Summary,
                    parts: [new SkillPart("injected", "not a valid summary", ["write_file"])])
            ];
            foreach (var fact in bad)
            {
                Check(ReferenceEquals(initial, History.Reduce(initial, fact)), "Bad fact changed Context: " + fact.Kind);
                // Known payloads survive canonical decoding; reducer policy, not
                // permissive JSON parsing, decides whether they can apply.
                Check(ReferenceEquals(initial, History.Reduce(initial, JsonCodec.ReadEvent(JsonCodec.WriteEvent(fact.Numbered(3, DateTimeOffset.UnixEpoch))))),
                    "Codec and reducer disagree: " + fact.Kind);
            }
            string path = Path.Combine(directory, "events.jsonl");
            var facts = new List<Event> { new(EventKind.MessageReceived, actor: Actor.Human, parts: [new TextPart("first")]) };
            facts.AddRange(bad);
            facts.Add(new Event(EventKind.MessageReceived, actor: Actor.Human, parts: [new TextPart("last")]));
            File.WriteAllLines(path, facts.Select((fact, index) => JsonCodec.WriteEvent(fact.Numbered(index + 1, DateTimeOffset.UnixEpoch))));
            agent.History().Load(path);
            var restored = agent.History().View();
            Check(restored.Dialogue.SelectMany(entry => entry.Parts).OfType<TextPart>()
                .Select(part => part.Text).SequenceEqual(new[] { "first", "last" }), "Later good event lost");
            Check(log.ToString().Contains("skipped event") && !log.ToString().Contains("not configuration"),
                "Skip diagnostic absent or exposed payload");
            Check(agent.History().Snapshot().Sequence == facts.Count, "Skipped fact lost sequence boundary");

            // Snapshot-tail and first-session journal replay share the owner path.
            string save = Path.Combine(directory, "save.json");
            var config = agent.Engine().SavedConfiguration(new Context());
            File.WriteAllText(save, JsonCodec.WriteSave(new SaveDocument(new Context(), 0, config,
                facts.Select((fact, index) => fact.Numbered(index + 1, DateTimeOffset.UnixEpoch)).ToImmutableArray())));
            Check(agent.History().LoadSave(save), "Save tail not loaded");
            File.Delete(save);
            File.Copy(path, save + ".journal.jsonl");
            Check(agent.History().LoadSave(save), "Journal-only recovery not loaded");
            Check(agent.History().View().Dialogue.Length == 2, "Replay paths differ");

            var unknown = JsonNode.Parse(JsonCodec.WriteEvent(facts[0].Numbered(1, DateTimeOffset.UnixEpoch)))!.AsObject();
            unknown["type"] = "future_kind";
            File.WriteAllText(path, unknown.ToJsonString() + "\n");
            var committed = agent.History().View();
            bool refused = false;
            try
            {
                agent.History().Load(path);
            }
            catch (FormatException)
            {
                refused = true;
            }
            Check(refused && ReferenceEquals(committed, agent.History().View()), "Unknown event was not refused transactionally");
        }
        finally
        {
            app.Destroy();
            Directory.Delete(directory, true);
        }
    }
}
