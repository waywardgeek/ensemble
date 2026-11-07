using System.Text.Json;
using Ensemble;
using Application = Ensemble.Ensemble;

internal static class SkillTests
{
    private static void Check(bool condition, string message)
    {
        if (!condition)
        {
            throw new Exception(message);
        }
    }

    internal static string Fixture()
    {
        string root = Path.Combine(Path.GetTempPath(), "skills-" + Guid.NewGuid().ToString("N"));
        void Skill(string name, string type, string fields, string body)
        {
            string directory = Path.Combine(root, name);
            Directory.CreateDirectory(directory);
            File.WriteAllText(Path.Combine(directory, "SKILL.md"),
                $"---\nname: {name}\ndescription: {name} description\ntype: {type}\n{fields}\n---\n{body}");
        }
        Skill("root", "primary", "tools: read_file\nloadable-skills: editing broken cycle", "PRIMARY-$WHO\n$TOOLS\n$SKILLS");
        Skill("editing", "loadable", "tools: write_file read_file\ndepends: helper\nloadable-skills: extra", "EDIT-$WHO\n$TOOLS\n$SKILLS");
        Skill("helper", "dependency", "tools: list_directory", "HELPER-$WHO");
        Skill("extra", "loadable", "tools: search_files", "EXTRA-$WHO");
        Skill("broken", "loadable", "depends: helper missing", "MUST-NOT-APPEAR");
        Skill("cycle", "loadable", "depends: loop", "CYCLE");
        Skill("loop", "dependency", "depends: cycle", "LOOP");
        return root;
    }

    private static ToolCallPart Call(string name, object arguments)
    {
        return new ToolCallPart(Guid.NewGuid().ToString("N"),
            new Provenance(Vendor.Anthropic, "model-a", Surface.Messages), name,
            JsonSerializer.SerializeToElement(arguments));
    }

    public static async Task Capabilities()
    {
        string root = Fixture();
        var app = new Application(TextWriter.Null);
        try
        {
            var agent = new Agent(app, new Configuration(Vendor.Anthropic, "model-a", "http://127.0.0.1:1"),
                skillSpec: new SkillSpec(root, "root", new Dictionary<string, string> { ["WHO"] = "before" }));
            var manager = agent.SkillManager()!;
            string primary = manager.Active().Single().Body;
            Check(primary.Contains("PRIMARY-before") && !primary.Contains("$TOOLS"), "Primary was not rendered at construction");
            Check(agent.ToolEnabled("read_file") && !agent.ToolEnabled("write_file"), "Initial grants are wrong");
            Check(manager.LoadableSkills().SequenceEqual(new[] { "broken", "cycle", "editing" }), "Progressive discovery is wrong");
            string target = Path.Combine(root, "effect.txt");
            var denied = await agent.DispatchAsync(Call("write_file", new { path = target, content = "no" }));
            Check(denied.IsError && !File.Exists(target), "Disabled tool had an effect");
            Check(!agent.History().Events().Any(value => value.Kind == EventKind.ToolStarted), "Refused tool claimed admission");
            foreach (string name in new[] { "extra", "broken", "cycle", "unknown" })
            {
                var refused = await agent.DispatchAsync(Call("load_skill", new { name }));
                Check(refused.IsError, "Invalid closure was accepted: " + name);
                Check(manager.Active().Length == 1 && !agent.ToolEnabled("list_directory"), "Failed load leaked a dependency");
            }
            manager.RegisterVariable("WHO", "after");
            var loaded = await agent.DispatchAsync(Call("load_skill", new { name = "editing" }));
            Check(!loaded.IsError && manager.Active().Length == 3, "Dependency closure did not activate");
            Check(manager.Active().Single(skill => skill.Primary).Body == primary, "Primary snapshot was rerendered");
            Check(manager.Active().Single(skill => skill.Name == "editing").Body.Contains("EDIT-after"), "Load did not capture current variable");
            Check(manager.Active().Single(skill => skill.Name == "helper").Body == "HELPER-after", "Dependency expansion failed");
            Check(manager.LoadableSkills().Contains("extra"), "New nested skill was not offered");
            var source = manager.Source("read_file");
            Check(source?.Skill == "root", "Later shared grant replaced first enabler");
            long activated = agent.History().Events().Count(value => value.Kind == EventKind.SkillLoaded);
            await agent.DispatchAsync(Call("load_skill", new { name = "editing" }));
            Check(agent.History().Events().Count(value => value.Kind == EventKind.SkillLoaded) == activated, "Repeated load duplicated instructions");
            var written = await agent.DispatchAsync(Call("write_file", new { path = target, content = "yes" }));
            Check(!written.IsError && File.ReadAllText(target) == "yes", "Enabled real tool did not execute");
            var unload = await agent.DispatchAsync(Call("unload_skill", new { name = "editing" }));
            Check(!unload.IsError && manager.Pending("editing") && !agent.ToolEnabled("write_file"), "Pending unload retained capability");
            Check((await agent.DispatchAsync(Call("unload_skill", new { name = "root" }))).IsError, "Primary unload accepted");
            var other = new Agent(app, new Configuration(Vendor.Anthropic, "model-a", "http://127.0.0.1:1"),
                skillSpec: new SkillSpec(root, "root"));
            Check(!other.ToolEnabled("write_file"), "Skill grants leaked between Agents");
            // Redacting dialogue must not silently revoke persistent instructions.
            agent.History().Append(new Event(EventKind.Redacted, redactFrom: 1, redactTo: agent.History().Events().Last().Seq, level: Redaction.Dialogue));
            Check(manager.Active().Length == 3, "Dialogue redaction removed instruction facts");
        }
        finally
        {
            app.Destroy();
            Directory.Delete(root, true);
        }
    }

