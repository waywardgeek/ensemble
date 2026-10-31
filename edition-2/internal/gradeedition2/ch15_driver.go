package gradeedition2

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/waywardgeek/ensemble/internal/fakevendor"
)

type invoiceOpts struct {
	executable         bool // Explicit C# host contract; legacy launch remains unchanged.
	dir, vendor, model string
	prompts            []string
	replies            []fakevendor.Reply
}
type invoiceOut struct {
	reqs   []fakevendor.Recorded
	events []map[string]json.RawMessage
	meters []map[string]any
	fatal  string
}

// Each snapshot comes from a fresh subscription after a completed turn. This
// observes the live meter without assuming its broadcast precedes turn_ended,
// or using the original arbitrary post-response sleep to guess at completion.
// Keep these subscribers alive through shutdown: mid-turn disconnect is a
// separately reviewed lifetime concern, not a prerequisite to meter arithmetic.
func invoiceSnapshot(url string) (map[string]any, *guiPeer, error) {
	return invoiceSnapshotFor(url, false)
}

func invoiceSnapshotFor(url string, executable bool) (map[string]any, *guiPeer, error) {
	p, err := openGUIPeer(url)
	if err != nil {
		return nil, nil, err
	}

	if err = invoiceSend(p, "subscribe", "", executable); err != nil {
		return nil, p, err
	}
	meter, err := invoiceReadSnapshot(p)
	return meter, p, err
}

// Retain usage arriving before settings, including the real event in a replay
// batch. Do not substitute diagnostic meter state or manufacture initial counts.
func invoiceReadSnapshot(p *guiPeer) (map[string]any, error) {
	var seen invoiceSnapshotFrames
	_, err := p.until(3*time.Second, func(m map[string]any) bool {
		seen.observe(m)
		return seen.settings || seen.err != nil
	})
	if err != nil {
		return nil, err
	}
	if seen.err != nil {
		return nil, seen.err
	}
	if seen.usage == nil {
		_, err = p.until(5*time.Second, func(m map[string]any) bool {
			seen.observe(m)
			return seen.usage != nil || seen.err != nil
		})
	}
	if err != nil {
		return nil, err
	}
	return seen.usage, seen.err
}

type invoiceSnapshotFrames struct {
	settings bool
	usage    map[string]any
	err      error
}

func (s *invoiceSnapshotFrames) observe(m map[string]any) {
	switch m["type"] {
	case "current_settings":
		if object(m["settings"]) == nil {
			s.err = fmt.Errorf("current_settings lacks a settings object")
		} else {
			s.settings = true
		}
	case "usage":
		s.usage = m
	case "replay":
		events, _ := object(m["data"])["events"].([]any)
		for _, event := range events {
			if e := object(event); e != nil && e["type"] == "usage" {
				s.usage = e
			}
		}
	}
}

