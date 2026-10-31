using System.Collections.Immutable;
using System.Text.Json;
using Ensemble.Gateway;

// The policy groups exercise the real graph, RPC dispatcher and filesystem.
// SessionChecks adds real loopback WebSockets, independent JSON and backpressure.
internal static class Program
{
    private static async Task<int> Main(string[] args)
    {
        int failed = 0;
        List<(string Name, Func<Task> Run)> cases =
        [
            ("reciprocal ownership and refusal-before-mutation", Relationships),
            ("accepted-group provenance and sequential queue admission", Routing),
            ("narrow child registration and live runtime administration", Administration),
            ("snapshot restore preserves policy, pending messages and drained receipt", Restart),
            ("independent Go-shaped files and fail-closed restore", DiskValidation),
            ("save failure is retryable; failed close still detaches", SaveFailure),
            ("RPC contracts and published strict tool schemas", Rpc),
            ("concurrent admission and save preserve contiguous queue", ConcurrentAdmission)
        ];
        if (args.SequenceEqual(["--sessions-only"])) cases.Clear();
        cases.AddRange(SessionChecks.Cases);
        if (args.Length == 2 && args[0] == "--go-fixture")
            cases.Add(("restore retained real-Go policy and receipt three", () => GoSnapshot(args[1])));
        else if (args.Length != 0 && !args.SequenceEqual(["--sessions-only"]))
            throw new ArgumentException("Usage: RouterTests [--sessions-only | --go-fixture DIRECTORY]");
        foreach (var test in cases)
        {
            try { await test.Run(); Console.WriteLine("PASS " + test.Name); }
            catch (Exception error) { failed++; Console.Error.WriteLine("FAIL " + test.Name + ": " + error); }
        }
        Console.WriteLine($"{cases.Count - failed} passed, {failed} failed");
        return failed == 0 ? 0 : 1;
    }

    private static string Home()
    {
        string path = Path.Combine(Path.GetTempPath(), "gateway-router-" + Guid.NewGuid().ToString("N"));
        Directory.CreateDirectory(path);
        return path;
    }

    private static RouterConfig Configuration(string home)
    {
        ImmutableDictionary<string, GroupDefinition> groups = new Dictionary<string, GroupDefinition>
        {
            ["admin"] = new(["group-admin"], "owner"),
            ["agents"] = new(["send-message", "declare-contactable"], "trusted"),
            ["high"] = new([], "owner"),
            ["external"] = new(["send-message", "declare-contactable"], "external")
        }.ToImmutableDictionary();
        IdentityDefinition Identity(params string[] memberships) =>
            new(Keys.EncodePublicKey(Keys.GenerateKeypair().Public.AsSpan()), memberships.ToImmutableArray());
        return new(new("127.0.0.1:0", home), groups, new Dictionary<string, IdentityDefinition>
        {
            ["admin"] = Identity("admin"), ["alice"] = Identity("high", "agents"),
            ["bob"] = Identity("agents"), ["outsider"] = Identity("external"), ["muted"] = Identity()
        }.ToImmutableDictionary());
    }

    private static void Check(bool condition, string message)
    {
        if (!condition) throw new Exception(message);
    }

    private static void Refuses(Action operation)
    {
        try { operation(); }
        catch (Exception error) when (error is ArgumentException or InvalidOperationException or InvalidDataException or JsonException or IOException or OverflowException)
        { return; }
        throw new Exception("Expected refusal, but operation succeeded.");
    }

    private static JsonElement Call(Router router, string caller, string name, object arguments, bool refused = false)
    {
        RpcMessage request = new(ID: JsonSerializer.SerializeToElement(42), Method: "tools/call",
            Params: JsonSerializer.SerializeToElement(new { name, arguments }));
        RpcMessage response = router.Request(router.Identity(caller), request)!;
        Check(response.Error is null && response.ID!.Value.GetInt32() == 42, "RPC correlation/error");
        JsonElement result = response.Result!.Value;
        string text = result.GetProperty("content")[0].GetProperty("text").GetString()!;
        Check(result.GetProperty("isError").GetBoolean() == refused, "Unexpected tool outcome: " + text);
        if (refused) return JsonSerializer.SerializeToElement(text);
        using JsonDocument document = JsonDocument.Parse(text);
        return document.RootElement.Clone();
    }

