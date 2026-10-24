package ensemble

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

// These are application-owned overrides, never conversation configuration.
func AuthOptionsFromEnv(get func(string) string) AuthOptions {
	if get == nil {
		get = func(string) string { return "" }
	}
	return AuthOptions{Directory: get("EN_HOME"), Issuer: get("OPENAI_OIDC_ISSUER"), RedirectURI: get("OPENAI_OAUTH_REDIRECT_URI"), ModelsURL: get("OPENAI_MODELS_ENDPOINT")}
}

// RunAuthCommand is shared by cmd/ch01 and gui/cmd, before Agent construction.
func RunAuthCommand(ctx context.Context, a Auth, args []string, out io.Writer, open func(string) error) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: auth login [--print-url] | status | logout [--all|ACCOUNT] | models")
	}
	switch args[0] {
	case "login":
		if len(args) != 1 && !(len(args) == 2 && args[1] == "--print-url") {
			return fmt.Errorf("usage: auth login [--print-url]")
		}
		if open == nil {
			open = func(u string) error {
				if browser := os.Getenv("EN_BROWSER_CMD"); browser != "" {
					// The operator owns this command. The URL is a positional
					// argument, not shell source (EN_BROWSER_CMD=true is headless).
					_ = exec.CommandContext(ctx, "/bin/sh", "-c", browser+" \"$1\"", "ensemble-browser", u).Run()
				} else {
					command := "xdg-open"
					if runtime.GOOS == "darwin" {
						command = "open"
					}
					if path, err := exec.LookPath(command); err == nil {
						_ = exec.CommandContext(ctx, path, u).Run()
					}
				}
				return nil // Manual navigation remains possible if launch fails.
			}
		}
		st, err := a.Login(ctx, func(u string) error {
			// Keep the URL available even if launch is suppressed or unavailable.
			// --print-url explicitly selects this documented form of login.
			fmt.Fprintln(out, "Open this authorization URL in a browser:\n"+u)
			return open(u)
		})
		if err != nil {
			return err
		}
		fmt.Fprintf(out, "ChatGPT identity verified for %s; inference eligible=%t; billing=%s\n", st.Account, st.Eligible, st.Mode)
		if st.Reason != "" {
			fmt.Fprintln(out, st.Reason)
		}
		return nil
	case "status":
		if len(args) != 1 {
			return fmt.Errorf("usage: auth status")
		}
		st, err := a.Status(ctx)
		if err != nil {
			return err
		}
		fmt.Fprintf(out, "billing=%s account=%s eligible=%t", st.Mode, st.Account, st.Eligible)
		if !st.ExpiresAt.IsZero() {
			fmt.Fprintf(out, " expires=%s", st.ExpiresAt.UTC().Format("2006-01-02T15:04:05Z"))
		}
		if st.Reason != "" {
			fmt.Fprintf(out, " (%s)", st.Reason)
		}
		fmt.Fprintln(out)
		return nil
	case "logout":
		if len(args) > 2 {
			return fmt.Errorf("usage: auth logout [--all|ACCOUNT]")
		}
		account := ""
		all := false
		if len(args) == 2 {
			all = args[1] == "--all"
			if !all {
				if strings.HasPrefix(args[1], "--") {
					return fmt.Errorf("unknown logout option")
				}
				account = args[1]
			}
		}
		err := a.Logout(ctx, account, all)
		if err != nil {
			return err
		}
		fmt.Fprintln(out, "Local ChatGPT tokens removed; API billing was not selected.")
		return nil
	case "models":
		if len(args) != 1 {
			return fmt.Errorf("usage: auth models")
		}
		models, err := a.Models(ctx, Config{AuthMode: AuthChatGPT})
		if err != nil {
			return err
		}
		for _, m := range models {
			fmt.Fprintf(out, "%s\t%s\n", m.Slug, m.Name)
		}
		return nil
	default:
		return fmt.Errorf("unknown auth command %q", args[0])
	}
}
