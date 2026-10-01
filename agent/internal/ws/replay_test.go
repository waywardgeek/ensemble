package ws

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/waywardgeek/ensemble/agent/internal/common"
)

// msgEvent builds the smallest event that renderEvent turns into a wire message.
func msgEvent(seq common.Seq, who common.Actor, text string) common.Event {
	return common.Event{
		Seq:  seq,
		Type: common.MessageReceived,
		Time: time.Unix(0, 0).UTC(),
		Message: &common.MessageData{
			Actor: who,
			Parts: common.PartList{common.TextPart{Text: text}},
		},
	}
}

// drainMessages collects the "message" frames already queued on a client.
func drainMessages(c *Client) []string {
	var texts []string
	for {
		select {
		case raw := <-c.send:
			var m map[string]any
			if err := json.Unmarshal(raw, &m); err != nil {
				continue
			}
			if m["type"] == "message" {
				if s, ok := m["text"].(string); ok {
					texts = append(texts, s)
				}
			}
		default:
			return texts
		}
	}
}

// A restored session hands the hub a log that is already full, but no
// observation has arrived yet. logLen used to be advanced only by Observe, so
// it stayed at zero and subscribe replayed nothing: the GUI opened blank on a
// conversation sitting right there on disk. Deleting the seeding in NewHub
// fails this test.
func TestSubscribeReplaysRestoredLog(t *testing.T) {
	log := &common.Log{
		Events: []common.Event{
			msgEvent(1, common.ActorHuman, "first question"),
			msgEvent(2, common.ActorAgent, "first answer"),
		},
		Next:  3,
		Clock: time.Now,
	}

	h := NewHub(nil, nil, "", log, nil, nil)
	c := &Client{send: make(chan []byte, 256)}

	h.subscribe(c)

	texts := drainMessages(c)
	if len(texts) != 2 {
		t.Fatalf("subscribe replayed %d messages, want 2 (got %v)", len(texts), texts)
	}
	if texts[0] != "first question" || texts[1] != "first answer" {
		t.Errorf("replayed in wrong order or wrong content: %v", texts)
	}
}

// The replay burst is larger than the client's buffer. A dropped replay frame
// is never re-sent, so it leaves a permanent hole in the transcript; sendReplay
// must wait for room rather than drop. Restoring the old select/default in the
// replay loop fails this test.
//
// The consumer must NOT start draining immediately: with a reader already
// pulling, a 256-slot buffer never actually fills and the dropping branch is
// never reached. An earlier version of this test made that mistake and passed
// against the bug. So let subscribe run ahead first, and only then read.
func TestSubscribeDoesNotDropWhenBufferOverflows(t *testing.T) {
	const n = 400 // comfortably more than the 256-slot client buffer

	events := make([]common.Event, 0, n)
	for i := 0; i < n; i++ {
		events = append(events, msgEvent(common.Seq(i+1), common.ActorHuman, "m"))
	}
	log := &common.Log{Events: events, Next: common.Seq(n + 1), Clock: time.Now}

	h := NewHub(nil, nil, "", log, nil, nil)
	c := &Client{send: make(chan []byte, 256)}

	subscribeDone := make(chan struct{})
	go func() {
		h.subscribe(c)
		close(subscribeDone)
	}()

	// Let subscribe run ahead until the buffer is full. A dropping send
	// finishes during this window, having discarded everything past slot 256;
	// a waiting send parks here until the reads below make room. Pushing 400
	// small messages into a channel takes well under a millisecond, so this is
	// a very large margin rather than a tuned one.
	time.Sleep(200 * time.Millisecond)

	count := func(raw []byte) bool {
		var m map[string]any
		if err := json.Unmarshal(raw, &m); err != nil {
			return false
		}
		return m["type"] == "message"
	}

	got := 0
	deadline := time.After(10 * time.Second)
collect:
	for got < n {
		select {
		case raw := <-c.send:
			if count(raw) {
				got++
			}
		case <-subscribeDone:
			// Producer is finished. Drain what is left and stop.
			for {
				select {
				case raw := <-c.send:
					if count(raw) {
						got++
					}
				default:
					break collect
				}
			}
		case <-deadline:
			break collect
		}
	}

	if got != n {
		t.Fatalf("replayed %d of %d messages; %d were silently dropped", got, n, n-got)
	}
}


