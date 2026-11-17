using System.Text.Json;
using Ensemble;
using Application = Ensemble.Ensemble;

// This is a host-side loader check, not a model trial. No request is sent.
if (args.Length is < 2 or > 3)
{
    throw new ArgumentException("Usage: SkillSurface REPOSITORY CANDIDATE_SKILL_DIRECTORY [flint-student|flint-reviewer]");
}
string repository = Path.GetFullPath(args[0]);
string candidate = Path.GetFullPath(args[1]);
string primary = args.Length == 3 ? args[2] : "flint-student";
if (primary is not ("flint-student" or "flint-reviewer"))
{
    throw new ArgumentException("Unknown experimental primary");
}
string root = Path.Combine(Path.GetTempPath(), "flint-skill-check-" + Guid.NewGuid().ToString("N"));
Directory.CreateDirectory(root);
var app = new Application(TextWriter.Null);
try
{
    // Copy only explicit dependencies. The full host catalogue is not the fixture.
    foreach (string name in new[] { "file-reading", "coding", "ensemble-coding" })
    {
        string source = name == "coding"
            ? Path.Combine(repository, "agent", "skills", "dev-driven-specs", name)
            : Path.Combine(repository, "agent", "skills", name);
        CopyTree(source, Path.Combine(root, name));
    }
    CopyTree(candidate, Path.Combine(root, primary));
    string workspace = Path.Combine(root, "workspace");
    Directory.CreateDirectory(workspace);
    var agent = new Agent(app,
        new Configuration(Vendor.Anthropic, "fixture", "http://127.0.0.1:1", sandboxRoot: workspace),
        skillSpec: new SkillSpec(root, primary));
    var manager = agent.SkillManager()!;
    string expectedTools = primary == "flint-student"
        ? "read_file list_directory search_files write_file edit_file run_command wait_for_job agent_status send_message_to_parent submit load_skill unload_skill micro_handoff"
        : "read_file list_directory search_files agent_status send_message_to_parent submit load_skill unload_skill micro_handoff";
    string[] expected = expectedTools.Split(' ').Order(StringComparer.Ordinal).ToArray();
    string[] before = agent.Tools().Select(tool => tool.Name()).Where(agent.ToolEnabled)
        .Order(StringComparer.Ordinal).ToArray();
    Require(before.SequenceEqual(expected), "Unexpected effective tool surface: " + string.Join(" ", before));
    Require(manager.Active().Count(skill => skill.Name == "coding") == 1, "Coding rules not active exactly once");
    Require(manager.LoadableSkills().SequenceEqual(new[] { "ensemble-coding" }), "Unexpected loadable skills");
    var call = new ToolCallPart(Guid.NewGuid().ToString("N"),
        new Provenance(Vendor.Anthropic, "fixture", Surface.Messages), "load_skill",
        JsonSerializer.SerializeToElement(new { name = "ensemble-coding" }));
    var result = await agent.DispatchAsync(call);
    Require(!result.IsError, "Required coding guidance failed to load");
    Require(manager.Active().Count(skill => skill.Name == "ensemble-coding") == 1,
        "Required coding guidance not active exactly once");
    string[] after = agent.Tools().Select(tool => tool.Name()).Where(agent.ToolEnabled)
        .Order(StringComparer.Ordinal).ToArray();
    Require(after.SequenceEqual(before), "Instruction-only activation changed tool authority");
    Console.WriteLine("PASS: real loader, exact tools, coding activation, no provider request");
}
finally
{
    // Teardown joins Agent-owned work before removing its disposable files.
    app.Destroy();
    Directory.Delete(root, recursive: true);
}

static void Require(bool condition, string message)
{
    if (!condition)
    {
        throw new InvalidOperationException(message);
    }
}

static void CopyTree(string source, string destination)
{
    Directory.CreateDirectory(destination);
    foreach (string file in Directory.GetFiles(source))
    {
        File.Copy(file, Path.Combine(destination, Path.GetFileName(file)));
    }
    foreach (string directory in Directory.GetDirectories(source))
    {
        CopyTree(directory, Path.Combine(destination, Path.GetFileName(directory)));
    }
}
