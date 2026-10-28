using System.Text.Json;
using Ensemble;
using Application = Ensemble.Ensemble;

// Real ordinary-tool dispatch over owned Agents, Jobs and durable files. The
// separate loopback case below the actor fixtures checks model-led submission.
internal static class ChildToolTests
{
    private static void Check(bool value, string why)
    {
        if (!value) throw new Exception(why);
    }

    private static Agent Parent(Application app, string temp, string name)
    {
        string work = Path.Combine(temp, name);
        Directory.CreateDirectory(work);
        var parent = new Agent(app, new Configuration(Vendor.Anthropic, "fixture",
            "http://127.0.0.1:1", "local-only", sandboxRoot: work));
        parent.InstallTools();
        parent.StartPersistence(new PersistenceOptions(Path.Combine(work, "save.json")));
        return parent;
    }

    private static JsonElement J(string text) => JsonSerializer.Deserialize<JsonElement>(text);
    private static Task<ToolResultPart> Call(Agent owner, string tool, object args, CancellationToken token)
        => owner.DispatchAsync(new ToolCallPart(Guid.NewGuid().ToString("N"),
            owner.Configuration().Target(), tool, JsonSerializer.SerializeToElement(args)), token);

    public static async Task Submission(string temp)
    {
        var app = new Application(TextWriter.Null);
        using var deadline = new CancellationTokenSource(TimeSpan.FromSeconds(15));
        try
        {
            Agent parent = Parent(app, temp, "tool-submission");
            JsonElement schema = J("""
                {"type":"object","properties":{"count":{"type":"integer","minimum":9007199254740993},"data":{"type":"object","properties":{"ok":{"const":true}},"required":["ok"],"additionalProperties":false}},"required":["count","data"],"additionalProperties":false}
                """);
            Agent child = await parent.CreateChildAsync("typed", submitSchema: schema, cancellation: deadline.Token);
            var declared = child.Tools().Single(tool => tool.Name() == "submit").Declaration(Vendor.Anthropic);
            Check(declared["input_schema"]?["properties"]?["count"]?["minimum"]?.ToJsonString() == "9007199254740993",
                "Submit declaration wrapped or replaced the child's exact schema.");
            foreach (string invalid in new[] {
                """{"count":9007199254740992,"data":{"ok":true}}""",
                """{"count":9007199254740993,"data":{"ok":false}}""",
                """{"count":9007199254740993,"data":{"ok":true,"extra":1}}""" })
            {
                ToolResultPart refused = await Call(child, "submit", J(invalid), deadline.Token);
                Check(refused.IsError && child.Submission() is null
                    && !File.Exists(Path.Combine(child.DataDirectory(), "submit.md")),
                    "Invalid submission published completion or a file.");
            }
            int beforeDuplicate = child.History().Events().Length;
            bool duplicateRefused = false;
            try
            {
                await Call(child, "submit", J("""{"count":9007199254740993,"count":9007199254740994,"data":{"ok":true}}"""), deadline.Token);
            }
            catch (ArgumentException) { duplicateRefused = true; }
            Check(duplicateRefused && !child.History().Events().Skip(beforeDuplicate)
                    .Any(fact => fact.Kind == EventKind.ToolStarted)
                && child.Submission() is null && !File.Exists(Path.Combine(child.DataDirectory(), "submit.md")),
                "Malformed duplicate input started work or published submission.");
            var descendant = await child.CreateChildAsync("unfinished", cancellation: deadline.Token);
            JsonElement valid = J("""{"count":9007199254740993,"data":{"ok":true},"ai_callback_delay":120}""");
            Check((await Call(child, "submit", valid, deadline.Token)).IsError,
                "Parent submitted while an unresolved child remained.");
            descendant.Destroy();
            ToolResultPart report = await Call(child, "submit", valid, deadline.Token);
            if (report.Job?.Status == "running")
            {
                report = await child.Jobs().Find(report.Job.Handle).CompletionAsync(deadline.Token);
            }
            Check(!report.IsError && child.Submission()!.Value.GetProperty("count").GetRawText() == "9007199254740993",
                "Exact typed submission did not survive dispatch.");
            JsonElement row = J(File.ReadAllText(Path.Combine(child.DataDirectory(), "submit.md")));
            Check(row.GetProperty("payload").GetProperty("data").GetProperty("ok").GetBoolean()
                && !row.GetProperty("payload").TryGetProperty("ai_callback_delay", out _),
                "Durable payload lost nesting or retained observer options.");
            Check((await Call(parent, "submit", new { result = "not delegated" }, deadline.Token)).IsError,
                "Standalone Agent accepted child submission.");
            Agent legacy = await parent.CreateChildAsync("legacy", cancellation: deadline.Token);
            Check(!(await Call(legacy, "submit", new { result = "text", ignored = 1 }, deadline.Token)).IsError,
                "Legacy submission failed.");
            Check(legacy.Submission()!.Value.EnumerateObject().Count() == 1,
                "Legacy submission was not normalized.");
            Agent booleans = await parent.CreateChildAsync("boolean-schema", submitSchema: J("""
                {"type":"object","properties":{"allowed":true,"blocked":false},"additionalProperties":false}
                """), cancellation: deadline.Token);
            var submit = booleans.Tools().Single(tool => tool.Name() == "submit");
            string original = submit.Declaration(Vendor.Anthropic).ToJsonString();
            _ = submit.Declaration(Vendor.Gemini);
            Check(submit.Declaration(Vendor.Anthropic).ToJsonString() == original,
                "Provider projection mutated a boolean submission schema.");
            Check((await Call(booleans, "submit", new { blocked = true }, deadline.Token)).IsError,
                "Boolean false schema admitted a prohibited value.");
            Check(!(await Call(booleans, "submit", new { allowed = new[] { 1, 2 } }, deadline.Token)).IsError,
                "Boolean true schema rejected valid arbitrary JSON.");
        }
        finally { app.Destroy(); }
    }

