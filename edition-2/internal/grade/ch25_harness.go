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
	"sync/atomic"
	"time"

	"github.com/waywardgeek/ensemble/internal/fakevendor"
)

// Chapter 25 is graded through the seven pinned tool names and nothing else.
// The contract fixes those names and deliberately leaves type names, file
// layout and wiring free, so every probe here drives the student's real
// binary and watches what comes back. Reaching into the student's packages
// would grade the reference implementation's shape instead of the behavior
// the chapter is about.
//
// The thesis under all of it: a fleet must report what the framework watched
// happen, not what an agent says about itself. So the observations are tool
// RESULTS, which the framework wrote, rather than tool calls, which are the
// model's words and prove nothing.

// ch25Marker appears only in the prompt the harness types to the top-level
// agent. A child's conversation never contains it, which is what lets one
// fake vendor answer a parent and its children without guessing at the order
// their requests arrive in.
const ch25Marker = "CH25-PARENT-GO"

// Distinct payloads so a check cannot pass on the wrong one.
const (
	ch25Payload  = "PAYLOAD-7f3a91"
	ch25TaskDone = "CH25-TASK-SUBMIT"

	// A second child with a payload of its own. One child is enough to show
	// that a result comes back; it takes two to show that a single join call
	// fans in, collecting every child rather than just the one it happened to
	// notice first.
	ch25Payload2  = "PAYLOAD-b48c20"
	ch25TaskDone2 = "CH25-TASK-SUBMIT-TWO"

	// Names the grader chooses itself. The contract pins the seven tool names
	// and nothing else, so every identity an assertion leans on has to be one
	// the grader supplied rather than a spelling the student was free to pick.
	ch25LiveChild = "stayer-4c1d"

	ch25MidChild  = "midder-9a22"
	ch25MidTask   = "CH25-TASK-MIDDLE"
	ch25LeafChild = "leafer-3b77"
	ch25LeafTask  = "CH25-TASK-LEAF"
	ch25Question  = "QUESTION-5e19d4"
	ch25Answer    = "ANSWER-8c02fa"
	ch25MidDone   = "MIDDONE-1f6b3e"
	ch25LeafDone  = "LEAFDONE-7a90c5"

	ch25TaskIdle = "CH25-TASK-IDLE"

	// A command that takes its time. The sentinel is arithmetic the shell
	// works out, so the literal the probe searches for never appears in the
	// tool arguments: a probe that matches its own request would pass without
	// the command having run at all.
	ch25SlowCmd = `sleep 12 && echo "SLOWDONE-$((123456+789))"`
	ch25SlowPay = "SLOWDONE-124245"

	// Typed at the agent while that wait is in progress. If this text reaches
	// the model, the only thing that can have carried it is the wait ending
	// early, because the callback was pushed out past the end of the run.
	ch25Interrupt = "INTERRUPT-9b24e7"

	// Long enough that the job's own callback cannot be what wakes the agent.
	ch25LongDelay = 60

	// The child that never finishes.
	ch25IdleChild = "driftr-6c81"
)

// Ch25Run grades a submission.
func Ch25Run(dir string) Ch25Result {
	var r Ch25Result
	bin, cleanup, err := Build(dir)
	if err != nil {
		r.Fatal = fmt.Sprintf("submission does not build: %v", err)
		return r
	}
	defer cleanup()

	ch25Fleet(bin, &r)
	ch25Visibility(bin, &r)
	ch25Messaging(bin, &r)
	ch25Interrupted(bin, &r)
	ch25Idle(bin, &r)
	return r
}

// ch25Env points the agent at the fake vendor.
func ch25Env(url string) []string {
	return append(os.Environ(),
		"LLM_BASE_URL="+url,
		"OPENAI_API_KEY=test",
		"ANTHROPIC_API_KEY=test",
	)
}

// ch25Session runs the binary, types one prompt, and returns the lines it
// printed. onLine, if set, sees every output line as it arrives and may write
// to stdin, which is how a probe delivers a message while a tool is still
// running.
type ch25Session struct {
	lines []string
	err   error
}

func ch25Launch(bin, workdir, url string, args []string, prompt string, onLine func(line string, stdin io.Writer)) ch25Session {
	return ch25LaunchOpts(bin, workdir, url, args, prompt, onLine, nil)
}

