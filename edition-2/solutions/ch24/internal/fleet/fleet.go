// Package fleet owns recursive child lifetimes and parent question/reply flow.
package fleet

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"ensemble/internal/common"
)

type question struct {
	message common.FleetMessage
	reply   chan common.FleetMessage
}
type child struct {
	id    string // the manager's reserved address, independent of a custom factory's label
	agent common.Agent
	// Both automatic completion and explicit kill share one drain/save result.
	stop      sync.Once
	stopErr   error
	cancelled bool // protected by the owning manager's mutex
}

// The mutex protects fleet facts, never conversation state. Model work still
// runs through each child's existing actor; no second execution loop is added.
type manager struct {
	parent    common.Agent
	mu        sync.Mutex
	children  map[string]*child
	pending   map[string]question
	messages  map[string]uint64
	changed   chan struct{}
	stopped   bool
	submitted bool
	result    common.Submission
	contract  *SubmissionContract
	failure   string
	wakes     int
	sequence  uint64
}

// New requires the Agent whose children and messages this manager owns.
func New(parent common.Agent, contract *SubmissionContract) common.SubAgentManager {
	if parent == nil || contract == nil {
		panic("fleet requires Agent and submission contract")
	}
	return &manager{parent: parent, contract: contract, messages: map[string]uint64{}, children: map[string]*child{}, pending: map[string]question{}, changed: make(chan struct{})}
}

// Agent exposes the required owner, including its configuration and services.
func (m *manager) Agent() common.Agent { return m.parent }
func (m *manager) signal()             { close(m.changed); m.changed = make(chan struct{}) }

// Changed broadcasts a child transition without retaining a stale wake flag.
func (m *manager) Changed() { m.mu.Lock(); defer m.mu.Unlock(); m.signal() }

// Spawn reserves the child's identity through construction. A failed factory
// publishes nothing, and no model starts until both ownership links exist.
func (m *manager) Spawn(id, task, model string) (common.ChildStatus, error) {
	return m.SpawnWithSchema(id, task, model, nil)
}

// SpawnWithSchema delegates contract validation to the public construction boundary.
// Ordinary grandchildren get defaults unless their own launch explicitly opts in.
func (m *manager) SpawnWithSchema(id, task, model string, schema json.RawMessage) (common.ChildStatus, error) {
	// Internal child construction does not enable ordinary delegation. Preserve
	// the previous refusal rather than start a child whose submit tool is hidden.
	if !m.parent.Ensemble().FleetEnabled() {
		return common.ChildStatus{}, errors.New("public delegation is not enabled")
	}
	c, result, err := m.launch(id, task, model, schema, false)
	if err != nil {
		return common.ChildStatus{}, err
	}
	go func() {
		<-result
		m.Changed()
	}()
	return common.ChildStatus{ID: c.id, State: "running"}, nil
}

// launch reserves identity through construction. Ordinary tasks restore their
// conversation; managed workers require fresh paths and never load old results.
func (m *manager) launch(id, task, model string, schema json.RawMessage, managed bool) (*child, <-chan common.Result, error) {
	if id == "" || id == "." || id == ".." || strings.ContainsAny(id, "/\\") || id == m.parent.Ensemble().UserID() || strings.TrimSpace(task) == "" {
		return nil, nil, errors.New("spawn requires a child ID and task; user and path addresses are reserved")
	}
	m.mu.Lock()
	if m.stopped {
		m.mu.Unlock()
		return nil, nil, common.ErrStopped
	}
	if m.children[id] != nil {
		m.mu.Unlock()
		return nil, nil, fmt.Errorf("child %q already exists", id)
	}
	spec := m.parent.Spec()
	spec.Compaction = managed
	spec.SubmitSchema = append(json.RawMessage(nil), schema...)
	spec.ID, spec.ParentID, spec.Model = id, m.parent.Spec().ID, model
	spec.DataDir = filepath.Join(m.parent.Config().DataDir, "children", id)
	spec.LogPath = filepath.Join(spec.DataDir, "events.jsonl")
	spec.SavePath = filepath.Join(spec.DataDir, "save.json")
	if managed {
		if _, err := m.parent.Sandbox().Stat(spec.DataDir); !os.IsNotExist(err) {
			m.mu.Unlock()
			if err == nil {
				err = errors.New("compaction child directory already exists")
			}
			return nil, nil, err
		}
	}
	a, err := m.parent.Ensemble().CreateChild(m.parent, spec)
	if err != nil {
		m.mu.Unlock()
		return nil, nil, err
	}
	if !managed {
		err = a.Load()
	}
	if err == nil {
		err = a.Save()
	}
	if err != nil {
		// Joining can notify the parent. Never retain its manager lock here.
		m.mu.Unlock()
		err = errors.Join(err, a.Shutdown())
		m.parent.Ensemble().ForgetCompaction(a)
		return nil, nil, err
	}
	c := &child{id: id, agent: a}
	m.children[id] = c
	if !managed {
		m.submitted = false
	}
	m.signal()
	result := a.Submit(context.Background(), task)
	m.mu.Unlock()
	m.parent.Ensemble().Observe(m.parent, common.FleetChanged{})
	return c, result, nil
}

