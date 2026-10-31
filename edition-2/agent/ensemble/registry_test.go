package ensemble_test

import (
	"context"
	"ensemble/ensemble"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// Restart the public root rather than reconstructing a peer manually. The next
// request must contain its old conversation, and deletion must remove its files.
func TestNamedRegistryRestartAndTree(t *testing.T) {
	replayed := make(chan bool, 1)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		if strings.Contains(string(body), "SECOND_TASK") {
			replayed <- strings.Contains(string(body), "FIRST_TASK")
		}
		fleetCall(w, "submit", map[string]string{"result": "saved answer"})
	}))
	t.Cleanup(server.Close)
	dir := t.TempDir()
	cfg := ensemble.Config{Model: "fake", BaseURL: server.URL, DataDir: filepath.Join(dir, "main"), BuiltinTools: true, Recall: ensemble.RecallSettings{Disabled: true}}
	start := func() *ensemble.Ensemble {
		e := ensemble.New(io.Discard)
		if _, err := e.Start(dir, cfg, ensemble.AgentSpec{}, make(humanInbox, 10)); err != nil {
			t.Fatal(err)
		}
		return e
	}
	app := start()
	peer, err := app.CreateAgent("reviewer", "fake", "")
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if _, err = peer.Ask(ctx, "FIRST_TASK"); err != nil {
		t.Fatal(err)
	}
	if err = app.Close(); err != nil {
		t.Fatal(err)
	}
	app = start()
	defer app.Close()
	peer, err = app.FindAgent("reviewer", nil)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = peer.Ask(ctx, "SECOND_TASK"); err != nil {
		t.Fatal(err)
	}
	if !<-replayed {
		t.Fatal("restart lost peer history")
	}
	if _, err = peer.Fleet().Spawn("reviewer", "CHILD_TASK", ""); err != nil {
		t.Fatal(err)
	}
	nested, err := app.FindAgent("reviewer/reviewer", nil)
	if err != nil || nested == peer || nested.ParentAgent() != peer {
		t.Fatal("path identity collapsed", err)
	}
	tree := app.TreeSnapshot(nil)
	if len(tree.Children) != 2 || tree.Children[1].Key != "reviewer" || !tree.Children[1].Persistent || !tree.Children[1].Deletable || tree.Children[1].Children[0].Persistent {
		t.Fatalf("wrong tree: %+v", tree)
	}
	if _, err = peer.Fleet().Join(ctx, []string{"reviewer"}); err != nil {
		t.Fatal(err)
	}
	if err = app.DeleteAgent("reviewer"); err != nil {
		t.Fatal(err)
	}
	if _, err = os.Stat(peer.Config().DataDir); !os.IsNotExist(err) {
		t.Fatal("deleted peer files remain", err)
	}
	if err = app.Close(); err != nil {
		t.Fatal(err)
	}
	app = start()
	defer app.Close()
	if _, err = app.FindAgent("reviewer", nil); err == nil {
		t.Fatal("deleted peer resurrected")
	}
}

// Deletion cannot remove the owner of a parked child. This exercises the actual
// root factory and child pointer, with a provider request that ends only on kill.
func TestNamedDeleteActiveChildGuard(t *testing.T) {
	entered := make(chan struct{})
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		io.Copy(io.Discard, r.Body)
		close(entered)
		<-r.Context().Done()
	}))
	t.Cleanup(server.Close)
	dir := t.TempDir()
	app := ensemble.New(io.Discard)
	cfg := ensemble.Config{Model: "fake", BaseURL: server.URL, DataDir: filepath.Join(dir, "main"), BuiltinTools: true, Recall: ensemble.RecallSettings{Disabled: true}}
	if _, err := app.Start(dir, cfg, ensemble.AgentSpec{}, make(humanInbox, 10)); err != nil {
		t.Fatal(err)
	}
	defer app.Close()
	peer, err := app.CreateAgent("worker", "fake", "")
	if err != nil {
		t.Fatal(err)
	}
	if _, err = peer.Fleet().Spawn("child", "TASK", ""); err != nil {
		t.Fatal(err)
	}
	<-entered
	if err = app.DeleteAgent("worker"); err == nil || !strings.Contains(err.Error(), "active children") {
		t.Fatal("deleted active tree", err)
	}
	if err = peer.Fleet().ShutdownChild("child"); err != nil {
		t.Fatal(err)
	}
	if err = app.DeleteAgent("worker"); err != nil {
		t.Fatal(err)
	}
}
