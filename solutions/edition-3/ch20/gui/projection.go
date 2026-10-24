package gui

import (
	"encoding/json"
	"fmt"
	"strings"

	"ensemble"
)

// Wire is a display projection, never a serialization of replay-only Part data.
type Wire struct {
	Question       *ensemble.Question     `json:"question,omitempty"`
	Usage          *UsageFrame            `json:"-"`
	Meter          *ensemble.EngineStatus `json:"meter,omitempty"`
	BrowserMCP     bool                   `json:"browser_mcp,omitempty"`
	Skills         []ensemble.SkillInfo   `json:"skills,omitempty"`
	LoadableSkills []ensemble.SkillInfo   `json:"loadable_skills,omitempty"`
	Payload        json.RawMessage        `json:"payload,omitempty"`
	Type           string                 `json:"type"`
	Settings       *Settings              `json:"settings,omitempty"`
	AgentID        string                 `json:"agent_id"`
	PartID         uint64                 `json:"part_id,omitempty"`
	TurnID         uint64                 `json:"turn_id,omitempty"`
	Kind           string                 `json:"kind,omitempty"`
	Chunk          string                 `json:"chunk"` // Empty starts still carry the public chunk field.
	Name           string                 `json:"name,omitempty"`
	Part           *Card                  `json:"part,omitempty"`
	From           ensemble.TurnState     `json:"from,omitempty"`
	To             ensemble.TurnState     `json:"to,omitempty"`
	State          *controlState          `json:"state,omitempty"`
	CallID         string                 `json:"call_id,omitempty"`
	Input          json.RawMessage        `json:"input,omitempty"`
	Result         string                 `json:"result,omitempty"`
	IsError        *bool                  `json:"is_error,omitempty"`
	Text           *string                `json:"text,omitempty"`
	Seq            ensemble.Seq           `json:"-"`
	Data           any                    `json:"data,omitempty"` // replay/speech extensions, not the public part protocol
	Replay         bool                   `json:"replay,omitempty"`
}
type Card struct {
	TurnID     uint64               `json:"turn_id,omitempty"`
	Job        *ensemble.JobData    `json:"job,omitempty"`
	CallPartID uint64               `json:"call_part_id,omitempty"`
	Kind       string               `json:"type"`
	Role       string               `json:"role,omitempty"`
	Text       string               `json:"text"`
	Name       string               `json:"name,omitempty"`
	CallID     string               `json:"call_id,omitempty"`
	Args       json.RawMessage      `json:"args,omitempty"`
	MIME       string               `json:"mime,omitempty"`
	Locator    string               `json:"locator,omitempty"`
	IsError    bool                 `json:"is_error,omitempty"`
	Dispatched bool                 `json:"dispatched,omitempty"`
	Partial    bool                 `json:"partial,omitempty"`
	From       *ensemble.Provenance `json:"from,omitempty"`
}

func readable(parts ensemble.PartList) string {
	var out []string
	for _, p := range parts {
		c := project(p)
		if c.Kind == "text" || c.Kind == "thinking" || c.Kind == "reasoning_summary" || c.Kind == "tool_result" {
			if c.Text != "" && !strings.HasPrefix(c.Text, "[opaque") {
				out = append(out, c.Text)
			}
		}
	}
	return strings.Join(out, "\n")
}
func project(p ensemble.Part) Card {
	c := Card{Kind: p.Type, Text: p.Text, Name: p.Name, CallID: p.CallID, MIME: p.MIME, IsError: p.IsError, From: p.From}
	switch p.Type {
	case "text", "thinking", "reasoning_summary":
	case "tool_call":
		c.Text = p.Name + "\n" + string(p.Args)
		c.Args = append(json.RawMessage(nil), p.Args...)
	case "tool_result":
		c.Text = readable(p.Parts)
	case "image", "audio", "video", "file":
		if p.Ref != nil {
			c.Locator = p.Ref.Locator
		}
		c.Text = p.Stub
	case "opaque":
		if p.From != nil && p.From.Surface == ensemble.Responses {
			c.Kind = "continuation"
			c.Text = "[private continuation retained]"
			break
		}
		// Only known readable reasoning fields cross the presentation boundary.
		var v struct {
			Type      string `json:"type"`
			Thinking  string `json:"thinking"`
			Text      string `json:"text"`
			Thought   bool   `json:"thought"`
			Reasoning string `json:"reasoning"`
			Content   string `json:"reasoning_content"`
		}
		_ = json.Unmarshal(p.Data, &v)
		c.Kind = "thinking"
		switch {
		case v.Type == "thinking":
			c.Text = v.Thinking
		case v.Thought:
			c.Text = v.Text
		case v.Reasoning != "":
			c.Text = v.Reasoning
		case v.Content != "":
			c.Text = v.Content
		default:
			c.Text = "[opaque or redacted reasoning retained privately]"
		}
	default:
		c.Text = p.Stub
	}
	return c
}

