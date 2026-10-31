using System.Diagnostics;
using System.Text.Json;
using Ensemble;
using Application = Ensemble.Ensemble;

// Real Jobs, real files and real PTYs. A latch replaces only an external slow
// operation; it cannot manufacture a job status, cursor or cancellation result.
internal static class JobTests
{
    private static void Check(bool condition, string message)
    {
        if (!condition)
        {
            throw new Exception(message);
        }
    }

    internal static async Task Run(string directory)
    {
        var diagnostics = new CompletionWriter();
        var app = new Application(diagnostics);
        var agent = new Agent(app, new Configuration(Vendor.Anthropic, "local", "http://127.0.0.1"));
        agent.InstallTools();
        int callNumber = 0;
        Task<ToolResultPart> Call(string name, object args, CancellationToken token = default)
        {
            var call = new ToolCallPart("jobs-" + ++callNumber, agent.Configuration().Target(), name, JsonSerializer.SerializeToElement(args));
            return agent.DispatchAsync(call, token);
        }
        static string Text(ToolResultPart result) => JsonCodec.Text(result.Parts);
        try
        {
            var body = "{\"type\":\"object\",\"properties\":{},\"required\":[],\"additionalProperties\":false}";
            var released = new TaskCompletionSource(TaskCreationOptions.RunContinuationsAsynchronously);
            _ = new Tool(agent, "slow", "Held external operation", JsonSerializer.Deserialize<JsonElement>(body), async (job, args, cancellation) =>
            {
                Check(!args.TryGetProperty("ai_callback_delay", out _), "metadata stripped for custom tool");
                Check(agent.History().Events().Any(e => e.Kind == EventKind.ToolCalled && e.Job?.Handle == job.Handle), "record exists before effects");
                job.AppendOutput("first\n");
                await released.Task.WaitAsync(cancellation);
                return "last\n";
            });
            var initial = await Call("slow", new { ai_callback_delay = 0.05 });
            Check(initial.Job is { Handle: 1, Status: "running" }, "first job survives wake delay");
            Check(Text(initial).Contains("first"), "initial output visible");
            var pending = Call("wait_for_job", new { handle = 1, ai_callback_delay = 0.08 });
            var wait = Stopwatch.StartNew();
            var empty = await pending;
            Check(wait.ElapsedMilliseconds >= 40 && !Text(empty).Contains("first"), "no old output on next wait");
            released.SetResult();
            var final = await Call("wait_for_job", new { handle = 1, ai_callback_pattern = "last", ai_callback_delay = 2 });
            Check(Text(final).Contains("last"), "fresh pattern result");
            // A pattern can precede finalization. Observe the same job again.
            while (final.Job!.Status == "running")
            {
                final = await Call("wait_for_job", new { handle = 1, ai_callback_delay = 0.05 });
            }
            Check(File.ReadAllText(final.Job.OutputPath) == "first\nlast\n", "disk contains complete custom output");
            using (var exclusive = new FileStream(final.Job.OutputPath, FileMode.Open, FileAccess.Read, FileShare.None))
            {
                Check(exclusive.Length == final.Job.Bytes, "completed job released its writer");
            }
            var value = new ToolResultPart("redacted-job", [new TextPart("visible")], job: final.Job);
            var view = new Context([new Entry(1, Actor.Tool, [value])]);
            view = History.Reduce(view, new Event(EventKind.Redacted, redactFrom: 1,
                redactTo: 1, level: Redaction.Result, reason: "reduce"));
            var redacted = view.Dialogue.Single().Parts.OfType<ToolResultPart>().Single();
            Check(redacted.Job?.Handle == final.Job.Handle
                && redacted.Parts.OfType<RedactedPart>().Single().Ref?.Locator == final.Job.OutputPath,
                "redaction retains the real output locator");
            Check((await Call("wait_for_job", new { handle = 999 })).IsError, "unknown handle is a result");
            var invalid = await Call("read_file", new { path = 32 });
            Check(invalid.IsError && invalid.Job?.Handle == 2 && invalid.Job.Status == "done", "bad args still get handle");
            var unknown = await Call("unknown", new { });
            Check(unknown.IsError && unknown.Job?.Handle == 3, "unknown ordinary tool gets next handle");
            var invalidRegex = await Call("read_file", new { path = "missing", ai_callback_pattern = "[" });
            Check(invalidRegex.IsError && invalidRegex.Job?.Handle == 4, "invalid metadata is localized job failure");

            var pty = await Call("run_command", new
            {
                command = "test -t 0 && test -t 1 && echo PTY; stty size; printf 'PROMPT>'; read answer; printf 'answer:%s\\n' \"$answer\"; sleep 0.1",
                cwd = directory, ai_callback_pattern = "PROMPT>", ai_callback_delay = 2
            });
            Check(pty.Job?.Status == "running" && Text(pty).Contains("PTY") && Text(pty).Contains("50 200"), "real PTY with dimensions");
            var input = await Call("send_input", new { handle = pty.Job!.Handle, input = "hello", ai_callback_pattern = "answer:hello", ai_callback_delay = 2 });
            Check(Text(input).Contains("answer:hello") && !Text(input).Contains("PROMPT>"), "input pattern starts at new bytes");

            var many = await Call("run_command", new { command = "printf HEAD; i=0; while [ $i -lt 10000 ]; do printf x; i=$((i+1)); done; printf TAIL", max_output_bytes = 128 });
            Check(many.Job?.Status == "done" && many.Job.Truncated && Text(many).Contains("HEAD") && Text(many).Contains("TAIL"), "head/tail inline cap");
            Check(new FileInfo(many.Job!.OutputPath).Length > 10000, "inline cap never caps job output");

            await Call("tool_limits", new { ai_callback_delay = 0.05 });
            var bad = await Call("unknown", new { });
            Check(bad.Job is not null && Text(bad).Contains("tool_limits consumed"), "bad call consumes pending observation policy");
            var next = await Call("run_command", new { command = "sleep 0.15; echo survives", ai_callback_delay = 1 });
            Check(next.Job?.Status == "done" && Text(next).Contains("survives") && !Text(next).Contains("tool_limits consumed"), "policy consumed exactly once");
            await Call("tool_limits", new { ai_callback_delay = 0.02 });
            var early = await Call("run_command", new { command = "sleep 0.3; echo finishes" });
            Check(early.Job?.Status == "running" && Text(early).Contains("tool_limits consumed"), "one-shot wakes without cancellation");
            await Call("tool_limits", new { ai_callback_delay = 0.01 });
            var controlled = await Call("wait_for_job", new { handle = early.Job!.Handle });
            Check(Text(controlled).Contains("tool_limits consumed"), "control consumes one-shot policy");
            await Call("tool_limits", new { max_output_bytes = 100 });
            var overridden = await Call("tool_limits", new { ai_callback_delay = 0 });
            Check(Text(overridden).Contains("tool_limits consumed"), "tool_limits itself consumes policy");
            var invalidControl = await Call("wait_for_job", new { handle = -5 });
            Check(invalidControl.IsError && Text(invalidControl).Contains("tool_limits consumed"), "failed control still consumes");

            var refuseCancellation = new TaskCompletionSource(TaskCreationOptions.RunContinuationsAsynchronously);
            var bodyEntered = new TaskCompletionSource(TaskCreationOptions.RunContinuationsAsynchronously);
            _ = new Tool(agent, "uncooperative", "Controlled uncooperative boundary", JsonSerializer.Deserialize<JsonElement>(body), async (job, args, cancellation) =>
            {
                bodyEntered.SetResult();
                await refuseCancellation.Task;
                return "must not resurrect";
            });
            var held = await Call("uncooperative", new { ai_callback_delay = 0.01 });
            await bodyEntered.Task;
            var stopped = await Call("kill_job", new { handle = held.Job!.Handle });
            Check(stopped.Job?.Status == "killed" && !stopped.IsError && Text(stopped).Contains("cannot forcibly stop"), "managed kill is honest and returns promptly");
            diagnostics.Match = $"job {held.Job.Handle} killed";
            refuseCancellation.SetResult();
            await diagnostics.Completion.Task.WaitAsync(TimeSpan.FromSeconds(3));
            Check(agent.Jobs().Find(held.Job.Handle).Snapshot().Status == "killed" && !File.ReadAllText(held.Job.OutputPath).Contains("resurrect"), "late completion cannot change killed state");

            var pidPath = Path.Combine(directory, "child.pid");
            var running = await Call("run_command", new { command = "sleep 30 & echo $! > '" + pidPath + "'; wait", ai_callback_delay = 0.1 });
            int pid = int.Parse(File.ReadAllText(pidPath));
            var killed = await Call("kill_job", new { handle = running.Job!.Handle });
            Check(killed.Job?.Status == "killed" && !killed.IsError, "explicit kill joins job");
            Check(agent.History().Events().Any(e => e.Kind == EventKind.JobKilled && e.Job?.Handle == running.Job.Handle && e.Reason == "kill_job"), "kill fact recorded");
            await Task.Delay(50);
            bool alive;
            try { using var child = Process.GetProcessById(pid); alive = !child.HasExited; }
            catch (ArgumentException) { alive = false; }
            Check(!alive, "kill ends process-group child");
            Check((await Call("kill_job", new { handle = running.Job.Handle })).Job?.Status == "killed", "kill idempotent");
            Check((await Call("send_input", new { handle = running.Job.Handle, input = "x" })).IsError, "stdin closed on finished job");

            using (var cancelWait = new CancellationTokenSource(50))
            {
                bool cancelled = false;
                try { var unexpected = await Call("run_command", new { command = "sleep 30", ai_callback_delay = 3 }, cancelWait.Token); throw new Exception("Expected cancelled observation; returned: " + Text(unexpected)); }
                catch (OperationCanceledException) { cancelled = true; }
                Check(cancelled && agent.Jobs().Children().Last().Snapshot().Status == "running", "observer cancellation does not kill");
            }
            var relativeDirectory = Path.GetRelativePath(Environment.CurrentDirectory, directory);
            var cwdResult = await Call("run_command", new { command = "pwd", cwd = relativeDirectory });
            Check(cwdResult.Job!.Cwd == Path.GetFullPath(relativeDirectory), "explicit cwd records its absolute path");
            _ = new Tool(agent, "broken", "Deliberately broken extension", JsonSerializer.SerializeToElement(new
            {
                type = "object", properties = new { }
            }), (job, args, token) => Task.FromException<string>(new InvalidOperationException("broken invariant")));
            bool propagated = false;
            try { await Call("broken", new { }); }
            catch (InvalidOperationException ex) when (ex.Message == "broken invariant") { propagated = true; }
            Check(propagated, "unexpected handler fault reaches the host, not the model");
            var broken = agent.Jobs().Children().MaxBy(j => j.Handle)!;
            try { broken.Destroy(); }
            catch (InvalidOperationException ex) when (ex.Message == "broken invariant") { }
            Check(!agent.Jobs().Children().Contains(broken), "faulted job cleanup still detaches");
            Check(!File.ReadAllText(pty.Job!.OutputPath).Contains("\r\n"), "PTY normalization happens before recording");
            var peer = new Agent(app, agent.Configuration());
            peer.InstallTools();
            var peerCall = new ToolCallPart("peer", peer.Configuration().Target(), "read_file",
                JsonSerializer.SerializeToElement(new { path = Path.Combine(directory, "absent") }));
            var peerResult = await peer.DispatchAsync(peerCall);
            Check(peerResult.Job!.Handle > agent.Jobs().Children().Max(j => j.Handle), "Ensemble-wide handles");
            bool foreign = false;
            try { peer.Jobs().Find(held.Job.Handle); }
            catch (ArgumentException) { foreign = true; }
            Check(foreign, "a global handle does not bypass Agent ownership");
            peer.History().Destroy();
            var noEffect = Path.Combine(directory, "must-not-be-written");
            bool refused = false;
            try
            {
                await peer.DispatchAsync(new ToolCallPart("no-log", peer.Configuration().Target(), "write_file",
                    JsonSerializer.SerializeToElement(new { path = noEffect, content = "bad" })));
            }
            catch (ObjectDisposedException) { refused = true; }
            Check(refused && !File.Exists(noEffect), "failed recording prevents tool effects");
            bool joinedFault = false;
            try
            {
                peer.Destroy();
            }
            catch (ObjectDisposedException)
            {
                joinedFault = true;
            }
            Check(joinedFault && peer.parent is null && !app.Agents().Contains(peer),
                "joining a faulted actor still completes destruction");
            var jobs = agent.Jobs();
            var children = jobs.Children();
            agent.Destroy();
            Check(jobs.parent is null && jobs.Children().Length == 0 && children.All(j => j.parent is null), "shutdown joins and detaches all jobs");
            Console.WriteLine("PASS jobs: durable output, handles, cursor, PTY input/kill, one-shot observation, shutdown");
        }
        finally
        {
            app.Destroy();
        }
    }
    // A working diagnostics sink lets the test observe the real worker's final
    // publication. No elapsed-time guess stands in for completion.
    private sealed class CompletionWriter : StringWriter
    {
        internal string? Match;
        internal readonly TaskCompletionSource Completion = new(TaskCreationOptions.RunContinuationsAsynchronously);
        public override void WriteLine(string? line)
        {
            base.WriteLine(line);
            if (Match is not null && line?.Contains(Match, StringComparison.Ordinal) == true)
            {
                Completion.TrySetResult();
            }
        }
    }

}
