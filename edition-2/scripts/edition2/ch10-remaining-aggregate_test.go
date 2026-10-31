package ensemble

// These are component controls over owned semantic values, not million-turn
// public imports. No reduced constant or invented admission index is used.
import (
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"testing"

	"example.com/ensemble/internal/common"
	"example.com/ensemble/internal/llm"
)

func ch10RemainingOwner(t *testing.T) (*Agent, common.Context) {
	t.Helper()
	e := New(io.Discard)
	t.Cleanup(func() { _ = e.Close() })
	w := t.TempDir()
	a, err := e.OpenSession(SessionOptions{Config: Config{Vendor: "openai", Model: "remaining-local", APIKey: "fixture-not-a-credential", BaseURL: "http://127.0.0.1:1", Workspace: w, DataDir: filepath.Join(w, "store")}})
	if err != nil {
		t.Fatal(err)
	}
	x, err := a.ExportCheckpoint()
	if err != nil {
		t.Fatal(err)
	}
	cp, err := a.Codec().Decode(x.Bytes)
	if err != nil {
		t.Fatal(err)
	}
	return a, cp.State.Context
}

func ch10RemainingCode(t *testing.T, err error, code string) {
	t.Helper()
	var failure *common.SessionError
	if !errors.As(err, &failure) || failure.Code != code {
		t.Fatalf("wanted %s; got %v", code, err)
	}
}

func TestCh10RemainingAggregateEntries(t *testing.T) {
	a, c := ch10RemainingOwner(t)
	// Split collections: neither individual list reaches the aggregate bound.
	c.Entries = make([]common.Entry, 500000)
	c.Instructions = make([]common.Entry, 500000)
	for i := range c.Entries {
		c.Entries[i] = common.Entry{Seq: uint64(i + 2), Actor: "human", Purpose: "dialogue", Parts: []common.Part{}}
		c.Instructions[i] = common.Entry{Seq: uint64(i + 500002), Actor: "system", Purpose: "instruction", Parts: []common.Part{}}
	}
	c.LastSeq = 1000001
	if err := llm.ValidateSemantic(a.engine, c); err != nil {
		t.Fatalf("million-entry semantic positive: %v", err)
	}
	llm.ReindexContext(a.engine, &c)
	if c.Index.Entries != 1000000 {
		t.Fatalf("derived entry count: %d", c.Index.Entries)
	}
	more := common.Entry{Seq: 1000002, Actor: "system", Purpose: "instruction", Parts: []common.Part{}}
	c.Instructions = c.Instructions[:499999]
	llm.ReindexContext(a.engine, &c)
	if err := llm.ValidateSessionCollections(a.engine, c, common.Event{Type: "message_received", Message: &more}); err != nil {
		t.Fatalf("exact entry admission: %v", err)
	}
	c.Instructions = c.Instructions[:500000]
	llm.ReindexContext(a.engine, &c)
	ch10RemainingCode(t, llm.ValidateSessionCollections(a.engine, c, common.Event{Type: "message_received", Message: &more}), "session_limit")
	c.Instructions = append(c.Instructions, more)
	c.LastSeq++
	ch10RemainingCode(t, llm.ValidateSemantic(a.engine, c), "session_corrupt")
}

func TestCh10RemainingAggregateParts(t *testing.T) {
	a, c := ch10RemainingOwner(t)
	parts := make([]common.Part, 500000)
	for i := range parts {
		parts[i] = Text("")
	}
	// Shared immutable input backing avoids two unnecessary half-million
	// allocations; the semantic counter still visits one million actual slots.
	c.Entries = []common.Entry{{Seq: 2, Actor: "human", Purpose: "dialogue", Parts: parts}}
	c.Instructions = []common.Entry{{Seq: 3, Actor: "system", Purpose: "instruction", Parts: parts}}
	c.LastSeq = 3
	if err := llm.ValidateSemantic(a.engine, c); err != nil {
		t.Fatalf("million-part semantic positive: %v", err)
	}
	llm.ReindexContext(a.engine, &c)
	if c.Index.Parts != 1000000 {
		t.Fatalf("derived part count: %d", c.Index.Parts)
	}
	more := common.Entry{Seq: 4, Actor: "system", Purpose: "instruction", Parts: []common.Part{Text("")}}
	c.Instructions[0].Parts = parts[:499999]
	llm.ReindexContext(a.engine, &c)
	if err := llm.ValidateSessionCollections(a.engine, c, common.Event{Type: "message_received", Message: &more}); err != nil {
		t.Fatalf("exact part admission: %v", err)
	}
	c.Instructions[0].Parts = parts
	llm.ReindexContext(a.engine, &c)
	ch10RemainingCode(t, llm.ValidateSessionCollections(a.engine, c, common.Event{Type: "message_received", Message: &more}), "session_limit")
	c.Instructions = append(c.Instructions, more)
	c.LastSeq++
	ch10RemainingCode(t, llm.ValidateSemantic(a.engine, c), "session_corrupt")
}

func TestCh10RemainingSeenRequests(t *testing.T) {
	a, c := ch10RemainingOwner(t)
	c.TurnIDs = make(map[string]bool, 1000001)
	c.Turns = make(map[string]common.TurnFact, 1000001)
	for i := 1; i <= 1000000; i++ {
		id := fmt.Sprintf("r%d", i)
		c.TurnIDs[id] = true
		c.Turns[id] = common.TurnFact{Index: uint64(i), Start: uint64(2 * i), End: uint64(2*i + 1), Outcome: "canceled"}
	}
	c.LastSeq, c.RequestCursor = 2000001, 1000000
	if err := llm.ValidateSemantic(a.engine, c); err != nil {
		t.Fatalf("million-request semantic positive: %v", err)
	}
	llm.ReindexContext(a.engine, &c)
	next := common.Event{Type: "turn_started", Turn: &common.TurnEvent{RequestID: "next", RequestIndex: 1000001}}
	delete(c.TurnIDs, "r1000000")
	if err := llm.ValidateSessionCollections(a.engine, c, next); err != nil {
		t.Fatalf("exact request admission: %v", err)
	}
	c.TurnIDs["r1000000"] = true
	ch10RemainingCode(t, llm.ValidateSessionCollections(a.engine, c, next), "session_limit")
	c.TurnIDs["last"] = true
	c.Turns["last"] = common.TurnFact{Index: 1000001, Start: 2000002, End: 2000003, Outcome: "canceled"}
	c.LastSeq, c.RequestCursor = 2000003, 1000001
	ch10RemainingCode(t, llm.ValidateSemantic(a.engine, c), "session_corrupt")
}
