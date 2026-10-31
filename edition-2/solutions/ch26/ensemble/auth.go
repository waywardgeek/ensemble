package ensemble

import (
	"ensemble/internal/auth"
	"ensemble/internal/common"
)

// AuthOptions identifies this application's credential directory and issuer.
type AuthOptions = common.AuthOptions

// AuthStatus exposes account metadata without token-bearing fields.
type AuthStatus = common.AuthStatus

// AuthStore is the public login, renewal and sign-out service.
type AuthStore = common.AuthStore

// CredentialProvider separates authentication from rendered conversation state.
type CredentialProvider = common.CredentialProvider

// CredentialKind identifies billing without revealing authentication material.
type CredentialKind = common.CredentialKind

const (
	// CredentialAPIKey explicitly selects metered API-key billing.
	CredentialAPIKey = common.CredentialAPIKey
	// CredentialChatGPT explicitly selects the saved subscription session.
	CredentialChatGPT = common.CredentialChatGPT
	// CredentialCodex selects the Codex endpoint with its existing sign-in.
	CredentialCodex = common.CredentialCodex
)

// ConfigureAuth installs the Ensemble-owned service before any Agent is created.
// Applications choose the private path; the library does not inspect other apps.
func (e *Ensemble) ConfigureAuth(options AuthOptions) {
	if len(e.agents) != 0 {
		panic("configure authentication before constructing Agents")
	}
	e.auth = auth.New(e, options)
}

// Auth returns the optional service installed by ConfigureAuth.
func (e *Ensemble) Auth() AuthStore { return e.auth }

// BillingSelected announces funding before the first inference request.
type BillingSelected = common.BillingSelected
