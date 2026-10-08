package gui

import (
	"bytes"
	"context"
	"encoding/json"
	"example.com/ensemble-gui/internal/common"
	"io"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"example.com/ensemble"
	"github.com/gorilla/websocket"
)

type outgoing struct {
	data    []byte
	message any
}

// Connector owns one socket's reader, writer, subscription and pause registration.
// The queue is never closed: a separate canceled context ends every producer.
type Connector struct {
	preferences         common.PreferencesWatch
	preferencesRevision uint64
	preferencesAdvanced chan struct{}
	snapshotDone        chan struct{}
	parent              ServerOwner
	id                  string
	socket              *websocket.Conn
	ctx                 context.Context
	cancel              context.CancelFunc
	once                sync.Once
	done                chan struct{}
	workers             sync.WaitGroup
	mu                  sync.Mutex
	queue               []outgoing
	deferred            []outgoing
	snapshotPending     bool
	bytes               int
	wake, space         chan struct{}
	watch               ensemble.Watch
	pause               ensemble.PauseRegistration
	generation          string
	revision            uint64
	advanced            chan struct{}
	ids                 map[string]bool
}

func NewConnector(parent ServerOwner, id string, socket *websocket.Conn) *Connector {
	ctx, cancel := context.WithCancel(context.Background())
	return &Connector{preferencesAdvanced: make(chan struct{}), snapshotDone: make(chan struct{}), parent: parent, id: id, socket: socket, ctx: ctx, cancel: cancel, done: make(chan struct{}), wake: make(chan struct{}, 1), space: make(chan struct{}, 1), advanced: make(chan struct{}), ids: map[string]bool{}}
}
func (c *Connector) Server() ServerOwner   { return c.parent }
func (c *Connector) Done() <-chan struct{} { return c.done }
func (c *Connector) Close()                { c.stop("connection closed") }
func (c *Connector) stop(reason string) {
	c.once.Do(func() {
		c.cancel()
		c.parent.Trace(c.id, "local", "closed", map[string]string{"reason": reason})
		_ = c.socket.WriteControl(websocket.CloseMessage, websocket.FormatCloseMessage(1008, reason), time.Now().Add(100*time.Millisecond))
		_ = c.socket.Close()
		c.mu.Lock()
		watch, pause := c.watch, c.pause
		preferences := c.preferences
		c.mu.Unlock()
		if preferences != nil {
			preferences.Close()
		}
		if watch != nil {
			watch.Close()
		}
		if pause != nil {
			pause.Close()
		}
	})
}
func (c *Connector) send(message any, wait bool) bool {
	data, err := json.Marshal(message)
	if err != nil {
		c.stop("cannot encode message")
		return false
	}
	for {
		c.mu.Lock()
		if c.ctx.Err() != nil {
			c.mu.Unlock()
			return false
		}
		fits := len(data) <= ensemble.WatchBytes-c.bytes && len(c.queue)+len(c.deferred) < ensemble.WatchItems
		if fits {
			if !wait && c.snapshotPending {
				c.deferred = append(c.deferred, outgoing{data, message})
			} else {
				c.queue = append(c.queue, outgoing{data, message})
			}
			c.bytes += len(data)
			c.parent.Trace(c.id, "server_to_client", "queued", message)
			c.mu.Unlock()
			select {
			case c.wake <- struct{}{}:
			default:
			}
			return true
		}
		deferredBlocked := len(c.deferred) > 0
		c.mu.Unlock()
		if !wait || len(data) > ensemble.WatchBytes || deferredBlocked {
			c.stop("outgoing overflow; resync required")
			return false
		}
		select {
		case <-c.ctx.Done():
			return false
		case <-c.space:
		}
	}
}
func (c *Connector) writer() {
	defer c.workers.Done()
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-c.ctx.Done():
			return
		case <-ticker.C:
			if c.socket.WriteControl(websocket.PingMessage, nil, time.Now().Add(5*time.Second)) != nil {
				c.stop("ping failed; resync required")
				return
			}
		case <-c.wake:
			for {
				c.mu.Lock()
				if len(c.queue) == 0 {
					c.mu.Unlock()
					break
				}
				item := c.queue[0]
				c.queue[0] = outgoing{}
				c.queue = c.queue[1:]
				c.bytes -= len(item.data)
				c.mu.Unlock()
				select {
				case c.space <- struct{}{}:
				default:
				}
				_ = c.socket.SetWriteDeadline(time.Now().Add(5 * time.Second))
				if c.socket.WriteMessage(websocket.TextMessage, item.data) != nil {
					c.stop("write failed; resync required")
					return
				}
				c.parent.Trace(c.id, "server_to_client", "written", item.message)
			}
		}
	}
}
func (c *Connector) Run() {
	defer close(c.done)
	defer func() { c.Close(); c.workers.Wait() }()
	// Bound the accumulated payload ourselves. Gorilla's header-level read limit
	// sends its own empty-reason close before we can explain an oversized command.
	_ = c.socket.SetReadDeadline(time.Now().Add(25 * time.Second))
	c.socket.SetPongHandler(func(string) error { return c.socket.SetReadDeadline(time.Now().Add(25 * time.Second)) })
	c.workers.Add(1)
	go c.writer()
	for {
		kind, reader, err := c.socket.NextReader()
		if err != nil {
			return
		}
		data, err := io.ReadAll(io.LimitReader(reader, 65537))
		if err != nil {
			return
		}
		if len(data) > 65536 {
			c.stop("message exceeds 65536 bytes")
			return
		}
		if kind != websocket.TextMessage || !utf8.Valid(data) {
			c.stop("expected UTF-8 JSON text")
			return
		}
		var fields map[string]json.RawMessage
		if json.Unmarshal(data, &fields) != nil || fields == nil {
			c.stop("malformed JSON object")
			return
		}
		c.parent.Trace(c.id, "client_to_server", "received", json.RawMessage(data))
		if err := commandObject(c, data); err != nil {
			var id string
			if json.Unmarshal(fields["id"], &id) != nil || strings.TrimSpace(id) == "" {
				c.stop("id must be a nonempty string")
			} else {
				c.refuse(id, err.Error())
			}
			continue
		}
		c.command(fields)
	}
}
func (c *Connector) refuse(id, message string) {
	c.send(map[string]any{"type": "error", "id": id, "code": "invalid_command", "message": message}, false)
}
func (c *Connector) command(f map[string]json.RawMessage) {
	var id, kind, text string
	if json.Unmarshal(f["id"], &id) != nil || strings.TrimSpace(id) == "" {
		c.stop("id must be a nonempty string")
		return
	}
	if len(c.ids) >= 4096 {
		c.stop("command limit reached; reconnect required")
		return
	}
	if c.ids[id] {
		c.refuse(id, "duplicate command id")
		return
	}
	if json.Unmarshal(f["type"], &kind) != nil {
		c.refuse(id, "type must be a command name")
		return
	}
	allowed := map[string]bool{"id": true, "type": true}
	switch kind {
	case "prompt", "hint":
		allowed["text"] = true
	case "pause":
		allowed["typing"] = true
		allowed["speaking"] = true
	case "preferences_update", "policy_update":
		allowed["base_revision"] = true
		allowed["patch"] = true
	case "subscribe", "interrupt":
	default:
		c.refuse(id, "unknown command")
		return
	}
	for k := range f {
		if !allowed[k] {
			c.refuse(id, "unknown command field")
			return
		}
	}
	if len(f) != len(allowed) {
		c.refuse(id, "missing command field")
		return
	}
	if kind == "prompt" || kind == "hint" {
		if string(f["text"]) == "null" || json.Unmarshal(f["text"], &text) != nil || strings.TrimSpace(text) == "" {
			c.refuse(id, "text must be nonempty")
			return
		}
	}
	var typing, speaking bool
	if kind == "pause" {
		if string(f["typing"]) == "null" || string(f["speaking"]) == "null" || json.Unmarshal(f["typing"], &typing) != nil || json.Unmarshal(f["speaking"], &speaking) != nil {
			c.refuse(id, "pause requires two Booleans")
			return
		}
	}
	var base uint64
	if kind == "preferences_update" || kind == "policy_update" {
		if bytes.Equal(bytes.TrimSpace(f["base_revision"]), []byte("null")) || json.Unmarshal(f["base_revision"], &base) != nil {
			c.refuse(id, "base_revision must be a nonnegative integer")
			return
		}
		var patch map[string]json.RawMessage
		if json.Unmarshal(f["patch"], &patch) != nil || len(patch) == 0 {
			c.refuse(id, "patch must be a nonempty object")
			return
		}
	}
	c.ids[id] = true
	c.mu.Lock()
	subscribed := c.watch != nil
	c.mu.Unlock()
	if kind != "subscribe" && !subscribed {
		c.refuse(id, "subscribe first")
		return
	}
	owner := c.parent.Ensemble()
	agent := c.parent.AgentID()
	switch kind {
	case "subscribe":
		if subscribed {
			c.refuse(id, "already subscribed")
			return
		}
		ps, pw, err := c.parent.Preferences().Subscribe()
		if err != nil {
			c.settingsError(id, err)
			return
		}
		p, err := owner.RegisterPause(agent)
		if err != nil {
			pw.Close()
			c.refuse(id, err.Error())
			return
		}
		s, w, err := owner.Watch(agent)
		if err != nil {
			p.Close()
			pw.Close()
			c.refuse(id, err.Error())
			return
		}
		c.mu.Lock()
		if c.ctx.Err() != nil {
			c.mu.Unlock()
			p.Close()
			pw.Close()
			w.Close()
			return
		}
		c.snapshotPending = true
		c.preferences = pw
		c.pause = p
		c.watch = w
		c.generation = s.Generation
		c.mu.Unlock()
		c.workers.Add(1)
		go c.deliverSnapshot(id, s, w, ps)
	case "preferences_update", "policy_update":
		c.workers.Add(1)
		go c.updateSettings(id, kind, base, append(json.RawMessage(nil), f["patch"]...))
	case "pause":
		c.mu.Lock()
		p := c.pause
		c.mu.Unlock()
		a, err := p.Update(typing, speaking)
		if err != nil {
			c.refuse(id, err.Error())
			return
		}
		c.send(map[string]any{"type": "ack", "id": id, "revision": a.Revision, "paused": a.Paused, "typing_clients": a.TypingClients, "speaking_clients": a.SpeakingClients}, false)
	case "prompt":
		h, err := owner.SubmitPrompt(agent, text)
		if err != nil {
			c.refuse(id, err.Error())
			return
		}
		if !c.send(map[string]any{"type": "accepted", "id": id, "request_id": h.ID()}, false) {
			return
		}
		c.workers.Add(1)
		go c.completion(h)
	case "hint":
		a, err := owner.Hint(agent, text)
		if err != nil {
			c.refuse(id, err.Error())
			return
		}
		c.send(map[string]any{"type": "ack", "id": id, "request_id": a.RequestID, "seq": a.Seq, "sent": a.Sent}, false)
	case "interrupt":
		a, err := owner.Interrupt(agent)
		if err != nil {
			c.refuse(id, err.Error())
			return
		}
		out := map[string]any{"type": "ack", "id": id, "accepted": a.Interrupted}
		if a.RequestID != "" {
			out["request_id"] = a.RequestID
		}
		c.send(out, false)
	}
}
func (c *Connector) advance(revision uint64) {
	c.mu.Lock()
	c.revision = revision
	close(c.advanced)
	c.advanced = make(chan struct{})
	c.mu.Unlock()
}
func (c *Connector) deliverSnapshot(id string, s ensemble.WatchSnapshot, w ensemble.Watch, ps common.PreferencesSnapshot) {
	defer c.workers.Done()
	// Loss while a large snapshot is sending invalidates the entire generation.
	lost := make(chan struct{})
	monitorDone := make(chan struct{})
	defer func() { close(lost); <-monitorDone }()
	go func() {
		defer close(monitorDone)
		select {
		case <-c.preferences.Done():
			c.stop("preferences watch lost; resync required")
		case <-w.Done():
			c.stop("watch lost; resync required")
		case <-lost:
		case <-c.ctx.Done():
		}
	}()
	if !c.send(map[string]any{"type": "preferences_snapshot", "revision": ps.Revision, "preferences": ps.Preferences}, true) {
		return
	}
	c.advancePreferences(ps.Revision)
	if !c.send(map[string]any{"type": "snapshot_begin", "id": id, "generation": s.Generation, "agent_id": s.AgentID, "watermark": s.Watermark, "first_seq": s.FirstSeq, "last_seq": s.LastSeq, "log_seq": s.LogSeq, "omitted": s.Omitted, "state": s.State}, true) {
		return
	}
	for _, e := range s.Events {
		if !c.send(map[string]any{"type": "snapshot_event", "generation": s.Generation, "event": ProjectEvent(c.parent, e)}, true) {
			return
		}
	}
	for _, p := range s.Partials {
		if !c.send(map[string]any{"type": "snapshot_partial", "generation": s.Generation, "agent_id": p.AgentID, "request_id": p.RequestID, "operation_id": p.OperationID, "part_id": p.PartID, "channels": p.Channels}, true) {
			return
		}
	}
	if !c.send(map[string]any{"type": "snapshot_end", "generation": s.Generation, "watermark": s.Watermark}, true) {
		return
	}
	c.advance(s.Watermark)
	c.mu.Lock()
	c.snapshotPending = false
	c.queue = append(c.queue, c.deferred...)
	c.deferred = nil
	c.mu.Unlock()
	close(c.snapshotDone)
	select {
	case c.wake <- struct{}{}:
	default:
	}
	c.workers.Add(1)
	go c.deliverPreferences()
	for {
		r, err := w.Next(c.ctx)
		if err != nil {
			c.stop("watch lost; resync required")
			return
		}
		if !c.send(map[string]any{"type": "observation", "generation": s.Generation, "revision": r.Revision, "observation": projectObservation(c.parent, r.Observation)}, false) {
			return
		}
		c.advance(r.Revision)
	}
}
func (c *Connector) completion(h ensemble.RequestHandle) {
	defer c.workers.Done()
	result, err := h.Wait(c.ctx)
	if err != nil {
		return
	}
	// An atomic cut after reliable completion supplies a barrier in the same tail.
	// This temporary watch is closed immediately; no duplicate observation is sent.
	s, w, err := c.parent.Ensemble().Watch(c.parent.AgentID())
	if err != nil {
		c.stop("Agent stopped; resync required")
		return
	}
	w.Close()
	for {
		c.mu.Lock()
		ready := c.revision >= s.Watermark
		advanced := c.advanced
		c.mu.Unlock()
		if ready {
			break
		}
		select {
		case <-advanced:
		case <-c.ctx.Done():
			return
		}
	}
	out := map[string]any{"type": "completion", "request_id": result.RequestID, "outcome": result.Outcome, "text": result.Text, "pending_hints": result.PendingHints}
	if result.Error != nil {
		out["error"] = result.Error
	}
	c.send(out, false)
}
