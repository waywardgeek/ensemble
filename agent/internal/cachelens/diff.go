package cachelens

import (
	"fmt"
	"regexp"
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

	// Breakpoints counts the cache_control markers in the current request.
	// The comparison above is deliberately blind to markers, so this is the
	// only place a lost breakpoint shows up. Report it: a request whose
	// sections are all identical and whose breakpoint count just fell to zero
	// is uncached, and every other number here looks perfect.
	Breakpoints int
}

// breakpointRE matches one cache_control marker, in any JSON formatting, along
// with the comma that separates it from the key before it. Stripped before
// comparison because a breakpoint is an instruction to the provider, not
// conversation content: the cache key is the content of the prefix.
//
// This matters because the markers MOVE. The rolling breakpoint advances every
// turn, so a block carrying a marker in one request carries none in the next.
// Comparing raw bytes would find that difference deep inside the common
// prefix, classify a healthy turn as "edited", and cut the reported cacheable
// prefix off at the old marker. The instrument would then under-report the
// cache on every single turn, which is the same false alarm the array-append
// case exists to prevent.
//
// A targeted strip rather than a JSON round trip, and the distinction is the
// point. Unmarshalling and re-marshalling would reorder object keys and
// rewrite the byte counts, and those counts are reported as a property of the
// real request and compared against the provider's own figures. Both sides
// getting the same distortion would keep equality honest and make the
// measurement a fiction.
//
// It is a pattern rather than a literal string for a reason worth keeping. The
// first version of this was a constant holding the exact bytes our renderer
// emits, on the argument that we control the emitter so the string is known.
// We do control the emitter, but the lens does not compare what the emitter
// wrote: it compares what the canonicalizer produced, and that is re-indented
// for human diffing. So the marker arrives as `"cache_control": {` with spaces
// the constant did not have, nothing ever matched, and the strip silently did
// nothing for every request the agent ever sent. Match the shape, not a
// rendering of it.
//
// The marker takes exactly ONE separator, and the whitespace on the same side,
// because a key can be last in its object or first. Removing the key but
// leaving its comma, or its indentation, turns an identical pair into an edited
// one just as effectively as not removing it at all: the leftover is still a
// byte that one request has and the other does not. Our renderers put
// cache_control last, so the leading-comma form is the production case; JSON
// marshalled from a Go map sorts keys and puts it first. Exactly one separator,
// never both, or a marker between two other keys would fuse them together.
var breakpointRE = regexp.MustCompile(
	`,\s*"cache_control"\s*:\s*\{[^{}]*\}` + `|` +
		`\s*"cache_control"\s*:\s*\{[^{}]*\}\s*,` + `|` +
		`\s*"cache_control"\s*:\s*\{[^{}]*\}`)

// stripBreakpoints removes markers, and returns the count it removed.
func stripBreakpoints(s string) (string, int) {
	locs := breakpointRE.FindAllStringIndex(s, -1)
	if len(locs) == 0 {
		return s, 0
	}
	return breakpointRE.ReplaceAllString(s, ""), len(locs)
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
		stripped, n := stripBreakpoints(string(current.Raw[name]))
		d.CurrentBytes += len(stripped)
		d.Breakpoints += n
	}
	allIdentical := true
	stillCacheable := true

	for _, name := range current.Order {
		p, _ := stripBreakpoints(string(prior.Raw[name]))
		c, _ := stripBreakpoints(string(current.Raw[name]))

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
		return fmt.Sprintf("identical (%d bytes, %d breakpoints)", d.CurrentBytes, d.Breakpoints)
	}

	parts := make([]string, 0, len(d.Sections))
	for _, s := range d.Sections {
		parts = append(parts, s.Name+":"+s.Status)
	}
	pct := 0.0
	if d.CurrentBytes > 0 {
		pct = float64(d.CacheableBytes) / float64(d.CurrentBytes) * 100
	}
	// The breakpoint count rides along because the comparison is blind to
	// markers by design. Every section can read "identical" while the request
	// carries no breakpoints at all, and that request is uncached.
	return fmt.Sprintf("%s — cacheable prefix %d of %d bytes (%.1f%%), %d breakpoints",
		strings.Join(parts, " "), d.CacheableBytes, d.CurrentBytes, pct, d.Breakpoints)
}
