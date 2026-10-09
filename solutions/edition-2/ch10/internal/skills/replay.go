package skills

import (
	"crypto/sha256"
	"fmt"
	"math"
	"reflect"
	"sort"
	"strings"
	"unicode/utf8"

	"example.com/ensemble/internal/common"
)

// Historical reconstructs recorded facts only. It never opens a source or
// renders a variable, and its nil catalog cannot service live operations.
func Historical(parent common.SkillAgent) *Service { return &Service{parent: parent} }

func (s *Service) PrepareRecorded(t common.SkillTransition) (common.SkillCandidate, error) {
	bad := func() (common.SkillCandidate, error) {
		return nil, s.failure("skill_dependency", t.Name, "invalid recorded skill transition")
	}
	st := t.State
	if !skillID(t.Name) || !skillID(st.Primary) || st.Roots == nil || st.Active == nil || st.Available == nil || st.Tools == nil || st.Retired == nil || t.Activated == nil || len(st.Active) > definitionLimit || len(t.Activated) > definitionLimit {
		return bad()
	}
	previous := s.committed
	roots := []string{}
	ceiling := t.Ceiling
	last := uint64(0)
	records := map[uint64]common.SkillMaterial{}
	retained := map[string]common.SkillActive{}
	if previous == nil {
		if t.Action != "initialize" || t.Name != st.Primary || st.Revision != 0 || len(st.Roots) != 0 || len(st.Retired) != 0 || !s.sortedNames(ceiling, true) {
			return bad()
		}
	} else {
		if t.Ceiling != nil || st.Primary != previous.state.Primary || previous.state.Revision == math.MaxUint64 || st.Revision != previous.state.Revision+1 {
			return bad()
		}
		ceiling = previous.ceiling
		last = previous.lastID
		roots = append(roots, previous.state.Roots...)
		for id, m := range previous.material {
			records[id] = m
		}
		for _, a := range previous.state.Active {
			retained[a.Name] = a
		}
		i := sort.SearchStrings(roots, t.Name)
		exists := i < len(roots) && roots[i] == t.Name
		switch t.Action {
		case "load":
			offered := false
			for _, o := range previous.state.Available {
				if o.Name == t.Name {
					offered = true
				}
			}
			if exists || !offered {
				return bad()
			}
			roots = append(roots, t.Name)
			sort.Strings(roots)
		case "unload":
			if !exists || len(t.Activated) != 0 {
				return bad()
			}
			roots = append(roots[:i], roots[i+1:]...)
		default:
			return bad()
		}
	}
	if !reflect.DeepEqual(st.Roots, roots) || !s.sortedNames(st.Roots, false) || !s.sortedNames(st.Tools, true) {
		return bad()
	}
	installed := map[string]bool{}
	for _, n := range ceiling {
		installed[n] = true
	}
	if !installed["load_skill"] || !installed["unload_skill"] {
		return bad()
	}
	if uint64(len(t.Activated)) > math.MaxUint64-last {
		return bad()
	}
	total := 0
	for _, a := range t.Activated {
		last++
		if a.Activation != last || !skillID(a.Name) || a.Type != "primary" && a.Type != "loadable" && a.Type != "dependency" || a.Type == "primary" && a.Body == "" || !utf8.ValidString(a.Body) || strings.ContainsRune(a.Body, 0) || len(a.Body) > sourceLimit || a.SHA256 != fmt.Sprintf("%x", sha256.Sum256([]byte(a.Body))) || !s.sortedNames(a.Tools, true) || a.Dependencies == nil || a.Offers == nil {
			return bad()
		}
		total += len(a.Body)
		if total > catalogLimit {
			return bad()
		}
		for i, id := range a.Dependencies {
			if id == 0 || i > 0 && id <= a.Dependencies[i-1] {
				return bad()
			}
		}
		for i, o := range a.Offers {
			if !skillID(o.Name) || o.Description == "" || len(o.Description) > 256 || !utf8.ValidString(o.Description) || strings.ContainsAny(o.Description, "\r\n\x00") || i > 0 && o.Name <= a.Offers[i-1].Name {
				return bad()
			}
		}
		if _, ok := retained[a.Name]; ok {
			return bad()
		}
		records[a.Activation] = common.SkillMaterial{Record: copyActivation(s, a)}
	}
	active := map[string]common.SkillActivation{}
	ids := map[uint64]bool{}
	for i, a := range st.Active {
		m, ok := records[a.Activation]
		if !ok || ids[a.Activation] || m.Retired || a.Name != m.Record.Name || a.Type != m.Record.Type || i > 0 && a.Name <= st.Active[i-1].Name {
			return bad()
		}
		if old, ok := retained[a.Name]; ok && old != a {
			return bad()
		}
		active[a.Name] = m.Record
		ids[a.Activation] = true
	}
	for _, a := range t.Activated {
		if !ids[a.Activation] {
			return bad()
		}
	}
	// Derive the complete closure and exact deterministic material order from
	// immutable records, rather than trusting the submitted replacement snapshot.
	visiting := map[uint64]uint8{}
	order := []uint64{}
	tools := map[string]bool{"load_skill": true, "unload_skill": true}
	offers := map[string]string{}
	var visit func(string, string) bool
	visit = func(name, kind string) bool {
		a, ok := active[name]
		if !ok || a.Type != kind || visiting[a.Activation] == 1 {
			return false
		}
		if visiting[a.Activation] == 2 {
			return true
		}
		visiting[a.Activation] = 1
		deps := []string{}
		for _, id := range a.Dependencies {
			m, ok := records[id]
			if !ok || !ids[id] || m.Record.Type != "dependency" {
				return false
			}
			deps = append(deps, m.Record.Name)
		}
		sort.Strings(deps)
		for _, n := range deps {
			if !visit(n, "dependency") {
				return false
			}
		}
		for _, n := range a.Tools {
			if !installed[n] {
				return false
			}
			tools[n] = true
		}
		for _, o := range a.Offers {
			if target, ok := active[o.Name]; ok && target.Type != "loadable" {
				return false
			}
			if desc, ok := offers[o.Name]; ok && desc != o.Description {
				return false
			}
			offers[o.Name] = o.Description
		}
		visiting[a.Activation] = 2
		order = append(order, a.Activation)
		return true
	}
	if !visit(st.Primary, "primary") {
		return bad()
	}
	for _, n := range roots {
		if !visit(n, "loadable") {
			return bad()
		}
	}
	if len(order) != len(st.Active) {
		return bad()
	}
	fresh := []uint64{}
	for _, id := range order {
		if previous == nil || id > previous.lastID {
			fresh = append(fresh, id)
		}
	}
	if len(fresh) != len(t.Activated) {
		return bad()
	}
	for i, id := range fresh {
		if id != t.Activated[i].Activation {
			return bad()
		}
	}
	expectedTools := []string{}
	for n := range tools {
		expectedTools = append(expectedTools, n)
	}
	sort.Strings(expectedTools)
	for _, n := range roots {
		delete(offers, n)
	}
	expectedOffers := []common.SkillOffer{}
	for n, d := range offers {
		expectedOffers = append(expectedOffers, common.SkillOffer{Name: n, Description: d})
	}
	sort.Slice(expectedOffers, func(i, j int) bool { return expectedOffers[i].Name < expectedOffers[j].Name })
	retired := []common.SkillRetired{}
	for id, m := range records {
		if !ids[id] {
			if m.Record.Type == "primary" {
				return bad()
			}
			m.Retired = true
			records[id] = m
			retired = append(retired, common.SkillRetired{Name: m.Record.Name, Activation: id})
		}
	}
	sort.Slice(retired, func(i, j int) bool { return retired[i].Activation < retired[j].Activation })
	if !reflect.DeepEqual(expectedTools, st.Tools) || !reflect.DeepEqual(expectedOffers, st.Available) || !reflect.DeepEqual(retired, st.Retired) {
		return bad()
	}
	if previous != nil && t.Action == "load" {
		for _, a := range previous.state.Active {
			if !ids[a.Activation] {
				return bad()
			}
		}
	}
	next := &ledger{state: copyState(s, st), material: records, lastID: last, ceiling: append([]string{}, ceiling...)}
	return &candidate{parent: s, prior: previous, next: next, transition: copyTransition(s, t), result: common.SkillResult{Name: t.Name, Revision: st.Revision, Changed: true}}, nil
}
func (s *Service) sortedNames(names []string, tool bool) bool {
	if names == nil {
		return false
	}
	for i, n := range names {
		if tool && !toolID(n) || !tool && !skillID(n) || i > 0 && n <= names[i-1] {
			return false
		}
	}
	return true
}