// ch25LaunchOpts adds onStart, which runs in its own goroutine once the prompt
// has been sent. A probe that has to deliver a message while a tool is still
// running cannot wait for a particular output line to cue it, because the
// agent is busy inside the tool and may print nothing at all until it is done.
func ch25LaunchOpts(bin, workdir, url string, args []string, prompt string, onLine func(line string, stdin io.Writer), onStart func(stdin io.Writer)) ch25Session {
	var s ch25Session
	cmd := exec.Command(bin, args...)
	cmd.Dir = workdir
	cmd.Env = ch25Env(url)

	stdin, err := cmd.StdinPipe()
	if err != nil {
		s.err = err
		return s
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		s.err = err
		return s
	}
	// Keep stderr: a crashed agent that is diagnosed from silence costs an
	// afternoon, and the message is already there.
	var errBuf strings.Builder
	cmd.Stderr = &errBuf

	if err := cmd.Start(); err != nil {
		s.err = err
		return s
	}

	done := make(chan struct{})
	go func() {
		defer close(done)
		sc := bufio.NewScanner(stdout)
		sc.Buffer(make([]byte, 0, 1<<20), 1<<20)
		for sc.Scan() {
			line := sc.Text()
			s.lines = append(s.lines, line)
			if onLine != nil {
				onLine(line, stdin)
			}
			// An assistant line ends a turn.
			var msg struct {
				Assistant *string `json:"assistant"`
			}
			if json.Unmarshal([]byte(line), &msg) == nil && msg.Assistant != nil {
				return
			}
		}
	}()

	fmt.Fprintf(stdin, "%s\n", ch25Prompt(prompt))
	if onStart != nil {
		go onStart(stdin)
	}

	select {
	case <-done:
	case <-time.After(90 * time.Second):
		s.err = fmt.Errorf("agent did not finish its turn within 90s; stderr: %s", errBuf.String())
	}

	// Close stdin and let the agent shut down on its own. Killing it here
	// would cut off whatever it does on the way out, including writing its
	// conversation, and a probe that looks for a save file would then be
	// measuring the harness rather than the submission.
	_ = stdin.Close()
	exited := make(chan struct{})
	go func() { _ = cmd.Wait(); close(exited) }()
	select {
	case <-exited:
	case <-time.After(20 * time.Second):
		_ = cmd.Process.Kill()
		<-exited
	}
	if s.err == nil && len(s.lines) == 0 {
		s.err = fmt.Errorf("agent printed nothing; stderr: %s", errBuf.String())
	}
	return s
}

func ch25Prompt(text string) string {
	b, _ := json.Marshal(map[string]string{"kind": "prompt", "text": text})
	return string(b)
}

// ch25Args renders tool arguments as the JSON string the fake vendor expects.
func ch25Args(m map[string]any) string {
	b, _ := json.Marshal(m)
	return string(b)
}

// ch25Bodies returns every request body the vendor saw, as strings.
func ch25Bodies(srv *fakevendor.Server) []string {
	var out []string
	for _, rec := range srv.Requests() {
		out = append(out, string(rec.Body))
	}
	return out
}

// ch25SawInResults reports whether want appears in a tool RESULT rather than
// anywhere in the conversation. A probe that searches the whole body matches
// its own tool call arguments and passes without the agent having done
// anything, which is a mistake this harness has made before.
func ch25SawInResults(srv *fakevendor.Server, want string) bool {
	for _, res := range ch23ToolResults(srv.Requests()) {
		if strings.Contains(res, want) {
			return true
		}
	}
	return false
}

// ch25ConversationDirs returns the directories holding a non-empty saved
// conversation. The contract leaves file layout free, so this looks for
// separate conversations rather than for a fixed path: a child that is a full
// agent keeps its own, and a child that is a goroutine wearing a label has
// its entries inside the parent's.
//
// A snapshot and its journal both count. A long lived agent writes the
// snapshot when it shuts down and appends to the journal before then, so
// requiring the snapshot would fail a child for the unrelated reason that it
// was still alive when the parent exited.
func ch25ConversationDirs(root string) []string {
	seen := map[string]bool{}
	_ = filepath.Walk(root, func(p string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || info.Size() == 0 {
			return nil
		}
		if base := filepath.Base(p); strings.HasPrefix(base, "save.json") {
			seen[filepath.Dir(p)] = true
		}
		return nil
	})
	out := make([]string, 0, len(seen))
	for d := range seen {
		out = append(out, d)
	}
	return out
}

