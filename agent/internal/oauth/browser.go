package oauth

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

// BrowserCmdEnvVar names the environment variable that overrides how a
// browser is launched.
//
// The value is a command line; the authorization URL is appended as the final
// argument. For example:
//
//	EN_BROWSER_CMD="firefox --private-window"
//	EN_BROWSER_CMD="wslview"
//
// This exists because "open a browser" has no portable answer. macOS has
// open(1), most Linux desktops have xdg-open, WSL needs a Windows host
// browser, and a user inside a container or a remote tmux may want something
// else entirely. Hardcoding one of them makes the agent unusable in the
// environments it is most likely to run in.
const BrowserCmdEnvVar = "EN_BROWSER_CMD"

// NoBrowserEnvVar forces the print-the-URL path without a code change.
//
// Set EN_NO_BROWSER=1 on a headless box or over SSH, where launching a
// browser either fails or — worse — succeeds on the WRONG machine, opening a
// consent screen on a server nobody is looking at while the user waits at
// their laptop for something to happen.
const NoBrowserEnvVar = "EN_NO_BROWSER"

// Opener starts the user's authorization journey by getting them to url.
//
// It is an interface-shaped function rather than a hardcoded call so that
// "open a browser", "print a URL", "show a QR code" and "drive a headless
// browser in a test" are all the same kind of thing to the flow. The flow
// does not care how the user gets there; it only cares that the callback
// eventually arrives.
type Opener func(ctx context.Context, url string) error

// PrintURLOpener returns an Opener that writes the URL to w instead of
// launching anything.
//
// This is the headless and SSH path, and it is a first-class way to run the
// flow rather than a degraded one: the user copies the URL into a browser on
// whatever machine actually has a display. The loopback listener is still on
// 127.0.0.1 of the agent's host, so the user must be able to reach that host
// — which is the normal case for a local terminal or an SSH session with a
// forwarded port.
func PrintURLOpener(w io.Writer) Opener {
	return func(_ context.Context, url string) error {
		_, err := fmt.Fprintf(w, "\nOpen this URL in a browser to sign in with ChatGPT:\n\n  %s\n\nWaiting for the callback...\n", url)
		return err
	}
}

// CommandOpener returns an Opener that launches a browser, honouring
// EN_BROWSER_CMD and falling back to the platform default.
//
// On failure it falls back to printing the URL to fallback rather than
// aborting the sign-in. A failed exec is not a reason to make the user start
// over: the URL is still perfectly usable by hand, and the listener is
// already waiting.
func CommandOpener(fallback io.Writer) Opener {
	return func(ctx context.Context, url string) error {
		name, args, ok := browserCommand()
		if ok {
			// The URL is passed as a separate argv element, never
			// interpolated into a shell string. It contains attacker-
			// influenceable material only indirectly, but a URL carrying a
			// quote or a semicolon into `sh -c` is a command-injection bug
			// that is entirely avoidable by not invoking a shell.
			cmd := exec.CommandContext(ctx, name, append(args, url)...)
			if err := cmd.Start(); err == nil {
				// Reap the child so it does not become a zombie, but do not
				// block on it: `open` returns immediately while `firefox`
				// may run for hours, and waiting would deadlock the flow
				// against the very browser it is waiting for.
				go func() { _ = cmd.Wait() }()
				if fallback != nil {
					fmt.Fprintf(fallback, "\nOpening a browser to sign in with ChatGPT.\nIf nothing opens, use this URL:\n\n  %s\n\n", url)
				}
				return nil
			}
		}
		if fallback == nil {
			return fmt.Errorf("oauth: no way to open a browser for the authorization URL")
		}
		return PrintURLOpener(fallback)(ctx, url)
	}
}

// browserCommand resolves the command used to open a URL.
func browserCommand() (name string, args []string, ok bool) {
	if custom := strings.TrimSpace(os.Getenv(BrowserCmdEnvVar)); custom != "" {
		// strings.Fields is a deliberate simplification: it splits on
		// whitespace and does not honour quoting. Paths with spaces are
		// handled by pointing the variable at a wrapper script, which is
		// clearer than half-implementing shell quoting rules.
		parts := strings.Fields(custom)
		return parts[0], parts[1:], true
	}
	switch runtime.GOOS {
	case "darwin":
		return "open", nil, true
	case "windows":
		return "rundll32", []string{"url.dll,FileProtocolHandler"}, true
	default:
		return "xdg-open", nil, true
	}
}

// headlessRequested reports whether the environment is asking us not to
// launch a browser.
//
// EN_NO_BROWSER is explicit. The SSH heuristic covers the common case where
// the user has not thought about it: if we are in an SSH session on a
// Unix-like host with no DISPLAY and no Wayland socket, there is no browser
// to launch and attempting one produces a confusing error instead of a
// usable URL. macOS is excluded because open(1) works there regardless of
// DISPLAY.
func headlessRequested() bool {
	if v := strings.TrimSpace(os.Getenv(NoBrowserEnvVar)); v != "" && v != "0" && !strings.EqualFold(v, "false") {
		return true
	}
	if runtime.GOOS == "darwin" || runtime.GOOS == "windows" {
		return false
	}
	inSSH := os.Getenv("SSH_CONNECTION") != "" || os.Getenv("SSH_TTY") != ""
	noDisplay := os.Getenv("DISPLAY") == "" && os.Getenv("WAYLAND_DISPLAY") == ""
	return inSSH && noDisplay
}
