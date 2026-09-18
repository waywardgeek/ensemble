package main

// ch07 exercise — a stream monitor.
//
// One agent, one observer, and a report about the SHAPE of the stream. The
// point of the exercise is not that text appears on a terminal; it is that a
// caller outside the module can see a response being built and measure it.
//
// Time to first delta is the number streaming exists to improve. A turn that
// takes nine seconds and shows its first word after eight is a different
// product from one that shows its first word after four hundred
// milliseconds, and the total is identical. Nothing except an observer on the
// stream can tell you which one you shipped.
//
// Protocol (stdin, one JSON object per line):
//
//	{"kind":"prompt","text":"..."}   — ask the agent
//	{"kind":"hint","text":"..."}     — steer it mid-turn
//	{"kind":"interrupt"}             — stop the turn
//	{"user":"..."}                   — same as kind:prompt, kept working
//
// Observations are written to stdout, one JSON object per line.
//
// Environment:
//
//	CH07_LOG        path for a copy of the observation stream
//	CH07_NO_STREAM  set to 1 to run with streaming disabled
//
// CH07_NO_STREAM is the interesting one. With it set, every response arrives
// as a single document and the agent emits ONE delta per part instead of
// dozens. The final text is identical, the event log is identical, and only
// the chunk count changes. That is the chapter's claim, and this is the
// switch that lets you check it rather than believe it.

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/waywardgeek/ensemble/agent"
)

func main() {
	host := cliHost{}
	fw := agent.NewFramework(host)

	cfg, err := agent.ConfigFromEnv()
	if err != nil {
		fmt.Fprintln(os.Stderr, "config:", err)
		os.Exit(2)
	}
	cfg.MaxTokens = 1024
	if os.Getenv("CH07_NO_STREAM") == "1" {
		cfg.DisableStreaming = true
	}

	ag := agent.NewAgent(cfg, envOr("CH07_AGENT_LOG", "ch07.log"))
	defer ag.Shutdown()

	actor := ag.NewActor()
	fw.Add("monitor", actor)

	out := bufio.NewWriter(os.Stdout)
	defer out.Flush()

	var outMu sync.Mutex
	emit := func(v any) {
		outMu.Lock()
		defer outMu.Unlock()
		b, err := json.Marshal(v)
		if err != nil {
			// A delta kind that was never set refuses to marshal. Report it
			// rather than dropping the line: a silently missing observation
			// is the hardest kind of bug to see from out here.
			fmt.Fprintln(os.Stderr, "emit:", err)
			return
		}
		out.Write(b)
		out.WriteByte('\n')
		out.Flush()
	}

	var logFile *os.File
	if p := os.Getenv("CH07_LOG"); p != "" {
		if f, err := os.Create(p); err == nil {
			logFile = f
			defer f.Close()
		}
	}
	logLine := func(v any) {
		if logFile == nil {
			return
		}
		if b, err := json.Marshal(v); err == nil {
			logFile.Write(b)
			logFile.Write([]byte{'\n'})
		}
	}

	mon := &monitor{}
	actor.Attach(observerFunc(func(obs agent.Observation) {
		mon.record(obs)
		line := observationJSON(obs)
		emit(line)
		logLine(line)

		// A turn ending is where the shape of the stream becomes reportable.
		if _, ok := obs.(agent.TurnEnded); ok {
			stats := mon.report()
			emit(stats)
			logLine(stats)
			mon.reset()
		}
	}))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go actor.Run(ctx)

	// stdin on its own goroutine, so a hint can arrive WHILE a turn is in
	// progress. Reading stdin from the loop that also waits for replies is
	// the deafness this framework spent a chapter removing.
	type line struct {
		msg inputMsg
		err error
	}
	lines := make(chan line, 16)
	go func() {
		defer close(lines)
		in := bufio.NewScanner(os.Stdin)
		in.Buffer(make([]byte, 0, 64*1024), 8*1024*1024)
		for in.Scan() {
			text := strings.TrimSpace(in.Text())
			if text == "" {
				continue
			}
			var msg inputMsg
			if err := json.Unmarshal([]byte(text), &msg); err != nil {
				lines <- line{err: err}
				continue
			}
			lines <- line{msg: msg}
		}
	}()

	var turns sync.WaitGroup
	ask := func(text string) {
		turns.Add(1)
		go func() {
			defer turns.Done()
			mon.start()
			reply, err := actor.Ask(text)
			if err != nil {
				emit(map[string]string{"error": err.Error()})
				return
			}
			emit(map[string]string{"assistant": reply})
		}()
	}

	for l := range lines {
		if l.err != nil {
			emit(map[string]string{"error": "bad input: " + l.err.Error()})
			continue
		}
		msg := l.msg

		if msg.Kind != nil {
			switch *msg.Kind {
			case "prompt":
				if msg.Text == nil {
					emit(map[string]string{"error": "prompt requires text"})
					continue
				}
				ask(*msg.Text)
			case "hint":
				if msg.Text == nil {
					emit(map[string]string{"error": "hint requires text"})
					continue
				}
				actor.Send(agent.Hint{Text: *msg.Text})
			case "interrupt":
				actor.Send(agent.Interrupt{})
			default:
				emit(map[string]string{"error": "unknown kind: " + *msg.Kind})
			}
			continue
		}

		if msg.User != nil {
			ask(*msg.User)
			continue
		}
		emit(map[string]string{"error": "no recognized field"})
	}

	turns.Wait()
	_ = fw.Shutdown()
	out.Flush()
}