// Eligibility belongs to the view, not History's selection/serialization owner.
// State/turn fences are replayed but do not count as renderable content events.
func eventVisible(e ensemble.Event) bool {
	if e.Stale {
		return false
	}
	switch e.Type {
	case ensemble.MessageReceived, ensemble.HintReceived:
		return e.Message != nil && len(e.Message.Parts) > 0
	case ensemble.ResponseEnded:
		return e.Response != nil && len(e.Response.Parts) > 0
	case ensemble.ToolCalled, ensemble.ToolReleased, ensemble.ToolReturned:
		return e.Tool != nil
	case ensemble.JobReported, ensemble.JobKilled:
		return e.Job != nil
	case ensemble.ErrorOccurred:
		return e.Error != nil
	}
	return false
}
func projectEvent(agent string, e ensemble.Event) []Wire {
	if e.Type == ensemble.TurnInterrupted && !e.Stale {
		return []Wire{{Type: "turn_interrupted", AgentID: agent, TurnID: e.TurnID, Seq: e.Seq}}
	}
	if e.Type == ensemble.TurnEnded && !e.Stale {
		text := ""
		if e.Message != nil {
			text = readable(e.Message.Parts)
		}
		return []Wire{{Type: "turn_ended", AgentID: agent, TurnID: e.TurnID, Text: &text, Seq: e.Seq}}
	}
	if !eventVisible(e) {
		return nil
	}
	var frames []Wire
	add := func(p ensemble.Part, role string) {
		c := project(p)
		c.Role = role
		c.TurnID = e.TurnID
		frames = append(frames, Wire{Type: "part_final", AgentID: agent, PartID: p.DisplayID, TurnID: e.TurnID, Seq: e.Seq, Part: &c})
	}
	switch e.Type {
	case ensemble.MessageReceived, ensemble.HintReceived:
		for _, p := range e.Message.Parts {
			add(p, string(e.Message.Actor))
		}
	case ensemble.ResponseEnded:
		for _, p := range e.Response.Parts {
			add(p, "agent")
		}
	case ensemble.ToolCalled, ensemble.ToolReleased:
		d := e.Tool
		c := Card{Kind: "tool_call", Role: "agent", Name: d.Name, Text: d.Name + "\n" + string(d.Args), CallID: d.CallID, Args: d.Args, Dispatched: e.Type == ensemble.ToolReleased && d.Dispatched, Job: d.Job, CallPartID: d.CallPartID, TurnID: e.TurnID}
		f := Wire{Type: "part_final", AgentID: agent, PartID: d.DisplayID, TurnID: e.TurnID, Seq: e.Seq, Part: &c}
		if c.Dispatched {
			f.Type = "tool_dispatched"
			f.CallID = d.CallID
			f.Name = d.Name
			f.Input = d.Args
		}
		frames = append(frames, f)
	case ensemble.ToolReturned:
		d := e.Tool
		add(ensemble.Part{DisplayID: d.DisplayID, Type: "tool_result", Name: d.Name, CallID: d.CallID, Parts: d.Parts, IsError: d.IsError}, "tool")
		f := &frames[len(frames)-1]
		f.Type = "tool_finished"
		f.Part.Job = d.Job
		f.Part.CallPartID = d.CallPartID
		f.CallID = d.CallID
		f.Name = d.Name
		f.Result = f.Part.Text
		isError := d.IsError
		f.IsError = &isError
	case ensemble.JobReported, ensemble.JobKilled:
		j := e.Job
		c := Card{Kind: "tool_call", Role: "agent", Name: j.Name, CallID: j.CallID, Args: j.Args, Text: j.Name + "\n" + string(j.Args), TurnID: j.TurnID, CallPartID: j.CallPartID, Job: j}
		frames = append(frames, Wire{Type: "job_update", AgentID: agent, PartID: j.CallPartID, TurnID: j.TurnID, Seq: e.Seq, Part: &c})
	case ensemble.ErrorOccurred:
		frames = append(frames, Wire{Type: "error", AgentID: agent, TurnID: e.TurnID, Seq: e.Seq, Data: e.Error.Message})
	}
	return frames
}
func cardKey(agent string, part uint64) string { return fmt.Sprintf("%s/%d", agent, part) }

// UsageFrame is a projection of one Engine snapshot, not GUI-owned accounting.
// Latest disjoint counts are flat; total holds the process-session counts.
// Tool-call Wire.Input remains untouched.
type UsageFrame struct {
	Input      int64          `json:"input"`
	Output     int64          `json:"output"`
	CacheRead  int64          `json:"cache_read"`
	CacheWrite int64          `json:"cache_write"`
	Total      ensemble.Usage `json:"total"`
	HitRate    *float64       `json:"hit_rate"`
	CostUSD    *float64       `json:"cost_usd"`
	Priced     bool           `json:"priced"`
}

func usageWire(id string, s ensemble.EngineStatus) Wire {
	last := s.Last
	rate := s.SessionHitRate
	// Before any response these are initial UI zeroes, not fabricated measurements.
	if s.Requests == 0 {
		last = ensemble.Usage{}
		zero := 0.0
		rate = &zero
	}
	u := &UsageFrame{Input: last.Input, Output: last.Output, CacheRead: last.CacheRead, CacheWrite: last.CacheWrite, Total: s.Session, HitRate: rate, Priced: s.Cost.Status == "known"}
	if u.Priced {
		cost := s.Cost.USD
		u.CostUSD = &cost
	}
	return Wire{Type: "usage", AgentID: id, Usage: u, Meter: &s}
}
func (w Wire) MarshalJSON() ([]byte, error) {
	type plain Wire
	if w.Type != "usage" {
		return json.Marshal(plain(w))
	}
	return json.Marshal(struct {
		Type    string `json:"type"`
		AgentID string `json:"agent_id"`
		Replay  bool   `json:"replay,omitempty"`
		*UsageFrame
		Meter *ensemble.EngineStatus `json:"meter"`
	}{w.Type, w.AgentID, w.Replay, w.Usage, w.Meter})
}
