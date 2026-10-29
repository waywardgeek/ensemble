using System.Diagnostics;
using System.Text;
using Ensemble;
using Application = Ensemble.Ensemble;

internal static class RecallCorpusTests
{
    private static Configuration Config(string root = "") =>
        new(Vendor.OpenAI, "test", "http://127.0.0.1:1", "unused", sandboxRoot: root);

    private static void Check(bool condition, string message)
    {
        if (!condition) throw new Exception(message);
    }

    private static void Refuses<T>(Action action) where T : Exception
    {
        try { action(); }
        catch (T) { return; }
        throw new Exception("Expected " + typeof(T).Name);
    }

    private static void Write(string home, string relative, string text)
    {
        string path = Path.Combine(home, relative);
        Directory.CreateDirectory(Path.GetDirectoryName(path)!);
        File.WriteAllText(path, text);
    }

    public static void Ownership(string temp)
    {
        string home = Path.Combine(temp, "recall-owner");
        Write(home, ".memory-stage.json", "deliberately unreadable stage");
        var app = new Application(TextWriter.Null);
        var first = new Agent(app, Config(), dataDirectory: home);
        var second = new Agent(app, Config());
        try
        {
            var child = first.Recaller();
            Check(child.Parent() == first, "construction did not attach both ends");
            Refuses<InvalidOperationException>(() => new Recaller(first));
            Refuses<InvalidOperationException>(() => second.RemoveRecaller(child));
            Check(first.Recaller() == child && child.Parent() == first,
                "ownership refusal changed a relationship");
            first.RemoveRecaller(child);
            Refuses<ObjectDisposedException>(() => child.Parent());
            Refuses<InvalidOperationException>(() => first.Recaller());
            first.AddRecaller(child);
            child.Destroy();
            child.Destroy();
            Refuses<InvalidOperationException>(() => first.Recaller());
            Refuses<InvalidOperationException>(() => first.AddRecaller(child));
            _ = new Recaller(first);
            Check(second.Recaller().Search("absent corpus", 0).IsEmpty, "no-home Agent read ambient files");
            Check(File.Exists(Path.Combine(home, ".memory-stage.json")), "destruction deleted corpus");
        }
        finally
        {
            // Construction must not have scanned the bad receipt. Remove this
            // fixture before the ordinary shutdown barrier attempts recovery.
            File.Delete(Path.Combine(home, ".memory-stage.json"));
            app.Destroy();
        }
        Refuses<ObjectDisposedException>(() => first.Recaller());
    }

