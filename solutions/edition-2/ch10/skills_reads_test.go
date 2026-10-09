package ensemble

import (
	"reflect"
	"testing"
)

func TestCurrentSkillReadsStayNarrowAndOwned(t *testing.T) {
	app := New(nil)
	defer app.Close()
	makeAgent := func(cycles int) *Agent {
		a, err := app.NewAgent(persistenceSkillConfig(t))
		if err != nil {
			t.Fatal(err)
		}
		for i := 0; i < cycles; i++ {
			if _, err = a.LoadSkill("edit"); err != nil {
				t.Fatal(err)
			}
			if _, err = a.UnloadSkill("edit"); err != nil {
				t.Fatal(err)
			}
		}
		if _, err = a.LoadSkill("edit"); err != nil {
			t.Fatal(err)
		}
		return a
	}
	short, long := makeAgent(1), makeAgent(40)
	var grants []string
	count := func(a *Agent) float64 {
		return testing.AllocsPerRun(20, func() { grants = (toolAgent{a}).GrantedTools() })
	}
	small, large := count(short), count(long)
	if small != large {
		t.Fatalf("Agent grant reads allocate with history: %g/%g", small, large)
	}
	if !reflect.DeepEqual((toolAgent{short}).GrantedTools(), grants) {
		t.Fatal("same active set has different grants")
	}
	grants[0] = "mutation"
	full, err := long.InspectSkills()
	if err != nil {
		t.Fatal(err)
	}
	state, err := long.SkillState()
	if err != nil || !reflect.DeepEqual(state, full.State) {
		t.Fatalf("state differs: %v", err)
	}
	state.Roots[0] = "mutation"
	state.Tools[0] = "mutation"
	state.Active[0].Name = "mutation"
	state.Retired[0].Name = "mutation"
	next, err := long.SkillState()
	if err != nil || !reflect.DeepEqual(next, full.State) {
		t.Fatal("public state aliases owner")
	}
	narrow := testing.AllocsPerRun(20, func() {
		if _, err := long.SkillState(); err != nil {
			t.Fatal(err)
		}
	})
	wide := testing.AllocsPerRun(20, func() {
		if _, err := long.InspectSkills(); err != nil {
			t.Fatal(err)
		}
	})
	if narrow >= wide {
		t.Fatalf("public state rebuilt inspection: state=%g inspect=%g", narrow, wide)
	}
	t.Logf("Agent grants short=%g long=%g allocations; state=%g inspection=%g", small, large, narrow, wide)
	snap, watch, err := long.Watch()
	if err != nil {
		t.Fatal(err)
	}
	defer watch.Close()
	if !reflect.DeepEqual(snap.State.Skills, full.State) {
		t.Fatal("watch state differs")
	}
	snap.State.Skills.Tools[0] = "mutation"
	if _, err = long.UnloadSkill("edit"); err != nil {
		t.Fatal(err)
	}
	next, err = long.SkillState()
	if err != nil || next.Revision != full.State.Revision+1 {
		t.Fatal("actor state query lost committed order")
	}
	for _, name := range next.Tools {
		if name == "write_file" {
			t.Fatal("unload retained write grant")
		}
	}
	// Offline state reads use the same narrow owner view without a live actor.
	offline, err := app.Load(long.config.LogPath, Config{})
	if err != nil {
		t.Fatal(err)
	}
	historical, err := offline.SkillState()
	if err != nil || !reflect.DeepEqual(historical, next) {
		t.Fatal("offline state changed")
	}
	cfg := persistenceSkillConfig(t)
	cfg.Skills = nil
	disabled, err := app.NewAgent(cfg)
	if err != nil {
		t.Fatal(err)
	}
	empty, err := disabled.SkillState()
	if err != nil || empty != nil || (toolAgent{disabled}).GrantedTools() != nil {
		t.Fatal("no-skills nil semantics changed")
	}
}
