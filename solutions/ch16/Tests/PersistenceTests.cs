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

    private static string ContextText(Context context)
    {
        // Compare the actual portable representation, including opaque JSON,
        // nullable usage, signatures, display identity and ephemeral state.
        var config = JsonSerializer.SerializeToElement(new { model = "x", vendor = "anthropic", system_prompt = "", tools = Array.Empty<object>() });
        return JsonNode.Parse(JsonCodec.WriteSave(new SaveDocument(context, 1000, config, [])))!["context"]!.ToJsonString();
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
            Check(document.Sequence == captured.Sequence && document.Events.Length == 4, "Wrong snapshot boundary");
            var restored = Agent(app);
            Check(restored.History().LoadSave(path), "Save not loaded");
            Check(ContextText(restored.History().View()) == expected, "Snapshot log applied twice or values lost");

            // A real additional event follows the snapshot boundary. Earlier
            // diagnostic events must be skipped, while this one is applied once.
            var tail = new Event(EventKind.HintReceived, actor: Actor.Human, seq: captured.Sequence + 1, parts: [new TextPart("tail")]);
            File.WriteAllText(path, JsonCodec.WriteSave(document with { Events = document.Events.Add(tail) }));
            restored.History().LoadSave(path);
            Check(ContextText(restored.History().View()) == ContextText(History.Reduce(captured.Context, tail)), "Tail replay differs");
            Check(restored.History().View().Dialogue.Last().Actor == Actor.Human, "Hint changed its recorded role");
            Check(restored.History().Append(new Event(EventKind.AgentPaused)).Seq == tail.Seq + 1, "Append identity regressed");

            File.WriteAllText(path, JsonCodec.WriteSave(document with { Context = null, Sequence = 0 }));
            restored.History().LoadSave(path);
            Check(ContextText(restored.History().View()) == expected, "Log-only replay differs");

            foreach (int keep in new[] { 0, 2, -1 })
            {
                source.History().LoadSave(path);
                string retainedPath = Path.Combine(directory, "retain-" + keep + ".json");
                string archive = Path.Combine(directory, "archive-" + keep + ".jsonl");
                source.History().Save(new PersistenceOptions(retainedPath, keep, archive));
                var retained = JsonCodec.ReadSave(File.ReadAllText(retainedPath));
                Check(retained.Events.Length == (keep < 0 ? 4 : keep), "Wrong diagnostic retention");
                restored.History().LoadSave(retainedPath);
                Check(ContextText(restored.History().View()) == expected, "Retention changed context");
                Check(restored.History().Append(new Event(EventKind.AgentPaused)).Seq == 5, "Retention reset sequence");
                Check(File.ReadAllText(archive).Contains("answer"), "Archive lost saved dialogue");
            }

            // Snapshot filtering keeps the surviving part's parser identity.
            var mixed = source.History().Append(new Event(EventKind.ResponseEnded, parts:
                [new TextPart("kept"), new ToolCallPart("exact-call", target, "read_file",
                    JsonSerializer.SerializeToElement(new { integer = 9007199254740993L }))])
                .WithDisplay(2, [3, 5]));
            source.History().Append(new Event(EventKind.Redacted, redactFrom: mixed.Seq,
                redactTo: mixed.Seq, level: Redaction.Tool));
            source.History().Save(new PersistenceOptions(path, 0));
            restored.History().LoadSave(path);
            var filtered = restored.History().View().Dialogue.Last();
            Check(filtered.PartIDs.SequenceEqual([3]) && filtered.Parts is [TextPart], "Redaction mismatched display IDs");
            source.History().Append(new Event(EventKind.Redacted, redactFrom: 1,
                redactTo: mixed.Seq, level: Redaction.Summary, parts: [new TextPart("durable summary")]));
            source.History().Save(new PersistenceOptions(path, 0));
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
            duplicate["log"]!.AsArray().Add(duplicate["log"]![0]!.DeepClone());
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
            badTail["log"]!.AsArray().RemoveAt(0);
            malformed.Add(badTail);
            foreach (var bad in malformed)
            {
                string bytes = bad.ToJsonString();
                File.WriteAllText(path, bytes);
                Refuses(() => restored.History().LoadSave(path));
                Check(ContextText(restored.History().View()) == live && File.ReadAllText(path) == bytes, "Failed load mutated state");
            }
            File.WriteAllText(path, second);
            Refuses(() => restored.History().Save(new PersistenceOptions(path, 0, path)));
            Refuses(() => restored.History().Save(new PersistenceOptions(path, 0, path + ".bak")));
            Check(File.ReadAllText(path) == second, "Archive alias touched recovery data");
            string blockedArchive = Path.Combine(directory, "blocked");
            File.WriteAllText(blockedArchive, "not a directory");
            long count = restored.History().Events().Length;
            Refuses(() => restored.History().Save(new PersistenceOptions(path, 0, Path.Combine(blockedArchive, "archive"))));
            Check(File.ReadAllText(path) == second && restored.History().Events().Length == count, "Failed save trimmed or published");
            Check(!File.Exists(path + ".bak"), "Archive failure rotated backup");
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
            Check(!first.StartPersistence(new PersistenceOptions(path, 0)), "Fresh startup claimed recovery");
            var manager = first.SkillManager()!;
            var origin = first.Configuration().Target();
            var load = await first.DispatchAsync(new ToolCallPart("load", origin, "load_skill", JsonSerializer.SerializeToElement(new { name = "editing" })));
            var unload = await first.DispatchAsync(new ToolCallPart("unload", origin, "unload_skill", JsonSerializer.SerializeToElement(new { name = "editing" })));
            Check(!load.IsError && !unload.IsError, "Runtime fixture did not activate and mark skill");
            first.History().Save(new PersistenceOptions(path, 0));
            Check(first.SkillManager()!.Pending("editing") && first.History().Events().IsEmpty,
                "Diagnostic retention erased pending unload state");
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
            second.History().Save(new PersistenceOptions(path, 0));
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
