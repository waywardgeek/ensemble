using System.Collections.Immutable;
using System.Net;
using System.Net.Sockets;
using System.Text;
using System.Text.Json;
using System.Text.Json.Nodes;
using Ensemble;
using Application = Ensemble.Ensemble;

// A separate process exercises the same protected credential transaction.
if (args is ["--auth-store-child", var authDirectory, var authLabel])
{
    return await AuthStoreTests.Child(authDirectory, authLabel);
}
if (args is ["--auth-renew-child", var renewDirectory, var renewIssuer])
{
    return await AuthRenewalTests.Child(renewDirectory, renewIssuer);
}

// A small dependency-free executable suite. Assertions inspect real model state
// and real wire bodies; the only fake is an actual loopback HTTP provider.
var failures = 0;
var passed = 0;
var temp = Path.Combine(Environment.GetEnvironmentVariable("CH02_TEST_DIR") ?? Path.GetTempPath(), "ch02-" + Guid.NewGuid().ToString("N"));
Directory.CreateDirectory(temp);
void CheckCase(string name, Action action)
{
    try
    {
        action();
        passed++;
        Console.WriteLine("PASS " + name);
    }
    catch (Exception ex) { failures++; Console.WriteLine("FAIL " + name + ": " + ex); }
}
async Task AsyncCase(string name, Func<Task> action)
{
    try
    {
        await action();
        passed++;
        Console.WriteLine("PASS " + name);
    }
    catch (Exception ex) { failures++; Console.WriteLine("FAIL " + name + ": " + ex); }
}
static void Check(bool yes, string why)
{
    if (!yes)
    {
        throw new Exception(why);
    }
}
static void Refuses(Action operation)
{
    try
    {
        operation();
    }
    catch (Exception e) when (e is ArgumentException or FormatException or InvalidOperationException or JsonException or NotSupportedException or KeyNotFoundException) { return; }
    throw new Exception("Expected refusal");
}
static JsonElement J(string text) { using var d = JsonDocument.Parse(text); return d.RootElement.Clone(); }
static Provenance Origin(Vendor vendor = Vendor.Anthropic, string model = "model-a") => new Configuration(vendor, model, "http://127.0.0.1").Target();
static ToolCallPart Call(string id = "a", Vendor vendor = Vendor.Anthropic) => new(id, Origin(vendor), "read_file", J("{\"n\":9007199254740993,\"path\":\"x\"}"));
static Agent NewAgent(Application app, Vendor vendor = Vendor.Anthropic, string url = "http://127.0.0.1", string resolved = "model-a") => new(app, new Configuration(vendor, "model-a", url, "dummy-local-only", resolved));
static string Wire(Vendor vendor, Context context, string model = "model-a")
{
    var app = new Application(TextWriter.Null);
    try
    {
        return NewAgent(app, vendor, resolved: model).Engine().Render(context).ToJsonString();
    }
    finally { app.Destroy(); }
}
try
{
    if (args is ["--cache-lens-tests"])
    {
        CheckCase("cache lens rooted captures", () => CacheLensTests.Confined(temp));
        CheckCase("cache lens byte views", CacheLensTests.ByteViews);
        await AsyncCase("cache lens transport", () => CacheLensTests.Transport(temp));
        await AsyncCase("cache lens failure", () => CacheLensTests.Failure(temp));
        Console.WriteLine($"{passed} passed; {failures} failed");
        return failures == 0 ? 0 : 1;
    }

    if (args is ["--child-creation-tests"] || args.Length == 0)
    {
        await AsyncCase("Chapter 20 child creation and template inheritance", () => ChildCreationTests.StorageAndTemplates(temp));
        await AsyncCase("Chapter 20 child creation refusal and recovery", () => ChildCreationTests.Refusals(temp));
        await AsyncCase("Chapter 20 inherited child MCP clients", () => McpTests.ChildSkills(temp));
        if (args.Length != 0)
        {
            Console.WriteLine($"{passed} passed; {failures} failed");
            return failures == 0 ? 0 : 1;
        }
    }
    if (args is ["--child-completion-tests"] || args.Length == 0)
    {
        await AsyncCase("Chapter 20 bounded submission corrections", () => ActorTests.ChildCompletionExhaustion(temp));
        await AsyncCase("Chapter 20 submission reset and draining", () => ActorTests.ChildCompletionResetAndDrain(temp));
        await AsyncCase("Chapter 20 child failures and standalone chat", () => ActorTests.ChildCompletionErrorsAndRoot(temp));
        if (args.Length != 0)
        {
            Console.WriteLine($"{passed} passed; {failures} failed");
            return failures == 0 ? 0 : 1;
        }
    }
    if (args is ["--question-tests"] || args.Length == 0)
    {
        await AsyncCase("Chapter 20 question routing and reply IDs", () => QuestionTests.Routing(temp));
        await AsyncCase("Chapter 20 question cancellation and ownership", () => QuestionTests.CancellationAndOwners(temp));
        await AsyncCase("Chapter 20 concurrent question shutdown", () => QuestionTests.ConcurrentShutdown(temp));
        if (args.Length != 0)
        {
            Console.WriteLine($"{passed} passed; {failures} failed");
            return failures == 0 ? 0 : 1;
        }
    }
    if (args is ["--child-submission-tests"] || args.Length == 0)
    {
        CheckCase("Chapter 20 durable child submission", () => ChildSubmissionTests.Durable(temp));
        CheckCase("Chapter 20 submission publication failure", () => ChildSubmissionTests.FailedPublication(temp));
        await AsyncCase("Chapter 20 admitted work generations", () => ActorTests.ChildSubmissionGenerations(temp));
        await AsyncCase("Chapter 20 publication and policy concurrency", () => ChildSubmissionTests.ConcurrentPolicy(temp));
        await AsyncCase("Chapter 20 reliable submission observers", () => SubmissionWaitTests.Observers(temp));
        await AsyncCase("Chapter 20 bottom-up submission", () => SubmissionWaitTests.BottomUp(temp));
        if (args.Length != 0)
        {
            Console.WriteLine($"{passed} passed; {failures} failed");
            return failures == 0 ? 0 : 1;
        }
    }
    if (args is ["--submission-contract-tests"] || args.Length == 0)
    {
        CheckCase("Chapter 20 structured submission contract", SubmissionContractTests.Structured);
        CheckCase("Chapter 20 schema admission", SubmissionContractTests.SchemaAdmission);
        CheckCase("Chapter 20 exact submission numbers", SubmissionContractTests.ExactNumbers);
        CheckCase("Chapter 20 nested submission constraints", SubmissionContractTests.NestedConstraints);
        CheckCase("Chapter 20 submission limits", SubmissionContractTests.Limits);
        if (args.Length != 0)
        {
            Console.WriteLine($"{passed} passed; {failures} failed");
            return failures == 0 ? 0 : 1;
        }
    }
    if (args is ["--child-ownership-tests"] || args.Length == 0)
    {
        await AsyncCase("Chapter 20 child ownership and refusal", () => ChildOwnershipTests.Relationships(temp));
        await AsyncCase("Chapter 20 joined descendant shutdown", ActorTests.ChildTreeShutdown);
        CheckCase("Chapter 20 child cleanup failure", () => ChildOwnershipTests.CleanupFailure(temp));
        if (args.Length != 0)
        {
            Console.WriteLine($"{passed} passed; {failures} failed");
            return failures == 0 ? 0 : 1;
        }
    }
    if (args is ["--child-policy-tests"] || args.Length == 0)
    {
        await AsyncCase("Chapter 19 child policy narrowing", ChildPolicyTests.Narrowing);
        if (args.Length != 0)
        {
            Console.WriteLine($"{passed} passed; {failures} failed");
            return failures == 0 ? 0 : 1;
        }
    }
    if (args is ["--diagnostic-tests"] || args.Length == 0)
    {
        await AsyncCase("Chapter 19 readable diagnostic redaction", DiagnosticTests.KnownKeys);
        await AsyncCase("Chapter 19 plan bearer trace redaction", AuthRenewalTests.BearerTrace);
        if (args.Length != 0)
        {
            await AsyncCase("Streaming trace compatibility", ActorTests.StreamingDelivery);
            Console.WriteLine($"{passed} passed; {failures} failed");
            return failures == 0 ? 0 : 1;
        }
    }
    if (args is ["--process-files-tests"] || args.Length == 0)
    {
        await AsyncCase("Chapter 19 confined process boundary", SandboxProcessTests.Boundary);
        await AsyncCase("Chapter 19 confined process useful work", SandboxProcessTests.UsefulWork);
        if (args.Length != 0)
        {
            await AsyncCase("PTY lifecycle compatibility", () => JobTests.Run(temp));
            Console.WriteLine($"{passed} passed; {failures} failed");
            return failures == 0 ? 0 : 1;
        }
    }
    if (args is ["--persistence-files-tests"] || args.Length == 0)
    {
        CheckCase("Chapter 19 rooted persistence replay", SandboxPersistenceTests.Replay);
        CheckCase("Chapter 19 rooted persistence refusal and poisoning", SandboxPersistenceTests.Refusals);
        if (args.Length != 0)
        {
            CheckCase("Persistence replay compatibility", PersistenceTests.Replay);
            CheckCase("Persistence recovery compatibility", PersistenceTests.Recovery);
            await AsyncCase("Persistence lifecycle compatibility", PersistenceTests.Runtime);
            CheckCase("Journal recovery compatibility", JournalTests.Recovery);
            CheckCase("Journal torn-tail compatibility", JournalTests.TornAndInvalid);
            CheckCase("Journal failed append compatibility", JournalTests.FailedAppend);
            await AsyncCase("Journal serialization compatibility", JournalTests.WriterSerialization);
            Console.WriteLine($"{passed} passed; {failures} failed");
            return failures == 0 ? 0 : 1;
        }
    }
    if (args is ["--job-files-tests"])
    {
        await AsyncCase("Chapter 19 rooted Job output", ActorTests.RootedJobFiles);
        await AsyncCase("Automatic snapshots and replay", ActorTests.AutomaticSnapshots);
        await AsyncCase("Automatic completion and server errors", ActorTests.AutomaticPauseAndError);
        Console.WriteLine($"{passed} passed; {failures} failed");
        return failures == 0 ? 0 : 1;
    }
    if (args is ["--media-tests"])
    {
        await AsyncCase("Media projection compatibility", ActorTests.MediaProjection);
        await AsyncCase("Chapter 19 rooted media projection", ActorTests.RootedMedia);
        Console.WriteLine($"{passed} passed; {failures} failed");
        return failures == 0 ? 0 : 1;
    }
    if (args is ["--tool-tests"])
    {
        await AsyncCase("Chapter 3 tools and loop", () => ToolTests.Run(temp));
        Console.WriteLine($"{passed} passed; {failures} failed");
        return failures == 0 ? 0 : 1;
    }
    if (args is ["--model-tests"] || args.Length == 0)
    {
        await AsyncCase("Chapter 18 actor-owned model selection", () => MeterTests.ModelSelection(temp));
        await AsyncCase("Chapter 18 direct-send selection guard and capabilities", MeterTests.DirectSendSelection);
        if (args.Length != 0)
        {
            Console.WriteLine($"{passed} passed; {failures} failed");
            return failures == 0 ? 0 : 1;
        }
    }
    if (args is ["--status-tests"] || args.Length == 0)
    {
        await AsyncCase("Chapter 18 status uses the owning Engine meter", () => MeterTests.Status(temp));
        await AsyncCase("Chapter 18 status observes its requesting response", () => MeterTests.StatusTurn(temp));
        if (args.Length != 0)
        {
            Console.WriteLine($"{passed} passed; {failures} failed");
            return failures == 0 ? 0 : 1;
        }
    }
    if (args is ["--nonregular-tests"])
    {
        await AsyncCase("Sandbox nonregular files", SandboxTests.NonregularFile);
        return failures == 0 ? 0 : 1;
    }

    if (args is ["--sandbox-tests"] || args.Length == 0)
    {
        CheckCase("Chapter 19 rooted filesystem and ownership", SandboxTests.Files);
        await AsyncCase("Chapter 19 rooted symlink-swap race", SandboxTests.SymlinkRace);
        await AsyncCase("Chapter 19 rooted file-tool dispatch", SandboxTests.ToolFiles);
        await AsyncCase("Sandbox nonregular files", SandboxTests.NonregularFile);
        if (args.Length != 0)
        {
            Console.WriteLine($"{passed} passed; {failures} failed");
            return failures == 0 ? 0 : 1;
        }
    }
    if (args is ["--safety-tests"] || args.Length == 0)
    {
        await AsyncCase("Chapter 19 safe-mode admission and projection", SafetyTests.SafeMode);
        await AsyncCase("Chapter 19 web policy and dependency admission", McpTests.WebPolicy);
        if (args.Length != 0)
        {
            Console.WriteLine($"{passed} passed; {failures} failed");
            return failures == 0 ? 0 : 1;
        }
    }
    if (args is ["--http-mcp-tests"] || args.Length == 0)
    {
        CheckCase("Chapter 18 HTTP MCP specification validation", McpTests.HttpSpecifications);
        await AsyncCase("Chapter 18 HTTP MCP real POST/SSE/session/cancellation", McpTests.HttpTransport);
        await AsyncCase("Chapter 18 HTTP MCP skill activation and unload", McpTests.HttpSkills);
        if (args.Length != 0)
        {
            Console.WriteLine($"{passed} passed; {failures} failed");
            return failures == 0 ? 0 : 1;
        }
    }
    if (args is ["--responses-tests"] || args.Length == 0)
    {
        CheckCase("Responses selected-model reasoning policy", ResponsesTests.ReasoningPolicy);
        await AsyncCase("Responses selected-model explanation dispatch", ResponsesTests.ReasoningDispatch);
        CheckCase("Responses projection and explicit route", ResponsesTests.Projection);
        CheckCase("Responses plain decoding and provenance", ResponsesTests.Plain);
        await AsyncCase("Responses SSE terminal gate and summaries", ResponsesTests.Streaming);
        await AsyncCase("Responses real HTTP lens and meter", () => ResponsesTests.Transport(temp));
        await AsyncCase("Responses explicit plan authority", () => ResponsesTests.Plan(temp));
        await AsyncCase("Responses real Agent dispatch", ResponsesTests.Dispatch);
        if (args.Length != 0)
        {
            Console.WriteLine($"{passed} passed; {failures} failed");
            return failures == 0 ? 0 : 1;
        }
    }
    if (args is ["--auth-inference-tests"])
    {
        await AsyncCase("auth inference 401 bounded same-source recovery", () => AuthRenewalTests.RejectedBearerRecovery(temp));
        await AsyncCase("auth delayed 401 reuses newest durable grant", () => AuthRenewalTests.RejectedOlderGeneration(temp));
        Console.WriteLine($"{passed} passed; {failures} failed");
        return failures == 0 ? 0 : 1;
    }
    CheckCase("auth signed identity valid login and refresh", AuthIdentityTests.ValidIdentities);
    CheckCase("auth signed identity cryptographic refusals", AuthIdentityTests.SignatureRefusals);
    CheckCase("auth signed identity claim and ambiguity refusals", AuthIdentityTests.ClaimRefusals);
    if (args is ["--auth-identity-tests"])
    {
        Console.WriteLine($"{passed} passed; {failures} failed");
        return failures == 0 ? 0 : 1;
    }
    await AsyncCase("auth store ownership, account separation and local logout", () => AuthStoreTests.OwnershipAndAccounts(temp));
    await AsyncCase("auth store strict decoding, permissions and safe refusal", () => AuthStoreTests.RefusalAndPermissions(temp));
    await AsyncCase("auth store real cross-process locking and cancellation", () => AuthStoreTests.Contention(temp));
    if (args is ["--auth-store-tests"])
    {
        Console.WriteLine($"{passed} passed; {failures} failed");
        return failures == 0 ? 0 : 1;
    }
    await AsyncCase("auth inference 401 bounded same-source recovery", () => AuthRenewalTests.RejectedBearerRecovery(temp));
    await AsyncCase("auth delayed 401 reuses newest durable grant", () => AuthRenewalTests.RejectedOlderGeneration(temp));
    await AsyncCase("auth rotating refresh across real processes", () => AuthRenewalTests.SerializedRotation(temp));
    await AsyncCase("auth refresh failure and plan-permission boundaries", () => AuthRenewalTests.FailuresAndPermission(temp));
    await AsyncCase("auth refresh durable-before-return failure", () => AuthRenewalTests.PublicationFailure(temp));
    await AsyncCase("auth local-first revocation and joined destruction", () => AuthRenewalTests.LogoutAndDestruction(temp));
    await AsyncCase("auth login real issuer, PKCE callback and registration reuse", () => AuthLoginTests.LoginAndReuse(temp));
    await AsyncCase("auth login callback ambiguity and foreign-state refusals", () => AuthLoginTests.CallbackRefusals(temp));
    await AsyncCase("auth login pending registration, retry and initial identity", () => AuthLoginTests.FailureAndPendingRetry(temp));
    await AsyncCase("auth login stale registration and joined destruction", () => AuthLoginTests.StaleRegistrationAndDestruction(temp));
    await AsyncCase("auth catalog ordered visible entries and fresh account tokens", () => AuthModelsTests.Catalog(Path.Combine(temp, "catalog")));
    await AsyncCase("auth catalog malformed, unsafe and denied response refusals", () => AuthModelsTests.Refusals(Path.Combine(temp, "catalog-refusals")));
    await AsyncCase("auth catalog cancellation and joined destruction", () => AuthModelsTests.Lifetime(Path.Combine(temp, "catalog-lifetime")));
    await AsyncCase("auth catalog real signed refresh before discovery", () => AuthRenewalTests.CatalogRenewal(temp));
    if (args is ["--auth-tests"])
    {
        Console.WriteLine($"{passed} passed; {failures} failed");
        return failures == 0 ? 0 : 1;
    }
    CheckCase("process meter disjoint arithmetic and unknowns", MeterTests.Arithmetic);
    await AsyncCase("process meter exact model-tier pricing", MeterTests.Pricing);
    await AsyncCase("process meter actual completion and restart isolation", () => MeterTests.Completion(temp));
    await AsyncCase("process meter final streaming usage", MeterTests.Streaming);
    CheckCase("cache markers final stable boundaries", CacheMarkerTests.Boundaries);
    await AsyncCase("cache markers deferred skill catalog", CacheMarkerTests.Deferred);
    await AsyncCase("cache markers actual transport and options", CacheMarkerTests.TransportAndOptions);
    CheckCase("cache lens rooted captures", () => CacheLensTests.Confined(temp));
    CheckCase("cache lens byte-preserving prompt views", CacheLensTests.ByteViews);
    await AsyncCase("cache lens actual-send correlation and cleanup", () => CacheLensTests.Transport(temp));
    await AsyncCase("cache lens capture failure isolation", () => CacheLensTests.Failure(temp));
    CheckCase("Chapter 14 once-only result selection boundaries", ResultStubbingTests.Selection);
    await AsyncCase("Chapter 14 opt-in keep/stub actor and snapshot replay", ActorTests.OnceOnlyResultStubbing);
    await AsyncCase("Chapter 14 interrupted keep cannot veto recorded cuts", ActorTests.InterruptedKeep);
    CheckCase("Chapter 14 full event archive and request-prefix replay", ArchiveTests.Replay);
    CheckCase("Chapter 14 archive refusal preserves recovery and diagnostic files", ArchiveTests.Refusal);
    CheckCase("Chapter 14 handoff clears private replay data but preserves visible identities", CompactionTests.HandoffPrivacy);
    CheckCase("Chapter 14 safe tool ladder selection and cached bytes", CompactionTests.Selection);
    await AsyncCase("Chapter 14 actor compaction after successful presentation", ActorTests.ToolCompaction);
    await AsyncCase("Chapter 11 MCP handshake and concurrent RPC correlation", McpTests.Correlation);
    await AsyncCase("Chapter 11 MCP reverse calls and peer closure", () => McpTests.ReverseAndClosure(temp));
    await AsyncCase("Chapter 11 MCP failed discovery and joined destruction", McpTests.RefusalAndShutdown);
    await AsyncCase("Chapter 11 remote Jobs preserve ordered content and cancellation", McpTests.RemoteJobs);
    await AsyncCase("Chapter 11 remote catalog registration is transactional", McpTests.RemoteRegistration);
    await AsyncCase("Chapter 11 MCP server frontmatter validation", McpTests.ServerSpecifications);
    await AsyncCase("Chapter 11 owned stdio process transport", McpTests.ProcessBoundary);
    await AsyncCase("Chapter 11 fragmented WebSocket RPC and closure", McpTests.WebSocketBoundary);
    await AsyncCase("Chapter 11 browser registration and dynamic skill binding", McpTests.BrowserSkillBinding);
    await AsyncCase("Chapter 11 browser primary startup", McpTests.BrowserPrimaryStartup);
    await AsyncCase("Chapter 11 browser discovery disconnect rollback", McpTests.BrowserDiscoveryDisconnect);
    await AsyncCase("Chapter 11 incoming MCP service and real file dispatch", McpTests.IncomingService);
    await AsyncCase("Chapter 11 service malformed input and joined termination", McpTests.ServiceTermination);
    await AsyncCase("Chapter 11 actor-owned skill process activation", McpTests.SkillActivation);
    await AsyncCase("Chapter 11 dependency activation and independent ownership", McpTests.SkillDependencies);
    await AsyncCase("Chapter 11 skill dependency and catalog rollback", McpTests.SkillRollback);
    await AsyncCase("Chapter 11 shutdown and interrupt cancel skill preparation", McpTests.SkillShutdown);
    await AsyncCase("Chapter 11 primary server startup is one transactional activation", McpTests.PrimaryStartup);
    await AsyncCase("Chapter 12 ordered startup selection and instruction placement", McpTests.StartupSelection);
    await AsyncCase("Chapter 12 optional server startup is transactional", McpTests.OptionalServerStartup);
    await AsyncCase("Chapter 12 failed optional startup preserves prior save", McpTests.OptionalStartupFailure);
    await AsyncCase("Chapter 11 failed primary startup preserves saved conversation", McpTests.PrimaryFailure);
    await AsyncCase("Chapter 11 shutdown joins primary startup", McpTests.PrimaryShutdown);
    await AsyncCase("Chapter 11 automatic snapshots and replay", ActorTests.AutomaticSnapshots);
    await AsyncCase("Chapter 11 automatic collection interrupt preserves jobs", ActorTests.AutomaticInterrupt);
    await AsyncCase("Chapter 11 automatic tools stay outside discovery", ActorTests.AutomaticDiscovery);
    await AsyncCase("Chapter 11 automatic completion respects pause and server errors", ActorTests.AutomaticPauseAndError);
    await AsyncCase("Chapter 11 failed HTTP consumes automatic ephemera", ActorTests.AutomaticFailedRequest);
    CheckCase("portable snapshot, tail, log-only, retention and archive", PersistenceTests.Replay);
CheckCase("validated backup recovery and failure nonmutation", PersistenceTests.Recovery);
await AsyncCase("automatic lifecycle and current runtime reconfiguration", PersistenceTests.Runtime);
await AsyncCase("Chapter 14 pending skill authority and frozen projections", SkillAuthorityTests.PendingAuthority);
CheckCase("Chapter 14 total reduction and strict unknown-event refusal", ReductionTests.TotalReplay);
await AsyncCase("Chapter 14 handoff waits for batch and replaces prior note", ActorTests.HandoffBatch);
CheckCase("Chapter 14 handoff snapshot-only replay", HandoffTests.Replay);
await AsyncCase("Chapter 14 interrupted handoff pairs tail and preserves Job", ActorTests.InterruptedHandoff);
await AsyncCase("Chapter 14 manual clear waits for outcomes and preserves context/Jobs", ActorTests.ManualToolClear);
CheckCase("Chapter 14 journal primary and backup recovery", JournalTests.Recovery);
CheckCase("Chapter 14 first-session journal and torn-tail recovery", JournalTests.TornAndInvalid);
CheckCase("Chapter 14 failed append never publishes", JournalTests.FailedAppend);
await AsyncCase("Chapter 14 serialized writers and unblocked committed readers", JournalTests.WriterSerialization);
await AsyncCase("Chapter 3 tools and loop", () => ToolTests.Run(temp));
    await AsyncCase("Chapter 4 jobs and PTY", () => JobTests.Run(temp));
    await AsyncCase("Chapter 13 atomic custom completion and shutdown", ChannelToolTests.Run);
    await AsyncCase("Chapter 5 transport isolation", ActorTests.TransportIsolation);
    await AsyncCase("Chapter 5 mailbox controls", ActorTests.MailboxControls);
    await AsyncCase("Chapter 5 busy hint", ActorTests.BusyHint);
    await AsyncCase("Chapter 5 immediate interrupt", ActorTests.ImmediateInterrupt);
    await AsyncCase("Chapter 5 boundary interrupt", ActorTests.BoundaryInterrupt);
    await AsyncCase("Chapter 5 shutdown joins", ActorTests.ShutdownJoins);
    await AsyncCase("Chapter 5 HTTP interrupt boundaries", ActorTests.HttpInterrupts);
    await AsyncCase("Chapter 5 direct tool hint", ActorTests.DirectHint);
    await AsyncCase("Chapter 5 input admission outcomes", ActorTests.InputAdmission);
    await AsyncCase("Chapter 5 reliable flush", ActorTests.FlushTurns);
    await AsyncCase("Chapter 5 unlimited tool batches", ActorTests.UnlimitedBatches);
    await AsyncCase("Chapter 5 local media projection and refusal", ActorTests.MediaProjection);
    await AsyncCase("Chapter 19 rooted media projection", ActorTests.RootedMedia);
    await AsyncCase("Chapter 19 rooted Job output", ActorTests.RootedJobFiles);
    await AsyncCase("Chapter 5 named Agents and caller roles", ActorTests.NamedAgents);
    await AsyncCase("Chapter 8 captured runtime tool budget", ActorTests.RuntimeBudget);
    await AsyncCase("Chapter 9 transactional skills and capability gating", SkillTests.Capabilities);
    await AsyncCase("Chapter 9 parser, ownership and provider instruction placement", SkillTests.Projection);
    await AsyncCase("Chapter 9 real discovery exchanges and deferred references", ActorTests.SkillExchange);
    await AsyncCase("Chapter 7 atomic snapshots and durable display identity", GuiCoreTests.SnapshotsAndIdentity);
    await AsyncCase("Chapter 7 truthful tool and pause lifecycle", GuiCoreTests.LifecycleFacts);
    await AsyncCase("Chapter 6 HTTP parser and stream identity", StreamingTests.Parser);
    await AsyncCase("Chapter 6 incremental actor publication", ActorTests.StreamingDelivery);
    await AsyncCase("Chapter 6 streaming tool policy and interrupt boundaries", ActorTests.StreamingToolBoundary);
    CheckCase("construction and complete idempotent destruction", () =>
    {
        var app = new Application(TextWriter.Null);
        var a = NewAgent(app);
        var b = NewAgent(app);
        var history = a.History();
        var engine = a.Engine();
        Check(app.Agents().Length == 2 && ReferenceEquals(a.Parent(), app), "both root links");
        Check(ReferenceEquals(history.Parent(), a) && ReferenceEquals(engine.Parent(), a), "both child links");
        engine.Destroy();
        Refuses(() => a.Engine());
        Refuses(() => engine.Parent());
        _ = new Engine(a);
        a.Destroy();
        a.Destroy();
        Check(app.Agents().Length == 1 && ReferenceEquals(app.Agents()[0], b), "sibling survives removal");
        Refuses(() => history.Parent());
        Refuses(() => a.Parent());
        app.Destroy();
        app.Destroy();
        Check(app.Agents().Length == 0, "root cascades");
        Refuses(() => b.Parent());
        Refuses(() => NewAgent(app));
    });
    CheckCase("each relationship operation completes both ends or changes neither", () =>
    {
        var root = new Application(TextWriter.Null);
        var foreign = new Application(TextWriter.Null);
        var a = NewAgent(root);
        var other = NewAgent(root);
        var h = a.History();
        var e = a.Engine();
        try
        {
            Refuses(() => foreign.RemoveAgent(a));
            Refuses(() => root.AddAgent(a));
            Check(a.Parent() == root && root.Agents().Contains(a), "refused root change leaves both ends");
            root.RemoveAgent(a);
            Check(!root.Agents().Contains(a) && a.parent is null && !a.destroyed, "remove detaches without destroying");
            root.AddAgent(a);
            Check(a.Parent() == root && root.Agents().Contains(a), "add establishes both ends");
            Refuses(() => other.RemoveHistory(h));
            Refuses(() => other.AddHistory(h));
            Check(h.Parent() == a && a.History() == h, "refused history change leaves both ends");
            a.RemoveHistory(h);
            Check(h.parent is null && !h.destroyed, "parent remove clears History back-reference");
            Refuses(() => a.History());
            a.AddHistory(h);
            Check(a.History() == h && h.Parent() == a, "History reattached at both ends");
            Refuses(() => other.RemoveEngine(e));
            Refuses(() => other.AddEngine(e));
            Check(e.Parent() == a && a.Engine() == e, "refused Engine change leaves both ends");
            a.RemoveEngine(e);
            Check(e.parent is null && !e.destroyed, "parent remove clears Engine back-reference");
            Refuses(() => a.Engine());
            a.AddEngine(e);
            Check(a.Engine() == e && e.Parent() == a, "Engine reattached at both ends");
            a.Destroy();
            Check(a.parent is null && h.parent is null && e.parent is null, "Destroy relies on complete removes");
            Refuses(() => root.AddAgent(a));
            Refuses(() => other.AddHistory(h));
            Refuses(() => other.AddEngine(e));
            Check(!root.Agents().Contains(a), "destroyed objects cannot be reattached");
        }
        finally { e.Destroy(); h.Destroy(); a.Destroy(); root.Destroy(); foreign.Destroy(); }
    });
    CheckCase("immutable exact-number values and JSON round trip", () =>
    {
        var call = Call();
        var copy = (ToolCallPart)JsonCodec.ReadPart(J(JsonCodec.WritePart(call).ToJsonString()));
        Check(copy.Args.GetProperty("n").GetInt64() == 9007199254740993, "no double rounding");
        var source = new List<Part> { new TextPart("original") };
        var entry = new Entry(1, Actor.Human, source);
        source.Clear();
        Check(entry.Parts.Length == 1, "immutable sequence copies input");
        Refuses(() => new BlobRef(0, "bad"));
        Refuses(() => new Provenance(0, "x", Surface.Messages));
        Refuses(() => new Event((EventKind)999));
        Refuses(() => JsonCodec.ReadPart(J("{\"type\":\"blob\",\"mime\":\"x\",\"path\":\"old\"}")));
        Refuses(() => JsonCodec.ReadPart(J("{\"type\":\"blob\",\"mime\":\"x\",\"ref\":{\"kind\":\"\",\"locator\":\"x\"}}")));
    });
    CheckCase("all 32 phase/event pairs follow the published transition table", () =>
    {
        Event[] values = [new(EventKind.MessageReceived, actor: Actor.Human), new(EventKind.RequestSent, from: Origin()),
            new(EventKind.ResponseStarted, from: Origin()), new(EventKind.ResponseEnded, from: Origin()),
            new(EventKind.ToolCalled, callID: "a", name: "read_file", args: J("{}")), new(EventKind.ToolReturned, callID: "a"),
            new(EventKind.Redacted, redactFrom: 1, redactTo: 1, level: Redaction.Result), new(EventKind.ErrorOccurred, error: "bad")];
        foreach (var phase in Enum.GetValues<Phase>())
        {
            foreach (var value in values)
            {
                var before = new Context(phase: phase);
                var expected = (phase, value.Kind) switch
                {
                    (Phase.Idle, EventKind.MessageReceived) => Phase.InputPending,
                    (Phase.InputPending, EventKind.RequestSent) => Phase.InFlight,
                    (Phase.InFlight, EventKind.ResponseEnded or EventKind.ErrorOccurred) => Phase.Idle,
                    _ => phase
                };
                Check(History.Reduce(before, value).Phase == expected, phase + "/" + value.Kind);
            }
        }

        var pending = History.Reduce(new Context(phase: Phase.InFlight), new Event(EventKind.ResponseEnded, parts: [Call("a"), Call("b")]));
        Check(pending.Phase == Phase.ToolsPending, "calls enter ToolsPending");
        pending = History.Reduce(pending, new Event(EventKind.ToolReturned, callID: "b"));
        Check(pending.Phase == Phase.ToolsPending, "not all results");
        pending = History.Reduce(pending, new Event(EventKind.ToolReturned, callID: "a"));
        Check(pending.Phase == Phase.InputPending, "last result releases");
    });
    CheckCase("published numeric Ref, legacy surface alias, and optional timestamps", () =>
    {
        for (int kind = 1; kind <= 3; kind++)
        {
            var part = (BlobPart)JsonCodec.ReadPart(J("""{"type":"blob","mime":"image/png","ref":{"kind":KIND,"locator":"kept"}}""".Replace("KIND", kind.ToString())));
            Check((int)part.Ref.Kind == kind && JsonCodec.WritePart(part)["ref"]!["kind"]!.GetValue<int>() == kind, "numeric Ref preserved");
        }
        Refuses(() => JsonCodec.ReadPart(J("""{"type":"blob","mime":"image/png","ref":{"kind":0,"locator":"bad"}}""")));
        Check(JsonCodec.ReadProvenance(J("""{"vendor":"gemini","model":"exact","surface":"generatecontent"}""")).Surface == Surface.GenerateContent, "documented alias reads");
        var e = JsonCodec.ReadEvent("""{"seq":1,"type":"message_received","message":{"actor":"human","parts":[{"type":"text","text":"hi"}]}}""");
        Check(e.Time is null && !JsonCodec.WriteEvent(e).Contains("time"), "absent time stays absent");
    });
    CheckCase("ephemera replaces, clears, stays out of dialogue, and consumes on send", () =>
    {
        var app = new Application(TextWriter.Null);
        try
        {
            var a = NewAgent(app);
            a.SetEphemeral("old");
            a.SetEphemeral("new");
            Check(a.History().View().Dialogue.Length == 0 && JsonCodec.Text(a.History().View().Ephemeral) == "new", "one pending slot");
            a.SetEphemeral("");
            Check(a.History().View().Ephemeral.IsEmpty, "empty clears");
            a.SetEphemeral("once");
            var first = a.Engine().Render(a.History().View()).ToJsonString();
            var second = a.Engine().Render(a.History().View()).ToJsonString();
            Check(first == second && first.Contains("once"), "render does not consume");
            a.History().Append(new Event(EventKind.RequestSent, from: Origin()));
            a.History().Append(new Event(EventKind.ErrorOccurred, error: "failed"));
            Check(!a.Engine().Render(a.History().View()).ToJsonString().Contains("once"), "attempt consumes even on error");
        }
        finally { app.Destroy(); }
    });
    CheckCase("log round trip, deterministic replay, and failed load isolation", () =>
    {
        var app = new Application(TextWriter.Null);
        var path = Path.Combine(temp, "roundtrip.log");
        try
        {
            var a = NewAgent(app);
            a.History().StartLog(path);
            a.History().Append(new Event(EventKind.MessageReceived, actor: Actor.Human, parts: [new TextPart("question")]));
            a.History().Append(new Event(EventKind.RequestSent, from: Origin()));
            a.History().Append(new Event(EventKind.ResponseEnded, from: Origin(), parts: [new TextPart("answer")], usage: new Usage(1, 2, 3, 4)));
            var b = NewAgent(app);
            b.History().Load(path);
            Check(a.Engine().Render(a.History().View()).ToJsonString() == b.Engine().Render(b.History().View()).ToJsonString(), "live equals replay");
            using var dumped = new StringWriter();
            b.History().Dump(dumped);
            var dump = dumped.ToString();
            File.WriteAllText(Path.Combine(temp, "dump.log"), dump);
            b.History().Load(Path.Combine(temp, "dump.log"));
            using var again = new StringWriter();
            b.History().Dump(again);
            Check(dump == again.ToString(), "dump stable");
            File.WriteAllText(Path.Combine(temp, "bad.log"), "{\"log_version\":1}\n" + JsonCodec.WriteEvent(new Event(EventKind.MessageReceived, 1, actor: Actor.Human)) + "\n{\"seq\":2,\"type\":\"future\"}\n");
            Refuses(() => b.History().Load(Path.Combine(temp, "bad.log")));
            Check(b.History().Events().Length == 3, "failed load preserved old events");
            var lines = File.ReadAllLines(path);
            File.WriteAllLines(Path.Combine(temp, "headerless.log"), lines.Skip(1));
            b.History().Load(Path.Combine(temp, "headerless.log"));
            Check(b.History().Events().Length == 3, "headerless first event preserved");
            File.WriteAllText(Path.Combine(temp, "future-version.log"), "{\"log_version\":2}\n");
            Refuses(() => b.History().Load(Path.Combine(temp, "future-version.log")));
            Check(b.History().Events().Length == 3, "unknown explicit version refused");
            File.WriteAllLines(Path.Combine(temp, "duplicate.log"), lines.Concat([lines[^1]]));
            Refuses(() => b.History().Load(Path.Combine(temp, "duplicate.log")));
            Check(b.History().Events().Length == 3, "duplicate sequence rejected transactionally");
        }
        finally { app.Destroy(); }
    });
    CheckCase("redaction retains raw facts and exact blob locator without I/O", () =>
    {
        foreach (var kind in Enum.GetValues<RefKind>())
        {
            var reference = new BlobRef(kind, "never-fetch-this-sensitive-location");
            var original = new ToolResultPart("a", [new TextPart("SECRET"), new BlobPart("image/png", reference)], true);
            var view = new Context([new Entry(1, Actor.Agent, [Call()]), new Entry(2, Actor.Tool, [original])]);
            var redacted = History.Reduce(view, new Event(EventKind.Redacted, redactFrom: 2, redactTo: 2, level: Redaction.Result, reason: "test"));
            var result = (ToolResultPart)redacted.Dialogue[1].Parts[0];
            var stub = (RedactedPart)result.Parts[1];
            Check(stub.Ref!.Kind == kind && stub.Ref.Locator == reference.Locator && result.IsError, "ref and error survived reducer");
            Check(original.Parts[0] is TextPart { Text: "SECRET" }, "original untouched");
            foreach (var vendor in Enum.GetValues<Vendor>())
            {
                var wire = Wire(vendor, redacted);
                Check(!wire.Contains("SECRET") && !wire.Contains(reference.Locator) && wire.Contains("read_file"), "only stub renders");
                Refuses(() => Wire(vendor, view));
            }
        }
        var dialogue = new Context([new Entry(1, Actor.Human, [new TextPart("SECRET")]), new Entry(2, Actor.Agent, [Call()])]);
        var textGone = History.Reduce(dialogue, new Event(EventKind.Redacted, redactFrom: 1, redactTo: 2, level: Redaction.Dialogue));
        Check(textGone.Dialogue[1].Parts[0] is ToolCallPart && textGone.Dialogue[0].Parts.IsEmpty, "dialogue keeps tool exchange");
        var toolsGone = History.Reduce(dialogue, new Event(EventKind.Redacted, redactFrom: 2, redactTo: 2, level: Redaction.Tool));
        Check(toolsGone.Dialogue[1].Parts.IsEmpty, "tool removal filter");
        var summary = History.Reduce(dialogue, new Event(EventKind.Redacted, redactFrom: 1, redactTo: 2, level: Redaction.Summary));
        Check(summary.Dialogue.Length == 1 && summary.Dialogue[0].Actor == Actor.System, "range collapses once");
        var replacement = new Event(EventKind.Redacted, seq: 3, parts: [new TextPart("compressed explanation")], redactFrom: 1, redactTo: 2, level: Redaction.Summary);
        var restored = JsonCodec.ReadEvent(JsonCodec.WriteEvent(replacement));
        Check(JsonCodec.Text(History.Reduce(dialogue, restored).Dialogue[0].Parts) == "compressed explanation", "supplied summary survives log and replay");
    });
    CheckCase("interleaved results reordered only in wire and missing result is honest", () =>
    {
        var entries = new[] { new Entry(1, Actor.Agent, [Call("a"), Call("b")]),
            new Entry(2, Actor.Tool, [new ToolResultPart("a", [new TextPart("FIRST")])]),
            new Entry(3, Actor.Human, [new TextPart("HUMAN")]),
            new Entry(4, Actor.Tool, [new ToolResultPart("b", [new TextPart("SECOND")])]) };
        var view = new Context(entries);
        foreach (var vendor in Enum.GetValues<Vendor>())
        {
            var wire = Wire(vendor, view);
            Check(wire.IndexOf("FIRST", StringComparison.Ordinal) < wire.IndexOf("SECOND", StringComparison.Ordinal) && wire.IndexOf("SECOND", StringComparison.Ordinal) < wire.IndexOf("HUMAN", StringComparison.Ordinal), "group results before human");
            var missing = Wire(vendor, new Context(entries.Take(3)));
            Check(missing.Contains("execution outcome unknown") && !missing.Contains("SECOND"), "honest missing-result repair");
            Check(wire.Contains("9007199254740993"), "wire exact number");
        }
        Check(ReferenceEquals(view.Dialogue[2], entries[2]), "canonical order unmodified");
        var repeated = new Context(entries.Concat([new Entry(5, Actor.Agent, [Call("a")]), new Entry(6, Actor.Tool, [new ToolResultPart("a", [new TextPart("LATER")])])]));
        Check(Wire(Vendor.Anthropic, repeated).Contains("LATER"), "IDs can be reused after completion");
        Refuses(() => Wire(Vendor.Anthropic, new Context([new Entry(1, Actor.Agent, [Call(), Call()])])));
        Refuses(() => Wire(Vendor.Anthropic, new Context([new Entry(1, Actor.Tool, [new ToolResultPart("foreign", [])])])));
    });
    CheckCase("opaque exact-model capture/replay and Gemini error objects", () =>
    {
        var app = new Application(TextWriter.Null);
        try
        {
            var a = NewAgent(app, Vendor.Gemini);
            var events = a.Engine().Parse(200, """{"modelVersion":"returned-model","candidates":[{"content":{"parts":[{"functionCall":{"name":"read_file","args":{"n":9007199254740993}},"thoughtSignature":"SIG"},{"functionCall":{"name":"read_file","args":{}}}]}}],"usageMetadata":{"promptTokenCount":4,"candidatesTokenCount":2}}""");
            var calls = events[^1].Parts.OfType<ToolCallPart>().ToArray();
            Check(calls.Length == 2 && calls[0].CallID != calls[1].CallID, "distinct response-local IDs");
            Check(calls[0].From.Model == "returned-model" && calls[0].Opaque!.Value.GetString() == "SIG", "capture returned identity and per-call opaque");
            var view = new Context([new Entry(1, Actor.Agent, [calls[0]]), new Entry(2, Actor.Tool,
                [new ToolResultPart(calls[0].CallID, [new TextPart("{\"message\":\"bad\",\"error\":\"original detail\"}")], true)])]);
            var exact = Wire(Vendor.Gemini, view, "returned-model");
            Check(exact.Contains("SIG") && exact.Contains("\"error\"") && exact.Contains("read_file"), "signature plus JSON error marker and actual name");
            var original = JsonNode.Parse(exact)!["contents"]![1]!["parts"]![0]!["functionResponse"]!["response"]!["error"]!;
            Check(original["message"]!.GetValue<string>() == "bad" && original["error"]!.GetValue<string>() == "original detail", "failure wrapper preserves every original object field");
            Check(!Wire(Vendor.Gemini, view, "different-model").Contains("SIG") && !Wire(Vendor.OpenAI, view).Contains("SIG"), "opaque gated by exact provenance");
            Check(a.Engine().Parse(403, "not json")[0].Kind == EventKind.ErrorOccurred, "status before decode");
            Check(a.Engine().Parse(200, "not json")[0].Kind == EventKind.ErrorOccurred, "bad response is an error event");
        }
        finally { app.Destroy(); }
    });
    CheckCase("empty assistant text is retained, not null or absent", () =>
    {
        var view = new Context([new Entry(1, Actor.Human, [new TextPart("question")]), new Entry(2, Actor.Agent, [new TextPart("")])]);
        var wire = JsonNode.Parse(Wire(Vendor.OpenAI, view))!;
        Check(wire["messages"]!.AsArray().Count == 3 && wire["messages"]![2]!["content"]!.GetValue<string>() == "", "empty assistant survives projection");
    });
    CheckCase("usage normalization, invalid operands, overflow, and known output", () =>
    {
        var app = new Application(new StringWriter());
        try
        {
            var a = NewAgent(app, Vendor.OpenAI);
            Usage Parse(string usage) => a.Engine().Parse(200, "{\"model\":\"m\",\"choices\":[{\"message\":{\"content\":\"ok\"}}],\"usage\":" + usage + "}")[^1].Usage!;
            var valid = Parse("{\"prompt_tokens\":100,\"completion_tokens\":7,\"prompt_tokens_details\":{\"cached_tokens\":20,\"cache_write_tokens\":30}}");
            Check(valid.Input == 50 && valid.CacheRead == 20 && valid.CacheWrite == 30 && valid.Output == 7, "disjoint categories");
            foreach (var bad in new[] { "-1", "1.5", "\"3\"", "null", "true", "9223372036854775808" })
            {
                var invalid = Parse("{\"prompt_tokens\":100,\"completion_tokens\":7,\"prompt_tokens_details\":{\"cached_tokens\":" + bad + "}}");
                Check(invalid.Input is null && invalid.CacheRead is null && invalid.Output == 7, "invalid count must not become zero: " + bad);
            }
            var inconsistent = Parse("{\"prompt_tokens\":1,\"completion_tokens\":7,\"prompt_tokens_details\":{\"cached_tokens\":2}}");
            Check(inconsistent.Input is null && inconsistent.Output == 7, "invalid subset does not erase output");
            Check(Parse("{\"prompt_tokens\":5,\"completion_tokens\":2}").Input == 5, "missing optional counts are zero");
            Check(new Usage(long.MaxValue).Add(new Usage(1)).Input is null, "overflow unknown");
        }
        finally { app.Destroy(); }
    });
    foreach (var vendor in Enum.GetValues<Vendor>())
    {
        await AsyncCase("real Agent HTTP exchange and failure consumption: " + vendor, async () =>
        {
            var listener = new TcpListener(IPAddress.Loopback, 0);
            listener.Start();
            using var deadline = new CancellationTokenSource(TimeSpan.FromSeconds(15));
            var requests = new List<string>();
            var port = ((IPEndPoint)listener.LocalEndpoint).Port;
            var server = Task.Run(async () =>
            {
                for (int turn = 0; turn < 2; turn++)
                {
                    using var client = await listener.AcceptTcpClientAsync(deadline.Token);
                    using var stream = client.GetStream();
                    requests.Add(await ReadRequest(stream, deadline.Token));
                    var body = vendor switch
                    {
                        Vendor.Anthropic => """{"model":"model-a","content":[{"type":"text","text":"local reply"}],"usage":{"input_tokens":3,"output_tokens":2}}""",
                        Vendor.OpenAI => """{"model":"model-a","choices":[{"message":{"content":"local reply"}}],"usage":{"prompt_tokens":3,"completion_tokens":2}}""",
                        _ => """{"modelVersion":"model-a","candidates":[{"content":{"parts":[{"text":"local reply"}]}}],"usageMetadata":{"promptTokenCount":3,"candidatesTokenCount":2}}"""
                    };
                    if (turn == 0)
                    {
                        body = "{\"error\":\"temporary local refusal\"}";
                    }

                    var bytes = Encoding.UTF8.GetBytes(body);
                    var header = Encoding.ASCII.GetBytes($"HTTP/1.1 {(turn == 0 ? "503 Service Unavailable" : "200 OK")}\r\nContent-Type: application/json\r\nContent-Length: {bytes.Length}\r\nConnection: close\r\n\r\n");
                    await stream.WriteAsync(header, deadline.Token);
                    await stream.WriteAsync(bytes, deadline.Token);
                }
            }, deadline.Token);
            var app = new Application(TextWriter.Null);
            try
            {
                var a = NewAgent(app, vendor, "http://127.0.0.1:" + port);
                a.SetEphemeral("ONCE_ONLY");
                try
                {
                    await a.SendAsync("first", deadline.Token);
                    throw new Exception("Expected provider refusal");
                }
                catch (InvalidOperationException) { }
                Check(a.History().View().Phase == Phase.Idle, "failure returns idle");
                var text = await a.SendAsync("second", deadline.Token);
                await server;
                Check(text == "local reply" && requests.Count == 2, "actual response and no retry");
                Check(requests[0].Contains("ONCE_ONLY") && !requests[1].Contains("ONCE_ONLY"), "ephemera consumed across real failure");
                Check(requests[1].Contains("first") && requests[1].Contains("second"), "recorded conversation retained");
                Check(a.History().View().Usage.Input == 3 && a.History().View().Usage.Output == 2, "real usage derived from events");
                Check(a.History().Events().All(e => !JsonCodec.WriteEvent(e).Contains("dummy-local-only")), "credentials not recorded");
            }
            finally { app.Destroy(); listener.Stop(); deadline.Cancel(); try { await server; } catch (OperationCanceledException) { } }
        });
    }
}
finally { Directory.Delete(temp, true); }
Console.WriteLine($"{passed} passed; {failures} failed");
return failures == 0 ? 0 : 1;

