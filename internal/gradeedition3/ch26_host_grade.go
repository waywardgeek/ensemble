package gradeedition3

import "github.com/waywardgeek/ensemble/internal/grade"

// A passing synthetic-candidate unit test is not native acceptance. Only invoke
// this against a real CLI after fresh authorization for the whole primary display.
func Chapter26HostScreens(config HostNativeConfig) (grade.Report, error) {
	if err := validateHostNativeConfig(config); err != nil {
		return grade.Report{}, err
	}
	check := grade.Check{ID: "host-screen-delivery", Title: "Authorized native primary-display pixels, delayed Jobs and replay", Points: 25}
	if err := runHostNativeCLI(config); err != nil {
		check.Details = []string{err.Error()}
	} else {
		check.Passed, check.Earned = true, check.Points
		check.Details = []string{"Host-screen grant loaded; display-only schema and primary inventory verified; default and explicit zero produced exact-dimension PNGs with matching predetermined synthetic band interiors through distinct delayed Jobs; both payloads replayed on a new human turn without recapture; candidate joined.",
			"This scope uses one agreed primary display. Other-display selection, unsupported platforms, pre-effect capability checks and capture-process cancellation still require their focused tests/review; a wire reply alone cannot prove absence of an OS side effect."}
	}
	return grade.NewTitledReport("Chapter 26 — host-screen scope only (25 points)", []grade.Check{check}, ""), nil
}
