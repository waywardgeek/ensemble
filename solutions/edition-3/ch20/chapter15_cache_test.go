package ensemble_test

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"ensemble"
	"ensemble/internal/common"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestChapter15SentLensAndRestartIsolation(t *testing.T) {
	var sent []byte
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		sent, _ = io.ReadAll(r.Body)
		if calls == 1 {
			fmt.Fprint(w, `{"model":"claude-sonnet-4-6","content":[{"type":"text","text":"first"}],"usage":{"input_tokens":10,"output_tokens":4,"cache_read_input_tokens":80,"cache_creation_input_tokens":10}}`)
		} else {
			fmt.Fprint(w, `{"model":"claude-sonnet-4-6","content":[{"type":"text","text":"second"}],"usage":{"input_tokens":20,"output_tokens":5,"cache_read_input_tokens":null}}`)
		}
	}))
	defer server.Close()
	cfg := ensemble.Config{Model: "claude-opus-4-6", BaseURL: server.URL, APIKey: "dummy", SystemPrompt: "stable"}
	a := r2Agent(t, cfg)
	defer a.Shutdown()
	if a.Engine().Cache().Parent() != a.Engine() {
		t.Fatal("cache not owned by Engine")
	}
	preview, err := a.Render()
	if err != nil {
		t.Fatal(err)
	}
	if a.Engine().Status().Cache.Current != nil {
		t.Fatal("preview counted as send")
	}
	var original map[string]any
	_ = json.Unmarshal(preview, &original)
	if original["stream"] != true {
		t.Fatal("fixture lacks streaming preview")
	}
	if _, err = a.Engine().ExchangeRequest(context.Background(), preview, common.ParseCallbacks{}); err != nil {
		t.Fatal(err)
	}
	s := a.Engine().Status()
	sum := sha256.Sum256(sent)
	if s.Cache.Current.Digest != hex.EncodeToString(sum[:]) || s.Cache.Current.Bytes != len(sent) || !bytes.Equal(sent, preview) {
		t.Fatalf("lens digest=%s want=%s bytes=%d/%d previewSame=%v", s.Cache.Current.Digest, hex.EncodeToString(sum[:]), s.Cache.Current.Bytes, len(sent), bytes.Equal(sent, preview))
	}
	var actual map[string]any
	_ = json.Unmarshal(sent, &actual)
	if actual["stream"] != true {
		t.Fatal("stream mismatch")
	}
	if s.LastHitRate == nil || *s.LastHitRate != .8 || s.Requests != 1 || s.Cost.Status != "known" {
		t.Fatalf("first status %+v", s)
	}
	charged := s.Cost.USD
	if a.Engine().Status().Cost.USD != charged {
		t.Fatal("model change repriced earlier response")
	}
	if _, err = a.Ask(context.Background(), "persisted question"); err == nil {
		t.Fatal("invalid supplied cache count was accepted")
	}
	s = a.Engine().Status()
	if s.Requests != 2 || len(s.Buckets) != 1 || s.Buckets[0].Model != "claude-sonnet-4-6" || s.Cost.Status != "partial" || s.SessionHitRate != nil || s.Session.CacheRead != -1 || s.Session.Output != 9 {
		t.Fatalf("unknown or identity lost: %+v", s)
	}
	dump := r2Dump(t, a)
	if bytes.Contains(dump, []byte(`"buckets"`)) || bytes.Contains(dump, []byte(`"marker_reason"`)) || bytes.Contains(dump, []byte(`"usd"`)) {
		t.Fatal("ephemeral diagnostics or costs persisted")
	}
	fresh := r2Agent(t, cfg)
	defer fresh.Shutdown()
	if err = fresh.Log().Load(bytes.NewReader(dump)); err != nil {
		t.Fatal(err)
	}
	clean := fresh.Engine().Status()
	if clean.Requests != 0 || clean.Cost.USD != 0 || len(clean.Buckets) != 0 || clean.Cache.Current != nil || clean.LastHitRate != nil || clean.SessionHitRate != nil || clean.Session.Output != 0 {
		t.Fatalf("replay charged or carried lens: %+v", clean)
	}
	if fresh.Log().Context().Usage.Output != 5 || fresh.Log().Context().Usage.CacheRead != -1 {
		t.Fatal("historical facts lost")
	}
	_, _ = fresh.Render()
	_, _ = fresh.ParseResponse(200, []byte(`{"content":[{"type":"text","text":"offline"}],"usage":{"input_tokens":99,"output_tokens":99}}`))
	if fresh.Engine().Status().Requests != 0 || fresh.Engine().Status().Cache.Current != nil || calls != 2 {
		t.Fatal("offline projection called or charged provider")
	}
}
