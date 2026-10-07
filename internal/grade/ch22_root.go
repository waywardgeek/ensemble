package grade

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"
)

// Chapter 22's single-composition-root check.
//
// The decay this check exists to catch is recorded in the chapter: the
// binary built the agent's parts itself and then stapled nine capabilities
// onto the engine -- Cache, Ctx, Journal, Target, ToolRoundLimit, Memory,
// Bands, Recall and the model -- none of which the library's own
// constructor set. There were two composition roots, and only the one
// inside main() could build an agent that ran.
//
// Like the chapter 5 repair and back-pointer-chain, this check names no
// type, no function and no package from the reference tree. Chapter 5's
// original check grepped for the literal strings "Logf" and "Host" and so
// reported a structure it had never inspected; it stayed green for sixteen
// chapters while protecting nothing. Every clause below is a statement
// about shape.

// compositionRootFanout is how many of the tree's own packages a function
// must reach before it counts as composing the program rather than merely
// constructing a value.
//
// The number is a judgement, and it is the one place this check could be
// said to be tuned. Three separates the two populations cleanly in any
// layered program: a leaf constructor stores its arguments and reaches
// nothing, a mid-level constructor reaches one or two collaborators, and a
// composition root reaches the whole of the layer beneath it. Measured on
// the reference tree the gap is not close -- see TestCompositionRootFanout.
const compositionRootFanout = 3

