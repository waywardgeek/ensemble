package grade

// Sharpness audit for backPointerChain, the chapter 22 thesis check.
//
// Every fixture below varies ALL the vocabulary: no package, interface,
// struct, field or method is named anything the reference implementation
// calls it. If this check can be made to pass or fail by renaming, it has
// become detectLogf again, which is the specific failure chapter 22 is about.
//
// The fixtures are tiny synthetic trees rather than copies of the reference
// tree, so they state the property in isolation and do not rot when the
// reference moves.

import (
	"strings"
	"testing"
)

// chainFixture builds a tree with a dispatch struct carrying two interfaces,
// one of which returns the other. Every identifier is a parameter.
//
// extraField is spliced into the dispatch struct, so a test can add a closure
// without changing anything else. deepReturnsIface controls whether the second
// interface has a method returning an interface, which is the chain's second
// link. passOutward controls whether any other package accepts the struct.
func chainFixture(pkg, ifaceA, ifaceB, dispatch, upMethod, extraField string,
	deepReturnsIface, passOutward bool) map[string]string {

	up := "\t" + upMethod + "() " + ifaceA + "\n"
	if !deepReturnsIface {
		// Same method, same name, but it hands back a string instead of an
		// interface. Nothing can be reached THROUGH it.
		up = "\t" + upMethod + "() string\n"
	}

	files := map[string]string{
		pkg + "/hub.go": "package " + pkg + "\n\n" +
			"type " + ifaceA + " interface {\n" +
			"\tNote(format string, args ...any)\n" +
			"}\n\n" +
			"type " + ifaceB + " interface {\n" +
			up +
			"\tLabel() string\n" +
			"}\n\n" +
			"type " + dispatch + " struct {\n" +
			"\t" + ifaceA + "\n" +
			"\tLink " + ifaceB + "\n" +
			"\tCount int\n" +
			extraField +
			"}\n",
	}
	if passOutward {
		files["leaf/leaf.go"] = "package leaf\n\n" +
			"import \"x/" + pkg + "\"\n\n" +
			"func Handle(c " + pkg + "." + dispatch + ") string {\n" +
			"\treturn c.Link.Label()\n" +
			"}\n"
	}
	return files
}

func chainResult(t *testing.T, files map[string]string) (string, bool) {
	t.Helper()
	s, err := scanTree(writeTree(t, files))
	if err != nil {
		t.Fatalf("scanTree: %v", err)
	}
	return s.backPointerChain()
}

// A correct tree passes no matter what anything is called. Three completely
// disjoint vocabularies, none of them the reference implementation's.
func TestBackPointerChainIsNameBlind(t *testing.T) {
	vocabularies := []struct {
		label                                   string
		pkg, ifaceA, ifaceB, dispatch, upMethod string
	}{
		{"reference-like", "common", "Agent", "Engine", "Call", "Agent"},
		{"spine", "core", "Parent", "Spine", "Packet", "Up"},
		{"nothing-in-common", "kernel", "Ancestor", "Conduit", "Envelope", "Origin"},
	}
	for _, v := range vocabularies {
		t.Run(v.label, func(t *testing.T) {
			ev, ok := chainResult(t, chainFixture(
				v.pkg, v.ifaceA, v.ifaceB, v.dispatch, v.upMethod, "", true, true))
			if !ok {
				t.Fatalf("correct tree rejected: %s", ev)
			}
			// The evidence must describe what was found, not a type the check
			// went looking for. It should name this vocabulary's own types.
			if !strings.Contains(ev, v.dispatch) || !strings.Contains(ev, v.ifaceB) {
				t.Errorf("evidence does not name the tree's own types: %s", ev)
			}
		})
	}
}

// A closure in the dispatch struct is the chapter's central mistake: it
// answers one question forever. It must fail even though every other clause
// holds.
func TestBackPointerChainRejectsClosureField(t *testing.T) {
	ev, ok := chainResult(t, chainFixture(
		"core", "Parent", "Spine", "Packet", "Up",
		"\tModel func() string\n", true, true))
	if ok {
		t.Fatalf("tree with a closure field accepted: %s", ev)
	}
	if !strings.Contains(ev, "Model") {
		t.Errorf("diagnostic does not name the offending field: %s", ev)
	}
}

// A lone back-pointer is not a chain. The struct carries interfaces, but
// nothing can be reached through them, so a new capability would still have to
// be stapled on.
func TestBackPointerChainRejectsDepthOne(t *testing.T) {
	ev, ok := chainResult(t, chainFixture(
		"core", "Parent", "Spine", "Packet", "Up", "", false, true))
	if ok {
		t.Fatalf("depth-one tree accepted: %s", ev)
	}
	if !strings.Contains(ev, "THROUGH") {
		t.Errorf("diagnostic does not explain what is missing: %s", ev)
	}
}

// A struct that never crosses a package boundary is an ordinary local
// aggregate, not a dispatch struct.
func TestBackPointerChainRequiresDispatchStruct(t *testing.T) {
	ev, ok := chainResult(t, chainFixture(
		"core", "Parent", "Spine", "Packet", "Up", "", true, false))
	if ok {
		t.Fatalf("tree with no dispatch struct accepted: %s", ev)
	}
	if !strings.Contains(ev, "dispatch struct") {
		t.Errorf("diagnostic does not say what was missing: %s", ev)
	}
}

// The reference tree must pass its own thesis check. This is the one fixture
// that is allowed to know where the real code lives, and it is deliberately
// last: if it were first, a check that only worked on the reference tree would
// look healthy.
func TestBackPointerChainAcceptsReferenceTree(t *testing.T) {
	s, err := scanTree("../../agent")
	if err != nil {
		t.Skipf("reference tree unavailable: %v", err)
	}
	ev, ok := s.backPointerChain()
	if !ok {
		t.Fatalf("reference tree fails its own thesis check: %s", ev)
	}
	t.Logf("reference evidence: %s", ev)
}
