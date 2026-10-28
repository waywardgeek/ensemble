using System.Diagnostics;
using System.Net;
using System.Net.Sockets;
using System.Security.Cryptography;
using System.Text;
using System.Text.Json;
using System.Text.Json.Nodes;
using Ensemble;
using Application = Ensemble.Ensemble;

internal static class AuthRenewalTests
{
    private static void Check(bool condition, string message)
    {
        if (!condition)
        {
            throw new Exception(message);
        }
    }

    // A small working token issuer: form parsing, a rotating one-use token,
    // RSA-signed identities, JWKS and idempotent revocation over real HTTP.
    private sealed class Server : IAsyncDisposable
    {
        internal readonly HttpListener Listener = new();
        internal readonly AuthIdentityTests.Issuer Signing = new();
        internal readonly Uri Issuer;
        internal readonly Task Pump;
        internal readonly TaskCompletionSource RefreshEntered = new(TaskCreationOptions.RunContinuationsAsynchronously);
        internal readonly TaskCompletionSource RefreshRelease = new(TaskCreationOptions.RunContinuationsAsynchronously);
        internal readonly TaskCompletionSource RevokeEntered = new(TaskCreationOptions.RunContinuationsAsynchronously);
        internal readonly TaskCompletionSource RevokeRelease = new(TaskCreationOptions.RunContinuationsAsynchronously);
        internal bool HoldRefresh;
        internal bool HoldRevoke;
        internal bool OmitIdentity;
        internal bool WrongIdentity;
        internal bool OmitPlanScope;
        internal bool BadSignature;
        internal bool RedirectToken;
        internal bool BadGrant;
        internal int TokenStatus = 200;
        internal int RevokeStatus = 200;
        internal string ErrorCode = "invalid_grant";
        internal string ExpectedRefresh = "old-refresh";
        internal int Posts;
        internal int InferenceRefusals;
        internal int InferenceStatus = 401;
        internal readonly List<string> InferenceBearers = new();
        internal readonly List<string> InferenceBodies = new();
        internal int KeyReads;
        internal int Revocations;

        internal Server()
        {
            var port = new TcpListener(IPAddress.Loopback, 0);
            port.Start();
            int number = ((IPEndPoint)port.LocalEndpoint).Port;
            port.Stop();
            Issuer = new Uri($"http://127.0.0.1:{number}");
            Listener.Prefixes.Add(Issuer.AbsoluteUri);
            Listener.Start();
            Pump = Run();
        }

        private async Task Run()
        {
            try
            {
                while (Listener.IsListening)
                {
                    var context = await Listener.GetContextAsync();
                    try
                    {
                        await Handle(context);
                    }
                    finally
                    {
                        context.Response.Close();
                    }
                }
            }
            catch (Exception error) when (!Listener.IsListening
                && error is HttpListenerException or ObjectDisposedException)
            {
            }
        }

        private static async Task<Dictionary<string, string>> Form(HttpListenerRequest request)
        {
            Check(request.HttpMethod == "POST" && request.ContentType == "application/x-www-form-urlencoded",
                "real form-encoded POST");
            Check(request.Headers["Authorization"] is null && request.Headers["Cookie"] is null,
                "OAuth endpoints receive neither bearer header nor cookies");
            using var reader = new StreamReader(request.InputStream);
            string body = await reader.ReadToEndAsync();
            return body.Split('&').Select(pair => pair.Split('=', 2)).ToDictionary(
                pair => WebUtility.UrlDecode(pair[0]), pair => WebUtility.UrlDecode(pair[1]), StringComparer.Ordinal);
        }

        private static async Task Reply(HttpListenerResponse response, object value)
        {
            byte[] bytes = Encoding.UTF8.GetBytes(JsonSerializer.Serialize(value));
            response.ContentType = "application/json";
            response.ContentLength64 = bytes.Length;
            await response.OutputStream.WriteAsync(bytes);
        }

