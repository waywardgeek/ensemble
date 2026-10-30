using System.Collections.Immutable;
using System.Text;
using System.Text.Json.Nodes;
using Ensemble;
using Application = Ensemble.Ensemble;

internal static class MemoryDiskTests
{
    private static readonly DateTimeOffset Now = DateTimeOffset.Parse("2026-10-28T12:00:00Z");

    private static void Check(bool condition, string message)
    {
        if (!condition)
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
        catch (Exception error) when (error is ArgumentException or FormatException
            or InvalidOperationException or IOException or UnauthorizedAccessException)
        {
            return;
        }
        throw new Exception("Expected memory disk refusal");
    }

    private static void InHome(Action<Agent, string> test)
    {
        string home = Path.Combine(Path.GetTempPath(), "memory-disk-" + Guid.NewGuid().ToString("N"));
        Directory.CreateDirectory(home);
        var app = new Application(TextWriter.Null);
        try
        {
            var agent = new Agent(app, new Configuration(Vendor.OpenAI, "fixture", "http://127.0.0.1:1",
                sandboxRoot: home), dataDirectory: home);
            test(agent, home);
        }
        finally
        {
            app.Destroy();
            Directory.Delete(home, recursive: true);
        }
    }

    private static void Write(string home, string relative, string text)
    {
        string path = Path.Combine(home, relative);
        Directory.CreateDirectory(Path.GetDirectoryName(path)!);
        File.WriteAllText(path, text, new UTF8Encoding(false, true));
    }

    private static void Conversation(Agent agent)
    {
        var history = agent.History();
        history.Append(new Event(EventKind.MessageReceived, actor: Actor.Human,
            parts: [new TextPart("Preserve the exact source and its reasons. ééé")]));
        history.Append(new Event(EventKind.HandoffRecorded, parts: [new HandoffPart("Completed work and rejected ideas")]));
        history.Append(new Event(EventKind.MessageReceived, actor: Actor.Human, parts: [new TextPart("New unfinished task")]));
    }

    private static void Launch(Agent agent, MemoryCapture capture)
    {
        agent.History().Append(new Event(EventKind.CompressionLaunched, compression: capture.Launch));
    }

    private static void Commit(Agent agent, MemoryCapture capture, string summary)
    {
        var launch = capture.Launch;
        Event removal = launch.Source is null
            ? new Event(EventKind.Redacted, redactFrom: launch.From, redactTo: launch.To,
                level: Redaction.Summary, reason: "memory compression", taskID: launch.TaskID)
            : new Event(EventKind.MemoryRetired,
                retirement: new MemoryRetirement(launch.Source.Value, launch.SourceIDs), taskID: launch.TaskID);
        agent.History().Append(removal);
        agent.History().Append(new Event(EventKind.MemoryPopulated,
            memory: new MemoryEntry(launch.Destination, launch.ID, launch.Timestamp, summary), taskID: launch.TaskID));
    }

    private static void Finish(Agent agent)
    {
        var memory = agent.Memory();
        var outcome = memory.Recover(agent.History().View()) ?? throw new Exception("Missing terminal outcome");
        agent.History().Append(new Event(EventKind.CompressionFinished, outcome: outcome));
        Check(memory.Recover(agent.History().View()) is null, "Terminal receipt did not release");
    }

    private static MemoryCapture Graduation(Agent agent, string home, MemoryBand band = MemoryBand.Session)
    {
        string directory = band == MemoryBand.Session ? "memory" : band == MemoryBand.Archive8 ? "bucket-0" : "bucket-1";
        string first = band == MemoryBand.Session ? "2026-10-28-1" : "2026-10-28-1--2026-10-28-2";
        string second = band == MemoryBand.Session ? "2026-10-28-3" : "2026-10-28-3--2026-10-28-4";
        Write(home, directory + "/" + first + ".md", "first selected source");
        Write(home, directory + "/" + second + ".md", "second selected source");
        foreach (var entry in agent.Memory().Scan(agent.History().View(), band))
        {
            agent.History().Append(new Event(EventKind.MemoryPopulated, memory: entry));
        }
        return agent.Memory().CaptureGraduation(agent.History().View(), band, 10, 5, Now)
            ?? throw new Exception("Expected two-file graduation");
    }

