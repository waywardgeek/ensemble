package common

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCh19ChildClamp(t *testing.T) {
	root := t.TempDir()
	outside := t.TempDir()
	if err := os.Symlink(outside, filepath.Join(root, "escape")); err != nil {
		t.Fatal(err)
	}
	parent := Config{SandboxRoot: root, SafeMode: true, EnableWebSearch: false, Tools: []string{}, SavePath: "parent-save", LogPath: "parent-log", MCPServers: []MCPServer{{Args: []string{"parent"}}}}
	child := parent.Child(filepath.Join(root, "child"))
	if child.SavePath != "" || child.LogPath != "" || child.Tools == nil {
		t.Fatal("child lost isolated paths or explicit empty tools")
	}
	child.MCPServers[0].Args[0] = "child"
	if parent.MCPServers[0].Args[0] != "parent" {
		t.Fatal("child aliases parent arguments")
	}
	for _, escape := range []string{"", outside, filepath.Join(root, "escape")} {
		child.SandboxRoot = escape
		child.SafeMode = false
		child.EnableWebSearch = true
		clamped, err := child.Clamp(parent)
		if err == nil || !clamped.SafeMode || clamped.EnableWebSearch {
			t.Fatalf("widening was not clamped: %+v, %v", clamped, err)
		}
		canonical, _ := CanonicalPath(root)
		if clamped.SandboxRoot != canonical {
			t.Fatal("root widened")
		}
	}
	child = parent.Child(filepath.Join(root, "nested"))
	child.SandboxRoot = child.DataDir
	if _, err := child.Clamp(parent); err != nil {
		t.Fatalf("narrow subtree rejected: %v", err)
	}
}
