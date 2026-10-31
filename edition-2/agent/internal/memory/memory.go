// Package memory selects completed work and owns the Agent-relative memory corpus.
// It does not mutate conversation: the actor records every decision in History.
package memory

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"ensemble/internal/common"
)

// Settings returns effective byte watermarks without copying policy into siblings.
func (m *memory) Settings(band common.Band) common.BandSettings {
	cfg := m.parent.Config().Memory
	s := cfg.Conversation
	switch band {
	case common.BandSoul:
		s = cfg.Soul
	case common.BandMemory:
		s = cfg.Memory
	case common.Band64x:
		s = cfg.Tier64
	case common.Band8x:
		s = cfg.Tier8
	case common.BandSession:
		s = cfg.Session
	}
	if s.Budget == 0 {
		s.Budget = 1500
		if band == 0 {
			s.Budget = 3000
		}
	}
	if s.HighWatermark == 0 {
		s.HighWatermark = 2 * s.Budget
	}
	if s.LowWatermark == 0 {
		s.LowWatermark = s.Budget
	}
	return s
}

// Validate rejects nonsensical settings before a client saves or applies them.
func Validate(m common.MemorySettings) error {
	for i, s := range []common.BandSettings{m.Conversation, m.Soul, m.Memory, m.Tier64, m.Tier8, m.Session} {
		if s.Budget < 0 || s.HighWatermark < 0 || s.LowWatermark < 0 {
			return fmt.Errorf("memory sizes must be nonnegative")
		}
		budget := s.Budget
		if budget == 0 {
			budget = 1500
			if i == 0 {
				budget = 3000
			}
		}
		high, low := s.HighWatermark, s.LowWatermark
		if high == 0 {
			high = 2 * budget
		}
		if low == 0 {
			low = budget
		}
		if low > high {
			return fmt.Errorf("memory low watermark exceeds high watermark")
		}
	}
	return nil
}

// Kind binds each memory band to a lifetime unaffected by tool clearing.
func Kind(b common.Band) common.EntryKind { return common.KindSoul + common.EntryKind(b-1) }

// Less orders corpus identities without relying on the event log that delivered them.
func Less(a, b common.MemoryFileID) bool { return a.Date < b.Date || a.Date == b.Date && a.Num < b.Num }

func stem(id common.MemoryFileID) string { return fmt.Sprintf("%s-%d", id.Date, id.Num) }
func parse(s string) (common.MemoryFileID, error) {
	if len(s) < 12 || s[10] != '-' {
		return common.MemoryFileID{}, fmt.Errorf("invalid memory filename %q", s)
	}
	if _, err := time.Parse("2006-01-02", s[:10]); err != nil {
		return common.MemoryFileID{}, err
	}
	n, err := strconv.Atoi(s[11:])
	if err != nil || n < 1 {
		return common.MemoryFileID{}, fmt.Errorf("invalid memory number")
	}
	return common.MemoryFileID{Date: s[:10], Num: n}, nil
}
func path(a common.Agent, b common.Band, file, thru common.MemoryFileID) string {
	dir := a.Config().DataDir
	switch b {
	case common.BandSoul:
		return filepath.Join(dir, "SOUL.md")
	case common.BandMemory:
		return filepath.Join(dir, "MEMORY.md")
	case common.BandSession:
		dir = filepath.Join(dir, "memory")
	case common.Band8x:
		dir = filepath.Join(dir, "bucket-0")
	case common.Band64x:
		dir = filepath.Join(dir, "bucket-1")
	}
	name := stem(file)
	if thru.Date != "" && thru != file {
		name += "_" + stem(thru)
	}
	return filepath.Join(dir, name+".md")
}

