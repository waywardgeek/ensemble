package gradeedition3

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/waywardgeek/ensemble/internal/fakevendor"
	"github.com/waywardgeek/ensemble/internal/grade"
	"github.com/waywardgeek/ensemble/internal/gradeedition2"
)

// Chapter5Executable leaves the historical Edition 2 gates unchanged. These
// additional checks drive the public CLI, not the submission's own test runner.
func Chapter5Executable(bin string, models map[string]string) grade.Report {
	checks := gradeedition2.Chapter5Executable(bin, models).Checks
	for _, mode := range []string{"early", "aged", "interrupt-now", "interrupt-after_tool"} {
		checks = append(checks, hintBatchCheck(bin, mode))
	}
	return grade.NewTitledReport("Edition 3, Chapter 5 — executable replay, media and non-disruptive hints", checks, "")
}

type hintRequest struct {
	body []byte
	at   time.Time
}

// Tool results are read only from Anthropic result blocks, never by searching
// the entire request (which also contains our own assistant call arguments).
func hintToolText(body []byte, id string) (string, error) {
	var request struct {
		Messages []struct{ Content json.RawMessage }
	}
	if err := json.Unmarshal(body, &request); err != nil {
		return "", err
	}
	var matched json.RawMessage
	failed := false
	count := 0
	for _, message := range request.Messages {
		var blocks []struct {
			Type    string
			ID      string `json:"tool_use_id"`
			Error   bool   `json:"is_error"`
			Content json.RawMessage
		}
		if json.Unmarshal(message.Content, &blocks) != nil {
			continue
		}
		for _, block := range blocks {
			if block.Type != "tool_result" || block.ID != id {
				continue
			}
			count++
			matched = block.Content
			failed = block.Error
		}
	}
	if count > 1 {
		return "", fmt.Errorf("duplicate results for %s", id)
	}
	if count == 1 {
		if failed {
			return "", fmt.Errorf("%s returned an error: %s", id, matched)
		}
		var text string
		if json.Unmarshal(matched, &text) == nil {
			return text, nil
		}
		var parts []struct{ Text string }
		if err := json.Unmarshal(matched, &parts); err != nil {
			return "", err
		}
		for _, part := range parts {
			text += part.Text
		}
		return text, nil
	}
	return "", fmt.Errorf("missing result for %s", id)
}

var hintHandle = regexp.MustCompile(`job ([0-9]+), status: running`)

func hintCall(id, name string, args map[string]any) fakevendor.ToolCall {
	encoded, _ := json.Marshal(args)
	return fakevendor.ToolCall{ID: id, Name: name, Args: string(encoded)}
}

func hintBatchCheck(bin, label string) (check grade.Check) {
	check = grade.Check{ID: "hint-" + label + "-batch", Title: label + " hint preserves pending effects and observation deadlines", Points: 10, Earned: 10, Passed: true}
	if err := runHintBatch(bin, label); err != nil {
		check.Passed = false
		check.Earned = 0
		check.Details = []string{err.Error()}
	} else if strings.HasPrefix(label, "interrupt-") {
		check.Title = label + " refuses undispatched effects without losing admitted jobs"
		check.Details = []string{"real CLI: interrupted turn ends, pending shell/write effects are absent with paired not-executed results; a fresh prompt recovers and retrieves the original job completion"}
	} else {
		check.Details = []string{"real CLI and gated shell jobs: current and remaining observations capped; pending write executes once; original call IDs and running handles retained", "repeated hints cannot reset the deadline; half-second waits and fast writes stay fast; same jobs finish after release; a fresh next-request wait remains uncapped; local fake provider only"}
	}
	return
}

