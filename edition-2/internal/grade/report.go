package grade

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

// Report is the full verdict for one submission.
type Report struct {
	Title    string  `json:"title"`
	Checks   []Check `json:"checks"`
	Score    int     `json:"score"`
	MaxScore int     `json:"max_score"`
	Passed   bool    `json:"passed"`
	Stderr   string  `json:"stderr,omitempty"`
}

// NewReport scores a Chapter 1 checklist.
func NewReport(checks []Check, stderr string) Report {
	return NewTitledReport("Chapter 1 — A Conversation, the Obvious Way", checks, stderr)
}

// NewTitledReport scores a checklist. A submission passes only if every check
// passes: there is no partial credit for a conversation that does not exist.
func NewTitledReport(title string, checks []Check, stderr string) Report {
	r := Report{Title: title, Checks: checks, Passed: true, Stderr: stderr}
	for _, c := range checks {
		r.MaxScore += c.Points
		r.Score += c.Earned
		if !c.Passed {
			r.Passed = false
		}
	}
	return r
}

// WriteText renders a human-readable report.
func (r Report) WriteText(w io.Writer) {
	title := r.Title
	if title == "" {
		title = "Chapter 1 — A Conversation, the Obvious Way"
	}
	fmt.Fprintln(w, title)
	fmt.Fprintln(w, strings.Repeat("=", 60))
	for _, c := range r.Checks {
		mark := "FAIL"
		if c.Passed {
			mark = "PASS"
		}
		fmt.Fprintf(w, "\n[%s] %-4s %s  (%d/%d)\n", mark, "", c.Title, c.Earned, c.Points)
		for _, d := range c.Details {
			fmt.Fprintf(w, "       %s\n", wrap(d, 70, "       "))
		}
	}
	fmt.Fprintln(w, "\n"+strings.Repeat("-", 60))
	fmt.Fprintf(w, "score: %d/%d — %s\n", r.Score, r.MaxScore, map[bool]string{true: "PASS", false: "FAIL"}[r.Passed])
	if strings.TrimSpace(r.Stderr) != "" {
		fmt.Fprintf(w, "\nprogram stderr:\n%s\n", indent(strings.TrimRight(r.Stderr, "\n"), "  | "))
	}
}

// WriteJSON renders the machine-readable report.
func (r Report) WriteJSON(w io.Writer) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(r)
}

func wrap(s string, width int, pad string) string {
	words := strings.Fields(s)
	var lines []string
	cur := ""
	for _, w := range words {
		if cur == "" {
			cur = w
			continue
		}
		if len(cur)+1+len(w) > width {
			lines = append(lines, cur)
			cur = w
			continue
		}
		cur += " " + w
	}
	if cur != "" {
		lines = append(lines, cur)
	}
	return strings.Join(lines, "\n"+pad)
}

func indent(s, pad string) string {
	lines := strings.Split(s, "\n")
	for i := range lines {
		lines[i] = pad + lines[i]
	}
	return strings.Join(lines, "\n")
}