    private static async Task Relationships()
    {
        Router first = new(Configuration(Home()), Keys.GenerateKeypair());
        await using Router second = new(Configuration(Home()), Keys.GenerateKeypair());
        RouterIdentity alice = first.Identity("alice");
        RouterGroup agents = first.Group("agents");
        DeliveryQueue queue = alice.Queue();
        Check(alice.parent == first && agents.parent == first && queue.parent == alice, "Strong ownership");
        Check(agents.HasMember(alice, GroupLink.Base) && alice.Groups(GroupLink.Base).Contains(agents), "Initial reciprocal link");
        Refuses(() => first.RemoveGroup(agents));
        Refuses(() => first.RemoveIdentity(alice));
        Refuses(() => alice.ReplaceGroups(GroupLink.Contact, [agents, second.Group("agents")]));
        Check(alice.Groups(GroupLink.Contact).Length == 0 && !agents.HasMember(alice, GroupLink.Contact), "Failed replacement leaked a link");
        agents.AddMember(alice, GroupLink.Contact);
        Check(alice.Groups(GroupLink.Contact).SequenceEqual([agents]), "Group-side addition missed identity");
        alice.RemoveGroup(agents, GroupLink.Contact);
        Check(!agents.HasMember(alice, GroupLink.Contact), "Identity-side removal missed group");
        await first.DisposeAsync();
        Check(alice.parent is null && agents.parent is null && queue.parent is null && alice.queue is null, "Destruction retained parents");
        await first.DisposeAsync();
    }

    private static async Task Routing()
    {
        await using Router router = new(Configuration(Home()), Keys.GenerateKeypair());
        RouterIdentity alice = router.Identity("alice");
        DeliveryQueue queue = router.Identity("bob").Queue();
        Refuses(() => router.Route(alice, new("bob", "not contactable")));
        Call(router, "bob", "declare-contactable", new { groups = new[] { "agents", "external" } });
        Envelope first = router.Route(alice, new("bob", "sender data", "conversation", Protocol.ResponseAsync));
        Check(first.Source.ChannelID == "agents" && first.Source.Trust == "trusted", "Unaccepted owner group raised trust");
        Check(first.Content == "[GATEWAY source=agents user=alice trust=trusted]\n\nsender data", "Exact stamped content");
        Check(first.ConversationID == "conversation" && first.ResponseMode == "async", "Routing metadata changed");
        QueueState before = queue.Snapshot();
        Refuses(() => router.Route(router.Identity("muted"), new("bob", "no capability")));
        Refuses(() => router.Route(alice, new("bob", new string('x', Wire.MaxPlaintext))));
        Check(queue.Snapshot().Next == 1, "Refused delivery allocated sequence");
        Call(router, "bob", "declare-contactable", new { groups = new[] { "agents", "high" } });
        Envelope second = router.Route(alice, new("bob", "second"));
        Check(second.Source.ChannelID == "high" && second.Source.Trust == "owner", "Strongest accepted group not selected");
        Check(before.Messages!.Value.Length == 1 && queue.After(1) == second, "Snapshot mutated or sequence ordering lost");
        Refuses(() => queue.Acknowledge(3));
        queue.Acknowledge(1);
        queue.Acknowledge(0);
        Check(queue.Snapshot().Ack == 1 && queue.After(0) == second, "Receipt regressed");
        Call(router, "bob", "declare-contactable", new { groups = Array.Empty<string>() });
        Refuses(() => router.Route(alice, new("bob", "revoked")));
        Check(queue.Snapshot().Messages!.Value.Length == 1, "Policy change erased accepted history");
    }

