package ensemble_test

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"ensemble/ensemble"
	"ensemble/internal/common"
)

// These tests now exercise the staged protocol. The old direct-Write failures
// remain in the recorded baseline source/output; they are not unchanged tests.
func stagedLaunch(t *testing.T, a ensemble.Agent, task *common.Compression) common.CompactorLaunchedData {
	t.Helper()
	if err := a.Save(); err != nil {
		t.Fatal(err)
	}
	launch := common.CompactorLaunchedData{TaskID: "staged-task", SelectionHash: task.SelectionHash, Source: task.Source, Band: task.Destination, File: task.File, Thru: task.Thru, From: task.From, To: task.To, TracePath: "children/staged-task"}
	appendCompactionFact(t, a, common.Event{Type: "compactor_launched", Compactor: &launch})
	return launch
}
func loadStagedAgent(t *testing.T, cfg ensemble.Config) ensemble.Agent {
	t.Helper()
	b, err := ensemble.New(io.Discard).NewAgent(cfg)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { b.Shutdown() })
	if err := b.Load(); err != nil {
		t.Fatal(err)
	}
	return b
}
func stageName(cfg ensemble.Config) string {
	return filepath.Join(cfg.DataDir, ".compaction", "pending.json")
}
func commitStaged(t *testing.T, a ensemble.Agent, task *common.Compression, launch common.CompactorLaunchedData, pop *common.BandPopulatedData) {
	t.Helper()
	if err := a.Memory().ValidateSelection(task); err != nil {
		t.Fatal(err)
	}
	if task.Source == 0 {
		appendCompactionFact(t, a, receiptCut(launch))
	} else {
		appendCompactionFact(t, a, common.Event{Type: "band_depopulated", Depopulated: &common.BandDepopulatedData{TaskID: launch.TaskID, Band: task.Source, Thru: &task.Thru, Reason: "graduation"}})
	}
	appendCompactionFact(t, a, common.Event{Type: "band_populated", Populated: pop})
}
func stagedGraduation(t *testing.T) (ensemble.Agent, ensemble.Config, *common.Compression, common.CompactorLaunchedData, *common.BandPopulatedData) {
	t.Helper()
	a, cfg := compactionWindowAgent(t)
	selectionCorpus(t, a, common.BandSession, "memory", 2)
	task, err := a.Memory().Graduation()
	if err != nil || task == nil {
		t.Fatalf("select: %+v %v", task, err)
	}
	launch := stagedLaunch(t, a, task)
	pop, err := a.Memory().Stage(launch.TaskID, task, "COMMITTED REPLACEMENT")
	if err != nil {
		t.Fatal(err)
	}
	return a, cfg, task, launch, pop
}

// TestUncommittedDiskReplacement now proves the hidden staged candidate cannot
// enter active memory on restart, before or after the held cut was journaled.
func TestUncommittedDiskReplacement(t *testing.T) {
	for _, cut := range []bool{false, true} {
		t.Run(map[bool]string{false: "staged", true: "held-cut"}[cut], func(t *testing.T) {
			a, cfg := compactionWindowAgent(t)
			cfg.Memory.Conversation.Budget = 1
			if err := a.SetMemorySettings(cfg.Memory); err != nil {
				t.Fatal(err)
			}
			appendCompactionFact(t, a, common.Event{Type: "message_received", Message: &common.MessageData{Actor: common.Human, Parts: []common.Part{{Type: "text", Text: "IRREPLACEABLE SOURCE"}}}})
			appendCompactionFact(t, a, common.Event{Type: "micro_handoff", Handoff: &common.MicroHandoffData{Text: "completed segment"}})
			task, err := a.Memory().Conversation()
			if err != nil || task == nil {
				t.Fatalf("select: %+v %v", task, err)
			}
			launch := stagedLaunch(t, a, task)
			if _, err := a.Memory().Stage(launch.TaskID, task, "UNCOMMITTED CANDIDATE"); err != nil {
				t.Fatal(err)
			}
			if files, err := a.Memory().Scan(common.BandSession); err != nil || len(files) != 0 {
				t.Fatalf("hidden candidate scanned: %+v %v", files, err)
			}
			if cut {
				appendCompactionFact(t, a, receiptCut(launch))
			}
			b := loadStagedAgent(t, cfg)
			if _, err := b.Ask(submissionContext(t), "ordinary restart; no checkpoint"); err != nil {
				t.Fatal(err)
			}
			source, candidate := false, false
			for _, e := range b.History().Context().Dialogue {
				for _, p := range e.Parts {
					source = source || strings.Contains(p.Text, "IRREPLACEABLE SOURCE")
					candidate = candidate || strings.Contains(p.Text, "UNCOMMITTED CANDIDATE")
				}
			}
			if !source || candidate {
				t.Fatalf("uncommitted recovery source=%t candidate=%t", source, candidate)
			}
			out := b.History().Context().CompactionOutcome
			if out == nil || out.TaskID != launch.TaskID || out.State != "abandoned" {
				t.Fatalf("missing abandonment receipt: %+v", out)
			}
			if _, err := os.Stat(stageName(cfg)); !os.IsNotExist(err) {
				t.Fatalf("staging not cleaned: %v", err)
			}
		})
	}
}

