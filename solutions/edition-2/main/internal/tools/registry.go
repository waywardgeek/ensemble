// Package tools owns the Agent's visible capabilities and synchronous execution.
package tools

import (
	"encoding/json"
	"example.com/ensemble/internal/common"
	"fmt"
	"reflect"
	"sort"
)

type field struct {
	Type        string `json:"type"`
	Description string `json:"description"`
	Default     any    `json:"default,omitempty"`
	Required    bool   `json:"-"`
	AllowEmpty  bool   `json:"-"`
	Minimum     *int   `json:"minimum,omitempty"`
}
type entry struct {
	description string
	fields      map[string]field
	run         func(*Registry, arguments) (string, error)
}
type arguments map[string]any
type Registry struct {
	parent  common.ToolAgent
	entries map[string]entry
}

func (r *Registry) Agent() common.Agent { return r.parent }
func textField(description string, required, empty bool, fallback string) field {
	var def any
	if !required && fallback != "" {
		def = fallback
	}
	return field{Type: "string", Description: description, Required: required, AllowEmpty: empty, Default: def}
}
func numberField(description string, fallback, minimum int) field {
	return field{Type: "integer", Description: description, Default: fallback, Minimum: &minimum}
}
func boolField(description string) field {
	return field{Type: "boolean", Description: description, Default: false}
}
func New(parent common.ToolAgent, selected []string) (*Registry, error) {
	r := &Registry{parent: parent, entries: map[string]entry{}}
	path := textField("File path, relative to the Agent workspace or absolute; not confined to workspace.", true, false, "")
	root := textField("File or directory path; defaults to the Agent workspace (.).", false, false, ".")
	cap := numberField("Maximum retained bytes; defaults to 65536. Omitted content is marked truncated.", 65536, 1)
	builtins := map[string]entry{
		"read_file":      {"Read an inclusive one-based line range; report I/O and range errors.", map[string]field{"path": path, "start_line": numberField("First line, default 1.", 1, 1), "end_line": numberField("Last line; default 0 reads through EOF.", 0, 0), "max_bytes": cap}, readFile},
		"list_directory": {"List one directory level in name order, distinguishing directories.", map[string]field{"path": root, "max_entries": numberField("Maximum entries; default 200. Further entries cause a truncation notice.", 200, 1), "max_bytes": cap}, listDirectory},
		"search_files":   {"Search regular text files recursively using Go regexp; skip .git, symlinks, and files with NUL in the first 8192 bytes.", map[string]field{"path": root, "pattern": textField("Nonempty Go regular expression.", true, false, ""), "file_pattern": textField("Optional filepath.Match glob matched against basenames; omitted means no filter.", false, false, ""), "context_lines": numberField("Neighbor lines around selected matching lines; default 0. Overlapping or adjacent groups merge.", 0, 0), "max_matches": numberField("Maximum matching lines, default 200. Further matches cause a truncation notice.", 200, 1), "max_bytes": cap}, searchFiles},
		"write_file":     {"Create a file and missing parents. Refuse replacing an existing file without overwrite:true; append mode creates or appends.", map[string]field{"path": path, "content": textField("Exact text to write; empty text is valid.", true, true, ""), "append": boolField("Append or create; default false. When true, overwrite is irrelevant."), "overwrite": boolField("Allow replacement of existing bytes; default false.")}, writeFile},
		"edit_file":      {"Replace exactly one non-overlapping exact anchor. Refuse zero or multiple matches; read the file and choose a unique anchor.", map[string]field{"path": path, "old_text": textField("Nonempty exact unique anchor.", true, false, ""), "new_text": textField("Replacement; empty text deletes the anchor. Identical text succeeds as a no-op.", true, true, "")}, editFile},
		"run_command":    {"Start a fresh POSIX shell on a PTY with merged stdout/stderr and input echo. Retain all output and exit_code on disk; wait delay never kills the job.", map[string]field{"command": textField("Nonempty shell command.", true, false, ""), "cwd": textField("Optional working-directory override, absolute or relative to Agent workspace.", false, false, "")}, nil},
		"wait_for_job":   {"Report unseen job output; wait for completion, new-output pattern, or delay. Finished jobs return immediately.", map[string]field{"handle": {Type: "integer", Required: true, Description: "Positive job handle owned by this Agent."}}, nil},
		"send_input":     {"Send input to a running PTY job, then wait using only output since this input for pattern matching.", map[string]field{"handle": {Type: "integer", Required: true}, "input": textField("Input bytes; empty is valid.", true, true, ""), "append_newline": {Type: "boolean", Default: true, Description: "Append one LF, default true; false sends exact bytes."}}, nil},
		"kill_job":       {"Kill a running job's process group. Repeated kill preserves the existing terminal status.", map[string]field{"handle": {Type: "integer", Required: true}}, nil},
		"tool_limits":    {"Set supplied wait/report overrides for the next attempted tool call only, including invalid calls and setters.", map[string]field{}, nil},
	}
	for _, name := range []string{"run_command", "wait_for_job", "send_input", "tool_limits"} {
		fields := builtins[name].fields
		fields["ai_callback_delay"] = field{Type: "number", Description: "Finite nonnegative wait seconds; default 3. Zero polls."}
		fields["ai_callback_pattern"] = textField("Go regular expression matching new output; empty clears the pattern.", false, true, "")
		fields["max_output_bytes"] = numberField("Positive inline content budget; default 16384. Complete output stays on disk.", 16384, 1)
	}
	for _, name := range selected {
		item, ok := builtins[name]
		if !ok {
			return nil, r.failure("unknown builtin %q", name)
		}
		if _, ok = r.entries[name]; ok {
			return nil, r.failure("duplicate builtin %q", name)
		}
		r.entries[name] = item
	}
	return r, nil
}
func (r *Registry) Declarations() []common.ToolDefinition {
	names := make([]string, 0, len(r.entries))
	for n := range r.entries {
		names = append(names, n)
	}
	sort.Strings(names)
	out := make([]common.ToolDefinition, 0, len(names))
	for _, name := range names {
		e := r.entries[name]
		required := []string{}
		for n, f := range e.fields {
			if f.Required {
				required = append(required, n)
			}
		}
		sort.Strings(required)
		schema, _ := json.Marshal(map[string]any{"type": "object", "properties": e.fields, "required": required, "additionalProperties": false})
		out = append(out, common.ToolDefinition{Name: name, Description: e.description, Schema: schema})
	}
	return out
}

