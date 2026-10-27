package ensemble_test

import (
	"bytes"
	"encoding/base64"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"ensemble/ensemble"
)

const downloadPDF = "%PDF-1.4\n% DOWNLOAD_PDF_BYTES_STAY_ON_DISK\n" +
	"1 0 obj\n<< /Type /Catalog /Pages 2 0 R >>\nendobj\n" +
	"2 0 obj\n<< /Type /Pages /Kids [3 0 R] /Count 1 >>\nendobj\n" +
	"3 0 obj\n<< /Type /Page /Parent 2 0 R /MediaBox [0 0 72 72] >>\nendobj\n" +
	"xref\n0 4\n0000000000 65535 f \n0000000043 00000 n \n0000000092 00000 n \n0000000149 00000 n \n" +
	"trailer\n<< /Size 4 /Root 1 0 R >>\nstartxref\n218\n%%EOF\n"

func downloadAgent(t *testing.T, workspace, dataDir string, spec ensemble.AgentSpec) ensemble.Agent {
	t.Helper()
	agent, err := ensemble.New(io.Discard).NewAgent(ensemble.Config{
		Model: "offline", BuiltinTools: true, Workspace: workspace, DataDir: dataDir,
	}, spec)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := agent.Shutdown(); err != nil {
			t.Error(err)
		}
	})
	return agent
}

// Both consumers must enforce the same boundary, including without a sandbox.
func TestDownloadWorkspaceAccess(t *testing.T) {
	for _, mode := range []string{"workspace", "sandbox", "safe-sandbox", "workspace-symlink"} {
		t.Run(mode, func(t *testing.T) {
			root := t.TempDir()
			workspace := filepath.Join(root, "workspace")
			outside := filepath.Join(root, "workspace-other")
			for _, directory := range []string{filepath.Join(workspace, "reports"), outside} {
				if err := os.MkdirAll(directory, 0700); err != nil {
					t.Fatal(err)
				}
			}
			source := filepath.Join(workspace, "reports", "report.pdf")
			unreadable := filepath.Join(workspace, "unreadable.pdf")
			for _, path := range []string{source, unreadable, filepath.Join(outside, "secret.pdf")} {
				if err := os.WriteFile(path, []byte(downloadPDF), 0600); err != nil {
					t.Fatal(err)
				}
			}
			if err := os.Chmod(unreadable, 0000); err != nil {
				t.Fatal(err)
			}
			for name, target := range map[string]string{
				"inside.pdf":   "reports/report.pdf",
				"escape.pdf":   filepath.Join(outside, "secret.pdf"),
				"escape-dir":   outside,
				"dangling.pdf": "absent.pdf",
			} {
				if err := os.Symlink(target, filepath.Join(workspace, name)); err != nil {
					t.Fatal(err)
				}
			}
			configuredWorkspace := workspace
			dataDir := filepath.Join(root, "state")
			spec := ensemble.AgentSpec{}
			switch mode {
			case "sandbox", "safe-sandbox":
				spec.SandboxRoot = workspace
				spec.SafeMode = mode == "safe-sandbox"
				dataDir = filepath.Join(workspace, "state")
				configuredWorkspace = outside
			case "workspace-symlink":
				configuredWorkspace = filepath.Join(root, "workspace-alias")
				if err := os.Symlink(workspace, configuredWorkspace); err != nil {
					t.Fatal(err)
				}
			}
			agent := downloadAgent(t, configuredWorkspace, dataDir, spec)
			if spec.SandboxRoot != "" && agent.Config().Workspace != agent.Spec().SandboxRoot {
				t.Fatal("sandbox did not become the effective workspace")
			}
			for _, test := range []struct {
				name, path string
				allowed    bool
			}{
				{"relative-pdf", "reports/report.pdf", true},
				{"absolute-pdf", filepath.Join(agent.Config().Workspace, "reports", "report.pdf"), true},
				{"internal-symlink", "inside.pdf", true},
				{"parent-traversal", "../workspace-other/secret.pdf", false},
				{"nested-traversal", "reports/../../workspace-other/secret.pdf", false},
				{"outside-absolute-prefix-lookalike", filepath.Join(outside, "secret.pdf"), false},
				{"symlink-file-escape", "escape.pdf", false},
				{"symlink-directory-escape", "escape-dir/secret.pdf", false},
				{"dangling-symlink", "dangling.pdf", false},
				{"missing", "reports/missing.pdf", false},
				{"directory", "reports", false},
				{"workspace-root", ".", false},
				{"empty", "", false},
				{"unreadable", "unreadable.pdf", false},
			} {
				t.Run(test.name, func(t *testing.T) {
					if test.name == "unreadable" {
						if file, err := os.Open(unreadable); err == nil {
							file.Close()
							t.Skip("current user can read mode-000 files")
						} else if !os.IsPermission(err) {
							t.Fatalf("permission fixture: %v", err)
						}
					}
					t.Run("shared-opener", func(t *testing.T) {
						file, err := agent.Sandbox().OpenWorkspaceFile(test.path)
						if file != nil {
							defer file.Close()
						}
						if !test.allowed {
							if err == nil || file != nil {
								t.Fatalf("rejected path returned file=%v, error=%v", file, err)
							}
							return
						}
						if err != nil || file == nil {
							t.Fatalf("allowed open: file=%v, error=%v", file, err)
						}
						got, err := io.ReadAll(file)
						if err != nil || string(got) != downloadPDF {
							t.Fatalf("PDF bytes = %q, error=%v", got, err)
						}
						opened, err := file.Stat()
						if err != nil {
							t.Fatal(err)
						}
						original, err := os.Stat(source)
						if err != nil || !os.SameFile(opened, original) {
							t.Fatalf("opener did not return the original file: %v", err)
						}
					})
					t.Run("builtin-invoke", func(t *testing.T) {
						out, err := sandboxCall(t, agent, "offer_download", map[string]any{"path": test.path})
						if !test.allowed {
							if err == nil || strings.Contains(out, "Download offered:") {
								t.Fatalf("rejected path offered: %q, error=%v", out, err)
							}
							return
						}
						if err != nil || out != "Download offered: "+filepath.Base(test.path) {
							t.Fatalf("offer: %q, error=%v", out, err)
						}
					})
				})
			}
		})
	}
}

