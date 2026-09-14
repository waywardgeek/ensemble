package main

// The tools. Six from Chapter 3, now dispatched as jobs, plus the four that
// supervise jobs.
//
// A tool is still a name, a description of its arguments, and a function from
// JSON to text. What changed is not the tools but the DISPATCH: every call to
// one of the six gets a handle, an output file and a status before it runs,
// and the caller waits on it rather than in it. Five of the six did not
// change at all beyond accepting a *Call they ignore. `run_command` changed,
// because it is the one tool whose output arrives over time, and streaming it
// into the job is what makes it watchable.

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/creack/pty"
)

// Call is what a tool is handed besides its arguments. Job is the job this
// call IS, for the tools that are jobs; it is nil for the supervision tools,
// which act on OTHER jobs through Jobs. Limits is what the dispatcher will
// wait on, resolved before the tool ran, so a supervision tool waits by the
// same rules the dispatcher does. Events is where a tool may leave records
// for the dispatcher to append after its result — kill_job uses it.
type Call struct {
	Job    *Job
	Jobs   *Jobs
	Limits Limits
	Events []Event
}

// ToolFunc executes one call. It returns the text the model will see.
//
// The error return is the tool saying "this did not work". It is NOT a crash
// and it is NOT the end of the turn: the caller turns a non-nil error into a
// tool_result marked as an error and hands it straight back to the model.
// Failure is a result, not an absence.
type ToolFunc func(c *Call, args json.RawMessage) (string, error)

type Tool struct {
	Name string
	// Description is read by the MODEL, not by a person. It is the only thing
	// that tells the model when to reach for this tool instead of another one,
	// so it says what the tool is FOR, not how it is implemented.
	Description string
	// Schema is the JSON Schema of the arguments object, exactly as it goes on
	// the wire. Vendor-neutral: all three vendors take this subset verbatim.
	Schema json.RawMessage
	Run    ToolFunc
	// NoJob marks the tools that do not become jobs: the ones that supervise
	// jobs, and tool_limits. They run on the dispatcher's goroutine and
	// return when they return. Everything else gets a handle.
	NoJob bool
}

// argSpec is a human-readable summary of each tool's arguments, kept as plain
// data so that an error message can describe a tool without depending on the
// registry that holds the tool's code.
var argSpec = map[string]string{
	"run_command":    `{"command":string,"ai_callback_delay":number?,"ai_callback_pattern":string?,"max_output_bytes":int?}`,
	"read_file":      `{"path":string,"start_line":int?,"end_line":int?,"max_bytes":int?}`,
	"write_file":     `{"path":string,"content":string,"append":bool?}`,
	"edit_file":      `{"path":string,"old_text":string,"new_text":string}`,
	"list_directory": `{"path":string?}`,
	"search_files":   `{"pattern":string,"path":string?,"file_pattern":string?}`,
	"wait_for_job":   `{"handle":int,"ai_callback_delay":number?,"ai_callback_pattern":string?,"max_output_bytes":int?}`,
	"send_input":     `{"handle":int,"input":string,"append_newline":bool?,"ai_callback_delay":number?,"ai_callback_pattern":string?,"max_output_bytes":int?}`,
	"kill_job":       `{"handle":int}`,
	"tool_limits":    `{"ai_callback_delay":number?,"ai_callback_pattern":string?,"max_output_bytes":int?}`,
}

// limitProps is the schema text for the three limit arguments, shared by
// every tool that waits on a job so the model sees one spelling of them.
const limitProps = `"ai_callback_delay":{"type":"number","description":"Seconds to wait before returning with whatever output exists so far. The job keeps running. Default 3."},
			"ai_callback_pattern":{"type":"string","description":"Return as soon as the output not yet shown to you matches this Go regular expression, e.g. a prompt such as \"\\(dlv\\) \"."},
			"max_output_bytes":{"type":"integer","description":"Largest result to return inline; more than this is cut to head and tail with the full output left in the job's file. Default 16384."}`

