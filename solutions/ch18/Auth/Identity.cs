using System.Security.Cryptography;
using System.Text;
using System.Text.Json;

namespace Ensemble;

internal static partial class AuthProtocol
{
    internal static partial string Encode(byte[] bytes) =>
        Convert.ToBase64String(bytes).TrimEnd('=').Replace('+', '-').Replace('/', '_');

    private static partial byte[] Decode(string text)
    {
        if (text.Length == 0 || text.Any(character =>
            !char.IsAsciiLetterOrDigit(character) && character != '-' && character != '_'))
        {
            throw new InvalidDataException("Invalid identity encoding.");
        }
        string padded = text.Replace('-', '+').Replace('_', '/');
        padded += new string('=', (4 - padded.Length % 4) % 4);
        byte[] bytes = Convert.FromBase64String(padded);
        if (Encode(bytes) != text)
        {
            throw new InvalidDataException("Noncanonical identity encoding.");
        }
        return bytes;
    }

    internal static partial void UniqueObject(JsonElement value)
    {
        if (value.ValueKind != JsonValueKind.Object)
        {
            throw new InvalidDataException("Expected an identity object.");
        }
        var names = new HashSet<string>(StringComparer.Ordinal);
        foreach (var member in value.EnumerateObject())
        {
            if (!names.Add(member.Name))
            {
                throw new InvalidDataException("Duplicate identity field.");
            }
        }
    }

    internal static partial string RequiredString(JsonElement value, string name)
    {
        string? text = OptionalString(value, name);
        if (text is null)
        {
            throw new InvalidDataException("Required identity field is missing.");
        }
        return text;
    }

    internal static partial string? OptionalString(JsonElement value, string name)
    {
        if (!value.TryGetProperty(name, out var member))
        {
            return null;
        }
        if (member.ValueKind != JsonValueKind.String || string.IsNullOrWhiteSpace(member.GetString()))
        {
            throw new InvalidDataException("Invalid identity field.");
        }
        return member.GetString()!;
    }

    private static partial long RequiredTime(JsonElement value, string name)
    {
        if (!value.TryGetProperty(name, out var member) || member.ValueKind != JsonValueKind.Number
            || !member.TryGetInt64(out long seconds) || seconds < 0 || seconds > 253402300799L)
        {
            throw new InvalidDataException("Invalid identity timestamp.");
        }
        return seconds;
    }

    private static partial void VerifySignature(string signingInput, byte[] signature,
        JsonElement header, JsonElement keys)
    {
        UniqueObject(header);
        if (RequiredString(header, "alg") != "RS256")
        {
            throw new InvalidDataException("Unsupported identity signature algorithm.");
        }
        // This profile has no critical extensions and uses ordinary base64url JWS.
        // Never follow jku/x5u or accept an embedded key from the token itself.
        if (header.TryGetProperty("crit", out _) || header.TryGetProperty("b64", out _))
        {
            throw new InvalidDataException("Unsupported identity signature extension.");
        }
        string kid = RequiredString(header, "kid");
        UniqueObject(keys);
        if (!keys.TryGetProperty("keys", out var items) || items.ValueKind != JsonValueKind.Array)
        {
            throw new InvalidDataException("Invalid issuer signing keys.");
        }
        JsonElement? selected = null;
        foreach (var key in items.EnumerateArray())
        {
            UniqueObject(key);
            if (OptionalString(key, "kid") != kid)
            {
                continue;
            }
            if (selected is not null)
            {
                throw new InvalidDataException("Ambiguous issuer signing key.");
            }
            selected = key;
        }
        if (selected is not JsonElement signingKey)
        {
            throw new InvalidDataException("Issuer signing key was not found.");
        }
        string? algorithm = OptionalString(signingKey, "alg");
        string? use = OptionalString(signingKey, "use");
        if (RequiredString(signingKey, "kty") != "RSA"
            || (algorithm is not null && algorithm != "RS256")
            || (use is not null && use != "sig"))
        {
            throw new InvalidDataException("Issuer key does not permit this signature.");
        }
        if (signingKey.TryGetProperty("key_ops", out var operations))
        {
            if (operations.ValueKind != JsonValueKind.Array || operations.GetArrayLength() != 1
                || operations[0].ValueKind != JsonValueKind.String || operations[0].GetString() != "verify")
            {
                throw new InvalidDataException("Issuer key does not permit verification.");
            }
        }
        byte[] modulus = Decode(RequiredString(signingKey, "n"));
        byte[] exponent = Decode(RequiredString(signingKey, "e"));
        if (modulus.Length < 256 || modulus.Length > 1024 || modulus[0] < 128
            || exponent.Length > 8 || exponent[0] == 0 || signature.Length != modulus.Length)
        {
            throw new InvalidDataException("Unsupported issuer RSA key or signature size.");
        }
        using var rsa = RSA.Create();
        rsa.ImportParameters(new RSAParameters { Modulus = modulus, Exponent = exponent });
        bool verified = rsa.VerifyData(Encoding.ASCII.GetBytes(signingInput), signature,
            HashAlgorithmName.SHA256, RSASignaturePadding.Pkcs1);
        if (!verified)
        {
            throw new InvalidDataException("Identity signature was not valid.");
        }
    }

