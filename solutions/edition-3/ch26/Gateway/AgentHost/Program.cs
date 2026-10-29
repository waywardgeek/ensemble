using System.Runtime.InteropServices;
using Ensemble;
using Ensemble.Gateway;
using Ensemble.GatewayAdapter;
using Application = Ensemble.Ensemble;

// Optional composition root: Ensemble owns the actor and persistence, Gateway
// owns encrypted transport. Only this application knows both libraries.
try
{
    return await Run(args);
}
catch (Exception error)
{
    Console.Error.WriteLine($"gateway-agent: {error.Message}");
    return 1;
}

static async Task<int> Run(string[] arguments)
{
    string address = "ws://127.0.0.1:8093/v1/ws", id = "ensemble", keyPath = "ensemble.key";
    string publicKey = "", groups = "agents", directory = "gateway-agent-data";
    string vendorName = Environment.GetEnvironmentVariable("EN_VENDOR") ?? "anthropic";
    string model = Environment.GetEnvironmentVariable("EN_MODEL") ?? "";
    string endpoint = Environment.GetEnvironmentVariable("EN_BASE_URL") ?? "";
    string keyEnvironment = "EN_API_KEY";
    string? skills = null;
    string primary = "full";
    bool printKey = false, noStream = false;
    for (int index = 0; index < arguments.Length; index++)
    {
        string option = arguments[index];
        if (option == "--help")
        {
            Console.WriteLine("GatewayAgent --gateway-key BASE64 --model MODEL [--gateway URL] [--agent ID] [--key PATH] [--groups agents] [--data PATH] [--vendor anthropic|openai|google] [--base-url URL] [--api-key-env EN_API_KEY] [--skills PATH --primary full] [--no-stream] [--print-key]");
            Console.WriteLine("API-key funding only; EN_VENDOR, EN_MODEL and EN_BASE_URL supply defaults. No skills or tools are activated unless --skills is supplied. --data is an Ensemble container; the Agent home is agents/<ID>/.");
            return 0;
        }
        if (option == "--print-key") { printKey = true; continue; }
        if (option == "--no-stream") { noStream = true; continue; }
        if (++index == arguments.Length)
            throw new ArgumentException($"Missing value for {option}.");
        string value = arguments[index];
        switch (option)
        {
            case "--gateway": address = value; break;
            case "--agent": id = value; break;
            case "--key": keyPath = value; break;
            case "--gateway-key": publicKey = value; break;
            case "--groups": groups = value; break;
            case "--data": directory = value; break;
            case "--vendor": vendorName = value; break;
            case "--model": model = value; break;
            case "--base-url": endpoint = value; break;
            case "--api-key-env": keyEnvironment = value; break;
            case "--skills": skills = value; break;
            case "--primary": primary = value; break;
            default: throw new ArgumentException($"Unknown option {option}.");
        }
    }
    // Print-key is independent of provider setup and opens no network listener.
    if (printKey)
    {
        Console.WriteLine(Keys.EncodePublicKey(Keys.LoadOrCreateKeypair(keyPath).Public.AsSpan()));
        return 0;
    }
    var pin = Keys.ParsePublicKey(publicKey);
    Vendor vendor = vendorName.ToLowerInvariant() switch
    {
        "anthropic" => Vendor.Anthropic,
        "openai" => Vendor.OpenAI,
        "google" => Vendor.Gemini,
        _ => throw new ArgumentException("Vendor must be anthropic, openai, or google.")
    };
    ArgumentException.ThrowIfNullOrWhiteSpace(model);
    string apiKey = Environment.GetEnvironmentVariable(keyEnvironment) ?? "";
    if (string.IsNullOrWhiteSpace(apiKey))
        throw new ArgumentException($"Set the API-key environment variable {keyEnvironment}.");
    if (endpoint.Length == 0)
        endpoint = vendor switch
        {
            Vendor.Anthropic => "https://api.anthropic.com",
            Vendor.OpenAI => "https://api.openai.com/v1",
            _ => "https://generativelanguage.googleapis.com"
        };
    var configuration = new Configuration(vendor, model, endpoint, apiKey, disableStreaming: noStream);
    var identity = Keys.LoadOrCreateKeypair(keyPath);
    var connection = new ClientConfig(address, id, identity, pin);
    var selectedSkills = skills is null ? null : new SkillSpec(skills, primary);
    var application = new Application(Console.Error, dataDirectory: Path.GetFullPath(directory));
    Agent? actor = null;
    bool saveOnExit = false;
    PersistenceOptions? persistence = null;
    List<Exception> failures = [];
    using var cancellation = new CancellationTokenSource();
    using var interrupt = PosixSignalRegistration.Create(PosixSignal.SIGINT, context =>
    {
        context.Cancel = true;
        cancellation.Cancel();
    });
    using var terminate = PosixSignalRegistration.Create(PosixSignal.SIGTERM, context =>
    {
        context.Cancel = true;
        cancellation.Cancel();
    });
    try
    {
        Console.Error.WriteLine(configuration.FundingDescription());
        actor = new Agent(application, configuration, name: id, skillSpec: selectedSkills);
        string home = actor.DataDirectory();
        persistence = new PersistenceOptions(Path.Combine(home, "session.save.json"));
        actor.StartPersistence(persistence);
        // Restore and journal recovery must precede opening the event writer.
        actor.History().StartLog(Path.Combine(home, "events.jsonl"));
        if (actor.SkillManager() is { } manager)
            await manager.InitializeAsync(cancellation.Token);
        saveOnExit = true;
        var bridge = new Bridge();
        Task running = bridge.RunAsync(actor, connection,
            groups.Split(',', StringSplitOptions.RemoveEmptyEntries | StringSplitOptions.TrimEntries), cancellation.Token);
        try
        {
            await bridge.Ready();
            Console.WriteLine($"Gateway Agent {id} ready");
        }
        finally
        {
            // Even a failed readiness wait must collect the invocation's joined
            // outcome; never abandon callback-created reply observers.
            await running;
        }
    }
    catch (OperationCanceledException) when (cancellation.IsCancellationRequested) { }
    catch (Exception error)
    {
        failures.Add(error);
    }
    finally
    {
        cancellation.Cancel();
        if (actor is not null)
        {
            try
            {
                await actor.ShutdownAsync();
            }
            catch (Exception error)
            {
                failures.Add(error);
            }
            if (saveOnExit)
            {
                try
                {
                    actor.History().Save(persistence!);
                }
                catch (Exception error)
                {
                    failures.Add(error);
                }
            }
        }
        try
        {
            // Log while Agent-owned immutable credential facts remain reachable.
            foreach (Exception error in failures)
                application.Log($"gateway-agent: {error.Message}");
        }
        finally
        {
            try
            {
                application.Destroy();
            }
            catch (Exception error)
            {
                failures.Add(error);
                // Destruction may already have detached every diagnostic owner.
                // A type is useful failure evidence without exposing raw details.
                Console.Error.WriteLine($"gateway-agent cleanup failed: {error.GetType().Name}");
            }
        }
    }
    return failures.Count == 0 ? 0 : 1;
}
