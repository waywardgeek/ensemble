package common

// The log: append-only, JSON-lines, greppable with ordinary tools.
//
// That last property is not a nicety. In Chapter 3 tool output starts arriving
// by the megabyte, and a log you cannot grep is a log you cannot debug.

import (
	"time"
)

// LogVersion is the semantic version of the log FORMAT. Replay happens with
// current code, not with historical code, so the version exists to let current
// code refuse a log it cannot faithfully interpret.
const LogVersion = 1

type Log struct {
	Version int
	Events  []Event

	Next  Seq
	Clock func() time.Time
}
