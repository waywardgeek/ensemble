using System.Diagnostics;
using System.Text.Json;
using Ensemble;
using Application = Ensemble.Ensemble;

internal static class AuthStoreTests
{
    private static void Check(bool condition, string reason)
    {
        if (!condition)
        {
            throw new InvalidOperationException(reason);
        }
    }

    private static AuthRegistration Grant(string label, string? client = null, string subject = "person") =>
        new(label, "https://issuer.invalid", subject, client ?? "client-" + label, "same@example.invalid",
            "private-access-" + label, "private-refresh-" + label, "private-id-" + label,
            DateTimeOffset.UtcNow.AddHours(1), "openid offline_access chatgpt.tokens.use.direct");

    private static async Task Refuses(Func<Task> operation)
    {
        try
        {
            await operation();
        }
        catch (Exception error) when (error is IOException or InvalidDataException or InvalidOperationException)
        {
            Check(!error.ToString().Contains("private-access", StringComparison.Ordinal)
                && !error.ToString().Contains("private-refresh", StringComparison.Ordinal)
                && !error.ToString().Contains("private-id", StringComparison.Ordinal), "errors never print credential bytes");
            return;
        }
        throw new Exception("Expected credential refusal.");
    }

    public static async Task OwnershipAndAccounts(string temp)
    {
        string directory = Path.Combine(temp, "auth-accounts");
        var app = new Application(TextWriter.Null);
        var store = new Credentials(app, directory);
        try
        {
            Check(ReferenceEquals(app.Credentials(), store) && ReferenceEquals(store.Parent(), app), "both ownership links");
            await Refuses(() => Task.Run(() => new Credentials(app, directory)));
            Check(ReferenceEquals(app.Credentials(), store), "failed duplicate construction preserves original");
            Check(!Directory.Exists(directory), "construction does not create credential files");
            Check((await store.AccountsAsync()).IsEmpty, "new store has no accounts");
            string host = await store.HostIDAsync();
            Check(host.StartsWith("urn:uuid:", StringComparison.Ordinal) && host.Length == 45
                && Guid.TryParseExact(host[9..], "D", out _) && host[23] == '4',
                "new host identity is an opaque version-four UUID URN");
            await store.StoreAsync(Grant("one"));
            await store.StoreAsync(Grant("two"));
            var accounts = await store.AccountsAsync();
            Check(accounts.Length == 2 && accounts.All(value => value.SignedIn), "same email keeps distinct registrations");
            string publicJSON = JsonSerializer.Serialize(accounts);
            Check(publicJSON.Contains("client-one", StringComparison.Ordinal)
                && publicJSON.Contains("same@example.invalid", StringComparison.Ordinal)
                && !publicJSON.Contains("private-", StringComparison.Ordinal),
                "public status contains account metadata but excludes all token fields");
            byte[] before = File.ReadAllBytes(Path.Combine(directory, "credentials.json"));
            await Refuses(() => store.StoreAsync(Grant("one", "different-client")));
            await Refuses(() => store.StoreAsync(Grant("renamed", "client-one")));
            await Refuses(() => store.StoreAsync(Grant("impostor", "client-one", "different-person")));
            Check(File.ReadAllBytes(Path.Combine(directory, "credentials.json")).SequenceEqual(before), "identity refusals preserve file bytes");
            await store.LogoutAsync("one");
            var after = await store.AccountsAsync();
            Check(!after[0].SignedIn && after[0].ClientID == "client-one" && after[1].SignedIn, "logout isolates one account and keeps registration");
            string file = File.ReadAllText(Path.Combine(directory, "credentials.json"));
            Check(!file.Contains("private-access-one") && !file.Contains("private-refresh-one")
                && !file.Contains("private-id-one") && file.Contains("private-access-two"), "all selected secrets cleared, unrelated tokens preserved");
            await store.LogoutAsync("one");
            await Refuses(() => store.LogoutAsync("missing"));
            Check(await store.HostIDAsync() == host, "host identity survives logout");
            app.Destroy();
            Check(app.Credentials() is null && store.parent is null, "root destruction detaches both ends");
            store.Destroy();
            await Refuses(() => store.AccountsAsync());
            var restored = new Application(TextWriter.Null);
            try
            {
                var reopened = new Credentials(restored, directory);
                Check(await reopened.HostIDAsync() == host && !(await reopened.AccountsAsync())[0].SignedIn,
                    "restart retains host and local logout");
                await reopened.StoreAsync(Grant("one"));
                Check((await reopened.AccountsAsync()).All(value => value.SignedIn), "same registration may sign in again");
                reopened.Destroy();
                Check(restored.Credentials() is null, "explicit child destruction removes parent membership");
                new Credentials(restored, directory); // A detached store does not reserve its old slot.
            }
            finally
            {
                restored.Destroy();
            }
        }
        finally
        {
            app.Destroy();
        }
    }

