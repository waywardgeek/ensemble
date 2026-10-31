using System.Net;
using System.Text.Json;

namespace Ensemble;

public sealed partial class Credentials
{
    private static partial int AccountIndex(AuthFile file, string label)
    {
        for (int index = 0; index < file.Registrations.Length; index++)
        {
            if (file.Registrations[index].Label == label)
            {
                return index;
            }
        }
        throw new InvalidOperationException("Unknown account label.");
    }

    public partial Task<string> GetBearerTokenAsync(string label, CancellationToken cancellation) =>
        GetBearerAfterRejectionAsync(label, null, cancellation);

    internal partial async Task<string> GetBearerAfterRejectionAsync(string label, string? rejectedBearer,
        CancellationToken cancellation)
    {
        ArgumentException.ThrowIfNullOrWhiteSpace(label);
        await gate.WaitAsync(cancellation).ConfigureAwait(false);
        try
        {
            ObjectDisposedException.ThrowIf(destroyed, this);
            using var active = CancellationTokenSource.CreateLinkedTokenSource(cancellation, lifetime.Token);
            cancellation = active.Token;
            return await ResolveBearerAsync(label, cancellation, rejectedBearer).ConfigureAwait(false);
        }
        finally
        {
            gate.Release();
        }
    }

    // Caller holds the local operation gate and links cancellation to lifetime.
    // The file lock covers renewal/publication, but not the subsequent catalog GET.
    private partial async Task<string> ResolveBearerAsync(string label, CancellationToken cancellation, string? rejectedBearer)
    {
        // The file lock deliberately spans the HTTP exchange AND publication.
        // A second process rereads the new generation, not a cached old token.
        using var held = await LockFileAsync(cancellation).ConfigureAwait(false);
        var file = ReadFile();
        int index = AccountIndex(file, label);
        var registration = file.Registrations[index];
        if (registration.Issuer != issuer.GetLeftPart(UriPartial.Authority))
        {
            throw new InvalidOperationException("Account issuer does not match this credential owner.");
        }
        if (registration.AccessToken is null)
        {
            throw new InvalidOperationException("This account requires sign-in.");
        }
        // Refresh before a request can outlive its bearer token. This is a
        // five-minute safety window, not a timer or a background scheduler.
        // A 401 can precede expiry. Compare while locked: another process may
        // already have renewed the rejected bearer while this request was away.
        if (registration.ExpiresAt <= DateTimeOffset.UtcNow.AddMinutes(5)
            || registration.AccessToken == rejectedBearer)
        {
            registration = await RefreshAsync(registration, cancellation).ConfigureAwait(false);
            WriteFile(new AuthFile(file.HostID, file.Registrations.SetItem(index, registration), file.SelectedPlanAccount));
        }
        if (registration.AccessToken is null)
        {
            throw new InvalidOperationException("This account requires sign-in after a terminal refresh refusal.");
        }
        if (!AuthProtocol.HasPlanScope(registration.Scope))
        {
            throw new InvalidOperationException("This account has not granted ChatGPT plan usage.");
        }
        return registration.AccessToken;
    }

    private partial async Task<AuthRegistration> RefreshAsync(AuthRegistration registration, CancellationToken cancellation)
    {
        var form = new Dictionary<string, string>
        {
            ["grant_type"] = "refresh_token", ["client_id"] = registration.ClientID,
            ["refresh_token"] = registration.RefreshToken!, ["resource"] = "https://api.openai.com/v1"
        };
        for (int attempt = 0; ; attempt++)
        {
            HttpResponseMessage response;
            try
            {
                response = await SendAuthAsync(HttpMethod.Post, new Uri(issuer, "/api/accounts/oauth/token"),
                    form, cancellation).ConfigureAwait(false);
            }
            catch (HttpRequestException) when (attempt < 2)
            {
                await Task.Delay(TimeSpan.FromMilliseconds(250 * (1 << attempt)), cancellation).ConfigureAwait(false);
                continue;
            }
            using (response)
            {
                if (response.StatusCode == HttpStatusCode.OK)
                {
                    var grant = await ReadAuthJsonAsync(response.Content, cancellation).ConfigureAwait(false);
                    return await ReadGrantAsync(grant, registration, cancellation).ConfigureAwait(false);
                }
                if ((int)response.StatusCode >= 500 || response.StatusCode == HttpStatusCode.TooManyRequests)
                {
                    if (attempt < 2)
                    {
                        await Task.Delay(TimeSpan.FromMilliseconds(250 * (1 << attempt)), cancellation).ConfigureAwait(false);
                        continue;
                    }
                    throw new HttpRequestException("Authentication service is temporarily unavailable.", null, response.StatusCode);
                }
                var failure = await ReadAuthJsonAsync(response.Content, cancellation).ConfigureAwait(false);
                string? code = AuthProtocol.OptionalString(failure, "error");
                if (code is "invalid_grant" or "invalid_refresh_token" or "token_expired"
                    or "refresh_token_expired" or "refresh_token_invalidated" or "refresh_token_reused")
                {
                    return registration.WithoutTokens();
                }
                // Do not copy arbitrary server diagnostics: they may echo tokens.
                // invalid_client is configuration trouble, not token revocation.
                string message = code == "invalid_client" ? "Authentication client configuration was refused."
                    : "Authentication refresh was refused.";
                throw new HttpRequestException(message, null, response.StatusCode);
            }
        }
    }

