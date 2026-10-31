package ensemble_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"ensemble/ensemble"
	"ensemble/internal/common"
)

const recallQuery = "Please explain the quartz zephyr migration decision and the rejected deployment approach from our archived project discussion."
const recallFact = "## Quartz zephyr migration\nWe selected the copper ferry because the silver bridge lost ordering."

func recallCorpus(t *testing.T, root string) {
	t.Helper()
	for _, dir := range []string{"memory", "docs", "handoffs", "learnings"} {
		if err := os.MkdirAll(filepath.Join(root, dir), 0700); err != nil {
			t.Fatal(err)
		}
	}
	// Retired memories are precisely the bytes no longer present in live bands.
	for i := 0; i < 3; i++ {
		recallWrite(t, filepath.Join(root, "memory", fmt.Sprintf("2026-10-10-%d.md.retired", i+1)), recallFact+fmt.Sprintf("\nMemory decision %d.", i))
	}
	for i := 0; i < 25; i++ {
		recallWrite(t, filepath.Join(root, "memory", fmt.Sprintf("padding-%d.md", i)), "## Gardening\nPlant carrots beneath the apple tree.")
		recallWrite(t, filepath.Join(root, "docs", fmt.Sprintf("padding-%d.md", i)), "## Kitchen\nBread and butter for dinner.")
	}
	for i := 0; i < 3; i++ {
		recallWrite(t, filepath.Join(root, "docs", fmt.Sprintf("design-%d.md", i)), "## Quartz zephyr migration\n"+strings.Repeat("quartz zephyr migration ", 20)+fmt.Sprintf("doc decision %d.", i))
	}
}
func recallWrite(t *testing.T, path, text string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(text), 0600); err != nil {
		t.Fatal(err)
	}
}

// recallPeer is an HTTP boundary. All indexing, judging, actor, history and
// renderer behavior stays real; it retains received payloads for assertions.
type recallPeer struct {
	mu           sync.Mutex
	vendor       ensemble.Vendor
	judgment     string
	fail         bool
	requests     []string
	judgeStarted chan struct{}
}

func (p *recallPeer) serve(w http.ResponseWriter, r *http.Request) {
	data, _ := io.ReadAll(r.Body)
	p.mu.Lock()
	p.requests = append(p.requests, string(data))
	isJudge := strings.Contains(string(data), "zero-based candidate indices")
	reply := "Received the recalled decision."
	if isJudge {
		reply = p.judgment
	}
	fail, started := p.fail, p.judgeStarted
	p.mu.Unlock()
	if isJudge && started != nil {
		select {
		case started <- struct{}{}:
		default:
		}
		<-r.Context().Done()
		return
	}
	if isJudge && fail {
		http.Error(w, "unavailable", 503)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	switch p.vendor {
	case ensemble.Anthropic:
		json.NewEncoder(w).Encode(map[string]any{"id": "r", "model": "fake-model", "type": "message", "role": "assistant", "content": []any{map[string]any{"type": "text", "text": reply}}, "stop_reason": "end_turn", "usage": map[string]int{"input_tokens": 10, "output_tokens": 2}})
	case ensemble.OpenAI:
		json.NewEncoder(w).Encode(map[string]any{"model": "fake-model", "choices": []any{map[string]any{"message": map[string]any{"role": "assistant", "content": reply}, "finish_reason": "stop"}}, "usage": map[string]int{"prompt_tokens": 10, "completion_tokens": 2}})
	case ensemble.Gemini:
		json.NewEncoder(w).Encode(map[string]any{"modelVersion": "fake-model", "candidates": []any{map[string]any{"content": map[string]any{"role": "model", "parts": []any{map[string]any{"text": reply}}}, "finishReason": "STOP"}}, "usageMetadata": map[string]int{"promptTokenCount": 10, "candidatesTokenCount": 2, "totalTokenCount": 12}})
	}
}
func (p *recallPeer) wires() []string {
	p.mu.Lock()
	defer p.mu.Unlock()
	return append([]string(nil), p.requests...)
}
func recallEntries(a ensemble.Agent) []common.Entry {
	var entries []common.Entry
	for _, e := range a.History().Context().Dialogue {
		if e.Kind == common.KindRecall {
			entries = append(entries, e)
		}
	}
	return entries
}
func recallAgent(t *testing.T, vendor ensemble.Vendor, peer *recallPeer, cfg ensemble.RecallSettings) ensemble.Agent {
	t.Helper()
	root := t.TempDir()
	recallCorpus(t, root)
	server := httptest.NewServer(http.HandlerFunc(peer.serve))
	t.Cleanup(server.Close)
	a, err := ensemble.New(io.Discard).NewAgent(ensemble.Config{Vendor: vendor, Model: "fake-model", APIKey: "test", BaseURL: server.URL, DataDir: root, DisableStreaming: true, Recall: cfg})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { a.Shutdown() })
	return a
}

