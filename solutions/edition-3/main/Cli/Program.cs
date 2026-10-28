using System.Net;
using System.Net.Sockets;
using System.Runtime.InteropServices;
using System.Threading.Channels;
using System.Text.Json;
using System.Text.Json.Nodes;
using Ensemble;
using Application = Ensemble.Ensemble;

// The executable is composition and input/output only. It does not own a second
// transcript, replay reducer, transport, or token counter.
return await Run(args);

static string Env(params string[] names) => names.Select(Environment.GetEnvironmentVariable).FirstOrDefault(s => !string.IsNullOrEmpty(s)) ?? "";
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

static Configuration Configuration(bool disableStreaming = false, bool stubsToolResults = false,
    bool cacheBreakpoints = false, string promptCacheKey = "",
    Surface? surface = null, BillingMode billing = BillingMode.ApiKey, string account = "",
    string reasoningEffort = "", string reasoningSummary = "", bool safeMode = false, bool enableWebSearch = true, string sandboxRoot = "")
{
    var vendor = Env("LLM_VENDOR") switch
    {
        "" when billing == BillingMode.ChatGptPlan => Vendor.OpenAI,
        "" or "anthropic" => Vendor.Anthropic,
        "openai" => Vendor.OpenAI,
        "gemini" => Vendor.Gemini,
        _ => throw new ArgumentException("Unknown LLM_VENDOR")
    };
    var prefix = vendor switch
    {
        Vendor.Anthropic => "ANTHROPIC",
        Vendor.OpenAI => "OPENAI",
        _ => "GEMINI"
    };
    var model = Env("LLM_MODEL", prefix + "_MODEL");
    var url = Env("LLM_BASE_URL", prefix + "_BASE_URL", prefix + "_API_URL");
    if (url == "")
    {
        url = vendor switch
        {
            Vendor.Anthropic => "https://api.anthropic.com",
            Vendor.OpenAI => "https://api.openai.com",
            _ => "https://generativelanguage.googleapis.com"
        };
    }

    var requireExplanation = Env("LLM_REQUIRE_TOOL_EXPLANATION") switch
    {
        "" or "false" => false,
        "true" => true,
        _ => throw new ArgumentException("LLM_REQUIRE_TOOL_EXPLANATION must be true or false")
    };
    // Billing is an explicit choice, not a fallback order. Ambient API keys
    // must never enter the plan configuration or its diagnostic redaction path.
    string key = billing == BillingMode.ApiKey ? Env("LLM_API_KEY", prefix + "_API_KEY") : "";
    return new(vendor, model == "" ? "offline" : model, url, key, Env("LLM_RESOLVED_MODEL"), requireExplanation,
        disableStreaming: disableStreaming || Env("EN_DISABLE_STREAMING") == "1",
        deferredTools: Env("EN_DEFERRED_TOOLS") == "1", stubsToolResults: stubsToolResults,
        cacheBreakpoints: cacheBreakpoints, promptCacheKey: promptCacheKey,
        surface: surface, billing: billing, account: account,
        reasoningEffort: reasoningEffort, reasoningSummary: reasoningSummary,
        safeMode: safeMode, enableWebSearch: enableWebSearch, sandboxRoot: sandboxRoot);
}
static void Print(JsonObject value)
{
    Console.WriteLine(value.ToJsonString());
    Console.Out.Flush();
}
static async Task<int> Run(string[] args)
{
    Application? app = null;
    StreamWriter? apiTrace = null;
    try
    {
        string skillsDirectory = Env("EN_SKILLS_DIR");
        string primarySkill = Env("EN_PRIMARY_SKILL");
        string startupSelection = Env("EN_STARTUP_SKILLS");
        string savePath = Env("EN_SAVE_PATH");
        string dataDirectory = Env("EN_DATA_DIR");
        string agentName = Env("EN_AGENT_NAME");
        string archivePath = Env("EN_ARCHIVE_LOG");
        string saveLog = Env("EN_SAVE_LOG");
        bool guiDebug = false;
        bool cacheBreakpoints = Env("EN_CACHE_BREAKPOINTS") == "1";
        string cacheLensDirectory = Env("EN_CACHE_LENS_DIR");
        string promptCacheKey = Env("EN_PROMPT_CACHE_KEY");
        string surfaceSelection = Env("EN_SURFACE");
        string billingSelection = Env("EN_BILLING");
        string account = Env("EN_ACCOUNT");
        string reasoningEffort = Env("OPENAI_REASONING_EFFORT");
        string reasoningSummary = Env("OPENAI_REASONING_SUMMARY");
        string credentialsDirectory = Env("EN_CREDENTIALS_DIR");
        bool stubResults = Env("EN_STUB_TOOL_RESULTS") == "1";
        string contextBudget = Env("EN_CONTEXT_BUDGET_BYTES");
        bool persistenceEnabled = Env("EN_DISABLE_SAVE") != "1";
        string sandboxRoot = Env("EN_SANDBOX_ROOT");
        string safeSelection = Env("EN_SAFE_MODE");
        string webSelection = Env("EN_ENABLE_WEB_SEARCH");
        var remaining = new List<string>();
        for (int index = 0; index < args.Length; index++)
        {
            if (args[index] is "--surface" or "--billing" or "--account" or "--credentials-dir" or "--reasoning-effort" or "--reasoning-summary")
            {
                string flag = args[index];
                if (++index == args.Length || string.IsNullOrWhiteSpace(args[index]) || args[index].StartsWith("--"))
                {
                    throw new ArgumentException("Missing route/account option value");
                }
                switch (flag)
                {
                    case "--surface": surfaceSelection = args[index]; break;
                    case "--billing": billingSelection = args[index]; break;
                    case "--account": account = args[index]; break;
                    case "--credentials-dir": credentialsDirectory = args[index]; break;
                    case "--reasoning-effort": reasoningEffort = args[index]; break;
                    case "--reasoning-summary": reasoningSummary = args[index]; break;
                }
            }
            else if (args[index] == "--sandbox")
            {
                if (++index == args.Length || args[index].StartsWith("--"))
                {
                    throw new ArgumentException("Missing sandbox root");
                }
                sandboxRoot = args[index];
            }
            else if (args[index] is "--data-dir" or "--agent-name")
            {
                string flag = args[index];
                if (++index == args.Length || args[index].StartsWith("--") || string.IsNullOrWhiteSpace(args[index]))
                {
                    throw new ArgumentException("Missing actor storage option value");
                }
                if (flag == "--data-dir")
                {
                    dataDirectory = args[index];
                }
                else
                {
                    agentName = args[index];
                }
            }
            else if (args[index] == "--safe")
            {
                safeSelection = "1";
            }
            else if (args[index] is "--safe-mode" or "--enable-web-search")
            {
                string flag = args[index];
                if (++index == args.Length || args[index].StartsWith("--") || string.IsNullOrWhiteSpace(args[index]))
                {
                    throw new ArgumentException("Missing policy boolean value");
                }
                if (flag == "--safe-mode")
                {
                    safeSelection = args[index];
                }
                else
                {
                    webSelection = args[index];
                }
            }
            else if (args[index] == "--cache-breakpoints")
            {
                cacheBreakpoints = true;
            }
            else if (args[index] is "--cache-lens" or "--prompt-cache-key")
            {
                string flag = args[index];
                if (++index == args.Length)
                {
                    throw new ArgumentException("Missing cache option value");
                }
                if (flag == "--cache-lens")
                {
                    cacheLensDirectory = args[index];
                }
                else
                {
                    promptCacheKey = args[index];
                }
            }
            else if (args[index] == "--stub-tool-results")
            {
                stubResults = true;
            }
            else if (args[index] == "--context-budget-bytes")
            {
                if (++index == args.Length)
                {
                    throw new ArgumentException("Missing context byte budget");
                }
                contextBudget = args[index];
            }
            else if (args[index] == "--no-save")
            {
                persistenceEnabled = false;
            }
            else if (args[index] == "--gui-debug")
            {
                guiDebug = true;
            }
            else if (args[index] is "--save" or "--save-log" or "--archive-log")
            {
                string flag = args[index];
                if (++index == args.Length)
                {
                    throw new ArgumentException("Missing persistence option value");
                }
                if (flag == "--save")
                {
                    savePath = args[index];
                }
                else if (flag == "--save-log")
                {
                    saveLog = args[index];
                }
                else
                {
                    archivePath = args[index];
                }
            }
            else if (args[index] is "--skills-dir" or "--primary-skill" or "--startup-skills")
            {
                string flag = args[index];
                if (++index == args.Length)
                {
                    throw new ArgumentException("Missing skill option value");
                }
                if (flag == "--skills-dir")
                {
                    skillsDirectory = args[index];
                }
                else if (flag == "--primary-skill")
                {
                    primarySkill = args[index];
                }
                else
                {
                    startupSelection = args[index];
                }
            }
            else
            {
                remaining.Add(args[index]);
            }
        }
        args = remaining.ToArray();
        if (skillsDirectory.Length == 0)
        {
            skillsDirectory = Path.Combine(AppContext.BaseDirectory, "skills");
        }
        if (primarySkill.Length == 0)
        {
            primarySkill = "ensemble";
        }
        bool disableStreaming = args.Contains("--no-stream") || args.Contains("-no-stream");
        args = args.Where(value => value is not ("--no-stream" or "-no-stream")).ToArray();
        // The application container is not the initial actor's home. Relative
        // explicit storage follows the same sandbox root as other host paths.
        string storageBase = sandboxRoot.Length == 0 ? Directory.GetCurrentDirectory() : Path.GetFullPath(sandboxRoot);
        string container = dataDirectory.Length == 0 ? storageBase : Path.GetFullPath(dataDirectory, storageBase);
        app = new Application(Console.Error, dataDirectory: container);
        if (args.Length > 2)
        {
            throw new ArgumentException("Usage: ch04 [--data-dir DIR] [--agent-name NAME] [--sandbox ROOT] [--safe | --safe-mode BOOL] [--enable-web-search BOOL] [--surface chat-completions|responses] [--billing api|chatgpt] [--account LABEL] [--credentials-dir DIR] [--no-stream] [--cache-lens DIR] [--cache-breakpoints] [--prompt-cache-key KEY] [--stub-tool-results] [--context-budget-bytes N] [chat | dump [logfile] | render logfile | auth status|models|login|logout|signout]");
        }

        // Authentication is a host operation, never a model turn or a tool.
        // Dispatch before Configuration so listing/logging out needs neither a
        // selected provider nor a valid model/key/skill configuration.
        if (args.FirstOrDefault() == "auth")
        {
            if (args.Length != 2 || args[1] is not ("status" or "models" or "login" or "logout" or "signout"))
            {
                throw new ArgumentException("Usage: ch04 --credentials-dir DIR [--account LABEL] auth status|models|login|logout|signout");
            }
            if (string.IsNullOrWhiteSpace(credentialsDirectory))
            {
                throw new ArgumentException("Authentication requires --credentials-dir or EN_CREDENTIALS_DIR");
            }
            if (args[1] != "status" && string.IsNullOrWhiteSpace(account))
            {
                throw new ArgumentException("This authentication command requires --account or EN_ACCOUNT");
            }
            var credentials = new Credentials(app, credentialsDirectory);
            return await RunAuth(credentials, args[1], account);
        }

        // Read only an explicitly supplied store. An explicit launch choice wins
        // but never overwrites the saved choice, including after local logout.
        Credentials? launchCredentials = string.IsNullOrWhiteSpace(credentialsDirectory)
            ? null : new Credentials(app, credentialsDirectory);
        string? savedAccount = launchCredentials is null ? null
            : await launchCredentials.SelectedPlanAccountAsync();
        if (billingSelection.Length == 0 && savedAccount is not null)
        {
            billingSelection = "chatgpt";
        }
        if (billingSelection == "chatgpt" && account.Length == 0)
        {
            account = savedAccount ?? "";
        }
        Surface? surface = surfaceSelection switch
        {
            "" => null,
            "chat-completions" => Surface.ChatCompletions,
            "responses" => Surface.Responses,
            _ => throw new ArgumentException("Unknown surface; use chat-completions or responses")
        };
        BillingMode billing = billingSelection switch
        {
            "" or "api" => BillingMode.ApiKey,
            "chatgpt" => BillingMode.ChatGptPlan,
            _ => throw new ArgumentException("Unknown billing mode; use api or chatgpt")
        };
        // Host launch defaults do not weaken the explicit library contract.
        if (billing == BillingMode.ChatGptPlan && surface is null)
        {
            surface = Surface.Responses;
        }
        var config = Configuration(disableStreaming, stubResults, cacheBreakpoints, promptCacheKey,
            surface, billing, account, reasoningEffort, reasoningSummary,
            PolicyBoolean(safeSelection, false), PolicyBoolean(webSelection, true), sandboxRoot);
        if (billing == BillingMode.ChatGptPlan)
        {
            if (string.IsNullOrWhiteSpace(credentialsDirectory))
            {
                throw new ArgumentException("Plan billing requires --credentials-dir or EN_CREDENTIALS_DIR");
            }
            // The root already owns the selected store. Engines obtain bearer
            // tokens through that parent path, never a host-held token copy.
        }
        Console.Error.WriteLine(config.FundingDescription());
        var agent = new Agent(app, config, name: agentName.Length == 0 ? "root" : agentName);
        string actorHome = agent.DataDirectory();
        if (args.Length == 0 || args[0] == "chat")
        {
            string tracePath = Env("CH06_API_LOG");
            apiTrace = new StreamWriter(agent.Sandbox().OpenFile(
                tracePath.Length == 0 ? Path.Combine(actorHome, "api.log") : tracePath, FileMode.Append, FileAccess.Write));
            app.AttachDiagnostics(Console.Error, apiTrace);
        }
        if (cacheLensDirectory.Length != 0)
        {
            // Captures contain private prompts. This directory is explicitly caller-owned.
            _ = new CacheLens(agent.Engine(), cacheLensDirectory);
        }
        if (contextBudget.Length != 0)
        {
            if (!int.TryParse(contextBudget, out int bytes) || bytes < 0)
            {
                throw new ArgumentException("Context byte budget must be a nonnegative integer");
            }
            agent.SetContextByteBudget(bytes);
        }
        agent.InstallTools();
        var log = Env("CH02_LOG");
        if (log == "")
        {
            log = Path.Combine(actorHome, "events.jsonl");
        }

        var command = args.FirstOrDefault() ?? "";
        if (command is "render" or "dump")
        {
            if (command == "render" && args.Length != 2)
            {
                throw new ArgumentException("render requires a log path");
            }

            agent.History().Load(args.Length == 2 ? args[1] : log);
            if (command == "dump")
            {
                agent.History().Dump(Console.Out);
            }
            else
            {
                // Replay needs the control schemas, not a new live manager or
                // a reread of mutable skill files. Recorded grants decide which
                // declarations are visible, just as they do for ordinary tools.
                _ = new Tool(agent, ToolKind.LoadSkill);
                _ = new Tool(agent, ToolKind.UnloadSkill);
                Print(agent.Engine().Render(agent.History().View()));
            }

            return 0;
        }
        if (command != "" && command != "chat" && command != "--mcp-pipe")
        {
            throw new ArgumentException("Unknown command");
        }

        if (args.Length > 1)
        {
            throw new ArgumentException("Unexpected command argument");
        }

        var startupSkills = startupSelection.Split(',', StringSplitOptions.TrimEntries | StringSplitOptions.RemoveEmptyEntries).ToList();
        if (guiDebug)
        {
            startupSkills.Add("gui-debug");
        }
        _ = new Skills(agent, new SkillSpec(skillsDirectory, primarySkill,
            new Dictionary<string, string> { ["CUSTOM_VAR"] = Env("EN_CUSTOM_VAR") },
            startupSkills, Environment.GetEnvironmentVariables()
                .Cast<System.Collections.DictionaryEntry>()
                .ToDictionary(entry => (string)entry.Key, entry => (string)entry.Value!)));
        if (persistenceEnabled)
        {
            string selectedPath = savePath.Length == 0 ? Path.Combine(actorHome, "save.json") : savePath;
            int retained = saveLog.Length == 0 ? -1 : int.Parse(saveLog, System.Globalization.CultureInfo.InvariantCulture);
            agent.StartPersistence(new PersistenceOptions(selectedPath, retained,
                archivePath.Length == 0 ? null : archivePath));
        }
        agent.History().StartLog(log);
        await agent.SkillManager()!.InitializeAsync();
        if (command == "--mcp-pipe")
        {
            // This executable is the incoming service. Its external client
            // initiates initialize/initialized; no conversation or HTTP starts.
            // The same Agent admission, Jobs and skill grants handle tool calls.
            var manager = agent.Mcp() ?? new Mcp(agent);
            var transport = new McpStreamTransport(Console.OpenStandardInput(), Console.OpenStandardOutput());
            var service = new McpClient(manager, "stdin-service", transport);
            try
            {
                await service.ServeAsync();
            }
            finally
            {
                await service.DestroyAsync();
                await agent.ShutdownAsync();
            }
            return 0;
        }
        JsonObject usage;
        if (command == "chat")
        {
            // Stdin never waits for a model turn. The actor decides whether a
            // line starts a turn or hints the current one; this queue only
            // prints reliable answers in admission order.
            var replies = Channel.CreateUnbounded<Task<JsonObject>>(new UnboundedChannelOptions
            {
                SingleReader = true,
                SingleWriter = true
            });
            Task writer = WriteReplies(replies.Reader, chat: true);
            Task preview = WriteChatProgress(agent);
            using var questionCancellation = new CancellationTokenSource();
            Task questions = WriteHumanQuestions(agent, chat: true, questionCancellation.Token);
            bool quit = false;
            try
            {
                while (true)
                {
                    Console.Error.Write("You: ");
                    Console.Error.Flush();
                    string? line = await Console.In.ReadLineAsync();
                    if (line is null)
                    {
                        break;
                    }
                    if (line is "/quit" or "quit" or "exit")
                    {
                        quit = true;
                        break;
                    }
                    if (string.IsNullOrWhiteSpace(line))
                    {
                        continue;
                    }
                    if (line == "/reply" || line.StartsWith("/reply ", StringComparison.Ordinal))
                    {
                        try
                        {
                            string response = line[6..].Trim();
                            if (response.Length == 0)
                            {
                                throw new ArgumentException("Use /reply TEXT or /reply MESSAGE_ID TEXT.");
                            }
                            long? id = null;
                            int separator = response.IndexOf(' ');
                            if (separator > 0 && long.TryParse(response[..separator], out long parsed))
                            {
                                id = parsed;
                                response = response[(separator + 1)..].Trim();
                            }
                            agent.RespondToQuestion(response, messageID: id);
                            Console.Error.WriteLine("Reply delivered.");
                        }
                        catch (Exception error) when (error is ArgumentException or InvalidOperationException)
                        {
                            Console.Error.WriteLine("Reply refused: " + error.Message);
                        }
                        continue;
                    }
                    if (line == "/model" || line.StartsWith("/model ", StringComparison.Ordinal))
                    {
                        if (line == "/model")
                        {
                            Console.Error.WriteLine("Model: " + agent.Configuration().Model);
                            continue;
                        }
                        try
                        {
                            await agent.SetModelAsync(line[7..].Trim());
                            Console.Error.WriteLine(agent.Configuration().FundingDescription());
                        }
                        catch (Exception error) when (error is ArgumentException or InvalidOperationException)
                        {
                            Console.Error.WriteLine("Model selection refused: " + error.Message);
                        }
                        continue;
                    }
                    RequireKey(agent.Configuration());
                    replies.Writer.TryWrite(AssistantReply(agent.PostAsync(new AgentInput(InputKind.Chat, line))));
                }
                if (!quit)
                {
                    await agent.FlushAsync(); // EOF drains; /quit explicitly stops.
                }
                usage = JsonCodec.WriteUsage(agent.History().View().Usage);
            }
            finally
            {
                questionCancellation.Cancel();
                try
                {
                    await agent.ShutdownAsync();
                }
                finally
                {
                    replies.Writer.TryComplete();
                    try
                    {
                        agent.Destroy();
                    }
                    finally
                    {
                        await Task.WhenAll(preview, questions, writer);
                    }
                }
            }
        }
        else
        {
            usage = await RunJson(agent);
        }
        if (command == "chat")
        {
            Console.Error.WriteLine("Usage: " + usage.ToJsonString());
        }
        else
        {
            Print(new JsonObject { ["usage"] = usage });
        }

        return 0;
    }
    catch (Exception ex) when (ex is ArgumentException or FormatException or InvalidOperationException or JsonException or IOException or UnauthorizedAccessException or InvalidDataException or OverflowException or NotSupportedException or KeyNotFoundException)
    {
        Console.Error.WriteLine(ex.Message);
        return 1;
    }
    finally
    {
        // Cascade to History's writer and Engine's HTTP client on success or failure.
        try
        {
            app?.Destroy();
        }
        finally
        {
            apiTrace?.Dispose();
        }
    }
}
// Auth output is private host output, not the JSON conversation protocol. The
// browser URL may carry an ID-token hint. Never send it through Print, History,
// an Agent, a trace writer, or a generic exception message.
static async Task<int> RunAuth(Credentials credentials, string operation, string account)
{
    using var stopping = new CancellationTokenSource(TimeSpan.FromMinutes(10));
    ConsoleCancelEventHandler interrupt = (_, signal) =>
    {
        signal.Cancel = true;
        stopping.Cancel();
    };
    Console.CancelKeyPress += interrupt;
    using var terminate = OperatingSystem.IsWindows() ? null :
        PosixSignalRegistration.Create(PosixSignal.SIGTERM, signal =>
        {
            signal.Cancel = true;
            stopping.Cancel();
        });
    try
    {
        if (operation == "status")
        {
            var accounts = await credentials.AccountsAsync(stopping.Token);
            // This explicit public value contains only redacted account metadata.
            Console.WriteLine(JsonSerializer.Serialize(accounts));
        }
        else if (operation == "models")
        {
            string endpoint = Env("LLM_BASE_URL", "OPENAI_BASE_URL", "OPENAI_API_URL");
            if (endpoint.Length == 0)
            {
                endpoint = "https://api.openai.com";
            }
            var models = await credentials.ModelsAsync(account, endpoint, stopping.Token);
            Console.WriteLine(JsonSerializer.Serialize(models));
        }
        else if (operation == "logout")
        {
            await credentials.LogoutAsync(account, stopping.Token);
            Console.WriteLine("Local tokens cleared; remote revocation was not requested.");
        }
        else if (operation == "signout")
        {
            bool revoked = await credentials.SignOutAsync(account, stopping.Token);
            Console.WriteLine(revoked
                ? "Local tokens cleared; remote sign-out confirmed."
                : "Local tokens cleared; remote revocation was not confirmed.");
        }
        else
        {
            await LoginFromConsole(credentials, account, stopping.Token);
        }
        return 0;
    }
    catch (OperationCanceledException) when (stopping.IsCancellationRequested)
    {
        Console.Error.WriteLine("Authentication cancelled or timed out.");
        return 1;
    }
    catch (Exception error) when (error is ArgumentException or InvalidOperationException
        or IOException or InvalidDataException or HttpRequestException or HttpListenerException
        or NotSupportedException or JsonException or System.Security.Cryptography.CryptographicException)
    {
        // Auth exceptions can originate in HTTP/file parsers. Do not print raw
        // messages which could contain a private callback URL or issuer body.
        Console.Error.WriteLine("Authentication operation failed. Check the private store and retry.");
        return 1;
    }
    finally
    {
        Console.CancelKeyPress -= interrupt;
        await credentials.CancelLoginAsync();
    }
}

