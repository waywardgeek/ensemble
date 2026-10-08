package gui

import (
	"bytes"
	"encoding/json"
	"errors"
	"example.com/ensemble"
	"fmt"
	"io"
)

// Check outer duplicate keys before map decoding can erase their presence.
func commandObject(c *Connector, data []byte) error {
	d := json.NewDecoder(bytes.NewReader(data))
	if _, err := d.Token(); err != nil {
		return fmt.Errorf("invalid command object")
	}
	seen := map[string]bool{}
	for d.More() {
		t, err := d.Token()
		if err != nil {
			return fmt.Errorf("invalid command object")
		}
		k := t.(string)
		if seen[k] {
			return fmt.Errorf("duplicate command field")
		}
		seen[k] = true
		var raw json.RawMessage
		if d.Decode(&raw) != nil {
			return fmt.Errorf("invalid command value")
		}
	}
	if _, err := d.Token(); err != nil {
		return fmt.Errorf("invalid command object")
	}
	if _, err := d.Token(); err != io.EOF {
		return fmt.Errorf("trailing command data")
	}
	return nil
}
func (c *Connector) settingsError(id string, err error) {
	var failure *ensemble.SettingsError
	if !errors.As(err, &failure) {
		failure = &ensemble.SettingsError{Code: "settings_closed", Message: "settings unavailable"}
	}
	out := map[string]any{"type": "error", "id": id, "code": failure.Code, "message": failure.Message}
	if failure.Domain != "" {
		out["domain"] = failure.Domain
		out["current"] = failure.Current
	}
	c.send(out, false)
}
func (c *Connector) advancePreferences(revision uint64) {
	c.mu.Lock()
	c.preferencesRevision = revision
	close(c.preferencesAdvanced)
	c.preferencesAdvanced = make(chan struct{})
	c.mu.Unlock()
}
func (c *Connector) waitSettingsRevision(policy bool, revision uint64) bool {
	select {
	case <-c.snapshotDone:
	case <-c.ctx.Done():
		return false
	}
	for {
		c.mu.Lock()
		current, advanced := c.preferencesRevision, c.preferencesAdvanced
		if policy {
			current, advanced = c.revision, c.advanced
		}
		c.mu.Unlock()
		if current >= revision {
			return true
		}
		select {
		case <-advanced:
		case <-c.ctx.Done():
			return false
		}
	}
}
func (c *Connector) deliverPreferences() {
	defer c.workers.Done()
	for {
		v, err := c.preferences.Next(c.ctx)
		if err != nil {
			c.stop("preferences watch lost; resync required")
			return
		}
		if !c.send(map[string]any{"type": "preferences_changed", "revision": v.Revision, "preferences": v.Preferences}, false) {
			return
		}
		c.advancePreferences(v.Revision)
	}
}
func (c *Connector) updateSettings(id, kind string, base uint64, patch json.RawMessage) {
	defer c.workers.Done()
	if kind == "preferences_update" {
		v, err := c.parent.Preferences().Update(base, patch)
		if err != nil {
			c.settingsError(id, err)
			return
		}
		if c.waitSettingsRevision(false, v.Revision) {
			c.send(map[string]any{"type": "preferences_ack", "id": id, "revision": v.Revision}, false)
		}
		return
	}
	ack, err := c.parent.Ensemble().UpdatePolicy(c.parent.AgentID(), base, patch)
	if err != nil {
		c.settingsError(id, err)
		return
	}
	if c.waitSettingsRevision(true, ack.WatchRevision) {
		c.send(map[string]any{"type": "policy_ack", "id": id, "revision": ack.Revision, "watch_revision": ack.WatchRevision}, false)
	}
}
