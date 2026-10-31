package agent

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	"ensemble/internal/common"
)

const goalReminder = "Automatic goal reminder: check the current goal and the actual state of jobs, children, and saved results. Continue unfinished authorized work. Do not replay commands merely because a response ended. If the goal is complete or you need the human, turn goal reminders off and explain why."

type goalRuntime struct {
	config     common.GoalConfig
	revision   uint64
	generation uint64
	timer      *time.Timer
	due        bool
	next       time.Time
	err        string
}
type goalChange struct {
	config common.GoalConfig
	reply  chan goalReply
}
type goalReply struct {
	status common.GoalStatus
	err    error
}
type goalWake struct{ generation uint64 }
type goalRefresh struct{}
type goalRead struct {
	revision uint64
	err      string
}

// GoalStatus reads a published snapshot without touching the timer.
func (a *agent) GoalStatus() common.GoalStatus {
	a.goalMu.RLock()
	defer a.goalMu.RUnlock()
	s := a.goalStatus
	if s.IdleSeconds == 0 {
		s.IdleSeconds = 480
	}
	if s.Phase == "" {
		s.Phase = "none"
	}
	return s
}

// ConfigureGoal routes updates through the same owner as turns and shutdown.
func (a *agent) ConfigureGoal(config common.GoalConfig) (common.GoalStatus, error) {
	reply := make(chan goalReply, 1)
	select {
	case <-a.done:
		return a.GoalStatus(), common.ErrStopped
	default:
	}
	if err := a.post(goalChange{config: config, reply: reply}); err != nil {
		return a.GoalStatus(), err
	}
	select {
	case r := <-reply:
		return r.status, r.err
	case <-a.done:
		return a.GoalStatus(), common.ErrStopped
	}
}

func (a *agent) goalAvailable() bool {
	if a.skills == nil {
		return false
	}
	for _, name := range a.skills.PendingUnload() {
		if name == "goals" {
			return false
		}
	}
	_, ok := a.skills.ToolMeta("goal_control")
	return ok
}
func (a *agent) readGoal(file string) (string, error) {
	data, err := a.Sandbox().ReadFile(file)
	if err != nil {
		return "", err
	}
	if strings.TrimSpace(string(data)) == "" {
		return "", errors.New("goal file is empty")
	}
	return string(data), nil
}

// GoalSystemPrompt reads once per rendered request, never from compacted history.
// A failed automatic read refuses the request. Human turns retain a repair notice.
func (a *agent) GoalSystemPrompt() (string, error) {
	if !a.goalAvailable() {
		return "", nil
	}
	a.goalMu.RLock()
	file, rev, automatic := a.goalStatus.File, a.goalRevision, a.goalAutomatic
	a.goalMu.RUnlock()
	if file == "" {
		return "", nil
	}
	text, err := a.readGoal(file)
	result := goalRead{revision: rev}
	if err != nil {
		result.err = err.Error()
	}
	a.box.Post(result)
	if err != nil {
		if automatic {
			return "", fmt.Errorf("goal file unavailable: %w", err)
		}
		return fmt.Sprintf("Current goal unavailable (%s): %v. Do not infer completion or pretend a cached goal is current. Help repair the file or clear it with goal_control; automatic reminders will not proceed until explicitly re-enabled.", file, err), nil
	}
	return "Current goal (" + file + "):\n" + text, nil
}