    public static async Task RefusalAndPermissions(string temp)
    {
        if (OperatingSystem.IsWindows())
        {
            throw new PlatformNotSupportedException("This test exercises the Unix credential storage policy.");
        }
        string directory = Path.Combine(temp, "auth-refusal");
        var app = new Application(TextWriter.Null);
        var store = new Credentials(app, directory);
        try
        {
            await store.StoreAsync(Grant("one"));
            string path = Path.Combine(directory, "credentials.json");
            string original = File.ReadAllText(path);
            Check(File.GetUnixFileMode(directory) == (UnixFileMode.UserRead | UnixFileMode.UserWrite | UnixFileMode.UserExecute), "private directory from creation");
            Check(File.GetUnixFileMode(path) == (UnixFileMode.UserRead | UnixFileMode.UserWrite), "private file from creation");
            foreach (string damaged in new[]
            {
                "{private-access-one", original.Replace("\"version\":2", "\"version\":\"private-access-one\""),
                original.Replace("\"version\":2", "\"version\":3"),
                original.Replace("\"version\":2", "\"version\":2,\"version\":2"),
                original.Replace("\"selected_plan_account\":null", "\"selected_plan_account\":\"missing\""),
                original.Replace("\"client-one\"", "\"dynamic_agent_client\""),
                original.Replace("\"private-refresh-one\"", "null")
            })
            {
                Check(damaged != original, "damage fixture must actually change bytes");
                File.WriteAllText(path, damaged);
                await Refuses(() => store.AccountsAsync());
                await Refuses(() => store.StoreAsync(Grant("two")));
                Check(File.ReadAllText(path) == damaged, "bad store is not silently replaced");
            }
            string legacy = original.Replace("\"version\":2", "\"version\":1")
                .Replace(",\"selected_plan_account\":null", "");
            File.WriteAllText(path, legacy);
            Check((await store.AccountsAsync()).Length == 1 && await store.SelectedPlanAccountAsync() is null,
                "legacy version one has no saved plan selection");
            Check(File.ReadAllText(path) == legacy, "reading legacy store does not rewrite it");
            File.WriteAllText(path, original);
            File.SetUnixFileMode(path, UnixFileMode.UserRead | UnixFileMode.UserWrite | UnixFileMode.OtherRead);
            await Refuses(() => store.AccountsAsync());
            Check(File.GetUnixFileMode(path).HasFlag(UnixFileMode.OtherRead), "refusal does not chmod existing files");
            File.SetUnixFileMode(path, UnixFileMode.UserRead | UnixFileMode.UserWrite);
            string held = Path.Combine(directory, "held.json");
            File.Move(path, held);
            File.CreateSymbolicLink(path, held);
            await Refuses(() => store.AccountsAsync());
            Check(File.ReadAllText(held) == original, "symlink target not modified");
            File.Delete(path);
            File.Move(held, path);
            string lockPath = Path.Combine(directory, "credentials.lock");
            File.Delete(lockPath);
            File.CreateSymbolicLink(lockPath, path);
            await Refuses(() => store.AccountsAsync());
            File.Delete(lockPath);
            Check((await store.AccountsAsync()).Length == 1, "repair enables later reads, no cached failure");
            Check(Directory.GetFiles(directory, ".credentials-*").Length == 0, "no secret-bearing temporary files remain");
        }
        finally
        {
            app.Destroy();
        }
    }

