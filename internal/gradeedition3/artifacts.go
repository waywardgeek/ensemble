package gradeedition3

import (
	"context"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"
)

//go:embed testdata/artifacts-consumer/*.cs testdata/artifacts-consumer/*.csproj testdata/artifacts-browser.cjs testdata/artifacts-actions.cjs
var artifactGradeFiles embed.FS

type ArtifactGradeConfig struct{ RuntimeDirectory, Dotnet, Node, PlaywrightModule string }
type ArtifactRow struct {
	ID      string `json:"id"`
	Points  int    `json:"points"`
	Earned  int    `json:"earned"`
	Status  string `json:"status"`
	Details string `json:"details"`
}
type ArtifactProbe struct {
	ID     string `json:"id"`
	Passed bool   `json:"passed"`
	Detail string `json:"detail"`
}
type ArtifactReport struct {
	Scope    string          `json:"scope"`
	Status   string          `json:"status"`
	Score    int             `json:"score"`
	MaxScore int             `json:"max_score"`
	Rows     []ArtifactRow   `json:"rows"`
	Probes   []ArtifactProbe `json:"probes"`
}

func newArtifactReport() ArtifactReport {
	r := ArtifactReport{Scope: "artifacts", Status: "ungraded", MaxScore: 100}
	for _, row := range []struct {
		id     string
		points int
	}{{"actions", 25}, {"snippets", 20}, {"recorded-time", 15}, {"work-products", 30}, {"interaction", 10}} {
		r.Rows = append(r.Rows, ArtifactRow{ID: row.id, Points: row.points, Status: "ungraded", Details: "Full independent row and sensitivity checks are not implemented; partial probes award no points."})
	}
	return r
}

// Only reviewed, independently exercised rows earn points. The shared Actions
// probe must finish all browser assertions and joined host checks: an early
// snippet or interaction success followed by a failed probe is insufficient evidence.
func (r *ArtifactReport) applyReviewedRows() {
	passed := make(map[string]bool)
	for _, probe := range r.Probes {
		passed[probe.ID] = probe.Passed
	}
	complete := passed["actions-browser"]
	// Failure presentation is exercised in both download viewer modes, not in
	// the Actions scenario. All three complete, joined probes are required.
	actionsComplete := complete && passed["work-product-browser-unnamed=false"] &&
		passed["work-product-browser-unnamed=true"]
	r.Score = 0
	for i := range r.Rows {
		row := &r.Rows[i]
		if row.ID == "actions" {
			row.Earned, row.Status = 0, "ungraded"
			row.Details = "Reviewed Actions contract; complete Actions and both failure-browser/host receipts required."
			if actionsComplete {
				row.Earned, row.Status = row.Points, "passed"
				row.Details = "Real successful, failed and delayed calls; exact pairing, factual states, raw detail, selected-Agent and Job-owner boundaries, and page-reconnect consistency."
			}
		}
		if row.ID == "snippets" || row.ID == "interaction" {
			row.Earned, row.Status = 0, "ungraded"
			row.Details = "Reviewed " + row.ID + " contract; complete Actions browser/host evidence unavailable."
			if complete {
				row.Earned, row.Status = row.Points, "passed"
				if row.ID == "snippets" {
					row.Details = "Real buffered/streaming payloads, exact copy, clipboard fallback, retained focus/selection and stale completion; clipboard boundary substituted."
				} else {
					row.Details = "Keyboard/touch targets, narrow-layout reading stability, textual states, draft preservation and speech noninterference; synthetic speech boundary, human usability separate."
				}
			}
		}
		r.Score += row.Earned
	}
	// Other rows remain unfinished. Never infer full acceptance from probe count.
}

