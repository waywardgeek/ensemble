// Package tools owns local operations and the four job-control verbs. It knows the owner
// through common interfaces and does not import Agent, Engine or History.
package tools

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"ensemble/internal/common"
)

// Each Agent receives its own registry slice. Schemas are declarations, not a
// second argument validator; typed decoding below still checks untrusted calls.
func Builtins() []common.ToolDefinition {
	definitions := []common.ToolDefinition{
		{
			ToolDeclaration: common.ToolDeclaration{
				Name:        "read_file",
				Description: "Read text with one-based inclusive start_line/end_line. Optional max_bytes bounds the read; job reports cap inline output and preserve the full result on disk.",
				Schema:      json.RawMessage(`{"type":"object","properties":{"path":{"type":"string"},"start_line":{"type":"integer"},"end_line":{"type":"integer"},"max_bytes":{"type":"integer"}},"required":["path"]}`),
			},
			Run: readFile,
		},
		{
			ToolDeclaration: common.ToolDeclaration{
				Name:        "list_directory",
				Description: "List immediate entries, sorted by name; directories end with /.",
				Schema:      json.RawMessage(`{"type":"object","properties":{"path":{"type":"string"}},"required":["path"]}`),
			},
			Run: listDirectory,
		},
		{
			ToolDeclaration: common.ToolDeclaration{
				Name:        "search_files",
				Description: "Search files recursively using a regular expression. context_lines defaults to zero. Output uses grep path:line:text and path-line-context with merged windows.",
				Schema:      json.RawMessage(`{"type":"object","properties":{"path":{"type":"string"},"pattern":{"type":"string"},"context_lines":{"type":"integer"}},"required":["path","pattern"]}`),
			},
			Run: searchFiles,
		},
		{
			ToolDeclaration: common.ToolDeclaration{
				Name:        "write_file",
				Description: "Write text, creating parent directories. Existing files require overwrite:true unless append:true. Append preserves existing content.",
				Schema:      json.RawMessage(`{"type":"object","properties":{"path":{"type":"string"},"content":{"type":"string"},"overwrite":{"type":"boolean"},"append":{"type":"boolean"}},"required":["path","content"]}`),
			},
			Run: writeFile,
		},
		{
			ToolDeclaration: common.ToolDeclaration{
				Name:        "edit_file",
				Description: "Replace one exact, unique old_text anchor with new_text. Zero or multiple matches refuse without changing the file; read and provide a more specific anchor.",
				Schema:      json.RawMessage(`{"type":"object","properties":{"path":{"type":"string"},"old_text":{"type":"string"},"new_text":{"type":"string"}},"required":["path","old_text","new_text"]}`),
			},
			Run: editFile,
		},
		{
			ToolDeclaration: common.ToolDeclaration{
				Name:        "run_command",
				Description: "Start a fresh /bin/sh in a PTY. cwd applies only to this call, relative to the workspace. Merged output and exit_code stay on disk; a running report is a handle, not cancellation. ai_callback_delay is seconds (default 3); pattern matches unseen output; max_output_bytes defaults to 16384.",
				Schema:      json.RawMessage(waitSchema(`"command":{"type":"string"},"cwd":{"type":"string"}`, `"command"`)),
			},
			Run: runCommand,
		},
	}
	return append(definitions, supervision()...)
}

// Presence is different from an empty value: writing an empty file is valid,
// omitting its content is not. JSON null does not satisfy a required argument.
func decode(raw json.RawMessage, into any, required ...string) error {
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(raw, &fields); err != nil {
		return fmt.Errorf("arguments must be an object: %w", err)
	}
	for _, name := range required {
		if len(fields[name]) == 0 || string(fields[name]) == "null" {
			return fmt.Errorf("missing required argument %s", name)
		}
	}
	if err := json.Unmarshal(raw, into); err != nil {
		return fmt.Errorf("invalid arguments: %w", err)
	}
	return nil
}

