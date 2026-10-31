package gradeedition2

import (
	"errors"
	"fmt"
	"strings"

	"github.com/waywardgeek/ensemble/internal/fakevendor"
	"github.com/waywardgeek/ensemble/internal/grade"
)

// The shipped agent_status tool is the consumer: raw outcome usage alone cannot
// prove that the owner imported it once or preserved its latest main response.
func memoryAccounting(run memoryRunner, dir string, failed bool) (accounting, replay error) {
	const model = "claude-ch16-course"
	const memory = "ACCOUNTED_MEMORY_15: I retained the decision and its reason."
	aux := fakevendor.Canonical{Input: 8000, CacheWrite: 700, CacheRead: 900, Output: 9000}
	replies := append(memoryTalk(1),
		fakevendor.Reply{ToolName: "agent_status", ToolID: "memory-account-first", ToolArgs: `{}`}, fakevendor.Reply{Text: "First status recorded."},
		fakevendor.Reply{ToolName: "agent_status", ToolID: "memory-account-repeat", ToolArgs: `{}`}, fakevendor.Reply{Text: "Second status recorded."})
	for i := range replies {
		replies[i].Usage = fakevendor.Canonical{Input: 100 + i*10, CacheWrite: 10 + i, CacheRead: 20 + i, Output: 30 + i}
	}
	workers := 0
	route := func(body []byte) *fakevendor.Reply {
		if !memoryCompressor(body) {
			return nil
		}
		workers++
		text := memory
		if failed {
			text = " "
		}
		return &fakevendor.Reply{ToolName: "submit", ToolID: "accounted-memory", ToolArgs: memoryJSON(text), Usage: aux}
	}
	prompts := append(memoryPrompts(1), "Report status after memory finishes, without a checkpoint.", "Report status again, without a checkpoint.")
	out := run(memoryOpts{dir: dir, model: model, settings: memorySettings(5000), prompts: prompts, expect: []int{2, 4, 6}, replies: replies, route: route, accounting: true})
	if out.fatal != "" {
		err := fmt.Errorf("accounting session: %s", out.fatal)
		return err, err
	}
	if workers != 1 {
		err := fmt.Errorf("accounting session made %d worker calls, want one", workers)
		return err, err
	}
	check := grade.Check{Passed: true}
	total := aux
	for i, response := range replies {
		u := response.Usage
		total.Input += u.Input
		total.CacheRead += u.CacheRead
		total.CacheWrite += u.CacheWrite
		total.Output += u.Output
		id := ""
		if i == 2 {
			id = "memory-account-first"
		}
		if i == 4 {
			id = "memory-account-repeat"
		}
		if i%2 == 1 {
			// The first meter is sampled after child completion but before the
			// status turn. A later main response must not hide an overwritten Last.
			if i/2 >= len(out.meters) || !invoiceCounts(out.meters[i/2], "", u) || !invoiceCounts(out.meters[i/2], "total", total) {
				fail(&check, "post-drain meter %d lost main Last or once-only totals", i/2+1)
			}
		}
		if id != "" {
			// The request that calls status has already incurred its own usage.
			// A later main response must add only its measured counts, not the child again.
			statusReport(&check, fleetResult(&check, out.reqs, id, false), model, u, total)
		}
	}
	if !check.Passed {
		accounting = fmt.Errorf("%s", strings.Join(check.Details, "; "))
	}
	last := memoryConversation(memoryLast(out))
	if strings.Contains(last, memory) == failed || strings.Contains(last, "END-WORK-0.") != failed {
		accounting = errors.Join(accounting, fmt.Errorf("accounted attempt did not have the intended success/failure source state"))
	}
	// Restoration retains historical outcome receipts but starts fresh measured
	// session totals. A real status call, not a zero-request startup alone, proves it.
	u := fakevendor.Canonical{Input: 7, CacheWrite: 3, CacheRead: 5, Output: 2}
	restarted := run(memoryOpts{dir: dir, model: model, settings: memorySettings(5000), prompts: []string{"Report restored status without another checkpoint."}, expect: []int{2}, replies: []fakevendor.Reply{
		{ToolName: "agent_status", ToolID: "memory-account-restart", ToolArgs: `{}`, Usage: u}, {Text: "Restored status recorded."},
	}, route: route})
	if restarted.fatal != "" {
		return accounting, fmt.Errorf("accounting restart: %s", restarted.fatal)
	}
	if workers != 1 || len(restarted.reqs) != 2 {
		return accounting, fmt.Errorf("startup/replay respawned measured memory work")
	}
	check = grade.Check{Passed: true}
	statusReport(&check, fleetResult(&check, restarted.reqs, "memory-account-restart", false), model, u, u)
	if !check.Passed {
		replay = fmt.Errorf("%s", strings.Join(check.Details, "; "))
	}
	return accounting, replay
}
