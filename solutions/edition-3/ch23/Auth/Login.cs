using System.Net;
using System.Security.Cryptography;
using System.Text;

namespace Ensemble;

internal sealed partial class AuthRegistration
{
    internal partial bool SameSnapshot(AuthRegistration other) =>
        Label == other.Label && Issuer == other.Issuer && Subject == other.Subject
        && ClientID == other.ClientID && Email == other.Email && AccessToken == other.AccessToken
        && RefreshToken == other.RefreshToken && IDToken == other.IDToken
        && ExpiresAt == other.ExpiresAt && Scope == other.Scope;
}

internal sealed partial class AuthLogin
{
    internal partial AuthLogin(string label, Uri redirect, AuthRegistration? before)
    {
        Label = label;
        Redirect = redirect;
        Before = before;
        State = AuthProtocol.Encode(RandomNumberGenerator.GetBytes(32));
        Nonce = AuthProtocol.Encode(RandomNumberGenerator.GetBytes(32));
        Verifier = AuthProtocol.Encode(RandomNumberGenerator.GetBytes(32));
        // Bound abandoned attempts without a timer, thread, or persisted secret.
        ExpiresAt = DateTimeOffset.UtcNow.AddMinutes(10);
    }
}

public sealed partial class Credentials
{
    public partial async Task<Uri> BeginLoginAsync(string label, Uri redirect,
        bool requestConsent, CancellationToken cancellation)
    {
        ArgumentException.ThrowIfNullOrWhiteSpace(label);
        ArgumentNullException.ThrowIfNull(redirect);
        // A fixed path makes registration reusable when the available port changes.
        // Do not normalize localhost, arbitrary paths, credentials or fragments.
        if (!redirect.IsAbsoluteUri || redirect.Scheme != "http" || redirect.Host != "127.0.0.1"
            || redirect.Port <= 0 || redirect.AbsolutePath != "/auth/callback"
            || redirect.UserInfo.Length != 0 || redirect.Query.Length != 0 || redirect.Fragment.Length != 0
            || redirect.OriginalString != redirect.AbsoluteUri)
        {
            throw new ArgumentException("Login requires an exact HTTP 127.0.0.1 /auth/callback URI.", nameof(redirect));
        }
        await gate.WaitAsync(cancellation).ConfigureAwait(false);
        try
        {
            ObjectDisposedException.ThrowIf(destroyed, this);
            if (login is not null)
            {
                throw new InvalidOperationException("A login attempt is already pending; complete or cancel it first.");
            }
            using var active = CancellationTokenSource.CreateLinkedTokenSource(cancellation, lifetime.Token);
            using var held = await LockFileAsync(active.Token).ConfigureAwait(false);
            var file = ReadFile();
            var before = file.Registrations.FirstOrDefault(account => account.Label == label);
            if (before is not null && before.Issuer != issuer.GetLeftPart(UriPartial.Authority))
            {
                throw new InvalidOperationException("Account issuer does not match this credential owner.");
            }
            // Persist host identity before exposing the first authorization URL.
            // No new account is published until a validated callback supplies its ID.
            WriteFile(file);
            var attempt = new AuthLogin(label, redirect, before);
            var parameters = new Dictionary<string, string>
            {
                ["client_id"] = before?.ClientID ?? "dynamic_agent_client",
                ["ext_agent_host_id"] = file.HostID,
                ["response_type"] = "code",
                ["redirect_uri"] = redirect.AbsoluteUri,
                ["scope"] = "openid profile email offline_access resource.invoke chatgpt.tokens.use.direct",
                ["resource"] = "https://api.openai.com/v1",
                ["state"] = attempt.State,
                ["nonce"] = attempt.Nonce,
                ["code_challenge_method"] = "S256",
                ["code_challenge"] = AuthProtocol.Encode(SHA256.HashData(Encoding.ASCII.GetBytes(attempt.Verifier)))
            };
            if (before is null)
            {
                parameters.Add("agent_name_hint", "Ensemble");
            }
            else
            {
                if (before.IDToken is not null)
                {
                    parameters.Add("id_token_hint", before.IDToken);
                }
                if (before.Email is not null)
                {
                    parameters.Add("login_hint", before.Email);
                }
            }
            if (requestConsent)
            {
                parameters.Add("prompt", "consent");
            }
            string query = string.Join('&', parameters.Select(pair =>
                Uri.EscapeDataString(pair.Key) + "=" + Uri.EscapeDataString(pair.Value)));
            var authorization = new Uri(issuer, "/api/accounts/authorize?" + query);
            login = attempt;
            return authorization;
        }
        finally
        {
            gate.Release();
        }
    }

