package grade

// ch11_harness.go — Persistence grader.
//
// Drives a multi-turn conversation, saves state to disk, then tests
// the three persistence invariants: deterministic rebuild, resume
// from save, and context-alone sufficiency.

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/waywardgeek/ensemble/internal/fakevendor"
)

// Ch11Result carries the outcomes of the persistence tests.
type Ch11Result struct {
	DeterministicOK  bool
	DeterministicErr string

	PartialReplayOK  bool
	PartialReplayErr string

	SaveConfigOK  bool
	SaveConfigErr string

	ResumeOK  bool
	ResumeErr string

	LogNotNeededOK  bool
	LogNotNeededErr string

	RoundtripOK  bool
	RoundtripErr string

	Ch10Parity    bool
	Ch10ParityErr string
}

// Ch11Run drives all persistence checks.
func Ch11Run(path string) Ch11Result {
	r := Ch11Result{}

	bin, cleanup, err := Build(path)
	if err != nil {
		msg := "build: " + err.Error()
		r.DeterministicErr = msg
		r.PartialReplayErr = msg
		r.SaveConfigErr = msg
		r.ResumeErr = msg
		r.LogNotNeededErr = msg
		r.RoundtripErr = msg
		return r
	}
	defer cleanup()

	tmp, err := os.MkdirTemp("", "ch11-grade-*")
	if err != nil {
		msg := "tmpdir: " + err.Error()
		r.DeterministicErr = msg
		return r
	}
	defer os.RemoveAll(tmp)

	guiDir := filepath.Join(tmp, "gui")
	os.MkdirAll(guiDir, 0755)

	skillsDir := filepath.Join(tmp, "skills")
	createCh11Skills(skillsDir)

	// Phase 1: Drive a 3-turn conversation and save.
	savePath := filepath.Join(tmp, "agent.json")
	if err := ch11DriveAndSave(bin, savePath, skillsDir, guiDir, tmp); err != "" {
		r.DeterministicErr = "phase1: " + err
		r.PartialReplayErr = "phase1: " + err
		r.SaveConfigErr = "phase1: " + err
		r.ResumeErr = "phase1: " + err
		r.LogNotNeededErr = "phase1: " + err
		r.RoundtripErr = "phase1: " + err
		return r
	}

	ch11CheckDeterministic(bin, savePath, &r)
	ch11CheckSaveConfig(savePath, &r)
	ch11CheckRoundtrip(savePath, &r)
	ch11CheckResume(bin, savePath, skillsDir, guiDir, tmp, &r)
	ch11CheckLogNotNeeded(bin, savePath, skillsDir, guiDir, tmp, &r)
	ch11CheckPartialReplay(savePath, &r)

	return r
}

func ch11DriveAndSave(bin, savePath, skillsDir, guiDir, tmp string) string {
	replies := []fakevendor.Reply{
		{Text: "Hello! I am your assistant.", Usage: fakevendor.Canonical{Input: 100, Output: 30}},
		{Text: "The meaning of life is 42.", Usage: fakevendor.Canonical{Input: 200, Output: 20}},
		{Text: "Goodbye, and thanks for all the fish.", Usage: fakevendor.Canonical{Input: 300, Output: 40}},
	}
	srv := fakevendor.New(replies)
	defer srv.Close()

	port := freePort()
	logDir := filepath.Join(tmp, "phase1-log")
	os.MkdirAll(logDir, 0755)

	cmd := exec.Command(bin, "--port", port, "--gui-dir", guiDir,
		"--save", savePath)
	cmd.Dir = tmp
	cmd.Env = append(os.Environ(),
		"LLM_BASE_URL="+srv.URL(),
		"LLM_MODEL=fake-model",
		"LLM_VENDOR=anthropic",
		"LLM_API_KEY=test-key",
		"EN_SKILLS_DIR="+skillsDir,
		"EN_PRIMARY_SKILL=base",
		"CH02_LOG="+filepath.Join(logDir, "events.jsonl"),
	)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return fmt.Sprintf("stdin pipe: %v", err)
	}
	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		return fmt.Sprintf("start: %v", err)
	}

	if !waitForHTTP("http://localhost:"+port+"/", 10*time.Second) {
		cmd.Process.Kill()
		cmd.Wait()
		return "GUI server did not start"
	}

	prompts := []string{
		`{"kind":"prompt","text":"Hello, who are you?"}`,
		`{"kind":"prompt","text":"What is the meaning of life?"}`,
		`{"kind":"prompt","text":"Goodbye!"}`,
	}
	for _, p := range prompts {
		fmt.Fprintln(stdin, p)
		time.Sleep(500 * time.Millisecond)
	}

	// Wait for final response to be processed.
	time.Sleep(2 * time.Second)

	// Close stdin to trigger save and exit.
	stdin.Close()
	done := make(chan error, 1)
	go func() { done <- cmd.Wait() }()
	select {
	case <-done:
	case <-time.After(10 * time.Second):
		cmd.Process.Kill()
		<-done
	}

	if _, err := os.Stat(savePath); os.IsNotExist(err) {
		return "save file not created"
	}
	return ""
}

