using System.Net;
using System.Net.Sockets;
using System.Security.Cryptography;
using System.Text;
using System.Text.Json;
using System.Text.Json.Nodes;
using Ensemble;
using Application = Ensemble.Ensemble;

internal static class AuthLoginTests
{
    private static void Check(bool condition, string message)
    {
        if (!condition)
        {
            throw new Exception(message);
        }
    }

    private static Dictionary<string, string> Query(string value) => value.TrimStart('?').Split('&')
        .Select(pair => pair.Split('=', 2)).ToDictionary(pair => WebUtility.UrlDecode(pair[0]),
            pair => WebUtility.UrlDecode(pair[1]), StringComparer.Ordinal);

    private static int FreePort()
    {
        var socket = new TcpListener(IPAddress.Loopback, 0);
        socket.Start();
        int port = ((IPEndPoint)socket.LocalEndpoint).Port;
        socket.Stop();
        return port;
    }

    // A functioning authorization-code issuer, not expectations on internal calls.
    // Tickets carry the actual challenge/nonce/redirect. An exchange consumes its
    // ticket, verifies the form and PKCE proof, then signs the returned identity.
    private sealed class Issuer : IAsyncDisposable
    {
        internal readonly HttpListener Listener = new();
        internal readonly AuthIdentityTests.Issuer Signing = new();
        internal readonly Uri Address;
        internal readonly Task Pump;
        internal readonly Dictionary<string, Dictionary<string, string>> Tickets = new();
        internal readonly TaskCompletionSource Entered = new(TaskCreationOptions.RunContinuationsAsynchronously);
        internal readonly TaskCompletionSource Release = new(TaskCreationOptions.RunContinuationsAsynchronously);
        internal string StorePath = "";
        internal string Subject = "subject-one";
        internal string? LastClient;
        internal bool WrongNonce;
        internal bool MissingIdentity;
        internal bool IdentityOnly;
        internal bool Hold;
        internal int Status = 200;
        internal int Exchanges;
        internal int Authorizations;
        internal int Registrations;
        internal int PendingObserved;