    public static void ScanAndRefresh(string temp)
    {
        string home = Path.Combine(temp, "recall-corpus");
        foreach (string path in new[] { "SOUL.md", "MEMORY.md", "memory/new.md", "memory/old.md.retired",
                     "bucket-0/prior.md.retired", "bucket-1/prior.md", "docs/nested/guide.md",
                     "handoffs/work.md", "learnings/lesson.MD" })
        {
            Write(home, path, "## Choice\nquartz decision in " + path);
        }
        foreach (string path in new[] { "docs/ignored.txt", "docs/old.md.retired", "docs/.hidden.md",
                     "memory/.stage/candidate.md", ".compaction/candidate.md", "children/one/note.md" })
        {
            Write(home, path, "quartz forbiddenhidden");
        }
        var app = new Application(TextWriter.Null);
        var agent = new Agent(app, Config(), dataDirectory: home);
        try
        {
            var initial = agent.Recaller().Search("quartz", 0);
            Check(initial.Length == 9, "wrong eligible corpus count: " + initial.Length);
            Check(initial.Count(hit => hit.Snippet.Source == RecallSource.Memory) == 6,
                "memory active/retired sources lost");
            Check(initial.All(hit => !Path.IsPathRooted(hit.Snippet.File)), "absolute source attribution");
            Check(agent.Recaller().Search("forbiddenhidden", 0).IsEmpty, "hidden or non-source file indexed");
            Check(initial.Select(hit => (hit.Snippet.File, hit.Score)).SequenceEqual(
                agent.Recaller().Search("quartz", 0).Select(hit => (hit.Snippet.File, hit.Score))),
                "unchanged scan reordered ties or changed scores");
            var memoryScores = initial.Where(hit => hit.Snippet.Source == RecallSource.Memory)
                .Select(hit => (hit.Snippet.File, hit.Score)).ToArray();
            Write(home, "docs/another.md", "quartz quartz quartz quartz");
            var refreshed = agent.Recaller().Search("quartz", 0);
            Check(memoryScores.SequenceEqual(refreshed.Where(hit => hit.Snippet.Source == RecallSource.Memory)
                .Select(hit => (hit.Snippet.File, hit.Score))), "docs changed independent memory BM25 statistics");
            File.Delete(Path.Combine(home, "docs/another.md"));
            Check(agent.History().Snapshot().Sequence == 0, "search published History");
            string file = Path.Combine(home, "docs/nested/guide.md");
            var stamp = File.GetLastWriteTimeUtc(file);
            long length = new FileInfo(file).Length;
            File.WriteAllText(file, File.ReadAllText(file).Replace("quartz", "cobalt", StringComparison.Ordinal));
            File.SetLastWriteTimeUtc(file, stamp);
            Check(new FileInfo(file).Length == length, "equal-size fixture changed length");
            Check(agent.Recaller().Search("cobalt", 0).Single().Snippet.Source == RecallSource.Docs,
                "same-size same-timestamp edit left stale index");
            Check(agent.Recaller().Search("quartz", 0).Length == 8, "old bytes still searchable");
            File.Delete(file);
            Check(agent.Recaller().Search("cobalt", 0).IsEmpty, "deleted file remains searchable");
            Write(home, "docs/fresh.md", "cobalt newly added");
            Check(agent.Recaller().Search("cobalt", 0).Single().Snippet.File == "docs/fresh.md",
                "new file was not indexed");
            File.WriteAllBytes(Path.Combine(home, "docs/fresh.md"), [0xff]);
            Refuses<DecoderFallbackException>(() => agent.Recaller().Search("cobalt", 0));
            Write(home, "docs/fresh.md", "jasper corrected");
            Check(agent.Recaller().Search("cobalt", 0).IsEmpty
                && agent.Recaller().Search("jasper", 0).Length == 1, "failed refresh leaked stale results");
            using var cancelled = new CancellationTokenSource();
            cancelled.Cancel();
            Refuses<OperationCanceledException>(() => agent.Recaller().Search("jasper", 0, cancelled.Token));
        }
        finally { app.Destroy(); }
    }

    public static void Recovery(string temp)
    {
        string home = Path.Combine(temp, "recall-recovery");
        Write(home, "memory/old.md.retired", "quartz retained original");
        var app = new Application(TextWriter.Null);
        var agent = new Agent(app, Config(), dataDirectory: home);
        try
        {
            Check(agent.Recaller().Search("quartz", 0).Length == 1, "initial source unavailable");
            Write(home, ".memory-stage.json", "not-yet-recovered");
            Refuses<InvalidOperationException>(() => agent.Recaller().Search("quartz", 0));
            File.Delete(Path.Combine(home, ".memory-stage.json"));
            agent.History().Append(new Event(EventKind.MessageReceived, actor: Actor.Human,
                parts: [new TextPart("original source") ]));
            agent.History().Append(new Event(EventKind.HandoffRecorded, parts: [new HandoffPart("checkpoint") ]));
            agent.History().Append(new Event(EventKind.MessageReceived, actor: Actor.Human,
                parts: [new TextPart("open tail") ]));
            var selection = History.SelectConversation(agent.History().View(), 1)!;
            var launch = new CompressionLaunch("scan-hold", null, MemoryBand.Session, "replacement",
                DateTimeOffset.Parse("2026-01-01T00:00:00Z"), new string('a', 64),
                selection.Entries[0].Seq, selection.Entries[^1].Seq,
                tracePath: "children/compress-scan-hold/events.jsonl");
            agent.History().Append(new Event(EventKind.CompressionLaunched, compression: launch));
            Check(agent.Recaller().Search("quartz", 0).Length == 1, "unstaged worker hid unchanged sources");
            agent.History().Append(new Event(EventKind.Redacted, redactFrom: launch.From,
                redactTo: launch.To, level: Redaction.Summary, reason: "memory compression", taskID: launch.TaskID));
            agent.History().Append(new Event(EventKind.MemoryPopulated, taskID: launch.TaskID,
                memory: new MemoryEntry(launch.Destination, launch.ID, launch.Timestamp, "summary")));
            Refuses<InvalidOperationException>(() => agent.Recaller().Search("quartz", 0));
        }
        finally
        {
            // Finish the synthetic replay without claiming a real disk commit.
            // Otherwise actor startup would correctly reject its missing receipt.
            var pending = agent.History().View().Compression;
            if (pending is not null)
            {
                var state = pending.Committed is null ? CompressionState.Abandoned : CompressionState.Committed;
                agent.History().Append(new Event(EventKind.CompressionFinished,
                    outcome: new CompressionOutcome(pending.Launch.TaskID, state)));
            }
            app.Destroy();
        }
    }

