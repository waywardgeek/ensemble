package ensemble_test

import (
	"ensemble/internal/common"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

// TestMemoryGraduationCoverage checks restored entries that bypass Scan, and
// proves that a normal fold still removes exactly its selected prefix.
func TestMemoryGraduationCoverage(t *testing.T) {
	for _, sample := range []struct {
		name  string
		ends  []int
		valid bool
	}{
		{"disjoint", []int{1, 2, 3}, true},
		{"reversed", []int{1, 1, 3}, false},
		{"overlapping", []int{3, 3, 3}, false},
	} {
		t.Run(sample.name, func(t *testing.T) {
			a, cfg := compactionWindowAgent(t)
			settings := cfg.Memory
			settings.Session.HighWatermark = 35
			settings.Session.LowWatermark = 20
			if err := a.SetMemorySettings(settings); err != nil {
				t.Fatal(err)
			}
			for i, end := range sample.ends {
				id := common.MemoryFileID{Date: "2026-10-01", Num: i + 1}
				thru := common.MemoryFileID{Date: id.Date, Num: end}
				text := fmt.Sprintf("source-%d ten bytes", i)
				if sample.valid {
					// Selection now fingerprints actual corpus bytes as well as replayed
					// entries. Invalid ranges still fail before filesystem selection.
					dir := filepath.Join(cfg.DataDir, "memory")
					if err := os.MkdirAll(dir, 0700); err != nil {
						t.Fatal(err)
					}
					if err := os.WriteFile(filepath.Join(dir, fmt.Sprintf("2026-10-01-%d.md", i+1)), []byte(text), 0600); err != nil {
						t.Fatal(err)
					}
				}
				appendCompactionFact(t, a, common.Event{Type: "band_populated", Populated: &common.BandPopulatedData{Band: common.BandSession, File: id, Thru: thru, Text: text}})
			}
			task, err := a.Memory().Graduation()
			if !sample.valid {
				if err == nil || task != nil {
					t.Fatalf("invalid restored coverage selected: task=%+v err=%v", task, err)
				}
				if countKind(a.History().Context(), common.KindSession) != 3 {
					t.Fatal("refusal altered source")
				}
				return
			}
			if err != nil || task == nil || len(task.Retired) != 2 {
				t.Fatalf("valid fold: task=%+v err=%v", task, err)
			}
			if task.Retired[0].File.Num != 1 || task.Retired[1].File.Num != 2 || task.File.Num != 1 || task.Thru.Num != 2 {
				t.Fatalf("valid fold selected wrong prefix: %+v", task)
			}
			appendCompactionFact(t, a, common.Event{Type: "compactor_launched", Compactor: &common.CompactorLaunchedData{Band: task.Destination, Thru: task.Thru}})
			appendCompactionFact(t, a, common.Event{Type: "band_depopulated", Depopulated: &common.BandDepopulatedData{Band: task.Source, Thru: &task.Thru, Reason: "graduation"}})
			appendCompactionFact(t, a, common.Event{Type: "band_populated", Populated: &common.BandPopulatedData{Band: task.Destination, File: task.File, Thru: task.Thru, Text: "folded", Source: "compression"}})
			if countKind(a.History().Context(), common.KindSession) != 1 {
				t.Fatal("valid fold did not retire exact prefix")
			}
			for _, entry := range a.History().Context().Dialogue {
				if entry.Kind == common.KindSession && entry.File.Num != 3 {
					t.Fatalf("valid fold retained wrong source: %+v", entry.File)
				}
			}
			if next, err := a.Memory().Graduation(); err != nil || next != nil {
				t.Fatalf("valid fold did not progress: %+v %v", next, err)
			}
		})
	}
}

// TestMemoryScanMalformedCoverage pairs real valid ranged files with refused
// reversed/extra-segment names, so refusing every corpus file cannot pass.
func TestMemoryScanMalformedCoverage(t *testing.T) {
	for _, name := range []string{"2026-10-01-1_2026-10-01-2.md", "2026-10-01-3_2026-10-01-1.md", "2026-10-01-1_2026-10-01-2_extra.md"} {
		t.Run(name, func(t *testing.T) {
			a, cfg := compactionWindowAgent(t)
			dir := filepath.Join(cfg.DataDir, "memory")
			if err := os.MkdirAll(dir, 0700); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(dir, name), []byte("source"), 0600); err != nil {
				t.Fatal(err)
			}
			files, err := a.Memory().Scan(common.BandSession)
			if name == "2026-10-01-1_2026-10-01-2.md" {
				if err != nil || len(files) != 1 || files[0].File.Num != 1 || files[0].Thru.Num != 2 {
					t.Fatalf("valid coverage scan: %+v %v", files, err)
				}
				return
			}
			if err == nil || len(files) != 0 {
				t.Fatalf("malformed range scanned: %+v %v", files, err)
			}
		})
	}
}

// TestReversedRangesCannotRepeatFold reproduces the baseline non-progress loop:
// a committed fold with a reversed endpoint retires nothing and selects it again.
func TestReversedRangesCannotRepeatFold(t *testing.T) {
	a, cfg := compactionWindowAgent(t)
	dir := filepath.Join(cfg.DataDir, "memory")
	if err := os.MkdirAll(dir, 0700); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"2026-10-01-2_2026-10-01-1.md", "2026-10-01-3_2026-10-01-1.md"} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte("completed source "+name), 0600); err != nil {
			t.Fatal(err)
		}
	}
	files, err := a.Memory().Scan(common.BandSession)
	if err != nil || len(files) == 0 {
		return
	} // Rejecting/skipping malformed ranges is safe.
	for _, f := range files {
		f := f
		appendCompactionFact(t, a, common.Event{Type: "band_populated", Populated: &f})
	}
	task, err := a.Memory().Graduation()
	if err != nil || task == nil {
		return
	}
	appendCompactionFact(t, a, common.Event{Type: "compactor_launched", Compactor: &common.CompactorLaunchedData{Band: task.Destination, Thru: task.Thru}})
	if err = a.Memory().Write(task, "REPLACEMENT"); err != nil {
		t.Fatal(err)
	}
	appendCompactionFact(t, a, common.Event{Type: "band_depopulated", Depopulated: &common.BandDepopulatedData{Band: task.Source, Thru: &task.Thru, Reason: "graduation"}})
	appendCompactionFact(t, a, common.Event{Type: "band_populated", Populated: &common.BandPopulatedData{Band: task.Destination, File: task.File, Thru: task.Thru, Text: "REPLACEMENT", Source: "compression"}})
	if err = a.Memory().Archive(task); err != nil {
		t.Fatal(err)
	}
	next, err := a.Memory().Graduation()
	if err != nil {
		t.Fatal(err)
	}
	if next != nil && next.Source == task.Source && next.Text == task.Text {
		t.Fatal("completed fold retires no entries and immediately reselects identical source")
	}
}