// Registry is the agent's entire capability surface, and therefore its
// permission boundary. A capability you do not put in this map is one the
// model cannot reach — which is the argument for having tools at all rather
// than only a shell.
//
// Each entry is also the model's only documentation of the tool. The schemas
// use the subset of JSON Schema every vendor accepts unmodified — object,
// properties, required, per-property type and description — and nothing
// else. `additionalProperties`, `$schema`, `format` and friends are where
// the vendors disagree, so they stay out.
var Registry = map[string]Tool{
	"run_command": {
		Name:        "run_command",
		Description: "Run a shell command in the working directory under a terminal. Returns its output so far and, when it has exited, its exit status; if it is still running after ai_callback_delay you get a job handle to wait on, talk to, or kill. Use for builds, tests, debuggers, and anything the other tools cannot do.",
		Schema: json.RawMessage(`{"type":"object","properties":{
			"command":{"type":"string","description":"The command line to run, as you would type it in a shell."},
			` + limitProps + `},
			"required":["command"]}`),
		Run: toolRunCommand,
	},
	"wait_for_job": {
		Name:        "wait_for_job",
		Description: "Wait for a job to finish, or until ai_callback_delay passes or ai_callback_pattern appears in its new output. Returns the output you have not yet seen and the job's status. Works on a job that has already finished.",
		Schema: json.RawMessage(`{"type":"object","properties":{
			"handle":{"type":"integer","description":"The job handle from an earlier tool result."},
			` + limitProps + `},
			"required":["handle"]}`),
		Run:   toolWaitForJob,
		NoJob: true,
	},
	"send_input": {
		Name:        "send_input",
		Description: "Write to the standard input of a running job, then wait as wait_for_job does and return what it said in reply. Use for debuggers, REPLs, and anything that prompts.",
		Schema: json.RawMessage(`{"type":"object","properties":{
			"handle":{"type":"integer","description":"The job handle of a running process."},
			"input":{"type":"string","description":"The text to send."},
			"append_newline":{"type":"boolean","description":"Send a newline after the text, as pressing Enter would. Default true."},
			` + limitProps + `},
			"required":["handle","input"]}`),
		Run:   toolSendInput,
		NoJob: true,
	},
	"kill_job": {
		Name:        "kill_job",
		Description: "Stop a running job. Its process and everything the process started are killed; its output so far stays in its file.",
		Schema: json.RawMessage(`{"type":"object","properties":{
			"handle":{"type":"integer","description":"The job handle to kill."}},
			"required":["handle"]}`),
		Run:   toolKillJob,
		NoJob: true,
	},
	"tool_limits": {
		Name:        "tool_limits",
		Description: "Set ai_callback_delay, ai_callback_pattern and max_output_bytes for the NEXT tool call only, for tools whose own arguments do not include them. run_command, wait_for_job and send_input take these directly; call this before other tools.",
		Schema: json.RawMessage(`{"type":"object","properties":{
			` + limitProps + `}}`),
		Run:   toolLimits,
		NoJob: true,
	},
	"read_file": {
		Name:        "read_file",
		Description: "Read a text file, or an inclusive range of lines from it. Ask for a range when the file is large; the result is truncated after max_bytes.",
		Schema: json.RawMessage(`{"type":"object","properties":{
			"path":{"type":"string","description":"Path to the file, relative to the working directory."},
			"start_line":{"type":"integer","description":"First line to return, 1-based. Defaults to the start of the file."},
			"end_line":{"type":"integer","description":"Last line to return, inclusive. Defaults to the end of the file."},
			"max_bytes":{"type":"integer","description":"Truncate the result after this many bytes."}},
			"required":["path"]}`),
		Run: toolReadFile,
	},
	"write_file": {
		Name:        "write_file",
		Description: "Create or overwrite a file with the given content. Set append to add to the end instead of replacing.",
		Schema: json.RawMessage(`{"type":"object","properties":{
			"path":{"type":"string","description":"Path to the file, relative to the working directory."},
			"content":{"type":"string","description":"The complete new content, or the text to append."},
			"append":{"type":"boolean","description":"Append instead of overwrite."}},
			"required":["path","content"]}`),
		Run: toolWriteFile,
	},
	"edit_file": {
		Name:        "edit_file",
		Description: "Replace one exact occurrence of old_text in a file with new_text. Refuses if old_text is absent or matches more than once — include enough context to make it unique.",
		Schema: json.RawMessage(`{"type":"object","properties":{
			"path":{"type":"string","description":"Path to the file, relative to the working directory."},
			"old_text":{"type":"string","description":"The exact text to find. Must occur exactly once."},
			"new_text":{"type":"string","description":"The text to put in its place."}},
			"required":["path","old_text","new_text"]}`),
		Run: toolEditFile,
	},
	"list_directory": {
		Name:        "list_directory",
		Description: "List the entries of a directory, one per line: subdirectories with a trailing slash, files with their size in bytes.",
		Schema: json.RawMessage(`{"type":"object","properties":{
			"path":{"type":"string","description":"Directory to list. Defaults to the working directory."}}}`),
		Run: toolListDirectory,
	},
	"search_files": {
		Name:        "search_files",
		Description: "Search files for a regular expression and return matching lines as path:line: text.",
		Schema: json.RawMessage(`{"type":"object","properties":{
			"pattern":{"type":"string","description":"Go regular expression to search for."},
			"path":{"type":"string","description":"Directory to search under. Defaults to the working directory."},
			"file_pattern":{"type":"string","description":"Glob restricting which file names are searched, e.g. *.go."}},
			"required":["pattern"]}`),
		Run: toolSearchFiles,
	},
}

