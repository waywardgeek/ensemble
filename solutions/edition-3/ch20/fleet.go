package ensemble

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"ensemble/internal/common"
)

type AgentOutcome = common.AgentOutcome
type AgentStatus = common.AgentStatus
type JoinResult = common.JoinResult
type PendingQuestion = common.PendingQuestion
type HumanQuestionSink func(Question)

// AgentFactory belongs to the application. It composes defaults, credentials,
// skills and state paths, then calls this Ensemble's NewAgent. It must return
// that registered Agent, with the requested ID, ParentID and submission contract.
// Until successful admission the factory owns its intended construction. On an
// error it either cleans that Agent itself, or returns Agent+error for rollback.
// It owns unrelated resources/Agents and must never return a foreign Agent.
type AgentFactory func(context.Context, Config) (Agent, error)

type spawnAdmission struct {
	parent string
	schema json.RawMessage
	agent  Agent
}

type agentEntry struct {
	id, parent         string
	agent              Agent
	state              string
	stopping, terminal bool
	resolved, retired  bool
	changed            chan struct{}
	cancel             context.CancelFunc
	done               chan struct{}
	outcome            AgentOutcome
	questions          map[string]*fleetQuestion
	nextQuestion       uint64
}

func validAgentID(id string) bool {
	if id == "" || len(id) > 128 {
		return false
	}
	for _, r := range id {
		if !(r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' || r == '-' || r == '_') {
			return false
		}
	}
	return true
}
func (e *Ensemble) entryLocked(id string) *agentEntry {
	for _, entry := range e.agents {
		if entry.id == id {
			return entry
		}
	}
	return nil
}
func (e *Ensemble) SetAgentFactory(factory AgentFactory) {
	e.mu.Lock()
	e.factory = factory
	e.mu.Unlock()
}

// CreateAgent uses application construction policy without starting a task or a server.
// NewAgent remains the low-level constructor called by the application factory.
func (e *Ensemble) CreateAgent(ctx context.Context, config Config) (Agent, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	e.mu.Lock()
	factory := e.factory
	closed := e.closed
	e.mu.Unlock()
	if closed {
		return nil, errors.New("ensemble is shut down")
	}
	if factory == nil {
		return e.NewAgent(config)
	}
	return factory(ctx, config)
}