// TestRetiredSourceBeforeArchive exercises the new staged/journal protocol at
// each durable boundary. It models saved states, not an actual process kill.
func TestRetiredSourceBeforeArchive(t *testing.T) {
	for _, boundary := range []string{"pair", "promoted", "one-archive", "finished", "terminal"} {
		t.Run(boundary, func(t *testing.T) {
			a, cfg, task, launch, pop := stagedGraduation(t)
			commitStaged(t, a, task, launch, pop)
			if boundary == "promoted" || boundary == "one-archive" {
				// Reproduce the exact disk state after destination promotion, before
				// Finish has renamed every selected source. History already committed.
				dest := filepath.Join(cfg.DataDir, task.Selection.Destination.Path)
				if err := os.MkdirAll(filepath.Dir(dest), 0700); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(dest, []byte(pop.Text), 0600); err != nil {
					t.Fatal(err)
				}
				if boundary == "one-archive" {
					name := filepath.Join(cfg.DataDir, task.Selection.Sources[0].Path)
					if err := os.Rename(name, name+".retired"); err != nil {
						t.Fatal(err)
					}
				}
			}
			if boundary == "finished" || boundary == "terminal" {
				if err := a.Memory().Finish(launch.TaskID); err != nil {
					t.Fatal(err)
				}
			}
			if boundary == "terminal" {
				appendCompactionFact(t, a, common.Event{Type: "compaction_outcome", CompactionOutcome: &common.CompactionOutcomeData{TaskID: launch.TaskID, State: "committed", TracePath: launch.TracePath}})
			}
			b := loadStagedAgent(t, cfg)
			if _, err := b.Ask(submissionContext(t), "ordinary restart; no checkpoint"); err != nil {
				t.Fatal(err)
			}
			c := b.History().Context()
			if countKind(c, common.KindSession) != 0 || countKind(c, common.Kind8x) != 1 || c.Compression != nil || c.CompactionOutcome == nil || c.CompactionOutcome.State != "committed" {
				t.Fatal("restart lost receipt or reintroduced retired source")
			}
			for _, source := range task.Selection.Sources {
				if _, err := os.Stat(filepath.Join(cfg.DataDir, source.Path)); !os.IsNotExist(err) {
					t.Fatalf("source still active: %v", err)
				}
				if _, err := os.Stat(filepath.Join(cfg.DataDir, source.Path+".retired")); err != nil {
					t.Fatal(err)
				}
			}
			if _, err := os.Stat(stageName(cfg)); !os.IsNotExist(err) {
				t.Fatalf("staging not cleaned: %v", err)
			}
			if out, err := b.Memory().Recover(); err != nil || out != nil {
				t.Fatalf("repeated recovery changed outcome: %+v %v", out, err)
			}
		})
	}
}

