using System.Net;
using System.Text;
using System.Text.Json;
using System.Text.Json.Nodes;
using Ensemble;
using Application = Ensemble.Ensemble;

// A real Engine/Agent and file reader run against a small working HTTP provider.
// The provider validates its route and paired grant, and emits actual SSE frames.
internal static class CodexTests
{
    private static void Check(bool condition, string message)
    {
        if (!condition) throw new Exception(message);
    }

    private static void Refuses<T>(Action action) where T : Exception
    {
        try { action(); }
        catch (T) { return; }
        throw new Exception("Expected " + typeof(T).Name);
    }

    private static Configuration Config(string endpoint = CodexCredentials.ServiceBase,
        string model = "gpt-6-astra", string cacheKey = "") =>
        new(Vendor.OpenAI, model, endpoint, surface: Surface.Responses,
            billing: BillingMode.Codex, promptCacheKey: cacheKey);

    private static string FilePath(string root) => Path.Combine(root, "codex-" + Guid.NewGuid().ToString("N") + ".json");

    private static void WriteGrant(string path, string generation)
    {
        string replacement = path + ".new";
        File.WriteAllText(replacement, JsonSerializer.Serialize(new
        {
            auth_mode = "chatgpt",
            OPENAI_API_KEY = (string?)null,
            tokens = new { access_token = "fake-bearer-" + generation, account_id = "fake-account-" + generation,
                id_token = "unused-id", refresh_token = "unused-refresh" }
        }));
        File.Move(replacement, path, true);
    }

    public static void Configuration()
    {
        Check((int)BillingMode.ApiKey == 1 && (int)BillingMode.ChatGptPlan == 2, "old billing numbers changed");
        Check(Config().FundingDescription().Contains("renewal belongs to Codex"), "funding hidden");
        _ = Config(CodexCredentials.ResponsesEndpoint);
        foreach (string endpoint in new[] { "https://api.openai.com", "http://127.0.0.1:1234",
            "https://chatgpt.com.evil.invalid/backend-api/codex", "https://chatgpt.com/backend-api/codex?x=1",
            "https://someone@chatgpt.com/backend-api/codex", "https://chatgpt.com:8443/backend-api/codex",
            "https://chatgpt.com/backend-api/codex#part" })
        {
            Refuses<ArgumentException>(() => Config(endpoint));
        }
        Refuses<ArgumentException>(() => new Configuration(Vendor.OpenAI, "gpt-6-astra", CodexCredentials.ServiceBase,
            apiKey: "not-allowed", surface: Surface.Responses, billing: BillingMode.Codex));
        Refuses<ArgumentException>(() => new Configuration(Vendor.OpenAI, "gpt-6-astra", CodexCredentials.ServiceBase,
            account: "public-account", surface: Surface.Responses, billing: BillingMode.Codex));
        Refuses<ArgumentException>(() => new Configuration(Vendor.OpenAI, "gpt-6-astra", CodexCredentials.ServiceBase,
            disableStreaming: true, surface: Surface.Responses, billing: BillingMode.Codex));
        Refuses<ArgumentException>(() => new Configuration(Vendor.Anthropic, "other", CodexCredentials.ServiceBase,
            billing: BillingMode.Codex));
    }

    public static async Task Credentials(string root)
    {
        string path = FilePath(root);
        var app = new Application(TextWriter.Null);
        var other = new Application(TextWriter.Null);
        try
        {
            // Construction performs no read; an explicitly absent file fails only
            // when authority is requested. There is no search of a default store.
            var source = new CodexCredentials(app, path);
            Check(ReferenceEquals(source.Parent(), app) && ReferenceEquals(app.CodexCredentials(), source), "ownership missing");
            Refuses<InvalidOperationException>(() => other.AddCodexCredentials(source));
            Check(other.CodexCredentials() is null && ReferenceEquals(source.Parent(), app), "failed add mutated parents");
            Refuses<InvalidOperationException>(() => other.RemoveCodexCredentials(source));
            await MustRefuse(source);
            WriteGrant(path, "A");
            byte[] original = File.ReadAllBytes(path);
            var first = await source.ReadAsync(default);
            Check(first.Bearer == "fake-bearer-A" && first.Account == "fake-account-A", "pair A");
            Check(File.ReadAllBytes(path).SequenceEqual(original), "reader rewrote file");
            WriteGrant(path, "B");
            var second = await source.ReadAsync(default);
            Check(second.Bearer == "fake-bearer-B" && second.Account == "fake-account-B", "stale pair after rotation");
            using var cancelled = new CancellationTokenSource();
            cancelled.Cancel();
            try { await source.ReadAsync(cancelled.Token); throw new Exception("ignored cancellation"); }
            catch (OperationCanceledException) { }
            app.Destroy();
            Check(app.CodexCredentials() is null && source.parent is null && source.destroyed, "destroy did not detach");
            Check(File.Exists(path), "destruction deleted Codex file");
            source.Destroy();
            try { await source.ReadAsync(default); throw new Exception("destroyed source read"); }
            catch (ObjectDisposedException) { }
        }
        finally
        {
            app.Destroy();
            other.Destroy();
        }
    }

