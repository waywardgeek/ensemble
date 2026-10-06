package llm

import (
	"encoding/json"
	"testing"

	"github.com/waywardgeek/ensemble/agent/internal/common"
)

// Reset clears the screen, not the agent.
//
// Every kind of entry lives in one slice and they are told apart only by their
// kind, so the failure this guards against is not subtle: a reset that clears
// the slice would delete the soul document, the memory bands and the loaded
// skills along with the conversation, and it would look like it worked.

func TestResetClearsTheConversationButNotTheMemory(t *testing.T) {
	cleared := []common.EntryKind{common.KindDialogue, common.KindRecall}
	preserved := []common.EntryKind{
		common.KindHandoff, common.KindSkill, common.KindTools,
		common.KindSoul, common.KindMemory, common.Kind64x,
		common.Kind8x, common.KindSession,
	}

	c := &common.Context{}
	for _, k := range append(append([]common.EntryKind{}, cleared...), preserved...) {
		c.Dialogue = append(c.Dialogue, common.Entry{Kind: k})
	}
	before := len(c.Dialogue)

	if err := Apply(c, common.Event{Seq: 1, Type: common.ConversationReset}); err != nil {
		t.Fatalf("Apply: %v", err)
	}

	present := map[common.EntryKind]int{}
	for _, d := range c.Dialogue {
		present[d.Kind]++
	}

	for _, k := range cleared {
		if present[k] != 0 {
			t.Errorf("kind %v survived the reset (%d left), want it cleared", k, present[k])
		}
	}
	for _, k := range preserved {
		if present[k] != 1 {
			t.Errorf("kind %v has %d entries after reset, want 1: reset destroyed something it must keep", k, present[k])
		}
	}
	if len(c.Dialogue) != before-len(cleared) {
		t.Errorf("context has %d entries, want %d", len(c.Dialogue), before-len(cleared))
	}
}

// A reset arriving mid-turn must not leave the machine waiting on a turn whose
// conversation no longer exists.
func TestResetIdlesTheTurn(t *testing.T) {
	c := &common.Context{Turn: common.InFlight}
	if err := Apply(c, common.Event{Seq: 1, Type: common.ConversationReset}); err != nil {
		t.Fatalf("Apply: %v", err)
	}
	if c.Turn != common.Idle {
		t.Errorf("turn state = %v after reset, want Idle", c.Turn)
	}
}

// The event type must survive a save file round trip. Event.UnmarshalJSON
// refuses names it does not know, so a type missing from the name table turns
// every save file written after a reset into one that will not load.
func TestResetEventSurvivesASaveFileRoundTrip(t *testing.T) {
	raw, err := json.Marshal(common.Event{Seq: 7, Type: common.ConversationReset})
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var back common.Event
	if err := json.Unmarshal(raw, &back); err != nil {
		t.Fatalf("unmarshal %s: %v", raw, err)
	}
	if back.Type != common.ConversationReset {
		t.Errorf("type round-tripped to %v, want ConversationReset (wire was %s)", back.Type, raw)
	}
}
