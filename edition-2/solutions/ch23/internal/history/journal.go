package history

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"ensemble/internal/common"
)

// A closed append reaches the OS before publication. No fsync is promised:
// this journal protects process crashes, not loss of machine power.
func (h *history) journalEvent(e common.Event) error {
	if h.journal == "" {
		return nil
	}
	f, err := h.parent.Sandbox().OpenFile(h.journal, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	err = json.NewEncoder(f).Encode(e)
	closeErr := f.Close()
	if err == nil {
		err = closeErr
	}
	return err
}

// Recover stages both sources before replacing live state. Overlap is expected
// if a process died after installing the snapshot but before resetting its journal.
func (h *history) Recover(path string) error {
	next := &history{parent: h.parent, context: common.Context{Turn: "idle"}}
	f, err := h.parent.Sandbox().OpenFile(path, os.O_RDONLY, 0)
	if err == nil {
		err = next.Restore(f)
		f.Close()
	}
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("load %s: %w", path, err)
	}
	f, err = h.parent.Sandbox().OpenFile(path+".journal", os.O_RDONLY, 0)
	if err == nil {
		defer f.Close()
		reader := bufio.NewReader(f)
		var previous uint64
		for {
			line, readErr := reader.ReadBytes('\n')
			if readErr != nil && readErr != io.EOF {
				return readErr
			}
			if len(line) > 0 {
				var e common.Event
				if err := json.Unmarshal(line, &e); err != nil {
					return fmt.Errorf("journal: %w", err)
				}
				if e.Seq == 0 || e.Seq <= previous {
					return fmt.Errorf("journal seq must ascend")
				}
				previous = e.Seq
				if !knownEvent(e.Type) {
					return fmt.Errorf("unknown event type %q", e.Type)
				}
				if e.Seq > next.asOf {
					next.replay(e)
					next.events = append(next.events, e)
					next.asOf = e.Seq
				}
			}
			if readErr == io.EOF {
				break
			}
		}
	} else if !os.IsNotExist(err) {
		return err
	}
	if err := h.parent.Sandbox().MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	h.events, h.context, h.asOf = next.events, next.context, next.asOf
	h.journal = path + ".journal"
	return nil
}

// Checkpoint keeps the old snapshot as a backup, installs the new one atomically,
// and only then resets the journal. The lock keeps the anchor and retained tail
// coherent even if an application requests a save while its Agent is running.
func (h *history) Checkpoint(path string) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	if err := h.parent.Sandbox().MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}
	f, err := h.parent.Sandbox().CreateTemp(filepath.Dir(path))
	if err != nil {
		return err
	}
	defer h.parent.Sandbox().Remove(f.Name())
	err = h.encodeSave(f)
	closeErr := f.Close()
	if err == nil {
		err = closeErr
	}
	if err != nil {
		return err
	}
	old, err := h.parent.Sandbox().ReadFile(path)
	if err == nil {
		if err = h.parent.Sandbox().WriteFile(path+".bak", old, 0600); err != nil {
			return err
		}
	} else if !os.IsNotExist(err) {
		return err
	}
	if err = h.parent.Sandbox().Rename(f.Name(), path); err != nil {
		return err
	}
	// Every journal fact is at or below the just-installed anchor. Keeping it
	// until this point makes either side of a crash recoverable without guessing.
	if err = h.parent.Sandbox().WriteFile(path+".journal", nil, 0600); err != nil {
		return err
	}
	// A fresh embedding client may Save before ever calling Load. Subsequent
	// launch/commit facts must still reach the journal established by this save.
	h.journal = path + ".journal"
	retention := h.parent.Config().LogRetention
	if retention > 0 && len(h.events) > retention {
		h.events = h.events[len(h.events)-retention:]
	}
	return nil
}