    public static void Filesystem(string temp)
    {
        if (!OperatingSystem.IsMacOS() && !OperatingSystem.IsLinux()) return;
        string outside = Path.Combine(temp, "recall-outside");
        Write(outside, "secret.md", "quartz outsideforbidden");
        foreach (bool confined in new[] { true, false })
        {
            string home = Path.Combine(temp, "recall-links-" + confined);
            Write(home, "docs/good.md", "quartz allowed");
            Write(home, "docs/.pending/secret.md", "quartz hidden candidate");
            Directory.CreateSymbolicLink(Path.Combine(home, "docs/visible-alias"), ".pending");
            File.CreateSymbolicLink(Path.Combine(home, "docs/visible-alias.md"), ".pending/secret.md");
            Directory.CreateSymbolicLink(Path.Combine(home, "docs/loop"), ".");
            if (confined)
            {
                Directory.CreateSymbolicLink(Path.Combine(home, "docs/escape"), outside);
                File.CreateSymbolicLink(Path.Combine(home, "docs/escape.md"), Path.Combine(outside, "secret.md"));
            }
            string pipe = Path.Combine(home, "docs/pipe.md");
            var start = new ProcessStartInfo("/usr/bin/mkfifo") { UseShellExecute = false };
            start.ArgumentList.Add(pipe);
            using (var process = Process.Start(start)!)
            {
                process.WaitForExit();
                Check(process.ExitCode == 0, "mkfifo fixture failed");
            }
            var app = new Application(TextWriter.Null);
            var agent = new Agent(app, Config(confined ? home : ""), dataDirectory: home);
            try
            {
                var hits = agent.Recaller().Search("quartz", 0);
                Check(hits.Length == 1 && hits[0].Snippet.Content.Contains("allowed", StringComparison.Ordinal),
                    "recursive scan escaped root, looped, or lost ordinary file");
            }
            finally { app.Destroy(); }
        }
    }
    private static string Manual(string name, string body, bool primary = false) =>
        "---\nname: " + name + "\ndescription: Recall fixture\ntype: "
        + (primary ? "primary" : "loadable") + "\ntools: "
        + (primary ? "read_file" : "run_command") + "\n---\n" + body + "\n";

