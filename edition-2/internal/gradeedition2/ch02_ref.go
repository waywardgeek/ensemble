package gradeedition2

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/waywardgeek/ensemble/internal/grade"
)

// refRedaction exercises the published CLI without requiring media rendering.
// The original text-only redaction check remains its positive control. Here a
// redacted blob must render as a result stub, while the original facts must
// remain in the log. Neither observation exposes the reducer's private Ref.
func refRedaction(bin, work string) grade.Check {
	check := grade.Check{ID: "ref-redaction", Title: "redacted view omits bytes; original Ref facts remain in the log", Points: 8, Earned: 8, Passed: true}
	for kind := 1; kind <= 3; kind++ {
		fixture := strings.Replace(grade.RefRedactionLog, `"kind":2`, fmt.Sprintf(`"kind":%d`, kind), 1)
		body, err := renderFixture(bin, work, "gemini", fixture)
		if err != nil {
			fail(&check, "Ref kind %d redacted render: %v", kind, err)
			continue
		}
		var request struct {
			// Contents retains all turns, including the redacted tool result.
			Contents []struct {
				// Parts separates function results from ordinary text.
				Parts []struct {
					// FunctionResponse must carry a readable replacement result.
					FunctionResponse *struct {
						// Name identifies the fixture call whose result was redacted.
						Name string
						// Response permits equivalent wrappers while requiring nonempty text.
						Response map[string]any
					}
				}
			}
		}
		if err := json.Unmarshal([]byte(body), &request); err != nil {
			fail(&check, "Ref kind %d render is invalid JSON: %v", kind, err)
			continue
		}
		resultSeen := false
		for _, turn := range request.Contents {
			for _, part := range turn.Parts {
				if result := part.FunctionResponse; result != nil && result.Name == "run_build" && hasText(result.Response) {
					resultSeen = true
				}
			}
		}
		if !resultSeen || strings.Contains(body, grade.RefRedactionSecret) {
			fail(&check, "Ref kind %d must retain a result stub without the removed text", kind)
		}
		dump, err := runChapter2(bin, work, "gemini", "http://127.0.0.1:1", "", "dump")
		if err != nil {
			fail(&check, "Ref kind %d dump after render: %v", kind, err)
			continue
		}
		events, err := readLog(dump)
		if err != nil {
			fail(&check, "Ref kind %d dumped log: %v", kind, err)
			continue
		}
		refSeen, textSeen, redactionSeen := false, false, false
		for _, event := range events {
			if event.Type == "redacted" {
				redactionSeen = true
			}
			if event.Type != "toolreturned" {
				continue
			}
			tool := object(event.Data["tool"])
			for _, value := range list(tool["parts"]) {
				part := object(value)
				if part["text"] == grade.RefRedactionSecret {
					textSeen = true
				}
				ref := object(part["ref"])
				if ref["locator"] != grade.RefRedactionLocator {
					continue
				}
				// As in the original grader, numeric and readable RefKind
				// encodings describe the same fact; the locator is verbatim.
				got := ref["kind"]
				name, _ := got.(string)
				refSeen = got == float64(kind) || normalize(name) == []string{"", "path", "uri", "handle"}[kind]
			}
		}
		if !refSeen || !textSeen || !redactionSeen {
			fail(&check, "Ref kind %d dump lost original Kind/Locator, original text, or redaction event", kind)
		}
	}
	if check.Passed {
		check.Details = []string{
			"all three RefKinds retain original log facts and render a secret-free result stub",
			"exact Ref Kind/Locator inside reduced stubs is not exposed by the CLI; verify it in source tests and review",
		}
	}
	return check
}

// A nonempty response object can contain only an empty output string. Walk
// its payload so either supported vendor wrapper spelling still needs a stub
// that a model can read; flags and keys alone are not replacement content.
func hasText(value any) bool {
	switch value := value.(type) {
	case string:
		return strings.TrimSpace(value) != ""
	case map[string]any:
		for _, child := range value {
			if hasText(child) {
				return true
			}
		}
	case []any:
		for _, child := range value {
			if hasText(child) {
				return true
			}
		}
	}
	return false
}
