package llm

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/waywardgeek/ensemble/agent/internal/common"
)

func bandFixture(t *testing.T) (*Engine, string) {
	t.Helper()
	dir := t.TempDir()
	write := func(rel, text string) {
		t.Helper()
		p := filepath.Join(dir, rel)
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(text), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write("SOUL.md", "I would rather hold uncertainty honestly.")
	write("MEMORY.md", "Bill owns the repo and pushes.")
	write("memory/2026-09-26-1.md", "I found six dead settings fields.")
	write("memory/2026-09-26-2.md", "BlobPart never resolves to bytes.")
	write("memory/bucket-0/2026-09-01-1_2026-09-20-3.md", "September, compressed once.")

	e := &Engine{
		Log:    common.NewLog(),
		Ctx:    common.NewContext(),
		Memory: NewStore(dir),
	}
	return e, dir
}

// Startup puts memory in the context, and it arrives as data rather than as
// instruction: one labelled entry per file, from the system actor.
func TestSyncBandsLoadsMemoryAtStartup(t *testing.T) {
	e, _ := bandFixture(t)
	if err := e.SyncBands("startup"); err != nil {
		t.Fatalf("sync: %v", err)
	}
	for _, b := range []common.Band{common.BandSoul, common.BandMemory, common.Band8x} {
		if got := len(e.Ctx.BandEntries(b)); got != 1 {
			t.Errorf("band %s holds %d entries, want 1", b, got)
		}
	}
	if got := len(e.Ctx.BandEntries(common.BandSession)); got != 2 {
		t.Errorf("session band holds %d entries, want 2", got)
	}
	// Bands sit ahead of the conversation, coarsest first.
	var order []common.Band
	for _, entry := range e.Ctx.Dialogue {
		if b := common.BandForKind(entry.Kind); b != 0 {
			order = append(order, b)
		}
	}
	want := []common.Band{common.BandSoul, common.BandMemory, common.Band8x, common.BandSession, common.BandSession}
	if len(order) != len(want) {
		t.Fatalf("band order %v, want %v", order, want)
	}
	for i := range want {
		if order[i] != want[i] {
			t.Fatalf("band order %v, want %v", order, want)
		}
	}
}

// Syncing twice must say what syncing once said. Otherwise every settings
// change quietly duplicates the agent's memory.
func TestSyncBandsIsIdempotent(t *testing.T) {
	e, _ := bandFixture(t)
	if err := e.SyncBands("startup"); err != nil {
		t.Fatalf("sync: %v", err)
	}
	before := len(e.Ctx.Dialogue)
	for i := 0; i < 3; i++ {
		if err := e.SyncBands("startup"); err != nil {
			t.Fatalf("resync: %v", err)
		}
	}
	if after := len(e.Ctx.Dialogue); after != before {
		t.Errorf("syncing four times produced %d entries, want %d", after, before)
	}
}

// Switching a band off and back on restores it exactly, and a hand edit made
// while it was off comes back with it. The second half is a deliberate
// divergence from strict replay-determinism: the files on disk are the truth
// about what memory is, so editing one and switching the band on tells the
// agent the edited thing.
func TestSyncBandsDisableRestoreReflectsDisk(t *testing.T) {
	e, dir := bandFixture(t)
	on := common.DefaultBandConfig()
	off := common.DefaultBandConfig()
	off.Session.Disabled = true
	cfg := on
	e.Bands = func() common.BandConfig { return cfg }

	if err := e.SyncBands("startup"); err != nil {
		t.Fatalf("sync: %v", err)
	}
	snapshot := mustMarshal(t, e.Ctx)

	cfg = off
	if err := e.SyncBands("settings"); err != nil {
		t.Fatalf("disable: %v", err)
	}
	if got := len(e.Ctx.BandEntries(common.BandSession)); got != 0 {
		t.Fatalf("session band still holds %d entries after being switched off", got)
	}

	cfg = on
	if err := e.SyncBands("restore"); err != nil {
		t.Fatalf("restore: %v", err)
	}
	if got := mustMarshal(t, e.Ctx); got != snapshot {
		t.Errorf("restoring a band did not reproduce the context.\nbefore:\n%s\nafter:\n%s", snapshot, got)
	}

	// Now edit a file while the band is off, and restore again.
	cfg = off
	if err := e.SyncBands("settings"); err != nil {
		t.Fatalf("disable again: %v", err)
	}
	edited := "I found six dead settings fields, and Bill deleted them."
	if err := os.WriteFile(filepath.Join(dir, "memory", "2026-09-26-1.md"), []byte(edited), 0o644); err != nil {
		t.Fatal(err)
	}
	cfg = on
	if err := e.SyncBands("restore"); err != nil {
		t.Fatalf("restore after edit: %v", err)
	}
	found := false
	for _, entry := range e.Ctx.BandEntries(common.BandSession) {
		for _, p := range entry.Parts {
			if tp, ok := p.(common.TextPart); ok && contains(tp.Text, "Bill deleted them") {
				found = true
			}
		}
	}
	if !found {
		t.Error("restoring a band did not reflect an edit made to its file while it was off")
	}
}

// A populate event must carry its bytes, so that replay needs no disk. If the
// event were a reference, deleting the file would change the past.
func TestBandEventsCarryTheirBytes(t *testing.T) {
	e, dir := bandFixture(t)
	if err := e.SyncBands("startup"); err != nil {
		t.Fatalf("sync: %v", err)
	}
	if err := os.RemoveAll(filepath.Join(dir, "memory")); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(filepath.Join(dir, "SOUL.md")); err != nil {
		t.Fatal(err)
	}

	// Replay the log into a fresh context with the files gone.
	replayed := common.NewContext()
	for _, ev := range e.Log.Events {
		if err := replayed.Apply(ev); err != nil {
			t.Fatalf("replay: %v", err)
		}
	}
	if mustMarshal(t, replayed) != mustMarshal(t, e.Ctx) {
		t.Error("replaying the log without the memory files produced a different context")
	}
	if replayed.BandBytes(common.BandSession) == 0 {
		t.Error("session memory vanished on replay: the events did not carry their bytes")
	}
}

func contains(haystack, needle string) bool {
	return strings.Contains(haystack, needle)
}

func mustMarshal(t *testing.T, c *common.Context) string {
	t.Helper()
	b, err := json.Marshal(c)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	return string(b)
}