        internal Issuer()
        {
            Address = new Uri($"http://127.0.0.1:{FreePort()}");
            Listener.Prefixes.Add(Address.AbsoluteUri);
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

        private static async Task Reply(HttpListenerResponse response, object value)
        {
            byte[] body = JsonSerializer.SerializeToUtf8Bytes(value);
            response.ContentType = "application/json";
            response.ContentLength64 = body.Length;
            await response.OutputStream.WriteAsync(body);
        }

        private async Task Handle(HttpListenerContext context)
        {
            var request = context.Request;
            if (request.Url!.AbsolutePath == "/api/accounts/authorize")
            {
                Check(request.HttpMethod == "GET", "authorization is browser GET");
                var parameters = Query(request.Url.Query);
                Check(parameters["response_type"] == "code" && parameters["code_challenge_method"] == "S256",
                    "code flow with S256");
                Check(parameters["resource"] == "https://api.openai.com/v1"
                    && parameters["scope"] == "openid profile email offline_access resource.invoke chatgpt.tokens.use.direct",
                    "authorization scopes/resource");
                Check(parameters["state"].Length >= 43 && parameters["nonce"].Length >= 43
                    && parameters["ext_agent_host_id"].Length != 0, "random state/nonce and stable host ID");
                string client = parameters["client_id"];
                if (client == "dynamic_agent_client")
                {
                    Check(parameters["agent_name_hint"] == "Ensemble" && !parameters.ContainsKey("id_token_hint"),
                        "bootstrap branding, no old identity hint");
                    client = "issued-" + ++Registrations;
                }
                else
                {
                    Check(!parameters.ContainsKey("agent_name_hint") && client.StartsWith("issued-"),
                        "issued registration retry is not a bootstrap");
                }
                Check(!parameters.ContainsKey("client_name") && !parameters.ContainsKey("issued_client_id"),
                    "no obsolete protocol aliases");
                parameters.Add("issued", client);
                string code = "one-use-" + ++Authorizations;
                Tickets.Add(code, parameters);
                context.Response.StatusCode = 302;
                context.Response.RedirectLocation = parameters["redirect_uri"] + "?code=" + code
                    + "&state=" + parameters["state"] + "&client_id=" + client;
                return;
            }
            if (request.Url.AbsolutePath == "/.well-known/jwks.json")
            {
                await Reply(context.Response, Signing.Keys());
                return;
            }
            Check(request.Url.AbsolutePath == "/api/accounts/oauth/token" && request.HttpMethod == "POST",
                "token endpoint and method");
            Check(request.ContentType == "application/x-www-form-urlencoded" && request.Headers["Authorization"] is null
                && request.Headers["Cookie"] is null, "private form, no bearer/cookie forwarding");
            using var reader = new StreamReader(request.InputStream);
            var form = Query(await reader.ReadToEndAsync());
            Check(form.Count == 6 && form["grant_type"] == "authorization_code"
                && form["resource"] == "https://api.openai.com/v1", "exact exchange fields and resource");
            Check(Tickets.Remove(form["code"], out var ticket), "one-use authorization ticket");
            Check(form["client_id"] == ticket!["issued"] && form["redirect_uri"] == ticket["redirect_uri"],
                "issued client and exact redirect at exchange");
            Check(AuthProtocol.Encode(SHA256.HashData(Encoding.ASCII.GetBytes(form["code_verifier"])))
                == ticket["code_challenge"], "actual PKCE proof matches authorization");
            using (var saved = JsonDocument.Parse(await File.ReadAllBytesAsync(StorePath)))
            {
                var account = saved.RootElement.GetProperty("accounts").EnumerateArray()
                    .Single(item => item.GetProperty("client_id").GetString() == form["client_id"]);
                if (account.GetProperty("subject").ValueKind == JsonValueKind.Null)
                {
                    Check(account.GetProperty("access_token").ValueKind == JsonValueKind.Null
                        && account.GetProperty("refresh_token").ValueKind == JsonValueKind.Null
                        && account.GetProperty("id_token").ValueKind == JsonValueKind.Null,
                        "pending registration is already durable and token-free before exchange");
                    PendingObserved++;
                }
            }
            LastClient = form["client_id"];
            Exchanges++;
            Entered.TrySetResult();
            if (Hold)
            {
                await Release.Task;
            }
            context.Response.StatusCode = Status;
            if (Status != 200)
            {
                await Reply(context.Response, new { error = "temporarily_unavailable", error_description = "private-token-echo" });
                return;
            }
            var claims = Signing.Claims();
            claims["sub"] = Subject;
            claims["nonce"] = WrongNonce ? "wrong-nonce" : ticket["nonce"];
            claims["iat"] = DateTimeOffset.UtcNow.ToUnixTimeSeconds();
            claims["exp"] = DateTimeOffset.UtcNow.AddHours(1).ToUnixTimeSeconds();
            claims["at_hash"] = AuthProtocol.Encode(SHA256.HashData(Encoding.ASCII.GetBytes("opaque-access"))[..16]);
            claims["iss"] = Address.GetLeftPart(UriPartial.Authority);
            claims["aud"] = form["client_id"];
            var grant = new JsonObject
            {
                ["token_type"] = "Bearer", ["access_token"] = "opaque-access", ["refresh_token"] = "refresh-secret",
                ["expires_in"] = 3600, ["scope"] = IdentityOnly ? "openid profile email offline_access" :
                    "openid profile email offline_access resource.invoke chatgpt.tokens.use.direct"
            };
            if (!MissingIdentity)
            {
                grant["id_token"] = Signing.Sign(claims);
            }
            await Reply(context.Response, grant);
        }

        internal async Task<string> Authorize(Uri authorization)
        {
            using var client = new HttpClient(new HttpClientHandler { AllowAutoRedirect = false, UseCookies = false });
            using var response = await client.GetAsync(authorization);
            Check(response.StatusCode == HttpStatusCode.Found, "browser reaches authorization endpoint");
            return response.Headers.Location!.AbsoluteUri;
        }

        public async ValueTask DisposeAsync()
        {
            Release.TrySetResult();
            Listener.Stop();
            await Pump;
            Listener.Close();
            Signing.Dispose();
        }
    }

    private static async Task Refuses<T>(Func<Task> operation) where T : Exception
    {
        try
        {
            await operation();
        }
        catch (T error)
        {
            Check(!error.ToString().Contains("private-token-echo"), "failure diagnostics omit server content");
            return;
        }
        throw new Exception("Expected " + typeof(T).Name);
    }

