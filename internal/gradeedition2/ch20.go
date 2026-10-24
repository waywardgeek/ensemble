package gradeedition2

import (
	"github.com/waywardgeek/ensemble/internal/grade"
)

// Chapter20 preserves the two original 100-point exercises: external GUI reuse
// and terminal fleet. Numeric import observations do not replace the separately
// required review of ownership, composition and recursive factory construction.
// legacy selects that export's existing public construction API, never fake hooks.
func Chapter20(bin, core, gui string, legacy bool) (grade.Report, error) {
	return chapter20(bin, core, gui, "", legacy, false)
}

// Chapter20Targets grades the importing core, separate GUI module and explicit
// application artifacts without requiring a library directory to be executable.
func Chapter20Targets(cli, core, guiModule, guiBin string) (grade.Report, error) {
	return chapter20(cli, core, guiModule, guiBin, false, true)
}

func chapter20(bin, core, gui, guiBin string, legacy, current bool) (grade.Report, error) {
	checks := append(grade.Ch24Checks(grade.Ch24Result{}), grade.Ch25Checks(grade.Ch25Result{})...)
	for i := range checks {
		checks[i].Passed = true
		checks[i].Earned = checks[i].Points
		checks[i].Details = nil
	}
	checks[1].Title = "An outside application uses real Agent history, prompt delivery and correlated browser relay"
	checks[5].Title = "The public headless consumer completes MCP without GUI/WebSocket dependencies"
	checks[8].Title = "Parent and two children retain distinct valid saved conversations"
	checks[11].Title = "Three generations yield a child question and return an addressed parent reply"
	checks[6].Title = "Public root has a resolved GUI-free dependency graph and constructible API"
	reuseChecksWith(core, gui, guiBin, legacy, current, checks[:7])
	fleetChecksWith(fleetDriver{bin: bin, current: current}, checks[7:])
	var cleanup = func() {}
	var err error
	if !current {
		guiBin, cleanup, err = grade.Build(gui)
	}
	if err != nil {
		fail(&checks[13], "root chat control: %v", err)
	} else {
		defer cleanup()
		fleetRootChatWith(guiBin, &checks[13], current)
	}
	title := "Edition 2, Chapter 20 — The Ensemble"
	if current {
		title = "Edition 3, Chapter 20 — The Ensemble"
	}
	return grade.NewTitledReport(title, checks, ""), nil
}

// A full real model-switch session already has a verified public driver. Reuse
// its observations without awarding its status/accounting points a second time.
func reuseWire(guiBin, core string, c *grade.Check, current bool) {
	models := statusModels
	var out statusRun
	if current {
		models = []string{"claude-sonnet-4-6", "claude-opus-4-6", "claude-sonnet-4-6"}
		out = statusLaunchModels(guiBin, core, models, true)
	} else {
		out = statusLaunch(guiBin, core)
	}
	if out.fatal != "" {
		fail(c, "actual GUI session: %s", out.fatal)
		return
	}
	if len(out.reqs) != 6 {
		fail(c, "got %d requests, want three two-request turns", len(out.reqs))
		return
	}
	for i, r := range out.reqs {
		if r.JSON()["model"] != models[i/2] {
			fail(c, "request %d model=%v want=%s", i, r.JSON()["model"], models[i/2])
		}
	}
	c.Details = append(c.Details, "Actual stdin turns and WebSocket selection reach the three selected models on subsequent provider requests.")
}

func reuseFail(checks []grade.Check, err error, indices ...int) {
	for _, i := range indices {
		fail(&checks[i], "%v", err)
	}
}
