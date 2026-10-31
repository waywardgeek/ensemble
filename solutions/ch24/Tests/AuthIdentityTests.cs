using System.Security.Cryptography;
using System.Text;
using System.Text.Json;
using System.Text.Json.Nodes;
using Ensemble;

internal static class AuthIdentityTests
{
    // This is an issuer, not a mocked validator: it signs real JWS bytes and
    // publishes the corresponding public RSA key. Its private key never leaves
    // the fixture. Mutable claims make each negative case a genuinely signed lie.
    internal sealed class Issuer : IDisposable
    {
        internal readonly RSA Key = RSA.Create(2048);
        internal readonly DateTimeOffset Now = DateTimeOffset.FromUnixTimeSeconds(1800000000);

        internal JsonObject Claims() => new()
        {
            ["iss"] = "https://issuer.invalid", ["aud"] = "issued-client", ["sub"] = "subject-one",
            ["email"] = "shared@example.invalid", ["nonce"] = "attempt-nonce",
            ["iat"] = Now.ToUnixTimeSeconds(), ["exp"] = Now.AddHours(1).ToUnixTimeSeconds(),
            ["at_hash"] = AuthProtocol.Encode(SHA256.HashData(Encoding.ASCII.GetBytes("private-access"))[..16])
        };

        internal JsonObject Header() => new() { ["alg"] = "RS256", ["kid"] = "local-key", ["typ"] = "JWT" };

        internal JsonObject PublicKey()
        {
            var key = Key.ExportParameters(false);
            return new JsonObject
            {
                ["kty"] = "RSA", ["alg"] = "RS256", ["use"] = "sig", ["kid"] = "local-key",
                ["key_ops"] = new JsonArray("verify"),
                ["n"] = AuthProtocol.Encode(key.Modulus!), ["e"] = AuthProtocol.Encode(key.Exponent!)
            };
        }

        internal JsonElement Keys(JsonObject? key = null) =>
            JsonSerializer.SerializeToElement(new JsonObject { ["keys"] = new JsonArray(key ?? PublicKey()) });

        internal string Sign(JsonObject? claims = null, JsonObject? header = null) =>
            SignRaw((claims ?? Claims()).ToJsonString(), (header ?? Header()).ToJsonString());

        internal string SignRaw(string claims, string header)
        {
            string unsigned = AuthProtocol.Encode(Encoding.UTF8.GetBytes(header)) + "."
                + AuthProtocol.Encode(Encoding.UTF8.GetBytes(claims));
            byte[] signature = Key.SignData(Encoding.ASCII.GetBytes(unsigned), HashAlgorithmName.SHA256,
                RSASignaturePadding.Pkcs1);
            return unsigned + "." + AuthProtocol.Encode(signature);
        }

        internal AuthIdentity Validate(string token, JsonElement? keys = null, string? nonce = "attempt-nonce",
            string? subject = null, DateTimeOffset? now = null, string access = "private-access") =>
            AuthProtocol.ValidateIdentity(token, keys ?? Keys(), "https://issuer.invalid", "issued-client",
                nonce, subject, access, now ?? Now);

        public void Dispose() => Key.Dispose();
    }

    private static void Check(bool condition, string reason)
    {
        if (!condition)
        {
            throw new Exception(reason);
        }
    }

    private static void Refuses(Action operation)
    {
        try
        {
            operation();
        }
        catch (InvalidDataException error)
        {
            Check(error.InnerException is null && !error.ToString().Contains("private-access", StringComparison.Ordinal)
                && !error.ToString().Contains("secret-payload", StringComparison.Ordinal),
                "credential refusal must not expose untrusted token contents");
            return;
        }
        throw new Exception("Expected signed identity refusal.");
    }

    internal static void ValidIdentities()
    {
        using var issuer = new Issuer();
        var identity = issuer.Validate(issuer.Sign());
        Check(identity.Issuer == "https://issuer.invalid" && identity.Subject == "subject-one"
            && identity.Email == "shared@example.invalid", "signature and expected identity verified");
        var claims = issuer.Claims();
        claims["aud"] = new JsonArray("other-service", "issued-client");
        claims["azp"] = "issued-client";
        Check(issuer.Validate(issuer.Sign(claims)).Subject == identity.Subject, "explicit multi-audience authorized party");
        claims.Remove("nonce");
        claims.Remove("at_hash");
        claims.Remove("email");
        Check(issuer.Validate(issuer.Sign(claims), nonce: null, subject: "subject-one").Email is null,
            "refresh pins saved identity without requiring new login nonce or optional claims");
        var key = issuer.PublicKey();
        key.Remove("alg");
        key.Remove("use");
        key.Remove("key_ops");
        Check(issuer.Validate(issuer.Sign(), issuer.Keys(key)).Subject == identity.Subject,
            "optional JWK restrictions may be absent; algorithm still fixed by validator");
        Check(issuer.Validate(issuer.Sign(), now: issuer.Now.AddSeconds(-5)).Subject == identity.Subject,
            "bounded five-second clock tolerance");
    }

