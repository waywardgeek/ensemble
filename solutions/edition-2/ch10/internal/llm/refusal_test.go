package llm

import (
	"encoding/json"
	"example.com/ensemble/internal/common"
	"strings"
	"testing"
)

func TestRefusalRenderingRejectsUnknownShapesAndOmitsForeignMaterial(t *testing.T) {
	op, _ := testOperation()
	owner := op.Engine()
	from := common.Provenance{Vendor: "openai", Model: "fixture", Surface: "chat_completions"}
	for _, raw := range []string{`{"future":"unsupported"}`, `{"refusal":null}`, `{"refusal":42}`, `{"refusal":"known","future":"unsupported"}`} {
		context := common.Context{Entries: []common.Entry{{Actor: "agent", Parts: []common.Part{Text("visible"), {Type: "opaque", From: &from, Data: json.RawMessage(raw)}}}}}
		if _, err := Render(owner, context, common.Config{Vendor: "openai", Model: "fixture"}); err == nil {
			t.Fatal("accepted unknown opaque shape", raw)
		}
		body, err := Render(owner, context, common.Config{Vendor: "openai", Model: "foreign"})
		if err != nil || strings.Contains(string(body), "unsupported") || strings.Contains(string(body), "refusal") {
			t.Fatal("foreign material leaked", string(body), err)
		}
	}
}