    private static async Task Administration()
    {
        await using Router router = new(Configuration(Home()), Keys.GenerateKeypair());
        string publicKey = Keys.EncodePublicKey(Keys.GenerateKeypair().Public.AsSpan());
        Call(router, "alice", "register-child", new { agent_id = "child:one", public_key = publicKey, groups = new[] { "agents", "admin" } }, refused: true);
        Refuses(() => router.Identity("child:one"));
        Call(router, "alice", "register-child", new { agent_id = "child:one", public_key = publicKey, groups = new[] { "agents" } });
        Call(router, "alice", "register-child", new { agent_id = "child:one", public_key = publicKey, groups = new[] { "agents" } }, refused: true);
        Check(router.Identity("child:one").Groups(GroupLink.Base).Single() == router.Group("agents"), "Child base grant");
        var definition = new { name = "runtime:team", capabilities = new[] { "send-message" }, trust = "trusted" };
        Call(router, "bob", "put-group", definition, refused: true);
        Call(router, "admin", "put-group", definition);
        Call(router, "admin", "set-groups", new { agent_id = "muted", groups = new[] { "runtime:team" } });
        RouterIdentity muted = router.Identity("muted");
        RouterGroup team = router.Group("runtime:team");
        Call(router, "admin", "set-groups", new { agent_id = "muted", groups = new[] { "runtime:team", "agents" } }, refused: true);
        Check(muted.Groups(GroupLink.Grant).SequenceEqual([team]), "Refused replacement changed grants");
        Call(router, "bob", "declare-contactable", new { groups = new[] { "runtime:team" } });
        router.Route(muted, new("bob", "before revoke"));
        Call(router, "admin", "put-group", new { name = "runtime:team", capabilities = Array.Empty<string>(), trust = "external" });
        Refuses(() => router.Route(muted, new("bob", "capability revoked")));
        // A child may use a parent's runtime group as its own attenuated base
        // grant. Deleting that group must remove this link too, not just grants.
        Call(router, "muted", "register-child", new { agent_id = "child:runtime", public_key = publicKey, groups = new[] { "runtime:team" } });
        Call(router, "admin", "delete-group", new { name = "agents" }, refused: true);
        Call(router, "admin", "delete-group", new { name = "runtime:team" });
        Check(team.parent is null && muted.Groups(GroupLink.Grant).Length == 0 &&
            router.Identity("bob").Groups(GroupLink.Contact).Length == 0 &&
            router.Identity("child:runtime").Groups(GroupLink.Base).Length == 0, "Deleted group retained links");
        Check(router.Identity("bob").Queue().After(0)!.Source.Trust == "trusted", "Historical provenance changed");
    }

    private static async Task Restart()
    {
        string home = Home();
        RouterConfig config = Configuration(home);
        Keypair key = Keys.GenerateKeypair();
        Router first = new(config, key);
        Call(first, "bob", "declare-contactable", new { groups = new[] { "agents" } });
        first.Route(first.Identity("alice"), new("bob", "one"));
        first.Route(first.Identity("alice"), new("bob", "two"));
        first.Identity("bob").Queue().Acknowledge(1);
        Call(first, "admin", "put-group", new { name = "runtime:team", capabilities = new[] { "send-message" }, trust = "trusted" });
        Call(first, "admin", "set-groups", new { agent_id = "muted", groups = new[] { "runtime:team" } });
        Call(first, "alice", "register-child", new { agent_id = "child:one", public_key = Keys.EncodePublicKey(Keys.GenerateKeypair().Public.AsSpan()), groups = new[] { "agents" } });
        await first.DisposeAsync();
        Check(File.Exists(Path.Combine(home, "queues", "child:one.json")), "Wrong Go queue filename");
        RouterConfig invalidBootstrap = config with
        {
            Agents = config.Agents.SetItem("alice", config.Agents["alice"] with { Groups = ["runtime:team"] })
        };
        Refuses(() => _ = new Router(invalidBootstrap, key));
        Router second = new(config, key);
        DeliveryQueue restored = second.Identity("bob").Queue();
        Check(restored.Snapshot().Next == 2 && restored.Snapshot().Ack == 1 && restored.After(0)!.Content.EndsWith("two"), "Queue restart lost pending delivery/receipt");
        Check(second.Group("runtime:team").HasMember(second.Identity("muted"), GroupLink.Grant), "Runtime links not restored");
        Check(second.Group("agents").HasMember(second.Identity("child:one"), GroupLink.Base), "Child not restored");
        restored.Acknowledge(2);
        await second.DisposeAsync();
        await using Router third = new(config, key);
        Check(third.Identity("bob").Queue().After(0) is null, "Drained queue replayed");
        Check(third.Route(third.Identity("alice"), new("bob", "three")).Seq == 3, "Drained sequence reused");
        if (!OperatingSystem.IsWindows())
        {
            Check(File.GetUnixFileMode(Path.Combine(home, "state.json")) == (UnixFileMode.UserRead | UnixFileMode.UserWrite), "State file permissions");
            Check(File.GetUnixFileMode(Path.Combine(home, "queues")) == (UnixFileMode.UserRead | UnixFileMode.UserWrite | UnixFileMode.UserExecute), "Queue directory permissions");
        }
    }

