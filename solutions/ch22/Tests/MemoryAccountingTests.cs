using System.Text.Json;
using System.Text.Json.Nodes;
using Ensemble;
using Application = Ensemble.Ensemble;

internal static partial class MeterTests
{
    private static string AccountReply(string model, string? tier, int input, int cached, int output,
        object? summary = null)
    {
        var body = summary is null
            ? JsonNode.Parse("{\"choices\":[{\"message\":{\"content\":\"done\"},\"finish_reason\":\"stop\"}]}")!
            : JsonNode.Parse(WorkerChat(JsonSerializer.Serialize(new { memory = summary })))!;
        body["model"] = model;
        body["service_tier"] = tier;
        body["usage"] = new JsonObject
        {
            ["prompt_tokens"] = input,
            ["prompt_tokens_details"] = new JsonObject { ["cached_tokens"] = cached },
            ["completion_tokens"] = output
        };
        return body.ToJsonString();
    }

    public static async Task MemoryAccountingCommit(string directory)
    {
        using var listener = Listen();
        var app = new Application(TextWriter.Null);
        var agent = await RuntimeAgent(app, Path.Combine(directory, "account-commit"), listener.Prefixes.Single());
        try
        {
            var engine = agent.Engine();
            var initial = engine.SendAsync("{}");
            await Reply(await listener.GetContextAsync().WaitAsync(TimeSpan.FromSeconds(10)),
                AccountReply("gpt-4o-mini-2024-07-18", "default", 100, 20, 4));
            await initial;
            var before = engine.Meter();
            await RuntimeCheckpoint(agent);
            var worker = await listener.GetContextAsync().WaitAsync(TimeSpan.FromSeconds(10));
            var turn = agent.SubmitAsync([new TextPart("Continue while the worker finishes")]);
            var main = await listener.GetContextAsync().WaitAsync(TimeSpan.FromSeconds(10));
            await Reply(worker, AccountReply("gpt-4o-2024-08-06", "default", 50, 10, 7, "COMPACT"));
            await RuntimeUntil(agent, () => engine.Meter().Totals.AllInput() == 150);
            var held = engine.Meter();
            Check(ReferenceEquals(held.Latest, before.Latest) && held.Rows.Length == 2
                && held.Totals.Output == 11 && agent.Children().Length == 0,
                "Joined worker changed latest main or was not imported before pending main finished");
            Check(agent.History().View().Compression?.Committed is null,
                "Accounting prematurely committed memory");
            Check(before.Totals.AllInput() == 100 && before.Rows.Length == 1, "Import mutated an earlier snapshot");
            await Reply(main, AccountReply("gpt-4o-mini-2024-07-18", "default", 20, 5, 3));
            await turn.WaitAsync(TimeSpan.FromSeconds(10));
            await RuntimeUntil(agent, () => agent.History().View().CompressionOutcome is not null);
            var outcome = agent.History().View().CompressionOutcome!;
            Check(outcome.State == CompressionState.Committed && outcome.Usage.Length == 1
                && outcome.Usage[0].Counts.AllInput() == 50, "Missing committed usage receipt");
            var final = engine.Meter();
            Check(final.Totals.AllInput() == 170 && final.Totals.Output == 14
                && final.Latest!.Counts.AllInput() == 20 && final.EstimatedUsd() is > 0,
                "Commit rebilled child work or overwrote latest main usage");
            await agent.WorkStatusAsync();
            await agent.ShutdownAsync().WaitAsync(TimeSpan.FromSeconds(10));
            Check(engine.Meter().Totals.AllInput() == 170, "Shutdown imported completion twice");

            // Restoration is documentary, even though the outcome contains a
            // full receipt. No current-process inference has occurred here.
            string saved = Path.Combine(agent.DataDirectory(), "save.json");
            var restored = new Agent(app, new Configuration(Vendor.OpenAI, "fixture", listener.Prefixes.Single()));
            Check(restored.History().LoadSave(saved), "Receipt save not restored");
            Check(restored.History().View().CompressionOutcome!.Usage.Single().Counts.AllInput() == 50
                && restored.Engine().Meter().Totals.AllInput() == 0 && restored.Engine().Meter().Latest is null,
                "Replay billed an old receipt");
        }
        finally { app.Destroy(); }
    }