    internal static async Task LoginAndReuse(string root)
    {
        await using var issuer = new Issuer();
        var app = new Application(TextWriter.Null);
        try
        {
            string directory = Path.Combine(root, "auth-login-success");
            issuer.StorePath = Path.Combine(directory, "credentials.json");
            var store = new Credentials(app, directory, issuer.Address);
            var redirect = new Uri($"http://127.0.0.1:{FreePort()}/auth/callback");
            // Own the real callback socket before presenting the URL, just as a
            // host must. This browser simulation follows the issuer's redirect.
            using var callback = new HttpListener();
            callback.Prefixes.Add(redirect.GetLeftPart(UriPartial.Authority) + "/");
            callback.Start();
            Check(await store.SelectedPlanAccountAsync() is null, "new store retains API default");
            var authorization = await store.BeginLoginAsync("work", redirect);
            var request = callback.GetContextAsync();
            using var browser = new HttpClient(new HttpClientHandler { UseCookies = false });
            var navigation = browser.GetAsync(authorization);
            var received = await request.WaitAsync(TimeSpan.FromSeconds(10));
            string complete = redirect.GetLeftPart(UriPartial.Authority) + received.Request.RawUrl;
            var account = await store.CompleteLoginAsync(complete);
            received.Response.StatusCode = 200;
            received.Response.Close();
            using var page = await navigation;
            Check(page.StatusCode == HttpStatusCode.OK, "real loopback callback completed browser navigation");
            Check(account.SignedIn && account.PlanEnabled && account.Subject == "subject-one"
                && issuer.PendingObserved == 1, "verified subject bound only after persisted pending state");
            Check(await store.SelectedPlanAccountAsync() == "work", "validated login saves plan selection");
            Check(await store.GetBearerTokenAsync("work") == "opaque-access", "fresh login usable without refresh");
            await Refuses<InvalidOperationException>(() => store.CompleteLoginAsync(complete));
            Check(issuer.Exchanges == 1, "callback cannot replay code");

            var again = await store.BeginLoginAsync("work", redirect, requestConsent: true);
            var firstQuery = Query(authorization.Query);
            var againQuery = Query(again.Query);
            Check(againQuery["client_id"] == account.ClientID && againQuery["prompt"] == "consent"
                && againQuery["login_hint"] == "shared@example.invalid" && againQuery.ContainsKey("id_token_hint")
                && againQuery["ext_agent_host_id"] == firstQuery["ext_agent_host_id"], "saved registration and hints reused");
            Check(firstQuery["nonce"] != againQuery["nonce"] && firstQuery["state"] != againQuery["state"]
                && firstQuery["code_challenge"] != againQuery["code_challenge"] && !firstQuery.ContainsKey("prompt"),
                "fresh randomness and explicit-only reconsent");
            issuer.IdentityOnly = true;
            account = await store.CompleteLoginAsync(await issuer.Authorize(again));
            Check(account.SignedIn && !account.PlanEnabled, "identity-only successful login remains signed in");
            Check(await store.SelectedPlanAccountAsync() == "work", "missing permission does not select API fallback");
            await Refuses<InvalidOperationException>(() => store.GetBearerTokenAsync("work"));

            issuer.Subject = "subject-two";
            var separate = await store.BeginLoginAsync("personal", redirect);
            var other = await store.CompleteLoginAsync(await issuer.Authorize(separate));
            Check(other.Email == account.Email && other.ClientID != account.ClientID && other.Subject != account.Subject
                && (await store.AccountsAsync()).Length == 2, "equal emails never merge separately registered accounts");
            Check(await store.SelectedPlanAccountAsync() == "personal", "latest successful login chooses its account");
            await store.LogoutAsync("personal");
            Check(await store.SelectedPlanAccountAsync() == "personal", "logout preserves funding consent boundary");
            var peer = new Application(TextWriter.Null);
            try
            {
                var restored = new Credentials(peer, directory, issuer.Address);
                Check(await restored.SelectedPlanAccountAsync() == "personal", "new owner reads durable selection");
                await Refuses<InvalidOperationException>(() => restored.GetBearerTokenAsync("personal"));
            }
            finally
            {
                peer.Destroy();
            }
        }
        finally
        {
            app.Destroy();
        }
    }