    public static void OwnershipAndScan()
    {
        InHome((agent, home) =>
        {
            var memory = agent.Memory();
            Check(memory.Parent() == agent, "Memory missing owning parent");
            Refuses(() => new Ensemble.Memory(agent));
            Check(ReferenceEquals(agent.Memory(), memory), "Refused duplicate changed ownership");
            Check(memory.Scan(new Context(), MemoryBand.Session).IsEmpty, "Missing corpus not empty");
            Check(!Directory.Exists(Path.Combine(home, "memory")), "Read eagerly created corpus");
            Write(home, "SOUL.md", "SOUL exact\r\n");
            Write(home, "MEMORY.md", "");
            Write(home, "memory/2026-10-28-10.md", "ten");
            Write(home, "memory/2026-10-28-2.md", "two");
            Write(home, "memory/2026-10-28-1.md.retired", "retired");
            Write(home, "memory/.candidate.tmp", "hidden");
            var entries = memory.Scan(new Context(), MemoryBand.Session);
            Check(entries.Select(entry => entry.ID).SequenceEqual(["2026-10-28-2", "2026-10-28-10"]),
                "Corpus ordering used lexical numbers or included inactive files");
            Check(new Context(memories: entries).Memories.Select(entry => entry.ID).SequenceEqual(entries.Select(entry => entry.ID)),
                "Context reordered numeric corpus identities");
            Check(memory.Scan(new Context(), MemoryBand.Soul).Single().Text == "SOUL exact\r\n"
                && memory.Scan(new Context(), MemoryBand.Memory).Single().Text == "", "Stable files lost exact bytes or empty presence");
            Write(home, "memory/2026-10-28-2.md", "edited");
            Check(memory.Scan(new Context(), MemoryBand.Session)[0].Text == "edited", "Scan returned cached bytes");
            Write(home, "bucket-0/2026-10-28-1--2026-10-28-3.md", "range one");
            Write(home, "bucket-0/2026-10-28-3--2026-10-28-4.md", "overlap");
            Refuses(() => memory.Scan(new Context(), MemoryBand.Archive8));
            File.Delete(Path.Combine(home, "bucket-0/2026-10-28-3--2026-10-28-4.md"));
            Write(home, "memory/2026-10-28-01.md", "ambiguous number");
            Refuses(() => memory.Scan(new Context(), MemoryBand.Session));
            File.Delete(Path.Combine(home, "memory/2026-10-28-01.md"));
            memory.Destroy();
            Refuses(() => memory.Parent());
            Refuses(() => agent.Memory());
            Check(File.Exists(Path.Combine(home, "SOUL.md")), "Logical destruction erased corpus");
            _ = new Ensemble.Memory(agent);
        });
        var app = new Application(TextWriter.Null);
        try
        {
            var transient = new Agent(app, new Configuration(Vendor.OpenAI, "fixture", "http://127.0.0.1:1"));
            Check(transient.Memory().Parent() == transient, "Transient actor lacks Memory relationship");
            Refuses(() => transient.Memory().Scan(new Context(), MemoryBand.Soul));
            var memory = transient.Memory();
            transient.Destroy();
            Refuses(() => memory.Parent());
        }
        finally { app.Destroy(); }
    }

    public static void ConversationStaging()
    {
        InHome((agent, home) =>
        {
            Conversation(agent);
            Write(home, "memory/2026-10-28-8.md.retired", "reserved sequence");
            var memory = agent.Memory();
            var capture = memory.CaptureConversation(agent.History().View(), 1, Now)!;
            Check(capture.Launch.ID == "2026-10-28-9" && capture.Input().Contains("rejected ideas"),
                "Capture reused retired identity or lost checkpoint input");
            Launch(agent, capture);
            agent.History().Append(new Event(EventKind.MessageReceived, actor: Actor.Human, parts: [new TextPart("Later work")]));
            Check(memory.StillCurrent(capture, agent.History().View()), "Unrelated appended work invalidated capture");
            var replacedDestination = History.Reduce(agent.History().View(), new Event(EventKind.MemoryPopulated,
                memory: new MemoryEntry(capture.Launch.Destination, capture.Launch.ID, Now, "independent replay replacement")));
            Check(!memory.StillCurrent(capture, replacedDestination), "Changed replay destination accepted");
            memory.Stage(capture, "  Exact summary\r\n");
            Check(!File.Exists(Path.Combine(home, "memory/" + capture.Launch.ID + ".md")), "Candidate became active before commit");
            Refuses(() => memory.Scan(agent.History().View(), MemoryBand.Session));
            Refuses(() => memory.Stage(capture, "second candidate"));
            var abandoned = memory.Recover(agent.History().View())!;
            Check(abandoned.State == CompressionState.Abandoned && agent.History().View().Dialogue.Length >= 3,
                "Uncommitted recovery removed source");
            agent.History().Append(new Event(EventKind.CompressionFinished, outcome: abandoned));
            memory.Recover(agent.History().View());
            capture = memory.CaptureConversation(agent.History().View(), 1, Now)!;
            Launch(agent, capture);
            memory.Stage(capture, "  Exact summary\r\n");
            Commit(agent, capture, "  Exact summary\r\n");
            Check(!memory.StillCurrent(capture, agent.History().View()), "Committed task still looks launchable");
            Finish(agent);
            Check(File.ReadAllText(Path.Combine(home, "memory/" + capture.Launch.ID + ".md")) == "  Exact summary\r\n",
                "Finalization rewrote accepted summary");
            Check(!File.Exists(Path.Combine(home, ".memory-stage.json"))
                && memory.Scan(agent.History().View(), MemoryBand.Session).Length == 1, "Final receipt or inactive source leaked into scan");
        });
    }

