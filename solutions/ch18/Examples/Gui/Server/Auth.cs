using Ensemble;
using System.Text.Json;

namespace EnsembleGui;

public static partial class Gui
{
    private static partial void MapAuth(WebApplication host, Agent agent, int port)
    {
        host.Map("/auth/{operation}", async (HttpContext context) =>
        {
            var request = context.Request;
            var response = context.Response;
            response.Headers.CacheControl = "no-store";
            response.Headers.Pragma = "no-cache";
            response.Headers["Referrer-Policy"] = "no-referrer";
            response.Headers["X-Content-Type-Options"] = "nosniff";
            response.Headers.ContentSecurityPolicy = "default-src 'none'; frame-ancestors 'none'";
            string operation = (string?)request.RouteValues["operation"] ?? "";
            bool callback = operation == "callback";

            // The host allowlist also runs globally. Check the exact listening
            // port here; never trust forwarded-host headers for local authority.
            string origin = $"http://{request.Host}";
            bool local = request.Scheme == "http" && request.Host.Port == port
                && request.Host.Host is "127.0.0.1" or "localhost" or "[::1]";
            bool sameOrigin = request.Headers.Origin.Count == 0
                || request.Headers.Origin.ToString() == origin;
            // Custom headers cannot be sent by a foreign HTML form or simple
            // cross-origin fetch. We provide no CORS/preflight permission. POST
            // additionally requires Origin, including rejecting the null origin.
            bool guarded = request.Headers["X-Ensemble-Auth"].ToString() == "1"
                && sameOrigin && (request.Method == "GET" || request.Headers.Origin.ToString() == origin);
            if (!local || (!callback && !guarded))
            {
                response.StatusCode = StatusCodes.Status403Forbidden;
                await response.WriteAsync("Account request refused.");
                return;
            }
            bool known = operation is "status" or "models" or "login" or "cancel" or "logout" or "signout" or "callback";
            string method = operation is "status" or "callback" ? "GET" : "POST";
            if (!known || request.Method != method)
            {
                response.StatusCode = known ? StatusCodes.Status405MethodNotAllowed : StatusCodes.Status404NotFound;
                await response.WriteAsync("Account operation unavailable.");
                return;
            }

            using var active = CancellationTokenSource.CreateLinkedTokenSource(
                request.HttpContext.RequestAborted, host.Lifetime.ApplicationStopping);
            // Bound a stalled issuer independently of the ten-minute user login
            // window. Cancellation consumes an admitted code exchange, not retries it.
            active.CancelAfter(TimeSpan.FromMinutes(2));
            var cancellation = active.Token;
            try
            {
                var credentials = agent.Parent().Credentials();
                if (operation == "status")
                {
                    var accounts = credentials is null ? [] : await credentials.AccountsAsync(cancellation);
                    bool pending = credentials is not null && await credentials.LoginPendingAsync(cancellation);
                    var config = agent.Configuration();
                    await response.WriteAsJsonAsync(new
                    {
                        enabled = credentials is not null,
                        billing = config.Billing == BillingMode.ChatGptPlan ? "chatgpt" : "api",
                        selected_account = config.Account,
                        pending,
                        accounts = accounts.Select(account => new
                        {
                            label = account.Label, email = account.Email,
                            signed_in = account.SignedIn, plan_enabled = account.PlanEnabled
                        })
                    }, cancellation);
                    return;
                }
                if (credentials is null)
                {
                    response.StatusCode = StatusCodes.Status409Conflict;
                    await response.WriteAsync("Account management requires an explicit credentials directory.", cancellation);
                    return;
                }
                if (callback)
                {
                    // Never log or reflect this URI. The credential owner checks
                    // exact endpoint, state, duplicates and registration identity.
                    string uri = origin + request.PathBase + request.Path + request.QueryString;
                    var account = await credentials.CompleteLoginAsync(uri, cancellation);
                    response.ContentType = "text/plain; charset=utf-8";
                    await response.WriteAsync(account.PlanEnabled
                        ? "Sign-in complete. ChatGPT plan funding is selected for the next launch using this store; this running session is unchanged. Close this window and refresh account status in Ensemble."
                        : "Signed in without direct-plan permission. Plan funding is selected for the next launch, but inference will require permission; there is no API-key fallback. This running session is unchanged. Close this window and refresh account status in Ensemble.", cancellation);
                    return;
                }

                if (!string.Equals(request.ContentType, "application/json", StringComparison.OrdinalIgnoreCase))
                {
                    response.StatusCode = StatusCodes.Status415UnsupportedMediaType;
                    await response.WriteAsync("Account operations require JSON.", cancellation);
                    return;
                }
                // Bound even chunked bodies before parsing. Do not buffer an
                // unbounded body just to discover that its account label is long.
                var bytes = new byte[4097];
                int count = 0;
                while (count < bytes.Length)
                {
                    int read = await request.Body.ReadAsync(bytes.AsMemory(count), cancellation);
                    if (read == 0)
                    {
                        break;
                    }
                    count += read;
                }
                if (count > 4096)
                {
                    response.StatusCode = StatusCodes.Status413PayloadTooLarge;
                    await response.WriteAsync("Account request is too large.", cancellation);
                    return;
                }
                using var body = JsonDocument.Parse(bytes.AsMemory(0, count));
                var root = body.RootElement;
                if (root.ValueKind != JsonValueKind.Object)
                {
                    throw new ArgumentException("Expected object.");
                }
                var names = new HashSet<string>(StringComparer.Ordinal);
                foreach (var property in root.EnumerateObject())
                {
                    bool allowed = operation != "cancel" && property.Name == "label"
                        || operation == "login" && property.Name == "consent";
                    if (!allowed || !names.Add(property.Name))
                    {
                        throw new ArgumentException("Unexpected or duplicate field.");
                    }
                }
                string label = "";
                if (operation != "cancel")
                {
                    if (!root.TryGetProperty("label", out var value) || value.ValueKind != JsonValueKind.String
                        || string.IsNullOrWhiteSpace(label = value.GetString() ?? "") || label.Length > 200)
                    {
                        throw new ArgumentException("Expected account label.");
                    }
                }
                bool consent = false;
                if (root.TryGetProperty("consent", out var consentValue))
                {
                    consent = consentValue.GetBoolean();
                }
                if (operation == "models")
                {
                    // The browser chooses an account, never a credential-bearing
                    // destination. Discovery does not mutate this Agent's model.
                    var config = agent.Configuration();
                    if (config.Vendor != Vendor.OpenAI)
                    {
                        throw new InvalidOperationException("Plan catalog requires an OpenAI endpoint.");
                    }
                    var models = await credentials.ModelsAsync(label, config.BaseUrl, cancellation);
                    await response.WriteAsJsonAsync(new
                    {
                        models = models.Select(model => new { slug = model.Slug, display_name = model.DisplayName })
                    }, cancellation);
                    return;
                }
                if (operation == "login")
                {
                    var redirect = new Uri($"http://127.0.0.1:{port}/auth/callback");
                    var url = await credentials.BeginLoginAsync(label, redirect, consent, cancellation);
                    // Private, request-local response only. Do not add this value
                    // to WS frames, traces, DOM links or browser local storage.
                    await response.WriteAsJsonAsync(new { url = url.AbsoluteUri }, cancellation);
                    return;
                }
                if (operation == "cancel")
                {
                    await credentials.CancelLoginAsync(cancellation);
                    await response.WriteAsJsonAsync(new { message = "Pending sign-in cancelled." }, cancellation);
                    return;
                }
                if (operation == "logout")
                {
                    await credentials.LogoutAsync(label, cancellation);
                    await response.WriteAsJsonAsync(new { message = "Local credentials cleared; remote revocation was not requested." }, cancellation);
                    return;
                }
                bool revoked = await credentials.SignOutAsync(label, cancellation);
                await response.WriteAsJsonAsync(new
                {
                    message = revoked ? "Local credentials cleared; remote revocation confirmed."
                        : "Local credentials cleared; remote revocation was not confirmed."
                }, cancellation);
            }
            catch (Exception error) when (error is ArgumentException or InvalidOperationException
                or IOException or InvalidDataException or JsonException or HttpRequestException or OperationCanceledException
                or UnauthorizedAccessException or System.Security.Cryptography.CryptographicException)
            {
                // Provider bodies and callback query strings may contain secrets.
                // Use a fixed diagnostic, never error.Message or exception logging.
                if (!response.HasStarted && !context.RequestAborted.IsCancellationRequested)
                {
                    response.StatusCode = StatusCodes.Status400BadRequest;
                    response.ContentType = "text/plain; charset=utf-8";
                    await response.WriteAsync(callback
                        ? "Sign-in was not completed. Return to Ensemble and refresh account status; cancel before retrying if a sign-in is still pending."
                        : "Account operation failed. Refresh status and cancel any pending sign-in before retrying.");
                }
            }
        });
    }
}