// ch25Fleet covers spawning, the tool surface, and submit.
//
// The parent spawns two children that submit, then joins both in one call.
// The payload has to survive the child exiting, because a result a parent can
// only read while the child is alive is not a result.
func ch25Fleet(bin string, r *Ch25Result) {
	work, err := os.MkdirTemp("", "ch25-fleet-*")
	if err != nil {
		r.failAll("cannot create a working directory: %v", err)
		return
	}
	defer os.RemoveAll(work)

	// Routing keys off the tool IDs echoed back in results rather than off a
	// count of result blocks, because the two vendors spell a result
	// differently and the grader must not care which one the student uses.
	srv := fakevendor.NewWithOptions(nil, fakevendor.Options{
		Route: func(body []byte) *fakevendor.Reply {
			b := string(body)
			switch {
			case !strings.Contains(b, ch25Marker):
				// A child. It submits once, then says it is finished.
				if strings.Contains(b, "ch25c1") {
					return &fakevendor.Reply{Text: "finished"}
				}
				payload := ch25Payload
				if strings.Contains(b, ch25TaskDone2) {
					payload = ch25Payload2
				}
				return &fakevendor.Reply{
					ToolName: "submit",
					ToolArgs: ch25Args(map[string]any{"result": payload}),
					ToolID:   "ch25c1",
				}
			case !strings.Contains(b, "ch25p1"):
				return &fakevendor.Reply{
					ToolName: "spawn_sub_agent",
					ToolArgs: ch25Args(map[string]any{"agent_id": "worker", "task": ch25TaskDone}),
					ToolID:   "ch25p1",
				}
			case !strings.Contains(b, "ch25p2"):
				return &fakevendor.Reply{
					ToolName: "spawn_sub_agent",
					ToolArgs: ch25Args(map[string]any{"agent_id": "worker2", "task": ch25TaskDone2}),
					ToolID:   "ch25p2",
				}
			case !strings.Contains(b, "ch25p3"):
				return &fakevendor.Reply{
					ToolName: "join_agents",
					ToolArgs: ch25Args(map[string]any{"agent_ids": []string{"worker", "worker2"}}),
					ToolID:   "ch25p3",
				}
			default:
				return &fakevendor.Reply{Text: "all done"}
			}
		},
	})
	defer srv.Close()

	sess := ch25Launch(bin, work, srv.URL(), nil, ch25Marker+" spawn two workers and collect their results", nil)
	if sess.err != nil {
		r.failAll("could not drive the agent: %v", sess.err)
		return
	}

	// A pinned name that is never advertised cannot carry its behaviour, so
	// attribute a missing name to the check that needs it. A student who omits
	// one tool should lose one check rather than all seven.
	ch25RequireTools(r, srv, "spawn-reports-back", "spawn_sub_agent", "submit")
	ch25RequireTools(r, srv, "join-fans-in", "join_agents")

	// Check 1: a spawned child runs and reports back. The payload comes back in
	// a tool RESULT, which is the framework reporting, rather than in a tool
	// call, which is only the model's words.
	r.ran("spawn-reports-back")
	if !ch25SawInResults(srv, ch25Payload) {
		r.fail("spawn-reports-back", "the child's submitted payload %q never came back to the parent in a tool result", ch25Payload)
	}

	// Check 2: children are isolated, which shows up as a conversation of its
	// own in a directory of its own rather than as entries in its parent's.
	r.ran("child-isolated")
	dirs := ch25ConversationDirs(work)
	if len(dirs) < 2 {
		r.fail("child-isolated", "found saved conversations in %d director(ies) under the data directory, want at least 2 (the parent and its children); a spawned child is a full agent and keeps its own conversation, not entries inside its parent's", len(dirs))
	}

	// Check 4: one join call fans in. Both payloads have to arrive in the SAME
	// tool result, which is what separates collecting every listed child from
	// returning whichever one happened to finish first.
	r.ran("join-fans-in")
	if !ch25OneResultHasAll(srv, ch25Payload, ch25Payload2) {
		r.fail("join-fans-in", "no single tool result carried both children's results (%q and %q); join_agents must block until every listed child finishes and then return each one's result", ch25Payload, ch25Payload2)
	}
}

