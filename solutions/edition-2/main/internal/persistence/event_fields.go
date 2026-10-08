package persistence

import (
	"encoding/json"
	"example.com/ensemble/internal/common"
	"reflect"
	"strconv"
)

// Earlier outer event annotations remain open. New session/limit payloads and
// identity are closed, so unknown or missing members cannot be silently dropped.
func (c *Codec) eventFields(value any) error {
	obj, ok := value.(map[string]any)
	if !ok {
		return nil
	}
	kind, _ := obj["type"].(string)
	keys := []string{}
	payload := "session"
	switch kind {
	case "session_initialized":
		keys = []string{"session_id", "identity"}
	case "session_anchor":
		keys = []string{"session_id", "origin_as_of", "origin_sha256", "high_watermarks"}
	case "tool_limits_set":
		payload = "limits"
		keys = []string{"call_id", "overrides"}
	case "tool_limits_consumed":
		payload = "limits"
		keys = []string{"call_id", "name", "overrides"}
	default:
		return nil
	}
	fields, ok := obj[payload].(map[string]any)
	if !ok || len(fields) != len(keys) {
		return c.bad("invalid session fact shape")
	}
	for _, key := range keys {
		if v, found := fields[key]; !found || v == nil {
			return c.bad("missing session fact field")
		}
	}
	if payload == "limits" {
		id, ok := fields["call_id"].(string)
		if !ok || id == "" {
			return c.bad("invalid limit call identity")
		}
		if kind == "tool_limits_consumed" {
			name, ok := fields["name"].(string)
			if !ok || name == "" {
				return c.bad("invalid consumed name")
			}
		}
		var limits common.LimitValues
		return c.unwire(fields["overrides"], reflect.ValueOf(&limits).Elem(), false)
	}
	id, ok := fields["session_id"].(string)
	if !ok || !hexadecimal(id, 32) {
		return c.bad("invalid session identity")
	}
	if kind == "session_initialized" {
		var identity common.SessionIdentity
		if err := c.unwire(fields["identity"], reflect.ValueOf(&identity).Elem(), false); err != nil {
			return err
		}
		return c.Identity(identity)
	}
	hash, ok := fields["origin_sha256"].(string)
	if !ok || !hexadecimal(hash, 64) {
		return c.bad("invalid origin hash")
	}
	seq, ok := fields["origin_as_of"].(json.Number)
	if !ok {
		return c.bad("invalid origin boundary")
	}
	n, err := strconv.ParseUint(string(seq), 10, 64)
	if err != nil || n == 0 {
		return c.bad("invalid origin boundary")
	}
	var w common.Watermarks
	if err = c.unwire(fields["high_watermarks"], reflect.ValueOf(&w).Elem(), false); err != nil {
		return err
	}
	if w.Event != n {
		return c.bad("invalid origin watermarks")
	}
	return nil
}
