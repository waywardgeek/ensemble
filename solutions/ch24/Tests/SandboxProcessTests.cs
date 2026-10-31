using System.Diagnostics;
using System.Text.Json;
using Ensemble;
using Application = Ensemble.Ensemble;

internal static class SandboxProcessTests
{
    public static async Task Boundary()
    {
        await InWorkspace(async (agent, inside, outside) =>
        {
            string sentinel = Path.Combine(outside, "sentinel");
            File.WriteAllText(sentinel, "OUTSIDE-CONTENT");
            Directory.CreateSymbolicLink(Path.Combine(inside, "escape"), outside);
            foreach (string path in new[] { sentinel, "../outside/sentinel", "escape/sentinel" })
            {
                var read = await Command(agent, "cat " + Quote(path));
                Check(read.Job?.ExitCode != 0 && !Text(read).Contains("OUTSIDE-CONTENT"), "Command read outside root");
            }
            string nested = "/bin/sh -c " + Quote("cat " + Quote(sentinel));
            var grandchild = await Command(agent, "/bin/sh -c " + Quote(nested));
            Check(grandchild.Job?.ExitCode != 0 && !Text(grandchild).Contains("OUTSIDE-CONTENT"),
                "Grandchild escaped process policy");
            var write = await Command(agent, "printf changed > " + Quote(sentinel));
            Check(write.Job?.ExitCode != 0 && File.ReadAllText(sentinel) == "OUTSIDE-CONTENT", "Command wrote outside root");
            var cwd = await Call(agent, "run_command", new { command = "echo MUST-NOT-RUN", cwd = outside });
            Check(cwd.IsError && !Text(cwd).Contains("MUST-NOT-RUN"), "Outside cwd accepted");
            var relativeCwd = await Call(agent, "run_command", new { command = "echo MUST-NOT-RUN", cwd = "escape" });
            Check(relativeCwd.IsError && Text(relativeCwd).Contains("escape")
                && !Text(relativeCwd).Contains(Path.GetDirectoryName(inside)!),
                "Relative cwd refusal leaked a constructed host path in its report");
            Check(relativeCwd.Job is not null && File.Exists(relativeCwd.Job.OutputPath),
                "Trusted host output metadata no longer locates retained output");
            // The cache exception permits only immediate xcrun_db* files. All
            // three paths below are newly owned by this fixture, not real caches.
            var query = new ProcessStartInfo("/usr/bin/getconf")
            {
                RedirectStandardOutput = true,
                UseShellExecute = false
            };
            query.ArgumentList.Add("DARWIN_USER_TEMP_DIR");
            using var getconf = Process.Start(query)!;
            string cacheDirectory = (await getconf.StandardOutput.ReadToEndAsync()).Trim();
            await getconf.WaitForExitAsync();
            Check(getconf.ExitCode == 0 && Path.IsPathRooted(cacheDirectory), "Could not locate toolchain cache fixture");
            string suffix = Guid.NewGuid().ToString("N");
            string cacheFile = Path.Combine(cacheDirectory, "xcrun_db-ensemble-" + suffix);
            string unrelated = Path.Combine(cacheDirectory, "ensemble-unrelated-" + suffix);
            string nestedDirectory = Path.Combine(cacheDirectory, "ensemble-nested-" + suffix);
            Directory.CreateDirectory(nestedDirectory);
            string nestedFile = Path.Combine(nestedDirectory, "xcrun_db-private");
            File.WriteAllText(cacheFile, "owned-cache");
            File.WriteAllText(unrelated, "OUTSIDE-CONTENT");
            File.WriteAllText(nestedFile, "OUTSIDE-CONTENT");
            try
            {
                var allowed = await Command(agent, "printf toolchain > " + Quote(cacheFile) + "; cat " + Quote(cacheFile));
                Check(allowed.Job?.ExitCode == 0 && Text(allowed).Contains("toolchain"), "Narrow cache write/read unavailable");
                foreach (string forbidden in new[] { unrelated, nestedFile })
                {
                    var denied = await Command(agent, "cat " + Quote(forbidden));
                    Check(denied.Job?.ExitCode != 0 && !Text(denied).Contains("OUTSIDE-CONTENT"), "Cache exception escaped its exact scope");
                }
            }
            finally
            {
                File.Delete(cacheFile);
                File.Delete(unrelated);
                Directory.Delete(nestedDirectory, recursive: true);
            }
            var listener = new System.Net.Sockets.TcpListener(System.Net.IPAddress.Loopback, 0);
            listener.Start();
            try
            {
                int port = ((System.Net.IPEndPoint)listener.LocalEndpoint).Port;
                string program = "import socket\ns=socket.socket()\ns.settimeout(2)\ntry:\n s.connect(('127.0.0.1',"
                    + port + "))\n print('NETWORK-ALLOWED')\nexcept OSError:\n print('NETWORK-DENIED')";
                var network = await Command(agent, "python3 -c " + Quote(program));
                Check(network.Job?.ExitCode == 0 && Text(network).Contains("NETWORK-DENIED") && !listener.Pending(),
                    "Confined child reached loopback network");
            }
            finally
            {
                listener.Stop();
            }
            Check(!Directory.EnumerateFiles(agent.Jobs().OutputDirectory, "*.pgid").Any(), "Writable PID control file retained");
        });
    }

