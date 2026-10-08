package skills

import (
	"crypto/sha256"
	"fmt"
	"math"
	"sort"
	"strings"

	"example.com/ensemble/internal/common"
)

// Service is actor-confined after construction. Its ledger is the authority;
// candidates share only immutable retained records and cannot publish themselves.
type Service struct {
	transitions []common.SkillTransitionRef
	parent      common.SkillAgent
	catalog     map[string]common.SkillDefinition
	committed   *ledger
}
type ledger struct {
	state    common.SkillState
	material map[uint64]common.SkillMaterial
	lastID   uint64
	ceiling  []string
}
type candidate struct {
	parent      common.Skills
	prior, next *ledger
	transition  common.SkillTransition
	result      common.SkillResult
}

func New(parent common.SkillAgent) (*Service, error) {
	s := &Service{parent: parent, catalog: map[string]common.SkillDefinition{}}
	if _, err := CopyConfiguration(parent, parent.SkillConfiguration()); err != nil {
		return nil, err
	}
	if err := s.readCatalog(); err != nil {
		return nil, err
	}
	return s, nil
}
func (s *Service) Agent() common.SkillAgent { return s.parent }
func (c *candidate) Transition() common.SkillTransition {
	return copyTransition(c.parent, c.transition)
}
func (c *candidate) Result() common.SkillResult { return c.result }

