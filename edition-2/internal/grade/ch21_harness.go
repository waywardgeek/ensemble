package grade

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/waywardgeek/ensemble/internal/fakevendor"
)

// ch21Model is deliberately a model that exists in BOTH the live tree and the
// frozen snapshot. A model row present only in one of them produces a grader
// that scores 100 where it was developed and silently lower where it is
// actually run — a trap this repo has already sprung once.
const ch21Model = "claude-sonnet-5"

type ch21Opts struct {
	dir     string
	prompts []string
	replies []fakevendor.Reply
}

type ch21Out struct {
	// requests is every request the fake vendor received, in order. This is
	// the grading surface: what came out of the socket, not what the source
	// says it does.
	requests []fakevendor.Recorded
	// calls is every tools/call the fake MCP server received.
	calls []ch21Call
	// replies is one entry per prompt: the agent's final text, or the error.
	replies []string
	fatal   string
}

// ch21Launch runs the student's agent against a fake vendor and a fake MCP
// server reached over the URL transport.
//
// The grader supplies the skill file, so the grader owns the endpoint and the
// tool names. That is what makes the scored checks deterministic without an
// internet connection, and it means a student who chose a different search
// backend is graded on their wiring rather than on their vendor.
func ch21Launch(bin, guiDir, nonce string, o ch21Opts) ch21Out {
	fake := newCh21FakeMCP(nonce)
	defer fake.Close()

	skillsDir, cleanupSkills, err := ch21Skills(fake.URL())
	if err != nil {
		return ch21Out{fatal: fmt.Sprintf("skills dir: %v", err)}
	}
	defer cleanupSkills()

	// Each scenario gets its own working directory. Since Chapter 11 the
	// agent loads save.json on startup by default, so two scenarios sharing a
	// directory means the second one silently resumes the first one's
	// conversation and every scripted reply lands one turn out of place.
	work, err := os.MkdirTemp("", "ch21-work-")
	if err != nil {
		return ch21Out{fatal: fmt.Sprintf("work dir: %v", err)}
	}
	defer os.RemoveAll(work)
	o.dir = work

	srv := fakevendor.New(o.replies)
	defer srv.Close()

	out := ch21Drive(bin, guiDir, skillsDir, srv.URL(), o)
	out.requests = srv.Requests()
	out.calls = fake.Calls()
	return out
}

// ch21Drive starts the agent, feeds it prompts on stdin and collects the
// reply the agent writes to stdout for each one.
//
// It drives the server mode rather than `chat`, because MCP servers are
// connected inside runActorLoop. stdin/stdout is enough: in server mode the
// agent answers each {"kind":"prompt"} with one {"assistant":...} or
// {"error":...} line, so turn completion is observable without opening a
// websocket to the GUI.
func ch21Drive(bin, guiDir, skillsDir, vendorURL string, o ch21Opts) ch21Out {
	var out ch21Out

	port := freePort()
	cmd := exec.Command(bin, "--port", port, "--gui-dir", guiDir)
	cmd.Dir = o.dir
	cmd.Env = append(os.Environ(),
		"LLM_BASE_URL="+vendorURL,
		"LLM_MODEL="+ch21Model,
		"LLM_VENDOR=anthropic",
		"LLM_API_KEY=test-key",
		"EN_SKILLS_DIR="+skillsDir,
		"EN_PRIMARY_SKILL=base",
	)

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return ch21Out{fatal: fmt.Sprintf("stdin pipe: %v", err)}
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return ch21Out{fatal: fmt.Sprintf("stdout pipe: %v", err)}
	}
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		return ch21Out{fatal: fmt.Sprintf("start: %v", err)}
	}

	// One reader goroutine owns stdout. Replies are delivered on a channel so
	// that a prompt which never completes times out instead of wedging the
	// whole grader — a hung grader reports nothing, which is worse than a
	// failing one because it cannot be distinguished from a slow machine.
	lines := make(chan string, 64)
	go func() {
		defer close(lines)
		sc := bufio.NewScanner(stdout)
		sc.Buffer(make([]byte, 0, 64*1024), 8*1024*1024)
		for sc.Scan() {
			lines <- sc.Text()
		}
	}()

	stop := func() {
		stdin.Close()
		if cmd.Process != nil {
			cmd.Process.Kill()
		}
		cmd.Wait()
		io.Copy(io.Discard, stdout)
	}
	defer stop()

	if !waitForHTTP("http://127.0.0.1:"+port+"/", 20*time.Second) {
		out.fatal = "agent did not come up within 20s"
		return out
	}

	for _, p := range o.prompts {
		msg, _ := json.Marshal(map[string]string{"kind": "prompt", "text": p})
		if _, err := stdin.Write(append(msg, '\n')); err != nil {
			out.fatal = fmt.Sprintf("write prompt: %v", err)
			return out
		}
		reply, err := ch21AwaitReply(lines, 90*time.Second)
		if err != nil {
			out.fatal = fmt.Sprintf("prompt %q: %v", p, err)
			return out
		}
		out.replies = append(out.replies, reply)
	}
	return out
}