    public static void GraduationRecovery()
    {
        foreach (var band in new[] { MemoryBand.Session, MemoryBand.Archive8, MemoryBand.Archive64 })
        {
            InHome((agent, home) =>
            {
                if (band == MemoryBand.Archive64)
                {
                    Write(home, "MEMORY.md", "existing curated reasons");
                    agent.History().Append(new Event(EventKind.MemoryPopulated,
                        memory: agent.Memory().Scan(agent.History().View(), MemoryBand.Memory).Single()));
                }
                var capture = Graduation(agent, home, band);
                var memory = agent.Memory();
                if (band == MemoryBand.Archive64)
                {
                    Check(capture.Input().Contains("existing curated reasons"), "Terminal input omitted current MEMORY.md");
                }
                Launch(agent, capture);
                Check(memory.StillCurrent(capture, agent.History().View()), "Fresh file capture refused");
                memory.Stage(capture, "replacement with reasons");
                Commit(agent, capture, "replacement with reasons");
                string sourceDirectory = band == MemoryBand.Session ? "memory" : band == MemoryBand.Archive8 ? "bucket-0" : "bucket-1";
                string destination = band == MemoryBand.Archive64 ? "MEMORY.md"
                    : (band == MemoryBand.Session ? "bucket-0/" : "bucket-1/") + capture.Launch.ID + ".md";
                // Simulate a crash after promotion and one completed rename.
                Write(home, destination, "replacement with reasons");
                string first = Path.Combine(home, sourceDirectory, capture.Sources[0].ID + ".md");
                File.Move(first, first + ".retired");
                var outcome = memory.Recover(agent.History().View())!;
                Check(outcome.State == CompressionState.Committed, "Interrupted retirement did not recover");
                Check(memory.Recover(agent.History().View())!.State == CompressionState.Committed, "Recovery not idempotent");
                Refuses(() => memory.Scan(agent.History().View(), band));
                foreach (var entry in capture.Sources)
                {
                    string path = Path.Combine(home, sourceDirectory, entry.ID + ".md");
                    Check(!File.Exists(path) && File.ReadAllText(path + ".retired") == entry.Text, "Retirement lost original bytes");
                }
                agent.History().Append(new Event(EventKind.CompressionFinished, outcome: outcome));
                memory.Recover(agent.History().View());
                Check(memory.Scan(agent.History().View(), band).IsEmpty, "Archived source resurrected");
            });
        }
    }

