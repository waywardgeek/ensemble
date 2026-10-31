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

// Chapter 17 launch machinery.
//
// Chapter 16 taught this harness that not every request is a turn. Chapter 17
// makes that lesson load-bearing, because recall adds a SECOND kind of
// out-of-turn request — the judge — and it fires BEFORE the turn request
// rather than after it.
//
// That ordering is what breaks the obvious approach. ch16 waits for a total
// request count, so if you write expect[i] = i+1 thinking "one turn per
// prompt", the judge call satisfies the wait on its own and the harness sails
// on to the next prompt while the turn it was supposed to wait for has not
// even been sent yet. The failures that produces are spectacular and look
// like everything except an off-by-one in a counter.
//
// So this harness does not count requests. It counts TURNS, by asking what
// each recorded request actually is:
//
//	tools == 0                   -> a judge call (see llm.Judge)
//	exactly one tool, "submit"   -> a compressor call (chapter 16)
//	otherwise                    -> a turn
//
// Classifying by shape rather than by position means the assertions stay true
// no matter how many judge calls a student's implementation decides to make.

// ch17Opts describes one launch.
type ch17Opts struct {
	dir     string
	model   string
	prompts []string
	// turns[i] is how many TURN requests the vendor must have recorded once
	// prompt i's turn is over. Judge and compressor requests are excluded, so
	// this is normally just i+1.
	turns   []int
	replies []fakevendor.Reply
	// route answers a request by its shape, ahead of the positional script.
	route func(body []byte) *fakevendor.Reply
}

type ch17Out struct {
	reqs  []fakevendor.Recorded
	log   []ch17Event
	fatal string
}

func ch17Launch(bin, skillsDir, guiDir string, o ch17Opts) ch17Out {
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
		return ch17Out{fatal: fmt.Sprintf("stdin pipe: %v", err)}
	}
	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		return ch17Out{fatal: fmt.Sprintf("start: %v", err)}
	}
	done := make(chan error, 1)
	go func() { done <- cmd.Wait() }()
	stop := func() {
		cmd.Process.Kill()
		<-done
	}

	if !waitForHTTP("http://localhost:"+port+"/", 10*time.Second) {
		stop()
		return ch17Out{fatal: "GUI server did not start"}
	}

	for i, p := range o.prompts {
		line, _ := json.Marshal(map[string]string{"kind": "prompt", "text": p})
		fmt.Fprintln(stdin, string(line))
		want := i + 1
		if i < len(o.turns) {
			want = o.turns[i]
		}
		// 40s rather than ch16's 25s: a judge that is being made to time out
		// burns most of a 15s budget before the turn request is even sent.
		if !ch17WaitTurns(srv, want, 40*time.Second) {
			stop()
			return ch17Out{reqs: srv.Requests(), log: ch17ReadLog(o.dir), fatal: fmt.Sprintf(
				"after prompt %d the vendor saw %d turn requests (%d total), expected %d turns",
				i+1, ch17CountTurns(srv.Requests()), len(srv.Requests()), want)}
		}
		// The last reply still has to land and end the turn.
		time.Sleep(500 * time.Millisecond)
	}

	stdin.Close()
	select {
	case <-done:
	case <-time.After(15 * time.Second):
		cmd.Process.Kill()
		<-done
		return ch17Out{reqs: srv.Requests(), log: ch17ReadLog(o.dir), fatal: "agent did not exit within 15s of stdin closing"}
	}
	return ch17Out{reqs: srv.Requests(), log: ch17ReadLog(o.dir)}
}

func ch17WaitTurns(srv *fakevendor.Server, n int, d time.Duration) bool {
	deadline := time.Now().Add(d)
	for time.Now().Before(deadline) {
		if ch17CountTurns(srv.Requests()) >= n {
			return true
		}
		time.Sleep(50 * time.Millisecond)
	}
	return ch17CountTurns(srv.Requests()) >= n
}

// ---------------------------------------------------------------------------
// Reading the wire
// ---------------------------------------------------------------------------

// ch17Req is the slice of an Anthropic request this chapter cares about.
type ch17Req struct {
	Model    string            `json:"model"`
	System   json.RawMessage   `json:"system"`
	Tools    []json.RawMessage `json:"tools"`
	Messages []ch17Msg         `json:"messages"`
}