    public static Task FileReadingDependency()
    {
        string root = Path.Combine(Directory.GetCurrentDirectory(), "agent", "skills");
        string reads = "read_file list_directory search_files";
        var primaries = new Dictionary<string, string>
        {
            ["ensemble"] = reads + " wait_for_job agent_status spawn_sub_agent agents_status get_sub_agent_status get_submission interrupt_agent join_agents submit send_message_to_parent respond_to_parent_message shutdown_agent kill_agent send_message ask_agent wait_for_agent wait_for_agent_change",
            ["full"] = reads + " write_file edit_file run_command wait_for_job send_input kill_job tool_limits",
            ["local-coding"] = reads + " write_file edit_file run_command"
        };
        var app = new Application(TextWriter.Null);
        try
        {
            foreach (var (primary, tools) in primaries)
            {
                var agent = new Agent(app,
                    new Configuration(Vendor.OpenAI, "gpt-6-astra", "http://127.0.0.1:1", surface: Surface.Responses),
                    skillSpec: new SkillSpec(root, primary));
                var manager = agent.SkillManager()!;
                var expected = (tools + " load_skill unload_skill micro_handoff").Split(' ').Order(StringComparer.Ordinal);
                var actual = agent.Tools().Select(tool => tool.Name()).Where(agent.ToolEnabled).Order(StringComparer.Ordinal);
                Check(actual.SequenceEqual(expected), primary + " changed its effective tools");
                Check(manager.Active().Count(skill => skill.Name == "file-reading") == 1,
                    primary + " must activate file-reading exactly once");
                foreach (string tool in reads.Split(' '))
                    Check(manager.Source(tool)?.Skill == "file-reading", primary + " has wrong read-tool owner");
                Check(!manager.Active().Single(skill => skill.Primary).Tools.Intersect(reads.Split(' ')).Any(),
                    primary + " still duplicates the read grants");
            }
        }
        finally { app.Destroy(); }
        return Task.CompletedTask;
    }

