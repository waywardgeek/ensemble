using System.Diagnostics;
using System.Text.Json;
using Ensemble.Gateway;

internal static partial class Program
{
    private static Process StartPersistenceChild(string mode, string home)
    {
        ProcessStartInfo start = new(Environment.ProcessPath!)
        {
            RedirectStandardOutput = true,
            RedirectStandardError = true,
            UseShellExecute = false
        };
        start.ArgumentList.Add(typeof(Program).Assembly.Location);
        start.ArgumentList.Add(mode);
        start.ArgumentList.Add(home);
        return Process.Start(start) ?? throw new Exception("Could not start persistence child.");
    }

    private static async Task<int> CrashWriter(string home)
    {
        RouterConfig config = Router.ReadConfig(Path.Combine(home, "config.json"));
        await using Router router = new(config, Keys.LoadOrCreateKeypair(Path.Combine(home, "test.key")));
        Call(router, "bob", "declare-contactable", new { groups = new[] { "agents" } });
        router.Route(router.Identity("alice"), new("bob", Path.GetFileName(home)));
        Console.WriteLine("ACCEPTED");
        Console.Out.Flush();
        await Task.Delay(Timeout.InfiniteTimeSpan);
        return 0;
    }

    private static async Task<int> LeaseContender(string home)
    {
        try
        {
            RouterConfig config = Router.ReadConfig(Path.Combine(home, "config.json"));
            await using Router router = new(config, Keys.LoadOrCreateKeypair(Path.Combine(home, "test.key")));
            return 2;
        }
        catch (IOException)
        {
            Console.WriteLine("LEASE REFUSED");
            return 0;
        }
    }

    private static async Task ProcessCrash()
    {
        string home = Home();
        RouterConfig config = Configuration(home);
        File.WriteAllText(Path.Combine(home, "config.json"), JsonSerializer.Serialize(config));
        Keypair key = Keys.LoadOrCreateKeypair(Path.Combine(home, "test.key"));
        using Process writer = StartPersistenceChild("--crash-writer", home);
        Task<string> writerErrors = writer.StandardError.ReadToEndAsync();
        try
        {
            string? accepted = await writer.StandardOutput.ReadLineAsync().WaitAsync(TimeSpan.FromSeconds(20));
            Check(accepted == "ACCEPTED", "Child failed before durable acceptance");
            using Process contender = StartPersistenceChild("--lease-contender", home);
            Task<string> contenderErrors = contender.StandardError.ReadToEndAsync();
            Task<string> contenderOutput = contender.StandardOutput.ReadToEndAsync();
            try
            {
                await contender.WaitForExitAsync().WaitAsync(TimeSpan.FromSeconds(20));
                Check(contender.ExitCode == 0 && (await contenderOutput).Trim() == "LEASE REFUSED",
                    "Independent process acquired live directory: " + await contenderErrors);
            }
            finally
            {
                if (!contender.HasExited) contender.Kill(entireProcessTree: true);
                await contender.WaitForExitAsync();
                await contenderErrors;
                await contenderOutput;
            }
            // No Dispose, host stop signal or final snapshot: acceptance alone
            // must have made the delivery recoverable before this hard exit.
            writer.Kill(entireProcessTree: true);
            await writer.WaitForExitAsync();
            await using (Router restored = new(config, key))
            {
                QueueState queue = restored.Identity("bob").Queue().Snapshot();
                Check(queue.Next == 1 && queue.Ack == 0 &&
                    queue.Messages!.Value.Single().Content.EndsWith(Path.GetFileName(home), StringComparison.Ordinal),
                    "Abrupt exit lost or changed accepted delivery");
                restored.Identity("bob").Queue().Acknowledge(1);
            }
            await using Router drained = new(config, key);
            QueueState after = drained.Identity("bob").Queue().Snapshot();
            Check(after.Next == 1 && after.Ack == 1 && after.Messages!.Value.IsEmpty,
                "Receipt removal did not survive restart");
        }
        finally
        {
            if (!writer.HasExited) writer.Kill(entireProcessTree: true);
            await writer.WaitForExitAsync();
            string errors = await writerErrors;
            if (errors.Length != 0) Console.Error.WriteLine(errors);
        }
    }
}
