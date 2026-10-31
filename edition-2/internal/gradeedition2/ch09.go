package gradeedition2

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"time"

	"github.com/waywardgeek/ensemble/internal/fakevendor"
	"github.com/waywardgeek/ensemble/internal/grade"
)

// Chapter9 preserves the original skill weights and composes corrected parity
// through separate clients. Graph validation, provenance and Agent isolation
// retain their source/integration evidence; a wire score does not prove them.
func Chapter9(bin, core, gui, guiModel, consumer, workflow, exercise, streamModel string, example ConsumerCase, media map[string]string) (grade.Report, error) {
	checks := skillsChecks(gui, guiModel)
	previous, err := Chapter8(bin, core, gui, guiModel, consumer, workflow, exercise, streamModel, example, media)
	if err != nil {
		return grade.Report{}, err
	}
	parity := grade.Check{ID: "ch8-parity", Title: "All corrected Chapter 8 checks pass", Points: 15, Earned: 15, Passed: true}
	for _, c := range previous.Checks {
		if !c.Passed {
			fail(&parity, "%s: %s", c.ID, strings.Join(c.Details, "; "))
		}
	}
	return grade.NewTitledReport("Edition 2, Chapter 9 — Skills", append(checks, parity), ""), nil
}

// Keep the historical eight weighted claims. Their evidence now comes from
// the actual optional client, while earlier core checks use their own binaries.
// Chapter9Executable retains the eight skill checks without inherited Go topology.
func Chapter9Executable(bin, static, model string) grade.Report {
	return grade.NewTitledReport("Edition 3 C# — Chapter 9 executable skill checks only (85 points)", skillsExecutableChecks(bin, static, model, nil), "")
}

func skillsChecks(dir, model string) []grade.Check {
	bin, cleanup, err := grade.Build(dir)
	if err != nil {
		return skillsExecutableChecks("", "", model, fmt.Errorf("build: %w", err))
	}
	defer cleanup()
	static, err := filepath.Abs(filepath.Join(dir, "web", "gui"))
	return skillsExecutableChecks(bin, static, model, err)
}

