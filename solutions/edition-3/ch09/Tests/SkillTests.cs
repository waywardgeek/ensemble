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
            Check(!unload.IsError && manager.Pending("editing") && agent.ToolEnabled("write_file"), "Pending unload removed capability early");
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

    public static async Task Projection()
    {
        string root = Fixture();
        var app = new Application(TextWriter.Null);
        try
        {
            var parsed = Skills.Parse("---\nname: list\ntype: loadable\ntools:\n - read_file\n - list_directory\n---\nbody\n---\n$UNKNOWN", "fixture");
            Check(parsed.Tools.Length == 2 && parsed.Body.EndsWith("---\n$UNKNOWN"), "Frontmatter consumed body or lost list");
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
            Check(declared.SequenceEqual(new[] { "load_skill", "read_file", "unload_skill" }),
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