// ch25RequireTools reports whether every named tool was advertised, failing
// check with the missing ones if not. Attributing a missing name to the one
// check that needs it keeps omitting a single tool from costing every check,
// so the diagnosis names the tool rather than the behaviour that silently
// could not happen without it.
func ch25RequireTools(r *Ch25Result, srv *fakevendor.Server, check string, names ...string) bool {
	advertised := map[string]bool{}
	for _, b := range ch25Bodies(srv) {
		for n := range ch23AdvertisedTools([]byte(b)) {
			advertised[n] = true
		}
	}
	var missing []string
	for _, n := range names {
		if !advertised[n] {
			missing = append(missing, n)
		}
	}
	if len(missing) > 0 {
		r.ran(check)
		r.fail(check, "these pinned tool names are never advertised to the model: %s", strings.Join(missing, ", "))
		return false
	}
	return true
}

// ch25OneResultHasAll reports whether a single tool result carried every one
// of the given substrings. Separate results carrying one each would also
// satisfy a naive scan, which is why this insists they arrive together.
func ch25OneResultHasAll(srv *fakevendor.Server, subs ...string) bool {
	for _, res := range ch23ToolResults(srv.Requests()) {
		all := true
		for _, s := range subs {
			if !strings.Contains(res, s) {
				all = false
				break
			}
		}
		if all {
			return true
		}
	}
	return false
}

// ch25Visibility grades check 3: a live child is listed, and a child that has
// been shut down is not.
//
// The teardown half is the one worth having. A fleet view that only ever grows
// looks correct for as long as nobody shuts anything down.
func ch25Visibility(bin string, r *Ch25Result) {
	work, err := os.MkdirTemp("", "ch25-vis-*")
	if err != nil {
		r.failAll("cannot create a working directory: %v", err)
		return
	}
	defer os.RemoveAll(work)

	srv := fakevendor.NewWithOptions(nil, fakevendor.Options{
		Route: func(body []byte) *fakevendor.Reply {
			b := string(body)
			switch {
			case !strings.Contains(b, ch25Marker):
				// The child does nothing but stay alive long enough to be seen.
				return &fakevendor.Reply{Text: "idling"}
			case !strings.Contains(b, "ch25v1"):
				return &fakevendor.Reply{
					ToolName: "spawn_sub_agent",
					ToolArgs: ch25Args(map[string]any{"agent_id": ch25LiveChild, "task": "stay up"}),
					ToolID:   "ch25v1",
				}
			case !strings.Contains(b, "ch25v2"):
				return &fakevendor.Reply{ToolName: "agents_status", ToolArgs: "{}", ToolID: "ch25v2"}
			case !strings.Contains(b, "ch25v3"):
				return &fakevendor.Reply{
					ToolName: "shutdown_agent",
					ToolArgs: ch25Args(map[string]any{"agent_id": ch25LiveChild}),
					ToolID:   "ch25v3",
				}
			case !strings.Contains(b, "ch25v4"):
				return &fakevendor.Reply{ToolName: "agents_status", ToolArgs: "{}", ToolID: "ch25v4"}
			default:
				return &fakevendor.Reply{Text: "all done"}
			}
		},
	})
	defer srv.Close()

	sess := ch25Launch(bin, work, srv.URL(), nil, ch25Marker+" check on the fleet", nil)
	if sess.err != nil {
		r.fail("fleet-visibility", "could not drive the agent: %v", sess.err)
		return
	}
	ch25RequireTools(r, srv, "fleet-visibility", "agents_status", "shutdown_agent")

	r.ran("fleet-visibility")
	res := ch23ToolResults(srv.Requests())
	if len(res) < 4 {
		r.fail("fleet-visibility", "expected at least 4 tool results (spawn, status, shutdown, status) but saw %d", len(res))
		return
	}

	// Endpoint shaped rather than positional: something between the spawn and
	// the final status has to have listed the child, and the final status is
	// the one that must no longer mention it.
	listedAlive := false
	for _, s := range res[1 : len(res)-1] {
		if strings.Contains(s, ch25LiveChild) {
			listedAlive = true
			break
		}
	}
	if !listedAlive {
		r.fail("fleet-visibility", "agents_status never listed the live child %q", ch25LiveChild)
	}
	if last := res[len(res)-1]; strings.Contains(last, ch25LiveChild) {
		r.fail("fleet-visibility", "after shutdown_agent, agents_status still lists %q: %s", ch25LiveChild, ch25Trim(last))
	}
}