    public partial async Task<bool> LoginPendingAsync(CancellationToken cancellation)
    {
        await gate.WaitAsync(cancellation).ConfigureAwait(false);
        try
        {
            ObjectDisposedException.ThrowIf(destroyed, this);
            return login is not null && DateTimeOffset.UtcNow < login.ExpiresAt;
        }
        finally
        {
            gate.Release();
        }
    }

    public partial async Task CancelLoginAsync(CancellationToken cancellation)
    {
        await gate.WaitAsync(cancellation).ConfigureAwait(false);
        try
        {
            ObjectDisposedException.ThrowIf(destroyed, this);
            login = null;
        }
        finally
        {
            gate.Release();
        }
    }

    public partial async Task<AuthAccount> CompleteLoginAsync(string callback, CancellationToken cancellation)
    {
        await gate.WaitAsync(cancellation).ConfigureAwait(false);
        try
        {
            ObjectDisposedException.ThrowIf(destroyed, this);
            var attempt = login ?? throw new InvalidOperationException("No login attempt is pending.");
            if (DateTimeOffset.UtcNow >= attempt.ExpiresAt)
            {
                login = null;
                throw new InvalidOperationException("Login attempt expired; begin a new authorization.");
            }
            var parameters = ReadCallback(callback, attempt);
            // A matched callback consumes the attempt, even when the user denied
            // consent or the exchange fails. Unsolicited/wrong-state callbacks do
            // not cancel a legitimate browser's outstanding attempt.
            login = null;
            if (parameters.ContainsKey("error"))
            {
                throw new InvalidOperationException("Authorization was declined or failed; begin a new login if desired.");
            }
            if (!parameters.TryGetValue("code", out string? code) || string.IsNullOrWhiteSpace(code))
            {
                throw new InvalidDataException("Authorization callback has no code.");
            }
            if (parameters.TryGetValue("iss", out string? callbackIssuer)
                && callbackIssuer != issuer.GetLeftPart(UriPartial.Authority))
            {
                throw new InvalidDataException("Authorization callback issuer did not match.");
            }
            parameters.TryGetValue("client_id", out string? suppliedClient);
            string? clientID = attempt.Before?.ClientID ?? suppliedClient;
            if (string.IsNullOrWhiteSpace(clientID) || clientID == "dynamic_agent_client"
                || clientID.Any(character => character < '!' || character > '~')
                || (suppliedClient is not null && suppliedClient != clientID))
            {
                throw new InvalidDataException("Authorization callback has no matching issued client ID.");
            }
            using var active = CancellationTokenSource.CreateLinkedTokenSource(cancellation, lifetime.Token);
            cancellation = active.Token;
            using var held = await LockFileAsync(cancellation).ConfigureAwait(false);
            var file = ReadFile();
            var current = file.Registrations.FirstOrDefault(account => account.Label == attempt.Label);
            if ((attempt.Before is null && current is not null)
                || (attempt.Before is not null && (current is null || !attempt.Before.SameSnapshot(current))))
            {
                throw new InvalidOperationException("Registration changed during login; begin a new authorization.");
            }
            if (current is null)
            {
                if (file.Registrations.Any(account => account.Issuer == issuer.GetLeftPart(UriPartial.Authority)
                    && account.ClientID == clientID))
                {
                    throw new InvalidOperationException("Issued client already belongs to another account label.");
                }
                current = new AuthRegistration(attempt.Label, issuer.GetLeftPart(UriPartial.Authority), null,
                    clientID, null, null, null, null, null, null);
                file = new AuthFile(file.HostID, file.Registrations.Add(current), file.SelectedPlanAccount);
                // This durable token-free registration deliberately survives a
                // failed code exchange. Retrying never reuses the bootstrap ID.
                WriteFile(file);
            }
            var form = new Dictionary<string, string>
            {
                ["grant_type"] = "authorization_code", ["client_id"] = clientID,
                ["code"] = code, ["code_verifier"] = attempt.Verifier,
                ["redirect_uri"] = attempt.Redirect.AbsoluteUri, ["resource"] = "https://api.openai.com/v1"
            };
            using var response = await SendAuthAsync(HttpMethod.Post,
                new Uri(issuer, "/api/accounts/oauth/token"), form, cancellation).ConfigureAwait(false);
            if (response.StatusCode != HttpStatusCode.OK)
            {
                // Codes are one-use. No automatic retry after an uncertain POST,
                // no copying token-service diagnostics, and no erasing old tokens.
                throw new HttpRequestException("Code exchange failed; begin a new authorization using the saved registration.",
                    null, response.StatusCode);
            }
            var grant = await ReadAuthJsonAsync(response.Content, cancellation).ConfigureAwait(false);
            var replacement = await ReadGrantAsync(grant, current, cancellation, attempt.Nonce).ConfigureAwait(false);
            int index = AccountIndex(file, attempt.Label);
            // Grant and future funding choice publish together. Even identity-only
            // login selects the plan: missing permission must refuse, never spend
            // through an ambient API key. Explicit launch overrides remain possible.
            WriteFile(new AuthFile(file.HostID, file.Registrations.SetItem(index, replacement), attempt.Label));
            return replacement.Account();
        }
        finally
        {
            gate.Release();
        }
    }

