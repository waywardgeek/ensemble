using Ensemble;
using EnsembleGui;
using Application = Ensemble.Ensemble;

// Composition owns the core graph and optional server operation. No hosted
// Agent, provider credential, browser profile, or public interface is borrowed.
Application? app = null;
StreamWriter? trace = null;
StreamWriter? diagnostics = null;
int exitCode = 0;
try
{
    string Env(string name) => Environment.GetEnvironmentVariable(name) ?? "";
    static bool PolicyBoolean(string value, bool fallback)
    {
        return value.Trim().ToLowerInvariant() switch
        {
            "" => fallback,
            "1" or "true" => true,
            "0" or "false" => false,
            _ => throw new ArgumentException("Policy values must be true, false, 1, or 0")
        };
    }

    var flags = new Dictionary<string, string>(StringComparer.Ordinal);
    var hostFlags = new Dictionary<string, string>(StringComparer.Ordinal);
    for (int i = 0; i < args.Length; i++)
    {
        string flag = args[i];
        if (flag == "--safe")
        {
            flags["--safe-mode"] = "1";
            continue;
        }
        if (flag is "--no-stream" or "--no-save" or "--gui-debug" or "--stub-tool-results" or "--cache-breakpoints")
        {
            flags[flag] = "1";
            continue;
        }
        if (flag is not ("--data-dir" or "--agent-name" or "--sandbox" or "--port" or "--gui-dir" or "--vendor" or "--model" or "--base-url" or "--workspace" or "--event-log" or "--settings" or "--skills-dir" or "--primary-skill" or "--startup-skills" or "--save" or "--save-log" or "--archive-log" or "--cache-lens" or "--prompt-cache-key" or "--surface" or "--billing" or "--account" or "--credentials-dir" or "--codex-auth-file" or "--reasoning-effort" or "--reasoning-summary" or "--safe-mode" or "--enable-web-search") || i + 1 >= args.Length)
        {
            throw new ArgumentException("Unknown or incomplete GUI flag");
        }
        string value = args[++i];
        if (flag is "--data-dir" or "--agent-name" or "--sandbox" or "--surface" or "--billing" or "--account" or "--credentials-dir" or "--codex-auth-file" or "--reasoning-effort" or "--reasoning-summary" or "--safe-mode" or "--enable-web-search")
        {
            if ((flag != "--sandbox" && string.IsNullOrWhiteSpace(value)) || value.StartsWith("--"))
            {
                throw new ArgumentException("Missing route/account option value");
            }
        }
        if (flag is "--port" or "--gui-dir" or "--settings")
        {
            hostFlags[flag] = value;
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
    string endpoint = Selected("--base-url", "LLM_BASE_URL");
    string sandboxRoot = Selected("--sandbox", "EN_SANDBOX_ROOT");
    string dataDirectory = Selected("--data-dir", "EN_DATA_DIR");
    string agentName = Selected("--agent-name", "EN_AGENT_NAME");
    string storageBase = sandboxRoot.Length == 0 ? Directory.GetCurrentDirectory() : Path.GetFullPath(sandboxRoot);
    string container = dataDirectory.Length == 0 ? storageBase : Path.GetFullPath(dataDirectory, storageBase);
    app = new Application(Console.Error, dataDirectory: container);
    Surface? surface = Selected("--surface", "EN_SURFACE") switch
    {
        "" => null,
        "chat-completions" => Surface.ChatCompletions,
        "responses" => Surface.Responses,
        _ => throw new ArgumentException("Unknown surface; use chat-completions or responses")
    };
    string credentialsDirectory = Selected("--credentials-dir", "EN_CREDENTIALS_DIR");
    string billingSelection = Selected("--billing", "EN_BILLING");
    string codexAuthFile = Selected("--codex-auth-file", "EN_CODEX_AUTH_FILE");
    // Codex and public OAuth have independent stores and renewal owners.
    // Do not inspect a public store when this launch explicitly selects Codex.
    Credentials? launchCredentials = billingSelection == "codex" || string.IsNullOrWhiteSpace(credentialsDirectory)
        ? null : new Credentials(app, credentialsDirectory);
    string? savedAccount = launchCredentials is null ? null
        : await launchCredentials.SelectedPlanAccountAsync();
    string accountSelection = Selected("--account", "EN_ACCOUNT");
    if (billingSelection.Length == 0 && savedAccount is not null)
    {
        billingSelection = "chatgpt";
    }
    if (billingSelection == "chatgpt" && accountSelection.Length == 0)
    {
        accountSelection = savedAccount ?? "";
    }
    BillingMode billing = billingSelection switch
    {
        "" or "api" => BillingMode.ApiKey,
        "chatgpt" => BillingMode.ChatGptPlan,
        "codex" => BillingMode.Codex,
        _ => throw new ArgumentException("Unknown billing mode; use api, chatgpt or codex")
    };
    var vendor = Selected("--vendor", "LLM_VENDOR") switch
    {
        "" when billing is BillingMode.ChatGptPlan or BillingMode.Codex => Vendor.OpenAI,
        "" or "anthropic" => Vendor.Anthropic,
        "openai" => Vendor.OpenAI,
        "gemini" => Vendor.Gemini,
        _ => throw new ArgumentException("Unknown LLM_VENDOR")
    };
    // Host launch defaults do not weaken the explicit library contract.
    if (billing is BillingMode.ChatGptPlan or BillingMode.Codex && surface is null)
    {
        surface = Surface.Responses;
    }
    if (endpoint.Length == 0)
    {
        endpoint = billing == BillingMode.Codex ? CodexCredentials.ServiceBase
            : throw new ArgumentException("Set LLM_BASE_URL explicitly");
    }
    // Selecting either plan route excludes ambient API keys.
    string key = billing == BillingMode.ApiKey ? Env("LLM_API_KEY") : "";
    var config = new Configuration(vendor, Selected("--model", "LLM_MODEL"), endpoint, key,
        requireToolExplanation: Env("LLM_REQUIRE_TOOL_EXPLANATION") == "true",
        disableStreaming: flags.ContainsKey("--no-stream") || Env("EN_DISABLE_STREAMING") == "1",
        deferredTools: Env("EN_DEFERRED_TOOLS") == "1",
        stubsToolResults: Selected("--stub-tool-results", "EN_STUB_TOOL_RESULTS") == "1",
        cacheBreakpoints: Selected("--cache-breakpoints", "EN_CACHE_BREAKPOINTS") == "1",
        promptCacheKey: Selected("--prompt-cache-key", "EN_PROMPT_CACHE_KEY"),
        surface: surface, billing: billing, account: accountSelection,
        reasoningEffort: Selected("--reasoning-effort", "OPENAI_REASONING_EFFORT"),
        reasoningSummary: Selected("--reasoning-summary", "OPENAI_REASONING_SUMMARY"),
        safeMode: PolicyBoolean(Selected("--safe-mode", "EN_SAFE_MODE"), false),
        enableWebSearch: PolicyBoolean(Selected("--enable-web-search", "EN_ENABLE_WEB_SEARCH"), true),
        sandboxRoot: sandboxRoot);
    if (billing == BillingMode.ChatGptPlan && string.IsNullOrWhiteSpace(credentialsDirectory))
    {
        throw new ArgumentException("Plan billing requires --credentials-dir or EN_CREDENTIALS_DIR");
    }
    if (billing == BillingMode.Codex)
    {
        if (string.IsNullOrWhiteSpace(codexAuthFile))
        {
            throw new ArgumentException("Codex billing requires --codex-auth-file or EN_CODEX_AUTH_FILE");
        }
        _ = new CodexCredentials(app, codexAuthFile);
    }
    // Account management changes the saved next-launch default, not this
    // immutable Configuration. No implicit credential location is searched.
    var agent = new Agent(app, config, name: agentName.Length == 0 ? "root" : agentName);
    string actorHome = agent.DataDirectory();
    trace = new StreamWriter(agent.Sandbox().OpenFile(
        Env("CH06_API_LOG") is { Length: > 0 } path ? path : Path.Combine(actorHome, "api.log"), FileMode.Append, FileAccess.Write))
    {
        AutoFlush = true
    };
    diagnostics = new StreamWriter(agent.Sandbox().OpenFile(Path.Combine(actorHome, "debug.log"), FileMode.Append, FileAccess.Write))
    {
        AutoFlush = true
    };
    app.AttachDiagnostics(diagnostics, trace);
    string cacheLensDirectory = Selected("--cache-lens", "EN_CACHE_LENS_DIR");
    if (cacheLensDirectory.Length != 0)
    {
        // Private diagnostics are opt-in and never part of the saved conversation.
        _ = new CacheLens(agent.Engine(), cacheLensDirectory);
    }
    agent.InstallTools();
    string log = Selected("--event-log", "CH02_LOG");
    if (log.Length == 0)
    {
        log = Path.Combine(actorHome, "events.jsonl");
    }
    string skillsDirectory = Selected("--skills-dir", "EN_SKILLS_DIR");
    string primarySkill = Selected("--primary-skill", "EN_PRIMARY_SKILL");
    var startupSkills = Selected("--startup-skills", "EN_STARTUP_SKILLS")
        .Split(',', StringSplitOptions.TrimEntries | StringSplitOptions.RemoveEmptyEntries).ToList();
    if (flags.ContainsKey("--gui-debug"))
    {
        startupSkills.Add("gui-debug");
    }
    _ = new Skills(agent, new SkillSpec(
        skillsDirectory.Length == 0 ? Path.Combine(AppContext.BaseDirectory, "skills") : skillsDirectory,
        primarySkill.Length == 0 ? "ensemble" : primarySkill,
        new Dictionary<string, string> { ["CUSTOM_VAR"] = Env("EN_CUSTOM_VAR") }, startupSkills,
        Environment.GetEnvironmentVariables().Cast<System.Collections.DictionaryEntry>()
            .ToDictionary(entry => (string)entry.Key, entry => (string)entry.Value!)));
    if (!flags.ContainsKey("--no-save") && Env("EN_DISABLE_SAVE") != "1")
    {
        string savePath = Selected("--save", "EN_SAVE_PATH");
        string retention = Selected("--save-log", "EN_SAVE_LOG");
        string archive = Selected("--archive-log", "EN_ARCHIVE_LOG");
        agent.StartPersistence(new PersistenceOptions(savePath.Length == 0 ? Path.Combine(actorHome, "save.json") : savePath,
            retention.Length == 0 ? -1 : int.Parse(retention, System.Globalization.CultureInfo.InvariantCulture),
            archive.Length == 0 ? null : archive));
    }
    agent.History().StartLog(log);
    // Shared GUI preferences belong to the application container. The reusable
    // GUI still borrows an Agent and accepts explicit caller-owned options.
    string defaultSettings = Path.Combine(app.DataDirectory(), "settings.json");
    var guiOptions = new GuiOptions(
        Port: hostFlags.TryGetValue("--port", out var portText)
            ? int.Parse(portText, System.Globalization.CultureInfo.InvariantCulture) : 8090,
        AssetsDirectory: hostFlags.GetValueOrDefault("--gui-dir"),
        SettingsPath: hostFlags.GetValueOrDefault("--settings", defaultSettings),
        TracePath: Env("CH07_GUI_LOG") is { Length: > 0 } guiLog ? guiLog : Path.Combine(actorHome, "gui.log"));
    // Fleet restoration belongs to this application composition root, not the
    // reusable GUI mount, which only borrows the Agent supplied by its caller.
    // A browser-dependent peer cannot finish startup before its own relay can
    // connect. Serve first, then restore alongside it; neither borrows the
    // initial Agent's browser. The tree exposes pending peers for a dedicated tab.
    using var fleetStop = new CancellationTokenSource();
    Task serving = Gui.RunAsync(agent, guiOptions, fleetStop.Token);
    Task restoring = app.RestorePeersAsync(agent, fleetStop.Token, waitForBrowser: true);
    Task first = await Task.WhenAny(serving, restoring);
    if (ReferenceEquals(first, restoring) && restoring.IsCompletedSuccessfully)
    {
        // Wait for serving to finish without choosing one failure before both
        // lifetimes are joined below. Task.WhenAny observes completion only.
        await Task.WhenAny(serving);
    }
    fleetStop.Cancel();
    Task joined = Task.WhenAll(serving, restoring);
    try
    {
        await joined;
    }
    catch (OperationCanceledException) when (fleetStop.IsCancellationRequested)
    {
        // Host exit cancels a peer still waiting for its independent browser.
    }
    catch when (joined.Exception is not null)
    {
        // If startup and hosting both fail, retain both causes after cleanup.
        throw joined.Exception;
    }
    await agent.ShutdownAsync();
    exitCode = 0;
}
catch (Exception error)
{
    // Report while the owning graph and its redaction policy still exist.
    // Console diagnostics contain only fixed guidance and type names; unknown
    // exception text may contain credentials, URLs or private filesystem paths.
    ReportFailure("GUI failed", error);
    exitCode = 1;
}
finally
{
    // Independent cleanup steps preserve the primary failure and attempt every
    // owned resource even when one cleanup fails. No exception is silently lost.
    try
    {
        app?.Destroy();
    }
    catch (Exception error)
    {
        ReportFailure("GUI cleanup failed", error);
        exitCode = 1;
    }
    try
    {
        trace?.Dispose();
    }
    catch (Exception error)
    {
        ReportFailure("Trace cleanup failed", error);
        exitCode = 1;
    }
    try
    {
        diagnostics?.Dispose();
    }
    catch (Exception error)
    {
        ReportFailure("Diagnostic cleanup failed", error);
        exitCode = 1;
    }
}
return exitCode;

static void ReportFailure(string prefix, Exception error)
{
    var causes = new Stack<Exception>();
    causes.Push(error);
    bool addressInUse = false;
    while (causes.TryPop(out var cause))
    {
        if (cause is System.Net.Sockets.SocketException socket
            && socket.SocketErrorCode == System.Net.Sockets.SocketError.AddressAlreadyInUse)
        {
            addressInUse = true;
        }
        if (cause is AggregateException aggregate)
        {
            foreach (var inner in aggregate.InnerExceptions)
            {
                causes.Push(inner);
            }
        }
        else if (cause.InnerException is { } inner)
        {
            causes.Push(inner);
        }
    }
    string advice = addressInUse
        ? "Loopback port is already in use. Choose an unused --port; do not stop its owner."
        : error is ArgumentException or FormatException or OverflowException
            ? "Invalid launch configuration. Check supported flags and value formats."
            : "Check private diagnostics and workspace/data permissions; no private error text is printed here.";
    Console.Error.WriteLine($"{prefix}: {error.GetType().Name}. {advice}");
}
