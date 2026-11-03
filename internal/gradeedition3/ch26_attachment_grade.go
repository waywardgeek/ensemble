package gradeedition3

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/waywardgeek/ensemble/internal/grade"
)

//go:embed testdata/ch26-attachments.cjs
var attachmentBrowserDriver []byte

// Chapter26Attachments distinguishes infrastructure refusal (error) from a
// candidate report. No browser package is installed and no live host is used.
func Chapter26Attachments(config AttachmentProbe) (report grade.Report, result error) {
	for _, path := range []string{config.Candidate, config.Node, config.PlaywrightModule} {
		if !filepath.IsAbs(path) {
			return report, fmt.Errorf("attachment setup requires absolute executable/module paths")
		}
		if _, err := os.Stat(path); err != nil {
			return report, fmt.Errorf("attachment setup: %w", err)
		}
	}
	root, err := os.MkdirTemp("", "ensemble-attachment-grade-")
	if err != nil {
		return report, err
	}
	defer func() {
		if err := os.RemoveAll(root); err != nil {
			result = errors.Join(result, err)
		}
	}()
	config.BrowserDriver = filepath.Join(root, "browser.cjs")
	if err := os.WriteFile(config.BrowserDriver, attachmentBrowserDriver, 0600); err != nil {
		return report, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()
	setup := exec.CommandContext(ctx, config.Node, "-e", `
(async () => {
 const {chromium} = require(process.env.PLAYWRIGHT_MODULE);
 const browser = await chromium.launch({headless:true, channel:'chrome', timeout:15000});
 try { const page = await browser.newPage(); await page.setContent('<p>setup</p>'); }
 finally { await browser.close(); }
})().catch(error => { console.error(error.message); process.exitCode=1; });`)
	setup.Dir = root
	setup.Env = []string{"PATH=" + os.Getenv("PATH"), "HOME=" + root, "TMPDIR=" + root, "PLAYWRIGHT_MODULE=" + config.PlaywrightModule}
	output := &probeOutput{}
	setup.Stdout, setup.Stderr = output, output
	if err := setup.Run(); err != nil {
		return report, fmt.Errorf("browser setup unavailable (not a candidate grade): %w; %s", err, output.text.String())
	}
	if output.overflow {
		return report, fmt.Errorf("browser setup diagnostic bound exceeded")
	}
	check := grade.Check{ID: "attachment-delivery", Title: "Browser attachment pixels, addressed admission and lifecycle", Points: 25}
	scenarios := []struct {
		name, input, fault string
		recipient          bool
	}{
		{"picker", "", "", false},
		{"paste/removal", "paste", "", false},
		{"recipient ownership", "paste", "", true},
		{"late reader", "paste", "late-reader", true},
		{"refusal recovery", "paste", "refusal", false},
		{"uncertain reconnect", "paste", "uncertain", false},
		{"acknowledgement ordering", "paste", "acknowledgements", false},
		{"wire replay", "paste", "duplicate-wire", false},
	}
	config.BrowserMutation = ""
	for _, scenario := range scenarios {
		config.InputMode, config.Fault, config.Recipient = scenario.input, scenario.fault, scenario.recipient
		if err := RunAttachmentPickerProbe(config); err != nil {
			check.Details = append(check.Details, scenario.name+": "+err.Error())
			return grade.NewTitledReport("Chapter 26 — attachments scope only (25 points)", []grade.Check{check}, ""), nil
		}
		check.Details = append(check.Details, scenario.name+": passed")
	}
	check.Passed, check.Earned = true, check.Points
	check.Details = append(check.Details, "Synthetic clipboard only; native phone acceptance is separate. Voice and host-screen scopes are not graded here.")
	return grade.NewTitledReport("Chapter 26 — attachments scope only (25 points)", []grade.Check{check}, ""), nil
}
