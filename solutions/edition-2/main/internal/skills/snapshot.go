package skills

import (
	"crypto/sha256"
	"encoding/json"
	"example.com/ensemble/internal/common"
	"fmt"
	"reflect"
	"sort"
)

func (s *Service) Identity() (common.SkillIdentity, error) {
	defs := make([]any, 0, len(s.catalog))
	names := []string{}
	for n := range s.catalog {
		names = append(names, n)
	}
	sort.Strings(names)
	for _, n := range names {
		d := s.catalog[n]
		defs = append(defs, map[string]any{"name": d.Name, "description": d.Description, "type": d.Type, "body": d.Body, "tools": append([]string{}, d.Tools...), "depends": append([]string{}, d.Depends...), "loadable-skills": append([]string{}, d.Loadable...)})
	}
	hash := func(v any) (string, error) {
		raw, err := json.Marshal(v)
		if err != nil {
			return "", err
		}
		raw, err = s.parent.Codec().Canonical(raw)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("%x", sha256.Sum256(raw)), nil
	}
	catalog, err := hash(defs)
	if err != nil {
		return common.SkillIdentity{}, err
	}
	variables := s.parent.SkillConfiguration().Variables
	if variables == nil {
		variables = map[string]string{}
	}
	bindings, err := hash(variables)
	return common.SkillIdentity{Primary: s.parent.SkillConfiguration().Primary, CatalogSHA256: catalog, BindingsSHA256: bindings}, err
}
func (s *Service) Snapshot() *common.SkillSnapshot {
	if s.committed == nil {
		return nil
	}
	inspection := s.Inspect()
	out := &common.SkillSnapshot{State: *inspection.State, Material: []common.SkillSnapshotMaterial{}, Transitions: []common.SkillTransitionRef{}, LastID: s.committed.lastID, Ceiling: append([]string{}, s.committed.ceiling...), Contributors: inspection.Contributors}
	sort.Strings(out.Ceiling)
	for _, m := range inspection.Material {
		out.Material = append(out.Material, common.SkillSnapshotMaterial{Record: m.Record, EventSeq: m.EventSeq, Retired: m.Retired})
	}
	for _, tr := range s.transitions {
		tr.State = copyState(s, tr.State)
		tr.Activated = append([]uint64{}, tr.Activated...)
		out.Transitions = append(out.Transitions, tr)
	}
	return out
}

// Restore validates each historical action using immutable records. The private
// service is not registered, and it starts no Actor, callback, process or worker.
func (s *Service) Restore(snapshot *common.SkillSnapshot, live bool) error {
	bad := func() error {
		return &common.SessionError{Code: "session_corrupt", Detail: "invalid semantic skill history"}
	}
	if snapshot == nil {
		return bad()
	}
	if len(snapshot.Material) > 1000000 || len(snapshot.Transitions) > 1000000 {
		return bad()
	}
	records := map[uint64]common.SkillSnapshotMaterial{}
	var last uint64
	for _, m := range snapshot.Material {
		if m.Record.Activation <= last || m.EventSeq == 0 {
			return bad()
		}
		last = m.Record.Activation
		records[last] = m
	}
	if last != snapshot.LastID {
		return bad()
	}
	var seq uint64
	s.committed = nil
	s.transitions = nil
	for i, tr := range snapshot.Transitions {
		if tr.Seq <= seq {
			return bad()
		}
		seq = tr.Seq
		t := common.SkillTransition{Action: tr.Action, Name: tr.Name, State: tr.State, Activated: []common.SkillActivation{}}
		if i == 0 {
			t.Ceiling = append([]string{}, snapshot.Ceiling...)
		}
		for _, id := range tr.Activated {
			m, ok := records[id]
			if !ok || m.EventSeq != tr.Seq {
				return bad()
			}
			t.Activated = append(t.Activated, m.Record)
		}
		candidate, err := s.PrepareRecorded(t)
		if err != nil {
			return &common.SessionError{Code: "session_corrupt", Detail: fmt.Sprintf("invalid skill transition at %d: %v", tr.Seq, err)}
		}
		if live {
			current, err := s.Prepare(common.SkillOperation{Action: t.Action, Name: t.Name})
			if err != nil || !reflect.DeepEqual(current.Transition(), t) {
				return &common.SessionError{Code: "session_incompatible", Detail: "recorded skill material differs from frozen candidate"}
			}
		}
		s.Apply(candidate, tr.Seq)
	}
	got := s.Snapshot()
	if got == nil || !reflect.DeepEqual(got, snapshot) {
		return &common.SessionError{Code: "session_corrupt", Detail: "restored skill projection mismatch"}
	}
	return nil
}