// Declarations is the registry as the model will be told about it: the
// request-direction half of the tool protocol, in the same stable order as
// ToolNames. It is the ONLY place the agent converts "a tool I can run" into
// "a tool the model may ask for", which is what makes the registry the
// permission boundary rather than just a lookup table.
//
// An empty registry yields a nil slice, and a nil slice renders to no field.
func Declarations() []ToolDecl {
	var decls []ToolDecl
	for _, n := range ToolNames() {
		t := Registry[n]
		decls = append(decls, ToolDecl{
			Name:        t.Name,
			Description: t.Description,
			Schema:      json.RawMessage(compactJSON(t.Schema)),
		})
	}
	return decls
}

// compactJSON strips the whitespace the source literals use for readability
// so the bytes on the wire are canonical.
func compactJSON(raw json.RawMessage) []byte {
	var buf bytes.Buffer
	if err := json.Compact(&buf, raw); err != nil {
		panic("tool schema is not valid JSON: " + err.Error())
	}
	return buf.Bytes()
}

// ToolNames returns the registry in a stable order. Map iteration order is
// random in Go, and a tool schema that reorders itself between requests
// changes the prompt bytes for no reason — which defeats prompt caching and
// makes the byte-identity tests of Chapter 2 flap.
func ToolNames() []string {
	names := make([]string, 0, len(Registry))
	for n := range Registry {
		names = append(names, n)
	}
	sort.Strings(names)
	return names
}

// Lookup finds a tool by name. An unknown name is an ERROR RESULT at the
// dispatch site, not a panic and not a silent skip: the model asked for
// something that does not exist, and the useful reply says what does.
func Lookup(name string) (Tool, error) {
	tool, ok := Registry[name]
	if !ok {
		return Tool{}, fmt.Errorf("no such tool %q; available tools: %s",
			name, strings.Join(ToolNames(), ", "))
	}
	return tool, nil
}

// decode parses a tool's arguments.
//
// Arguments come from a language model, so malformed JSON is an ordinary
// Tuesday rather than an exceptional condition. The error names the tool so
// the model can tell which of several calls it got wrong.
func decode(name string, args json.RawMessage, into any) error {
	if len(args) == 0 {
		return fmt.Errorf("%s: no arguments given, want %s", name, argSpec[name])
	}
	dec := json.NewDecoder(strings.NewReader(string(args)))
	dec.DisallowUnknownFields()
	if err := dec.Decode(into); err != nil {
		// Retry permissively: an unknown field is the model being chatty, not
		// the model being wrong, and refusing the whole call over a spurious
		// key costs a round trip to discover.
		if err2 := json.Unmarshal(args, into); err2 != nil {
			return fmt.Errorf("%s: arguments did not parse: %v (want %s)", name, err2, argSpec[name])
		}
	}
	return nil
}