    public static async Task MemoryAccountingFailure(string directory)
    {
        using var listener = Listen();
        var app = new Application(TextWriter.Null);
        var agent = await RuntimeAgent(app, Path.Combine(directory, "account-failure"), listener.Prefixes.Single());
        try
        {
            await RuntimeCheckpoint(agent);
            // A numeric memory value fails authoritative submit validation. The real
            // child retries until its three-send bound, retaining every bill.
            for (int i = 0; i < 3; i++)
            {
                var request = await listener.GetContextAsync().WaitAsync(TimeSpan.FromSeconds(10));
                await Reply(request, AccountReply("gpt-4o-mini-2024-07-18", i == 1 ? "priority" : "default",
                    30, 10, 2, 7));
            }
            await RuntimeUntil(agent, () => agent.History().View().CompressionOutcome is not null);
            var outcome = agent.History().View().CompressionOutcome!;
            var meter = agent.Engine().Meter();
            Check(outcome.State == CompressionState.Failed && outcome.Usage.Length == 2
                && meter.Rows.Length == 2 && meter.Totals.AllInput() == 90 && meter.Totals.Output == 6
                && meter.EstimatedUsd() is null && meter.Latest is null,
                "Failed worker lost rows, invented price, or changed main usage");
            Check(outcome.Usage.Sum(row => row.Counts.AllInput()) == 90, "Failure receipt lost completed attempts");
            await agent.ShutdownAsync().WaitAsync(TimeSpan.FromSeconds(10));
            Check(agent.Engine().Meter().Totals.AllInput() == 90, "Failed work rebilled at shutdown");
        }
        finally { app.Destroy(); }
    }

    public static async Task MemoryAccountingCancellation(string directory)
    {
        foreach (string mode in new[] { "policy", "delivered", "shutdown", "journal-fault" })
        {
            using var listener = Listen();
            var app = new Application(TextWriter.Null);
            var agent = await RuntimeAgent(app, Path.Combine(directory, "account-" + mode), listener.Prefixes.Single());
            var engine = agent.Engine();
            try
            {
                await RuntimeCheckpoint(agent);
                await Reply(await listener.GetContextAsync().WaitAsync(TimeSpan.FromSeconds(10)),
                    AccountReply("gpt-4o-mini-2024-07-18", "default", 40, 10, 3, 7));
                var held = await listener.GetContextAsync().WaitAsync(TimeSpan.FromSeconds(10));
                Task<string>? mainTurn = null;
                System.Net.HttpListenerContext? mainRequest = null;
                long expected = mode == "delivered" ? 80 : 40;
                if (mode == "delivered")
                {
                    mainTurn = agent.SubmitAsync([new TextPart("Keep main work in flight")]);
                    mainRequest = await listener.GetContextAsync().WaitAsync(TimeSpan.FromSeconds(10));
                    await Reply(held, AccountReply("gpt-4o-mini-2024-07-18", "default", 40, 10, 3, "ready"));
                    await RuntimeUntil(agent, () => engine.Meter().Totals.AllInput() == 80);
                }
                if (mode == "journal-fault")
                {
                    string journal = Path.Combine(agent.DataDirectory(), "save.json.journal.jsonl");
                    File.Delete(journal);
                    Directory.CreateDirectory(journal);
                }
                bool refused = false;
                try
                {
                    if (mode is "policy" or "delivered")
                    {
                        await agent.SetMemorySettingsAsync(RuntimePolicy(sessionOff: true)).WaitAsync(TimeSpan.FromSeconds(10));
                    }
                    else
                    {
                        await agent.ShutdownAsync().WaitAsync(TimeSpan.FromSeconds(10));
                    }
                }
                catch (Exception error) when (mode == "journal-fault"
                    && error is IOException or UnauthorizedAccessException)
                {
                    refused = true;
                }
                if (mode != "delivered")
                {
                    held.Response.Abort();
                }
                Check(refused == (mode == "journal-fault") && agent.Children().Length == 0,
                    "Cancellation failed to join or swallowed journal failure");
                Check(engine.Meter().Totals.AllInput() == expected && engine.Meter().Totals.Output == (expected == 80 ? 6 : 3)
                    && engine.Meter().Latest is null, "Canceled/faulted attempt lost or doubled measured usage");
                if (mode != "journal-fault")
                {
                    var outcome = agent.History().View().CompressionOutcome!;
                    Check(outcome.Usage.Single().Counts.AllInput() == expected
                        && outcome.State == (mode is "policy" or "delivered" ? CompressionState.Invalidated : CompressionState.Canceled),
                        "Cancellation receipt lost earlier completed response");
                }
                if (mainRequest is not null)
                {
                    await Reply(mainRequest, AccountReply("gpt-4o-mini-2024-07-18", "default", 0, 0, 0));
                    await mainTurn!.WaitAsync(TimeSpan.FromSeconds(10));
                }
            }
            finally
            {
                try { app.Destroy(); }
                catch (Exception error) when (mode == "journal-fault"
                    && error is IOException or UnauthorizedAccessException or AggregateException) { }
            }
            Check(engine.Meter().Totals.AllInput() == (mode == "delivered" ? 80 : 40), "Final teardown billed again");
        }
    }