// StartCompaction retains the reliable Result before automatic drain/removal.
// A concurrent explicit kill shares this exact entry's save result, even after
// its map entry disappears; no dismissed-ID registry is needed.
func (m *manager) StartCompaction(id, task string) (common.Agent, <-chan common.ChildCompletion, error) {
	c, result, err := m.launch(id, task, "", nil, true)
	if err != nil {
		return nil, nil, err
	}
	done := make(chan common.ChildCompletion, 1)
	go func() {
		out := common.ChildCompletion{Result: <-result}
		out.CleanupErr = m.stopChild(c)
		status := c.agent.Fleet().Status()
		out.Submission = common.Submission{Result: status.Result, Payload: append(json.RawMessage(nil), status.Payload...)}
		out.Usage = c.agent.Engine().UsageBuckets()
		m.mu.Lock()
		out.Cancelled = c.cancelled
		m.mu.Unlock()
		done <- out
		close(done)
	}()
	return c.agent, done, nil
}

// Status reports explicit submission; ordinary text never means finished.
func (m *manager) Status() common.ChildStatus {
	m.mu.Lock()
	defer m.mu.Unlock()
	s := common.ChildStatus{ID: m.parent.Spec().ID, State: "running", Activity: m.parent.Activity(), Usage: m.parent.Engine().Usage()}
	if s.Activity.Phase == common.ActivityIdle {
		s.State = "idle"
	}
	if m.failure != "" {
		s.State = "failed"
		s.Error = m.failure
	}
	if m.submitted {
		s.State = "submitted"
		s.Result = m.result.Result
		s.Payload = append(json.RawMessage(nil), m.result.Payload...)
	}
	return s
}
func (m *manager) list() []common.ChildStatus {
	out := make([]common.ChildStatus, 0, len(m.children))
	for id, c := range m.children {
		s := c.agent.Fleet().Status()
		s.ID = id
		out = append(out, s)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}

// List returns an owned snapshot, never a mutable child map.
func (m *manager) List() []common.ChildStatus { m.mu.Lock(); defer m.mu.Unlock(); return m.list() }
func encoded(v any) string                    { b, _ := json.Marshal(v); return string(b) }

// Join checks condition and subscribes under the same lock, so completion or a
// question cannot slip between those operations. Questions yield without submit.
func (m *manager) Join(ctx context.Context, ids []string) (string, error) {
	for {
		m.mu.Lock()
		if m.stopped {
			m.mu.Unlock()
			return "", common.ErrStopped
		}
		if len(ids) == 0 {
			for id := range m.children {
				ids = append(ids, id)
			}
			sort.Strings(ids)
		}
		var results []common.ChildStatus
		done := true
		for _, id := range ids {
			c := m.children[id]
			if c == nil {
				m.mu.Unlock()
				return "", fmt.Errorf("unknown agent %q", id)
			}
			s := c.agent.Fleet().Status()
			if s.State == "failed" {
				m.mu.Unlock()
				return "", fmt.Errorf("child %s failed: %s", id, s.Error)
			}
			results = append(results, s)
			done = done && s.State == "submitted"
		}
		if len(m.pending) > 0 {
			messages := make([]common.FleetMessage, 0, len(m.pending))
			for _, q := range m.pending {
				messages = append(messages, q.message)
			}
			sort.Slice(messages, func(i, j int) bool { return messages[i].ID < messages[j].ID })
			m.mu.Unlock()
			return encoded(messages), nil
		}
		if done {
			m.mu.Unlock()
			return encoded(results), nil
		}
		changed := m.changed
		m.mu.Unlock()
		select {
		case <-changed:
		case <-ctx.Done():
			return "", ctx.Err()
		}
	}
}

// ShutdownChild preserves files and waits for the full recursive lifetime before
// freeing an ID. Holding no manager lock while joining avoids reverse-call cycles.
func (m *manager) ShutdownChild(id string) error {
	m.mu.Lock()
	c := m.children[id]
	if c != nil {
		c.cancelled = true
		m.signal()
	}
	m.mu.Unlock()
	if c == nil {
		return fmt.Errorf("unknown agent %q", id)
	}
	// Publish the cancellation request before a potentially long shared drain.
	m.parent.Ensemble().Observe(m.parent, common.FleetChanged{})
	return m.stopChild(c)
}

func (m *manager) stopChild(c *child) error {
	c.stop.Do(func() {
		c.stopErr = errors.Join(c.agent.Shutdown(), c.agent.Save())
		m.mu.Lock()
		id := c.id
		if m.children[id] == c {
			delete(m.children, id)
		}
		m.signal()
		m.mu.Unlock()
		m.parent.Ensemble().ForgetCompaction(c.agent)
		m.parent.Ensemble().Observe(m.parent, common.FleetChanged{})
	})
	return c.stopErr
}

// Shutdown closes spawning first, then joins the retained child set recursively.
func (m *manager) Shutdown() error {
	m.mu.Lock()
	m.stopped = true
	var children []*child
	for _, c := range m.children {
		c.cancelled = true
		children = append(children, c)
	}
	m.signal()
	m.mu.Unlock()
	var errs []error
	for _, c := range children {
		errs = append(errs, m.stopChild(c))
	}
	return errors.Join(errs...)
}

// AskParent stores the reply rendezvous at its recipient. Cancellation removes
// abandoned questions; no stale channel can answer a later job by accident.
func (m *manager) AskParent(ctx context.Context, content string) (common.FleetMessage, error) {
	msg := common.FleetMessage{From: m.parent.Spec().ID, To: m.parent.Spec().ParentID, Content: content}
	reply := make(chan common.FleetMessage, 1)
	var recipient common.SubAgentManager = m
	if parent := m.parent.ParentAgent(); parent != nil {
		recipient = parent.Fleet()
	}
	id, err := recipient.Receive(msg, reply)
	if err != nil {
		return common.FleetMessage{}, err
	}
	select {
	case result := <-reply:
		return result, nil
	case <-ctx.Done():
		// A canceled rendezvous is retired through the same recipient operation.
		recipient.Reply(msg.From, id, "question canceled")
		return common.FleetMessage{}, ctx.Err()
	}
}

// Receive retains pending questions before waking a parent. Top-level delivery
// goes to the user sink, never back into the asking model's own conversation.
func (m *manager) Receive(msg common.FleetMessage, reply chan common.FleetMessage) (string, error) {
	m.mu.Lock()
	if m.stopped {
		m.mu.Unlock()
		return "", common.ErrStopped
	}
	if msg.To != m.parent.Ensemble().UserID() {
		m.messages[msg.From]++
	}
	// IDs belong to the rendezvous owner. Sender-local counters collide when
	// a child legitimately has the same address as its parent.
	if reply != nil {
		m.sequence++
		msg.ID = fmt.Sprintf("%s-%d", m.parent.Spec().ID, m.sequence)
		m.pending[msg.ID] = question{msg, reply}
	}
	m.signal()
	m.mu.Unlock()
	if msg.To == m.parent.Ensemble().UserID() {
		if err := m.parent.Record(common.Event{Type: "fleet_notice", Message: &common.MessageData{Actor: common.System, Parts: []common.Part{{Type: "text", Text: encoded(msg)}}}}); err != nil {
			return "", err
		}
		if err := m.parent.Ensemble().DeliverUser(msg); err != nil {
			m.mu.Lock()
			delete(m.pending, msg.ID)
			m.mu.Unlock()
			return "", err
		}
		return msg.ID, nil
	}
	return msg.ID, m.parent.Post(common.UserMessage{Text: "Fleet message: " + encoded(msg)})
}

// Reply optionally resolves the sole pending question from one child. Ambiguity
// is an error; guessing would send a correct answer to the wrong waiting job.
func (m *manager) Reply(agentID, messageID, content string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	var selected *question
	for id, q := range m.pending {
		humanQuestion := q.message.To == m.parent.Ensemble().UserID()
		recipientMatches := messageID != "" || (agentID == "") == humanQuestion
		if recipientMatches && (agentID == "" || q.message.From == agentID) && (messageID == "" || id == messageID) {
			if selected != nil {
				return errors.New("multiple questions pending; provide message_id")
			}
			copy := q
			selected = &copy
		}
	}
	if selected == nil {
		return errors.New("no matching pending question")
	}
	delete(m.pending, selected.message.ID)
	from := m.parent.Spec().ID
	if selected.message.To == m.parent.Ensemble().UserID() {
		from = m.parent.Ensemble().UserID()
	}
	selected.reply <- common.FleetMessage{ID: selected.message.ID, From: from, To: selected.message.From, Content: content}
	m.signal()
	return nil
}
func (m *manager) active() []string {
	var ids []string
	for _, s := range m.list() {
		if s.State != "submitted" {
			ids = append(ids, s.ID)
		}
	}
	return ids
}

// Begin clears completion for new work while preserving the consecutive-wake bound.
func (m *manager) Begin() {
	m.mu.Lock()
	m.submitted = false
	m.result = common.Submission{}
	m.failure = ""
	m.mu.Unlock()
}

// Idle applies reactive correction only to a successful turn; the actor never
// calls it for transport, persistence or interruption failures.
func (m *manager) Idle() (string, string) {
	// Human-facing Agents finish in chat. Only delegated Agents owe a submit.
	if m.parent.ParentAgent() == nil {
		return "", ""
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.submitted || m.stopped {
		return "", ""
	}
	if m.wakes >= 3 {
		m.failure = fmt.Sprintf("Agent %s stopped without submit after three framework wakes", m.parent.Spec().ID)
		return "", m.failure
	}
	m.wakes++
	return fmt.Sprintf("Framework wake %d/3 for agent %s: work is not submitted. Active children: %v. Use join_agents or shutdown_agent for unfinished children, then submit your result.", m.wakes, m.parent.Spec().ID, m.active()), ""
}

// Escalate is a notice, not a question: reporting a failed child cannot itself
// leave the framework blocked waiting for another model to answer.
func (m *manager) Escalate(content string) error {
	msg := common.FleetMessage{From: m.parent.Spec().ID, To: m.parent.Spec().ParentID, Content: content}
	if parent := m.parent.ParentAgent(); parent != nil {
		_, err := parent.Fleet().Receive(msg, nil)
		return err
	}
	return m.parent.Ensemble().DeliverUser(msg)
}

// End retains real failures without treating human interruption as agent death.
func (m *manager) End(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if err != nil && !errors.Is(err, common.ErrInterrupted) && !errors.Is(err, common.ErrStopped) {
		m.failure = err.Error()
	}
	m.signal()
}

// PrepareDelete closes child admission under the same lock as the active check.
// This prevents a racing spawn from orphaning new work after a GUI deletion gate.
func (m *manager) PrepareDelete() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if len(m.active()) > 0 {
		return errors.New("agent has active children; join or shut them down first")
	}
	m.stopped = true
	m.signal()
	return nil
}