        private async Task Handle(HttpListenerContext context)
        {
            switch (context.Request.Url!.AbsolutePath)
            {
                case "/v1/models":
                    Check(context.Request.HttpMethod == "GET"
                        && context.Request.Headers["Authorization"] == "Bearer rotated-access-1",
                        "catalog receives the published rotated token");
                    await Reply(context.Response, new { models = new[] {
                        new { slug = "available-model", display_name = "Available model", visibility = "list" }
                    } });
                    return;
                case "/.well-known/jwks.json":
                    Interlocked.Increment(ref KeyReads);
                    await Reply(context.Response, Signing.Keys());
                    return;
                case "/.well-known/openid-configuration":
                    await Reply(context.Response, new
                    {
                        issuer = Issuer.GetLeftPart(UriPartial.Authority),
                        revocation_endpoint = new Uri(Issuer, "/revoke").AbsoluteUri
                    });
                    return;
                case "/revoke":
                    var revoke = await Form(context.Request);
                    Check(revoke.Count == 3 && revoke["token_type_hint"] == "refresh_token"
                        && revoke["client_id"] == "issued-client" && revoke["token"] == ExpectedRefresh,
                        "revoke the selected renewable session only");
                    Interlocked.Increment(ref Revocations);
                    RevokeEntered.TrySetResult();
                    if (HoldRevoke)
                    {
                        await RevokeRelease.Task;
                    }
                    context.Response.StatusCode = RevokeStatus;
                    context.Response.ContentLength64 = 0;
                    return;
                case "/v1/responses":
                {
                    Check(context.Request.HttpMethod == "POST", "inference uses POST");
                    using var reader = new StreamReader(context.Request.InputStream);
                    InferenceBodies.Add(await reader.ReadToEndAsync());
                    InferenceBearers.Add(context.Request.Headers["Authorization"] ?? "");
                    if (InferenceBearers.Count <= InferenceRefusals)
                    {
                        context.Response.StatusCode = InferenceStatus;
                        await Reply(context.Response, new { error = "private remote detail must not escape" });
                        return;
                    }
                    context.Response.ContentType = "text/event-stream";
                    string body = "data: " + JsonSerializer.Serialize(new
                    {
                        type = "response.completed",
                        response = new
                        {
                            status = "completed", model = "local-plan-model",
                            output = new[] { new { type = "message", role = "assistant", status = "completed",
                                content = new[] { new { type = "output_text", text = "recovered" } } } },
                            usage = new { input_tokens = 10, output_tokens = 2 }
                        }
                    }) + "\n\n";
                    byte[] bytes = Encoding.UTF8.GetBytes(body);
                    context.Response.ContentLength64 = bytes.Length;
                    await context.Response.OutputStream.WriteAsync(bytes);
                    return;
                }
                case "/api/accounts/oauth/token":
                    var form = await Form(context.Request);
                    Check(form.Count == 4 && form["grant_type"] == "refresh_token" && form["client_id"] == "issued-client"
                        && form["resource"] == "https://api.openai.com/v1" && !form.ContainsKey("scope"),
                        "refresh uses issued client, resource and retained scope");
                    int generation = Interlocked.Increment(ref Posts);
                    RefreshEntered.TrySetResult();
                    if (HoldRefresh)
                    {
                        await RefreshRelease.Task;
                    }
                    if (RedirectToken)
                    {
                        context.Response.StatusCode = 307;
                        context.Response.RedirectLocation = new Uri(Issuer, "/unexpected-redirect").AbsoluteUri;
                        await Reply(context.Response, new { error = "redirect_refused" });
                        return;
                    }
                    if (TokenStatus != 200 || form["refresh_token"] != ExpectedRefresh)
                    {
                        context.Response.StatusCode = TokenStatus == 200 ? 400 : TokenStatus;
                        await Reply(context.Response, new { error = ErrorCode, error_description = "private-token-echo" });
                        return;
                    }
                    string access = "rotated-access-" + generation;
                    ExpectedRefresh = "rotated-refresh-" + generation;
                    var claims = Signing.Claims();
                    claims["iss"] = Issuer.GetLeftPart(UriPartial.Authority);
                    claims.Remove("nonce");
                    claims["iat"] = DateTimeOffset.UtcNow.ToUnixTimeSeconds();
                    claims["exp"] = DateTimeOffset.UtcNow.AddHours(1).ToUnixTimeSeconds();
                    claims["at_hash"] = AuthProtocol.Encode(SHA256.HashData(Encoding.ASCII.GetBytes(access))[..16]);
                    if (WrongIdentity)
                    {
                        claims["sub"] = "another-person";
                    }
                    string idToken = Signing.Sign(claims);
                    if (BadSignature)
                    {
                        using var attacker = new AuthIdentityTests.Issuer();
                        idToken = attacker.Sign(claims);
                    }
                    var grant = new JsonObject
                    {
                        ["access_token"] = access, ["refresh_token"] = ExpectedRefresh,
                        ["token_type"] = "Bearer", ["expires_in"] = 3600
                    };
                    if (!OmitIdentity)
                    {
                        grant["id_token"] = idToken;
                    }
                    if (OmitPlanScope)
                    {
                        grant["scope"] = "openid email profile";
                    }
                    if (BadGrant)
                    {
                        grant["expires_in"] = "private-token-echo";
                    }
                    await Reply(context.Response, grant);
                    return;
                default:
                    throw new Exception("Unexpected issuer route; redirects must not be followed.");
            }
        }