func skillsExecutableChecks(bin, static, model string, setupError error) []grade.Check {
	checks := []grade.Check{
		{ID: "initial-tools", Title: "Primary limits declarations and actual dispatch", Points: 10},
		{ID: "system-prompt", Title: "Loaded instructions reach system context", Points: 10},
		{ID: "ensemble-tools", Title: "Full primary declares the installed core tools", Points: 10},
		{ID: "load-skill", Title: "Loading enables instructions and tools; unload remains pending", Points: 15},
		{ID: "progressive-disclosure", Title: "Loaded skills reveal exactly the next discovery level", Points: 15},
		{ID: "depends-autoload", Title: "Dependencies supply tools and instructions", Points: 10},
		{ID: "var-substitution", Title: "Skill variables render at point of use", Points: 10},
		{ID: "blocked-skill", Title: "Unreachable skill returns an error without granting tools", Points: 5},
	}
	for i := range checks {
		checks[i].Earned = checks[i].Points
		checks[i].Passed = true
	}
	if setupError != nil {
		for i := range checks {
			fail(&checks[i], "setup: %v", setupError)
		}
		return checks
	}
	// One real conversation preserves the primary while capabilities change. The
	// hidden write uses a different path: later success cannot conceal its effect.
	calls := []fakevendor.Reply{
		skillCall("hidden", "write_file", `{"path":"hidden.txt","content":"FORBIDDEN"}`),
		skillCall("code", "load_skill", `{"name":"code-tools"}`),
		skillCall("write", "write_file", `{"path":"enabled.txt","content":"ENABLED"}`),
		skillCall("search", "load_skill", `{"name":"search-tools"}`),
		skillCall("unload", "unload_skill", `{"name":"code-tools"}`),
		skillCall("pending", "write_file", `{"path":"pending.txt","content":"PENDING"}`),
		{Text: "Skills exercise complete."},
	}
	requests, files, err := driveSkills(bin, static, model, "base", calls)
	if err != nil {
		for _, i := range []int{0, 1, 3, 4, 5, 6} {
			fail(&checks[i], "skill conversation: %v", err)
		}
	} else {
		initial := []string{"read_file", "load_skill", "unload_skill"}
		loaded := append(append([]string{}, initial...), "write_file", "edit_file")
		all := append(append([]string{}, loaded...), "search_files", "list_directory")
		skillDeclarations(&checks[0], requests[0], initial)
		skillResult(&checks[0], requests[1], "hidden", true)
		if _, exists := files["hidden.txt"]; exists {
			fail(&checks[0], "hidden write executed")
		}
		primary := skillSystem(requests[0])
		if !strings.Contains(primary, "PRIMARY_BODY") || !strings.Contains(primary, "BODY_AFTER_RULE") || strings.Contains(primary, "$TOOLS") || strings.Contains(primary, "$SKILLS") {
			fail(&checks[1], "primary body/variables/horizontal rule not rendered: %q", primary)
		}
		if strings.Contains(primary, "CODE_BODY") || strings.Contains(primary, "SEARCH_BODY") {
			fail(&checks[1], "unloaded instructions present at startup")
		}
		for i, r := range requests {
			system := skillSystem(r)
			if !strings.Contains(system, "PRIMARY_BODY") || !strings.Contains(system, "BODY_AFTER_RULE") {
				fail(&checks[1], "request %d lost primary instructions", i)
			}
			if i >= 2 && !strings.Contains(system, "CODE_BODY") {
				fail(&checks[1], "request %d lacks loaded code instructions", i)
			}
			if i >= 4 && (!strings.Contains(system, "SEARCH_BODY") || !strings.Contains(system, "HELPER_BODY")) {
				fail(&checks[1], "request %d lacks loaded search/dependency instructions", i)
			}
		}
		// Results are read from the next provider request: a GUI acknowledgement or
		// a scripted final answer cannot establish that the real handler succeeded.
		skillDeclarations(&checks[3], requests[2], loaded)
		skillResult(&checks[3], requests[2], "code", false)
		_, body, hasCode := strings.Cut(skillSystem(requests[2]), "CODE_BODY")
		if !hasCode {
			fail(&checks[3], "loaded instructions absent from system context")
		}
		skillResult(&checks[3], requests[3], "write", false)
		skillResult(&checks[3], requests[5], "unload", false)
		skillResult(&checks[3], requests[6], "pending", false)
		skillDeclarations(&checks[3], requests[6], all)
		for name, want := range map[string]string{"enabled.txt": "ENABLED", "pending.txt": "PENDING"} {
			if files[name] != want {
				fail(&checks[3], "%s: want actual bytes %q, got %q", name, want, files[name])
			}
		}
		skillResult(&checks[4], requests[4], "search", false)
		searchSystem := skillSystem(requests[4])
		_, search, hasSearch := strings.Cut(searchSystem, "SEARCH_BODY")
		if !hasSearch {
			fail(&checks[4], "search instructions absent from system context")
		}
		skillDeclarations(&checks[4], requests[4], all)
		// Delimited text is authored by the fixture, not a required renderer format.
		// Names/descriptions may be formatted freely inside each expansion.
		skillExpansion(&checks[4], primary, "SKILLS", []string{"code-tools"}, []string{"search-tools", "search-helpers", "blocked"})
		skillExpansion(&checks[4], body, "SKILLS", []string{"search-tools"}, []string{"code-tools", "search-helpers", "blocked"})
		skillExpansion(&checks[4], search, "SKILLS", nil, []string{"code-tools", "search-tools", "search-helpers", "blocked"})
		skillDeclarations(&checks[5], requests[4], all)
		if !strings.Contains(searchSystem, "HELPER_BODY") {
			fail(&checks[5], "dependency instructions absent from system context")
		}
		if !strings.Contains(searchSystem, "SEARCH_BODY hello-from-grader") || strings.Contains(search, "$CUSTOM_VAR") || strings.Contains(body, "hello-from-grader") || strings.Contains(primary, "hello-from-grader") {
			fail(&checks[6], "custom variable missing/unrendered or appears before its skill loads")
		}
		skillExpansion(&checks[6], primary, "TOOLS", initial, []string{"write_file", "edit_file", "search_files", "list_directory"})
		skillExpansion(&checks[6], body, "TOOLS", loaded, []string{"search_files", "list_directory"})
		skillExpansion(&checks[6], search, "TOOLS", all, nil)
	}
	// Independent initial configurations must use new Agents, not another prompt
	// in the already-expanded conversation. Both still enter through WebSocket.
	full, _, err := driveSkills(bin, static, model, "ensemble", []fakevendor.Reply{{Text: "Full primary ready."}})
	if err != nil {
		fail(&checks[2], "full primary: %v", err)
	} else {
		skillDeclarations(&checks[2], full[0], strings.Fields("read_file write_file edit_file list_directory search_files run_command wait_for_job send_input kill_job tool_limits load_skill unload_skill"))
		if !strings.Contains(skillSystem(full[0]), "FULL_PRIMARY_BODY") {
			fail(&checks[2], "full primary body missing")
		}
	}
	blocked, _, err := driveSkills(bin, static, model, "base", []fakevendor.Reply{skillCall("blocked", "load_skill", `{"name":"blocked"}`), {Text: "Refusal observed."}})
	if err != nil {
		fail(&checks[7], "blocked skill: %v", err)
	} else {
		skillResult(&checks[7], blocked[1], "blocked", true)
		skillDeclarations(&checks[7], blocked[1], []string{"read_file", "load_skill", "unload_skill"})
	}
	return checks
}

