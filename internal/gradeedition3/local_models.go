package gradeedition3

import (
	"bytes"
	"context"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"time"
)

//go:embed testdata/local-models-consumer/*.cs testdata/local-models-consumer/*.csproj testdata/local-models-consumer/NuGet.Config
var localModelFiles embed.FS

type LocalModelGradeConfig struct{ RuntimeDirectory, Dotnet string }
type LocalModelReport struct {
	Scope       string `json:"scope"`
	Status      string `json:"status"`
	ProbePassed bool   `json:"probe_passed"`
	Detail      string `json:"detail"`
}

func (r LocalModelReport) ExitCode() int {
	if r.Status == "failed" {
		return 1
	}
	return 3 // Remaining chapter coverage is deliberately not certified.
}
func (r LocalModelReport) WriteJSON(w io.Writer) error { return json.NewEncoder(w).Encode(r) }
func (r LocalModelReport) WriteText(w io.Writer) {
	fmt.Fprintf(w, "Local Models: %s (no chapter points)\nPublic conversations: %t\n%s\n", r.Status, r.ProbePassed, r.Detail)
}

func validateLocalModelConfig(c LocalModelGradeConfig) error {
	if runtime.GOOS != "darwin" && runtime.GOOS != "linux" {
		return fmt.Errorf("local-models runner requires macOS or Linux")
	}
	if !filepath.IsAbs(c.RuntimeDirectory) || !filepath.IsAbs(c.Dotnet) {
		return fmt.Errorf("runtime and SDK paths must be absolute")
	}
	info, err := os.Stat(c.Dotnet)
	if err != nil || !info.Mode().IsRegular() || info.Mode().Perm()&0111 == 0 {
		return fmt.Errorf("SDK must be an existing executable")
	}
	info, err = os.Stat(filepath.Join(c.RuntimeDirectory, "Ensemble.dll"))
	if err != nil || !info.Mode().IsRegular() {
		return fmt.Errorf("candidate directory must contain Ensemble.dll")
	}
	return nil
}
func localModelEnvironment(root string) []string {
	return []string{"PATH=/usr/bin:/bin", "HOME=" + root, "TMPDIR=" + root, "DOTNET_CLI_HOME=" + root,
		"DOTNET_SKIP_FIRST_TIME_EXPERIENCE=1", "DOTNET_CLI_TELEMETRY_OPTOUT=1", "DOTNET_GENERATE_ASPNET_CERTIFICATE=false"}
}

// Separate streams keep compiler diagnostics out of the machine receipt. WaitDelay
// bounds inherited output pipes even if a descendant outlives its direct parent.
// This owns a process group, not a sandbox against deliberately escaped processes.
func runLocalModelCommand(ctx context.Context, dir string, env []string, exe string, args ...string) (string, string, error) {
	cmd := exec.CommandContext(ctx, exe, args...)
	cmd.Dir, cmd.Env = dir, env
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	cmd.Cancel = func() error { return syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL) }
	cmd.WaitDelay = time.Second
	stdout, stderr := &probeOutput{}, &probeOutput{}
	cmd.Stdout, cmd.Stderr = stdout, stderr
	err := cmd.Run()
	if cmd.Process != nil {
		killErr := syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
		if killErr != nil && !errors.Is(killErr, syscall.ESRCH) {
			err = errors.Join(err, killErr)
		}
	}
	if ctx.Err() != nil {
		err = errors.Join(err, ctx.Err())
	}
	if stdout.overflow || stderr.overflow {
		err = errors.Join(err, fmt.Errorf("diagnostic output exceeds 64 KiB per stream"))
	}
	return stdout.text.String(), stderr.text.String(), err
}

func GradeLocalModels(c LocalModelGradeConfig) (report LocalModelReport, result error) {
	report = LocalModelReport{Scope: "local-models", Status: "ungraded"}
	if err := validateLocalModelConfig(c); err != nil {
		return report, err
	}
	root, err := os.MkdirTemp("", "ensemble-local-model-grade-")
	if err != nil {
		return report, err
	}
	defer func() { result = errors.Join(result, os.RemoveAll(root)) }()
	env := localModelEnvironment(root)
	setup, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	sdk, diagnostic, err := runLocalModelCommand(setup, root, env, c.Dotnet, "--list-sdks")
	if err != nil {
		return report, fmt.Errorf("SDK preflight: %w; %s", err, diagnostic)
	}
	found := false
	for _, line := range strings.Split(sdk, "\n") {
		if strings.HasPrefix(line, "10.") {
			found = true
		}
	}
	if !found {
		return report, fmt.Errorf("an installed .NET 10 SDK is required")
	}
	for _, name := range []string{"Program.cs", "ConfigurationChecks.cs", "LifecycleChecks.cs", "LocalModelsConsumer.csproj", "NuGet.Config"} {
		data, err := localModelFiles.ReadFile("testdata/local-models-consumer/" + name)
		if err != nil {
			return report, err
		}
		if err := os.WriteFile(filepath.Join(root, name), data, 0600); err != nil {
			return report, err
		}
	}
	output, diagnostic, err := runLocalModelCommand(setup, root, env, c.Dotnet, "build", filepath.Join(root, "LocalModelsConsumer.csproj"),
		"-p:CandidateDirectory="+c.RuntimeDirectory, "--configfile", filepath.Join(root, "NuGet.Config"),
		"--artifacts-path", filepath.Join(root, "artifacts"), "-p:NuGetAudit=false", "--tl:off", "-v:q")
	if err != nil {
		return report, fmt.Errorf("public consumer build (setup/API compatibility): %w; %s%s", err, output, diagnostic)
	}
	work := filepath.Join(root, "work")
	if err := os.Mkdir(work, 0700); err != nil {
		return report, err
	}
	execution, stop := context.WithTimeout(context.Background(), 75*time.Second)
	defer stop()
	binary := filepath.Join(root, "artifacts", "bin", "LocalModelsConsumer", "debug", "LocalModelsConsumer.dll")
	output, diagnostic, err = runLocalModelCommand(execution, work, localModelEnvironment(work), c.Dotnet, binary)
	if err == nil {
		err = checkLocalModelReceipt([]byte(output))
	}
	if err != nil {
		report.Status, report.Detail = "failed", fmt.Sprintf("%v; stdout: %s; stderr: %s", err, output, diagnostic)
	} else {
		report.ProbePassed = true
		report.Detail = fmt.Sprintf("%d named public conversation/configuration/restoration/lifecycle/accounting checks passed. Broader lifecycle/refusal coverage and real-model usefulness remain incomplete.", len(localModelCheckNames()))
	}
	return report, nil
}

