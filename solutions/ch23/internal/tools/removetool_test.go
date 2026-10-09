package tools

import (
	"testing"
)

// RemoveTool had no test, and the gap was not harmless: it deleted only the
// normalized form of the name while builtins are keyed by their literal
// name, so removing run_command did nothing at all. Nothing failed when it
// did nothing. The tool stayed in the registry and kept working, which is
// exactly what a tool is supposed to do, so the only way to notice was to
// ask whether a tool that should be gone was gone.
func TestRemoveToolRemovesBuiltinsWithUnderscores(t *testing.T) {
	// Every builtin worth withholding has an underscore in its name, which
	// is precisely the case the old implementation could not handle.
	for _, name := range []string{"run_command", "wait_for_job", "send_input", "kill_job"} {
		t.Run(name, func(t *testing.T) {
			r := NewRegistry()

			if _, err := r.Lookup(name); err != nil {
				t.Fatalf("%q is not a builtin, so this test proves nothing: %v", name, err)
			}

			r.RemoveTool(name)

			if _, err := r.Lookup(name); err == nil {
				t.Errorf("%q is still present after RemoveTool", name)
			}
		})
	}
}

// Removal must not be a registry-wide reset.
func TestRemoveToolLeavesOtherToolsAlone(t *testing.T) {
	r := NewRegistry()
	r.RemoveTool("run_command")

	if _, err := r.Lookup("read_file"); err != nil {
		t.Errorf("removing run_command also removed read_file: %v", err)
	}
}

// Lookup accepts the normalized spelling too, so removal has to close that
// door as well. Otherwise a withheld tool is reachable by asking for it
// slightly differently, which is a bypass rather than a bug.
func TestRemoveToolClosesTheNormalizedSpelling(t *testing.T) {
	r := NewRegistry()
	r.RemoveTool("run_command")

	if _, err := r.Lookup("runcommand"); err == nil {
		t.Error("run_command is still reachable under its normalized name")
	}
}

func TestRemoveToolIsSafeForUnknownNames(t *testing.T) {
	r := NewRegistry()
	r.RemoveTool("no_such_tool_exists")
}
