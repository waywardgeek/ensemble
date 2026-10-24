package main

import (
	"context"
	"encoding/json"
	"flag"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"ensemble"
)

func TestChapter12GUIPublicImports(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	output, err := exec.CommandContext(ctx, "go", "list", "-f", "{{join .Imports \"\\n\"}}", "../...").CombinedOutput()
	if err != nil {
		t.Fatalf("GUI build imports: %v\n%s", err, output)
	}
	for _, name := range strings.Fields(string(output)) {
		if strings.HasPrefix(name, "ensemble/internal/") {
			t.Errorf("GUI consumer imports private core package %s", name)
		}
	}
}

// Run the real GUI entry point with EOF, so flags and environment pass through
// production wiring without a model turn, browser, or native device.
func TestChapter12GUIStartupSelection(t *testing.T) {
	if os.Getenv("E3_CH12_GUI_CHILD") == "1" {
		var args []string
		if err := json.Unmarshal([]byte(os.Getenv("E3_CH12_GUI_ARGS")), &args); err != nil {
			t.Fatal(err)
		}
		flag.CommandLine = flag.NewFlagSet("gui", flag.ExitOnError)
		os.Args = append([]string{"gui"}, args...)
		main()
		return
	}
	assets, err := filepath.Abs("../web/gui")
	if err != nil {
		t.Fatal(err)
	}
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	for _, mode := range []string{"plain", "default", "debug-default", "flag", "environment", "flag-over-environment", "debug-missing"} {
		t.Run(mode, func(t *testing.T) {
			dir := t.TempDir()
			writeSkills := func(directory, primary, marker string) {
				t.Helper()
				for name, text := range map[string]string{
					primary:     "---\nname: " + primary + "\ntype: primary\ndescription: fixture\nloadable-skills: gui-debug\n---\nPRIMARY:" + marker,
					"gui-debug": "---\nname: gui-debug\ntype: loadable\ndescription: fixture browser\ntools: gui_snapshot tts_queue gui_click gui_input\nmcp_servers:\n  - name: browser-debug\n    transport: websocket\n---\nSTARTUP:browser",
				} {
					path := filepath.Join(directory, name, "SKILL.md")
					if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
						t.Fatal(err)
					}
					if err := os.WriteFile(path, []byte(text), 0600); err != nil {
						t.Fatal(err)
					}
				}
			}
			args := []string{"-port", "0", "-gui-dir", assets}
			envDir, envPrimary, marker := "", "", ""
			debug := mode == "debug-default" || mode == "debug-missing" || mode == "flag-over-environment"
			if mode != "plain" && mode != "debug-missing" {
				writeSkills(filepath.Join(dir, "skills"), "ensemble", "default")
				marker = "default"
			}
			if mode == "environment" || mode == "flag-over-environment" {
				envDir, envPrimary, marker = filepath.Join(dir, "env-skills"), "custom", "environment"
				writeSkills(envDir, envPrimary, marker)
			}
			if mode == "flag" || mode == "flag-over-environment" {
				primary := "ensemble"
				if envPrimary != "" {
					primary = envPrimary
				}
				flagDir := filepath.Join(dir, "flag-skills")
				writeSkills(flagDir, primary, "flag")
				args, marker = append(args, "--skills-dir", flagDir), "flag"
			}
			if debug {
				args = append(args, "--gui-debug")
			}
			rawArgs, _ := json.Marshal(args)
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			cmd := exec.CommandContext(ctx, executable, "-test.run=^TestChapter12GUIStartupSelection$")
			cmd.Dir = dir
			cmd.Env = append(os.Environ(), "E3_CH12_GUI_CHILD=1", "E3_CH12_GUI_ARGS="+string(rawArgs), "LLM_MODEL=fixture", "LLM_VENDOR=openai", "LLM_API_KEY=fixture", "LLM_TOOLS=", "LLM_BASE_URL=", "EN_SKILLS_DIR="+envDir, "EN_PRIMARY_SKILL="+envPrimary, "ENSEMBLE_SAVE=", "LLM_PAUSE_ON_START=false")
			output, err := cmd.CombinedOutput()
			if mode == "debug-missing" {
				if err == nil || !strings.Contains(string(output), "--gui-debug requires") || !strings.Contains(string(output), "skills/ensemble/SKILL.md") {
					t.Fatalf("missing debug skills: %v\n%s", err, output)
				}
				return
			}
			if err != nil {
				t.Fatalf("GUI startup: %v\n%s", err, output)
			}
			raw, err := os.ReadFile(filepath.Join(dir, "save.json"))
			if err != nil {
				t.Fatalf("startup save: %v\n%s", err, output)
			}
			var save ensemble.SaveFile
			if err := json.Unmarshal(raw, &save); err != nil {
				t.Fatal(err)
			}
			if mode == "plain" {
				if len(save.Config.Skills) != 0 {
					t.Fatalf("plain startup selected skills: %+v", save.Config.Skills)
				}
				return
			}
			if !strings.Contains(save.Config.SystemPrompt, "PRIMARY:"+marker) || strings.Contains(save.Config.SystemPrompt, "STARTUP:browser") != debug {
				t.Fatalf("wrong initial instructions: %s", save.Config.SystemPrompt)
			}
			for _, skill := range save.Config.Skills {
				if skill.Name == "gui-debug" && (skill.Enabled != debug || (debug && (skill.Origin != "initial" || skill.EventSeq != 0))) {
					t.Fatalf("startup skill provenance: %+v", skill)
				}
			}
			browserTools := 0
			for _, tool := range save.Config.Tools {
				if tool.Name == "gui_snapshot" || tool.Name == "tts_queue" || tool.Name == "gui_click" || tool.Name == "gui_input" {
					browserTools++
				}
			}
			want := 0
			if debug {
				want = 4
			}
			if browserTools != want {
				t.Fatalf("startup browser tools=%d, want %d", browserTools, want)
			}
		})
	}
}