// A range bounds the text the model pays to resend on every later request.
// The byte cap is applied to that range, and a marker makes omission visible.
// Reading the whole local file here keeps Chapter 3 simple; this is an output
// cap, not a claim of bounded filesystem memory or background I/O.
func readFile(owner common.ToolContext, raw json.RawMessage) (string, error) {
	var args struct {
		// Path identifies the requested file or directory before tool validation.
		Path string `json:"path"`
		// Start selects the first requested line using the tool's one-based convention.
		Start int `json:"start_line"`
		// End selects the inclusive last requested line when supplied.
		End int `json:"end_line"`
		// Max bounds the visible output without changing retained file contents.
		Max int `json:"max_bytes"`
	}
	if err := decode(raw, &args, "path"); err != nil {
		return "", err
	}
	if args.Start == 0 {
		args.Start = 1
	}
	if args.Start < 1 || args.End < 0 || (args.End > 0 && args.End < args.Start) || args.Max < 0 {
		return "", errors.New("invalid line range or max_bytes")
	}
	data, err := os.ReadFile(args.Path)
	if err != nil {
		return "", err
	}
	// SplitAfter retains real newline bytes, including an unterminated last line.
	// Slice the requested lines before applying the cap, so later ranges work.
	lines := strings.SplitAfter(string(data), "\n")
	if lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	end := len(lines)
	if args.End > 0 && args.End < end {
		end = args.End
	}
	if args.Start > end {
		return "", nil
	}
	result := strings.Join(lines[args.Start-1:end], "")
	if args.Max > 0 && len(result) > args.Max {
		return result[:args.Max] + "\n[truncated; use a narrower line range or larger max_bytes]", nil
	}
	return result, nil
}

// Orientation needs names, not a recursive dump. ReadDir sorts entries for
// reproducible output; the slash distinguishes a directory without another stat.
func listDirectory(owner common.ToolContext, raw json.RawMessage) (string, error) {
	var args struct {
		// Path identifies the requested file or directory before tool validation.
		Path string `json:"path"`
	}
	if err := decode(raw, &args, "path"); err != nil {
		return "", err
	}
	entries, err := os.ReadDir(args.Path)
	if err != nil {
		return "", err
	}
	var out strings.Builder
	for _, entry := range entries {
		out.WriteString(entry.Name())
		if entry.IsDir() {
			out.WriteByte('/')
		}
		out.WriteByte('\n')
	}
	return out.String(), nil
}

// New files and appends are routine. Replacing existing bytes requires an
// explicit overwrite flag so an accidental whole-file rewrite is recoverable.
// Creating parents lets the model write a new package in one ordinary call.
func writeFile(owner common.ToolContext, raw json.RawMessage) (string, error) {
	var args struct {
		// Path identifies the requested file or directory before tool validation.
		Path string `json:"path"`
		// Content is the exact text to write after overwrite/append validation.
		Content string `json:"content"`
		// Overwrite explicitly permits replacing an existing file.
		Overwrite bool `json:"overwrite"`
		// Append requests extension of an existing file instead of replacement.
		Append bool `json:"append"`
	}
	if err := decode(raw, &args, "path", "content"); err != nil {
		return "", err
	}
	if err := os.MkdirAll(filepath.Dir(args.Path), 0755); err != nil {
		return "", err
	}
	flags := os.O_WRONLY | os.O_CREATE | os.O_EXCL
	if args.Append {
		flags = os.O_WRONLY | os.O_CREATE | os.O_APPEND
	} else if args.Overwrite {
		flags = os.O_WRONLY | os.O_CREATE | os.O_TRUNC
	}
	// O_EXCL enforces the guard at open, not only at an earlier existence check.
	// The diagnostic tells the model what it must explicitly acknowledge next.
	file, err := os.OpenFile(args.Path, flags, 0644)
	if errors.Is(err, fs.ErrExist) {
		info, statErr := os.Stat(args.Path)
		if statErr != nil {
			return "", statErr
		}
		return "", fmt.Errorf("%s exists (%d bytes); read it first, then use overwrite:true to replace it", args.Path, info.Size())
	}
	if err != nil {
		return "", err
	}
	n, writeErr := file.WriteString(args.Content)
	closeErr := file.Close()
	if writeErr != nil {
		return "", writeErr
	}
	if closeErr != nil {
		return "", closeErr
	}
	return fmt.Sprintf("wrote %d bytes to %s", n, args.Path), nil
}

