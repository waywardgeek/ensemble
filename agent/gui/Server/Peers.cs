using System.Collections.Immutable;
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
            frame["sandbox_root"] = source.Sandbox().Root();
            var imports = new System.Text.Json.Nodes.JsonArray();
            var inherited = source.Configuration().ProcessReadPaths ?? [];
            foreach (string path in inherited)
                imports.Add(new System.Text.Json.Nodes.JsonObject {
                    ["label"] = "Inherited import", ["path"] = path, ["available"] = true, ["selected"] = true });
            // No filesystem discovery outside a confined source. Its existing
            // grants are the only suggestions; creation still validates subsets.
            if (source.Sandbox().Root().Length == 0)
            {
                bool available;
                try { available = source.Sandbox().IsDirectory("/usr/local/go"); }
                catch (Exception error) when (error is IOException or UnauthorizedAccessException) { available = false; }
                imports.Add(new System.Text.Json.Nodes.JsonObject {
                    ["label"] = "Go toolchain", ["path"] = "/usr/local/go", ["available"] = available, ["selected"] = false });
            }
            frame["process_read_suggestions"] = imports;
            return Results.Json(frame);
        }

        IResult Directories(HttpContext context)
        {
            if (!Guard(context)) return Results.StatusCode(StatusCodes.Status403Forbidden);
            try
            {
                var sandbox = source.Sandbox();
                string requested = context.Request.Query["path"].ToString();
                if (requested.Length > 4096) return Results.BadRequest();
                if (requested.Length == 0) requested = sandbox.Root().Length == 0 ? Path.GetTempPath() : sandbox.Root();
                if (!Path.IsPathFullyQualified(requested)) return Results.BadRequest();
                string path = sandbox.Resolve(requested);
                string? parent = Path.GetDirectoryName(path);
                if (parent is not null)
                {
                    try { parent = sandbox.Resolve(parent); }
                    catch (IOException) { parent = null; }
                }
                var folders = new List<string>();
                foreach (string name in sandbox.ReadDirectory(path))
                {
                    try
                    {
                        if (sandbox.IsDirectory(Path.Combine(path, name))) folders.Add(name);
                    }
                    catch (Exception error) when (error is IOException or UnauthorizedAccessException) { }
                    if (folders.Count == 201) break;
                }
                return Results.Json(new { path, parent, directories = folders.Take(200), truncated = folders.Count > 200 });
            }
            catch (Exception error) when (error is IOException or UnauthorizedAccessException or ArgumentException)
            {
                return Results.BadRequest(new { error = "Directory is unavailable or outside the permitted boundary." });
            }
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
                    bool allowed = offering || (!funding && property.Name == "name") || (create && property.Name is "model" or "skill" or "sandbox_root" or "process_read_paths");
                    if (!allowed || !seen.Add(property.Name)
                        || property.Value.ValueKind != (offering ? JsonValueKind.Object : property.Name == "process_read_paths" ? JsonValueKind.Array : JsonValueKind.String))
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
                    ImmutableArray<string>? reads = null;
                    if (body.TryGetProperty("process_read_paths", out var imports))
                    {
                        if (imports.GetArrayLength() > 64 || imports.EnumerateArray().Any(value => value.ValueKind != JsonValueKind.String))
                            throw new InvalidDataException();
                        reads = imports.EnumerateArray().Select(value => value.GetString()!).ToImmutableArray();
                    }
                    var peer = restore
                        ? await ensemble.RestorePeerAsync(source, name, offering, token, waitForBrowser: true)
                        : await ensemble.CreatePeerAsync(source, name, model, skill, token,
                            waitForBrowser: true, offering: offering,
                            sandboxRoot: body.TryGetProperty("sandbox_root", out var workspace) ? workspace.GetString() : null,
                            processReadPaths: reads);
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

        routes.MapGet("/peers/directories", (Func<HttpContext, IResult>)Directories);
        routes.MapGet("/peers/saved", (Func<HttpContext, IResult>)Saved);
        routes.MapPost("/peers/restore", (Func<HttpContext, Task<IResult>>)(context => Change(context, "restore")));
        routes.MapPost("/models/funding", (Func<HttpContext, Task<IResult>>)(context => Change(context, "funding")));
        routes.MapGet("/peers/choices", (Func<HttpContext, IResult>)Choices);
        routes.MapPost("/peers/create", (Func<HttpContext, Task<IResult>>)(context => Change(context, "create")));
        routes.MapPost("/peers/delete", (Func<HttpContext, Task<IResult>>)(context => Change(context, "delete")));
    }
}