    public static async Task Questions(string temp)
    {
        var app = new Application(TextWriter.Null);
        using var deadline = new CancellationTokenSource(TimeSpan.FromSeconds(15));
        try
        {
            Agent parent = Parent(app, temp, "tool-questions");
            Agent child = await parent.CreateChildAsync("worker", cancellation: deadline.Token);
            var announced = parent.WaitForQuestionsAsync(cancellation: deadline.Token);
            ToolResultPart asking = await Call(child, "send_message_to_parent",
                new { content = "Which file?", ai_callback_delay = 0 }, deadline.Token);
            QuestionInfo question = (await announced).Single();
            Check(asking.Job?.Status == "running" && question.Content == "Which file?",
                "Question did not remain an owned running Job.");
            Check((await Call(parent, "respond_to_parent_message", new {
                agent_id = "missing", response = "wrong", message_id = question.MessageID }, deadline.Token)).IsError,
                "Reply accepted an unowned child.");
            Task<QuestionReply> second = child.SendMessageToParentAsync("Which format?", deadline.Token);
            Check((await Call(parent, "respond_to_parent_message", new {
                agent_id = "worker", response = "ambiguous" }, deadline.Token)).IsError,
                "Reply silently selected among multiple pending questions.");
            Check(!(await Call(parent, "respond_to_parent_message", new {
                agent_id = "worker", response = "note.txt", message_id = question.MessageID }, deadline.Token)).IsError,
                "Exact reply failed.");
            ToolResultPart answer = await child.Jobs().Find(asking.Job!.Handle).CompletionAsync(deadline.Token);
            Check(!answer.IsError && JsonCodec.Text(answer.Parts).Contains("note.txt")
                && child.Submission() is null, "Answer became completion or lost content.");
            Check(JsonCodec.Text(answer.Parts).Contains("\"sender\":\"root\""),
                "Question reply lost its actual sender address.");
            Check(!(await Call(parent, "respond_to_parent_message", new {
                agent_id = "worker", response = "text" }, deadline.Token)).IsError,
                "Unique pending question required an unnecessary ID.");
            Check((await second).Response == "text" && parent.PendingQuestions().IsEmpty, "Reply left an inbox edge.");

            var cancellationQuestion = parent.WaitForQuestionsAsync(cancellation: deadline.Token);
            ToolResultPart cancelled = await Call(child, "send_message_to_parent",
                new { content = "Keep the question", ai_callback_delay = 0 }, deadline.Token);
            await cancellationQuestion;
            await child.Jobs().Find(cancelled.Job!.Handle).KillAsync();
            Check(parent.PendingQuestions().Length == 1, "Killing the observing Job destroyed its question.");
            parent.RespondToQuestion("still replyable", "worker");
            ToolResultPart human = await Call(parent, "send_message_to_parent",
                new { content = "Human decision", ai_callback_delay = 0 }, deadline.Token);
            QuestionInfo inbox = (await parent.WaitForQuestionsAsync(humanOnly: true, cancellation: deadline.Token)).Single();
            Check(inbox.AgentID is null && inbox.Content == "Human decision", "Root question was routed as child work.");
            parent.RespondToQuestion("human answer", messageID: inbox.MessageID);
            ToolResultPart humanAnswer = await parent.Jobs().Find(human.Job!.Handle).CompletionAsync(deadline.Token);
            Check(JsonCodec.Text(humanAnswer.Parts).Contains("human answer"), "Root answer did not reach Job.");
            // Closing the inbox precedes Job teardown. Its cancelled reply is
            // an expected tool outcome, not an unhandled Job task failure.
            ToolResultPart pending = await Call(parent, "send_message_to_parent",
                new { content = "Pending at shutdown", ai_callback_delay = 0 }, deadline.Token);
            await parent.WaitForQuestionsAsync(humanOnly: true, cancellation: deadline.Token);
            Check(pending.Job?.Status == "running", "Shutdown regression did not retain a pending question Job.");
            await parent.ShutdownAsync().WaitAsync(deadline.Token);
            parent.Destroy();
        }
        finally { app.Destroy(); }
    }

