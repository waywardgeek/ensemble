// Current Chapter 3 tool-output supplement; historical scores are unchanged.
using System.Text.Json;
using Ensemble;

static class ReadLineChecks
{
    public static async Task<int> Run()
    {
        var checks = new List<string>();
        var directory = Path.Combine(Path.GetTempPath(), "read-lines-" + Guid.NewGuid().ToString("N"));
        Directory.CreateDirectory(directory);
        var app = new Ensemble.Ensemble(TextWriter.Null);
        try
        {
            var owner = new Agent(app, new Configuration(Vendor.Anthropic, "local", "http://127.0.0.1:1"));
            var tool = new Tool(owner, ToolKind.ReadFile);
            var path = Path.Combine(directory, "source.txt");
            async Task<string> Read(object args)
            {
                var call = new ToolCallPart(Guid.NewGuid().ToString("N"), owner.Configuration().Target(),
                    "read_file", JsonSerializer.SerializeToElement(args));
                var result = await tool.ExecuteAsync(call, CancellationToken.None);
                if (result.IsError) throw new InvalidOperationException("read_file refused planted input");
                var text = JsonCodec.Text(result.Parts);
                return text.StartsWith("job ", StringComparison.Ordinal) ? text[(text.IndexOf('\n') + 1)..] : text;
            }
            void Check(bool condition, string name)
            {
                if (!condition) throw new InvalidOperationException(name);
                checks.Add(name);
            }
            File.WriteAllText(path, "first\r\n\r\n\t  code\r\n😀tail");
            Check(await Read(new { path, start_line = 2, end_line = 3 }) == "2: \r\n3: \t  code\r\n", "actual-range-blank-indentation-crlf");
            Check(await Read(new { path, start_line = 4 }) == "4: 😀tail", "unterminated-final-line");
            Check(await Read(new { path, start_line = 4, max_bytes = 5 }) == "4: 😀t\n[truncated]\n", "source-cap-before-prefix");
            Check(await Read(new { path, start_line = 4, max_bytes = 3 }) == "\n[truncated]\n", "utf8-boundary-no-invented-source");
            Check(await Read(new { path, start_line = 8 }) == "", "past-eof");
            File.WriteAllText(path, "");
            Check(await Read(new { path }) == "", "empty-file");
            File.WriteAllText(path, "last\n");
            Check(await Read(new { path }) == "1: last\n", "no-phantom-final-line");
            File.WriteAllText(path, new string('x', 32769));
            string capped = "1: " + new string('x', 32768) + "\n[truncated]\n";
            Check(await Read(new { path, max_output_bytes = 65536 }) == capped, "default-source-cap");
            Check(await Read(new { path, max_bytes = 0, max_output_bytes = 65536 }) == capped, "zero-source-cap");
            Check(await Read(new { path, max_bytes = 40000, max_output_bytes = 65536 }) == "1: " + new string('x', 32769), "explicit-larger-cap");
            Console.WriteLine(JsonSerializer.Serialize(new { contract = "chapter3-read-lines-v1", passed = true, checks }));
            return 0;
        }
        catch (Exception error)
        {
            Console.WriteLine(JsonSerializer.Serialize(new { contract = "chapter3-read-lines-v1", passed = false, checks, failure = error.Message }));
            return 1;
        }
        finally
        {
            app.Destroy();
            Directory.Delete(directory, recursive: true);
        }
    }
}
