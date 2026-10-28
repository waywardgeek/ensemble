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
        if (flag is not ("--sandbox" or "--port" or "--gui-dir" or "--vendor" or "--model" or "--base-url" or "--workspace" or "--event-log" or "--settings" or "--skills-dir" or "--primary-skill" or "--startup-skills" or "--save" or "--save-log" or "--archive-log" or "--cache-lens" or "--prompt-cache-key" or "--surface" or "--billing" or "--account" or "--credentials-dir" or "--reasoning-effort" or "--reasoning-summary" or "--safe-mode" or "--enable-web-search") || i + 1 >= args.Length)
        {
            throw new ArgumentException("Unknown or incomplete GUI flag");
        }
        string value = args[++i];
        if (flag is "--sandbox" or "--surface" or "--billing" or "--account" or "--credentials-dir" or "--reasoning-effort" or "--reasoning-summary" or "--safe-mode" or "--enable-web-search")
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
    if (endpoint.Length == 0)
    {
        throw new ArgumentException("Set LLM_BASE_URL explicitly");
    }
    app = new Application(Console.Error);
    Surface? surface = Selected("--surface", "EN_SURFACE") switch
    {
        "" => null,
        "chat-completions" => Surface.ChatCompletions,
        "responses" => Surface.Responses,
        _ => throw new ArgumentException("Unknown surface; use chat-completions or responses")
    };
    string credentialsDirectory = Selected("--credentials-dir", "EN_CREDENTIALS_DIR");
    Credentials? launchCredentials = string.IsNullOrWhiteSpace(credentialsDirectory)
        ? null : new Credentials(app, credentialsDirectory);
    string? savedAccount = launchCredentials is null ? null
        : await launchCredentials.SelectedPlanAccountAsync();
    string billingSelection = Selected("--billing", "EN_BILLING");
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
        _ => throw new ArgumentException("Unknown billing mode; use api or chatgpt")
    };
    var vendor = Selected("--vendor", "LLM_VENDOR") switch
    {
        "" when billing == BillingMode.ChatGptPlan => Vendor.OpenAI,
        "" or "anthropic" => Vendor.Anthropic,
        "openai" => Vendor.OpenAI,
        "gemini" => Vendor.Gemini,
        _ => throw new ArgumentException("Unknown LLM_VENDOR")
    };
    // Host launch defaults do not weaken the explicit library contract.
    if (billing == BillingMode.ChatGptPlan && surface is null)
    {
        surface = Surface.Responses;
    }
    // Selecting plan billing excludes API keys even when the process inherits one.
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
        sandboxRoot: Selected("--sandbox", "EN_SANDBOX_ROOT"));
    if (billing == BillingMode.ChatGptPlan && string.IsNullOrWhiteSpace(credentialsDirectory))
    {
        throw new ArgumentException("Plan billing requires --credentials-dir or EN_CREDENTIALS_DIR");
    }
    // Account management changes the saved next-launch default, not this
    // immutable Configuration. No implicit credential location is searched.
    var agent = new Agent(app, config, name: "root");
    trace = new StreamWriter(agent.Sandbox().OpenFile(
        Env("CH06_API_LOG") is { Length: > 0 } path ? path : "api.log", FileMode.Append, FileAccess.Write))
    {
        AutoFlush = true
    };
    diagnostics = new StreamWriter(agent.Sandbox().OpenFile("debug.log", FileMode.Append, FileAccess.Write))
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
        log = "event.log";
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
        agent.StartPersistence(new PersistenceOptions(savePath.Length == 0 ? "save.json" : savePath,
            retention.Length == 0 ? -1 : int.Parse(retention, System.Globalization.CultureInfo.InvariantCulture),
            archive.Length == 0 ? null : archive));
    }
    agent.History().StartLog(log);
    // Environment and user-home defaults belong to this executable, not the
    // reusable GUI library. An outside host supplies its own explicit options.
    string defaultSettings = agent.Sandbox().Root().Length == 0
        ? Path.Combine(Environment.GetFolderPath(Environment.SpecialFolder.UserProfile), ".ensemble", "settings.json")
        : "settings.json";
    var guiOptions = new GuiOptions(
        Port: hostFlags.TryGetValue("--port", out var portText)
            ? int.Parse(portText, System.Globalization.CultureInfo.InvariantCulture) : 8090,
        AssetsDirectory: hostFlags.GetValueOrDefault("--gui-dir"),
        SettingsPath: hostFlags.GetValueOrDefault("--settings", defaultSettings),
        TracePath: Env("CH07_GUI_LOG") is { Length: > 0 } guiLog ? guiLog : "gui.log");
    await Gui.RunAsync(agent, guiOptions);
    await agent.ShutdownAsync();
    return 0;
}
catch (Exception error) when (error is ArgumentException or IOException or InvalidDataException or InvalidOperationException or FormatException or OverflowException)
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