    private static async Task DiskValidation()
    {
        try
        {
            _ = Router.ReadConfig("absent-gateway-" + Guid.NewGuid().ToString("N") + ".json");
            throw new Exception("Missing configuration accepted.");
        }
        catch (FileNotFoundException) { }
        string obstructed = Home();
        File.WriteAllText(Path.Combine(obstructed, "queues"), "not a directory");
        Refuses(() => _ = new Router(Configuration(obstructed), Keys.GenerateKeypair()));
        Check(File.ReadAllText(Path.Combine(obstructed, "queues")) == "not a directory", "Startup changed obstruction");
        string dangling = Home();
        File.CreateSymbolicLink(Path.Combine(dangling, "state.json"), Path.Combine(dangling, "missing.json"));
        Refuses(() => _ = new Router(Configuration(dangling), Keys.GenerateKeypair()));
        string home = Home();
        RouterConfig config = Configuration(home);
        Keypair key = Keys.GenerateKeypair();
        Directory.CreateDirectory(Path.Combine(home, "queues"));
        // Independent JSON reproduces Go omitempty/null forms, not a round trip
        // through the implementation's records or serializer.
        string valid = """
            {"next":2,"ack":1,"messages":[{"id":"bob:2","seq":2,"to_agent":"bob","content":"historical",
            "timestamp":"2026-01-02T03:04:05Z","source":{"type":"agent","channel_id":"removed","user_id":"old-sender","trust":"trusted"}}]}
            """;
        string path = Path.Combine(home, "queues", "bob.json");
        File.WriteAllText(path, valid);
        await using (Router router = new(config, key))
        {
            Envelope envelope = router.Identity("bob").Queue().After(0)!;
            Check(envelope.ConversationID == "" && envelope.ResponseMode == "" && envelope.Source.DisplayName == "", "Go omissions rejected");
            router.Identity("bob").Queue().Acknowledge(2);
        }
        File.WriteAllText(path, "{\"next\":2,\"ack\":2,\"messages\":null}");
        await using (Router router = new(config, key)) Check(router.Identity("bob").Queue().After(0) is null, "Go drained null queue rejected");
        string[] corrupt =
        [
            "{\"next\":2,\"ack\":3,\"messages\":[]}",
            "{\"next\":2,\"ack\":1,\"messages\":[]}",
            "{\"next\":2,\"next\":2,\"ack\":2,\"messages\":[]}",
            "{\"next\":2,\"ack\":2}", "null", "{",
            valid.Replace("bob:2", "wrong"), valid.Replace("\"seq\":2", "\"seq\":1")
        ];
        foreach (string bad in corrupt)
        {
            File.WriteAllText(path, bad);
            Refuses(() => _ = new Router(config, key));
            Check(File.ReadAllText(path) == bad, "Failed restore overwrote evidence");
        }
        File.WriteAllText(path, "{\"next\":18446744073709551615,\"ack\":18446744073709551615,\"messages\":null}");
        await using (Router full = new(config, key))
        {
            Call(full, "bob", "declare-contactable", new { groups = new[] { "agents" } });
            Refuses(() => full.Route(full.Identity("alice"), new("bob", "overflow")));
            Check(full.Identity("bob").Queue().Snapshot().Next == ulong.MaxValue, "Overflow mutated sequence");
        }
        string statePath = Path.Combine(home, "state.json");
        string state = File.ReadAllText(statePath);
        File.WriteAllText(statePath, state.Replace("\"contactable\":", "\"unrecognized\":"));
        Refuses(() => _ = new Router(config, key));
    }

    private static async Task SaveFailure()
    {
        string home = Home();
        Router router = new(Configuration(home), Keys.GenerateKeypair());
        string queues = Path.Combine(home, "queues");
        File.WriteAllText(queues, "owned obstruction");
        Refuses(router.Save);
        Check(router.Identity("alice").parent == router, "Save failure detached live identity");
        File.Delete(queues);
        router.Save();
        RouterIdentity alice = router.Identity("alice");
        DeliveryQueue queue = alice.Queue();
        // Only remove our own empty test snapshot files; never shared resources.
        Directory.Delete(queues, recursive: true);
        File.WriteAllText(queues, "owned obstruction");
        try { await router.DisposeAsync(); throw new Exception("Expected final save failure."); }
        catch (IOException) { }
        Check(alice.parent is null && queue.parent is null, "Save failure blocked destruction");
        Check(File.ReadAllText(queues) == "owned obstruction", "Save overwrote obstruction");
        // A completed failed shutdown remains failed for subsequent observers;
        // joining twice must not turn an unsuccessful final save into success.
        Task first = router.DisposeAsync().AsTask();
        Task second = router.DisposeAsync().AsTask();
        Check(ReferenceEquals(first, second), "Failed shutdown result was not shared");
        try { await first; throw new Exception("Repeated close hid its save failure."); }
        catch (IOException) { }
    }

