using Ensemble;
using Microsoft.AspNetCore.Routing;
using System.Net;

namespace EnsembleGui;

public static partial class Gui
{
    private static partial void MapJobOutput(IEndpointRouteBuilder routes, Agent source, int port, CancellationToken stopping)
    {
        routes.MapGet("/jobs/output", async (HttpContext context) =>
        {
            context.Response.Headers.CacheControl = "no-store";
            context.Response.Headers["X-Content-Type-Options"] = "nosniff";
            context.Response.Headers.ContentSecurityPolicy = "default-src 'none'; sandbox";
            var request = context.Request;
            string expectedOrigin = $"{request.Scheme}://{request.Host}";
            bool local = context.Connection.RemoteIpAddress is { } remote && IPAddress.IsLoopback(remote);
            bool host = request.Host.Host is "localhost" or "127.0.0.1" or "::1" or "[::1]";
            bool origin = !request.Headers.ContainsKey("Origin")
                || request.Headers.Origin.ToString() == expectedOrigin;
            // Custom-header fetch excludes simple cross-site requests. Origin and
            // loopback checks retain the existing local GUI boundary.
            bool sameSite = request.Headers["Sec-Fetch-Site"].ToString() is "" or "none" or "same-origin";
            if (!local || !host || request.Host.Port != port || !origin || !sameSite
                || request.Headers["X-Ensemble-Output"] != "1")
            {
                return Results.StatusCode(StatusCodes.Status403Forbidden);
            }
            var query = request.Query;
            if (query.Keys.Any(key => key is not ("agent" or "handle" or "download"))
                || query["agent"].Count != 1 || query["handle"].Count != 1
                || query["download"].Count > 1
                || (query.ContainsKey("download") && query["download"] != "1")
                || !int.TryParse(query["handle"], out int handle) || handle <= 0)
            {
                return Results.BadRequest(new { error = "Expected Agent identity and positive Job handle." });
            }
            try
            {
                string address = query["agent"].ToString();
                Agent? owner = address.Length == 0 ? source : source.Parent().FindAgentAddress(address);
                if (owner is null)
                {
                    return Results.NotFound();
                }
                // Find checks the owner's actual collection. Global handle
                // uniqueness alone is not permission to read another Agent's Job.
                var job = owner.Jobs().Find(handle);
                var snapshot = job.Snapshot();
                if (snapshot.Status == "running")
                {
                    return Results.Conflict(new { error = "Output is not complete while the Job is running." });
                }
                using var linked = CancellationTokenSource.CreateLinkedTokenSource(stopping, context.RequestAborted);
                linked.Token.ThrowIfCancellationRequested();
                // The path comes only from this live owned Job. Use the existing
                // nonblocking regular-file operation, not a browser path or a
                // broad unconfined OpenFile call that could wait on a FIFO.
                // Preserve that Sandbox's policy: unconfined means full access;
                // rooted resolution retains its existing admissible-link rules.
                // Ownership does not authenticate bytes against a hostile local
                // writer replacing this file. A metadata-then-open symlink walk
                // would not establish that stronger, unrequested guarantee.
                await using var stream = owner.Sandbox().OpenRegularRead(snapshot.OutputPath);
                long remaining = stream.Length;
                const long browserLimit = 16 * 1024 * 1024;
                if (remaining > browserLimit)
                {
                    return Results.Json(new { error = "Output exceeds the 16 MiB browser download limit. Use the owned workspace file locally." }, statusCode: 413);
                }
                context.Response.ContentType = "text/plain; charset=utf-8";
                context.Response.ContentLength = remaining;
                if (query.ContainsKey("download"))
                {
                    context.Response.Headers.ContentDisposition = $"attachment; filename=\"job-{handle}-output.txt\"";
                }
                // Bound browser allocation without truncating a result. This
                // independently opened stream never moves the Job observer's
                // cursor. Both client abort and mount shutdown cancel the full
                // copy before the mounted request is considered drained.
                byte[] buffer = new byte[64 * 1024];
                while (remaining > 0)
                {
                    int count = (int)Math.Min(buffer.Length, remaining);
                    int read = await stream.ReadAsync(buffer.AsMemory(0, count), linked.Token);
                    if (read == 0)
                    {
                        throw new EndOfStreamException("Job output changed while being read.");
                    }
                    await context.Response.Body.WriteAsync(buffer.AsMemory(0, read), linked.Token);
                    remaining -= read;
                }
                return Results.Empty;
            }
            catch (Exception error) when (!context.Response.HasStarted && error is
                (ArgumentException or IOException or UnauthorizedAccessException
                    or ObjectDisposedException or InvalidOperationException))
            {
                // Do not disclose paths or distinguish a foreign handle from one
                // destroyed while the request was being admitted.
                return Results.NotFound();
            }
        });
    }
}
