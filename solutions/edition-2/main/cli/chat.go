package cli

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"

	"example.com/ensemble"
)

const chatLineLimit = 1024 * 1024

// ErrQuit asks an embedding application to shut down after /quit.
var ErrQuit = fmt.Errorf("terminal requested application shutdown")

// Chat reuses the human CLI on an existing Agent. With detachEOF, EOF drains
// this client's accepted requests and leaves the Agent running; /quit returns ErrQuit.
func Chat(owner ensemble.ClientOwner, agent *ensemble.Agent, input io.Reader, output io.Writer, detachEOF bool) error {
	return runChatOptions(owner, agent, input, output, detachEOF)
}
func runChat(owner ensemble.ClientOwner, agent *ensemble.Agent, input io.Reader, output io.Writer) error {
	return runChatOptions(owner, agent, input, output, false)
}
func runChatOptions(owner ensemble.ClientOwner, agent *ensemble.Agent, input io.Reader, output io.Writer, detachEOF bool) error {
	quitRequested := false
	writer := bufio.NewWriter(output)
	progress, err := newProgress(owner, agent.ID())
	if err != nil {
		return err
	}
	defer owner.Unsubscribe(progress.subscription)
	show := func(o ensemble.Observation) error { return progress.chat(writer, o) }
	gap := func() error {
		fmt.Fprintln(writer, "\nIncomplete display: observation overflow; recovering from reliable completion.")
		return writer.Flush()
	}
	config := agent.Config()
	fmt.Fprintf(writer, "Ensemble — %s / %s\nType /help for commands.\n", config.Vendor, config.Model)
	done := make(chan struct{})
	defer close(done)
	lines := inputLines(owner, input, true, done)
	completed := make(chan completedRequest)
	pending := 0
	reading := true
	var fatal error
	prompt := func() error { fmt.Fprint(writer, "You> "); return writer.Flush() }
	if err := prompt(); err != nil {
		return err
	}
	finish := func() error {
		if !detachEOF {
			if err := agent.Close(); fatal == nil {
				fatal = err
			}
		}
		if quitRequested && detachEOF {
			return ErrQuit
		}
		if fatal != nil {
			return fatal
		}
		return showUsage(owner, writer, agent.Usage(), true)
	}
	for reading || pending > 0 {
		select {
		case item := <-lines:
			if item.err != nil {
				reading = false
				lines = nil
				if item.err != io.EOF {
					fatal = item.err
					_ = agent.Close()
				}
				continue
			}
			line := item.line
			if strings.TrimSpace(line) == "" {
				if err := prompt(); err != nil {
					return err
				}
				continue
			}
			if strings.HasPrefix(line, "/") && !strings.HasPrefix(line, "//") {
				command, rest := commandWord(owner, line)
				switch command {
				case "/hint":
					ack, err := agent.Hint(rest)
					if err != nil {
						fmt.Fprintf(writer, "Command error: %s.\n", err)
					} else {
						fmt.Fprintf(writer, "Hint received for %s at seq %d; sent=false (pending next request).\n", ack.RequestID, ack.Seq)
					}
				case "/interrupt":
					if rest != "" {
						fmt.Fprintln(writer, "Command error: /interrupt takes no text.")
					} else {
						ack, err := agent.Interrupt()
						if err != nil {
							return err
						}
						fmt.Fprintf(writer, "Interrupt: request=%s interrupted=%t.\n", ack.RequestID, ack.Interrupted)
					}
				default:
					quit, err := chatCommand(owner, agent, writer, line)
					if err != nil {
						if strings.Contains(err.Error(), "busy") {
							fmt.Fprintln(writer, "Command refused: Agent busy.")
						} else {
							fatal = err
							reading = false
							lines = nil
							_ = agent.Close()
						}
					}
					if quit {
						quitRequested = true
						reading = false
						lines = nil
						_ = agent.Close()
					}
				}
				if reading {
					if err := prompt(); err != nil {
						return err
					}
				} else {
					writer.Flush()
				}
				continue
			}
			if strings.HasPrefix(line, "//") {
				line = line[1:]
			}
			h, err := agent.Submit(line)
			if err != nil {
				return err
			}
			pending++
			fmt.Fprintf(writer, "Accepted %s.\n", h.ID())
			watchCompletion(h, completed, false)
			if err := prompt(); err != nil {
				return err
			}
		case <-progress.wake:
			if err := progress.drain(show, gap); err != nil {
				return err
			}
		case item := <-completed:
			if err := progress.beforeCompletion(agent, item.value, show, gap); err != nil {
				return err
			}
			pending--
			c := item.value
			text := c.Text
			if text == "" && c.Outcome == "success" {
				text = "[No text returned]"
			}
			fmt.Fprintf(writer, "\nRequest %s (%s; pending hints=%d)\n", c.RequestID, c.Outcome, c.PendingHints)
			if text != "" && (progress.reported || !progress.streamed[c.RequestID] || c.Text == "") {
				if progress.reported {
					fmt.Fprintln(writer, "Complete final answer (recovered):")
				}
				fmt.Fprintf(writer, "Assistant:\n%s\n", text)
			}
			switch c.StopReason {
			case "max_tokens", "length", "MAX_TOKENS":
				fmt.Fprintln(writer, "Generation limit reached; accepted answer may be incomplete.")
			}
			if c.Error != nil {
				fmt.Fprintf(writer, "%s: %s\n", c.Error.Code, c.Error.Message)
			}
			if c.Outcome == "error" || c.Outcome == "round_limit" || item.err != nil {
				fatal = completionError(item)
				reading = false
				lines = nil
				_ = agent.Close()
			}
			if reading {
				if err := prompt(); err != nil {
					return err
				}
			} else {
				writer.Flush()
			}
		}
	}
	return finish()
}

