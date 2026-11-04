using Ensemble;
using Microsoft.AspNetCore.Routing;

namespace EnsembleGui;

public static partial class Gui
{
    private static partial void MapWorkProducts(IEndpointRouteBuilder routes, Agent source,
        int port, string? publicOrigin, CancellationToken stopping)
    {
        routes.MapGet("/work-products/download", async (HttpContext context) =>
        {
            context.Response.Headers.CacheControl = "no-store";
            context.Response.Headers["X-Content-Type-Options"] = "nosniff";
            context.Response.Headers.ContentSecurityPolicy = "default-src 'none'; sandbox";
            var request = context.Request;
            string? expectedOrigin = RequestOrigin(context, port, publicOrigin);
            bool sameSite = request.Headers["Sec-Fetch-Site"].ToString() is "" or "none" or "same-origin";
            if (expectedOrigin is null || !MatchesOrigin(request, expectedOrigin) || !sameSite
                || request.Headers["X-Ensemble-Work-Product"] != "1")
                return Results.StatusCode(StatusCodes.Status403Forbidden);
            var query = request.Query;
            if (query.Keys.Any(key => key is not ("agent" or "offer"))
                || query["agent"].Count != 1 || query["offer"].Count != 1
                || query["offer"].ToString().Length != 32)
                return Results.BadRequest(new { error = "Expected Agent identity and opaque offer ID." });
            try
            {
                string address = query["agent"].ToString();
                Agent? owner = address.Length == 0 ? source : source.Parent().FindAgentAddress(address);
                if (owner is null) return Results.NotFound();
                using var linked = CancellationTokenSource.CreateLinkedTokenSource(stopping, context.RequestAborted);
                context.Response.ContentType = "application/octet-stream";
                context.Response.Headers.ContentDisposition = "attachment; filename=\"work-product.bin\"";
                // No supplied path or inferred History grant reaches filesystem
                // access. The live Agent reopens and bounds this copy, owns its
                // input and cancellation, and joins it before Sandbox teardown.
                await owner.CopyWorkProductAsync(query["offer"].ToString(), context.Response.Body, linked.Token);
                return Results.Empty;
            }
            catch (OperationCanceledException)
            {
                context.Abort();
                return Results.Empty;
            }
            catch (Exception error) when (error is ArgumentException or IOException
                or UnauthorizedAccessException or InvalidOperationException or KeyNotFoundException)
            {
                if (context.Response.HasStarted)
                {
                    // Partial bytes are never a successful truncated download.
                    context.Abort();
                    return Results.Empty;
                }
                context.Response.Headers.Remove("Content-Disposition");
                return Results.NotFound(new { error = "Offer expired, unavailable, or exceeds the 64 MiB transfer limit." });
            }
        });
    }
}
