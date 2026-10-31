package grade

// Edition2OAuthOptions selects the existing signed OAuth peer's protocol cases.
// This alias lets the revised grader reuse it without changing the original run.
type Edition2OAuthOptions = ch19OAuthOptions

// Edition2OAuthObservations is a detached snapshot of actual issuer exchanges.
// Its token fields contain only credentials minted by the local grading peer.
type Edition2OAuthObservations = ch19OAuthObs

// Edition2ResponsesOptions selects the original peer's captured Responses shapes.
type Edition2ResponsesOptions = ch19RespOptions

// Edition2ResponsesObservations preserves the original peer's copied wire report.
type Edition2ResponsesObservations = ch19RespObs

// Edition2APIKey identifies the original peer's metered route. It is a public
// fake canary, never an application credential or provider configuration default.
const Edition2APIKey = ch19APIKey

// NewEdition2OAuthPeer starts the unchanged issuer, including real RSA signing
// and PKCE validation. The caller owns Close and the isolated client profile.
func NewEdition2OAuthPeer(o Edition2OAuthOptions) *ch19FakeOAuth {
	s := ch19NewFakeOAuth(o)
	// The original constructor starts serving before assigning srv. Publish
	// that final assignment through the mutex every handler acquires before
	// reading it, before this adapter exposes the URL to an external child.
	s.mu.Lock()
	s.mu.Unlock()
	return s
}

// NewEdition2ResponsesPeer starts the unchanged protocol peer; the caller must
// close it after all submitted-client requests have finished.
func NewEdition2ResponsesPeer(o Edition2ResponsesOptions) *ch19FakeResponses {
	return ch19NewFakeResponses(o)
}
