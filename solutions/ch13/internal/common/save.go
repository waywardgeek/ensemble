package common

import (
	"encoding/json"
	"fmt"
	"os"
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
type SaveFile struct {
	Config  SaveConfig `json:"config"`
	Context *Context   `json:"context"`
	Log     []Event    `json:"log"`
}

// Save writes the agent's complete state to a JSON file.
func Save(path string, ctx *Context, log *Log, cfg Config) error {
	sf := SaveFile{
		Config: SaveConfig{
			Model:        cfg.Model,
			Vendor:       cfg.Vendor.String(),
			SystemPrompt: cfg.SystemPrompt,
			Tools:        cfg.Tools,
		},
		Context: ctx,
		Log:     log.Events,
	}
	data, err := json.MarshalIndent(sf, "", "  ")
	if err != nil {
		return fmt.Errorf("save: marshal: %w", err)
	}
	return os.WriteFile(path, data, 0644)
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

// Rebuild creates a fresh Context by replaying every event in
// the log through Apply. If the reducer is deterministic, the
// result must be byte-identical to a Context that was built
// incrementally during the original conversation.
func Rebuild(events []Event) (*Context, error) {
	ctx := NewContext()
	for _, e := range events {
		if err := ctx.Apply(e); err != nil {
			return nil, fmt.Errorf("rebuild: event %d: %w", e.Seq, err)
		}
	}
	return ctx, nil
}
