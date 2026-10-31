package grade

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

// Chapter 16 launch machinery.
//
// The shape is chapter 15's, with one difference that drives everything
// else: not every request the agent makes is a turn. A compressor asks
// the model to fold finished work into a memory, and it does so in the
// middle of a turn that the grader started for another reason. So the
// script is not positional here. Replies are chosen by what the request
// IS, through fakevendor's Route, and the positional list is left for
// the main conversation.

// ch16Opts describes one launch.
type ch16Opts struct {
	dir     string
	model   string
	prompts []string
	// expect[i] is how many requests the vendor must have recorded once
	// prompt i's turn is over. Compressor requests count, so this is
	// usually larger than the number of prompts.
	expect  []int
	replies []fakevendor.Reply
	// route answers a request by its shape, ahead of the script.
	route func(body []byte) *fakevendor.Reply
	// settings is written to settings.json before launch. Empty writes
	// nothing, which leaves the student on their own defaults.
	settings string
	// kill ends the run with SIGKILL instead of closing stdin, so the
	// agent gets no chance to snapshot or to finish a compaction.
	kill bool
}

type ch16Out struct {
	reqs  []fakevendor.Recorded
	fatal string
}

func ch16Launch(bin, skillsDir, guiDir string, o ch16Opts) ch16Out {
	if o.settings != "" {
		if err := os.WriteFile(filepath.Join(o.dir, "settings.json"), []byte(o.settings), 0o644); err != nil {
			return ch16Out{fatal: fmt.Sprintf("write settings.json: %v", err)}
		}
	}

	srv := fakevendor.NewWithOptions(o.replies, fakevendor.Options{Route: o.route})
	defer srv.Close()

	port := freePort()
	cmd := exec.Command(bin, "--port", port, "--gui-dir", guiDir)
	cmd.Dir = o.dir
	cmd.Env = append(os.Environ(),
		"LLM_BASE_URL="+srv.URL(),
		"LLM_MODEL="+o.model,
		"LLM_VENDOR=anthropic",
		"LLM_API_KEY=test-key",
		"EN_SKILLS_DIR="+skillsDir,
		"EN_PRIMARY_SKILL=base",
		"CH02_LOG="+filepath.Join(o.dir, "events.jsonl"),
	)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return ch16Out{fatal: fmt.Sprintf("stdin pipe: %v", err)}
	}
	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		return ch16Out{fatal: fmt.Sprintf("start: %v", err)}
	}
	done := make(chan error, 1)
	go func() { done <- cmd.Wait() }()
	stop := func() {
		cmd.Process.Kill()
		<-done
	}

	if !waitForHTTP("http://localhost:"+port+"/", 10*time.Second) {
		stop()
		return ch16Out{fatal: "GUI server did not start"}
	}

	for i, p := range o.prompts {
		line, _ := json.Marshal(map[string]string{"kind": "prompt", "text": p})
		fmt.Fprintln(stdin, string(line))
		want := i + 1
		if i < len(o.expect) {
			want = o.expect[i]
		}
		if !ch16WaitRequests(srv, want, 25*time.Second) {
			stop()
			return ch16Out{reqs: srv.Requests(), fatal: fmt.Sprintf(
				"after prompt %d the vendor saw %d requests, expected %d",
				i+1, len(srv.Requests()), want)}
		}
		// The last reply still has to land and end the turn.
		time.Sleep(500 * time.Millisecond)
	}

	if o.kill {
		stop()
		stdin.Close()
		return ch16Out{reqs: srv.Requests()}
	}
	stdin.Close()
	select {
	case <-done:
	case <-time.After(15 * time.Second):
		cmd.Process.Kill()
		<-done
		return ch16Out{reqs: srv.Requests(), fatal: "agent did not exit within 15s of stdin closing"}
	}
	return ch16Out{reqs: srv.Requests()}
}

func ch16WaitRequests(srv *fakevendor.Server, n int, d time.Duration) bool {
	deadline := time.Now().Add(d)
	for time.Now().Before(deadline) {
		if len(srv.Requests()) >= n {
			return true
		}
		time.Sleep(50 * time.Millisecond)
	}
	return len(srv.Requests()) >= n
}

// ----------------------------------------------------------------
// Reading requests
// ----------------------------------------------------------------

// ch16Req is the part of an Anthropic Messages request these checks read.
type ch16Req struct {
	System   json.RawMessage `json:"system"`
	Tools    []ch16Tool      `json:"tools"`
	Messages []ch16Msg       `json:"messages"`
}

type ch16Tool struct {
	Name string `json:"name"`
}

type ch16Msg struct {
	Role    string          `json:"role"`
	Content json.RawMessage `json:"content"`
}

func ch16Parse(body []byte) (ch16Req, error) {
	var r ch16Req
	err := json.Unmarshal(body, &r)
	return r, err
}

// ch16IsCompressor reports whether a request is a compressor asking for a
// memory rather than a turn of the conversation. The signature is the
// tool set: a compressor is offered exactly one tool, and it is submit.
//
// This is the whole reason the grader can stay behavioural. It never asks
// the student which function made the call; it looks at what went on the
// wire and recognises a request that can only be a compressor.
func ch16IsCompressor(body []byte) bool {
	r, err := ch16Parse(body)
	if err != nil {
		return false
	}
	return len(r.Tools) == 1 && r.Tools[0].Name == "submit"
}

// ch16ToolNames lists the tools offered in a request.
func ch16ToolNames(body []byte) []string {
	r, err := ch16Parse(body)
	if err != nil {
		return nil
	}
	out := make([]string, 0, len(r.Tools))
	for _, t := range r.Tools {
		out = append(out, t.Name)
	}
	return out
}

// ch16Turns returns only the requests that are turns, dropping compressor
// traffic. Checks about what the agent can see should read these.
func ch16Turns(reqs []fakevendor.Recorded) []fakevendor.Recorded {
	var out []fakevendor.Recorded
	for _, r := range reqs {
		if !ch16IsCompressor(r.Body) {
			out = append(out, r)
		}
	}
	return out
}

// ch16Compressors returns only the compressor requests.
func ch16Compressors(reqs []fakevendor.Recorded) []fakevendor.Recorded {
	var out []fakevendor.Recorded
	for _, r := range reqs {
		if ch16IsCompressor(r.Body) {
			out = append(out, r)
		}
	}
	return out
}

// ch16Last returns the last request, or false when there is none.
func ch16Last(reqs []fakevendor.Recorded) (fakevendor.Recorded, bool) {
	if len(reqs) == 0 {
		return fakevendor.Recorded{}, false
	}
	return reqs[len(reqs)-1], true
}

// ch16Conversation is everything in a request that is not the system
// prompt: the messages, flattened to text. Memory has to arrive here
// rather than in the system prompt, so the two are kept apart.
func ch16Conversation(body []byte) string {
	r, err := ch16Parse(body)
	if err != nil {
		return ""
	}
	var b strings.Builder
	for _, m := range r.Messages {
		b.Write(m.Content)
		b.WriteByte('\n')
	}
	return b.String()
}

// ch16System is the system prompt of a request, flattened to text.
func ch16System(body []byte) string {
	r, err := ch16Parse(body)
	if err != nil {
		return ""
	}
	return string(r.System)
}

// ch16Whole is system plus conversation, for checks that only ask whether
// a string reached the model at all.
func ch16Whole(body []byte) string {
	return ch16System(body) + "\n" + ch16Conversation(body)
}
