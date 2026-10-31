package ensemble_test

import (
	"bytes"
	"io"
	"testing"

	"ensemble/ensemble"
	"ensemble/internal/common"
)

func receiptAgent(t *testing.T) (ensemble.Agent, ensemble.Config, common.CompactorLaunchedData) {
	t.Helper()
	a, cfg := compactionWindowAgent(t)
	appendCompactionFact(t, a, common.Event{Type: "message_received", Message: &common.MessageData{Actor: common.Human, Parts: []common.Part{{Type: "text", Text: "original source"}}}})
	from := a.History().LastSeq()
	appendCompactionFact(t, a, common.Event{Type: "micro_handoff", Handoff: &common.MicroHandoffData{Text: "completed source"}})
	id := common.MemoryFileID{Date: "2026-10-01", Num: 1}
	return a, cfg, common.CompactorLaunchedData{TaskID: "compact-one", SelectionHash: "captured-selection", Band: common.BandSession, File: id, Thru: id, From: from, To: a.History().LastSeq()}
}
func receiptCut(launch common.CompactorLaunchedData) common.Event {
	return common.Event{Type: "redacted", Redact: &common.RedactData{TaskID: launch.TaskID, From: launch.From, To: launch.To, Level: "redact_summary", Reason: "memory compression"}}
}
func receiptPopulation(launch common.CompactorLaunchedData) common.Event {
	return common.Event{Type: "band_populated", Populated: &common.BandPopulatedData{TaskID: launch.TaskID, Band: launch.Band, File: launch.File, Thru: launch.Thru, Text: "committed memory", Source: "compression"}}
}

// TestCompactionCommitReceipt proves a correlated pair applies once, retains
// its disk-finalization receipt across replay, and cannot be abandoned afterward.
func TestCompactionCommitReceipt(t *testing.T) {
	a, cfg, launch := receiptAgent(t)
	if err := a.Save(); err != nil {
		t.Fatal(err)
	}
	appendCompactionFact(t, a, common.Event{Type: "compactor_launched", Compactor: &launch})
	appendCompactionFact(t, a, receiptCut(launch))
	if countKind(a.History().Context(), common.KindDialogue) != 1 {
		t.Fatal("cut applied before replacement")
	}
	wrong := receiptPopulation(launch)
	wrong.Populated.TaskID = "stale-task"
	if err := a.History().Append(wrong); err == nil {
		t.Fatal("wrong task committed held cut")
	}
	// Even matching legacy band/range metadata cannot satisfy a managed pair.
	disk := receiptPopulation(launch)
	disk.Populated.TaskID = ""
	disk.Populated.Source = "disk"
	appendCompactionFact(t, a, disk)
	if countKind(a.History().Context(), common.KindDialogue) != 1 || a.History().Context().Compression == nil {
		t.Fatal("disk population committed managed cut")
	}
	appendCompactionFact(t, a, receiptPopulation(launch))
	c := a.History().Context()
	if countKind(c, common.KindDialogue) != 0 || countKind(c, common.KindSession) != 1 || c.Compression == nil || c.Compression.Committed == nil {
		t.Fatalf("missing applied receipt: %+v", c.Compression)
	}
	if err := a.History().Append(receiptPopulation(launch)); err == nil {
		t.Fatal("replacement applied twice")
	}
	if err := a.History().Append(common.Event{Type: "compaction_outcome", CompactionOutcome: &common.CompactionOutcomeData{TaskID: launch.TaskID, State: "abandoned", Reason: "restart"}}); err == nil {
		t.Fatal("committed source was abandoned")
	}
	legacyLaunch := common.CompactorLaunchedData{Band: common.BandSession, Thru: launch.Thru}
	if err := a.History().Append(common.Event{Type: "compactor_launched", Compactor: &legacyLaunch}); err == nil {
		t.Fatal("new legacy launch erased pending commit receipt")
	}
	// Recover the actual snapshot+journal, without any Memory or model activity.
	b, err := ensemble.New(io.Discard).NewAgent(cfg)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { b.Shutdown() })
	if err := b.Load(); err != nil {
		t.Fatal(err)
	}
	if got := b.History().Context().Compression; got == nil || got.Committed == nil || got.Committed.Text != "committed memory" {
		t.Fatalf("journal lost committed receipt: %+v", got)
	}
	before := b.History().Context().Turn
	outcome := common.Event{Type: "compaction_outcome", CompactionOutcome: &common.CompactionOutcomeData{TaskID: launch.TaskID, State: "committed"}}
	appendCompactionFact(t, b, outcome)
	appendCompactionFact(t, b, outcome)
	if c := b.History().Context(); c.Compression != nil || c.CompactionOutcome == nil || c.Turn != before {
		t.Fatal("terminal outcome lost identity or altered main turn")
	}
	var saved bytes.Buffer
	if err := b.History().Save(&saved); err != nil {
		t.Fatal(err)
	}
	if err := a.History().Restore(&saved); err != nil {
		t.Fatal(err)
	}
	if c := a.History().Context(); c.Compression != nil || c.CompactionOutcome.TaskID != launch.TaskID {
		t.Fatal("snapshot lost terminal cleanup receipt")
	}
}

