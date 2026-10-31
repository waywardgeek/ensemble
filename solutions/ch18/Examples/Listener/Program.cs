using System.Runtime.InteropServices;
using Ensemble;
using SpeechListener;

// This application never attaches to the human's browser or running Ensemble.
// Caller-selected binaries and a fresh directory are composition inputs, not tools.
Session? session = null;
using var lifetime = new CancellationTokenSource();
Console.CancelKeyPress += (_, e) => { e.Cancel = true; lifetime.Cancel(); };
using var termination = PosixSignalRegistration.Create(PosixSignal.SIGTERM, context =>
{
    context.Cancel = true;
    lifetime.Cancel();
});
try
{
    var flags = new Dictionary<string, string>(StringComparer.Ordinal);
    for (int i = 0; i < args.Length; i++)
    {
        if (args[i] is not ("--dotnet" or "--gui" or "--browser" or "--directory" or "--task" or "--timeout") || i + 1 == args.Length)
        {
            throw new ArgumentException("Expected --dotnet PATH --gui DLL --browser PATH --directory FRESH_DIR [--task TEXT] [--timeout SECONDS]");
        }
        flags.Add(args[i], args[++i]);
    }
    string Required(string name) => flags.TryGetValue(name, out string? value) && value.Length > 0
        ? value : throw new ArgumentException("Missing " + name);
    string Env(string name) => Environment.GetEnvironmentVariable(name) ?? "";
    double seconds = double.Parse(flags.GetValueOrDefault("--timeout", "300"), System.Globalization.CultureInfo.InvariantCulture);
    if (!double.IsFinite(seconds) || seconds <= 0 || seconds > 86400) throw new ArgumentException("Timeout must be positive and at most one day");
    lifetime.CancelAfter(TimeSpan.FromSeconds(seconds));
    var provider = Env("LISTENER_VENDOR") switch
    {
        "anthropic" => Vendor.Anthropic, "openai" => Vendor.OpenAI, "gemini" => Vendor.Gemini,
        _ => throw new ArgumentException("Set LISTENER_VENDOR to anthropic, openai or gemini")
    };
    const string instructions = "You are a speech-only virtual user testing another coding agent. " +
        "You have exactly five tools: say, listen, sleep, set_speech_speed, save_history. " +
        "Use say to submit one target turn, then listen repeatedly while pending is true. " +
        "A timeout is not idle. Only source-labeled heard entries are evidence of target output. " +
        "You cannot inspect the DOM, files, screenshots, target model responses or tool results. " +
        "Heard means admitted to the speech channel, not acoustic hearing. Save semantic history when asked.";
    var configuration = new Configuration(provider, Env("LISTENER_MODEL"), Env("LISTENER_BASE_URL"), Env("LISTENER_API_KEY"), instructions);
    string directory = Path.GetFullPath(Required("--directory"));
    session = new Session(configuration, directory);
    var targetEnvironment = new Dictionary<string, string>();
    foreach (string name in new[] { "LLM_VENDOR", "LLM_MODEL", "LLM_BASE_URL", "LLM_API_KEY" }) targetEnvironment[name] = Env(name);
    await session.StartAsync(Required("--dotnet"), Required("--gui"), Required("--browser"), targetEnvironment, lifetime.Token);
    if (flags.TryGetValue("--task", out string? task))
    {
        Console.WriteLine(await session.Listener().SendAsync(task, lifetime.Token));
    }
    else
    {
        // The input loop remains responsive while model work and Jobs run. EOF
        // ends this session, not merely observation of its latest request.
        Task? turn = null;
        async Task RunTurn(string instruction)
        {
            Console.WriteLine(await session.Listener().SendAsync(instruction, lifetime.Token));
        }
        while (await Console.In.ReadLineAsync(lifetime.Token) is { } line)
        {
            if (turn is not null && !turn.IsCompleted) throw new ArgumentException("Wait for the listener turn before submitting another instruction");
            if (turn is not null) await turn;
            turn = RunTurn(line);
        }
        lifetime.Cancel();
        if (turn is not null)
        {
            try { await turn; }
            catch (OperationCanceledException) { }
        }
    }
}
catch (OperationCanceledException)
{
    Console.Error.WriteLine("Listener stopped by interruption, EOF or deadline");
    Environment.ExitCode = 1;
}
catch (Exception error)
{
    Console.Error.WriteLine(error.Message);
    Environment.ExitCode = 1;
}
finally
{
    if (session is not null)
    {
        try { await session.DestroyAsync(); }
        catch (Exception error) { Console.Error.WriteLine("Listener cleanup failed: " + error.Message); Environment.ExitCode = 1; }
    }
}
