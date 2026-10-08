package ensemble

import (
	"example.com/ensemble/internal/common"
	"fmt"
)

type SkillConfig = common.SkillConfig
type SkillDefinition = common.SkillDefinition
type SkillOffer = common.SkillOffer
type SkillActive = common.SkillActive
type SkillRetired = common.SkillRetired
type SkillState = common.SkillState
type SkillActivation = common.SkillActivation
type SkillMaterial = common.SkillMaterial
type SkillContributors = common.SkillContributors
type SkillInspection = common.SkillInspection
type SkillResult = common.SkillResult
type SkillError = common.SkillError
type SkillOperation = common.SkillOperation
type SkillTransition = common.SkillTransition

// These adapters borrow only on the serialized owner path. Calling the public
// mailbox getters here would synchronously enqueue back into the same actor.
type skillAgent struct{ *Agent }

func (a skillAgent) SkillConfiguration() *common.SkillConfig { return a.config.Skills }
func (a skillAgent) SkillCeiling() []string                  { return a.config.Builtins }
func (a toolAgent) GrantedTools() []string {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.skills == nil {
		return nil
	}
	state := a.skills.Inspect().State
	if state == nil {
		return nil
	}
	return state.Tools
}
func (a turnAgent) SkillView() common.SkillInspection {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.skills == nil {
		return common.SkillInspection{}
	}
	return a.skills.Inspect()
}
func (a turnAgent) ChangeSkill(op common.SkillOperation) (common.SkillResult, error) {
	a.mu.Lock()
	if a.skills == nil {
		a.mu.Unlock()
		return common.SkillResult{}, &common.SkillError{Code: "skills_disabled", Name: op.Name, Detail: "skills disabled"}
	}
	candidate, err := a.skills.Prepare(op)
	a.mu.Unlock()
	if err != nil {
		return common.SkillResult{}, err
	}
	result := candidate.Result()
	if !result.Changed {
		return result, nil
	}
	fact := candidate.Transition()
	if err = a.appendPrepared(Event{Type: "skills_changed", Skills: &fact}, true, true, nil, candidate); err != nil {
		return common.SkillResult{}, err
	}
	return result, nil
}
func (a *Agent) InspectSkills() (SkillInspection, error) {
	if a.actor != nil {
		return a.actor.InspectSkills()
	}
	return (turnAgent{a}).SkillView(), nil
}
func (a *Agent) SkillState() (*SkillState, error) {
	view, err := a.InspectSkills()
	return view.State, err
}
func (a *Agent) LoadSkill(name string) (SkillResult, error) {
	return a.changeSkill(common.SkillOperation{Action: "load", Name: name})
}
func (a *Agent) UnloadSkill(name string) (SkillResult, error) {
	return a.changeSkill(common.SkillOperation{Action: "unload", Name: name})
}
func (a *Agent) changeSkill(op common.SkillOperation) (SkillResult, error) {
	if a.actor == nil {
		return SkillResult{}, fmt.Errorf("Agent is read-only")
	}
	return a.actor.ChangeSkill(op)
}
func (e *Ensemble) SkillState(id string) (*SkillState, error) {
	a, err := e.Agent(id)
	if err != nil {
		return nil, err
	}
	return a.SkillState()
}
func (e *Ensemble) InspectSkills(id string) (SkillInspection, error) {
	a, err := e.Agent(id)
	if err != nil {
		return SkillInspection{}, err
	}
	return a.InspectSkills()
}
func (e *Ensemble) LoadSkill(id, name string) (SkillResult, error) {
	a, err := e.Agent(id)
	if err != nil {
		return SkillResult{}, err
	}
	return a.LoadSkill(name)
}
func (e *Ensemble) UnloadSkill(id, name string) (SkillResult, error) {
	a, err := e.Agent(id)
	if err != nil {
		return SkillResult{}, err
	}
	return a.UnloadSkill(name)
}