func skillCall(id, name, args string) fakevendor.Reply {
	return fakevendor.Reply{ToolID: id, ToolName: name, ToolArgs: args}
}

// Correlation excludes earlier call arguments and unrelated old results, which
// can contain the expected marker even when the target handler failed.
func skillResult(c *grade.Check, r fakevendor.Recorded, id string, wantError bool) string {
	text, bad, err := anthropicToolText(r.Body, id)
	meaningful := strings.TrimSpace(text) != "" || (!wantError && skillReferenceResult(r, id))
	if err != nil || bad != wantError || !meaningful {
		fail(c, "result %s: error=%t want=%t text=%q decode=%v", id, bad, wantError, text, err)
	}
	return text
}

// Anthropic expands references instead of ordinary text in successful discovery
// results. This exception requires the actual preceding load call, not merely a
// reference appearing somewhere in the conversation. Catalog validity is checked
// separately by skillDeclarations.
func skillReferenceResult(r fakevendor.Recorded, id string) bool {
	paired := false
	for _, rawMessage := range list(r.JSON()["messages"]) {
		message := object(rawMessage)
		for _, rawBlock := range list(message["content"]) {
			block := object(rawBlock)
			if message["role"] == "assistant" && block["type"] == "tool_use" && block["id"] == id {
				paired = block["name"] == "load_skill"
			}
			if block["type"] != "tool_result" || block["tool_use_id"] != id {
				continue
			}
			if !paired || message["role"] != "user" || block["is_error"] == true {
				return false
			}
			content := list(block["content"])
			if len(content) == 0 {
				return false
			}
			for _, rawReference := range content {
				reference := object(rawReference)
				name, _ := reference["tool_name"].(string)
				if reference["type"] != "tool_reference" || name == "" {
					return false
				}
			}
			return true
		}
	}
	return false
}

// Deferred Anthropic definitions are a server-side catalog, not yet exposed
// capabilities. A standard tool_result's tool_reference expands a catalog entry.
// Repeated references in replay are harmless; duplicate definitions are not.
func skillDeclarations(c *grade.Check, r fakevendor.Recorded, want []string) {
	body := r.JSON()
	catalog := map[string]bool{}
	exposed := map[string]bool{}
	tools, _ := body["tools"].([]any)
	for _, tool := range tools {
		definition := object(tool)
		name, _ := definition["name"].(string)
		if name == "" || catalog[name] {
			fail(c, "empty or duplicate tool definition %q", name)
		}
		catalog[name] = true
		if deferred, _ := definition["defer_loading"].(bool); !deferred {
			exposed[name] = true
		}
	}
	messages, _ := body["messages"].([]any)
	pending := map[string]bool{}
	for _, message := range messages {
		m := object(message)
		blocks, _ := m["content"].([]any)
		for _, block := range blocks {
			result := object(block)
			if m["role"] == "assistant" && result["type"] == "tool_use" {
				if id, _ := result["id"].(string); id != "" {
					pending[id] = true
				}
				continue
			}
			if result["type"] != "tool_result" {
				continue
			}
			id, _ := result["tool_use_id"].(string)
			paired := m["role"] == "user" && id != "" && pending[id]
			if paired {
				delete(pending, id)
			}
			content, _ := result["content"].([]any)
			referenceCount := 0
			for _, item := range content {
				if object(item)["type"] == "tool_reference" {
					referenceCount++
				}
			}
			if referenceCount > 0 && referenceCount != len(content) {
				fail(c, "tool references are mixed with other result content")
			}
			for _, item := range content {
				reference := object(item)
				if reference["type"] != "tool_reference" {
					continue
				}
				if !paired {
					fail(c, "tool reference is not in a paired user tool result")
					continue
				}
				name, _ := reference["tool_name"].(string)
				if !catalog[name] {
					fail(c, "tool reference %q has no definition", name)
				}
				exposed[name] = true
			}
		}
	}
	var names []string
	for name := range exposed {
		names = append(names, name)
	}
	sort.Strings(names)
	expected := append([]string{}, want...)
	sort.Strings(expected)
	if !reflect.DeepEqual(names, expected) {
		fail(c, "exposed tools: want %v, got %v", expected, names)
	}
}

// Both documented Anthropic system representations carry the same text.
// Compare that instruction content without requiring one serializer shape.
func skillSystem(r fakevendor.Recorded) string {
	switch value := r.JSON()["system"].(type) {
	case string:
		return value
	case []any:
		var parts []string
		for _, part := range value {
			if text, ok := object(part)["text"].(string); ok {
				parts = append(parts, text)
			}
		}
		return strings.Join(parts, "\n")
	}
	return ""
}

