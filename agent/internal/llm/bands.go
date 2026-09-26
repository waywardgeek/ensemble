package llm

// Putting memory into the context, and taking it out again.
//
// One function does startup, a settings change, and a restore, because they
// are the same operation: make the context agree with what is on disk and
// what the settings say. Three separate paths would be three chances for them
// to disagree.

import (
	"fmt"

	"github.com/waywardgeek/ensemble/agent/internal/common"
)

// bandConfig reads the current settings, defaulting to every band on.
func (e *Engine) bandConfig() common.BandConfig {
	if e.Bands == nil {
		return common.DefaultBandConfig()
	}
	return e.Bands()
}

// SyncBands makes the context agree with the memory directory and the band
// settings.
//
// Enabled bands gain one event per file they do not already hold; disabled
// bands are emptied. It is idempotent by construction: a file already in the
// context is skipped, and a band already empty is left alone. Calling it
// twice does what calling it once did.
//
// Every event it emits carries the bytes it read, so replaying the log needs
// no disk. That is also what makes the restore behave the way Bill ruled it
// should: switch a band off and on again and the same files come back, and if
// someone edited one of those files in between, the edit comes back with it,
// because the populate event was built from the file as it is now.
func (e *Engine) SyncBands(source string) error {
	if e.Memory == nil {
		return nil
	}
	cfg := e.bandConfig()
	for _, b := range common.AllBands() {
		settings := cfg.For(b)

		if !settings.Enabled {
			if len(e.Ctx.BandEntries(b)) == 0 {
				continue
			}
			if err := e.Record(common.Event{
				Type:     common.BandDepopulated,
				BandDrop: &common.BandDepopulatedData{Band: b, Reason: "disabled"},
			}); err != nil {
				return fmt.Errorf("switching off band %s: %w", b, err)
			}
			continue
		}

		populates, err := e.Memory.PopulateEvents(b, source)
		if err != nil {
			return fmt.Errorf("reading band %s: %w", b, err)
		}
		for _, p := range populates {
			if e.Ctx.BandHas(b, p.File) {
				continue
			}
			data := p
			if err := e.Record(common.Event{
				Type:    common.BandPopulated,
				BandAdd: &data,
			}); err != nil {
				return fmt.Errorf("loading band %s file %s: %w", b, p.File, err)
			}
		}
	}
	return nil
}
