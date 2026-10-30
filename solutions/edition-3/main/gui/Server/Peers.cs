using System.Net;
using System.Text.Json;
using Ensemble;
using Microsoft.AspNetCore.Routing;

namespace EnsembleGui;

public static partial class Gui
{
    private static partial bool CanDeleteAgent(Agent source, Agent target)
    {
        for (Agent? owner = source; owner is not null; owner = owner.AgentParent())
        {
            if (ReferenceEquals(owner, target)) return false;
        }
        return true;
    }

    private static partial void MapPeers(IEndpointRouteBuilder routes, Agent source, int port,
        CancellationToken stopping)
    {
        // Trusted host UI authority, not a model tool. The mounted Agent is an
        // explicit source of launch defaults, not a retained defaults owner.
        // No CORS is enabled. The custom header excludes simple cross-site
        // requests; Origin/Host checks also defend against DNS rebinding.
        bool Guard(HttpContext context)
        {
            context.Response.Headers.CacheControl = "no-store";
            context.Response.Headers["X-Content-Type-Options"] = "nosniff";
            var request = context.Request;
            bool host = request.Host.Host is "localhost" or "127.0.0.1" or "::1" or "[::1]";
            bool remote = context.Connection.RemoteIpAddress is { } ip && IPAddress.IsLoopback(ip);
            string expected = $"{request.Scheme}://{request.Host}";
            bool origin = !request.Headers.ContainsKey("Origin")
                || string.Equals(request.Headers.Origin.ToString(), expected, StringComparison.Ordinal);
            return host && remote && request.Host.Port == port && origin
                && request.Headers["X-Ensemble-Peers"] == "1";
        }

        async Task<IResult> Change(HttpContext context, bool create)
        {
            if (!Guard(context)) return Results.StatusCode(StatusCodes.Status403Forbidden);
            if (!context.Request.HasJsonContentType()) return Results.StatusCode(StatusCodes.Status415UnsupportedMediaType);
            // Both abort and mount shutdown reach the existing peer lifetime.
            // Once deletion closes admission, it ignores caller cancellation
            // and joins to a definite outcome rather than leaving half a delete.
            using var operation = CancellationTokenSource.CreateLinkedTokenSource(
                context.RequestAborted, stopping);
            var token = operation.Token;
            try
            {
                const int limit = 4096;
                if (context.Request.ContentLength > limit) throw new InvalidDataException();
                using var bytes = new MemoryStream();
                byte[] buffer = new byte[1024];
                while (true)
                {
                    int count = await context.Request.Body.ReadAsync(buffer, token);
                    if (count == 0) break;
                    if (bytes.Length + count > limit) throw new InvalidDataException();
                    bytes.Write(buffer, 0, count);
                }
                using var document = JsonDocument.Parse(bytes.ToArray(), new JsonDocumentOptions { MaxDepth = 4 });
                var body = document.RootElement;
                if (body.ValueKind != JsonValueKind.Object) throw new InvalidDataException();
                var seen = new HashSet<string>(StringComparer.Ordinal);
                foreach (var property in body.EnumerateObject())
                {
                    bool allowed = property.Name == "name" || (create && property.Name is "model" or "skill");
                    if (!allowed || !seen.Add(property.Name) || property.Value.ValueKind != JsonValueKind.String)
                    {
                        throw new InvalidDataException();
                    }
                }
                if (!body.TryGetProperty("name", out var nameValue) || string.IsNullOrWhiteSpace(nameValue.GetString()))
                {
                    throw new InvalidDataException();
                }
                string name = nameValue.GetString()!;
                var ensemble = source.Parent();
                if (create)
                {
                    string model = body.TryGetProperty("model", out var modelValue) ? modelValue.GetString()! : "";
                    string skill = body.TryGetProperty("skill", out var skillValue) ? skillValue.GetString()! : "";
                    var peer = await ensemble.CreatePeerAsync(source, name, model, skill, token, waitForBrowser: true);
                    return Results.Json(new { name, address = peer.Address(), created = true });
                }
                // A borrowed mount cannot destroy the Agent its caller promised
                // to keep alive until all mounted requests have drained.
                var target = ensemble.FindAgentAddress(name)
                    ?? throw new ArgumentException("No live Agent has that address.");
                if (!CanDeleteAgent(source, target))
                {
                    return Results.Conflict(new { error = "The mounted Agent and its owners cannot be deleted here." });
                }
                if (target.AgentParent() is not null)
                {
                    // Explicit human deletion cancels and joins the owned subtree,
                    // then detaches both ends of its relationships. It never edits
                    // the top-level launch registry or deletes saved child files.
                    token.ThrowIfCancellationRequested();
                    await target.ShutdownAsync().ConfigureAwait(ConfigureAwaitOptions.SuppressThrowing);
                    target.Destroy();
                }
                else
                {
                    await ensemble.DeletePeerAsync(source, target.Name!, token);
                }
                return Results.Json(new { name, deleted = true, files_retained = true });
            }
            catch (OperationCanceledException) when (token.IsCancellationRequested)
            {
                return Results.StatusCode(StatusCodes.Status503ServiceUnavailable);
            }
            catch (Exception error) when (error is JsonException or InvalidDataException or ArgumentException)
            {
                return Results.BadRequest(new { error = "Invalid peer request, skill, or stored configuration." });
            }
            catch (Exception error) when (error is InvalidOperationException or IOException or UnauthorizedAccessException or AggregateException)
            {
                // Never reflect paths, provider diagnostics or credentials.
                return Results.Conflict(new { error = "Peer operation refused. Check the name, startup skill, storage, and unfinished children." });
            }
        }

        routes.MapPost("/peers/create", (Func<HttpContext, Task<IResult>>)(context => Change(context, create: true)));
        routes.MapPost("/peers/delete", (Func<HttpContext, Task<IResult>>)(context => Change(context, create: false)));
    }
}
