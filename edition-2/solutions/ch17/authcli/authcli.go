// Package authcli is a small public-API client shared by the terminal and GUI commands.
package authcli

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"ensemble/ensemble"
)

// Configure selects this application's private store without borrowing another
// application's profile. EN_HOME makes isolated launches and tests explicit.
func Configure(root *ensemble.Ensemble) error {
	directory := os.Getenv("EN_HOME")
	if directory == "" {
		home, err := os.UserConfigDir()
		if err != nil {
			return err
		}
		directory = filepath.Join(home, "ensemble")
	}
	root.ConfigureAuth(ensemble.AuthOptions{Directory: directory, Issuer: os.Getenv("OPENAI_OIDC_ISSUER")})
	return nil
}

// Run handles only explicit auth commands; provider/model flags belong to the
// invoking client. No token-bearing structure is printable through this API.
func Run(ctx context.Context, root *ensemble.Ensemble, args []string, out io.Writer) error {
	if len(args) == 0 {
		return errors.New("usage: auth login [--print-url] | status | logout | models")
	}
	switch args[0] {
	case "login":
		printURL := false
		for _, arg := range args[1:] {
			switch arg {
			case "--print-url":
				printURL = true
			case "openai-chatgpt":
			default:
				return errors.New("unknown auth login argument")
			}
		}
		ctx, cancel := context.WithTimeout(ctx, 5*time.Minute)
		defer cancel()
		status, err := root.Auth().Login(ctx, func(url string) error {
			if printURL {
				if _, err := fmt.Fprintln(out, url); err != nil {
					return err
				}
			}
			command := strings.Fields(os.Getenv("EN_BROWSER_CMD"))
			if len(command) == 0 {
				if runtime.GOOS == "darwin" {
					command = []string{"open"}
				} else {
					command = []string{"xdg-open"}
				}
			}
			// Argument vectors keep browser URLs out of shell evaluation and diagnostics.
			cmd := exec.CommandContext(ctx, command[0], append(command[1:], url)...)
			return cmd.Run()
		})
		if err != nil {
			return err
		}
		display(out, status)
		if !status.PlanEnabled {
			fmt.Fprintln(out, "Identity verified; ChatGPT plan permission was not granted. Inference is disabled.")
		}
		return nil
	case "status":
		if len(args) != 1 {
			return errors.New("auth status takes no arguments")
		}
		status, err := root.Auth().Status(ctx)
		if err != nil {
			return err
		}
		display(out, status)
		return nil
	case "logout":
		if len(args) != 1 {
			return errors.New("auth logout takes no arguments")
		}
		err := root.Auth().Logout(ctx)
		fmt.Fprintln(out, "Signed out locally. ChatGPT billing remains selected; sign in again or explicitly select api billing.")
		return err
	case "models":
		if len(args) != 1 {
			return errors.New("auth models takes no arguments")
		}
		base := os.Getenv("LLM_BASE_URL")
		if base == "" {
			base = "https://api.openai.com"
		}
		models, err := root.Auth().Models(ctx, base)
		if err != nil {
			return err
		}
		for _, m := range models {
			fmt.Fprintf(out, "%s\t%s\n", m.Slug, m.DisplayName)
		}
		return nil
	}
	return errors.New("unknown auth command")
}
func display(out io.Writer, status ensemble.AuthStatus) {
	billing := status.Billing
	if billing == "" {
		billing = ensemble.CredentialAPIKey
	}
	fmt.Fprintf(out, "Account: %s\nSubject: %s\nClient ID: %s\nBilling: %s\nSigned in: %t\nPlan enabled: %t\nExpires: %s\nScopes: %s\n", status.Email, status.Subject, status.ClientID, billing, status.SignedIn, status.PlanEnabled, status.ExpiresAt.Format(time.RFC3339), strings.Join(status.Scopes, " "))
}
