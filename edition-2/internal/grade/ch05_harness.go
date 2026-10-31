package grade

// Chapter 5 harness: verify the refactoring produced a reusable framework.
//
// Takes a single directory — the student's library tree. Builds the binary
// from cmd/, inspects the import graph, and tests behavior.

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/waywardgeek/ensemble/internal/fakevendor"
)

// Ch5Result holds evidence for the ch5 checks.
type Ch5Result struct {
	AgentBuildOK   bool
	AgentBuildErr  string
	ImportGraph    map[string][]string
	ToolCalled     bool
	ToolOutput     string
	Base           string
	HasLogf        bool
	LogfEvidence   string
	MutableGlobals []string
	Ch4Result      *Ch4Result
	HelpersErr     string
}

func Ch5Run(dir string) (*Ch5Result, error) {
	dir, _ = filepath.Abs(dir)
	r := &Ch5Result{ImportGraph: make(map[string][]string)}

	// Discover module path.
	r.Base = DiscoverBase(dir)

	// Build binary.
	bin, cleanup, err := Build(dir)
	if err != nil {
		r.AgentBuildErr = err.Error()
		return r, nil
	}
	defer cleanup()
	r.AgentBuildOK = true

	// Inspect import graph.
	r.ImportGraph = DiscoverImportGraph(dir)

	// Verify the parent chain structurally: a back-pointer interface on the
	// dispatch struct, walked from outside the hub package. Not by name.
	r.LogfEvidence, r.HasLogf = detectParentChain(dir)

	// Check for mutable package-level vars.
	r.MutableGlobals = detectMutableGlobals(dir)

	// Drive the binary: fake vendor sends a tool call for "calculate",
	// binary should execute it and return the result.
	r.ToolCalled, r.ToolOutput = driveCh5Tool(bin, dir)

	// ch4 parity: run ch4's harness on the built binary.
	ch4r, err := Ch4Run(bin)
	if err != nil {
		r.HelpersErr = err.Error()
	} else {
		r.Ch4Result = ch4r
	}

	return r, nil
}

func GradeCh5(dir string) ([]Check, string) {
	r, err := Ch5Run(dir)
	if err != nil {
		return nil, err.Error()
	}
	return Ch5Evaluate(r), r.HelpersErr
}

// driveCh5Tool tests that the student's binary can execute a tool call.
// The fake vendor tells the model to call "calculate", and the binary
// should execute it and send the result back.
func driveCh5Tool(bin, workDir string) (toolCalled bool, toolOutput string) {
	replies := []fakevendor.Reply{
		{ToolName: "calculate", ToolArgs: `{"expression":"6 * 7"}`, ToolID: "call_001"},
		{Text: "The answer is 42."},
	}
	fake := fakevendor.New(replies)
	defer fake.Close()

	cmd := exec.Command(bin)
	// Not workDir: this launch must not load or leave a save.json in
	// the student's source tree (chapter 11). Nothing here reads a
	// file, so an empty directory is all the agent needs.
	runDir, cleanupDir := freshRunDir("ch5-tool")
	defer cleanupDir()
	cmd.Dir = runDir
	cmd.Env = append(os.Environ(),
		"LLM_VENDOR=anthropic",
		"LLM_MODEL=fake-model",
		"LLM_API_KEY=fake-key",
		"LLM_BASE_URL="+fake.URL(),
	)

	var inBuf bytes.Buffer
	inBuf.WriteString(`{"user":"What is 6 * 7?"}` + "\n")
	cmd.Stdin = &inBuf

	var outBuf, errBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf

	done := make(chan error, 1)
	if err := cmd.Start(); err != nil {
		return false, ""
	}
	go func() { done <- cmd.Wait() }()

	select {
	case <-done:
	case <-time.After(15 * time.Second):
		cmd.Process.Kill()
		return false, ""
	}

	// The fake received at least two requests if the tool was called.
	reqs := fake.Requests()
	if len(reqs) >= 2 {
		s := string(reqs[1].Body)
		if strings.Contains(s, "tool_result") || strings.Contains(s, "calculate") {
			toolCalled = true
			if i := strings.Index(s, "Result:"); i >= 0 {
				end := i + 40
				if end > len(s) {
					end = len(s)
				}
				toolOutput = s[i:end]
				if qi := strings.IndexByte(toolOutput, '"'); qi > 0 {
					toolOutput = toolOutput[:qi]
				}
			} else {
				toolOutput = "(tool was called)"
			}
		}
	}
	return
}

