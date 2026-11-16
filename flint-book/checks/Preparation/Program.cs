using System.Collections.Immutable;
using System.Text;
using System.Text.Json;
using System.Text.RegularExpressions;
using Ensemble;
using Application = Ensemble.Ensemble;

// Provider-free host checker. Only disposable, self-authored code is executed.
// The command runs through the actual Agent tool and OS confinement boundary.
if (args.Length != 3)
{
    throw new ArgumentException("Usage: Preparation REPOSITORY SDK_ROOT GO_EXECUTABLE");
}
string repository = Path.GetFullPath(args[0]);
string sdk = Path.GetFullPath(args[1]);
string go = Path.GetFullPath(args[2]);
string goRoot = Directory.GetParent(Path.GetDirectoryName(go)!)!.FullName;
string workspace = Path.Combine(Path.GetTempPath(), "flint-preparation-check-" + Guid.NewGuid().ToString("N"));
Directory.CreateDirectory(workspace);
var app = new Application(TextWriter.Null);
try
{
    string skill = File.ReadAllText(Path.Combine(repository, "flint-book/skills/flint-student/SKILL.md"));
    var blocks = Regex.Matches(skill, "```csharp\\r?\\n(.*?)\\r?\\n```", RegexOptions.Singleline);
    Require(blocks.Count == 1, "Expected exactly one canonical worked C# example");
    File.WriteAllText(Path.Combine(workspace, "Owners.cs"), blocks[0].Groups[1].Value);
    File.WriteAllText(Path.Combine(workspace, "Example.csproj"), """
        <Project Sdk="Microsoft.NET.Sdk">
          <PropertyGroup>
            <OutputType>Exe</OutputType>
            <TargetFramework>net10.0</TargetFramework>
            <LangVersion>14</LangVersion>
            <Nullable>enable</Nullable>
            <ImplicitUsings>enable</ImplicitUsings>
            <TreatWarningsAsErrors>true</TreatWarningsAsErrors>
          </PropertyGroup>
        </Project>
        """);
    File.WriteAllText(Path.Combine(workspace, "NuGet.Config"),
        "<configuration><packageSources><clear /></packageSources></configuration>");
    File.WriteAllText(Path.Combine(workspace, "Program.cs"), """
        using OwnershipExample;

        static void Check(bool value, string message)
        {
            if (!value)
            {
                throw new Exception(message);
            }
        }
        var first = new Graph();
        var second = new Graph();
        var node = new Node(first);
        Check(first.nodes.Contains(node) && ReferenceEquals(node.parent, first), "Constructor must attach both ends");
        bool refused = false;
        try
        {
            second.AddNode(node);
        }
        catch (InvalidOperationException)
        {
            refused = true;
        }
        Check(refused && first.nodes.Contains(node) && second.nodes.Count == 0
            && ReferenceEquals(node.parent, first), "Second-parent refusal must change neither end");
        first.RemoveNode(node);
        Check(node.parent is null && !node.destroyed && first.nodes.Count == 0, "Remove detaches without destruction");
        second.AddNode(node);
        node.Destroy();
        Check(node.parent is null && node.destroyed && second.nodes.Count == 0 && !second.destroyed,
            "Direct child destruction must detach but preserve its parent");
        var survivor = new Node(second);
        second.Destroy();
        Check(survivor.destroyed && survivor.parent is null && second.nodes.Count == 0,
            "Parent destruction must cascade through complete child operations");
        refused = false;
        try
        {
            _ = new Node(second);
        }
        catch (ObjectDisposedException)
        {
            refused = true;
        }
        Check(refused && second.nodes.Count == 0, "Refused construction must leave no membership");
        first.Destroy();
        Console.WriteLine("PASS: exact skill example ownership postconditions");
        """);
    File.WriteAllText(Path.Combine(workspace, "hello.go"),
        "package main\nimport \"fmt\"\nfunc main() { fmt.Println(\"PASS: offline Go recipe\") }\n");
    // These are test-selected paths, not arbitrary template code from a student.
    string recipe = File.ReadAllText(Path.Combine(repository, "flint-book/setup/env.sh"));
    foreach (string path in new[] { workspace, sdk, go })
    {
        Require(!path.Contains('\'') && !path.Contains('\n'), "Fixture paths must fit single-quoted template values");
    }
    recipe = recipe.Replace("@PROJECT_DIR@", workspace).Replace("@GO_EXE@", go).Replace("@DOTNET_ROOT@", sdk);
    File.WriteAllText(Path.Combine(workspace, "env.sh"), recipe);
    var agent = new Agent(app, new Configuration(Vendor.Anthropic, "fixture", "http://127.0.0.1:1",
        sandboxRoot: workspace, processReadPaths: ImmutableArray.Create(sdk, goRoot)),
        dataDirectory: Path.Combine(workspace, "state"));
    // This primary-free host fixture installs only the operations it actually uses.
    _ = new Tool(agent, ToolKind.RunCommand);
    _ = new Tool(agent, ToolKind.WaitForJob);
    var success = await Command("""
        . ./env.sh
        test "$(command -v dotnet)" = "$DOTNET_EXE"
        test "$(command -v go)" = "$GO_EXE"
        dotnet build Example.csproj --configfile NuGet.Config -p:NuGetAudit=false
        dotnet bin/Debug/net10.0/Example.dll
        go run hello.go
        """);
    Require(success.Exit == 0, "Confined preparation command failed:\n" + success.Output);
    Require(success.Output.Contains("PASS: exact skill example ownership postconditions")
        && success.Output.Contains("PASS: offline Go recipe"), "Missing behavioral markers");
    Require(!Regex.IsMatch(success.Output, "\\x1b\\[[0-9;]*[FGHJ]"),
        "Interactive build cursor movement remains enabled");
    var failure = await Command(". ./env.sh\nfalse\nprintf 'MASKED_FAILURE\\n'");
    Require(failure.Exit != 0 && !failure.Output.Contains("MASKED_FAILURE"),
        "Recipe masked a failing command");
    Console.WriteLine("PASS: compiled skill example, confined offline SDK/Go PATH, quiet build, fail-fast recipe");

    async Task<(int? Exit, string Output)> Command(string command)
    {
        var output = new StringBuilder();
        var result = await Dispatch("run_command", new { command, cwd = workspace, ai_callback_delay = 3, max_output_bytes = 16000 });
        while (true)
        {
            Require(!result.IsError, "Tool refused: " + JsonCodec.Text(result.Parts));
            output.Append(JsonCodec.Text(result.Parts));
            if (result.Job?.Status != "running")
            {
                return (result.Job?.ExitCode, output.ToString());
            }
            result = await Dispatch("wait_for_job", new { handle = result.Job.Handle, ai_callback_delay = 3, max_output_bytes = 16000 });
        }
    }

    Task<ToolResultPart> Dispatch(string name, object arguments)
    {
        return agent.DispatchAsync(new ToolCallPart(Guid.NewGuid().ToString("N"),
            new Provenance(Vendor.Anthropic, "fixture", Surface.Messages), name,
            JsonSerializer.SerializeToElement(arguments)));
    }
}
finally
{
    // Join owned processes before deleting only this disposable fixture.
    app.Destroy();
    Directory.Delete(workspace, recursive: true);
}

static void Require(bool condition, string message)
{
    if (!condition)
    {
        throw new InvalidOperationException(message);
    }
}