        public async ValueTask DisposeAsync()
        {
            RefreshRelease.TrySetResult();
            RevokeRelease.TrySetResult();
            Listener.Stop();
            Listener.Close();
            await Pump;
            Signing.Dispose();
        }
    }

    private static AuthRegistration Seed(Server server, string label = "main", string? client = null, int remainingMinutes = -1) =>
        new(label, server.Issuer.GetLeftPart(UriPartial.Authority), "subject-one", client ?? "issued-client",
            "shared@example.invalid", "old-access", "old-refresh", "retained-id-token",
            DateTimeOffset.UtcNow.AddMinutes(remainingMinutes), "openid email profile offline_access resource.invoke chatgpt.tokens.use.direct");

    private static async Task Refuses<T>(Func<Task> operation) where T : Exception
    {
        try
        {
            await operation();
        }
        catch (T error)
        {
            Check(!error.ToString().Contains("private-token-echo", StringComparison.Ordinal), "safe refusal diagnostics");
            return;
        }
        throw new Exception("Expected refusal: " + typeof(T).Name);
    }

    internal static async Task<int> Child(string directory, string issuer)
    {
        var app = new Application(TextWriter.Null);
        try
        {
            var store = new Credentials(app, directory, new Uri(issuer));
            Console.WriteLine("ready");
            Console.Out.Flush();
            using var timeout = new CancellationTokenSource(TimeSpan.FromSeconds(15));
            Console.WriteLine(await store.GetBearerTokenAsync("main", timeout.Token));
            return 0;
        }
        finally
        {
            app.Destroy();
        }
    }

    private static Process StartChild(string directory, Uri issuer)
    {
        string executable = Environment.ProcessPath ?? throw new Exception("Missing dotnet process path");
        var info = new ProcessStartInfo(executable) { RedirectStandardOutput = true, RedirectStandardError = true };
        if (Path.GetFileNameWithoutExtension(executable) == "dotnet")
        {
            info.ArgumentList.Add(typeof(AuthRenewalTests).Assembly.Location);
        }
        foreach (string argument in new[] { "--auth-renew-child", directory, issuer.AbsoluteUri })
        {
            info.ArgumentList.Add(argument);
        }
        return Process.Start(info) ?? throw new Exception("Could not start auth refresh child");
    }