    // This subprocess mode invokes the real store, not a scripted lock mock.
    public static async Task<int> Child(string directory, string label)
    {
        var app = new Application(TextWriter.Null);
        try
        {
            var store = new Credentials(app, directory);
            File.WriteAllText(Path.Combine(directory, label + ".ready"), "ready");
            string host = await store.HostIDAsync();
            await store.StoreAsync(Grant(label));
            Console.WriteLine(host);
            return 0;
        }
        finally
        {
            app.Destroy();
        }
    }

    public static async Task Contention(string temp)
    {
        string directory = Path.Combine(temp, "auth-contention");
        var app = new Application(TextWriter.Null);
        var store = new Credentials(app, directory);
        var children = new List<Process>();
        using var deadline = new CancellationTokenSource(TimeSpan.FromSeconds(20));
        try
        {
            string host = await store.HostIDAsync();
            using (var held = new FileStream(Path.Combine(directory, "credentials.lock"), FileMode.Open,
                FileAccess.ReadWrite, FileShare.None))
            {
                using var cancelled = new CancellationTokenSource(TimeSpan.FromMilliseconds(150));
                try
                {
                    await store.AccountsAsync(cancelled.Token);
                    throw new Exception("Contended lock did not wait.");
                }
                catch (OperationCanceledException)
                {
                    Check(cancelled.IsCancellationRequested, "caller cancellation ends only the wait");
                }
                catch (IOException error)
                {
                    throw new Exception($"Lock contention HResult: 0x{error.HResult:X8}", error);
                }
                foreach (string label in new[] { "child-a", "child-b" })
                {
                    var start = new ProcessStartInfo(Environment.ProcessPath!)
                    {
                        UseShellExecute = false, RedirectStandardOutput = true, RedirectStandardError = true
                    };
                    // dotnet run may use an apphost or the dotnet host itself.
                    if (Path.GetFileNameWithoutExtension(Environment.ProcessPath) == "dotnet")
                    {
                        start.ArgumentList.Add(typeof(AuthStoreTests).Assembly.Location);
                    }
                    start.ArgumentList.Add("--auth-store-child");
                    start.ArgumentList.Add(directory);
                    start.ArgumentList.Add(label);
                    children.Add(Process.Start(start)!);
                }
                while (!File.Exists(Path.Combine(directory, "child-a.ready"))
                    || !File.Exists(Path.Combine(directory, "child-b.ready")))
                {
                    await Task.Delay(20, deadline.Token);
                }
                Check(children.All(child => !child.HasExited), "both real processes wait behind held lock");
            }
            foreach (var child in children)
            {
                await child.WaitForExitAsync(deadline.Token);
                string output = await child.StandardOutput.ReadToEndAsync();
                string errors = await child.StandardError.ReadToEndAsync();
                Check(child.ExitCode == 0 && output.Trim() == host && errors.Length == 0,
                    "each child reused host and published successfully: " + errors);
            }
            var accounts = await store.AccountsAsync();
            Check(accounts.Length == 2 && accounts.Select(account => account.Label).Distinct().Count() == 2,
                "concurrent whole-file transactions lose neither registration");
        }
        finally
        {
            foreach (var child in children)
            {
                if (!child.HasExited)
                {
                    child.Kill(entireProcessTree: true);
                    await child.WaitForExitAsync();
                }
                child.Dispose();
            }
            app.Destroy();
        }
    }
}