static async Task LoginFromConsole(Credentials credentials, string account, CancellationToken cancellation)
{
    // HttpListener cannot allocate port zero. Probe a candidate, then require a
    // successful real bind BEFORE creating a login. If another process wins the
    // race, Start throws and no callback URL or authorization has been issued.
    int port;
    using (var candidate = new TcpListener(IPAddress.Loopback, 0))
    {
        candidate.Start();
        port = ((IPEndPoint)candidate.LocalEndpoint).Port;
    }
    var redirect = new Uri($"http://127.0.0.1:{port}/auth/callback");
    using var listener = new HttpListener();
    listener.Prefixes.Add($"http://127.0.0.1:{port}/");
    listener.Start();
    var authorization = await credentials.BeginLoginAsync(account, redirect, cancellation: cancellation);
    Console.Error.WriteLine("Open this private sign-in URL locally. Do not share it or paste it into a conversation:");
    Console.Error.WriteLine(authorization.AbsoluteUri);
    Console.Error.Flush();
    while (true)
    {
        // Disposing the listener below releases an outstanding accept even when
        // WaitAsync is cancelled. The login and listener have the same lifetime.
        var request = await listener.GetContextAsync().WaitAsync(cancellation);
        if (request.Request.HttpMethod != "GET" || request.Request.Url?.Host != "127.0.0.1"
            || request.Request.Url.Port != port || request.Request.Url.AbsolutePath != "/auth/callback")
        {
            await AuthReply(request.Response, 404, "Not found.");
            continue;
        }
        AuthAccount status;
        try
        {
            status = await credentials.CompleteLoginAsync(request.Request.Url.AbsoluteUri, cancellation);
        }
        catch (Exception error) when (error is ArgumentException or InvalidOperationException
            or InvalidDataException or IOException or HttpRequestException or JsonException
            or System.Security.Cryptography.CryptographicException)
        {
            await AuthReply(request.Response, 400, "Sign-in was not completed. Return to Ensemble.");
            if (await credentials.LoginPendingAsync(cancellation))
            {
                continue;
            }
            throw new InvalidOperationException("Sign-in was not completed.");
        }
        await AuthReply(request.Response, 200, "Sign-in completed. You may close this window.");
        Console.WriteLine(status.PlanEnabled
            ? "Signed in; ChatGPT-plan permission granted."
            : "Signed in, but ChatGPT-plan permission was not granted. Plan inference remains disabled.");
        Console.WriteLine("ChatGPT plan funding is selected for future launches using this store. "
            + "Logout preserves that selection; API billing requires an explicit override.");
        return;
    }
}