    internal static async Task RejectedBearerRecovery(string root)
    {
        // Each scenario owns real issuer/inference sockets and its own protected
        // store. No pre-expiry refresh: only the inference 401 can trigger renewal.
        foreach (string scenario in new[] { "success", "second-401", "forbidden", "api-401", "terminal", "transient" })
        {
            await using var server = new Server
            {
                InferenceRefusals = scenario == "second-401" ? 2 : 1,
                InferenceStatus = scenario == "forbidden" ? 403 : 401,
                TokenStatus = scenario == "terminal" ? 400 : scenario == "transient" ? 503 : 200,
                ErrorCode = "invalid_grant"
            };
            var app = new Application(TextWriter.Null);
            using var deadline = new CancellationTokenSource(TimeSpan.FromSeconds(15));
            try
            {
                string directory = Path.Combine(root, "auth-inference-recovery-" + scenario);
                var store = new Credentials(app, directory, server.Issuer);
                await store.StoreAsync(Seed(server, remainingMinutes: 60));
                var agent = new Agent(app, new Configuration(Vendor.OpenAI, "local-plan-model",
                    server.Issuer.AbsoluteUri, scenario == "api-401" ? "dummy-api-key" : "", surface: Surface.Responses,
                    billing: scenario == "api-401" ? BillingMode.ApiKey : BillingMode.ChatGptPlan,
                    account: scenario == "api-401" ? "" : "main"));
                const string wire = "{ \"input\": [ ], \"stream\": true }";
                if (scenario is "terminal" or "transient")
                {
                    await Refuses<PlanAuthorizationException>(() => agent.Engine().SendAsync(wire, deadline.Token));
                }
                else
                {
                    var result = await agent.Engine().SendAsync(wire, deadline.Token);
                    Check(result.Events.Count(fact => fact.Kind == EventKind.ResponseEnded)
                        == (scenario == "success" ? 1 : 0), "only success returns a completion");
                    Check(result.Events.Count(fact => fact.Kind == EventKind.ErrorOccurred)
                        == (scenario == "success" ? 0 : 1), "last refusal returns one error");
                }
                int requests = scenario is "success" or "second-401" ? 2 : 1;
                int refreshes = scenario is "forbidden" or "api-401" ? 0 : scenario == "transient" ? 3 : 1;
                Check(server.InferenceBodies.Count == requests && server.InferenceBodies.All(body => body == wire),
                    "at most one byte-identical inference retry: " + scenario);
                Check(server.InferenceBearers[0] == (scenario == "api-401" ? "Bearer dummy-api-key" : "Bearer old-access")
                    && (requests == 1 || server.InferenceBearers[1] == "Bearer rotated-access-1"),
                    "same-source retry uses the durable rotated bearer");
                Check(server.Posts == refreshes, "bounded refresh count: " + scenario);
                Check(agent.Engine().Meter().Totals.Output == (scenario == "success" ? 2 : 0)
                    && agent.History().Events().IsEmpty, "meter once and no Engine-owned History writes");
                Check((await store.AccountsAsync()).Single().SignedIn == (scenario != "terminal"),
                    "only terminal refresh error clears local credentials");
                if (scenario is "success" or "second-401")
                {
                    Check(await store.GetBearerTokenAsync("main") == "rotated-access-1" && server.Posts == 1,
                        "retry bearer was published, not borrowed from an unsaved response");
                }
            }
            finally
            {
                app.Destroy();
            }
        }
    }

    internal static async Task RejectedOlderGeneration(string root)
    {
        await using var server = new Server();
        var firstApp = new Application(TextWriter.Null);
        var secondApp = new Application(TextWriter.Null);
        try
        {
            string directory = Path.Combine(root, "auth-delayed-rejection");
            var first = new Credentials(firstApp, directory, server.Issuer);
            var second = new Credentials(secondApp, directory, server.Issuer);
            await first.StoreAsync(Seed(server, remainingMinutes: 60));
            string rejected = await first.GetBearerTokenAsync("main");
            string rotated = await second.GetBearerAfterRejectionAsync("main", rejected, default);
            Check(rotated == "rotated-access-1" && server.Posts == 1, "second owner rotates rejected generation");
            Check(await first.GetBearerAfterRejectionAsync("main", rejected, default) == rotated && server.Posts == 1,
                "delayed 401 rereads newer generation without rotating again");
        }
        finally
        {
            firstApp.Destroy();
            secondApp.Destroy();
        }
    }

    internal static async Task CatalogRenewal(string root)
    {
        await using var server = new Server();
        var app = new Application(TextWriter.Null);
        try
        {
            string directory = Path.Combine(root, "auth-catalog-renewal");
            var store = new Credentials(app, directory, server.Issuer);
            await store.StoreAsync(Seed(server));
            var models = await store.ModelsAsync("main", server.Issuer.AbsoluteUri);
            Check(models.Single().Slug == "available-model" && server.Posts == 1 && server.KeyReads == 1,
                "discovery refreshes expired grant and validates signed identity first");
            Check(await store.GetBearerTokenAsync("main") == "rotated-access-1" && server.Posts == 1,
                "catalog renewal is durably reusable by inference without another refresh");
        }
        finally
        {
            app.Destroy();
        }
    }

