package agent

import (
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/waywardgeek/ensemble/agent/sandbox"
)

// The tests below are deliberately split along the seam Clamp is built on.
//
// "Effective" tests assert the narrowing: what permissions the child ends up
// with. "Reports" tests assert the diagnostic: that the attempt was named.
// They are separate tests rather than two assertions in one, because the
// whole point of the design is that these two properties fail independently.
// Deleting the error should break only the second group. Deleting the
// assignment that narrows should break only the first. A single test
// asserting both would hide which one a change actually broke.

func within(t *testing.T, root, path string) (bool, error) {
	t.Helper()
	return sandbox.Within(root, path)
}

func parentSpec(t *testing.T) AgentSpec {
	t.Helper()
	root := t.TempDir()
	return AgentSpec{
		DataDir:         root,
		SandboxRoot:     root,
		SafeMode:        true,
		EnableWebSearch: false,
	}
}

// --- derivation -------------------------------------------------------

func TestChildInheritsRestrictionsByConstruction(t *testing.T) {
	parent := parentSpec(t)
	child := parent.Child(filepath.Join(parent.DataDir, "sub"))

	if !child.SafeMode {
		t.Error("child did not inherit SafeMode")
	}
	if child.EnableWebSearch {
		t.Error("child did not inherit the absence of web search")
	}
	if child.SandboxRoot != parent.SandboxRoot {
		t.Errorf("child did not inherit SandboxRoot: got %q want %q",
			child.SandboxRoot, parent.SandboxRoot)
	}
}

// A child that inherits its parent's save file would corrupt it: two agents
// would write the same file.
func TestChildDoesNotInheritSingleFilePaths(t *testing.T) {
	parent := parentSpec(t)
	parent.LogPath = "/tmp/parent.log"
	parent.SavePath = "/tmp/parent.json"

	child := parent.Child(filepath.Join(parent.DataDir, "sub"))

	if child.LogPath != "" {
		t.Errorf("child inherited LogPath %q", child.LogPath)
	}
	if child.SavePath != "" {
		t.Errorf("child inherited SavePath %q", child.SavePath)
	}
	if child.DataDir == parent.DataDir {
		t.Error("child shares its parent's data directory")
	}
}

// --- the wall: effective permissions ----------------------------------

func TestClampForcesSafeModeOnChildOfSafeParent(t *testing.T) {
	parent := parentSpec(t)
	child := parent.Child(filepath.Join(parent.DataDir, "sub"))
	child.SafeMode = false // the compromised child tries to escape

	got, _ := child.Clamp(parent)

	if !got.SafeMode {
		t.Error("child escaped safe mode")
	}
}

func TestClampWithholdsWebSearchParentLacks(t *testing.T) {
	parent := parentSpec(t)
	child := parent.Child(filepath.Join(parent.DataDir, "sub"))
	child.EnableWebSearch = true

	got, _ := child.Clamp(parent)

	if got.EnableWebSearch {
		t.Error("child granted itself web search its parent lacked")
	}
}

func TestClampPullsOutsideRootBackInside(t *testing.T) {
	parent := parentSpec(t)
	child := parent.Child(filepath.Join(parent.DataDir, "sub"))
	child.SandboxRoot = "/etc"

	got, _ := child.Clamp(parent)

	inside, err := within(t, parent.SandboxRoot, got.SandboxRoot)
	if err != nil {
		t.Fatalf("checking containment: %v", err)
	}
	if !inside {
		t.Errorf("child root %q is not inside parent root %q",
			got.SandboxRoot, parent.SandboxRoot)
	}
}

// An unconfined child of a confined parent is a widening too, and the
// quietest one: it asks for nothing and would receive the whole filesystem.
func TestClampRefusesToLeaveChildUnconfined(t *testing.T) {
	parent := parentSpec(t)
	child := parent.Child(filepath.Join(parent.DataDir, "sub"))
	child.SandboxRoot = ""

	got, _ := child.Clamp(parent)

	if got.SandboxRoot == "" {
		t.Error("child of a confined parent was left unconfined")
	}
}

// --- the alarm: the error -------------------------------------------

func TestClampReportsEachWidening(t *testing.T) {
	for _, tc := range []struct {
		name   string
		mutate func(*AgentSpec)
		want   string
	}{
		{"safe mode", func(s *AgentSpec) { s.SafeMode = false }, "safe mode"},
		{"web search", func(s *AgentSpec) { s.EnableWebSearch = true }, "web search"},
		{"sandbox root", func(s *AgentSpec) { s.SandboxRoot = "/etc" }, "sandbox root"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			parent := parentSpec(t)
			child := parent.Child(filepath.Join(parent.DataDir, "sub"))
			tc.mutate(&child)

			_, err := child.Clamp(parent)

			if err == nil {
				t.Fatalf("widening %s was not reported", tc.name)
			}
			if !errors.Is(err, ErrWidenedPermissions) {
				t.Errorf("error does not wrap ErrWidenedPermissions: %v", err)
			}
			if !strings.Contains(err.Error(), tc.want) {
				t.Errorf("error does not name %q: %v", tc.want, err)
			}
		})
	}
}

// POSITIVE CONTROL. A clamp that rejected every child would pass all of the
// tests above and be useless. An ordinary child -- one that inherits, or
// narrows further -- must be clamped silently.
func TestClampAcceptsHonestChildren(t *testing.T) {
	parent := parentSpec(t)
	parent.SafeMode = false
	parent.EnableWebSearch = true

	t.Run("plain inheritance", func(t *testing.T) {
		child := parent.Child(filepath.Join(parent.DataDir, "sub"))

		got, err := child.Clamp(parent)

		if err != nil {
			t.Fatalf("an inheriting child was reported as widening: %v", err)
		}
		if got.SafeMode || !got.EnableWebSearch {
			t.Error("inheritance did not preserve the parent's permissions")
		}
	})

	t.Run("narrowing further", func(t *testing.T) {
		child := parent.Child(filepath.Join(parent.DataDir, "sub"))
		child.SafeMode = true
		child.EnableWebSearch = false
		child.SandboxRoot = filepath.Join(parent.SandboxRoot, "sub")

		got, err := child.Clamp(parent)

		if err != nil {
			t.Fatalf("a narrowing child was reported as widening: %v", err)
		}
		if !got.SafeMode {
			t.Error("child's own stricter safe mode was discarded")
		}
		if got.EnableWebSearch {
			t.Error("child's own refusal of web search was discarded")
		}
		if got.SandboxRoot != filepath.Join(parent.SandboxRoot, "sub") {
			t.Errorf("child's narrower root was discarded: %q", got.SandboxRoot)
		}
	})
}

// A child of an unconfined parent may be confined, unconfined, or anything
// between: there is nothing to widen past.
func TestClampLeavesUnconfinedParentAlone(t *testing.T) {
	parent := AgentSpec{DataDir: t.TempDir(), EnableWebSearch: true}
	child := parent.Child(filepath.Join(parent.DataDir, "sub"))
	child.SandboxRoot = t.TempDir()

	got, err := child.Clamp(parent)

	if err != nil {
		t.Fatalf("confining a child of an unconfined parent was refused: %v", err)
	}
	if got.SandboxRoot != child.SandboxRoot {
		t.Error("child's self-imposed sandbox was discarded")
	}
}