    public static async Task UsefulWork()
    {
        await InWorkspace(async (agent, inside, outside) =>
        {
            const string variable = "SANDBOX_TEST_UNUSUAL_CREDENTIAL";
            string? previous = Environment.GetEnvironmentVariable(variable);
            string? previousHook = Environment.GetEnvironmentVariable("BASH_ENV");
            string hook = Path.Combine(inside, "startup-hook");
            File.WriteAllText(hook, "printf HOOK-RAN\n");
            Environment.SetEnvironmentVariable("BASH_ENV", hook);
            Environment.SetEnvironmentVariable(variable, "synthetic-secret");
            try
            {
                var environment = await Command(agent,
                    "test -z \"${" + variable + "+set}\" && test \"$HOME\" = " + Quote(inside)
                    + " && test \"$TMPDIR\" = " + Quote(inside)
                    + " && test \"$PWD\" = " + Quote(inside)
                     + " && test \"$GIT_CONFIG_SYSTEM\" = /dev/null && /bin/bash -c 'echo FILTERED'");
                Check(environment.Job?.ExitCode == 0 && Text(environment).Contains("FILTERED")
                    && !Text(environment).Contains("HOOK-RAN"), "Confined environment not allowlisted");
            }
            finally
            {
                Environment.SetEnvironmentVariable(variable, previous);
                Environment.SetEnvironmentVariable("BASH_ENV", previousHook);
            }
            var concurrent = await Task.WhenAll(Enumerable.Range(0, 4).Select(_ =>
                Command(agent, "git config --global --get user.name")));
            Check(concurrent.All(result => result.Job?.ExitCode == 0 && Text(result).Contains("Ensemble")),
                "Concurrent launches observed incomplete Git identity");
            Check(!Directory.EnumerateFiles(inside, ".ensemble-gitconfig-*").Any(), "Generated Git temporary files remain");
            Directory.CreateDirectory(Path.Combine(inside, "work"));
            var useful = await Command(agent,
                "test -t 0 && test -t 1 && stty size && "
                + "python3 -c 'from pathlib import Path; Path(\"python.txt\").write_text(\"local\"); print(\"PYTHON-OK\")' && "
                + "git init -q repo && git -C repo config --get user.name && "
                + "printf local > repo/tracked && git -C repo add tracked && git -C repo status --porcelain", "work");
            Check(useful.Job?.ExitCode == 0 && Text(useful).Contains("50 200")
                && Text(useful).Contains("PYTHON-OK") && Text(useful).Contains("Ensemble")
                && Text(useful).Contains("A  tracked"), "Legitimate PTY/Git/Python work failed: " + Text(useful));
            Check(useful.Job?.Cwd == Path.Combine(inside, "work"), "Confined cwd report used host directory");

            var prompt = await Call(agent, "run_command", new
            {
                command = "printf 'PROMPT>'; read answer; printf 'answer:%s\\n' \"$answer\"",
                ai_callback_pattern = "PROMPT>", ai_callback_delay = 3
            });
            Check(prompt.Job?.Status == "running" && Text(prompt).Contains("PROMPT>"), "Confined PTY prompt unavailable");
            var input = await Call(agent, "send_input", new { handle = prompt.Job!.Handle, input = "hello", ai_callback_delay = 3 });
            input = await Finish(agent, input);
            Check(Text(input).Contains("answer:hello") && input.Job?.ExitCode == 0, "Confined PTY stdin failed");

            var running = await Call(agent, "run_command", new
            {
                command = "sleep 60 & printf 'child:%s\\n' \"$!\"; wait",
                ai_callback_pattern = "child:[0-9]+", ai_callback_delay = 3
            });
            string line = File.ReadLines(running.Job!.OutputPath).Single(value => value.StartsWith("child:", StringComparison.Ordinal));
            int pid = int.Parse(line[6..]);
            var killed = await Call(agent, "kill_job", new { handle = running.Job.Handle });
            Check(killed.Job?.Status == "killed", "Confined job did not join cancellation");
            using var deadline = new CancellationTokenSource(TimeSpan.FromSeconds(5));
            while (Alive(pid))
            {
                await Task.Delay(10, deadline.Token);
            }
            foreach (string path in Directory.EnumerateFiles(agent.Jobs().OutputDirectory))
            {
                Check(!File.ReadAllText(path).Contains("ENSEMBLE_PGID="), "Startup control entered retained output");
            }
        });
    }

