// Package preferences owns display defaults; it has no Agent execution policy.
package preferences

import (
	"bytes"
	"context"
	"encoding/json"
	"example.com/ensemble"
	"example.com/ensemble-gui/internal/common"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"sync"
	"unicode/utf8"
)

type Service struct {
	parent       common.Server
	path         string
	mu           sync.Mutex
	value        common.PreferencesSnapshot
	busy, closed bool
	writer       sync.WaitGroup
	closeDone    chan struct{}
	watches      map[*watch]bool
}
type watch struct {
	parent     common.PreferencesService
	mu         sync.Mutex
	values     []common.PreferencesSnapshot
	done, wake chan struct{}
	closed     bool
}

func (s *Service) Server() common.Server            { return s.parent }
func (w *watch) Service() common.PreferencesService { return w.parent }
func (w *watch) Done() <-chan struct{}              { return w.done }
func (w *watch) stop() {
	w.mu.Lock()
	defer w.mu.Unlock()
	if !w.closed {
		w.closed = true
		w.values = nil
		close(w.done)
	}
}
func (w *watch) Close() { w.parent.CloseWatch(w) }
func (s *Service) CloseWatch(subscription common.PreferencesWatch) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for w := range s.watches {
		if w == subscription {
			delete(s.watches, w)
			w.stop()
			break
		}
	}
}
func (w *watch) push(value common.PreferencesSnapshot) {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.closed {
		return
	}
	if len(w.values) >= ensemble.WatchItems {
		w.closed = true
		w.values = nil
		close(w.done)
		return
	}
	w.values = append(w.values, value)
	select {
	case w.wake <- struct{}{}:
	default:
	}
}
func (w *watch) Next(ctx context.Context) (common.PreferencesSnapshot, error) {
	for {
		w.mu.Lock()
		if w.closed {
			w.mu.Unlock()
			return common.PreferencesSnapshot{}, fmt.Errorf("preferences watch lost")
		}
		if len(w.values) > 0 {
			v := w.values[0]
			w.values = w.values[1:]
			w.mu.Unlock()
			return v, nil
		}
		w.mu.Unlock()
		select {
		case <-w.wake:
		case <-w.done:
		case <-ctx.Done():
			return common.PreferencesSnapshot{}, ctx.Err()
		}
	}
}
func New(parent common.Server, path string) (*Service, error) {
	s := &Service{parent: parent, closeDone: make(chan struct{}), watches: map[*watch]bool{}, value: common.PreferencesSnapshot{Preferences: common.Preferences{Theme: "dark", FontSize: 16, SidebarWidth: 260, ActionsWidth: 380, SpeechRate: 1}}}
	resolved, err := parent.Ensemble().ClaimSettingsPath(path)
	if err != nil {
		return nil, err
	}
	s.path = resolved
	good := false
	defer func() {
		if !good {
			parent.Ensemble().ReleaseSettingsPath(resolved)
		}
	}()
	if os.MkdirAll(filepath.Dir(resolved), 0700) != nil {
		return nil, fmt.Errorf("cannot create preferences directory")
	}
	f, err := os.Open(resolved)
	if os.IsNotExist(err) {
		good = true
		return s, nil
	}
	if err != nil {
		return nil, fmt.Errorf("cannot read preferences file")
	}
	defer f.Close()
	data, err := io.ReadAll(io.LimitReader(f, 65537))
	if err != nil || len(data) > 65536 {
		return nil, fmt.Errorf("preferences file exceeds limit or cannot be read")
	}
	fields, err := object(s, data)
	if err != nil || len(fields) != 3 {
		return nil, fmt.Errorf("invalid complete preferences file")
	}
	var version int
	var revision uint64
	if !decode(s, fields["version"], &version) || version != 1 || !decode(s, fields["revision"], &revision) {
		return nil, fmt.Errorf("unsupported preferences version or invalid revision")
	}
	values, err := object(s, fields["preferences"])
	if err != nil || len(values) != 6 {
		return nil, fmt.Errorf("incomplete preferences file")
	}
	value, err := merge(s, s.value.Preferences, values)
	if err != nil {
		return nil, err
	}
	s.value = common.PreferencesSnapshot{Revision: revision, Preferences: value}
	good = true
	return s, nil
}
func decode(s *Service, raw json.RawMessage, target any) bool {
	return len(raw) > 0 && !bytes.Equal(bytes.TrimSpace(raw), []byte("null")) && json.Unmarshal(raw, target) == nil
}
func object(s *Service, data []byte) (map[string]json.RawMessage, error) {
	if !utf8.Valid(data) {
		return nil, fmt.Errorf("preferences must be UTF-8")
	}
	d := json.NewDecoder(bytes.NewReader(data))
	t, err := d.Token()
	if err != nil || t != json.Delim('{') {
		return nil, fmt.Errorf("preferences must be an object")
	}
	f := map[string]json.RawMessage{}
	for d.More() {
		t, err = d.Token()
		if err != nil {
			return nil, fmt.Errorf("invalid preferences object")
		}
		k, ok := t.(string)
		if !ok {
			return nil, fmt.Errorf("invalid preferences key")
		}
		if _, ok = f[k]; ok {
			return nil, fmt.Errorf("duplicate preferences field")
		}
		var raw json.RawMessage
		if d.Decode(&raw) != nil {
			return nil, fmt.Errorf("invalid preferences value")
		}
		f[k] = raw
	}
	if _, err = d.Token(); err != nil {
		return nil, fmt.Errorf("invalid preferences object")
	}
	if _, err = d.Token(); err != io.EOF {
		return nil, fmt.Errorf("trailing preferences data")
	}
	return f, nil
}
func merge(s *Service, v common.Preferences, f map[string]json.RawMessage) (common.Preferences, error) {
	if len(f) == 0 {
		return v, fmt.Errorf("preferences patch must be nonempty")
	}
	for k, raw := range f {
		switch k {
		case "theme":
			if !decode(s, raw, &v.Theme) || (v.Theme != "dark" && v.Theme != "light" && v.Theme != "system") {
				return v, fmt.Errorf("theme must be dark, light or system")
			}
		case "font_size":
			if !decode(s, raw, &v.FontSize) || v.FontSize < 12 || v.FontSize > 28 {
				return v, fmt.Errorf("font_size must be an integer between 12 and 28")
			}
		case "sidebar_width":
			if !decode(s, raw, &v.SidebarWidth) || v.SidebarWidth < 180 || v.SidebarWidth > 480 {
				return v, fmt.Errorf("sidebar_width must be an integer between 180 and 480")
			}
		case "actions_width":
			if !decode(s, raw, &v.ActionsWidth) || v.ActionsWidth < 200 || v.ActionsWidth > 640 {
				return v, fmt.Errorf("actions_width must be an integer between 200 and 640")
			}
		case "autoplay":
			if !decode(s, raw, &v.Autoplay) {
				return v, fmt.Errorf("autoplay must be Boolean")
			}
		case "speech_rate":
			if !decode(s, raw, &v.SpeechRate) || math.IsNaN(v.SpeechRate) || math.IsInf(v.SpeechRate, 0) || v.SpeechRate < 0.5 || v.SpeechRate > 2 {
				return v, fmt.Errorf("speech_rate must be between 0.5 and 2")
			}
		default:
			return v, fmt.Errorf("unknown preferences field")
		}
	}
	return v, nil
}
func (s *Service) Snapshot() common.PreferencesSnapshot {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.value
}
func (s *Service) Subscribe() (common.PreferencesSnapshot, common.PreferencesWatch, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return s.value, nil, &ensemble.SettingsError{Code: "settings_closed", Message: "preferences closed"}
	}
	w := &watch{parent: s, done: make(chan struct{}), wake: make(chan struct{}, 1)}
	s.watches[w] = true
	return s.value, w, nil
}
func (s *Service) Update(base uint64, patch json.RawMessage) (common.PreferencesSnapshot, error) {
	fields, err := object(s, patch)
	if err != nil {
		return common.PreferencesSnapshot{}, &ensemble.SettingsError{Code: "invalid_preferences", Message: err.Error()}
	}
	s.mu.Lock()
	candidate, err := merge(s, s.value.Preferences, fields)
	if err != nil {
		s.mu.Unlock()
		return common.PreferencesSnapshot{}, &ensemble.SettingsError{Code: "invalid_preferences", Message: err.Error()}
	}
	fail := func(code, message string) (common.PreferencesSnapshot, error) {
		v := s.value
		s.mu.Unlock()
		return v, &ensemble.SettingsError{Code: code, Message: message}
	}
	if s.closed {
		return fail("settings_closed", "preferences closed")
	}
	if s.busy {
		return fail("settings_busy", "preferences write in progress")
	}
	if base != s.value.Revision {
		v := s.value
		s.mu.Unlock()
		return v, &ensemble.SettingsError{Code: "revision_conflict", Message: "preferences revision changed", Domain: "preferences", Current: v}
	}
	if candidate == s.value.Preferences {
		v := s.value
		s.mu.Unlock()
		return v, nil
	}
	if base == ^uint64(0) {
		return fail("settings_persist_failed", "preferences revision exhausted")
	}
	next := common.PreferencesSnapshot{Revision: base + 1, Preferences: candidate}
	s.busy = true
	s.writer.Add(1)
	s.mu.Unlock()
	defer s.writer.Done()
	data, _ := json.Marshal(struct {
		Version     int                `json:"version"`
		Revision    uint64             `json:"revision"`
		Preferences common.Preferences `json:"preferences"`
	}{1, next.Revision, candidate})
	err = replace(s, data)
	s.mu.Lock()
	defer s.mu.Unlock()
	s.busy = false
	if err != nil {
		s.parent.Ensemble().Logf("preferences persistence failed before replacement")
		return s.value, &ensemble.SettingsError{Code: "settings_persist_failed", Message: "cannot persist preferences"}
	}
	s.value = next
	for w := range s.watches {
		w.push(next)
		select {
		case <-w.done:
			delete(s.watches, w)
		default:
		}
	}
	return next, nil
}
func replace(s *Service, data []byte) error {
	f, err := os.CreateTemp(filepath.Dir(s.path), ".preferences-*")
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
func (s *Service) Close() {
	s.mu.Lock()
	if s.closed {
		s.mu.Unlock()
		<-s.closeDone
		return
	}
	s.closed = true
	s.mu.Unlock()
	s.writer.Wait()
	s.mu.Lock()
	defer s.mu.Unlock()
	for w := range s.watches {
		w.stop()
	}
	s.watches = map[*watch]bool{}
	s.parent.Ensemble().ReleaseSettingsPath(s.path)
	close(s.closeDone)
}
