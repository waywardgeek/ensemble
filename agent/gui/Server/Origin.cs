using System.Net;

namespace EnsembleGui;

public static partial class Gui
{
    private static partial string? ValidatePublicOrigin(string? origin)
    {
        if (origin is null) return null;
        if (!Uri.TryCreate(origin, UriKind.Absolute, out var uri)
            || uri.Scheme != "https" || uri.HostNameType != UriHostNameType.Dns
            || uri.IsLoopback || uri.UserInfo.Length != 0 || uri.AbsolutePath != "/"
            || uri.Query.Length != 0 || uri.Fragment.Length != 0
            || origin.Contains('\\') || origin.Any(char.IsWhiteSpace))
        {
            throw new ArgumentException("Public origin must be one HTTPS DNS origin, without credentials, path, query or fragment.");
        }
        return uri.GetLeftPart(UriPartial.Authority);
    }

    private static partial string? RequestOrigin(HttpContext context, int port, string? publicOrigin)
    {
        // Both native local access and the authenticated proxy arrive over
        // loopback. Never use X-Forwarded-Host/Proto to widen this boundary.
        if (context.Connection.RemoteIpAddress is not { } remote || !IPAddress.IsLoopback(remote)) return null;
        var request = context.Request;
        if (request.Scheme == "http" && request.Host.Port == port
            && request.Host.Host is "localhost" or "127.0.0.1" or "::1" or "[::1]")
        {
            return $"http://{request.Host}";
        }
        if (publicOrigin is null) return null;
        var configured = new Uri(publicOrigin, UriKind.Absolute);
        if (!string.Equals(request.Host.Host, configured.Host, StringComparison.OrdinalIgnoreCase)
            || (request.Host.Port ?? 443) != configured.Port) return null;
        return publicOrigin;
    }

    private static partial bool MatchesOrigin(HttpRequest request, string expected, bool required)
    {
        var origins = request.Headers.Origin;
        if (origins.Count == 0) return !required;
        return origins.Count == 1 && string.Equals(origins[0], expected, StringComparison.Ordinal);
    }
}