func (r ArtifactReport) ExitCode() int {
	if r.Status == "passed" {
		return 0
	}
	if r.Status == "failed" {
		return 1
	}
	return 3
}
func (r ArtifactReport) WriteJSON(w io.Writer) error {
	e := json.NewEncoder(w)
	e.SetIndent("", "  ")
	return e.Encode(r)
}
func (r ArtifactReport) WriteText(w io.Writer) {
	fmt.Fprintln(w, "Artifacts: independent public contract (in development)")
	for _, p := range r.Probes {
		status := "FAIL"
		if p.Passed {
			status = "PASS"
		}
		fmt.Fprintf(w, "[%s] probe %s: %s\n", status, p.ID, p.Detail)
	}
	for _, row := range r.Rows {
		fmt.Fprintf(w, "[%s] %s (%d/%d): %s\n", strings.ToUpper(row.Status), row.ID, row.Earned, row.Points, row.Details)
	}
	fmt.Fprintf(w, "score: %d/%d; %s (exit 3 means unfinished coverage, not candidate acceptance)\n", r.Score, r.MaxScore, strings.ToUpper(r.Status))
}
func ValidateArtifactGradeConfig(c ArtifactGradeConfig) error {
	if runtime.GOOS != "darwin" && runtime.GOOS != "linux" {
		return fmt.Errorf("artifact runner currently requires macOS or Linux")
	}
	for _, p := range []string{c.RuntimeDirectory, c.Dotnet, c.Node, c.PlaywrightModule} {
		if !filepath.IsAbs(p) {
			return fmt.Errorf("artifact runtime and tool paths must be absolute")
		}
		if _, err := os.Stat(p); err != nil {
			return err
		}
	}
	for _, p := range []string{c.Dotnet, c.Node} {
		info, err := os.Stat(p)
		if err != nil || !info.Mode().IsRegular() || info.Mode().Perm()&0111 == 0 {
			return fmt.Errorf("artifact tool must be executable: %s", p)
		}
	}
	for _, name := range []string{"Ensemble.dll", "Ensemble.Gui.dll"} {
		info, err := os.Stat(filepath.Join(c.RuntimeDirectory, name))
		if err != nil || !info.Mode().IsRegular() {
			return fmt.Errorf("artifact candidate must contain %s", name)
		}
	}
	return nil
}
func artifactEnvironment(root string) []string {
	return []string{"PATH=" + os.Getenv("PATH"), "HOME=" + root, "TMPDIR=" + root, "DOTNET_CLI_HOME=" + root, "DOTNET_SKIP_FIRST_TIME_EXPERIENCE=1", "DOTNET_CLI_TELEMETRY_OPTOUT=1", "DOTNET_GENERATE_ASPNET_CERTIFICATE=false"}
}

// Each process group belongs to this grader operation. Cleanup never targets a
// discovered PID or an installed application; descendants cannot outlive a run.
func artifactCommand(ctx context.Context, dir string, env []string, exe string, args ...string) (*exec.Cmd, *probeOutput) {
	cmd := exec.CommandContext(ctx, exe, args...)
	cmd.Dir, cmd.Env = dir, env
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	cmd.Cancel = func() error { return syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL) }
	out := &probeOutput{}
	cmd.Stdout, cmd.Stderr = out, out
	return cmd, out
}
func runArtifactCommand(ctx context.Context, dir string, env []string, exe string, args ...string) (string, error) {
	cmd, out := artifactCommand(ctx, dir, env, exe, args...)
	err := cmd.Run()
	if cmd.Process != nil {
		_ = syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
	}
	if out.overflow {
		return out.text.String(), fmt.Errorf("grader diagnostic output exceeds bound")
	}
	return out.text.String(), err
}
func artifactBrowserPreflight(c ArtifactGradeConfig, root string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()
	env := append(artifactEnvironment(root), "PLAYWRIGHT_MODULE="+c.PlaywrightModule)
	output, err := runArtifactCommand(ctx, root, env, c.Node, "-e", `
(async () => {
 const {chromium}=require(process.env.PLAYWRIGHT_MODULE);
 const browser=await chromium.launch({headless:true,channel:'chrome',timeout:15000});
 try {const page=await browser.newPage();await page.setContent('<p>artifact grader setup</p>');}
 finally {await browser.close();}
})().catch(error=>{console.error(error.message);process.exitCode=1;});`)
	if err != nil {
		return fmt.Errorf("browser setup unavailable (not a candidate grade): %w; %s", err, output)
	}
	return nil
}