static async Task AuthReply(HttpListenerResponse response, int status, string message)
{
    // A callback never serves scripts, reflects parameters, or links elsewhere.
    // Its address contains a one-use code: prohibit caching and referrer leakage.
    response.StatusCode = status;
    response.ContentType = "text/plain; charset=utf-8";
    response.Headers["Cache-Control"] = "no-store";
    response.Headers["Referrer-Policy"] = "no-referrer";
    response.Headers["Content-Security-Policy"] = "default-src 'none'; frame-ancestors 'none'";
    byte[] body = System.Text.Encoding.UTF8.GetBytes(message);
    response.ContentLength64 = body.Length;
    try
    {
        await response.OutputStream.WriteAsync(body);
    }
    catch (Exception error) when (error is IOException or HttpListenerException or ObjectDisposedException)
    {
        // Closing the browser must not undo a grant already saved by Credentials.
    }
    finally
    {
        response.Close();
    }
}

// Read input independently of request completion. This queue owns only pending
// output Tasks; the Agent owns scheduling, History, and every reliable reply.
// Display only. Neither streamed argument fragments nor display text may
// trigger tools or supply the reliable response returned by SendAsync.
static async Task WriteChatProgress(Agent agent)
{
    bool color = !Console.IsErrorRedirected;
    try
    {
        while (true)
        {
            var observation = await agent.WaitAsync(_ => true);
            if (observation.Delta is Chunk chunk)
            {
                string style = chunk.Kind switch
                {
                    ChunkKind.Thinking => "\u001b[2m",
                    ChunkKind.ToolCall => "\u001b[33m",
                    _ => ""
                };
                string text = chunk.Kind == ChunkKind.ToolCall
                    ? chunk.ToolName + chunk.ArgsText : chunk.Text;
                if (color && style.Length != 0)
                {
                    Console.Error.Write(style);
                }
                Console.Error.Write(text);
                if (color && style.Length != 0)
                {
                    Console.Error.Write("\u001b[0m");
                }
                Console.Error.Flush();
            }
            else if (observation.Kind == ObservationKind.PartFinal && observation.PartID is not null)
            {
                Console.Error.WriteLine();
            }
        }
    }
    catch (ObjectDisposedException)
    {
        // Destruction closes the queue; queued previews have now been drained.
    }
}