    internal static async Task SerializedRotation(string root)
    {
        await using var server = new Server { HoldRefresh = true };
        var app = new Application(TextWriter.Null);
        using var deadline = new CancellationTokenSource(TimeSpan.FromSeconds(15));
        try
        {
            string directory = Path.Combine(root, "auth-renew-processes");
            var store = new Credentials(app, directory, server.Issuer);
            // Still-valid access must proactively rotate inside the documented
            // five-minute margin, not wait until the token actually expires.
            await store.StoreAsync(Seed(server, remainingMinutes: 3));
            await store.StoreAsync(Seed(server, "other", "separate-client"));
            using var first = StartChild(directory, server.Issuer);
            using var second = StartChild(directory, server.Issuer);
            try
            {
                Check(await first.StandardOutput.ReadLineAsync(deadline.Token) == "ready", "first process started");
                Check(await second.StandardOutput.ReadLineAsync(deadline.Token) == "ready", "second process started");
                await server.RefreshEntered.Task.WaitAsync(deadline.Token);
                server.RefreshRelease.TrySetResult();
                await Task.WhenAll(first.WaitForExitAsync(deadline.Token), second.WaitForExitAsync(deadline.Token));
                string one = await first.StandardOutput.ReadToEndAsync();
                string two = await second.StandardOutput.ReadToEndAsync();
                Check(first.ExitCode == 0 && second.ExitCode == 0, "both refresh processes exited successfully: "
                    + await first.StandardError.ReadToEndAsync() + await second.StandardError.ReadToEndAsync());
                Check(one.Trim() == "rotated-access-1" && two.Trim() == one.Trim() && server.Posts == 1 && server.KeyReads == 1,
                    "only one process rotates; both use the saved newest token");
                Check(await store.GetBearerTokenAsync("main") == "rotated-access-1" && server.Posts == 1,
                    "third owner rereads current file without HTTP");
                using var disk = JsonDocument.Parse(File.ReadAllBytes(Path.Combine(directory, "credentials.json")));
                var accounts = disk.RootElement.GetProperty("accounts");
                Check(accounts[0].GetProperty("refresh_token").GetString() == "rotated-refresh-1"
                    && accounts[1].GetProperty("access_token").GetString() == "old-access", "rotation preserves other account");
            }
            finally
            {
                server.RefreshRelease.TrySetResult();
                foreach (var process in new[] { first, second })
                {
                    if (!process.HasExited)
                    {
                        process.Kill(entireProcessTree: true);
                        await process.WaitForExitAsync();
                    }
                }
            }
        }
        finally
        {
            app.Destroy();
        }
    }

    internal static async Task FailuresAndPermission(string root)
    {
        await using var server = new Server();
        var app = new Application(TextWriter.Null);
        using var deadline = new CancellationTokenSource(TimeSpan.FromSeconds(20));
        try
        {
            string directory = Path.Combine(root, "auth-renew-errors");
            string path = Path.Combine(directory, "credentials.json");
            var store = new Credentials(app, directory, server.Issuer);
            await store.StoreAsync(Seed(server));
            byte[] original = File.ReadAllBytes(path);
            foreach (string failure in new[] { "identity", "signature", "grant", "redirect", "temporary", "client" })
            {
                server.ExpectedRefresh = "old-refresh";
                server.WrongIdentity = failure == "identity";
                server.BadSignature = failure == "signature";
                server.BadGrant = failure == "grant";
                server.RedirectToken = failure == "redirect";
                server.TokenStatus = failure == "temporary" ? 503 : failure == "client" ? 400 : 200;
                server.ErrorCode = "invalid_client";
                int before = server.Posts;
                if (failure is "identity" or "signature" or "grant")
                {
                    await Refuses<InvalidDataException>(() => store.GetBearerTokenAsync("main", deadline.Token));
                }
                else
                {
                    await Refuses<HttpRequestException>(() => store.GetBearerTokenAsync("main", deadline.Token));
                }
                Check(File.ReadAllBytes(path).SequenceEqual(original), "failed refresh preserves prior file: " + failure);
                Check(server.Posts - before == (failure == "temporary" ? 3 : 1), "bounded retry only for transient status");
            }
            server.TokenStatus = 400;
            foreach (string code in new[] { "invalid_grant", "invalid_refresh_token", "token_expired",
                "refresh_token_expired", "refresh_token_invalidated", "refresh_token_reused" })
            {
                await store.StoreAsync(Seed(server));
                server.ErrorCode = code;
                await Refuses<InvalidOperationException>(() => store.GetBearerTokenAsync("main", deadline.Token));
                var account = (await store.AccountsAsync())[0];
                Check(!account.SignedIn && account.ClientID == "issued-client", "terminal refresh clears only tokens");
            }
            server.TokenStatus = 200;
            server.ExpectedRefresh = "old-refresh";
            server.OmitIdentity = true;
            server.OmitPlanScope = true;
            await store.StoreAsync(Seed(server));
            await Refuses<InvalidOperationException>(() => store.GetBearerTokenAsync("main", deadline.Token));
            var identityOnly = (await store.AccountsAsync())[0];
            Check(identityOnly.SignedIn && !identityOnly.PlanEnabled,
                "identity-only grant retained and reported without silently enabling inference");
            int savedPosts = server.Posts;
            await Refuses<InvalidOperationException>(() => store.GetBearerTokenAsync("main", deadline.Token));
            Check(server.Posts == savedPosts, "permission refusal does not loop through OAuth");
        }
        finally
        {
            app.Destroy();
        }
    }