// ReadSlice bounds memory while allowing exactly the ceiling plus CRLF. A
// scanner's default split/limit would conflate data bytes and line endings.
func readChatLine(owner ensemble.ClientOwner, reader *bufio.Reader) (string, error) {
	var line []byte
	for {
		piece, err := reader.ReadSlice('\n')
		if len(line)+len(piece) > chatLineLimit+2 {
			return "", fmt.Errorf("input line exceeds 1 MiB limit")
		}
		line = append(line, piece...)
		if err == bufio.ErrBufferFull {
			continue
		}
		if err != nil && err != io.EOF {
			return "", fmt.Errorf("cannot read input")
		}
		if len(line) == 0 && err == io.EOF {
			return "", io.EOF
		}
		if len(line) > 0 && line[len(line)-1] == '\n' {
			line = line[:len(line)-1]
			if len(line) > 0 && line[len(line)-1] == '\r' {
				line = line[:len(line)-1]
			}
		}
		if len(line) > chatLineLimit {
			return "", fmt.Errorf("input line exceeds 1 MiB limit")
		}
		if !utf8.Valid(line) {
			return "", fmt.Errorf("input must be valid UTF-8")
		}
		return string(line), nil
	}
}

func commandWord(owner ensemble.ClientOwner, line string) (string, string) {
	i := strings.IndexFunc(line, unicode.IsSpace)
	if i < 0 {
		return line, ""
	}
	return line[:i], strings.TrimLeftFunc(line[i:], unicode.IsSpace)
}