    public static void MemoryAccountingCodec()
    {
        var identity = new UsageIdentity(Vendor.OpenAI, "actual-model", null, BillingMode.ChatGptPlan);
        var rows = new[] { new MeterEntry(identity, new Usage(null, 3, 4, 5)) };
        var outcome = new CompressionOutcome("receipt", CompressionState.Failed, "bounded failure", rows);
        rows[0] = new MeterEntry(identity, new Usage(999));
        string encoded = JsonCodec.WriteEvent(new Event(EventKind.CompressionFinished, seq: 1, outcome: outcome));
        var decoded = JsonCodec.ReadEvent(encoded).Outcome!;
        Check(decoded.Usage.Single().Identity == identity && decoded.Usage[0].Counts.Input is null
            && decoded.Usage[0].Counts.CacheWrite == 3 && decoded.Usage[0].Counts.Output == 5,
            "Receipt lost identity, null counts, or value ownership");
        var old = JsonNode.Parse(encoded)!;
        old["outcome"]!.AsObject().Remove("usage");
        Check(JsonCodec.ReadEvent(old.ToJsonString()).Outcome!.Usage.IsEmpty, "Old outcome compatibility lost");
        var app = new Application(TextWriter.Null);
        try
        {
            var agent = new Agent(app, new Configuration(Vendor.OpenAI, "fixture", "http://127.0.0.1:1"));
            agent.Engine().ImportUsage(new MeterSnapshot(decoded.Usage[0], decoded.Usage));
            var imported = agent.Engine().Meter();
            Check(imported.Latest is null && imported.Rows.Single().Identity == identity
                && imported.Totals.Input is null && imported.Totals.CacheWrite == 3
                && imported.Totals.CacheRead == 4 && imported.EstimatedUsd() is null,
                "Import fabricated latest response, unknown counts, or plan pricing");
        }
        finally
        {
            app.Destroy();
        }
        foreach (string defect in new[] { "null", "duplicate", "missing-count", "negative" })
        {
            var damaged = JsonNode.Parse(encoded)!;
            var receipt = damaged["outcome"]!;
            if (defect == "null") receipt["usage"] = null;
            if (defect == "duplicate") receipt["usage"]!.AsArray().Add(receipt["usage"]![0]!.DeepClone());
            if (defect == "missing-count") receipt["usage"]![0]!["counts"]!.AsObject().Remove("input");
            if (defect == "negative") receipt["usage"]![0]!["counts"]!["input"] = -1;
            bool refused = false;
            try { _ = JsonCodec.ReadEvent(damaged.ToJsonString()); }
            catch (Exception error) when (error is ArgumentException or InvalidOperationException or KeyNotFoundException
                or InvalidDataException or JsonException or FormatException) { refused = true; }
            Check(refused, "Damaged receipt accepted: " + defect);
        }
    }
}