func runHintBatch(bin, mode string) (result error) {
	aged := mode == "aged"
	interrupt := strings.HasPrefix(mode, "interrupt-")
	work, err := os.MkdirTemp("", "edition3-hints-*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(work)
	ctx, cancel := context.WithTimeout(context.Background(), 35*time.Second)
	defer cancel()
	requests := make(chan hintRequest, 16)
	// Route captures receipt time before returning a scripted reply. The fixture
	// consumes these records sequentially; the provider never waits for the test.
	replies := make(chan fakevendor.Reply, 4)
	shell := func(index int) string {
		return fmt.Sprintf("printf 'started\\n' >> started%d; n=0; while [ ! -f release ] && [ \"$n\" -lt 250 ]; do sleep 0.1; n=$((n+1)); done; printf 'finished\\n' > finished%d; printf FINISHED_%d", index, index, index)
	}
	replies <- fakevendor.Reply{Tools: []fakevendor.ToolCall{
		hintCall("held-1", "run_command", map[string]any{"command": shell(1), "ai_callback_delay": 20}),
		hintCall("held-2", "run_command", map[string]any{"command": shell(2), "ai_callback_delay": 20}),
		hintCall("held-3", "run_command", map[string]any{"command": shell(3), "ai_callback_delay": 0.5}),
		hintCall("pending-write", "write_file", map[string]any{"path": "written.txt", "content": "PENDING_WRITE_SURVIVED", "ai_callback_delay": 20}),
	}}
	fake := fakevendor.NewWithOptions(nil, fakevendor.Options{Route: func(body []byte) *fakevendor.Reply {
		select {
		case requests <- hintRequest{append([]byte(nil), body...), time.Now()}:
		default:
		}
		select {
		case reply := <-replies:
			return &reply
		default:
			return &fakevendor.Reply{Text: "Fixture complete."}
		}
	}})
	defer fake.Close()
	cmd := exec.CommandContext(ctx, bin)
	cmd.Dir = work
	cmd.Env = []string{"PATH=/usr/bin:/bin:/usr/sbin:/sbin", "HOME=" + work, "TMPDIR=" + work, "LLM_VENDOR=anthropic", "LLM_MODEL=fake-model", "LLM_BASE_URL=" + fake.URL(), "LLM_API_KEY=fixture-only", "EN_DISABLE_STREAMING=1", "EN_STARTUP_SKILLS=code-tools", "EN_DISABLE_SAVE=1", "EN_DATA_DIR=" + filepath.Join(work, "data"), "LLM_REQUIRE_TOOL_EXPLANATION=false"}
	output := hintOutput{ended: make(chan struct{}, 16)}
	cmd.Stdout = &output
	cmd.Stderr = &output
	input, err := cmd.StdinPipe()
	if err != nil {
		return err
	}
	if err = cmd.Start(); err != nil {
		return err
	}
	done := make(chan struct{})
	var exitError error
	go func() { exitError = cmd.Wait(); close(done) }()
	// Always release real work before asking the CLI to drain. A watchdog only
	// terminates a broken submission; it is not the behavior being graded.
	defer func() {
		_ = os.WriteFile(filepath.Join(work, "release"), nil, 0600)
		_ = input.Close()
		select {
		case <-done:
		case <-time.After(3 * time.Second):
			cancel()
			<-done
		}
		if result == nil && exitError != nil {
			result = fmt.Errorf("CLI failed while draining: %w", exitError)
		}
		if result != nil {
			result = fmt.Errorf("%w; CLI output: %s", result, output.String())
		}
	}()
	send := func(kind, text string) error {
		encoded, _ := json.Marshal(map[string]string{"kind": kind, "text": text})
		_, err := fmt.Fprintln(input, string(encoded))
		return err
	}
	next := func() (hintRequest, error) {
		select {
		case request := <-requests:
			return request, nil
		case <-done:
			return hintRequest{}, fmt.Errorf("CLI exited before expected request: %v", exitError)
		case <-ctx.Done():
			return hintRequest{}, fmt.Errorf("timed out awaiting provider request")
		}
	}
	waitFile := func(name string) (time.Time, error) {
		for {
			if _, err := os.Stat(filepath.Join(work, name)); err == nil {
				return time.Now(), nil
			}
			select {
			case request := <-requests:
				// A model request may race the marker poll. Preserve it if the
				// effect is present; otherwise the batch advanced without it.
				if _, err := os.Stat(filepath.Join(work, name)); err == nil {
					requests <- request
					return time.Now(), nil
				}
				return time.Time{}, fmt.Errorf("provider advanced before expected effect %s (discarded or reordered work)", name)
			case <-done:
				return time.Time{}, fmt.Errorf("CLI exited before %s: %v", name, exitError)
			case <-ctx.Done():
				return time.Time{}, fmt.Errorf("timed out awaiting effect %s", name)
			case <-time.After(10 * time.Millisecond):
			}
		}
	}
	if err := send("prompt", "Run the fixture batch."); err != nil {
		return err
	}
	if _, err := next(); err != nil {
		return err
	}
	start, err := waitFile("started1")
	if err != nil {
		return err
	}
	pause := 600 * time.Millisecond
	if aged {
		pause = 3500 * time.Millisecond
	}
	time.Sleep(pause)
	if interrupt {
		encoded, _ := json.Marshal(map[string]string{"kind": "interrupt", "mode": strings.TrimPrefix(mode, "interrupt-")})
		if _, err := fmt.Fprintln(input, string(encoded)); err != nil {
			return err
		}
		select {
		case <-output.ended:
		case <-ctx.Done():
			return fmt.Errorf("interrupt did not end turn")
		}
		elapsed := time.Since(start)
		if mode == "interrupt-now" && elapsed > 1800*time.Millisecond {
			return fmt.Errorf("immediate interrupt did not end current observation promptly: %s", elapsed)
		}
		if mode == "interrupt-after_tool" && (elapsed < 18*time.Second || elapsed > 23*time.Second) {
			return fmt.Errorf("after_tool did not retain current twenty-second observation: %s", elapsed)
		}
		for _, name := range []string{"started2", "started3", "written.txt"} {
			if _, err := os.Stat(filepath.Join(work, name)); !os.IsNotExist(err) {
				return fmt.Errorf("interrupted tail had effect %s: %v", name, err)
			}
		}
		if err := os.WriteFile(filepath.Join(work, "release"), nil, 0600); err != nil {
			return err
		}
		if _, err := waitFile("finished1"); err != nil {
			return err
		}
		// A fresh prompt exposes the interrupted batch's durable result pairing.
		if err := send("prompt", "Inspect the interrupted batch and finish."); err != nil {
			return err
		}
		request, err := next()
		if err != nil {
			return err
		}
		text, err := hintToolText(request.body, "held-1")
		if err != nil {
			return err
		}
		match := hintHandle.FindStringSubmatch(text)
		if match == nil {
			return fmt.Errorf("interrupt lost original running handle: %s", text)
		}
		for _, id := range []string{"held-2", "held-3", "pending-write"} {
			text, err := hintToolText(request.body, id)
			if err == nil && !strings.Contains(text, "not executed") {
				return fmt.Errorf("interrupt tail %s lacks not-executed result: %s", id, text)
			}
			if err != nil && !strings.Contains(err.Error(), "not executed") {
				return err
			}
		}
		// Wait for the inspection prompt to end before admitting another prompt.
		select {
		case <-output.ended:
		case <-ctx.Done():
			return fmt.Errorf("recovery prompt did not finish")
		}
		handle, _ := strconv.Atoi(match[1])
		replies <- fakevendor.Reply{Tools: []fakevendor.ToolCall{hintCall("interrupt-finish", "wait_for_job", map[string]any{"handle": handle, "ai_callback_delay": 1})}}
		if err := send("prompt", "Retrieve the original job completion."); err != nil {
			return err
		}
		if _, err := next(); err != nil {
			return err
		}
		completion, err := next()
		if err != nil {
			return err
		}
		text, err = hintToolText(completion.body, "interrupt-finish")
		if err != nil || !strings.Contains(text, "FINISHED_1") {
			return fmt.Errorf("interrupt lost job completion: %q (%v)", text, err)
		}
		return nil
	}
	hinted := time.Now()
	// Queue the next response before sending the hint: it starts a fresh, longer
	// observation whose delay must not inherit the previous batch's cap.
	replies <- fakevendor.Reply{Tools: []fakevendor.ToolCall{hintCall("fresh-wait", "run_command", map[string]any{"command": "sleep 4.2; printf FRESH_WAIT_COMPLETE", "ai_callback_delay": 8})}}
	if err := send("hint", "HINT_BATCH_MARKER"); err != nil {
		return err
	}
	if !aged {
		time.Sleep(1200 * time.Millisecond)
		if err := send("hint", "HINT_REPEAT_MARKER"); err != nil {
			return err
		}
	}
	secondStart, err := waitFile("started2")
	if err != nil {
		return err
	}
	elapsed := secondStart.Sub(start)
	if !aged && (elapsed < 2500*time.Millisecond || elapsed > 3900*time.Millisecond) {
		return fmt.Errorf("early observation lasted %s; expected original three-second deadline (not instant wake or reset)", elapsed)
	}
	if aged && secondStart.Sub(hinted) > 1200*time.Millisecond {
		return fmt.Errorf("aged observation did not return promptly: %s", secondStart.Sub(hinted))
	}
	shortStart, err := waitFile("started3")
	if err != nil {
		return err
	}
	batch, err := next()
	if err != nil {
		return err
	}
	shortElapsed := batch.at.Sub(shortStart)
	if shortElapsed < 300*time.Millisecond || shortElapsed > 1400*time.Millisecond {
		return fmt.Errorf("short observation plus fast write lasted %s; expected half-second wait and prompt write", shortElapsed)
	}
	remaining := shortStart.Sub(secondStart)
	if remaining < 2500*time.Millisecond || remaining > 4*time.Second {
		return fmt.Errorf("remaining observation lasted %s, expected three-second ceiling", remaining)
	}
	if !bytes.Contains(batch.body, []byte("HINT_BATCH_MARKER")) {
		return fmt.Errorf("hint absent from next request")
	}
	if !aged && !bytes.Contains(batch.body, []byte("HINT_REPEAT_MARKER")) {
		return fmt.Errorf("repeated hint absent from next request")
	}
	handles := make([]int, 0, 3)
	for _, id := range []string{"held-1", "held-2", "held-3"} {
		text, err := hintToolText(batch.body, id)
		if err != nil {
			return err
		}
		match := hintHandle.FindStringSubmatch(text)
		if match == nil {
			return fmt.Errorf("%s lacks truthful running handle: %s", id, text)
		}
		handle, _ := strconv.Atoi(match[1])
		handles = append(handles, handle)
	}
	if handles[0] == handles[1] || handles[0] == handles[2] || handles[1] == handles[2] {
		return fmt.Errorf("distinct jobs reused handle %d", handles[0])
	}
	if _, err := hintToolText(batch.body, "pending-write"); err != nil {
		return err
	}
	data, err := os.ReadFile(filepath.Join(work, "written.txt"))
	if err != nil || string(data) != "PENDING_WRITE_SURVIVED" {
		return fmt.Errorf("pending write missing or corrupt: %v", err)
	}
	// Release the original jobs, then have the provider wait on their same IDs.
	if err := os.WriteFile(filepath.Join(work, "release"), nil, 0600); err != nil {
		return err
	}
	for i := 1; i <= 3; i++ {
		if _, err := waitFile(fmt.Sprintf("finished%d", i)); err != nil {
			return err
		}
		data, err := os.ReadFile(filepath.Join(work, fmt.Sprintf("started%d", i)))
		if err != nil || string(data) != "started\n" {
			return fmt.Errorf("job %d did not execute exactly once: %q (%v)", i, data, err)
		}
	}
	waits := []fakevendor.ToolCall{}
	for i, handle := range handles {
		waits = append(waits, hintCall(fmt.Sprintf("finish-%d", i+1), "wait_for_job", map[string]any{"handle": handle, "ai_callback_delay": 1}))
	}
	replies <- fakevendor.Reply{Tools: waits}
	fresh, err := next()
	if err != nil {
		return err
	}
	text, err := hintToolText(fresh.body, "fresh-wait")
	if err != nil || !strings.Contains(text, "FRESH_WAIT_COMPLETE") || hintHandle.MatchString(text) {
		return fmt.Errorf("fresh observation inherited hint cap or failed: %q (%v)", text, err)
	}
	final, err := next()
	if err != nil {
		return err
	}
	for i := 1; i <= 3; i++ {
		text, err := hintToolText(final.body, fmt.Sprintf("finish-%d", i))
		if err != nil || !strings.Contains(text, fmt.Sprintf("FINISHED_%d", i)) {
			return fmt.Errorf("original job completion lost: %q (%v)", text, err)
		}
	}
	return nil
}

// hintOutput is written by os/exec's single shared stdout/stderr copier. The
// retained bytes are read only after cmd.Wait has joined that copier.
type hintOutput struct {
	bytes.Buffer
	pending []byte
	ended   chan struct{}
}

func (output *hintOutput) Write(data []byte) (int, error) {
	output.Buffer.Write(data)
	output.pending = append(output.pending, data...)
	for {
		newline := bytes.IndexByte(output.pending, '\n')
		if newline < 0 {
			break
		}
		var record struct{ Observation string }
		if json.Unmarshal(output.pending[:newline], &record) == nil && record.Observation == "turn_ended" {
			select {
			case output.ended <- struct{}{}:
			default:
			}
		}
		output.pending = output.pending[newline+1:]
	}
	return len(data), nil
}

func (output *hintOutput) ReadFrom(reader io.Reader) (int64, error) {
	return io.Copy(struct{ io.Writer }{output}, reader)
}