    private static partial Dictionary<string, string> ReadCallback(string callback, AuthLogin attempt)
    {
        // Compare the literal endpoint before parsing. URI normalization must not
        // make a foreign authority, path spelling, fragment, or userinfo acceptable.
        if (callback is null || callback.Length > 32768 || callback.Contains('#'))
        {
            throw new InvalidDataException("Invalid authorization callback URI.");
        }
        int query = callback.IndexOf('?');
        if (query < 0 || callback[..query] != attempt.Redirect.AbsoluteUri)
        {
            throw new InvalidDataException("Authorization callback endpoint did not match.");
        }
        var parameters = new Dictionary<string, string>(StringComparer.Ordinal);
        foreach (string part in callback[(query + 1)..].Split('&'))
        {
            int equals = part.IndexOf('=');
            if (equals <= 0)
            {
                throw new InvalidDataException("Invalid authorization callback parameters.");
            }
            string name = DecodeCallbackPart(part[..equals]);
            string value = DecodeCallbackPart(part[(equals + 1)..]);
            if (!parameters.TryAdd(name, value))
            {
                throw new InvalidDataException("Ambiguous authorization callback parameters.");
            }
        }
        if (!parameters.TryGetValue("state", out string? state)
            || !CryptographicOperations.FixedTimeEquals(Encoding.UTF8.GetBytes(state), Encoding.UTF8.GetBytes(attempt.State)))
        {
            throw new InvalidDataException("Authorization callback state did not match.");
        }
        return parameters;
    }

    private static partial string DecodeCallbackPart(string text)
    {
        // Form decoding with strict percent/UTF-8 validation. Permissive URI
        // decoding would silently retain malformed escapes or replacement text.
        using var bytes = new MemoryStream();
        for (int index = 0; index < text.Length; index++)
        {
            char character = text[index];
            if (character == '%')
            {
                if (index + 2 >= text.Length || !byte.TryParse(text.AsSpan(index + 1, 2),
                    System.Globalization.NumberStyles.AllowHexSpecifier, null, out byte value))
                {
                    throw new InvalidDataException("Invalid authorization callback encoding.");
                }
                bytes.WriteByte(value);
                index += 2;
            }
            else if (character >= '!' && character <= '~')
            {
                bytes.WriteByte(character == '+' ? (byte)' ' : (byte)character);
            }
            else
            {
                throw new InvalidDataException("Invalid authorization callback encoding.");
            }
        }
        try
        {
            string decoded = new UTF8Encoding(false, true).GetString(bytes.ToArray());
            if (decoded.Any(char.IsControl))
            {
                throw new InvalidDataException("Invalid authorization callback character.");
            }
            return decoded;
        }
        catch (DecoderFallbackException)
        {
            throw new InvalidDataException("Invalid authorization callback encoding.");
        }
    }
}
