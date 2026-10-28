using System.Text.Json;
using System.Text.Json.Nodes;
using Ensemble;
using Application = Ensemble.Ensemble;

// The executable is composition and input/output only. It does not own a second
// transcript, replay reducer, transport, or token counter.
return await Run(args);

static string Env(params string[] names) => names.Select(Environment.GetEnvironmentVariable).FirstOrDefault(s => !string.IsNullOrEmpty(s)) ?? "";
static Configuration Configuration()
{
    var vendor = Env("LLM_VENDOR") switch
    {
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
    return new(vendor, model == "" ? "offline" : model, url, Env("LLM_API_KEY", prefix + "_API_KEY"), Env("LLM_RESOLVED_MODEL"), requireExplanation);
}
static void Print(JsonObject value) => Console.WriteLine(value.ToJsonString());
static async Task<int> Run(string[] args)
{
    var app = new Application(Console.Error);
    try
    {
        if (args.Length > 2)
        {
            throw new ArgumentException("Usage: ch03 [chat | dump [logfile] | render logfile]");
        }

        var config = Configuration();
        var agent = new Agent(app, config);
        agent.InstallTools();
        var log = Env("CH02_LOG");
        if (log == "")
        {
            log = "ch02.jsonl";
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
                Print(agent.Engine().Render(agent.History().View()));
            }

            return 0;
        }
        if (command != "" && command != "chat")
        {
            throw new ArgumentException("Unknown command");
        }

        if (args.Length > 1)
        {
            throw new ArgumentException("Unexpected command argument");
        }

        agent.History().StartLog(log);
        while (true)
        {
            if (command == "chat")
            {
                Console.Write("You: ");
                Console.Out.Flush();
            }
            var line = Console.ReadLine();
            if (line is null)
            {
                break;
            }

            if (command == "chat")
            {
                if (line is "/quit" or "quit" or "exit")
                {
                    break;
                }

                if (string.IsNullOrWhiteSpace(line))
                {
                    continue;
                }

                RequireKey(config);
                Console.WriteLine("Assistant: " + await agent.SendAsync(line));
                continue;
            }
            if (string.IsNullOrWhiteSpace(line))
            {
                continue;
            }

            using var doc = JsonDocument.Parse(line);
            var input = doc.RootElement;
            var commands = new[] { "user", "ephemeral" }.Count(k => input.TryGetProperty(k, out _));
            if (commands != 1 || input.EnumerateObject().Count() != 1)
            {
                throw new FormatException("Expected one user or ephemeral operation");
            }

            if (input.TryGetProperty("user", out var user))
            {
                RequireKey(config);
                var text = user.GetString() ?? throw new FormatException("user must be a string");
                var answer = await agent.SendAsync(text);
                Print(new JsonObject { ["assistant"] = answer });
            }
            else
            {
                var text = input.GetProperty("ephemeral").GetString()
                    ?? throw new FormatException("ephemeral must be a string");
                agent.SetEphemeral(text);
                Print(new JsonObject { ["ack"] = "ephemeral" });
            }
        }
        var usage = JsonCodec.WriteUsage(agent.History().View().Usage);
        if (command == "chat")
        {
            Console.WriteLine("Usage: " + usage.ToJsonString());
        }
        else
        {
            Print(new JsonObject { ["usage"] = usage });
        }

        return 0;
    }
    catch (Exception ex) when (ex is ArgumentException or FormatException or InvalidOperationException or JsonException or IOException or NotSupportedException or KeyNotFoundException)
    {
        Console.Error.WriteLine(ex.Message);
        return 1;
    }
    finally
    {
        // Cascade to History's writer and Engine's HTTP client on success or failure.
        app.Destroy();
    }
}
static void RequireKey(Configuration config)
{
    if (config.ApiKey.Length == 0)
    {
        throw new InvalidOperationException("No API key configured; render and dump remain available offline");
    }

    if (config.Model == "offline")
    {
        throw new InvalidOperationException("Set LLM_MODEL before sending a live request");
    }
}
