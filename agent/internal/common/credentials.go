package common

import "context"

// Credentials: where a bearer token comes from.
//
// Every chapter up to this one treated the credential as a fact of
// configuration. An API key is a string. It is read once at startup, it never
// changes, and it can live in Config next to the model name and the base URL
// without anyone noticing that it is different in kind from its neighbours.
//
// An OAuth access token is not a fact. It expires in about an hour, it is
// replaced by a refresh exchange that may cross the network and fail, and the
// refresh token it is replaced with is itself replaced on every use. Nothing
// about that fits in a string field resolved at startup.
//
// The remedy is one interface with one interesting method. The code that
// builds an HTTP request asks for a bearer token and gets one. It does not
// learn whether the string it received was typed into a settings file two
// months ago or minted ninety seconds ago by a refresh it triggered itself.
// That ignorance is the entire point: it is what lets a static key and a
// rotating token travel the same path.
//
// Note what did NOT change. Config stays data. A provider is behaviour, so it
// does not go in Config; the engine holds it, resolves a token once per
// request, and hands the result to the renderer through the Config field that
// already existed. The renderers are untouched. Each one still knows where its
// vendor wants the credential — Authorization for OpenAI, x-api-key for
// Anthropic, a query parameter for Gemini — because that is vendor knowledge
// and it belongs with the vendor. What they no longer know is where the
// credential came from.

// CredentialKind names the sort of credential a provider holds.
//
// It exists for display and for policy — "which account am I signed in as",
// "is a metered key available as a fallback" — and never to branch on while
// building a request. A request built differently depending on this value
// would mean the abstraction had failed.
type CredentialKind int

const (
	// CredentialAPIKey is a static, long-lived secret supplied by the
	// operator. It does not expire and cannot be refreshed.
	CredentialAPIKey CredentialKind = iota + 1
	// CredentialChatGPTOAuth is an OAuth access token obtained from a
	// consumer ChatGPT plan. It expires, it refreshes, and the refresh
	// rotates.
	CredentialChatGPTOAuth
)

var credentialKindNames = map[CredentialKind]string{
	CredentialAPIKey:       "api_key",
	CredentialChatGPTOAuth: "chatgpt_oauth",
}

func (k CredentialKind) String() string {
	if s, ok := credentialKindNames[k]; ok {
		return s
	}
	return "unknown"
}

// CredentialProvider yields a bearer credential for an inference request.
//
// GetBearerToken returns a token that is valid at the moment it returns. A
// provider backed by a static key returns the same string forever. A provider
// backed by OAuth may refresh before returning, which is why the method can
// fail and why it takes a context.
//
// The context deserves a word, because the engine does not currently thread a
// request context through Turn and so passes context.Background(). The
// parameter is still right: a refresh is an outbound HTTP call, and a credential
// provider that cannot be cancelled or bounded is a provider that can hang a
// turn forever. Providers that make network calls apply their own timeout
// rather than trusting the caller to have supplied one.
type CredentialProvider interface {
	GetBearerToken(ctx context.Context) (string, error)
	Kind() CredentialKind
}

// APIKeyProvider is the trivial provider: it hands back the key it was given.
//
// This is deliberately the first implementation. Introducing an interface and
// immediately writing something complicated behind it hides whether the
// interface was the right shape, because every bug is then ambiguous between
// the abstraction and its first user. Wrapping the behaviour that already
// worked proves the seam is in the right place before anything depends on it:
// if every chapter before this one still passes with an APIKeyProvider
// installed, the seam costs nothing.
type APIKeyProvider struct {
	Key string
}

// NewAPIKeyProvider wraps a static key.
func NewAPIKeyProvider(key string) *APIKeyProvider {
	return &APIKeyProvider{Key: key}
}

// GetBearerToken returns the static key. It cannot fail, and it ignores the
// context, because there is nothing to wait for.
func (p *APIKeyProvider) GetBearerToken(context.Context) (string, error) {
	return p.Key, nil
}

// Kind reports that this is a static key.
func (p *APIKeyProvider) Kind() CredentialKind { return CredentialAPIKey }
