package gui

import (
	"encoding/json"
	"ensemble"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"
)

const clientQueue = 64

type controlState struct {
	Paused        bool               `json:"paused"`
	Speaking      bool               `json:"speaking"`
	Reason        string             `json:"reason,omitempty"`
	Turn          ensemble.TurnState `json:"turn"`
	Epoch         uint64             `json:"epoch"`
	ContinueEpoch uint64             `json:"continue_epoch"`
}

func controlsOf(s ensemble.ViewSnapshot) controlState {
	return controlState{s.Paused, s.Speaking, s.Reason, s.State, s.ControlEpoch, s.ContinueEpoch}
}

type agentView struct {
	meterModel         string
	meterRequests      int
	meterDigest        string
	ids                *displayIDs
	contextReplay      []ensemble.Event
	unsubscribeSkills  func()
	agent              ensemble.Agent
	cursor             ensemble.Seq
	revision, speechID uint64
	state              controlState
	partial            map[uint64]Card
	spoken             map[uint64]*speechText
	turn               uint64
}
type client struct {
	mcpMu sync.Mutex
	mcp   map[string]*browserTransport
	id    string
	queue chan []byte
	done  chan struct{}
	once  sync.Once
}

func (c *client) stop() { c.once.Do(func() { close(c.done) }) }

type Hub struct {
	selected    string     // empty is the explicitly global legacy/workbench mode.
	settingsMu  sync.Mutex // Serialize policy changes without holding mu across callbacks.
	mu          sync.Mutex
	en          *ensemble.Ensemble
	agents      map[string]*agentView
	clients     map[*client]bool
	nextClient  uint64
	session     string
	unsubscribe func()
	closed      bool
	speech      *speaker
	settings    Settings
	store       settingsStore
}

func NewHub(en *ensemble.Ensemble) *Hub {
	return newHub(en, DefaultSettings(), settingsStore{})
}

func newHub(en *ensemble.Ensemble, settings Settings, store settingsStore) *Hub {
	return newScopedHub(en, settings, store, "")
}
func (h *Hub) allows(id string) bool { return h.selected == "" || h.selected == id }
func newScopedHub(en *ensemble.Ensemble, settings Settings, store settingsStore, selected string) *Hub {
	h := &Hub{selected: selected, en: en, agents: map[string]*agentView{}, clients: map[*client]bool{}, session: fmt.Sprint(time.Now().UnixNano())}
	h.settings = settings
	h.store = store
	h.speech = newSpeaker(h)
	for _, a := range en.Agents() {
		if !h.allows(a.ID()) {
			continue
		}
		v := h.view(a)
		s := a.View()
		v.cursor = s.Through
		v.revision = s.Revision
		v.speechID = s.SpeechID
		v.state = controlsOf(s)
		v.partial = partialCards(s, v.ids)
		v.turn = s.TurnID
	}
	h.unsubscribe = en.SubscribeIdentified(h.observe)
	return h
}
func (h *Hub) Close() {
	h.unsubscribe()
	h.mu.Lock()
	defer h.mu.Unlock()
	h.closed = true
	for _, v := range h.agents {
		if v.unsubscribeSkills != nil {
			v.unsubscribeSkills()
		}
	}
	h.speech.close()
	for c := range h.clients {
		c.stop()
	}
	h.clients = map[*client]bool{}
}
func (h *Hub) trace(kind, agent, detail string) {
	h.en.Trace(ensemble.TraceRecord{Stage: "gui_" + kind, AgentID: agent, Detail: detail})
}
func (h *Hub) view(a ensemble.Agent) *agentView {
	v := h.agents[a.ID()]
	if v == nil {

		cfg := a.Config()
		v = &agentView{agent: a, state: controlState{Paused: cfg.PauseOnStart, Speaking: cfg.Speaking, Turn: ensemble.Idle}, partial: map[uint64]Card{}, spoken: map[uint64]*speechText{}}
		all := contextEvents(a)
		v.ids = newDisplayIDs(all)
		first := a.Log().LastSeq() + 1
		if events := a.Log().Events(); len(events) > 0 {
			first = events[0].Seq
		}
		for _, e := range all {
			if e.Seq < first {
				v.contextReplay = append(v.contextReplay, e)
			}
		}
		h.agents[a.ID()] = v
		v.unsubscribeSkills = a.Skills().SubscribeSkills(func() {
			h.mu.Lock()
			defer h.mu.Unlock()
			if h.closed {
				return
			}
			h.broadcast(skillFrame(a))
		})
	}
	return v
}
func partialCards(s ensemble.ViewSnapshot, ids ...*displayIDs) map[uint64]Card {
	parts := make(map[uint64]Card, len(s.Partials))
	for _, p := range s.Partials {
		id := p.PartID
		if len(ids) > 0 {
			id = ids[0].partial(id)
		}
		parts[id] = Card{Kind: string(p.Kind), Role: "agent", Name: p.Name, Text: p.Chunk, Partial: true, TurnID: p.TurnID}
	}
	return parts
}

