// ch05 is the exercise binary: a custom agent that imports the student's
// framework and adds a tool the builtin agent doesn't have.
//
// The grader sends a prompt that triggers the custom tool, then checks that
// the tool was called and its output appeared in the response.
//
// Usage:
//
//	echo '{"user":"What is 6 * 7?"}' | ./ch05
package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	agent "github.com/waywardgeek/coding-agents-course/agent"
)

func main() {
	// Register a custom tool before creating the agent. This is the whole
	// point of the refactoring: external code can extend the agent's
	// capabilities without modifying the framework.
	agent.RegisterTool(
		"calculate",
		"Evaluate a simple arithmetic expression and return the result.",
		json.RawMessage(`{
			"type": "object",
			"properties": {
				"expression": {
					"type": "string",
					"description": "The arithmetic expression to evaluate, e.g. '6 * 7'"
				}
			},
			"required": ["expression"]
		}`),
		func(args json.RawMessage) (string, error) {
			var p struct {
				Expression string `json:"expression"`
			}
			if err := json.Unmarshal(args, &p); err != nil {
				return "", err
			}
			// Dead-simple: we only need to prove the tool was called.
			// A real calculator would parse and evaluate. For the exercise,
			// the model sees the tool exists and will call it; the grader
			// only checks that the output appeared.
			result := evalSimple(p.Expression)
			return fmt.Sprintf("Result: %s = %s", p.Expression, result), nil
		},
	)

	cfg := agent.Config{
		Vendor:       agent.VendorAnthropic,
		Surface:      agent.DefaultSurface(agent.VendorAnthropic),
		SystemPrompt: "You are a calculator assistant. Use the calculate tool for any arithmetic.",
		MaxTokens:    1024,
	}

	a := agent.NewAgent(cfg, "ch05.log")
	defer a.Shutdown()

	in := bufio.NewScanner(os.Stdin)
	in.Buffer(make([]byte, 0, 64*1024), 8*1024*1024)
	out := bufio.NewWriter(os.Stdout)
	defer out.Flush()

	for in.Scan() {
		line := strings.TrimSpace(in.Text())
		if line == "" {
			continue
		}
		var msg struct {
			User *string `json:"user"`
		}
		if err := json.Unmarshal([]byte(line), &msg); err != nil {
			emit(out, map[string]string{"error": "bad input: " + err.Error()})
			continue
		}
		if msg.User == nil {
			emit(out, map[string]string{"error": "no user field"})
			continue
		}

		reply, err := a.Ask(*msg.User)
		if err != nil {
			emit(out, map[string]string{"error": err.Error()})
			continue
		}
		emit(out, map[string]string{"assistant": reply})
	}

	emit(out, map[string]any{"usage": a.Usage()})
}

func emit(out *bufio.Writer, v any) {
	b, _ := json.Marshal(v)
	out.Write(b)
	out.WriteByte('\n')
	out.Flush()
}

// evalSimple does trivial arithmetic. It's intentionally limited — the point
// is that the tool EXISTS and gets CALLED, not that it's a full calculator.
func evalSimple(expr string) string {
	// Strip spaces, try basic a op b patterns.
	expr = strings.TrimSpace(expr)
	var a, b float64
	var op string
	n, _ := fmt.Sscanf(expr, "%f %s %f", &a, &op, &b)
	if n == 3 {
		switch op {
		case "+":
			return fmt.Sprintf("%g", a+b)
		case "-":
			return fmt.Sprintf("%g", a-b)
		case "*", "×":
			return fmt.Sprintf("%g", a*b)
		case "/", "÷":
			if b == 0 {
				return "error: division by zero"
			}
			return fmt.Sprintf("%g", a/b)
		}
	}
	return "unsupported expression: " + expr
}
