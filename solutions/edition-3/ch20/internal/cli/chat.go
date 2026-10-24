package cli

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strings"

	"ensemble"
)

type chatLine struct {
	text string
	err  error
}

func chatInput(in io.Reader, lines chan<- chatLine, done <-chan struct{}) {
	defer close(lines)
	r := bufio.NewReader(in)
	for {
		line, err := r.ReadString('\n')
		if len(line) > 0 {
			line = strings.TrimSuffix(strings.TrimSuffix(line, "\n"), "\r")
			select {
			case lines <- chatLine{text: line}:
			case <-done:
				return
			}
		}
		if err != nil {
			if err != io.EOF {
				select {
				case lines <- chatLine{err: err}:
				case <-done:
				}
			}
			return
		}
	}
}
func runChat(ctx context.Context, a ensemble.Agent, in io.Reader, out, diagnostics io.Writer) error {
	lines := make(chan chatLine)
	done := make(chan struct{})
	defer close(done)
	humanQuestions := make(chan ensemble.Question, 16)
	root, _ := a.Parent().(*ensemble.Ensemble)
	if root != nil {
		root.SetHumanQuestionSink(func(q ensemble.Question) {
			select {
			case humanQuestions <- q:
			case <-done:
			}
		})
		defer root.SetHumanQuestionSink(nil)
	}
	go chatInput(in, lines, done)
	interrupts := make(chan os.Signal, 1)
	signal.Notify(interrupts, os.Interrupt)
	defer signal.Stop(interrupts)
	progress := a.Observe()
	defer progress.Close()
	type partKey struct{ turn, response, part uint64 }
	streamed := make(map[partKey]bool)
	shown := make(map[uint64]bool)
	thinking := false
	writeText := func(turn uint64, text string) error {
		if !shown[turn] {
			shown[turn] = true
			_, err := fmt.Fprint(out, "\nassistant> "+text)
			return err
		}
		_, err := fmt.Fprint(out, text)
		return err
	}
	drain := func() error {
		for _, observation := range progress.Drain() {
			switch v := observation.(type) {
			case ensemble.BillingSelected:
				fmt.Fprintf(diagnostics, "\nfunding> %s\n", v.Mode)
			case ensemble.PartDelta:
				key := partKey{v.TurnID, v.ResponseID, v.PartID}
				firstDelta := !streamed[key]
				streamed[key] = true
				switch v.Kind {
				case "text":
					if err := writeText(v.TurnID, v.Chunk); err != nil {
						return err
					}
				case "tool_call":
					if firstDelta {
						fmt.Fprint(diagnostics, "\ntool> ")
					}
					fmt.Fprint(diagnostics, v.Chunk)
				case "thinking", "reasoning_summary":
					if !thinking {
						fmt.Fprint(diagnostics, "\nthinking> ")
						thinking = true
					}
					fmt.Fprint(diagnostics, v.Chunk)
				}
				a.Parent().Trace(ensemble.TraceRecord{Stage: "cli_delta", TurnID: v.TurnID, ResponseID: v.ResponseID, PartID: v.PartID, Kind: string(v.Kind), Bytes: len(v.Chunk)})
			case ensemble.PartFinal:
				key := partKey{v.TurnID, v.ResponseID, v.PartID}
				switch v.Part.Type {
				case "text":
					if !streamed[key] {
						if err := writeText(v.TurnID, v.Part.Text); err != nil {
							return err
						}
					}
				case "thinking", "reasoning_summary":
					if !streamed[key] {
						fmt.Fprintln(diagnostics, "\nthinking> "+v.Part.Text)
					}
				case "tool_call":
					if !streamed[key] {
						fmt.Fprintf(diagnostics, "\ntool> %s %s\n", v.Part.Name, v.Part.Args)
					} else {
						fmt.Fprintln(diagnostics)
					}
				case "tool_result":
					for _, p := range v.Part.Parts {
						if p.Type == "text" {
							fmt.Fprintln(diagnostics, p.Text)
						}
					}
				}
				a.Parent().Trace(ensemble.TraceRecord{Stage: "cli_final", TurnID: v.TurnID, ResponseID: v.ResponseID, PartID: v.PartID, Kind: v.Part.Type})
			case ensemble.HintObserved:
				if v.Sender == "human" {
					fmt.Fprintln(diagnostics, "hint queued: "+hintText(v.Parts))
				} else {
					fmt.Fprintln(diagnostics, "message queued from "+v.Sender)
				}
			}
		}
		return nil
	}
	var pending []<-chan ensemble.Result
	var printedTurn uint64
	fmt.Fprintln(diagnostics, "Chat ready — /interrupt or Ctrl-C ends the turn; Ctrl-D quits")
	fmt.Fprint(diagnostics, "you> ")
	for lines != nil || len(pending) > 0 {
		var reply <-chan ensemble.Result
		if len(pending) > 0 {
			reply = pending[0]
		}
		select {
		case question := <-humanQuestions:
			fmt.Fprintf(out, "Human question from %s [%s]: %s\nReply: /reply %s RESPONSE\n", question.Sender, question.RequestID, question.Text, question.RequestID)
		case line, ok := <-lines:
			if !ok {
				lines = nil
				continue
			}
			if line.err != nil {
				return line.err
			}
			text := line.text
			if strings.TrimSpace(text) == "" {
				continue
			}
			if (text == "/reply" || strings.HasPrefix(text, "/reply ")) && root != nil {
				if err := replyToHuman(root, strings.TrimSpace(strings.TrimPrefix(text, "/reply"))); err != nil {
					fmt.Fprintln(out, err)
				}
			} else if strings.HasPrefix(text, "/answer ") && root != nil {
				parts := strings.SplitN(strings.TrimPrefix(text, "/answer "), " ", 3)
				if len(parts) != 3 {
					fmt.Fprintln(out, "usage: /answer AGENT REQUEST RESPONSE")
				} else if err := root.RespondToHumanQuestion(parts[0], parts[1], parts[2]); err != nil {
					fmt.Fprintln(out, err)
				}
			} else if text == "/help" {
				fmt.Fprintln(out, "/status /model [name] /save /interrupt /reply [MESSAGE_ID] TEXT /help. Launch flags (before chat): --safe-mode --no-web-search --sandbox PATH.")
			} else if text == "/status" {
				cfg := a.Config()
				fmt.Fprintf(out, "safe=%t web-search=%t sandbox-root=%q\n", cfg.SafeMode, cfg.EnableWebSearch, cfg.SandboxRoot)
				_ = json.NewEncoder(out).Encode(a.Engine().Status())
			} else if text == "/model" {
				fmt.Fprintln(diagnostics, "model:", a.Config().Model)
			} else if strings.HasPrefix(text, "/model ") {
				if err := a.SetModel(strings.TrimSpace(strings.TrimPrefix(text, "/model "))); err != nil {
					fmt.Fprintln(diagnostics, "model:", err)
				} else {
					fmt.Fprintln(diagnostics, "model:", a.Config().Model)
				}
			} else if text == "/save" {
				if err := a.Save(); err != nil {
					fmt.Fprintln(diagnostics, "save:", err)
				} else {
					fmt.Fprintln(diagnostics, "save complete (if configured)")
				}
			} else if text == "/interrupt" {
				if err := a.Post(ensemble.Interrupt{}); err != nil {
					return err
				}
				fmt.Fprintln(diagnostics, "interrupt: turn ending; jobs remain available")
			} else {
				// Only the actor knows whether this is idle input or a live hint.
				// Keep every private completion so EOF also drains newly started turns.
				pending = append(pending, a.SubmitLive(ctx, text))
				fmt.Fprintln(diagnostics, "input queued: "+text)
			}
			fmt.Fprint(diagnostics, "you> ")
		case <-interrupts:
			if err := a.Post(ensemble.Interrupt{}); err != nil {
				return err
			}
			fmt.Fprintln(diagnostics, "interrupt: turn ending; jobs remain available")
		case result := <-reply:
			if err := drain(); err != nil {
				return err
			}
			pending = pending[1:]
			if result.TurnID != 0 && result.TurnID == printedTurn {
				continue
			}
			if result.Err != nil {
				if !errors.Is(result.Err, ensemble.ErrInterrupted) {
					return result.Err
				}
				fmt.Fprintln(diagnostics, "turn interrupted")
			} else {
				if !shown[result.TurnID] {
					if err := writeText(result.TurnID, result.Text); err != nil {
						return err
					}
				}
				if _, err := fmt.Fprint(out, "\n\n"); err != nil {
					return err
				}
			}
			if thinking {
				fmt.Fprintln(diagnostics)
				thinking = false
			}
			delete(shown, result.TurnID)
			for key := range streamed {
				if key.turn == result.TurnID {
					delete(streamed, key)
				}
			}
			printedTurn = result.TurnID
			u := a.Usage()
			fmt.Fprintf(diagnostics, "[%d turns | %d input tokens | %d output tokens]\n", len(a.History()), u.Input, u.Output)
		case <-progress.Wake():
			if err := drain(); err != nil {
				return err
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	return nil
}

func hintText(parts ensemble.PartList) string {
	var b strings.Builder
	for _, p := range parts {
		if p.Type == "text" {
			b.WriteString(p.Text)
		}
	}
	return b.String()
}

// Parsing never falls through to ordinary chat: stale/ambiguous answers are errors.
func replyToHuman(root *ensemble.Ensemble, text string) error {
	if text == "" {
		return fmt.Errorf("usage: /reply [MESSAGE_ID] TEXT")
	}
	first, rest, hasRest := strings.Cut(text, " ")
	marker := strings.LastIndex(first, "-question-")
	isID := marker > 0 && marker+len("-question-") < len(first)
	if isID {
		for _, c := range first[marker+len("-question-"):] {
			if c < '0' || c > '9' {
				isID = false
				break
			}
		}
	}
	if isID {
		if !hasRest || strings.TrimSpace(rest) == "" {
			return fmt.Errorf("usage: /reply MESSAGE_ID TEXT")
		}
		return root.ReplyToHuman(first, strings.TrimSpace(rest))
	}
	return root.ReplyToHuman("", text)
}