// detectParentChain verifies -- structurally, and without knowing any of the
// reference implementation's identifiers -- that the back-pointer interface
// taught in this chapter exists, is genuinely walked, and is reachable from
// the struct the hub hands outward to tool code.
//
// The property, stated without vocabulary. There is an interface I declared
// in the tree with a variadic logging method, and:
//
//   - WALKED: some struct in a package other than I's stores I as an
//     interface-typed field -- which is what a back-pointer constructor
//     parameter becomes -- and a method on that struct calls I's logging
//     method through that field.
//
//   - REACHABLE: some struct declared alongside I in I's own package carries
//     I as a field, and is passed as a parameter to a function in another
//     package. That is the tool dispatch struct: the value the hub hands
//     outward to code that was given nothing else.
//
// Together these say the parent chain is real rather than decorative, and
// that tool code has a route onto it. Neither clause names a type.
//
// Both clauses reject func-typed fields. That refusal is the whole
// distinction between a back-pointer and a closure stapled on at the wiring
// site, which is the subject of Chapter 22: a closure answers exactly the
// one question its author anticipated, and an interface answers every
// question the parent can answer.
//
// What this replaces grepped the tree for "Logf" and for "Host" and reported
// "Host interface with Logf found, embedded in Call" without ever looking at
// an interface or at a struct. It scored any tree containing a logging
// helper and an unrelated identifier, and a rename would have killed it.
// See P11.
//
// The evidence is returned so the check can report what it found rather than
// assert what it assumed.
func detectParentChain(dir string) (string, bool) {
	scan, err := scanTree(dir)
	if err != nil {
		return "no parsable Go source found: " + err.Error(), false
	}

	ifaces := scan.loggingInterfaces()
	if len(ifaces) == 0 {
		return "no interface in the tree declares a variadic logging method", false
	}

	// Report the near miss that got furthest, so a failing student is told
	// which clause broke rather than just being told "no".
	best := ""
	note := func(s string) {
		if best == "" {
			best = s
		}
	}
	for _, iface := range ifaces {
		use, usedOK := scan.storedBackPointerUse(iface)
		if !usedOK {
			note(fmt.Sprintf("%s has %v, but no struct outside package %s stores it "+
				"as a field and logs through it (a closure field does not count)",
				iface.Key(), loggingMethods(iface), iface.Pkg))
			continue
		}
		disp, dispOK := scan.dispatchStructFor(iface)
		if !dispOK {
			note(fmt.Sprintf("%s is walked as a back-pointer (%s), but no struct in "+
				"package %s carries it and is passed out to another package, so tool "+
				"code has no way to reach it", iface.Key(), use, iface.Pkg))
			continue
		}
		return use + "; " + disp, true
	}
	return best, false
}

// detectMutableGlobals finds package-level var declarations that are mutable
// state (not immutable lookup maps). Returns the list of offending locations.
func detectMutableGlobals(dir string) []string {
	cmd := exec.Command("grep", "-rn", "^var ", dir, "--include=*.go")
	out, _ := cmd.Output()
	var globals []string
	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if strings.Contains(line, "_test.go:") {
			continue
		}
		lower := strings.ToLower(line)
		if strings.Contains(lower, "names") || strings.Contains(lower, "name =") {
			continue
		}
		// go:embed has no var-free form: an embedded filesystem must be a
		// package-level var. The compiler populates it before main and
		// nothing reassigns it, so it is not mutable state.
		if strings.Contains(line, "embed.FS") {
			continue
		}
		globals = append(globals, line)
	}
	return globals
}