static async Task<JsonObject> RunJson(Agent agent)
{
    var replies = Channel.CreateUnbounded<Task<JsonObject>>(new UnboundedChannelOptions
    {
        SingleReader = true,
        SingleWriter = true
    });
    var writer = WriteReplies(replies.Reader);
    using var observing = new CancellationTokenSource();
    Task? progress = null;
    using var questionCancellation = new CancellationTokenSource();
    Task questions = WriteHumanQuestions(agent, chat: false, questionCancellation.Token);
    bool shutdown = false;
    JsonObject? usage = null;
    try
    {
        while (await Console.In.ReadLineAsync() is string line)
        {
            if (string.IsNullOrWhiteSpace(line))
            {
                continue;
            }
            using var doc = JsonDocument.Parse(line);
            var input = doc.RootElement;
            if (input.ValueKind != JsonValueKind.Object)
            {
                throw new FormatException("Expected a JSON input object");
            }
            if (!input.TryGetProperty("kind", out var kindValue))
            {
                if (input.TryGetProperty("user", out var user))
                {
                    Fields(input, "user", "parts");
                    RequireKey(agent.Configuration());
                    // Structured parts take precedence, including an empty list.
                    // The public codec is shared with replay; no CLI media model.
                    Part[] parts = input.TryGetProperty("parts", out var values)
                        ? values.EnumerateArray().Select(JsonCodec.ReadPart).ToArray()
                        : [new TextPart(user.GetString() ?? throw new FormatException("user must be a string"))];
                    replies.Writer.TryWrite(AssistantReply(agent.SubmitAsync(parts)));
                }
                else
                {
                    Fields(input, "ephemeral");
                    string text = RequiredText(input, "ephemeral");
                    await agent.PostAsync(new AgentInput(InputKind.Ephemeral, text));
                    Print(new JsonObject { ["ack"] = "ephemeral" });
                }
                continue;
            }

            string kind = kindValue.GetString() ?? throw new FormatException("kind must be a string");
            // Only explicit live mode enables progress output. Legacy users
            // still receive exactly one assistant record per request.
            progress ??= WriteProgress(agent, observing.Token);
            switch (kind)
            {
                case "prompt":
                case "hint":
                    Fields(input, "kind", "text");
                    RequireKey(agent.Configuration());
                    var messageKind = kind == "prompt" ? InputKind.Message : InputKind.Hint;
                    await agent.PostAsync(new AgentInput(messageKind, RequiredText(input, "text")));
                    break;
                case "pause":
                case "resume":
                    Fields(input, "kind");
                    await agent.PostAsync(new AgentInput(kind == "pause" ? InputKind.Pause : InputKind.Resume));
                    break;
                case "interrupt":
                    Fields(input, "kind", "mode");
                    string mode = input.TryGetProperty("mode", out var modeValue)
                        ? modeValue.GetString() ?? "" : "now";
                    var interrupt = mode switch
                    {
                        "now" => InputKind.InterruptNow,
                        "after_tool" => InputKind.InterruptAfterTool,
                        _ => throw new FormatException("interrupt mode must be now or after_tool")
                    };
                    await agent.PostAsync(new AgentInput(interrupt));
                    break;
                case "tool":
                    Fields(input, "kind", "id", "name", "arguments");
                    var call = new ToolCallPart(RequiredText(input, "id"), agent.Configuration().Target(),
                        RequiredText(input, "name"), input.GetProperty("arguments"));
                    replies.Writer.TryWrite(ToolReply(agent.DispatchAsync(call)));
                    break;
                case "set_model":
                    Fields(input, "kind", "model");
                    try
                    {
                        await agent.SetModelAsync(RequiredText(input, "model"));
                        Print(new JsonObject { ["ack"] = kind, ["model"] = agent.Configuration().Model });
                    }
                    catch (Exception error) when (error is ArgumentException or InvalidOperationException)
                    {
                        // Busy/invalid selection is a control refusal, not a
                        // failed conversation or a reason to terminate stdin.
                        Print(new JsonObject { ["error"] = "Model selection refused: " + error.Message,
                            ["model"] = agent.Configuration().Model });
                    }
                    continue;
                case "reply":
                    Fields(input, "kind", "text", "message_id");
                    try
                    {
                        long? id = input.TryGetProperty("message_id", out var idValue)
                            ? idValue.GetInt64() : null;
                        agent.RespondToQuestion(RequiredText(input, "text"), messageID: id);
                        Print(new JsonObject { ["ack"] = kind, ["message_id"] = id });
                    }
                    catch (Exception error) when (error is ArgumentException or InvalidOperationException or FormatException)
                    {
                        Print(new JsonObject { ["error"] = "Reply refused: " + error.Message });
                    }
                    continue;
                case "clear_tools":
                    Fields(input, "kind");
                    await agent.PostAsync(new AgentInput(InputKind.ClearTools));
                    break;
                case "shutdown":
                    Fields(input, "kind");
                    await agent.ShutdownAsync();
                    shutdown = true;
                    break;
                default:
                    throw new FormatException("Unknown input kind: " + kind);
            }
            Print(new JsonObject { ["ack"] = kind });
            if (shutdown)
            {
                break;
            }
        }
        if (!shutdown)
        {
            // EOF drains turns; it is not an implicit interrupt. Paused work
            // stays paused, so send resume or shutdown before ending such input.
            await agent.FlushAsync();
        }
    }
    finally
    {
        questionCancellation.Cancel();
        // On malformed input or read failure, shutdown unblocks pending replies
        // before we join the output writer. Independent Jobs are stopped here.
        try
        {
            await agent.ShutdownAsync();
        }
        finally
        {
            replies.Writer.TryComplete();
            usage = JsonCodec.WriteUsage(agent.History().View().Usage);
            // Closing the writer lets display drain queued observations. Reliable
            // completion alone does not prove that a slower reader caught up.
            try
            {
                agent.Destroy();
            }
            finally
            {
                if (progress is not null)
                {
                    await progress;
                }
                await Task.WhenAll(questions, writer);
            }
        }
    }
    return usage!;
}

