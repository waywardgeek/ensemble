package auth

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"strings"

	"ensemble/internal/common"
)

// The reader owns no token cache. A single read keeps account and bearer paired,
// and the next request observes any atomic replacement made by Codex itself.
type codex struct {
	parent common.Ensemble
	path   string
}

// NewCodex constructs a read-only credential reader owned by the application.
// Only the application chooses a path; the library never searches other profiles.
func NewCodex(parent common.Ensemble, path string) common.CodexCredentials {
	if parent == nil {
		panic("Codex credentials require Ensemble")
	}
	return &codex{parent: parent, path: path}
}

// Ensemble exposes diagnostics through the actual owner.
func (c *codex) Ensemble() common.Ensemble { return c.parent }

// Kind distinguishes this route from Ensemble's separately registered OAuth grant.
func (c *codex) Kind() common.CredentialKind { return common.CredentialCodex }

// GetBearerToken satisfies the shared credential boundary; transport uses Read
// so its account header cannot come from a different credential-file generation.
func (c *codex) GetBearerToken(ctx context.Context) (string, error) {
	token, _, err := c.Read(ctx)
	return token, err
}

// Read loads only the file-backed ChatGPT sign-in shape, without returning JSON
// errors or file contents that could expose secrets. Codex owns token renewal.
func (c *codex) Read(ctx context.Context) (string, string, error) {
	if err := ctx.Err(); err != nil {
		return "", "", err
	}
	data, err := os.ReadFile(c.path)
	if err != nil {
		return "", "", errors.New("cannot read Codex auth file; sign in with Codex and check EN_CODEX_AUTH_FILE")
	}
	var value struct {
		// Mode distinguishes a ChatGPT sign-in from an API-key Codex profile.
		Mode string `json:"auth_mode"`
		// Tokens keeps the account and bearer in one file generation.
		Tokens struct {
			// Access is used only for the current HTTP request.
			Access string `json:"access_token"`
			// Account selects the account associated with Access.
			Account string `json:"account_id"`
		} `json:"tokens"`
	}
	if json.Unmarshal(data, &value) != nil || (value.Mode != "" && value.Mode != "chatgpt") ||
		strings.TrimSpace(value.Tokens.Access) == "" || strings.TrimSpace(value.Tokens.Account) == "" ||
		strings.ContainsAny(value.Tokens.Access+value.Tokens.Account, "\r\n") {
		return "", "", errors.New("Codex auth file has no usable ChatGPT token/account pair; sign in with Codex")
	}
	return value.Tokens.Access, value.Tokens.Account, nil
}
