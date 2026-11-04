using System.Text.Json;
using Ensemble.Gateway;

internal static partial class Program
{
    private static async Task DiskValidation()
    {
        try
        {
            _ = Router.ReadConfig(Path.Combine(Home(), "absent.json"));
            throw new Exception("Missing configuration accepted.");
        }
        catch (FileNotFoundException) { }
        string valid = """
            {"next":2,"ack":1,"messages":[{"id":"bob:2","seq":2,"to_agent":"bob","content":"historical",
            "timestamp":"2026-01-02T03:04:05Z","source":{"type":"agent","channel_id":"removed","user_id":"old-sender","trust":"trusted"}}]}
            """;
        string[] cases =
        [
            valid, "{\"next\":2,\"ack\":2,\"messages\":null}",
            "{\"next\":18446744073709551615,\"ack\":18446744073709551615,\"messages\":null}",
            "{\"next\":2,\"ack\":3,\"messages\":[]}",
            "{\"next\":2,\"ack\":1,\"messages\":[]}",
            "{\"next\":2,\"next\":2,\"ack\":2,\"messages\":[]}",
            "{\"next\":2,\"ack\":2}", "null", "{",
            valid.Replace("bob:2", "wrong"), valid.Replace("\"seq\":2", "\"seq\":1")
        ];
        for (int index = 0; index < cases.Length; index++)
        {
            string home = Home();
            RouterConfig config = Configuration(home);
            Keypair key = Keys.GenerateKeypair();
            Directory.CreateDirectory(Path.Combine(home, "queues"));
            string path = Path.Combine(home, "queues", "bob.json");
            File.WriteAllText(path, cases[index]);
            Refuses(() => _ = new Router(config, key));
            if (index < 3)
            {
                await using Router router = Router.MigrateLegacy(config, key);
                QueueState queue = router.Identity("bob").Queue().Snapshot();
                if (index == 0)
                    Check(queue.Messages!.Value.Single().ConversationID == "", "Legacy omitted fields not normalized");
                if (index == 1) Check(queue.Ack == 2 && queue.Messages!.Value.Length == 0, "Drained receipt lost");
                if (index == 2)
                {
                    Call(router, "bob", "declare-contactable", new { groups = new[] { "agents" } });
                    Refuses(() => router.Route(router.Identity("alice"), new("bob", "overflow")));
                    Check(router.Identity("bob").Queue().Snapshot().Next == ulong.MaxValue, "Overflow allocated sequence");
                }
            }
            else Refuses(() => _ = Router.MigrateLegacy(config, key));
            Check(File.ReadAllText(path) == cases[index], "Migration overwrote original queue");
        }
        string emptyHome = Home();
        RouterConfig emptyConfig = Configuration(emptyHome) with
        {
            Agents = System.Collections.Immutable.ImmutableDictionary<string, IdentityDefinition>.Empty
        };
        File.WriteAllText(Path.Combine(emptyHome, "queues"), "must not disappear");
        Refuses(() => _ = Router.MigrateLegacy(emptyConfig, Keys.GenerateKeypair()));
        Check(!File.Exists(Path.Combine(emptyHome, "gateway-state.json")),
            "Empty-identity migration ignored obstructed legacy queue path");
        string obstructed = Home();
        File.WriteAllText(Path.Combine(obstructed, "queues"), "obstruction");
        Refuses(() => _ = Router.MigrateLegacy(Configuration(obstructed), Keys.GenerateKeypair()));
        string dangling = Home();
        File.CreateSymbolicLink(Path.Combine(dangling, "state.json"), Path.Combine(dangling, "missing.json"));
        Refuses(() => _ = Router.MigrateLegacy(Configuration(dangling), Keys.GenerateKeypair()));
        await CompleteStateRecovery();
    }