    private static async Task MustRefuse(CodexCredentials source)
    {
        try { await source.ReadAsync(default); }
        catch (CodexAuthorizationException failure)
        {
            Check(failure.InnerException is null && !failure.ToString().Contains("fake-bearer"), "secret diagnostic");
            return;
        }
        throw new Exception("invalid credentials accepted");
    }

    public static async Task InvalidFiles(string root)
    {
        string path = FilePath(root);
        var app = new Application(TextWriter.Null);
        try
        {
            var source = new CodexCredentials(app, path);
            foreach (string json in new[] {
                "[]", "{", "{}", "{\"tokens\":null}",
                "{\"OPENAI_API_KEY\":\"fake-bearer-key\"}",
                "{\"auth_mode\":\"apikey\",\"tokens\":{\"access_token\":\"fake-bearer\",\"account_id\":\"a\"}}",
                "{\"tokens\":{\"access_token\":\"fake-bearer\",\"access_token\":\"other\",\"account_id\":\"a\"}}",
                "{\"tokens\":{\"access_token\":\"fake-bearer\",\"account_id\":null}}",
                "{\"tokens\":{\"access_token\":\"fake-bearer\\r\\nheader\",\"account_id\":\"a\"}}",
                "{\"tokens\":{\"access_token\":\"fake-bearer\",\"account_id\":\"a b\"}}",
                new string('x', 1024 * 1024 + 1) })
            {
                File.WriteAllText(path, json);
                await MustRefuse(source);
                Check(File.ReadAllText(path) == json, "invalid file replaced");
            }
            File.WriteAllBytes(path, [0xff, 0xfe, 0x00]);
            await MustRefuse(source);
            // Legacy Codex files may omit auth_mode but still require the pair.
            File.WriteAllText(path, "{\"tokens\":{\"access_token\":\"fake-bearer-L\",\"account_id\":\"fake-account-L\"}}");
            Check((await source.ReadAsync(default)).Account == "fake-account-L", "legacy paired grant refused");
        }
        finally { app.Destroy(); }
    }

    private sealed record Request(string Body, string Bearer, string Account, string Session);

    private sealed class Provider : HttpMessageHandler
    {
        public readonly List<Request> Requests = [];
        public HttpStatusCode Status = HttpStatusCode.OK;
        public bool Complete = true;
        public bool Disposed;
        protected override async Task<HttpResponseMessage> SendAsync(HttpRequestMessage request, CancellationToken cancellation)
        {
            Check(request.Method == HttpMethod.Post && request.RequestUri!.AbsoluteUri == CodexCredentials.ResponsesEndpoint, "wrong endpoint");
            string body = await request.Content!.ReadAsStringAsync(cancellation);
            string bearer = request.Headers.Authorization?.Parameter ?? "";
            string account = request.Headers.GetValues("ChatGPT-Account-ID").Single();
            string session = request.Headers.GetValues("session-id").Single();
            Requests.Add(new(body, bearer, account, session));
            Check(bearer.StartsWith("fake-bearer-", StringComparison.Ordinal)
                && account == "fake-account-" + bearer["fake-bearer-".Length..], "grant generations mixed");
            string payload;
            if (Status != HttpStatusCode.OK)
            {
                payload = "remote error containing " + bearer + " " + account;
            }
            else
            {
                var response = new JsonObject
                {
                    ["id"] = "response-1", ["status"] = "completed", ["model"] = "gpt-6-astra",
                    ["output"] = new JsonArray(new JsonObject
                    {
                        ["id"] = "message-1", ["type"] = "message", ["role"] = "assistant", ["status"] = "completed",
                        ["content"] = new JsonArray(new JsonObject { ["type"] = "output_text", ["text"] = "OK" })
                    }),
                    ["usage"] = new JsonObject { ["input_tokens"] = 100, ["output_tokens"] = 5,
                        ["input_tokens_details"] = new JsonObject { ["cached_tokens"] = 80 } },
                    ["ignored_diagnostic"] = bearer + " " + account
                };
                payload = Complete ? "data: " + new JsonObject { ["type"] = "response.completed", ["response"] = response }.ToJsonString() + "\n\n"
                    : "data: {\"type\":\"response.created\",\"response\":{\"id\":\"response-1\",\"status\":\"in_progress\"}}\n\n";
            }
            // Intentionally no Content-Type: this matches the observed endpoint.
            var reply = new HttpResponseMessage(Status) { Content = new ByteArrayContent(Encoding.UTF8.GetBytes(payload)) };
            if (Status == HttpStatusCode.Redirect) reply.Headers.Location = new Uri("https://untrusted.invalid/");
            return reply;
        }
        protected override void Dispose(bool disposing)
        {
            Disposed = true;
            base.Dispose(disposing);
        }
    }

