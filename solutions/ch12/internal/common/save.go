package common

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// SaveConfig captures everything the renderer needs to reproduce
// exactly the same wire format: model, vendor, system prompt, and
// tool declarations. The API key is deliberately absent.
type SaveConfig struct {
	Model        string     `json:"model"`
	Vendor       string     `json:"vendor"` // stored as string for readability
	SystemPrompt string     `json:"system_prompt"`
	Tools        []ToolDecl `json:"tools"`
}

// SaveFile is the on-disk representation of a persisted agent.
// It contains enough information to reconstruct the agent's
// exact state, resume the conversation, or interview a snapshot.
//
// AsOf is the anchor: the Seq of the last event already folded into
// Context. Everything in Log above it is the tail still to apply. A
// snapshot with no log is a complete save; so is a log with no
// snapshot.
type SaveFile struct {
	Config  SaveConfig `json:"config"`
	AsOf    Seq        `json:"as_of"`
	Context *Context   `json:"context"`
	Log     []Event    `json:"log"`
}

// Save writes the agent's complete state to a JSON file.
//
// asOf is the Seq of the last event folded into ctx. The write goes to
// a temporary file in the same directory and is then renamed over the
// target, because rename is atomic within a directory: a crash halfway
// through leaves the previous save intact rather than a truncated file
// where the only copy of the conversation used to be.
func Save(path string, asOf Seq, ctx *Context, log *Log, cfg Config) error {
	sf := SaveFile{
		Config: SaveConfig{
			Model:        cfg.Model,
			Vendor:       cfg.Vendor.String(),
			SystemPrompt: cfg.SystemPrompt,
			Tools:        cfg.Tools,
		},
		AsOf:    asOf,
		Context: ctx,
		Log:     log.Events,
	}
	data, err := json.MarshalIndent(sf, "", "  ")
	if err != nil {
		return fmt.Errorf("save: marshal: %w", err)
	}
	dir := filepath.Dir(path)
	tmp, err := os.CreateTemp(dir, ".save-*.json")
	if err != nil {
		return fmt.Errorf("save: temp file: %w", err)
	}
	tmpName := tmp.Name()
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		os.Remove(tmpName)
		return fmt.Errorf("save: write: %w", err)
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmpName)
		return fmt.Errorf("save: close: %w", err)
	}
	// CreateTemp makes the file 0600; the save is ordinary user data,
	// not a secret (the API key is deliberately not in it).
	if err := os.Chmod(tmpName, 0644); err != nil {
		os.Remove(tmpName)
		return fmt.Errorf("save: chmod: %w", err)
	}
	if err := os.Rename(tmpName, path); err != nil {
		os.Remove(tmpName)
		return fmt.Errorf("save: rename: %w", err)
	}
	return nil
}

// Load reads a SaveFile from disk.
func Load(path string) (*SaveFile, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("load: read: %w", err)
	}
	var sf SaveFile
	if err := json.Unmarshal(data, &sf); err != nil {
		return nil, fmt.Errorf("load: unmarshal: %w", err)
	}
	return &sf, nil
}

// Restore turns a save file back into a live Context. It is the only
// replay loop in the agent, and it implements rule 3 of the chapter:
//
//   - Context non-null: install it, then apply every log event whose
//     Seq is strictly greater than AsOf. Events at or below AsOf are
//     already inside the snapshot, and applying one twice is a bug —
//     the conversation would grow a duplicate turn.
//   - Context null: there is no anchor to respect, so rebuild from
//     nothing by applying the whole log to a fresh context.
//
// Both paths must land on the same Context for the same save. That
// equivalence is the chapter's falsifiable claim, and the grader
// checks it by comparing the vendor requests the two produce.
func (sf *SaveFile) Restore() (*Context, error) {
	ctx, after := sf.Context, sf.AsOf
	if ctx == nil {
		ctx, after = NewContext(), 0
	}
	for _, e := range sf.Log {
		if e.Seq <= after {
			continue
		}
		if err := ctx.Apply(e); err != nil {
			return nil, fmt.Errorf("restore: event %d: %w", e.Seq, err)
		}
	}
	return ctx, nil
}

// NextSeq is the Seq the first new event should get: one past the
// larger of the anchor and the last event in the log (rule 5). The two
// can disagree — a save with a snapshot and an empty log knows its
// anchor and nothing else — so neither alone is enough.
func (sf *SaveFile) NextSeq() Seq {
	last := sf.AsOf
	for _, e := range sf.Log {
		if e.Seq > last {
			last = e.Seq
		}
	}
	return last + 1
}