// TestRecallTurnReplayCheckpoint covers both public turn paths and all provider
// renderers. Replaying after disk deletion must reproduce the exact request.
func TestRecallTurnReplayCheckpoint(t *testing.T) {
	for _, vendor := range []ensemble.Vendor{ensemble.Anthropic, ensemble.OpenAI, ensemble.Gemini} {
		t.Run(vendor.String(), func(t *testing.T) {
			peer := &recallPeer{vendor: vendor}
			a := recallAgent(t, vendor, peer, ensemble.RecallSettings{DisableJudge: true})
			if a.Recaller().Agent() != a {
				t.Fatal("recall owner missing")
			}
			if _, err := a.Ask(context.Background(), recallQuery); err != nil {
				t.Fatal(err)
			}
			entries := recallEntries(a)
			if len(entries) != 1 {
				t.Fatalf("recall entries: %v", entries)
			}
			text := entries[0].Parts[0].Text
			if strings.Count(text, "From memory/") != 2 || strings.Count(text, "From docs/") != 1 {
				t.Fatalf("mechanical quota: %s", text)
			}
			for _, e := range a.History().Context().Dialogue {
				if e.Kind == common.KindDialogue && e.Actor == common.Human && e.Parts[0].Text != recallQuery {
					t.Fatal("user message modified")
				}
			}
			if !strings.Contains(peer.wires()[0], "[Auto-recalled memories]") {
				t.Fatal("attachment not rendered")
			}
			before, _ := a.Engine().Render(a.History().Context())
			var log bytes.Buffer
			if err := a.History().Dump(&log); err != nil {
				t.Fatal(err)
			}
			if !strings.Contains(log.String(), `"recall_attached":{"parts"`) {
				t.Fatal("recorded bytes missing")
			}
			os.RemoveAll(a.Config().DataDir)
			if err := a.History().Load(&log); err != nil {
				t.Fatal(err)
			}
			after, _ := a.Engine().Render(a.History().Context())
			if !bytes.Equal(before, after) || len(peer.wires()) != 1 {
				t.Fatal("replay changed bytes or called model")
			}
			// Restore corpus for real Post path, then checkpoint and recall again.
			recallCorpus(t, a.Config().DataDir)
			if err := a.Record(common.Event{Type: "micro_handoff", Handoff: &common.MicroHandoffData{Text: "checkpoint"}}); err != nil {
				t.Fatal(err)
			}
			if len(recallEntries(a)) != 0 {
				t.Fatal("checkpoint retained recall")
			}
			if err := a.Post(ensemble.UserMessage{Text: recallQuery}); err != nil {
				t.Fatal(err)
			}
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()
			if err := a.Flush(ctx); err != nil {
				t.Fatal(err)
			}
			if len(recallEntries(a)) != 1 {
				t.Fatal("Post path skipped recall")
			}
		})
	}
}

// TestRecallJudge distinguishes successful filtering (including empty) from
// transport/format failure. A successful all-doc verdict overrides memory quotas.
func TestRecallJudge(t *testing.T) {
	for _, tc := range []struct {
		name, reply string
		fail        bool
		mem, docs   int
	}{
		{"selected", `["0",1,2,3,2,900,-1]`, false, 0, 3},
		{"empty", `[]`, false, 0, 0},
		{"malformed", "not an array", false, 2, 1},
		{"failure", "", true, 2, 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			peer := &recallPeer{vendor: ensemble.OpenAI, judgment: tc.reply, fail: tc.fail}
			a := recallAgent(t, ensemble.OpenAI, peer, ensemble.RecallSettings{JudgeModel: "small-judge"})
			if _, err := a.Ask(context.Background(), recallQuery); err != nil {
				t.Fatal(err)
			}
			text := ""
			for _, e := range recallEntries(a) {
				text += e.Parts[0].Text
			}
			if strings.Count(text, "From memory/") != tc.mem || strings.Count(text, "From docs/") != tc.docs {
				t.Fatalf("selection: %s", text)
			}
			wires := peer.wires()
			if len(wires) != 2 {
				t.Fatalf("judge/turn calls: %d", len(wires))
			}
			var judge map[string]any
			json.Unmarshal([]byte(wires[0]), &judge)
			if judge["tools"] != nil || judge["model"] != "small-judge" || len(judge["messages"].([]any)) != 1 {
				t.Fatal("judge is not isolated tool-free route")
			}
			if a.History().Context().LastUsage.Input != 10 || a.Engine().Usage().Input < 10 {
				t.Fatal("judge replaced main usage")
			}
		})
	}
}

