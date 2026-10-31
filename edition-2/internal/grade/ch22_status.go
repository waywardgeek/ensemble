package grade

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

// Parsing and assertions for the two behavioural chapter 22 checks.
//
// The report is read with tolerant patterns anchored on the quantity being
// named rather than on any particular sentence. Chapter 21 shipped a check
// that asserted the literal text "404" and so passed agents that had
// dropped the is_error flag entirely -- it was testing the prose and not
// the thing the prose described. Here the labels are the only fixed part;
// every number is compared against a figure computed independently of the
// agent.

var (
	ch22ReModel = regexp.MustCompile(`(?i)model:\s*(\S+)`)
	ch22ReLast  = regexp.MustCompile(
		`(?i)last[^:\n]*:\s*(\d+)\s*in\D+(\d+)\s*out\D+(\d+)\s*cache read\D+(\d+)\s*cache write`)
	ch22ReTotal = regexp.MustCompile(
		`(?i)(?:session total|total)[^:\n]*:\s*(\d+)\s*in\D+(\d+)\s*out\D+(\d+)\s*cache read\D+(\d+)\s*cache write`)
	ch22ReRate = regexp.MustCompile(`(?i)cache hit rate[^:\n]*:\s*([0-9.]+)\s*%`)
	ch22ReCost = regexp.MustCompile(`(?i)cost[^:\n]*:\s*~?\$\s*([0-9.]+)`)
)

func ch22Int(m []string, i int) int {
	if len(m) <= i {
		return -1
	}
	n, err := strconv.Atoi(m[i])
	if err != nil {
		return -1
	}
	return n
}

func ch22Float(m []string, i int) float64 {
	if len(m) <= i {
		return math.NaN()
	}
	f, err := strconv.ParseFloat(m[i], 64)
	if err != nil {
		return math.NaN()
	}
	return f
}

// ch22CheckStatus verifies what agent_status reported against figures the
// grader computed from the scripted session. Returns a list of specific
// complaints; empty means the report was right.
func ch22CheckStatus(status string, e ch22Expect) []string {
	var bad []string
	note := func(f string, a ...any) { bad = append(bad, fmt.Sprintf(f, a...)) }

	if m := ch22ReModel.FindStringSubmatch(status); m == nil {
		note("the report does not name the current model")
	} else if got := strings.Trim(m[1], `",`); got != e.model {
		note("reports model %q; the session's last request was sent as %q", got, e.model)
	}

	if m := ch22ReLast.FindStringSubmatch(status); m == nil {
		note("the report does not give the last response's four token counts")
	} else {
		for _, c := range []struct {
			label string
			got   int
			want  int
		}{
			{"input", ch22Int(m, 1), e.lastIn},
			{"output", ch22Int(m, 2), e.lastOut},
			{"cache read", ch22Int(m, 3), e.lastRead},
			{"cache write", ch22Int(m, 4), e.lastWrite},
		} {
			if c.got != c.want {
				note("last response %s: reports %d, the vendor returned %d", c.label, c.got, c.want)
			}
		}
	}

	if m := ch22ReTotal.FindStringSubmatch(status); m == nil {
		note("the report does not give the session's four token totals")
	} else {
		for _, c := range []struct {
			label string
			got   int
			want  int
		}{
			{"input", ch22Int(m, 1), e.totIn},
			{"output", ch22Int(m, 2), e.totOut},
			{"cache read", ch22Int(m, 3), e.totRead},
			{"cache write", ch22Int(m, 4), e.totWrite},
		} {
			if c.got != c.want {
				note("session total %s: reports %d, the vendor returned %d across the session",
					c.label, c.got, c.want)
			}
		}
	}

	if m := ch22ReRate.FindStringSubmatch(status); m == nil {
		note("the report does not give a cache hit rate")
	} else if got := ch22Float(m, 1); math.Abs(got-e.cacheRatePct) > 0.15 {
		note("cache hit rate: reports %.1f%%, but %d of %d input tokens came from cache (%.1f%%)",
			got, e.totRead, e.totIn+e.totRead, e.cacheRatePct)
	}
	return bad
}

// ch22CheckCost is the one assertion with an observably wrong answer
// available. It asserts BOTH that the figure is the sum of each model's
// tokens at that model's own rate, AND that it is not the whole session
// priced at the current model's rate -- which is the bug that shipped.
//
// The second clause is what makes this a bug test rather than an
// arithmetic test, and it is worthless unless the two formulas actually
// disagree for the session that was driven. So that is checked first, and
// if they agree the GRADER is reported as broken rather than the student
// as wrong.
func ch22CheckCost(status string, e ch22Expect) (bad []string, graderBroken string) {
	if len(e.models) < 2 {
		return nil, fmt.Sprintf(
			"the scripted session billed only %d model(s) (%v); a single-model "+
				"session cannot distinguish per-model pricing from one flat rate",
			len(e.models), e.models)
	}
	correct := fmt.Sprintf("%.4f", e.costCorrect)
	flat := fmt.Sprintf("%.4f", e.costIfOneRate)
	if correct == flat {
		return nil, fmt.Sprintf(
			"per-model pricing and one flat rate both come to $%s for this "+
				"session, so the check would pass either way", correct)
	}

	m := ch22ReCost.FindStringSubmatch(status)
	if m == nil {
		return []string{"the report does not give a session cost"}, ""
	}
	got := ch22Float(m, 1)

	// Tolerance, not equality: the cost is a sum over a map, so the order
	// of the floating-point additions varies from run to run and the last
	// digit is not reproducible.
	const tol = 0.0005
	if math.Abs(got-e.costCorrect) <= tol {
		return nil, ""
	}
	if math.Abs(got-e.costIfOneRate) <= tol {
		return []string{fmt.Sprintf(
			"session cost $%.4f is the WHOLE session priced at the current "+
				"model's rate. Each model's tokens must be priced at that "+
				"model's own rate, which comes to $%.4f here (%v were used)",
			got, e.costCorrect, e.models)}, ""
	}
	return []string{fmt.Sprintf(
		"session cost $%.4f; each model's tokens at its own rate comes to $%.4f (%v were used)",
		got, e.costCorrect, e.models)}, ""
}

// ch22StatusToolDeclared reports whether any skill the student SHIPS
// declares the status tool.
//
// The driven run uses a skill file the grader writes, which is what makes
// it deterministic -- and also what makes it blind. Chapter 21's feature
// was complete, correct and dead on arrival because the shipped skill
// never listed its tool, and every check was green. Every isolation a test
// buys is a thing it stops observing, so the shipped file is inspected
// separately.
func ch22StatusToolDeclared(dir string) (bool, string) {
	var found []string
	_ = filepath.Walk(dir, func(p string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || !strings.EqualFold(info.Name(), "SKILL.md") {
			return nil
		}
		b, err := os.ReadFile(p)
		if err != nil {
			return nil
		}
		if strings.Contains(string(b), "agent_status") {
			rel, _ := filepath.Rel(dir, p)
			found = append(found, rel)
		}
		return nil
	})
	if len(found) == 0 {
		return false, "no skill file in the tree declares agent_status, so the " +
			"tool is unreachable however well it is implemented"
	}
	return true, "declared in " + strings.Join(found, ", ")
}
