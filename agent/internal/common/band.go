package common

// Memory bands: the five tiers of what the agent remembers.
//
// Chapter 15 taught the context to throw bytes away. This is how it throws
// bytes away WITHOUT losing the information: a band holds compressed prose
// that an LLM wrote, and the log records that writing as a decision, never as
// a recipe to re-run.
//
// The rule this file exists to enforce:
//
//	A populate event CARRIES THE BYTES. It never carries a reference to
//	them.
//
// That is not a preference. The reducer is a pure function over (context,
// event) with no filesystem, and no renderer dereferences anything either. So
// a populate event holding a path would describe memory that nothing in the
// pipeline could ever turn into text. Worse, it would replay tomorrow against
// whatever the file says tomorrow, and a context that changes under replay is
// not a context, it is a rumour.
//
// The cost is real and was accepted deliberately: the log carries the memory
// bytes. It buys back determinism (replay re-applies a decision; it never
// remakes it) and it costs nothing in cache, because a populate appends at the
// tail.

import (
	"encoding/json"
	"fmt"
)

// Band names one memory tier. The order of the constants is the order the
// bands appear in the context, coarsest and most stable first, which is also
// the order that costs the least cache: a band that rarely changes sits ahead
// of one that changes often, so churn never invalidates what is stable.
type Band uint8

const (
	BandSoul    Band = iota + 1 // SOUL.md — identity. Never cascades.
	BandMemory                  // MEMORY.md — curated, terminal.
	Band64x                     // 64x compression tier (bucket-1)
	Band8x                      // 8x compression tier (bucket-0)
	BandSession                 // session memories, written at micro_handoff
)

var bandNames = map[Band]string{
	BandSoul:    "soul",
	BandMemory:  "memory",
	Band64x:     "64x",
	Band8x:      "8x",
	BandSession: "session",
}

func (b Band) String() string {
	if s, ok := bandNames[b]; ok {
		return s
	}
	return fmt.Sprintf("band(%d)", uint8(b))
}

func (b Band) MarshalJSON() ([]byte, error) {
	s, ok := bandNames[b]
	if !ok {
		return nil, fmt.Errorf("refusing to marshal invalid band %d", uint8(b))
	}
	return json.Marshal(s)
}

func (b *Band) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	for k, v := range bandNames {
		if NormalizeName(v) == NormalizeName(s) {
			*b = k
			return nil
		}
	}
	return fmt.Errorf("unknown band %q: refusing to load this log", s)
}

// Kind is the EntryKind a band's entries land as. Each band gets its own
// kind rather than one parameterized kind, because Chapter 15's removal rule
// wants every non-dialogue kind literally checkable in the tool-clearing
// path, the same way KindSkill and KindTools already are.
func (b Band) Kind() EntryKind {
	switch b {
	case BandSoul:
		return KindSoul
	case BandMemory:
		return KindMemory
	case Band64x:
		return Kind64x
	case Band8x:
		return Kind8x
	case BandSession:
		return KindSession
	}
	return 0
}

// BandForKind is Kind's inverse. Zero for any kind that is not a band.
func BandForKind(k EntryKind) Band {
	switch k {
	case KindSoul:
		return BandSoul
	case KindMemory:
		return BandMemory
	case Kind64x:
		return Band64x
	case Kind8x:
		return Band8x
	case KindSession:
		return BandSession
	}
	return 0
}

// MemoryFileID identifies one memory file by the convention the memory
// directory already uses on disk: a date and a same-day sequence number.
//
// Deliberately NOT the event log's Seq. Seq orders events inside one log;
// this names a position in a memory corpus that outlives any one log, and
// which a different session or a different agent can read with no log in
// hand at all. Keying a compressed file's range to a number meaningful only
// inside one specific log was the wrong coupling.
type MemoryFileID struct {
	Date string `json:"date"` // "2026-09-26"
	Num  int    `json:"num"`  // 1, 2, 3... within that date
}

func (m MemoryFileID) String() string { return fmt.Sprintf("%s-%d", m.Date, m.Num) }

// Zero reports whether this names no file at all.
func (m MemoryFileID) Zero() bool { return m.Date == "" && m.Num == 0 }

// Before orders two memory files. Dates are ISO-8601, so they compare
// correctly as strings; the same-day number breaks ties. This is the ordering
// that "oldest first" means everywhere in the cascade.
func (m MemoryFileID) Before(o MemoryFileID) bool {
	if m.Date != o.Date {
		return m.Date < o.Date
	}
	return m.Num < o.Num
}

// AtOrBefore is Before plus equality: the inclusive sense "through", which is
// what a graduation's Thru means.
func (m MemoryFileID) AtOrBefore(o MemoryFileID) bool { return !o.Before(m) }

// BandPopulatedData adds one unit of memory to a band, bytes included.
//
// Text is the memory itself, verbatim. The event builder reads the file and
// puts its contents here; nothing downstream reads anything. See this file's
// header for why that is not negotiable.
//
// File identifies the unit for ordering and for the cascade's "oldest first".
// It is also what makes restoring a band idempotent: the same files produce
// the same entries in the same places, whatever order the events arrive in.
type BandPopulatedData struct {
	Band Band         `json:"band"`
	File MemoryFileID `json:"file"`
	Text string       `json:"text"`
	// Thru is set by the cascade only: the newest source file folded into
	// this one. It records what this unit supersedes, so an auditor can see
	// the graduation without replaying it.
	Thru MemoryFileID `json:"thru,omitempty"`
	// Source is provenance for a reader, never a switch for the reducer:
	// "startup", "compaction", "graduation", "curator", "restore".
	Source string `json:"source,omitempty"`
}

