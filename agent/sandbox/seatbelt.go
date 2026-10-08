package sandbox

// The kernel half of the boundary.
//
// A userspace path check cannot confine a subprocess. `cat ../../etc/passwd`
// never calls our Go code, so resolveSandboxedPath is not merely weak
// against a shell, it is absent. The operating system has to enforce this
// one.
//
// On macOS that means Seatbelt, reached through sandbox-exec. The approach
// is the same one OpenAI's Codex CLI ships: wrap the CHILD, leave the host
// binary outside. Wrapping the child is what lets the profile deny network
// absolutely. If the host binary were inside the sandbox it would need
// network access to reach the model API, and because macOS sandboxes are
// inherited by descendants, every command the agent ran would inherit that
// access too, which hands exfiltration straight back.
//
// Everything in this file was measured on macOS 26.3.1 (arm64). The
// comments record WHY each rule is present, because every one of them was
// added to fix an observed failure and would otherwise look like cargo
// cult.

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// ErrUnsupportedPlatform is returned when confinement is requested on a
// platform with no implementation here.
//
// This FAILS CLOSED. Running the command unconfined because the host
// happens to be Linux would be the worst possible behavior: the caller
// asked for a boundary, got none, and received no error saying so. A
// Linux implementation belongs behind Landlock and seccomp; until it
// exists, the honest answer is a refusal.
var ErrUnsupportedPlatform = fmt.Errorf("sandbox: command confinement is not implemented on %s", runtime.GOOS)

// Command returns an exec.Cmd that runs name with args confined to the
// sandbox.
//
// When unconfined, this is exactly exec.Command and nothing is wrapped.
// When confined, the command is wrapped in the platform sandbox, its
// working directory is the sandbox root, and its environment has been
// stripped of credentials.
func (s *Sandbox) Command(name string, args ...string) (*exec.Cmd, error) {
	root, confined, err := s.root()
	if err != nil {
		return nil, err
	}
	if !confined {
		return exec.Command(name, args...), nil
	}
	if runtime.GOOS != "darwin" {
		return nil, ErrUnsupportedPlatform
	}
	if _, err := exec.LookPath(seatbeltBinary); err != nil {
		return nil, fmt.Errorf("sandbox: %s not found, refusing to run unconfined: %w", seatbeltBinary, err)
	}

	profile, err := s.Profile()
	if err != nil {
		return nil, err
	}
	// Git is pointed at a config inside the sandbox, so the file has to be
	// there. A failure here is not fatal: everything except git still
	// works, and refusing to run any command because git would be
	// unconfigured would be a worse trade than running with git
	// unconfigured.
	_ = WriteGitConfig(root)
	// -p takes the profile inline, which avoids writing a profile file and
	// then having to decide who deletes it and when. The command outlives
	// this function (output streams back over a PTY), so a temp file would
	// need a lifetime tied to process exit.
	full := append([]string{"-p", profile, name}, args...)
	cmd := exec.Command(seatbeltBinary, full...)
	cmd.Dir = root
	cmd.Env = SanitizedEnv(os.Environ(), root)
	return cmd, nil
}

const seatbeltBinary = "sandbox-exec"

// Available reports whether command confinement can be enforced here.
func Available() bool {
	if runtime.GOOS != "darwin" {
		return false
	}
	_, err := exec.LookPath(seatbeltBinary)
	return err == nil
}

