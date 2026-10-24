package engine

import (
	"bytes"
	"encoding/json"
	"ensemble/internal/common"
	"strings"
	"testing"
)

func TestCh19R2RawCredentialFields(t *testing.T) {
	const token = "exact\"<&\\credential"
	for _, html := range []bool{false, true} {
		var encoded bytes.Buffer
		encoder := json.NewEncoder(&encoded)
		encoder.SetEscapeHTML(html)
		_ = encoder.Encode(token)
		value := strings.TrimSpace(encoded.String())
		raw := `{ "n" : 1e+100, "echo" : ` + value + `, "ordinary" : [1, 2] }`
		want := `{ "n" : 1e+100, "echo" : "[REDACTED]", "ordinary" : [1, 2] }`
		parts := common.PartList{{Type: "tool_call", Name: "local", Args: json.RawMessage(raw)}}
		events := []common.Event{{Type: common.ResponseEnded, Response: &common.ResponseData{Parts: parts}}}
		replacer := credentialReplacer(credentialSpellings([]string{token}))
		redactCredentialEvents(events, replacer, replacer)
		if string(parts[0].Args) != want {
			t.Fatalf("ordinary JSON bytes rewritten: %s", parts[0].Args)
		}
		var deltas strings.Builder
		filter := newCredentialDeltas([]string{token}, func(d common.PartDelta) { deltas.WriteString(d.Chunk) })
		filter.push(common.PartDelta{Chunk: value[:4]})
		filter.push(common.PartDelta{Chunk: value[4:]})
		filter.flush()
		if deltas.String() != `"[REDACTED]"` {
			t.Fatalf("escaped credential delta: %s", deltas.String())
		}
	}
	var count int
	filter := newCredentialDeltas([]string{"known"}, func(d common.PartDelta) { count++ })
	filter.push(common.PartDelta{})
	filter.flush()
	if count != 1 {
		t.Fatal("empty notification lost")
	}
}
