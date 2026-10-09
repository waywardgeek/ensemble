package skills

import (
	"reflect"
	"testing"

	"example.com/ensemble/internal/common"
)

func TestNarrowReadsExcludeRetainedMaterial(t *testing.T) {
	short, long := service(t, fixture()), service(t, fixture())
	for _, s := range []*Service{short, long} {
		if s.State() != nil || s.GrantedTools() != nil {
			t.Fatal("uninitialized read must be nil")
		}
		commit(t, s, "initialize", "base")
	}
	for i := 0; i < 128; i++ {
		commit(t, long, "load", "edit")
		commit(t, long, "unload", "edit")
	}
	for _, s := range []*Service{short, long} {
		commit(t, s, "load", "edit")
	}
	want := short.GrantedTools()
	if !reflect.DeepEqual(want, long.GrantedTools()) {
		t.Fatal("retired history changed grants")
	}
	var escaped []string
	measure := func(s *Service) float64 {
		return testing.AllocsPerRun(30, func() { escaped = s.GrantedTools() })
	}
	small, large := measure(short), measure(long)
	if small != large {
		t.Fatalf("grant allocations depend on retained history: %v vs %v", small, large)
	}
	inspection := testing.AllocsPerRun(30, func() { escaped = long.Inspect().State.Tools })
	if inspection <= large {
		t.Fatal("full-inspection control did not distinguish retained work")
	}
	t.Logf("grant allocations short=%g long=%g; full long inspection=%g; retired=%d", small, large, inspection, len(long.State().Retired))
	escaped = long.GrantedTools()
	escaped[0] = "caller mutation"
	if !reflect.DeepEqual(want, long.GrantedTools()) {
		t.Fatal("grant slice aliases owner")
	}

	before := long.Inspect()
	state := long.State()
	state.Roots[0] = "changed"
	state.Active[0].Name = "changed"
	state.Available[0].Description = "changed"
	state.Tools[0] = "changed"
	state.Retired[0].Name = "changed"
	state.Revision = 0
	if !reflect.DeepEqual(before.State, long.State()) {
		t.Fatal("state aliases owner")
	}
	// An explicit owner-test seam removes access to historical material while
	// retaining the exact state. No live log or valid transition is claimed here.
	material := long.committed.material
	long.committed.material = nil
	isolatedState, isolatedGrants := long.State(), long.GrantedTools()
	long.committed.material = material
	if !reflect.DeepEqual(before.State, isolatedState) || !reflect.DeepEqual(want, isolatedGrants) {
		t.Fatal("narrow reads depend on historical material")
	}

	after := long.Inspect()
	if !reflect.DeepEqual(before, after) || len(after.Material) != 259 {
		t.Fatal("full inspection lost history")
	}
	active := map[uint64]bool{}
	for _, a := range after.State.Active {
		active[a.Activation] = true
	}
	for _, c := range after.Contributors {
		if c.Mandatory != (c.Tool == "load_skill" || c.Tool == "unload_skill") {
			t.Fatal("mandatory contributor changed")
		}
		for _, id := range c.Activations {
			if !active[id] {
				t.Fatal("retired grant still contributes")
			}
		}
		if c.Tool == "write_file" && len(c.Activations) != 1 {
			t.Fatal("current editing contributor missing")
		}
	}
	for i, m := range after.Material {
		if m.Record.Activation != uint64(i+1) {
			t.Fatal("material order changed")
		}
		if len(m.Record.Tools) > 0 {
			after.Material[i].Record.Tools[0] = "mutated"
		}
		if len(m.Record.Dependencies) > 0 {
			after.Material[i].Record.Dependencies[0] = 0
		}
		if len(m.Record.Offers) > 0 {
			after.Material[i].Record.Offers[0].Name = "mutated"
		}
	}
	for i := range after.Contributors {
		after.Contributors[i].Activations = append(after.Contributors[i].Activations, 999)
	}
	if !reflect.DeepEqual(before, long.Inspect()) {
		t.Fatal("full inspection aliases retained values")
	}
}

var _ common.Skills = (*Service)(nil)

func TestLastActivationIsCommittedDurableMaximum(t *testing.T) {
	s := service(t, fixture())
	var owner common.Skills = s
	if owner.LastActivation() != 0 {
		t.Fatal("uninitialized activation maximum")
	}
	commit(t, s, "initialize", "base")
	prepared := prepare(t, s, "load", "edit")
	if owner.LastActivation() != 1 {
		t.Fatal("uncommitted candidate advanced maximum")
	}
	s.Apply(prepared, 2)
	commit(t, s, "unload", "edit")
	want := owner.LastActivation()
	if want != 3 || s.Snapshot().LastID != want {
		t.Fatal("retired activations lost from durable maximum")
	}
	commit(t, s, "unload", "edit") // unchanged candidate burns no identity
	if _, err := s.Prepare(common.SkillOperation{Action: "load", Name: "hidden"}); err == nil {
		t.Fatal("negative candidate control did not refuse")
	}
	if owner.LastActivation() != want {
		t.Fatal("failed or unchanged candidate advanced maximum")
	}
	// Distinguish the scalar operation from building a full retained inspection.
	if allocations := testing.AllocsPerRun(10, func() { owner.LastActivation() }); allocations != 0 {
		t.Fatalf("scalar read allocated an inspection: %g", allocations)
	}
	before := s.Snapshot()
	copy := s.Snapshot()
	copy.Material[0].Record.Body = "caller mutation"
	copy.Transitions[0].State.Active[0].Name = "caller mutation"
	copy.LastID = 0
	if !reflect.DeepEqual(before, s.Snapshot()) || owner.LastActivation() != want {
		t.Fatal("owned snapshot mutation reached Skills")
	}
}