    internal static async Task CallbackRefusals(string root)
    {
        await using var issuer = new Issuer();
        var app = new Application(TextWriter.Null);
        try
        {
            string directory = Path.Combine(root, "auth-login-callbacks");
            issuer.StorePath = Path.Combine(directory, "credentials.json");
            var store = new Credentials(app, directory, issuer.Address);
            foreach (string bad in new[] { "https://127.0.0.1:43219/auth/callback", "http://localhost:43219/auth/callback",
                "http://127.0.0.1:43219/foreign", "http://127.0.0.1:43219/auth/callback?extra=1",
                "http://name@127.0.0.1:43219/auth/callback", "http://127.0.0.1:43219/auth/callback#fragment" })
            {
                await Refuses<ArgumentException>(() => store.BeginLoginAsync("work", new Uri(bad)));
            }
            var redirect = new Uri("http://127.0.0.1:43219/auth/callback");
            var authorization = await store.BeginLoginAsync("work", redirect);
            Check(await store.LoginPendingAsync(), "host sees active login without secret state");
            await Refuses<InvalidOperationException>(() => store.BeginLoginAsync("second", redirect));
            string callback = await issuer.Authorize(authorization);
            string original = await File.ReadAllTextAsync(issuer.StorePath);
            string state = Query(authorization.Query)["state"];
            foreach (string bad in new[] { callback.Replace(state, "wrong-state"), callback.Replace("127.0.0.1", "localhost"),
                callback.Replace(":43219", ":43220"), callback.Replace("/auth/callback", "/foreign"), callback + "#fragment",
                callback + "&state=" + state, callback + "&client%5fid=duplicate", callback + "&bad=%FF",
                callback + "&bad=%xy", callback + "&bad=%00", callback.Replace("?code=", "?missing=").Replace("&state=", "&other=") })
            {
                await Refuses<InvalidDataException>(() => store.CompleteLoginAsync(bad));
                Check(await store.LoginPendingAsync(), "unsolicited callback leaves host listener active");
                Check(await File.ReadAllTextAsync(issuer.StorePath) == original && issuer.Exchanges == 0,
                    "invalid callback changes neither disk nor network");
            }
            string missingClient = callback[..callback.IndexOf("&client_id=", StringComparison.Ordinal)];
            await Refuses<InvalidDataException>(() => store.CompleteLoginAsync(missingClient));
            Check(issuer.Exchanges == 0 && await File.ReadAllTextAsync(issuer.StorePath) == original,
                "bootstrap requires issued client ID before writes or exchange");
            authorization = await store.BeginLoginAsync("work", redirect);
            callback = await issuer.Authorize(authorization);
            await store.CompleteLoginAsync(callback);
            Check(issuer.Exchanges == 1, "valid callback alone exchanges the code");
            var substituted = await store.BeginLoginAsync("work", redirect);
            string foreignClient = (await issuer.Authorize(substituted)).Replace("client_id=issued-2", "client_id=foreign");
            await Refuses<InvalidDataException>(() => store.CompleteLoginAsync(foreignClient));
            Check(issuer.Exchanges == 1, "saved client ID is authoritative, not callback substitution");
            var cancelled = await store.BeginLoginAsync("work", redirect);
            string stale = await issuer.Authorize(cancelled);
            await store.CancelLoginAsync();
            Check(!await store.LoginPendingAsync(), "cancelled login is no longer pending");
            await Refuses<InvalidOperationException>(() => store.CompleteLoginAsync(stale));
            var next = await store.BeginLoginAsync("work", redirect);
            await Refuses<InvalidDataException>(() => store.CompleteLoginAsync(stale));
            string denied = redirect.AbsoluteUri + "?state=" + Query(next.Query)["state"]
                + "&error=access_denied&error_description=private-token-echo";
            await Refuses<InvalidOperationException>(() => store.CompleteLoginAsync(denied));
            Check(!await store.LoginPendingAsync(), "matched denial consumes pending login");
            Check(issuer.Exchanges == 1 && (await store.AccountsAsync()).Single().SignedIn,
                "denied reauthorization preserves established credentials");
        }
        finally
        {
            app.Destroy();
        }
    }

