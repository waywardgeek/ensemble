package tools

import (
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

func (r *Registry) path(path string) string {
	if filepath.IsAbs(path) {
		return path
	}
	return filepath.Join(r.parent.Workspace(), path)
}
func (r *Registry) bounded(text string, limit int, label string) string {
	prefix := textPrefix(r, text, limit)
	if len(prefix) == len(text) {
		return text
	}
	return prefix + "\n[truncated: " + label + " limit reached]\n"
}
func (r *Registry) lines(data []byte) []string {
	if len(data) == 0 {
		return nil
	}
	rows := strings.SplitAfter(string(data), "\n")
	if rows[len(rows)-1] == "" {
		rows = rows[:len(rows)-1]
	}
	return rows
}
func readFile(r *Registry, a arguments) (string, error) {
	path := a["path"].(string)
	data, err := os.ReadFile(r.path(path))
	if err != nil {
		return "", r.failure("read_file path %q: %v", path, err)
	}
	start, end := a["start_line"].(int), a["end_line"].(int)
	if end > 0 && end < start {
		return "", r.failure("read_file end_line: must not precede start_line")
	}
	rows := r.lines(data)
	if len(rows) == 0 && start == 1 && end == 0 {
		// end_line:0 means through EOF even when supplied. An explicit start
		// still names a line target, which an empty file cannot satisfy.
		if a["explicit_start_line"] == true {
			return "", r.failure("read_file start_line/end_line: explicit range into empty path %q", path)
		}
		return "", nil
	}
	if start > len(rows) {
		return "", r.failure("read_file start_line: beyond EOF for path %q", path)
	}
	if end == 0 || end > len(rows) {
		end = len(rows)
	}
	return r.bounded(strings.Join(rows[start-1:end], ""), a["max_bytes"].(int), "max_bytes"), nil
}
func listDirectory(r *Registry, a arguments) (string, error) {
	path := a["path"].(string)
	entries, err := os.ReadDir(r.path(path))
	if err != nil {
		return "", r.failure("list_directory path %q: %v", path, err)
	}
	n := len(entries)
	limit := a["max_entries"].(int)
	if n > limit {
		n = limit
	}
	var out strings.Builder
	for _, e := range entries[:n] {
		out.WriteString(e.Name())
		if e.IsDir() {
			out.WriteByte('/')
		}
		out.WriteByte('\n')
	}
	text := r.bounded(out.String(), a["max_bytes"].(int), "max_bytes")
	if len(entries) > n {
		text += "\n[truncated: max_entries limit reached]\n"
	}
	return text, nil
}
func searchFiles(r *Registry, a arguments) (string, error) {
	expr, err := regexp.Compile(a["pattern"].(string))
	if err != nil {
		return "", r.failure("search_files pattern: invalid Go regular expression")
	}
	glob, _ := a["file_pattern"].(string)
	if _, err := filepath.Match(glob, ""); err != nil {
		return "", r.failure("search_files file_pattern: invalid glob")
	}
	path := a["path"].(string)
	root := r.path(path)
	limit := a["max_matches"].(int)
	context := a["context_lines"].(int)
	matches := 0
	omitted := false
	var out strings.Builder
	err = filepath.WalkDir(root, func(full string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return r.failure("search_files path %q: %v", full, walkErr)
		}
		if d.IsDir() {
			if d.Name() == ".git" {
				return filepath.SkipDir
			}
			return nil
		}
		if !d.Type().IsRegular() {
			return nil
		}
		if glob != "" {
			ok, _ := filepath.Match(glob, d.Name())
			if !ok {
				return nil
			}
		}
		data, err := os.ReadFile(full)
		if err != nil {
			return r.failure("search_files path %q: %v", full, err)
		}
		first := len(data)
		if first > 8192 {
			first = 8192
		}
		if bytes.IndexByte(data[:first], 0) >= 0 {
			return nil
		}
		lines := r.lines(data)
		selected := map[int]bool{}
		indices := []int{}
		for i, line := range lines {
			if expr.MatchString(strings.TrimSuffix(line, "\n")) {
				if matches < limit {
					selected[i] = true
					indices = append(indices, i)
					matches++
				} else {
					omitted = true
				}
			}
		}
		if len(indices) == 0 {
			return nil
		}
		display := full
		if !filepath.IsAbs(path) {
			rel, e := filepath.Rel(root, full)
			if e != nil {
				return r.failure("search_files path: cannot form relative result")
			}
			if rel == "." {
				display = path
			} else {
				display = filepath.Join(path, rel)
			}
		}
		last := -2
		for _, index := range indices {
			start := 0
			if context < index {
				start = index - context
			}
			end := len(lines) - 1
			if context < len(lines)-1-index {
				end = index + context
			}
			if start > last+1 && last >= 0 {
				out.WriteString("--\n")
			}
			if start <= last {
				start = last + 1
			}
			for i := start; i <= end; i++ {
				marker := "-"
				if selected[i] {
					marker = ":"
				}
				fmt.Fprintf(&out, "%s%s%d%s%s\n", display, marker, i+1, marker, strings.TrimSuffix(lines[i], "\n"))
			}
			if end > last {
				last = end
			}
		}
		return nil
	})
	if err != nil {
		return "", err
	}
	text := out.String()
	if matches == 0 {
		text = "No matches.\n"
	}
	text = r.bounded(text, a["max_bytes"].(int), "max_bytes")
	if omitted {
		text += "\n[truncated: max_matches limit reached]\n"
	}
	return text, nil
}
func writeFile(r *Registry, a arguments) (string, error) {
	path := a["path"].(string)
	full := r.path(path)
	content := a["content"].(string)
	if err := os.MkdirAll(filepath.Dir(full), 0755); err != nil {
		return "", r.failure("write_file path %q: %v", path, err)
	}
	flags := os.O_WRONLY | os.O_CREATE
	operation := "created"
	if a["append"].(bool) {
		flags |= os.O_APPEND
		operation = "appended"
	} else if a["overwrite"].(bool) {
		flags |= os.O_TRUNC
		operation = "wrote"
	} else {
		flags |= os.O_EXCL
	}
	f, err := os.OpenFile(full, flags, 0644)
	if err != nil {
		if os.IsExist(err) {
			if info, e := os.Stat(full); e == nil {
				return "", r.failure("write_file path %q: refusing replacement of existing file (%d bytes); set overwrite:true", path, info.Size())
			}
		}
		return "", r.failure("write_file path %q: %v", path, err)
	}
	n, writeErr := f.WriteString(content)
	closeErr := f.Close()
	if writeErr != nil {
		return "", r.failure("write_file path %q after %d bytes: %v", path, n, writeErr)
	}
	if closeErr != nil {
		return "", r.failure("write_file path %q close: %v", path, closeErr)
	}
	return fmt.Sprintf("%s %q: %d bytes", operation, path, n), nil
}
func editFile(r *Registry, a arguments) (string, error) {
	path := a["path"].(string)
	full := r.path(path)
	data, err := os.ReadFile(full)
	if err != nil {
		return "", r.failure("edit_file path %q: %v", path, err)
	}
	old, next := a["old_text"].(string), a["new_text"].(string)
	count := strings.Count(string(data), old)
	if count != 1 {
		return "", r.failure("edit_file old_text in %q: found %d matches; read the file and choose a unique anchor", path, count)
	}
	if old == next {
		return fmt.Sprintf("edit_file %q: successful no-op (identical replacement)", path), nil
	}
	info, err := os.Stat(full)
	if err != nil {
		return "", r.failure("edit_file path %q: %v", path, err)
	}
	if err = os.WriteFile(full, []byte(strings.Replace(string(data), old, next, 1)), info.Mode().Perm()); err != nil {
		return "", r.failure("edit_file path %q: %v", path, err)
	}
	return fmt.Sprintf("edited %q: replaced 1 exact match", path), nil
}
