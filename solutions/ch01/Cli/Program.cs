using System.Text.Json;
using Ensemble;

// Composition belongs here, not inside the provider adapter. Stdout is reserved
// for protocol records in default mode; diagnostics always go to stderr.
bool chat = args.Length == 1 && args[0] == "chat";
if (args.Length != 0 && !chat)
{
    Console.Error.WriteLine("Usage: ch01 [chat]");
    return 1;
}

string? model = Environment.GetEnvironmentVariable("ANTHROPIC_MODEL");
if (string.IsNullOrWhiteSpace(model))
{
    Console.Error.WriteLine("ANTHROPIC_MODEL is not set; ask GET /v1/models which models exist");
    return 1;
}

string? key = Environment.GetEnvironmentVariable("ANTHROPIC_API_KEY");
if (string.IsNullOrWhiteSpace(key))
{
    Console.Error.WriteLine("ANTHROPIC_API_KEY is not set.");
    return 1;
}

string baseUrl = Environment.GetEnvironmentVariable("ANTHROPIC_BASE_URL")
    ?? "https://api.anthropic.com";
var config = new Configuration(baseUrl, key, model);
var application = new Ensemble.Ensemble(Console.Error);
try
{
    var agent = new Agent(application, config);
    while (true)
    {
        if (chat)
        {
            Console.Error.Write("You: ");
        }

        // ReadLineAsync reads the whole physical line, including an unterminated
        // last line. Only null means normal EOF; I/O exceptions are failures.
        string? line = await Console.In.ReadLineAsync();
        if (line is null)
        {
            break;
        }
        if (string.IsNullOrWhiteSpace(line))
        {
            continue;
        }

        string question;
        if (chat)
        {
            question = line;
        }
        else
        {
            using JsonDocument input = JsonDocument.Parse(line);
            question = input.RootElement.GetProperty("user").GetString()
                ?? throw new InvalidDataException("Missing user message.");
        }

        string answer = await agent.AskAsync(question);
        Usage usage = agent.Engine().Usage();
        if (chat)
        {
            Console.WriteLine(answer);
            Console.Error.WriteLine($"Usage: input={usage.Input} output={usage.Output}");
        }
        else
        {
            Console.WriteLine(JsonSerializer.Serialize(new { assistant = answer }));
        }
    }

    if (!chat)
    {
        Usage total = agent.Engine().Usage();
        Console.WriteLine(JsonSerializer.Serialize(new
        {
            usage = new { input = total.Input, output = total.Output }
        }));
    }
    return 0;
}
catch (Exception error) when (
    error is IOException or InvalidDataException or JsonException or InvalidOperationException or
    ArgumentException or KeyNotFoundException)
{
    // Do not print raw exception text: input and remote failures may contain
    // secrets. In particular, a failed read must not emit the normal EOF record.
    Console.Error.WriteLine("Conversation stopped because input or the provider exchange failed.");
    return 1;
}
finally
{
    application.Destroy();
}
