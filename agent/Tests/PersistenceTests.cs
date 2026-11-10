using System.Collections.Immutable;
using System.Text.Json;
using System.Text.Json.Nodes;
using Ensemble;
using Application = Ensemble.Ensemble;

internal static class PersistenceTests
{
    private static void Check(bool value, string message)
    {
        if (!value)
        {
            throw new Exception(message);
        }
    }

    private static void Refuses(Action action)
    {
        try
        {
            action();
        }
        catch (Exception error) when (error is FormatException or IOException or InvalidDataException or ArgumentException or InvalidOperationException)
        {
            return;
        }
        throw new Exception("Expected refusal");
    }

    private static string DirectoryForTest()
    {
        string path = Path.Combine(Path.GetTempPath(), "persistence-" + Guid.NewGuid().ToString("N"));
        Directory.CreateDirectory(path);
        return path;
    }

    private static Agent Agent(Application app, Vendor vendor = Vendor.Anthropic, string model = "model-a")
    {
        return new Agent(app, new Configuration(vendor, model, "http://127.0.0.1:1", "SECRET-KEY"));
    }

    // Manufacture historical input explicitly: production WriteSave must never
    // regain an event-list parameter merely to support a compatibility fixture.
    internal static string LegacySave(Context? context, long sequence, JsonElement config, IEnumerable<Event> events)
    {
        var value = JsonNode.Parse(JsonCodec.WriteSave(new SaveDocument(context ?? new Context(), sequence, config)))!;
        if (context is null)
        {
            value["context"] = null;
        }
        value["log"] = new JsonArray(events.Select(fact => JsonNode.Parse(JsonCodec.WriteEvent(fact))).ToArray());
        return value.ToJsonString();
    }

    private static string ContextText(Context context)
    {
        // Compare the actual portable representation, including opaque JSON,
        // nullable usage, signatures, display identity and ephemeral state.
        var config = JsonSerializer.SerializeToElement(new { model = "x", vendor = "anthropic", system_prompt = "", tools = Array.Empty<object>() });
        return JsonNode.Parse(JsonCodec.WriteSave(new SaveDocument(context, 1000, config)))!["context"]!.ToJsonString();
    }

    public static void HandoffBoundaries()
    {
        string path = Path.Combine(DirectoryForTest(), "save.json");
        var app = new Application(TextWriter.Null);
        try
        {
            var source = Agent(app).History();
            source.StartJournal(path);
            var first = source.Append(new Event(EventKind.HandoffRecorded, parts: [new HandoffPart("first")]));
            source.Append(new Event(EventKind.AgentPaused));
            var second = source.Append(new Event(EventKind.HandoffRecorded, parts: [new HandoffPart("second")]));
            source.Save(new PersistenceOptions(path));
            var saved = JsonCodec.ReadSave(File.ReadAllText(path));
            Check(saved.LatestHandoff == second.Seq && saved.PreviousHandoff == first.Seq,
                "State-only snapshot lost handoff boundaries");
            var third = source.Append(new Event(EventKind.HandoffRecorded, parts: [new HandoffPart("third")]));
            var restored = Agent(app).History();
            Check(restored.LoadSave(path), "Handoff snapshot did not restore");
            var snapshot = restored.Snapshot();
            Check(snapshot.LatestHandoff == third.Seq && snapshot.PreviousHandoff == second.Seq,
                "Journal replay did not advance both handoff boundaries");
            var legacy = JsonCodec.ReadSave(LegacySave(source.View(), third.Seq,
                saved.Configuration, source.Events()));
            Check(legacy.LatestHandoff == third.Seq && legacy.PreviousHandoff == second.Seq,
                "Legacy decoding did not recover folded handoff markers");
            var malformed = JsonNode.Parse(File.ReadAllText(path))!;
            malformed["previous_handoff"] = second.Seq;
            Refuses(() => JsonCodec.ReadSave(malformed.ToJsonString()));
            malformed["previous_handoff"] = 0;
            malformed["latest_handoff"] = saved.Sequence + 1;
            Refuses(() => JsonCodec.ReadSave(malformed.ToJsonString()));
        }
        finally
        {
            app.Destroy();
        }
    }

