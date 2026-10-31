package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

// The shipped binary must supply the human sink and accept terminal replies
// while the model is blocked. Library fixtures cannot establish this wiring.
func TestFleetCLIHumanReplyAndEscalation(t *testing.T) {
	binary := filepath.Join(t.TempDir(), "ensemble")
	if out, err := exec.Command("go", "build", "-o", binary, ".").CombinedOutput(); err != nil {
		t.Fatal(err, string(out))
	}
	for _, idle := range []bool{false, true} {
		t.Run(fmt.Sprint(idle), func(t *testing.T) {
			var calls atomic.Int32
			var replySeen atomic.Bool
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				body, _ := io.ReadAll(r.Body)
				n := calls.Add(1)
				if idle {
					fmt.Fprint(w, `{"content":[{"type":"text","text":"forgot to submit"}],"usage":{"input_tokens":1,"output_tokens":1}}`)
					return
				}
				name := "send_message_to_parent"
				args := map[string]any{"content": "HUMAN_QUESTION", "ai_callback_delay": 60}
				if n > 1 {
					// Inspect only returned blocks; the prompt and tool request are insufficient.
					var request struct {
						// Messages preserves provider dialogue boundaries.
						Messages []struct {
							// Content identifies result blocks independently of requested tool arguments.
							Content []struct {
								// Type distinguishes a result from a requested call.
								Type string
								// Content carries the actual returned reply.
								Content json.RawMessage
							}
						}
					}
					json.Unmarshal(body, &request)
					for _, m := range request.Messages {
						for _, p := range m.Content {
							if p.Type == "tool_result" && strings.Contains(string(p.Content), "COBALT_REPLY") && strings.Contains(string(p.Content), "user") {
								replySeen.Store(true)
							}
						}
					}
					name = "submit"
					args = map[string]any{"result": "CLI_COMPLETED"}
				}
				json.NewEncoder(w).Encode(map[string]any{"content": []any{map[string]any{"type": "tool_use", "id": "call", "name": name, "input": args}}, "usage": map[string]int{"input_tokens": 1, "output_tokens": 1}})
			}))
			defer server.Close()
			ctx, cancel := context.WithTimeout(context.Background(), 6*time.Second)
			defer cancel()
			cmd := exec.CommandContext(ctx, binary, "chat")
			cmd.Dir = t.TempDir()
			cmd.Env = []string{"PATH=" + os.Getenv("PATH"), "HOME=" + t.TempDir(), "EN_HOME=" + t.TempDir(), "LLM_VENDOR=anthropic", "LLM_MODEL=fake", "LLM_API_KEY=fake", "LLM_BASE_URL=" + server.URL, "EN_DISABLE_STREAMING=1", "EN_RECALL_DISABLED=1"}
			stdin, _ := cmd.StdinPipe()
			stderr, _ := cmd.StderrPipe()
			var stdout bytes.Buffer
			cmd.Stdout = &stdout
			if err := cmd.Start(); err != nil {
				t.Fatal(err)
			}
			userNotice := make(chan string, 1)
			var diagnostics bytes.Buffer
			drained := make(chan struct{})
			go func() {
				defer close(drained)
				scanner := bufio.NewScanner(stderr)
				for scanner.Scan() {
					line := scanner.Text()
					fmt.Fprintln(&diagnostics, line)
					if strings.HasPrefix(line, "[from main to user]") {
						select {
						case userNotice <- line:
						default:
						}
					}
				}
			}()
			io.WriteString(stdin, "Please carry out the terminal task.\n")
			select {
			case notice := <-userNotice:
				if idle {
					if !strings.Contains(notice, "three framework wakes") {
						t.Error("wrong escalation", notice)
					}
				} else {
					if !strings.Contains(notice, "HUMAN_QUESTION") {
						t.Error("question lost", notice)
					}
					io.WriteString(stdin, "/reply COBALT_REPLY\n")
				}
			case <-ctx.Done():
				t.Error("human sink was unreachable")
			}
			stdin.Close()
			err := cmd.Wait()
			<-drained
			if err != nil {
				t.Fatal(err, diagnostics.String())
			}
			if idle {
				if calls.Load() != 4 {
					t.Fatal("CLI idle bound", calls.Load())
				}
			} else if calls.Load() != 2 || !replySeen.Load() || !strings.Contains(stdout.String(), "CLI_COMPLETED") {
				t.Fatalf("human reply never resumed model: calls=%d reply=%v stdout=%s stderr=%s", calls.Load(), replySeen.Load(), stdout.String(), diagnostics.String())
			}
		})
	}
}
