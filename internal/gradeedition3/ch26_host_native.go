package gradeedition3

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"image/color"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"

	"github.com/waywardgeek/ensemble/internal/fakevendor"
)

// HostNativePattern describes the human-arranged whole primary display. The
// dimensions are the expected prepared PNG, not a rectangle to capture.
type HostNativePattern struct {
	Width  int         `json:"prepared_width"`
	Height int         `json:"prepared_height"`
	Colors [4][3]uint8 `json:"colors"`
}
type HostNativeConfig struct {
	Candidate        string
	NativeAuthorized bool
	Display          int
	Pattern          HostNativePattern
}

// runHostNativeCLI is never an automatic source-only test. A caller must obtain
// fresh human authorization and cover the entire primary display first. Unit
// tests may exercise this runner only with synthetic, noncapturing candidates.
func runHostNativeCLI(config HostNativeConfig) (result error) {
	if !config.NativeAuthorized {
		return fmt.Errorf("native host-screen grading requires fresh explicit synthetic-display authorization")
	}
	if config.Display != 0 {
		return fmt.Errorf("default/explicit-zero grading requires the agreed synthetic primary display; no fallback capture is permitted")
	}
	if config.Pattern.Width < 32 || config.Pattern.Height < 8 || config.Pattern.Width > 2048 || config.Pattern.Height > 2048 {
		return fmt.Errorf("invalid prepared synthetic-display dimensions")
	}
	for index, rgb := range config.Pattern.Colors {
		for previous := 0; previous < index; previous++ {
			if config.Pattern.Colors[previous] == rgb {
				return fmt.Errorf("synthetic display bands must have distinct colors")
			}
		}
	}
	if !filepath.IsAbs(config.Candidate) {
		return fmt.Errorf("host candidate path must be absolute")
	}
	script := &hostScreenScript{width: config.Pattern.Width, height: config.Pattern.Height}
	for index, rgb := range config.Pattern.Colors {
		script.challenge.colors[index] = color.NRGBA{rgb[0], rgb[1], rgb[2], 255}
	}
	root, err := os.MkdirTemp("", "ensemble-host-native-grade-")
	if err != nil {
		return err
	}
	defer func() { result = errors.Join(result, os.RemoveAll(root)) }()
	skills := filepath.Join(root, "skills")
	for _, directory := range []string{filepath.Join(root, "home"), filepath.Join(skills, "primary"), filepath.Join(skills, "host-screen")} {
		if err := os.MkdirAll(directory, 0700); err != nil {
			return err
		}
	}
	manuals := map[string]string{
		"primary":     "---\nname: primary\ntype: primary\ndescription: Authorized synthetic host grading\ntools: load_skill wait_for_job\nloadable-skills: host-screen\n---\nUse only the explicitly approved fully synthetic primary display.\n",
		"host-screen": "---\nname: host-screen\ntype: loadable\ndescription: Explicit synthetic-display capture authority\ntools: list_displays screenshot\n---\nCapture only the approved whole primary display. No other screen or rectangle is authorized.\n",
	}
	for name, text := range manuals {
		if err := os.WriteFile(filepath.Join(skills, name, "SKILL.md"), []byte(text), 0600); err != nil {
			return err
		}
	}
	var gate sync.Mutex
	var failure error
	exchanges := 0
	provider := fakevendor.NewWithOptions(nil, fakevendor.Options{Route: func(body []byte) *fakevendor.Reply {
		gate.Lock()
		defer gate.Unlock()
		exchanges++
		if failure != nil {
			return &fakevendor.Reply{Text: "HOST_GRADE_REFUSED"}
		}
		if exchanges > 24 {
			failure = fmt.Errorf("host grading exceeded bounded provider exchanges")
			return &fakevendor.Reply{Text: "HOST_GRADE_REFUSED"}
		}
		reply, err := script.step(body)
		if err != nil {
			failure = err
			return &fakevendor.Reply{Text: "HOST_GRADE_REFUSED"}
		}
		return reply
	}})
	defer provider.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	command := exec.CommandContext(ctx, config.Candidate)
	command.Dir = root
	command.Env = append(voiceGradeEnvironment(filepath.Join(root, "home")), "LLM_VENDOR=anthropic",
		"LLM_MODEL=claude-3-haiku-20240307", "LLM_BASE_URL="+provider.URL(), "LLM_API_KEY=synthetic-host-native-key",
		"LLM_REQUIRE_TOOL_EXPLANATION=false", "EN_DISABLE_STREAMING=1", "EN_DISABLE_SAVE=1",
		"EN_SKILLS_DIR="+skills, "EN_PRIMARY_SKILL=primary", "EN_STUB_TOOL_RESULTS=0", "EN_CONTEXT_BUDGET_BYTES=400000")
	input, err := command.StdinPipe()
	if err != nil {
		return err
	}
	output := &mediaOutput{ended: make(chan struct{}, 4)}
	diagnostics := &probeOutput{}
	command.Stdout, command.Stderr = output, diagnostics
	if err := command.Start(); err != nil {
		input.Close()
		return err
	}
	exited := make(chan error, 1)
	done := make(chan struct{})
	go func() { exited <- command.Wait(); close(done) }()
	defer func() {
		input.Close()
		select {
		case err := <-exited:
			if err != nil {
				result = errors.Join(result, fmt.Errorf("native candidate exit: %w; %s", err, diagnostics.text.String()))
			}
		case <-time.After(5 * time.Second):
			cancel()
			<-exited
			result = errors.Join(result, fmt.Errorf("native candidate failed joined exit"))
		}
		if output.overflow || diagnostics.overflow {
			result = errors.Join(result, fmt.Errorf("native candidate diagnostics exceeded bound"))
		}
		gate.Lock()
		defer gate.Unlock()
		result = errors.Join(result, failure)
		if result == nil && (script.stage != 6 || len(script.payloads) != 2) {
			result = fmt.Errorf("native capture/replay did not finish")
		}
	}()
	send := func(text string) error {
		data, _ := json.Marshal(map[string]any{"kind": "prompt", "text": text})
		_, err := fmt.Fprintln(input, string(data))
		return err
	}
	ended := func() error {
		select {
		case <-output.ended:
			return nil
		case <-done:
			return fmt.Errorf("candidate exited before host turn completed")
		case <-ctx.Done():
			return fmt.Errorf("native host turn timed out")
		}
	}
	if err := send("Use only the approved synthetic primary display. Capture default and explicit zero through the scripted tool calls."); err != nil {
		return err
	}
	if err := ended(); err != nil {
		return err
	}
	gate.Lock()
	firstError := failure
	if firstError == nil && script.stage != 5 {
		firstError = fmt.Errorf("first host turn did not complete both captures")
	}
	if firstError == nil {
		script.replayAllowed = true
	}
	gate.Unlock()
	if firstError != nil {
		return firstError
	}
	if err := send("Verify both retained screenshots on this new turn. Do not capture again."); err != nil {
		return err
	}
	return ended()
}
