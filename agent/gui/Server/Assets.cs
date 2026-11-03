using Microsoft.AspNetCore.Routing;
using System.Security.Cryptography;
using System.Text;

namespace EnsembleGui;

public static partial class Gui
{
    public static partial void MapAssets(IEndpointRouteBuilder routes, string? directory)
    {
        // Freeze this host's fixed public asset set before accepting requests.
        // A version URL must never change bytes, including with a disk override.
        var assembly = typeof(Gui).Assembly;
        const string prefix = "EnsembleGui.Assets/";
        string? disk = directory is null ? null : Path.GetFullPath(directory);
        if (disk is not null && !Directory.Exists(disk))
        {
            throw new DirectoryNotFoundException("GUI asset override directory does not exist.");
        }
        var assets = new SortedDictionary<string, byte[]>(StringComparer.Ordinal);
        foreach (string resource in assembly.GetManifestResourceNames().Order(StringComparer.Ordinal))
        {
            if (!resource.StartsWith(prefix, StringComparison.Ordinal)) continue;
            string name = resource[prefix.Length..];
            if (disk is not null)
            {
                string file = Path.Combine(disk, name);
                if (File.Exists(file)) assets.Add(name, File.ReadAllBytes(file));
            }
            else
            {
                using Stream stream = assembly.GetManifestResourceStream(resource)
                    ?? throw new InvalidOperationException("Embedded GUI asset is missing.");
                using var bytes = new MemoryStream();
                stream.CopyTo(bytes);
                assets.Add(name, bytes.ToArray());
            }
        }
        // One content fingerprint covers HTML and every static/dynamic dependency.
        // Paths, rather than query parameters, also work with caches ignoring queries.
        string manifest = string.Join("\n", assets.Select(asset =>
            asset.Key + ":" + Convert.ToHexString(SHA256.HashData(asset.Value))));
        string version = Convert.ToHexString(SHA256.HashData(Encoding.UTF8.GetBytes(manifest))).ToLowerInvariant();
        string versionPath = "_assets/" + version + "/";
        foreach (var asset in assets)
        {
            string name = asset.Key;
            byte[] content = asset.Value;
            string extension = Path.GetExtension(name);
            bool html = extension == ".html";
            string contentType = extension switch
            {
                ".html" => "text/html; charset=utf-8",
                ".js" => "text/javascript; charset=utf-8",
                ".css" => "text/css; charset=utf-8",
                ".png" => "image/png",
                ".svg" => "image/svg+xml",
                _ => "application/octet-stream"
            };
            if (html)
            {
                string page = Encoding.UTF8.GetString(content);
                // Only rewrite exact local references to shipped assets. Fragment
                // links, API routes, form values and user content are not rewritten.
                foreach (string target in assets.Keys.Where(target => Path.GetExtension(target) != ".html"))
                {
                    page = page.Replace("\"./" + target + "\"", "\"./" + versionPath + target + "\"", StringComparison.Ordinal);
                }
                content = Encoding.UTF8.GetBytes(page);
            }
            IResult Read(HttpContext context, bool versioned)
            {
                context.Response.Headers.CacheControl = versioned
                    ? "public, max-age=31536000, immutable" : "no-store";
                if (name == "index.html" && !context.Request.Path.Value!.EndsWith('/')
                    && !context.Request.Path.Value.EndsWith("/index.html", StringComparison.Ordinal))
                {
                    return Results.Redirect(context.Request.PathBase + context.Request.Path + "/" + context.Request.QueryString);
                }
                return Results.Bytes(content, contentType);
            }
            routes.MapGet("/" + name, (HttpContext context) => Read(context, false));
            if (name == "index.html") routes.MapGet("/", (HttpContext context) => Read(context, false));
            if (!html) routes.MapGet("/" + versionPath + name, (HttpContext context) => Read(context, true));
        }
    }
}
