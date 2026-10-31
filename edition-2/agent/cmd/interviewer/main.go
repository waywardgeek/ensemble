// The interviewer is an independent application of the public gateway library.
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"ensemble/gateway"
)

type interviewer struct {
	gateway.NopHandler
	answers chan gateway.Envelope
	ctx     context.Context
}

// OnDeliver accepts answers into the application-owned interview queue.
func (h *interviewer) OnDeliver(e gateway.Envelope) {
	select {
	case h.answers <- e:
	case <-h.ctx.Done():
	}
}
func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
func run() error {
	address := flag.String("gateway", "ws://127.0.0.1:8093/v1/ws", "gateway WebSocket URL")
	id := flag.String("agent", "interviewer", "configured interviewer identity")
	keyPath := flag.String("key", "interviewer.key", "private identity file")
	public := flag.String("gateway-key", "", "pinned gateway public key")
	groups := flag.String("groups", "agents", "comma-separated reply groups")
	subject := flag.String("subject", "ensemble", "agent to interview")
	report := flag.String("report", "interview.md", "markdown report path")
	timeout := flag.Duration("timeout", 2*time.Minute, "per-question deadline")
	printKey := flag.Bool("print-key", false, "print this identity's public key and exit")
	flag.Parse()
	key, err := gateway.LoadOrCreateKeypair(*keyPath)
	if err != nil {
		return err
	}
	if *printKey {
		fmt.Println(gateway.EncodePublicKey(key.Public))
		return nil
	}
	peer, err := gateway.ParsePublicKey(*public)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	h := &interviewer{answers: make(chan gateway.Envelope, 8), ctx: ctx}
	c, err := gateway.Connect(ctx, gateway.ClientConfig{GatewayURL: *address, AgentID: *id, StaticKey: key, GatewayPubKey: peer}, h)
	if err != nil {
		return err
	}
	defer c.Close()
	setup, stop := context.WithTimeout(ctx, *timeout)
	err = c.DeclareContactable(setup, strings.Split(*groups, ","))
	stop()
	if err != nil {
		return err
	}
	// Fixed questions make routing and correlation reproducible without asking
	// another model to invent the interview. Reported opinions remain answers
	// from this subject, not independently measured performance claims.
	questions := []string{"What slows you down most in your coding workflow?", "What burns the most tokens in your work?", "Where do you need a human's help?", "What new tool would help you most?"}
	answers := make([]string, len(questions))
	for i, q := range questions {
		// Bound each question separately. One lost answer cannot leave this
		// application waiting forever, and a successful early answer does not
		// consume the time budget for the next question.
		deadline, stop := context.WithTimeout(ctx, *timeout)
		conversation := fmt.Sprintf("interview-%d-%d", time.Now().UnixNano(), i)
		err = c.Send(deadline, gateway.OutboundMessage{ToAgent: *subject, Content: q, ConversationID: conversation, ResponseMode: gateway.ResponseAsync})
		if err != nil {
			stop()
			return err
		}
		waiting := true
		for waiting {
			select {
			case e := <-h.answers:
				// Other authorized messages are valid deliveries, but cannot answer this
				// question merely because they happened to arrive first.
				if e.ConversationID != conversation || e.Source.UserID != *subject {
					continue
				}
				text := e.Content
				if _, body, ok := strings.Cut(text, "\n\n"); ok {
					text = body
				}
				answers[i] = text
				waiting = false
			case <-deadline.Done():
				stop()
				return fmt.Errorf("question %d: %w", i+1, deadline.Err())
			case <-c.Done():
				stop()
				return fmt.Errorf("gateway disconnected during question %d", i+1)
			}
		}
		stop()
		fmt.Printf("Received answer %d of %d\n", i+1, len(questions))
	}
	var out strings.Builder
	out.WriteString("# Ensemble workflow interview\n\n## Friction points\n\n")
	for i := 0; i < 3; i++ {
		fmt.Fprintf(&out, "%s\n\n%s\n\n", questions[i], answers[i])
	}
	fmt.Fprintf(&out, "## Improvement opportunities\n\n%s\n\n## Raw transcript\n\n", answers[3])
	for i, q := range questions {
		fmt.Fprintf(&out, "### Question %d\n\n%s\n\n%s\n\n", i+1, q, answers[i])
	}
	return os.WriteFile(*report, []byte(out.String()), 0600)
}
