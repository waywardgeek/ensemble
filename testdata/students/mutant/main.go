// Command mutant is a deliberately defective Chapter 1 submission, used to
// prove the grader is sensitive rather than decorative. The defect is selected
// by the COURSE_MUTATION environment variable, one axis at a time.
//
// This lives under testdata/ so the go tool ignores it in ./... patterns.
package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type request struct {
	Model     string    `json:"model"`
	MaxTokens int       `json:"max_tokens,omitempty"`
	System    string    `json:"system,omitempty"`
	Messages  []Message `json:"messages"`
}

type response struct {
	Content []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
	Usage struct {
		InputTokens  int `json:"input_tokens"`
		OutputTokens int `json:"output_tokens"`
	} `json:"usage"`
}

var mutation = os.Getenv("COURSE_MUTATION")

// knownMutations is every defect this program can express. A COURSE_MUTATION
// value outside this set is a typo, and a typo must not be allowed to look
// like a clean submission.
var knownMutations = map[string]bool{
	"nomaxtokens": true,
	"hardkey":     true,
	"hardmodel":   true,
	"noversion":   true,
	"nosystem":    true,
	"firstblock":  true,
	"fabricate":   true,
	"amnesiac":    true,
	"twocalls":    true,
	"useronly":    true,
	"chatty":      true,
	"nousage":     true,
	"fakeusage":   true,
}

// mutationHits counts how many times the selected mutation's branch was
// actually reached.
var mutationHits int

// mutating reports whether name is the selected mutation, recording that its
// branch was reached.
//
// This exists because of the failure mode that makes a mutation suite worse
// than useless: a mutation that does not apply. The program then behaves
// correctly, the grader says 100/100, and the suite reports that the grader
// "caught" a defect that was never actually introduced — a manufactured
// finding. Counting reaches lets main refuse to exit quietly in that case.
func mutating(name string) bool {
	if mutation != name {
		return false
	}
	mutationHits++
	return true
}

func main() {
	// "none" is the control: the mutant with no defect, which must score 100
	// or the other thirteen prove nothing. Normalise it to "no mutation" so
	// the guards below do not mistake the control for a typo.
	if mutation == "none" {
		mutation = ""
	}
	if mutation != "" && !knownMutations[mutation] {
		fmt.Fprintf(os.Stderr, "unknown COURSE_MUTATION %q\n", mutation)
		os.Exit(2)
	}
	base := strings.TrimSuffix(os.Getenv("ANTHROPIC_BASE_URL"), "/")
	key := os.Getenv("ANTHROPIC_API_KEY")
	model := os.Getenv("ANTHROPIC_MODEL")

	var conv []Message
	var inTok, outTok int
	out := json.NewEncoder(os.Stdout)

	send := func(msgs []Message) (string, error) {
		req := request{Model: model, MaxTokens: 1024, System: "You are helpful.", Messages: msgs}
		if mutating("nomaxtokens") {
			req.MaxTokens = 0
		}
		if mutating("hardmodel") {
			// Never reads ANTHROPIC_MODEL; names a real-looking model inline.
			req.Model = "claude-3-5-sonnet-20241022"
		}
		if mutating("nosystem") {
			// Sends no system prompt at all; `omitempty` drops the field.
			req.System = ""
		}
		body, _ := json.Marshal(req)
		hreq, err := http.NewRequest(http.MethodPost, base+"/v1/messages", bytes.NewReader(body))
		if err != nil {
			return "", err
		}
		hdrKey := key
		if mutating("hardkey") {
			hdrKey = "sk-ant-hardcoded-by-student"
		}
		hreq.Header.Set("x-api-key", hdrKey)
		hreq.Header.Set("content-type", "application/json")
		if !mutating("noversion") {
			hreq.Header.Set("anthropic-version", "2023-06-01")
		}
		resp, err := http.DefaultClient.Do(hreq)
		if err != nil {
			return "", err
		}
		defer resp.Body.Close()
		raw, _ := io.ReadAll(resp.Body)
		var parsed response
		if err := json.Unmarshal(raw, &parsed); err != nil {
			return "", err
		}
		inTok += parsed.Usage.InputTokens
		outTok += parsed.Usage.OutputTokens
		var text strings.Builder
		if mutating("firstblock") {
			// The classic day-one stumble: read content[0].text and ignore
			// the rest of the list.
			if len(parsed.Content) > 0 {
				text.WriteString(parsed.Content[0].Text)
			}
		} else {
			for _, b := range parsed.Content {
				text.WriteString(b.Text)
			}
		}
		return text.String(), nil
	}

	in := bufio.NewScanner(os.Stdin)
	for in.Scan() {
		line := strings.TrimSpace(in.Text())
		if line == "" {
			continue
		}
		var msg struct {
			User string `json:"user"`
		}
		if err := json.Unmarshal([]byte(line), &msg); err != nil {
			fmt.Fprintln(os.Stderr, "bad input:", err)
			os.Exit(1)
		}

		if mutating("fabricate") {
			// Never calls the API at all; answers from thin air.
			_ = out.Encode(map[string]string{"assistant": "Certainly! Here is your answer."})
			continue
		}

		conv = append(conv, Message{Role: "user", Content: msg.User})

		// What gets sent is the axis most of these mutations turn.
		payload := conv
		if mutating("amnesiac") {
			// A fresh single-message call every round: no conversation at all.
			payload = []Message{{Role: "user", Content: msg.User}}
		}

		if mutating("twocalls") {
			if _, err := send(payload); err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
		}

		reply, err := send(payload)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}

		if !mutating("useronly") {
			conv = append(conv, Message{Role: "assistant", Content: reply})
		}

		if mutating("chatty") {
			fmt.Println("DEBUG: round complete")
		}
		_ = out.Encode(map[string]string{"assistant": reply})
	}

	switch {
	case mutating("nousage"):
		// Reports nothing.
	case mutating("fakeusage"):
		_ = out.Encode(map[string]map[string]int{"usage": {"input": 1, "output": 1}})
	default:
		_ = out.Encode(map[string]map[string]int{"usage": {"input": inTok, "output": outTok}})
	}

	// A selected mutation that never reached its branch would leave this
	// program behaving exactly like a correct submission: the grader would
	// score it 100/100 and the mutation suite would record that as "the
	// grader caught nothing", when in truth nothing was ever broken. Refuse
	// to exit quietly.
	if mutation != "" && mutationHits == 0 {
		fmt.Fprintf(os.Stderr,
			"mutation %q was selected but never applied — the mutation suite would be reporting a fake result\n",
			mutation)
		os.Exit(3)
	}
}
