package main

// ch06 exercise — three-agent author / editor / reviewer workflow.
//
// The binary reads JSON events on stdin and emits observations as JSON on stdout.
// It uses the agent library to create three agents, each with its own prompt.
//
// Protocol:
//   {"kind":"prompt","text":"..."}                — start the author; pipeline flows through editor → reviewer
//   {"kind":"hint","text":"...", "agent":"<id>"}   — deliver hint to specific agent (default: all)
//   {"kind":"interrupt"}                          — interrupt all agents
//   {"user":"..."}                                — backward compat, same as kind:prompt
//
// Observations are emitted on stdout, one JSON object per line.
// The event log path is read from CH06_LOG (default: none).

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/waywardgeek/coding-agents-course/agent"
)

const (
	authorName   agent.AgentID = "author"
	editorName   agent.AgentID = "editor"
	reviewerName agent.AgentID = "reviewer"
)

func main() {
	host := cliHost{}
	fw := agent.NewFramework(host)

	authorAgent, err := makeAgent("author", authorPrompt)
	if err != nil {
		fmt.Fprintln(os.Stderr, "author:", err)
		os.Exit(2)
	}
	editorAgent, err := makeAgent("editor", editorPrompt)
	if err != nil {
		fmt.Fprintln(os.Stderr, "editor:", err)
		os.Exit(2)
	}
	reviewerAgent, err := makeAgent("reviewer", reviewerPrompt)
	if err != nil {
		fmt.Fprintln(os.Stderr, "reviewer:", err)
		os.Exit(2)
	}
	defer authorAgent.Shutdown()
	defer editorAgent.Shutdown()
	defer reviewerAgent.Shutdown()

	authorActor := authorAgent.NewActor()
	editorActor := editorAgent.NewActor()
	reviewerActor := reviewerAgent.NewActor()

	fw.Add(authorName, authorActor)
	fw.Add(editorName, editorActor)
	fw.Add(reviewerName, reviewerActor)

	out := bufio.NewWriter(os.Stdout)
	defer out.Flush()

	var outMu sync.Mutex
	emit := func(v any) {
		outMu.Lock()
		defer outMu.Unlock()
		b, _ := json.Marshal(v)
		out.Write(b)
		out.WriteByte('\n')
		out.Flush()
	}

	// Optional event log.
	var logFile *os.File
	if p := os.Getenv("CH06_LOG"); p != "" {
		f, err := os.Create(p)
		if err == nil {
			logFile = f
			defer f.Close()
		}
	}
	logEvent := func(v any) {
		if logFile != nil {
			b, _ := json.Marshal(v)
			logFile.Write(b)
			logFile.Write([]byte{'\n'})
		}
	}

	// Attach observer that emits to stdout and logs.
	actors := map[agent.AgentID]*agent.Actor{
		authorName: authorActor, editorName: editorActor, reviewerName: reviewerActor,
	}
	for id, act := range actors {
		capturedID := id
		act.Attach(observerFunc(func(obs agent.Observation) {
			tagged := observationJSON(obs, capturedID)
			emit(tagged)
			logEvent(tagged)
		}))
	}

	// Start all actor loops.
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	for _, act := range actors {
		a := act
		go a.Run(ctx)
	}

	// Read stdin.
	in := bufio.NewScanner(os.Stdin)
	in.Buffer(make([]byte, 0, 64*1024), 8*1024*1024)

	for in.Scan() {
		line := strings.TrimSpace(in.Text())
		if line == "" {
			continue
		}

		var msg inputMsg
		if err := json.Unmarshal([]byte(line), &msg); err != nil {
			emit(map[string]string{"error": "bad input: " + err.Error()})
			continue
		}

		// Chapter 6 protocol.
		if msg.Kind != nil {
			switch *msg.Kind {
			case "prompt":
				if msg.Text == nil {
					emit(map[string]string{"error": "prompt requires text"})
					continue
				}
				runPipeline(ctx, fw, *msg.Text, emit, logEvent)
			case "hint":
				if msg.Text == nil {
					emit(map[string]string{"error": "hint requires text"})
					continue
				}
				// Deliver to specific agent or to all.
				if msg.Agent != nil && *msg.Agent != "" {
					target := agent.AgentID(*msg.Agent)
					if a, ok := fw.Get(target); ok {
						a.Send(agent.Hint{Text: *msg.Text})
					}
				} else {
					// Deliver to all.
					for _, act := range actors {
						act.Send(agent.Hint{Text: *msg.Text})
					}
				}
			case "interrupt":
				for _, act := range actors {
					act.Send(agent.Interrupt{})
				}
			default:
				emit(map[string]string{"error": "unknown kind: " + *msg.Kind})
			}
			continue
		}

		// Backward compat: {"user":"..."}
		if msg.User != nil {
			runPipeline(ctx, fw, *msg.User, emit, logEvent)
			continue
		}

		emit(map[string]string{"error": "no recognized field"})
	}

	_ = fw.Shutdown()
	out.Flush()
}

