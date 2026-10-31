namespace Ensemble;

// Value-like immutable protocol snapshot, not a service or another credential
// authority. Credentials owns its lifetime. The old registration is retained
// only to detect intervening writes (including another process's logout). No
// generated ToString may disclose the verifier or saved account tokens.
internal sealed partial class AuthLogin
{
    internal readonly string Label;
    internal readonly Uri Redirect;
    internal readonly string State;
    internal readonly string Nonce;
    internal readonly string Verifier;
    internal readonly DateTimeOffset ExpiresAt;
    internal readonly AuthRegistration? Before;
    internal partial AuthLogin(string label, Uri redirect, AuthRegistration? before);
}

public sealed partial class Credentials
{
    // The host starts a listener at this exact loopback URI before presenting the
    // returned browser URL. This operation never opens a browser. URLs may carry
    // an ID-token hint: display/open them locally, never put them in History/logs.
    // A second Begin refuses until completion or explicit cancellation.
    public partial Task<Uri> BeginLoginAsync(string label, Uri redirect,
        bool requestConsent = false, CancellationToken cancellation = default);
    // Accept the complete callback URI, not just an authorization code. No
    // exchange or registration write happens before endpoint/state validation.
    public partial Task<AuthAccount> CompleteLoginAsync(string callback,
        CancellationToken cancellation = default);
    public partial Task CancelLoginAsync(CancellationToken cancellation = default);
    // Hosts use this after callback refusal: an unsolicited callback leaves the
    // attempt pending; a matched denial or exchange failure consumes it. No
    // state/nonce/token copy is exposed. Expired attempts are no longer pending.
    public partial Task<bool> LoginPendingAsync(CancellationToken cancellation = default);
    private static partial Dictionary<string, string> ReadCallback(string callback, AuthLogin attempt);
    private static partial string DecodeCallbackPart(string text);
}
