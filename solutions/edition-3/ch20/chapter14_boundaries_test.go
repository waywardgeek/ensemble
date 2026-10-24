package ensemble_test

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"ensemble"
)

func TestChapter14HandoffSaveDiagnosticBoundary(t *testing.T) {
	for _, failSave := range []bool{false, true} {
		name := "successful-save"
		if failSave {
			name = "failed-save"
		}
		t.Run(name, func(t *testing.T) {
			dir := t.TempDir()
			savePath, diagPath := filepath.Join(dir, "session.save"), filepath.Join(dir, "trace.log")
			diagnostic, err := ensemble.OpenDiagnosticLog(diagPath, func() int { return 2 })
			if err != nil {
				t.Fatal(err)
			}
			defer diagnostic.Close()
			var requests atomic.Int32
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if requests.Add(1) == 1 {
					ch5Reply(w, ch5Call("handoff", "micro_handoff", map[string]string{"text": "checkpoint before continuation"}))
					return
				}
				ch5Reply(w, map[string]string{"type": "text", "text": "continued"})
			}))
			defer server.Close()
			app := ensemble.New(log.New(diagnostic, "", 0))
			defer func() {
				if err := app.Shutdown(); err != nil && !failSave {
					t.Error(err)
				}
			}()
			a, err := app.NewAgent(ensemble.Config{Vendor: ensemble.Anthropic, Model: "claude-sonnet-4", APIKey: "local-fixture-only", BaseURL: server.URL, BaseDir: t.TempDir(), DataDir: t.TempDir(), SavePath: savePath})
			if err != nil {
				t.Fatal(err)
			}
			if failSave {
				// Journal writes still use the sibling path; only snapshot replacement is blocked.
				if err := os.Mkdir(savePath, 0700); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(filepath.Join(savePath, "blocker"), []byte("owned failure fixture"), 0600); err != nil {
					t.Fatal(err)
				}
			}
			initial := []byte("before-retain-0\nbefore-retain-1\nbefore-retain-2\n")
			if _, err := diagnostic.Write(initial); err != nil {
				t.Fatal(err)
			}
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			_, askErr := a.Ask(ctx, "make a checkpoint")
			if failSave {
				if askErr == nil || !strings.Contains(askErr.Error(), savePath) {
					t.Fatalf("expected real snapshot failure, got %v (requests=%d)", askErr, requests.Load())
				}
				if requests.Load() != 1 {
					t.Fatalf("continued model requests after failed save: %d", requests.Load())
				}
				t.Logf("expected snapshot failure: %v", askErr)
				info, err := os.Stat(savePath)
				if err != nil || !info.IsDir() {
					t.Fatalf("failed save replaced blocker: %v", err)
				}
			} else {
				if askErr != nil {
					t.Fatal(askErr)
				}
				if requests.Load() != 2 {
					t.Fatalf("missing handoff continuation: %d", requests.Load())
				}
				saved, err := os.ReadFile(savePath)
				if err != nil {
					t.Fatal(err)
				}
				var snapshot struct {
					Context ensemble.Context `json:"context"`
				}
				if err := json.Unmarshal(saved, &snapshot); err != nil {
					t.Fatal(err)
				}
				if snapshot.Context.Handoff != "checkpoint before continuation" {
					t.Fatalf("successful handoff was not saved: %+v", snapshot.Context)
				}
			}
			before, err := os.ReadFile(diagPath)
			if err != nil {
				t.Fatal(err)
			}
			retainedOld := bytes.Contains(before, []byte("before-retain-0\n"))
			if retainedOld != failSave {
				t.Fatalf("diagnostic retention crossed wrong boundary (failed save=%v): %q", failSave, before)
			}
			if len(before) == 0 || before[len(before)-1] != '\n' {
				t.Fatalf("incomplete diagnostic record: %q", before)
			}
			continued := []byte(strings.Repeat("continued-record-", 500) + "\n")
			if _, err := diagnostic.Write(continued); err != nil {
				t.Fatal(err)
			}
			after, err := os.ReadFile(diagPath)
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.HasPrefix(after, before) || bytes.Count(after, continued) != 1 {
				t.Fatal("continued diagnostic write was overwritten, split, or duplicated")
			}
		})
	}
}