    internal static void SignatureRefusals()
    {
        using var issuer = new Issuer();
        using var impostor = new Issuer();
        Refuses(() => issuer.Validate(impostor.Sign()));
        string token = issuer.Sign();
        string[] pieces = token.Split('.');
        Refuses(() => issuer.Validate(pieces[0] + "." + pieces[1] + ".AA"));
        Refuses(() => issuer.Validate(pieces[0] + "." + pieces[1] + "." + pieces[2] + "="));
        Refuses(() => issuer.Validate(token + ".extra"));
        Refuses(() => issuer.Validate(new string('x', 65537)));
        Refuses(() => issuer.Validate("secret-payload"));
        foreach (string algorithm in new[] { "none", "HS256", "RS512", "rs256" })
        {
            var header = issuer.Header();
            header["alg"] = algorithm;
            Refuses(() => issuer.Validate(issuer.Sign(header: header)));
        }
        foreach (string field in new[] { "crit", "b64" })
        {
            var header = issuer.Header();
            header[field] = false;
            Refuses(() => issuer.Validate(issuer.Sign(header: header)));
        }
        var unknown = issuer.Header();
        unknown["kid"] = "not-published";
        // A token-supplied JWK/URL must not override the trusted published set.
        unknown["jwk"] = issuer.PublicKey();
        unknown["jku"] = "https://attacker.invalid/keys";
        Refuses(() => issuer.Validate(issuer.Sign(header: unknown)));
        string duplicateHeader = "{\"alg\":\"HS256\",\"alg\":\"RS256\",\"kid\":\"local-key\"}";
        Refuses(() => issuer.Validate(issuer.SignRaw(issuer.Claims().ToJsonString(), duplicateHeader)));
        var duplicates = JsonSerializer.SerializeToElement(new JsonObject
        {
            ["keys"] = new JsonArray(issuer.PublicKey(), issuer.PublicKey())
        });
        Refuses(() => issuer.Validate(token, duplicates));
        foreach (var (field, value) in new[] { ("alg", "HS256"), ("kty", "EC"), ("use", "enc"),
            ("n", "AAAA"), ("e", "AA"), ("kid", "wrong") })
        {
            var key = issuer.PublicKey();
            key[field] = value;
            Refuses(() => issuer.Validate(token, issuer.Keys(key)));
        }
        var operations = issuer.PublicKey();
        operations["key_ops"] = new JsonArray("encrypt");
        Refuses(() => issuer.Validate(token, issuer.Keys(operations)));
        Refuses(() => issuer.Validate(token, JsonSerializer.SerializeToElement(new { keys = "secret-payload" })));
    }

    internal static void ClaimRefusals()
    {
        using var issuer = new Issuer();
        foreach (string field in new[] { "iss", "aud", "sub", "iat", "exp", "nonce" })
        {
            var claims = issuer.Claims();
            claims.Remove(field);
            Refuses(() => issuer.Validate(issuer.Sign(claims)));
            claims = issuer.Claims();
            claims[field] = null;
            Refuses(() => issuer.Validate(issuer.Sign(claims)));
        }
        foreach (var (field, value) in new[] { ("iss", "https://issuer.invalid/"), ("aud", "other-client"),
            ("nonce", "other-attempt"), ("sub", " "), ("sid", ""), ("email", ""), ("azp", "other-client"),
            ("at_hash", "AAAA"), ("iat", "secret-payload") })
        {
            var claims = issuer.Claims();
            claims[field] = value;
            Refuses(() => issuer.Validate(issuer.Sign(claims)));
        }
        var multiple = issuer.Claims();
        multiple["aud"] = new JsonArray("issued-client", "other-service");
        Refuses(() => issuer.Validate(issuer.Sign(multiple)));
        multiple["azp"] = "issued-client";
        multiple["aud"] = new JsonArray("issued-client", "issued-client");
        Refuses(() => issuer.Validate(issuer.Sign(multiple)));
        var future = issuer.Claims();
        future["iat"] = issuer.Now.AddSeconds(6).ToUnixTimeSeconds();
        Refuses(() => issuer.Validate(issuer.Sign(future)));
        future = issuer.Claims();
        future["nbf"] = issuer.Now.AddSeconds(6).ToUnixTimeSeconds();
        Refuses(() => issuer.Validate(issuer.Sign(future)));
        var expired = issuer.Claims();
        expired["iat"] = issuer.Now.AddHours(-2).ToUnixTimeSeconds();
        expired["exp"] = issuer.Now.AddSeconds(-5).ToUnixTimeSeconds();
        Refuses(() => issuer.Validate(issuer.Sign(expired)));
        var inverted = issuer.Claims();
        inverted["exp"] = inverted["iat"]!.DeepClone();
        Refuses(() => issuer.Validate(issuer.Sign(inverted)));
        var overflow = issuer.Claims();
        overflow["exp"] = double.MaxValue;
        Refuses(() => issuer.Validate(issuer.Sign(overflow)));
        string duplicate = issuer.Claims().ToJsonString().Replace("\"sub\":", "\"sub\":\"impostor\",\"sub\":", StringComparison.Ordinal);
        Refuses(() => issuer.Validate(issuer.SignRaw(duplicate, issuer.Header().ToJsonString())));
        Refuses(() => issuer.Validate(issuer.Sign(), access: "different-access"));
        Refuses(() => issuer.Validate(issuer.Sign(), nonce: null, subject: "different-person"));
        Refuses(() => issuer.Validate(issuer.Sign(), nonce: null, subject: null));
    }
}