    public static void SkillDocuments(string temp)
    {
        if (!OperatingSystem.IsMacOS() && !OperatingSystem.IsLinux()) return;
        string root = Path.Combine(temp, "recall-skill-source");
        string workspace = Path.Combine(temp, "recall-skill-workspace");
        Directory.CreateDirectory(workspace);
        Write(root, "primary/SKILL.md", Manual("primary", "cobalt originalmanual", true));
        Write(root, "dormant/SKILL.md", Manual("dormant", "cobalt dormantmanual"));
        var app = new Application(TextWriter.Null);
        var agent = new Agent(app, Config(workspace), skillSpec: new SkillSpec(root, "primary"));
        var child = new Agent(app, Config(workspace));
        _ = new Tool(child, ToolKind.ReadFile);
        var manager = agent.SkillManager()!;
        var inherited = new Skills(child, manager.CaptureBootstrap());
        try
        {
            var initial = agent.Recaller().Search("cobalt", 0);
            Check(initial.Length == 2 && initial.All(hit => hit.Snippet.Source == RecallSource.Skills),
                "no-home Agent did not search its explicit external skill corpus");
            Check(initial.All(hit => hit.Snippet.File.StartsWith("skills/", StringComparison.Ordinal)
                && !Path.IsPathRooted(hit.Snippet.File)), "external source path leaked into attribution");
            Check(child.Recaller().Search("originalmanual", 0).Length == 1,
                "bootstrap lost skill directory provenance");
            Refuses<IOException>(() => agent.Sandbox().Resolve(Path.Combine(root, "primary/SKILL.md")));
            var active = manager.Active();
            var grants = manager.CatalogTools();
            long sequence = agent.History().Snapshot().Sequence;
            Write(root, "primary/SKILL.md", Manual("primary", "cobalt changedmanual", true));
            Check(agent.Recaller().Search("originalmanual", 0).IsEmpty
                && agent.Recaller().Search("changedmanual", 0).Length == 1,
                "skill refresh used cached activation text");
            Write(root, "new/SKILL.md", "cobalt newmanual");
            Check(agent.Recaller().Search("newmanual", 0).Length == 1, "new documentation not discovered");
            File.Delete(Path.Combine(root, "dormant/SKILL.md"));
            Check(agent.Recaller().Search("dormantmanual", 0).IsEmpty, "deleted documentation retained");
            Check(manager.Active().SequenceEqual(active) && manager.CatalogTools().SequenceEqual(grants)
                && !agent.ToolEnabled("run_command") && agent.History().Snapshot().Sequence == sequence,
                "document refresh changed activation, grants or History");
            File.WriteAllBytes(Path.Combine(root, "new/SKILL.md"), [0xc3, 0x28]);
            Refuses<DecoderFallbackException>(() => agent.Recaller().Search("cobalt", 0));
            Write(root, "new/SKILL.md", "cobalt repairedmanual");
            Check(agent.Recaller().Search("repairedmanual", 0).Length == 1, "failed scan poisoned later refresh");
            using (var cancelled = new CancellationTokenSource())
            {
                cancelled.Cancel();
                Refuses<OperationCanceledException>(() => manager.RecallFiles(cancelled.Token));
            }
            // Existing readers retain the selected directory identity. Replacing
            // its pathname must not retarget either already-opened descriptor.
            Directory.Move(root, root + "-moved");
            Write(root, "primary/SKILL.md", "replacementforbidden");
            Check(agent.Recaller().Search("replacementforbidden", 0).IsEmpty
                && child.Recaller().Search("changedmanual", 0).Length == 1,
                "replaced root pathname redirected an existing reader");
            manager.Destroy();
            manager.Destroy();
            Check(agent.SkillManager() is null, "destroyed Skills remained attached");
            Refuses<ObjectDisposedException>(() => manager.RecallFiles(default));
            Check(agent.Recaller().Search("cobalt", 0).IsEmpty, "removed Skills left stale source hits");
            Check(inherited.Parent() == child && child.Recaller().Search("changedmanual", 0).Length == 1,
                "destroying one Skills closed another manager's source");
        }
        finally { app.Destroy(); }
        Check(File.Exists(Path.Combine(root + "-moved", "primary/SKILL.md")), "destruction removed skill files");
    }

    public static void SkillFilesystem(string temp)
    {
        if (!OperatingSystem.IsMacOS() && !OperatingSystem.IsLinux()) return;
        string root = Path.Combine(temp, "recall-skill-links");
        string outside = Path.Combine(temp, "recall-skill-outside");
        Write(root, "primary/SKILL.md", Manual("primary", "cobalt allowedmanual", true));
        var app = new Application(TextWriter.Null);
        var agent = new Agent(app, Config(), skillSpec: new SkillSpec(root, "primary"));
        try
        {
            // Add hostile sources only after trusted skill-template construction;
            // this exercises the recall reader, not the separate bootstrap loader.
            Write(root, ".hidden/SKILL.md", "cobalt forbiddenmanual");
            Write(root, "primary/notes.md", "cobalt forbiddenmanual");
            Write(root, "lowercase/skill.md", "cobalt forbiddenmanual");
            Write(root, "web-search/SKILL.md", "cobalt forbiddenmanual");
            Write(outside, "SKILL.md", "cobalt forbiddenmanual");
            Directory.CreateSymbolicLink(Path.Combine(root, "escape"), outside);
            Directory.CreateSymbolicLink(Path.Combine(root, "hidden-alias"), ".hidden");
            Directory.CreateSymbolicLink(Path.Combine(root, "loop"), ".");
            Directory.CreateDirectory(Path.Combine(root, "file-alias"));
            File.CreateSymbolicLink(Path.Combine(root, "file-alias/SKILL.md"), Path.Combine(outside, "SKILL.md"));
            Directory.CreateDirectory(Path.Combine(root, "pipe"));
            var start = new ProcessStartInfo("/usr/bin/mkfifo") { UseShellExecute = false };
            start.ArgumentList.Add(Path.Combine(root, "pipe/SKILL.md"));
            using (var process = Process.Start(start)!)
            {
                process.WaitForExit();
                Check(process.ExitCode == 0, "skill FIFO fixture failed");
            }
            var hits = agent.Recaller().Search("cobalt", 0);
            Check(hits.Length == 1 && hits[0].Snippet.Content.Contains("allowedmanual", StringComparison.Ordinal),
                "skill scan followed a link, indexed non-manual data, or ignored web policy");
        }
        finally { app.Destroy(); }
    }
}