// Spawn admits a delegated task before construction, then starts its prompt once
// through the child's normal mailbox. Config remains the per-Agent specification.
func (e *Ensemble) Spawn(ctx context.Context, parentID string, config Config) (Agent, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if !validAgentID(config.ID) {
		return nil, errors.New("spawn requires a valid unique agent ID")
	}
	if strings.TrimSpace(config.Prompt) == "" {
		return nil, errors.New("spawn requires a nonempty prompt")
	}
	canonical, err := common.CanonicalSubmitSchema(config.SubmitSchema)
	if err != nil {
		return nil, fmt.Errorf("submit schema: %w", err)
	}
	config.SubmitSchema = canonical
	config.ParentID = parentID
	e.mu.Lock()
	if config.ID == e.humanAddressLocked() {
		e.mu.Unlock()
		return nil, errors.New("human address is reserved")
	}
	parent := e.entryLocked(parentID)
	if e.closed || parent == nil || parent.stopping || parent.terminal || parent.state == "submitting" {
		e.mu.Unlock()
		return nil, fmt.Errorf("parent %q is unavailable", parentID)
	}
	if e.factory == nil {
		e.mu.Unlock()
		return nil, errors.New("application AgentFactory is not configured")
	}
	if e.entryLocked(config.ID) != nil || e.admissions[config.ID] != nil {
		e.mu.Unlock()
		return nil, fmt.Errorf("agent %q already exists", config.ID)
	}
	if e.admissions == nil {
		e.admissions = map[string]*spawnAdmission{}
	}
	admission := &spawnAdmission{parent: parentID, schema: append(json.RawMessage(nil), canonical...)}
	e.admissions[config.ID] = admission
	e.mu.Unlock()
	defer func() { e.mu.Lock(); delete(e.admissions, config.ID); e.mu.Unlock() }()
	a, err := e.CreateAgent(ctx, config)
	rollback := func() {
		e.mu.Lock()
		owned := a != nil && admission.agent == a
		entry := e.entryLocked(config.ID)
		owned = owned && entry != nil && entry.agent == a
		e.mu.Unlock()
		if !owned {
			return
		}
		_ = e.ShutdownAgent(config.ID)
		e.mu.Lock()
		for i, entry := range e.agents {
			if entry.agent == a {
				e.agents = append(e.agents[:i], e.agents[i+1:]...)
				break
			}
		}
		e.mu.Unlock()
	}
	if err != nil {
		rollback()
		return nil, err
	}
	e.mu.Lock()
	entry := e.entryLocked(config.ID)
	parent = e.entryLocked(parentID)
	if a == nil || admission.agent != a || entry == nil || entry.agent != a || entry.parent != parentID || !bytes.Equal(a.Config().SubmitSchema, admission.schema) {
		e.mu.Unlock()
		rollback()
		return nil, errors.New("factory did not return the requested registered child")
	}
	if err := ctx.Err(); err != nil {
		e.mu.Unlock()
		rollback()
		return nil, err
	}
	if e.closed || parent == nil || parent.stopping || parent.terminal || parent.state == "submitting" {
		e.mu.Unlock()
		rollback()
		return nil, errors.New("parent stopped during construction")
	}
	if entry.terminal || entry.state == "recovering" {
		e.mu.Unlock()
		return a, nil
	}
	if entry.stopping {
		e.mu.Unlock()
		rollback()
		return nil, errors.New("child stopped during construction")
	}
	taskCtx, cancel := context.WithCancel(context.Background())
	entry.cancel = cancel
	entry.state = "running"
	e.mu.Unlock()
	reply := a.Submit(taskCtx, config.Prompt)
	go func() {
		result := <-reply
		err := errors.Join(result.Err, e.stopChildren(entry.id), a.Shutdown())
		cancel()
		outcome := AgentOutcome{AgentID: entry.id, ParentID: entry.parent, State: "submitted", Submission: append(json.RawMessage(nil), result.Submission...)}
		if err == nil && len(result.Submission) == 0 {
			err = errors.New("delegated Agent ended without a valid submit")
		}
		if err != nil {
			outcome.State = "failed"
			outcome.Error = err.Error()
			outcome.Submission = nil
		}
		e.mu.Lock()
		e.settleLocked(entry, outcome)
		e.mu.Unlock()
	}()
	return a, nil
}
func (e *Ensemble) settleLocked(entry *agentEntry, outcome AgentOutcome) {
	if entry.terminal {
		return
	}
	entry.outcome = outcome
	entry.state = outcome.State
	entry.terminal = true
	entry.stopping = true
	close(entry.done)
	e.signalLocked(entry)
	e.signalLocked(e.entryLocked(entry.parent))
}
func (e *Ensemble) Children(parentID string) []AgentStatus {
	e.mu.Lock()
	defer e.mu.Unlock()
	result := []AgentStatus{}
	for _, entry := range e.agents {
		if entry.parent == parentID && !entry.retired {
			result = append(result, statusOf(entry))
		}
	}
	return result
}
func (e *Ensemble) Status(id string) (AgentStatus, error) {
	e.mu.Lock()
	defer e.mu.Unlock()
	entry := e.entryLocked(id)
	if entry == nil || entry.retired {
		return AgentStatus{}, fmt.Errorf("unknown agent %q", id)
	}
	return statusOf(entry), nil
}
func (e *Ensemble) Outcome(id string) (AgentOutcome, error) {
	e.mu.Lock()
	defer e.mu.Unlock()
	entry := e.entryLocked(id)
	if entry == nil {
		return AgentOutcome{}, fmt.Errorf("unknown agent %q", id)
	}
	if !entry.terminal {
		return AgentOutcome{}, fmt.Errorf("agent %q is not terminal", id)
	}
	outcome := entry.outcome
	outcome.Submission = append(json.RawMessage(nil), outcome.Submission...)
	return outcome, nil
}
func (e *Ensemble) Wait(ctx context.Context, id string) (AgentOutcome, error) {
	e.mu.Lock()
	entry := e.entryLocked(id)
	e.mu.Unlock()
	if entry == nil {
		return AgentOutcome{}, fmt.Errorf("unknown agent %q", id)
	}
	select {
	case <-entry.done:
		return e.Outcome(id)
	case <-ctx.Done():
		return AgentOutcome{}, ctx.Err()
	}
}