// Profile renders the Seatbelt profile confining a command to the sandbox.
//
// The shape is `(allow default)` followed by targeted denials, not
// `(deny default)` followed by allowances. That is not a style preference.
// MEASURED: a profile beginning `(deny default)` fails before main with
// "execvp() of '/bin/cat' failed: Operation not permitted", and importing
// the system bsd.sb profile does not rescue it. Codex uses the same
// allow-then-deny shape.
func (s *Sandbox) Profile() (string, error) {
	root, confined, err := s.root()
	if err != nil {
		return "", err
	}
	if !confined {
		return "", fmt.Errorf("sandbox: no profile for an unconfined sandbox")
	}

	var b strings.Builder
	w := func(format string, a ...any) { fmt.Fprintf(&b, format+"\n", a...) }

	w("(version 1)")
	w("(allow default)")
	w("(deny file-read*)")
	w("(deny file-write*)")

	// Metadata (stat) stays allowed. Denying it breaks far more than it
	// protects: a stat reveals that a path exists, which is a weak leak,
	// while shells and build tools stat constantly.
	w("(allow file-read-metadata)")

	// The root directory ENTRY must be readable or nothing starts at all.
	// MEASURED: without it the system log shows `deny(1) file-read-data /`
	// and every exec dies. This is the single least obvious line here: the
	// denial is on "/" itself, not on any library under it.
	w(`(allow file-read* (literal "/"))`)

	// System paths, for the dynamic loader and the standard tools.
	w(`(allow file-read* %s)`, subpaths(
		"/usr", "/bin", "/sbin", "/System", "/Library",
		"/private/var/db/dyld", "/dev",
		// TLS trust store and openssl.cnf. Without it anything using
		// LibreSSL dies reading /private/etc/ssl/openssl.cnf, which looks
		// like a network error and is not.
		"/private/etc/ssl",
	))

	// Writable device files. git in particular opens /dev/null for
	// READING AND WRITING and fails outright without this.
	w(`(allow file-write* %s)`, literals("/dev/null", "/dev/tty", "/dev/stdout", "/dev/stderr"))

	// The Xcode shim cache.
	//
	// Every developer tool in /usr/bin (git, python3, clang) is an xcrun
	// shim that reads and WRITES a cache in the Darwin per-user temp
	// directory, and fails hard without it.
	//
	// This is scoped by REGEX to files named xcrun_db*, not by subpath to
	// the directory. That distinction is the whole point. MEASURED:
	// allowing the directory wholesale also grants read and write to
	// everything else living there, and `mktemp -d` creates its
	// directories inside it, so the allowance silently covers other
	// processes' temporary files. The shim also creates randomly suffixed
	// cache files (xcrun_db-eozhFtQb), so a literal cannot match and a
	// pattern is required.
	if dut := darwinUserTempDir(); dut != "" {
		w(`(allow file-read* file-write* (regex #"^%s/xcrun_db"))`, regexpQuoteDir(dut))
	}

	// The workspace itself.
	w(`(allow file-read* (subpath "%s"))`, root)
	w(`(allow file-write* (subpath "%s"))`, root)

	// The most valuable line in the file. A confined command that cannot
	// reach the network cannot exfiltrate anything it reads, which is what
	// breaks the lethal trifecta for subprocesses.
	w("(deny network*)")

	return b.String(), nil
}

// subpaths renders a list of (subpath "...") terms.
func subpaths(paths ...string) string {
	parts := make([]string, 0, len(paths))
	for _, p := range paths {
		parts = append(parts, fmt.Sprintf("(subpath %q)", p))
	}
	return strings.Join(parts, " ")
}

// literals renders a list of (literal "...") terms.
func literals(paths ...string) string {
	parts := make([]string, 0, len(paths))
	for _, p := range paths {
		parts = append(parts, fmt.Sprintf("(literal %q)", p))
	}
	return strings.Join(parts, " ")
}

// darwinUserTempDir returns the per-user temp directory, symlink-resolved.
//
// The resolution is mandatory, not tidiness. SBPL matches PHYSICAL paths:
// the system reports this directory as /var/folders/..., /var is a symlink
// to /private/var, and MEASURED, a rule written with the /var form NEVER
// matches. Trailing slashes, by contrast, are irrelevant: all four
// combinations were tested and only the realpath mattered.
//
// This is the same bug as an uncanonicalized sandbox root in Resolve.
// Compare canonical paths on both sides, or compare neither.
func darwinUserTempDir() string {
	dir := os.TempDir()
	resolved, err := filepath.EvalSymlinks(dir)
	if err != nil {
		return ""
	}
	return strings.TrimRight(resolved, string(filepath.Separator))
}