// --- run_command -----------------------------------------------------------

// toolRunCommand starts a shell command under a pseudo-terminal and streams
// everything it says into the job as it says it.
//
// A terminal, not pipes, because the programs worth supervising behave
// differently without one: debuggers and REPLs block-buffer their prompts, so
// the "(dlv) " you are waiting for never arrives, and pagers wait for a
// keypress nobody will send. The costs are real and are stated rather than
// hidden — stdout and stderr are one stream, input you send is echoed back
// into the output, and lines end in \r\n, which is normalized to \n here.
// If a terminal cannot be allocated that is an error, not a quiet fallback
// to pipes; a job the model believes is interactive and is not would fail in
// a way that looks like the program's fault.
//
// This function BLOCKS until the process exits. It does not know about the
// delay, the pattern or the handle: the dispatcher is waiting on the job it
// is writing into, and returns to the model without it when the delay
// passes. Nothing in here was made asynchronous. It was made observable.
//
// A non-zero exit is NOT an error return. The command ran; the agent asked a
// question and got an answer, and "the tests failed" is the answer. Marking it
// as a tool error would tell the model its CALL was malformed, which is a
// different and false claim. Errors are reserved for "I could not run this."
func toolRunCommand(c *Call, args json.RawMessage) (string, error) {
	var a struct {
		Command string `json:"command"`
		limitArgs
	}
	if err := decode("run_command", args, &a); err != nil {
		return "", err
	}
	if strings.TrimSpace(a.Command) == "" {
		return "", fmt.Errorf("run_command: empty command")
	}

	cmd := exec.Command("sh", "-c", a.Command)
	cmd.Env = append(os.Environ(), "TERM=dumb")
	// pty.Start puts the child in its own session with the pty as its
	// controlling terminal. Its own session means its own process group, so
	// kill_job can take down everything it started with one signal.
	f, err := pty.StartWithSize(cmd, &pty.Winsize{Rows: 50, Cols: 200})
	if err != nil {
		return "", fmt.Errorf("run_command: could not start under a terminal: %v", err)
	}
	c.Job.Attach(cmd.Process, f)

	// Copy until the terminal closes. On Linux that is an EIO once the last
	// process holding the slave side exits; on macOS it is EOF. Either way the
	// stream is over and the only thing left to learn is the exit status.
	buf := make([]byte, 32*1024)
	for {
		n, rerr := f.Read(buf)
		if n > 0 {
			_, _ = c.Job.Write(bytes.ReplaceAll(buf[:n], []byte("\r\n"), []byte("\n")))
		}
		if rerr != nil {
			break
		}
	}
	_ = f.Close()

	werr := cmd.Wait()
	if c.Job.Status() == StatusKilled {
		// kill_job got here first. The model was already told; the exit
		// status of a process we killed is not news.
		return "", nil
	}
	exitCode := 0
	if ee, ok := werr.(*exec.ExitError); ok {
		exitCode = ee.ExitCode()
	} else if werr != nil {
		return "", fmt.Errorf("run_command: %v", werr)
	}
	c.Job.SetExit(exitCode)
	fmt.Fprintf(c.Job, "exit_code: %d\n", exitCode)
	return "", nil
}

// --- read_file -------------------------------------------------------------

const defaultMaxBytes = 64 * 1024