// TestCompactionRecoveryHold preserves mismatched files and receipts, blocks
// every ordinary corpus/render entry point, and retries after concrete repair.
func TestCompactionRecoveryHold(t *testing.T) {
	for _, failure := range []string{"missing-stage", "malformed-stage", "invalid-stage", "unknown-stage", "selection-hash", "candidate-hash", "task-id", "receipt-text", "destination-edit", "source-edit", "archive-conflict"} {
		t.Run(failure, func(t *testing.T) {
			a, cfg, task, launch, pop := stagedGraduation(t)
			commitStaged(t, a, task, launch, pop)
			stage, err := os.ReadFile(stageName(cfg))
			if err != nil {
				t.Fatal(err)
			}
			source := filepath.Join(cfg.DataDir, task.Selection.Sources[0].Path)
			original, err := os.ReadFile(source)
			if err != nil {
				t.Fatal(err)
			}
			destination := filepath.Join(cfg.DataDir, task.Selection.Destination.Path)
			switch failure {
			case "missing-stage":
				if err := os.Remove(stageName(cfg)); err != nil {
					t.Fatal(err)
				}
			case "malformed-stage":
				if err := os.WriteFile(stageName(cfg), []byte(`{"version":`), 0600); err != nil {
					t.Fatal(err)
				}
			case "invalid-stage", "unknown-stage", "selection-hash", "candidate-hash", "task-id", "receipt-text":
				var record map[string]any
				if err := json.Unmarshal(stage, &record); err != nil {
					t.Fatal(err)
				}
				switch failure {
				case "invalid-stage":
					record["version"] = 999
				case "unknown-stage":
					record["unrecognized"] = true
				case "selection-hash":
					record["selection"].(map[string]any)["entries_hash"] = "changed"
				case "task-id":
					record["launch"].(map[string]any)["task_id"] = "foreign"
				case "candidate-hash":
					record["candidate_hash"] = "wrong digest"
				case "receipt-text":
					record["population"].(map[string]any)["text"] = "different memory"
					record["candidate_hash"] = fmt.Sprintf("%x", sha256.Sum256([]byte("different memory")))
				}
				corrupt, err := json.Marshal(record)
				if err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(stageName(cfg), corrupt, 0600); err != nil {
					t.Fatal(err)
				}
			case "destination-edit":
				if err := os.MkdirAll(filepath.Dir(destination), 0700); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(destination, []byte("HUMAN DESTINATION"), 0600); err != nil {
					t.Fatal(err)
				}
			case "source-edit":
				if err := os.WriteFile(source, []byte("HUMAN SOURCE"), 0600); err != nil {
					t.Fatal(err)
				}
			case "archive-conflict":
				if err := os.WriteFile(source+".retired", []byte("UNRELATED ARCHIVE"), 0600); err != nil {
					t.Fatal(err)
				}
			}

			b := loadStagedAgent(t, cfg)
			if _, err := b.Ask(submissionContext(t), "original source memory"); err == nil {
				t.Fatal("unfinalized commit reached main model")
			}
			if pending := b.History().Context().Compression; pending == nil || pending.Committed == nil {
				t.Fatal("failure discarded commit receipt")
			}
			if _, err := b.Memory().Scan(common.BandSession); err == nil {
				t.Fatal("active scan bypassed recovery hold")
			}
			if _, err := b.Memory().Graduation(); err == nil {
				t.Fatal("new selection bypassed recovery hold")
			}
			if _, err := b.Engine().Render(b.History().Context()); err == nil {
				t.Fatal("render bypassed recovery hold")
			}
			if err := b.SetMemorySettings(cfg.Memory); err == nil {
				t.Fatal("re-enable bypassed recovery failure")
			}
			for _, e := range b.History().EventsAfter(0) {
				if e.Type == "request_sent" {
					t.Fatal("recovery failure sent model request")
				}
			}
			// The explicit repair restores only the conflicting test artifact.
			switch failure {
			case "missing-stage", "malformed-stage", "invalid-stage", "unknown-stage", "selection-hash", "candidate-hash", "task-id", "receipt-text":
				err = os.WriteFile(stageName(cfg), stage, 0600)
			case "destination-edit":
				data, _ := os.ReadFile(destination)
				if string(data) != "HUMAN DESTINATION" {
					t.Fatal("human destination overwritten")
				}
				err = os.Remove(destination)
			case "source-edit":
				data, _ := os.ReadFile(source)
				if string(data) != "HUMAN SOURCE" {
					t.Fatal("human source overwritten")
				}
				err = os.WriteFile(source, original, 0600)
			case "archive-conflict":
				data, _ := os.ReadFile(source + ".retired")
				if string(data) != "UNRELATED ARCHIVE" {
					t.Fatal("archive overwritten")
				}
				err = os.Remove(source + ".retired")
			}
			if err != nil {
				t.Fatal(err)
			}
			if _, err := b.Ask(submissionContext(t), "continue after repair"); err != nil {
				t.Fatal(err)
			}
			if b.History().Context().Compression != nil {
				t.Fatal("successful repair retained recovery hold")
			}
		})
	}
}