static void Fields(JsonElement input, params string[] allowed)
{
    var seen = new HashSet<string>(StringComparer.Ordinal);
    foreach (var field in input.EnumerateObject())
    {
        if (!allowed.Contains(field.Name) || !seen.Add(field.Name))
        {
            throw new FormatException("Unexpected or duplicate input field: " + field.Name);
        }
    }
}

static string RequiredText(JsonElement input, string name)
{
    return input.GetProperty(name).GetString()
        ?? throw new FormatException(name + " must be a string");
}

static async Task<JsonObject> AssistantReply(Task<string> reply)
{
    return new JsonObject { ["assistant"] = await reply };
}

static async Task<JsonObject> ToolReply(Task<ToolResultPart> reply)
{
    return new JsonObject { ["tool_result"] = JsonCodec.WritePart(await reply) };
}

// The human inbox is authoritative and separate from the lossy progress queue.
// IDs here track display only; replying never removes entries from a UI cache.
static async Task WriteHumanQuestions(Agent agent, bool chat, CancellationToken cancellation)
{
    var displayed = new HashSet<long>();
    try
    {
        while (true)
        {
            var questions = await agent.WaitForQuestionsAsync(humanOnly: true, cancellation, displayed);
            foreach (QuestionInfo question in questions)
            {
                displayed.Add(question.MessageID);
                if (chat)
                {
                    Console.Error.WriteLine($"Question {question.MessageID}: {question.Content}");
                    Console.Error.WriteLine($"Use /reply {question.MessageID} TEXT (or /reply TEXT for the sole question).");
                }
                else
                {
                    Print(new JsonObject
                    {
                        ["human_question"] = new JsonObject
                        {
                            ["message_id"] = question.MessageID,
                            ["content"] = question.Content
                        }
                    });
                }
            }
        }
    }
    catch (OperationCanceledException) when (cancellation.IsCancellationRequested) { }
    catch (ObjectDisposedException) { }
}