// Join snapshots current direct children when ids is empty; it does not join
// future children. A failed child is a terminal outcome, not a reason to hang.
func (e *Ensemble) Join(ctx context.Context, parentID string, ids []string) (JoinResult, error) {
	for {
		e.mu.Lock()
		parent := e.entryLocked(parentID)
		if parent == nil {
			e.mu.Unlock()
			return JoinResult{}, fmt.Errorf("unknown parent %q", parentID)
		}
		selected := append([]string(nil), ids...)
		if len(selected) == 0 {
			for _, child := range e.agents {
				if child.parent == parentID && !child.retired {
					selected = append(selected, child.id)
				}
			}
		}
		result := JoinResult{Outcomes: []AgentOutcome{}}
		all := true
		wanted := map[string]bool{}
		for _, id := range selected {
			child := e.entryLocked(id)
			if child == nil || child.parent != parentID {
				e.mu.Unlock()
				return JoinResult{}, fmt.Errorf("%q is not your direct child", id)
			}
			wanted[id] = true
			if child.terminal {
				result.Outcomes = append(result.Outcomes, copyOutcome(child.outcome))
			} else {
				all = false
			}
		}
		for _, recipient := range e.agents {
			for request, q := range recipient.questions {
				for sender := e.entryLocked(q.sender); sender != nil; sender = e.entryLocked(sender.parent) {
					if wanted[sender.id] {
						result.Questions = append(result.Questions, PendingQuestion{AgentID: q.sender, Recipient: q.recipient, RequestID: request, Text: q.text, Human: q.human})
						break
					}
				}
			}
		}

		sort.Slice(result.Questions, func(i, j int) bool { return result.Questions[i].RequestID < result.Questions[j].RequestID })
		if all || len(result.Questions) > 0 {
			e.mu.Unlock()
			return result, nil
		}
		changed := parent.changed
		done := parent.done
		e.mu.Unlock()
		select {
		case <-ctx.Done():
			return JoinResult{}, ctx.Err()
		case <-done:
			return JoinResult{}, errors.New("parent stopped while joining")
		case <-changed:
		}
	}
}
func (e *Ensemble) signalLocked(entry *agentEntry) {
	if entry != nil {
		if entry.changed != nil {
			close(entry.changed)
		}
		entry.changed = make(chan struct{})
	}
}

// CompleteSubmission is called only after the actor has closed admission, saved
// its final history and submit row, and stopped its owned jobs and resources.
func (e *Ensemble) CompleteSubmission(id string, payload json.RawMessage, err error) {
	e.mu.Lock()
	defer e.mu.Unlock()
	entry := e.entryLocked(id)
	if entry == nil {
		return
	}
	outcome := AgentOutcome{AgentID: id, ParentID: entry.parent, State: "submitted", Submission: append(json.RawMessage(nil), payload...)}
	if err != nil {
		outcome.State = "failed"
		outcome.Error = err.Error()
		outcome.Submission = nil
	}
	e.settleLocked(entry, outcome)
}

func (e *Ensemble) AcceptSubmission(id string, payload json.RawMessage) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	entry := e.entryLocked(id)
	if entry == nil {
		return errors.New("unknown Agent")
	}
	if entry.stopping || entry.terminal || entry.state == "submitting" {
		return errors.New("Agent is already terminal or submitting")
	}
	if err := common.ValidateSubmission(entry.agent.Config().SubmitSchema, payload); err != nil {
		return err
	}
	for _, child := range e.agents {
		if child.parent == id && (!child.terminal || (child.outcome.State != "submitted" && !child.resolved)) {
			return fmt.Errorf("child %q is unfinished or failed without resolution; explicitly shut it down before submit", child.id)
		}
	}
	entry.state = "submitting"
	return nil
}
func (e *Ensemble) relatedLocked(sender, target *agentEntry) bool {
	return sender != nil && target != nil && !sender.stopping && !sender.terminal && !target.stopping && !target.terminal && (target.parent == sender.id || sender.parent == target.id)
}