type ch17Msg struct {
	Role    string          `json:"role"`
	Content json.RawMessage `json:"content"`
}

// Text flattens a message's content, which may be a bare string or a list of
// typed blocks.
func (m ch17Msg) Text() string {
	var s string
	if json.Unmarshal(m.Content, &s) == nil {
		return s
	}
	var blocks []struct {
		Type    string          `json:"type"`
		Text    string          `json:"text"`
		Content json.RawMessage `json:"content"`
	}
	if json.Unmarshal(m.Content, &blocks) != nil {
		return string(m.Content)
	}
	var sb strings.Builder
	for _, b := range blocks {
		sb.WriteString(b.Text)
		if len(b.Content) > 0 {
			sb.Write(b.Content)
		}
		sb.WriteString("\n")
	}
	return sb.String()
}

func ch17Parse(body []byte) ch17Req {
	var r ch17Req
	_ = json.Unmarshal(body, &r)
	return r
}

// ch17IsJudge reports whether a request is a snippet-judge call.
//
// The signature is structural but it is not a code-shape assertion: it is a
// statement about what the model was ASKED. A judge call carries no tools,
// because a component whose only job is to answer "which of these are
// relevant" has nothing to act on. A turn always carries the agent's toolbelt,
// and chapter 16's compressor carries exactly one tool named submit.
func ch17IsJudge(body []byte) bool {
	r := ch17Parse(body)
	return len(r.Tools) == 0 && len(r.Messages) > 0
}

func ch17IsCompressor(body []byte) bool {
	r := ch17Parse(body)
	if len(r.Tools) != 1 {
		return false
	}
	var td struct {
		Name string `json:"name"`
	}
	_ = json.Unmarshal(r.Tools[0], &td)
	return td.Name == "submit"
}

func ch17IsTurn(body []byte) bool {
	return !ch17IsJudge(body) && !ch17IsCompressor(body)
}

func ch17CountTurns(reqs []fakevendor.Recorded) int {
	n := 0
	for _, r := range reqs {
		if ch17IsTurn(r.Body) {
			n++
		}
	}
	return n
}

func ch17Judges(reqs []fakevendor.Recorded) []fakevendor.Recorded {
	var out []fakevendor.Recorded
	for _, r := range reqs {
		if ch17IsJudge(r.Body) {
			out = append(out, r)
		}
	}
	return out
}

func ch17Turns(reqs []fakevendor.Recorded) []fakevendor.Recorded {
	var out []fakevendor.Recorded
	for _, r := range reqs {
		if ch17IsTurn(r.Body) {
			out = append(out, r)
		}
	}
	return out
}

// ch17LastTurn returns the final turn request body, which is the one that has
// seen the most context.
func ch17LastTurn(reqs []fakevendor.Recorded) (fakevendor.Recorded, bool) {
	turns := ch17Turns(reqs)
	if len(turns) == 0 {
		return fakevendor.Recorded{}, false
	}
	return turns[len(turns)-1], true
}

// ch17AllText returns every message's text in one string, for "did the model
// ever see this token" questions.
func ch17AllText(body []byte) string {
	r := ch17Parse(body)
	var sb strings.Builder
	sb.Write(r.System)
	sb.WriteString("\n")
	for _, m := range r.Messages {
		sb.WriteString(m.Text())
		sb.WriteString("\n")
	}
	return sb.String()
}

// ---------------------------------------------------------------------------
// Reading the event log
// ---------------------------------------------------------------------------

type ch17Event struct {
	Type   string `json:"type"`
	Recall *struct {
		Parts json.RawMessage `json:"parts"`
	} `json:"recall_attached"`
}

func ch17ReadLog(dir string) []ch17Event {
	raw, err := os.ReadFile(filepath.Join(dir, "events.jsonl"))
	if err != nil {
		return nil
	}
	var out []ch17Event
	for _, line := range strings.Split(string(raw), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		var e ch17Event
		if json.Unmarshal([]byte(line), &e) != nil {
			continue
		}
		out = append(out, e)
	}
	return out
}