func (r *runtime) stopGoalTimer() {
	r.goal.generation++
	if r.goal.timer != nil {
		r.goal.timer.Stop()
		r.goal.timer = nil
	}
	r.goal.due = false
	r.goal.next = time.Time{}
}
func (r *runtime) changeGoal(m goalChange) {
	cfg := m.config
	if cfg.IdleSeconds == 0 {
		cfg.IdleSeconds = 480
	}
	var err error
	switch {
	case r.stopping:
		err = common.ErrStopped
	case !r.owner.goalAvailable():
		err = errors.New("load the goals skill before configuring a goal")
	case r.owner.spec.Compaction:
		err = errors.New("compaction workers cannot configure goals")
	case math.IsNaN(cfg.IdleSeconds) || math.IsInf(cfg.IdleSeconds, 0) || cfg.IdleSeconds <= 0 || cfg.IdleSeconds >= float64(math.MaxInt64)/float64(time.Second) || time.Duration(cfg.IdleSeconds*float64(time.Second)) <= 0:
		err = errors.New("idle_seconds must be a positive representable duration")
	}
	if cfg.File == "" {
		cfg.Enabled = false
	}
	if err == nil && cfg.File != "" && (cfg.Enabled || cfg.File != r.goal.config.File) {
		_, err = r.owner.readGoal(cfg.File)
	}
	if err != nil {
		m.reply <- goalReply{r.owner.GoalStatus(), err}
		return
	}
	r.stopGoalTimer()
	r.goal.revision++
	r.goal.config = cfg
	r.goal.err = ""
	r.updateGoal()
	m.reply <- goalReply{status: r.owner.GoalStatus()}
}

func (r *runtime) publishGoal(phase string) {
	s := common.GoalStatus{GoalConfig: r.goal.config, Phase: phase, NextReminder: r.goal.next, Error: r.goal.err}
	if s.IdleSeconds == 0 {
		s.IdleSeconds = 480
	}
	r.owner.goalMu.Lock()
	changed := r.owner.goalStatus != s
	r.owner.goalStatus = s
	r.owner.goalRevision = r.goal.revision
	r.owner.goalMu.Unlock()
	// Rejected construction has not been published. Its parent may still hold
	// the child-reservation lock while joining this discard.
	if changed && !r.discard {
		r.owner.notify(common.GoalChanged{Status: s})
	}
}

// updateGoal runs only on the actor, after ordinary queued turns have precedence.
func (r *runtime) updateGoal() {
	if r.goal.config.File != "" && !r.owner.goalAvailable() {
		r.stopGoalTimer()
		r.goal.config = common.GoalConfig{}
		r.goal.revision++
		r.goal.err = ""
	}
	terminal := r.owner.fleet != nil && r.owner.fleet.Status().State == "submitted"
	if r.stopping || terminal || r.owner.spec.Compaction {
		r.stopGoalTimer()
		r.goal.config.Enabled = false
	}
	phase := "none"
	switch {
	case r.goal.config.File == "":
		r.stopGoalTimer()
	case r.goal.err != "":
		r.stopGoalTimer()
		phase = "error"
	case r.owner.Paused():
		r.stopGoalTimer()
		phase = "paused"
	case r.active != nil || len(r.pending) > 0:
		r.stopGoalTimer()
		phase = "working"
	case !r.goal.config.Enabled:
		r.stopGoalTimer()
		phase = "off"
	default:
		phase = "waiting"
		if r.goal.due {
			r.stopGoalTimer()
			if _, err := r.owner.readGoal(r.goal.config.File); err != nil {
				r.goal.err = err.Error()
				r.goal.config.Enabled = false
				phase = "error"
			} else {
				// Buffered completion needs no observer goroutine and uses ordinary lifetime.
				r.begin(request{ctx: context.Background(), text: goalReminder, reply: make(chan common.Result, 1), automatic: true})
				// Startup can refuse synchronously. With due cleared above, this
				// publishes the real active state or arms a fresh full interval.
				r.updateGoal()
				return
			}
		} else if r.goal.timer == nil {
			delay := time.Duration(r.goal.config.IdleSeconds * float64(time.Second))
			r.goal.generation++
			generation := r.goal.generation
			r.goal.next = time.Now().Add(delay)
			r.goal.timer = time.AfterFunc(delay, func() { r.owner.box.Post(goalWake{generation: generation}) })
		}
	}
	r.publishGoal(phase)
}

// GoalSkillUnloaded is a nonblocking registry notification, never a second owner.
func (a *agent) GoalSkillUnloaded() { a.box.Post(goalRefresh{}) }
