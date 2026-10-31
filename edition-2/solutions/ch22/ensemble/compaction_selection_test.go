package ensemble_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"ensemble/ensemble"
	"ensemble/internal/common"
)

// TestCompactionConversationSelection separates a selected completed segment
// from later dialogue and independent survivors that may change while it runs.
func TestCompactionConversationSelection(t *testing.T) {
	a, cfg := compactionWindowAgent(t)
	cfg.Memory.Conversation.Budget = 1
	if err := a.SetMemorySettings(cfg.Memory); err != nil {
		t.Fatal(err)
	}
	appendCompactionFact(t, a, common.Event{Type: "message_received", Message: &common.MessageData{Actor: common.Human, Parts: []common.Part{{Type: "text", Text: "original decision and its reason"}}}})
	appendCompactionFact(t, a, common.Event{Type: "micro_handoff", Handoff: &common.MicroHandoffData{Text: "completed source"}})
	task, err := a.Memory().Conversation()
	if err != nil || task == nil {
		t.Fatalf("selection: %+v %v", task, err)
	}
	if err := a.Memory().ValidateSelection(task); err != nil {
		t.Fatal(err)
	}
	appendCompactionFact(t, a, common.Event{Type: "message_received", Message: &common.MessageData{Actor: common.Human, Parts: []common.Part{{Type: "text", Text: "new unrelated dialogue"}}}})
	appendCompactionFact(t, a, common.Event{Type: "skill_loaded", Skill: &common.SkillData{Name: "independent", Body: "new skill instructions"}})
	if err := a.Memory().ValidateSelection(task); err != nil {
		t.Fatalf("unrelated append invalidated selection: %v", err)
	}
	appendCompactionFact(t, a, common.Event{Type: "redacted", Redact: &common.RedactData{From: task.From, To: task.To, Level: "redact_dialogue", Reason: "test changed selected parts"}})
	if err := a.Memory().ValidateSelection(task); err == nil {
		t.Fatal("changed selected conversation accepted")
	}
}

func selectionCorpus(t *testing.T, a ensemble.Agent, band common.Band, dir string, count int) {
	t.Helper()
	if err := os.MkdirAll(filepath.Join(a.Config().DataDir, dir), 0700); err != nil {
		t.Fatal(err)
	}
	for i := 1; i <= count; i++ {
		name := filepath.Join(a.Config().DataDir, dir, fmt.Sprintf("2026-10-01-%d.md", i))
		if err := os.WriteFile(name, []byte(fmt.Sprintf("source-%d ten bytes", i)), 0600); err != nil {
			t.Fatal(err)
		}
	}
	files, err := a.Memory().Scan(band)
	if err != nil {
		t.Fatal(err)
	}
	for _, f := range files {
		f := f
		appendCompactionFact(t, a, common.Event{Type: "band_populated", Populated: &f})
	}
}

// TestCompactionFileSelection uses real corpus edits and policy changes. The
// unchanged prefix remains valid when a later unselected file is edited.
func TestCompactionFileSelection(t *testing.T) {
	for _, change := range []string{"unchanged", "selected-file", "unselected-file", "destination", "existing-destination", "policy", "projection"} {
		t.Run(change, func(t *testing.T) {
			a, cfg := compactionWindowAgent(t)
			cfg.Memory.Session = ensemble.BandSettings{Budget: 20, HighWatermark: 35, LowWatermark: 20}
			if err := a.SetMemorySettings(cfg.Memory); err != nil {
				t.Fatal(err)
			}
			selectionCorpus(t, a, common.BandSession, "memory", 3)
			if change == "existing-destination" {
				dir := filepath.Join(cfg.DataDir, "bucket-0")
				if err := os.MkdirAll(dir, 0700); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(filepath.Join(dir, "2026-10-01-1_2026-10-01-2.md"), []byte("previous result"), 0600); err != nil {
					t.Fatal(err)
				}
			}
			task, err := a.Memory().Graduation()
			if change == "existing-destination" {
				if err == nil || task != nil {
					t.Fatal("preexisting destination selected for overwrite")
				}
				return
			}
			if err != nil || task == nil {
				t.Fatalf("select: %+v %v", task, err)
			}
			if len(task.Retired) != 2 || len(task.Selection.Sources) != 2 {
				t.Fatal("selection did not capture exact prefix")
			}
			write := func(name, text string) {
				t.Helper()
				if err := os.MkdirAll(filepath.Dir(name), 0700); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(name, []byte(text), 0600); err != nil {
					t.Fatal(err)
				}
			}
			switch change {
			case "selected-file":
				write(filepath.Join(cfg.DataDir, "memory", "2026-10-01-1.md"), "edited-1 ten bytes")
			case "unselected-file":
				write(filepath.Join(cfg.DataDir, "memory", "2026-10-01-3.md"), "edited-3 ten bytes")
			case "destination":
				write(filepath.Join(cfg.DataDir, task.Selection.Destination.Path), "human replacement")
			case "policy":
				cfg.Memory.Session.Budget++
				if err := a.SetMemorySettings(cfg.Memory); err != nil {
					t.Fatal(err)
				}
			case "projection":
				appendCompactionFact(t, a, common.Event{Type: "band_populated", Populated: &common.BandPopulatedData{Band: common.BandSession, File: task.File, Text: "different context bytes"}})
			}
			err = a.Memory().ValidateSelection(task)
			valid := change == "unchanged" || change == "unselected-file"
			if valid && err != nil {
				t.Fatalf("unchanged selected prefix refused: %v", err)
			}
			if !valid && err == nil {
				t.Fatal("changed selected source/destination/policy accepted")
			}
		})
	}
}

// TestCompactionCuratedSelection protects the existing terminal document both
// before launch and after selection; matching file/context bytes are the control.
func TestCompactionCuratedSelection(t *testing.T) {
	for _, when := range []string{"unchanged", "before", "after"} {
		t.Run(when, func(t *testing.T) {
			a, cfg := compactionWindowAgent(t)
			cfg.Memory.Tier64 = ensemble.BandSettings{Budget: 1, HighWatermark: 1, LowWatermark: 1}
			if err := a.SetMemorySettings(cfg.Memory); err != nil {
				t.Fatal(err)
			}
			name := filepath.Join(cfg.DataDir, "MEMORY.md")
			if err := os.WriteFile(name, []byte("old curated memory"), 0600); err != nil {
				t.Fatal(err)
			}
			appendCompactionFact(t, a, common.Event{Type: "band_populated", Populated: &common.BandPopulatedData{Band: common.BandMemory, Text: "old curated memory"}})
			selectionCorpus(t, a, common.Band64x, "bucket-1", 2)
			edit := func() {
				if err := os.WriteFile(name, []byte("new curated memory"), 0600); err != nil {
					t.Fatal(err)
				}
			}
			if when == "before" {
				edit()
			}
			task, err := a.Memory().Graduation()
			if when == "before" {
				if err == nil || task != nil {
					t.Fatal("stale curated source selected")
				}
				return
			}
			if err != nil || task == nil {
				t.Fatalf("select: %+v %v", task, err)
			}
			if when == "after" {
				edit()
			}
			err = a.Memory().ValidateSelection(task)
			if when == "after" && err == nil {
				t.Fatal("curated memory edit accepted")
			}
			if when == "unchanged" && err != nil {
				t.Fatal(err)
			}
		})
	}
}
