package cachelens

import (
	"strings"
	"testing"

	"github.com/waywardgeek/ensemble/agent/internal/common"
)

// The verdict the lens prints when the prefix held but the provider served
// little or nothing is the part a human acts on, so it is the part most worth
// pinning down. Before the model table carried a caching style and a minimum
// cacheable size, every one of these cases produced the SAME sentence, which
// offered three possible causes and therefore diagnosed nothing. Worse, it
// named "no breakpoint is set" as a cause on models where setting one is
// impossible, which is how we once talked ourselves into believing a vendor
// had no prompt caching at all.

// lensProbe drives a Lens through two requests and a usage report, returning
// every line it printed.
func lensProbe(t *testing.T, model, prior, current string, u common.Usage) []string {
	t.Helper()

	var lines []string
	l := New(t.TempDir(), func(s string) { lines = append(lines, s) })

	l.ObserveRequest(model, []byte(prior))
	l.ObserveRequest(model, []byte(current))
	l.ObserveUsage(u)

	return lines
}

func joined(lines []string) string { return strings.Join(lines, "\n") }

// geminiBodies returns a prior/current pair that differs ONLY by an appended
// turn, so the prefix is stable and any miss is not our fault.
func geminiBodies() (string, string) {
	big := strings.Repeat("stable system preamble that is reused every turn. ", 200)
	prior := `{"contents":[{"role":"user","parts":[{"text":"` + big + `"}]}],` +
		`"tools":[{"functionDeclarations":[{"name":"read"}]}]}`
	current := `{"contents":[{"role":"user","parts":[{"text":"` + big + `"}]},` +
		`{"role":"model","parts":[{"text":"ok"}]}],` +
		`"tools":[{"functionDeclarations":[{"name":"read"}]}]}`
	return prior, current
}

func TestBelowTheFloorIsReportedAsNotEligible(t *testing.T) {
	// A prompt under the vendor's minimum was never a caching candidate, so a
	// zero is the correct answer rather than a symptom. Reporting it as a miss
	// is precisely how three short live requests to Gemini got written up as
	// "caching is not implemented".
	prior, current := geminiBodies()

	out := joined(lensProbe(t, "gemini-3.8-flash", prior, current,
		common.Usage{Input: 500}))

	if !strings.Contains(out, "NOT ELIGIBLE") {
		t.Errorf("want NOT ELIGIBLE for a 500-token prompt under the 4096 floor, got:\n%s", out)
	}
	if strings.Contains(out, "NO BREAKPOINT") {
		t.Errorf("must not blame a missing marker below the floor, got:\n%s", out)
	}
}

func TestImplicitModelIsNeverBlamedForAMissingMarker(t *testing.T) {
	// Gemini finds the repeated prefix itself. Sending no cache directives is
	// the entire interface, not an omission, so a miss here must never point
	// at a marker: there is nothing to mark, and a reader sent looking for one
	// will "fix" a request that was already correct.
	prior, current := geminiBodies()

	out := joined(lensProbe(t, "gemini-3.8-flash", prior, current,
		common.Usage{Input: 21660}))

	if !strings.Contains(out, "IMPLICIT-CACHING") {
		t.Errorf("want the implicit-caching verdict above the floor, got:\n%s", out)
	}
	if strings.Contains(out, "NO BREAKPOINT") {
		t.Errorf("blamed a missing marker on an implicit-caching model, got:\n%s", out)
	}
	if strings.Contains(out, "NOT ELIGIBLE") {
		t.Errorf("21660 tokens is over the 4096 floor, got:\n%s", out)
	}
}

func TestExplicitModelWithNoMarkerIsOurBug(t *testing.T) {
	// The mirror image: on Anthropic an unmarked prefix really is billed in
	// full every turn, and that IS ours to fix. The same silence must produce
	// opposite verdicts on the two caching styles, or the field is pointless.
	big := strings.Repeat("stable system preamble that is reused every turn. ", 200)
	prior := `{"system":[{"type":"text","text":"` + big + `"}],` +
		`"messages":[{"role":"user","content":"hi"}],"tools":[{"name":"read"}]}`
	current := `{"system":[{"type":"text","text":"` + big + `"}],` +
		`"messages":[{"role":"user","content":"hi"},{"role":"assistant","content":"yo"}],` +
		`"tools":[{"name":"read"}]}`

	out := joined(lensProbe(t, "claude-sonnet-5", prior, current,
		common.Usage{Input: 20000}))

	if !strings.Contains(out, "NO BREAKPOINT") {
		t.Errorf("want NO BREAKPOINT on an explicit model carrying zero markers, got:\n%s", out)
	}
	if strings.Contains(out, "IMPLICIT") {
		t.Errorf("claude-sonnet-5 caches explicitly, got:\n%s", out)
	}
}

func TestUnknownModelAdmitsItDoesNotKnow(t *testing.T) {
	// An unknown model has no floor and no caching style, so the honest report
	// is that no diagnosis is available — not a guess dressed as one.
	prior, current := geminiBodies()

	out := joined(lensProbe(t, "no-such-model-9000", prior, current,
		common.Usage{Input: 20000}))

	if !strings.Contains(out, "not in the model table") {
		t.Errorf("want an explicit admission of ignorance, got:\n%s", out)
	}
}

func TestEveryModelInTheTableDeclaresACachingStyle(t *testing.T) {
	// The whole point of the column is that it is never absent: a zero value
	// would silently fall through the verdict switch and print nothing, which
	// is the failure mode this test exists to prevent.
	for _, m := range []string{
		"claude-opus-4-6", "claude-opus-5", "claude-sonnet-5",
		"gpt-6-astra", "gpt-5.6-sol",
		"gemini-3.8-flash", "gemini-3.1-pro-preview",
	} {
		f, ok := common.LookupModel(m)
		if !ok {
			t.Errorf("%s: not in the model table", m)
			continue
		}
		if f.Caching != common.CacheExplicit && f.Caching != common.CacheImplicit {
			t.Errorf("%s: Caching is unset — every model we speak to supports caching", m)
		}
		if f.MinCacheTokens <= 0 {
			t.Errorf("%s: MinCacheTokens is unset, so a miss cannot be told from an "+
				"ineligible request", m)
		}
	}
}
