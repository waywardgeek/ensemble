package grade

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestBuild runs the selected executable: compiling a library archive or a
// different healthy command must not hide the intended entry point's failure.
func TestBuild(t *testing.T) {
	for _, tc := range []struct {
		name, cmdSource, want string
		nested, broken        bool
	}{
		{name: "root", want: "root"},
		{name: "command container", nested: true, want: "root"},
		{name: "preferred command", cmdSource: "package main\nfunc main() { println(\"cmd\") }", want: "cmd"},
		{name: "command library", cmdSource: "package library", want: "root"},
		{name: "broken preferred command", cmdSource: "package main\nfunc main() { undefinedFunction() }", broken: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			files := map[string]string{
				"go.mod":  "module example.org/grader-build\n\ngo 1.25\n",
				"main.go": "package main\nfunc main() { println(\"root\") }",
			}
			if tc.cmdSource != "" {
				files["cmd/main.go"] = tc.cmdSource
			}
			if tc.nested {
				files["cmd/gateway/main.go"] = "package main\nfunc main() { println(\"gateway\") }"
			}
			for name, source := range files {
				path := filepath.Join(dir, name)
				if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(path, []byte(source), 0600); err != nil {
					t.Fatal(err)
				}
			}
			bin, clean, err := Build(dir)
			if tc.broken {
				if clean != nil {
					clean()
				}
				if err == nil || !strings.Contains(err.Error(), "undefinedFunction") {
					t.Fatalf("selected command's compile failure was lost: %v", err)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			defer clean()
			out, err := exec.Command(bin).CombinedOutput()
			if err != nil || strings.TrimSpace(string(out)) != tc.want {
				t.Fatalf("executed %q, error %v; want %q", out, err, tc.want)
			}
		})
	}
}