// SendMessage asks any live Agent, or the configured human address, and returns
// a correlated envelope with the address to use for a follow-up. Empty target
// means the sender's parent, falling back to the human for a parentless Agent.
func (e *Ensemble) SendMessage(ctx context.Context, senderID, targetID, text string) (Question, error) {
	if err := ctx.Err(); err != nil {
		return Question{}, err
	}
	e.mu.Lock()
	sender := e.entryLocked(senderID)
	if sender == nil || sender.stopping || sender.terminal || sender.retired {
		e.mu.Unlock()
		return Question{}, errors.New("message sender is unavailable")
	}
	if targetID == "" {
		targetID = sender.parent
		if targetID == "" {
			targetID = e.humanAddressLocked()
		}
	}
	human := targetID == e.humanAddressLocked()
	holder := e.entryLocked(targetID)
	sink := e.humanSink
	var requestID string
	if human {
		if sink == nil {
			e.mu.Unlock()
			return Question{}, errors.New("human question sink is unavailable")
		}
		holder = sender
		e.nextHumanQuestion++
		requestID = fmt.Sprintf("%s-question-%d", targetID, e.nextHumanQuestion)
	} else {
		if holder == nil || holder.stopping || holder.terminal || holder.retired {
			e.mu.Unlock()
			return Question{}, errors.New("message recipient is unavailable")
		}
		holder.nextQuestion++
		requestID = fmt.Sprintf("%s-question-%d", targetID, holder.nextQuestion)
	}
	q := &fleetQuestion{sender: senderID, recipient: targetID, text: text, human: human, reply: make(chan Question, 1)}
	if holder.questions == nil {
		holder.questions = map[string]*fleetQuestion{}
	}
	holder.questions[requestID] = q
	e.signalLocked(holder)
	e.signalAncestorsLocked(sender)
	e.mu.Unlock()
	defer func() {
		e.mu.Lock()
		delete(holder.questions, requestID)
		e.signalLocked(holder)
		e.signalAncestorsLocked(sender)
		e.mu.Unlock()
	}()
	envelope := Question{Sender: senderID, Recipient: targetID, RequestID: requestID, Text: text}
	if human {
		sink(envelope)
	} else if err := holder.agent.Post(envelope); err != nil {
		return Question{}, err
	}
	select {
	case reply := <-q.reply:
		return reply, nil
	case <-ctx.Done():
		return Question{}, ctx.Err()
	case <-sender.done:
		return Question{}, errors.New("sender stopped before receiving a reply")
	case <-holder.done:
		return Question{}, errors.New("recipient stopped before answering")
	}
}
func (e *Ensemble) signalAncestorsLocked(entry *agentEntry) {
	for entry != nil {
		e.signalLocked(entry)
		entry = e.entryLocked(entry.parent)
	}
}

func (e *Ensemble) AskAgent(ctx context.Context, senderID, targetID, text string) (Result, error) {
	e.mu.Lock()
	sender, target := e.entryLocked(senderID), e.entryLocked(targetID)
	allowed := e.relatedLocked(sender, target)
	e.mu.Unlock()
	if !allowed {
		return Result{}, errors.New("asking requires a live parent/direct-child relationship")
	}
	result := <-target.agent.SubmitLive(ctx, text)
	return result, result.Err
}

// ShutdownChildren settles owned descendants through the same shutdown owner.
func (e *Ensemble) ShutdownChildren(id string) error { return e.stopChildren(id) }
func (e *Ensemble) stopChildren(id string) error {
	var err error
	for _, child := range e.Children(id) {
		err = errors.Join(err, e.ShutdownAgent(child.AgentID))
	}
	return err
}

// ShutdownAgent recursively stops descendants, actor work and jobs. Terminal
// records are retained; callers may still Wait, Join and inspect the outcome.
func (e *Ensemble) ShutdownAgent(id string) error {
	e.mu.Lock()
	entry := e.entryLocked(id)
	if entry == nil {
		e.mu.Unlock()
		return fmt.Errorf("unknown agent %q", id)
	}
	if entry.terminal {
		entry.retired = true
		entry.resolved = true
		e.signalLocked(e.entryLocked(entry.parent))
		e.mu.Unlock()
		return nil
	}
	entry.resolved = true
	entry.stopping = true
	entry.state = "stopping"
	cancel := entry.cancel
	e.mu.Unlock()
	if cancel != nil {
		cancel()
	}
	err := errors.Join(e.stopChildren(id), entry.agent.Shutdown())
	if cancel != nil {
		<-entry.done
	} else {
		outcome := AgentOutcome{AgentID: id, ParentID: entry.parent, State: "stopped"}
		if err != nil {
			outcome.Error = err.Error()
			outcome.State = "failed"
		}
		e.mu.Lock()
		e.settleLocked(entry, outcome)
		e.mu.Unlock()
	}
	e.mu.Lock()
	entry.retired = true
	e.signalAncestorsLocked(entry)
	e.mu.Unlock()
	return err
}

// Pending questions belong to the recipient's fleet entry. Delivery still uses
// the Agent mailbox; this table only correlates one explicit reply per question.
type fleetQuestion struct {
	sender, recipient, text string
	human                   bool
	reply                   chan Question
}

func (e *Ensemble) AskParent(ctx context.Context, senderID, text string) (string, error) {
	reply, err := e.SendMessage(ctx, senderID, "", text)
	return reply.Text, err
}
func (e *Ensemble) RespondToQuestion(recipientID, senderID, requestID, response string) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	recipient, sender := e.entryLocked(recipientID), e.entryLocked(senderID)
	if recipient == nil || sender == nil || recipient.stopping || recipient.terminal || sender.stopping || sender.terminal {
		return errors.New("question participant is unavailable")
	}
	if requestID == "" {
		for id, q := range recipient.questions {
			if !q.human && q.sender == senderID {
				if requestID != "" {
					return errors.New("multiple pending questions; specify message_id")
				}
				requestID = id
			}
		}
	}
	q := recipient.questions[requestID]
	if q == nil || q.human || q.sender != senderID || q.recipient != recipientID {
		return errors.New("unknown pending question")
	}
	delete(recipient.questions, requestID)
	q.reply <- Question{Sender: recipientID, Recipient: senderID, RequestID: requestID, Text: response}
	e.signalLocked(recipient)
	e.signalAncestorsLocked(sender)
	return nil
}

