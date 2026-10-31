package history

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"

	"ensemble/internal/common"
)

// Save takes the snapshot and anchor under one lock, so a concurrent append
// cannot leave a projection paired with the wrong sequence number.
func (h *history) Save(w io.Writer) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.encodeSave(w)
}

// Caller holds the history lock through encoding and any disk checkpoint.
func (h *history) encodeSave(w io.Writer) error {
	cfg := h.parent.Config()
	log := h.events
	if log == nil {
		log = []common.Event{}
	}
	if cfg.LogRetention > 0 && len(log) > cfg.LogRetention {
		log = log[len(log)-cfg.LogRetention:]
	}
	tools := h.parent.Declarations()
	if h.context.StartupTools != nil {
		tools = *h.context.StartupTools
	}
	if tools == nil {
		tools = []common.ToolDeclaration{}
	}
	return json.NewEncoder(w).Encode(common.SaveFile{
		Config: common.SaveConfig{Model: cfg.Model, Vendor: cfg.Vendor, SystemPrompt: h.parent.Engine().SystemPrompt(), Tools: tools},
		AsOf:   h.asOf, Context: &h.context, Log: log,
	})
}

// Restore stages the replacement before touching live history. A bad existing
// save is an error, not permission to overwrite it with a fresh conversation.
func (h *history) Restore(r io.Reader) error {
	data, err := io.ReadAll(r)
	if err != nil {
		return err
	}
	var sf *common.SaveFile
	if err := json.Unmarshal(data, &sf); err != nil {
		return err
	}
	if sf == nil || sf.Config.Model == "" || sf.Config.Vendor < common.Anthropic || sf.Config.Vendor > common.Gemini {
		return errors.New("invalid save configuration")
	}
	ctx, after := sf.Context, sf.AsOf
	if ctx == nil {
		ctx, after = &common.Context{Turn: "idle"}, 0
	}
	switch ctx.Turn {
	case "idle", "input_pending", "in_flight", "tools_pending", "interrupted":
	default:
		return errors.New("invalid saved turn state")
	}
	// Retain the anchor even without any log: the next appended fact must not
	// reuse an identity already present in the snapshot.
	next := history{parent: h.parent, context: *ctx}
	last := sf.AsOf
	for i, event := range sf.Log {
		if event.Seq == 0 || (i > 0 && event.Seq <= sf.Log[i-1].Seq) {
			return errors.New("save log seq must ascend")
		}
		if !knownEvent(event.Type) {
			return fmt.Errorf("unknown event type %q", event.Type)
		}
		if event.Seq > after {
			next.replay(event)
		}
		if event.Seq > last {
			last = event.Seq
		}
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	h.events, h.context, h.asOf = sf.Log, next.context, last
	return nil
}
