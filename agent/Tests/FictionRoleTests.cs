using System.Diagnostics;
using Ensemble;
using Application = Ensemble.Ensemble;

internal static partial class MeterTests
{
    // Python prepares actual homes; real owned Agents load them and send requests
    // to the local protocol fake. No paid provider or model is involved. Learned
    // memory below is synthetic fixture data, not a claim of model learning.
    public static async Task FictionRoleHomes(string directory)
    {
        string home = Path.Combine(directory, "fiction-role-context");
        Directory.CreateDirectory(home);
        string baselines = Path.Combine(home, "baselines");
        foreach (string role in new[] { "author", "editor" })
        {
            string source = Path.Combine(baselines, role);
            Directory.CreateDirectory(source);
            File.WriteAllText(Path.Combine(source, "SOUL.md"), role + "_ROLE_IDENTITY");
            File.WriteAllText(Path.Combine(source, "MEMORY.md"), "");
        }
        using var listener = Listen();
        var root = new Application(TextWriter.Null);
        try
        {
            var parent = MemoryAgent(root, home, listener.Prefixes.Single());
            File.WriteAllText(Path.Combine(home, "SOUL.md"), "PARENT_PRIVATE_MEMORY");
            await parent.SetMemorySettingsAsync(new MemorySettings());
            string children = Path.Combine(home, "children");
            Directory.CreateDirectory(children);
            string firstHome = Path.Combine(children, "author1");
            await PrepareFictionHome(firstHome, "author", baselines, null);
            var first = await parent.CreateChildAsync("author1", null, null, CancellationToken.None);
            await FictionMemoryTurn(first, listener, "author_ROLE_IDENTITY", "PARENT_PRIVATE_MEMORY");
            first.Destroy();

            // Simulate the role's saved learning after joined destruction. Its
            // real saved conversation remains in the old home but must not move.
            File.WriteAllText(Path.Combine(firstHome, "MEMORY.md"), "AUTHOR_LEARNED_VOICE");
            Check(File.Exists(Path.Combine(firstHome, "save.json")), "First child History was not saved");
            string secondHome = Path.Combine(children, "author2");
            await PrepareFictionHome(secondHome, "author", baselines, firstHome);
            Check(!File.Exists(Path.Combine(secondHome, "save.json")), "Preparation copied History");
            var second = await parent.CreateChildAsync("author2", null, null, CancellationToken.None);
            // The fake's previous response is a conversation-only canary. Its
            // absence proves the new request did not reuse chapter one's turns.
            await FictionMemoryTurn(second, listener, "AUTHOR_LEARNED_VOICE", "memory observed");
            Check(!second.History().View().Memories.Any(entry => entry.Text.Contains("PARENT_PRIVATE_MEMORY")),
                "Parent memory entered child corpus");
            second.Destroy();

            string editorHome = Path.Combine(children, "editor1");
            await PrepareFictionHome(editorHome, "editor", baselines, null);
            var editor = await parent.CreateChildAsync("editor1", null, null, CancellationToken.None);
            await FictionMemoryTurn(editor, listener, "editor_ROLE_IDENTITY", "AUTHOR_LEARNED_VOICE");
            Check(File.ReadAllText(Path.Combine(baselines, "author", "MEMORY.md")) == "",
                "Run changed the baseline");
        }
        finally
        {
            root.Destroy();
        }
    }

    private static async Task FictionMemoryTurn(Agent child, System.Net.HttpListener listener,
        string expected, string absent)
    {
        var turn = child.SubmitAsync([new TextPart("Inspect role memory")]);
        var request = await listener.GetContextAsync().WaitAsync(TimeSpan.FromSeconds(10));
        using var reader = new StreamReader(request.Request.InputStream);
        string wire = await reader.ReadToEndAsync();
        Check(wire.Contains(expected, StringComparison.Ordinal), "Role memory missing from actual request");
        Check(!wire.Contains(absent, StringComparison.Ordinal), "Unwanted memory or dialogue reached request");
        // This fixture tests context, not the model's ability to invoke submit.
        // Real child completion is required even for a fake provider's final text.
        child.PublishSubmission("context checked");
        await Reply(request, System.Text.Json.JsonSerializer.Serialize(new
        {
            choices = new[] { new { message = new { content = "memory observed" }, finish_reason = "stop" } }
        }));
        await turn.WaitAsync(TimeSpan.FromSeconds(10));
        await child.FlushAsync().WaitAsync(TimeSpan.FromSeconds(10));
    }

    private static async Task PrepareFictionHome(string destination, string role,
        string baselines, string? previous)
    {
        var start = new ProcessStartInfo("/usr/bin/python3")
        {
            UseShellExecute = false,
            RedirectStandardOutput = true,
            RedirectStandardError = true
        };
        start.ArgumentList.Add("-c");
        start.ArgumentList.Add("import sys; sys.path.insert(0, sys.argv[1]); "
            + "from role_homes import prepare_home; "
            + "prepare_home(sys.argv[2], sys.argv[3], sys.argv[4], sys.argv[5] or None)");
        start.ArgumentList.Add(Path.GetFullPath("agent/skills/fiction-workflow"));
        start.ArgumentList.Add(destination);
        start.ArgumentList.Add(role);
        start.ArgumentList.Add(baselines);
        start.ArgumentList.Add(previous ?? "");
        using var process = Process.Start(start) ?? throw new IOException("Could not start Python fixture");
        var output = process.StandardOutput.ReadToEndAsync();
        var errors = process.StandardError.ReadToEndAsync();
        try
        {
            await process.WaitForExitAsync().WaitAsync(TimeSpan.FromSeconds(10));
            Check(process.ExitCode == 0, "Role preparation failed: " + await errors);
        }
        finally
        {
            if (!process.HasExited)
            {
                process.Kill(entireProcessTree: true);
            }
            await process.WaitForExitAsync();
            await output;
            await errors;
        }
    }
}
