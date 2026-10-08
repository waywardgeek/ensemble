package gui

import (
	"encoding/json"
	"example.com/ensemble"
	"strings"
	"testing"
)

func TestSkillProjectionKeepsUint64MaterialIdentities(t *testing.T) {
	app := ensemble.New(nil)
	defer app.Close()
	server := &Server{parent: app}
	state := ensemble.SkillState{Revision: ^uint64(0), Active: []ensemble.SkillActive{{Name: "one", Activation: 9007199254740992}}, Retired: []ensemble.SkillRetired{{Name: "two", Activation: 9007199254740993}}}
	transition := ensemble.SkillTransition{State: state, Activated: []ensemble.SkillActivation{{Activation: 9007199254740993, Body: "revision:9007199254740993 is literal", Dependencies: []uint64{9007199254740992}}}}
	event := ensemble.Event{Type: "skills_changed", Skills: &transition}
	wire, err := json.Marshal(projectObservation(server, ensemble.Observation{Kind: "skills_changed", Skills: &state, Event: event}))
	if err != nil {
		t.Fatal(err)
	}
	for _, fragment := range []string{`"revision":18446744073709551615`, `"activation":9007199254740992`, `"activation":9007199254740993`, `"dependencies":[9007199254740992]`, `"body":"revision:9007199254740993 is literal"`} {
		if !strings.Contains(string(wire), fragment) {
			t.Fatalf("lost %s in %s", fragment, wire)
		}
	}
}
