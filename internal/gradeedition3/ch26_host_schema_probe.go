package gradeedition3

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sync/atomic"
	"time"

	"github.com/waywardgeek/ensemble/internal/fakevendor"
)

// runHostScreenSchemaCLI loads the public skill and inspects declarations only.
// Its fake provider never requests enumeration or capture. This earns no points.
func runHostScreenSchemaCLI(bin string) (result error) {
	root, err := os.MkdirTemp("", "ensemble-host-schema-")
	if err != nil {
		return err
	}
	defer func() { result = errors.Join(result, os.RemoveAll(root)) }()
	skills := filepath.Join(root, "skills")
	for _, dir := range []string{filepath.Join(root, "home"), filepath.Join(skills, "primary"), filepath.Join(skills, "host-screen")} {
		if err := os.MkdirAll(dir, 0700); err != nil {
			return err
		}
	}
	primary := "---\nname: primary\ntype: primary\ndescription: Host-screen declaration probe\ntools: load_skill wait_for_job\nloadable-skills: host-screen\n---\nLoad host-screen to inspect its declaration. Do not enumerate or capture displays.\n"
	screen := "---\nname: host-screen\ntype: loadable\ndescription: Explicit host-screen capability\ntools: list_displays screenshot\n---\nNo native action is authorized by this schema-only probe.\n"
	for name, text := range map[string]string{"primary": primary, "host-screen": screen} {
		if err := os.WriteFile(filepath.Join(skills, name, "SKILL.md"), []byte(text), 0600); err != nil {
			return err
		}
	}
	var calls atomic.Int32
	bodies := make(chan []byte, 4)
	provider := fakevendor.NewWithOptions(nil, fakevendor.Options{Route: func(body []byte) *fakevendor.Reply {
		if hostRecallRequest(body) {
			return &fakevendor.Reply{Text: "[]"}
		}
		number := calls.Add(1)
		select {
		case bodies <- append([]byte(nil), body...):
		default:
		}
		if number == 1 {
			return &fakevendor.Reply{Text: "Use only the agreed fully synthetic primary display.", Tools: []fakevendor.ToolCall{hintCall("host-schema-load", "load_skill", map[string]any{"name": "host-screen", "ai_callback_delay": 10})}}
		}
		return &fakevendor.Reply{Text: "HOST_SCHEMA_COMPLETE; no enumeration or capture requested."}
	}})
	defer provider.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()
	command := exec.CommandContext(ctx, bin)
	command.Dir = root
	command.Env = append(voiceGradeEnvironment(filepath.Join(root, "home")),
		"LLM_VENDOR=anthropic", "LLM_MODEL=claude-3-haiku-20240307", "LLM_BASE_URL="+provider.URL(),
		"LLM_API_KEY=synthetic-host-schema-key", "LLM_REQUIRE_TOOL_EXPLANATION=false",
		"EN_DISABLE_STREAMING=1", "EN_DISABLE_SAVE=1", "EN_SKILLS_DIR="+skills, "EN_PRIMARY_SKILL=primary",
		"EN_STUB_TOOL_RESULTS=0", "EN_CONTEXT_BUDGET_BYTES=400000")
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
				result = errors.Join(result, fmt.Errorf("schema candidate exit: %w; %s", err, diagnostics.text.String()))
			}
		case <-time.After(3 * time.Second):
			cancel()
			<-exited
			result = errors.Join(result, fmt.Errorf("schema candidate failed joined exit"))
		}
		if output.overflow || diagnostics.overflow {
			result = errors.Join(result, fmt.Errorf("schema diagnostics exceeded bound"))
		}
		if result == nil && calls.Load() != 2 {
			result = fmt.Errorf("expected two declaration exchanges, got %d", calls.Load())
		}
	}()
	if _, err := fmt.Fprintln(input, `{"kind":"prompt","text":"Use only the approved synthetic primary display. Capture default and explicit zero through the scripted tool calls."}`); err != nil {
		return err
	}
	var last []byte
	for range 2 {
		select {
		case last = <-bodies:
		case <-done:
			return fmt.Errorf("candidate exited before skill declaration")
		case <-ctx.Done():
			return fmt.Errorf("host schema probe timed out")
		}
	}
	if len(last) > 16<<20 {
		return fmt.Errorf("host schema request exceeds bound")
	}
	var request struct {
		Tools []struct {
			Name   string          `json:"name"`
			Schema json.RawMessage `json:"input_schema"`
		} `json:"tools"`
	}
	if err := json.Unmarshal(last, &request); err != nil {
		return err
	}
	screenshots, displays := 0, 0
	for _, tool := range request.Tools {
		if tool.Name == "screenshot" {
			screenshots++
			if err := screenshotSelectionSchema(tool.Schema); err != nil {
				return err
			}
		}
		if tool.Name == "list_displays" {
			displays++
		}
	}
	if screenshots != 1 || displays != 1 {
		load, loadError := hostToolText(last, "host-schema-load")
		return fmt.Errorf("expected one screenshot and one list_displays declaration, got %d/%d; load result: %s; error: %v", screenshots, displays, load, loadError)
	}
	select {
	case <-output.ended:
		return nil
	case <-ctx.Done():
		return fmt.Errorf("host schema turn did not finish")
	}
}
