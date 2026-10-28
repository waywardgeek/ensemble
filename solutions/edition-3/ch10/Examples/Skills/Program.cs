using System.Text.Json;
using Ensemble;
using Application = Ensemble.Ensemble;

// A separate assembly uses only the public API. No provider is needed to test
// capability admission: DispatchAsync enters the same Agent mailbox as a model call.
if (args.Length != 1 || Directory.Exists(args[0]) || File.Exists(args[0]))
{
    Console.Error.WriteLine("Supply one new owned exercise directory.");
    return 2;
}
string workspace = Path.GetFullPath(args[0]);
string catalog = Path.Combine(workspace, "skills");
WriteSkill("base", """
    ---
    name: base
    type: primary
    tools: read_file
    loadable-skills: writer
    ---
    Primary marker: $MARKER
    Available skills:
    $SKILLS
    """);
WriteSkill("writer", """
    ---
    name: writer
    type: loadable
    tools: write_file
    ---
    Loaded marker: $MARKER
    Enabled tools:
    $TOOLS
    """);

var app = new Application(Console.Error);
try
{
    // The endpoint is deliberately unused. The library reads no environment.
    var configuration = new Configuration(Vendor.Anthropic, "local", "http://127.0.0.1:1");
    var specification = new SkillSpec(catalog, "base",
        new Dictionary<string, string> { ["MARKER"] = "creation" });
    var agent = new Agent(app, configuration, skillSpec: specification);
    var manager = agent.SkillManager()!;
    string primary = manager.Active().Single().Body;
    string target = Path.Combine(workspace, "note.txt");

    var refused = await Invoke("write_file", new { path = target, content = "not yet" });
    Require(refused.IsError && !File.Exists(target), "A hidden tool executed.");

    manager.RegisterVariable("MARKER", "load-time");
    var loaded = await Invoke("load_skill", new { name = "writer" });
    Require(!loaded.IsError, "The offered writer failed to load.");
    Require(manager.Active().Single(skill => skill.Primary).Body == primary,
        "Loading rewrote the primary snapshot.");
    Require(manager.Active().Single(skill => skill.Name == "writer").Body.Contains("load-time"),
        "Loading did not capture the new variable.");

    // Pending removal is deliberately not revocation. Test the newly granted
    // writer after unload, rather than a reader that the primary already owned.
    var pending = await Invoke("unload_skill", new { name = "writer" });
    Require(!pending.IsError && manager.Pending("writer"), "Unload was not recorded.");
    var written = await Invoke("write_file", new { path = target, content = "SKILLS_OK" });
    Require(!written.IsError && File.ReadAllText(target) == "SKILLS_OK",
        "Pending unload prematurely disabled the real writer.");
    var source = manager.Source("write_file");
    Require(source?.Skill == "writer" && source.Value.Sequence > 0, "Missing enabling provenance.");
    Console.WriteLine("Public skills consumer: refusal, snapshot, activation, provenance and pending write passed.");
    Console.WriteLine(target);
    return 0;

    Task<ToolResultPart> Invoke(string name, object arguments)
    {
        var provenance = new Provenance(Vendor.Anthropic, "local", Surface.Messages);
        var call = new ToolCallPart(Guid.NewGuid().ToString("N"), provenance, name,
            JsonSerializer.SerializeToElement(arguments));
        return agent.DispatchAsync(call);
    }
}
finally
{
    app.Destroy();
}

void WriteSkill(string name, string text)
{
    string directory = Path.Combine(catalog, name);
    Directory.CreateDirectory(directory);
    File.WriteAllText(Path.Combine(directory, "SKILL.md"), text);
}

static void Require(bool condition, string message)
{
    if (!condition)
    {
        throw new InvalidOperationException(message);
    }
}
