package cli

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"ensemble"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestChapter18HumanModelAndStatus(t *testing.T) {
	peer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"model":"claude-sonnet-4-6","content":[{"type":"text","text":"ok"}],"usage":{"input_tokens":4,"output_tokens":2}}`)
	}))
	defer peer.Close()
	app := ensemble.New(nil)
	defer app.Shutdown()
	a, err := app.NewAgent(ensemble.Config{Vendor: ensemble.Anthropic, Model: "claude-sonnet-4-6", APIKey: "dummy", BaseURL: peer.URL, BaseDir: t.TempDir(), DataDir: t.TempDir(), Tools: []string{}})
	if err != nil {
		t.Fatal(err)
	}
	if _, err = a.Ask(context.Background(), "fixture"); err != nil {
		t.Fatal(err)
	}
	var out, diag bytes.Buffer
	input := "/model\n/status\n/model new-private-model\n/model\n/status\n/model gpt-5.4\n/status\n"
	if err = runChat(context.Background(), a, strings.NewReader(input), &out, &diag); err != nil {
		t.Fatal(err)
	}
	reader := bufio.NewReader(&out)
	var first, changed, failed ensemble.EngineStatus
	for _, status := range []*ensemble.EngineStatus{&first, &changed, &failed} {
		prefix, readErr := reader.ReadString('\n')
		if readErr != nil {
			t.Fatal(readErr)
		}
		cfg := a.Config()
		want := fmt.Sprintf("safe=%t web-search=%t sandbox-root=%q\n", cfg.SafeMode, cfg.EnableWebSearch, cfg.SandboxRoot)
		if prefix != want {
			t.Fatalf("status policy prefix: %q want %q", prefix, want)
		}
		raw, readErr := reader.ReadBytes('\n')
		if readErr != nil {
			t.Fatal(readErr)
		}
		if err = json.Unmarshal(raw, status); err != nil {
			t.Fatal(err)
		}
	}
	if first.SelectedModel != "claude-sonnet-4-6" || changed.SelectedModel != "new-private-model" || failed.SelectedModel != changed.SelectedModel || changed.LastModel != first.LastModel || changed.Cost != first.Cost || failed.Session != first.Session {
		t.Fatalf("human controls changed producer accounting: %+v %+v %+v", first, changed, failed)
	}
	if !strings.Contains(diag.String(), "model: claude-sonnet-4-6") || !strings.Contains(diag.String(), "model: new-private-model") || !strings.Contains(diag.String(), "provider") {
		t.Fatalf("human command feedback: %s", diag.String())
	}
}