    internal static async Task PublicationFailure(string root)
    {
        await using var server = new Server { HoldRefresh = true };
        var app = new Application(TextWriter.Null);
        using var deadline = new CancellationTokenSource(TimeSpan.FromSeconds(15));
        string directory = Path.Combine(root, "auth-renew-save-failure");
        string path = Path.Combine(directory, "credentials.json");
        string preserved = Path.Combine(directory, "preserved.json");
        try
        {
            var store = new Credentials(app, directory, server.Issuer);
            await store.StoreAsync(Seed(server));
            byte[] original = File.ReadAllBytes(path);
            var refresh = store.GetBearerTokenAsync("main", deadline.Token);
            await server.RefreshEntered.Task.WaitAsync(deadline.Token);
            // Move the owned fixture's old file aside and obstruct the final
            // replacement. The test, not another writer, injects this I/O fault.
            File.Move(path, preserved);
            Directory.CreateDirectory(path);
            server.RefreshRelease.TrySetResult();
            await Refuses<IOException>(async () => await refresh);
            Check(File.ReadAllBytes(preserved).SequenceEqual(original), "failed publication preserves original bytes");
            Check(Directory.GetFiles(directory, ".credentials.*.tmp").Length == 0, "failed publication cleans its private temp");
            Directory.Delete(path);
            File.Move(preserved, path);
            Check((await store.AccountsAsync())[0].ExpiresAt < DateTimeOffset.UtcNow,
                "no in-memory successful grant can conceal the failed durable update");
        }
        finally
        {
            server.RefreshRelease.TrySetResult();
            app.Destroy();
        }
    }

    internal static async Task LogoutAndDestruction(string root)
    {
        await using var server = new Server { HoldRevoke = true, RevokeStatus = 503 };
        var app = new Application(TextWriter.Null);
        using var deadline = new CancellationTokenSource(TimeSpan.FromSeconds(15));
        try
        {
            string directory = Path.Combine(root, "auth-renew-logout");
            var store = new Credentials(app, directory, server.Issuer);
            await store.StoreAsync(Seed(server));
            var signout = store.SignOutAsync("main", deadline.Token);
            await server.RevokeEntered.Task.WaitAsync(deadline.Token);
            string bytes = File.ReadAllText(Path.Combine(directory, "credentials.json"));
            Check(!bytes.Contains("old-refresh", StringComparison.Ordinal) && !bytes.Contains("old-access", StringComparison.Ordinal)
                && !bytes.Contains("retained-id-token", StringComparison.Ordinal), "local tokens absent before remote reply");
            server.RevokeRelease.TrySetResult();
            Check(!await signout && server.Revocations == 3, "remote 503 bounded and explicitly unconfirmed");
            Check(!(await store.AccountsAsync())[0].SignedIn, "failed revocation cannot restore tokens");
            await store.StoreAsync(Seed(server));
            server.RevokeStatus = 200;
            Check(await store.SignOutAsync("main", deadline.Token), "empty HTTP 200 confirms revocation");
            await store.StoreAsync(Seed(server));
            server.HoldRefresh = true;
            var refresh = store.GetBearerTokenAsync("main", deadline.Token);
            await server.RefreshEntered.Task.WaitAsync(deadline.Token);
            await Task.Run(app.Destroy).WaitAsync(deadline.Token);
            await Refuses<OperationCanceledException>(async () => await refresh);
            Check(app.Credentials() is null, "destruction cancels owned HTTP, joins operation, then detaches");
            Check(File.ReadAllText(Path.Combine(directory, "credentials.json")).Contains("old-refresh", StringComparison.Ordinal),
                "cancelled refresh does not publish a partial replacement");
            store.Destroy();
        }
        finally
        {
            server.RefreshRelease.TrySetResult();
            app.Destroy();
        }
    }
}