// Scan rereads enabled corpus files; no file identity cache can hide human edits.
func (m *memory) Scan(b common.Band) ([]common.BandPopulatedData, error) {
	if err := m.ready(); err != nil {
		return nil, err
	}
	a := m.parent
	var paths []string
	if b <= common.BandMemory {
		paths = []string{path(a, b, common.MemoryFileID{}, common.MemoryFileID{})}
	} else {
		var err error
		paths, err = filepath.Glob(filepath.Join(filepath.Dir(path(a, b, common.MemoryFileID{}, common.MemoryFileID{})), "*.md"))
		if err != nil {
			return nil, err
		}
	}
	var out []common.BandPopulatedData
	for _, p := range paths {
		data, err := m.parent.Sandbox().ReadFile(p)
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			return nil, err
		}
		item := common.BandPopulatedData{Band: b, Text: string(data), Source: "disk"}
		if b > common.BandMemory {
			names := strings.Split(strings.TrimSuffix(filepath.Base(p), ".md"), "_")
			if len(names) > 2 {
				return nil, fmt.Errorf("invalid memory coverage filename %q", p)
			}
			item.File, err = parse(names[0])
			if err != nil {
				a.Ensemble().Logf("memory: skipped %s: %v", p, err)
				continue
			}
			if len(names) == 2 {
				item.Thru, err = parse(names[1])
				if err != nil {
					return nil, err
				}
				if Less(item.Thru, item.File) {
					return nil, fmt.Errorf("reversed memory coverage in %q", p)
				}
			}
		}
		out = append(out, item)
	}
	sort.Slice(out, func(i, j int) bool { return Less(out[i].File, out[j].File) })
	return out, nil
}

func text(e common.Entry) string {
	var b strings.Builder
	for _, p := range e.Parts {
		if p.Type == "text" {
			b.WriteString(p.Text)
		}
	}
	return b.String()
}
func nextID(a common.Agent) (common.MemoryFileID, error) {
	id := common.MemoryFileID{Date: time.Now().UTC().Format("2006-01-02"), Num: 1}
	// Archived session names still reserve their identity across fresh starts.
	paths, err := filepath.Glob(filepath.Join(a.Config().DataDir, "memory", id.Date+"-*.md*"))
	if err != nil {
		return id, err
	}
	for _, p := range paths {
		n := strings.TrimSuffix(strings.TrimSuffix(filepath.Base(p), ".retired"), ".md")
		prev, err := parse(n)
		if err == nil {
			id.Num = max(id.Num, prev.Num+1)
		}
	}
	return id, nil
}

// Conversation selects oldest whole checkpoint segments until dialogue meets its budget.
// Handoff notes enter the compressor input too; their surviving kind is not counted as dialogue.
func (m *memory) Conversation() (*common.Compression, error) {
	if err := m.ready(); err != nil {
		return nil, err
	}
	a := m.parent
	s := m.Settings(0)
	if s.Disabled || m.Settings(common.BandSession).Disabled {
		return nil, nil
	}
	c := a.History().Context()
	total := 0
	for _, e := range c.Dialogue {
		if e.Kind <= common.KindDialogue {
			total += len(text(e))
		}
	}
	if total <= s.Budget {
		return nil, nil
	}
	task := &common.Compression{Destination: common.BandSession, Ratio: 2}
	var input strings.Builder
	removed := 0
	for _, e := range c.Dialogue {
		if e.Kind > common.KindHandoff {
			continue
		}
		if task.From == 0 {
			task.From = e.Seq
		}
		fmt.Fprintf(&input, "%s: %s\n", e.Actor, text(e))
		if e.Kind <= common.KindDialogue {
			removed += len(text(e))
		}
		if e.Kind == common.KindHandoff {
			task.To = e.Seq
			task.Text = input.String()
			if total-removed <= s.Budget {
				break
			}
		}
	}
	if task.To == 0 {
		return nil, nil
	}
	var err error
	task.File, err = nextID(a)
	task.Thru = task.File
	if err != nil {
		return nil, err
	}
	return m.selected(task)
}

