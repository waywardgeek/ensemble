package gui

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/waywardgeek/ensemble/agent/internal/common"
)

// Replay is deliberately complete: it does NOT window the log. A reconnecting
// client rebuilds the cleared screen because the reset is itself an event, so
// the stale transcript is replayed and then cleared, in order. That ordering
// is the whole reason a refresh is self-healing, so it is worth pinning.
//
// The live path cannot rely on this (it replays nothing), which is why the
// actor also notifies ConversationCleared.
func TestReplayCarriesResetInOrder(t *testing.T) {
	log := &common.Log{
		Events: []common.Event{
			msgEvent(1, common.ActorHuman, "old question"),
			msgEvent(2, common.ActorAgent, "old answer"),
			{Seq: 3, Type: common.ConversationReset, Time: time.Unix(0, 0).UTC()},
			msgEvent(4, common.ActorHuman, "new question"),
		},
		Next:  5,
		Clock: time.Now,
	}

	h := New(AgentHooks{EventLog: log}, "")
	c := &Client{send: make(chan []byte, 256)}

	h.subscribe(c)
	close(c.send)

	var kinds []string
	for raw := range c.send {
		var m struct {
			Type string `json:"type"`
		}
		if err := json.Unmarshal(raw, &m); err != nil {
			continue
		}
		if m.Type == "message" || m.Type == "conversation_reset" {
			kinds = append(kinds, m.Type)
		}
	}

	// The reset must arrive after the messages it cleared, and before the
	// one that followed it. Drop the reset and the stale transcript survives
	// the refresh.
	want := []string{"message", "message", "conversation_reset", "message"}
	if len(kinds) != len(want) {
		t.Fatalf("replay frames = %v, want %v", kinds, want)
	}
	for i := range want {
		if kinds[i] != want[i] {
			t.Fatalf("replay frames = %v, want %v", kinds, want)
		}
	}
}
