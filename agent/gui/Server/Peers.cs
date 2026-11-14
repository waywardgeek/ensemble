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

    private static partial void MapPeers(IEndpointRouteBuilder routes, Agent source, int port, string? publicOrigin, CancellationToken stopping)
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
            string? expected = RequestOrigin(context, port, publicOrigin);
            // Browsers omit Origin on same-origin GETs. The custom header and
            // checked Host still guard reads; public mutations require Origin.
            return expected is not null && MatchesOrigin(request, expected,
                required: request.Method != "GET" && expected == publicOrigin)
                && request.Headers["X-Ensemble-Peers"] == "1";
        }

        IResult Choices(HttpContext context)
        {
            if (!Guard(context)) return Results.StatusCode(StatusCodes.Status403Forbidden);
            // Reuse the same credential-free model projection as the main picker.
            var frame = MeterFrame(source);
            frame["funding_choices"] = FundingChoices(source);
            frame["primary_skills"] = new System.Text.Json.Nodes.JsonArray(
                (source.SkillManager()?.PrimaryChoices() ?? [])
                .Select(value => (System.Text.Json.Nodes.JsonNode)new System.Text.Json.Nodes.JsonObject
                {
                    ["name"] = value.Name, ["description"] = value.Description
                }).ToArray());
            return Results.Json(frame);
        }

        async Task<IResult> Change(HttpContext context, string operationName)
        {
            if (!Guard(context)) return Results.StatusCode(StatusCodes.Status403Forbidden);
            if (!context.Request.HasJsonContentType()) return Results.StatusCode(StatusCodes.Status415UnsupportedMediaType);
            bool create = operationName == "create";
            bool restore = operationName == "restore";
            bool funding = operationName == "funding";
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
                    bool offering = (create || restore || funding) && property.Name is "offering" or "preference";
                    bool allowed = offering || (!funding && property.Name == "name") || (create && property.Name is "model" or "skill");
                    if (!allowed || !seen.Add(property.Name)
                        || property.Value.ValueKind != (offering ? JsonValueKind.Object : JsonValueKind.String))
                    {
                        throw new InvalidDataException();
                    }
                }
                var ensemble = source.Parent();
                if (funding)
                {
                    if (seen.Count != 1 || !body.TryGetProperty("offering", out var selected))
                        throw new InvalidDataException();
                    var chosen = ensemble.ResolveModel(JsonCodec.ReadModelIdentity(selected));
                    var settings = ensemble.GlobalSettings()
                        ?? throw new InvalidOperationException("Global settings are unavailable.");
                    settings.RememberFunding(chosen);
                    return Results.Json(new { saved = true });
                }
                if (!body.TryGetProperty("name", out var nameValue) || string.IsNullOrWhiteSpace(nameValue.GetString()))
                {
                    throw new InvalidDataException();
                }
                string name = nameValue.GetString()!;
                if (create || restore)
                {
                    string model = body.TryGetProperty("model", out var modelValue) ? modelValue.GetString()! : "";
                    string skill = body.TryGetProperty("skill", out var skillValue) ? skillValue.GetString()! : "";
                    ModelIdentity? offering = body.TryGetProperty("offering", out var identity)
                        ? JsonCodec.ReadModelIdentity(identity) : null;
                    if (body.TryGetProperty("preference", out var preference))
                    {
                        if (offering is not null || model.Length != 0)
                            throw new ArgumentException("Conflicting model selection forms.");
                        offering = ResolvePreferredModel(source, preference).Identity();
                    }
                    var peer = restore
                        ? await ensemble.RestorePeerAsync(source, name, offering, token, waitForBrowser: true)
                        : await ensemble.CreatePeerAsync(source, name, model, skill, token,
                            waitForBrowser: true, offering: offering);
                    return Results.Json(new { name, address = peer.Address(), created = create, restored = restore });
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
                    // Retire restart registration as well as live membership;
                    // saved child History/artifacts remain, but cannot auto-restore.
                    await target.AgentParent()!.DeleteChildAsync(target.ChildID()!, token);
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
            catch (Exception error) when (error is JsonException or InvalidDataException or ArgumentException or FormatException)
            {
                return Results.BadRequest(new { error = "Invalid request, model, skill, or stored configuration." });
            }
            catch (Exception error) when (error is InvalidOperationException or IOException or UnauthorizedAccessException or AggregateException)
            {
                // Never reflect paths, provider diagnostics or credentials.
                string message = operationName switch
                {
                    "funding" => "Funding default was not saved. Check the configured funding source and settings storage. Running Agents are unchanged.",
                    "restore" => "Agent could not start. Check its model and funding in Settings, primary skill, and storage. An explicitly saved recovery choice is retained for retry.",
                    "delete" => "Agent deletion refused. Check storage and unfinished children; saved files are retained.",
                    _ => "Agent creation refused. Check the name, model funding, primary skill, and storage."
                };
                return Results.Conflict(new { error = message });
            }
        }

        IResult Saved(HttpContext context)
        {
            if (!Guard(context)) return Results.StatusCode(StatusCodes.Status403Forbidden);
            try
            {
                var root = source.Parent();
                var saved = root.SavedPeers(source).Where(row => root.FindAgent(row.Name) is null)
                    .Select(row => new { name = row.Name, model = row.Model,
                        message = "Not running. Retry the saved configuration, or choose a model to repair it." });
                return Results.Json(saved);
            }
            catch (Exception error) when (error is IOException or InvalidDataException or UnauthorizedAccessException)
            {
                return Results.Conflict(new { error = "The saved-agent registry could not be read. Files have been preserved." });
            }
        }

        routes.MapGet("/peers/saved", (Func<HttpContext, IResult>)Saved);
        routes.MapPost("/peers/restore", (Func<HttpContext, Task<IResult>>)(context => Change(context, "restore")));
        routes.MapPost("/models/funding", (Func<HttpContext, Task<IResult>>)(context => Change(context, "funding")));
        routes.MapGet("/peers/choices", (Func<HttpContext, IResult>)Choices);
        routes.MapPost("/peers/create", (Func<HttpContext, Task<IResult>>)(context => Change(context, "create")));
        routes.MapPost("/peers/delete", (Func<HttpContext, Task<IResult>>)(context => Change(context, "delete")));
    }
}