// Match compares JSON values so insignificant whitespace cannot create two authorities.
func (r *Registry) Match(defs []common.ToolDefinition) bool {
	if len(defs) == 0 {
		return true
	}
	expected := r.Declarations()
	if len(defs) != len(expected) {
		return false
	}
	seen := map[string]bool{}
	for _, got := range defs {
		var found *common.ToolDefinition
		for i := range expected {
			if expected[i].Name == got.Name {
				found = &expected[i]
			}
		}
		if found == nil || seen[got.Name] || got.Description != found.Description {
			return false
		}
		seen[got.Name] = true
		var a, b any
		if json.Unmarshal(got.Schema, &a) != nil || json.Unmarshal(found.Schema, &b) != nil || !reflect.DeepEqual(a, b) {
			return false
		}
	}
	return true
}
func (r *Registry) Execute(call common.Part) common.ToolEvent {
	result := r.execute(call)
	text := result.Text + result.Note
	return common.ToolEvent{CallID: call.CallID, IsError: result.IsError, Parts: []common.Part{{Type: "text", Text: &text}}}
}
func (r *Registry) execute(call common.Part) common.ExecutionResult {
	result := common.ExecutionResult{}
	e, ok := r.entries[call.Name]
	var text, note string
	var err error
	if !ok {
		err = r.failure("%s: tool is unavailable to this Agent", call.Name)
	} else {
		var args arguments
		args, err = r.decode(call.Name, e, call.Args)
		if err == nil {
			if e.run == nil {
				err = r.failure("%s requires managed dispatch", call.Name)
			} else if call.Name == "read_file" {
				text, note, err = readSelection(r, args)
			} else {
				text, err = e.run(r, args)
			}
		}
	}
	if err != nil {
		result.IsError = true
		text = fmt.Sprintf("%s failed: %s", call.Name, err)
	}
	result.Text = text
	result.Note = note
	return result
}
func (r *Registry) decode(name string, e entry, raw json.RawMessage) (arguments, error) {
	var data map[string]json.RawMessage
	if json.Unmarshal(raw, &data) != nil || data == nil {
		return nil, r.failure("%s: arguments must be an object", name)
	}
	out := arguments{}
	if name == "read_file" {
		for _, key := range []string{"start_line"} {
			if _, ok := data[key]; ok {
				out["explicit_"+key] = true
			}
		}
	}
	for key := range data {
		if _, ok := e.fields[key]; !ok {
			return nil, r.failure("%s: unknown field %s", name, key)
		}
	}
	for key, f := range e.fields {
		raw, ok := data[key]
		if !ok {
			if f.Required {
				return nil, r.failure("%s: missing field %s (%s)", name, key, f.Description)
			}
			if f.Default != nil {
				out[key] = f.Default
			}
			continue
		}
		bad := func() (arguments, error) {
			return nil, r.failure("%s: invalid field %s (%s)", name, key, f.Description)
		}
		if string(raw) == "null" {
			return bad()
		}
		switch f.Type {
		case "string":
			var v string
			if json.Unmarshal(raw, &v) != nil || (!f.AllowEmpty && v == "") {
				return bad()
			}
			out[key] = v
		case "number":
			var v float64
			if json.Unmarshal(raw, &v) != nil {
				return bad()
			}
			out[key] = v
		case "integer":
			var v int
			if json.Unmarshal(raw, &v) != nil || (f.Minimum != nil && v < *f.Minimum) {
				return bad()
			}
			out[key] = v
		case "boolean":
			var v bool
			if json.Unmarshal(raw, &v) != nil {
				return bad()
			}
			out[key] = v
		}
	}
	return out, nil
}
func (r *Registry) failure(format string, args ...any) error {
	err := fmt.Errorf(format, args...)
	r.parent.Ensemble().Logf("%s", err)
	return err
}

func (r *Registry) Kind(name string) (bool, bool) {
	_, available := r.entries[name]
	supervision := name == "wait_for_job" || name == "send_input" || name == "kill_job" || name == "tool_limits"
	return available, supervision
}

// ExecuteJob leaves a successfully started process to its lifecycle worker.
// Local results return to Jobs, which owns their artifact and terminal event.
func (r *Registry) ExecuteJob(call common.Part, job common.Job) *common.ExecutionResult {
	if call.Name != "run_command" {
		result := r.execute(call)
		return &result
	}
	result := common.ExecutionResult{}
	entry, ok := r.entries[call.Name]
	var err error
	if !ok {
		err = r.failure("run_command unavailable")
	} else {
		var args arguments
		args, err = r.decode(call.Name, entry, call.Args)
		if err == nil {
			err = startCommand(r, job, args)
		}
	}
	if err == nil {
		return nil
	}
	text := fmt.Sprintf("run_command failed: %v", err)
	result.IsError = true
	result.Text = text
	return &result
}
