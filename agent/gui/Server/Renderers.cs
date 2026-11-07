using System.Collections.Immutable;
using System.Text.Json;
using Microsoft.AspNetCore.Routing;

namespace EnsembleGui;

public static partial class Gui
{
    private static partial void MapRenderers(IEndpointRouteBuilder routes, RendererStore? store,
        int port, string? publicOrigin, CancellationToken stopping)
    {
        // The bounded first experiment supports one real renderer, not a registry
        // of speculative renderer names. No request may supply a filesystem path.
        const string renderer = "list_directory";
        bool Allowed(HttpContext context, bool mutation)
        {
            context.Response.Headers.CacheControl = "no-store";
            context.Response.Headers["X-Content-Type-Options"] = "nosniff";
            string? origin = RequestOrigin(context, port, publicOrigin);
            return origin is not null && MatchesOrigin(context.Request, origin, required: mutation)
                && context.Request.Headers["Sec-Fetch-Site"].ToString() is "" or "none" or "same-origin"
                && (!mutation || context.Request.Headers["X-Ensemble-Renderer"] == "1");
        }
        static bool ExpectedFailure(Exception error) => error is ArgumentException or IOException
            or InvalidDataException or InvalidOperationException or JsonException or UnauthorizedAccessException;
        static IResult Refusal() => Results.BadRequest(new { error = "Renderer data is invalid, unavailable, or changed. Refresh before trying again." });

        routes.MapGet("/renderers/selection", (HttpContext context) =>
        {
            if (!Allowed(context, false)) return Results.StatusCode(403);
            if (store is null) return Results.Json(new { available = false });
            try { return Results.Json(new { available = true, selection = store.Selection(renderer) }); }
            catch (Exception error) when (ExpectedFailure(error)) { return Refusal(); }
        });
        routes.MapGet("/renderers/packages/{version}", (HttpContext context, string version) =>
        {
            if (!Allowed(context, false)) return Results.StatusCode(403);
            if (store is null) return Results.NotFound();
            try
            {
                var package = store.Read(version);
                return package.Renderer == renderer ? Results.Json(package) : Results.NotFound();
            }
            catch (Exception error) when (ExpectedFailure(error)) { return Refusal(); }
        });
        routes.MapGet("/renderers/modules/{version}/{**file}", (HttpContext context, string version, string file) =>
        {
            if (!Allowed(context, false)) return Results.StatusCode(403);
            if (store is null) return Results.NotFound();
            try
            {
                var package = store.Read(version);
                if (package.Renderer != renderer || !package.Files.TryGetValue(file, out string? source))
                    return Results.NotFound();
                // Validate the snapshot on every read; a changed file must never
                // be served under the identity of approved code.
                return Results.Text(source, "text/javascript; charset=utf-8");
            }
            catch (Exception error) when (ExpectedFailure(error)) { return Refusal(); }
        });
        foreach (string action in new[] { "stage", "accept", "rollback" })
        {
            routes.MapPost("/renderers/" + action, async (HttpContext context) =>
            {
                if (!Allowed(context, true)) return Results.StatusCode(403);
                if (store is null) return Results.NotFound();
                using var linked = CancellationTokenSource.CreateLinkedTokenSource(stopping, context.RequestAborted);
                try
                {
                    using var bytes = new MemoryStream();
                    var buffer = new byte[8192];
                    while (true)
                    {
                        int count = await context.Request.Body.ReadAsync(buffer, linked.Token);
                        if (count == 0) break;
                        if (bytes.Length + count > 1024 * 1024) return Results.StatusCode(413);
                        bytes.Write(buffer, 0, count);
                    }
                    using var document = JsonDocument.Parse(bytes.ToArray(), new JsonDocumentOptions { MaxDepth = 8 });
                    var body = document.RootElement;
                    static void Fields(JsonElement value, params string[] names)
                    {
                        if (value.ValueKind != JsonValueKind.Object) throw new InvalidDataException();
                        var seen = new HashSet<string>(StringComparer.Ordinal);
                        foreach (var field in value.EnumerateObject())
                            if (!names.Contains(field.Name, StringComparer.Ordinal) || !seen.Add(field.Name))
                                throw new InvalidDataException();
                        if (seen.Count != names.Length) throw new InvalidDataException();
                    }
                    if (action == "stage")
                    {
                        Fields(body, "entry", "files");
                        string entry = body.GetProperty("entry").GetString() ?? throw new InvalidDataException();
                        var files = ImmutableSortedDictionary.CreateBuilder<string, string>(StringComparer.Ordinal);
                        foreach (var file in body.GetProperty("files").EnumerateObject())
                        {
                            if (files.Count >= 64 || files.ContainsKey(file.Name)) throw new InvalidDataException();
                            files.Add(file.Name, file.Value.GetString() ?? throw new InvalidDataException());
                        }
                        string version = store.Stage(new(renderer, entry, files.ToImmutable()));
                        return Results.Json(new { version });
                    }
                    Fields(body, action == "accept" ? ["version", "expected"] : ["expected"]);
                    var expectedValue = body.GetProperty("expected");
                    string? expected = expectedValue.ValueKind == JsonValueKind.Null ? null : expectedValue.GetString();
                    var selection = action == "accept"
                        ? store.Accept(renderer, body.GetProperty("version").GetString() ?? throw new InvalidDataException(), expected)
                        : store.Rollback(renderer, expected);
                    return Results.Json(new { selection });
                }
                catch (Exception error) when (ExpectedFailure(error)) { return Refusal(); }
            });
        }
    }
}
