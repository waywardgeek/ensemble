package gradeedition3

import (
	"context"
	"crypto/rand"
	"embed"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/waywardgeek/ensemble/internal/grade"
)

//go:embed testdata/voice-consumer/*.cs testdata/voice-consumer/*.csproj testdata/ch26-voice.cjs
var voiceGradeFiles embed.FS

type VoiceGradeConfig struct{ RuntimeDirectory, Dotnet, Node, PlaywrightModule string }

func voiceGradeEnvironment(root string) []string {
	return []string{"PATH=" + os.Getenv("PATH"), "HOME=" + root, "TMPDIR=" + root,
		"DOTNET_CLI_HOME=" + root, "DOTNET_SKIP_FIRST_TIME_EXPERIENCE=1",
		"DOTNET_CLI_TELEMETRY_OPTOUT=1", "DOTNET_GENERATE_ASPNET_CERTIFICATE=false"}
}

// Chapter26Voice builds a grader-owned public consumer, never candidate test
// code. Setup errors return no score; missing candidate APIs fail the scope.
func Chapter26Voice(config VoiceGradeConfig) (report grade.Report, result error) {
	for _, path := range []string{config.RuntimeDirectory, config.Dotnet, config.Node, config.PlaywrightModule} {
		if !filepath.IsAbs(path) {
			return report, fmt.Errorf("voice setup requires absolute runtime/tool paths")
		}
		if _, err := os.Stat(path); err != nil {
			return report, fmt.Errorf("voice setup: %w", err)
		}
	}
	for _, name := range []string{"Ensemble.dll", "Ensemble.Gui.dll"} {
		if info, err := os.Stat(filepath.Join(config.RuntimeDirectory, name)); err != nil || !info.Mode().IsRegular() {
			return report, fmt.Errorf("voice candidate runtime must contain %s", name)
		}
	}
	root, err := os.MkdirTemp("", "ensemble-voice-grade-")
	if err != nil {
		return report, err
	}
	defer func() { result = errors.Join(result, os.RemoveAll(root)) }()
	if err := checkMediaBrowserSetup(config.Node, config.PlaywrightModule, root); err != nil {
		return report, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	sdk := exec.CommandContext(ctx, config.Dotnet, "--list-sdks")
	sdk.Dir, sdk.Env = root, voiceGradeEnvironment(root)
	output := &probeOutput{}
	sdk.Stdout, sdk.Stderr = output, output
	if err := sdk.Run(); err != nil {
		return report, fmt.Errorf(".NET SDK setup: %w", err)
	}
	sdkFound := false
	for _, line := range strings.Split(output.text.String(), "\n") {
		if strings.HasPrefix(line, "10.") {
			sdkFound = true
		}
	}
	if !sdkFound || output.overflow {
		return report, fmt.Errorf("voice setup requires an installed .NET 10 SDK")
	}
	for _, name := range []string{"Program.cs", "ExternalVoiceHost.cs", "VoiceConsumer.csproj"} {
		data, err := voiceGradeFiles.ReadFile("testdata/voice-consumer/" + name)
		if err != nil {
			return report, err
		}
		if err := os.WriteFile(filepath.Join(root, name), data, 0600); err != nil {
			return report, err
		}
	}
	browserSource, err := voiceGradeFiles.ReadFile("testdata/ch26-voice.cjs")
	if err != nil {
		return report, err
	}
	driver := filepath.Join(root, "browser.cjs")
	if err := os.WriteFile(driver, browserSource, 0600); err != nil {
		return report, err
	}
	nuget := filepath.Join(root, "NuGet.Config")
	if err := os.WriteFile(nuget, []byte("<configuration><packageSources><clear /></packageSources></configuration>"), 0600); err != nil {
		return report, err
	}
	artifacts := filepath.Join(root, "artifacts")
	build := exec.CommandContext(ctx, config.Dotnet, "build", filepath.Join(root, "VoiceConsumer.csproj"),
		"-p:CandidateDirectory="+config.RuntimeDirectory, "--configfile", nuget, "--artifacts-path", artifacts, "--tl:off", "-v:q")
	build.Dir, build.Env = root, voiceGradeEnvironment(root)
	buildOutput := &probeOutput{}
	build.Stdout, build.Stderr = buildOutput, buildOutput
	check := grade.Check{ID: "voice-lifecycle", Title: "Public voice capture, funding, admission and joined cancellation", Points: 30}
	finish := func() grade.Report {
		return grade.NewTitledReport("Chapter 26 — voice scope only (30 points)", []grade.Check{check}, "")
	}
	if err := build.Run(); err != nil {
		check.Details = []string{"Public consumer build failed: " + err.Error() + "; " + buildOutput.text.String()}
		return finish(), nil
	}
	if buildOutput.overflow {
		check.Details = []string{"Consumer build exceeded diagnostic bound"}
		return finish(), nil
	}
	binary := filepath.Join(artifacts, "bin", "VoiceConsumer", "debug", "VoiceConsumer.dll")
	for _, mode := range []string{"lifecycle", "funding-off", "permission-delay", "disconnect-recording", "recipient-recording", "recipient-transcription", "disconnect-transcription", "transcription-error"} {
		if err := runVoiceGradeMode(config, root, binary, driver, mode, ""); err != nil {
			check.Details = append(check.Details, mode+": "+err.Error())
			return finish(), nil
		}
		check.Details = append(check.Details, mode+": passed")
	}
	check.Passed, check.Earned = true, check.Points
	check.Details = append(check.Details, "Synthetic browser audio and HTTP transports only; native microphone/provider acceptance and host-screen grade remain separate.")
	return finish(), nil
}

func runVoiceGradeMode(config VoiceGradeConfig, root, binary, driver, mode, mutation string) (result error) {
	work, err := os.MkdirTemp(root, "mode-")
	if err != nil {
		return err
	}
	var nonce [16]byte
	if _, err := rand.Read(nonce[:]); err != nil {
		return err
	}
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return err
	}
	port := listener.Addr().(*net.TCPAddr).Port
	if err := listener.Close(); err != nil {
		return err
	}
	base := "http://127.0.0.1:" + strconv.Itoa(port) + "/assistant/"
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()
	env := append(voiceGradeEnvironment(work), "VOICE_GRADE_PORT="+strconv.Itoa(port),
		"VOICE_GRADE_TRANSCRIPT=VOICE-"+hex.EncodeToString(nonce[:]), "VOICE_GRADE_MODE="+mode,
		"VOICE_GRADE_MUTATION="+mutation, "VOICE_GUI_URL="+base, "PLAYWRIGHT_MODULE="+config.PlaywrightModule)
	host := exec.CommandContext(ctx, config.Dotnet, binary)
	host.Dir, host.Env = work, env
	hostOutput := &probeOutput{}
	host.Stdout, host.Stderr = hostOutput, hostOutput
	if err := host.Start(); err != nil {
		return err
	}
	exited := make(chan error, 1)
	go func() { exited <- host.Wait() }()
	stopped := false
	defer func() {
		if !stopped {
			result = errors.Join(result, os.WriteFile(filepath.Join(work, "stop"), []byte("stop\n"), 0600))
			select {
			case err := <-exited:
				if err != nil {
					result = errors.Join(result, fmt.Errorf("consumer exit: %w; %s", err, hostOutput.text.String()))
				}
			case <-time.After(20 * time.Second):
				cancel()
				<-exited
				result = errors.Join(result, fmt.Errorf("consumer failed joined shutdown"))
			}
		}
		if hostOutput.overflow {
			result = errors.Join(result, fmt.Errorf("consumer diagnostic bound exceeded"))
		}
		if result != nil {
			return
		}
		resultPath := filepath.Join(work, "result.json")
		info, err := os.Stat(resultPath)
		if err != nil {
			result = err
			return
		}
		if info.Size() > 64<<10 {
			result = fmt.Errorf("consumer result exceeds bound")
			return
		}
		data, err := os.ReadFile(resultPath)
		if err != nil {
			result = err
			return
		}
		var fields map[string]json.RawMessage
		if err := json.Unmarshal(data, &fields); err != nil {
			result = err
			return
		}
		audio, admitted := 0, 0
		if mode == "lifecycle" || mode == "transcription-error" {
			audio, admitted = 2, 1
		} else if strings.HasSuffix(mode, "-transcription") {
			audio = 1
		}
		for key, want := range map[string]int{"audioCalls": audio, "audioCompleted": audio, "modelRequests": admitted, "admitted": admitted} {
			var got int
			if err := json.Unmarshal(fields[key], &got); err != nil || got != want {
				result = fmt.Errorf("joined %s count mismatch", key)
				return
			}
		}
		if string(fields["failure"]) != "null" {
			result = fmt.Errorf("consumer reported oracle failure")
		}
	}()
	client := &http.Client{Timeout: time.Second}
	for {
		select {
		case err := <-exited:
			stopped = true
			return fmt.Errorf("consumer exited during startup: %v; %s", err, hostOutput.text.String())
		case <-ctx.Done():
			return fmt.Errorf("voice consumer readiness timed out")
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
	browser := exec.CommandContext(ctx, config.Node, driver)
	browser.Dir, browser.Env = work, env
	browserOutput := &probeOutput{}
	browser.Stdout, browser.Stderr = browserOutput, browserOutput
	// Let the driver close Chrome before escalating a cancelled wait.
	browser.Cancel = func() error { return browser.Process.Signal(os.Interrupt) }
	browser.WaitDelay = 5 * time.Second
	if err := browser.Run(); err != nil {
		return fmt.Errorf("voice browser: %w; %s", err, browserOutput.text.String())
	}
	if browserOutput.overflow {
		return fmt.Errorf("voice browser diagnostic bound exceeded")
	}
	return nil
}
