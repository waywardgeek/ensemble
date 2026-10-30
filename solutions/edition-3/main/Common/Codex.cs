namespace Ensemble;

public sealed partial class Ensemble
{
    // A separate strong child: Codex owns its grant, while our public OAuth
    // Credentials child owns a different registration and renewal protocol.
    // The existing credentialsGate protects both optional relationships.
    private CodexCredentials? codexCredentials;
    public partial CodexCredentials? CodexCredentials();
    internal partial void AddCodexCredentials(CodexCredentials child);
    internal partial void RemoveCodexCredentials(CodexCredentials child);
}

/// <summary>Ensemble-owned read-only source of Codex's file-backed ChatGPT grant.
/// Construction selects a path, never discovers profiles or reads credentials.
/// Every send reads a fresh pair; Codex alone signs in, rotates and revokes it.</summary>
public sealed partial class CodexCredentials
{
    public const string ServiceBase = "https://chatgpt.com/backend-api/codex";
    public const string ResponsesEndpoint = ServiceBase + "/responses";
    internal Ensemble? parent;
    internal bool destroyed;
    private readonly string path;
    private readonly SemaphoreSlim gate = new(1, 1);
    private readonly CancellationTokenSource lifetime = new();

    public partial CodexCredentials(Ensemble parent, string path);
    public partial Ensemble Parent();
    // Internal secret boundary. Tokens never become configuration or History.
    internal partial Task<CodexGrant> ReadAsync(CancellationToken cancellation);
    // Cancel/join active reads, then detach. Never delete or rewrite the file.
    public partial void Destroy();
    internal static partial void ValidateEndpoint(Uri endpoint);
}

/// <summary>Value-like, exchange-local secret pair from one file read. Not a
/// record: compiler-generated formatting must not expose secret fields.</summary>
internal sealed partial class CodexGrant
{
    internal readonly string Bearer;
    internal readonly string Account;
    internal partial CodexGrant(string bearer, string account);
}

/// <summary>Safe failure narration shared by direct callers and actor History.
/// Original file/JSON/provider exceptions are deliberately not retained.</summary>
public sealed partial class CodexAuthorizationException : InvalidOperationException
{
    public partial CodexAuthorizationException();
}
