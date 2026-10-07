package grade

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/gorilla/websocket"

	"github.com/waywardgeek/ensemble/internal/fakevendor"
)

// The behavioural half of chapter 22's grader: it builds the student's
// binary, drives a real conversation against a fake vendor, switches model
// mid-session over the GUI socket, and then reads what the agent_status
// tool reported back.
//
// Why drive the binary rather than call the library: chapter 21 shipped a
// feature that every structural check passed and that was dead on arrival,
// because the skill file never listed its tool. A grader that calls the
// library directly cannot see that class of fault. Every isolation a test
// buys is a thing it stops observing.
//
// The two models below are PRODUCTION rows, not course rows. That is
// deliberate and slightly unusual: the model table warns that a grader
// naming a row the frozen solution snapshots have never heard of loses
// marks silently. These two have been in the table for many chapters, they
// share a vendor -- so switching between them does not change the wire
// shape -- and, uniquely among same-vendor pairs, they are both PRICED.
// The course rows carry no price sheet at all, and an unpriced pair cannot
// demonstrate a per-model cost bug.
const (
	// claude-opus-4-6: Input 5, CacheRead 0.5, Output 25 per million.
	ch22ModelDear = "claude-opus-4-6"
	// claude-sonnet-5: Input 3, CacheRead 0.3, Output 15 per million.
	ch22ModelCheap = "claude-sonnet-5"
)

// Prices, in dollars per million tokens, copied from the model table.
//
// The grader hardcodes these rather than reading the student's table, and
// it has no choice: internal/grade lives in the root module and the agent's
// common package is internal to the agent module, so it cannot be imported.
// The coupling is a feature here -- the expected cost is computed from an
// independent copy of the price sheet, so a student cannot make a wrong
// figure agree with itself.
const (
	ch22DearIn, ch22DearOut, ch22DearCacheRead    = 5.0, 25.0, 0.5
	ch22CheapIn, ch22CheapOut, ch22CheapCacheRead = 3.0, 15.0, 0.3
)

// ch22Usage is one scripted response's token tally.
type ch22Usage struct{ in, out, cacheRead, cacheWrite int }

// The scripted session. Token counts are large enough that cost, which is
// quoted to four decimal places, lands well clear of $0.0000 under both
// the correct and the buggy formula -- a realistic hundred-token fake
// conversation prices out as zero either way, which would make the whole
// check vacuous.
//
// They are NOT larger than that, and the ceiling matters as much as the
// floor. An earlier draft used millions of tokens, and the driven agent
// duly decided its context was full and withdrew every tool but the
// checkpoint one. The checks still passed, but only by luck: the grader
// had quietly started testing the student's compaction policy instead of
// their status tool. A fake usage figure is still an input to the thing
// under test.
var (
	ch22TurnOne   = ch22Usage{in: 10_000, out: 1_000}                // first turn
	ch22TurnTwo   = ch22Usage{in: 2_000, out: 200, cacheRead: 3_000} // after the switch is asked for
	ch22TurnThree = ch22Usage{in: 1_000, out: 100}                   // the turn that calls the tool
)

// ch22Expected returns the figures agent_status must report, computed here
// rather than taken from the agent.
//
// The per-turn model is read off the wire rather than assumed. Switching
// model is asynchronous -- the settings frame reaches the actor through its
// mailbox, and a prompt written to stdin a moment later can and does
// overtake it, so the turn after the switch is sometimes still billed to
// the old model. That race belongs to the harness, not to the student, and
// an expectation built on "the switch landed where I asked for it" would
// fail honest work intermittently. Taking the model from each recorded
// request removes the race completely: however the turns were attributed,
// the grader knows exactly how they were attributed, and the cost it
// predicts is still computed independently of the agent's own arithmetic.
type ch22Expect struct {
	model                                string
	lastIn, lastOut, lastRead, lastWrite int
	totIn, totOut, totRead, totWrite     int
	cacheRatePct                         float64
	costCorrect, costIfOneRate           float64
	models                               []string // distinct models billed
}

