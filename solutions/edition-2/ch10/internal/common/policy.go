package common

import "encoding/json"

const DefaultMaxModelRequests = 16

type TurnPolicy struct {
	Revision                  uint64 `json:"revision"`
	MaxModelRequests          int    `json:"max_model_requests"`
	EffectiveMaxModelRequests int    `json:"effective_max_model_requests"`
}
type PolicySnapshot struct {
	Revision                  uint64 `json:"revision"`
	Persistent                bool   `json:"persistent"`
	MaxModelRequests          int    `json:"max_model_requests"`
	EffectiveMaxModelRequests int    `json:"effective_max_model_requests"`
}
type PolicyAck struct {
	Revision      uint64 `json:"revision"`
	WatchRevision uint64 `json:"watch_revision"`
}
type SettingsError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Domain  string `json:"domain,omitempty"`
	Current any    `json:"current,omitempty"`
}

func (e *SettingsError) Error() string { return e.Message }

type PolicyService interface {
	Agent() Agent
	Snapshot() PolicySnapshot
	Prepare(uint64, json.RawMessage) (PolicySnapshot, bool, error)
	Persist(PolicySnapshot) error
	Apply(PolicySnapshot, error)
	Close()
}
