using Ensemble;
using EnsembleGui;
using Application = Ensemble.Ensemble;

// Composition owns the core graph and optional server operation. No hosted
// Agent, provider credential, browser profile, or public interface is borrowed.
Application? app = null;
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
        if (flag is not ("--global-settings" or "--model-config" or "--data-dir" or "--agent-name" or "--sandbox" or "--port" or "--public-origin" or "--gui-dir" or "--vendor" or "--model" or "--base-url" or "--workspace" or "--event-log" or "--settings" or "--skills-dir" or "--primary-skill" or "--startup-skills" or "--save" or "--save-log" or "--archive-log" or "--cache-lens" or "--prompt-cache-key" or "--surface" or "--billing" or "--transcription" or "--transcription-key-file" or "--account" or "--credentials-dir" or "--codex-auth-file" or "--reasoning-effort" or "--reasoning-summary" or "--safe-mode" or "--enable-web-search") || i + 1 >= args.Length)
        {
            throw new ArgumentException("Unknown or incomplete GUI flag");
        }
        string value = args[++i];
        if (flag is "--global-settings" or "--model-config" or "--data-dir" or "--agent-name" or "--sandbox" or "--surface" or "--billing" or "--transcription" or "--transcription-key-file" or "--account" or "--credentials-dir" or "--codex-auth-file" or "--reasoning-effort" or "--reasoning-summary" or "--safe-mode" or "--enable-web-search")
        {
            if ((flag != "--sandbox" && string.IsNullOrWhiteSpace(value)) || value.StartsWith("--"))
            {
                throw new ArgumentException("Missing route/account option value");
            }
        }
        if (flag is "--port" or "--public-origin" or "--gui-dir" or "--settings")
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
    // Speech has its own explicit payer. A model API key or subscription never
    // enables transcription, and this dedicated key is not a command-line flag.
    string transcription = Selected("--transcription", "EN_TRANSCRIPTION");
    if (transcription is not ("" or "off" or "openai-api-key"))
        throw new ArgumentException("Transcription must be off or openai-api-key");
    string endpoint = Selected("--base-url", "LLM_BASE_URL");
    string sandboxRoot = Selected("--sandbox", "EN_SANDBOX_ROOT");
    string dataDirectory = Selected("--data-dir", "EN_DATA_DIR");
    string agentName = Selected("--agent-name", "EN_AGENT_NAME");
    string storageBase = sandboxRoot.Length == 0 ? Directory.GetCurrentDirectory() : Path.GetFullPath(sandboxRoot);
    string container = dataDirectory.Length == 0 ? storageBase : Path.GetFullPath(dataDirectory, storageBase);
    app = new Application(Console.Error, dataDirectory: container, agentDiagnostics: true);
    app.StartDiagnostics();
    if (transcription == "openai-api-key")
    {
        string keyFile = Selected("--transcription-key-file", "EN_TRANSCRIPTION_KEY_FILE");
        string transcriptionKey = string.IsNullOrEmpty(keyFile)
            ? Env("EN_TRANSCRIPTION_API_KEY")
            : ReadTranscriptionKey(keyFile);
        if (string.IsNullOrWhiteSpace(transcriptionKey))
            throw new ArgumentException("Explicit transcription requires its dedicated key or key file");
        _ = new VoiceTranscriber(app, transcriptionKey);
    }

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
        "none" => BillingMode.None,
        _ => throw new ArgumentException("Unknown billing mode; use api, chatgpt, codex or none")
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
    CodexCredentials? launchCodex = null;
    if (billing == BillingMode.Codex)
    {
        if (string.IsNullOrWhiteSpace(codexAuthFile))
        {
            throw new ArgumentException("Codex billing requires --codex-auth-file or EN_CODEX_AUTH_FILE");
        }
        launchCodex = new CodexCredentials(app, codexAuthFile);
    }
    string globalSettingsPath = Selected("--global-settings", "EN_GLOBAL_SETTINGS");
    if (globalSettingsPath.Length == 0)
    {
        globalSettingsPath = Path.Combine(
            Environment.GetFolderPath(Environment.SpecialFolder.UserProfile), ".en", "settings.json");
    }
    var globalSettings = new GlobalSettings(app, globalSettingsPath);
    RegisterApiOfferings(app, globalSettings);
    string modelConfigurationPath = Selected("--model-config", "EN_MODEL_CONFIG");
    if (modelConfigurationPath.Length != 0)
    {
        // Host authority supplies the file and the existing payer. The document
        // can reference these bindings, never credential locations or secrets.
        byte[] modelDocument = ReadModelConfiguration(modelConfigurationPath);
        _ = new ModelAuthentication(app, "none");
        ModelAuthentication? launchBinding = billing switch
        {
            BillingMode.None => new ModelAuthentication(app, "launch"),
            BillingMode.ApiKey when !string.IsNullOrWhiteSpace(key) => new ModelAuthentication(app, "launch", key),
            BillingMode.ChatGptPlan => new ModelAuthentication(app, "launch", launchCredentials!, accountSelection),
            BillingMode.Codex => new ModelAuthentication(app, "launch", launchCodex!),
            _ => null
        };
        if (launchBinding is not null)
        {
            // A normal configured offering is the explicit return choice. It
            // owns request policy, not a hidden backup Configuration or secrets.
            // Reserve its vendor before loading: the file cannot replace it.
            var launchVendor = new ModelVendor(app, "host-launch", "Launch configuration");
            var launchEndpoint = new ModelEndpoint(launchVendor, "launch", config.BaseUrl,
                config.Surface, launchBinding);
            string summary = config.ReasoningSummary;
            if (summary.Length == 0 && config.Surface == Surface.Responses
                && !config.Features.ReasoningEfforts.IsDefaultOrEmpty)
            {
                summary = "off";
            }
            var launchPolicy = new ModelRequestPolicy(config.DisableStreaming, config.RequireToolExplanation,
                config.DeferredTools, config.StubsToolResults, config.CacheBreakpoints,
                config.PromptCacheKey, config.ReasoningEffort, summary);
            _ = new Model(launchEndpoint, "launch", config.Model, "Return to launch",
                config.Features, config.ServingContextTokens, launchPolicy);
        }
        // Complete loading before any Agent, persisted conversation, or fleet
        // restoration. This is additive; it does not rewrite launch policy.
        app.LoadModelConfiguration(modelDocument);
    }
    // Account management changes the saved next-launch default, not this
    // immutable Configuration. No implicit credential location is searched.
    var agent = new Agent(app, config, name: agentName.Length == 0 ? "root" : agentName);
    string actorHome = agent.DataDirectory();
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
    // User information has one saved source, shared with the GUI. Skills capture
    // it at activation; restoring History does not rewrite earlier instructions.
    string settingsPath = hostFlags.GetValueOrDefault("--settings", Path.Combine(app.DataDirectory(), "settings.json"));
    var user = Gui.ReadPreferences(agent, settingsPath);
    _ = new Skills(agent, new SkillSpec(
        skillsDirectory.Length == 0 ? Path.Combine(AppContext.BaseDirectory, "skills") : skillsDirectory,
        primarySkill.Length == 0 ? "ensemble" : primarySkill,
        new Dictionary<string, string>
        {
            ["CUSTOM_VAR"] = Env("EN_CUSTOM_VAR"),
            ["HUMAN_NAME"] = string.IsNullOrWhiteSpace(user.HumanName) ? "Human" : user.HumanName,
            ["HUMAN_DESCRIPTION"] = user.HumanDescription
        }, startupSkills,
        Environment.GetEnvironmentVariables().Cast<System.Collections.DictionaryEntry>()
            // The dedicated speech key is not a skill/MCP startup variable.
            .Where(entry => (string)entry.Key != "EN_TRANSCRIPTION_API_KEY")
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
    // Load durable state before replacing this Agent's previous diagnostics.
    // These legacy overrides affect only the root, not its children or peers.
    agent.StartDiagnostics(
        Env("CH06_API_LOG") is { Length: > 0 } apiPath ? apiPath : null,
        Env("CH07_GUI_LOG") is { Length: > 0 } guiPath ? guiPath : null);
    agent.History().StartLog(log);
    // Shared GUI preferences belong to the application container. The reusable
    // GUI still borrows an Agent and accepts explicit caller-owned options.
    var guiOptions = new GuiOptions(
        Port: hostFlags.TryGetValue("--port", out var portText)
            ? int.Parse(portText, System.Globalization.CultureInfo.InvariantCulture) : 8090,
        AssetsDirectory: hostFlags.GetValueOrDefault("--gui-dir"),
        SettingsPath: settingsPath,
        PublicOrigin: hostFlags.GetValueOrDefault("--public-origin"));
    // Fleet restoration belongs to this application composition root, not the
    // reusable GUI mount, which only borrows the Agent supplied by its caller.
    // A browser-dependent peer cannot finish startup before its own relay can
    // connect. Serve first, then restore alongside it; neither borrows the
    // initial Agent's browser. The tree exposes pending peers for a dedicated tab.
    using var fleetStop = new CancellationTokenSource();
    Task serving = Gui.RunAsync(agent, guiOptions, fleetStop.Token);
    Task restoring = RestoreFleet();
    async Task RestoreFleet()
    {
        await agent.RestoreChildrenAsync(fleetStop.Token, waitForBrowser: true);
        await app.RestorePeersAsync(agent, fleetStop.Token, waitForBrowser: true);
    }
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


// Composition-root credential input, not an Agent filesystem operation. The caller
// explicitly selects this private file; no discovery, writes, or payer fallback.
static string ReadTranscriptionKey(string path)
{
    const int maximumBytes = 1024 * 1024;
    using var input = new FileStream(Path.GetFullPath(path), FileMode.Open, FileAccess.Read,
        FileShare.ReadWrite | FileShare.Delete);
    if (input.Length > maximumBytes)
        throw new InvalidDataException("Transcription key file exceeds its size limit");
    // Read one opened generation and bound growth as well as the initial length.
    byte[] bytes = new byte[maximumBytes + 1];
    int used = 0;
    while (used < bytes.Length)
    {
        int count = input.Read(bytes.AsSpan(used));
        if (count == 0) break;
        used += count;
    }
    if (used > maximumBytes)
        throw new InvalidDataException("Transcription key file exceeds its size limit");
    using var document = System.Text.Json.JsonDocument.Parse(bytes.AsMemory(0, used),
        new System.Text.Json.JsonDocumentOptions { MaxDepth = 16 });
    var root = document.RootElement;
    if (root.ValueKind != System.Text.Json.JsonValueKind.Object)
        throw new InvalidDataException("Transcription key file must contain an object");
    var names = new HashSet<string>(StringComparer.Ordinal);
    foreach (var property in root.EnumerateObject())
    {
        if (!names.Add(property.Name))
            throw new InvalidDataException("Transcription key file contains a duplicate field");
    }
    if (!root.TryGetProperty("openai_api_key", out var value)
        || value.ValueKind != System.Text.Json.JsonValueKind.String
        || string.IsNullOrWhiteSpace(value.GetString()))
        throw new InvalidDataException("Transcription key file has no OpenAI API key");
    return value.GetString()!;
}

// The host explicitly names this file; an Agent never supplies this pathname.
// Bound both the initial file size and growth through the same opened handle.
static byte[] ReadModelConfiguration(string path)
{
    const int maximumBytes = 1024 * 1024;
    using var input = new FileStream(Path.GetFullPath(path), FileMode.Open, FileAccess.Read,
        FileShare.ReadWrite | FileShare.Delete);
    if (input.Length > maximumBytes)
        throw new InvalidDataException("Model configuration exceeds 1 MiB.");
    byte[] bytes = new byte[maximumBytes + 1];
    int used = 0;
    while (used < bytes.Length)
    {
        int count = input.Read(bytes.AsSpan(used));
        if (count == 0) break;
        used += count;
    }
    if (used > maximumBytes)
        throw new InvalidDataException("Model configuration exceeds 1 MiB.");
    return bytes.AsSpan(0, used).ToArray();
}

// Host-approved API routes use the existing curated names and exact capability
// table. Presence of a key is not a provider entitlement or a network probe.
// Stable graph IDs are persisted by Agent settings and global funding defaults.
static void RegisterApiOfferings(Application app, GlobalSettings settings)
{
    var vendors = new (Vendor Vendor, string ID, string Label, string Address, Surface Surface)[]
    {
        (Vendor.OpenAI, "openai", "OpenAI", "https://api.openai.com", Surface.Responses),
        (Vendor.Anthropic, "anthropic", "Anthropic", "https://api.anthropic.com", Surface.Messages),
        (Vendor.Gemini, "gemini", "Gemini", "https://generativelanguage.googleapis.com", Surface.GenerateContent)
    };
    foreach (var row in vendors)
    {
        ModelAuthentication? binding = settings.BindApiKey(row.Vendor, row.ID + "-api");
        if (binding is null) continue;
        var vendor = new ModelVendor(app, row.ID, row.Label);
        var endpoint = new ModelEndpoint(vendor, "api", row.Address, row.Surface, binding);
        foreach (var recommendation in Engine.RecommendedModels)
        {
            if (recommendation.Vendor != row.Vendor) continue;
            Engine.LookupModel(row.Vendor, recommendation.Name, out ModelFeatures features);
            // Provider names may contain dots; graph IDs have a narrower grammar.
            string modelID = recommendation.Name.Replace('.', '_');
            _ = new Model(endpoint, modelID, recommendation.Name,
                recommendation.Label, features);
        }
    }
}