    private static async Task ConcurrentAdmission()
    {
        string home = Home();
        RouterConfig config = Configuration(home);
        Keypair key = Keys.GenerateKeypair();
        Router router = new(config, key);
        Call(router, "bob", "declare-contactable", new { groups = new[] { "agents" } });
        Task[] work = Enumerable.Range(0, 96).Select(index => Task.Run(() =>
        {
            router.Route(router.Identity("alice"), new("bob", index.ToString()));
            if (index % 16 == 0) router.Save();
        })).ToArray();
        await Task.WhenAll(work);
        QueueState snapshot = router.Identity("bob").Queue().Snapshot();
        Check(snapshot.Next == 96 && snapshot.Messages!.Value.Select(message => message.Seq)
            .SequenceEqual(Enumerable.Range(1, 96).Select(value => (ulong)value)), "Concurrent admission lost/reordered sequences");
        await router.DisposeAsync();
        await using Router restored = new(config, key);
        Check(restored.Identity("bob").Queue().Snapshot().Messages!.Value.Length == 96, "Concurrent save lost final deliveries");
    }

    private static async Task GoSnapshot(string fixture)
    {
        RouterConfig original = Router.ReadConfig(Path.Combine(fixture, "gateway.json"));
        string home = Home();
        Directory.CreateDirectory(Path.Combine(home, "queues"));
        File.Copy(Path.Combine(original.Gateway.DataDirectory, "state.json"), Path.Combine(home, "state.json"));
        foreach (string id in original.Agents.Keys)
        {
            // Validate copied filenames through the already public config read;
            // the retained test fixture contains only generated public identities.
            Check(id is "alice" or "bob", "Unexpected reference fixture identity");
            File.Copy(Path.Combine(original.Gateway.DataDirectory, "queues", id + ".json"), Path.Combine(home, "queues", id + ".json"));
        }
        RouterConfig config = original with { Gateway = original.Gateway with { DataDirectory = home } };
        await using Router router = new(config, Keys.GenerateKeypair());
        QueueState bob = router.Identity("bob").Queue().Snapshot();
        Check(bob.Next == 3 && bob.Ack == 3 && bob.Messages!.Value.IsEmpty, "Real Go receipt did not restore");
        Check(router.Group("agents").HasMember(router.Identity("bob"), GroupLink.Contact), "Real Go contacts did not restore");
        Check(router.Route(router.Identity("alice"), new("bob", "after-Go-restore")).Seq == 4, "Real Go allocation frontier lost");
    }

    private static async Task Rpc()
    {
        await using Router router = new(Configuration(Home()), Keys.GenerateKeypair());
        RouterIdentity alice = router.Identity("alice");
        Check(router.Request(alice, new(Method: "notifications/initialized")) is null, "Notification got response");
        Check(router.Request(alice, new(Jsonrpc: "1.0", Method: "initialize"))!.Error!.Code == -32600, "Version check");
        Check(router.Request(alice, new(Method: "ping"))!.Error!.Code == -32601, "Invented method accepted");
        Check(router.Request(alice, new(Method: "tools/call"))!.Error!.Code == -32602, "Malformed call");
        JsonElement tools = router.Request(alice, new(Method: "tools/list"))!.Result!.Value.GetProperty("tools");
        Check(tools.GetArrayLength() == 7 && tools.EnumerateArray().All(tool => !tool.GetProperty("inputSchema").GetProperty("additionalProperties").GetBoolean()), "Published tool contract");
        Call(router, "alice", "send-message", new { to_agent = "bob", content = "text", source = "admin" }, refused: true);
        Call(router, "bob", "declare-contactable", new { groups = new[] { "agents", "missing" } }, refused: true);
        Check(router.Identity("bob").Groups(GroupLink.Contact).Length == 0, "Failed contact declaration mutated");
        using JsonDocument duplicate = JsonDocument.Parse("{\"name\":\"list-groups\",\"arguments\":{\"groups\":[],\"groups\":[]}}");
        RpcMessage reply = router.Request(alice, new(Method: "tools/call", Params: duplicate.RootElement))!;
        Check(reply.Error?.Code == -32602, "Duplicate nested property not rejected");
    }
}