static async Task WriteReplies(ChannelReader<Task<JsonObject>> replies, bool chat = false)
{
    bool failed = false;
    await foreach (var reply in replies.ReadAllAsync())
    {
        try
        {
            JsonObject result = await reply;
            if (chat)
            {
                string? answer = result["assistant"]?.GetValue<string>();
                // Busy chat input is an acknowledged hint, not another answer.
                if (!string.IsNullOrEmpty(answer))
                {
                    Console.WriteLine(answer);
                    Console.Out.Flush();
                }
            }
            else
            {
                Print(result);
            }
        }
        catch (Exception error) when (error is OperationCanceledException or ObjectDisposedException)
        {
            // Interrupt/shutdown is an explicit terminal outcome, not success.
            string message = error is OperationCanceledException
                ? "Turn interrupted." : "Agent shut down before completion.";
            if (chat)
            {
                Console.Error.WriteLine(message);
            }
            else
            {
                Print(new JsonObject { ["error"] = message });
            }
        }
        catch (Exception error) when (error is ArgumentException or InvalidOperationException
            or IOException or JsonException or NotSupportedException)
        {
            failed = true;
            Console.Error.WriteLine(error.Message);
            if (!chat)
            {
                Print(new JsonObject { ["error"] = error.Message });
            }
        }
    }
    if (failed)
    {
        throw new InvalidOperationException("One or more requests failed.");
    }
}

