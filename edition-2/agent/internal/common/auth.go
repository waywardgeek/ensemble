package common

import (
	"context"
	"time"
)

// CredentialKind reports the funding identity without exposing a credential.
type CredentialKind string

const (
	// CredentialAPIKey selects the explicitly configured metered key.
	CredentialAPIKey CredentialKind = "api"
	// CredentialChatGPT selects the registered subscription session.
	CredentialChatGPT CredentialKind = "chatgpt"
	// CredentialCodex explicitly selects an existing Codex sign-in and endpoint.
	CredentialCodex CredentialKind = "codex"
)

// CredentialProvider keeps credential renewal out of rendering and conversation.
type CredentialProvider interface {
	// GetBearerToken returns usable transport-only credentials or a safe error.
	GetBearerToken(context.Context) (string, error)
	// Kind identifies the billing source without returning secret material.
	Kind() CredentialKind
}

// CodexCredentials reads an access token and account from the same file version.
// The external Codex application owns renewal; consumers never write its store.
type CodexCredentials interface {
	CredentialProvider
	// Ensemble is the actual owner of this credential reader.
	Ensemble() Ensemble
	// Read returns a request-local bearer/account pair, or a secret-free error.
	Read(context.Context) (bearer, account string, err error)
}

// AuthOptions identifies this application's private store and trusted issuer.
type AuthOptions struct {
	// Directory stores only this app's protected registration and token records.
	Directory string
	// Issuer defaults to OpenAI; loopback HTTP supports external protocol fakes.
	Issuer string
}

// AuthStatus is safe to display; it deliberately cannot carry token material.
type AuthStatus struct {
	// ClientID retains the registration after expiry or logout.
	ClientID string
	// Subject is the validated account identity, never inferred from email.
	Subject string
	// Email is display metadata, not identity.
	Email string
	// HostID remains stable across sign-ins on this installation.
	HostID string
	// ExpiresAt describes the current access token's recorded lifetime.
	ExpiresAt time.Time
	// Scopes records the token response's actual grants.
	Scopes []string
	// SignedIn means tokens are present; PlanEnabled additionally requires permission.
	SignedIn bool
	// PlanEnabled distinguishes identity-only login from inference consent.
	PlanEnabled bool
	// Billing preserves the user's explicit login choice across launches and failures.
	Billing CredentialKind
}

// AuthStore is an Ensemble-owned credential service shared by its Agents.
type AuthStore interface {
	CredentialProvider
	// Ensemble exposes the actual owner, including application diagnostics.
	Ensemble() Ensemble
	// Login launches one fresh attempt; opening its URL is a client responsibility.
	Login(context.Context, func(string) error) (AuthStatus, error)
	// Status reads safe account information without refreshing any credentials.
	Status(context.Context) (AuthStatus, error)
	// Logout attempts revocation then clears local tokens even on remote failure.
	Logout(context.Context) error
	// Refresh rejects a previously failed access token without racing a newer one.
	Refresh(context.Context, string) error
	// Models lists visible account-specific models with the active OAuth bearer.
	Models(context.Context, string) ([]ModelOption, error)
}

// BillingSelected announces the chosen funding source before inference spends.
type BillingSelected struct {
	// Source distinguishes subscription use from metered API requests.
	Source CredentialKind `json:"source"`
}

func (BillingSelected) observation() {}

// ModelOption preserves the distinction between request identity and UI label.
type ModelOption struct {
	// Slug is sent unchanged to inference.
	Slug string
	// DisplayName is the catalog label shown to a person.
	DisplayName string
}