// ch21AwaitReply reads stdout until the agent reports the turn finished.
// Lines that are not the completion envelope are ignored rather than treated
// as errors: the agent is free to log whatever it likes on stdout, and a
// grader that breaks when a student adds a log line is grading formatting.
func ch21AwaitReply(lines <-chan string, limit time.Duration) (string, error) {
	deadline := time.After(limit)
	for {
		select {
		case line, ok := <-lines:
			if !ok {
				return "", fmt.Errorf("agent exited before replying")
			}
			line = strings.TrimSpace(line)
			if !strings.HasPrefix(line, "{") {
				continue
			}
			var env map[string]any
			if json.Unmarshal([]byte(line), &env) != nil {
				continue
			}
			if s, ok := env["assistant"].(string); ok {
				return s, nil
			}
			if s, ok := env["error"].(string); ok {
				return "", fmt.Errorf("agent reported: %s", s)
			}
		case <-deadline:
			return "", fmt.Errorf("no reply within %s", limit)
		}
	}
}

// ch21Skills writes the skills directory the agent will read: a minimal
// primary skill, plus the loadable web-search skill pointing at the fake MCP
// server.
//
// web-search is `type: loadable` on purpose. onSkillMCPConnect has exactly one
// call site, inside the load_skill tool handler, so mcp_servers declared on a
// PRIMARY skill parses cleanly and then silently never connects. Loadable is
// both the architecture's intent and the only shape that works.
func ch21Skills(mcpURL string) (string, func(), error) {
	dir, err := os.MkdirTemp("", "ch21-skills-")
	if err != nil {
		return "", func() {}, err
	}
	cleanup := func() { os.RemoveAll(dir) }

	// loadable-skills is not decoration: Chapter 10's progressive disclosure
	// means load_skill refuses any skill that the currently loaded set does
	// not whitelist. Omitting it here makes every web check fail with the
	// tools simply absent, which looks exactly like a broken transport.
	base := strings.Join([]string{
		"---",
		"name: base",
		"description: Base skill.",
		"type: primary",
		"tools: read_file think load_skill unload_skill",
		"loadable-skills: web-search",
		"---",
		"",
		"You are a helpful agent.",
		"",
	}, "\n")

	if err := ch21WriteSkill(dir, "base", base); err != nil {
		cleanup()
		return "", func() {}, err
	}
	if err := ch21WriteSkill(dir, "web-search", ch21SkillMD(mcpURL)); err != nil {
		cleanup()
		return "", func() {}, err
	}
	return dir, cleanup, nil
}

func ch21WriteSkill(root, name, body string) error {
	d := filepath.Join(root, name)
	if err := os.MkdirAll(d, 0o755); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(d, "SKILL.md"), []byte(body), 0o644)
}