    public static void RecordedTime()
    {
        var original = new DateTimeOffset(2026, 11, 5, 9, 10, 11, TimeSpan.FromHours(-8));
        var later = original.AddMinutes(3);
        var context = History.Reduce(new Context(), new Event(EventKind.MessageReceived,
            seq: 1, time: original, actor: Actor.Human, parts: [new TextPart("retained")]));
        Check(context.Dialogue.Single().Time == original, "Reduction lost recorded time");
        var config = JsonSerializer.SerializeToElement(new { model = "x", vendor = "anthropic", system_prompt = "", tools = Array.Empty<object>() });
        var saved = JsonNode.Parse(JsonCodec.WriteSave(new SaveDocument(context, 1, config)))!;
        var restored = JsonCodec.ReadSave(saved.ToJsonString()).Context!.Dialogue.Single();
        Check(restored.Time == original && restored.Time.Value.Offset == original.Offset,
            "Snapshot changed recorded time or offset");
        saved["context"]!["dialogue"]![0]!.AsObject().Remove("time");
        Check(JsonCodec.ReadSave(saved.ToJsonString()).Context!.Dialogue.Single().Time is null,
            "Legacy snapshot invented a timestamp");
        saved["context"]!["dialogue"]![0]!["time"] = "not a timestamp";
        Refuses(() => JsonCodec.ReadSave(saved.ToJsonString()));
        var filtered = History.Reduce(context, new Event(EventKind.Redacted, seq: 2,
            time: later, redactFrom: 1, redactTo: 1, level: Redaction.Tool));
        Check(filtered.Dialogue.Single().Time == original, "Filtering changed original time");
        var summarized = History.Reduce(context, new Event(EventKind.Redacted, seq: 2,
            time: later, redactFrom: 1, redactTo: 1, level: Redaction.Summary,
            parts: [new TextPart("summary")]));
        Check(summarized.Dialogue.Single().Time == later, "Summary borrowed replaced entry time");
        var handoff = History.Reduce(context, new Event(EventKind.HandoffRecorded, seq: 2,
            time: later, parts: [new HandoffPart("checkpoint")]));
        Check(handoff.Dialogue[0].Time == original && handoff.Dialogue[1].Time == later,
            "Handoff changed retained or checkpoint time");
        var measured = TimeSpan.FromTicks(1234567);
        var returned = new Event(EventKind.ToolReturned, seq: 2, time: later,
            callID: "timed", parts: [new TextPart("result")], observationDuration: measured);
        var timed = History.Reduce(context, returned);
        Check(timed.Dialogue.Last().ObservationDuration == measured, "Reduction dropped duration");
        var roundtrip = JsonCodec.ReadSave(JsonCodec.WriteSave(new SaveDocument(timed, 2, config)));
        Check(roundtrip.Context!.Dialogue.Last().ObservationDuration == measured, "Snapshot rounded duration");
        Check(returned.Numbered(3, later).ObservationDuration == measured, "Numbering dropped duration");
        Refuses(() => new Event(EventKind.ToolReturned, callID: "negative", observationDuration: TimeSpan.FromTicks(-1)));
        Refuses(() => new Event(EventKind.MessageReceived, observationDuration: measured));
        var app = new Application(TextWriter.Null);
        try
        {
            var agent = Agent(app);
            var fact = agent.History().Append(new Event(EventKind.MessageReceived,
                actor: Actor.Human, parts: [new TextPart("actual append")]));
            Check(fact.Time is not null && agent.History().View().Dialogue.Last().Time == fact.Time,
                "History-assigned time did not reach retained Context");
        }
        finally { app.Destroy(); }
    }

