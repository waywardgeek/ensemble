using System.Collections.Immutable;
using System.Text.Json;
using Microsoft.AspNetCore.Routing;

namespace EnsembleGui;

public static partial class Gui
{
    private static partial void MapRenderers(IEndpointRouteBuilder routes, RendererStore? store,
        int port, string? publicOrigin, CancellationToken stopping)
    {
        // The query selects a shipped output type, never a filesystem path.
        static string? Renderer(HttpContext context)
        {
            var values = context.Request.Query["renderer"];
            if (values.Count == 0) return "list_directory"; // Existing clients.
            if (values.Count != 1) return null;
            return values[0] is "list_directory" or "read_file" or "search_files"
                or "write_file" or "edit_file" or "tool_default" ? values[0] : null;
        }
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

        // Read only this build's explicit embedded source. Discovery does not
        // stage a version, change acceptance, or expose a filesystem directory.
        routes.MapGet("/renderers/shipped", (HttpContext context) =>
        {
            if (!Allowed(context, false)) return Results.StatusCode(403);
            string? renderer = Renderer(context);
            if (renderer is null) return Refusal();
            string skill = renderer is "write_file" or "edit_file" ? "code-tools" : "file-reading";
            var assembly = typeof(Gui).Assembly;
            string? Source(string resource)
            {
                using var stream = assembly.GetManifestResourceStream(resource);
                if (stream is null) return null;
                using var reader = new StreamReader(stream);
                return reader.ReadToEnd();
            }
            if (renderer == "tool_default")
            {
                const string defaultEntry = "shared/tool-default.js";
                string? source = Source("EnsembleGui.Assets/skills/" + defaultEntry);
                return source is null ? Results.NotFound() : Results.Json(new
                {
                    entry = defaultEntry, files = new Dictionary<string, string> { [defaultEntry] = source }
                });
            }
            string? manual = Source("EnsembleGui.RendererSkill/" + skill);
            if (manual is null) return Results.NotFound();
            var definition = Ensemble.Skills.Parse(manual, "embedded " + skill);
            var association = definition.Renderers.Single(value => value.Name == renderer);
            // Retain the skill-relative directory layout so ordinary module
            // imports resolve inside this immutable package, including helpers.
            string entry = renderer == "list_directory" ? association.Entry : skill + "/" + association.Entry;
            var files = new Dictionary<string, string>();
            string? code = Source("EnsembleGui.Assets/skills/" + skill + "/" + association.Entry);
            if (code is null) return Results.NotFound();
            files.Add(entry, code);
            if (renderer is "read_file" or "write_file" or "edit_file")
                foreach (string helper in new[] { "file-excerpt.js", "source-code.js" })
                {
                    string? source = Source("EnsembleGui.Assets/skills/shared/" + helper);
                    if (source is null) return Results.NotFound();
                    files.Add("shared/" + helper, source);
                }
            return Results.Json(new { entry, files });
        });

        routes.MapGet("/renderers/selection", (HttpContext context) =>
        {
            if (!Allowed(context, false)) return Results.StatusCode(403);
            string? renderer = Renderer(context);
            if (renderer is null) return Refusal();
            if (store is null) return Results.Json(new { available = false });
            try { return Results.Json(new { available = true, selection = store.Selection(renderer) }); }
            catch (Exception error) when (ExpectedFailure(error)) { return Refusal(); }
        });
        routes.MapGet("/renderers/packages/{version}", (HttpContext context, string version) =>
        {
            if (!Allowed(context, false)) return Results.StatusCode(403);
            string? renderer = Renderer(context);
            if (renderer is null) return Refusal();
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
            string? renderer = Renderer(context);
            if (renderer is null) return Refusal();
            if (store is null) return Results.NotFound();
            try
            {
                var package = store.Read(version);
                if (package.Renderer is not ("list_directory" or "read_file" or "search_files" or "write_file" or "edit_file" or "tool_default")
                    || !package.Files.TryGetValue(file, out string? source))
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
                string? renderer = Renderer(context);
                if (renderer is null) return Refusal();
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