    private static bool Alive(int pid)
    {
        try
        {
            using var process = Process.GetProcessById(pid);
            return !process.HasExited;
        }
        catch (ArgumentException)
        {
            return false;
        }
    }

    private static async Task<ToolResultPart> Command(Agent agent, string command, string cwd = "")
    {
        return await Finish(agent, await Call(agent, "run_command", new { command, cwd, ai_callback_delay = 3 }));
    }

    private static async Task<ToolResultPart> Finish(Agent agent, ToolResultPart result)
    {
        using var deadline = new CancellationTokenSource(TimeSpan.FromSeconds(15));
        while (result.Job?.Status == "running")
        {
            result = await Call(agent, "wait_for_job", new { handle = result.Job.Handle, ai_callback_delay = 1 }, deadline.Token);
        }
        // The last observation need not contain earlier chunks. Assertions use
        // the real retained file, not a test-side reconstruction of those chunks.
        if (result.Job is not null)
        {
            return new ToolResultPart(result.CallID, [new TextPart(File.ReadAllText(result.Job.OutputPath))], result.IsError, result.Job);
        }
        return result;
    }

    private static Task<ToolResultPart> Call(Agent agent, string name, object args, CancellationToken token = default) =>
        agent.DispatchAsync(new ToolCallPart(Guid.NewGuid().ToString("N"), agent.Configuration().Target(), name,
            JsonSerializer.SerializeToElement(args)), token);

    private static string Text(ToolResultPart result) => JsonCodec.Text(result.Parts);
    private static string Quote(string value) => "'" + value.Replace("'", "'\\''", StringComparison.Ordinal) + "'";
    private static void Check(bool value, string message)
    {
        if (!value)
        {
            throw new Exception(message);
        }
    }

    private static async Task InWorkspace(Func<Agent, string, string, Task> test)
    {
        string workspace = Path.Combine(Path.GetTempPath(), "rooted-process-" + Guid.NewGuid().ToString("N"));
        string inside = Path.Combine(workspace, "inside ' quoted \" root");
        string outside = Path.Combine(workspace, "outside");
        Directory.CreateDirectory(inside);
        Directory.CreateDirectory(outside);
        var app = new Application(TextWriter.Null);
        var agent = new Agent(app, new Configuration(Vendor.OpenAI, "offline", "http://127.0.0.1:1", sandboxRoot: inside));
        agent.InstallTools();
        try
        {
            await test(agent, agent.Sandbox().Root(), outside);
        }
        finally
        {
            app.Destroy();
            Directory.Delete(workspace, recursive: true);
        }
    }
}