    private partial async Task<AuthRegistration> ReadGrantAsync(JsonElement grant,
        AuthRegistration registration, CancellationToken cancellation, string? nonce)
    {
        // Start expiry at receipt, not after potentially slow signing-key I/O.
        DateTimeOffset receivedAt = DateTimeOffset.UtcNow;
        AuthProtocol.UniqueObject(grant);
        string tokenType = AuthProtocol.RequiredString(grant, "token_type");
        string access = AuthProtocol.RequiredString(grant, "access_token");
        string refresh = AuthProtocol.RequiredString(grant, "refresh_token");
        if (!tokenType.Equals("Bearer", StringComparison.OrdinalIgnoreCase)
            || access.Any(character => character < '!' || character > '~')
            || refresh.Any(character => character < '!' || character > '~')
            || !grant.TryGetProperty("expires_in", out var expiry) || expiry.ValueKind != JsonValueKind.Number
            || !expiry.TryGetInt32(out int seconds) || seconds <= 0)
        {
            throw new InvalidDataException("Invalid renewable credential grant.");
        }
        string scope = nonce is null
            ? AuthProtocol.OptionalString(grant, "scope") ?? registration.Scope!
            : AuthProtocol.RequiredString(grant, "scope");
        string? idToken = nonce is null
            ? AuthProtocol.OptionalString(grant, "id_token")
            : AuthProtocol.RequiredString(grant, "id_token");
        string? email = registration.Email;
        string? subject = registration.Subject;
        if (idToken is not null)
        {
            var keys = await GetAuthDocumentAsync("/.well-known/jwks.json", cancellation).ConfigureAwait(false);
            var identity = AuthProtocol.ValidateIdentity(idToken, keys, registration.Issuer,
                registration.ClientID, nonce, registration.Subject, access, DateTimeOffset.UtcNow);
            subject = identity.Subject;
            email = identity.Email ?? email;
        }
        else
        {
            // OIDC refresh may omit a new ID token. Preserve the identity already
            // validated at login; an old ID token's expiry does not expire refresh.
            idToken = registration.IDToken;
        }
        var replacement = new AuthRegistration(registration.Label, registration.Issuer, subject,
            registration.ClientID, email, access, refresh, idToken, receivedAt.AddSeconds(seconds), scope);
        Validate(replacement);
        return replacement;
    }

