using Ensemble;
using Ensemble.Gateway;

static void Check(bool condition, string message)
{
    if (!condition) throw new InvalidOperationException(message);
}
int passed = 0, failed = 0;
async Task Test(string name, Func<Task> body)
{
    try { await body(); passed++; Console.WriteLine("PASS " + name); }
    catch (Exception error) { failed++; Console.WriteLine("FAIL " + name + ": " + error); }
}
await Test("busy actor replies retain exact request correlation and admission receipts", async () =>
{
    await using var f = new Fixture();
    await f.Start();
    await f.Send("FIRST", "first");
    var first = await f.Next();
    await f.Send("SECOND", "second");
    await f.Receipt(2);
    Check(!f.Answers.Messages.Reader.TryRead(out _), "Accepted delivery was mistaken for completion.");
    Check(first.Body.Contains("FIRST") && !first.Body.Contains("SECOND"), "Queued work changed the frozen first request.");
    first.Reply.SetResult((200, "first answer"));
    var second = await f.Next();
    Check(second.Body.Contains("SECOND"), "Second admitted turn did not run.");
    second.Reply.SetResult((200, "second answer"));
    var a = await f.Answer();
    var b = await f.Answer();
    var answers = new[] { a, b }.ToDictionary(x => x.ConversationID);
    Check(answers["first"].Content.EndsWith("first answer", StringComparison.Ordinal), "First result was misrouted.");
    Check(answers["second"].Content.EndsWith("second answer", StringComparison.Ordinal), "Second result was misrouted.");
    Check(a.Source.UserID == "subject" && b.Source.UserID == "subject", "Reply lost authenticated identity.");
    Check(a.ResponseMode == "" && b.ResponseMode == "", "Reply requested recursive replies.");
    Check(f.Agent.History().Events().Where(e => e.Kind == EventKind.MessageReceived)
        .SelectMany(e => e.Parts).OfType<TextPart>().Any(p => p.Text.Contains("[GATEWAY") && p.Text.Contains("FIRST")),
        "Router provenance was stripped from ordinary input.");
    Check(f.Agent.Tools().Length == 0, "Message content granted tools.");
    f.Connection.Cancel();
    await f.Running.WaitAsync(f.Deadline.Token);
    Check(f.QueuedReplies() == 2, "Each successful request must enqueue exactly one reply.");
});
await Test("no-response deliveries still run; failed turns do not leak exceptions", async () =>
{
    await using var f = new Fixture();
    await f.Start();
    await f.Send("NO_REPLY", "silent", "");
    (await f.Next()).Reply.SetResult((200, "silent result"));
    await f.Send("FAIL", "failed");
    (await f.Next()).Reply.SetResult((400, "private provider failure"));
    await f.Send("AFTER", "after");
    (await f.Next()).Reply.SetResult((200, "after result"));
    var answer = await f.Answer();
    Check(answer.ConversationID == "after" && answer.Content.EndsWith("after result", StringComparison.Ordinal), "A failed/silent turn replied.");
    f.Connection.Cancel();
    await f.Running.WaitAsync(f.Deadline.Token);
    Check(f.QueuedReplies() == 1, "Failed or no-response delivery enqueued an extra reply.");
    Check(f.Diagnostics.ToString().Contains("gateway turn/reply failed"), "Failure disappeared from diagnostics.");
});
await Test("disconnect joins observers but does not cancel admitted actor work", async () =>
{
    await using var f = new Fixture();
    await f.Start();
    await f.Send("CONTINUE", "detached");
    var pending = await f.Next();
    await f.Receipt(1);
    f.Connection.Cancel();
    await f.Running.WaitAsync(f.Deadline.Token);
    pending.Reply.SetResult((200, "completed after disconnect"));
    await f.Agent.FlushAsync(f.Deadline.Token);
    Check(f.Agent.History().Events().Any(e => e.Kind == EventKind.TurnEnded), "Connection cancellation killed admitted work.");
    Check(f.QueuedReplies() == 0, "Detached observer sent a reply.");
});
await Test("closed actor refuses delivery before receipt", async () =>
{
    await using var f = new Fixture();
    await f.Start();
    await f.Agent.ShutdownAsync();
    await f.Send("REFUSE", "refused");
    try { await f.Running.WaitAsync(f.Deadline.Token); throw new Exception("Closed actor was accepted."); }
    catch (IOException) { }
    // The fixture's teardown still collects the already-observed failed run.
    f.Running = Task.CompletedTask;
    f.Router.Save();
    using var json = System.Text.Json.JsonDocument.Parse(await File.ReadAllTextAsync(Path.Combine(f.DirectoryPath, "queues", "subject.json")));
    Check(json.RootElement.GetProperty("ack").GetUInt64() == 0, "Refused delivery was acknowledged.");
    Check(!f.Requests.Reader.TryRead(out _), "Closed actor performed inference.");
});
await Test("authentication failure settles readiness and joined outcome together", async () =>
{
    await using var f = new Fixture();
    Exception? readiness = null;
    try { await f.Start(wrongPin: true); }
    catch (Exception error) when (error is not OperationCanceledException and not TimeoutException)
    {
        readiness = error;
    }
    Check(readiness is not null && f.Bridge.Ready().IsFaulted, "Authentication did not fail readiness.");
    Exception? outcome = null;
    try { await f.Running.WaitAsync(f.Deadline.Token); }
    catch (Exception error) when (error is not OperationCanceledException and not TimeoutException)
    {
        outcome = error;
    }
    Check(ReferenceEquals(readiness, outcome), "Readiness and run did not preserve the same startup failure.");
    f.Running = Task.CompletedTask;
    Check(!f.Requests.Reader.TryRead(out _), "Failed authentication admitted a model request.");
});
await Test("remote shutdown fails the run and leaves admitted actor work independently owned", async () =>
{
    await using var f = new Fixture();
    await f.Start();
    await f.Send("REMOTE_CLOSE", "closed");
    var pending = await f.Next();
    await f.Receipt(1);
    await f.Router.DisposeAsync();
    try { await f.Running.WaitAsync(f.Deadline.Token); throw new Exception("Remote shutdown was reported as success."); }
    catch (IOException) { }
    f.Running = Task.CompletedTask;
    pending.Reply.SetResult((200, "still completed"));
    await f.Agent.FlushAsync(f.Deadline.Token);
    Check(f.Agent.History().Events().Any(e => e.Kind == EventKind.TurnEnded), "Remote shutdown killed actor work.");
});
Console.WriteLine($"{passed} passed, {failed} failed");
return failed == 0 ? 0 : 1;
