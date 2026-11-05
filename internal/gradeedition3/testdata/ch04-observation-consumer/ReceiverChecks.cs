using System.Text.Json;
using Ensemble;

internal static class ReceiverChecks
{
    internal static async Task<int> Run()
    {
        var checks = new List<string>();
        void Check(bool condition, string name)
        {
            if (!condition) throw new InvalidOperationException(name);
            checks.Add(name);
        }
        try
        {
            using var deadline = new CancellationTokenSource(TimeSpan.FromSeconds(20));
            var root = new Ensemble.Ensemble(TextWriter.Null);
            try
            {
                var config = new Configuration(Vendor.Anthropic, "local", "http://127.0.0.1:1", "synthetic-unused");
                var owner = new Agent(root, config);
                int first = 0, second = 0;
                string nonce = Guid.NewGuid().ToString("N");
                var schema = JsonSerializer.SerializeToElement(new { type = "object", properties = new { } });
                var receiver = new Tool(owner, "receiver_first", "First harmless callback", schema,
                    (_, _, _) => { Interlocked.Increment(ref first); return Task.FromResult("first-" + nonce); });
                _ = new Tool(owner, "receiver_second", "Second harmless callback", schema,
                    (_, _, _) => { Interlocked.Increment(ref second); return Task.FromResult("second-" + nonce); });
                ToolCallPart Call(string id, string name) => new(id, config.Target(), name,
                    JsonSerializer.SerializeToElement(new { ai_callback_delay = 3 }));

                int facts = owner.History().Events().Length;
                var wrong = await receiver.ExecuteAsync(Call("wrong-" + nonce, "receiver_second"), deadline.Token);
                Check(wrong.IsError, "receiver-mismatch-is-error");
                Check(wrong.CallID == "wrong-" + nonce, "receiver-mismatch-retains-call-id");
                Check(wrong.Job is null, "receiver-mismatch-no-job-metadata");
                Check(first == 0 && second == 0, "receiver-mismatch-no-handler-effects");
                Check(owner.Jobs().Children().Length == 0, "receiver-mismatch-no-job-admission");
                Check(owner.History().Events().Length == facts, "receiver-mismatch-no-dispatch-facts");

                var own = await receiver.ExecuteAsync(Call("own-" + nonce, receiver.Name()), deadline.Token);
                Check(!own.IsError && JsonCodec.Text(own.Parts).Contains("first-" + nonce, StringComparison.Ordinal),
                    "receiver-own-call-real-output");
                Check(first == 1 && second == 0, "receiver-own-call-executes-only-receiver");
                var selected = await owner.DispatchAsync(Call("selected-" + nonce, "receiver_second"), deadline.Token);
                Check(!selected.IsError && JsonCodec.Text(selected.Parts).Contains("second-" + nonce, StringComparison.Ordinal),
                    "agent-dispatch-selects-by-name");
                Check(first == 1 && second == 1, "agent-dispatch-executes-selected-once");
            }
            finally { root.Destroy(); }
            Console.WriteLine(JsonSerializer.Serialize(new { passed = true, checks }));
            return 0;
        }
        catch (Exception error)
        {
            Console.WriteLine(JsonSerializer.Serialize(new { passed = false, checks, failure = error.Message }));
            return 1;
        }
    }
}