    public static async Task NativeCodingGuidance()
    {
        string repository = Directory.GetCurrentDirectory();
        string skillsRoot = Path.Combine(repository, "agent", "Skills");
        string nativePath = Path.Combine(skillsRoot, "ensemble-coding", "SKILL.md");
        string repositoryPath = Path.Combine(repository, ".agents", "skills", "ensemble-coding", "SKILL.md");
        var native = Skills.Parse(File.ReadAllText(nativePath), nativePath);
        string repositoryText = File.ReadAllText(repositoryPath).Replace("\r\n", "\n");
        int headerEnd = repositoryText.IndexOf("\n---\n", StringComparison.Ordinal);
        Check(headerEnd >= 0, "Repository coding manual has no frontmatter boundary");
        string repositoryBody = repositoryText[(headerEnd + "\n---\n".Length)..].Trim();
        Check(native.Body.Trim() == repositoryBody, "Native and repository coding guidance drifted");
        Check(native.Tools.IsEmpty && native.Depends.IsEmpty, "Coding guidance must not add tool authority");

        var app = new Application(TextWriter.Null);
        try
        {
            var variables = new Dictionary<string, string>
            {
                ["HUMAN_NAME"] = "Fixture",
                ["HUMAN_DESCRIPTION"] = "Isolated coding-skill test"
            };
            var agent = new Agent(app,
                new Configuration(Vendor.OpenAI, "gpt-6-astra", "http://127.0.0.1:1", surface: Surface.Responses),
                skillSpec: new SkillSpec(skillsRoot, "ensemble", variables));
            var manager = agent.SkillManager()!;
            var explanationPath = Path.Combine(skillsRoot, "visible-reasoning", "SKILL.md");
            var explanation = Skills.Parse(File.ReadAllText(explanationPath), explanationPath);
            Check(explanation.Type == SkillType.Dependency && explanation.Tools.IsEmpty
                && explanation.Depends.IsEmpty, "Visible reasoning must be a tool-free dependency");
            var guidance = manager.Active().Single(skill => skill.Name == "visible-reasoning");
            Check(!guidance.Primary && guidance.Body.Contains("Pair with Fixture")
                && !guidance.Body.Contains("$HUMAN_NAME"), "Primary dependency did not activate with variables");
            string initialInstructions = agent.Engine().Render(agent.History().View())["instructions"]!.GetValue<string>();
            Check(initialInstructions.Contains(guidance.Body, StringComparison.Ordinal)
                && initialInstructions.Split("# Visible engineering reasoning").Length == 2,
                "Visible reasoning must appear once in initial provider instructions");
            Check(manager.LoadableSkills().Contains("ensemble-coding"), "Native coding skill is not offered");
            string[] toolsBefore = manager.CatalogTools().Order(StringComparer.Ordinal).ToArray();
            var result = await agent.DispatchAsync(Call("load_skill", new { name = "ensemble-coding" }));
            Check(!result.IsError, "Native coding skill failed to activate");
            var loaded = manager.Active().Single(skill => skill.Name == "ensemble-coding");
            Check(loaded.Body.Trim() == repositoryBody, "Activation changed the coding guidance");
            Check(manager.CatalogTools().Order(StringComparer.Ordinal).SequenceEqual(toolsBefore),
                "Loading instruction-only guidance changed the tool grants");
            var request = agent.Engine().Render(agent.History().View());
            string instructions = request["instructions"]!.GetValue<string>();
            Check(instructions.Contains(repositoryBody, StringComparison.Ordinal),
                "Responses instructions lack the complete coding guidance");
            string heading = repositoryBody.Split('\n').First(line => line.StartsWith('#'));
            string encodedHeading = JsonSerializer.Serialize(heading)[1..^1];
            Check(!request["input"]!.ToJsonString().Contains(encodedHeading, StringComparison.Ordinal),
                "Coding guidance appeared as ordinary dialogue");
        }
        finally
        {
            app.Destroy();
        }
    }