// toolReadFile reads a file, or a range of its lines.
//
// The line range and the size cap are the point of the tool, not a nicety.
// `cat` on a four-thousand-line file floods the context window, and the
// student pays for those tokens again on every subsequent turn of the
// conversation. The tool that reads is also the tool that decides how much of
// the window to spend.
func toolReadFile(_ *Call, args json.RawMessage) (string, error) {
	var a struct {
		Path      string `json:"path"`
		StartLine int    `json:"start_line"`
		EndLine   int    `json:"end_line"`
		MaxBytes  int    `json:"max_bytes"`
	}
	if err := decode("read_file", args, &a); err != nil {
		return "", err
	}
	if a.Path == "" {
		return "", fmt.Errorf("read_file: path is required")
	}
	data, err := os.ReadFile(a.Path)
	if err != nil {
		return "", fmt.Errorf("read_file: %v", err)
	}

	text := string(data)
	if a.StartLine > 0 || a.EndLine > 0 {
		lines := strings.Split(text, "\n")
		start := a.StartLine
		if start <= 0 {
			start = 1
		}
		end := a.EndLine
		if end <= 0 || end > len(lines) {
			end = len(lines)
		}
		if start > len(lines) {
			return "", fmt.Errorf("read_file: %s has %d lines, start_line %d is past the end",
				a.Path, len(lines), a.StartLine)
		}
		if end < start {
			return "", fmt.Errorf("read_file: end_line %d is before start_line %d", a.EndLine, start)
		}
		text = strings.Join(lines[start-1:end], "\n")
	}

	cap := a.MaxBytes
	if cap <= 0 {
		cap = defaultMaxBytes
	}
	if len(text) > cap {
		text = text[:cap] + fmt.Sprintf("\n[truncated at %d bytes]", cap)
	}
	return text, nil
}

// --- write_file ------------------------------------------------------------

func toolWriteFile(_ *Call, args json.RawMessage) (string, error) {
	var a struct {
		Path    string `json:"path"`
		Content string `json:"content"`
		Append  bool   `json:"append"`
	}
	if err := decode("write_file", args, &a); err != nil {
		return "", err
	}
	if a.Path == "" {
		return "", fmt.Errorf("write_file: path is required")
	}
	if dir := filepath.Dir(a.Path); dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return "", fmt.Errorf("write_file: %v", err)
		}
	}
	if a.Append {
		f, err := os.OpenFile(a.Path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
		if err != nil {
			return "", fmt.Errorf("write_file: %v", err)
		}
		defer f.Close()
		if _, err := f.WriteString(a.Content); err != nil {
			return "", fmt.Errorf("write_file: %v", err)
		}
		return fmt.Sprintf("appended %d bytes to %s", len(a.Content), a.Path), nil
	}
	if err := os.WriteFile(a.Path, []byte(a.Content), 0o644); err != nil {
		return "", fmt.Errorf("write_file: %v", err)
	}
	return fmt.Sprintf("wrote %d bytes to %s", len(a.Content), a.Path), nil
}

// --- edit_file -------------------------------------------------------------

// toolEditFile replaces an exact anchor with new text.
//
// THIS IS THE CHAPTER'S DECLINED DECISION, and this file is where this
// solution makes its choice. When the anchor does not match, three answers are
// defensible: refuse, fuzzy-match, or rewrite the file.
//
// THIS SOLUTION REFUSES. The file is not touched, and the error says what was
// looked for, in which file, and what was found instead — because a refusal
// that does not say what it saw is only half a loud failure: it declines to
// guess, and then costs a round trip to find out why.
//
// The reasoning, for the record, is that an edit is a claim about the current
// contents of a file. If the claim is false, the model's belief about the file
// is stale, and the cheapest correct move is to say so and let it re-read. A
// fuzzy match applied to a stale belief writes the right text in the wrong
// place, and that failure is silent.
//
// Ambiguity is refused for the same reason: an anchor matching three places
// does not identify an edit site. Guessing the first is a coin flip the model
// cannot see being tossed.
func toolEditFile(_ *Call, args json.RawMessage) (string, error) {
	var a struct {
		Path    string `json:"path"`
		OldText string `json:"old_text"`
		NewText string `json:"new_text"`
	}
	if err := decode("edit_file", args, &a); err != nil {
		return "", err
	}
	if a.Path == "" {
		return "", fmt.Errorf("edit_file: path is required")
	}
	if a.OldText == "" {
		return "", fmt.Errorf("edit_file: old_text is required; to create or replace a whole file use write_file")
	}
	data, err := os.ReadFile(a.Path)
	if err != nil {
		return "", fmt.Errorf("edit_file: %v", err)
	}
	text := string(data)

	switch n := strings.Count(text, a.OldText); {
	case n == 0:
		return "", fmt.Errorf("edit_file: refused: old_text not found in %s.\n"+
			"looked for:\n%s\nThe file was not modified. Re-read %s and retry with text that matches it exactly.",
			a.Path, quoteAnchor(a.OldText), a.Path)
	case n > 1:
		return "", fmt.Errorf("edit_file: refused: old_text appears %d times in %s, so it does not identify one place.\n"+
			"looked for:\n%s\nThe file was not modified. Include more surrounding context to make the anchor unique.",
			n, a.Path, quoteAnchor(a.OldText))
	}

	updated := strings.Replace(text, a.OldText, a.NewText, 1)
	if err := os.WriteFile(a.Path, []byte(updated), 0o644); err != nil {
		return "", fmt.Errorf("edit_file: %v", err)
	}
	return fmt.Sprintf("edited %s: replaced %d bytes with %d bytes",
		a.Path, len(a.OldText), len(a.NewText)), nil
}

