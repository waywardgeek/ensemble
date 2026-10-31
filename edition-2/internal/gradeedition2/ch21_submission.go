package gradeedition2

import (
	"encoding/json"
	"reflect"
	"strings"

	"github.com/waywardgeek/ensemble/internal/grade"
)

// Inspect the result value itself. A marker in the original tool arguments or
// inside a doubly encoded JSON string cannot establish lossless retrieval.
func managementPayload(c *grade.Check, text, id string) {
	var got map[string]any
	if err := json.Unmarshal([]byte(text), &got); err != nil || got["agent_id"] != id || !reflect.DeepEqual(got["payload"], fleetJSON(fleetReviewPayload)) {
		fail(c, "typed retrieval lost owner or exact nested payload: %s (%v)", text, err)
	}
}

// The existing submit result has a deliberate text prefix; retrieval is a JSON
// envelope. Both remain ordinary tool_finished records in live and replay views.
func managementSubmissionFrame(c *grade.Check, frames []map[string]any, call, prefix, id string) {
	for _, frame := range frames {
		if frame["type"] != "tool_finished" || frame["call_id"] != call {
			continue
		}
		text, ok := frame["result"].(string)
		if !ok || !strings.HasPrefix(text, prefix) {
			fail(c, "%s result lacks its expected text boundary", call)
			return
		}
		text = strings.TrimPrefix(text, prefix)
		if id != "" {
			managementPayload(c, text, id)
		} else {
			var value any
			if err := json.Unmarshal([]byte(text), &value); err != nil || !reflect.DeepEqual(value, fleetJSON(fleetReviewPayload)) {
				fail(c, "%s lost the direct structured submit result: %s (%v)", call, text, err)
			}
		}
		return
	}
	fail(c, "no rendered %s tool result in selected conversation", call)
}
