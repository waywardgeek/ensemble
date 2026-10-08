// Package policy owns execution-policy validation and persistence. The actor
// alone orders application of its candidate with turn activation.
package policy

import (
	"bytes"
	"encoding/json"
	"example.com/ensemble/internal/common"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"unicode/utf8"
)

type Service struct {
	parent       common.Agent
	path         string
	mu           sync.Mutex
	value        common.PolicySnapshot
	busy, closed bool
}

func (s *Service) Agent() common.Agent { return s.parent }
func New(parent common.Agent, path string) (*Service, error) {
	s := &Service{parent: parent, value: common.PolicySnapshot{EffectiveMaxModelRequests: common.DefaultMaxModelRequests}}
	if path == "" {
		return s, nil
	}
	resolved, err := parent.Ensemble().ClaimSettingsPath(path)
	if err != nil {
		return nil, err
	}
	s.path = resolved
	s.value.Persistent = true
	good := false
	defer func() {
		if !good {
			parent.Ensemble().ReleaseSettingsPath(resolved)
		}
	}()
	if err = os.MkdirAll(filepath.Dir(resolved), 0700); err != nil {
		return nil, fmt.Errorf("cannot create policy directory")
	}
	f, err := os.Open(resolved)
	if os.IsNotExist(err) {
		good = true
		return s, nil
	}
	if err != nil {
		return nil, fmt.Errorf("cannot read policy file")
	}
	defer f.Close()
	data, err := io.ReadAll(io.LimitReader(f, 65537))
	if err != nil || len(data) > 65536 {
		return nil, fmt.Errorf("policy file exceeds limit or cannot be read")
	}
	fields, err := object(s, data)
	if err != nil || len(fields) != 3 {
		return nil, fmt.Errorf("invalid complete policy file")
	}
	var version int
	var revision uint64
	if !decode(s, fields["version"], &version) || version != 1 || !decode(s, fields["revision"], &revision) {
		return nil, fmt.Errorf("unsupported policy version or invalid revision")
	}
	value, err := patch(s, fields["policy"])
	if err != nil {
		return nil, err
	}
	s.value.Revision = revision
	s.value.MaxModelRequests = value
	s.value.EffectiveMaxModelRequests = effective(s, value)
	good = true
	return s, nil
}

// These strict parser helpers keep the service argument even where parsing is
// currently stateless: future diagnostics must still reach the actual owner.
func decode(s *Service, raw json.RawMessage, target any) bool {
	return len(raw) > 0 && !bytes.Equal(bytes.TrimSpace(raw), []byte("null")) && json.Unmarshal(raw, target) == nil
}
func object(s *Service, data []byte) (map[string]json.RawMessage, error) {
	if !utf8.Valid(data) {
		return nil, fmt.Errorf("settings must be UTF-8")
	}
	d := json.NewDecoder(bytes.NewReader(data))
	t, err := d.Token()
	if err != nil || t != json.Delim('{') {
		return nil, fmt.Errorf("settings must be an object")
	}
	fields := map[string]json.RawMessage{}
	for d.More() {
		t, err = d.Token()
		if err != nil {
			return nil, fmt.Errorf("invalid settings object")
		}
		key, ok := t.(string)
		if !ok {
			return nil, fmt.Errorf("invalid settings key")
		}
		if _, ok = fields[key]; ok {
			return nil, fmt.Errorf("duplicate settings field")
		}
		var raw json.RawMessage
		if d.Decode(&raw) != nil {
			return nil, fmt.Errorf("invalid settings value")
		}
		fields[key] = raw
	}
	if _, err = d.Token(); err != nil {
		return nil, fmt.Errorf("invalid settings object")
	}
	if _, err = d.Token(); err != io.EOF {
		return nil, fmt.Errorf("trailing settings data")
	}
	return fields, nil
}
func patch(s *Service, data []byte) (int, error) {
	f, err := object(s, data)
	if err != nil {
		return 0, err
	}
	if len(f) != 1 {
		return 0, fmt.Errorf("policy requires only max_model_requests")
	}
	var n int
	if !decode(s, f["max_model_requests"], &n) || n < 0 || n > 256 {
		return 0, fmt.Errorf("max_model_requests must be an integer between 0 and 256")
	}
	return n, nil
}
func effective(s *Service, n int) int {
	if n == 0 {
		return common.DefaultMaxModelRequests
	}
	return n
}
func (s *Service) Snapshot() common.PolicySnapshot { s.mu.Lock(); defer s.mu.Unlock(); return s.value }
func (s *Service) Prepare(base uint64, raw json.RawMessage) (common.PolicySnapshot, bool, error) {
	n, err := patch(s, raw)
	if err != nil {
		return common.PolicySnapshot{}, false, &common.SettingsError{Code: "invalid_policy", Message: err.Error()}
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	fail := func(code, message string) (common.PolicySnapshot, bool, error) {
		return s.value, false, &common.SettingsError{Code: code, Message: message}
	}
	if s.closed {
		return fail("settings_closed", "policy service closed")
	}
	if s.busy {
		return fail("settings_busy", "policy write in progress")
	}
	if base != s.value.Revision {
		return s.value, false, &common.SettingsError{Code: "revision_conflict", Message: "policy revision changed", Domain: "policy", Current: s.value}
	}
	if n == s.value.MaxModelRequests {
		return s.value, false, nil
	}
	if base == ^uint64(0) {
		return fail("settings_persist_failed", "policy revision exhausted")
	}
	next := s.value
	next.Revision++
	next.MaxModelRequests = n
	next.EffectiveMaxModelRequests = effective(s, n)
	s.busy = true
	return next, true, nil
}
func (s *Service) Persist(next common.PolicySnapshot) error {
	if s.path == "" {
		return nil
	}
	data, err := json.Marshal(struct {
		Version  int    `json:"version"`
		Revision uint64 `json:"revision"`
		Policy   struct {
			Max int `json:"max_model_requests"`
		} `json:"policy"`
	}{1, next.Revision, struct {
		Max int `json:"max_model_requests"`
	}{next.MaxModelRequests}})
	if err == nil {
		err = replace(s, data)
	}
	if err != nil {
		s.parent.Ensemble().Logf("policy persistence failed before replacement")
		return &common.SettingsError{Code: "settings_persist_failed", Message: "cannot persist policy"}
	}
	return nil
}
func replace(s *Service, data []byte) error {
	f, err := os.CreateTemp(filepath.Dir(s.path), ".policy-*")
	if err != nil {
		return err
	}
	name := f.Name()
	defer os.Remove(name)
	n, err := f.Write(data)
	if err == nil && n != len(data) {
		err = io.ErrShortWrite
	}
	if err == nil {
		err = f.Sync()
	}
	closeErr := f.Close()
	if err == nil {
		err = closeErr
	}
	if err != nil {
		return err
	}
	return os.Rename(name, s.path)
}
func (s *Service) Apply(next common.PolicySnapshot, err error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err == nil {
		s.value = next
	}
	s.busy = false
}

// Close is called after the actor has joined the writer and applied its result.
func (s *Service) Close() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return
	}
	s.closed = true
	if s.path != "" {
		s.parent.Ensemble().ReleaseSettingsPath(s.path)
	}
}