// ch25Messaging grades check 5.
//
// The fleet is three deep on purpose. The contract asks that a reply carry its
// sender's address, and the only addresses the grader can name without pinning
// a spelling the contract leaves free are the ones it chose itself. So the
// REPLY under test is the one a grader named child sends to a grader named
// grandchild, and the address looked for is that child's id.
func ch25Messaging(bin string, r *Ch25Result) {
	work, err := os.MkdirTemp("", "ch25-msg-*")
	if err != nil {
		r.failAll("cannot create a working directory: %v", err)
		return
	}
	defer os.RemoveAll(work)

	// A join that the framework backgrounds comes back saying the job is still
	// running, which is correct behaviour and not an answer. The scripted
	// model has to do what a real one would and ask again, so these count the
	// attempts and keep the ids unique.
	var rootJoins, midJoins atomic.Int32

	srv := fakevendor.NewWithOptions(nil, fakevendor.Options{
		Route: func(body []byte) *fakevendor.Reply {
			b := string(body)
			switch {
			case strings.Contains(b, ch25Marker):
				// The root: spawn the middle agent and wait for it.
				switch {
				case !strings.Contains(b, "ch25m1"):
					return &fakevendor.Reply{
						ToolName: "spawn_sub_agent",
						ToolArgs: ch25Args(map[string]any{"agent_id": ch25MidChild, "task": ch25MidTask}),
						ToolID:   "ch25m1",
					}
				case strings.Contains(b, ch25MidDone) || rootJoins.Load() > 8:
					return &fakevendor.Reply{Text: "root done"}
				default:
					return &fakevendor.Reply{
						ToolName: "join_agents",
						ToolArgs: ch25Args(map[string]any{"agent_ids": []string{ch25MidChild}}),
						ToolID:   fmt.Sprintf("ch25m2-%d", rootJoins.Add(1)),
					}
				}

			case strings.Contains(b, ch25MidTask):
				// The middle agent: spawn a grandchild, wait for it, and when
				// a question arrives answer it and report upward.
				switch {
				case !strings.Contains(b, "ch25m3"):
					return &fakevendor.Reply{
						ToolName: "spawn_sub_agent",
						ToolArgs: ch25Args(map[string]any{"agent_id": ch25LeafChild, "task": ch25LeafTask}),
						ToolID:   "ch25m3",
					}
				case strings.Contains(b, ch25Question) && !strings.Contains(b, "ch25m5"):
					return &fakevendor.Reply{
						ToolName: "respond_to_parent_message",
						ToolArgs: ch25Args(map[string]any{"agent_id": ch25LeafChild, "response": ch25Answer}),
						ToolID:   "ch25m5",
					}
				case strings.Contains(b, ch25LeafDone) && !strings.Contains(b, "ch25m7"):
					return &fakevendor.Reply{
						ToolName: "submit",
						ToolArgs: ch25Args(map[string]any{"result": ch25MidDone}),
						ToolID:   "ch25m7",
					}
				case strings.Contains(b, "ch25m7") || midJoins.Load() > 8:
					return &fakevendor.Reply{Text: "mid done"}
				default:
					return &fakevendor.Reply{
						ToolName: "join_agents",
						ToolArgs: ch25Args(map[string]any{"agent_ids": []string{ch25LeafChild}}),
						ToolID:   fmt.Sprintf("ch25m4-%d", midJoins.Add(1)),
					}
				}

			default:
				// The grandchild: ask upward, and once answered report what
				// it was told. Submitting rather than just stopping is what
				// lets the middle agent wait for it, so the reply is observed
				// instead of being raced by the process exiting.
				switch {
				case strings.Contains(b, "ch25m9"):
					return &fakevendor.Reply{Text: "leaf done"}
				case strings.Contains(b, "ch25m6"):
					return &fakevendor.Reply{
						ToolName: "submit",
						ToolArgs: ch25Args(map[string]any{"result": ch25LeafDone}),
						ToolID:   "ch25m9",
					}
				default:
					return &fakevendor.Reply{
						ToolName: "send_message_to_parent",
						ToolArgs: ch25Args(map[string]any{"content": ch25Question}),
						ToolID:   "ch25m6",
					}
				}
			}
		},
	})
	defer srv.Close()

	sess := ch25Launch(bin, work, srv.URL(), nil, ch25Marker+" delegate and answer questions", nil)
	if sess.err != nil {
		r.fail("addressed-messaging", "could not drive the agent: %v", sess.err)
		return
	}
	ch25RequireTools(r, srv, "addressed-messaging", "send_message_to_parent", "respond_to_parent_message")

	r.ran("addressed-messaging")
	res := ch23ToolResults(srv.Requests())

	// The question must come back out of a join. That is what shows the child
	// blocked and the parent got control back to answer it, rather than the
	// join sitting there until a deadline.
	asked := false
	for _, s := range res {
		if strings.Contains(s, ch25Question) {
			asked = true
			break
		}
	}
	if !asked {
		r.fail("addressed-messaging", "the question %q never reached the waiting parent; join_agents must return early when a child is blocked on send_message_to_parent", ch25Question)
	}

	// The reply has to arrive carrying who sent it, or the agent that receives
	// it has an answer and no way to address a follow up.
	addressed := false
	for _, s := range res {
		if strings.Contains(s, ch25Answer) && strings.Contains(s, ch25MidChild) {
			addressed = true
			break
		}
	}
	if !addressed {
		r.fail("addressed-messaging", "the reply %q never arrived carrying its sender's address %q; a reply has to name who sent it so the receiver can answer back", ch25Answer, ch25MidChild)
	}
}

