package persistence

import "example.com/ensemble/internal/common"

// Added only by the independent review overlay, before any save is in flight.
// The wrapper delegates to the real owned I/O and records the injected boundary.
func (s *Store) Ch10ReviewWrapIO(wrap func(common.CheckpointIO) common.CheckpointIO) {
	s.io = wrap(s.io)
}
