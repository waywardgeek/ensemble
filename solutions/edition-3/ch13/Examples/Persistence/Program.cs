using Ensemble;

// This independent assembly uses public owners and value types only. The caller
// supplies an empty directory; the example never touches an existing session.
if (args.Length != 1 || Directory.Exists(args[0]))
{
    Console.Error.WriteLine("Supply one new output directory.");
    return 2;
}
Directory.CreateDirectory(args[0]);
string path = Path.Combine(args[0], "save.json");
var policy = new PersistenceOptions(path, RetainLog: 0);

var firstRoot = new Ensemble.Ensemble(Console.Error);
try
{
    var first = new Agent(firstRoot, new Configuration(Vendor.Anthropic, "before", "http://127.0.0.1:1", ""));
    if (first.StartPersistence(policy))
    {
        throw new InvalidOperationException("An empty directory unexpectedly restored a save");
    }
    // Explicit history import is a public startup operation, before actor work.
    first.History().Append(new Event(EventKind.MessageReceived, actor: Actor.Human,
        parts: [new TextPart("Remember the violet notebook.")]));
    first.History().Append(new Event(EventKind.ResponseEnded,
        parts: [new TextPart("The violet notebook is recorded.")]));
    await first.ShutdownAsync();
}
finally
{
    firstRoot.Destroy();
}

var saved = JsonCodec.ReadSave(File.ReadAllText(path));
if (!saved.Events.IsEmpty || saved.Context is null)
{
    throw new InvalidOperationException("Snapshot-only policy was not honored");
}
var secondRoot = new Ensemble.Ensemble(Console.Error);
try
{
    // Saved configuration is documentary, never a replacement for these inputs.
    var second = new Agent(secondRoot, new Configuration(Vendor.OpenAI, "after", "http://127.0.0.1:1", ""));
    if (!second.StartPersistence(policy))
    {
        throw new InvalidOperationException("The save was not restored");
    }
    string request = second.Engine().Render(second.History().View()).ToJsonString();
    if (!request.Contains("violet notebook") || !request.Contains("after"))
    {
        throw new InvalidOperationException("Conversation or current configuration was lost");
    }
    var fact = second.History().Append(new Event(EventKind.HintReceived, actor: Actor.Human,
        parts: [new TextPart("Continue from the saved notebook.")]));
    if (fact.Seq <= saved.Sequence)
    {
        throw new InvalidOperationException("Event numbering restarted");
    }
    await second.ShutdownAsync();
    Console.WriteLine("PASS independent C# consumer: snapshot-only load, current config, continuous sequence, automatic save.");
}
finally
{
    secondRoot.Destroy();
}
return 0;