    internal static async Task FailureAndPendingRetry(string root)
    {
        await using var issuer = new Issuer { Status = 503 };
        var app = new Application(TextWriter.Null);
        var secondApp = new Application(TextWriter.Null);
        try
        {
            string directory = Path.Combine(root, "auth-login-failed");
            issuer.StorePath = Path.Combine(directory, "credentials.json");
            var redirect = new Uri("http://127.0.0.1:43219/auth/callback");
            var store = new Credentials(app, directory, issuer.Address);
            var authorization = await store.BeginLoginAsync("work", redirect);
            await Refuses<HttpRequestException>(async () => await store.CompleteLoginAsync(await issuer.Authorize(authorization)));
            Check(await store.SelectedPlanAccountAsync() is null, "failed exchange never selects plan");
            var pending = (await store.AccountsAsync()).Single();
            Check(pending.Subject is null && !pending.SignedIn && !pending.PlanEnabled && pending.ClientID == "issued-1"
                && issuer.Exchanges == 1, "failed exchange retains only issued registration; no automatic code replay");
            await Refuses<InvalidDataException>(() => store.StoreAsync(new AuthRegistration("invalid",
                issuer.Address.GetLeftPart(UriPartial.Authority), null, "issued-invalid", null,
                "access", "refresh", "identity", DateTimeOffset.UtcNow.AddHours(1), "openid")));
            Check((await store.AccountsAsync()).Length == 1, "pending records cannot contain unbound tokens");
            app.Destroy();
            var reopened = new Credentials(secondApp, directory, issuer.Address);
            var retry = await reopened.BeginLoginAsync("work", redirect);
            Check(Query(retry.Query)["client_id"] == "issued-1"
                && Query(retry.Query)["ext_agent_host_id"] == Query(authorization.Query)["ext_agent_host_id"],
                "restart reuses stable host and pending issued client");
            issuer.Status = 200;
            issuer.WrongNonce = true;
            await Refuses<InvalidDataException>(async () => await reopened.CompleteLoginAsync(await issuer.Authorize(retry)));
            Check(!(await reopened.AccountsAsync()).Single().SignedIn, "wrong signed nonce never publishes login");
            Check(await reopened.SelectedPlanAccountAsync() is null, "invalid identity never publishes selection");
            issuer.WrongNonce = false;
            issuer.MissingIdentity = true;
            retry = await reopened.BeginLoginAsync("work", redirect);
            await Refuses<InvalidDataException>(async () => await reopened.CompleteLoginAsync(await issuer.Authorize(retry)));
            issuer.MissingIdentity = false;
            retry = await reopened.BeginLoginAsync("work", redirect);
            var signedIn = await reopened.CompleteLoginAsync(await issuer.Authorize(retry));
            Check(signedIn.SignedIn && issuer.Registrations == 1, "pending retry binds identity without bootstrapping again");
            string good = await File.ReadAllTextAsync(issuer.StorePath);
            issuer.Subject = "foreign-subject";
            retry = await reopened.BeginLoginAsync("work", redirect);
            await Refuses<InvalidDataException>(async () => await reopened.CompleteLoginAsync(await issuer.Authorize(retry)));
            Check(await File.ReadAllTextAsync(issuer.StorePath) == good, "reauthorization cannot rebind established subject");
        }
        finally
        {
            app.Destroy();
            secondApp.Destroy();
        }
    }

    internal static async Task StaleRegistrationAndDestruction(string root)
    {
        await using var issuer = new Issuer();
        var app = new Application(TextWriter.Null);
        var peerApp = new Application(TextWriter.Null);
        try
        {
            string directory = Path.Combine(root, "auth-login-stale");
            issuer.StorePath = Path.Combine(directory, "credentials.json");
            var redirect = new Uri("http://127.0.0.1:43219/auth/callback");
            var store = new Credentials(app, directory, issuer.Address);
            var peer = new Credentials(peerApp, directory, issuer.Address);
            var authorization = await store.BeginLoginAsync("work", redirect);
            await store.CompleteLoginAsync(await issuer.Authorize(authorization));
            authorization = await store.BeginLoginAsync("work", redirect);
            string callback = await issuer.Authorize(authorization);
            await peer.LogoutAsync("work");
            await Refuses<InvalidOperationException>(() => store.CompleteLoginAsync(callback));
            Check(issuer.Exchanges == 1 && !(await peer.AccountsAsync()).Single().SignedIn,
                "intervening owner's logout cannot be overwritten by stale login");
            issuer.Hold = true;
            authorization = await store.BeginLoginAsync("work", redirect);
            callback = await issuer.Authorize(authorization);
            var completion = store.CompleteLoginAsync(callback);
            // Entered was already set by the first login; wait for the new POST.
            using var deadline = new CancellationTokenSource(TimeSpan.FromSeconds(10));
            while (Volatile.Read(ref issuer.Exchanges) < 2)
            {
                await Task.Delay(10, deadline.Token);
            }
            var destruction = Task.Run(app.Destroy);
            await Refuses<OperationCanceledException>(() => completion);
            await destruction.WaitAsync(deadline.Token);
            issuer.Release.TrySetResult();
            Check(store.parent is null && app.Credentials() is null && !(await peer.AccountsAsync()).Single().SignedIn,
                "destruction cancels exchange, joins/detaches owner, preserves signed-out state");
        }
        finally
        {
            issuer.Release.TrySetResult();
            app.Destroy();
            peerApp.Destroy();
        }
    }
}
