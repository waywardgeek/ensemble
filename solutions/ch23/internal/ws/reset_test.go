package ws

import (
	"encoding/json"
	"testing"

	"github.com/waywardgeek/ensemble/agent/internal/common"
)

// A reset reaches the screen through the ordinary event stream rather than as
// a reply to the button press. That is what makes a reconnecting client, which
// rebuilds its page by replaying the log, land on the same cleared screen
// instead of a stale transcript the agent no longer has.
func TestResetReachesTheScreenAsAnEvent(t *testing.T) {
	e := common.Event{Seq: 9, Type: common.ConversationReset}

	if !isRenderable(e) {
		t.Fatal("ConversationReset is not renderable: the agent would clear its context and the screen would keep showing the old conversation")
	}

	msgs := renderEvent(e)
	if len(msgs) != 1 {
		t.Fatalf("renderEvent produced %d wire messages, want exactly 1", len(msgs))
	}

	var got map[string]any
	if err := json.Unmarshal(msgs[0], &got); err != nil {
		t.Fatalf("unmarshal %s: %v", msgs[0], err)
	}
	if got["type"] != "conversation_reset" {
		t.Errorf("wire type = %v, want %q (full message: %s)", got["type"], "conversation_reset", msgs[0])
	}
}
