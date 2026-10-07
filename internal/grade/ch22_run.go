package grade

import (
	"fmt"
	"strings"
)

// Ch22Run observes a student's tree and reports what it found.
//
// Chapter 22 is about structure, so most of these checks read the source
// rather than running it. That is a hazard as much as a convenience: a
// structural check can pass while the program does not work, which is exactly
// how chapter 5's original parent-chain check sat green for sixteen chapters.
// Every check below therefore states a property that is false in a tree with
// the defect and true in one without it, and each is covered by a mutant in
// scripts/ch22-mutants.sh.
func Ch22Run(dir string) Ch22Result {
	var r Ch22Result

	// agent-builds. A tree that does not compile cannot be judged on anything
	// else, so this failure is fatal to the whole run rather than worth five
	// points on its own.
	r.ran("agent-builds")
	_, cleanup, err := Build(dir)
	if err != nil {
		r.fail("agent-builds", "go build failed: %v", err)
		r.Fatal = "the agent does not build, so no other check could be exercised"
		return r
	}
	defer cleanup()

	scan, err := scanTree(dir)
	if err != nil {
		r.Fatal = fmt.Sprintf("cannot parse the tree: %v", err)
		return r
	}

	// back-pointer-chain. The chapter's thesis. Framework-blind: see
	// backPointerChain in ch22_chain.go, and the sharpness fixtures in
	// ch22_chain_test.go which rename every identifier.
	r.ran("back-pointer-chain")
	evidence, ok := scan.backPointerChain()
	if !ok {
		r.fail("back-pointer-chain", "%s", evidence)
	}

	// reaches-through-the-chain. The anti-cheat for agent-status-tool: a
	// student could report the model from a package-level global or by
	// importing the engine directly. A tool must ask its parent.
	r.ran("reaches-through-the-chain")
	if msg, ok := toolPackagesReachOnlyTheHub(dir, scan); !ok {
		r.fail("reaches-through-the-chain", "%s", msg)
	}

	// ch21-parity. Chapter 22 rearranges the spine of the agent -- it renames
	// the hub interface, moves usage onto the engine, and replaces the
	// binary's composition root. Every one of those is a change that can
	// pass its own new checks while quietly breaking the chapter before it.
	// Parity is the guard, and it is worth its cost: it re-runs chapter 21's
	// own checks rather than re-asserting a summary of them, so it cannot
	// drift away from what chapter 21 actually requires.
	r.ran("ch21-parity")
	var broken []string
	for _, c := range Ch21Checks(Ch21Run(dir)) {
		if !c.Passed {
			broken = append(broken, c.ID)
		}
	}
	if len(broken) > 0 {
		r.fail("ch21-parity", "chapter 21 checks now fail: %s", strings.Join(broken, ", "))
	}

	// single-composition-root. One place assembles the agent, and the
	// binary uses it. Framework-blind: see detectSingleCompositionRoot in
	// ch22_root.go. Its sharpness was settled not by fixtures but by
	// running it against the real pre-repair tree, which an earlier draft
	// passed.
	r.ran("single-composition-root")
	if ok, detail := detectSingleCompositionRoot(scan); !ok {
		r.fail("single-composition-root", "%s", detail)
	}

	// TODO(ch22): these two are not implemented yet. They are failed
	// EXPLICITLY, with a reason naming the harness rather than the student,
	// so an incomplete grader cannot be mistaken for a failing tree. The
	// alternative -- leaving them unexercised -- produced the message "an
	// earlier failure stopped the run", which blames a student for work the
	// grader has not done. Design for each is in docs/ch22-grader-design.md.
	for _, id := range []string{
		"agent-status-tool",
		"per-model-cost",
	} {
		r.fail(id, "GRADER INCOMPLETE: this check is not implemented yet; see docs/ch22-grader-design.md. This is not a failure of the tree under test.")
	}

	return r
}

// toolPackagesReachOnlyTheHub reports whether every package that defines tool
// functions imports nothing from this module except the hub.
//
// Framework-blind in both halves. The hub is whichever package declares the
// dispatch struct; the tool packages are whichever packages declare functions
// taking that struct. Neither is named here.
//
// In-module imports are identified without knowing the module path: go list
// ./... enumerates exactly this module's packages, so an import that is itself
// a key in the graph is in-module by construction.
func toolPackagesReachOnlyTheHub(dir string, scan *astScan) (string, bool) {
	disp, _ := scan.bestDispatchStruct()
	if disp == nil {
		return "no dispatch struct found, so no tool package could be identified", false
	}

	// Which packages define functions taking the dispatch struct?
	toolPkgs := map[string]bool{}
	for _, fn := range scan.Funcs {
		if fn.Pkg == disp.Pkg {
			continue
		}
		for _, p := range fn.Params {
			if p.Key(fn.Pkg) == disp.Key() {
				toolPkgs[fn.Pkg] = true
			}
		}
	}
	if len(toolPkgs) == 0 {
		return fmt.Sprintf("no package outside %s defines a function taking %s, so nothing dispatches through the hub",
			disp.Pkg, disp.Key()), false
	}

	graph := DiscoverImportGraph(dir)
	if len(graph) == 0 {
		return "go list produced no import graph", false
	}
	inModule := func(path string) bool { _, ok := graph[path]; return ok }

	// Map package NAMES to import paths. A package's import path ends with its
	// directory, which is its name in every tree Go tooling will accept here.
	pathFor := func(pkg string) []string {
		var out []string
		for path := range graph {
			if path == pkg || strings.HasSuffix(path, "/"+pkg) {
				out = append(out, path)
			}
		}
		return out
	}

	hubPaths := map[string]bool{}
	for _, p := range pathFor(disp.Pkg) {
		hubPaths[p] = true
	}

	for pkg := range toolPkgs {
		for _, path := range pathFor(pkg) {
			for _, imp := range graph[path] {
				imp = strings.TrimSpace(imp)
				if imp == "" || !inModule(imp) || hubPaths[imp] || imp == path {
					continue
				}
				return fmt.Sprintf("%s imports %s directly; a tool must reach the engine through %s, not by importing its implementation",
					path, imp, disp.Pkg), false
			}
		}
	}
	return "", true
}
