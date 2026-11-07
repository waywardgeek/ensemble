using Ensemble;
using Microsoft.AspNetCore.Routing;
using System.Net;

namespace EnsembleGui;

public static partial class Gui
{
    private static partial void MapJobOutput(IEndpointRouteBuilder routes, Agent source, int port, string? publicOrigin, CancellationToken stopping)
    {
        // The live bounded window and the completed full download share the
        // same ownership boundary, but have distinct response contracts.
        foreach (bool window in new[] { false, true })
        {
            routes.MapGet(window ? "/jobs/window" : "/jobs/output", async (HttpContext context) =>
            {
                context.Response.Headers.CacheControl = "no-store";
                context.Response.Headers["X-Content-Type-Options"] = "nosniff";
                context.Response.Headers.ContentSecurityPolicy = "default-src 'none'; sandbox";
                var request = context.Request;
                string? expectedOrigin = RequestOrigin(context, port, publicOrigin);
                bool origin = expectedOrigin is not null
                    && MatchesOrigin(request, expectedOrigin);
                // Preserve custom-header and same-site requirements through the proxy.
                bool sameSite = request.Headers["Sec-Fetch-Site"].ToString() is "" or "none" or "same-origin";
                if (!origin || !sameSite
                    || request.Headers["X-Ensemble-Output"] != "1")
                {
                    return Results.StatusCode(StatusCodes.Status403Forbidden);
                }
                var query = request.Query;
                if (query.Keys.Any(key => key is not ("agent" or "handle" or "download"))
                    || query["agent"].Count != 1 || query["handle"].Count != 1
                    || (window && query.ContainsKey("download"))
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
                    if (!window && snapshot.Status == "running")
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
                    if (window)
                    {
                        // Snapshot.Bytes counts flushed publication. Exclude later
                        // output rather than mix offsets from two observations.
                        long end = snapshot.Bytes;
                        long start = Math.Max(0, end - 64 * 1024);
                        if (stream.Length < end) throw new EndOfStreamException("Job output changed while being read.");
                        stream.Position = start;
                        byte[] tail = new byte[checked((int)(end - start))];
                        await stream.ReadExactlyAsync(tail, linked.Token);
                        return Results.Json(new
                        {
                            handle, status = snapshot.Status, start, end,
                            rolled_out = start > 0, encoding = "base64", data = Convert.ToBase64String(tail),
                            origin_call_id = snapshot.OriginCallID, origin_tool = snapshot.OriginTool
                        });
                    }
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
}
