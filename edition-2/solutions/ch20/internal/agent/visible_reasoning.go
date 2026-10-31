package agent

import (
	"ensemble/internal/common"
	"strings"
)

// Check only this provider response, not earlier conversation or tool output.
// Private reasoning and public reasoning summaries are different part kinds;
// neither substitutes for the ordinary explanation the person will retain.
// This deliberately checks presence and order, not length or semantic quality.
func visibleExplanationBeforeTools(events []common.Event) bool {
	visible := false
	for _, event := range events {
		if event.Response == nil {
			continue
		}
		for _, part := range event.Response.Parts {
			switch part.Type {
			case "text":
				visible = visible || strings.TrimSpace(part.Text) != ""
			case "tool_call":
				// One explanation may cover a coherent batch, but narration after
				// its first call is too late to authorize any of that batch.
				return visible
			}
		}
	}
	return false
}

// Pair every rejected call without dispatching it, consuming one-shot tool
// limits, or creating a job. Preserve the original assistant response, including
// opaque provider replay data, and return corrective feedback on the next wire
// request so the model can try again within the existing round budget.
func (r *runtime) rejectUnexplainedTools(toolFirstStream bool) {
	message := "Not executed: this model requires a visible explanation before tool calls. In your next response, first explain what you will do and why in ordinary user-visible text, then retry the calls. Private reasoning and reasoning summaries do not satisfy this requirement. No tool in this batch ran."
	if toolFirstStream {
		message += " A tool call arrived in the stream before any visible explanation."
	}
	for _, call := range r.active.calls {
		if err := r.append(common.Event{Type: "tool_returned", Tool: &common.ToolData{
			CallID: call.CallID, Name: call.Name, IsError: true,
			Parts: []common.Part{{Type: "text", Text: message}},
		}}); err != nil {
			r.finish("", err)
			return
		}
		r.owner.notify(common.ToolFinished{CallID: call.CallID, Result: message, IsError: true})
	}
	r.active.calls = nil
	r.send()
}