// Text anchors survive unrelated earlier edits, unlike line-number patches.
// The same unique-anchor rule is advertised in the declaration and enforced
// here, so the model can repair a failed call using the reported match count.
func editFile(owner common.ToolContext, raw json.RawMessage) (string, error) {
	var args struct {
		// Path identifies the requested file or directory before tool validation.
		Path string `json:"path"`
		// Old identifies the exact text that the edit must match before replacement.
		Old string `json:"old_text"`
		// New supplies replacement text only after the old text is uniquely located.
		New string `json:"new_text"`
	}
	if err := decode(raw, &args, "path", "old_text", "new_text"); err != nil {
		return "", err
	}
	if args.Old == "" {
		return "", errors.New("old_text must be a nonempty exact anchor")
	}
	data, err := os.ReadFile(args.Path)
	if err != nil {
		return "", err
	}
	// Exact uniqueness is the chosen edit contract in both mismatch directions.
	// Never fall back to rewriting or guessing; the failed result enables repair.
	// Count overlapping sites too: "aa" in "aaa" names two possible edits.
	count := 0
	for rest := string(data); ; {
		index := strings.Index(rest, args.Old)
		if index < 0 {
			break
		}
		count++
		rest = rest[index+1:]
	}
	if count != 1 {
		return "", fmt.Errorf("exact edit refused: old_text matched %d times in %s; read the file and supply one unique anchor; file unchanged", count, args.Path)
	}
	updated := strings.Replace(string(data), args.Old, args.New, 1)
	if err := os.WriteFile(args.Path, []byte(updated), 0644); err != nil {
		return "", err
	}
	return "replaced one exact match in " + args.Path, nil
}

// Shell setup belongs to the tool; process/output lifetime belongs to its job.
// No persistent shell state or execution deadline is hidden behind a call.
func runCommand(owner common.ToolContext, raw json.RawMessage) (string, error) {
	var args struct {
		// Command is the explicit shell program, with no persistent shell state.
		Command string `json:"command"`
		// Cwd optionally overrides this call's workspace without changing Agent
		// configuration.
		Cwd string `json:"cwd"`
	}
	if err := decode(raw, &args, "command"); err != nil {
		return "", err
	}
	if strings.TrimSpace(args.Command) == "" {
		return "", errors.New("command must not be empty")
	}
	command := exec.Command("/bin/sh", "-c", args.Command)
	command.Dir = owner.Engine().Agent().Config().Workspace
	if args.Cwd != "" {
		if filepath.IsAbs(args.Cwd) {
			command.Dir = args.Cwd
		} else {
			command.Dir = filepath.Join(command.Dir, args.Cwd)
		}
	}
	// Provider credentials belong to transport, never an unrelated shell child.
	// TERM is set once after filtering so an inherited terminal cannot override it.
	for _, variable := range os.Environ() {
		name, _, _ := strings.Cut(variable, "=")
		switch name {
		case "LLM_API_KEY", "ANTHROPIC_API_KEY", "OPENAI_API_KEY", "GEMINI_API_KEY", "TERM":
		default:
			command.Env = append(command.Env, variable)
		}
	}
	command.Env = append(command.Env, "TERM=dumb")
	return "", owner.Job().StartProcess(command)
}

// Search uses local files and Go regular expressions, without shell quoting
// or an external grep dependency. The familiar grep-style output identifies
// matches versus context and preserves separated groups with a visible marker.
func searchFiles(owner common.ToolContext, raw json.RawMessage) (string, error) {
	var args struct {
		// Path identifies the requested file or directory before tool validation.
		Path string `json:"path"`
		// Pattern is a regular expression matched against each file line.
		Pattern string `json:"pattern"`
		// Context selects the number of neighboring lines included around grep matches.
		Context int `json:"context_lines"`
	}
	if err := decode(raw, &args, "path", "pattern"); err != nil {
		return "", err
	}
	if args.Context < 0 {
		return "", errors.New("context_lines must not be negative")
	}
	pattern, err := regexp.Compile(args.Pattern)
	if err != nil {
		return "", err
	}
	var out strings.Builder
	printed := false
	// WalkDir gives stable path order and does not follow symlink directories.
	// Match on lines, then merge adjacent context windows exactly as grep -C does.
	err = filepath.WalkDir(args.Path, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if !entry.Type().IsRegular() {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		// An empty file has no lines, even for a regexp matching empty text.
		if len(data) == 0 {
			return nil
		}
		lines := strings.Split(strings.TrimSuffix(string(data), "\n"), "\n")
		matches := make([]bool, len(lines))
		var starts, ends []int
		for i, line := range lines {
			if !pattern.MatchString(line) {
				continue
			}
			matches[i] = true
			start, end := max(0, i-args.Context), min(len(lines), i+args.Context+1)
			if len(ends) > 0 && start <= ends[len(ends)-1] {
				ends[len(ends)-1] = end
			} else {
				starts = append(starts, start)
				ends = append(ends, end)
			}
		}
		for group, start := range starts {
			if printed && args.Context > 0 {
				out.WriteString("--\n")
			}
			for i := start; i < ends[group]; i++ {
				separator := "-"
				if matches[i] {
					separator = ":"
				}
				fmt.Fprintf(&out, "%s%s%d%s%s\n", path, separator, i+1, separator, lines[i])
			}
			printed = true
		}
		return nil
	})
	return out.String(), err
}
