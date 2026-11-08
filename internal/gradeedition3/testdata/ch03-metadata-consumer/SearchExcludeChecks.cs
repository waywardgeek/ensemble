using System.Text.Json;
using Ensemble;

static class SearchExcludeChecks
{
    public static async Task<int> Run()
    {
        var checks = new List<string>();
        var directory = Path.Combine(Path.GetTempPath(), "search-exclude-" + Guid.NewGuid().ToString("N"));
        var app = new Ensemble.Ensemble(TextWriter.Null);
        try
        {
            string[] files = ["keep.cs", ".hidden", ".git/config", "a/node_modules/vendor.js", "a/generated/out.cs", "generated/out.cs", "a/code.min.js"];
            foreach (string file in files)
            {
                string path = Path.Combine(directory, file);
                Directory.CreateDirectory(Path.GetDirectoryName(path)!);
                File.WriteAllText(path, "MARK " + file + "\n");
            }
            // Ignore rules exist, but this contract never consults them.
            File.WriteAllText(Path.Combine(directory, ".gitignore"), "keep.cs\n");
            var owner = new Agent(app, new Configuration(Vendor.Anthropic, "local", "http://127.0.0.1:1"));
            var tool = new Tool(owner, ToolKind.SearchFiles);
            async Task<ToolResultPart> Search(object args) => await tool.ExecuteAsync(new ToolCallPart(
                Guid.NewGuid().ToString("N"), owner.Configuration().Target(), "search_files",
                JsonSerializer.SerializeToElement(args)), CancellationToken.None);
            void Check(bool value, string name)
            {
                if (!value) throw new InvalidOperationException(name);
                checks.Add(name);
            }
            var all = await Search(new { path = directory, pattern = "MARK" });
            Check(!all.IsError && files.All(file => JsonCodec.Text(all.Parts).Contains("MARK " + file)), "default-includes-hidden-and-gitignored");
            var empty = await Search(new { path = directory, pattern = "MARK", exclude = Array.Empty<string>() });
            Check(!empty.IsError && files.All(file => JsonCodec.Text(empty.Parts).Contains("MARK " + file)), "empty-preserves-default");
            var filtered = await Search(new { path = directory, pattern = "MARK", exclude = new[] { ".git/", "node_modules/", "*.min.js" } });
            string text = JsonCodec.Text(filtered.Parts);
            Check(!filtered.IsError && !text.Contains("MARK .git/") && !text.Contains("MARK a/node_modules/") && !text.Contains("MARK a/code.min.js") && text.Contains("MARK keep.cs"), "names-and-directory-pruning");
            var relative = await Search(new { path = directory, pattern = "MARK", exclude = new[] { "a/generated/" } });
            text = JsonCodec.Text(relative.Parts);
            Check(!relative.IsError && !text.Contains("MARK a/generated/") && text.Contains("MARK generated/out.cs"), "relative-subtree");
            var recursive = await Search(new { path = directory, pattern = "MARK", exclude = new[] { "**/generated/" } });
            text = JsonCodec.Text(recursive.Parts);
            Check(!recursive.IsError && !text.Contains("MARK a/generated/") && !text.Contains("MARK generated/") && text.Contains("MARK keep.cs"), "double-star-zero-and-nested");
            var direct = await Search(new { path = Path.Combine(directory, ".hidden"), pattern = "MARK", exclude = new[] { "*" } });
            Check(!direct.IsError && JsonCodec.Text(direct.Parts).Contains("MARK .hidden"), "explicit-file-override");
            Check((await Search(new { path = directory, pattern = "MARK", exclude = new[] { "../outside" } })).IsError, "malformed-glob-refused");
            Console.WriteLine(JsonSerializer.Serialize(new { contract = "chapter3-search-exclude-v1", passed = true, checks }));
            return 0;
        }
        catch (Exception error)
        {
            Console.WriteLine(JsonSerializer.Serialize(new { contract = "chapter3-search-exclude-v1", passed = false, checks, failure = error.Message }));
            return 1;
        }
        finally
        {
            app.Destroy();
            if (Directory.Exists(directory)) Directory.Delete(directory, true);
        }
    }
}