// TestRecallDedupRefreshIsolation checks actual-context filtering, latest files,
// same-size human edits and optional disabling without another Agent's corpus.
func TestRecallDedupRefreshIsolation(t *testing.T) {
	peer := &recallPeer{vendor: ensemble.OpenAI}
	a := recallAgent(t, ensemble.OpenAI, peer, ensemble.RecallSettings{DisableJudge: true})
	if got, err := recallText(a, context.Background(), "quartz", &common.Context{}); err != nil || got != "" {
		t.Fatal("short query recalled")
	}
	path := filepath.Join(a.Config().DataDir, "memory", "2099-12-31-1.md")
	recallWrite(t, path, recallFact+" newest fact")
	c := &common.Context{}
	text, err := recallText(a, context.Background(), recallQuery+" newest", c)
	if err != nil || !strings.Contains(text, "2099-12-31-1.md") {
		t.Fatalf("newest excluded: %s %v", text, err)
	}
	recallWrite(t, path, strings.ReplaceAll(recallFact+" newest fact", "copper", "golden"))
	text, err = recallText(a, context.Background(), recallQuery+" newest", c)
	if err != nil || !strings.Contains(text, "golden ferry") {
		t.Fatalf("edit not indexed: %s %v", text, err)
	}
	// Every file's content is actual context data, independent of entry kind.
	var content string
	filepath.WalkDir(a.Config().DataDir, func(path string, d os.DirEntry, err error) error {
		if err == nil && !d.IsDir() {
			b, _ := os.ReadFile(path)
			content += string(b) + "\n"
		}
		return err
	})
	c.Dialogue = []common.Entry{{Parts: []common.Part{{Type: "text", Text: content}}}}
	if text, err = recallText(a, context.Background(), recallQuery+" newest", c); err != nil || text != "" {
		t.Fatal("already-present content recalled")
	}
	other, err := ensemble.New(io.Discard).NewAgent(ensemble.Config{Model: "offline", DataDir: t.TempDir(), Recall: ensemble.RecallSettings{DisableJudge: true}})
	if err != nil {
		t.Fatal(err)
	}
	defer other.Shutdown()
	if text, err = recallText(other, context.Background(), recallQuery, &common.Context{}); err != nil || text != "" {
		t.Fatal("cross-Agent corpus leak")
	}
	disabled, err := ensemble.New(io.Discard).NewAgent(ensemble.Config{Model: "offline", Recall: ensemble.RecallSettings{Disabled: true}})
	if err != nil {
		t.Fatal(err)
	}
	defer disabled.Shutdown()
	if disabled.Recaller() != nil {
		t.Fatal("disabled recall installed")
	}
}

// TestRecallInterrupt keeps the actor responsive while a slow judge owns HTTP.
func TestRecallInterrupt(t *testing.T) {
	peer := &recallPeer{vendor: ensemble.OpenAI, judgeStarted: make(chan struct{}, 1)}
	a := recallAgent(t, ensemble.OpenAI, peer, ensemble.RecallSettings{})
	result := a.Submit(context.Background(), recallQuery)
	select {
	case <-peer.judgeStarted:
	case <-time.After(time.Second):
		t.Fatal("judge never ran")
	}
	// A recorded barrier proves the actor processed the hint before inspection.
	if err := a.Post(ensemble.Hint{Text: "steer during recall"}); err != nil {
		t.Fatal(err)
	}
	if err := a.Ephemeral("barrier"); err != nil {
		t.Fatal(err)
	}
	view := a.History().Context()
	if len(view.Ephemera) != 2 || view.Ephemera[0].Text != "steer during recall" {
		t.Fatal("hint became retained dialogue during recall")
	}
	if err := a.Post(ensemble.Interrupt{}); err != nil {
		t.Fatal(err)
	}
	select {
	case got := <-result:
		if !errors.Is(got.Err, ensemble.ErrInterrupted) {
			t.Fatal(got.Err)
		}
	case <-time.After(time.Second):
		t.Fatal("judge blocked interrupt")
	}
	if _, err := a.Ask(context.Background(), "hello"); err != nil {
		t.Fatal(err)
	}
	if len(recallEntries(a)) != 0 {
		t.Fatal("interrupted recall attached to next turn")
	}
}

