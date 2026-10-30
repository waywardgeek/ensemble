using System.Text.Json;
using Ensemble;

internal static class MemorySelectionTests
{
    private static void Check(bool condition, string message)
    {
        if (!condition)
        {
            throw new Exception(message);
        }
    }

    private static Entry Said(long seq, string text) => new(seq, Actor.Human, [new TextPart(text)]);
    private static Entry Note(long seq, string text) => new(seq, Actor.System, [new HandoffPart(text)]);

    public static void Prefix()
    {
        var entries = new[]
        {
            new Entry(1, Actor.System, [new SkillPart("manual", "not dialogue", ["write_file"])]),
            Said(2, "ééé"), // Six UTF-8 bytes, not three characters.
            Note(3, "retain this reason even though notes are not budgeted"),
            Said(4, "second"),
            Note(5, "second reason"),
            Said(6, "open")
        };
        var context = new Context(entries);
        Check(History.SelectConversation(context, 16) is null, "At-budget conversation selected");
        var first = History.SelectConversation(context, 10)
            ?? throw new Exception("Over-budget completed source not selected");
        Check(first.Entries.Select(entry => entry.Seq).SequenceEqual([2L, 3L]), "Selection split/overshot first segment");
        Check(first.Text() == "Human: ééé\nSystem: retain this reason even though notes are not budgeted\n",
            "Compressor input lost readable source, attribution or note");
        var both = History.SelectConversation(context, 5)
            ?? throw new Exception("Multiple segments not selected");
        Check(both.Entries.Select(entry => entry.Seq).SequenceEqual([2L, 3L, 4L, 5L]), "Selection did not reach budget in whole segments");
        var allCompleted = History.SelectConversation(context, 1)!;
        Check(allCompleted.Entries.Length == 4, "Oversized unfinished tail entered selection");
        Check(context.Dialogue.SequenceEqual(entries) && context.EnabledTools().Contains("write_file"),
            "Selection changed source or authority");
        entries[1] = Said(2, "caller changed its array");
        Check(first.Entries[0].Parts.OfType<TextPart>().Single().Text == "ééé", "Selection retained a mutable collection");
        var later = new Context(context.Dialogue.Append(Said(7, "later append")));
        Check(first.Entries.All(entry => later.Dialogue.Contains(entry)), "Unrelated append invalidated exact selected source");

        // One completed checkpoint is sufficient when over budget. The two-file
        // minimum belongs to graduation, not conversation selection.
        Check(History.SelectConversation(new Context([Said(1, "first task"), Note(2, "done")]), 1)
            ?.Entries.Length == 2, "Invented a two-checkpoint prerequisite");
    }

    public static void Boundaries()
    {
        Check(History.SelectConversation(new Context([Said(1, "unfinished")]), 1) is null,
            "Uncheckpointed conversation selected");
        Check(History.SelectConversation(new Context([Note(1, "note only"), Said(2, "unfinished")]), 1) is null,
            "A note-only prefix manufactured compression progress");
        var call = new ToolCallPart("call", new Configuration(Vendor.OpenAI, "model-a", "http://127.0.0.1:1").Target(), "work", JsonSerializer.SerializeToElement(new { }));
        var unsafePrefix = new Context([
            Said(1, "old"), new Entry(2, Actor.Agent, [call]), Note(3, "not yet safe")]);
        Check(History.SelectConversation(unsafePrefix, 1) is null, "Unpaired tool source selected");
        var safeThenPending = new Context([
            Said(1, "old"), Note(2, "complete"), new Entry(3, Actor.Agent, [call]), Said(4, "unfinished")]);
        Check(History.SelectConversation(safeThenPending, 1)?.Entries.Length == 2,
            "Pending work prevented selection of an earlier complete prefix");
        var signed = new TextPart("private envelope", from: new Configuration(Vendor.OpenAI, "model-a", "http://127.0.0.1:1").Target(), opaque: JsonSerializer.SerializeToElement(new { signature = "private" }));
        Check(History.SelectConversation(new Context([new Entry(1, Actor.Agent, [signed]), Note(2, "unsafe")]), 1) is null,
            "Private replay selected before handoff stripped it");
        foreach (long invalid in new[] { 0L, -1L })
        {
            try
            {
                _ = History.SelectConversation(safeThenPending, invalid);
                throw new Exception("Unresolved/negative budget accepted");
            }
            catch (ArgumentOutOfRangeException) { }
        }
    }
}
