using Microsoft.AspNetCore.Routing;

namespace EnsembleGui;

public static partial class Gui
{
    public static partial void MapAssets(IEndpointRouteBuilder routes, string? directory)
    {
        // Enumerate the assembly's fixed public assets, not request-supplied
        // paths. A disk override changes their bytes, not the set of readable
        // host files. No arbitrary file-server authority is exposed here.
        var assembly = typeof(Gui).Assembly;
        const string prefix = "EnsembleGui.Assets/";
        string? disk = directory is null ? null : Path.GetFullPath(directory);
        if (disk is not null && !Directory.Exists(disk))
        {
            throw new DirectoryNotFoundException("GUI asset override directory does not exist.");
        }
        foreach (string resource in assembly.GetManifestResourceNames().Order(StringComparer.Ordinal))
        {
            if (!resource.StartsWith(prefix, StringComparison.Ordinal))
            {
                continue;
            }
            string name = resource[prefix.Length..];
            string contentType = Path.GetExtension(name) switch
            {
                ".html" => "text/html; charset=utf-8",
                ".js" => "text/javascript; charset=utf-8",
                ".css" => "text/css; charset=utf-8",
                _ => "application/octet-stream"
            };
            IResult Read(HttpContext context)
            {
                // Relative browser URLs need a directory base. Routing accepts
                // both spellings, but /assistant must become /assistant/ before
                // the browser resolves app.js, ws, or auth/status.
                if (name == "index.html" && !context.Request.Path.Value!.EndsWith('/')
                    && !context.Request.Path.Value.EndsWith("/index.html", StringComparison.Ordinal))
                {
                    return Results.Redirect(context.Request.PathBase + context.Request.Path + "/" + context.Request.QueryString);
                }
                if (disk is not null)
                {
                    string file = Path.Combine(disk, name);
                    return File.Exists(file) ? Results.File(file, contentType) : Results.NotFound();
                }
                // Each request owns a fresh stream. Results.Stream disposes it
                // after delivery, including when the caller disconnects.
                Stream stream = assembly.GetManifestResourceStream(resource)
                    ?? throw new InvalidOperationException("Embedded GUI asset is missing.");
                return Results.Stream(stream, contentType);
            }
            routes.MapGet("/" + name, Read);
            if (name == "index.html")
            {
                routes.MapGet("/", Read);
            }
        }
    }
}
