package main

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

func runChat(owner ensemble.ClientOwner, agent *ensemble.Agent, input io.Reader, output io.Writer) error {
	reader := bufio.NewReader(input)
	writer := bufio.NewWriter(output)
	config := agent.Config()
	fmt.Fprintf(writer, "Ensemble — %s / %s\nType /help for commands.\n", config.Vendor, config.Model)
	for {
		fmt.Fprint(writer, "You> ")
		// Flush before reading so a person can see that the client is ready.
		if err := writer.Flush(); err != nil {
			return fmt.Errorf("cannot write chat output")
		}
		line, err := readChatLine(owner, reader)
		if err == io.EOF {
			fmt.Fprintln(writer)
			return showUsage(owner, writer, agent.Usage(), true)
		}
		if err != nil {
			return err
		}
		if strings.TrimSpace(line) == "" {
			continue
		}
		if strings.HasPrefix(line, "/") && !strings.HasPrefix(line, "//") {
			quit, err := chatCommand(owner, agent, writer, line)
			if err != nil {
				return err
			}
			if quit {
				return showUsage(owner, writer, agent.Usage(), true)
			}
			continue
		}
		if strings.HasPrefix(line, "//") {
			line = line[1:]
		}
		result, err := owner.Submit(context.Background(), ensemble.ClientRequest{AgentID: agent.ID(), Prompt: &line})
		if err != nil {
			return err
		}
		text := result.Text
		if text == "" {
			text = "[No text returned]"
		}
		fmt.Fprintf(writer, "Assistant:\n%s\n", text)
	}
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
	case "/help", "/usage", "/history", "/quit":
		if rest != "" {
			return localError()
		}
		switch command {
		case "/help":
			fmt.Fprintln(out, "/help — show commands\n/usage — token totals\n/history — event sequences and tool call IDs\n/ephemeral TEXT — one-request directive\n/redact FROM TO REASON — redact tool results in a sequence span\n/quit — finish the session\n//TEXT — submit a literal leading slash\nInput: one UTF-8 line, at most 1 MiB (1048576 bytes), excluding LF or CRLF.")
		case "/usage":
			return false, showUsage(owner, out, agent.Usage(), false)
		case "/history":
			fmt.Fprintln(out, "History:")
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