    public static void RefusalAndConfinement()
    {
        InHome((agent, home) =>
        {
            var capture = Graduation(agent, home);
            var memory = agent.Memory();
            Launch(agent, capture);
            string first = "memory/" + capture.Sources[0].ID + ".md";
            Write(home, first, "human changed selected file");
            Check(!memory.StillCurrent(capture, agent.History().View()), "Selected-byte edit accepted");
            Write(home, first, capture.Sources[0].Text);
            Write(home, "memory/2026-10-28-99.md", "unrelated later file");
            Check(memory.StillCurrent(capture, agent.History().View()), "Unselected later file invalidated capture");
            memory.Stage(capture, "accepted summary");
            Commit(agent, capture, "accepted summary");
            string receiptPath = Path.Combine(home, ".memory-stage.json");
            string original = File.ReadAllText(receiptPath);
            File.Delete(receiptPath);
            Refuses(() => memory.Recover(agent.History().View()));
            var damaged = JsonNode.Parse(original)!.AsObject();
            damaged["conversation"] = null;
            File.WriteAllText(receiptPath, damaged.ToJsonString());
            Refuses(() => memory.Recover(agent.History().View()));
            damaged = JsonNode.Parse(original)!.AsObject();
            damaged["sources"]![0]!["text"] = "changed captured source";
            File.WriteAllText(receiptPath, damaged.ToJsonString());
            Refuses(() => memory.Recover(agent.History().View()));
            damaged = JsonNode.Parse(original)!.AsObject();
            damaged["summary"] = "different candidate";
            File.WriteAllText(receiptPath, damaged.ToJsonString());
            Refuses(() => memory.Recover(agent.History().View()));
            Check(File.Exists(Path.Combine(home, first)) && !Directory.Exists(Path.Combine(home, "bucket-0")),
                "Mismatched candidate caused disk effects");
            File.WriteAllText(receiptPath, original.Replace("\"version\":1", "\"version\":1,\"version\":1", StringComparison.Ordinal));
            Refuses(() => memory.Recover(agent.History().View()));
            damaged = JsonNode.Parse(original)!.AsObject();
            damaged["unexpected"] = "not accepted";
            File.WriteAllText(receiptPath, damaged.ToJsonString());
            Refuses(() => memory.Recover(agent.History().View()));
            File.WriteAllText(receiptPath, original);
            Write(home, "bucket-0/" + capture.Launch.ID + ".md", "external destination edit");
            Refuses(() => memory.Recover(agent.History().View()));
            Check(File.ReadAllText(Path.Combine(home, "bucket-0/" + capture.Launch.ID + ".md")) == "external destination edit",
                "Recovery overwrote a conflicting destination");
            File.Delete(Path.Combine(home, "bucket-0/" + capture.Launch.ID + ".md"));
            Write(home, first + ".retired", "external archive edit");
            Refuses(() => memory.Recover(agent.History().View()));
            Check(File.Exists(Path.Combine(home, first)), "Conflicting archive removed live source");
            File.Delete(Path.Combine(home, first + ".retired"));
            Finish(agent);
        });
        InHome((agent, home) =>
        {
            string outside = Path.Combine(Path.GetTempPath(), "memory-outside-" + Guid.NewGuid().ToString("N"));
            File.WriteAllText(outside, "outside sentinel");
            try
            {
                File.CreateSymbolicLink(Path.Combine(home, "SOUL.md"), outside);
                Refuses(() => agent.Memory().Scan(new Context(), MemoryBand.Soul));
                Check(File.ReadAllText(outside) == "outside sentinel", "Confinement changed outside file");
                File.Delete(Path.Combine(home, "SOUL.md"));
                File.WriteAllBytes(Path.Combine(home, "SOUL.md"), [0xff, 0xfe]);
                Refuses(() => agent.Memory().Scan(new Context(), MemoryBand.Soul));
            }
            finally { File.Delete(outside); }
        });
    }

    public static void PersistedReceiptRecovery()
    {
        string home = Path.Combine(Path.GetTempPath(), "memory-restart-" + Guid.NewGuid().ToString("N"));
        Directory.CreateDirectory(home);
        string save = Path.Combine(home, "save.json");
        var config = new Configuration(Vendor.OpenAI, "fixture", "http://127.0.0.1:1", sandboxRoot: home);
        var app = new Application(TextWriter.Null);
        MemoryCapture? capture = null;
        try
        {
            var agent = new Agent(app, config, dataDirectory: home);
            agent.StartPersistence(new PersistenceOptions(save, RetainLog: 0));
            capture = Graduation(agent, home);
            Launch(agent, capture);
            agent.Memory().Stage(capture, "durable accepted result");
            Commit(agent, capture, "durable accepted result");
            agent.History().Save(new PersistenceOptions(save, RetainLog: 0));
            // Destroy never finalizes a receipt implicitly or erases the corpus.
            app.Destroy();
            app = new Application(TextWriter.Null);
            agent = new Agent(app, config, dataDirectory: home);
            Check(agent.History().LoadSave(save), "Saved History did not restore");
            agent.StartPersistence(new PersistenceOptions(save, RetainLog: 0));
            Check(agent.History().View().Compression?.Committed?.Text == "durable accepted result", "Restart lost committed replay receipt");
            Refuses(() => agent.Memory().Scan(agent.History().View(), MemoryBand.Session));
            Finish(agent);
            agent.History().Save(new PersistenceOptions(save, RetainLog: 0));
            Check(agent.History().View().Memories.Count(entry => entry.Band == MemoryBand.Archive8) == 1,
                "Recovery duplicated accepted replacement");
            Check(File.ReadAllText(Path.Combine(home, "bucket-0", capture.Launch.ID + ".md")) == "durable accepted result",
                "Restored receipt did not promote staged bytes");
        }
        finally
        {
            app.Destroy();
            Directory.Delete(home, recursive: true);
        }
    }
}
