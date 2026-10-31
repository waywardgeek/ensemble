package ensemble

import (
	"bytes"
	"encoding/json"
	"ensemble/internal/common"
	"ensemble/internal/fleet"
	"errors"
)

// These instructions belong to delegated Agents, never the human-facing root.
const subAgentInstructions = `You are a delegated sub-agent. Ask your parent questions with send_message_to_parent
and answer child questions with respond_to_parent_message. Join or shut down
unfinished children, then call submit with your completed result. Ordinary text
does not submit your work.`

// AgentFactory supplies application-specific child configuration and construction.
type AgentFactory func(parent Agent, spec AgentSpec) (Agent, error)

// FleetMessage carries sender identity for questions, replies and escalation.
type FleetMessage = common.FleetMessage

// UserSink receives questions and escalation addressed to a human.
type UserSink = common.UserSink

// ChildStatus reports a direct child's explicit completion state.
type ChildStatus = common.ChildStatus

// Submission contains default result text or a custom raw JSON payload.
type Submission = common.Submission

// SetAgentFactory enables fleet execution before Agents are started. The callback
// must construct through NewChildAgent so both actual parent links are retained.
func (e *Ensemble) SetAgentFactory(factory AgentFactory) { e.factory = factory }

// FleetEnabled distinguishes a fleet application from earlier single-Agent clients.
func (e *Ensemble) FleetEnabled() bool { return e.factory != nil }

// SetUserSink selects the human address and delivery boundary before startup.
func (e *Ensemble) SetUserSink(id string, sink UserSink) { e.userID = id; e.userSink = sink }

// UserID returns the configured human address, defaulting to the role user.
func (e *Ensemble) UserID() string {
	if e.userID != "" {
		return e.userID
	}
	return "user"
}

// DeliverUser reaches the application sink; an unwired human is a visible error.
func (e *Ensemble) DeliverUser(message FleetMessage) error {
	if e.userSink == nil {
		return errors.New("no user message sink configured")
	}
	return e.userSink.Deliver(message)
}

// CreateChild delegates construction, leaving provider choices in the application.
func (e *Ensemble) CreateChild(parent Agent, spec common.AgentSpec) (Agent, error) {
	if parent == nil || parent.Ensemble() != e {
		return nil, errors.New("child requires parent in this Ensemble")
	}
	if e.factory == nil {
		return nil, errors.New("no agent factory configured")
	}
	// Direct public callers and Fleet spawning share this boundary. Validate
	// before the factory can create files, load skills or register an Agent.
	contract, err := fleet.CompileSubmissionSchema(spec.SubmitSchema)
	if err != nil {
		return nil, err
	}
	var expected json.RawMessage
	if len(spec.SubmitSchema) != 0 {
		expected = contract.Schema()
		spec.SubmitSchema = contract.Schema()
	}
	a, err := e.factory(parent, AgentSpec(spec))
	if err != nil {
		return nil, err
	}
	if a == nil {
		return nil, errors.New("factory returned no child")
	}
	if a.ParentAgent() != parent || a.Ensemble() != e {
		// A bad factory may return the parent or an unrelated live Agent.
		// Refusal does not give us ownership of that Agent's lifetime.
		return nil, errors.New("factory returned child with wrong ownership")
	}
	// Factories choose model configuration, not a replacement submit contract.
	// Keep our comparison bytes independent of the factory's mutable input.
	if !bytes.Equal(a.Spec().SubmitSchema, expected) || !bytes.Equal(a.Fleet().SubmitSchema(), contract.Schema()) {
		a.Shutdown()
		return nil, errors.New("factory changed submission contract")
	}
	return a, nil
}

// NewChildAgent enforces inherited restrictions at the construction boundary.
// Routing IDs complement these pointers; they never replace actual ownership.
func (e *Ensemble) NewChildAgent(parent Agent, cfg Config, spec AgentSpec) (Agent, error) {
	if parent == nil || parent.Ensemble() != e {
		return nil, errors.New("child requires parent in this Ensemble")
	}
	narrowed, err := spec.Clamp(AgentSpec(parent.Spec()))
	if err != nil {
		return nil, err
	}
	return e.newAgent(parent, cfg, narrowed)
}
