using System.Text.Json;
using Ensemble;
using Application = Ensemble.Ensemble;

internal static class ChildPolicyTests
{
    public static async Task Narrowing()
    {
        string workspace = Path.Combine(Path.GetTempPath(), "ensemble-child-policy-" + Guid.NewGuid().ToString("N"));
        string inside = Path.Combine(workspace, "inside");
        string outside = Path.Combine(workspace, "outside");
        Directory.CreateDirectory(inside);
        Directory.CreateDirectory(outside);
        File.WriteAllText(Path.Combine(outside, "sentinel"), "outside-only");
        Directory.CreateDirectory(Path.Combine(inside, "existing"));
        Directory.CreateSymbolicLink(Path.Combine(inside, "alias"), "existing");
        Directory.CreateSymbolicLink(Path.Combine(inside, "escape"), outside);
        var app = new Application(TextWriter.Null);
        try
        {
            var parent = new Agent(app, Request(inside, safe: true, web: false));
            var sandbox = parent.Sandbox();
            string physicalRoot = sandbox.Root();
            var inherited = sandbox.DeriveChild();
            Check(inherited.Refusal is null && inherited.Effective.SafeMode
                && !inherited.Effective.EnableWebSearch && inherited.Effective.SandboxRoot == physicalRoot,
                "Omitted request did not inherit effective policy");
            foreach (string path in new[] { "", "existing", "alias", "future/sub" })
            {
                var narrowed = sandbox.DeriveChild(Request(path, safe: true, web: false));
                Check(narrowed.Refusal is null && narrowed.Effective.SandboxRoot == sandbox.Resolve(path.Length == 0 ? "." : path),
                    "Valid child root refused or not physically resolved");
            }
            Check(!Directory.Exists(Path.Combine(inside, "future")), "Policy derivation created directories");
            foreach (string path in new[] { outside, "../outside", "escape", inside + "-sibling" })
            {
                var refused = sandbox.DeriveChild(Request(path, safe: false, web: true));
                Check(refused.Refusal is not null && !refused.Refusal.Contains(path)
                    && refused.Effective.SandboxRoot == physicalRoot && refused.Effective.SafeMode
                    && !refused.Effective.EnableWebSearch, "Widening refusal returned unsafe policy or target path");
            }

            // Ignore the warning deliberately. The result must still enforce the
            // parent's policy when consumed by the real Agent and file tools.
            var greedy = sandbox.DeriveChild(Request("escape", safe: false, web: true));
            var child = new Agent(app, greedy.Effective);
            child.InstallTools();
            Check(!child.ToolEnabled("run_command") && !child.ToolEnabled("wait_for_job"),
                "Derived child restored safe-mode tools");
            var read = await child.DispatchAsync(new ToolCallPart("outside", child.Configuration().Target(), "read_file",
                JsonSerializer.SerializeToElement(new { path = Path.Combine(outside, "sentinel") })));
            Check(read.IsError && !JsonCodec.Text(read.Parts).Contains("outside-only"), "Derived child escaped through file tool");
            var write = await child.DispatchAsync(new ToolCallPart("inside", child.Configuration().Target(), "write_file",
                JsonSerializer.SerializeToElement(new { path = "child-note", content = "works" })));
            Check(!write.IsError && File.ReadAllText(Path.Combine(inside, "child-note")) == "works", "Child positive file control failed");

            var fullParent = new Agent(app, Request("", safe: false, web: true));
            var strict = fullParent.Sandbox().DeriveChild(Request(inside, safe: true, web: false));
            Check(strict.Refusal is null && strict.Effective.SafeMode && !strict.Effective.EnableWebSearch
                && strict.Effective.SandboxRoot == physicalRoot, "Full-access parent could not narrow child");

            // Cloning policy must not reapply model defaults. In particular an
            // explicitly disabled summary is stored as an empty validated value.
            var plan = new Configuration(Vendor.OpenAI, "gpt-6-astra", "https://api.openai.com",
                resolvedModel: "resolved-plan", requireToolExplanation: true, toolRoundLimit: 37,
                primaryInstructions: "child instructions", surface: Surface.Responses,
                billing: BillingMode.ChatGptPlan, account: "local-label", reasoningEffort: "high",
                reasoningSummary: "off", safeMode: true, sandboxRoot: "existing");
            var preserved = sandbox.DeriveChild(plan).Effective;
            Check(preserved.Billing == plan.Billing && preserved.Account == plan.Account
                && preserved.Surface == plan.Surface && preserved.ResolvedModel == plan.ResolvedModel
                && preserved.ReasoningEffort == "high" && preserved.ReasoningSummary == ""
                && preserved.ToolRoundLimit == 37 && preserved.PrimaryInstructions == plan.PrimaryInstructions,
                "Child policy changed unrelated validated configuration");
            Check(parent.Configuration().SandboxRoot == inside && parent.Configuration().SafeMode
                && !parent.Configuration().EnableWebSearch, "Derivation mutated parent");
        }
        finally
        {
            app.Destroy();
            Directory.Delete(workspace, true);
        }
    }

    private static Configuration Request(string root, bool safe, bool web)
    {
        return new Configuration(Vendor.OpenAI, "gpt-4.1", "http://127.0.0.1:1",
            safeMode: safe, enableWebSearch: web, sandboxRoot: root);
    }

    private static void Check(bool condition, string message)
    {
        if (!condition)
        {
            throw new Exception(message);
        }
    }
}