static async Task<string> ReadRequest(NetworkStream stream, CancellationToken cancellation)
{
    using var header = new MemoryStream();
    var one = new byte[1];
    while (true)
    {
        if (await stream.ReadAsync(one, cancellation) != 1)
        {
            throw new IOException("EOF in headers");
        }

        header.WriteByte(one[0]);
        if (header.Length > 65536)
        {
            throw new IOException("oversized headers");
        }

        var bytes = header.GetBuffer();
        var n = (int)header.Length;
        if (n >= 4 && bytes[n - 4] == 13 && bytes[n - 3] == 10 && bytes[n - 2] == 13 && bytes[n - 1] == 10)
        {
            break;
        }
    }
    var headers = Encoding.ASCII.GetString(header.ToArray());
    var lengthLine = headers.Split("\r\n").Single(x => x.StartsWith("Content-Length:", StringComparison.OrdinalIgnoreCase));
    var length = int.Parse(lengthLine.Split(':', 2)[1].Trim());
    if (length is < 0 or > 1048576)
    {
        throw new IOException("oversized body");
    }

    var body = new byte[length];
    await stream.ReadExactlyAsync(body, cancellation);
    using var parsed = JsonDocument.Parse(body); // A real protocol fake rejects malformed JSON.
    return Encoding.UTF8.GetString(body);
}