func (s *Service) Prepare(op common.SkillOperation) (common.SkillCandidate, error) {
	if !skillID(op.Name) {
		return nil, s.failure("invalid_skill_arguments", "", "invalid skill identifier")
	}
	config := s.parent.SkillConfiguration()
	if config == nil {
		return nil, s.failure("skills_disabled", op.Name, "skills disabled")
	}
	previous := s.committed
	roots := []string{}
	revision, lastID := uint64(0), uint64(0)
	active := map[string]common.SkillActive{}
	if previous != nil {
		roots = append(roots, previous.state.Roots...)
		revision, lastID = previous.state.Revision, previous.lastID
		for _, item := range previous.state.Active {
			active[item.Name] = item
		}
	}
	result := common.SkillResult{Name: op.Name, Revision: revision, Changed: true}
	definition, exists := s.catalog[op.Name]
	if op.Action == "initialize" {
		if previous != nil || op.Name != config.Primary || !exists || definition.Type != "primary" {
			return nil, s.failure("skill_unavailable", op.Name, "invalid primary initialization")
		}
		result.Status = "initialized"
	} else {
		if previous == nil {
			return nil, s.failure("skills_disabled", op.Name, "skills not initialized")
		}
		if !exists || definition.Type != "loadable" {
			return nil, s.failure("skill_unavailable", op.Name, "skill is not a dynamic root")
		}
		index := sort.SearchStrings(roots, op.Name)
		loaded := index < len(roots) && roots[index] == op.Name
		switch op.Action {
		case "load":
			if loaded {
				result.Changed = false
				break
			}
			offered := false
			for _, offer := range previous.state.Available {
				if offer.Name == op.Name {
					offered = true
					break
				}
			}
			if !offered {
				return nil, s.failure("skill_unavailable", op.Name, "skill is not currently discoverable")
			}
			roots = append(roots, op.Name)
			sort.Strings(roots)
			result.Status = "loaded"
		case "unload":
			if !loaded {
				result.Changed = false
				break
			}
			roots = append(roots[:index], roots[index+1:]...)
			result.Status = "unloaded"
		default:
			return nil, s.failure("invalid_skill_arguments", op.Name, "invalid skill operation")
		}
		if !result.Changed {
			result.Status = "unchanged"
			return &candidate{parent: s, prior: previous, next: previous, result: result}, nil
		}
		if revision == math.MaxUint64 {
			return nil, s.failure("skill_too_large", op.Name, "skill revision exhausted")
		}
		revision++
	}
	order, tools, offers, err := s.resolve(config.Primary, roots, op.Name)
	if err != nil {
		return nil, err
	}
	newCount := 0
	for _, name := range order {
		if _, ok := active[name]; !ok {
			newCount++
		}
	}
	if uint64(newCount) > math.MaxUint64-lastID {
		return nil, s.failure("skill_too_large", op.Name, "skill activation IDs exhausted")
	}
	state := common.SkillState{Revision: revision, Primary: config.Primary, Roots: roots, Active: []common.SkillActive{}, Available: offers, Tools: tools, Retired: []common.SkillRetired{}}
	next := &ledger{state: state, lastID: lastID, material: map[uint64]common.SkillMaterial{}, ceiling: append([]string{}, s.parent.SkillCeiling()...)}
	if previous != nil {
		next.state.Retired = append(next.state.Retired, previous.state.Retired...)
		for id, record := range previous.material {
			next.material[id] = record
		}
	}
	activated := []common.SkillActivation{}
	current := map[string]uint64{}
	total := 0
	for _, name := range order {
		def := s.catalog[name]
		retained, ok := active[name]
		if ok {
			current[name] = retained.Activation
			next.state.Active = append(next.state.Active, retained)
			continue
		}
		body, err := s.expand(def.Body, tools, offers, config.Variables, op.Name)
		if err != nil {
			return nil, err
		}
		if len(body) > catalogLimit-total {
			return nil, s.failure("skill_too_large", op.Name, "transition material byte limit exceeded")
		}
		total += len(body)
		next.lastID++
		record := common.SkillActivation{Activation: next.lastID, Name: name, Type: def.Type, Body: body, SHA256: fmt.Sprintf("%x", sha256.Sum256([]byte(body))), Tools: append([]string{}, def.Tools...), Dependencies: []uint64{}, Offers: []common.SkillOffer{}}
		for _, dependency := range def.Depends {
			record.Dependencies = append(record.Dependencies, current[dependency])
		}
		sort.Slice(record.Dependencies, func(i, j int) bool { return record.Dependencies[i] < record.Dependencies[j] })
		for _, offered := range def.Loadable {
			record.Offers = append(record.Offers, common.SkillOffer{Name: offered, Description: s.catalog[offered].Description})
		}
		activated = append(activated, record)
		next.material[record.Activation] = common.SkillMaterial{Record: record}
		current[name] = record.Activation
		next.state.Active = append(next.state.Active, common.SkillActive{Name: name, Type: def.Type, Activation: record.Activation})
	}
	for name, old := range active {
		if _, retained := current[name]; retained {
			continue
		}
		next.state.Retired = append(next.state.Retired, common.SkillRetired{Name: name, Activation: old.Activation})
		record := next.material[old.Activation]
		record.Retired = true
		next.material[old.Activation] = record
	}
	sort.Slice(next.state.Active, func(i, j int) bool { return next.state.Active[i].Name < next.state.Active[j].Name })
	sort.Slice(next.state.Retired, func(i, j int) bool { return next.state.Retired[i].Activation < next.state.Retired[j].Activation })
	transition := common.SkillTransition{Action: op.Action, Name: op.Name, State: next.state, Activated: activated}
	if op.Action == "initialize" {
		transition.Ceiling = append([]string{}, s.parent.SkillCeiling()...)
		sort.Strings(transition.Ceiling)
	}
	result.Revision = revision
	return &candidate{parent: s, prior: previous, next: next, transition: transition, result: result}, nil
}

// Apply is an actor-only operation after successful durable append. A foreign or
// stale candidate is a programming error, never a recoverable post-commit error.
func (s *Service) Apply(value common.SkillCandidate, eventSeq uint64) {
	c, ok := value.(*candidate)
	if !ok || c.parent != s || c.prior != s.committed {
		panic("invalid skill candidate application")
	}
	if !c.result.Changed {
		return
	}
	if eventSeq == 0 {
		panic("skill commit requires durable sequence")
	}
	for _, activation := range c.transition.Activated {
		record := c.next.material[activation.Activation]
		record.EventSeq = eventSeq
		c.next.material[activation.Activation] = record
	}
	ids := []uint64{}
	for _, record := range c.transition.Activated {
		ids = append(ids, record.Activation)
	}
	s.transitions = append(s.transitions, common.SkillTransitionRef{Seq: eventSeq, Action: c.transition.Action, Name: c.transition.Name, State: copyState(s, c.transition.State), Activated: ids})
	s.committed = c.next
}