func GradeArtifacts(c ArtifactGradeConfig) (report ArtifactReport, result error) {
	report = newArtifactReport()
	if err := ValidateArtifactGradeConfig(c); err != nil {
		return report, err
	}
	root, err := os.MkdirTemp("", "ensemble-artifact-grade-")
	if err != nil {
		return report, err
	}
	defer func() { result = errors.Join(result, os.RemoveAll(root)) }()
	env := artifactEnvironment(root)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	sdk, err := runArtifactCommand(ctx, root, env, c.Dotnet, "--list-sdks")
	if err != nil {
		return report, fmt.Errorf("SDK preflight: %w", err)
	}
	found := false
	for _, line := range strings.Split(sdk, "\n") {
		if strings.HasPrefix(line, "10.") {
			found = true
		}
	}
	if !found {
		return report, fmt.Errorf("artifact setup requires an installed .NET 10 SDK")
	}
	if err := artifactBrowserPreflight(c, root); err != nil {
		return report, err
	}
	for _, name := range []string{"Program.cs", "ArtifactHost.cs", "ActionsHost.cs", "TimingChecks.cs", "TransferChecks.cs", "ResponseBarrier.cs", "ArtifactsConsumer.csproj"} {
		data, err := artifactGradeFiles.ReadFile("testdata/artifacts-consumer/" + name)
		if err != nil {
			return report, err
		}
		if err = os.WriteFile(filepath.Join(root, name), data, 0600); err != nil {
			return report, err
		}
	}
	data, err := artifactGradeFiles.ReadFile("testdata/artifacts-browser.cjs")
	if err != nil {
		return report, err
	}
	driver := filepath.Join(root, "browser.cjs")
	if err = os.WriteFile(driver, data, 0600); err != nil {
		return report, err
	}
	nuget := filepath.Join(root, "NuGet.Config")
	if err = os.WriteFile(nuget, []byte("<configuration><packageSources><clear /></packageSources></configuration>"), 0600); err != nil {
		return report, err
	}
	output, err := runArtifactCommand(ctx, root, env, c.Dotnet, "build", filepath.Join(root, "ArtifactsConsumer.csproj"), "-p:CandidateDirectory="+c.RuntimeDirectory, "--configfile", nuget, "--artifacts-path", filepath.Join(root, "artifacts"), "--tl:off", "-v:q")
	if err != nil {
		report.Status = "failed"
		report.Probes = append(report.Probes, ArtifactProbe{"public-consumer-build", false, err.Error() + ": " + output})
		return report, nil
	}
	binary := filepath.Join(root, "artifacts", "bin", "ArtifactsConsumer", "debug", "ArtifactsConsumer.dll")
	core, err := os.MkdirTemp(root, "core-")
	if err != nil {
		return report, err
	}
	coreCtx, coreCancel := context.WithTimeout(context.Background(), 70*time.Second)
	output, err = runArtifactCommand(coreCtx, core, artifactEnvironment(core), c.Dotnet, binary, "work-product-core")
	coreCancel()
	if err == nil {
		err = checkArtifactCoreReceipt([]byte(output))
	}
	p := ArtifactProbe{ID: "work-product-core", Passed: err == nil, Detail: "Independent public APIs and randomized grader-owned files; no points yet."}
	if err != nil {
		p.Detail = err.Error() + ": " + output
		report.Status = "failed"
	}
	report.Probes = append(report.Probes, p)
	for _, unnamed := range []bool{false, true} {
		err := runArtifactBrowser(c, root, binary, driver, unnamed, false)
		p := ArtifactProbe{ID: fmt.Sprintf("work-product-browser-unnamed=%t", unnamed), Passed: err == nil, Detail: "Real model/tool sequence, failure presentation, download guards and joined host; contributes to the reviewed Actions row, work-product row incomplete."}
		if err != nil {
			p.Detail = err.Error()
			report.Status = "failed"
		}
		report.Probes = append(report.Probes, p)
	}
	actionsData, err := artifactGradeFiles.ReadFile("testdata/artifacts-actions.cjs")
	if err != nil {
		return report, err
	}
	actionsDriver := filepath.Join(root, "actions.cjs")
	if err = os.WriteFile(actionsDriver, actionsData, 0600); err != nil {
		return report, err
	}
	err = runArtifactBrowser(c, root, binary, actionsDriver, false, true)
	actionsProbe := ArtifactProbe{ID: "actions-browser", Passed: err == nil, Detail: "Real Actions, snippet and interaction browser checks with joined host; Actions credit also requires both failure-browser probes."}
	if err != nil {
		actionsProbe.Detail = err.Error()
		report.Status = "failed"
	}
	report.Probes = append(report.Probes, actionsProbe)
	timing, err := os.MkdirTemp(root, "timing-")
	if err != nil {
		return report, err
	}
	timingCtx, timingCancel := context.WithTimeout(context.Background(), 30*time.Second)
	output, err = runArtifactCommand(timingCtx, timing, artifactEnvironment(timing), c.Dotnet, binary, "timing-core")
	timingCancel()
	if err == nil {
		err = checkTimingCoreReceipt([]byte(output))
	}
	timingProbe := ArtifactProbe{ID: "timing-core", Passed: err == nil, Detail: "Public snapshot load and projection of known, zero and unknown timing; no points yet."}
	if err != nil {
		timingProbe.Detail = err.Error() + ": " + output
		report.Status = "failed"
	}
	report.Probes = append(report.Probes, timingProbe)
	report.applyReviewedRows()
	return report, nil
}
func checkArtifactCoreReceipt(data []byte) error {
	var r struct {
		Contract string   `json:"contract"`
		Passed   bool     `json:"passed"`
		Checks   []string `json:"checks"`
	}
	if err := json.Unmarshal(data, &r); err != nil {
		return err
	}
	if r.Contract != "artifact-work-product-core-v1" || !r.Passed {
		return fmt.Errorf("core contract did not pass")
	}
	for _, want := range []string{"ungranted-offer-refused", "explicit-grant-loaded", "successful-result-has-typed-offer", "offer-description-matches-file", "exact-click-time-bytes", "download-destination-remains-borrowed", "foreign-owner-refused", "pathname-is-not-offer-id", "event-roundtrip-retains-description", "snapshot-only-gui-retains-description", "grown-file-over-limit-refused", "grant-unloaded", "revoked-new-offer-refused", "same-name-replacement-does-not-revive-offer", "exact-64mib-copy-hash", "missing-offered-file-refused", "outside-root-offer-refused", "replacement-symlink-escape-refused", "failed-offer-refused-before-cap", "offer-cap-admits-64-unique-grants", "offer-cap-refuses-65th", "offer-cap-preserves-earliest-grant", "failed-result-cannot-offer", "failed-event-cannot-offer", "blocked-copy-does-not-block-another", "caller-cancel-joins-copy-cleanup", "caller-cancel-preserves-borrowed-sink", "agent-destruction-joins-transfer-cleanup", "agent-destruction-cancels-without-owning-sink", "snapshot-file-restored", "restored-history-retains-description", "restored-description-cannot-grant-read"} {
		found := false
		for _, got := range r.Checks {
			if got == want {
				found = true
			}
		}
		if !found {
			return fmt.Errorf("missing core receipt: %s", want)
		}
	}
	return nil
}
func checkArtifactHostReceipt(data []byte) error {
	var r map[string]json.RawMessage
	if err := json.Unmarshal(data, &r); err != nil {
		return err
	}
	for key, want := range map[string]int{"requests": 4, "returned": 3, "offers": 1, "failedOffers": 1} {
		var got int
		if err := json.Unmarshal(r[key], &got); err != nil || got != want {
			return fmt.Errorf("joined host %s count mismatch", key)
		}
	}
	if string(r["failure"]) != "null" {
		return fmt.Errorf("host oracle failure or missing terminal field")
	}
	return nil
}
func runArtifactBrowser(c ArtifactGradeConfig, root, binary, driver string, unnamed, actions bool) (result error) {
	work, err := os.MkdirTemp(root, "browser-")
	if err != nil {
		return err
	}
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return err
	}
	port := listener.Addr().(*net.TCPAddr).Port
	if err = listener.Close(); err != nil {
		return err
	}
	base := "http://127.0.0.1:" + strconv.Itoa(port) + "/assistant/"
	env := append(artifactEnvironment(work), "ARTIFACT_GRADE_PORT="+strconv.Itoa(port), "ARTIFACT_GUI_URL="+base, "PLAYWRIGHT_MODULE="+c.PlaywrightModule)
	if unnamed {
		env = append(env, "ARTIFACT_UNNAMED_VIEWER=1")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 110*time.Second)
	defer cancel()
	mode := "browser-host"
	if actions {
		mode = "actions-host"
	}
	host, out := artifactCommand(ctx, work, env, c.Dotnet, binary, mode)
	if err = host.Start(); err != nil {
		return err
	}
	exited := make(chan error, 1)
	go func() { exited <- host.Wait() }()
	collected := false
	defer func() {
		if !collected {
			result = errors.Join(result, os.WriteFile(filepath.Join(work, "stop"), []byte("stop\n"), 0600))
			select {
			case err := <-exited:
				result = errors.Join(result, err)
			case <-time.After(20 * time.Second):
				_ = syscall.Kill(-host.Process.Pid, syscall.SIGKILL)
				<-exited
				result = errors.Join(result, fmt.Errorf("host failed joined shutdown"))
			}
		}
		_ = syscall.Kill(-host.Process.Pid, syscall.SIGKILL)
		if out.overflow {
			result = errors.Join(result, fmt.Errorf("host output exceeded bound"))
		}
		if result != nil {
			result = fmt.Errorf("%w; host: %s", result, out.text.String())
			return
		}
		path := filepath.Join(work, "result.json")
		info, err := os.Stat(path)
		if err != nil {
			result = err
			return
		}
		if info.Size() > 64<<10 {
			result = fmt.Errorf("host receipt exceeds bound")
			return
		}
		data, err := os.ReadFile(path)
		if err != nil {
			result = err
			return
		}
		if actions {
			result = checkActionsHostReceipt(data)
		} else {
			result = checkArtifactHostReceipt(data)
		}
	}()
	client := &http.Client{Timeout: time.Second}
	for {
		select {
		case err := <-exited:
			collected = true
			return fmt.Errorf("host exited before readiness: %v", err)
		case <-ctx.Done():
			return fmt.Errorf("host readiness timed out")
		default:
		}
		response, err := client.Get(base)
		if err == nil {
			response.Body.Close()
			if response.StatusCode == http.StatusOK {
				break
			}
		}
		time.Sleep(50 * time.Millisecond)
	}
	output, err := runArtifactCommand(ctx, work, env, c.Node, driver)
	if err != nil {
		return fmt.Errorf("browser failed: %w; %s", err, output)
	}
	if actions {
		return checkActionsBrowserReceipt([]byte(output))
	}
	return checkArtifactBrowserReceipt([]byte(output))
}

