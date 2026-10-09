package cachelens

import "testing"

// A minimal body in the shape the OpenAI Responses API actually sends: the
// conversation lives under "input", not "messages".
func responsesBody(convo string) []byte {
	return []byte(`{"model":"gpt-6-astra","instructions":"be helpful",` +
		`"tools":[{"name":"read"}],"input":` + convo + `}`)
}

func TestResponsesAPIConversationIsRecognisedAsTheDialogue(t *testing.T) {
	s, err := Split("gpt-6-astra", responsesBody(`[{"role":"user","content":"hi"}]`))
	if err != nil {
		t.Fatalf("Split: %v", err)
	}
	if s.Dialogue != "input" {
		t.Errorf("Dialogue = %q, want \"input\". Anything else files the "+
			"conversation under the frozen sections, and a conversation is "+
			"not frozen: it changes every turn by design.", s.Dialogue)
	}
}

func TestAppendingOnTheResponsesAPIIsNotReportedAsInstability(t *testing.T) {
	// This is the regression that matters, and the one whose absence let a
	// real session burn millions of uncached tokens unnoticed.
	//
	// With "input" unrecognised, the conversation counted as a frozen
	// section. Frozen sections are not allowed to change, so every ordinary
	// appended turn was reported as PREFIX DIVERGED. The alarm fired on all
	// 100% of requests, which is the same as not having one: when a genuine
	// break happened it was indistinguishable from the constant noise.
	prior, err := Split("gpt-6-astra", responsesBody(
		`[{"role":"user","content":"hi"}]`))
	if err != nil {
		t.Fatalf("Split prior: %v", err)
	}
	current, err := Split("gpt-6-astra", responsesBody(
		`[{"role":"user","content":"hi"},{"role":"assistant","content":"yo"}]`))
	if err != nil {
		t.Fatalf("Split current: %v", err)
	}

	d := Compare(prior, current)

	if d.Unstable != "" {
		t.Errorf("Unstable = %q for a plain append; the lens is crying wolf "+
			"and a real divergence will be lost in the noise", d.Unstable)
	}
	if d.DialogueRewritten {
		t.Error("DialogueRewritten = true for a plain append on the " +
			"Responses API")
	}
	if !d.DialogueChanged {
		t.Error("DialogueChanged = false; the conversation did change, so " +
			"the lens is not looking at the conversation at all")
	}
}

func TestRetroactiveEditOnTheResponsesAPIIsCaught(t *testing.T) {
	// The other half: once "input" is understood as the dialogue, a rewrite
	// inside it must still be caught. A detector that never fires is as
	// useless as one that always does.
	prior, err := Split("gpt-6-astra", responsesBody(
		`[{"role":"user","content":"hi"},`+
			`{"type":"function_call","call_id":"call_1","name":"read"}]`))
	if err != nil {
		t.Fatalf("Split prior: %v", err)
	}
	current, err := Split("gpt-6-astra", responsesBody(
		`[{"role":"user","content":"hi"},`+
			`{"type":"reasoning","id":"rs_1","encrypted_content":"gAAAAAB"},`+
			`{"type":"function_call","call_id":"call_1","name":"read"}]`))
	if err != nil {
		t.Fatalf("Split current: %v", err)
	}

	d := Compare(prior, current)

	if !d.DialogueRewritten {
		t.Error("DialogueRewritten = false; a reasoning item inserted into " +
			"an already-sent turn rewrote history and must be reported")
	}
}