// quoteAnchor renders the anchor the tool was given, bounded, so a refusal is
// diagnosable without pasting an entire file back into the context window.
func quoteAnchor(s string) string {
	const max = 400
	if len(s) > max {
		s = s[:max] + "..."
	}
	return "---\n" + s + "\n---"
}

// --- list_directory --------------------------------------------------------

// toolListDirectory is the one tool in this set justified by judgement rather
// than by the measurement: it is well under one percent of real calls. It
// stays because orientation is cheap, and an agent that cannot see the tree
// guesses at paths.
func toolListDirectory(_ *Call, args json.RawMessage) (string, error) {
	var a struct {
		Path string `json:"path"`
	}
	if err := decode("list_directory", args, &a); err != nil {
		return "", err
	}
	if a.Path == "" {
		a.Path = "."
	}
	entries, err := os.ReadDir(a.Path)
	if err != nil {
		return "", fmt.Errorf("list_directory: %v", err)
	}
	var b strings.Builder
	fmt.Fprintf(&b, "%s:\n", a.Path)
	for _, e := range entries {
		if e.IsDir() {
			fmt.Fprintf(&b, "dir   %s/\n", e.Name())
			continue
		}
		size := int64(-1)
		if info, err := e.Info(); err == nil {
			size = info.Size()
		}
		fmt.Fprintf(&b, "file  %s (%d bytes)\n", e.Name(), size)
	}
	return b.String(), nil
}

// --- search_files ----------------------------------------------------------

const maxMatches = 200

// toolSearchFiles is what makes read_file usable on a codebase bigger than one
// directory. An agent that cannot grep cannot find what to read.
func toolSearchFiles(_ *Call, args json.RawMessage) (string, error) {
	var a struct {
		Pattern     string `json:"pattern"`
		Path        string `json:"path"`
		FilePattern string `json:"file_pattern"`
	}
	if err := decode("search_files", args, &a); err != nil {
		return "", err
	}
	if a.Pattern == "" {
		return "", fmt.Errorf("search_files: pattern is required")
	}
	re, err := regexp.Compile(a.Pattern)
	if err != nil {
		return "", fmt.Errorf("search_files: bad pattern %q: %v", a.Pattern, err)
	}
	root := a.Path
	if root == "" {
		root = "."
	}
	if _, err := os.Stat(root); err != nil {
		return "", fmt.Errorf("search_files: %v", err)
	}

	var out []string
	err = filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() || len(out) >= maxMatches {
			if d != nil && d.IsDir() && d.Name() == ".git" {
				return filepath.SkipDir
			}
			return nil
		}
		if a.FilePattern != "" {
			ok, _ := filepath.Match(a.FilePattern, d.Name())
			if !ok {
				return nil
			}
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return nil
		}
		for i, line := range strings.Split(string(data), "\n") {
			if re.MatchString(line) {
				out = append(out, fmt.Sprintf("%s:%d:%s", path, i+1, line))
				if len(out) >= maxMatches {
					break
				}
			}
		}
		return nil
	})
	if err != nil {
		return "", fmt.Errorf("search_files: %v", err)
	}
	if len(out) == 0 {
		return fmt.Sprintf("no matches for %q under %s", a.Pattern, root), nil
	}
	return strings.Join(out, "\n"), nil
}
