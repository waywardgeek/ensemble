package history

import (
	"errors"
	"strings"

	"ensemble/internal/common"
)

// Correlation prevents an unrelated disk population or stale worker from
// committing a held cut. Legacy uncorrelated facts retain their old projection.
func validateCompaction(c *common.Context, e common.Event) error {
	pending := c.Compression
	if e.Compactor != nil && pending != nil && pending.Launch.TaskID != "" {
		return errors.New("previous compaction has no terminal outcome")
	}
	if e.Compactor != nil && e.Compactor.TaskID != "" {
		launch := e.Compactor
		if launch.SelectionHash == "" || launch.File.Date == "" || launch.File.Num < 1 || launch.Thru.Date == "" || launch.Thru.Num < 1 {
			return errors.New("managed compaction requires source identity")
		}
		if launch.Source == 0 {
			if launch.Band != common.BandSession || launch.From == 0 || launch.To < launch.From {
				return errors.New("invalid managed conversation boundaries")
			}
		} else if launch.Source < common.Band64x || launch.Source > common.BandSession || launch.Band != launch.Source-1 {
			return errors.New("invalid managed graduation bands")
		}
	}
	id := ""
	if e.Redact != nil {
		id = e.Redact.TaskID
	}
	if e.Depopulated != nil {
		id = e.Depopulated.TaskID
	}
	if e.Populated != nil {
		id = e.Populated.TaskID
	}
	if id != "" {
		if pending == nil || pending.Launch.TaskID != id || pending.Committed != nil {
			return errors.New("compaction event has no matching uncommitted launch")
		}
		launch := pending.Launch
		if cut := e.Redact; cut != nil {
			if launch.Source != 0 || cut.From != launch.From || cut.To != launch.To || cut.Level != "redact_summary" || len(cut.Replacement) != 0 || cut.Reason != "memory compression" {
				return errors.New("cut differs from managed selection")
			}
		}
		if retire := e.Depopulated; retire != nil {
			if launch.Source == 0 || retire.Band != launch.Source || retire.Thru == nil || *retire.Thru != launch.Thru || retire.Reason != "graduation" {
				return errors.New("retirement differs from managed selection")
			}
		}
		if pop := e.Populated; pop != nil {
			if pop.Band != launch.Band || pop.File != launch.File || pop.Thru != launch.Thru || strings.TrimSpace(pop.Text) == "" || pop.Source != "compression" || (pending.Cut == nil && pending.Retire == nil) {
				return errors.New("replacement differs from managed selection")
			}
		}
	}
	// An old uncorrelated cut must not borrow a managed launch's eventual
	// replacement. Non-compaction edits and disk populations remain independent.
	if pending != nil && pending.Launch.TaskID != "" && id == "" {
		if e.Redact != nil && e.Redact.Reason == "memory compression" || e.Depopulated != nil && e.Depopulated.Reason == "graduation" {
			return errors.New("managed retirement requires task identity")
		}
	}
	if outcome := e.CompactionOutcome; outcome != nil {
		if pending == nil {
			if c.CompactionOutcome == nil || c.CompactionOutcome.TaskID != outcome.TaskID || c.CompactionOutcome.State != outcome.State {
				return errors.New("outcome has no matching compaction")
			}
			return nil // Repeated completion is harmless and retains bounded state.
		}
		if pending.Launch.TaskID != outcome.TaskID || (pending.Committed != nil) != (outcome.State == "committed") {
			return errors.New("outcome disagrees with compaction commit state")
		}
	}
	return nil
}
