package common

// SkillRecordLimit includes every encoded byte and the framing LF, if present.
const SkillRecordLimit = 67_108_864

// SkillConfig contains creation-only inputs. Agent owns their copied values;
// Skills owns parsed definitions, not another mutable configuration copy.
type SkillConfig struct {
	Directory string
	Catalog   map[string][]byte
	Primary   string
	Variables map[string]string
}

type SkillDefinition struct {
	Name, Description, Type, Body string
	Tools, Depends, Loadable      []string
}
type SkillOffer struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}
type SkillActive struct {
	Name       string `json:"name"`
	Type       string `json:"type"`
	Activation uint64 `json:"activation"`
}
type SkillRetired struct {
	Name       string `json:"name"`
	Activation uint64 `json:"activation"`
}
type SkillState struct {
	Revision  uint64         `json:"revision"`
	Primary   string         `json:"primary"`
	Roots     []string       `json:"roots"`
	Active    []SkillActive  `json:"active"`
	Available []SkillOffer   `json:"available"`
	Tools     []string       `json:"tools"`
	Retired   []SkillRetired `json:"retired"`
}
type SkillActivation struct {
	Activation   uint64       `json:"activation"`
	Name         string       `json:"name"`
	Type         string       `json:"type"`
	Body         string       `json:"body"`
	SHA256       string       `json:"sha256"`
	Tools        []string     `json:"tools"`
	Dependencies []uint64     `json:"dependencies"`
	Offers       []SkillOffer `json:"offers"`
}
type SkillMaterial struct {
	Record   SkillActivation
	EventSeq uint64
	Retired  bool
}
type SkillContributors struct {
	Tool        string
	Activations []uint64
	Mandatory   bool
}
type SkillInspection struct {
	State        *SkillState
	Contributors []SkillContributors
	Material     []SkillMaterial
}
type SkillResult struct {
	Status   string `json:"status"`
	Name     string `json:"name"`
	Revision uint64 `json:"revision"`
	Changed  bool   `json:"changed"`
}
type SkillError struct {
	Code, Name string
	Revision   uint64
	Detail     string
}

func (e *SkillError) Error() string { return e.Detail }

type SkillOperation struct{ Action, Name string }
type SkillTransition struct {
	Action    string            `json:"action"`
	Name      string            `json:"name"`
	Ceiling   []string          `json:"ceiling,omitempty"`
	State     SkillState        `json:"state"`
	Activated []SkillActivation `json:"activated"`
}

// SkillAgent is the real owning Agent. These accessors return owned creation
// inputs directly; they must not enqueue to the actor already using Skills.
type SkillAgent interface {
	Agent
	SkillConfiguration() *SkillConfig
	SkillCeiling() []string
}

// Candidates are prepared capabilities, not client-editable replacement state.
// Transition and Result return owned values; only the creating service applies.
type SkillCandidate interface {
	Transition() SkillTransition
	Result() SkillResult
}
type Skills interface {
	Agent() SkillAgent
	Prepare(SkillOperation) (SkillCandidate, error)
	PrepareRecorded(SkillTransition) (SkillCandidate, error)
	Apply(SkillCandidate, uint64)
	Inspect() SkillInspection
}