    public static void Replay()
    {
        string directory = DirectoryForTest();
        string path = Path.Combine(directory, "save.json");
        var app = new Application(TextWriter.Null);
        try
        {
            var source = Agent(app);
            var history = source.History();
            var target = source.Configuration().Target();
            history.Append(new Event(EventKind.MessageReceived, actor: Actor.Human,
                parts: [new TextPart("remember"), new BlobPart("image/png", new BlobRef(RefKind.Uri, "data:image/png;base64,AA=="))]));
            var sent = history.Append(new Event(EventKind.RequestSent, from: target));
            var response = new Event(EventKind.ResponseEnded, parts:
                [new TextPart("answer"), new OpaquePart(target, JsonSerializer.SerializeToElement(new { exact = 9007199254740993L }))],
                usage: new Usage(10, null, 2, 4)).WithDisplay(sent.Seq, [0, 7]);
            history.Append(response);
            history.Append(new Event(EventKind.MessageReceived, actor: Actor.System, parts: [new TextPart("one-use") ]));
            var captured = history.Snapshot();
            string expected = ContextText(captured.Context);
            history.Save(new PersistenceOptions(path));
            string original = File.ReadAllText(path);
            Check(!original.Contains("SECRET-KEY") && !original.Contains("127.0.0.1"), "Runtime secrets leaked");
            var document = JsonCodec.ReadSave(original);
            Check(document.Sequence == captured.Sequence && !JsonNode.Parse(original)!.AsObject().ContainsKey("log"), "Wrong snapshot boundary");
            var restored = Agent(app);
            Check(restored.History().LoadSave(path), "Save not loaded");
            Check(ContextText(restored.History().View()) == expected, "Snapshot log applied twice or values lost");

            // A real additional event follows the snapshot boundary. Earlier
            // diagnostic events must be skipped, while this one is applied once.
            var tail = new Event(EventKind.HintReceived, actor: Actor.Human, seq: captured.Sequence + 1, parts: [new TextPart("tail")]);
            File.WriteAllText(path, LegacySave(document.Context, document.Sequence, document.Configuration, captured.Events.Add(tail)));
            restored.History().LoadSave(path);
            Check(ContextText(restored.History().View()) == ContextText(History.Reduce(captured.Context, tail)), "Tail replay differs");
            Check(restored.History().View().Dialogue.Last().Actor == Actor.Human, "Hint changed its recorded role");
            Check(restored.History().Append(new Event(EventKind.AgentPaused)).Seq == tail.Seq + 1, "Append identity regressed");

            // Migrating an old primary must not carry its duplicate event array
            // into the backup, or lose the tail folded during legacy decoding.
            restored.History().Save(new PersistenceOptions(path));
            string migratedBackup = File.ReadAllText(path + ".bak");
            var backup = JsonCodec.ReadSave(migratedBackup);
            Check(!JsonNode.Parse(migratedBackup)!.AsObject().ContainsKey("log"),
                "Backup retained the legacy event array");
            Check(backup.Sequence == tail.Seq
                && ContextText(backup.Context) == ContextText(History.Reduce(captured.Context, tail)),
                "Backup migration lost the legacy recovery tail");

            File.WriteAllText(path, LegacySave(null, 0, document.Configuration, captured.Events));
            restored.History().LoadSave(path);
            Check(ContextText(restored.History().View()) == expected, "Log-only replay differs");

            foreach (bool audit in new[] { false, true })
            {
                source.History().LoadSave(path);
                string retainedPath = Path.Combine(directory, "audit-" + audit + ".json");
                source.History().Save(new PersistenceOptions(retainedPath, Audit: audit));
                var retained = JsonCodec.ReadSave(File.ReadAllText(retainedPath));
                Check(!JsonNode.Parse(File.ReadAllText(retainedPath))!.AsObject().ContainsKey("log"), "Snapshot duplicated events");
                restored.History().LoadSave(retainedPath);
                Check(ContextText(restored.History().View()) == expected, "Retention changed context");
                Check(restored.History().Append(new Event(EventKind.AgentPaused)).Seq == 5, "Retention reset sequence");
            }

            // Snapshot filtering keeps the surviving part's parser identity.
            var mixed = source.History().Append(new Event(EventKind.ResponseEnded, parts:
                [new TextPart("kept"), new ToolCallPart("exact-call", target, "read_file",
                    JsonSerializer.SerializeToElement(new { integer = 9007199254740993L }))])
                .WithDisplay(2, [3, 5]));
            source.History().Append(new Event(EventKind.Redacted, redactFrom: mixed.Seq,
                redactTo: mixed.Seq, level: Redaction.Tool));
            source.History().Save(new PersistenceOptions(path));
            restored.History().LoadSave(path);
            var filtered = restored.History().View().Dialogue.Last();
            Check(filtered.PartIDs.SequenceEqual([3]) && filtered.Parts is [TextPart], "Redaction mismatched display IDs");
            source.History().Append(new Event(EventKind.Redacted, redactFrom: 1,
                redactTo: mixed.Seq, level: Redaction.Summary, parts: [new TextPart("durable summary")]));
            source.History().Save(new PersistenceOptions(path));
            restored.History().LoadSave(path);
            Check(restored.History().View().Dialogue is [{ Parts: [TextPart { Text: "durable summary" }] }],
                "Summary did not survive snapshot-only restoration");
        }
        finally
        {
            app.Destroy();
        }
    }

