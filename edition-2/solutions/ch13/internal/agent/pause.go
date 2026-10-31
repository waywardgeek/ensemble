package agent

// Pause holds the Agent gate, not a browser-owned share. Every explicit unpause releases
// it; disconnecting or another tab's local typing state cannot retain a veto.
func (a *agent) Pause() {
	a.pauseMu.Lock()
	defer a.pauseMu.Unlock()
	if a.resume == nil {
		a.resume = make(chan struct{})
	}
}

// Unpause wakes the current boundary waiter without cancelling running jobs.
func (a *agent) Unpause() {
	a.pauseMu.Lock()
	defer a.pauseMu.Unlock()
	if a.resume != nil {
		close(a.resume)
		a.resume = nil
	}
}

// Paused exposes the effective Agent-wide state to every client.
func (a *agent) Paused() bool { return a.pauseChannel() != nil }
func (a *agent) pauseChannel() <-chan struct{} {
	a.pauseMu.Lock()
	defer a.pauseMu.Unlock()
	return a.resume
}