// detectSingleCompositionRoot reports whether the program has exactly one
// place that assembles the agent.
//
// Two clauses, both shape-only:
//
//  1. ONE ROOT. Within a single main package there may be at most one call
//     to a composition root, and there must be at least one. Zero is the
//     decayed shape this chapter found: the binary reached past the
//     library's constructor and wired the parts itself.
//
//  2. NO STAPLES INSIDE THE ROOT'S REACH. A main package must not take an
//     object from a package the composition root itself builds from and
//     then assign its fields. Receiving a part and setting its
//     capabilities is the decay: the constructor's signature stops
//     describing what it takes to get a working object.
//
// Clause 1 is scoped per main package, not per tree, deliberately: a
// repository may legitimately ship more than one binary, and the reference
// tree does -- cmd/virtual-user is a second main package with its own root
// call. What is forbidden is one binary assembling the agent twice, not a
// repository containing two binaries.
//
// Clause 2 is restricted to the root's reach, and that restriction is the
// part that took measurement rather than design. "Main sets a field on a
// library value" is NOT on its own a decay signal: preparing a
// configuration struct before handing it to a constructor has exactly the
// same shape as stapling a capability onto a live engine, and no amount of
// syntax separates them -- only types do, and this scanner has none. Both
// shapes are present in the reference tree and only one is a fault. What
// does separate them is provenance: if the root already builds from
// package P, then a binary that builds a P-object and configures it is
// doing the root's job a second time. A value from the root's OWN package
// is a request to the root, not a part of it, so that package is excluded.
func detectSingleCompositionRoot(s *astScan) (bool, string) {
	if s == nil {
		return false, "no scan"
	}

	// Locate every composition-root call, grouped by the binary making it.
	type rootCall struct {
		pkg, name, site string
		reach           map[string]bool
	}
	roots := map[string][]rootCall{}
	byDir := map[string][]*astFunc{}
	for _, fn := range s.Funcs {
		if fn.Pkg != "main" {
			continue
		}
		dir := filepath.Dir(fn.File)
		byDir[dir] = append(byDir[dir], fn)
		for _, c := range s.callsIn(fn) {
			if c.Pkg == "" || !s.isTreePkg(fn.File, c.Pkg) {
				continue
			}
			target := s.findFunc(c.Pkg, c.Name)
			if target == nil {
				continue
			}
			reach := s.treeFanout(target)
			if len(reach) < compositionRootFanout {
				continue
			}
			roots[dir] = append(roots[dir], rootCall{
				pkg: c.Pkg, name: c.Name, reach: reach,
				site: fmt.Sprintf("%s:%d calls %s.%s (which reaches %d packages: %s)",
					shortPath(s, fn.File), c.Line, c.Pkg, c.Name,
					len(reach), strings.Join(sortedKeys(reach), ", ")),
			})
		}
	}

	var problems []string

	// Clause 1: exactly one assembly site per binary.
	//
	// "At most one" is the obvious half. The other half is "at least one",
	// and it has to be asked of each binary separately rather than of the
	// tree as a whole -- a lesson learned by running this check against
	// the actual pre-repair tree, where it passed. The decayed binary
	// called no root at all; a SECOND, minor binary happened to call one,
	// and a tree-wide count was satisfied by it. The decay was invisible
	// precisely because the broken program was not the one being counted.
	//
	// A binary is only required to have a root if it is assembling
	// something: the test is whether the main package itself reaches as
	// far into the tree as a composition root would. A small helper binary
	// that touches one package is not assembling anything and is left
	// alone; a binary that touches the whole library and calls no root is
	// doing the root's job inline, which is exactly what this chapter
	// found.
	for _, dir := range sortedKeys(toSet(byDir)) {
		own := map[string]bool{}
		for _, fn := range byDir[dir] {
			for _, c := range s.callsIn(fn) {
				if c.Pkg != "" && s.isTreePkg(fn.File, c.Pkg) {
					own[c.Pkg] = true
				}
			}
		}
		switch {
		case len(roots[dir]) > 1:
			sites := make([]string, 0, len(roots[dir]))
			for _, r := range roots[dir] {
				sites = append(sites, r.site)
			}
			sort.Strings(sites)
			problems = append(problems, fmt.Sprintf(
				"%s assembles the program in %d places:\n      %s",
				shortPath(s, dir), len(sites), strings.Join(sites, "\n      ")))
		case len(roots[dir]) == 0 && len(own) >= compositionRootFanout:
			problems = append(problems, fmt.Sprintf(
				"%s reaches %d of the library's packages (%s) but calls no "+
					"composition root: this binary assembles the agent itself "+
					"instead of asking the library to",
				shortPath(s, dir), len(own), strings.Join(sortedKeys(own), ", ")))
		}
	}

	// Clause 2: no configuring of parts the root already builds.
	for _, fn := range s.Funcs {
		if fn.Pkg != "main" {
			continue
		}
		called := roots[filepath.Dir(fn.File)]
		if len(called) == 0 {
			continue
		}
		owned := map[string]bool{} // packages the root builds from
		for _, r := range called {
			for p := range r.reach {
				owned[p] = true
			}
		}
		for _, r := range called {
			delete(owned, r.pkg) // the root's own package is a request, not a part
		}

		vars := s.varsFromTreeCalls(fn)
		for _, st := range s.fieldAssignsOn(fn, vars) {
			origin := vars[st.Pkg]
			if !owned[strings.SplitN(origin, ".", 2)[0]] {
				continue
			}
			problems = append(problems, fmt.Sprintf(
				"%s:%d: %s.%s is assigned after %s returned it -- a capability "+
					"stapled on by the caller, from a package the composition "+
					"root already builds from",
				shortPath(s, fn.File), st.Line, st.Pkg, st.Name, origin))
		}
	}

	if len(problems) > 0 {
		return false, strings.Join(problems, "\n    ")
	}

	total := 0
	for _, v := range roots {
		total += len(v)
	}
	if total == 0 {
		return false, "no main package calls a composition root: " +
			"nothing in this tree assembles the agent from the library, " +
			"so the binary is still building the parts itself"
	}
	return true, fmt.Sprintf(
		"one composition root per binary (%d main package(s), %d root call(s)); "+
			"no parts rebuilt or reconfigured by a caller", len(roots), total)
}

func shortPath(s *astScan, p string) string {
	if rel, err := filepath.Rel(s.Root, p); err == nil {
		return rel
	}
	return p
}

func sortedKeys(m map[string]bool) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

func toSet[T any](m map[string][]T) map[string]bool {
	out := map[string]bool{}
	for k := range m {
		out[k] = true
	}
	return out
}