    public static void Recovery()
    {
        string directory = DirectoryForTest();
        string path = Path.Combine(directory, "save.json");
        var diagnostics = new StringWriter();
        var app = new Application(diagnostics);
        try
        {
            var source = Agent(app);
            source.History().Append(new Event(EventKind.HintReceived, actor: Actor.System, parts: [new TextPart("first") ]));
            source.History().Save(new PersistenceOptions(path));
            string first = File.ReadAllText(path);
            source.History().Append(new Event(EventKind.HintReceived, actor: Actor.System, parts: [new TextPart("second") ]));
            source.History().Save(new PersistenceOptions(path));
            string second = File.ReadAllText(path);
            Check(File.ReadAllText(path + ".bak") == first, "Backup is not previous validated primary");
            File.WriteAllText(path, "broken SECRET-KEY");
            var restored = Agent(app);
            Check(restored.History().LoadSave(path), "Backup recovery failed");
            Check(restored.History().Snapshot().Sequence == 1, "Wrong recovery candidate");
            restored.History().Save(new PersistenceOptions(path));
            Check(File.ReadAllText(path + ".bak") == first, "Corrupt primary replaced good backup");
            Check(!diagnostics.ToString().Contains("SECRET-KEY"), "Unsafe corruption diagnostic");

            // Isolate malformed candidates: a valid backup must not mask decoder
            // rejection. Each failure preserves the already committed live state.
            File.Delete(path + ".bak");
            string live = ContextText(restored.History().View());
            var valid = JsonNode.Parse(second)!.AsObject();
            var malformed = new List<JsonObject>();
            JsonObject Variant() => valid.DeepClone().AsObject();
            var missingConfig = Variant();
            missingConfig["config"] = null;
            malformed.Add(missingConfig);
            var invalidTool = Variant();
            invalidTool["config"]!["tools"] = new JsonArray(new JsonObject { ["name"] = "incomplete" });
            malformed.Add(invalidTool);
            var negative = Variant();
            negative["as_of"] = -1;
            malformed.Add(negative);
            var duplicate = Variant();
            var legacyFact = JsonNode.Parse(JsonCodec.WriteEvent(new Event(EventKind.AgentPaused, seq: 1)))!;
            duplicate["log"] = new JsonArray(legacyFact.DeepClone(), legacyFact.DeepClone());
            malformed.Add(duplicate);
            var missing = Variant();
            missing.Remove("context");
            malformed.Add(missing);
            var phase = Variant();
            phase["context"]!["phase"] = "invented";
            malformed.Add(phase);
            var usage = Variant();
            usage["context"]!["usage"]!["input"] = -2;
            malformed.Add(usage);
            var badTail = Variant();
            badTail["as_of"] = 0;
            badTail["context"] = null;
            badTail["log"] = new JsonArray(JsonNode.Parse(JsonCodec.WriteEvent(new Event(EventKind.AgentPaused, seq: 2))));
            malformed.Add(badTail);
            foreach (var bad in malformed)
            {
                string bytes = bad.ToJsonString();
                File.WriteAllText(path, bytes);
                Refuses(() => restored.History().LoadSave(path));
                Check(ContextText(restored.History().View()) == live && File.ReadAllText(path) == bytes, "Failed load mutated state");
            }
            File.WriteAllText(path, second);
            string blocked = Path.Combine(directory, "blocked");
            File.WriteAllText(blocked, "not a directory");
            long count = restored.History().Events().Length;
            Refuses(() => restored.History().Save(new PersistenceOptions(Path.Combine(blocked, "save.json"))));
            Check(File.ReadAllText(path) == second && restored.History().Events().Length == count,
                "Failed save changed existing state");
            Check(!File.Exists(path + ".bak"), "Unrelated failed save rotated backup");
        }
        finally
        {
            app.Destroy();
        }
    }

