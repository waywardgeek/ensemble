using System.Text;
using Ensemble;
using Application = Ensemble.Ensemble;

internal static class SandboxTests
{
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
        catch (Exception error) when (error is IOException or ObjectDisposedException or InvalidOperationException)
        {
            return;
        }
        throw new Exception("Expected rooted operation to refuse");
    }

    public static void Files()
    {
        string workspace = Path.Combine(Path.GetTempPath(), "ensemble-root-" + Guid.NewGuid().ToString("N"));
        string inside = Path.Combine(workspace, "inside");
        string outside = Path.Combine(workspace, "outside");
        Directory.CreateDirectory(inside);
        Directory.CreateDirectory(outside);
        string sentinel = Path.Combine(outside, "private.txt");
        File.WriteAllText(sentinel, "outside-private");
        var app = new Application(TextWriter.Null);
        try
        {
            var configuration = new Configuration(Vendor.OpenAI, "gpt-4.1", "http://127.0.0.1",
                sandboxRoot: inside);
            var agent = new Agent(app, configuration, Path.Combine(inside, "jobs"));
            var files = agent.Sandbox();
            Check(ReferenceEquals(files.Parent(), agent), "Sandbox must retain its actual Agent parent");
            Check(configuration.WithModel("gpt-4.1-mini").SandboxRoot == inside,
                "Model selection must preserve the requested boundary");
            Check(files.Root() == files.Resolve("."), "Root must be canonical");
            files.CreateDirectory("nested/deep");
            files.WriteFile("nested/deep/note", "rooted"u8);
            if (!OperatingSystem.IsWindows())
            {
                var permissions = File.GetUnixFileMode(Path.Combine(inside, "nested/deep/note"));
                Check(permissions == (UnixFileMode.UserRead | UnixFileMode.UserWrite),
                    "Created file permissions were " + permissions);
            }
            Check(Encoding.UTF8.GetString(files.ReadFile("nested/deep/note")) == "rooted", "Rooted round trip");
            using (var append = files.OpenFile("nested/deep/note", FileMode.Append, FileAccess.Write))
            {
                append.Write("+append"u8);
            }
            Check(Encoding.UTF8.GetString(files.ReadFile("nested/deep/note")) == "rooted+append", "Rooted append");
            Refuses(() => files.OpenFile("nested/deep/note", FileMode.CreateNew, FileAccess.Write).Dispose());
            files.WriteFile("nested/new", "replacement"u8);
            files.MoveFile("nested/new", "nested/deep/note");
            Check(Encoding.UTF8.GetString(files.ReadFile("nested/deep/note")) == "replacement", "Atomic replacement");

            Directory.CreateSymbolicLink(Path.Combine(inside, "inward"), "nested");
            File.CreateSymbolicLink(Path.Combine(inside, "absolute-inward"), Path.Combine(inside, "nested/deep/note"));
            Check(Encoding.UTF8.GetString(files.ReadFile("inward/deep/note")) == "replacement", "Inward directory link");
            Check(files.ReadDirectory("inward").SequenceEqual(new[] { "deep" }), "Rooted enumeration through inward link");
            Check(files.ReadDirectory("inward").SequenceEqual(new[] { "deep" }), "Independent enumeration offsets");
            Check(files.IsDirectory("inward") && !files.IsDirectory("inward/deep/note"), "Rooted classification");
            // Long UTF-8 names force multiple native directory buffers and
            // exercise the byte layout, not just an ASCII one-entry fixture.
            files.CreateDirectory("names");
            var names = Enumerable.Range(0, 160)
                .Select(index => $"{index:D3}-é-音-" + new string('x', 100)).ToArray();
            foreach (string name in names)
            {
                files.WriteFile("names/" + name, []);
            }
            Check(files.ReadDirectory("names").SequenceEqual(names), "Complete UTF-8 multi-buffer enumeration");
            Check(Encoding.UTF8.GetString(files.ReadFile("absolute-inward")) == "replacement", "Absolute inward file link");
            Directory.CreateSymbolicLink(Path.Combine(inside, "escape"), outside);
            File.CreateSymbolicLink(Path.Combine(inside, "escape-file"), sentinel);
            foreach (string path in new[] { sentinel, "../outside/private.txt", "escape/private.txt", "escape-file" })
            {
                Refuses(() => files.ReadFile(path));
                Refuses(() => files.WriteFile(path, "bad"u8));
                Refuses(() => files.DeleteFile(path));
                Refuses(() => files.MoveFile("nested/deep/note", path));
            }
            Refuses(() => files.CreateDirectory("escape/new-dir"));
            Refuses(() => files.ReadDirectory("escape"));
            Check(!Directory.Exists(Path.Combine(outside, "new-dir")), "No outside directory created");
            Check(File.ReadAllText(sentinel) == "outside-private", "Outside content unchanged");
            files.DeleteFile("nested/deep/note");
            files.DeleteFile("nested/deep/note");
            Check(!File.Exists(Path.Combine(inside, "nested/deep/note")), "Rooted deletion");
            Refuses(() => new Sandbox(agent));
            Check(ReferenceEquals(agent.Sandbox(), files), "Duplicate owner refusal must preserve both links");
            files.Destroy();
            files.Destroy();
            Refuses(() => files.Parent());
            Refuses(() => agent.Sandbox());
            Refuses(() => files.ReadFile("anything"));
            agent.Destroy();
            Check(app.Agents().Length == 0, "Explicit destruction detaches Agent");
            Refuses(() => new Agent(app, new Configuration(Vendor.OpenAI, "gpt-4.1", "http://127.0.0.1",
                sandboxRoot: Path.Combine(workspace, "missing"))));
            Check(app.Agents().Length == 0, "Failed root construction must detach Agent");

            // Full access is still an explicit, usable mode; the owner is not
            // a deny-everything substitute for the preexisting filesystem.
            var full = new Agent(app, new Configuration(Vendor.OpenAI, "gpt-4.1", "http://127.0.0.1"),
                Path.Combine(workspace, "full-jobs"));
            Check(full.Sandbox().Root() == "", "Full-access root is empty");
            Check(Encoding.UTF8.GetString(full.Sandbox().ReadFile(sentinel)) == "outside-private", "Full access positive");
        }
        finally
        {
            app.Destroy();
            Directory.Delete(workspace, recursive: true);
        }
    }

    internal static async Task ToolFiles()
    {
        string tree = Path.Combine(Path.GetTempPath(), "sandbox-tools-" + Guid.NewGuid().ToString("N"));
        string inside = Path.Combine(tree, "inside");
        string outside = Path.Combine(tree, "outside");
        Directory.CreateDirectory(inside);
        Directory.CreateDirectory(outside);
        string sentinel = Path.Combine(outside, "secret.txt");
        File.WriteAllText(sentinel, "outside-only-secret");
        var root = new Application(TextWriter.Null);
        try
        {
            var agent = new Agent(root, new Configuration(Vendor.OpenAI, "gpt-4.1", "http://127.0.0.1",
                sandboxRoot: inside), Path.Combine(inside, "jobs"));
            agent.InstallTools();
            async Task<ToolResultPart> Call(string name, object args)
            {
                return await agent.DispatchAsync(new ToolCallPart(Guid.NewGuid().ToString("N"),
                    agent.Configuration().Target(), name, System.Text.Json.JsonSerializer.SerializeToElement(args)));
            }
            static string Text(ToolResultPart result)
            {
                string report = JsonCodec.Text(result.Parts);
                return report.StartsWith("job ", StringComparison.Ordinal)
                    ? report[(report.IndexOf('\n') + 1)..] : report;
            }

            Check((await Call("read_file", new { path = sentinel })).IsError,
                "confined tool must refuse outside read before publishing bytes");
            Directory.CreateSymbolicLink(Path.Combine(inside, "escape"), outside);
            foreach (string path in new[] { "escape/secret.txt", "../outside/secret.txt", sentinel })
            {
                Check((await Call("read_file", new { path })).IsError, "tool read escape refused");
                Check((await Call("write_file", new { path, content = "changed", overwrite = true })).IsError,
                    "tool write escape refused");
                Check((await Call("edit_file", new { path, old_text = "outside", new_text = "changed" })).IsError,
                    "tool edit escape refused");
                Check((await Call("search_files", new { path, pattern = "secret" })).IsError,
                    "tool search escape refused");
            }
            Check((await Call("list_directory", new { path = "escape" })).IsError,
                "tool enumeration escape refused");
            Check(File.ReadAllText(sentinel) == "outside-only-secret", "all refusals preserve outside bytes");

            Check(!(await Call("write_file", new { path = "work/note.txt", content = "one\nneedle\n" })).IsError,
                "tool creates rooted parents and file");
            Check((await Call("write_file", new { path = "work/note.txt", content = "bad" })).IsError,
                "tool no-overwrite refusal");
            Check(!(await Call("write_file", new { path = "work/note.txt", content = "tail\n", append = true })).IsError,
                "tool append");
            Check(Text(await Call("read_file", new { path = "work/note.txt", start_line = 2, end_line = 2 })) == "needle\n",
                "rooted inclusive line range");
            Check(!(await Call("edit_file", new { path = "work/note.txt", old_text = "needle", new_text = "found" })).IsError,
                "rooted unique edit");
            Check((await Call("edit_file", new { path = "work/note.txt", old_text = "absent", new_text = "bad" })).IsError,
                "rooted missing-anchor refusal");
            Check(File.ReadAllText(Path.Combine(inside, "work/note.txt")) == "one\nfound\ntail\n",
                "rooted tool effects are real");
            Directory.CreateSymbolicLink(Path.Combine(inside, "alias"), "work");
            Check(Text(await Call("read_file", new { path = "alias/note.txt" })).Contains("found"),
                "inward symlink tool read");
            var listing = await Call("list_directory", new { path = "." });
            Check(!listing.IsError && Text(listing) == "alias/\nescape\njobs/\nwork/\n",
                "rooted sorted listing does not follow outside target for decoration");
            Directory.CreateSymbolicLink(Path.Combine(inside, "work/loop"), ".");
            var search = await Call("search_files", new { path = "work", pattern = "found", context_lines = 1 });
            Check(!search.IsError && Text(search) == "work/note.txt-1-one\nwork/note.txt:2:found\nwork/note.txt-3-tail\n",
                "recursive rooted search preserves context and terminates inward cycle");
            var rootSearch = await Call("search_files", new { path = ".", pattern = "found" });
            Check(!rootSearch.IsError && Text(rootSearch).Contains("note.txt:2:found")
                && !Text(rootSearch).Contains("outside-only-secret"),
                "recursive search skips outward links without losing inside matches");
            var missing = await Call("read_file", new { path = "missing/file" });
            Check(missing.IsError && !Text(missing).Contains(inside), "relative error does not reveal root");
            File.CreateSymbolicLink(Path.Combine(inside, "link-loop"), "link-loop");
            var loop = await Call("read_file", new { path = "link-loop" });
            Check(loop.IsError && !Text(loop).Contains(inside), "canonicalization failure does not reveal root");
        }
        finally
        {
            root.Destroy();
            Directory.Delete(tree, true);
        }
    }

    internal static async Task NonregularFile()
    {
        string tree = Path.Combine(Path.GetTempPath(), "sandbox-fifo-" + Guid.NewGuid().ToString("N"));
        Directory.CreateDirectory(tree);
        var app = new Application(TextWriter.Null);
        try
        {
            var agent = new Agent(app, new Configuration(Vendor.OpenAI, "gpt-4.1", "http://127.0.0.1",
                sandboxRoot: tree), Path.Combine(tree, "jobs"));
            string pipe = Path.Combine(tree, "pipe");
            var start = new System.Diagnostics.ProcessStartInfo("/usr/bin/mkfifo") { UseShellExecute = false };
            start.ArgumentList.Add(pipe);
            using (var create = System.Diagnostics.Process.Start(start)!)
            {
                await create.WaitForExitAsync();
                Check(create.ExitCode == 0, "owned FIFO fixture created");
            }
            Task<bool> read = Task.Run(() =>
            {
                try
                {
                    using var stream = agent.Sandbox().OpenFile("pipe", FileMode.Open, FileAccess.Read);
                    return false;
                }
                catch (IOException)
                {
                    return true;
                }
            });
            bool completed = await Task.WhenAny(read, Task.Delay(TimeSpan.FromSeconds(3))) == read;
            if (!completed)
            {
                // Rescue the deliberately failing baseline's blocking open.
                // Never leave an owned FIFO reader or the Sandbox lock stuck.
                using var rescue = new FileStream(pipe, FileMode.Open, FileAccess.ReadWrite);
                await read.WaitAsync(TimeSpan.FromSeconds(3));
            }
            Check(completed && await read, "confined FIFO open must refuse without waiting for a writer");
            agent.InstallTools();
            var call = new ToolCallPart("fifo", agent.Configuration().Target(), "read_file",
                System.Text.Json.JsonSerializer.SerializeToElement(new { path = "pipe" }));
            Check((await agent.DispatchAsync(call)).IsError, "real read tool reports non-file refusal");
            File.WriteAllText(Path.Combine(tree, "note"), "regular-positive");
            var search = new ToolCallPart("search", agent.Configuration().Target(), "search_files",
                System.Text.Json.JsonSerializer.SerializeToElement(new { path = ".", pattern = "regular-positive" }));
            var result = await agent.DispatchAsync(search);
            Check(!result.IsError && JsonCodec.Text(result.Parts).Contains("note:1:regular-positive"),
                "recursive search skips FIFO and retains regular-file results");
        }
        finally
        {
            app.Destroy();
            Directory.Delete(tree, true);
        }
    }

    public static async Task SymlinkRace()
    {
        string workspace = Path.Combine(Path.GetTempPath(), "ensemble-root-race-" + Guid.NewGuid().ToString("N"));
        string inside = Path.Combine(workspace, "inside");
        string outside = Path.Combine(workspace, "outside");
        string live = Path.Combine(inside, "live");
        string parked = Path.Combine(inside, "parked");
        Directory.CreateDirectory(live);
        Directory.CreateDirectory(outside);
        File.WriteAllText(Path.Combine(live, "note"), "inside");
        File.WriteAllText(Path.Combine(outside, "note"), "outside-private");
        File.WriteAllText(Path.Combine(outside, "outside-only"), "sentinel");
        var app = new Application(TextWriter.Null);
        using var stop = new CancellationTokenSource();
        using var ready = new ManualResetEventSlim();
        Task? swapper = null;
        try
        {
            var agent = new Agent(app, new Configuration(Vendor.OpenAI, "gpt-4.1", "http://127.0.0.1",
                sandboxRoot: inside), Path.Combine(inside, "jobs"));
            var files = agent.Sandbox();
            // A stream already opened inside remains tied to that object even
            // when its former pathname is replaced with an outside symlink.
            using (var opened = files.OpenFile("live/note", FileMode.Open, FileAccess.Read))
            {
                Directory.Move(live, parked);
                Directory.CreateSymbolicLink(live, outside);
                using var reader = new StreamReader(opened);
                Check(reader.ReadToEnd() == "inside", "Open handle followed a replacement link");
                Directory.Delete(live);
                Directory.Move(parked, live);
            }
            int swaps = 0;
            swapper = Task.Run(() =>
            {
                ready.Set();
                while (!stop.IsCancellationRequested)
                {
                    Directory.Move(live, parked);
                    RetryAfterToolCreation(() => Directory.CreateSymbolicLink(live, outside));
                    Thread.Yield();
                    Directory.Delete(live);
                    RetryAfterToolCreation(() => Directory.Move(parked, live));
                    Interlocked.Increment(ref swaps);
                }
            });
            // write_file legitimately creates a missing parent. Preserve such
            // a directory under another in-root name rather than letting the
            // attacker's own setup fail, or deleting a tool's still-open file.
            void RetryAfterToolCreation(Action operation)
            {
                while (true)
                {
                    try
                    {
                        operation();
                        return;
                    }
                    catch (IOException) when (Directory.Exists(live))
                    {
                        Check(new DirectoryInfo(live).LinkTarget is null,
                            "Unexpected link in the race fixture's creation gap");
                        Directory.Move(live, Path.Combine(inside, "created-" + Guid.NewGuid().ToString("N")));
                    }
                }
            }
            ready.Wait();
            for (int attempt = 0; attempt < 2000; attempt++)
            {
                try
                {
                    Check(!files.ReadDirectory("live").Contains("outside-only"),
                        "Racing enumeration exposed outside directory");
                    files.WriteFile("live/note", "inside"u8);
                    string text = Encoding.UTF8.GetString(files.ReadFile("live/note"));
                    Check(text == "inside", "Racing read escaped the root");
                }
                catch (IOException)
                {
                    // Concurrent rename/removal is allowed to refuse. It must
                    // never fall back to an unconfined path-based operation.
                }
            }
            // Exercise the public dispatch path during the same attack, not
            // only the primitive. Every completed result must remain private
            // to the root; transient refusals are legitimate under swapping.
            agent.InstallTools();
            var calls = new (string Name, object Arguments)[]
            {
                ("read_file", new { path = "live/note" }),
                ("write_file", new { path = "live/note", content = "inside", overwrite = true }),
                ("edit_file", new { path = "live/note", old_text = "inside", new_text = "inside" }),
                ("list_directory", new { path = "live" }),
                ("search_files", new { path = "live", pattern = "." })
            };
            for (int attempt = 0; attempt < 10; attempt++)
            {
                foreach (var call in calls)
                {
                    var result = await agent.DispatchAsync(new ToolCallPart(Guid.NewGuid().ToString("N"),
                        agent.Configuration().Target(), call.Name,
                        System.Text.Json.JsonSerializer.SerializeToElement(call.Arguments)));
                    string text = JsonCodec.Text(result.Parts);
                    Check(!text.Contains("outside-private") && !text.Contains("outside-only"),
                        "Racing file-tool result escaped the root: " + call.Name);
                }
            }
            stop.Cancel();
            await swapper;
            swapper = null;
            Check(swaps > 0, "Race fixture did not exchange directory entries");
            // A continuously changing name may legitimately refuse every
            // attempt. Stable positives before and after the race establish
            // usability without depending on thread scheduling during attack.
            Check(File.ReadAllText(Path.Combine(outside, "note")) == "outside-private", "Racing write escaped the root");
            files.WriteFile("live/note", "after-race"u8);
            Check(Encoding.UTF8.GetString(files.ReadFile("live/note")) == "after-race", "Boundary unusable after race");
        }
        finally
        {
            stop.Cancel();
            if (swapper is not null)
            {
                await swapper;
            }
            app.Destroy();
            Directory.Delete(workspace, recursive: true);
        }
    }
}