static async Task WriteProgress(Agent agent, CancellationToken cancellation)
{
    try
    {
        while (true)
        {
            var value = await agent.WaitAsync(_ => true, cancellation);
            string kind = JsonNamingPolicy.SnakeCaseLower.ConvertName(value.Kind.ToString());
            if (value.Kind == ObservationKind.PartFinal && value.PartID is null)
            {
                kind = value.Part is ToolResultPart ? "tool_result" : "input_part";
            }
            var record = new JsonObject
            {
                ["observation"] = kind,
                ["seq"] = value.Seq,
                ["text"] = value.Text,
                ["accepted"] = value.Accepted,
                ["from"] = value.From is Phase from ? JsonNamingPolicy.SnakeCaseLower.ConvertName(from.ToString()) : null,
                ["to"] = value.To is Phase to ? JsonNamingPolicy.SnakeCaseLower.ConvertName(to.ToString()) : null,
                ["part"] = value.Part is null ? null : JsonCodec.WritePart(value.Part)
            };
            if (value.PartID is int id)
            {
                record["part_id"] = id;
                record["response_id"] = value.ResponseID;
            }
            if (value.Delta is Chunk delta)
            {
                record["kind"] = delta.Kind switch
                {
                    ChunkKind.Reply => "text",
                    ChunkKind.Thinking => "thinking",
                    _ => "tool_call"
                };
                record["chunk"] = delta.Kind == ChunkKind.ToolCall
                    ? delta.ToolName + delta.ArgsText : delta.Text;
            }
            if (value.Succeeded is bool succeeded)
            {
                record["succeeded"] = succeeded;
                if (succeeded && !value.HasReply)
                {
                    record["assistant"] = value.Text;
                }
            }
            Print(record);
        }
    }
    catch (OperationCanceledException) when (cancellation.IsCancellationRequested) { }
    catch (ObjectDisposedException)
    {
        // Destroy closes the stream after shutdown; buffered records were read.
    }
}

static void RequireKey(Configuration config)
{
    if (config.Billing == BillingMode.ApiKey && config.ApiKey.Length == 0)
    {
        throw new InvalidOperationException("No API key configured; render and dump remain available offline");
    }

    if (config.Model == "offline")
    {
        throw new InvalidOperationException("Set LLM_MODEL before sending a live request");
    }
}