type inputMsg struct {
	Kind  *string `json:"kind"`
	Text  *string `json:"text"`
	Agent *string `json:"agent"`
	User  *string `json:"user"`
}

// runPipeline executes the author → editor → reviewer pipeline.
func runPipeline(ctx context.Context, fw *agent.Framework, text string, emit func(any), logEvent func(any)) {
	// Step 1: Author writes.
	a, _ := fw.Get(authorName)
	authorReply, err := a.Ask(text)
	if err != nil {
		emit(map[string]string{"error": "author: " + err.Error()})
		return
	}
	logEvent(map[string]string{"pipeline": "author_done", "text": authorReply})

	// Step 2: Editor improves.
	e, _ := fw.Get(editorName)
	editorReply, err := e.Ask("Please improve this draft:\n\n" + authorReply)
	if err != nil {
		emit(map[string]string{"error": "editor: " + err.Error()})
		return
	}
	logEvent(map[string]string{"pipeline": "editor_done", "text": editorReply})

	// Step 3: Reviewer approves.
	r, _ := fw.Get(reviewerName)
	reviewerReply, err := r.Ask("Please review this revised draft:\n\n" + editorReply)
	if err != nil {
		emit(map[string]string{"error": "reviewer: " + err.Error()})
		return
	}
	logEvent(map[string]string{"pipeline": "reviewer_done", "text": reviewerReply})

	emit(map[string]string{"assistant": reviewerReply, "pipeline": "complete"})
}

// makeAgent creates an agent for a given role.
func makeAgent(role, systemPrompt string) (*agent.Agent, error) {
	cfg, err := agent.ConfigFromEnv()
	if err != nil {
		return nil, err
	}
	cfg.SystemPrompt = systemPrompt
	cfg.MaxTokens = 1024

	logPath := envOr("CH06_"+strings.ToUpper(role)+"_LOG", role+".log")
	return agent.NewAgent(cfg, logPath), nil
}

// observationJSON converts an observation to a JSON-friendly map.
func observationJSON(obs agent.Observation, defaultAgent agent.AgentID) map[string]any {
	agentID := string(defaultAgent)
	switch v := obs.(type) {
	case agent.PartDelta:
		if v.Agent != "" {
			agentID = string(v.Agent)
		}
		return map[string]any{"observation": "part_delta", "agent": agentID, "part_id": v.PartID, "chunk": v.Chunk}
	case agent.PartFinal:
		if v.Agent != "" {
			agentID = string(v.Agent)
		}
		m := map[string]any{"observation": "part_final", "agent": agentID, "part_id": v.PartID, "seq": v.Seq}
		if tp, ok := v.Part.(agent.TextPart); ok {
			m["text"] = tp.Text
		}
		return m
	case agent.StateChanged:
		if v.Agent != "" {
			agentID = string(v.Agent)
		}
		return map[string]any{"observation": "state_changed", "agent": agentID, "from": v.From.String(), "to": v.To.String()}
	case agent.TurnEnded:
		if v.Agent != "" {
			agentID = string(v.Agent)
		}
		m := map[string]any{"observation": "turn_ended", "agent": agentID, "text": v.Text}
		if v.Err != "" {
			m["error"] = v.Err
		}
		return m
	}
	return map[string]any{"observation": "unknown", "agent": agentID}
}

// observerFunc adapts a plain function to the Observer interface.
type observerFunc func(agent.Observation)

func (f observerFunc) Observe(o agent.Observation) { f(o) }

const authorPrompt = `You are an author. Write creative, engaging prose based on the given topic.
Use vivid language and create compelling narratives.`

const editorPrompt = `You are an editor. Improve the given draft by fixing grammar, improving flow,
and strengthening the writing while preserving the author's voice.`

const reviewerPrompt = `You are a reviewer. Evaluate the revised draft and provide a final assessment.
Approve the draft if it meets quality standards, or suggest specific improvements.`

type cliHost struct{}

func (cliHost) Logf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
}

func envOr(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