func (s *Service) resolve(primary string, roots []string, requested string) ([]string, []string, []common.SkillOffer, error) {
	ceiling := map[string]bool{}
	for _, name := range s.parent.SkillCeiling() {
		ceiling[name] = true
	}
	for _, name := range []string{"load_skill", "unload_skill"} {
		if !ceiling[name] {
			return nil, nil, nil, s.failure("skill_tool_unavailable", requested, "management handlers must be installed")
		}
	}
	tools := map[string]bool{"load_skill": true, "unload_skill": true}
	offers := map[string]string{}
	visits := map[string]uint8{}
	order := []string{}
	var visit func(string, string) error
	visit = func(name, kind string) error {
		def, exists := s.catalog[name]
		if !exists || def.Type != kind {
			return s.failure("skill_dependency", requested, "missing or wrong-type graph target")
		}
		if visits[name] == 1 {
			return s.failure("skill_dependency", requested, "dependency cycle")
		}
		if visits[name] == 2 {
			return nil
		}
		visits[name] = 1
		for _, dependency := range def.Depends {
			if err := visit(dependency, "dependency"); err != nil {
				return err
			}
		}
		for _, tool := range def.Tools {
			if !ceiling[tool] {
				return s.failure("skill_tool_unavailable", requested, "skill grants an uninstalled handler")
			}
			tools[tool] = true
		}
		for _, name := range def.Loadable {
			offered, exists := s.catalog[name]
			if !exists || offered.Type != "loadable" {
				return s.failure("skill_dependency", requested, "missing or wrong-type offer target")
			}
			offers[name] = offered.Description
		}
		visits[name] = 2
		order = append(order, name)
		return nil
	}
	if err := visit(primary, "primary"); err != nil {
		return nil, nil, nil, err
	}
	for _, root := range roots {
		if err := visit(root, "loadable"); err != nil {
			return nil, nil, nil, err
		}
	}
	// A later root can advertise an earlier root; subtract after the full walk.
	for _, root := range roots {
		delete(offers, root)
	}
	names := make([]string, 0, len(tools))
	for name := range tools {
		names = append(names, name)
	}
	sort.Strings(names)
	available := make([]common.SkillOffer, 0, len(offers))
	for name, description := range offers {
		available = append(available, common.SkillOffer{Name: name, Description: description})
	}
	sort.Slice(available, func(i, j int) bool { return available[i].Name < available[j].Name })
	return order, names, available, nil
}

