package grade

import "fmt"

// Chapter 19 checks: leaving one vendor's account system for another's.
//
// What this chapter grades is not "did you call a different URL". It is
// whether eighteen chapters of vendor-seam work actually paid for itself when
// a vendor changed the deal. So the checks are deliberately weighted toward
// the things a shortcut would skip: verifying a signature rather than
// decoding a token, noticing a failure that arrives with a 200, and keeping
// the cache boundary while moving it onto a different wire format.
//
// Every check reads either the wire the student's agent produced or the
// observations of a fake server that enforced the vendor's published rules.
// None of them reads the student's source. A student may lay the code out
// however they like; what is graded is what came out of the socket.

// ch19Table is the published contract: seven checks, one hundred points.
var ch19Table = []struct {
	ID     string
	Points int
	Why    string
}{
	{"credential-provider", 15,
		"An API key and an OAuth token reach the HTTP layer through one interface, " +
			"and the code that builds a request never learns which it got."},
	{"oauth-flow", 15,
		"PKCE with S256, state and nonce validated, the ID token's signature checked " +
			"against JWKS, and the issued client_id persisted instead of the registration " +
			"placeholder."},
	{"token-refresh", 15,
		"Refresh happens before expiry, rotates the stored refresh token atomically, " +
			"and treats the terminal errors as needing a new authorization."},
	{"responses-format", 15,
		"store:false, stream:true, no parameter the plan route forbids, a developer " +
			"role rather than a system one, and success claimed only on response.completed."},
	{"cache-breakpoints", 10,
		"Chapter 18's caching carried exactly as far as each route allows: none on the " +
			"plan route, which refuses it, and explicit breakpoints on content blocks on " +
			"the metered route, with the mode declared only when one attached."},
	{"reasoning-summaries", 15,
		"Summaries stream incrementally and arrive distinguishable from the answer, " +
			"which is the accessibility payoff the migration exists for."},
	{"billing-mode", 15,
		"Plan-usage errors stop inference rather than silently spending money, and " +
			"any fallback to a metered key is announced before the first metered request."},
}

// Ch19Result is what the runner observes. Errs maps a check ID to the reason
// it failed; a check ID present with an empty string passed.
type Ch19Result struct {
	Errs  map[string]string
	Fatal string
}

// ran marks a check as exercised. A check that never ran reports that fact
// rather than silently scoring zero, so a harness failure is distinguishable
// from a student failure.
func (r *Ch19Result) ran(id string) {
	if r.Errs == nil {
		r.Errs = map[string]string{}
	}
	if _, ok := r.Errs[id]; !ok {
		r.Errs[id] = ""
	}
}

func (r *Ch19Result) fail(id, format string, args ...any) {
	r.ran(id)
	if r.Errs == nil {
		r.Errs = map[string]string{}
	}
	if existing := r.Errs[id]; existing != "" {
		return // keep the first failure: it is usually the cause of the rest
	}
	r.Errs[id] = fmt.Sprintf(format, args...)
}

// Ch19Checks converts observations into the graded result.
func Ch19Checks(r Ch19Result) []Check {
	out := make([]Check, 0, len(ch19Table))
	for _, w := range ch19Table {
		c := Check{ID: w.ID, Title: w.Why, Points: w.Points, Earned: w.Points, Passed: true}
		switch msg, ran := r.Errs[w.ID]; {
		case r.Fatal != "":
			c.failf("%s", r.Fatal)
		case !ran:
			c.failf("not exercised: an earlier failure stopped the run before this check")
		case msg != "":
			c.failf("%s", msg)
		}
		out = append(out, c)
	}
	return out
}
