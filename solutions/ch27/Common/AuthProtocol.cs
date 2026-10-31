using System.Text.Json;

namespace Ensemble;

/// <summary>Value-like verified identity, not a session or credential owner.</summary>
internal sealed record AuthIdentity(string Issuer, string Subject, string? Email);

/// <summary>
/// The deliberately narrow OIDC wire vocabulary. Auth owns its implementations;
/// callers supply trusted issuer keys, never key URLs taken from an untrusted JWT.
/// No credentials or validation state are retained here.
/// </summary>
internal static partial class AuthProtocol
{
    // A login requires its original nonce. Refresh instead pins the saved subject;
    // it may omit nonce because no new authorization transaction was initiated.
    internal static partial AuthIdentity ValidateIdentity(string token, JsonElement keys,
        string issuer, string clientID, string? nonce, string? subject,
        string accessToken, DateTimeOffset now);

    // Canonical unpadded base64url serves both JWS and PKCE. Decoding refuses other
    // alphabets and alternate spellings rather than accepting ambiguous tokens.
    internal static partial bool HasPlanScope(string? scope);
    internal static partial string Encode(byte[] bytes);
    private static partial byte[] Decode(string text);
    internal static partial void UniqueObject(JsonElement value);
    internal static partial string RequiredString(JsonElement value, string name);
    internal static partial string? OptionalString(JsonElement value, string name);
    private static partial long RequiredTime(JsonElement value, string name);
    private static partial void VerifySignature(string signingInput, byte[] signature,
        JsonElement header, JsonElement keys);
    private static partial void ValidateAudience(JsonElement claims, string clientID);
}
