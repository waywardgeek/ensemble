package common

// The outbound seam: what leaves the agent.
//
// waywardgeest — the ghost in the machine watches, and what it sees leaves
// through here.
//
// Observe MUST NOT BLOCK. An observer that blocks parks the actor's loop and
// re-creates the exact deafness this chapter exists to remove — the GUI would
// become able to freeze the agent by being slow. A slow observer buffers, or
// drops, on its own time.

// AgentID names which agent an observation came from.
//
// Empty means "the one agent", which is the only case the basic engine builds.
// It is omitempty everywhere, so the single-agent log is byte-identical to one
// written before this field existed.
type AgentID string

// Observer watches one or more agents. It never calls back into the agent.
type Observer interface {
	Observe(Observation)
}

// Observation is what an observer receives. Sealed union, same discipline as
// Part: the interface method is unexported, so only types in this package can
// implement it.
type Observation interface{ isObservation() }

// PartDelta is content arriving incrementally.
//
// THE RULE THAT KEEPS THIS SEAM STABLE: streaming is not a mode. A
// non-streaming vendor emits exactly one delta and then a final. Adding real
// streaming later therefore adds NO new observation kinds — only a different
// chunk count.
type PartDelta struct {
	Agent  AgentID `json:"agent,omitempty"`
	PartID uint64  `json:"part_id"`
	Chunk  string  `json:"chunk"`
}

// PartFinal is the authoritative, complete part.
type PartFinal struct {
	Agent  AgentID `json:"agent,omitempty"`
	Seq    Seq     `json:"seq"`
	PartID uint64  `json:"part_id"`
	Part   Part    `json:"part"`
}

// StateChanged is a transition of the state machine. Chapter 5 does not
// introduce agent state. It EXPOSES it.
type StateChanged struct {
	Agent AgentID   `json:"agent,omitempty"`
	From  TurnState `json:"from"`
	To    TurnState `json:"to"`
}

// TurnEnded signals that a turn completed. Observers use this to know when
// the agent stopped processing and what its final reply text was.
type TurnEnded struct {
	Agent AgentID `json:"agent,omitempty"`
	Text  string  `json:"text"`
	Err   string  `json:"error,omitempty"`
}

func (PartDelta) isObservation()    {}
func (PartFinal) isObservation()    {}
func (StateChanged) isObservation() {}
func (TurnEnded) isObservation()    {}