    public partial async Task<bool> SignOutAsync(string label, CancellationToken cancellation)
    {
        ArgumentException.ThrowIfNullOrWhiteSpace(label);
        await gate.WaitAsync(cancellation).ConfigureAwait(false);
        try
        {
            ObjectDisposedException.ThrowIf(destroyed, this);
            using var active = CancellationTokenSource.CreateLinkedTokenSource(cancellation, lifetime.Token);
            cancellation = active.Token;
            using var held = await LockFileAsync(cancellation).ConfigureAwait(false);
            var file = ReadFile();
            int index = AccountIndex(file, label);
            var registration = file.Registrations[index];
            // Local revocation is unconditional and durable before any network
            // operation. A timeout cannot leave this application's tokens usable.
            WriteFile(new AuthFile(file.HostID, file.Registrations.SetItem(index, registration.WithoutTokens()), file.SelectedPlanAccount));
            if (registration.RefreshToken is null)
            {
                return false; // No token remains with which to confirm remote revocation.
            }
            if (registration.Issuer != issuer.GetLeftPart(UriPartial.Authority))
            {
                return false; // Never send another issuer's secret to this owner.
            }
            try
            {
                var discovery = await GetAuthDocumentAsync("/.well-known/openid-configuration", cancellation).ConfigureAwait(false);
                if (AuthProtocol.RequiredString(discovery, "issuer") != registration.Issuer)
                {
                    return false;
                }
                string endpoint = AuthProtocol.RequiredString(discovery, "revocation_endpoint");
                if (!Uri.TryCreate(endpoint, UriKind.Absolute, out var revoke)
                    || revoke.GetLeftPart(UriPartial.Authority) != issuer.GetLeftPart(UriPartial.Authority)
                    || revoke.UserInfo.Length != 0 || revoke.Fragment.Length != 0)
                {
                    return false;
                }
                var form = new Dictionary<string, string>
                {
                    ["token"] = registration.RefreshToken, ["token_type_hint"] = "refresh_token",
                    ["client_id"] = registration.ClientID
                };
                for (int attempt = 0; ; attempt++)
                {
                    try
                    {
                        using var response = await SendAuthAsync(HttpMethod.Post, revoke, form, cancellation).ConfigureAwait(false);
                        if (response.StatusCode == HttpStatusCode.OK)
                        {
                            return true;
                        }
                        if ((int)response.StatusCode < 500)
                        {
                            return false;
                        }
                    }
                    catch (HttpRequestException)
                    {
                        // Revocation is idempotent; a bounded retry is safe even
                        // when a connection failed after the server processed it.
                    }
                    if (attempt == 2)
                    {
                        return false;
                    }
                    await Task.Delay(TimeSpan.FromMilliseconds(250 * (1 << attempt)), cancellation).ConfigureAwait(false);
                }
            }
            catch (Exception error) when (error is HttpRequestException or InvalidDataException or IOException)
            {
                return false;
            }
        }
        finally
        {
            gate.Release();
        }
    }

    private partial async Task<HttpResponseMessage> SendAuthAsync(HttpMethod method, Uri endpoint,
        IReadOnlyDictionary<string, string>? form, CancellationToken cancellation)
    {
        using var request = new HttpRequestMessage(method, endpoint);
        if (form is not null)
        {
            request.Content = new FormUrlEncodedContent(form);
        }
        try
        {
            return await http.SendAsync(request, HttpCompletionOption.ResponseHeadersRead, cancellation).ConfigureAwait(false);
        }
        catch (HttpRequestException)
        {
            throw new HttpRequestException("Authentication service could not be reached.");
        }
    }

    private partial async Task<JsonElement> GetAuthDocumentAsync(string path, CancellationToken cancellation)
    {
        using var response = await SendAuthAsync(HttpMethod.Get, new Uri(issuer, path), null, cancellation).ConfigureAwait(false);
        if (response.StatusCode != HttpStatusCode.OK)
        {
            throw new HttpRequestException("Authentication document was unavailable.", null, response.StatusCode);
        }
        return await ReadAuthJsonAsync(response.Content, cancellation).ConfigureAwait(false);
    }

    private static partial async Task<JsonElement> ReadAuthJsonAsync(HttpContent content, CancellationToken cancellation)
    {
        // Headers-only transport plus a bounded reader prevents an untrusted
        // endpoint from allocating an unbounded credential/document body.
        const int limit = 1024 * 1024;
        using var input = await content.ReadAsStreamAsync(cancellation).ConfigureAwait(false);
        using var bytes = new MemoryStream();
        byte[] buffer = new byte[8192];
        while (true)
        {
            int count = await input.ReadAsync(buffer, cancellation).ConfigureAwait(false);
            if (count == 0)
            {
                break;
            }
            if (bytes.Length + count > limit)
            {
                throw new InvalidDataException("Authentication response is too large.");
            }
            bytes.Write(buffer, 0, count);
        }
        try
        {
            using var document = JsonDocument.Parse(bytes.ToArray());
            AuthProtocol.UniqueObject(document.RootElement);
            return document.RootElement.Clone();
        }
        catch (JsonException)
        {
            throw new InvalidDataException("Invalid authentication response JSON.");
        }
    }
}

internal static partial class AuthProtocol
{
    internal static partial bool HasPlanScope(string? scope) =>
        scope?.Split(' ', StringSplitOptions.RemoveEmptyEntries)
            .Contains("chatgpt.tokens.use.direct", StringComparer.Ordinal) == true;
}
