package cachelens

import (
	"fmt"
	"strings"
)

// Status classifies how one section changed between two requests.
const (
	// StatusIdentical: byte-for-byte the same. This is the only status that
	// lets a provider serve the section from cache.
	StatusIdentical = "identical"
	// StatusAppended: the section grew at the end and kept its prefix. Healthy
	// and expected for the dialogue, which gains a turn every round.
	StatusAppended = "appended"
	// StatusTruncated: the section shrank but kept its prefix. Expected after a
	// compaction or a reset.
	StatusTruncated = "truncated"
	// StatusEdited: bytes changed in place. Harmless in the dialogue, a bug
	// anywhere above it, because it voids the cache for everything after.
	StatusEdited = "edited"
)

// SectionDiff is the verdict for one section.
type SectionDiff struct {
	Name         string
	Status       string
	PrefixBytes  int
	PriorBytes   int
	CurrentBytes int
}

// Divergence is the verdict for a whole request.
//
// The comparison is per-section rather than over one flat blob, because that is
// how the provider sees it. A cache breakpoint sits at a section boundary, so
// the question that decides a hit is not "how many bytes match" but "which
// sections are untouched, in order, from the front". Flattening first answers a
// question nobody is charged for.
type Divergence struct {
	Identical bool

	// Sections is every section in cache order, each with its own verdict.
	Sections []SectionDiff

	// CacheableBytes is the length of the leading run of identical sections,
	// plus the matching prefix of the first section that changed. This is the
	// most a provider could serve from cache given a perfectly placed
	// breakpoint.
	CacheableBytes int
	CurrentBytes   int

	// Unstable names the first section above the dialogue that changed, or ""
	// if everything above the dialogue held. This is the field worth alerting
	// on: a change here is nondeterminism in content that was supposed to be
	// fixed, and it is ours to fix.
	Unstable       string
	UnstableOffset int

	// DialogueChanged is true when only the dialogue moved. That is the
	// healthy case and must never raise an alarm, because it happens on every
	// single turn.
	DialogueChanged bool
}

// dialogueSection is the one section expected to change every turn.
const dialogueSection = "messages"

// Compare produces a per-section verdict for two consecutive requests.
func Compare(prior, current Sections) Divergence {
	// CurrentBytes counts section payloads only, deliberately excluding the
	// framing Canonical adds. CacheableBytes is summed from the same payloads,
	// and two numbers that get divided must be measured with the same ruler.
	// Using the canonical length here instead would fold this tool's own
	// formatting into a percentage reported as a property of the request.
	d := Divergence{}
	for _, name := range current.Order {
		d.CurrentBytes += len(current.Raw[name])
	}
	allIdentical := true
	stillCacheable := true

	for _, name := range current.Order {
		p := string(prior.Raw[name])
		c := string(current.Raw[name])

		sd := SectionDiff{
			Name:         name,
			PriorBytes:   len(p),
			CurrentBytes: len(c),
			PrefixBytes:  commonPrefix(p, c),
		}
		sd.Status = classify(p, c, sd.PrefixBytes)

		if sd.Status != StatusIdentical {
			allIdentical = false
			if name == dialogueSection {
				d.DialogueChanged = true
			} else if d.Unstable == "" {
				// A section other than the dialogue changed. Record the first
				// one; later ones are downstream of it and not independently
				// interesting.
				d.Unstable = name
				d.UnstableOffset = sd.PrefixBytes
			}
		}

		// The cacheable run ends at the first section that is not identical.
		// Everything after it is past the breakpoint no matter how well it
		// matches, so counting those bytes would overstate what is reachable.
		if stillCacheable {
			if sd.Status == StatusIdentical {
				d.CacheableBytes += len(c)
			} else {
				d.CacheableBytes += sd.PrefixBytes
				stillCacheable = false
			}
		}

		d.Sections = append(d.Sections, sd)
	}

	d.Identical = allIdentical
	return d
}

func commonPrefix(a, b string) int {
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	i := 0
	for i < n && a[i] == b[i] {
		i++
	}
	return i
}

// classify decides how a section changed.
//
// The non-obvious case is an append inside a JSON array, and getting it wrong
// makes the whole instrument useless. Adding a message turns
//
//	[{"a":1}]
//
// into
//
//	[{"a":1},{"b":2}]
//
// so the two strings diverge at the closing bracket: the prior has ']' where
// the current has ','. A naive comparison calls that an edit, and since it
// happens on literally every turn, the tool would cry wolf every single time
// and be switched off within a day. A false alarm that fires constantly is
// worse than no alarm, because it also teaches you to ignore the real one.
//
// So: if the prior's trailing JSON framing is dropped and what remains is a
// prefix of the current, the container grew rather than changed.
func classify(p, c string, prefix int) string {
	switch {
	case p == c:
		return StatusIdentical
	case prefix == len(p):
		return StatusAppended
	case prefix == len(c):
		return StatusTruncated
	}

	const framing = "]}\n\t\r "
	if trimmed := strings.TrimRight(p, framing); trimmed != "" && strings.HasPrefix(c, trimmed) {
		return StatusAppended
	}
	if trimmed := strings.TrimRight(c, framing); trimmed != "" && strings.HasPrefix(p, trimmed) {
		return StatusTruncated
	}
	return StatusEdited
}

// Describe renders a divergence as one line fit for a log.
func Describe(d Divergence) string {
	if d.Identical {
		return fmt.Sprintf("identical (%d bytes)", d.CurrentBytes)
	}

	parts := make([]string, 0, len(d.Sections))
	for _, s := range d.Sections {
		parts = append(parts, s.Name+":"+s.Status)
	}
	pct := 0.0
	if d.CurrentBytes > 0 {
		pct = float64(d.CacheableBytes) / float64(d.CurrentBytes) * 100
	}
	return fmt.Sprintf("%s — cacheable prefix %d of %d bytes (%.1f%%)",
		strings.Join(parts, " "), d.CacheableBytes, d.CurrentBytes, pct)
}