func (s *Service) expand(body string, tools []string, offers []common.SkillOffer, variables map[string]string, requested string) (string, error) {
	toolLines, offerLines := []string{}, []string{}
	for _, name := range tools {
		toolLines = append(toolLines, "- "+name)
	}
	for _, offer := range offers {
		offerLines = append(offerLines, "- "+offer.Name+": "+offer.Description)
	}
	list := func(lines []string) string {
		if len(lines) == 0 {
			return "(none)"
		}
		return strings.Join(lines, "\n")
	}
	lookup := func(name string) (string, bool) {
		switch name {
		case "TOOLS":
			return list(toolLines), true
		case "SKILLS":
			return list(offerLines), true
		}
		value, ok := variables[name]
		return value, ok
	}
	var out strings.Builder
	write := func(text string) error {
		if len(text) > sourceLimit-out.Len() {
			return s.failure("skill_too_large", requested, "expanded skill body byte limit exceeded")
		}
		out.WriteString(text)
		return nil
	}
	letter := func(b byte) bool { return b >= 'a' && b <= 'z' || b >= 'A' && b <= 'Z' || b == '_' }
	for pos := 0; pos < len(body); {
		start := pos
		for pos < len(body) && body[pos] != '$' {
			pos++
		}
		if err := write(body[start:pos]); err != nil {
			return "", err
		}
		if pos == len(body) {
			break
		}
		pos++
		if pos < len(body) && body[pos] == '$' {
			if err := write("$"); err != nil {
				return "", err
			}
			pos++
			continue
		}
		braced := pos < len(body) && body[pos] == '{'
		if braced {
			pos++
		}
		start = pos
		if pos == len(body) || !letter(body[pos]) {
			return "", s.failure("skill_variable", requested, "malformed variable token")
		}
		pos++
		for pos < len(body) && (letter(body[pos]) || body[pos] >= '0' && body[pos] <= '9') {
			pos++
		}
		name := body[start:pos]
		if braced {
			if pos == len(body) || body[pos] != '}' {
				return "", s.failure("skill_variable", requested, "malformed variable token")
			}
			pos++
		}
		value, ok := lookup(name)
		if !ok {
			return "", s.failure("skill_variable", requested, "unknown skill variable")
		}
		if err := write(value); err != nil {
			return "", err
		}
	}
	return out.String(), nil
}

func copyState(owner common.Skills, state common.SkillState) common.SkillState {
	state.Roots = append([]string{}, state.Roots...)
	state.Active = append([]common.SkillActive{}, state.Active...)
	state.Available = append([]common.SkillOffer{}, state.Available...)
	state.Tools = append([]string{}, state.Tools...)
	state.Retired = append([]common.SkillRetired{}, state.Retired...)
	return state
}
func copyActivation(owner common.Skills, record common.SkillActivation) common.SkillActivation {
	record.Tools = append([]string{}, record.Tools...)
	record.Dependencies = append([]uint64{}, record.Dependencies...)
	record.Offers = append([]common.SkillOffer{}, record.Offers...)
	return record
}
func copyTransition(owner common.Skills, t common.SkillTransition) common.SkillTransition {
	if t.Ceiling != nil {
		t.Ceiling = append([]string{}, t.Ceiling...)
	}
	t.State = copyState(owner, t.State)
	activated := make([]common.SkillActivation, 0, len(t.Activated))
	for _, record := range t.Activated {
		activated = append(activated, copyActivation(owner, record))
	}
	t.Activated = activated
	return t
}

// GrantedTools serves declarations and admission without traversing retained
// material or even the retired summaries required by a complete state view.
func (s *Service) GrantedTools() []string {
	if s.committed == nil {
		return nil
	}
	return append([]string{}, s.committed.state.Tools...)
}

func (s *Service) State() *common.SkillState {
	if s.committed == nil {
		return nil
	}
	state := copyState(s, s.committed.state)
	return &state
}

func (s *Service) Inspect() common.SkillInspection {
	out := common.SkillInspection{Contributors: []common.SkillContributors{}, Material: []common.SkillMaterial{}}
	if s.committed == nil {
		return out
	}
	out.State = s.State()
	for _, record := range s.committed.material {
		record.Record = copyActivation(s, record.Record)
		out.Material = append(out.Material, record)
	}
	sort.Slice(out.Material, func(i, j int) bool { return out.Material[i].Record.Activation < out.Material[j].Record.Activation })
	for _, tool := range out.State.Tools {
		contributors := common.SkillContributors{Tool: tool, Activations: []uint64{}, Mandatory: tool == "load_skill" || tool == "unload_skill"}
		for _, record := range out.Material {
			if record.Retired {
				continue
			}
			for _, grant := range record.Record.Tools {
				if grant == tool {
					contributors.Activations = append(contributors.Activations, record.Record.Activation)
					break
				}
			}
		}
		out.Contributors = append(out.Contributors, contributors)
	}
	return out
}

var _ common.Skills = (*Service)(nil)
