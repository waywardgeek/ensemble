// Package cachelens diagnoses prefix changes without interpreting conversation text.
package cachelens

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"

	"ensemble/internal/common"
)

// Lens belongs to one Engine and retains only its preceding main request.
// Interrupted workers can overlap new turns; each response keeps its own prediction.
type Lens struct {
	parent   common.Engine
	mu       sync.Mutex
	previous map[string]json.RawMessage
	name     string
}

// New establishes the required ownership link; files are created only on live use.
func New(parent common.Engine) *Lens {
	if parent == nil {
		panic("CacheLens requires Engine")
	}
	return &Lens{parent: parent}
}

// Engine returns the transport that owns this diagnostic.
func (l *Lens) Engine() common.Engine { return l.parent }

// Marker metadata is the sole byte edit. Do not marshal decoded content: that
// would change key order, escaping and the byte counts we claim to measure.
// These vendor markers contain only scalar fields. Both comma positions matter
// because a provider serializer may put the marker first or last in an object.
func strip(body []byte) []byte {
	marker := `"(?:cache_control|prompt_cache_breakpoint)"\s*:\s*\{[^{}]*\}`
	body = regexp.MustCompile(`,\s*`+marker).ReplaceAll(body, nil)
	body = regexp.MustCompile(marker+`\s*,\s*`).ReplaceAll(body, nil)
	return regexp.MustCompile(marker).ReplaceAll(body, nil)
}

// ObserveRequest rotates readable section files and reports real serialized
// prefix bytes, returning this request's prediction for its eventual response.
// Section order follows the vendor's cache order, not Go field order.
func (l *Lens) ObserveRequest(name string, body []byte) float64 {
	// Interrupt ends the turn before its worker necessarily exits. Serialize the
	// comparison and artifact rotation together, never the HTTP exchange.
	l.mu.Lock()
	defer l.mu.Unlock()
	root := l.parent.Agent().Ensemble()
	var sections map[string]json.RawMessage
	if err := json.Unmarshal(strip(body), &sections); err != nil {
		root.Logf("cache lens: invalid request JSON")
		return 0
	}
	// A model/route switch starts a fresh comparison rather than accusing its
	// deliberately different system/tools prefix of an accidental rewrite.
	if l.name != name {
		l.previous = nil
		l.name = name
	}
	keys := []string{}
	for _, key := range []string{"tools", "system", "systemInstruction", "messages", "contents"} {
		if _, ok := sections[key]; ok {
			keys = append(keys, key)
		}
	}
	// Retain every remaining field in the artifact for ordinary diff inspection;
	// generation controls are not prompt bytes and do not enter the prediction.
	var other []string
	for key := range sections {
		found := false
		for _, k := range keys {
			found = found || key == k
		}
		if !found {
			other = append(other, key)
		}
	}
	sort.Strings(other)
	var artifact strings.Builder
	var verdicts []string
	total, prefix := 0, 0
	open := l.previous != nil
	for _, key := range keys {
		now, old := sections[key], l.previous[key]
		state := "changed"
		if l.previous == nil {
			state = "cold"
		} else if bytes.Equal(old, now) {
			state = "identical"
		} else if appended(old, now) {
			state = "appended"
		}
		verdicts = append(verdicts, key+":"+state)
		// Chat Completions embeds its stable system instruction in messages.
		// Inspect that head separately, but count its bytes only once.
		if l.previous != nil && key == "messages" {
			var a, b []json.RawMessage
			json.Unmarshal(old, &a)
			json.Unmarshal(now, &b)
			if len(a) > 0 && len(b) > 0 {
				var head struct {
					// Role identifies the embedded OpenAI system instruction.
					Role string `json:"role"`
				}
				json.Unmarshal(b[0], &head)
				if head.Role == "system" && !bytes.Equal(a[0], b[0]) {
					root.Logf("cache prefix changed: system (changed)")
				}
			}
		}
		total += len(now)
		if open {
			n := 0
			for n < len(old) && n < len(now) && old[n] == now[n] {
				n++
			}
			prefix += n
			if n != len(now) {
				open = false
			}
		}
		// Conversation arrays are recognized by their role-bearing elements, including
		// Gemini's contents. Only unexpected non-conversation changes raise the alarm.
		if l.previous != nil && state != "identical" && !conversation(now) {
			root.Logf("cache prefix changed: %s (%s)", key, state)
		}
	}
	for _, key := range append(keys, other...) {
		fmt.Fprintf(&artifact, "%s:\n%s\n", key, sections[key])
	}
	predicted := 0.0
	if total > 0 {
		predicted = float64(prefix) / float64(total)
	}
	root.Logf("cache lens %s: %s; prefix_bytes=%d prompt_bytes=%d predicted=%.4f", name, strings.Join(verdicts, " "), prefix, total, predicted)
	dir := l.parent.Agent().Config().DataDir
	if err := os.MkdirAll(dir, 0700); err != nil {
		root.Logf("cache lens: create directory: %v", err)
	} else {
		current := filepath.Join(dir, "cache-current.txt")
		if err := os.Rename(current, filepath.Join(dir, "cache-prior.txt")); err != nil && !os.IsNotExist(err) {
			root.Logf("cache lens: rotate: %v", err)
		}
		if err := os.WriteFile(current, []byte(artifact.String()), 0600); err != nil {
			root.Logf("cache lens: write: %v", err)
		}
	}
	l.previous = sections
	return predicted
}

func appended(old, now []byte) bool {
	var a, b []json.RawMessage
	if json.Unmarshal(old, &a) != nil || json.Unmarshal(now, &b) != nil || len(b) <= len(a) {
		return false
	}
	for i := range a {
		if !bytes.Equal(a[i], b[i]) {
			return false
		}
	}
	return true
}

func conversation(raw []byte) bool {
	var entries []map[string]json.RawMessage
	if json.Unmarshal(raw, &entries) != nil || len(entries) == 0 {
		return false
	}
	for _, entry := range entries {
		if _, ok := entry["role"]; !ok {
			return false
		}
	}
	return true
}

// ObserveUsage keeps tokens and bytes separate: only their dimensionless ratios
// are comparable. A zero grant alone says nothing about minimum eligibility.
func (l *Lens) ObserveUsage(u common.Usage, predicted float64) {
	total := u.Input + u.CacheWrite + u.CacheRead
	rate := 0.0
	if total > 0 {
		rate = float64(u.CacheRead) / float64(total)
	}
	l.parent.Agent().Ensemble().Logf("cache grant: input_tokens=%d read_tokens=%d granted=%.4f predicted=%.4f", total, u.CacheRead, rate, predicted)
}
