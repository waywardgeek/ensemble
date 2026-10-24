package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"io"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"ensemble/gui"
)

type launcherOutput struct {
	mu sync.Mutex
	b  bytes.Buffer
}

func (b *launcherOutput) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.b.Write(p)
}
func (b *launcherOutput) String() string { b.mu.Lock(); defer b.mu.Unlock(); return b.b.String() }

// This executes the actual current main in an owned, race-instrumented child
// process. Readiness alone is not a pass: the bound server must serve the assets.
func TestCh20LauncherAssets(t *testing.T) {
	if os.Getenv("E3_CH20_GUI_LAUNCH") == "1" {
		var args []string
		if err := json.Unmarshal([]byte(os.Getenv("E3_CH20_GUI_ARGS")), &args); err != nil {
			t.Fatal(err)
		}
		flag.CommandLine = flag.NewFlagSet("gui", flag.ExitOnError)
		os.Args = append([]string{"gui"}, args...)
		main()
		return
	}
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	readyPattern := regexp.MustCompile(`Ensemble GUI (http://127\.0\.0\.1:[0-9]+)`)
	for _, mode := range []string{"default", "empty", "override", "invalid-override"} {
		t.Run(mode, func(t *testing.T) {
			cwd := t.TempDir()
			var providerCalls atomic.Int32
			peer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				providerCalls.Add(1)
				http.Error(w, "no provider work requested", 400)
			}))
			defer peer.Close()
			args := []string{"-port", "0"}
			expected := map[string][]byte{}
			if mode == "override" || mode == "invalid-override" {
				disk := t.TempDir()
				args = append(args, "-gui-dir", disk)
				if mode == "override" {
					for _, name := range []string{"index.html", "app.js", "style.css"} {
						data := []byte("OVERRIDE-" + name)
						if err := os.WriteFile(filepath.Join(disk, name), data, 0600); err != nil {
							t.Fatal(err)
						}
						expected[name] = data
					}
				}
			} else {
				if mode == "empty" {
					args = append(args, "-gui-dir", "")
				}
				for _, name := range []string{"index.html", "app.js", "style.css", "artifact-scroll.js", "renderers.js"} {
					data, err := fs.ReadFile(gui.Assets(), name)
					if err != nil {
						t.Fatal(err)
					}
					expected[name] = data
				}
			}
			encoded, _ := json.Marshal(args)
			ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
			defer cancel()
			cmd := exec.CommandContext(ctx, executable, "-test.run=^TestCh20LauncherAssets$")
			cmd.Dir = cwd
			cmd.Env = append(os.Environ(), "E3_CH20_GUI_LAUNCH=1", "E3_CH20_GUI_ARGS="+string(encoded),
				"LLM_VENDOR=anthropic", "LLM_MODEL=fixture", "LLM_API_KEY=owned-loopback-only", "ANTHROPIC_BASE_URL="+peer.URL,
				"LLM_BASE_URL="+peer.URL, "EN_SKILLS_DIR=", "EN_PRIMARY_SKILL=", "ENSEMBLE_SAVE=", "LLM_LOG_PATH=", "LLM_PAUSE_ON_START=false", "EN_DISABLE_STREAMING=true")
			input, err := cmd.StdinPipe()
			if err != nil {
				t.Fatal(err)
			}
			var output launcherOutput
			cmd.Stdout, cmd.Stderr = &output, &output
			if err := cmd.Start(); err != nil {
				t.Fatal(err)
			}
			done := make(chan struct{})
			var exitErr error
			go func() { exitErr = cmd.Wait(); close(done) }()
			defer func() { _ = input.Close(); cancel(); <-done }()
			if mode == "invalid-override" {
				_ = input.Close()
				<-done
				if exitErr == nil || !strings.Contains(output.String(), "GUI asset missing") || readyPattern.MatchString(output.String()) {
					t.Fatalf("invalid override startup: %v\n%s", exitErr, output.String())
				}
			} else {
				var address string
				for address == "" {
					if match := readyPattern.FindStringSubmatch(output.String()); len(match) == 2 {
						address = match[1]
						break
					}
					select {
					case <-done:
						t.Fatalf("launcher exited before readiness: %v\n%s", exitErr, output.String())
					case <-ctx.Done():
						t.Fatalf("launcher readiness timeout\n%s", output.String())
					case <-time.After(10 * time.Millisecond):
					}
				}
				if strings.HasSuffix(address, ":0") {
					t.Fatalf("readiness did not name the bound address: %s\n%s", address, output.String())
				}
				client := &http.Client{Timeout: time.Second}
				for name, want := range expected {
					path := "/" + name
					if name == "index.html" {
						path = "/"
					}
					response, err := client.Get(address + path)
					if err != nil {
						t.Fatalf("HTTP %s: %v\n%s", name, err, output.String())
					}
					got, readErr := io.ReadAll(response.Body)
					response.Body.Close()
					if readErr != nil || response.StatusCode != http.StatusOK || !bytes.Equal(got, want) {
						t.Fatalf("asset %s status=%d bytes=%d/%d err=%v", name, response.StatusCode, len(got), len(want), readErr)
					}
				}
				_ = input.Close()
				<-done
				if exitErr != nil || ctx.Err() != nil {
					t.Fatalf("normal EOF: %v %v\n%s", exitErr, ctx.Err(), output.String())
				}
				data, err := os.ReadFile(filepath.Join(cwd, "save.json"))
				if err != nil || !json.Valid(data) {
					t.Fatalf("EOF did not preserve a valid save: %v\n%s", err, output.String())
				}
			}
			if providerCalls.Load() != 0 {
				t.Fatalf("startup/EOF performed %d provider calls", providerCalls.Load())
			}
		})
	}
}
