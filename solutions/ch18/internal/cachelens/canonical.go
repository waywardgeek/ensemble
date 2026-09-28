// Package cachelens analyzes prompt-cache behavior by comparing each outbound
// request against the one before it.
//
// It is a spoke: it depends on internal/common and on nothing else in the tree.
// The engine reaches it through common.CacheLens, wired at the composition root,
// so the request path does not import this package and no cycle arises.
package cachelens

import (
	"bytes"
	"encoding/json"
	"fmt"
)

// Sections holds one request decomposed into its top-level parts, with each
// part's bytes preserved exactly as the vendor serialized them.
//
// json.RawMessage is the point: unmarshalling into map[string]any and
// re-marshalling would re-sort keys and re-format numbers, so the bytes being
// compared would be ours rather than the ones actually sent. A divergence
// introduced by our own round trip is indistinguishable from one that really
// broke the cache.
type Sections struct {
	// Order is the sequence in which the provider concatenates these parts for
	// caching. Recorded per request rather than assumed, so a request captured
	// today stays interpretable after the order changes.
	Order []string
	Raw   map[string]json.RawMessage
}

// cacheOrder returns the order in which the provider concatenates a request for
// cache-prefix purposes.
//
// This is deliberately NOT the JSON key order. anthRequest declares its fields
// as model, max_tokens, system, messages, tools — so the bytes on the wire put
// system before messages before tools. Anthropic's documented cache hierarchy
// is tools, then system, then messages. Diffing in key order would therefore
// show a divergence in "system" while the genuine first divergence was in a
// tool declaration above it, which is the opposite of the tool's purpose.
//
// The trailing parts are the volatile ones. Ordering stable-before-volatile is
// what makes a common prefix long: everything up to the first change is
// cacheable, so the parts that change every turn have to come last.
func cacheOrder(model string) []string {
	// One order for now. It is a function rather than a constant because the
	// next vendor added will need a different one, and the call site should not
	// have to learn that.
	return []string{"tools", "system", "messages"}
}

// Split decomposes a request body into cache-ordered sections.
//
// Keys present in the body but absent from the known order are appended in
// sorted order rather than dropped. Silently discarding an unrecognized key
// would make the canonical form disagree with what was sent, and the first
// symptom would be a prefix that looks stable while the real request changed.
func Split(model string, body []byte) (Sections, error) {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(body, &raw); err != nil {
		return Sections{}, fmt.Errorf("cachelens: request body is not a JSON object: %w", err)
	}

	order := make([]string, 0, len(raw))
	seen := map[string]bool{}
	for _, k := range cacheOrder(model) {
		if _, ok := raw[k]; ok {
			order = append(order, k)
			seen[k] = true
		}
	}

	// Remaining keys, sorted, so the leftovers are deterministic too.
	rest := make([]string, 0, len(raw))
	for k := range raw {
		if !seen[k] {
			rest = append(rest, k)
		}
	}
	sortStrings(rest)
	order = append(order, rest...)

	return Sections{Order: order, Raw: raw}, nil
}

// Canonical renders sections as the byte sequence used for prefix comparison:
// each section's exact bytes, in cache order, one section per line with its
// name. The name is included so a diff names the section it landed in instead
// of leaving the reader to count braces.
func Canonical(s Sections) []byte {
	var b bytes.Buffer
	for _, k := range s.Order {
		b.WriteString(k)
		b.WriteByte('=')
		b.Write(s.Raw[k])
		b.WriteByte('\n')
	}
	return b.Bytes()
}

// Pretty renders sections for human reading: cache order preserved, each
// section indented so that a line-oriented diff lands on a single field rather
// than reporting that one very long line changed.
//
// Prefix computations must use Canonical, never this. Indentation changes byte
// offsets, so an offset measured here would not correspond to anything the
// provider saw.
func Pretty(s Sections) []byte {
	var b bytes.Buffer
	for _, k := range s.Order {
		b.WriteString("// ── ")
		b.WriteString(k)
		b.WriteString(" ──\n")
		var indented bytes.Buffer
		if err := json.Indent(&indented, s.Raw[k], "", "  "); err != nil {
			// Unindentable means unparseable, which Split would already have
			// caught. Emit the raw bytes rather than losing the section.
			b.Write(s.Raw[k])
		} else {
			b.Write(indented.Bytes())
		}
		b.WriteString("\n\n")
	}
	return b.Bytes()
}

// sortStrings is a tiny insertion sort, kept local so this package imports only
// encoding/json, bytes and fmt. The slice is the number of unrecognized keys in
// a request, which is a handful.
func sortStrings(s []string) {
	for i := 1; i < len(s); i++ {
		for j := i; j > 0 && s[j] < s[j-1]; j-- {
			s[j], s[j-1] = s[j-1], s[j]
		}
	}
}
