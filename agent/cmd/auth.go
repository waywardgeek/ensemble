package main

// The `auth` subcommand: sign in with ChatGPT, inspect the stored grant, or
// throw it away.
//
// Nothing here prints a token. `auth status` exists so an operator can answer
// "am I signed in, as whom, and until when" without reaching for the
// credential file, and a status command that leaked the credential would
// defeat its own purpose.

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/waywardgeek/ensemble/agent/internal/common"
	"github.com/waywardgeek/ensemble/agent/internal/llm"
	"github.com/waywardgeek/ensemble/agent/internal/oauth"
)

// attachCredentials gives the engine an OAuth provider when the operator has
// signed in, and otherwise leaves it alone so the API key keeps working.
//
// The choice is made here, in the host, rather than inside the agent library.
// Reading the operator's home directory is a policy decision belonging to the
// program the operator launched; a library that went looking for credentials
// on its own would be making that decision for every embedder.
//
// The provider is registered against the vendor it belongs to, not merely
// installed when that vendor happens to be the one in force at startup. An
// operator who starts on Anthropic and switches to OpenAI from the model
// picker is still entitled to their plan, so resolving every vendor up front
// turns the switch into a lookup — the same shape endpoints already use.
// Gating on the startup vendor instead left the provider unattached for the
// rest of the session, quietly billing the metered key.
func attachCredentials(eng *llm.Engine, cfg common.Config) {
	p := openAIProvider()
	if p == nil {
		return // not signed in: the API key path is unchanged
	}
	if eng.CredsByVendor == nil {
		eng.CredsByVendor = make(map[common.Vendor]common.CredentialProvider)
	}
	eng.CredsByVendor[common.VendorOpenAI] = p
	if cfg.Vendor == common.VendorOpenAI {
		eng.Creds = p
	}
}

// openAIProvider loads the stored ChatGPT-plan credentials, returning nil when
// the operator has not signed in or the stored credentials cannot be used.
func openAIProvider() common.CredentialProvider {
	acfg, err := authConfig()
	if err != nil {
		return nil
	}
	creds, err := acfg.Store.Load()
	if err != nil || creds == nil {
		return nil // not signed in: the API key path is unchanged
	}
	p, err := oauth.NewProvider(acfg)
	if err != nil {
		fmt.Fprintln(os.Stderr, "auth: stored credentials unusable:", err)
		return nil
	}
	return p
}

func runAuth(sub string, printURL bool) error {
	switch sub {
	case "login":
		return authLogin(printURL)
	case "status":
		return authStatus(os.Stdout)
	case "logout":
		return authLogout(os.Stdout)
	case "":
		return fmt.Errorf("expected a subcommand: login, status, or logout")
	default:
		return fmt.Errorf("unknown subcommand %q: expected login, status, or logout", sub)
	}
}

func authConfig() (oauth.Config, error) {
	credPath, err := oauth.DefaultCredentialPath()
	if err != nil {
		return oauth.Config{}, err
	}
	hostIDPath, err := oauth.DefaultHostIDPath()
	if err != nil {
		return oauth.Config{}, err
	}
	return oauth.Config{
		Issuer:     oauth.Issuer(),
		AgentName:  "ensemble",
		Store:      oauth.NewStore(credPath),
		HostIDPath: hostIDPath,
		Output:     os.Stdout,
	}, nil
}

func authLogin(printURL bool) error {
	cfg, err := authConfig()
	if err != nil {
		return err
	}
	if printURL {
		// The operator opens the URL somewhere else. The callback still
		// arrives on this machine's loopback listener, so the flow completes
		// only if the browser can reach it.
		cfg.NoBrowser = true
		cfg.Opener = oauth.PrintURLOpener(os.Stdout)
	} else {
		cfg.Opener = oauth.CommandOpener(os.Stdout)
	}

	flow, err := oauth.NewFlow(cfg)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	creds, err := flow.Authorize(ctx)
	if err != nil {
		return err
	}
	fmt.Fprintf(os.Stdout, "signed in as %s\n", authWho(creds))
	if !creds.HasDirectTokenScope() {
		// A grant without the direct-token scope authenticates the operator
		// but authorizes nothing. Saying so now beats a puzzling 403 on the
		// first turn.
		fmt.Fprintln(os.Stdout,
			"warning: this grant does not carry the plan-usage scope, so inference will be refused")
	}
	return nil
}

func authStatus(w io.Writer) error {
	cfg, err := authConfig()
	if err != nil {
		return err
	}
	creds, err := cfg.Store.Load()
	if err != nil || creds == nil {
		fmt.Fprintln(w, "not signed in")
		return nil
	}
	fmt.Fprintf(w, "signed in as %s\n", authWho(creds))
	fmt.Fprintf(w, "issuer:  %s\n", creds.Issuer)
	fmt.Fprintf(w, "scopes:  %s\n", strings.Join(creds.GrantedScopes, " "))
	if creds.ExpiresAt.IsZero() {
		fmt.Fprintln(w, "expires: unknown")
	} else {
		fmt.Fprintf(w, "expires: %s (in %s)\n",
			creds.ExpiresAt.Format(time.RFC3339),
			time.Until(creds.ExpiresAt).Round(time.Second))
	}
	if !creds.HasDirectTokenScope() {
		fmt.Fprintln(w, "warning: no plan-usage scope; inference will be refused")
	}
	return nil
}

func authLogout(w io.Writer) error {
	cfg, err := authConfig()
	if err != nil {
		return err
	}
	if err := cfg.Store.Clear(); err != nil {
		return err
	}
	fmt.Fprintln(w, "signed out")
	return nil
}

// authWho names the operator without printing anything secret.
func authWho(c *oauth.Credentials) string {
	switch {
	case c.Email != "":
		return c.Email
	case c.Subject != "":
		return c.Subject
	default:
		return "an unnamed account"
	}
}