// responseEvent builds a ResponseEnded event with the given parts.
func responseEvent(seq common.Seq, parts ...common.Part) common.Event {
	return common.Event{
		Seq:  seq,
		Type: common.ResponseEnded,
		Time: time.Unix(0, 0).UTC(),
		Response: &common.ResponseData{
			Parts: common.PartList(parts),
		},
	}
}

// drainPartFinals collects part_final frames from a client, returning
// their part_id, kind, and text fields.
type partFinalMsg struct {
	PartID string
	Kind   string
	Text   string
}

func drainPartFinals(c *Client) []partFinalMsg {
	var out []partFinalMsg
	for {
		select {
		case raw := <-c.send:
			var m map[string]any
			if err := json.Unmarshal(raw, &m); err != nil {
				continue
			}
			if m["type"] != "part_final" {
				continue
			}
			pf := partFinalMsg{}
			if v, ok := m["part_id"].(string); ok {
				pf.PartID = v
			}
			if v, ok := m["kind"].(string); ok {
				pf.Kind = v
			}
			if v, ok := m["text"].(string); ok {
				pf.Text = v
			}
			out = append(out, pf)
		default:
			return out
		}
	}
}

// Before the fix, every replayed part_final lacked a part_id, so the GUI's
// artifact Map collapsed all agent responses into one DOM element. This test
// verifies that multiple ResponseEnded events produce distinct part_id values
// and that the text content survives.
func TestResponsesGetDistinctPartIDs(t *testing.T) {
	log := &common.Log{
		Events: []common.Event{
			msgEvent(1, common.ActorHuman, "hello"),
			responseEvent(2, common.TextPart{Text: "first response"}),
			msgEvent(3, common.ActorHuman, "more"),
			responseEvent(4, common.TextPart{Text: "second response"}),
		},
		Next:  5,
		Clock: time.Now,
	}

	h := NewHub(nil, nil, "", log, nil, nil)
	c := &Client{send: make(chan []byte, 256)}
	h.subscribe(c)

	finals := drainPartFinals(c)
	if len(finals) != 2 {
		t.Fatalf("got %d part_finals, want 2: %+v", len(finals), finals)
	}
	if finals[0].PartID == finals[1].PartID {
		t.Errorf("both responses share part_id %q — GUI will collapse them", finals[0].PartID)
	}
	if finals[0].Text != "first response" {
		t.Errorf("first response text = %q, want %q", finals[0].Text, "first response")
	}
	if finals[1].Text != "second response" {
		t.Errorf("second response text = %q, want %q", finals[1].Text, "second response")
	}
}

// Thinking blocks (OpaquePart with type="thinking") should render as
// kind="thinking" on reconnect, the same way they appear during live streaming.
func TestThinkingRendersOnReconnect(t *testing.T) {
	thinkingData, _ := json.Marshal(map[string]any{
		"type":      "thinking",
		"thinking":  "Let me analyze this problem...",
		"signature": "abc123",
	})
	log := &common.Log{
		Events: []common.Event{
			responseEvent(1,
				common.OpaquePart{Data: thinkingData},
				common.TextPart{Text: "Here is my answer"},
			),
		},
		Next:  2,
		Clock: time.Now,
	}

	h := NewHub(nil, nil, "", log, nil, nil)
	c := &Client{send: make(chan []byte, 256)}
	h.subscribe(c)

	finals := drainPartFinals(c)
	if len(finals) != 2 {
		t.Fatalf("got %d part_finals, want 2 (thinking + text): %+v", len(finals), finals)
	}

	// Thinking part comes first.
	if finals[0].Kind != "thinking" {
		t.Errorf("first part kind = %q, want %q", finals[0].Kind, "thinking")
	}
	if finals[0].Text != "Let me analyze this problem..." {
		t.Errorf("thinking text = %q", finals[0].Text)
	}

	// Text part comes second.
	if finals[1].Kind != "text" {
		t.Errorf("second part kind = %q, want %q", finals[1].Kind, "text")
	}
	if finals[1].Text != "Here is my answer" {
		t.Errorf("text content = %q", finals[1].Text)
	}

	// They must have distinct part_ids.
	if finals[0].PartID == finals[1].PartID {
		t.Errorf("thinking and text share part_id %q", finals[0].PartID)
	}
}