    public static async Task Runtime()
    {
        string directory = DirectoryForTest();
        string path = Path.Combine(directory, "save.json");
        string root = SkillTests.Fixture();
        var app = new Application(TextWriter.Null);
        try
        {
            var first = new Agent(app, new Configuration(Vendor.Anthropic, "model-a", "http://127.0.0.1:1"),
                skillSpec: new SkillSpec(root, "root", new Dictionary<string, string> { ["WHO"] = "old" }));
            Check(!first.StartPersistence(new PersistenceOptions(path)), "Fresh startup claimed recovery");
            var manager = first.SkillManager()!;
            var origin = first.Configuration().Target();
            var load = await first.DispatchAsync(new ToolCallPart("load", origin, "load_skill", JsonSerializer.SerializeToElement(new { name = "editing" })));
            var unload = await first.DispatchAsync(new ToolCallPart("unload", origin, "unload_skill", JsonSerializer.SerializeToElement(new { name = "editing" })));
            Check(!load.IsError && !unload.IsError, "Runtime fixture did not activate and mark skill");
            first.History().Save(new PersistenceOptions(path));
            Check(first.SkillManager()!.Pending("editing"),
                "Saving erased pending unload state");
            var rawRestore = Agent(app);
            rawRestore.History().LoadSave(path);
            Check(rawRestore.History().View().Dialogue.SelectMany(entry => entry.Parts)
                .OfType<SkillPart>().Any(skill => skill.Name == "editing" && skill.Pending),
                "Snapshot did not preserve pending unload state");
            rawRestore.Destroy();
            first.History().Append(new Event(EventKind.HintReceived, actor: Actor.System, parts: [new TextPart("ordinary system entry")]));
            first.History().Append(new Event(EventKind.MessageReceived, actor: Actor.Human, parts: [new TextPart("remember me")]));
            await first.ShutdownAsync();
            long savedSequence = JsonCodec.ReadSave(File.ReadAllText(path)).Sequence;
            first.Destroy();
            var second = new Agent(app, new Configuration(Vendor.OpenAI, "model-b", "http://127.0.0.1:2"),
                skillSpec: new SkillSpec(root, "root", new Dictionary<string, string> { ["WHO"] = "new" }));
            Check(second.StartPersistence(new PersistenceOptions(path)), "Runtime startup did not restore");
            Check(second.SkillManager()!.Active().Length == 1 && !second.ToolEnabled("write_file"), "Old dynamic grant survived restart");
            Check(!second.SkillManager()!.Pending("editing"), "Old unload marker survived restart");
            string wire = second.Engine().Render(second.History().View()).ToJsonString();
            Check(wire.Contains("PRIMARY-new") && !wire.Contains("PRIMARY-old") && !wire.Contains("EDIT-old"), "Saved instructions overrode current runtime");
            Check(wire.Contains("ordinary system entry") && wire.Contains("remember me") && wire.Contains("model-b"), "Conversation or current model lost");
            Check(second.History().Snapshot().Sequence == savedSequence + 1 && second.History().View().Phase == Phase.Idle, "Restart boundary incorrect");
            await second.PostAsync(new AgentInput(InputKind.Pause));
            Refuses(() => second.StartPersistence(new PersistenceOptions(path)));
            // A summary range may include a retained primary skill. The summary
            // must not borrow that preserved entry's sequence identity.
            long summaryEnd = second.History().Snapshot().Sequence;
            second.History().Append(new Event(EventKind.Redacted, redactFrom: 1,
                redactTo: summaryEnd, level: Redaction.Summary,
                parts: [new TextPart("resumed conversation summary")]));
            second.History().Save(new PersistenceOptions(path));
            var summarized = JsonCodec.ReadSave(File.ReadAllText(path));
            Check(summarized.Context!.Dialogue.Select(entry => entry.Seq).Distinct().Count()
                == summarized.Context.Dialogue.Length, "Summary collided with a preserved skill identity");
            await second.ShutdownAsync();
            second.Destroy();

            // Invalid startup never arms saving, even when cleanup joins an actor.
            File.WriteAllText(path, "damaged");
            File.WriteAllText(path + ".bak", "also damaged");
            var refused = Agent(app);
            Refuses(() => refused.StartPersistence(new PersistenceOptions(path)));
            refused.Destroy();
            Check(File.ReadAllText(path) == "damaged" && File.ReadAllText(path + ".bak") == "also damaged", "Cleanup clobbered refused startup");
        }
        finally
        {
            app.Destroy();
        }
    }
}