var artifactBrowserChecks = []string{
	"real-model-tool-sequence", "click-only", "exact-current-bytes", "filename",
	"origin-header", "no-path-query", "foreign-owner", "exact-64mib-browser-hash",
	"oversize-http-refused", "oversize-browser-refused", "replacement-expiry", "replay-no-effects", "cancel-joins-http", "selection-replay-cancels-http", "failed-offer-no-download", "failed-offer-label-payload-and-pair-replay",
}

func checkArtifactBrowserReceipt(data []byte) error {
	var receipt struct {
		Contract string   `json:"contract"`
		Passed   bool     `json:"passed"`
		Checks   []string `json:"checks"`
	}
	if err := json.Unmarshal(data, &receipt); err != nil {
		return err
	}
	if receipt.Contract != "artifact-browser-work-product-v1" || !receipt.Passed {
		return fmt.Errorf("browser receipt did not pass")
	}
	for _, want := range artifactBrowserChecks {
		found := false
		for _, got := range receipt.Checks {
			if got == want {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("missing browser receipt: %s", want)
		}
	}
	return nil
}

func checkActionsHostReceipt(data []byte) error {
	var r map[string]json.RawMessage
	if err := json.Unmarshal(data, &r); err != nil {
		return err
	}
	for key, want := range map[string]int{"requests": 3, "executions": 1, "returned": 2, "running": 1, "done": 1, "streamRequests": 1, "streamCompleted": 1, "stateRequests": 2, "stateExecutions": 2} {
		var got int
		if err := json.Unmarshal(r[key], &got); err != nil || got != want {
			return fmt.Errorf("Actions host %s count mismatch", key)
		}
	}
	if string(r["failure"]) != "null" || string(r["streamFailure"]) != "null" || string(r["stateFailure"]) != "null" {
		return fmt.Errorf("Actions host oracle failure or missing field")
	}
	return nil
}

var actionsBrowserChecks = []string{"real-running-observation", "call-result-navigation", "historical-running-preserved", "terminal-pathname-not-grant", "replay-no-execution", "replay-exact-pair-and-agent-isolation", "focused-expanded-reader-preserved", "canonical-recorded-time-and-duration", "timing-survives-replay", "legacy-unknown-and-zero-browser", "provider-snippets-exact-and-fallback", "real-streaming-snippet-stability", "delayed-clipboard-replay-fenced", "raw-selection-and-scroll-preserved", "job-output-owner-and-running-guards", "requested-and-dispatched-batch-states", "successful-batch-exact-pairs", "keyboard-touch-and-draft-preserved", "speech-copy-and-pair-noninterference", "streamed-narration-once-and-replay-silent"}

func checkActionsBrowserReceipt(data []byte) error {
	var r struct {
		Contract string
		Passed   bool
		Checks   []string
	}
	if err := json.Unmarshal(data, &r); err != nil {
		return err
	}
	if r.Contract != "artifact-actions-v1" || !r.Passed {
		return fmt.Errorf("Actions browser receipt did not pass")
	}
	for _, want := range actionsBrowserChecks {
		found := false
		for _, got := range r.Checks {
			if got == want {
				found = true
			}
		}
		if !found {
			return fmt.Errorf("missing Actions receipt: %s", want)
		}
	}
	return nil
}

var timingCoreChecks = []string{"timing-event-codec-preserves-exact-values", "timing-snapshot-actually-loaded", "timing-snapshot-retains-three-results", "timing-known-retained-after-load", "timing-zero-retained-after-load", "timing-unknown-remains-unknown-after-load", "timing-snapshot-only-projects-three-results", "timing-known-snapshot-only-projection", "timing-zero-snapshot-only-projection", "timing-unknown-snapshot-only-projection"}

func checkTimingCoreReceipt(data []byte) error {
	var r struct {
		Contract string
		Passed   bool
		Checks   []string
	}
	if err := json.Unmarshal(data, &r); err != nil {
		return err
	}
	if r.Contract != "artifact-timing-core-v1" || !r.Passed {
		return fmt.Errorf("timing core receipt did not pass")
	}
	for _, want := range timingCoreChecks {
		found := false
		for _, got := range r.Checks {
			if got == want {
				found = true
			}
		}
		if !found {
			return fmt.Errorf("missing timing receipt: %s", want)
		}
	}
	return nil
}
