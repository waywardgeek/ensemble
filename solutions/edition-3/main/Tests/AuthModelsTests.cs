using System.Net;
using System.Net.Sockets;
using System.Text;
using Ensemble;
using Application = Ensemble.Ensemble;

internal static class AuthModelsTests
{
    private static void Check(bool condition, string message)
    {
        if (!condition)
        {
            throw new Exception(message);
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
            Check(!error.ToString().Contains("private-catalog-body", StringComparison.Ordinal), "safe diagnostics");
            return;
        }
        throw new Exception("Expected " + typeof(T).Name);
    }

    // A real one-request-at-a-time HTTP boundary; no replacement of credential
    // reads, locking, request construction, cancellation, or response parsing.
    private sealed class Server : IDisposable
    {
        internal readonly HttpListener Listener = new();
        internal readonly string Address;
        internal Server()
        {
            var reservation = new TcpListener(IPAddress.Loopback, 0);
            reservation.Start();
            int port = ((IPEndPoint)reservation.LocalEndpoint).Port;
            reservation.Stop();
            Address = $"http://127.0.0.1:{port}";
            Listener.Prefixes.Add(Address + "/");
            Listener.Start();
        }
        internal async Task Reply(string bearer, string body, int status = 200)
        {
            var context = await Listener.GetContextAsync().WaitAsync(TimeSpan.FromSeconds(10));
            try
            {
                Check(context.Request.HttpMethod == "GET" && context.Request.RawUrl == "/v1/models", "catalog route");
                Check(context.Request.Headers["Authorization"] == "Bearer " + bearer, "selected account bearer");
                Check(context.Request.Headers["Cookie"] is null && !context.Request.HasEntityBody, "no cookies/body");
                context.Response.StatusCode = status;
                if (status == 307)
                {
                    context.Response.RedirectLocation = Address + "/forbidden-redirect";
                }
                context.Response.ContentType = "application/json";
                byte[] bytes = Encoding.UTF8.GetBytes(body);
                context.Response.ContentLength64 = bytes.Length;
                await context.Response.OutputStream.WriteAsync(bytes);
            }
            finally
            {
                context.Response.Close();
            }
        }
        public void Dispose() => Listener.Close();
    }

    private static AuthRegistration Seed(string label, string token) =>
        new(label, "https://auth.openai.com", "subject-" + label, "client-" + label,
            "shared@example.invalid", token, "refresh-" + label, "id-" + label,
            DateTimeOffset.UtcNow.AddHours(1), "openid chatgpt.tokens.use.direct");

    internal static async Task Catalog(string directory)
    {
        var app = new Application(TextWriter.Null);
        using var server = new Server();
        try
        {
            var credentials = new Credentials(app, directory);
            await credentials.StoreAsync(Seed("one", "bearer-one"));
            await credentials.StoreAsync(Seed("two", "bearer-two"));
            const string list = """
                {"models":[
                  {"slug":"z-unknown","display_name":"Visible Z","visibility":"list"},
                  {"slug":"hidden","visibility":"hidden"},
                  {"slug":"a-unknown","display_name":"Visible A","visibility":"list"},
                  {"slug":"not-listed"}]}
                """;
            var reply = server.Reply("bearer-one", list);
            var first = await credentials.ModelsAsync("one", server.Address);
            await reply;
            Check(first.Select(m => m.Slug).SequenceEqual(new[] { "z-unknown", "a-unknown" })
                && first[0].DisplayName == "Visible Z", "filter visibility, retain server order and public labels");
            reply = server.Reply("bearer-two", "{\"models\":[]}");
            var second = await credentials.ModelsAsync("two", server.Address + "/v1");
            await reply;
            Check(second.IsEmpty && first.Length == 2, "empty is valid; prior result remains immutable");
            await credentials.StoreAsync(Seed("one", "rotated-one"));
            reply = server.Reply("rotated-one", list);
            var third = await credentials.ModelsAsync("one", server.Address + "/v1/responses");
            await reply;
            Check(third.SequenceEqual(first), "fresh disk token and fresh catalog each call");
            Check(!Engine.LookupModel(Vendor.OpenAI, "z-unknown", out _),
                "discovery never promotes unknown names into capabilities");
        }
        finally
        {
            app.Destroy();
        }
    }

    internal static async Task Refusals(string directory)
    {
        var app = new Application(TextWriter.Null);
        using var server = new Server();
        try
        {
            var credentials = new Credentials(app, directory);
            await credentials.StoreAsync(Seed("one", "bearer-one"));
            foreach (string body in new[]
            {
                "{\"data\":[]}", "{\"models\":null}", "{\"models\":[null]}",
                "{\"models\":[{\"visibility\":\"list\",\"slug\":\"x\"}]}",
                "{\"models\":[{\"visibility\":3}]}",
                "{\"models\":[],\"models\":[]}",
                "{\"models\":[{\"visibility\":\"list\",\"slug\":\"x\",\"display_name\":\"X\"},{\"visibility\":\"list\",\"slug\":\"x\",\"display_name\":\"Y\"}]}",
                "private-catalog-body"
            })
            {
                var reply = server.Reply("bearer-one", body);
                await Refuses<InvalidDataException>(() => credentials.ModelsAsync("one", server.Address));
                await reply;
            }
            foreach (int status in new[] { 401, 403, 429, 500, 307 })
            {
                var reply = server.Reply("bearer-one", "private-catalog-body", status);
                await Refuses<HttpRequestException>(() => credentials.ModelsAsync("one", server.Address));
                await reply;
            }
            Check((await credentials.AccountsAsync()).Single().SignedIn, "catalog failures never erase credentials");
            // Each unsafe endpoint refuses before any credential or network work.
            foreach (string address in new[] { "relative", "http://example.invalid", "http://localhost:1234", "https://user@example.invalid", server.Address + "?key=private", server.Address + "#fragment" })
            {
                await Refuses<ArgumentException>(() => credentials.ModelsAsync("one", address));
            }
            await credentials.LogoutAsync("one");
            await Refuses<InvalidOperationException>(() => credentials.ModelsAsync("one", server.Address));
            await Refuses<InvalidOperationException>(() => credentials.ModelsAsync("missing", server.Address));
        }
        finally
        {
            app.Destroy();
        }
    }

    internal static async Task Lifetime(string directory)
    {
        var app = new Application(TextWriter.Null);
        using var server = new Server();
        try
        {
            var credentials = new Credentials(app, directory);
            await credentials.StoreAsync(Seed("one", "bearer-one"));
            using var cancellation = new CancellationTokenSource(TimeSpan.FromSeconds(15));
            Task request = credentials.ModelsAsync("one", server.Address, cancellation.Token);
            var context = await server.Listener.GetContextAsync().WaitAsync(TimeSpan.FromSeconds(10));
            cancellation.Cancel();
            await Refuses<OperationCanceledException>(() => request);
            context.Response.Close();
            // A cancelled request releases the operation gate and retains login.
            Check((await credentials.AccountsAsync()).Single().SignedIn, "cancel preserves account and releases gate");
            request = credentials.ModelsAsync("one", server.Address);
            context = await server.Listener.GetContextAsync().WaitAsync(TimeSpan.FromSeconds(10));
            await Task.Run(credentials.Destroy).WaitAsync(TimeSpan.FromSeconds(10));
            await Refuses<OperationCanceledException>(() => request);
            context.Response.Close();
            Check(app.Credentials() is null, "destruction cancels, joins and detaches catalog work");
        }
        finally
        {
            app.Destroy();
        }
    }
}