// BandDepopulatedData retires memory from a band.
//
// It carries no content, and that asymmetry with BandPopulatedData is the
// point: removing needs only to say WHAT GOES, while adding must say what
// arrives, because nothing downstream can go and look.
//
// Thru nil means the whole band. The event does not enumerate what it wiped
// because "all of it" is computable from whatever the context holds at the
// moment it is applied — and an enumeration written now would be a second
// copy of a fact the context already has, free to drift.
type BandDepopulatedData struct {
	Band   Band          `json:"band"`
	Thru   *MemoryFileID `json:"thru,omitempty"`
	Reason string        `json:"reason,omitempty"` // "graduation", "disabled"
}

// CompactorLaunchData records that a compressor started work on a band.
//
// It exists so a half-finished compaction is VISIBLE rather than inferred. A
// compressor is an LLM call: slow, costly, and killable half way through by a
// crash. A launch with no matching BandPopulated is therefore an abandoned
// one, and the rule is to leave the source band exactly as it was and let the
// next checkpoint measure again.
//
// Resuming would be the tempting alternative and it is wrong: the process
// that died is the only thing that knew how far it got, and re-running from a
// guess produces a second compressed file covering an overlapping range.
type CompactorLaunchData struct {
	Band Band         `json:"band"`
	Thru MemoryFileID `json:"thru"`
}

// BandSettings controls one band: whether it is in the context at all, how
// many bytes it should occupy, and when the cascade should move its oldest
// contents up a tier.
//
// Watermarks rather than a single threshold, because compaction that fires at
// exactly its budget fires on almost every checkpoint. The band is allowed to
// grow to High, and when it does, enough of its oldest content graduates to
// bring it back down to Low. The gap is what stops an expensive LLM call from
// running every time a memory is written.
type BandSettings struct {
	Enabled       bool `json:"enabled"`
	Budget        int  `json:"budget,omitempty"`
	HighWatermark int  `json:"high_watermark,omitempty"` // 0 means 2x budget
	LowWatermark  int  `json:"low_watermark,omitempty"`  // 0 means 1x budget
}

// High is the size at which this band graduates its oldest contents.
func (s BandSettings) High() int {
	if s.HighWatermark > 0 {
		return s.HighWatermark
	}
	return 2 * s.Budget
}

// Low is the size a graduation brings this band back down to.
func (s BandSettings) Low() int {
	if s.LowWatermark > 0 {
		return s.LowWatermark
	}
	return s.Budget
}

// BandConfig is one settings row per band.
//
// BandSoul's watermarks are present and unused: soul never graduates
// anywhere. Kept as one type rather than splitting leaf bands from cascading
// ones, because a second type would exist to describe a single member and
// every consumer would then have to handle both. The asymmetry is documented
// instead of designed around.
type BandConfig struct {
	Soul    BandSettings `json:"soul"`
	Memory  BandSettings `json:"memory"`
	B64x    BandSettings `json:"64x"`
	B8x     BandSettings `json:"8x"`
	Session BandSettings `json:"session"`

	// Conversation is not a band. It is the live dialogue, and it gets the
	// same two watermarks for the same reason: something has to say when
	// there is too much of it and how much to leave behind. Giving it the
	// same shape as a band means the compaction loop reads the same way at
	// every rung, rather than special-casing the one rung that feeds all
	// the others.
	Conversation BandSettings `json:"conversation"`
}

// For returns the settings row for a band.
func (c BandConfig) For(b Band) BandSettings {
	switch b {
	case BandSoul:
		return c.Soul
	case BandMemory:
		return c.Memory
	case Band64x:
		return c.B64x
	case Band8x:
		return c.B8x
	case BandSession:
		return c.Session
	}
	return BandSettings{}
}

// Up is the band one tier coarser, where this band's oldest contents go when
// it crosses its high watermark. Zero for bands that graduate nowhere.
func (b Band) Up() Band {
	switch b {
	case BandSession:
		return Band8x
	case Band8x:
		return Band64x
	case Band64x:
		return BandMemory
	}
	return 0
}

// DefaultBandConfig is every band on, sized so the whole memory system fits
// comfortably inside a normal context budget.
//
// The numbers are a starting point and nothing more. They were chosen to be
// obviously adjustable rather than tuned: nothing here has been measured
// against recall quality, and pretending otherwise would be worse than saying
// so.
func DefaultBandConfig() BandConfig {
	return BandConfig{
		Soul:    BandSettings{Enabled: true, Budget: 4 * 1024},
		Memory:  BandSettings{Enabled: true, Budget: 12 * 1024},
		B64x:    BandSettings{Enabled: true, Budget: 12 * 1024},
		B8x:     BandSettings{Enabled: true, Budget: 12 * 1024},
		Session: BandSettings{Enabled: true, Budget: 12 * 1024},

		// The conversation is allowed to be much larger than any single
		// band, because it is where the work actually happens. The bands
		// are what is left of the work after it stops being current.
		Conversation: BandSettings{Enabled: true, Budget: 64 * 1024},
	}
}

// AllBands lists the bands in context order, coarsest and most stable first.
func AllBands() []Band {
	return []Band{BandSoul, BandMemory, Band64x, Band8x, BandSession}
}
