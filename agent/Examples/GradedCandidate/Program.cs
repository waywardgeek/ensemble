using System.Collections.Immutable;
using System.Diagnostics;
using Ensemble;
using Ensemble.Grading;
using Application = Ensemble.Ensemble;

// Trusted host entry point. A host-owned wrapper supplies all three absolute
// paths; never place that wrapper or this launcher in a student-writable tree.
// The unchanged Chapter 1 grader supplies only its fake endpoint/key/model.
try
{
    if (!OperatingSystem.IsMacOS() || args.Length != 3
        || args.Any(path => !Path.IsPathFullyQualified(path)))
    {
        throw new ArgumentException("Usage: GradedCandidate ABS_WORKSPACE ABS_EXECUTABLE ABS_DOTNET_ROOT (macOS).");
    }
    string workspace = args[0];
    string executable = args[1];
    string sdk = args[2];
    string key = Environment.GetEnvironmentVariable("ANTHROPIC_API_KEY") ?? "";
    string model = Environment.GetEnvironmentVariable("ANTHROPIC_MODEL") ?? "";
    string address = Environment.GetEnvironmentVariable("ANTHROPIC_BASE_URL") ?? "";
    if (key != "sk-ant-course-grader-fake" || string.IsNullOrWhiteSpace(model)
        || !Uri.TryCreate(address, UriKind.Absolute, out Uri? provider))
    {
        throw new ArgumentException("Expected the Chapter 1 grader's fake-provider environment.");
    }
    var app = new Application(TextWriter.Null);
    ProcessStartInfo start;
    try
    {
        // This Agent is only a policy owner. It admits no model turn or tools.
        var policy = new Configuration(Vendor.Anthropic, "grading-policy", "http://127.0.0.1:1",
            sandboxRoot: workspace, processReadPaths: ImmutableArray.Create(sdk));
        var owner = new Agent(app, policy, dataDirectory: workspace);
        start = owner.Sandbox().PrepareGradedProcess(executable, workspace, provider);
        start.Environment["DOTNET_ROOT"] = sdk;
        start.Environment["DOTNET_CLI_TELEMETRY_OPTOUT"] = "1";
        if (provider.Host == "127.0.0.1")
        {
            // Seatbelt's host-local predicate rejects IPv4-mapped IPv6 sockets
            // on this host. Match the fake's IPv4 family without widening access.
            start.Environment["DOTNET_SYSTEM_NET_DISABLEIPV6"] = "1";
        }
        start.Environment["ANTHROPIC_BASE_URL"] = provider.GetLeftPart(UriPartial.Authority);
        start.Environment["ANTHROPIC_API_URL"] = provider.GetLeftPart(UriPartial.Authority);
        start.Environment["ANTHROPIC_API_KEY"] = key;
        start.Environment["ANTHROPIC_MODEL"] = model;
        start.Environment["LLM_MODEL"] = model;
    }
    finally
    {
        // Close all host-owned resources before replacing this process. Prepared
        // launch values remain usable; no Agent work escapes into the candidate.
        app.Destroy();
    }
    ProcessReplacement.Execute(start);
    return 127;
}
catch (Exception error)
{
    // Neither candidate paths nor provider values need appear in diagnostics.
    Console.Error.WriteLine("Confined grading launch refused: " + error.GetType().Name);
    return 126;
}
