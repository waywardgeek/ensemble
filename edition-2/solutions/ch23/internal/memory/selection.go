package memory

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"

	"ensemble/internal/common"
)

func digest(data []byte) string { return fmt.Sprintf("%x", sha256.Sum256(data)) }

func (m *memory) fingerprint(name string) (common.MemoryFingerprint, error) {
	rel, err := filepath.Rel(m.parent.Config().DataDir, name)
	if err != nil || !filepath.IsLocal(rel) {
		return common.MemoryFingerprint{}, fmt.Errorf("memory file is outside owner directory")
	}
	f := common.MemoryFingerprint{Path: rel}
	data, err := m.parent.Sandbox().ReadFile(name)
	if os.IsNotExist(err) {
		return f, nil
	}
	if err != nil {
		return f, err
	}
	f.Exists, f.Hash = true, digest(data)
	return f, nil
}

// A snapshot includes only the entries the cut will actually remove. Growing
// dialogue, skills and later corpus files cannot make an otherwise valid result
// stale. Full part identities matter even when the displayed source text matches.
func (m *memory) selection(t *common.Compression) (*common.CompressionSelection, error) {
	if t == nil || t.Source == 0 && t.Destination != common.BandSession || t.Source != 0 && (t.Source < common.Band64x || t.Source > common.BandSession || t.Destination != t.Source-1) {
		return nil, fmt.Errorf("invalid compaction selection")
	}
	c := m.parent.History().Context()
	var entries []common.Entry
	for _, e := range c.Dialogue {
		if t.Source == 0 {
			if e.Seq >= t.From && e.Seq <= t.To && (e.Kind <= common.KindDialogue || e.Kind == common.KindHandoff) {
				entries = append(entries, e)
			}
		} else if e.Kind == Kind(t.Source) && !Less(t.Thru, e.File) {
			entries = append(entries, e)
		}
	}
	if len(entries) == 0 || t.Source == 0 && (t.From == 0 || t.To < t.From || entries[len(entries)-1].Kind != common.KindHandoff) {
		return nil, fmt.Errorf("selected completed source no longer exists")
	}
	if t.Source != 0 && !reflect.DeepEqual(entries, t.Retired) {
		return nil, fmt.Errorf("graduation no longer retires exactly its selected prefix")
	}
	// Bind task boundaries and text as well as projected entries. A caller may
	// retain the selection value, but changing the proposed cut cannot reuse it.
	encoded, err := json.Marshal(struct {
		// Source and Destination bind the lifetime transition to these entries.
		Source, Destination common.Band
		// File and Thru bind the replacement identity and covered prefix.
		File, Thru common.MemoryFileID
		// From and To bind the inclusive conversation cut.
		From, To uint64
		// Text is the exact task input sent to the worker.
		Text string
		// Entries preserve selected identities and all content parts.
		Entries []common.Entry
	}{t.Source, t.Destination, t.File, t.Thru, t.From, t.To, t.Text, entries})
	if err != nil {
		return nil, err
	}
	s := &common.CompressionSelection{EntriesHash: digest(encoded), SourcePolicy: m.Settings(t.Source), DestinationPolicy: m.Settings(t.Destination)}
	if s.SourcePolicy.Disabled || s.DestinationPolicy.Disabled {
		return nil, fmt.Errorf("compaction source or destination is disabled")
	}
	for _, e := range t.Retired {
		f, err := m.fingerprint(path(m.parent, t.Source, e.File, e.Thru))
		if err != nil {
			return nil, err
		}
		if !f.Exists || f.Hash != digest([]byte(text(e))) {
			return nil, fmt.Errorf("selected memory file differs from source context: %s", f.Path)
		}
		s.Sources = append(s.Sources, f)
	}
	s.Destination, err = m.fingerprint(path(m.parent, t.Destination, t.File, t.Thru))
	if err != nil {
		return nil, err
	}
	if t.Destination != common.BandMemory && s.Destination.Exists {
		return nil, fmt.Errorf("memory destination already exists: %s", s.Destination.Path)
	}
	if t.Destination == common.BandMemory {
		// The terminal file is also source material. A stale context must not
		// overwrite an edit made after the last scan, even before worker launch.
		var existing string
		for _, e := range c.Dialogue {
			if e.Kind == common.KindMemory {
				existing = text(e)
			}
		}
		if s.Destination.Exists && s.Destination.Hash != digest([]byte(existing)) || !s.Destination.Exists && existing != "" {
			return nil, fmt.Errorf("curated memory differs from source context")
		}
	}
	return s, nil
}

func (m *memory) selected(t *common.Compression) (*common.Compression, error) {
	selection, err := m.selection(t)
	if err != nil {
		return nil, err
	}
	t.Selection = selection
	encoded, err := json.Marshal(selection)
	if err != nil {
		return nil, err
	}
	t.SelectionHash = digest(encoded)
	return t, nil
}

// ValidateSelection rechecks exact selected content and destination bytes. The
// actor additionally compares its generation so disable/re-enable cannot revive
// old work merely because settings happen to have returned to their old values.
func (m *memory) ValidateSelection(t *common.Compression) error {
	if t == nil || t.Selection == nil {
		return fmt.Errorf("compaction has no captured selection")
	}
	current, err := m.selection(t)
	if err != nil {
		return err
	}
	encoded, err := json.Marshal(current)
	if err != nil {
		return err
	}
	if !reflect.DeepEqual(current, t.Selection) || digest(encoded) != t.SelectionHash {
		return fmt.Errorf("compaction selection changed while worker ran")
	}
	return nil
}