type inputMsg struct {
	Kind *string `json:"kind"`
	Text *string `json:"text"`
	User *string `json:"user"`
}

// monitor measures the shape of one turn's stream.
type monitor struct {
	mu        sync.Mutex
	began     time.Time
	firstAt   time.Time
	counts    map[string]int
	chunkText strings.Builder
}

func (m *monitor) start() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.began = time.Now()
}

func (m *monitor) record(obs agent.Observation) {
	d, ok := obs.(agent.PartDelta)
	if !ok {
		return
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.firstAt.IsZero() {
		m.firstAt = time.Now()
	}
	if m.counts == nil {
		m.counts = map[string]int{}
	}
	m.counts[d.Kind.String()]++
	if d.Kind == agent.DeltaText {
		m.chunkText.WriteString(d.Chunk)
	}
}

func (m *monitor) report() map[string]any {
	m.mu.Lock()
	defer m.mu.Unlock()
	rep := map[string]any{
		"observation": "stream_stats",
		"deltas":      m.counts,
		// The concatenation of every text chunk. A caller that appends
		// chunks in arrival order must end up with the finished reply; if
		// this and the assistant line ever disagree, the stream is lying.
		"text": m.chunkText.String(),
	}
	if !m.began.IsZero() && !m.firstAt.IsZero() {
		rep["first_delta_ms"] = m.firstAt.Sub(m.began).Milliseconds()
	}
	return rep
}

func (m *monitor) reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.counts = map[string]int{}
	m.firstAt = time.Time{}
	m.chunkText.Reset()
}

func observationJSON(obs agent.Observation) map[string]any {
	switch v := obs.(type) {
	case agent.PartDelta:
		// kind is what makes a delta usable. Without it a reader cannot tell
		// reasoning from the reply, and dumping reasoning into the answer is
		// the first bug every streaming client writes.
		return map[string]any{
			"observation": "part_delta",
			"part_id":     v.PartID,
			"kind":        v.Kind.String(),
			"chunk":       v.Chunk,
		}
	case agent.PartFinal:
		m := map[string]any{"observation": "part_final", "part_id": v.PartID, "seq": v.Seq}
		switch p := v.Part.(type) {
		case agent.TextPart:
			m["text"] = p.Text
		case agent.ToolCallPart:
			m["tool"] = p.Name
			m["args"] = string(p.Args)
		}
		return m
	case agent.StateChanged:
		return map[string]any{"observation": "state_changed", "from": v.From.String(), "to": v.To.String()}
	case agent.TurnEnded:
		m := map[string]any{"observation": "turn_ended", "text": v.Text}
		if v.Err != "" {
			m["error"] = v.Err
		}
		return m
	}
	return map[string]any{"observation": "unknown"}
}

type observerFunc func(agent.Observation)

func (f observerFunc) Observe(o agent.Observation) { f(o) }

type cliHost struct{}

func (cliHost) Logf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
}
func (cliHost) APILogf(format string, args ...any) {}
func (cliHost) Debugf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
}

func envOr(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
