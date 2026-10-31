// Package gradeedition2 holds corrections to the course graders without
// changing the first edition's checks or historical student results.
package gradeedition2

import (
	"fmt"

	"github.com/waywardgeek/ensemble/internal/grade"
)

// Chapter1 evaluates the original run evidence, preserving all seven checks.
// The original fake records headers but checks only whether the API version
// exists. Check its documented value here without changing the original fake.
func Chapter1(res *grade.RunResult) grade.Report {
	checks := grade.Evaluate(res)
	for i := range checks {
		if checks[i].ID != "wire" {
			continue
		}
		for _, request := range res.Records {
			if got := request.Header.Get("anthropic-version"); got != "2023-06-01" {
				if checks[i].Passed {
					// The original success note no longer describes this verdict.
					checks[i].Details = nil
				}
				checks[i].Passed = false
				checks[i].Earned = 0
				checks[i].Details = append(checks[i].Details,
					fmt.Sprintf("request %d: anthropic-version is %q, expected 2023-06-01", request.Seq, got))
			}
		}
	}
	return grade.NewTitledReport("Edition 2, Chapter 1 — One Loop, Sixty Billion Dollars", checks, res.Stderr)
}
