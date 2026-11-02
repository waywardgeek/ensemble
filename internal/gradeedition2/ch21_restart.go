package gradeedition2

import (
	"fmt"
	"strings"

	"github.com/waywardgeek/ensemble/internal/fakevendor"
	"github.com/waywardgeek/ensemble/internal/grade"
)

// Two actual host processes share only an isolated data home. No implementation
// file format is assumed, and no spawn is sent to the second process.
func c21Restart(d c20Executable, check *grade.Check) {
	c20Temp([]*grade.Check{check}, func(work string) {
		first, err := d.run(work, func(body []byte) *fakevendor.Reply {
			if !fleetHas(body, "before-restart-submit") {
				return fleetCall("before-restart-submit", "submit", map[string]any{"result": "BEFORE_RESTART_RESULT"})
			}
			return &fakevendor.Reply{Text: "BEFORE_RESTART_FINAL"}
		}, func(p *mcpCommand) error {
			c := &c21Controls{p: p}
			defer c.save(work)
			_, err := c21OK(c.completed("spawn_sub_agent", map[string]any{
				"agent_id": "persistent", "task": "RESTART_CONVERSATION_NONCE: finish this bounded task.",
			}))
			return err
		})
		if err != nil {
			fail(check, "first restart session: %v", err)
			return
		}
		c20Require(check, len(first) == 2, "initial restart fixture did not complete exactly one submit turn")
		resumed, err := d.run(work, func(body []byte) *fakevendor.Reply {
			if !strings.Contains(memoryConversation(body), "RESTART_CONVERSATION_NONCE") {
				return &fakevendor.Reply{Text: "MISSING_RESTORED_CONVERSATION"}
			}
			if !fleetHas(body, "after-restart-submit") {
				return fleetCall("after-restart-submit", "submit", map[string]any{"result": "AFTER_RESTART_RESULT"})
			}
			return &fakevendor.Reply{Text: "AFTER_RESTART_FINAL"}
		}, func(p *mcpCommand) error {
			c := &c21Controls{p: p}
			defer c.save(work)
			status, err := c21OK(c.completed("get_sub_agent_status", map[string]any{"agent_id": "persistent"}))
			if err != nil {
				return fmt.Errorf("child was not automatically restored: %w", err)
			}
			if status["activity"] != "idle" || status["submitted"] != true {
				return fmt.Errorf("restored completion was not idle/submitted: %v", status)
			}
			answer, err := c.completed("ask_agent", map[string]any{
				"agent_id": "persistent", "content": "NEW_RESTART_TASK: use the retained conversation and submit a fresh answer.",
			})
			if err != nil {
				return err
			}
			if answer.bad || !strings.Contains(answer.text, "AFTER_RESTART_FINAL") {
				return fmt.Errorf("restored child did not answer fresh work: %s", answer.text)
			}
			return nil
		})
		if err != nil {
			fail(check, "second restart session: %v", err)
			return
		}
		c20Require(check, len(resumed) == 2, "restart replayed work, lost History, or failed to complete exactly one fresh turn")
	})
}
