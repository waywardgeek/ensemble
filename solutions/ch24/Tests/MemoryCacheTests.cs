using System.Text.Json.Nodes;
using Ensemble;
using Application = Ensemble.Ensemble;

internal static partial class MeterTests
{
    public static void MemoryCacheBoundary()
    {
        var app = new Application(TextWriter.Null);
        try
        {
            var agent = new Agent(app, new Configuration(Vendor.Anthropic, "claude-3-haiku-20240307",
                "http://127.0.0.1:1", cacheBreakpoints: true));
            var stamp = DateTimeOffset.Parse("2026-10-01T00:00:00Z");
            MemoryEntry[] memories = [new(MemoryBand.Soul, "SOUL", stamp, "STABLE_SOUL"),
                new(MemoryBand.Session, "1", stamp, "RECENT_MEMORY")];
            Entry[] dialogue = [new(1, Actor.Human, [new TextPart("DIALOGUE_TAIL")])];
            var context = new Context(dialogue: dialogue, memories: memories, ephemeral: [new TextPart("TRANSIENT")]);
            string[] Marked(Context view)
            {
                var request = agent.Engine().Render(view);
                return request["messages"]!.AsArray().SelectMany(message => message!["content"]!.AsArray())
                    .OfType<JsonObject>().Where(block => block.ContainsKey("cache_control"))
                    .Select(block => block["text"]?.GetValue<string>() ?? "").ToArray();
            }
            string[] markers = Marked(context);
            Check(markers.Length == 2 && markers.Any(text => text.Contains("RECENT_MEMORY"))
                && markers.Contains("DIALOGUE_TAIL") && !markers.Any(text => text.Contains("TRANSIENT")),
                "Missing final memory fallback or transient block received marker");
            var withHandoffs = new Context(dialogue:
                [new(1, Actor.System, [new HandoffPart("OLDER_CHECKPOINT")]),
                 new(2, Actor.System, [new HandoffPart("NEWEST_CHECKPOINT")]),
                 new(3, Actor.Human, [new TextPart("DIALOGUE_TAIL")])], memories: memories);
            markers = Marked(withHandoffs);
            Check(markers.SequenceEqual(["NEWEST_CHECKPOINT", "DIALOGUE_TAIL"]), "Handoff did not supersede memory boundary");
            var retired = new Context(dialogue: dialogue, memories: memories.Take(1));
            markers = Marked(retired);
            Check(markers.Any(text => text.Contains("STABLE_SOUL")) && !markers.Any(text => text.Contains("RECENT_MEMORY")),
                "Retired memory retained a breakpoint");
            Check(Marked(new Context(dialogue: dialogue)).SequenceEqual(["DIALOGUE_TAIL"]), "Empty memory invented boundary");
            Check(Marked(new Context(memories: memories)).Length == 1, "Coincident memory/tail duplicated marker");
            Check(context.Memories.Length == 2 && context.Dialogue[0].Parts[0] is TextPart { Text: "DIALOGUE_TAIL" },
                "Projection changed canonical context");
        }
        finally { app.Destroy(); }
    }
}
