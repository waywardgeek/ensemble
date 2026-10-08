package common

import "context"

const WatchItems = 256
const WatchBytes = 128 * 1024 * 1024

type PauseState struct {
	Revision        uint64 `json:"revision"`
	Paused          bool   `json:"paused"`
	TypingClients   int    `json:"typing_clients"`
	SpeakingClients int    `json:"speaking_clients"`
}
type PauseRegistration interface {
	Actor() Actor
	Update(typing, speaking bool) (PauseState, error)
	Close() error
}
type ActiveOperation struct {
	AgentID     string `json:"agent_id"`
	RequestID   string `json:"request_id"`
	OperationID string `json:"operation_id"`
	Delivery    string `json:"delivery"`
}
type WatchPartial struct {
	ActiveOperation
	PartID   int               `json:"part_id"`
	Channels map[string]string `json:"channels"`
}
type UsageAccount struct {
	From  Provenance `json:"from"`
	Usage Usage      `json:"usage"`
}
type WatchState struct {
	Skills                 *SkillState      `json:"skills"`
	ExecutionPolicy        PolicySnapshot   `json:"execution_policy"`
	ActiveMaxModelRequests *int             `json:"active_max_model_requests"`
	Lifecycle              string           `json:"lifecycle"`
	ActiveRequestID        *string          `json:"active_request_id"`
	ActiveOperation        *ActiveOperation `json:"active_operation"`
	QueuedRequestIDs       []string         `json:"queued_request_ids"`
	Paused                 bool             `json:"paused"`
	TypingClients          int              `json:"typing_clients"`
	SpeakingClients        int              `json:"speaking_clients"`
	Model                  string           `json:"model"`
	Usage                  []UsageAccount   `json:"usage"`
}
type WatchSnapshot struct {
	AgentID    string         `json:"agent_id"`
	Generation string         `json:"generation"`
	Watermark  uint64         `json:"watermark"`
	FirstSeq   *uint64        `json:"first_seq"`
	LastSeq    *uint64        `json:"last_seq"`
	LogSeq     uint64         `json:"log_seq"`
	Omitted    int            `json:"omitted"`
	State      WatchState     `json:"state"`
	Events     []Event        `json:"events"`
	Partials   []WatchPartial `json:"partials"`
}
type WatchRecord struct {
	Revision    uint64      `json:"revision"`
	Observation Observation `json:"observation"`
}

// Watch has one consumer. Close invalidates queued records and wakes Next.
// Done and Status expose loss independently of a blocked consumer.
type Watch interface {
	Actor() Actor
	Next(context.Context) (WatchRecord, error)
	Done() <-chan struct{}
	Status() string
	Close() error
}
