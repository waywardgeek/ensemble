package main

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestNormalHarnessWiring(t *testing.T) {
	root, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	t.Setenv("WORKSPACE", t.TempDir())
	t.Setenv("LLM_BASE_URL", "http://127.0.0.1:9999/v1")
	log := filepath.Join(t.TempDir(), "transcript.jsonl")
	t.Setenv("TTS_LOG", log)
	t.Setenv("STREAMING", "off")
	t.Setenv("LLM_API_KEY", "must-not-propagate")
	t.Setenv("EN_MCP_SERVERS", "must-not-propagate")
	t.Setenv("LLM_MODEL", "current-model")
	c, prompt, err := normalConfig(root, []string{"question", "--type-during-turn", " "})
	if err != nil {
		t.Fatal(err)
	}
	if prompt != "question" || c.logPath != log || c.typeDuring != " " || c.streaming {
		t.Fatalf("normal wiring: %+v", c)
	}
	env := strings.Join(guiEnvironment(c), "\n")
	for _, want := range []string{"LLM_DISABLE_STREAMING=true", "LLM_STREAM_CAPABILITY=plain", "LLM_MODEL=current-model", "LLM_BASE_URL=http://127.0.0.1:9999/v1"} {
		if !strings.Contains(env, want) {
			t.Fatalf("missing %s", want)
		}
	}
	if strings.Contains(env, "must-not-propagate") {
		t.Fatal("unrelated credentials/config propagated")
	}
	t.Setenv("STREAMING", "")
	c, _, err = normalConfig(root, []string{"question"})
	if err != nil || !c.streaming {
		t.Fatalf("default stream: %+v %v", c, err)
	}
	t.Setenv("TTS_LOG", "")
	if _, _, err = normalConfig(root, []string{"question"}); err == nil {
		t.Fatal("missing TTS_LOG accepted")
	}
}
func TestOwnedProcessStops(t *testing.T) {
	// A shell process only. No browser, GUI binary, provider, or network.
	cmd := exec.Command("/bin/sh", "-c", "trap 'exit 0' TERM; while read line; do :; done")
	process, err := startOwned(cmd, filepath.Join(t.TempDir(), "process.log"))
	if err != nil {
		t.Fatal(err)
	}
	finished := make(chan error, 1)
	go func() { finished <- process.stop() }()
	select {
	case err = <-finished:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(4 * time.Second):
		t.Fatal("owned process did not stop")
	}
	if cmd.ProcessState == nil {
		t.Fatal("owned process was not reaped")
	}
}
func TestReadinessReportsOwnedExit(t *testing.T) {
	cmd := exec.Command("/bin/sh", "-c", "exit 7")
	process, err := startOwned(cmd, filepath.Join(t.TempDir(), "process.log"))
	if err != nil {
		t.Fatal(err)
	}
	defer process.stop()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err = awaitHTTP(ctx, "http://127.0.0.1:1", process); err == nil {
		t.Fatal("early child exit reported ready")
	}
}
func TestNormalGuardBeforeLaunch(t *testing.T) {
	t.Setenv("LLM_BASE_URL", "")
	t.Setenv("WORKSPACE", "")
	t.Setenv("TTS_LOG", "")
	_, _, err := normalConfig("../..", []string{"question"})
	if err == nil {
		t.Fatal("missing public inputs accepted")
	}
	if _, err = os.Stat("../../tts-harness.sh"); err != nil {
		t.Fatal(err)
	}
}