    private static partial void ValidateAudience(JsonElement claims, string clientID)
    {
        if (!claims.TryGetProperty("aud", out var audience))
        {
            throw new InvalidDataException("Identity audience is missing.");
        }
        var audiences = new HashSet<string>(StringComparer.Ordinal);
        if (audience.ValueKind == JsonValueKind.String)
        {
            audiences.Add(audience.GetString()!);
        }
        else if (audience.ValueKind == JsonValueKind.Array)
        {
            foreach (var value in audience.EnumerateArray())
            {
                if (value.ValueKind != JsonValueKind.String || string.IsNullOrWhiteSpace(value.GetString())
                    || !audiences.Add(value.GetString()!))
                {
                    throw new InvalidDataException("Invalid identity audiences.");
                }
            }
        }
        if (!audiences.Contains(clientID))
        {
            throw new InvalidDataException("Identity audience did not match the registration.");
        }
        string? authorizedParty = OptionalString(claims, "azp");
        if ((audiences.Count > 1 && authorizedParty is null)
            || (authorizedParty is not null && authorizedParty != clientID))
        {
            throw new InvalidDataException("Identity authorized party did not match the registration.");
        }
    }

    internal static partial AuthIdentity ValidateIdentity(string token, JsonElement keys,
        string issuer, string clientID, string? nonce, string? subject,
        string accessToken, DateTimeOffset now)
    {
        // These expectations come from the caller's trusted transaction/store,
        // not from unverified claims. Refresh must still pin an existing identity.
        if (string.IsNullOrWhiteSpace(issuer) || string.IsNullOrWhiteSpace(clientID)
            || clientID == "dynamic_agent_client" || string.IsNullOrWhiteSpace(accessToken)
            || accessToken.Any(character => character < '!' || character > '~')
            || (string.IsNullOrWhiteSpace(nonce) && string.IsNullOrWhiteSpace(subject)))
        {
            throw new InvalidDataException("Identity validation requires trusted expectations.");
        }
        try
        {
            if (token.Length > 65536)
            {
                throw new InvalidDataException("Identity token is too large.");
            }
            string[] parts = token.Split('.');
            if (parts.Length != 3)
            {
                throw new InvalidDataException("Identity token is not a signed JWT.");
            }
            using var header = JsonDocument.Parse(Decode(parts[0]));
            VerifySignature(parts[0] + "." + parts[1], Decode(parts[2]), header.RootElement, keys);
            using var payload = JsonDocument.Parse(Decode(parts[1]));
            var claims = payload.RootElement;
            UniqueObject(claims);
            if (RequiredString(claims, "iss") != issuer)
            {
                throw new InvalidDataException("Identity issuer did not match.");
            }
            ValidateAudience(claims, clientID);
            long issued = RequiredTime(claims, "iat");
            long expires = RequiredTime(claims, "exp");
            long current = now.ToUnixTimeSeconds();
            // Five seconds absorbs small clock differences, not expired sessions.
            if (expires <= issued || issued > current + 5 || expires <= current - 5)
            {
                throw new InvalidDataException("Identity token is outside its validity interval.");
            }
            if (claims.TryGetProperty("nbf", out _) && RequiredTime(claims, "nbf") > current + 5)
            {
                throw new InvalidDataException("Identity token is not yet valid.");
            }
            string identitySubject = RequiredString(claims, "sub");
            if (subject is not null && identitySubject != subject)
            {
                throw new InvalidDataException("Identity subject changed for this registration.");
            }
            string? tokenNonce = OptionalString(claims, "nonce");
            if (nonce is not null && tokenNonce != nonce)
            {
                throw new InvalidDataException("Identity nonce did not match the sign-in attempt.");
            }
            // Session identifiers are metadata, but malformed supplied values are
            // still refused. Email is never used to merge registrations.
            _ = OptionalString(claims, "sid");
            string? email = OptionalString(claims, "email");
            string? accessHash = OptionalString(claims, "at_hash");
            if (accessHash is not null)
            {
                byte[] hash = SHA256.HashData(Encoding.ASCII.GetBytes(accessToken));
                byte[] supplied = Decode(accessHash);
                if (!CryptographicOperations.FixedTimeEquals(supplied, hash.AsSpan(0, hash.Length / 2)))
                {
                    throw new InvalidDataException("Identity access-token hash did not match.");
                }
            }
            return new AuthIdentity(issuer, identitySubject, email);
        }
        catch (Exception error) when (error is JsonException or FormatException
            or CryptographicException or ArgumentException or OverflowException)
        {
            // Parser/crypto exceptions can include input. Do not attach an inner
            // exception or copy its message across the credential boundary.
            throw new InvalidDataException("Invalid signed identity token or issuer key.");
        }
    }
}
