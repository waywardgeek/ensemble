package grade

// Sharpness audit for detectParentChain, the Chapter 5 back-pointer check.
//
// The check this replaced grepped for the identifiers "Logf" and "Host".
// These fixtures pin down the two directions that mattered and that the old
// check got wrong in both: a tree that renames every identifier must still
// pass, and a tree that keeps the identifiers while destroying the structure
// must fail.
//
// Fixtures are tiny synthetic trees rather than copies of the reference
// implementation, so they state the property in isolation and do not rot
// when the reference moves.

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// writeTree materializes map[relative path]contents under a temp dir.
func writeTree(t *testing.T, files map[string]string) string {
	t.Helper()
	root := t.TempDir()
	for name, body := range files {
		full := filepath.Join(root, name)
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return root
}

// The reference shape, parameterized over every identifier the old check
// hardcoded. hubPkg/iface/logMethod/dispatch all vary independently.
func chainTree(hubPkg, iface, logMethod, dispatch, deepPkg, deepType, field string) map[string]string {
	return map[string]string{
		hubPkg + "/hub.go": "package " + hubPkg + "\n\n" +
			"type " + iface + " interface {\n" +
			"\t" + logMethod + "(format string, args ...any)\n" +
			"}\n\n" +
			"type " + dispatch + " struct {\n" +
			"\t" + iface + "\n" +
			"\tName string\n" +
			"}\n",
		deepPkg + "/deep.go": "package " + deepPkg + "\n\n" +
			"import \"x/" + hubPkg + "\"\n\n" +
			"type " + deepType + " struct {\n" +
			"\t" + field + " " + hubPkg + "." + iface + "\n" +
			"}\n\n" +
			"func New" + deepType + "(p " + hubPkg + "." + iface + ") *" + deepType + " {\n" +
			"\treturn &" + deepType + "{" + field + ": p}\n" +
			"}\n\n" +
			"func (d *" + deepType + ") Run() {\n" +
			"\td." + field + "." + logMethod + "(\"running\")\n" +
			"}\n",
		"tools/tool.go": "package tools\n\n" +
			"import \"x/" + hubPkg + "\"\n\n" +
			"func Echo(c *" + hubPkg + "." + dispatch + ") string {\n" +
			"\treturn c.Name\n" +
			"}\n",
	}
}

func TestParentChainAcceptsReferenceShape(t *testing.T) {
	dir := writeTree(t, chainTree("common", "Host", "Logf", "Call", "jobs", "Jobs", "host"))
	ev, ok := detectParentChain(dir)
	if !ok {
		t.Fatalf("reference shape rejected: %s", ev)
	}
	if !strings.Contains(ev, "common.Host") || !strings.Contains(ev, "common.Call") {
		t.Errorf("evidence should name what it found, got: %s", ev)
	}
}

// THE REGRESSION TEST FOR THE ACTUAL BUG. The old check grepped for "Logf"
// and "Host" and so was pinned to this vocabulary; renaming every identifier
// would have failed a structurally perfect tree. A student who calls the
// interface Parent and the dispatch struct Invocation must score the same.
func TestParentChainIsNameBlind(t *testing.T) {
	dir := writeTree(t, chainTree("core", "Parent", "Printf", "Invocation", "worker", "Pool", "up"))
	ev, ok := detectParentChain(dir)
	if !ok {
		t.Fatalf("renamed-but-correct tree rejected: %s", ev)
	}
	for _, forbidden := range []string{"Host", "Call", "Logf", "common"} {
		if strings.Contains(ev, forbidden) {
			t.Errorf("evidence mentions reference vocabulary %q: %s", forbidden, ev)
		}
	}
}

// THE OTHER HALF OF THE BUG. The old check passed this tree: it contains the
// strings "Logf" and "Host" and nothing else the check wanted. There is no
// interface, no back-pointer, and no chain -- the logger is a package-level
// function that anyone can call from anywhere.
func TestParentChainRejectsVocabularyWithoutStructure(t *testing.T) {
	dir := writeTree(t, map[string]string{
		"common/log.go": "package common\n\n" +
			"func Logf(format string, args ...any) {}\n\n" +
			"type Call struct {\n\tName string\n}\n",
		"jobs/jobs.go": "package jobs\n\n" +
			"import \"x/common\"\n\n" +
			"type Jobs struct {\n\thost string\n}\n\n" +
			"func (j *Jobs) Run() {\n\tcommon.Logf(\"running\")\n}\n",
	})
	if ev, ok := detectParentChain(dir); ok {
		t.Fatalf("accepted a tree with the vocabulary but no chain: %s", ev)
	}
}

// A closure is not a back-pointer. This is the distinction Chapter 22 exists
// to make: the field answers exactly the question its author anticipated and
// no other, which is why a flat root produces dead settings.
func TestParentChainRejectsClosureInPlaceOfBackPointer(t *testing.T) {
	dir := writeTree(t, map[string]string{
		"common/hub.go": "package common\n\n" +
			"type Host interface {\n\tLogf(format string, args ...any)\n}\n\n" +
			"type Call struct {\n\tHost\n\tName string\n}\n",
		"jobs/jobs.go": "package jobs\n\n" +
			"type Jobs struct {\n\tlogf func(string, ...any)\n}\n\n" +
			"func NewJobs(logf func(string, ...any)) *Jobs {\n\treturn &Jobs{logf: logf}\n}\n\n" +
			"func (j *Jobs) Run() {\n\tj.logf(\"running\")\n}\n",
		"tools/tool.go": "package tools\n\n" +
			"import \"x/common\"\n\n" +
			"func Echo(c *common.Call) string {\n\treturn c.Name\n}\n",
	})
	ev, ok := detectParentChain(dir)
	if ok {
		t.Fatalf("accepted a closure staple as a back-pointer: %s", ev)
	}
	if !strings.Contains(ev, "closure") {
		t.Errorf("failure message should explain the closure rule, got: %s", ev)
	}
}

// Declaring the interface is not walking it. A tree where nothing ever logs
// through the stored parent has an ornamental chain.
func TestParentChainRejectsUnwalkedInterface(t *testing.T) {
	dir := writeTree(t, map[string]string{
		"common/hub.go": "package common\n\n" +
			"type Host interface {\n\tLogf(format string, args ...any)\n}\n\n" +
			"type Call struct {\n\tHost\n\tName string\n}\n",
		"jobs/jobs.go": "package jobs\n\n" +
			"import \"x/common\"\n\n" +
			"type Jobs struct {\n\thost common.Host\n}\n\n" +
			"func (j *Jobs) Run() string {\n\treturn \"running\"\n}\n",
		"tools/tool.go": "package tools\n\n" +
			"import \"x/common\"\n\n" +
			"func Echo(c *common.Call) string {\n\treturn c.Name\n}\n",
	})
	if ev, ok := detectParentChain(dir); ok {
		t.Fatalf("accepted an interface that is never walked: %s", ev)
	}
}

// The chain may be walked internally while the hub hands tool code nothing
// it can reach the parent through. That is the state Chapter 22 repairs for
// the engine, and the check must be able to see it.
func TestParentChainRejectsUnreachableDispatchStruct(t *testing.T) {
	dir := writeTree(t, map[string]string{
		"common/hub.go": "package common\n\n" +
			"type Host interface {\n\tLogf(format string, args ...any)\n}\n\n" +
			"type Call struct {\n\tName string\n}\n",
		"jobs/jobs.go": "package jobs\n\n" +
			"import \"x/common\"\n\n" +
			"type Jobs struct {\n\thost common.Host\n}\n\n" +
			"func (j *Jobs) Run() {\n\tj.host.Logf(\"running\")\n}\n",
		"tools/tool.go": "package tools\n\n" +
			"import \"x/common\"\n\n" +
			"func Echo(c *common.Call) string {\n\treturn c.Name\n}\n",
	})
	if ev, ok := detectParentChain(dir); ok {
		t.Fatalf("accepted a dispatch struct with no route to the parent: %s", ev)
	}
}
