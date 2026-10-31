package gradeedition2

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/waywardgeek/ensemble/internal/fakevendor"
	"github.com/waywardgeek/ensemble/internal/grade"
)

// Chapter3 preserves the ten original checks, carries corrected Chapter 2
// parity forward, and exercises the chapter's omitted search/append contracts.
func Chapter3(bin string) (grade.Report, error) {
	res, err := grade.Ch3Run(bin)
	if err != nil {
		return grade.Report{}, err
	}
	// The original runner retains only its old Chapter 2 verdict. Reuse the
	// corrected grader rather than duplicating its scope and capture fixes here.
	parity, err := Chapter2(bin)
	if err != nil {
		return grade.Report{}, err
	}
	res.Ch2, res.Ch2Err = parity.Checks, ""
	searchErr, appendErr := fileToolExtensions(bin)
	checks := grade.Ch3Evaluate(res)
	for i := range checks {
		check := &checks[i]
		switch check.ID {
		case "readtools":
			if searchErr != nil {
				if !check.Passed {
					for j := range check.Details {
						check.Details[j] = "original probes: " + check.Details[j]
					}
				}
				// Charge this tool once, preserving deductions for other read tools.
				// A group-score cap would miss a simultaneous read/search defect.
				earned := check.Earned
				if originalSearchPassed(res.LocalTools) {
					earned = max(0, earned-4)
				}
				fail(check, "search_files context: %v", searchErr)
				check.Earned = earned
			} else {
				check.Details = append(check.Details, "context_lines retains, merges and separates grep-style windows")
			}
		case "mutatetools":
			if appendErr != nil {
				fail(check, "write_file append: %v", appendErr)
			} else {
				check.Details = append(check.Details, "append:true preserves existing bytes without requiring overwrite:true")
			}
		}
	}
	return grade.NewTitledReport("Edition 2, Chapter 3 — Six Tools", checks, ""), nil
}

// These operations were already required by section 3.3. Exercise
// the real loop/filesystem with the existing provider fake, not tool doubles.
// The planted search lines include adjacent windows and a separated group;
// append operates on a different file so one defect cannot conceal the other.
func fileToolExtensions(bin string) (searchErr, appendErr error) {
	work, err := os.MkdirTemp("", "edition2-ch03-*")
	if err != nil {
		return err, err
	}
	defer os.RemoveAll(work)
	const notes = "one\nneedle\nthree\nfour\nneedle\nsix\nseven\neight\nneedle\nten\n"
	for name, data := range map[string]string{"notes.txt": notes, "append.txt": "original\n"} {
		if err := os.WriteFile(filepath.Join(work, name), []byte(data), 0600); err != nil {
			return err, err
		}
	}
	fake := fakevendor.New([]fakevendor.Reply{
		{Tools: []fakevendor.ToolCall{
			{ID: "search-context", Name: "search_files", Args: `{"path":"notes.txt","pattern":"needle","context_lines":1}`},
			{ID: "append-content", Name: "write_file", Args: `{"path":"append.txt","content":"tail\n","append":true}`},
		}},
		{Text: "File probes complete."},
	})
	defer fake.Close()
	if _, err := runChapter2(bin, work, "anthropic", fake.URL(), "{\"user\":\"exercise the file tools\"}\n"); err != nil {
		return err, err
	}
	requests := fake.Requests()
	if len(requests) < 2 {
		err := fmt.Errorf("no subsequent request delivered the tool results")
		return err, err
	}
	// Inspect the last actual request: a tool log alone cannot establish that
	// the model received a readable result or a failure flag.
	body := requests[len(requests)-1].Body
	search, searchFailed, searchErr := anthropicToolText(body, "search-context")
	if searchFailed {
		searchErr = fmt.Errorf("call search-context returned a tool error")
	}
	const expected = "notes.txt-1-one\nnotes.txt:2:needle\nnotes.txt-3-three\nnotes.txt-4-four\nnotes.txt:5:needle\nnotes.txt-6-six\n--\nnotes.txt-8-eight\nnotes.txt:9:needle\nnotes.txt-10-ten\n"
	if searchErr == nil && search != expected {
		searchErr = fmt.Errorf("expected grep -C1 matching/context delimiters and merged groups; got %q", search)
	}
	_, appendFailed, appendErr := anthropicToolText(body, "append-content")
	if appendFailed {
		appendErr = fmt.Errorf("call append-content returned a tool error")
	}
	data, err := os.ReadFile(filepath.Join(work, "append.txt"))
	if appendErr == nil && (err != nil || string(data) != "original\ntail\n") {
		appendErr = fmt.Errorf("existing bytes plus appended bytes were not preserved: read error %v, content %q", err, data)
	}
	return searchErr, appendErr
}

// Anthropic allows string content or an array of text blocks in tool results.
// Accept both spellings, checking only the tool's output and error flag.
func anthropicToolText(body []byte, callID string) (string, bool, error) {
	var request map[string]any
	if err := json.Unmarshal(body, &request); err != nil {
		return "", false, err
	}
	for _, message := range list(request["messages"]) {
		for _, raw := range list(object(message)["content"]) {
			part := object(raw)
			if part["type"] != "tool_result" || part["tool_use_id"] != callID {
				continue
			}
			if text, ok := part["content"].(string); ok {
				return text, part["is_error"] == true, nil
			}
			var text strings.Builder
			for _, block := range list(part["content"]) {
				if value, ok := object(block)["text"].(string); ok {
					text.WriteString(value)
				}
			}
			return text.String(), part["is_error"] == true, nil
		}
	}
	return "", false, fmt.Errorf("call %s has no result in the model request", callID)
}

// Mirror only the original search verdict's two text predicates to account for
// its four-point deduction. Original checks still judge their complete scenario;
// their human-readable Details are not a machine-readable per-tool breakdown.
func originalSearchPassed(session *grade.Ch3Session) bool {
	if session == nil || session.Err != "" {
		return false
	}
	for _, request := range session.Requests {
		text, _, err := anthropicToolText(request.Body, "toolu_lt_grep")
		if err == nil {
			return strings.Contains(text, "notes.md") && strings.Contains(text, "gamma line three")
		}
	}
	return false
}