// TestCompactionReceiptRefusal protects the uncommitted side of the protocol:
// mismatched cuts/populations and premature success cannot retire source.
func TestCompactionReceiptRefusal(t *testing.T) {
	for _, mode := range []string{"cut-task", "cut-range", "cut-legacy", "population-before-cut", "population-file", "premature-success", "duplicate-launch"} {
		t.Run(mode, func(t *testing.T) {
			a, _, launch := receiptAgent(t)
			appendCompactionFact(t, a, common.Event{Type: "compactor_launched", Compactor: &launch})
			e := receiptCut(launch)
			switch mode {
			case "cut-task":
				e.Redact.TaskID = "other"
			case "cut-range":
				e.Redact.From++
			case "cut-legacy":
				e.Redact.TaskID = ""
			case "population-before-cut":
				e = receiptPopulation(launch)
			case "population-file":
				appendCompactionFact(t, a, e)
				e = receiptPopulation(launch)
				e.Populated.File.Num++
			case "premature-success":
				e = common.Event{Type: "compaction_outcome", CompactionOutcome: &common.CompactionOutcomeData{TaskID: launch.TaskID, State: "committed"}}
			case "duplicate-launch":
				e = common.Event{Type: "compactor_launched", Compactor: &launch}
			}
			if err := a.History().Append(e); err == nil {
				t.Fatal("mismatched or premature compaction fact accepted")
			}
			if countKind(a.History().Context(), common.KindDialogue) != 1 {
				t.Fatal("refused fact changed source")
			}
			appendCompactionFact(t, a, common.Event{Type: "compaction_outcome", CompactionOutcome: &common.CompactionOutcomeData{TaskID: launch.TaskID, State: "abandoned", Reason: "no committed replacement"}})
			if a.History().Context().Compression != nil || countKind(a.History().Context(), common.KindDialogue) != 1 {
				t.Fatal("abandonment changed source or retained pending cut")
			}
		})
	}
}

// TestCompactionGraduationReceipt applies the correlated source-band prefix
// with its replacement, refusing a retirement naming a different band.
func TestCompactionGraduationReceipt(t *testing.T) {
	a, cfg := compactionWindowAgent(t)
	cfg.Memory.Session = ensemble.BandSettings{Budget: 20, HighWatermark: 35, LowWatermark: 20}
	if err := a.SetMemorySettings(cfg.Memory); err != nil {
		t.Fatal(err)
	}
	selectionCorpus(t, a, common.BandSession, "memory", 3)
	task, err := a.Memory().Graduation()
	if err != nil || task == nil {
		t.Fatalf("select: %+v %v", task, err)
	}
	launch := common.CompactorLaunchedData{TaskID: "fold", SelectionHash: "captured-fold", Source: task.Source, Band: task.Destination, File: task.File, Thru: task.Thru}
	appendCompactionFact(t, a, common.Event{Type: "compactor_launched", Compactor: &launch})
	wrong := common.BandDepopulatedData{TaskID: launch.TaskID, Band: common.Band8x, Thru: &task.Thru, Reason: "graduation"}
	if err := a.History().Append(common.Event{Type: "band_depopulated", Depopulated: &wrong}); err == nil {
		t.Fatal("wrong source band retired")
	}
	wrong.Band = task.Source
	appendCompactionFact(t, a, common.Event{Type: "band_depopulated", Depopulated: &wrong})
	if countKind(a.History().Context(), common.KindSession) != 3 {
		t.Fatal("graduation retired before population")
	}
	appendCompactionFact(t, a, receiptPopulation(launch))
	c := a.History().Context()
	if c.Compression == nil || c.Compression.Committed == nil || countKind(c, common.KindSession) != 1 || countKind(c, common.Kind8x) != 1 {
		t.Fatal("graduation pair/receipt incomplete")
	}
	for _, e := range c.Dialogue {
		if e.Kind == common.KindSession && e.File.Num != 3 {
			t.Fatal("graduation retired wrong prefix")
		}
	}
}