    private static async Task CompleteStateRecovery()
    {
        string home = Home();
        RouterConfig config = Configuration(home);
        Keypair key = Keys.GenerateKeypair();
        string primary = Path.Combine(home, "gateway-state.json");
        string backup = Path.Combine(home, "gateway-state.backup.json");
        Router router = new(config, key);
        Refuses(() => _ = new Router(config, key));
        Call(router, "bob", "declare-contactable", new { groups = new[] { "agents" } });
        router.Route(router.Identity("alice"), new("bob", "one"));
        router.Route(router.Identity("alice"), new("bob", "two"));
        // Read disk before Dispose: shutdown must not be what makes acceptance durable.
        using (JsonDocument disk = JsonDocument.Parse(File.ReadAllText(primary)))
            Check(disk.RootElement.GetProperty("queues").GetProperty("bob").GetProperty("next").GetUInt64() == 2,
                "Accepted queue was not on disk before shutdown");
        string committed = File.ReadAllText(primary);
        await router.DisposeAsync();
        Check(File.ReadAllText(primary) == committed, "Shutdown rewrote durable state");
        Refuses(() => _ = new Router(config, Keys.GenerateKeypair()));
        string[] corrupt =
        [
            "{", "null", committed.Replace("\"version\":1", "\"version\":2"),
            committed.Replace("\"version\":1", "\"version\":1,\"version\":1"),
            committed.Replace("\"version\":1", "\"unexpected\":0,\"version\":1"),
            committed.Replace("\"next\":2", "\"next\":3")
        ];
        foreach (string bad in corrupt)
        {
            File.WriteAllText(primary, bad);
            Refuses(() => _ = new Router(config, key));
            Check(File.ReadAllText(primary) == bad, "Startup silently repaired corrupt state");
        }
        string evidence = File.ReadAllText(primary);
        await using (Router recovered = Router.RecoverBackup(config, key))
            Check(recovered.Identity("bob").Queue().Snapshot().Next == 1, "Recovery did not restore validated prior generation");
        Check(Directory.GetFiles(home, "gateway-state.recovery-*.json").Any(path => File.ReadAllText(path) == evidence), "Recovery lost primary evidence");
        File.Delete(primary);
        Refuses(() => _ = new Router(config, key));
        await using (Router recovered = Router.RecoverBackup(config, key))
            Check(recovered.Identity("bob").Queue().Snapshot().Next == 1, "Missing-primary recovery failed");
        string before = File.ReadAllText(primary);
        File.WriteAllText(backup, "{");
        Refuses(() => _ = Router.RecoverBackup(config, key));
        Check(File.ReadAllText(primary) == before, "Invalid backup overwrote primary");
    }

    private static async Task SaveFailure()
    {
        // A real filesystem boundary failure, not a mocked Router or queue.
        // Test admission, removal and permission publication separately.
        foreach (string mutation in new[] { "enqueue", "receipt", "grant" })
        {
            string home = Home();
            RouterConfig config = Configuration(home);
            Keypair key = Keys.GenerateKeypair();
            Router router = new(config, key);
            Call(router, "bob", "declare-contactable", new { groups = new[] { "agents" } });
            router.Route(router.Identity("alice"), new("bob", "retained"));
            RouterIdentity alice = router.Identity("alice");
            RouterIdentity bob = router.Identity("bob");
            DeliveryQueue queue = bob.Queue();
            string primary = Path.Combine(home, "gateway-state.json");
            string before = File.ReadAllText(primary);
            string backup = Path.Combine(home, "gateway-state.backup.json");
            File.Delete(backup);
            Directory.CreateDirectory(backup);
            if (mutation == "enqueue") Refuses(() => router.Route(alice, new("bob", "must not publish")));
            if (mutation == "receipt") Refuses(() => queue.Acknowledge(1));
            if (mutation == "grant") Refuses(() => Call(router, "bob", "declare-contactable", new { groups = Array.Empty<string>() }));
            Check(queue.Snapshot().Next == 1 && queue.Snapshot().Ack == 0 && queue.After(0) is not null,
                "Failed save changed live queue");
            Check(bob.Groups(GroupLink.Contact).Length == 1, "Failed save changed live permission");
            Check(File.ReadAllText(primary) == before, "Failed backup write changed primary");
            Directory.Delete(backup);
            Refuses(router.Save);
            Refuses(() => router.Route(alice, new("bob", "must remain closed")));
            await router.DisposeAsync();
            Check(alice.parent is null && bob.parent is null && queue.parent is null, "Poisoned teardown retained ownership");
            Check(File.ReadAllText(primary) == before, "Teardown overwrote commit evidence");
            await using Router restored = new(config, key);
            Check(restored.Identity("bob").Queue().Snapshot().Next == 1, "Restart lost committed state");
        }
        await PublicationFailure();
    }

    private static async Task PublicationFailure()
    {
        string home = Home();
        RouterConfig config = Configuration(home);
        Keypair key = Keys.GenerateKeypair();
        Router router = new(config, key);
        Call(router, "bob", "declare-contactable", new { groups = new[] { "agents" } });
        router.Route(router.Identity("alice"), new("bob", "old live value"));
        RouterIdentity bob = router.Identity("bob");
        QueueState live = bob.Queue().Snapshot();
        Envelope changed = live.Messages!.Value.Single() with { Content = "committed candidate" };
        QueueState candidate = live with { Messages = [changed] };
        // The failure is after actual replacement, before live publication.
        // This exercises ambiguous-outcome handling without a production hook.
        lock (router.gate)
            Refuses(() => router.CommitQueue(bob, candidate, () => throw new IOException("publication fault")));
        Check(bob.Queue().Snapshot().Messages!.Value.Single().Content != changed.Content,
            "Injected publication unexpectedly changed live queue");
        Refuses(router.Save);
        string primary = Path.Combine(home, "gateway-state.json");
        string committed = File.ReadAllText(primary);
        Check(committed.Contains("committed candidate", StringComparison.Ordinal), "Candidate was not committed before fault");
        await router.DisposeAsync();
        Check(File.ReadAllText(primary) == committed, "Shutdown overwrote post-commit evidence");
        await using Router restored = new(config, key);
        Check(restored.Identity("bob").Queue().Snapshot().Messages!.Value.Single().Content == changed.Content,
            "Restart failed to recover committed candidate after publication fault");
    }
}