// TestRecallJudgeTimeout verifies the chapter's own 15-second deadline, rather
// than relying on a caller deadline or the transport's longer timeout.
func TestRecallJudgeTimeout(t *testing.T) {
	peer := &recallPeer{vendor: ensemble.OpenAI, judgeStarted: make(chan struct{}, 1)}
	a := recallAgent(t, ensemble.OpenAI, peer, ensemble.RecallSettings{})
	ctx, cancel := context.WithTimeout(context.Background(), 18*time.Second)
	defer cancel()
	start := time.Now()
	text, err := recallText(a, ctx, recallQuery, &common.Context{})
	elapsed := time.Since(start)
	if err != nil || strings.Count(text, "From memory/") != 2 || elapsed < 14*time.Second || elapsed > 17*time.Second {
		t.Fatalf("judge timeout/fallback: %v %v %s", elapsed, err, text)
	}
}

// TestRecallJudgeContext bounds the prompt and keeps conversation as data in
// one request. No messages or selected candidates survive in a judge context.
func TestRecallJudgeContext(t *testing.T) {
	peer := &recallPeer{vendor: ensemble.OpenAI, judgment: "[]"}
	a := recallAgent(t, ensemble.OpenAI, peer, ensemble.RecallSettings{Candidates: 2, ContextMsgs: 3})
	c := &common.Context{}
	for _, text := range []string{"old-excluded", "recent-one", "recent-two", "recent-three"} {
		c.Dialogue = append(c.Dialogue, common.Entry{Kind: common.KindDialogue, Actor: common.Human, Parts: []common.Part{{Type: "text", Text: text}}})
	}
	if _, err := recallText(a, context.Background(), recallQuery, c); err != nil {
		t.Fatal(err)
	}
	wire := peer.wires()[0]
	if strings.Contains(wire, "old-excluded") || !strings.Contains(wire, "recent-one") || !strings.Contains(wire, "recent-three") || strings.Contains(wire, "[2]") || !strings.Contains(wire, "[1]") {
		t.Fatal("judge prompt limits wrong")
	}
}

// TestRecallSources discovers manuals without loading their tools, and keeps
// optional document corpora under the Agent's own directory.
func TestRecallSources(t *testing.T) {
	root := t.TempDir()
	// The CLI uses a relative data directory with an independently absolute skill root.
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	root, err = filepath.Rel(cwd, root)
	if err != nil {
		t.Fatal(err)
	}
	skills := skillFiles(t, map[string]string{
		"base":     "---\nname: base\ntype: primary\nloadable-skills: calendar\n---\nAnswer briefly.",
		"calendar": "---\nname: calendar\ntype: loadable\n---\n## Iridium calendar\nUse the iridium calendar to schedule telescope bookings.",
	})
	for _, source := range []string{"docs", "handoffs", "learnings"} {
		if err := os.MkdirAll(filepath.Join(root, source), 0700); err != nil {
			t.Fatal(err)
		}
		recallWrite(t, filepath.Join(root, source, "record.md"), "## Iridium "+source+"\nThe iridium telescope booking belongs to "+source+".")
	}
	a, err := ensemble.New(io.Discard).NewAgent(ensemble.Config{Model: "offline", DataDir: root, Recall: ensemble.RecallSettings{DisableJudge: true, MinScore: .1, MaxSnippets: 4}}, ensemble.AgentSpec{SkillsDir: skills, PrimarySkill: "base"})
	if err != nil {
		t.Fatal(err)
	}
	defer a.Shutdown()
	text, err := recallText(a, context.Background(), "Please explain the iridium telescope booking calendar instructions and the archived documentation from our previous project discussion.", &common.Context{})
	if err != nil {
		t.Fatal(err)
	}
	for _, source := range []string{"docs/record.md", "handoffs/record.md", "learnings/record.md", "skills/calendar/SKILL.md"} {
		if !strings.Contains(text, "From "+source) {
			t.Fatalf("missing source %s: %s", source, text)
		}
	}
	if got := a.Skills().LoadableSkills(); len(got) != 1 || got[0] != "calendar" {
		t.Fatal("recall activated skill")
	}
}

