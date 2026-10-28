package gradeedition2

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/waywardgeek/ensemble/internal/fakevendor"
	"github.com/waywardgeek/ensemble/internal/grade"
)

// Keep actual paired requests, not a rewritten transcript. A completed wait may
// contain only the output after the initial observation's consumed cursor.
type sandboxPair struct {
	request fakevendor.Recorded
	id      string
}
type sandboxObservation []sandboxPair

func (o sandboxObservation) result() (string, bool, error) {
	if len(o) == 0 {
		return "", false, fmt.Errorf("missing observation")
	}
	var text []string
	var bad bool
	for _, p := range o {
		s, b, err := anthropicToolText(p.request.Body, p.id)
		if err != nil {
			return "", false, err
		}
		text = append(text, s)
		bad = b
	}
	return strings.Join(text, "\n"), bad, nil
}

func (o sandboxObservation) skill(c *grade.Check, id string, wantError bool) string {
	text, bad, err := o.result()
	if err != nil || bad != wantError || strings.TrimSpace(text) == "" {
		fail(c, "result %s: error=%t want=%t text=%q decode=%v", id, bad, wantError, text, err)
	}
	return text
}

var sandboxRunning = regexp.MustCompile(`(?m)^job ([0-9]+) running[;\n ]`)

// Only the current route uses this sequential model script. Observation delay
// is not command lifetime: wait on a real returned handle before advancing.
// The process/session budget and finite wait count bound a broken fixture.
type sandboxScript struct {
	probes                  []sandboxProbe
	firstText, final        string
	next, waits, totalWaits int
	pending                 string
	handle                  int
	finished                bool
	explained               bool // C# requires explanation before every tool, including waits.
	observed                map[string]sandboxObservation
	err                     error
}

func (s *sandboxScript) route(body []byte) *fakevendor.Reply {
	finish := func(err error) *fakevendor.Reply {
		s.err = err
		s.finished = true
		return &fakevendor.Reply{Text: s.final}
	}
	if s.finished {
		return finish(fmt.Errorf("unexpected request after final response"))
	}
	if s.pending != "" {
		text, bad, err := anthropicToolText(body, s.pending)
		if err != nil {
			return finish(fmt.Errorf("result %s: %w", s.pending, err))
		}
		probe := s.probes[s.next]
		s.observed[probe.id] = append(s.observed[probe.id], sandboxPair{fakevendor.Recorded{Body: append([]byte(nil), body...)}, s.pending})
		if m := sandboxRunning.FindStringSubmatch(text); !bad && len(m) != 0 {
			h, err := strconv.Atoi(m[1])
			if err != nil || h <= 0 || (s.handle != 0 && h != s.handle) {
				return finish(fmt.Errorf("probe %s returned invalid or changed job handle", probe.id))
			}
			if s.waits >= 8 {
				return finish(fmt.Errorf("probe %s did not settle within eight waits", probe.id))
			}
			s.handle = h
			s.waits++
			s.totalWaits++
			s.pending = fmt.Sprintf("sandbox-wait-%s-%d", probe.id, s.waits)
			r := (sandboxProbe{s.pending, "wait_for_job", map[string]any{"handle": h, "ai_callback_delay": 5, "max_output_bytes": 65536}}).reply()
			if s.explained {
				r.Text = "I will observe the running command before continuing."
			}
			return &r
		}
		s.next++
		s.waits, s.handle = 0, 0
	}
	if s.next == len(s.probes) {
		return finish(nil)
	}
	probe := s.probes[s.next]
	s.pending = probe.id
	r := probe.reply()
	if s.explained {
		r.Text = "I will exercise the next sandbox boundary."
	}
	if s.next == 0 {
		r.Text = s.firstText
	}
	return &r
}