// sync consumes exactly one immutable saved checkpoint. No view callback runs
// on the actor, and a dropped/coalesced wakeup cannot lose content or a final.
func (h *Hub) sync(v *agentView, s ensemble.ViewSnapshot) {
	id := v.agent.ID()
	meter := v.agent.Engine().Status()
	digest := ""
	if meter.Cache.Current != nil {
		digest = meter.Cache.Current.Digest
	}
	if meter.Requests != v.meterRequests || digest != v.meterDigest || meter.SelectedModel != v.meterModel {
		h.broadcast(usageWire(id, meter))
		v.meterRequests, v.meterDigest, v.meterModel = meter.Requests, digest, meter.SelectedModel
	}
	if s.Revision == v.revision {
		return
	}
	old := v.state
	v.state = controlsOf(s)
	v.turn = s.TurnID
	if old != v.state {
		h.broadcast(Wire{Type: "state_changed", AgentID: id, From: old.Turn, To: v.state.Turn, State: &v.state})
	}
	if old.Epoch != v.state.Epoch {
		h.speech.cancelAgent(id, old.Speaking && !v.state.Speaking)
	}
	for _, e := range v.agent.Log().EventsRange(v.cursor, s.Through) {
		for _, f := range v.ids.project(id, e) {
			if f.Type == "part_final" && f.Part != nil && f.Part.Role == "agent" && (f.Part.Kind == "text" || f.Part.Kind == "reasoning_summary") {
				// A coalesced checkpoint may already contain the final and no
				// partial. Deliver its unseen suffix first under the same ID.
				seen := v.partial[f.PartID].Text
				if strings.HasPrefix(f.Part.Text, seen) && len(f.Part.Text) > len(seen) {
					h.broadcast(Wire{Type: "part_delta", AgentID: id, PartID: f.PartID, TurnID: f.Part.TurnID, Kind: f.Part.Kind, Chunk: f.Part.Text[len(seen):]})
				}
			}
			if f.Type != "job_update" {
				if v.state.Speaking && (f.Part != nil || f.Type == "error") {
					h.flushBefore(v, f.PartID)
				}
				delete(v.partial, f.PartID)
			}
			h.broadcast(f)
			if v.state.Speaking && f.Type != "job_update" {
				if f.Part != nil {
					h.narrate(v, f.PartID, *f.Part, true)
				} else if f.Type == "error" {
					if text, ok := f.Data.(string); ok {
						h.speech.enqueue(id, spokenPart{Text: cleanSpeech(text), Source: "error"}, "automatic", e.TurnID)
					}
				}
			}
		}
		if e.Type == ensemble.TurnEnded || e.Type == ensemble.TurnInterrupted {
			v.spoken = map[uint64]*speechText{}
		}
	}
	v.cursor = s.Through
	next := partialCards(s, v.ids)
	for id := range v.partial {
		if _, ok := next[id]; !ok {
			h.broadcast(Wire{Type: "part_removed", AgentID: v.agent.ID(), PartID: id})
		}
	}
	ids := make([]uint64, 0, len(next))
	for id := range next {
		ids = append(ids, id)
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
	for index, part := range ids {
		c := next[part]
		old, ok := v.partial[part]
		if !ok || old.Text != c.Text || old.Name != c.Name {
			h.broadcast(Wire{Type: "part_delta", AgentID: id, PartID: part, TurnID: c.TurnID, Kind: c.Kind, Name: c.Name, Chunk: c.Text[len(old.Text):]})
		}
		// Arrival of a later part closes the previous source even if its bytes are unchanged.
		if v.state.Speaking {
			h.narrate(v, part, c, index < len(ids)-1)
		}
	}
	v.partial = next
	v.revision = s.Revision
	if s.SpeechID > v.speechID {
		v.speechID = s.SpeechID
		if s.SpeechEpoch == s.ControlEpoch {
			h.speech.request(id, s.Speech)
		}
	}
}
func (h *Hub) observe(p ensemble.AgentObservation) {
	if !h.allows(p.Agent.ID()) {
		return
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	if !h.closed {
		h.sync(h.view(p.Agent), p.Agent.View())
	}
}

// Flush older retained tails before an event changes source. The existing raw
// cursor makes a later final publication idempotent; zero means an error boundary.
func (h *Hub) flushBefore(v *agentView, next uint64) {
	ids := make([]uint64, 0, len(v.partial))
	for id := range v.partial {
		if next == 0 || id < next {
			ids = append(ids, id)
		}
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
	for _, id := range ids {
		h.narrate(v, id, v.partial[id], true)
	}
}
func (h *Hub) narrate(v *agentView, id uint64, c Card, final bool) {
	// Provider failures already have a dedicated ErrorOccurred speech path.
	if c.Kind == "error" && c.Role != "tool" {
		return
	}
	if c.Text == "" || strings.HasPrefix(c.Text, "[opaque") {
		return
	}
	if c.Kind != "text" && c.Kind != "thinking" && c.Kind != "reasoning_summary" && c.Kind != "tool_call" && !(c.Kind == "tool_result" && c.IsError) {
		return
	}
	if c.Role != "agent" && c.Role != "human" && !(c.Role == "tool" && c.IsError) {
		return
	}
	if c.Kind == "tool_call" && (!final || !c.Dispatched) {
		return
	}
	state := v.spoken[id]
	if state == nil {
		state = &speechText{}
		v.spoken[id] = state
	}
	text := c.Text
	// Dispatch is an announcement, never automatic narration of arguments/results.
	if c.Kind == "tool_call" {
		text = c.Name
	}
	if spoken := state.update(text, final); spoken != "" {
		reason, source := "automatic", speechSource(c)
		if c.Role == "human" {
			reason, source = "prompt", "human"
		}
		h.speech.enqueue(v.agent.ID(), spokenPart{Text: spoken, Source: source, PartID: id, TurnID: c.TurnID}, reason, c.TurnID)
	}
}
func abbreviate(s string, n int) string {
	r := []rune(s)
	if len(r) > n {
		return string(r[:n]) + "…"
	}
	return s
}
func (h *Hub) broadcast(w Wire) {
	if w.AgentID != "" && !h.allows(w.AgentID) {
		return
	}
	if w.Question != nil && !h.allows(w.Question.Sender) {
		return
	}
	raw, err := json.Marshal(w)
	if err != nil {
		h.trace("encode_error", w.AgentID, err.Error())
		return
	}
	for c := range h.clients {
		h.enqueue(c, raw, w.Type, w.AgentID)
	}
}
func (h *Hub) enqueue(c *client, raw []byte, kind, agent string) {
	select {
	case c.queue <- raw:
	default:
		h.trace("slow_client_disconnect", agent, c.id)
		delete(h.clients, c)
		c.stop()
	}
}

type replayAgent struct {
	Meter  ensemble.EngineStatus `json:"meter"`
	ID     string                `json:"id"`
	Config struct {
		Vendor      ensemble.Vendor   `json:"vendor"`
		Model       string            `json:"model"`
		Surface     ensemble.Surface  `json:"surface"`
		BillingMode ensemble.AuthMode `json:"billing_mode,omitempty"`
	} `json:"config"`
	State controlState `json:"state"`
}
type replayEnvelope struct {
	Session  string        `json:"session"`
	Client   string        `json:"client"`
	Agents   []replayAgent `json:"agents"`
	Events   []Wire        `json:"events"`
	Partials []Wire        `json:"partials"`
}

func (h *Hub) connect() *client {
	h.mu.Lock()
	defer h.mu.Unlock()
	snapshots := map[string]ensemble.ViewSnapshot{}
	for _, a := range h.en.Agents() {
		if !h.allows(a.ID()) {
			continue
		}
		snapshots[a.ID()] = a.View()
	}
	return h.connectAt(snapshots)
}

// connectAt consumes the already captured checkpoint under the hub lock.
func (h *Hub) connectAt(snapshots map[string]ensemble.ViewSnapshot) *client {
	visible := map[string]ensemble.ViewSnapshot{}
	for id, s := range snapshots {
		if h.allows(id) {
			if _, ok := h.en.Agent(id); ok {
				visible[id] = s
			}
		}
	}
	snapshots = visible
	h.nextClient++
	c := &client{id: fmt.Sprint(h.nextClient), queue: make(chan []byte, clientQueue), done: make(chan struct{})}
	if h.closed {
		c.stop()
		return c
	}
	for id, s := range snapshots {
		if a, ok := h.en.Agent(id); ok {
			h.sync(h.view(a), s)
		}
	}
	data := replayEnvelope{Session: h.session, Client: c.id, Agents: []replayAgent{}, Events: []Wire{}, Partials: []Wire{}}
	ids := make([]string, 0, len(snapshots))
	for id := range snapshots {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	for _, id := range ids {
		v := h.agents[id]
		s := snapshots[id]
		replay := append([]ensemble.Event(nil), v.contextReplay...)
		replay = append(replay, v.agent.Log().RecentThrough(s.Through, time.Now().Add(-time.Hour), 100, eventVisible)...)
		for _, e := range replay {
			frames := v.ids.project(id, e)
			for i := range frames {
				frames[i].Replay = true
			}
			data.Events = append(data.Events, frames...)
		}
		cfg := v.agent.Config()
		a := replayAgent{ID: id, State: controlsOf(s), Meter: v.agent.Engine().Status()}
		a.Config.Vendor = cfg.Vendor
		a.Config.Model = cfg.Model
		a.Config.Surface = cfg.Surface
		a.Config.BillingMode = cfg.AuthMode
		data.Agents = append(data.Agents, a)
		data.Events = append(data.Events, skillFrame(v.agent))
		initialUsage := usageWire(id, a.Meter)
		initialUsage.Replay = true
		data.Events = append(data.Events, initialUsage)
		data.Events = append(data.Events, Wire{Type: "state_changed", AgentID: id, Replay: true, From: s.State, To: s.State, State: &a.State})
		for _, p := range s.Partials {
			data.Partials = append(data.Partials, Wire{Type: "part_delta", AgentID: id, PartID: v.ids.partial(p.PartID), TurnID: p.TurnID, Kind: string(p.Kind), Name: p.Name, Chunk: p.Chunk, Replay: true})
		}
	}
	raw, err := json.Marshal(Wire{Type: "replay", Data: data, Replay: true})
	if err != nil {
		c.stop()
		return c
	}
	c.queue <- raw
	settings := h.currentSettings()
	settingsRaw, _ := json.Marshal(Wire{Type: "current_settings", Settings: &settings})
	c.queue <- settingsRaw
	h.clients[c] = true
	h.trace("client_connect", "", c.id)
	return c
}
func (h *Hub) disconnect(c *client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.clients, c)
	c.stop()
	h.speech.cancelClient(c.id)
}
func (h *Hub) sendError(c *client, agent string, err error) {
	h.mu.Lock()
	defer h.mu.Unlock()
	b, _ := json.Marshal(Wire{Type: "error", AgentID: agent, Data: err.Error()})
	h.enqueue(c, b, "error", agent)
}
func (h *Hub) emit(w Wire) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if !h.closed {
		h.broadcast(w)
	}
}

func (h *Hub) sendSettingsError(c *client, err error) {
	h.mu.Lock()
	defer h.mu.Unlock()
	raw, _ := json.Marshal(Wire{Type: "settings_error", Data: err.Error()})
	h.enqueue(c, raw, "settings_error", "")
}

func skillFrame(a ensemble.Agent) Wire {
	frame := Wire{Type: "skill_state", AgentID: a.ID(), Skills: a.Skills().ListSkills(), LoadableSkills: a.Skills().LoadableSkills()}
	// Publish only tunnel eligibility, not endpoint commands or addresses.
	// Canonical skill declarations remain unchanged for display and admission.
	for _, skill := range frame.Skills {
		if skill.PendingUnload {
			continue
		}
		for _, declaration := range skill.MCPServers {
			if browserWebSocket(a, declaration) {
				frame.BrowserMCP = true
				return frame
			}
		}
	}
	return frame
}

func (h *Hub) refreshStatus(a ensemble.Agent) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if !h.closed {
		h.broadcast(usageWire(a.ID(), a.Engine().Status()))
	}
}

// PresentHumanQuestion is presentation only. Applications explicitly install it
// with Ensemble.SetHumanQuestionSink; replies remain correlated by Ensemble.
func (h *Hub) PresentHumanQuestion(question ensemble.Question) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if !h.closed {
		h.broadcast(Wire{Type: "human_question", Question: &question})
	}
}