func invoiceLaunch(bin, skillsDir, guiDir string, o invoiceOpts) (out invoiceOut) {
	srv := fakevendor.New(o.replies)
	defer srv.Close()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		out.fatal = err.Error()
		return
	}
	port := fmt.Sprint(listener.Addr().(*net.TCPAddr).Port)
	listener.Close()
	url := "ws://localhost:" + port + "/ws"
	args := []string{"--port", port, "--gui-dir", guiDir}
	if o.executable {
		args = append(args, "--no-stream", "--vendor", o.vendor, "--model", o.model,
			"--base-url", srv.URL(), "--skills-dir", skillsDir, "--primary-skill", "base",
			"--event-log", filepath.Join(o.dir, "events.jsonl"), "--settings", filepath.Join(o.dir, "settings.json"),
			"--save", filepath.Join(o.dir, "save.json"), "--save-log", "-1",
			"--cache-lens", filepath.Join(o.dir, "lens"))
		if o.vendor == "anthropic" {
			args = append(args, "--cache-breakpoints")
		}
	}
	cmd := exec.Command(bin, args...)
	cmd.Dir = o.dir
	cmd.Env = invoiceEnv(os.Environ(), skillsDir, srv.URL(), o)
	var stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = io.Discard, &stderr
	stdin, err := cmd.StdinPipe()
	if err != nil {
		out.fatal = err.Error()
		return
	}
	if err = cmd.Start(); err != nil {
		stdin.Close()
		out.fatal = err.Error()
		return
	}
	done := make(chan error, 1)
	go func() { done <- cmd.Wait() }()
	finished := false
	defer func() {
		stdin.Close()
		if !finished {
			cmd.Process.Kill()
			<-done
		}
		out.reqs = srv.Requests()
		if o.executable {
			os.WriteFile(filepath.Join(o.dir, "host.stderr"), stderr.Bytes(), 0600)
			if b, e := json.Marshal(out.reqs); e == nil {
				os.WriteFile(filepath.Join(o.dir, "provider-requests.json"), b, 0600)
			}
			// Current-process event log, not the saved log containing earlier sessions.
			out.events, err = contextReadEvents(filepath.Join(o.dir, "events.jsonl"))
		} else {
			out.events, err = recallReadEvents(o.dir)
		}
		if err != nil && out.fatal == "" {
			out.fatal = err.Error()
		}
		if out.fatal != "" {
			out.fatal += ": " + stderr.String()
		}
	}()
	var peer *guiPeer
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		peer, err = openGUIPeer(url)
		if err == nil {
			break
		}
		select {
		case e := <-done:
			finished = true
			out.fatal = fmt.Sprintf("GUI exited before ready: %v", e)
			return
		default:
		}
		time.Sleep(20 * time.Millisecond)
	}
	if err != nil {
		out.fatal = "GUI socket did not start: " + err.Error()
		return
	}
	defer peer.close()
	if err = invoiceSend(peer, "subscribe", "", o.executable); err != nil {
		out.fatal = err.Error()
		return
	}
	if _, err = readSettings(peer, "current_settings"); err != nil {
		out.fatal = err.Error()
		return
	}
	// The initial frame is measured even when the save contains earlier usage.
	initial, snapshot, err := invoiceSnapshotFor(url, o.executable)
	if snapshot != nil {
		defer snapshot.close()
	}
	if err != nil {
		out.fatal = err.Error()
		return
	}
	out.meters = append(out.meters, initial)
	for i, prompt := range o.prompts {
		if o.executable {
			err = invoiceSend(peer, "prompt", prompt, true)
		} else {
			err = json.NewEncoder(stdin).Encode(map[string]string{"kind": "prompt", "text": prompt})
		}
		if err != nil {
			out.fatal = err.Error()
			return
		}
		messages, e := peer.until(10*time.Second, messageKind("turn_ended"))
		if e != nil {
			out.fatal = fmt.Sprintf("prompt %d completion: %v", i+1, e)
			return
		}
		if problem, ok := messages[len(messages)-1]["error"].(string); ok && problem != "" {
			out.fatal = "turn failed: " + problem
			return
		}
		meter, snapshot, e := invoiceSnapshotFor(url, o.executable)
		if snapshot != nil {
			defer snapshot.close()
		}
		if e != nil {
			out.fatal = e.Error()
			return
		}
		out.meters = append(out.meters, meter)
		if n := len(srv.Requests()); n != i+1 {
			out.fatal = fmt.Sprintf("completed turn %d made %d requests", i+1, n)
			return
		}
	}
	stdin.Close()
	ending := "EOF"
	if o.executable {
		ending = "SIGTERM"
		if err = cmd.Process.Signal(syscall.SIGTERM); err != nil {
			out.fatal = err.Error()
			return
		}
	}
	select {
	case err = <-done:
		finished = true
		if err != nil {
			out.fatal = "nonzero GUI exit after " + ending + ": " + err.Error()
		}
	case <-time.After(5 * time.Second):
		out.fatal = "GUI did not save and exit after " + ending
	}
	return
}

// Artifact paths belong to the submitted client, not the core library. They
// are explicit grader fixture inputs and never become application flags.
func invoiceRead(dir, path string) ([]byte, error) {
	return os.ReadFile(filepath.Join(dir, filepath.FromSlash(path)))
}

// Control application settings without changing PATH or offline build settings.
// Restart scenarios reuse only their owned run directory.
func invoiceEnv(base []string, skills, url string, o invoiceOpts) []string {
	env := make([]string, 0, len(base)+24)
	for _, item := range base {
		key, _, _ := strings.Cut(item, "=")
		if strings.HasPrefix(key, "LLM_") || strings.HasPrefix(key, "EN_") ||
			strings.HasPrefix(key, "ENSEMBLE_") || strings.HasPrefix(key, "ANTHROPIC_") ||
			strings.HasPrefix(key, "OPENAI_") || strings.HasPrefix(key, "GEMINI_") {
			continue
		}
		env = append(env, item)
	}
	return append(env,
		"ANTHROPIC_API_KEY=fake", "OPENAI_API_KEY=fake", "GEMINI_API_KEY=fake",
		"ANTHROPIC_BASE_URL="+url, "OPENAI_BASE_URL="+url, "GEMINI_BASE_URL="+url,
		"LLM_VENDOR="+o.vendor, "LLM_MODEL="+o.model, "LLM_RESOLVED_MODEL="+o.model,
		"LLM_BASE_URL="+url, "LLM_API_KEY=fake", "LLM_TOOLS=default",
		"ENSEMBLE_SAVE="+filepath.Join(o.dir, "save.json"), "ENSEMBLE_LOG_MODE=full",
		"LLM_LOG_PATH="+filepath.Join(o.dir, "events.jsonl"),
		"EN_SKILLS_DIR="+skills, "EN_PRIMARY_SKILL=base",
		"LLM_PAUSE_ON_START=false", "LLM_SPEAKING=false",
		"LLM_STREAM_CAPABILITY=plain", "EN_DISABLE_STREAMING=1")
}
