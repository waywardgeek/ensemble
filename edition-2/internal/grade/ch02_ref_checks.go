package grade

// Chapter 2, Ref amendment: a blob's location is a Ref (a Kind saying WHAT the
// locator is, plus the locator), not a bare path.
//
// Five checks, three skills:
//
//	ref-roundtrip  serialization — all three kinds survive the log
//	ref-oldformat  serialization — a pre-Ref log is refused, not coerced
//	ref-zerokind   serialization — the zero value is not a kind
//	ref-render     rendering     — RefURI becomes the vendor's remote form
//	ref-redaction  carry-forward — a stub keeps the superseded Ref
//
// The three serialization IDs share one point budget. They are separate IDs so
// a submission is told WHICH symptom fired, not charged three times for one
// bug.

import (
	"fmt"
	"strings"
)

// refKindOf reads a normalized "kind" value, accepting either the wire number
// the reference emits or a spelled-out name. The numbers are what the fixtures
// supply and what a correct submission echoes back; the names are accepted
// because failing someone for a String()-based encoding is a spelling lesson,
// not a design one.
func refKindOf(v any) (int, bool) {
	switch t := v.(type) {
	case float64:
		return int(t), true
	case string:
		switch normName(t) {
		case "path":
			return 1, true
		case "uri":
			return 2, true
		case "handle":
			return 3, true
		}
	}
	return 0, false
}

// refBlobs walks a dumped log and returns every blob part it can find, in
// order, as (mime, kind, locator) triples. It also reports whether any blob
// still carries the pre-Ref "path" spelling.
func refBlobs(lines []Ch2LogLine) (blobs [][3]string, sawPath bool) {
	var walk func(parts []any)
	walk = func(parts []any) {
		for _, p := range parts {
			m, ok := p.(map[string]any)
			if !ok {
				continue
			}
			if inner := getSlice(m, "parts"); inner != nil {
				walk(inner)
			}
			if normName(getStr(m, "type", "kind")) != normName("blob") {
				continue
			}
			if getStr(m, "path") != "" {
				sawPath = true
			}
			mime := getStr(m, "mime", "mime_type", "media_type")
			ref := getMap(m, "ref")
			if ref == nil {
				blobs = append(blobs, [3]string{mime, "<no ref>", ""})
				continue
			}
			k, ok := refKindOf(get(ref, "kind"))
			if !ok {
				blobs = append(blobs, [3]string{mime, "<bad kind>", getStr(ref, "locator")})
				continue
			}
			blobs = append(blobs, [3]string{mime, fmt.Sprint(k), getStr(ref, "locator")})
		}
	}
	for _, l := range lines {
		if msg := getMap(l.Data, "message"); msg != nil {
			walk(getSlice(msg, "parts"))
		}
		if resp := getMap(l.Data, "response"); resp != nil {
			walk(getSlice(resp, "parts"))
		}
		if tool := getMap(l.Data, "tool"); tool != nil {
			walk(getSlice(tool, "parts"))
		}
	}
	return blobs, sawPath
}

// ch2RefRoundTripCheck: a BlobPart of each of the three kinds survives a trip
// through the log unchanged.
//
// Graded on `dump`, which LOADS the log into typed parts and re-marshals them,
// rather than on a render. The three kinds are a property of our own format,
// and no vendor's opinion of them should be able to make this pass or fail.
func ch2RefRoundTripCheck(r *Ch2Result) Check {
	c := Check{ID: "ref-roundtrip", Title: "all three RefKinds survive the log", Points: ch2RefRoundTrip, Passed: true, Earned: ch2RefRoundTrip}

	if strings.TrimSpace(r.RefDump) == "" {
		c.failf("`dump` produced nothing for a log containing three blobs, one of each RefKind "+
			"(stderr: %.300s). All three kinds are valid and this log must load.", r.RefDumpErr)
		return c
	}
	lines, perr := parseLogLines(r.RefDump)
	if perr != "" {
		c.failf("the re-emitted log is not valid JSON-lines: %s", perr)
		return c
	}
	blobs, sawPath := refBlobs(lines)
	if sawPath {
		c.failf("a re-emitted blob still carries the pre-Ref %q field. A blob's location is a Ref now; "+
			"%q is retained only to RECOGNIZE an old log and refuse it.", "path", "path")
	}
	want := [][3]string{
		{"text/plain", "1", RefLocatorPath},
		{"image/png", "2", RefLocatorURI},
		{"application/json", "3", RefLocatorHandle},
	}
	if len(blobs) != len(want) {
		c.failf("expected %d blob parts to survive the round trip, found %d: %v. "+
			"Each of RefPath, RefURI and RefHandle must survive marshal and unmarshal.",
			len(want), len(blobs), blobs)
		return c
	}
	for i, w := range want {
		got := blobs[i]
		if got == w {
			continue
		}
		c.failf("blob %d round-tripped as (mime=%s kind=%s locator=%s), want (mime=%s kind=%s locator=%s). "+
			"A Ref carries BOTH halves: the kind says what the locator is, and dropping either one "+
			"turns a remote reference into a filename that never existed.",
			i+1, got[0], got[1], got[2], w[0], w[1], w[2])
	}
	if c.Passed {
		c.Details = append(c.Details, "RefPath, RefURI and RefHandle all survive marshal/unmarshal")
	}
	return c
}