func localModelCheckNames() []string {
	var names []string
	for _, mode := range []string{"buffered", "streamed", "prose"} {
		for _, suffix := range []string{"selection-membership", "serving-not-advertised-context", "final-answer", "exact-effects-and-requests", "qualified-response-provenance", "cached-input-counted-once", "known-usage-unknown-cost"} {
			names = append(names, mode+"-"+suffix)
		}
		count := 2
		if mode == "prose" {
			count = 1
		}
		for index := 0; index < count; index++ {
			for _, suffix := range []string{"route", "no-borrowed-credentials", "wire-name", "declared-stream-policy"} {
				names = append(names, fmt.Sprintf("%s-%s-%d", mode, suffix, index))
			}
		}
		if mode != "prose" {
			names = append(names, mode+"-real-correlated-result", mode+"-result-recorded-before-next-request")
		}
	}
	for _, suffix := range []string{"selection-membership", "serving-not-advertised-context", "final-answer", "exact-effects-and-requests", "qualified-response-provenance", "input-remains-unknown", "meter-remains-unknown", "route-0", "no-borrowed-credentials-0", "wire-name-0", "declared-stream-policy-0"} {
		names = append(names, "missing-usage-"+suffix)
	}
	for _, mode := range []string{"interrupt", "provider-error"} {
		for _, suffix := range []string{"selection-membership", "serving-not-advertised-context", "final-answer", "exact-effects-and-requests", "qualified-response-provenance", "cached-input-counted-once", "known-usage-unknown-cost", "partial-has-no-execution-authority", "controlled-failure", "failed-fragments-have-no-effects", "fresh-request-has-no-orphan-tools"} {
			names = append(names, mode+"-"+suffix)
		}
		for index := 0; index < 2; index++ {
			for _, suffix := range []string{"route", "no-borrowed-credentials", "wire-name", "declared-stream-policy"} {
				names = append(names, fmt.Sprintf("%s-%s-%d", mode, suffix, index))
			}
		}
	}
	return append(names, "streamed-partial-has-no-execution-authority",
		"configuration-load-preserves-launch", "configuration-real-binding-and-capacity",
		"configuration-late-refusal-rolls-back", "child-inherits-real-offering",
		"child-registration-identity-not-secret", "child-restores-exact-offering",
		"child-restores-conversation-without-request", "model-loss-refuses-without-agent-destruction",
		"child-missing-route-no-wire-name-fallback", "child-refusal-preserves-registration-no-request",
		"lifecycle-real-configured-request", "lifecycle-real-tool-started",
		"lifecycle-cancellation-reaches-executing-tool", "lifecycle-destruction-joins-held-cleanup",
		"lifecycle-worker-exited-before-destruction-return", "lifecycle-detaches-without-destroying-sibling")
}

func checkLocalModelReceipt(data []byte) error {
	// Decode fields individually so duplicate keys cannot overwrite an earlier claim.
	decoder := json.NewDecoder(bytes.NewReader(data))
	token, err := decoder.Token()
	if err != nil || token != json.Delim('{') {
		return fmt.Errorf("receipt must be one JSON object")
	}
	fields := map[string]json.RawMessage{}
	for decoder.More() {
		token, err = decoder.Token()
		if err != nil {
			return err
		}
		name, ok := token.(string)
		if !ok {
			return fmt.Errorf("invalid receipt field")
		}
		if _, exists := fields[name]; exists {
			return fmt.Errorf("duplicate receipt field %s", name)
		}
		var value json.RawMessage
		if err = decoder.Decode(&value); err != nil {
			return err
		}
		fields[name] = value
	}
	if _, err = decoder.Token(); err != nil {
		return err
	}
	var trailing any
	if err = decoder.Decode(&trailing); err != io.EOF {
		return fmt.Errorf("trailing receipt data")
	}
	if len(fields) != 3 || string(fields["passed"]) != "true" || string(fields["failure"]) != "null" {
		return fmt.Errorf("receipt requires passed:true, checks, failure:null")
	}
	var checks []string
	if err = json.Unmarshal(fields["checks"], &checks); err != nil {
		return err
	}
	expected := localModelCheckNames()
	if len(checks) != len(expected) {
		return fmt.Errorf("receipt requires %d named checks", len(expected))
	}
	seen := map[string]bool{}
	for _, name := range checks {
		if seen[name] {
			return fmt.Errorf("duplicate check %s", name)
		}
		seen[name] = true
	}
	for _, name := range expected {
		if !seen[name] {
			return fmt.Errorf("missing check %s", name)
		}
	}
	return nil
}