func ch22Expected(perRequest []string) ch22Expect {
	// agent_status runs as a tool DURING the third turn, so the usage of
	// the first three responses has been recorded when it reads the
	// counters -- and no more. The fourth response closes the turn after
	// the tool has already answered.
	turns := []ch22Usage{ch22TurnOne, ch22TurnTwo, ch22TurnThree}

	e := ch22Expect{}
	type rates struct{ in, out, read float64 }
	price := map[string]rates{
		ch22ModelDear:  {ch22DearIn, ch22DearOut, ch22DearCacheRead},
		ch22ModelCheap: {ch22CheapIn, ch22CheapOut, ch22CheapCacheRead},
	}
	seen := map[string]bool{}

	for i, u := range turns {
		if i >= len(perRequest) {
			break
		}
		m := perRequest[i]
		if !seen[m] {
			seen[m] = true
			e.models = append(e.models, m)
		}
		e.totIn += u.in
		e.totOut += u.out
		e.totRead += u.cacheRead
		e.totWrite += u.cacheWrite
		p := price[m]
		e.costCorrect += perMillion(u.in, p.in) + perMillion(u.out, p.out) +
			perMillion(u.cacheRead, p.read)
		e.lastIn, e.lastOut = u.in, u.out
		e.lastRead, e.lastWrite = u.cacheRead, u.cacheWrite
		e.model = m
	}

	if in := e.totIn + e.totRead; in > 0 {
		e.cacheRatePct = 100 * float64(e.totRead) / float64(in)
	}

	// The bug: the whole session priced at whichever model happens to be
	// current. This is what shipped, and telling it apart from the correct
	// answer is the entire point of the check.
	p := price[e.model]
	e.costIfOneRate = perMillion(e.totIn, p.in) + perMillion(e.totOut, p.out) +
		perMillion(e.totRead, p.read)
	return e
}

func perMillion(tokens int, rate float64) float64 {
	return float64(tokens) * rate / 1_000_000
}

// ch22Out is what a driven session yields.
type ch22Out struct {
	status   string   // the text the agent_status tool returned
	replies  []string // assistant turns, in order
	models   []string // the model named in each recorded request, in order
	requests int
	fatal    string
}

// ch22Drive builds and runs the student's agent through the scripted
// session. Each scenario gets its own work directory: chapter 21 lost time
// to a shared one, where save.json let a run resume the previous
// scenario's conversation.
func ch22Drive(dir string) ch22Out {
	bin, cleanup, err := Build(dir)
	if err != nil {
		return ch22Out{fatal: fmt.Sprintf("build failed: %v", err)}
	}
	defer cleanup()

	// Three scripted responses. The third asks for agent_status; the
	// fourth closes the turn once the tool has answered.
	replies := []fakevendor.Reply{
		{Text: "first", Usage: usageOf(ch22TurnOne)},
		{Text: "second", Usage: usageOf(ch22TurnTwo)},
		{ToolName: "agent_status", ToolArgs: `{}`, Usage: usageOf(ch22TurnThree)},
		{Text: "done", Usage: fakevendor.Canonical{}},
	}
	srv := fakevendor.New(replies)
	defer srv.Close()

	work, err := os.MkdirTemp("", "ch22-run-")
	if err != nil {
		return ch22Out{fatal: fmt.Sprintf("tempdir: %v", err)}
	}
	defer os.RemoveAll(work)

	skills := ch22Skills(work)
	port := freePort()

	cmd := exec.Command(bin, "--port", port)
	cmd.Dir = work
	cmd.Env = append(os.Environ(),
		"LLM_BASE_URL="+srv.URL(),
		"LLM_MODEL="+ch22ModelDear,
		"LLM_API_KEY=test",
		"EN_SKILL_DIR="+skills,
		"EN_PRIMARY_SKILL=base",
	)
	stdin, _ := cmd.StdinPipe()
	stdout, _ := cmd.StdoutPipe()
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		return ch22Out{fatal: fmt.Sprintf("start: %v", err)}
	}
	defer func() {
		_ = stdin.Close()
		_ = cmd.Process.Kill()
		_ = cmd.Wait()
	}()

	out := ch22Out{}
	lines := bufio.NewScanner(stdout)
	lines.Buffer(make([]byte, 1<<20), 1<<20)

	say := func(text string) bool {
		msg, _ := json.Marshal(map[string]string{"kind": "prompt", "text": text})
		if _, err := stdin.Write(append(msg, '\n')); err != nil {
			out.fatal = fmt.Sprintf("write prompt: %v", err)
			return false
		}
		for lines.Scan() {
			var ev struct {
				Assistant string `json:"assistant"`
			}
			if err := json.Unmarshal(lines.Bytes(), &ev); err != nil {
				continue
			}
			if ev.Assistant != "" {
				out.replies = append(out.replies, ev.Assistant)
				return true
			}
		}
		out.fatal = "agent produced no assistant reply"
		return false
	}

	if !ch22WaitPort(port, 10*time.Second) {
		return ch22Out{fatal: "agent never opened its GUI port"}
	}
	if !say("one") {
		return out
	}
	if err := ch22SetModel(port, ch22ModelCheap); err != nil {
		out.fatal = fmt.Sprintf("switching model over the GUI socket: %v", err)
		return out
	}
	if !say("two") {
		return out
	}
	if !say("three") {
		return out
	}

	reqs := srv.Requests()
	out.requests = len(reqs)
	out.models = ch22RequestModels(reqs)
	out.status = ch22FindToolResult(reqs)
	return out
}

