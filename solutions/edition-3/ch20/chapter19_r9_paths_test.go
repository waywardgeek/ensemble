package ensemble_test

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"ensemble"
)

func TestCh19R9FileRefusals(t *testing.T) {
	top := t.TempDir()
	workspace := filepath.Join(top, "workspace")
	outside := filepath.Join(top, "outside")
	for _, dir := range []string{workspace, outside} {
		if err := os.MkdirAll(dir, 0700); err != nil {
			t.Fatal(err)
		}
	}
	outsideFile := filepath.Join(outside, "file")
	if err := os.WriteFile(outsideFile, []byte("OUTSIDE_BYTES"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(workspace, "inside"), []byte("INSIDE_BYTES"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, filepath.Join(workspace, "out-link")); err != nil {
		t.Fatal(err)
	}
	app := ensemble.New(nil)
	defer app.Shutdown()
	a, err := app.NewAgent(ensemble.Config{Model: "offline", BaseURL: "http://127.0.0.1:1", BaseDir: workspace, SandboxRoot: workspace, DataDir: filepath.Join(workspace, "data")})
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"read_file", "list_directory", "search_files", "write_file", "edit_file"} {
		t.Run(name, func(t *testing.T) {
			paths := []string{"../outside/file", "out-link/file", outsideFile}
			if name == "list_directory" || name == "search_files" {
				paths = []string{"../outside", "out-link", outside}
			}
			if name == "write_file" {
				paths = append(paths, "out-link/new-dir/file")
			}
			for _, path := range paths {
				args := map[string]any{"path": path}
				switch name {
				case "search_files":
					args["pattern"] = "OUTSIDE"
				case "write_file":
					args["content"] = "MODIFIED"
					args["overwrite"] = true
				case "edit_file":
					args["old_text"] = "OUTSIDE"
					args["new_text"] = "MODIFIED"
				}
				text, failed := c3Call(t, a, name, args)
				if !failed || !strings.Contains(text, path) {
					t.Errorf("%s must refuse supplied %q: failed=%v text=%q", name, path, failed, text)
				}
				if !filepath.IsAbs(path) && strings.Contains(text, top) {
					t.Errorf("synthesized host context for %q: %q", path, text)
				}
				if strings.Contains(text, "OUTSIDE_BYTES") {
					t.Error("outside bytes disclosed")
				}
			}
		})
	}
	data, err := os.ReadFile(outsideFile)
	if err != nil || string(data) != "OUTSIDE_BYTES" {
		t.Fatalf("outside mutation: %q %v", data, err)
	}
	if _, err := os.Stat(filepath.Join(outside, "new-dir")); !os.IsNotExist(err) {
		t.Fatalf("outside parent creation: %v", err)
	}
	for _, path := range []string{"inside", filepath.Join(workspace, "inside")} {
		if text, failed := c3Call(t, a, "read_file", map[string]any{"path": path}); failed || text != "INSIDE_BYTES" {
			t.Fatalf("allowed %q: %q %v", path, text, failed)
		}
	}
	for _, path := range []string{"allowed/relative", filepath.Join(workspace, "allowed", "absolute")} {
		if text, failed := c3Call(t, a, "write_file", map[string]any{"path": path, "content": "BEFORE"}); failed {
			t.Fatalf("allowed write: %q", text)
		}
		if text, failed := c3Call(t, a, "edit_file", map[string]any{"path": path, "old_text": "BEFORE", "new_text": "AFTER"}); failed {
			t.Fatalf("allowed edit: %q", text)
		}
		if text, failed := c3Call(t, a, "read_file", map[string]any{"path": path}); failed || text != "AFTER" {
			t.Fatalf("allowed read: %q %v", text, failed)
		}
	}
	if text, failed := c3Call(t, a, "search_files", map[string]any{"path": "allowed", "pattern": "AFTER"}); failed || !strings.Contains(text, "allowed/relative:1:AFTER") || !strings.Contains(text, "allowed/absolute:1:AFTER") || strings.Contains(text, top) {
		t.Fatalf("relative search/display: %q %v", text, failed)
	}
	if text, failed := c3Call(t, a, "list_directory", map[string]any{"path": "allowed"}); failed || !strings.Contains(text, "relative") || !strings.Contains(text, "absolute") || strings.Index(text, "absolute") > strings.Index(text, "relative") {
		t.Fatalf("allowed list: %q %v", text, failed)
	}
	if text, failed := c3Call(t, a, "read_file", map[string]any{"path": "./allowed"}); !failed || !strings.Contains(text, "./allowed") || strings.Contains(text, top) {
		t.Fatalf("handle read error lost requested spelling: %q %v", text, failed)
	}

}

func TestCh19R9CwdRefusalEnvelope(t *testing.T) {
	top := t.TempDir()
	workspace := filepath.Join(top, "workspace")
	outside := filepath.Join(top, "outside")
	for _, dir := range []string{filepath.Join(workspace, "sub"), outside} {
		if err := os.MkdirAll(dir, 0700); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.Symlink(outside, filepath.Join(workspace, "out-link")); err != nil {
		t.Fatal(err)
	}
	app := ensemble.New(nil)
	defer app.Shutdown()
	a, err := app.NewAgent(ensemble.Config{Model: "offline", BaseURL: "http://127.0.0.1:1", BaseDir: workspace, SandboxRoot: workspace, DataDir: filepath.Join(workspace, "data")})
	if err != nil {
		t.Fatal(err)
	}
	for _, cwd := range []string{"../outside", "./out-link", outside} {
		before := len(app.JobTable())
		raw, _ := json.Marshal(map[string]any{"command": "printf MUST_NOT_EXECUTE", "cwd": cwd})
		op := a.Tools().Prepare(context.Background(), ensemble.Part{Type: "tool_call", Name: "run_command", Args: raw}, "", 0)
		called := op.Called()
		if called.Job != nil {
			// Before evidence stops at reservation; never release a confined process.
			op.Abort(errors.New("fixture abort before command launch"))
			t.Errorf("rejected cwd %q reserved host context: %+v", cwd, called.Job)
			continue
		}
		op.Release()
		data := op.Await(context.Background())
		if !data.IsError || data.Dispatched || data.Job != nil || len(app.JobTable()) != before {
			t.Fatalf("refusal dispatched or reserved a job: %+v", data)
		}
		encoded, _ := json.Marshal(data)
		var text strings.Builder
		for _, part := range data.Parts {
			text.WriteString(part.Text)
		}
		if !strings.Contains(text.String(), cwd) {
			t.Errorf("missing supplied cwd %q: %q", cwd, text.String())
		}
		if strings.Contains(text.String(), "MUST_NOT_EXECUTE") {
			t.Error("command output in refusal")
		}
		if !filepath.IsAbs(cwd) && strings.Contains(string(encoded), top) {
			t.Errorf("synthesized host envelope: %s", encoded)
		}
		if strings.Contains(string(encoded), `"output"`) || strings.Contains(string(encoded), `"handle"`) {
			t.Errorf("refusal has job locator/handle: %s", encoded)
		}
	}
	// Admitted reservations retain existing metadata. Abort without launching.
	for _, cwd := range []string{"", "sub", filepath.Join(workspace, "sub")} {
		args := map[string]any{"command": "printf NOT_LAUNCHED"}
		if cwd != "" {
			args["cwd"] = cwd
		}
		raw, _ := json.Marshal(args)
		op := a.Tools().Prepare(context.Background(), ensemble.Part{Type: "tool_call", Name: "run_command", Args: raw}, "", 0)
		called := op.Called()
		op.Abort(errors.New("fixture abort: metadata-only check"))
		if called.Job == nil || called.Job.Handle <= 0 || !strings.HasPrefix(called.Job.Output.Locator, a.Config().DataDir) {
			t.Fatalf("admitted reservation lost handle/output: %+v", called)
		}
		want := cwd
		if cwd != "" && !filepath.IsAbs(cwd) {
			want = filepath.Join(workspace, cwd)
		}
		if called.Job.Cwd != want {
			t.Fatalf("admitted cwd=%q want %q", called.Job.Cwd, want)
		}
	}
}
