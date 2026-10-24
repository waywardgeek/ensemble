package cli

import (
	"context"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	"ensemble"
)

func TestCh20Round2ReplyCommand(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("reply became provider input")
		w.WriteHeader(400)
	}))
	defer server.Close()
	root := ensemble.New(nil)
	defer root.Shutdown()
	base := t.TempDir()
	create := func(id string) ensemble.Agent {
		a, err := root.NewAgent(ensemble.Config{ID: id, Model: "fixture", Vendor: ensemble.OpenAI, APIKey: "dummy", BaseURL: server.URL, BaseDir: base, DataDir: filepath.Join(base, id), SavePath: filepath.Join(base, id, "save.json")})
		if err != nil {
			t.Fatal(err)
		}
		return a
	}
	first, second := create("first"), create("second")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	questions := make(chan ensemble.Question, 2)
	replies := make(chan ensemble.Question, 2)
	failures := make(chan error, 2)
	root.SetHumanQuestionSink(func(q ensemble.Question) { questions <- q })
	ask := func(a ensemble.Agent) {
		go func() {
			q, err := root.SendMessage(ctx, a.ID(), root.HumanAddress(), "help")
			if err != nil {
				failures <- err
				return
			}
			replies <- q
		}()
	}
	receive := func() ensemble.Question {
		select {
		case q := <-questions:
			return q
		case err := <-failures:
			t.Fatal(err)
		case <-ctx.Done():
			t.Fatal(ctx.Err())
		}
		return ensemble.Question{}
	}
	answer := func() {
		select {
		case <-replies:
		case err := <-failures:
			t.Fatal(err)
		case <-ctx.Done():
			t.Fatal(ctx.Err())
		}
	}
	ask(first)
	receive()
	if err := replyToHuman(root, "the-question-answer is yes"); err != nil {
		t.Fatal(err)
	}
	answer()
	ask(first)
	ask(second)
	q1, q2 := receive(), receive()
	if q1.RequestID == q2.RequestID {
		t.Fatal("human message IDs collided")
	}
	if err := replyToHuman(root, "ambiguous answer"); err == nil {
		t.Fatal("ambiguous /reply accepted")
	}
	if err := replyToHuman(root, q1.RequestID+" explicit answer"); err != nil {
		t.Fatal(err)
	}
	answer()
	if err := replyToHuman(root, q1.RequestID+" stale answer"); err == nil {
		t.Fatal("stale /reply accepted")
	}
	if err := replyToHuman(root, "remaining answer"); err != nil {
		t.Fatal(err)
	}
	answer()
	if err := replyToHuman(root, "not a new prompt"); err == nil {
		t.Fatal("reply without pending question accepted")
	}
}
