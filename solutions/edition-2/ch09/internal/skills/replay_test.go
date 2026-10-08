package skills

import (
	"encoding/json"
	"example.com/ensemble/internal/common"
	"reflect"
	"testing"
)

func TestRecordedTransitionsWithoutCatalog(t *testing.T) {
	live := service(t, fixture())
	history := Historical(newOwner(t, nil))
	for i, op := range []common.SkillOperation{{Action: "initialize", Name: "base"}, {Action: "load", Name: "edit"}, {Action: "load", Name: "search"}, {Action: "load", Name: "review"}, {Action: "unload", Name: "edit"}, {Action: "unload", Name: "review"}, {Action: "load", Name: "edit"}} {
		c := prepare(t, live, op.Action, op.Name)
		wire, err := json.Marshal(c.Transition())
		if err != nil {
			t.Fatal(err)
		}
		var fact common.SkillTransition
		if err = json.Unmarshal(wire, &fact); err != nil {
			t.Fatal(err)
		}
		replay, err := history.PrepareRecorded(fact)
		if err != nil {
			t.Fatalf("%s: %v", op.Action, err)
		}
		live.Apply(c, uint64(i+1))
		history.Apply(replay, uint64(i+1))
		if !reflect.DeepEqual(live.Inspect(), history.Inspect()) {
			t.Fatal("historical projection differs")
		}
	}
}
func TestRecordedMutationRefusesBeforePublication(t *testing.T) {
	live := service(t, fixture())
	initial := commit(t, live, "initialize", "base").Transition()
	load := prepare(t, live, "load", "edit").Transition()
	for name, mutate := range map[string]func(*common.SkillTransition){
		"hidden root":         func(x *common.SkillTransition) { x.Name = "hidden" },
		"unrelated root":      func(x *common.SkillTransition) { x.State.Roots = append(x.State.Roots, "review") },
		"reused ID":           func(x *common.SkillTransition) { x.Activated[0].Activation = 1 },
		"altered digest":      func(x *common.SkillTransition) { x.Activated[0].Body += "tampered" },
		"missing management":  func(x *common.SkillTransition) { x.State.Tools = x.State.Tools[1:] },
		"replacement ceiling": func(x *common.SkillTransition) { x.Ceiling = []string{"load_skill", "unload_skill"} },
		"retire primary":      func(x *common.SkillTransition) { x.State.Active = x.State.Active[1:] },
		"extra material":      func(x *common.SkillTransition) { x.Activated = append(x.Activated, x.Activated[0]) },
		"revision leap":       func(x *common.SkillTransition) { x.State.Revision++ },
	} {
		t.Run(name, func(t *testing.T) {
			h := Historical(newOwner(t, nil))
			c, err := h.PrepareRecorded(initial)
			if err != nil {
				t.Fatal(err)
			}
			h.Apply(c, 1)
			before := h.Inspect()
			good, err := h.PrepareRecorded(load)
			if err != nil || good == nil {
				t.Fatal("positive control", err)
			}
			bad := copyTransition(live, load)
			mutate(&bad)
			if _, err = h.PrepareRecorded(bad); err == nil {
				t.Fatal("accepted mutation")
			}
			if !reflect.DeepEqual(before, h.Inspect()) {
				t.Fatal("failed preparation mutated state")
			}
		})
	}
}
func TestRecordedJSONShapes(t *testing.T) {
	s := service(t, fixture())
	wire, _ := json.Marshal(prepare(t, s, "initialize", "base").Transition())
	var raw map[string]json.RawMessage
	_ = json.Unmarshal(wire, &raw)
	for _, field := range []string{"state", "activated", "ceiling", "name", "action"} {
		copy := map[string]json.RawMessage{}
		for k, v := range raw {
			copy[k] = v
		}
		delete(copy, field)
		bad, _ := json.Marshal(copy)
		var out common.SkillTransition
		if json.Unmarshal(bad, &out) == nil {
			t.Fatalf("missing %s accepted", field)
		}
	}
	for _, bad := range []string{`{"name":"edit","description":"One","description":"Two"}`, `{"name":"edit","description":null}`, `{"name":"edit","description":"One","unknown":1}`} {
		var out common.SkillOffer
		if json.Unmarshal([]byte(bad), &out) == nil {
			t.Fatal("invalid nested shape accepted")
		}
	}
}