// ch22RequestModels reads the model named in each recorded request, in
// order. This is the ground truth for which model a turn was billed to.
func ch22RequestModels(reqs []fakevendor.Recorded) []string {
	out := make([]string, 0, len(reqs))
	for _, r := range reqs {
		var body struct {
			Model string `json:"model"`
		}
		_ = json.Unmarshal(r.Body, &body)
		out = append(out, body.Model)
	}
	return out
}

func usageOf(u ch22Usage) fakevendor.Canonical {
	return fakevendor.Canonical{
		Input: u.in, Output: u.out,
		CacheRead: u.cacheRead, CacheWrite: u.cacheWrite,
	}
}

// ch22SetModel switches the running agent's model the only way anything
// can: the GUI's websocket. There is no stdin verb for it and the settings
// file is read once at startup, so this is not a convenience -- it is the
// single available seam, and driving it is also what proves the switch
// works end to end.
func ch22SetModel(port string, model string) error {
	url := fmt.Sprintf("ws://127.0.0.1:%s/ws", port)
	c, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		return fmt.Errorf("dial %s: %w", url, err)
	}
	defer c.Close()

	// The GUI socket is client-pull: a freshly accepted connection is sent
	// nothing at all until it asks. Without this the agent looks hung, and
	// an earlier draft of this harness timed out here waiting for a state
	// frame that was never going to arrive unprompted.
	if err := c.WriteJSON(map[string]any{"type": "subscribe"}); err != nil {
		return fmt.Errorf("subscribe: %w", err)
	}

	frame := map[string]any{
		"type":     "update_settings",
		"settings": json.RawMessage(fmt.Sprintf(`{"model":%q}`, model)),
	}
	if err := c.WriteJSON(frame); err != nil {
		return fmt.Errorf("send update_settings: %w", err)
	}

	// Wait for the agent to echo the new model back before returning, so
	// the next prompt cannot race the switch.
	_ = c.SetReadDeadline(time.Now().Add(10 * time.Second))
	for {
		_, data, err := c.ReadMessage()
		if err != nil {
			return fmt.Errorf("never saw the model change confirmed: %w", err)
		}
		if strings.Contains(string(data), model) {
			return nil
		}
	}
}

// ch22FindToolResult digs the agent_status output out of the conversation
// the agent sent back to the vendor. A tool result is returned to the model
// in the NEXT request, so the transcript the fake vendor recorded is a
// faithful record of what the tool actually produced -- no instrumentation
// of the student's code required.
func ch22FindToolResult(reqs []fakevendor.Recorded) string {
	for i := len(reqs) - 1; i >= 0; i-- {
		var body any
		if err := json.Unmarshal(reqs[i].Body, &body); err != nil {
			continue
		}
		if found := findStatusText(body); found != "" {
			return found
		}
	}
	return ""
}

// findStatusText walks decoded JSON looking for a string that carries the
// shape of a status report. It matches on the REPORTED FIELDS rather than
// on any particular sentence, so a student may word the report however
// they like.
func findStatusText(n any) string {
	switch v := n.(type) {
	case string:
		if strings.Contains(v, "model:") && strings.Contains(v, "session") {
			return v
		}
	case []any:
		for _, e := range v {
			if s := findStatusText(e); s != "" {
				return s
			}
		}
	case map[string]any:
		for _, e := range v {
			if s := findStatusText(e); s != "" {
				return s
			}
		}
	}
	return ""
}

// ch22Skills writes the skill file the run uses. Supplying it makes the run
// deterministic, but it also makes the grader blind to whether the agent's
// OWN shipped skill lists the tool -- which is exactly how chapter 21 nearly
// shipped a dead feature. The shipped file is checked separately, by
// detectStatusToolDeclared.
func ch22Skills(work string) string {
	dir := filepath.Join(work, "skills", "base")
	_ = os.MkdirAll(dir, 0o755)
	_ = os.WriteFile(filepath.Join(dir, "SKILL.md"), []byte(
		"---\nname: base\ndescription: grader skill for chapter 22\n"+
			"tools: [agent_status]\n---\n\nReport status when asked.\n"), 0o644)
	return filepath.Join(work, "skills")
}

func ch22WaitPort(port string, limit time.Duration) bool {
	deadline := time.Now().Add(limit)
	for time.Now().Before(deadline) {
		c, err := net.DialTimeout("tcp", "127.0.0.1:"+port, 200*time.Millisecond)
		if err == nil {
			_ = c.Close()
			return true
		}
		time.Sleep(50 * time.Millisecond)
	}
	return false
}