// refRefusal grades "this log must not load". The load must FAIL: nothing that
// looks like a loaded log may reach stdout, and the program must say why.
func refRefusal(c *Check, what, out, errOut string, code int, locator string) {
	loaded := strings.TrimSpace(out) != ""
	if loaded && locator != "" && strings.Contains(out, locator) {
		c.failf("%s LOADED: the locator %q came back out of `dump`. This log must be refused, "+
			"not repaired. Silently coercing it is the exact anti-pattern the loader already "+
			"refuses for unknown part types — a value you quietly convert is a value you will "+
			"debug in production.", what, locator)
		return
	}
	if loaded {
		c.failf("%s LOADED: `dump` emitted %d bytes instead of refusing the log.", what, len(strings.TrimSpace(out)))
		return
	}
	if code == 0 && strings.TrimSpace(errOut) == "" {
		c.failf("%s was not loaded, but the program exited 0 and said nothing. A refusal must be "+
			"LOUD: a non-zero exit and a diagnostic naming what was wrong.", what)
		return
	}
	if strings.TrimSpace(errOut) == "" {
		c.failf("%s was refused (exit %d) with no diagnostic on stderr. Say what was wrong with "+
			"the log, or the next person to hit this has only an exit code.", what, code)
	}
}

// ch2RefOldFormatCheck: a log written BEFORE the Ref type must be refused.
//
// This is a HARD BREAK, on purpose. The tempting one-liner reads the old
// "path" into a RefPath, and it would survive every local-file log ever
// written — right up to the first remote reference, which it would silently
// turn into a local filename.
func ch2RefOldFormatCheck(r *Ch2Result) Check {
	c := Check{ID: "ref-oldformat", Title: "a pre-Ref log is refused, not coerced", Points: ch2RefOldFormat, Passed: true, Earned: ch2RefOldFormat}
	refRefusal(&c, "a blob carrying the pre-Ref \"path\" field and no \"ref\"",
		r.RefOldFormatOut, r.RefOldFormatErr, r.RefOldFormatCode, RefLocatorPath)
	if c.Passed {
		c.Details = append(c.Details, "old-format log refused loudly; \"path\" is not coerced into a Ref")
	}
	return c
}

// ch2RefZeroKindCheck: RefKind's constants start at iota+1 precisely so that
// the zero value cannot be mistaken for a real kind. A Ref that was never
// filled in must be detectably empty rather than silently "a path".
func ch2RefZeroKindCheck(r *Ch2Result) Check {
	c := Check{ID: "ref-zerokind", Title: "the zero value is not a RefKind", Points: ch2RefZeroKind, Passed: true, Earned: ch2RefZeroKind}
	refRefusal(&c, "a blob whose ref has kind 0",
		r.RefZeroKindOut, r.RefZeroKindErr, r.RefZeroKindCode, RefLocatorPath)
	if c.Passed {
		c.Details = append(c.Details, "kind 0 refused; the constants start at iota+1 so the zero value is not a kind")
	}
	return c
}

