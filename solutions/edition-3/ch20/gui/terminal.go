package gui

import (
	"bufio"
	"context"
	"encoding/json"
	"ensemble"
	"fmt"
	"io"
	"strings"
)

// Terminal uses unbounded line input. The returned channel closes on EOF or
// /quit after all admitted input has completed; the application owns shutdown.
func (s *Server) Terminal(ctx context.Context, in io.Reader, out io.Writer) <-chan struct{} {
	done := make(chan struct{})
	c := s.Hub.connect()
	go func() {
		defer s.Hub.disconnect(c)
		view := terminalView{out: out, parts: map[string]terminalPart{}}
		for {
			select {
			case <-ctx.Done():
				return
			case <-c.done:
				return
			case raw := <-c.queue:
				view.render(raw)
			}
		}
	}()
	go func() {
		defer close(done)
		var pending []<-chan ensemble.Result
		// SubmitLive completion is reliable; observer frames and Running are not
		// completion acknowledgements. Keep reading so live hints still arrive.
		defer func() {
			for _, result := range pending {
				select {
				case reply := <-result:
					if reply.Err != nil {
						s.Hub.sendError(c, s.DefaultAgent, reply.Err)
					}
				case <-ctx.Done():
					return
				}
			}
		}()
		reader := bufio.NewReader(in)
		for {
			line, err := reader.ReadString('\n')
			line = strings.TrimSpace(line)
			if line == "/quit" || line == "/exit" {
				return
			}
			if line != "" {
				m := command{AgentID: s.DefaultAgent, Type: "input", Text: line}
				switch {
				case line == "/save":
					m.Type = "save"
				case line == "/pause":
					m.Type = "pause"
				case line == "/continue":
					m.Type = "unpause"
				case line == "/interrupt" || line == "/stop":
					m.Type = "interrupt"
				case line == "/speak on":
					m.Type = "speaking"
					m.Enabled = true
				case line == "/speak off":
					m.Type = "speaking"
				case strings.HasPrefix(line, "/hint "):
					m.Type = "hint"
					m.Text = strings.TrimPrefix(line, "/hint ")
				}
				if m.Type == "input" {
					a, ok := s.Root.Agent(m.AgentID)
					if !ok {
						s.Hub.sendError(c, m.AgentID, fmt.Errorf("unknown Agent %q", m.AgentID))
						continue
					}
					pending = append(pending, a.SubmitLive(ctx, m.Text))
				} else if cmdErr := s.command(c, m); cmdErr != nil {
					s.Hub.sendError(c, m.AgentID, cmdErr)
				}
			}
			if err != nil {
				if err != io.EOF {
					s.Hub.sendError(c, s.DefaultAgent, err)
				}
				s.Hub.trace("terminal_eof", s.DefaultAgent, "input ended; draining admitted work before shutdown")
				return
			}
		}
	}()
	return done
}

type terminalPart struct {
	text              string
	final, dispatched bool
	job               string
}
type terminalView struct {
	out     io.Writer
	parts   map[string]terminalPart
	current string
}

func (v *terminalView) render(raw []byte) {
	var w struct {
		Type    string          `json:"type"`
		AgentID string          `json:"agent_id"`
		PartID  uint64          `json:"part_id"`
		Data    json.RawMessage `json:"data"`
		Part    *Card           `json:"part"`
		Kind    string          `json:"kind"`
		Chunk   string          `json:"chunk"`
		Name    string          `json:"name"`
		State   *controlState   `json:"state"`
	}
	if json.Unmarshal(raw, &w) != nil {
		return
	}
	switch w.Type {
	case "replay":
		var s struct{ Events, Partials []json.RawMessage }
		if json.Unmarshal(w.Data, &s) == nil {
			for _, f := range s.Events {
				v.render(f)
			}
			for _, f := range s.Partials {
				v.render(f)
			}
		}
	case "part_delta", "part_final", "tool_dispatched", "tool_finished", "job_update":
		key := cardKey(w.AgentID, w.PartID)
		old, exists := v.parts[key]
		var c Card
		if w.Type == "part_delta" {
			c = Card{Kind: w.Kind, Name: w.Name, Role: "agent", Text: old.text + w.Chunk, Partial: true}
		} else if w.Part != nil {
			c = *w.Part
		} else {
			return
		}
		text := c.Text
		if c.Kind == "tool_call" && len(c.Args) > 0 {
			text = string(c.Args)
		}
		if w.Type != "job_update" && !(old.final && c.Partial) {
			if !exists || text != old.text {
				if v.current != key {
					fmt.Fprintf(v.out, "\n[%s · %s %s %s]\n", w.AgentID, c.Role, c.Kind, c.Name)
					v.current = key
				}
				if strings.HasPrefix(text, old.text) {
					fmt.Fprint(v.out, text[len(old.text):])
				} else {
					fmt.Fprintf(v.out, "\n[%s final]\n%s", key, text)
				}
				old.text = text
			}
			if !c.Partial && !old.final {
				fmt.Fprintln(v.out)
				old.final = true
				v.current = ""
			}
			if c.Dispatched && !old.dispatched {
				fmt.Fprintf(v.out, "[%s dispatched]\n", key)
				old.dispatched = true
				v.current = ""
			}
		}
		if c.Job != nil {
			state := fmt.Sprintf("%d/%s/%d/%v/%s", c.Job.Handle, c.Job.Status, c.Job.Bytes, c.Job.ExitCode, c.Job.Reason)
			if state != old.job {
				fmt.Fprintf(v.out, "\n[%s · job %d · %s · %d bytes · %s]\n", key, c.Job.Handle, c.Job.Status, c.Job.Bytes, c.Job.Reason)
				old.job = state
				v.current = ""
			}
		}
		v.parts[key] = old
	case "state_changed":
		if w.State == nil {
			return
		}
		fmt.Fprintf(v.out, "\n[%s · %s · paused=%t · speaking=%t]\n", w.AgentID, w.State.Turn, w.State.Paused, w.State.Speaking)
		v.current = ""
	case "error":
		var text string
		_ = json.Unmarshal(w.Data, &text)
		fmt.Fprintf(v.out, "\n[error] %s\n", text)
		v.current = ""
	}
}
