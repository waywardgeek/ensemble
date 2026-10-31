package common

import "time"

// GoalConfig is explicit, process-local configuration. Empty File clears it.
type GoalConfig struct {
	File        string  `json:"file"`
	Enabled     bool    `json:"enabled"`
	IdleSeconds float64 `json:"idle_seconds"`
}

// GoalStatus is a snapshot, not a judgment that the assignment is complete.
type GoalStatus struct {
	GoalConfig
	Phase        string    `json:"phase"`
	NextReminder time.Time `json:"next_reminder,omitempty"`
	Error        string    `json:"error,omitempty"`
}

// GoalController is an optional Agent capability; ordinary Agent implementations
// and callers need not implement or enable reminders.
type GoalController interface {
	ConfigureGoal(GoalConfig) (GoalStatus, error)
	GoalStatus() GoalStatus
	GoalSystemPrompt() (string, error)
}

// GoalChanged publishes accepted state. Replay obtains a fresh GoalStatus.
type GoalChanged struct{ Status GoalStatus }

func (GoalChanged) observation() {}
