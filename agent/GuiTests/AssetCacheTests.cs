using System.Net;
using System.Text.RegularExpressions;
using EnsembleGui;

static class AssetCacheTests
{
    private static void Check(bool condition, string message)
    {
        if (!condition) throw new InvalidOperationException(message);
    }

    public static async Task RunAsync()
    {
        string directory = Directory.CreateTempSubdirectory("ensemble-asset-cache-").FullName;
        using var deadline = new CancellationTokenSource(TimeSpan.FromSeconds(30));
        var cancellation = deadline.Token;
        var builder = WebApplication.CreateBuilder(Array.Empty<string>());
        builder.Logging.ClearProviders();
        builder.WebHost.UseUrls("http://127.0.0.1:0");
        await using var host = builder.Build();
        Gui.MapAssets(host);
        Gui.MapAssets(host.MapGroup("/mounted"));
        string page = "<script src=\"./app.js\"></script><link href=\"./style.css\" rel=\"stylesheet\">";
        await File.WriteAllTextAsync(Path.Combine(directory, "index.html"), page, cancellation);
        await File.WriteAllTextAsync(Path.Combine(directory, "app.js"), "FIRST", cancellation);
        await File.WriteAllTextAsync(Path.Combine(directory, "style.css"), "body{}", cancellation);
        Gui.MapAssets(host.MapGroup("/first"), directory);
        await File.WriteAllTextAsync(Path.Combine(directory, "app.js"), "SECOND", cancellation);
        Gui.MapAssets(host.MapGroup("/second"), directory);
        // Later edits must not mutate either published immutable snapshot.
        await File.WriteAllTextAsync(Path.Combine(directory, "app.js"), "UNPUBLISHED", cancellation);
        await host.StartAsync(cancellation);
        try
        {
            using var client = new HttpClient { BaseAddress = new Uri(host.Urls.Single()) };
            async Task<string> Script(string path)
            {
                using var response = await client.GetAsync(path, cancellation);
                response.EnsureSuccessStatusCode();
                Check(response.Headers.CacheControl?.NoStore == true, "HTML must be no-store");
                string html = await response.Content.ReadAsStringAsync(cancellation);
                var match = Regex.Match(html, "\\./_assets/[a-f0-9]{64}/app\\.js");
                Check(match.Success, "Missing versioned relative script URL");
                return match.Value[2..];
            }
            string root = await Script("/");
            string mounted = await Script("/mounted/");
            string rendererPath = mounted.Replace("app.js", "skills/file-reading/renderers/list-directory.js");
            using var renderer = await client.GetAsync("/mounted/" + rendererPath, cancellation);
            Check(renderer.IsSuccessStatusCode && renderer.Content.Headers.ContentType?.MediaType == "text/javascript",
                "Skill renderer must be served as JavaScript in the same immutable asset version");
            Check((await client.GetStringAsync("/mounted/", cancellation)).Contains("./" + rendererPath),
                "Skill renderer script reference must use the mounted asset version");
            Check(root == mounted, "Identical snapshots must have identical versions");
            string first = await Script("/first/");
            string second = await Script("/second/");
            Check(first != second, "Content change did not change asset version");
            using var asset = await client.GetAsync("/first/" + first, cancellation);
            Check(await asset.Content.ReadAsStringAsync(cancellation) == "FIRST", "Snapshot bytes changed");
            Check(asset.Headers.CacheControl?.Public == true &&
                asset.Headers.CacheControl.MaxAge == TimeSpan.FromDays(365) &&
                asset.Headers.CacheControl.Extensions.Any(item => item.Name == "immutable"), "Versioned cache policy missing");
            Check(await client.GetStringAsync("/second/" + second, cancellation) == "SECOND", "Second snapshot wrong");
            using var old = await client.GetAsync("/second/" + first, cancellation);
            Check(old.StatusCode == HttpStatusCode.NotFound, "Old version silently served new bytes");
            using var alias = await client.GetAsync("/first/app.js", cancellation);
            Check(alias.Headers.CacheControl?.NoStore == true, "Legacy alias may be cached");
            Check(await alias.Content.ReadAsStringAsync(cancellation) == "FIRST", "Legacy alias escaped snapshot");
            using var module = await client.GetAsync("/mounted/" + mounted.Replace("app.js", "voice-session.js"), cancellation);
            Check(module.IsSuccessStatusCode, "Nested module missing from same version directory");
            Check(!(await client.GetStringAsync("/mounted/", cancellation)).Contains("src=\"./app.js\""), "Unversioned app reference remains");
            Console.WriteLine("PASS asset cache: root/mount, no-store HTML/aliases, immutable snapshot, changed versions, old-version refusal, nested module");
        }
        finally { await host.StopAsync(CancellationToken.None); }
    }
}