// regexpQuoteDir escapes a directory path for use inside an SBPL regex.
func regexpQuoteDir(dir string) string {
	r := strings.NewReplacer(
		`\`, `\\`, `.`, `\.`, `+`, `\+`, `*`, `\*`, `?`, `\?`,
		`(`, `\(`, `)`, `\)`, `[`, `\[`, `]`, `\]`, `{`, `\{`, `}`, `\}`,
		`^`, `\^`, `$`, `\$`, `|`, `\|`,
	)
	return r.Replace(dir)
}

// --- environment confinement ----------------------------------------------

// sensitiveSuffixes name environment variables that tend to carry
// credentials. Suffix matching rather than an exact list, because the next
// vendor's key will be called something this file has never heard of.
var sensitiveSuffixes = []string{
	"_KEY", "_SECRET", "_TOKEN", "_PASSWORD", "_PASSWD", "_CREDENTIALS",
}

// sensitiveNames are exact matches the suffix rule would miss.
var sensitiveNames = []string{
	"LLM_API_KEY",
	"OPENAI_API_KEY",
	"ANTHROPIC_API_KEY",
	"GOOGLE_API_KEY",
	// A forwarded SSH agent socket is a live credential: whatever can
	// reach the socket can authenticate as the user without ever seeing a
	// private key. Stripping the address is the whole mitigation.
	"SSH_AUTH_SOCK",
}

// SanitizedEnv builds the environment for a command running inside the
// sandbox: the caller's environment minus anything that looks like a
// credential, with HOME pointed at the sandbox root.
//
// This closes one of the three paths a credential can take into the
// context window. The model runs `env`, or `curl -H "Authorization:
// $OPENAI_API_KEY"`, and the output comes back as a tool result. A tool
// result IS context.
//
// HOME is the interesting one. It is not a credential, it is the ADDRESS
// of every credential on the machine: ~/.ssh/id_rsa, ~/.aws/credentials,
// ~/.config/gh/hosts.yml. Pointing it at the sandbox turns the most common
// exfiltration one-liner into a read of a file that is not there.
func SanitizedEnv(environ []string, root string) []string {
	out := make([]string, 0, len(environ))
	for _, kv := range environ {
		eq := strings.Index(kv, "=")
		if eq < 0 {
			continue
		}
		name := kv[:eq]
		if isSensitiveEnv(name) || name == "HOME" || isGitConfigEnv(name) {
			continue
		}
		out = append(out, kv)
	}
	out = append(out, "HOME="+root)

	// Git is the one tool that reads a credential store as a matter of
	// routine, so it gets named handling rather than being left to HOME.
	//
	// A real ~/.gitconfig is not merely a preferences file. It can carry
	// url.<base>.insteadOf rewrites with a token embedded in the URL, and
	// credential.helper = store points at ~/.git-credentials, which is
	// plaintext. Allowing the host config into the sandbox to make git
	// work would hand over exactly the thing the sandbox exists to keep
	// back, and it is the obvious fix, which is what makes it dangerous.
	//
	// Pointing HOME at the sandbox already redirects the lookup. These
	// two variables say it outright, so the protection does not rest on
	// a side effect of HOME that some later change might undo.
	out = append(out, "GIT_CONFIG_GLOBAL="+filepath.Join(root, gitConfigName))
	out = append(out, "GIT_CONFIG_SYSTEM=/dev/null")

	return out
}

// gitConfigName is the config git is pointed at inside the sandbox.
const gitConfigName = ".gitconfig"

// isGitConfigEnv reports whether a variable would let git find a config
// outside the sandbox. They are dropped before the sandbox values are set,
// so an inherited value cannot survive.
func isGitConfigEnv(name string) bool {
	switch name {
	case "GIT_CONFIG_GLOBAL", "GIT_CONFIG_SYSTEM", "GIT_CONFIG", "GIT_CONFIG_NOSYSTEM":
		return true
	}
	return false
}

// WriteGitConfig creates the minimal git config the sandbox points at, if
// it is not already there.
//
// Git refuses to commit without an identity, so redirecting it at a file
// that does not exist trades a credential leak for a tool that does not
// work. This supplies the identity and nothing else: no credential helper,
// no URL rewrites, no includes.
//
// An existing file is left alone. It lives inside the sandbox, so the agent
// is entitled to have written it, and overwriting the agent's own config on
// every command would be surprising in a way that confinement does not
// require.
func WriteGitConfig(root string) error {
	if root == "" {
		return nil
	}
	path := filepath.Join(root, gitConfigName)
	if _, err := os.Stat(path); err == nil {
		return nil
	}
	const minimal = "[user]\n" +
		"\tname = Ensemble Agent\n" +
		"\temail = agent@ensemble.invalid\n" +
		"[init]\n" +
		"\tdefaultBranch = main\n"
	return os.WriteFile(path, []byte(minimal), 0o600)
}

// isSensitiveEnv reports whether a variable name is likely to carry a
// credential.
func isSensitiveEnv(name string) bool {
	upper := strings.ToUpper(name)
	for _, n := range sensitiveNames {
		if upper == n {
			return true
		}
	}
	for _, s := range sensitiveSuffixes {
		if strings.HasSuffix(upper, s) {
			return true
		}
	}
	return false
}

// SanitizeForLog replaces known credentials in s with a truncated form, so
// a log written inside the sandbox cannot hand the model a usable key.
//
// This closes the second path into the context window. The host binary
// writes api.log and debug.log into the agent's data directory; if that
// directory is inside the sandbox, the model can simply read_file it.
// Every other protection here would be pointless with the key sitting in
// a log.
//
// It runs at WRITE time, never as a later scrub. A file that briefly held
// the real key is a file the model may have read in the window before the
// scrub, and a race whose timing the attacker controls is not a
// mitigation.
//
// The truncated form keeps the log useful: "the key starting sk-p and
// ending 7xQ2" identifies which credential was used without being one.
// Credentials of nine characters or fewer are replaced outright, because
// showing four characters from each end of an eight-character secret shows
// the secret.
func SanitizeForLog(s string, creds []string) string {
	for _, c := range creds {
		if c == "" {
			continue
		}
		if len(c) > 8 {
			s = strings.ReplaceAll(s, c, c[:4]+"…"+c[len(c)-4:])
			continue
		}
		s = strings.ReplaceAll(s, c, "…")
	}
	return s
}