// Creating or listing a work product must not offer it or copy it into state.
func TestOfferDownloadIsExplicitAndDoesNotCopy(t *testing.T) {
	root := t.TempDir()
	workspace := filepath.Join(root, "workspace")
	if err := os.Mkdir(workspace, 0700); err != nil {
		t.Fatal(err)
	}
	agent := downloadAgent(t, workspace, filepath.Join(root, "state"), ensemble.AgentSpec{})
	found := false
	for _, declaration := range agent.Declarations() {
		if declaration.Name == "offer_download" {
			found = true
			description := strings.ToLower(declaration.Description)
			if !strings.Contains(description, "only when") || !strings.Contains(description, "user explicitly request") {
				t.Errorf("description omits explicit user request requirement: %q", declaration.Description)
			}
		}
	}
	if !found {
		t.Fatal("built-in offer_download is missing")
	}
	source := filepath.Join(workspace, "report.pdf")
	if err := os.WriteFile(source, []byte(downloadPDF), 0600); err != nil {
		t.Fatal(err)
	}
	if out, err := sandboxCall(t, agent, "list_directory", map[string]any{"path": workspace}); err != nil || !strings.Contains(out, "report.pdf") {
		t.Fatalf("list work product: %q, error=%v", out, err)
	}
	var before bytes.Buffer
	if err := agent.History().Dump(&before); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(before.String(), "offer_download") || strings.Contains(before.String(), "Download offered:") {
		t.Fatal("work product offered without an explicit tool call")
	}
	if out, err := sandboxCall(t, agent, "offer_download", map[string]any{"path": "report.pdf"}); err != nil || out != "Download offered: report.pdf" {
		t.Fatalf("explicit offer: %q, error=%v", out, err)
	}
	var after bytes.Buffer
	if err := agent.History().Dump(&after); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(after.String(), "Download offered: report.pdf") {
		t.Fatal("explicit offer result was not recorded")
	}
	for _, forbidden := range []string{"DOWNLOAD_PDF_BYTES_STAY_ON_DISK", base64.StdEncoding.EncodeToString([]byte(downloadPDF))} {
		if strings.Contains(after.String(), forbidden) {
			t.Fatal("download bytes entered history")
		}
	}
	if err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil || !entry.Type().IsRegular() {
			return err
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if path == source {
			if string(data) != downloadPDF {
				t.Error("offering changed the source PDF")
			}
		} else if bytes.Contains(data, []byte("DOWNLOAD_PDF_BYTES_STAY_ON_DISK")) || bytes.Contains(data, []byte(base64.StdEncoding.EncodeToString([]byte(downloadPDF)))) {
			t.Errorf("download bytes copied to %s", path)
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
}

// Invoke may return before the actor records completion; inspect durable facts
// until the matching job_finished arrives, rather than relying on a fixed sleep.
func TestOfferDownloadJobFinishedStatus(t *testing.T) {
	workspace := t.TempDir()
	if err := os.WriteFile(filepath.Join(workspace, "report.pdf"), []byte(downloadPDF), 0600); err != nil {
		t.Fatal(err)
	}
	agent := downloadAgent(t, workspace, t.TempDir(), ensemble.AgentSpec{})
	for _, test := range []struct {
		name, path string
		failed     bool
	}{
		{"failed", "missing.pdf", true},
		{"successful", "report.pdf", false},
	} {
		t.Run(test.name, func(t *testing.T) {
			sequence := agent.History().LastSeq()
			out, err := sandboxCall(t, agent, "offer_download", map[string]any{"path": test.path})
			if (err != nil) != test.failed {
				t.Fatalf("offer returned %q, error=%v; want failed=%v", out, err, test.failed)
			}
			if test.failed {
				if out != err.Error() || !strings.Contains(out, test.path) || strings.Contains(out, "Download offered:") {
					t.Fatalf("failed offer result: %q, error=%v", out, err)
				}
			} else if out != "Download offered: report.pdf" {
				t.Fatalf("successful offer result: %q", out)
			}
			timeout := time.NewTimer(5 * time.Second)
			defer timeout.Stop()
			poll := time.NewTicker(time.Millisecond)
			defer poll.Stop()
			handle := 0
			var calledSeq uint64
			for {
				for _, event := range agent.History().EventsAfter(sequence) {
					sequence = event.Seq
					if event.Type == "tool_called" && event.Tool != nil && event.Tool.Name == "offer_download" {
						if event.Tool.Job == nil || event.Tool.Job.Handle <= 0 || event.Tool.Job.Status != "running" || event.Tool.Job.Failed {
							t.Fatalf("offer did not start a normal running job: %+v", event.Tool.Job)
						}
						handle = event.Tool.Job.Handle
						calledSeq = event.Seq
					}
					if event.Type != "job_finished" || event.Tool == nil || event.Tool.Job == nil || handle == 0 || event.Tool.Job.Handle != handle {
						continue
					}
					job := event.Tool.Job
					if event.Tool.CalledSeq != calledSeq {
						t.Fatalf("job_finished dispatch sequence = %d, want %d", event.Tool.CalledSeq, calledSeq)
					}
					if job.Status != "done" || job.Failed != test.failed {
						t.Fatalf("job_finished = %+v; want done with Failed=%v", job, test.failed)
					}
					data, err := os.ReadFile(job.Output.Locator)
					if err != nil || string(data) != out || job.Bytes != int64(len(data)) {
						t.Fatalf("finished job output = %q, bytes=%d, error=%v; want %q", data, job.Bytes, err, out)
					}
					return
				}
				select {
				case <-poll.C:
				case <-timeout.C:
					t.Fatalf("no job_finished fact for offer_download job %d", handle)
				}
			}
		})
	}
}
