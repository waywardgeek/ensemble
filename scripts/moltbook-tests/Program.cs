using System.Diagnostics;
using System.Text.Json;
using Ensemble;

if (args.Length != 2) throw new ArgumentException("Expected source skill directory and Python executable");
string source = Path.GetFullPath(args[0]);
string directory = Path.Combine(Path.GetTempPath(), "moltbook-consumer-" + Guid.NewGuid().ToString("N"));
Directory.CreateDirectory(directory);
Console.WriteLine("Evidence: " + directory);
Directory.CreateDirectory(Path.Combine(directory, "primary"));
Directory.CreateDirectory(Path.Combine(directory, "moltbook"));
File.WriteAllText(Path.Combine(directory, "primary", "SKILL.md"),
    "---\nname: primary\ntype: primary\ntools: wait_for_job kill_job\nloadable-skills: moltbook\n---\nOffline test");
string wrapper = Path.Combine(directory, "fake.py");
string pidPath = Path.Combine(directory, "pid");
string entered = Path.Combine(directory, "entered");
File.WriteAllText(wrapper, $$"""
import sys, os, time
sys.path.insert(0, {{JsonSerializer.Serialize(source)}})
import moltbook_server as m
open({{JsonSerializer.Serialize(pidPath)}}, 'w').write(str(os.getpid()))
def fake(method, path, body=None):
    if path == '/home': return {'your_account': {'karma': 7}, 'activity_on_your_posts': []}
    open({{JsonSerializer.Serialize(entered)}}, 'w').close()
    time.sleep(60)
    return {'posts': []}
m.api_request = fake
m.main()
""");
string manual = File.ReadAllText(Path.Combine(source, "SKILL.md"));
if (manual.Split('\n').Count(line => line.StartsWith("    args: ")) != 1)
    throw new Exception("Expected exactly one MCP argv declaration; refuse unsafe fixture");
manual = string.Join('\n', manual
    .Split('\n').Select(line => line.StartsWith("    args: ")
        ? "    args: " + JsonSerializer.Serialize(new[] { wrapper, "--credentials", Path.Combine(directory, "absent-credentials") })
        : line.StartsWith("    command: ") ? "    command: " + args[1] : line));
File.WriteAllText(Path.Combine(directory, "moltbook", "SKILL.md"), manual);
var root = new Ensemble.Ensemble(TextWriter.Null, dataDirectory: Path.Combine(directory, "data"));
try
{
    var agent = new Agent(root, new Configuration(Vendor.Anthropic, "fake", "http://127.0.0.1:1"),
        skillSpec: new SkillSpec(directory, "primary"));
    async Task<ToolResultPart> Call(string name, object arguments) =>
        await agent.DispatchAsync(new ToolCallPart(Guid.NewGuid().ToString("N"),
            agent.Configuration().Target(), name, JsonSerializer.SerializeToElement(arguments)))
            .WaitAsync(TimeSpan.FromSeconds(15));
    void Check(bool condition, string message) { if (!condition) throw new Exception(message); }
    Check(!agent.ToolEnabled("moltbook_status"), "tools granted before load");
    var loaded = await Call("load_skill", new { name = "moltbook" });
    Check(!loaded.IsError, "skill load failed: " + string.Join(" ", loaded.Parts.OfType<TextPart>().Select(p => p.Text)));
    Check(agent.Tools().Count(t => t.Name().StartsWith("moltbook_")) == 12, "wrong tool count");
    var status = await Call("moltbook_status", new { ai_callback_delay = 5 });
    Check(!status.IsError && status.Parts.OfType<TextPart>().Any(p => p.Text.Contains("7")), "fake status not delivered");
    var held = await Call("moltbook_feed", new { ai_callback_delay = 0 });
    Check(held.Job is { Status: "running" }, "missing running Job");
    var deadline = DateTime.UtcNow.AddSeconds(5);
    while (!File.Exists(entered) && DateTime.UtcNow < deadline) await Task.Delay(10);
    Check(File.Exists(entered), "call never entered");
    await Call("kill_job", new { handle = held.Job!.Handle, ai_callback_delay = 5 });
    int pid = int.Parse(File.ReadAllText(pidPath));
    using var process = Process.GetProcessById(pid);
    Check(!process.HasExited, "unexpected process exit before unload");
    var unloaded = await Call("unload_skill", new { name = "moltbook" });
    Check(!unloaded.IsError && !agent.ToolEnabled("moltbook_status"), "unload failed");
    Check(process.HasExited, "unload did not join server process");
    Check(agent.SkillManager()!.Clients().IsEmpty, "client still attached");
    Console.WriteLine("PASS real skill discovery, grants, API result, delayed Job, cancellation and joined unload");
}
finally { root.Destroy(); }
