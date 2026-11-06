package gradeedition3

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"
)

func TestLocalModelReceipt(t *testing.T) {
	names := localModelCheckNames()
	if len(names) != 73 {
		t.Fatal(len(names))
	}
	encode := func(checks []string) []byte {
		b, err := json.Marshal(map[string]any{"passed": true, "checks": checks, "failure": nil})
		if err != nil {
			t.Fatal(err)
		}
		return b
	}
	good := encode(names)
	if err := checkLocalModelReceipt(good); err != nil {
		t.Fatal(err)
	}
	for i, name := range names {
		t.Run(name, func(t *testing.T) {
			reduced := append([]string{}, names[:i]...)
			reduced = append(reduced, names[i+1:]...)
			if checkLocalModelReceipt(encode(reduced)) == nil {
				t.Fatal("omission accepted")
			}
		})
	}
	duplicate := append([]string{}, names...)
	duplicate[1] = duplicate[0]
	unknown := append([]string{}, names...)
	unknown[0] = "invented"
	for _, bad := range [][]byte{encode(duplicate), encode(unknown), []byte(`{"passed":true}`), []byte(`{"passed":true,"checks":[],"failure":null,"passed":true}`), append(append([]byte{}, good...), good...), []byte(strings.Replace(string(good), `"failure":null,`, "", 1)), []byte(strings.Replace(string(good), `"passed":true`, `"passed":false`, 1))} {
		if checkLocalModelReceipt(bad) == nil {
			t.Fatalf("invalid receipt accepted: %s", bad)
		}
	}
}
func TestLocalModelProcessDeadline(t *testing.T) {
	root := t.TempDir()
	script := filepath.Join(root, "held.sh")
	if err := os.WriteFile(script, []byte("#!/bin/sh\ntrap '' TERM\necho $$ > pid\nwhile :; do sleep 1; done\n"), 0700); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()
	start := time.Now()
	_, _, err := runLocalModelCommand(ctx, root, localModelEnvironment(root), script)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("deadline not reported: %v", err)
	}
	if time.Since(start) > 4*time.Second {
		t.Fatal("deadline cleanup exceeded bound")
	}
	data, err := os.ReadFile(filepath.Join(root, "pid"))
	if err != nil {
		t.Fatal(err)
	}
	pid, err := strconv.Atoi(strings.TrimSpace(string(data)))
	if err != nil {
		t.Fatal(err)
	}
	if err = syscall.Kill(pid, 0); !errors.Is(err, syscall.ESRCH) {
		t.Fatalf("direct child not reaped: %v", err)
	}
}
func TestLocalModelInheritedPipeBound(t *testing.T) {
	root := t.TempDir()
	script := filepath.Join(root, "pipe.sh")
	if err := os.WriteFile(script, []byte("#!/bin/sh\nsleep 30 &\nexit 0\n"), 0700); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	start := time.Now()
	_, _, err := runLocalModelCommand(ctx, root, localModelEnvironment(root), script)
	if err == nil {
		t.Fatal("inherited open pipe was accepted")
	}
	if time.Since(start) > 4*time.Second {
		t.Fatal("inherited pipe did not reach WaitDelay")
	}
}
func TestLocalModelReportIncomplete(t *testing.T) {
	for _, passed := range []bool{false, true} {
		if (LocalModelReport{Status: "ungraded", ProbePassed: passed}).ExitCode() != 3 {
			t.Fatal("incomplete scope passed")
		}
	}
	if (LocalModelReport{Status: "failed"}).ExitCode() != 1 {
		t.Fatal("failure exit")
	}
}
func TestLocalModelExternal(t *testing.T) {
	candidate, sdk := os.Getenv("LOCAL_MODEL_RUNTIME"), os.Getenv("LOCAL_MODEL_DOTNET")
	if candidate == "" || sdk == "" {
		t.Skip("explicit existing candidate and SDK required")
	}
	report, err := GradeLocalModels(LocalModelGradeConfig{candidate, sdk})
	if err != nil {
		t.Fatal(err)
	}
	if !report.ProbePassed || report.ExitCode() != 3 || !strings.HasPrefix(report.Detail, strconv.Itoa(len(localModelCheckNames()))+" named public ") {
		t.Fatalf("%+v", report)
	}
}

func TestLocalModelClearedEnvironment(t *testing.T) {
	t.Setenv("LLM_API_KEY", "private-sentinel-not-forwarded")
	root := t.TempDir()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	output, diagnostic, err := runLocalModelCommand(ctx, root, localModelEnvironment(root), "/bin/sh", "-c", `test -z "$LLM_API_KEY" && test "$HOME" = "$PWD" && printf clean`)
	if err != nil || output != "clean" || diagnostic != "" {
		t.Fatalf("environment not isolated: %v %q %q", err, output, diagnostic)
	}
}
