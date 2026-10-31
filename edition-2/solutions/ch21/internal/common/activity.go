package common

// Activity describes current work independently of submission or turn state.
type Activity struct {
	// Phase is idle, thinking, tool_call, or waiting.
	Phase string `json:"phase"`
	// Tool identifies the dispatched tool, if any.
	Tool string `json:"tool,omitempty"`
	// Detail is a short display hint, never the full tool input.
	Detail string `json:"detail,omitempty"`
}

const (
	// ActivityIdle means no turn is executing.
	ActivityIdle = "idle"
	// ActivityThinking means a provider request is in flight.
	ActivityThinking = "thinking"
	// ActivityToolCall means a tool is executing.
	ActivityToolCall = "tool_call"
	// ActivityWaiting means the turn is waiting on fleet work.
	ActivityWaiting = "waiting"
)

// ActivityChanged carries the actor's current phase through the existing Observer.
type ActivityChanged struct {
	// Activity is the new display snapshot.
	Activity Activity `json:"activity"`
}

func (ActivityChanged) observation() {}

// FleetChanged asks observers to refresh topology after admission or removal.
type FleetChanged struct{}

func (FleetChanged) observation() {}