// ch25Trim shortens a tool result for a failure message, so a diagnosis names
// what went wrong without pasting an entire conversation into the output.
func ch25Trim(s string) string {
	s = strings.ReplaceAll(s, "\n", " ")
	if len(s) > 200 {
		return s[:200] + "..."
	}
	return s
}

// ch25Interrupted covers wait-interrupt. The agent starts a genuinely slow
// command and pushes the wait's own callback out past the end of the run, so
// that nothing except an arriving message can return control to it. A message
// typed while that wait is in flight should end the wait, and the command it
// walked away from should still deliver its output afterwards.
//
// The slow thing has to be a real tool rather than a slow vendor reply,
// because the fake vendor answers one request at a time: a sleep taken inside
// a reply stalls every other agent in the run, including the one under test.
func ch25Interrupted(bin string, r *Ch25Result) {
	work, err := os.MkdirTemp("", "ch25-int-*")
	if err != nil {
		r.fail("wait-interrupt", "cannot create a working directory: %v", err)
		return
	}
	defer os.RemoveAll(work)

	var waits atomic.Int32
	srv := fakevendor.NewWithOptions(nil, fakevendor.Options{
		Route: func(body []byte) *fakevendor.Reply {
			b := string(body)
			switch {
			case strings.Contains(b, ch25SlowPay):
				return &fakevendor.Reply{Text: "interrupt run done"}
			case !strings.Contains(b, "ch25i1"):
				// Push the wait's own callback out past the end of the run,
				// so that waking early can only be the message's doing.
				return &fakevendor.Reply{
					ToolName: "tool_limits",
					ToolArgs: ch25Args(map[string]any{"ai_callback_delay": ch25LongDelay}),
					ToolID:   "ch25i1",
				}
			case !strings.Contains(b, "ch25i2"):
				return &fakevendor.Reply{
					ToolName: "run_command",
					ToolArgs: ch25Args(map[string]any{"command": ch25SlowCmd}),
					ToolID:   "ch25i2",
				}
			default:
				// Back in control with the command still out. Keep asking
				// after it, the way a model would, until it reports.
				n := waits.Add(1)
				return &fakevendor.Reply{
					ToolName: "wait_for_job",
					ToolArgs: ch25Args(map[string]any{"handle": 1}),
					ToolID:   fmt.Sprintf("ch25iw%d", n),
				}
			}
		},
	})
	defer srv.Close()

	sess := ch25LaunchOpts(bin, work, srv.URL(), nil,
		ch25Marker+" run the slow command and wait for it", nil,
		func(stdin io.Writer) {
			// Late enough that the wait is certainly in flight, early enough
			// that the command is certainly not finished.
			time.Sleep(5 * time.Second)
			fmt.Fprintf(stdin, "%s\n", ch25Prompt(ch25Interrupt))
		})
	if sess.err != nil {
		r.fail("wait-interrupt", "could not drive the agent: %v", sess.err)
		return
	}
	if !ch25RequireTools(r, srv, "wait-interrupt", "tool_limits") {
		return
	}
	r.ran("wait-interrupt")
	// The wake is a turn carrying the typed message but not yet the command's
	// output: the agent has control back with the command still running. Had
	// the message merely queued, the next turn would carry both at once.
	woke := false
	for _, b := range ch25Bodies(srv) {
		if strings.Contains(b, ch25Interrupt) && !strings.Contains(b, ch25SlowPay) {
			woke = true
			break
		}
	}
	if !woke {
		r.fail("wait-interrupt", "the agent took no turn between the message arriving and the command finishing: a message that lands during a tool wait has to end the wait, not queue behind it")
		return
	}
	if !ch25SawInResults(srv, ch25SlowPay) {
		r.fail("wait-interrupt", "the abandoned command never delivered %q: ending a wait early has to leave the tool running rather than cancel it", ch25SlowPay)
	}
}

