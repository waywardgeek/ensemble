package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	"ensemble"
)

func TestCh20FactoryResolvesChildModel(t *testing.T) {
	models := make(chan string, 4)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Model string `json:"model"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Error(err)
		}
		models <- body.Model
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"id":"child","usage":{"prompt_tokens":1,"completion_tokens":1,"total_tokens":2},"choices":[{"message":{"role":"assistant","content":null,"tool_calls":[{"id":"finish","type":"function","function":{"name":"submit","arguments":"{\"result\":\"done\"}"}}]},"finish_reason":"tool_calls"}]}`)
	}))
	defer server.Close()
	root := ensemble.New(nil)
	defer root.Shutdown()
	if err := installAgentFactory(root); err != nil {
		t.Fatal(err)
	}
	base := t.TempDir()
	parent, err := root.CreateAgent(context.Background(), ensemble.Config{ID: "root", Vendor: ensemble.OpenAI, Model: "parent-alias", ResolvedModel: "parent-wire-model", APIKey: "dummy", BaseURL: server.URL, BaseDir: base, DataDir: filepath.Join(base, "root"), SavePath: filepath.Join(base, "root", "save.json"), RequireToolExplanation: true})
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	child, err := root.Spawn(ctx, parent.ID(), ensemble.Config{ID: "child", Prompt: "finish", Model: "child-wire-model"})
	if err != nil {
		t.Fatal(err)
	}
	outcome, err := root.Wait(ctx, child.ID())
	if err != nil || outcome.State != "submitted" {
		t.Fatalf("outcome %+v %v", outcome, err)
	}
	if model := <-models; model != "child-wire-model" {
		t.Fatalf("child request used %q", model)
	}
	if parent.Config().ResolvedModel != "parent-wire-model" {
		t.Fatal("parent model was mutated")
	}
}
