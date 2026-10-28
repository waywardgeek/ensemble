package gradeedition2

import (
	"fmt"
	"runtime"
	"strings"

	"github.com/waywardgeek/ensemble/internal/fakevendor"
	"github.com/waywardgeek/ensemble/internal/grade"
)

// Reuse the real host probes and their positive controls. The Go-only child
// consumer is explicitly excluded; C# DeriveChild has separate component
// evidence and earns no executable points here. No commit is made by this scope.
func Chapter19Executable(bin string) (grade.Report, error) {
	if runtime.GOOS != "darwin" {
		return grade.Report{}, fmt.Errorf("Chapter 19 C# executable scope requires macOS command confinement")
	}
	checks := grade.Ch23Checks(grade.Ch23Result{})
	for i := range checks {
		checks[i].Passed = true
		checks[i].Earned = checks[i].Points
		checks[i].Details = nil
	}
	// Clarify the tested known-secret diagnostic boundary, not a promise that
	// arbitrary successful provider content can never contain a credential.
	checks[3].Title = "Child environment omits configured credentials; readable diagnostics redact known values"

	sandboxChecksFor(bin, checks, true, true)
	sandboxSafeFor(bin, &checks[5], true, true)
	return grade.NewTitledReport("Edition 3 C#, Chapter 19 — executable sandbox scope (90 points; Go child consumer excluded)", append(checks[:4], checks[5:]...),
		"Owned local peers and files on macOS only. No live provider, arbitrary-secret detection, or trusted MCP confinement claim. DeriveChild is tested separately."), nil
}

const sandboxExecutableGitCommand = "git init -q repo && cd repo && printf git-control > note && git add note && git show :note && test -n \"$(git config user.name)\" && test -n \"$(git config user.email)\""

// Keep this wording adaptation out of the original Go refusal contract. An
// actual correlated error is mandatory; successful text cannot earn a refusal.
func sandboxExecutableRefusal(c *grade.Check, r fakevendor.Recorded, id string) {
	text := strings.ToLower(skillResult(c, r, id, true))
	if !strings.Contains(text, "tool is not enabled") {
		fail(c, "%s was not a disabled-tool refusal: %s", id, text)
	}
}
