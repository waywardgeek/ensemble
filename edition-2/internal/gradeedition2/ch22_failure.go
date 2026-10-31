package gradeedition2

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/waywardgeek/ensemble/internal/fakevendor"
)

// Protocol failures are terminal outcomes of real children. Retaining the
// source is insufficient if an automatic retry silently starts another attempt.
func memoryProtocolFailure(run memoryRunner, dir, mode string) error {
	const candidate = "TRUNCATED_CANDIDATE_15 must never be committed."
	o := memoryOptions(dir, 1, memorySettings(5000), &memoryRouter{})
	o.prompts[len(o.prompts)-1] = "Verify retained source after refused candidate, without a checkpoint."
	o.lengthStop = mode == "length"
	requests := 0
	o.route = func(body []byte) *fakevendor.Reply {
		if !memoryCompressor(body) {
			return nil
		}
		requests++
		args := memoryJSON(candidate)
		switch mode {
		case "empty":
			args = memoryJSON(" \n\t")
		case "type":
			args = `{"memory":7}`
		}
		return &fakevendor.Reply{ToolName: "submit", ToolID: fmt.Sprintf("invalid-memory-%d", requests), ToolArgs: args, Usage: fakevendor.Canonical{Input: 80, CacheRead: 20, Output: 30}}
	}
	out := run(o)
	if out.fatal != "" {
		return fmt.Errorf("protocol session: %s", out.fatal)
	}
	// Three bounded Engine sends can have separate bounded authentication retry
	// behavior. This faithful local peer supplies no authentication challenge.
	if requests < 1 || requests > 3 {
		return fmt.Errorf("failed worker made %d model requests, want1–3", requests)
	}
	if mode != "type" && requests != 1 {
		return fmt.Errorf("terminal %s failure retried %d times", mode, requests)
	}
	if mode == "type" {
		// The child must receive the validation error before any owner-side JSON
		// parsing can reject it. A failed owner alone does not prove tool admission.
		rejected := false
		for _, request := range out.reqs {
			if _, isError, err := anthropicToolText(request.Body, "invalid-memory-1"); err == nil && isError {
				rejected = true
			}
		}
		if !rejected {
			return fmt.Errorf("wrong-type submit never produced its actual correction error")
		}
	}
	records, err := memoryRecords(dir)
	if err != nil {
		return err
	}
	launches, outcomes := map[string]bool{}, map[string]map[string]any{}
	for _, raw := range records {
		var event map[string]any
		if err = json.Unmarshal(raw, &event); err != nil {
			return err
		}
		if event["type"] == "compactor_launched" {
			id, _ := object(event["compactor"])["task_id"].(string)
			launches[id] = true
		}
		if event["type"] == "compaction_outcome" {
			value := object(event["compaction_outcome"])
			id, _ := value["task_id"].(string)
			outcomes[id] = value
		}
	}
	if len(launches) != 1 || launches[""] {
		return fmt.Errorf("failed candidate did not retain one correlated attempt: %v", launches)
	}
	for id := range launches {
		value := outcomes[id]
		reason, _ := value["reason"].(string)
		if value["state"] != "failed" || strings.TrimSpace(reason) == "" {
			return fmt.Errorf("candidate lacks reasoned failure: %v", value)
		}
	}
	last := memoryConversation(memoryLast(out))
	if !strings.Contains(last, o.prompts[len(o.prompts)-1]) || !strings.Contains(last, "Work item 0.") || !strings.Contains(last, "END-WORK-0.") || strings.Contains(last, candidate) {
		return fmt.Errorf("failed candidate removed selected source or entered the next request")
	}
	return nil
}
