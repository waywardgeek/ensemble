using System.Globalization;
using System.Runtime.InteropServices;
using Ensemble.Gateway;
using GatewayInterview;

// Independent public-client application: no reference to Ensemble's runtime.
try
{
    var options = new Dictionary<string, string>(StringComparer.Ordinal)
    {
        ["--gateway"] = "ws://127.0.0.1:8093/v1/ws",
        ["--agent"] = "interviewer",
        ["--key"] = "interviewer.key",
        ["--gateway-key"] = "",
        ["--groups"] = "agents",
        ["--subject"] = "ensemble",
        ["--report"] = "interview.md",
        ["--timeout"] = "120"
    };
    bool printKey = false;
    for (int index = 0; index < args.Length; index++)
    {
        if (args[index] == "--print-key")
        {
            printKey = true;
        }
        else if (args[index] == "--help")
        {
            Console.WriteLine("Interviewer [--gateway URL] [--agent ID] [--key PATH] [--gateway-key PUBLIC_KEY]\n" +
                "  [--groups a,b] [--subject ID] [--report PATH] [--timeout SECONDS] [--print-key]");
            return 0;
        }
        else if (options.ContainsKey(args[index]) && index + 1 < args.Length)
        {
            string name = args[index++];
            options[name] = args[index];
        }
        else
        {
            throw new ArgumentException("Unknown option or missing value; use --help for usage.");
        }
    }
    Keypair key = Keys.LoadOrCreateKeypair(options["--key"]);
    if (printKey)
    {
        Console.WriteLine(Keys.EncodePublicKey(key.Public.AsSpan()));
        return 0;
    }
    if (!double.TryParse(options["--timeout"], NumberStyles.AllowDecimalPoint, CultureInfo.InvariantCulture,
            out double seconds) || !double.IsFinite(seconds) || seconds <= 0 || seconds > uint.MaxValue / 1000.0)
        throw new ArgumentException("--timeout must be a positive number of seconds within the timer range.");
    var config = new ClientConfig(options["--gateway"], options["--agent"], key,
        Keys.ParsePublicKey(options["--gateway-key"]));
    using var lifetime = new CancellationTokenSource();
    // Signals request cancellation, not immediate process termination. RunAsync
    // then joins the connection through its await-using boundary.
    using var interrupt = PosixSignalRegistration.Create(PosixSignal.SIGINT, signal =>
    {
        signal.Cancel = true;
        lifetime.Cancel();
    });
    using var terminate = PosixSignalRegistration.Create(PosixSignal.SIGTERM, signal =>
    {
        signal.Cancel = true;
        lifetime.Cancel();
    });
    await new Interview().RunAsync(new InterviewOptions(config, options["--subject"],
        options["--groups"].Split(','), options["--report"], TimeSpan.FromSeconds(seconds)), lifetime.Token);
    return 0;
}
catch (Exception error)
{
    Console.Error.WriteLine($"interviewer: {error.Message}");
    return 1;
}
