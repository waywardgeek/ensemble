package llm

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/waywardgeek/ensemble/agent/internal/common"
)

// fakeCompressor stands in for the vendor during compaction.
//
// It answers every request with a submit call, which is what a compressor is
// required to do. The bodies it received are kept so a test can assert on
// what the compressor was actually shown, which is the part of this chapter
// that is easy to get wrong and impossible to see from the outside.
type fakeCompressor struct {
	srv    *httptest.Server
	bodies []string
	reply  func(n int) string
}

func newFakeCompressor(t *testing.T) *fakeCompressor {
	t.Helper()
	f := &fakeCompressor{}
	f.srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Errorf("reading request body: %v", err)
		}
		f.bodies = append(f.bodies, string(body))

		text := fmt.Sprintf("MEMORY-%d", len(f.bodies))
		if f.reply != nil {
			text = f.reply(len(f.bodies))
		}
		input, _ := json.Marshal(map[string]string{"memory": text})

		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"id":"msg_1","type":"message","role":"assistant","model":"test",
			"content":[{"type":"tool_use","id":"tu_1","name":"submit","input":%s}],
			"stop_reason":"tool_use","usage":{"input_tokens":10,"output_tokens":5}}`, input)
	}))
	t.Cleanup(f.srv.Close)
	return f
}

// compactEngine builds an engine wired to a fake vendor and a real memory
// directory, with watermarks low enough that a handful of turns trips them.
func compactEngine(t *testing.T, f *fakeCompressor, cfg common.BandConfig) (*Engine, *Store, string) {
	t.Helper()
	dir := t.TempDir()
	store := NewStore(dir)
	e := &Engine{
		Log:    NewLog(),
		Ctx:    common.NewContext(),
		parent: testHost{t},
		Cfg: common.Config{
			Vendor:           common.VendorAnthropic,
			Surface:          common.SurfaceMessages,
			Model:            "test",
			BaseURL:          f.srv.URL,
			APIKey:           "test",
			MaxTokens:        1024,
			DisableStreaming: true,
		},
		HTTP:   f.srv.Client(),
		Memory: store,
		Bands:  func() common.BandConfig { return cfg },
	}
	return e, store, dir
}

type testHost struct{ t *testing.T }

func (h testHost) Logf(format string, args ...any)    { h.t.Logf(format, args...) }
func (h testHost) APILogf(format string, args ...any) { h.t.Logf(format, args...) }
func (h testHost) Debugf(format string, args ...any)  { h.t.Logf(format, args...) }

// testHost is a value type, so it cannot embed common.UsageCounter: that would
// copy a mutex on every assignment. These tests exercise compaction and assert
// nothing about spend, so the counts are discarded rather than tracked.
func (h testHost) RecordUsage(common.Usage)   {}
func (h testHost) SessionUsage() common.Usage { return common.Usage{} }

// bandText pulls the text of a band's entries, which is what the model
// actually sees.
func bandText(c *common.Context, b common.Band) []string {
	var out []string
	for _, e := range BandEntries(c, b) {
		for _, p := range e.Parts {
			if tp, ok := p.(common.TextPart); ok {
				out = append(out, tp.Text)
			}
		}
	}
	return out
}

// say adds one user turn of roughly n bytes.
func say(e *Engine, n int) {
	e.Record(common.Event{
		Type: common.MessageReceived,
		Message: &common.MessageData{
			Actor: common.ActorHuman,
			Parts: common.PartList{common.TextPart{Text: strings.Repeat("u", n)}},
		},
	})
}

func checkpoint(e *Engine) {
	e.Record(common.Event{
		Type:    common.MicroHandoff,
		Handoff: &common.MicroHandoffData{Text: "checkpoint"},
	})
}

// TestCompactionSwapsConversationForMemory is the chapter's central claim:
// finished work leaves the conversation and comes back as one memory entry,
// and the memory that comes back is the one written to disk.
func TestCompactionSwapsConversationForMemory(t *testing.T) {
	f := newFakeCompressor(t)
	cfg := common.DefaultBandConfig()
	cfg.Conversation.Budget = 1000 // High() is 2*Budget, so 2000
	e, store, _ := compactEngine(t, f, cfg)

	for i := 0; i < 4; i++ {
		say(e, 500)
	}
	checkpoint(e)
	say(e, 400) // the live segment, which must survive

	before := ConversationBytes(e.Ctx)
	e.maybeCompact()
	after := ConversationBytes(e.Ctx)

	if after >= before {
		t.Fatalf("conversation did not shrink: %d then %d", before, after)
	}
	if len(f.bodies) != 1 {
		t.Fatalf("want exactly one compressor call, got %d", len(f.bodies))
	}

	got := bandText(e.Ctx, common.BandSession)
	if len(got) != 1 || !strings.Contains(got[0], "MEMORY-1") {
		t.Fatalf("session band should hold the compressed memory, got %q", got)
	}

	files, err := store.Files(common.BandSession)
	if err != nil {
		t.Fatalf("Files: %v", err)
	}
	if len(files) != 1 || files[0].Text != "MEMORY-1" {
		t.Fatalf("disk and context disagree: %+v vs %q", files, got)
	}

	// The live segment is the one thing compaction may never take.
	if !strings.Contains(renderConversation(e.Ctx, 0, common.Seq(^uint64(0))), strings.Repeat("u", 400)) {
		t.Fatal("the live segment was compacted away")
	}
}

// TestCompactionWillNotEatTheLiveSegment is the refusal case: a conversation
// over its watermark with no checkpoint at all must be left entirely alone.
func TestCompactionWillNotEatTheLiveSegment(t *testing.T) {
	f := newFakeCompressor(t)
	cfg := common.DefaultBandConfig()
	cfg.Conversation.Budget = 500
	e, _, _ := compactEngine(t, f, cfg)

	for i := 0; i < 6; i++ {
		say(e, 500)
	}
	// No checkpoint has ever landed.

	before := ConversationBytes(e.Ctx)
	e.maybeCompact()

	if ConversationBytes(e.Ctx) != before {
		t.Fatalf("compacted unfinished work: %d became %d", before, ConversationBytes(e.Ctx))
	}
	if len(f.bodies) != 0 {
		t.Fatalf("called the compressor with nothing safe to compress: %d calls", len(f.bodies))
	}
}

// TestGraduationFoldsEightIntoOne walks the second rung: once the session
// band holds FoldFactor files they become one 8x file, the sources are
// retired from disk, and the context follows.
func TestGraduationFoldsEightIntoOne(t *testing.T) {
	f := newFakeCompressor(t)
	cfg := common.DefaultBandConfig()
	e, store, _ := compactEngine(t, f, cfg)

	for i := 0; i < FoldFactor; i++ {
		file, err := store.WriteSession(fmt.Sprintf("session memory %d", i))
		if err != nil {
			t.Fatalf("WriteSession: %v", err)
		}
		e.Record(common.Event{
			Type: common.BandPopulated,
			BandAdd: &common.BandPopulatedData{
				Band: common.BandSession, File: file.ID, Text: file.Text, Source: "compaction",
			},
		})
	}

	e.graduate(common.BandSession, cfg)

	sess, err := store.Files(common.BandSession)
	if err != nil {
		t.Fatalf("Files: %v", err)
	}
	if len(sess) != 0 {
		t.Fatalf("sources were not retired: %d left", len(sess))
	}
	up, err := store.Files(common.Band8x)
	if err != nil {
		t.Fatalf("Files: %v", err)
	}
	if len(up) != 1 || up[0].Text != "MEMORY-1" {
		t.Fatalf("want one 8x memory holding the fold, got %+v", up)
	}
	if got := bandText(e.Ctx, common.BandSession); len(got) != 0 {
		t.Fatalf("session band should be empty in context, got %q", got)
	}
	if got := bandText(e.Ctx, common.Band8x); len(got) != 1 || !strings.Contains(got[0], "MEMORY-1") {
		t.Fatalf("8x band should hold the fold, got %q", got)
	}

	// The compressor must have been shown all eight, or it folded a lie.
	for i := 0; i < FoldFactor; i++ {
		if !strings.Contains(f.bodies[0], fmt.Sprintf("session memory %d", i)) {
			t.Fatalf("compressor never saw session memory %d", i)
		}
	}
}

// TestGraduationRefusesWhenTheBandAboveIsOff is the loud-refusal rule. The
// two silent alternatives both lose: dropping destroys memory, carrying on
// grows without bound.
func TestGraduationRefusesWhenTheBandAboveIsOff(t *testing.T) {
	f := newFakeCompressor(t)
	cfg := common.DefaultBandConfig()
	cfg.B8x.Disabled = true
	e, store, _ := compactEngine(t, f, cfg)

	for i := 0; i < FoldFactor; i++ {
		if _, err := store.WriteSession(fmt.Sprintf("session memory %d", i)); err != nil {
			t.Fatalf("WriteSession: %v", err)
		}
	}

	e.graduate(common.BandSession, cfg)

	if len(f.bodies) != 0 {
		t.Fatalf("compressed into a band that is switched off: %d calls", len(f.bodies))
	}
	sess, err := store.Files(common.BandSession)
	if err != nil {
		t.Fatalf("Files: %v", err)
	}
	if len(sess) != FoldFactor {
		t.Fatalf("memories were lost to a disabled band: %d of %d left", len(sess), FoldFactor)
	}
}

// TestAbandonedGraduationWaitsForACheckpoint covers the crash case. A launch
// with no completion must not be retried on the very next turn, or a
// compressor that kills the agent kills it again forever.
func TestAbandonedGraduationWaitsForACheckpoint(t *testing.T) {
	f := newFakeCompressor(t)
	cfg := common.DefaultBandConfig()
	e, store, _ := compactEngine(t, f, cfg)

	for i := 0; i < FoldFactor; i++ {
		if _, err := store.WriteSession(fmt.Sprintf("session memory %d", i)); err != nil {
			t.Fatalf("WriteSession: %v", err)
		}
	}
	files, err := store.Files(common.BandSession)
	if err != nil {
		t.Fatalf("Files: %v", err)
	}
	thru := files[FoldFactor-1].ID

	// A launch that never completed, exactly as a crash would leave it.
	e.Record(common.Event{
		Type:    common.CompactorLaunched,
		Compact: &common.CompactorLaunchData{Band: common.BandSession, Thru: thru},
	})

	e.graduate(common.BandSession, cfg)
	if len(f.bodies) != 0 {
		t.Fatalf("retried an abandoned graduation immediately: %d calls", len(f.bodies))
	}

	// A checkpoint is the evidence that something has changed.
	checkpoint(e)
	e.graduate(common.BandSession, cfg)
	if len(f.bodies) != 1 {
		t.Fatalf("did not retry after a checkpoint: %d calls", len(f.bodies))
	}
}

// TestCompressorIsShownNothingNewerThanTheSpan is the isolation rule: the
// compressor summarises the range it was given, so it must not be able to
// see past the end of it.
func TestCompressorIsShownNothingNewerThanTheSpan(t *testing.T) {
	f := newFakeCompressor(t)
	cfg := common.DefaultBandConfig()
	cfg.Conversation.Budget = 1000 // High() is 2*Budget, so 2000
	e, _, _ := compactEngine(t, f, cfg)

	for i := 0; i < 4; i++ {
		say(e, 500)
	}
	checkpoint(e)
	e.Record(common.Event{
		Type: common.MessageReceived,
		Message: &common.MessageData{
			Actor: common.ActorHuman,
			Parts: common.PartList{common.TextPart{Text: "FUTURE-SECRET"}},
		},
	})

	e.maybeCompact()

	if len(f.bodies) == 0 {
		t.Fatal("no compaction happened")
	}
	if strings.Contains(f.bodies[0], "FUTURE-SECRET") {
		t.Fatal("the compressor was shown work newer than the span it was asked to summarise")
	}
}