// TestCompactionStageRefusal proves invalid output and a real staging I/O
// refusal leave original context and active files untouched.
func TestCompactionStageRefusal(t *testing.T) {
	for _, failure := range []string{"empty", "task", "io"} {
		t.Run(failure, func(t *testing.T) {
			a, _ := compactionWindowAgent(t)
			selectionCorpus(t, a, common.BandSession, "memory", 2)
			task, err := a.Memory().Graduation()
			if err != nil || task == nil {
				t.Fatalf("select: %+v %v", task, err)
			}
			launch := stagedLaunch(t, a, task)
			id, text := launch.TaskID, "candidate"
			if failure == "empty" {
				text = " \n "
			}
			if failure == "task" {
				id = "stale-task"
			}
			if failure == "io" {
				if err := os.WriteFile(filepath.Dir(stageName(a.Config())), []byte("not a directory"), 0600); err != nil {
					t.Fatal(err)
				}
			}
			if _, err := a.Memory().Stage(id, task, text); err == nil {
				t.Fatal("invalid/stale/unwritable candidate staged")
			}
			if countKind(a.History().Context(), common.KindSession) != 2 {
				t.Fatal("failed staging retired source")
			}
			if _, err := os.Stat(filepath.Join(a.Config().DataDir, task.Selection.Destination.Path)); !os.IsNotExist(err) {
				t.Fatalf("failed staging exposed candidate: %v", err)
			}
		})
	}
}

// TestCompactionCuratedRecovery covers replacement of an existing terminal
// document, including a human edit after journal commit but before promotion.
func TestCompactionCuratedRecovery(t *testing.T) {
	for _, conflict := range []bool{false, true} {
		t.Run(map[bool]string{false: "replace", true: "human-edit"}[conflict], func(t *testing.T) {
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
			task, err := a.Memory().Graduation()
			if err != nil || task == nil || task.Destination != common.BandMemory {
				t.Fatalf("curated selection: %+v %v", task, err)
			}
			launch := stagedLaunch(t, a, task)
			pop, err := a.Memory().Stage(launch.TaskID, task, "exact curated replacement")
			if err != nil {
				t.Fatal(err)
			}
			commitStaged(t, a, task, launch, pop)
			if conflict {
				if err := os.WriteFile(name, []byte("HUMAN CURATED EDIT"), 0600); err != nil {
					t.Fatal(err)
				}
			}
			b := loadStagedAgent(t, cfg)
			_, err = b.Ask(submissionContext(t), "continue")
			data, readErr := os.ReadFile(name)
			if readErr != nil {
				t.Fatal(readErr)
			}
			if conflict {
				if err == nil || string(data) != "HUMAN CURATED EDIT" || b.History().Context().Compression == nil {
					t.Fatalf("curated conflict lost: bytes=%q error=%v", data, err)
				}
				return
			}
			if err != nil || string(data) != pop.Text || b.History().Context().Compression != nil {
				t.Fatalf("curated recovery: bytes=%q error=%v", data, err)
			}
			files, err := b.Memory().Scan(common.BandMemory)
			if err != nil || len(files) != 1 || files[0].Text != pop.Text || countKind(b.History().Context(), common.Kind64x) != 0 {
				t.Fatalf("curated corpus differs from commit: %+v %v", files, err)
			}
		})
	}
}
