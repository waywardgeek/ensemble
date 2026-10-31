package memory

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"ensemble/internal/common"
)

// One hidden record is sufficient because the owning actor permits one attempt.
// It is outside every active/recall scan and carries no model-authored paths.
type stagedCompaction struct {
	// Version refuses unknown record formats instead of guessing disk mutations.
	Version int `json:"version"`
	// Launch binds source/task identity to its already journaled parent decision.
	Launch common.CompactorLaunchedData `json:"launch"`
	// Selection holds prior file fingerprints, not recoverable original file bytes.
	Selection common.CompressionSelection `json:"selection"`
	// Population contains the exact replacement bytes journaled at commit.
	Population common.BandPopulatedData `json:"population"`
	// CandidateHash detects a damaged candidate before promotion.
	CandidateHash string `json:"candidate_hash"`
}

func (m *memory) stagePath() string {
	return filepath.Join(m.parent.Config().DataDir, ".compaction", "pending.json")
}

// Use the owner's confined temporary-file/rename operations. As with History's
// journal, atomic installation protects process crashes, not power-loss/fsync.
func (m *memory) atomicWrite(name string, data []byte) error {
	sandbox := m.parent.Sandbox()
	if err := sandbox.MkdirAll(filepath.Dir(name), 0700); err != nil {
		return err
	}
	f, err := sandbox.CreateTemp(filepath.Dir(name))
	if err != nil {
		return err
	}
	defer sandbox.Remove(f.Name())
	_, err = f.Write(data)
	closeErr := f.Close()
	if err == nil {
		err = closeErr
	}
	if err != nil {
		return err
	}
	return sandbox.Rename(f.Name(), name)
}

func (m *memory) readStage() (*stagedCompaction, error) {
	data, err := m.parent.Sandbox().ReadFile(m.stagePath())
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var record stagedCompaction
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&record); err != nil {
		return nil, fmt.Errorf("invalid staged compaction: %w", err)
	}
	var extra any
	if decoder.Decode(&extra) != io.EOF {
		return nil, fmt.Errorf("staged compaction has trailing data")
	}
	p, launch := record.Population, record.Launch
	selection, err := json.Marshal(record.Selection)
	if err != nil {
		return nil, err
	}
	if record.Version != 1 || launch.TaskID == "" || digest(selection) != launch.SelectionHash || p.TaskID != launch.TaskID || p.Band != launch.Band || p.File != launch.File || p.Thru != launch.Thru || p.Source != "compression" || strings.TrimSpace(p.Text) == "" || digest([]byte(p.Text)) != record.CandidateHash {
		return nil, fmt.Errorf("staged compaction identity/hash mismatch")
	}
	// A corrupted record cannot name another owner's files, even when no sandbox
	// root was configured. The destination is derived again from trusted metadata.
	destination, err := filepath.Rel(m.parent.Config().DataDir, path(m.parent, p.Band, p.File, p.Thru))
	if err != nil || !filepath.IsLocal(destination) || destination != record.Selection.Destination.Path {
		return nil, fmt.Errorf("invalid staged destination")
	}
	for _, source := range record.Selection.Sources {
		if launch.Source < common.Band64x || launch.Source > common.BandSession || !filepath.IsLocal(source.Path) || !source.Exists {
			return nil, fmt.Errorf("invalid staged source")
		}
		dir, err := filepath.Rel(m.parent.Config().DataDir, filepath.Dir(path(m.parent, launch.Source, common.MemoryFileID{}, common.MemoryFileID{})))
		if err != nil || filepath.Dir(source.Path) != dir {
			return nil, fmt.Errorf("staged source is outside its band")
		}
	}
	return &record, nil
}

// Stage persists only hidden bytes. The actor revalidates selection again after
// this write, before recording its held cut and matching population.
func (m *memory) Stage(id string, task *common.Compression, text string) (*common.BandPopulatedData, error) {
	if strings.TrimSpace(text) == "" {
		return nil, fmt.Errorf("compaction memory is empty")
	}
	if err := m.ValidateSelection(task); err != nil {
		return nil, err
	}
	pending := m.parent.History().Context().Compression
	if pending == nil || id == "" || pending.Launch.TaskID != id || pending.Committed != nil {
		return nil, fmt.Errorf("candidate has no matching uncommitted launch")
	}
	launch := pending.Launch
	if launch.SelectionHash != task.SelectionHash || launch.Source != task.Source || launch.Band != task.Destination || launch.File != task.File || launch.Thru != task.Thru || launch.From != task.From || launch.To != task.To {
		return nil, fmt.Errorf("candidate differs from launched selection")
	}
	if old, err := m.readStage(); err != nil {
		return nil, err
	} else if old != nil {
		return nil, fmt.Errorf("previous staged compaction requires recovery")
	}
	population := common.BandPopulatedData{TaskID: id, Band: task.Destination, File: task.File, Thru: task.Thru, Text: text, Source: "compression"}
	record := stagedCompaction{Version: 1, Launch: launch, Selection: *task.Selection, Population: population, CandidateHash: digest([]byte(text))}
	data, err := json.Marshal(record)
	if err != nil {
		return nil, err
	}
	if err := m.atomicWrite(m.stagePath(), data); err != nil {
		return nil, err
	}
	return &population, nil
}

