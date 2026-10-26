package ensemble_test

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"ensemble/ensemble"
	"ensemble/internal/common"
)

func TestHistoryNarrowReads(t *testing.T) {
	for _, log := range []string{"", conversation, conversation + firstResult + lastResult} {
		agent := loaded(t, ensemble.Gemini, log)
		history := agent.History()
		full := history.Context()
		state := history.State()
		if state.Turn != full.Turn || state.SentThrough != full.SentThrough || state.LastUsage != full.LastUsage ||
			state.DisabledBands != full.DisabledBands || state.HasEphemera != (len(full.Ephemera) > 0) || state.HasCompression {
			t.Fatalf("metadata differs from full snapshot: %+v", state)
		}
		state.DisabledBands[common.BandMemory] = !state.DisabledBands[common.BandMemory]
		if history.State().DisabledBands != full.DisabledBands {
			t.Fatal("returned metadata aliases history")
		}
		for _, entry := range full.Dialogue {
			got, err := json.Marshal(history.DialogueEntry(entry.Seq))
			if err != nil {
				t.Fatal(err)
			}
			want, err := json.Marshal(entry)
			if err != nil || !bytes.Equal(got, want) {
				t.Fatalf("entry %d differs from full snapshot", entry.Seq)
			}
		}
		if history.DialogueEntry(9999) != nil {
			t.Fatal("missing dialogue entry was fabricated")
		}
	}
}

func TestHistoryNarrowEntryOwnership(t *testing.T) {
	log := strings.Replace(conversation, `{"type":"text","text":"decision"}`, `{"type":"text","text":"decision"},{"type":"opaque","from":{"vendor":"gemini","model":"model","surface":"generate_content"},"data":{"signature":"preserved"}}`, 1)
	agent := loaded(t, ensemble.Gemini, log+firstResult+lastResult)
	history := agent.History()
	before, err := json.Marshal(history.Context())
	if err != nil {
		t.Fatal(err)
	}
	response := history.DialogueEntry(3)
	response.Parts[0].Args[0] = '!'
	response.Parts[0].Opaque[0] = '!'
	response.Parts[len(response.Parts)-1].Data[0] = '!'
	response.Parts[0].Name = "changed"
	result := history.DialogueEntry(4)
	result.Parts[0].Parts[0].Text = "changed"
	result.Parts[0].Parts[1].Ref.Locator = "changed"
	after, err := json.Marshal(history.Context())
	if err != nil || !bytes.Equal(before, after) {
		t.Fatal("entry mutation changed owned history")
	}
	if history.DialogueEntry(3).Parts[0].Name != "read" || history.DialogueEntry(4).Parts[0].Parts[0].Text != "secret" {
		t.Fatal("entry reads share mutable parts")
	}
}

func TestHistoryNarrowMemoryState(t *testing.T) {
	agent, _, launch := receiptAgent(t)
	history := agent.History()
	appendCompactionFact(t, agent, common.Event{Type: "band_depopulated", Depopulated: &common.BandDepopulatedData{Band: common.BandMemory, Reason: "disabled"}})
	if !history.State().DisabledBands[common.BandMemory] {
		t.Fatal("recorded memory toggle missing from metadata")
	}
	appendCompactionFact(t, agent, common.Event{Type: "compactor_launched", Compactor: &launch})
	if state := history.State(); !state.HasCompression || state.CompressionCommitted {
		t.Fatalf("pending receipt metadata: %+v", state)
	}
	appendCompactionFact(t, agent, receiptCut(launch))
	appendCompactionFact(t, agent, receiptPopulation(launch))
	if state := history.State(); !state.HasCompression || !state.CompressionCommitted {
		t.Fatalf("committed receipt metadata: %+v", state)
	}
}

func TestHistoryNarrowEphemeraState(t *testing.T) {
	agent := loaded(t, ensemble.Anthropic, "")
	history := agent.History()
	if err := history.Append(common.Event{Type: "message_received", Message: &common.MessageData{Actor: common.System, Parts: []common.Part{{Type: "text", Text: "one request only"}}}}); err != nil {
		t.Fatal(err)
	}
	if !history.State().HasEphemera {
		t.Fatal("pending instructions missing from metadata")
	}
	if err := history.Append(common.Event{Type: "request_sent", Request: &common.RequestData{To: common.Provenance{Vendor: common.Anthropic, Model: "model", Surface: common.Messages}}}); err != nil {
		t.Fatal(err)
	}
	if state := history.State(); state.HasEphemera || state.SentThrough != 1 {
		t.Fatalf("request did not consume pending instructions: %+v", state)
	}
}