func chatCommand(owner ensemble.ClientOwner, agent *ensemble.Agent, out *bufio.Writer, line string) (bool, error) {
	command, rest := commandWord(owner, line)
	localError := func() (bool, error) {
		fmt.Fprintln(out, "Command error: use /help for command syntax.")
		return false, nil
	}
	switch command {
	case "/help", "/usage", "/history", "/mcp", "/skills", "/session", "/checkpoint", "/quit":
		if rest != "" {
			return localError()
		}
		switch command {
		case "/session":
			s, err := agent.Session()
			if err != nil {
				return false, err
			}
			if s == nil {
				fmt.Fprintln(out, "Standalone fresh-log Agent.")
			} else {
				anchor := "none"
				if s.CheckpointSeq != nil {
					anchor = strconv.FormatUint(*s.CheckpointSeq, 10)
				}
				fmt.Fprintf(out, "Session %s; resumed=%t; durable sequence=%d; checkpoint=%s\n", s.ID, s.Resumed, agent.Snapshot().LastSeq, anchor)
			}
		case "/checkpoint":
			ack, err := agent.Checkpoint()
			if err != nil {
				fmt.Fprintf(out, "Checkpoint refused: %s\n", err)
			} else {
				fmt.Fprintf(out, "Checkpoint saved at %d.\n", ack.AsOf)
			}
		case "/help":
			fmt.Fprintln(out, "/session — session identity and saved boundary\n/checkpoint — save the settled session")
			fmt.Fprintln(out, "/help — show commands\n/usage — token totals\n/mcp — selected remote tools and connection status\n/skills — current skills and tool grants\n/history — event sequences and tool call IDs\n/hint TEXT — guide the next request of the active turn\n/interrupt — interrupt the active turn\n/ephemeral TEXT — one-request directive\n/redact FROM TO REASON — redact tool results in a sequence span\n/quit — finish the session\n//TEXT — submit a literal leading slash\nInput: one UTF-8 line, at most 1 MiB (1048576 bytes), excluding LF or CRLF.")
		case "/mcp":
 bindings,err:=agent.MCPState();if err!=nil{return false,err};if len(bindings)==0{fmt.Fprintln(out,"No selected MCP bindings.")};for _,b:=range bindings{generation:="none";if b.Generation!=nil{generation=fmt.Sprint(*b.Generation)};fmt.Fprintf(out,"%s: %s/%s — %s — generation %s — %s\n",b.Alias,b.Connection,b.RemoteName,b.State,generation,b.ErrorCode)}
 case "/skills":
			state, err := agent.SkillState()
			if err != nil {
				return false, err
			}
			if state == nil {
				fmt.Fprintln(out, "Skills disabled.")
				break
			}
			fmt.Fprintf(out, "Primary: %s\nRevision: %d\nRoots: %s\n", state.Primary, state.Revision, strings.Join(state.Roots, ", "))
			fmt.Fprintln(out, "Active:")
			for _, a := range state.Active {
				fmt.Fprintf(out, "  %s (%s, activation %d)\n", a.Name, a.Type, a.Activation)
			}
			fmt.Fprintln(out, "Available:")
			for _, o := range state.Available {
				fmt.Fprintf(out, "  %s: %s\n", o.Name, o.Description)
			}
			fmt.Fprintf(out, "Tools: %s\n", strings.Join(state.Tools, ", "))
		case "/usage":
			return false, showUsage(owner, out, agent.Usage(), false)
		case "/history":
			fmt.Fprintln(out, "History:")
			for _, e := range agent.Events() {
				if e.Type == "session_anchor" && e.Session != nil {
					fmt.Fprintf(out, "Snapshot origin at %d; only retained anchor/tail events follow.\n", e.Session.OriginAsOf)
					break
				}
			}
			hasResult := false
			for _, event := range agent.Events() {
				fmt.Fprintf(out, "%d %s", event.Seq, event.Type)
				if event.Tool != nil {
					fmt.Fprintf(out, " call_id=%s", event.Tool.CallID)
				}
				fmt.Fprintln(out)
				hasResult = hasResult || event.Type == "tool_returned"
			}
			if !hasResult {
				fmt.Fprintln(out, "No tool_returned events to redact in this session.")
			}
		case "/quit":
			return true, nil
		}
	case "/ephemeral", "/redact":
		request := ensemble.ClientRequest{AgentID: agent.ID()}
		if command == "/ephemeral" {
			if strings.TrimSpace(rest) == "" {
				return localError()
			}
			request.Ephemeral = &rest
		} else {
			first, rest := commandWord(owner, rest)
			last, reason := commandWord(owner, rest)
			from, err1 := strconv.ParseUint(first, 10, 64)
			to, err2 := strconv.ParseUint(last, 10, 64)
			if err1 != nil || err2 != nil || from == 0 || to < from || strings.TrimSpace(reason) == "" {
				return localError()
			}
			request.Redact = &ensemble.Redaction{From: from, To: to, Level: "redact_result", Reason: reason}
		}
		if _, err := owner.Submit(context.Background(), request); err != nil {
			return false, err
		}
		fmt.Fprintf(out, "Recorded %s.\n", command[1:])
	default:
		return localError()
	}
	return false, nil
}

func showUsage(owner ensemble.ClientOwner, out *bufio.Writer, usage ensemble.Usage, final bool) error {
	label := "Usage"
	if final {
		label = "Final usage"
	}
	fmt.Fprintf(out, "%s: input=%d, cache write=%d, cache read=%d, output=%d\n", label, usage.Input, usage.CacheWrite, usage.CacheRead, usage.Output)
	if err := out.Flush(); err != nil {
		return fmt.Errorf("cannot write chat output")
	}
	return nil
}