    public static Task AgentQueueGuidance()
    {
        string skillsRoot = Path.Combine(Directory.GetCurrentDirectory(), "agent", "Skills");
        string root = Path.Combine(Path.GetTempPath(), "queue-guidance-" + Guid.NewGuid().ToString("N"));
        var app = new Application(TextWriter.Null);
        try
        {
            string queuePath = Path.Combine(skillsRoot, "task-queue", "SKILL.md");
            var definition = Skills.Parse(File.ReadAllText(queuePath), queuePath);
            Check(definition.Type == SkillType.Dependency && definition.Tools.IsEmpty
                && definition.Depends.IsEmpty, "Task queue must grant no tools");
            var config = new Configuration(Vendor.OpenAI, "gpt-6-astra", "http://127.0.0.1:1", surface: Surface.Responses);
            Agent Make(string? home) => new(app, config, skillSpec: new SkillSpec(skillsRoot, "ensemble"), dataDirectory: home);
            var first = Make(Path.Combine(root, "first"));
            var second = Make(Path.Combine(root, "second"));
            var transient = Make(null);
            foreach (var agent in new[] { first, second, transient })
            {
                string expected = agent == transient ? "UNAVAILABLE" : agent.DataDirectory();
                var manager = agent.SkillManager()!;
                var queue = manager.Active().Single(part => part.Name == "task-queue");
                Check(queue.Body.Contains("directory is: " + expected)
                    && !queue.Body.Contains("$AGENT_DATA_DIRECTORY"), "Queue used the wrong Agent directory");
                Check(manager.Active().Single(part => part.Primary).Body.Contains("directory is: " + expected),
                    "Primary must know its own directory too");
                string instructions = agent.Engine().Render(agent.History().View())["instructions"]!.GetValue<string>();
                Check(instructions.Contains(queue.Body) && instructions.Split("# Per-Agent task queue").Length == 2,
                    "Queue dependency must appear once in provider instructions");
                bool refused = false;
                try { manager.RegisterVariable("AGENT_DATA_DIRECTORY", "/forged"); }
                catch (ArgumentException) { refused = true; }
                Check(refused, "Host variables must not override Agent storage identity");
                if (agent != transient)
                    Check(!File.Exists(Path.Combine(expected, "queue.md")), "Skill activation created a queue without a request");
            }
            Check(!first.SkillManager()!.Active().Single(part => part.Name == "task-queue").Body.Contains(second.DataDirectory()),
                "One Agent's queue guidance leaked another Agent's directory");
        }
        finally
        {
            app.Destroy();
            if (Directory.Exists(root)) Directory.Delete(root, recursive: true);
        }
        return Task.CompletedTask;
    }

