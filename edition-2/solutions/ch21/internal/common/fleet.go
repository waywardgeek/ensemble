package common

import (
	"context"
	"encoding/json"
	"time"
)

// UserSink is the application's delivery boundary for questions and escalation.
// Deliver must return promptly; a question's reply comes through Fleet.Reply.
type UserSink interface {
	// Deliver presents the sender and message without impersonating a fleet member.
	Deliver(FleetMessage) error
}

// FleetMessage retains sender identity so questions and replies can be continued.
type FleetMessage struct {
	// ID identifies a pending question; notices need no reply.
	ID string `json:"message_id,omitempty"`
	// From identifies the sender in the recipient's immediate ownership tree.
	From string `json:"from"`
	// To is the receiving Agent or configured human address.
	To string `json:"to"`
	// Content carries the question, answer or escalation.
	Content string `json:"content"`
}

// Submission preserves the distinction between legacy text and a custom JSON object.
// Returned Payload bytes belong to the caller, not the live Fleet.
type Submission struct {
	// Result holds default-schema text; it can legitimately be empty.
	Result string `json:"result,omitempty"`
	// Payload holds the full custom argument object, without a result wrapper.
	Payload json.RawMessage `json:"payload,omitempty"`
}

// ChildStatus is a snapshot of one owned child's completion, not a global registry.
type ChildStatus struct {
	// Activity distinguishes provider work, tools, waiting and idle.
	Activity Activity `json:"activity"`
	// Usage contains the child's measured session tokens.
	Usage Usage `json:"usage"`
	// ID names this parent's child.
	ID string `json:"agent_id"`
	// State is idle, running, submitted, or failed.
	State string `json:"state"`
	// Result is default-schema text, never inferred from an idle reply.
	Result string `json:"result,omitempty"`
	// Payload is the custom submitted object, with caller-owned backing bytes.
	Payload json.RawMessage `json:"payload,omitempty"`
	// Error records a failed turn without retrying it.
	Error string `json:"error,omitempty"`
}

// SubAgentManager owns child lifetimes and pending questions for one Agent.
type SubAgentManager interface {
	// Submission reads default text with durable fallback; custom data requires SubmissionData.
	Submission(string) (string, error)
	// SubmissionData reads typed text/custom data, resolving ownership before disk fallback.
	SubmissionData(string) (Submission, error)
	// Children returns the actual owned children for trusted tree traversal.
	Children() []Agent
	// Find resolves only this manager's children; foreign and absent IDs are identical.
	Find(string) (Agent, error)
	// WaitChange waits for a state, output, question, exit, pattern or stuck condition.
	WaitChange(context.Context, []string, string, time.Duration) (string, error)
	// Output returns finalized output since an event cursor, optionally waiting.
	Output(context.Context, string, uint64, bool) (ChildOutput, error)
	// GracefulShutdown waits for queued turns before ending one child's lifetime.
	GracefulShutdown(context.Context, string) error
	// Agent returns the required owner.
	Agent() Agent
	// Spawn constructs, restores and starts a child through the application factory.
	Spawn(id, task, model string) (ChildStatus, error)
	// SpawnWithSchema validates and freezes the optional whole-argument submit contract.
	SpawnWithSchema(id, task, model string, schema json.RawMessage) (ChildStatus, error)
	// List returns only this Agent's children in stable ID order.
	List() []ChildStatus
	// Join waits for all requested submissions, yielding early to pending questions.
	Join(context.Context, []string) (string, error)
	// ShutdownChild joins and saves one child, then removes it, preserving its files.
	ShutdownChild(string) error
	// PrepareDelete atomically refuses active children or closes child admission.
	PrepareDelete() error
	// Shutdown closes admission and recursively joins all owned children.
	Shutdown() error
	// AskParent blocks until the addressed parent or human replies.
	AskParent(context.Context, string) (FleetMessage, error)
	// Receive delivers a child question or notice and allocates its recipient-local ID.
	Receive(FleetMessage, chan FleetMessage) (string, error)
	// Reply answers a pending question; blank agentID selects human-addressed questions.
	// Omitted messageID selects the sole matching question.
	Reply(agentID, messageID, content string) error
	// Submit wraps result text in JSON and obeys the same custom contract as SubmitJSON.
	Submit(string) error
	// SubmitJSON validates the direct argument object before publishing completion.
	SubmitJSON(json.RawMessage) error
	// SubmitSchema returns an owned copy of the runtime submission contract.
	SubmitSchema() json.RawMessage
	// Status reports this Agent's explicit completion state.
	Status() ChildStatus
	// End records an actual turn failure; interruption leaves the child available.
	End(error)
	// Changed wakes fan-in observers after a child transition.
	Changed()
	// Begin starts a new task without forgetting consecutive idle corrections.
	Begin()
	// Idle returns a correction or escalation after a successful unsubmitted turn.
	Idle() (correction, escalation string)
	// Escalate reports failure to the parent without opening a blocking question.
	Escalate(string) error
}

// ChildOutput carries only new finalized output and its owning event cursor.
type ChildOutput struct {
	// ID is the caller-local child address.
	ID string `json:"agent_id"`
	// Cursor resumes after the last inspected history event.
	Cursor uint64 `json:"cursor"`
	// Content includes assistant text and real tool results, never user input.
	Content string `json:"content"`
	// Status describes the child at the same read boundary.
	Status ChildStatus `json:"status"`
}
