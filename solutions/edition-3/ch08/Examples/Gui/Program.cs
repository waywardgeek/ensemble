using Ensemble;
using EnsembleGui;
using Application = Ensemble.Ensemble;

// Composition owns the core graph and optional server operation. No hosted
// Agent, provider credential, browser profile, or public interface is borrowed.
Application? app = null;
StreamWriter? trace = null;
StreamWriter? diagnostics = null;
try
{
    string Env(string name) => Environment.GetEnvironmentVariable(name) ?? "";
    var flags = new Dictionary<string, string>(StringComparer.Ordinal);
    var hostArgs = new List<string>();
    for (int i = 0; i < args.Length; i++)
    {
        string flag = args[i];
        if (flag == "--no-stream")
        {
            flags[flag] = "1";
            continue;
        }
        if (flag is not ("--port" or "--gui-dir" or "--vendor" or "--model" or "--base-url" or "--workspace" or "--event-log" or "--settings") || i + 1 >= args.Length)
        {
            throw new ArgumentException("Unknown or incomplete GUI flag");
        }
        string value = args[++i];
        if (flag is "--port" or "--gui-dir" or "--settings")
        {
            hostArgs.Add(flag);
            hostArgs.Add(value);
        }
        else
        {
            flags[flag] = value;
        }
    }
    string Selected(string flag, string environment) => flags.GetValueOrDefault(flag, Env(environment));
    if (flags.TryGetValue("--workspace", out var workspace))
    {
        Directory.SetCurrentDirectory(Path.GetFullPath(workspace));
    }
    var vendor = Selected("--vendor", "LLM_VENDOR") switch
    {
        "" or "anthropic" => Vendor.Anthropic,
        "openai" => Vendor.OpenAI,
        "gemini" => Vendor.Gemini,
        _ => throw new ArgumentException("Unknown LLM_VENDOR")
    };
    string endpoint = Selected("--base-url", "LLM_BASE_URL");
    if (endpoint.Length == 0)
    {
        throw new ArgumentException("Set LLM_BASE_URL explicitly");
    }
    trace = new StreamWriter(Env("CH06_API_LOG") is { Length: > 0 } path ? path : "api.log", append: true)
    {
        AutoFlush = true
    };
    diagnostics = new StreamWriter("debug.log", append: true) { AutoFlush = true };
    app = new Application(TextWriter.Synchronized(diagnostics), trace);
    var config = new Configuration(vendor, Selected("--model", "LLM_MODEL"), endpoint, Env("LLM_API_KEY"),
        requireToolExplanation: Env("LLM_REQUIRE_TOOL_EXPLANATION") == "true",
        disableStreaming: flags.ContainsKey("--no-stream") || Env("EN_DISABLE_STREAMING") == "1");
    var agent = new Agent(app, config, name: "root");
    agent.InstallTools();
    string log = Selected("--event-log", "CH02_LOG");
    if (log.Length == 0)
    {
        log = "event.log";
    }
    if (log.Length != 0)
    {
        agent.History().StartLog(log);
    }
    await Gui.RunAsync(agent, hostArgs.ToArray());
    return 0;
}
catch (Exception error) when (error is ArgumentException or IOException or InvalidOperationException or FormatException or OverflowException)
{
    Console.Error.WriteLine("GUI failed: " + error.GetType().Name);
    return 1;
}
finally
{
    try
    {
        app?.Destroy();
    }
    finally
    {
        try
        {
            trace?.Dispose();
        }
        finally
        {
            diagnostics?.Dispose();
        }
    }
}