// ch25Idle covers idle-corrected. The child answers in prose and never calls
// submit. Stopping without submitting is a mistake rather than an answer, so
// the framework should wake the child to finish, and give up after a fixed
// number of attempts by telling the parent rather than trying forever.
func ch25Idle(bin string, r *Ch25Result) {
	work, err := os.MkdirTemp("", "ch25-idle-*")
	if err != nil {
		r.fail("idle-corrected", "cannot create a working directory: %v", err)
		return
	}
	defer os.RemoveAll(work)

	srv := fakevendor.NewWithOptions(nil, fakevendor.Options{
		Route: func(body []byte) *fakevendor.Reply {
			b := string(body)
			switch {
			case !strings.Contains(b, ch25Marker):
				// The child, however many times it is asked, never finishes.
				return &fakevendor.Reply{Text: "I had a look around."}
			case !strings.Contains(b, "ch25d1"):
				return &fakevendor.Reply{
					ToolName: "spawn_sub_agent",
					ToolArgs: ch25Args(map[string]any{"agent_id": ch25IdleChild, "task": ch25TaskIdle}),
					ToolID:   "ch25d1",
				}
			case !strings.Contains(b, "ch25d1b"):
				// Try to submit while the child is active. The framework
				// must reject this: submits happen bottom-up.
				return &fakevendor.Reply{
					ToolName: "submit",
					ToolArgs: ch25Args(map[string]any{"result": "premature"}),
					ToolID:   "ch25d1b",
				}
			case !strings.Contains(b, "ch25d2"):
				return &fakevendor.Reply{
					ToolName: "join_agents",
					ToolArgs: ch25Args(map[string]any{"agent_ids": []string{ch25IdleChild}}),
					ToolID:   "ch25d2",
				}
			default:
				return &fakevendor.Reply{Text: "idle run done"}
			}
		},
	})
	defer srv.Close()

	sess := ch25Launch(bin, work, srv.URL(), nil, ch25Marker+" start the drifting worker", nil)
	if sess.err != nil {
		r.fail("idle-corrected", "could not drive the agent: %v", sess.err)
		return
	}
	if !ch25RequireTools(r, srv, "idle-corrected", "spawn_sub_agent", "submit", "join_agents") {
		return
	}
	r.ran("idle-corrected")

	// Verify submit-rejection: the parent tried to submit while the child
	// was active, and the tool result must contain an error. The error text
	// appears in the next request body as a tool_result message.
	sawReject := false
	for _, b := range ch25Bodies(srv) {
		if strings.Contains(b, ch25Marker) &&
			(strings.Contains(b, "cannot submit") || strings.Contains(b, "still active")) {
			sawReject = true
			break
		}
	}
	if !sawReject {
		r.fail("idle-corrected", "submit while the child %q is active was not rejected: submit must fail when active children exist", ch25IdleChild)
		return
	}

	// The child makes no tool calls, so one turn is one request. Counting them
	// pins both halves of the rule at once: a framework that treats stopping
	// as finishing leaves this at one, and a framework that nags without a
	// bound runs until the harness times out.
	turns := 0
	for _, b := range ch25Bodies(srv) {
		if !strings.Contains(b, ch25Marker) {
			turns++
		}
	}
	const want = 4 // the child's own turn, plus three wakes
	if turns != want {
		r.fail("idle-corrected", "the child that never submitted took %d turns, expected %d: its own, plus three wakes before the framework gives up", turns, want)
		return
	}
	if !ch25SawInResults(srv, ch25IdleChild) {
		r.fail("idle-corrected", "after giving up on %q the framework told nobody: the parent is the one that can act on it", ch25IdleChild)
	}
}