func ch11CheckDeterministic(bin, savePath string, r *Ch11Result) {
	data, err := os.ReadFile(savePath)
	if err != nil {
		r.DeterministicErr = fmt.Sprintf("read save: %v", err)
		return
	}

	var raw struct {
		Context json.RawMessage   `json:"context"`
		Log     []json.RawMessage `json:"log"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		r.DeterministicErr = fmt.Sprintf("parse save: %v", err)
		return
	}

	if len(raw.Log) == 0 {
		r.DeterministicErr = "save file has no events in log"
		return
	}

	var ctx struct {
		Turn     int `json:"turn"`
		Dialogue []struct {
			Seq   int    `json:"seq"`
			Actor string `json:"actor"`
		} `json:"dialogue"`
	}
	if err := json.Unmarshal(raw.Context, &ctx); err != nil {
		r.DeterministicErr = fmt.Sprintf("parse context: %v", err)
		return
	}

	humanCount := 0
	agentCount := 0
	for _, entry := range ctx.Dialogue {
		switch entry.Actor {
		case "human":
			humanCount++
		case "agent":
			agentCount++
		}
	}
	if humanCount < 3 {
		r.DeterministicErr = fmt.Sprintf("expected ≥3 human entries, got %d", humanCount)
		return
	}
	if agentCount < 3 {
		r.DeterministicErr = fmt.Sprintf("expected ≥3 agent entries, got %d", agentCount)
		return
	}
	if len(raw.Log) < 6 {
		r.DeterministicErr = fmt.Sprintf("expected ≥6 log events, got %d", len(raw.Log))
		return
	}

	// Run the binary's verify command to test Rebuild determinism
	verifyCmd := exec.Command(bin, "verify", "--save", savePath)
	verifyOut, err := verifyCmd.CombinedOutput()
	if err != nil {
		r.DeterministicErr = fmt.Sprintf("verify command failed: %s (%v)", string(verifyOut), err)
		return
	}
	if !strings.Contains(string(verifyOut), "MATCH") {
		r.DeterministicErr = fmt.Sprintf("verify output: %s", string(verifyOut))
		return
	}

	r.DeterministicOK = true
}

func ch11CheckSaveConfig(savePath string, r *Ch11Result) {
	data, err := os.ReadFile(savePath)
	if err != nil {
		r.SaveConfigErr = fmt.Sprintf("read: %v", err)
		return
	}

	var sf struct {
		Config struct {
			Model        string `json:"model"`
			Vendor       string `json:"vendor"`
			SystemPrompt string `json:"system_prompt"`
			Tools        []struct {
				Name string `json:"name"`
			} `json:"tools"`
		} `json:"config"`
	}
	if err := json.Unmarshal(data, &sf); err != nil {
		r.SaveConfigErr = fmt.Sprintf("parse: %v", err)
		return
	}

	if sf.Config.Model == "" {
		r.SaveConfigErr = "config.model is empty"
		return
	}
	if sf.Config.Vendor == "" {
		r.SaveConfigErr = "config.vendor is empty"
		return
	}
	if sf.Config.SystemPrompt == "" {
		r.SaveConfigErr = "config.system_prompt is empty"
		return
	}
	if len(sf.Config.Tools) == 0 {
		r.SaveConfigErr = "config.tools is empty"
		return
	}

	toolNames := make(map[string]bool)
	for _, t := range sf.Config.Tools {
		toolNames[t.Name] = true
	}
	for _, want := range []string{"read_file", "think"} {
		if !toolNames[want] {
			r.SaveConfigErr = fmt.Sprintf("missing tool %q", want)
			return
		}
	}

	r.SaveConfigOK = true
}

func ch11CheckRoundtrip(savePath string, r *Ch11Result) {
	// Read, parse, re-marshal, parse again, compare structurally.
	data, err := os.ReadFile(savePath)
	if err != nil {
		r.RoundtripErr = fmt.Sprintf("read: %v", err)
		return
	}

	// First parse: verify it's valid JSON.
	var first interface{}
	if err := json.Unmarshal(data, &first); err != nil {
		r.RoundtripErr = fmt.Sprintf("parse: %v", err)
		return
	}

	// Re-marshal and re-parse: verify the JSON survives a roundtrip.
	remarshaled, err := json.Marshal(first)
	if err != nil {
		r.RoundtripErr = fmt.Sprintf("remarshal: %v", err)
		return
	}

	var second interface{}
	if err := json.Unmarshal(remarshaled, &second); err != nil {
		r.RoundtripErr = fmt.Sprintf("reparse: %v", err)
		return
	}

	// Structural comparison via re-marshaling to canonical form.
	canon1, _ := json.Marshal(first)
	canon2, _ := json.Marshal(second)

	if string(canon1) != string(canon2) {
		r.RoundtripErr = "structural mismatch after roundtrip"
		return
	}

	r.RoundtripOK = true
}

func ch11CheckResume(bin, savePath, skillsDir, guiDir, tmp string, r *Ch11Result) {
	replies := []fakevendor.Reply{
		{Text: "I remember our conversation!", Usage: fakevendor.Canonical{Input: 500, Output: 30}},
	}
	srv := fakevendor.New(replies)
	defer srv.Close()

	port := freePort()
	logDir := filepath.Join(tmp, "resume-log")
	os.MkdirAll(logDir, 0755)

	cmd := exec.Command(bin, "--port", port, "--gui-dir", guiDir,
		"--load", savePath)
	cmd.Dir = tmp
	cmd.Env = append(os.Environ(),
		"LLM_BASE_URL="+srv.URL(),
		"LLM_MODEL=fake-model",
		"LLM_VENDOR=anthropic",
		"LLM_API_KEY=test-key",
		"EN_SKILLS_DIR="+skillsDir,
		"EN_PRIMARY_SKILL=base",
		"CH02_LOG="+filepath.Join(logDir, "events.jsonl"),
	)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		r.ResumeErr = fmt.Sprintf("stdin pipe: %v", err)
		return
	}
	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		r.ResumeErr = fmt.Sprintf("start: %v", err)
		return
	}

	if !waitForHTTP("http://localhost:"+port+"/", 10*time.Second) {
		cmd.Process.Kill()
		cmd.Wait()
		r.ResumeErr = "GUI server did not start for resume"
		return
	}

	fmt.Fprintln(stdin, `{"kind":"prompt","text":"Do you remember me?"}`)
	time.Sleep(2 * time.Second)

	stdin.Close()
	done := make(chan error, 1)
	go func() { done <- cmd.Wait() }()
	select {
	case <-done:
	case <-time.After(10 * time.Second):
		cmd.Process.Kill()
		<-done
	}

	reqs := srv.Requests()
	if len(reqs) == 0 {
		r.ResumeErr = "no requests to fakevendor after resume"
		return
	}

	lastReq := string(reqs[len(reqs)-1].Body)
	if !strings.Contains(lastReq, "Hello") &&
		!strings.Contains(lastReq, "meaning of life") {
		r.ResumeErr = "resumed request missing original conversation"
		return
	}
	if !strings.Contains(lastReq, "Do you remember me") {
		r.ResumeErr = "resumed request missing new prompt"
		return
	}

	r.ResumeOK = true
}

func ch11CheckLogNotNeeded(bin, savePath, skillsDir, guiDir, tmp string, r *Ch11Result) {
	data, err := os.ReadFile(savePath)
	if err != nil {
		r.LogNotNeededErr = fmt.Sprintf("read save: %v", err)
		return
	}

	var sf map[string]interface{}
	if err := json.Unmarshal(data, &sf); err != nil {
		r.LogNotNeededErr = fmt.Sprintf("parse: %v", err)
		return
	}

	sf["log"] = []interface{}{}
	noLogData, err := json.MarshalIndent(sf, "", "  ")
	if err != nil {
		r.LogNotNeededErr = fmt.Sprintf("marshal: %v", err)
		return
	}

	noLogPath := filepath.Join(tmp, "no-log.json")
	if err := os.WriteFile(noLogPath, noLogData, 0644); err != nil {
		r.LogNotNeededErr = fmt.Sprintf("write: %v", err)
		return
	}

	replies := []fakevendor.Reply{
		{Text: "Context-only agent works!", Usage: fakevendor.Canonical{Input: 500, Output: 30}},
	}
	srv := fakevendor.New(replies)
	defer srv.Close()

	port := freePort()
	logDir := filepath.Join(tmp, "no-log-dir")
	os.MkdirAll(logDir, 0755)

	cmd := exec.Command(bin, "--port", port, "--gui-dir", guiDir,
		"--load", noLogPath)
	cmd.Dir = tmp
	cmd.Env = append(os.Environ(),
		"LLM_BASE_URL="+srv.URL(),
		"LLM_MODEL=fake-model",
		"LLM_VENDOR=anthropic",
		"LLM_API_KEY=test-key",
		"EN_SKILLS_DIR="+skillsDir,
		"EN_PRIMARY_SKILL=base",
		"CH02_LOG="+filepath.Join(logDir, "events.jsonl"),
	)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		r.LogNotNeededErr = fmt.Sprintf("stdin pipe: %v", err)
		return
	}
	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		r.LogNotNeededErr = fmt.Sprintf("start: %v", err)
		return
	}

	if !waitForHTTP("http://localhost:"+port+"/", 10*time.Second) {
		cmd.Process.Kill()
		cmd.Wait()
		r.LogNotNeededErr = "GUI server did not start"
		return
	}

	fmt.Fprintln(stdin, `{"kind":"prompt","text":"Are you alive?"}`)
	time.Sleep(2 * time.Second)

	stdin.Close()
	done := make(chan error, 1)
	go func() { done <- cmd.Wait() }()
	select {
	case <-done:
	case <-time.After(10 * time.Second):
		cmd.Process.Kill()
		<-done
	}

	reqs := srv.Requests()
	if len(reqs) == 0 {
		r.LogNotNeededErr = "no requests to fakevendor"
		return
	}

	lastReq := string(reqs[len(reqs)-1].Body)
	if !strings.Contains(lastReq, "Hello") {
		r.LogNotNeededErr = "context-only load missing original conversation"
		return
	}

	r.LogNotNeededOK = true
}

func ch11CheckPartialReplay(savePath string, r *Ch11Result) {
	data, err := os.ReadFile(savePath)
	if err != nil {
		r.PartialReplayErr = fmt.Sprintf("read: %v", err)
		return
	}

	var sf struct {
		Log []json.RawMessage `json:"log"`
	}
	if err := json.Unmarshal(data, &sf); err != nil {
		r.PartialReplayErr = fmt.Sprintf("parse: %v", err)
		return
	}

	if len(sf.Log) < 4 {
		r.PartialReplayErr = fmt.Sprintf("need ≥4 log events, got %d", len(sf.Log))
		return
	}

	type seqEvent struct {
		Seq int `json:"seq"`
	}
	prevSeq := -1
	for i, raw := range sf.Log {
		var e seqEvent
		if err := json.Unmarshal(raw, &e); err != nil {
			r.PartialReplayErr = fmt.Sprintf("event %d: %v", i, err)
			return
		}
		if e.Seq <= prevSeq {
			r.PartialReplayErr = fmt.Sprintf("event %d: seq %d not monotonic (prev %d)", i, e.Seq, prevSeq)
			return
		}
		prevSeq = e.Seq
	}

	r.PartialReplayOK = true
}

func createCh11Skills(dir string) {
	os.MkdirAll(filepath.Join(dir, "base"), 0755)
	os.WriteFile(filepath.Join(dir, "base", "SKILL.md"), []byte(`---
name: base
description: Base assistant skill
type: primary
tools: read_file think
---

You are a helpful assistant.
`), 0644)
}