    public static async Task Shutdown(string temp)
    {
        var app = new Application(TextWriter.Null);
        using var deadline = new CancellationTokenSource(TimeSpan.FromSeconds(15));
        try
        {
            Agent parent = Parent(app, temp, "tool-shutdown");
            Agent child = await parent.CreateChildAsync("selected", cancellation: deadline.Token);
            Agent sibling = await parent.CreateChildAsync("retained", cancellation: deadline.Token);
            Agent leaf = await child.CreateChildAsync("leaf", cancellation: deadline.Token);
            string directory = child.DataDirectory();
            leaf.PublishSubmission("leaf done");
            child.PublishSubmission("selected done");
            ToolResultPart result = await Call(parent, "shutdown_agent", new { agent_id = "selected" }, deadline.Token);
            Check(!result.IsError && parent.FindChild("selected") is null && child.AgentParent() is null
                && leaf.AgentParent() is null && parent.FindChild("retained") == sibling,
                "Shutdown failed to detach exactly the selected subtree.");
            Check(File.Exists(Path.Combine(directory, "submit.md")) && File.Exists(Path.Combine(directory, "save.json")),
                "Shutdown deleted durable child work.");
            Check((await Call(parent, "shutdown_agent", new { agent_id = "selected" }, deadline.Token)).IsError,
                "Missing child shutdown claimed success.");
            Agent replacement = await parent.CreateChildAsync("selected", cancellation: deadline.Token);
            Check(replacement.Submission() is null, "Recreated child accepted old completion.");
            // A real journal write failure must be reported, while teardown still
            // detaches both owners and leaves unrelated children usable.
            string journal = Path.Combine(replacement.DataDirectory(), "save.json.journal.jsonl");
            File.Delete(journal);
            Directory.CreateDirectory(journal);
            bool poisoned = false;
            try
            {
                replacement.History().Append(new Event(EventKind.MessageReceived,
                    actor: Actor.Human, parts: [new TextPart("cannot persist")]));
            }
            catch (Exception error) when (error is IOException or UnauthorizedAccessException) { poisoned = true; }
            Check(poisoned, "Failure fixture did not poison its own journal.");
            ToolResultPart failure = await Call(parent, "shutdown_agent", new { agent_id = "selected" }, deadline.Token);
            Check(failure.IsError && parent.FindChild("selected") is null && replacement.AgentParent() is null
                && parent.FindChild("retained") == sibling, "Failed cleanup lost its error or retained ownership.");
        }
        finally { app.Destroy(); }
    }
}