    public static async Task Projection()
    {
        string root = Fixture();
        var app = new Application(TextWriter.Null);
        try
        {
            var parsed = Skills.Parse("---\nname: list\ntype: loadable\ntools:\n - read_file\n - list_directory\n---\nbody\n---\n$UNKNOWN", "fixture");
            Check(parsed.Tools.Length == 2 && parsed.Body.EndsWith("---\n$UNKNOWN"), "Frontmatter consumed body or lost list");
            var described = Skills.Parse("---\nname: display\ntype: dependency\nrenderers:\n  list_directory: renderers/list-directory.js\n---\nManual", "fixture");
            Check(described.Renderers.Single() == new SkillRenderer("list_directory", "renderers/list-directory.js")
                && described.Tools.IsEmpty && described.Body == "Manual", "Renderer metadata changed grants or body");
            Check(parsed.Renderers.IsEmpty, "Legacy skill acquired a renderer");
            foreach (string mapping in new[] {
                "  list_directory: ../outside.js", "  list_directory: /absolute.js",
                "  list_directory: a//b.js", "  list_directory: a/./b.js",
                "  list_directory: https://example.com/a.js", "  list_directory: a.txt",
                "  list_directory: a.js\n  list_directory: b.js",
                "  list_directory: a.js\n   read_file: b.js", "" })
            {
                bool refused = false;
                try { Skills.Parse("---\nname: bad\ntype: dependency\nrenderers:\n" + mapping + "\n---\nBody", "fixture"); }
                catch (InvalidDataException) { refused = true; }
                Check(refused, "Invalid renderer declaration accepted: " + mapping);
            }
            bool invalid = false;
            try { Skills.Parse("---\nname: x\ntype: mystery\n---\nbody", "fixture"); }
            catch (InvalidDataException) { invalid = true; }
            Check(invalid, "Invalid type accepted");
            // No offered children must not hide the universal eager controls in
            // deferred mode. Declarations alone still do not grant a writer.
            string lone = Path.Combine(root, "lone");
            Directory.CreateDirectory(lone);
            File.WriteAllText(Path.Combine(lone, "SKILL.md"),
                "---\nname: lone\ntype: primary\ntools: read_file\n---\nOnly read.");
            var deferred = new Agent(app,
                new Configuration(Vendor.Anthropic, "model-a", "http://127.0.0.1:1", deferredTools: true),
                skillSpec: new SkillSpec(root, "lone"));
            var declarations = deferred.Engine().Render(deferred.History().View())["tools"]!.AsArray();
            string[] declared = declarations.Select(tool => tool!["name"]!.GetValue<string>()).Order().ToArray();
            Check(declared.SequenceEqual(new[] { "load_skill", "micro_handoff", "read_file", "unload_skill" }),
                "Deferred primary without offers lost its eager controls");
            Check(declarations.All(tool => tool!["defer_loading"] is null), "Initial control was deferred");
            foreach (var vendor in Enum.GetValues<Vendor>())
            {
                var agent = new Agent(app, new Configuration(vendor, "model-a", "http://127.0.0.1:1"),
                    skillSpec: new SkillSpec(root, "root", new Dictionary<string, string> { ["WHO"] = "once" }));
                await agent.DispatchAsync(Call("load_skill", new { name = "editing" }));
                var body = agent.Engine().Render(agent.History().View());
                string instructions = vendor switch
                {
                    Vendor.Anthropic => body["system"]!.GetValue<string>(),
                    Vendor.OpenAI => body["messages"]![0]!["content"]!.GetValue<string>(),
                    _ => body["systemInstruction"]!["parts"]![0]!["text"]!.GetValue<string>()
                };
                Check(instructions.Contains("PRIMARY-once") && instructions.Contains("EDIT-once")
                    && instructions.Contains("HELPER-once"), "Instructions missing from provider role: " + vendor);
                string serialized = body.ToJsonString();
                Check(serialized.Split("EDIT-once").Length == 2, "Instruction body duplicated in dialogue: " + vendor);
                foreach (var fact in agent.History().Events().Where(value => value.Kind is EventKind.SkillLoaded or EventKind.SkillUnloadPending))
                {
                    string line = JsonCodec.WriteEvent(fact);
                    Check(JsonCodec.WriteEvent(JsonCodec.ReadEvent(line)) == line, "Skill event codec changed payload");
                }
                var manager = agent.SkillManager()!;
                agent.RemoveSkills(manager);
                Check(agent.SkillManager() is null, "Remove left parent membership");
                bool detached = false;
                try { manager.Parent(); } catch (ObjectDisposedException) { detached = true; }
                Check(detached, "Remove left reverse parent link");
                var unavailable = await agent.DispatchAsync(Call("load_skill", new { name = "editing" }));
                Check(unavailable.IsError, "Detached manager left a callable loader");
                Check(!unavailable.ToolReferences.Any(), "Refused loader advertised discovery");
                agent.AddSkills(manager);
                manager.Destroy();
                Check(agent.SkillManager() is null, "Destroy left membership");
            }
        }
        finally
        {
            app.Destroy();
            Directory.Delete(root, true);
        }
    }
}