func skillExpansion(c *grade.Check, body, key string, want, absent []string) {
	_, tail, ok := strings.Cut(body, key+"_BEGIN")
	expansion, _, end := strings.Cut(tail, key+"_END")
	if !ok || !end || strings.Contains(expansion, "$"+key) {
		fail(c, "missing rendered %s expansion", key)
		return
	}
	for _, name := range want {
		if !strings.Contains(expansion, name) {
			fail(c, "%s expansion lacks %s", key, name)
		}
	}
	for _, name := range absent {
		if strings.Contains(expansion, name) {
			fail(c, "%s expansion prematurely/incorrectly exposes %s", key, name)
		}
	}
}

// Only the provider is fake: the submitted binary, skill parser, registry,
// tool dispatch, file writes and next-request rendering all execute normally.
// Each process owns its scratch settings and skill selection; no global test
// environment mutation can change another concurrent fixture's identity.
func driveSkills(bin, static, model, primary string, replies []fakevendor.Reply) ([]fakevendor.Recorded, map[string]string, error) {
	work, err := os.MkdirTemp("", "edition2-skills-")
	if err != nil {
		return nil, nil, err
	}
	defer os.RemoveAll(work)
	skills := filepath.Join(work, "skills")
	if err := writeSkillFixtures(skills); err != nil {
		return nil, nil, err
	}
	provider := fakevendor.New(replies)
	defer provider.Close()
	url, stop, err := startGUI(bin, static, model, provider.URL(), work, "EN_SKILLS_DIR="+skills, "EN_PRIMARY_SKILL="+primary, "EN_CUSTOM_VAR=hello-from-grader")
	if err != nil {
		return nil, nil, err
	}
	defer stop()
	peer, err := openGUIPeer(url)
	if err != nil {
		return nil, nil, err
	}
	defer peer.close()
	if err = peer.send("subscribe", ""); err != nil {
		return nil, nil, err
	}
	if err = peer.send("prompt", "Exercise the fixture skills."); err != nil {
		return nil, nil, err
	}
	if _, err = peer.until(10*time.Second, messageKind("turn_ended")); err != nil {
		return nil, nil, err
	}
	requests := provider.Requests()
	if len(requests) != len(replies) {
		return nil, nil, fmt.Errorf("expected %d actual requests, got %d", len(replies), len(requests))
	}
	files := map[string]string{}
	for _, name := range []string{"hidden.txt", "enabled.txt", "pending.txt"} {
		data, e := os.ReadFile(filepath.Join(work, name))
		if e == nil {
			files[name] = string(data)
		} else if !os.IsNotExist(e) {
			return nil, nil, e
		}
	}
	return requests, files, nil
}

func writeSkillFixtures(dir string) error {
	// The original fixture advertised undefined think. Only already taught tools
	// are declared here, including the inherited one-call tool_limits capability.
	expansion := "\nTOOLS_BEGIN\n$TOOLS\nTOOLS_END\nSKILLS_BEGIN\n$SKILLS\nSKILLS_END\n"
	files := map[string]string{
		"base":           "---\nname: base\ndescription: Skills grading fixture\ntype: primary\ntools: read_file\nloadable-skills: code-tools\n---\nPRIMARY_BODY" + expansion + "\n---\nBODY_AFTER_RULE\n",
		"code-tools":     "---\nname: code-tools\ndescription: Skills grading fixture\ntype: loadable\ntools:\n  - edit_file\n  - write_file\nloadable-skills: search-tools\n---\nCODE_BODY" + expansion,
		"search-tools":   "---\nname: search-tools\ndescription: Skills grading fixture\ntype: loadable\ntools: search_files\ndepends: search-helpers\n---\nSEARCH_BODY $CUSTOM_VAR" + expansion,
		"search-helpers": "---\nname: search-helpers\ndescription: Skills grading fixture\ntype: dependency\ntools: list_directory\n---\nHELPER_BODY\n",
		"blocked":        "---\nname: blocked\ndescription: Skills grading fixture\ntype: loadable\ntools: run_command\n---\nBLOCKED_BODY\n",
		"ensemble":       "---\nname: ensemble\ndescription: Skills grading fixture\ntype: primary\ntools: read_file write_file edit_file list_directory search_files run_command wait_for_job send_input kill_job tool_limits load_skill unload_skill\n---\nFULL_PRIMARY_BODY\n",
	}
	for name, body := range files {
		path := filepath.Join(dir, name)
		if err := os.MkdirAll(path, 0755); err != nil {
			return err
		}
		if err := os.WriteFile(filepath.Join(path, "SKILL.md"), []byte(body), 0644); err != nil {
			return err
		}
	}
	return nil
}
