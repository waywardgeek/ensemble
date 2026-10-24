package common

import (
	"context"
	"encoding/json"
)

// AgentOutcome is retained after an Agent's actor and jobs have stopped.
// Submitted means validated submit, settled descendants, and a successful final save.
type AgentOutcome struct {
	AgentID    string          `json:"agent_id"`
	ParentID   string          `json:"parent_id,omitempty"`
	State      string          `json:"state"`
	Submission json.RawMessage `json:"submission,omitempty"`
	Error      string          `json:"error,omitempty"`
}
type AgentStatus struct {
	AgentID  string `json:"agent_id"`
	ParentID string `json:"parent_id,omitempty"`
	State    string `json:"state"`
	// Outcome is a defensive projection of the retained terminal outcome.
	// Running/starting status has no outcome; failed outcomes stay failed.
	Outcome *AgentOutcome `json:"outcome,omitempty"`
}

type PendingQuestion struct {
	AgentID   string `json:"agent_id"`
	Recipient string `json:"recipient"`
	RequestID string `json:"request_id"`
	Text      string `json:"text"`
	Human     bool   `json:"human,omitempty"`
}
type JoinResult struct {
	Outcomes  []AgentOutcome    `json:"outcomes"`
	Questions []PendingQuestion `json:"questions,omitempty"`
}

// Fleet is the library seam used by built-in delegation tools. The application
// supplies construction policy; tools never reach application-owned registries.
type Fleet interface {
	Spawn(context.Context, string, Config) (Agent, error)
	Children(string) []AgentStatus
	Status(string) (AgentStatus, error)
	Outcome(string) (AgentOutcome, error)
	Wait(context.Context, string) (AgentOutcome, error)
	Join(context.Context, string, []string) (JoinResult, error)
	SendMessage(context.Context, string, string, string) (Question, error)
	AskAgent(context.Context, string, string, string) (Result, error)
	AskParent(context.Context, string, string) (string, error)
	AskHuman(context.Context, string, string) (string, error)
	HumanAddress() string
	RespondToQuestion(string, string, string, string) error
	ShutdownAgent(string) error
	ShutdownChildren(string) error
	AcceptSubmission(string, json.RawMessage) error
	CompleteSubmission(string, json.RawMessage, error)
}
