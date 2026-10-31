package agent

import (
	"ensemble/internal/common"
	"time"
)

// Activity returns an immutable copy for callers outside the actor.
func (a *agent) Activity() common.Activity {
	a.activityMu.Lock()
	defer a.activityMu.Unlock()
	if a.activity.Phase == "" {
		return common.Activity{Phase: common.ActivityIdle}
	}
	return a.activity
}

// Only the actor changes activity. A backgrounded tool cannot overwrite the
// newer thinking/idle phase when its independent execution finally finishes.
func (a *agent) setActivity(phase, tool, detail string) {
	runes := []rune(detail)
	if len(runes) > 50 {
		detail = string(runes[:50])
	}
	activity := common.Activity{Phase: phase, Tool: tool, Detail: detail}
	a.activityMu.Lock()
	a.activity = activity
	a.activityMu.Unlock()
	a.notify(common.ActivityChanged{Activity: activity})
}

// ProgressTime includes streamed progress, so a long live response is not stuck.
func (a *agent) ProgressTime() time.Time {
	a.activityMu.Lock()
	defer a.activityMu.Unlock()
	return a.progressTime
}
