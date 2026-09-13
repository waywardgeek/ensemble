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

func main() {
	base := strings.TrimSuffix(os.Getenv("ANTHROPIC_BASE_URL"), "/")
	key := os.Getenv("ANTHROPIC_API_KEY")
	model := os.Getenv("ANTHROPIC_MODEL")

	var conv []Message
	var inTok, outTok int
	out := json.NewEncoder(os.Stdout)

	send := func(msgs []Message) (string, error) {
		req := request{Model: model, MaxTokens: 1024, System: "You are helpful.", Messages: msgs}
		if mutation == "nomaxtokens" {
			req.MaxTokens = 0
		}
		body, _ := json.Marshal(req)
		hreq, err := http.NewRequest(http.MethodPost, base+"/v1/messages", bytes.NewReader(body))
		if err != nil {
			return "", err
		}
		hdrKey := key
		if mutation == "hardkey" {
			hdrKey = "sk-ant-hardcoded-by-student"
		}
		hreq.Header.Set("x-api-key", hdrKey)
		hreq.Header.Set("content-type", "application/json")
		if mutation != "noversion" {
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
		if mutation == "firstblock" {
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

		if mutation == "fabricate" {
			// Never calls the API at all; answers from thin air.
			_ = out.Encode(map[string]string{"assistant": "Certainly! Here is your answer."})
			continue
		}

		conv = append(conv, Message{Role: "user", Content: msg.User})

		// What gets sent is the axis most of these mutations turn.
		payload := conv
		if mutation == "amnesiac" {
			// A fresh single-message call every round: no conversation at all.
			payload = []Message{{Role: "user", Content: msg.User}}
		}

		if mutation == "twocalls" {
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

		if mutation != "useronly" {
			conv = append(conv, Message{Role: "assistant", Content: reply})
		}

		if mutation == "chatty" {
			fmt.Println("DEBUG: round complete")
		}
		_ = out.Encode(map[string]string{"assistant": reply})
	}

	switch mutation {
	case "nousage":
		// Reports nothing.
	case "fakeusage":
		_ = out.Encode(map[string]map[string]int{"usage": {"input": 1, "output": 1}})
	default:
		_ = out.Encode(map[string]map[string]int{"usage": {"input": inTok, "output": outTok}})
	}
}
