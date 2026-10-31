package gradeedition2

import "github.com/waywardgeek/ensemble/internal/grade"

// Chapter4 retains the original jobs score and separately grades an external
// consumer. Source review still establishes ownership and absence of globals;
// observable diagnostics and an import graph cannot prove those properties.
func Chapter4(bin, core, consumer string, example ConsumerCase) (grade.Report, error) {
	report, err := Chapter4Jobs(bin)
	if err != nil {
		return grade.Report{}, err
	}
	checks := append(report.Checks, consumerTopology(core, consumer))
	checks = append(checks, exerciseConsumer(consumer, example)...)
	return grade.NewTitledReport("Edition 2, Chapter 4 — Jobs (100) and external consumer (60)", checks, report.Stderr), nil
}

// Chapter4Jobs grades the executable protocol without assuming the consumer's
// implementation language. The independent consumer remains a separate task.
func Chapter4Jobs(bin string) (grade.Report, error) {
	res, err := grade.Ch4Run(bin)
	if err != nil {
		return grade.Report{}, err
	}
	parity, err := Chapter3(bin)
	if err != nil {
		return grade.Report{}, err
	}
	res.Ch3, res.Ch3Err = parity.Checks, ""
	checks := grade.Ch4Evaluate(res)
	// The original kill scenario starts its last wait AFTER the kill. Keep the
	// useful check without claiming it proves notification of a blocked waiter.
	for i := range checks {
		if checks[i].ID == "killjob" {
			checks[i].Title = "kill_job stops the process; subsequent waits report killed"
		}
	}
	return grade.NewTitledReport("Chapter 4 — Executable jobs only (100); external consumer not included", checks, res.HelpersErr), nil
}