func statePaths(c Config) ([]string, error) {
	raw := []string{c.LogPath}
	if c.DataDir != "" {
		raw = append(raw, filepath.Join(c.DataDir, "submit.md"))
	}
	if c.SavePath != "" {
		raw = append(raw, c.SavePath, c.SavePath+".bak")
	}
	var paths []string
	for _, path := range raw {
		if path == "" {
			continue
		}
		canonical, err := common.CanonicalPath(path)
		if err != nil {
			return nil, err
		}
		paths = append(paths, canonical)
	}
	return paths, nil
}

// SetHumanQuestionSink installs application presentation only. Correlation,
// cancellation, and replies stay in the library's existing question owner.
func (e *Ensemble) SetHumanQuestionSink(sink HumanQuestionSink) {
	e.mu.Lock()
	e.humanSink = sink
	e.mu.Unlock()
}

// HumanAddress is a reserved routing address, never a managed Agent ID.
func (e *Ensemble) humanAddressLocked() string {
	if e.humanAddress == "" {
		return "human"
	}
	return e.humanAddress
}
func (e *Ensemble) HumanAddress() string {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.humanAddressLocked()
}
func (e *Ensemble) SetHumanAddress(address string) error {
	if !validAgentID(address) {
		return errors.New("invalid human address")
	}
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.entryLocked(address) != nil || e.admissions[address] != nil {
		return errors.New("human address collides with an Agent")
	}
	for _, entry := range e.agents {
		for _, q := range entry.questions {
			if q.human {
				return errors.New("cannot change human address while questions are pending")
			}
		}
	}
	e.humanAddress = address
	return nil
}
func (e *Ensemble) AskHuman(ctx context.Context, senderID, text string) (string, error) {
	reply, err := e.SendMessage(ctx, senderID, e.HumanAddress(), text)
	return reply.Text, err
}
func (e *Ensemble) PendingHumanQuestions() []Question {
	e.mu.Lock()
	defer e.mu.Unlock()
	result := []Question{}
	for _, entry := range e.agents {
		for id, q := range entry.questions {
			if q.human {
				result = append(result, Question{Sender: q.sender, Recipient: q.recipient, RequestID: id, Text: q.text})
			}
		}
	}
	sort.Slice(result, func(i, j int) bool { return result[i].RequestID < result[j].RequestID })
	return result
}
func (e *Ensemble) respondHumanLocked(senderID, requestID, response string) error {
	var owner *agentEntry
	var question *fleetQuestion
	for _, entry := range e.agents {
		for id, q := range entry.questions {
			if q.human && (senderID == "" || q.sender == senderID) && (requestID == "" || id == requestID) {
				if question != nil {
					return errors.New("multiple human questions; specify message_id")
				}
				owner, question = entry, q
			}
		}
	}
	if question == nil || owner.terminal || owner.stopping {
		return errors.New("unknown pending human question")
	}
	if requestID == "" {
		for id, q := range owner.questions {
			if q == question {
				requestID = id
				break
			}
		}
	}
	delete(owner.questions, requestID)
	question.reply <- Question{Sender: question.recipient, Recipient: question.sender, RequestID: requestID, Text: response}
	e.signalAncestorsLocked(owner)
	return nil
}
func (e *Ensemble) RespondToHumanQuestion(senderID, requestID, response string) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.respondHumanLocked(senderID, requestID, response)
}

// ReplyToHuman answers the sole pending question if messageID is empty; ambiguous
// and stale requests are errors and can never become ordinary chat input.
func (e *Ensemble) ReplyToHuman(messageID, response string) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.respondHumanLocked("", messageID, response)
}

// statusOf is called under the registry lock; it owns no additional state.
func statusOf(entry *agentEntry) AgentStatus {
	status := AgentStatus{AgentID: entry.id, ParentID: entry.parent, State: entry.state}
	if entry.terminal {
		outcome := copyOutcome(entry.outcome)
		status.Outcome = &outcome
	}
	return status
}

func copyOutcome(out AgentOutcome) AgentOutcome {
	out.Submission = append(json.RawMessage(nil), out.Submission...)
	return out
}
