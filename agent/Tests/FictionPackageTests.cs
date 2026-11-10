using System.Collections.Immutable;
using System.Text.Json;
using Ensemble;
using Application = Ensemble.Ensemble;

static partial class McpTests
{
    // Copy the distributed package unchanged: miniature fixture manuals cannot
    // prove that its literal launch paths and dependency graph work together.
    public static async Task FictionPackage(string directory, bool sharedReadOnly = false)
    {
        if (!OperatingSystem.IsMacOS())
        {
            throw new PlatformNotSupportedException("Package test requires macOS confinement");
        }
        string home = Path.Combine(directory, sharedReadOnly ? "fiction-package-mapped" : "fiction-package");
        string skills = sharedReadOnly
            ? Path.Combine(directory, "fiction-shared-skills")
            : Path.Combine(home, "skills");
        string run = Path.Combine(home, "run");
        string baseline = Path.Combine(home, "baseline");
        Directory.CreateDirectory(run);
        void CopyTree(string source, string destination)
        {
            Directory.CreateDirectory(destination);
            foreach (string file in Directory.GetFiles(source))
            {
                File.Copy(file, Path.Combine(destination, Path.GetFileName(file)));
            }
            foreach (string child in Directory.GetDirectories(source))
            {
                if (Path.GetFileName(child) != "__pycache__")
                {
                    CopyTree(child, Path.Combine(destination, Path.GetFileName(child)));
                }
            }
        }
        foreach (string name in new[] { "fiction-workflow", "shared", "file-reading", "workflows" })
        {
            CopyTree(Path.Combine("agent/skills", name), Path.Combine(skills, name));
        }
        if (sharedReadOnly)
        {
            Directory.CreateSymbolicLink(Path.Combine(home, "skills"), skills);
        }
        CopyTree("docs/edition-3-notes/fiction-workflow", baseline);
        File.WriteAllText(Path.Combine(run, "outline.md"), "One chapter, a harmless fictional journey.");
        File.WriteAllText(Path.Combine(run, "voice.md"), "Clear and concrete.");
        File.WriteAllText(Path.Combine(home, "fiction-workflow-setup.json"), JsonSerializer.Serialize(new
        {
            version = 1, owner_directory = home, run_directory = run,
            baseline_root = baseline, chapter_count = 1, outline = "outline.md",
            voice = "voice.md", source_limit = 32768, result_limit = 65536
        }));
        var root = new Application(TextWriter.Null);
        try
        {
            var agent = new Agent(root, new Configuration(Vendor.Anthropic, "fake",
                "http://127.0.0.1:1", sandboxRoot: home,
                processReadPaths: sharedReadOnly ? ImmutableArray.Create(skills) : null), dataDirectory: home,
                skillSpec: new SkillSpec(skills, "fiction-supervisor"));
            // The authoring manual is discovered through the real primary graph.
            // Loading instructions must not acquire capabilities or begin book work.
            // Compare a shipped, fixed contract with the native declaration, never
            // regenerate the expected schema from the implementation under test.
            var expectedRead = System.Text.Json.Nodes.JsonNode.Parse(File.ReadAllText(
                Path.Combine(skills, "workflows", "contracts.json")))!["read_file"];
            var nativeRead = agent.Tools().Single(tool => tool.Name() == "read_file")
                .Declaration(Vendor.Anthropic)["input_schema"];
            Check(System.Text.Json.Nodes.JsonNode.DeepEquals(expectedRead, nativeRead),
                "Workflow example read_file contract differs from the native declaration");
            var toolsBefore = agent.Tools().Select(tool => tool.Name()).Order().ToArray();
            var taught = await agent.DispatchAsync(SkillCall(agent, "load_skill", "workflows"));
            Check(!taught.IsError, "Workflow authoring skill failed to load: " + JsonCodec.Text(taught.Parts));
            var manual = agent.SkillManager()!.Active().Single(part => part.Name == "workflows");
            var expectedManual = Skills.Parse(File.ReadAllText(Path.Combine(skills, "workflows", "SKILL.md")),
                Path.Combine(skills, "workflows", "SKILL.md"));
            Check(manual.Body == expectedManual.Body, "Loaded workflow manual differs from its package");
            Check(toolsBefore.SequenceEqual(agent.Tools().Select(tool => tool.Name()).Order()),
                "Instructional workflow skill changed tool authority");
            Check(agent.SkillManager()!.Clients().IsEmpty && agent.Children().Length == 0
                && !File.Exists(Path.Combine(run, "state.json")),
                "Instructional workflow skill started execution");
            var teachingUnloaded = await agent.DispatchAsync(SkillCall(agent, "unload_skill", "workflows"));
            Check(!teachingUnloaded.IsError, "Workflow authoring skill failed to unload");
            var loaded = await agent.DispatchAsync(SkillCall(agent, "load_skill", "fiction-workflow"))
                .WaitAsync(TimeSpan.FromSeconds(15));
            Check(!loaded.IsError, "Packaged workflow load failed: " + JsonCodec.Text(loaded.Parts));
            Check(agent.Tools().Any(tool => tool.Name() == "continue_fiction_book"),
                "Packaged workflow tool missing");
            Check(agent.SkillManager()!.Clients().Length == 1, "Expected one owned package server");
            foreach (string role in new[] { "fiction-author", "fiction-editor", "fiction-reviewer" })
            {
                var roleLoaded = await agent.DispatchAsync(SkillCall(agent, "load_skill", role));
                Check(!roleLoaded.IsError, "Role failed to load: " + role + " " + JsonCodec.Text(roleLoaded.Parts));
                Check(agent.Tools().Any(tool => tool.Name() == "read_file"), "Role reading dependency missing");
                var roleUnloaded = await agent.DispatchAsync(SkillCall(agent, "unload_skill", role));
                Check(!roleUnloaded.IsError, "Role failed to unload: " + role);
            }
            Check(agent.Children().Length == 0 && !File.Exists(Path.Combine(run, "state.json")),
                "Loading package unexpectedly started book work");
            var unloaded = await agent.DispatchAsync(SkillCall(agent, "unload_skill", "fiction-workflow"));
            Check(!unloaded.IsError && agent.SkillManager()!.Clients().IsEmpty,
                "Package unload did not detach/join its server");
        }
        finally
        {
            root.Destroy();
        }
    }
}
