using System.Collections.Immutable;

namespace Ensemble;

// Value-like, public provider catalog entry. Availability is not an Engine
// capability or a price; never promote a name into the local capability table.
public sealed record PlanModel(string Slug, string DisplayName);

// Value-like, deliberately redacted account status. Hosts may display this value;
// it contains no access, refresh or ID token. Email is a label, never identity.
public sealed record AuthAccount(string Label, string Issuer, string? Subject,
    string ClientID, string? Email, DateTimeOffset? ExpiresAt, bool SignedIn, bool PlanEnabled);

// Value-like private storage vocabulary, not model conversation. Keep these
// records internal and never serialize them with the History/event codecs.
// A class rather than a record avoids a generated ToString that prints secrets.
internal sealed partial class AuthRegistration
{
    internal readonly string Label;
    internal readonly string Issuer;
    // Null only for an issued registration awaiting its first verified login.
    internal readonly string? Subject;
    internal readonly string ClientID;
    internal readonly string? Email;
    internal readonly string? AccessToken;
    internal readonly string? RefreshToken;
    internal readonly string? IDToken;
    internal readonly DateTimeOffset? ExpiresAt;
    internal readonly string? Scope;
    internal partial AuthRegistration(string label, string issuer, string? subject,
        string clientID, string? email, string? accessToken, string? refreshToken,
        string? idToken, DateTimeOffset? expiresAt, string? scope);
    internal partial AuthAccount Account();
    internal partial AuthRegistration WithoutTokens();
    // Compare an optimistic login snapshot with a freshly locked disk record.
    internal partial bool SameSnapshot(AuthRegistration other);
}

// Value-like disk snapshot. It is staged inside one transaction, not retained as
// a second credential authority. Host identity survives logout and process exit.
internal sealed partial class AuthFile
{
    internal readonly string HostID;
    internal readonly ImmutableArray<AuthRegistration> Registrations;
    // Saved launch default, not a guarantee of a currently usable credential.
    internal readonly string? SelectedPlanAccount;
    internal partial AuthFile(string hostID, ImmutableArray<AuthRegistration> registrations,
        string? selectedPlanAccount = null);
}

/// <summary>Optional Ensemble-owned protected credential storage shared by its
/// Agents. Construction registers both ends but performs no file or network I/O.
/// Supply a dedicated private directory, not a conversation workspace.</summary>
public sealed partial class Credentials
{
    internal Ensemble? parent;
    internal bool destroyed;
    private readonly string directory;
    private readonly Uri issuer;
    private readonly HttpClient http;
    private readonly CancellationTokenSource lifetime = new();
    // Immutable, secret-bearing protocol state; never serialize into History.
    private AuthLogin? login;
    // Local lifecycle serialization plus an OS-held file-share lock for other
    // cooperating processes. A leftover lock FILE is not a held lock.
    private readonly SemaphoreSlim gate = new(1, 1);

    public partial Credentials(Ensemble parent, string directory);
    // The alternate issuer is internal: local tests exercise the real HTTP
    // protocol, not a production provider registry or user-supplied token URL.
    internal partial Credentials(Ensemble parent, string directory, Uri issuer);
    public partial Ensemble Parent();
    // Transport-only secret. Never pass this result to rendering, History or logs.
    public partial Task<string> GetBearerTokenAsync(string label, CancellationToken cancellation = default);
    // Fresh account-specific catalog, in server order. No History/meter effects.
    // The base follows inference URL conventions; HTTPS except owned loopback tests.
    public partial Task<ImmutableArray<PlanModel>> ModelsAsync(string label,
        string baseUrl = "https://api.openai.com", CancellationToken cancellation = default);
    // A rejected bearer is a compare value, not a cached credential. Under the
    // file lock, reuse a newer grant or rotate this exact rejected generation.
    internal partial Task<string> GetBearerAfterRejectionAsync(string label, string? rejectedBearer,
        CancellationToken cancellation);
    private partial Task<string> ResolveBearerAsync(string label, CancellationToken cancellation,
        string? rejectedBearer = null);
    // Shared credential-bearing endpoint policy for inference and discovery.
    internal static partial void ValidatePlanEndpoint(Uri endpoint);
    // Local logout commits first. False means remote revocation was not confirmed;
    // the caller must tell the user, not describe this as confirmed server logout.
    public partial Task<bool> SignOutAsync(string label, CancellationToken cancellation = default);
    private partial Task<AuthRegistration> RefreshAsync(AuthRegistration registration, CancellationToken cancellation);
    private partial Task<AuthRegistration> ReadGrantAsync(System.Text.Json.JsonElement grant,
        AuthRegistration registration, CancellationToken cancellation, string? nonce = null);
    private partial Task<HttpResponseMessage> SendAuthAsync(HttpMethod method, Uri endpoint,
        IReadOnlyDictionary<string, string>? form, CancellationToken cancellation);
    private static partial Task<System.Text.Json.JsonElement> ReadAuthJsonAsync(HttpContent content,
        CancellationToken cancellation);
    private partial Task<System.Text.Json.JsonElement> GetAuthDocumentAsync(string path, CancellationToken cancellation);
    private static partial int AccountIndex(AuthFile file, string label);
    public partial Task<ImmutableArray<AuthAccount>> AccountsAsync(CancellationToken cancellation = default);
    // Successful login chooses plan billing for future launches. Null preserves
    // the legacy API default. Logout does not authorize switching to an API key.
    public partial Task<string?> SelectedPlanAccountAsync(CancellationToken cancellation = default);
    // Clear local tokens while retaining account/client/host identity. Provider
    // revocation is a separate operation; local success does not claim revocation.
    public partial Task LogoutAsync(string label, CancellationToken cancellation = default);
    public partial void Destroy();

    internal partial Task<string> HostIDAsync(CancellationToken cancellation = default);
    // Auth calls this only after validating an OAuth grant, including identity.
    internal partial Task StoreAsync(AuthRegistration registration, CancellationToken cancellation = default);
    private partial Task<FileStream> LockFileAsync(CancellationToken cancellation);
    private partial AuthFile ReadFile();
    private partial void WriteFile(AuthFile value);
}

public sealed partial class Ensemble
{
    private readonly object credentialsGate = new();
    private Credentials? credentials;
    public partial Credentials? Credentials();
    internal partial void AddCredentials(Credentials child);
    internal partial void RemoveCredentials(Credentials child);
}