    public static async Task Transport(string root)
    {
        string path = FilePath(root);
        WriteGrant(path, "A");
        var trace = new StringWriter();
        var app = new Application(TextWriter.Null, trace);
        var provider = new Provider();
        try
        {
            _ = new CodexCredentials(app, path);
            var agent = new Agent(app, Config(cacheKey: "fixed-cache-key"));
            agent.Engine().Destroy();
            var engine = new Engine(agent, provider);
            var context = new Context([new Entry(1, Actor.Human, [new TextPart("reply OK")])]);
            string body = engine.Render(context).ToJsonString();
            var wire = JsonNode.Parse(body)!;
            Check(wire["stream"]!.GetValue<bool>() && !wire["store"]!.GetValue<bool>(), "Codex response options");
            Check(wire["instructions"] is not null && wire["prompt_cache_key"]!.GetValue<string>() == "fixed-cache-key", "instruction/key projection");
            string documentary = engine.SavedConfiguration(context).GetRawText();
            Check(!documentary.Contains(path) && !documentary.Contains("fake-bearer"), "configuration captured credentials");
            for (int i = 0; i < 2; i++)
            {
                if (i == 1) WriteGrant(path, "B");
                var response = await engine.SendAsync(body);
                Check(response.Events.Any(item => item.Kind == EventKind.ResponseEnded), "missing terminal success without Content-Type");
            }
            Check(provider.Requests.Count == 2 && provider.Requests.All(item => item.Body == body && item.Session == "fixed-cache-key"), "request/affinity changed");
            Check(provider.Requests[1].Account == "fake-account-B", "transport cached old grant");
            var meter = engine.Meter();
            Check(meter.Totals.Input == 40 && meter.Totals.AllInput() == 200 && meter.Totals.Output == 10 && meter.Totals.CacheRead == 160, "completed usage not counted exactly once");
            Check(!trace.ToString().Contains("fake-bearer-") && !trace.ToString().Contains("fake-account-"), "trace leaked grant");
        }
        finally { app.Destroy(); }
        Check(provider.Disposed, "Engine did not own handler lifetime");
    }

    public static async Task ActorFailure(string root)
    {
        using var deadline = new CancellationTokenSource(TimeSpan.FromSeconds(15));
        var app = new Application(TextWriter.Null);
        try
        {
            string path = FilePath(root);
            WriteGrant(path, "A");
            _ = new CodexCredentials(app, path);
            var agent = new Agent(app, Config());
            agent.Engine().Destroy();
            var provider = new Provider { Status = HttpStatusCode.Unauthorized };
            var engine = new Engine(agent, provider);
            bool refused = false;
            try { await agent.SendAsync("reply OK", deadline.Token); }
            catch (CodexAuthorizationException) { refused = true; }
            var facts = agent.History().Events();
            var errors = facts.Where(fact => fact.Kind == EventKind.ErrorOccurred).ToArray();
            Check(refused && errors.Length == 1 && errors[0].Error!.Contains("renew in Codex")
                && facts.Count(fact => fact.Kind == EventKind.TurnEnded) == 1
                && !facts.Any(fact => fact.Kind == EventKind.ResponseEnded), "actor lost or duplicated refusal");
            Check(provider.Requests.Count == 1 && engine.Meter().Totals.AllInput() is null or 0,
                "actor retried or metered rejection");
            Check(!JsonSerializer.Serialize(facts.Select(JsonCodec.WriteEvent)).Contains("fake-bearer"), "History leaked bearer");
        }
        finally { app.Destroy(); }
    }

    public static async Task Refusals(string root)
    {
        string path = FilePath(root);
        WriteGrant(path, "A");
        var app = new Application(TextWriter.Null);
        try
        {
            var agent = new Agent(app, Config());
            agent.Engine().Destroy();
            var provider = new Provider();
            var engine = new Engine(agent, provider);
            string body = engine.Render(new Context([])).ToJsonString();
            Check(body == engine.Render(new Context([])).ToJsonString(), "fallback affinity changed between renders");
            try { await engine.SendAsync(body); throw new Exception("missing source accepted"); }
            catch (CodexAuthorizationException) { }
            Check(provider.Requests.Count == 0, "HTTP before credential refusal");
            _ = new CodexCredentials(app, path);
            byte[] before = File.ReadAllBytes(path);
            foreach (var status in new[] { HttpStatusCode.Unauthorized, HttpStatusCode.TooManyRequests, HttpStatusCode.Redirect })
            {
                provider.Status = status;
                int count = provider.Requests.Count;
                if (status == HttpStatusCode.Unauthorized)
                {
                    try { await engine.SendAsync(body); throw new Exception("401 not refused"); }
                    catch (CodexAuthorizationException) { }
                }
                else
                {
                    var response = await engine.SendAsync(body);
                    Check(!response.Events.Any(item => item.Kind == EventKind.ResponseEnded), "HTTP error accepted");
                }
                Check(provider.Requests.Count == count + 1, "implicit retry");
                Check(File.ReadAllBytes(path).SequenceEqual(before), "rejection rewrote credentials");
            }
            provider.Status = HttpStatusCode.OK;
            provider.Complete = false;
            var incomplete = await engine.SendAsync(body);
            Check(!incomplete.Events.Any(item => item.Kind == EventKind.ResponseEnded), "EOF accepted as completed");
            Check(engine.Meter().Totals.Input is null or 0, "failed attempts metered as success");
        }
        finally { app.Destroy(); }
    }
}
