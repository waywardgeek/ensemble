package engine

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"ensemble/internal/common"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// Only the current/prior actual sends are retained. Files and predictions are
// local diagnostics, never provider cache keys or durable conversation state.
type cache struct {
	files             common.Sandbox
	parent            common.Engine
	mu                sync.Mutex
	current, previous []byte
	currentCfg        common.Config
	currentReport     []byte
	nextID            uint64
	dir               string
	status            common.CacheStatus
}

const noSentRequest = "No main HTTP request has been sent in this Engine process. Prior-process request diagnostics are not restored.\n"

func NewCache(parent common.Engine) common.Cache {
	c := &cache{parent: parent}
	if parent != nil {
		c.files = parent.Parent().Sandbox()
		cfg := parent.Parent().Config()
		c.dir = cfg.DataDir
		if !filepath.IsAbs(c.dir) {
			c.dir = filepath.Join(cfg.BaseDir, c.dir)
		}
		c.write("cache-current.txt", []byte(noSentRequest))
		c.write("cache-prior.txt", []byte(noSentRequest))
	}
	return c
}
func (c *cache) Parent() common.Engine { return c.parent }
func digest(b []byte) string           { sum := sha256.Sum256(b); return hex.EncodeToString(sum[:]) }
func requestRoute(cfg common.Config, body []byte) string {
	var request struct{ Model string }
	_ = json.Unmarshal(body, &request)
	if request.Model == "" {
		request.Model = cfg.Model
	}
	return fmt.Sprintf("%s / %s / %s / %s / resolved=%s", cfg.Vendor, cfg.Surface, cfg.BaseURL, request.Model, cfg.ResolvedModel)
}
func requestSnapshot(cfg common.Config, body, clean []byte, id uint64) (*common.RequestSnapshot, int) {
	s := &common.RequestSnapshot{Digest: digest(body), Bytes: len(body), RequestID: id, Route: requestRoute(cfg, body)}
	sections := promptSections(clean)
	s.PromptBytes = len(promptBytes(sections))
	original, _ := wireRoot(body)
	_, marked := markerMembers(cfg, original, body)
	count := 0
	add := func(path, kind string, b []byte, isMarked bool) {
		s.Segments = append(s.Segments, common.RequestSegment{Path: path, Kind: kind, Digest: digest(b), Bytes: len(b), Marked: isMarked})
	}
	for _, section := range sections {
		old := original.field(section.source)
		if len(section.node.items) == 0 {
			add(section.name, section.name, section.raw, false)
			continue
		}
		for i, n := range section.node.items {
			sourceIndex := i + section.offset
			isMarked := sourceIndex < len(old.items) && marked[old.items[sourceIndex].start]
			if isMarked {
				count++
			}
			// A message can contain both history and handoff markers after coalescing.
			if section.source == "messages" && sourceIndex < len(old.items) {
				count -= boolInt(isMarked)
				for _, part := range old.items[sourceIndex].field("content").items {
					if marked[part.start] {
						count++
					}
				}
			}
			add(fmt.Sprintf("%s[%d]", section.name, i), section.name, clean[n.start:n.end], isMarked)
		}
	}
	return s, count
}
func boolInt(v bool) int {
	if v {
		return 1
	}
	return 0
}
func (c *cache) Observe(cfg common.Config, body []byte) common.CachePrediction {
	c.mu.Lock()
	defer c.mu.Unlock()
	clean := stripCacheMarkers(cfg, body)
	sections := promptSections(clean)
	prompt := promptBytes(sections)
	prior := c.status.Current
	c.nextID++
	current, count := requestSnapshot(cfg, body, clean, c.nextID)
	_, reason := common.CacheBreakpoints(cfg)
	status := common.CacheStatus{Current: current, Previous: prior, MarkerCount: count, MarkerReason: reason, OptionsChange: "first send", Comparison: c.status.Comparison}
	prediction := common.CachePrediction{RequestID: c.nextID, PromptBytes: len(prompt), Reason: "first send in this process; no prior request comparison"}
	for _, s := range sections {
		status.Sections = append(status.Sections, common.CacheSection{Name: s.name, Change: "first send", Bytes: len(s.raw)})
	}
	if prior != nil {
		previousClean := stripCacheMarkers(c.currentCfg, c.current)
		previousSections := promptSections(previousClean)
		if bytes.Equal(optionBytes(previousClean), optionBytes(clean)) {
			status.OptionsChange = "unchanged"
		} else {
			status.OptionsChange = "changed"
		}
		if prior.Route != current.Route {
			prediction.Reason = "route/model changed; prefix comparison reset"
			status.FirstDifference = "route/model changed"
			for i := range status.Sections {
				status.Sections[i].Change = "not comparable across route/model"
			}
		} else {
			shared := bytePrefix(promptBytes(previousSections), prompt)
			status.SharedBytes = shared
			prediction.SharedBytes = shared
			prediction.Reason = "marker-free local prompt byte prefix estimate; not a token count or provider cache grant"
			if len(prompt) > 0 {
				fraction := float64(shared) / float64(len(prompt))
				prediction.Fraction = &fraction
			}
			previousByName := map[string]promptSection{}
			for _, s := range previousSections {
				previousByName[s.name] = s
			}
			for i, s := range sections {
				old, ok := previousByName[s.name]
				change := "added"
				if ok {
					change = sectionChange(old, s)
				}
				status.Sections[i].Change = change
				delete(previousByName, s.name)
				if change != "unchanged" && status.FirstDifference == "" {
					at := bytePrefix(old.raw, s.raw)
					status.FirstDifference = fmt.Sprintf("%s byte %d: prior %q; current %q", s.name, at, nearByte(old.raw, at), nearByte(s.raw, at))
				}
			}
			for _, s := range previousSections {
				if _, ok := previousByName[s.name]; ok {
					status.Sections = append(status.Sections, common.CacheSection{Name: s.name, Change: "removed"})
					if status.FirstDifference == "" {
						status.FirstDifference = s.name + " removed"
					}
				}
			}
			for i, segment := range current.Segments {
				if i >= len(prior.Segments) || segment.Path != prior.Segments[i].Path || segment.Digest != prior.Segments[i].Digest {
					break
				}
				status.SharedSegments++
			}
		}
	}
	if cfg.Surface == common.Responses {
		prediction.Reason += "; private continuation redacted from local comparison"
	}
	status.Prediction = prediction
	c.previous, c.current = c.current, append([]byte(nil), body...)
	// Retain only routing metadata, never a Config copy containing credentials.
	c.currentCfg = common.Config{Vendor: cfg.Vendor, Surface: cfg.Surface}
	c.status = status
	if len(c.currentReport) > 0 {
		c.write("cache-prior.txt", c.currentReport)
	}
	c.currentReport = diagnosticReport(status, body, clean)
	c.write("cache-current.txt", c.currentReport)
	return prediction
}
func nearByte(b []byte, at int) string {
	if at > len(b) {
		at = len(b)
	}
	start := at - 24
	if start < 0 {
		start = 0
	}
	end := at + 48
	if end > len(b) {
		end = len(b)
	}
	return string(b[start:end])
}
func diagnosticReport(s common.CacheStatus, body, clean []byte) []byte {
	var b strings.Builder
	fmt.Fprintf(&b, "Local sent-request diagnostics; not a provider cache grant.\nRequest %d: %s\nWire bytes: %d; prompt bytes: %d\n%s\n", s.Current.RequestID, s.Current.Route, len(body), s.Current.PromptBytes, s.Prediction.Reason)
	for _, section := range s.Sections {
		fmt.Fprintf(&b, "%s: %s (%d bytes)\n", section.Name, section.Change, section.Bytes)
	}
	fmt.Fprintf(&b, "Non-prompt options: %s\nShared prompt byte prefix estimate: %d/%d\nFirst difference: %s\nMarkers: %d; %s\n", s.OptionsChange, s.SharedBytes, s.Current.PromptBytes, s.FirstDifference, s.MarkerCount, s.MarkerReason)
	b.WriteString("\nPrompt sections in provider cache order (marker-free raw spans; not necessarily standalone JSON):\n")
	sections := promptSections(clean)
	for _, name := range []string{"tools", "system", "messages"} {
		fmt.Fprintf(&b, "=== %s ===\n", strings.ToUpper(name))
		found := false
		for _, section := range sections {
			if section.name == name {
				b.Write(section.raw)
				found = true
			}
		}
		if !found {
			b.WriteString("(absent from request)")
		}
		b.WriteByte('\n')
	}
	b.WriteString("=== NONPROMPT OPTIONS ===\n")
	b.Write(optionBytes(clean))
	b.WriteString("\n\nExact sent request:\n")
	b.Write(body)
	b.WriteString("\n\nMarker-free comparison request (all other bytes unchanged):\n")
	b.Write(clean)
	b.WriteByte('\n')
	return []byte(b.String())
}
func (c *cache) write(name string, b []byte) {
	if c.dir == "" {
		return
	}
	mkdir, rename := os.MkdirAll, os.Rename
	if c.files != nil {
		mkdir = c.files.MkdirAll
		rename = c.files.Rename
	}
	err := mkdir(c.dir, 0700)
	if err == nil {
		path := filepath.Join(c.dir, name)
		tmp := path + ".tmp"
		if c.files == nil {
			err = os.WriteFile(tmp, b, 0600)
		} else {
			var f *os.File
			f, err = c.files.OpenFile(tmp, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
			if err == nil {
				_, err = f.Write(b)
				closeErr := f.Close()
				if err == nil {
					err = closeErr
				}
			}
		}
		if err == nil {
			err = rename(tmp, path)
		}
	}
	if err != nil {
		c.status.DiagnosticError = err.Error()
	}
}
func (c *cache) Report(prediction common.CachePrediction, u common.Usage) {
	c.mu.Lock()
	defer c.mu.Unlock()
	// The token is captured at this request's send, not read from current state.
	c.status.Comparison = &common.CacheComparison{Prediction: prediction, ActualHitRate: common.CacheHitRate(u)}
}
func (c *cache) Status() common.CacheStatus {
	c.mu.Lock()
	defer c.mu.Unlock()
	s := c.status
	clone := func(in *common.RequestSnapshot) *common.RequestSnapshot {
		if in == nil {
			return nil
		}
		out := *in
		out.Segments = append([]common.RequestSegment(nil), in.Segments...)
		return &out
	}
	s.Current, s.Previous = clone(s.Current), clone(s.Previous)
	s.Sections = append([]common.CacheSection(nil), s.Sections...)
	if s.Prediction.Fraction != nil {
		v := *s.Prediction.Fraction
		s.Prediction.Fraction = &v
	}
	if s.Comparison != nil {
		v := *s.Comparison
		s.Comparison = &v
		if v.ActualHitRate != nil {
			n := *v.ActualHitRate
			s.Comparison.ActualHitRate = &n
		}
		if v.Prediction.Fraction != nil {
			n := *v.Prediction.Fraction
			s.Comparison.Prediction.Fraction = &n
		}
	}
	return s
}