// Graduation walks bottom-up, returning one oldest-first fold at a time.
func (m *memory) Graduation() (*common.Compression, error) {
	if err := m.ready(); err != nil {
		return nil, err
	}
	a := m.parent
	c := a.History().Context()
	for b := common.BandSession; b >= common.Band64x; b-- {
		s := m.Settings(b)
		if s.Disabled {
			continue
		}
		var files []common.Entry
		size := 0
		for _, e := range c.Dialogue {
			if e.Kind == Kind(b) {
				files = append(files, e)
				size += len(text(e))
			}
		}
		if size <= s.HighWatermark || len(files) < 2 {
			continue
		}
		if m.Settings(b - 1).Disabled {
			a.Ensemble().Logf("memory: cannot graduate band %d into disabled band %d", b, b-1)
			continue
		}
		// Retirement removes a prefix by coverage endpoint. Overlapping or reversed
		// ranges could retire unselected files, or none at all and repeat forever.
		// Validate restored entries too: they need not have come from today's Scan.
		var previous common.MemoryFileID
		for i, e := range files {
			end := e.Thru
			if end.Date == "" {
				end = e.File
			}
			if _, err := parse(stem(e.File)); err != nil {
				return nil, fmt.Errorf("invalid memory coverage start: %w", err)
			}
			if _, err := parse(stem(end)); err != nil {
				return nil, fmt.Errorf("invalid memory coverage end: %w", err)
			}
			if Less(end, e.File) || i > 0 && !Less(previous, e.File) {
				return nil, fmt.Errorf("memory band %d has reversed or overlapping coverage", b)
			}
			previous = end
		}
		task := &common.Compression{Source: b, Destination: b - 1, Ratio: 8}
		var input strings.Builder
		// Watermarks, not a required batch of eight, choose how much must retire.
		for _, e := range files {
			if size <= s.LowWatermark && len(task.Retired) >= 2 {
				break
			}
			if len(task.Retired) == 0 {
				task.File = e.File
			}
			task.Thru = e.File
			if e.Thru.Date != "" {
				task.Thru = e.Thru
			}
			task.Retired = append(task.Retired, e)
			size -= len(text(e))
			fmt.Fprintf(&input, "Memory %s: %s\n", stem(e.File), text(e))
		}
		if task.Destination == common.BandMemory {
			// MEMORY.md is one terminal document; retain the old knowledge when updating it.
			for _, e := range c.Dialogue {
				if e.Kind == common.KindMemory {
					fmt.Fprintf(&input, "Existing curated memory: %s\n", text(e))
				}
			}
		}
		task.Text = input.String()
		return m.selected(task)
	}
	return nil, nil
}

// Write persists new bytes before the actor retires any source information.
func (m *memory) Write(t *common.Compression, summary string) error {
	a := m.parent
	p := path(a, t.Destination, t.File, t.Thru)
	if err := m.parent.Sandbox().MkdirAll(filepath.Dir(p), 0700); err != nil {
		return err
	}
	if err := m.parent.Sandbox().WriteFile(p+".tmp", []byte(summary), 0600); err != nil {
		return err
	}
	return m.parent.Sandbox().Rename(p+".tmp", p)
}

// Archive removes graduated files from future scans without destroying the corpus.
func (m *memory) Archive(t *common.Compression) error {
	a := m.parent
	for _, e := range t.Retired {
		p := path(a, t.Source, e.File, e.Thru)
		if err := m.parent.Sandbox().Rename(p, p+".retired"); err != nil && !os.IsNotExist(err) {
			return err
		}
	}
	return nil
}

// The corpus has one actual parent. It reaches policy, history and diagnostics
// through that owner instead of retaining sibling services or copied settings.
type memory struct{ parent common.Agent }

// New creates the Agent's memory child before any startup corpus scans.
func New(a common.Agent) common.Memory {
	if a == nil {
		panic("Memory requires Agent")
	}
	return &memory{parent: a}
}

// Agent exposes the required ownership link used by the composition root.
func (m *memory) Agent() common.Agent { return m.parent }

// Validate shares the pure settings boundary with external clients.
func (m *memory) Validate(settings common.MemorySettings) error { return Validate(settings) }