// TestRecallJudgeRoute keeps model-specific headers with the selected stateless
// request, even when the main Agent uses the course inline-tools fixture model.
func TestRecallJudgeRoute(t *testing.T) {
	beta := make(chan string, 1)
	peer := &recallPeer{vendor: ensemble.Anthropic, judgment: "[]"}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { beta <- r.Header.Get("anthropic-beta"); peer.serve(w, r) }))
	defer server.Close()
	a, err := ensemble.New(io.Discard).NewAgent(ensemble.Config{Vendor: ensemble.Anthropic, Model: "claude-opus-5-course", BaseURL: server.URL, DataDir: t.TempDir(), Recall: ensemble.RecallSettings{JudgeModel: "claude-sonnet-5"}})
	if err != nil {
		t.Fatal(err)
	}
	defer a.Shutdown()
	if _, err := a.Engine().Judge(context.Background(), "Return a JSON array of zero-based candidate indices: []"); err != nil {
		t.Fatal(err)
	}
	if <-beta != "" {
		t.Fatal("judge inherited parent model header")
	}
	if !strings.Contains(peer.wires()[0], "claude-sonnet-5") {
		t.Fatal("judge route model missing")
	}
}

// recallText exercises the real service while leaving presentation assertions concise.
func recallText(a ensemble.Agent, ctx context.Context, query string, c *common.Context) (string, error) {
	data, err := a.Recaller().Recall(ctx, query, c)
	if len(data.Parts) == 0 {
		return "", err
	}
	return data.Parts[0].Text, err
}

// TestRecallCappedDedup retains the identity of a partially shown snippet through
// replay, but permits a real source edit even beyond the visible capped prefix.
func TestRecallCappedDedup(t *testing.T) {
	peer := &recallPeer{vendor: ensemble.OpenAI}
	a := recallAgent(t, ensemble.OpenAI, peer, ensemble.RecallSettings{DisableJudge: true, MaxSnippets: 1})
	root := a.Config().DataDir
	if err := os.RemoveAll(root); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(root, "memory"), 0700); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 25; i++ {
		recallWrite(t, filepath.Join(root, "memory", fmt.Sprintf("padding-%d.md.retired", i)), "Garden carrots and apples.")
	}
	path := filepath.Join(root, "memory", "2026-10-01-1.md.retired")
	source := "## Quartz zephyr migration\n" + strings.Repeat("quartz zephyr migration preserves transaction order. ", 230) + "OLD_END"
	recallWrite(t, path, source)
	if _, err := a.Ask(context.Background(), recallQuery); err != nil {
		t.Fatal(err)
	}
	entries := recallEntries(a)
	if len(entries) != 1 || len(entries[0].Parts[0].Text) != 6144 || len(entries[0].RecallHashes) != 1 {
		t.Fatal("capped attachment or identity missing")
	}
	var saved bytes.Buffer
	if err := a.History().Dump(&saved); err != nil {
		t.Fatal(err)
	}
	if err := a.History().Load(&saved); err != nil {
		t.Fatal(err)
	}
	if _, err := a.Ask(context.Background(), recallQuery); err != nil {
		t.Fatal(err)
	}
	if len(recallEntries(a)) != 1 || strings.Count(peer.wires()[1], "[Auto-recalled memories]") != 1 {
		t.Fatal("capped snippet recalled twice")
	}
	// The changed suffix was not shown before. Identity must describe source
	// content, not merely the filename or the clipped display bytes.
	recallWrite(t, path, strings.ReplaceAll(source, "OLD_END", "NEW_END"))
	if _, err := a.Ask(context.Background(), recallQuery); err != nil {
		t.Fatal(err)
	}
	if len(recallEntries(a)) != 2 {
		t.Fatal("edited capped source not recalled")
	}
}