func (m *memory) currentFile(f common.MemoryFingerprint) (common.MemoryFingerprint, error) {
	if !filepath.IsLocal(f.Path) {
		return common.MemoryFingerprint{}, fmt.Errorf("invalid owner-relative memory path")
	}
	return m.fingerprint(filepath.Join(m.parent.Config().DataDir, f.Path))
}

// Finish rechecks every precondition immediately before its mutation. It accepts
// matching already-installed/retired bytes, but never overwrites a conflicting
// destination or archive. External concurrent edits can require manual repair;
// portable renames cannot provide a transaction against another writer.
func (m *memory) Finish(id string) error {
	record, err := m.readStage()
	if err != nil {
		return err
	}
	pending := m.parent.History().Context().Compression
	if record == nil || pending == nil || pending.Committed == nil || pending.Launch.TaskID != id || !reflect.DeepEqual(pending.Launch, record.Launch) || !reflect.DeepEqual(*pending.Committed, record.Population) {
		return fmt.Errorf("committed compaction has no matching staged receipt")
	}
	destination := record.Selection.Destination
	current, err := m.currentFile(destination)
	if err != nil {
		return err
	}
	if !current.Exists || current.Hash != record.CandidateHash {
		if current != destination {
			return fmt.Errorf("compaction destination changed: %s", destination.Path)
		}
		if err := m.atomicWrite(filepath.Join(m.parent.Config().DataDir, destination.Path), []byte(record.Population.Text)); err != nil {
			return err
		}
	}
	for _, source := range record.Selection.Sources {
		current, err := m.currentFile(source)
		if err != nil {
			return err
		}
		retired := source
		retired.Path += ".retired"
		archive, err := m.currentFile(retired)
		if err != nil {
			return err
		}
		if !current.Exists && archive == retired {
			continue
		}
		if current != source || archive.Exists {
			return fmt.Errorf("compaction source/archive changed: %s", source.Path)
		}
		if err := m.parent.Sandbox().Rename(filepath.Join(m.parent.Config().DataDir, source.Path), filepath.Join(m.parent.Config().DataDir, retired.Path)); err != nil {
			return err
		}
	}
	return nil
}

// Recover makes no projection decisions of its own. It returns the terminal
// fact for the actor, and deletes staging only after that fact is recorded. A
// missing receipt is an error, not permission to guess whether files committed.
func (m *memory) Recover() (*common.CompactionOutcomeData, error) {
	record, err := m.readStage()
	if err != nil {
		return nil, err
	}
	c := m.parent.History().Context()
	pending := c.Compression
	if pending == nil || pending.Launch.TaskID == "" {
		if record == nil {
			return nil, nil
		}
		if c.CompactionOutcome == nil || c.CompactionOutcome.TaskID != record.Launch.TaskID {
			return nil, fmt.Errorf("staged compaction has no recovered launch/outcome")
		}
		if err := m.parent.Sandbox().Remove(m.stagePath()); err != nil && !os.IsNotExist(err) {
			return nil, err
		}
		return nil, nil
	}
	if record != nil && !reflect.DeepEqual(record.Launch, pending.Launch) {
		return nil, fmt.Errorf("staged compaction differs from recovered launch")
	}
	outcome := &common.CompactionOutcomeData{TaskID: pending.Launch.TaskID, TracePath: pending.Launch.TracePath}
	if pending.Committed == nil {
		outcome.State, outcome.Reason = "abandoned", "restart before replacement commit; source retained"
		return outcome, nil
	}
	if err := m.Finish(pending.Launch.TaskID); err != nil {
		return nil, err
	}
	outcome.State = "committed"
	return outcome, nil
}

func (m *memory) ready() error {
	if pending := m.parent.History().Context().Compression; pending != nil && pending.Committed != nil {
		return fmt.Errorf("memory commit requires disk recovery before corpus access")
	}
	return nil
}