// ch2RefRenderCheck: a RefURI must render to the vendor's REMOTE-REFERENCE
// form, not to a local path and not to inlined bytes.
//
// Graded on Gemini, whose remote form was verified against the published
// schema: a Part carries `fileData`, and FileData is {mimeType, fileUri}.
// See book/.wire-facts.md for the citation.
func ch2RefRenderCheck(r *Ch2Result) Check {
	c := Check{ID: "ref-render", Title: "RefURI renders to the vendor's remote-reference form", Points: ch2RefRender, Passed: true, Earned: ch2RefRender}

	if strings.TrimSpace(r.RefRender) == "" {
		c.failf("gemini: `render` produced nothing for a log containing a RefURI blob (stderr: %.300s)", r.RefRenderErr)
		return c
	}
	// Ruling 1: base64 is a rendering decision, and a RefURI is precisely the
	// case where the vendor fetches the bytes itself. Inlining here would also
	// mean reading a URL as though it were a filename.
	if strings.Contains(r.RefRender, "inlineData") || strings.Contains(r.RefRender, "inline_data") {
		c.failf("gemini: the RefURI blob was rendered as inline bytes. A RefURI is a reference the " +
			"vendor resolves for itself; inlining it means something tried to READ a URI as a path.")
	}
	m := decode(&c, "gemini(ref)", r.RefRender)
	if m == nil {
		return c
	}
	contents, _ := m["contents"].([]any)
	if len(contents) == 0 {
		c.failf("gemini: no `contents` array in the RefURI render")
		return c
	}
	found := false
	for _, raw := range contents {
		turn, _ := raw.(map[string]any)
		parts, _ := turn["parts"].([]any)
		for _, p := range parts {
			part, _ := p.(map[string]any)
			// Vendor field names are graded EXACTLY: `fileData` and `fileUri`
			// are Gemini's spellings, not the student's.
			fd, ok := part["fileData"].(map[string]any)
			if !ok {
				continue
			}
			found = true
			uri, _ := fd["fileUri"].(string)
			if uri != RefLocatorURI {
				c.failf("gemini: fileData.fileUri is %q, want %q. The locator is carried verbatim.", uri, RefLocatorURI)
			}
			if mt, _ := fd["mimeType"].(string); mt != "image/png" {
				c.failf("gemini: fileData.mimeType is %q, want %q.", mt, "image/png")
			}
		}
	}
	if !found {
		c.failf("gemini: no `fileData` part in the request. A RefURI is the one kind the vendor can " +
			"fetch for itself, and Gemini spells that as a fileData part carrying {mimeType, fileUri}. " +
			"This is the case a bare path could never express.")
	}
	if c.Passed {
		c.Details = append(c.Details, "RefURI rendered as a Gemini fileData part, not a path and not inline bytes")
	}
	return c
}

// ch2RefRedactionCheck: redaction is recoverable BY CONSTRUCTION.
//
// A RedactedPart is synthesized by the reducer and carries forward the Ref of
// the part it supersedes. The stub says how many bytes went; the Ref still
// says where they are. Nothing records the redaction separately, because the
// log already does, permanently.
func ch2RefRedactionCheck(r *Ch2Result) Check {
	c := Check{ID: "ref-redaction", Title: "a stub carries forward the superseded Ref", Points: ch2RefRedaction, Passed: true, Earned: ch2RefRedaction}

	if strings.TrimSpace(r.RefRedacted) == "" {
		c.failf("gemini: rendering the redacted log produced nothing (stderr: %.300s)", r.RefRedactedErr)
		return c
	}
	// NEGATIVE CONTROL. Without this, a submission that simply never renders
	// tool results would pass the "secret is absent" half for the wrong reason.
	if strings.TrimSpace(r.RefRedactedPlain) == "" {
		c.failf("gemini: rendering the UN-redacted control log produced nothing (stderr: %.300s). "+
			"Without the control this check cannot show that redaction did anything.", r.RefRedactedPlainErr)
		return c
	}
	if !strings.Contains(r.RefRedactedPlain, RefRedactionSecret) {
		c.failf("gemini: the un-redacted control does not contain the tool output (%q) at all, "+
			"so this check cannot show that redaction removed it.", RefRedactionSecret)
		return c
	}
	if !strings.Contains(r.RefRedactedPlain, RefRedactionLocator) {
		c.failf("gemini: the un-redacted control does not mention the blob's locator (%q), "+
			"so this check cannot show that the locator SURVIVED redaction.", RefRedactionLocator)
		return c
	}
	if strings.Contains(r.RefRedacted, RefRedactionSecret) {
		c.failf("gemini: redacted tool output (%q) still appears in the rendered request.", RefRedactionSecret)
	}
	if !strings.Contains(r.RefRedacted, RefRedactionLocator) {
		c.failf("gemini: the superseded blob's locator (%q) did NOT survive the redaction. "+
			"A RedactedPart carries forward the Ref of the part it supersedes: the stub says the "+
			"bytes are gone, and the Ref still says where they are, so a later turn can fetch them "+
			"again if it must. Dropping it makes redaction unrecoverable and needs a side table to "+
			"undo — which is the storage this design exists to avoid.", RefRedactionLocator)
	}
	if c.Passed {
		c.Details = append(c.Details, "tool output stubbed; the superseded blob's Ref carried forward into the stub")
	}
	return c
}
