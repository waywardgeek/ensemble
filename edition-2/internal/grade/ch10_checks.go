package grade

import (
	"fmt"
	"strings"
)

// Ch10Evaluate scores chapter 10 (skills engine).
func Ch10Evaluate(r *Ch10Result) []Check {
	checks := []Check{
		ch10InitialTools(r),
		ch10SystemPrompt(r),
		ch10EnsembleTools(r),
		ch10LoadSkill(r),
		ch10ProgressiveDisclosure(r),
		ch10DependsAutoload(r),
		ch10VarSubstitution(r),
		ch10BlockedSkill(r),
		ch10Ch9Parity(r),
	}
	return checks
}

// initial-tools (10 pts): Agent starts with only primary skill's tools.
func ch10InitialTools(r *Ch10Result) Check {
	c := Check{ID: "initial-tools", Title: "Initial tool declarations match primary skill", Points: 10, Earned: 10, Passed: true}

	if !r.BuildOK {
		c.failf("build failed: %s", r.BuildErr)
		return c
	}

	if r.InitialToolsErr != "" {
		c.failf("%s", r.InitialToolsErr)
		return c
	}

	if len(r.InitialTools) == 0 {
		c.failf("no tools declared in first request")
		return c
	}

	// The base skill declares tools: read_file think.
	// load_skill and unload_skill are framework-provided (always present).
	// So we expect at least read_file, think, load_skill, unload_skill.
	expected := map[string]bool{
		"read_file":    false,
		"think":        false,
		"load_skill":   false,
		"unload_skill": false,
	}

	for _, t := range r.InitialTools {
		if _, ok := expected[t]; ok {
			expected[t] = true
		}
	}

	var missing []string
	for name, found := range expected {
		if !found {
			missing = append(missing, name)
		}
	}

	if len(missing) > 0 {
		c.failf("missing expected initial tools: %v (got: %v)", missing, r.InitialTools)
		return c
	}

	// Verify code-tools' tools are NOT present initially.
	for _, t := range r.InitialTools {
		if t == "edit_file" || t == "write_file" {
			c.failf("%s should not be in initial tools — code-tools not loaded yet", t)
			return c
		}
	}

	c.notef("initial tools: %v", r.InitialTools)
	return c
}

// system-prompt (10 pts): System prompt contains primary skill's rendered body.
func ch10SystemPrompt(r *Ch10Result) Check {
	c := Check{ID: "system-prompt", Title: "System prompt contains primary skill body", Points: 10, Earned: 10, Passed: true}

	if !r.BuildOK {
		c.failf("build failed: %s", r.BuildErr)
		return c
	}

	if r.SystemPromptErr != "" {
		c.failf("%s", r.SystemPromptErr)
		return c
	}

	if !r.SystemPromptOK {
		c.failf("system prompt does not contain primary skill body text")
		return c
	}

	c.notef("system prompt contains base skill body (len=%d)", len(r.SystemPrompt))
	return c
}

// ensemble-tools (10 pts): With EN_PRIMARY_SKILL=ensemble, all core tools visible.
func ch10EnsembleTools(r *Ch10Result) Check {
	c := Check{ID: "ensemble-tools", Title: "Ensemble skill declares all core tools", Points: 10, Earned: 10, Passed: true}

	if !r.BuildOK {
		c.failf("build failed: %s", r.BuildErr)
		return c
	}

	if r.EnsembleToolsErr != "" {
		c.failf("%s", r.EnsembleToolsErr)
		return c
	}

	if len(r.EnsembleTools) == 0 {
		c.failf("no tools in ensemble request")
		return c
	}

	// The ensemble skill declares all 12 core tools.
	expected := []string{
		"run_command", "wait_for_job", "send_input", "kill_job",
		"read_file", "write_file", "edit_file", "list_directory",
		"search_files", "think", "load_skill", "unload_skill",
	}

	toolSet := make(map[string]bool)
	for _, t := range r.EnsembleTools {
		toolSet[t] = true
	}

	var missing []string
	for _, e := range expected {
		if !toolSet[e] {
			missing = append(missing, e)
		}
	}

	if len(missing) > 0 {
		c.failf("missing tools: %v (got %d: %v)", missing, len(r.EnsembleTools), r.EnsembleTools)
		return c
	}

	// Also verify the system prompt contains the ensemble skill body.
	if r.EnsemblePrompt != "" && !strings.Contains(r.EnsemblePrompt, "Ensemble agent") {
		c.failf("ensemble system prompt does not contain expected body text")
		return c
	}

	c.notef("all %d core tools present", len(expected))
	return c
}

// load-skill (15 pts): Loading code-tools adds its tools.
func ch10LoadSkill(r *Ch10Result) Check {
	c := Check{ID: "load-skill", Title: "load_skill adds new tool declarations", Points: 15, Earned: 15, Passed: true}

	if !r.BuildOK {
		c.failf("build failed: %s", r.BuildErr)
		return c
	}

	if !r.LoadSkillOK {
		c.failf("load_skill failed: %s", r.LoadSkillErr)
		return c
	}

	// After loading code-tools (tools: edit_file write_file), both should appear.
	foundEdit := false
	foundWrite := false
	for _, t := range r.PostLoadTools {
		if t == "edit_file" {
			foundEdit = true
		}
		if t == "write_file" {
			foundWrite = true
		}
	}

	if !foundEdit {
		c.failf("edit_file not in tools after loading code-tools: %v", r.PostLoadTools)
		return c
	}
	if !foundWrite {
		c.failf("write_file not in tools after loading code-tools: %v", r.PostLoadTools)
		return c
	}

	c.notef("post-load tools: %v", r.PostLoadTools)
	return c
}

// progressive-disclosure (15 pts): Loading code-tools makes search-tools loadable.
func ch10ProgressiveDisclosure(r *Ch10Result) Check {
	c := Check{ID: "progressive-disclosure", Title: "Loading a skill reveals new loadable skills", Points: 15, Earned: 15, Passed: true}

	if !r.BuildOK {
		c.failf("build failed: %s", r.BuildErr)
		return c
	}

	if r.DisclosureErr != "" {
		c.failf("%s", r.DisclosureErr)
		return c
	}

	if !r.DisclosureOK {
		c.failf("search_files not found in tools — search-tools was not loadable after code-tools")
		return c
	}

	c.notef("post-disclosure tools: %v", r.PostDisclosureTools)
	return c
}

// depends-autoload (10 pts): Loading search-tools auto-loads search-helpers.
func ch10DependsAutoload(r *Ch10Result) Check {
	c := Check{ID: "depends-autoload", Title: "Skill dependencies auto-loaded", Points: 10, Earned: 10, Passed: true}

	if !r.BuildOK {
		c.failf("build failed: %s", r.BuildErr)
		return c
	}

	if r.DependsErr != "" {
		c.failf("%s", r.DependsErr)
		return c
	}

	if !r.DependsOK {
		c.failf("list_directory not found — search-helpers dependency not auto-loaded")
		return c
	}

	c.notef("depends chain verified: search-tools → search-helpers (list_directory present)")
	return c
}

// var-substitution (10 pts): $CUSTOM_VAR in skill body renders the registered value.
func ch10VarSubstitution(r *Ch10Result) Check {
	c := Check{ID: "var-substitution", Title: "$VAR tokens rendered in skill body", Points: 10, Earned: 10, Passed: true}

	if !r.BuildOK {
		c.failf("build failed: %s", r.BuildErr)
		return c
	}

	if r.VarSubErr != "" {
		c.failf("%s", r.VarSubErr)
		return c
	}

	if !r.VarSubOK {
		c.failf("$CUSTOM_VAR not found in rendered skill body")
		return c
	}

	// Also verify the literal $CUSTOM_VAR is NOT present (it was substituted).
	if strings.Contains(r.VarSubBody, "$CUSTOM_VAR") {
		c.failf("literal $CUSTOM_VAR still present — substitution did not occur")
		return c
	}

	c.notef("rendered body contains 'hello-from-grader' — substitution works")
	return c
}

// blocked-skill (5 pts): A skill not in any loadable-skills chain is rejected.
func ch10BlockedSkill(r *Ch10Result) Check {
	c := Check{ID: "blocked-skill", Title: "Unreachable skill rejected by load_skill", Points: 5, Earned: 5, Passed: true}

	if !r.BuildOK {
		c.failf("build failed: %s", r.BuildErr)
		return c
	}

	if r.BlockedErr != "" {
		c.failf("%s", r.BlockedErr)
		return c
	}

	if !r.BlockedOK {
		c.failf("load_skill(blocked) should have returned an error")
		return c
	}

	c.notef("blocked skill correctly rejected: %s", r.BlockedResult)
	return c
}

// ch9-parity (15 pts): All ch9 checks pass.
func ch10Ch9Parity(r *Ch10Result) Check {
	c := Check{ID: "ch9-parity", Title: "Chapter 9 checks still pass", Points: 15, Earned: 15, Passed: true}

	if r.Ch9Err != "" {
		c.failf("ch9 harness error: %s", r.Ch9Err)
		return c
	}
	if r.Ch9Result == nil {
		c.failf("ch9 result is nil")
		return c
	}

	ch9Checks := Ch9Evaluate(r.Ch9Result)
	var failures []string
	for _, ch := range ch9Checks {
		if !ch.Passed {
			failures = append(failures, fmt.Sprintf("%s: %s", ch.ID, ch.Details))
		}
	}

	if len(failures) > 0 {
		c.failf("ch9 regressions:\n  %s", strings.Join(failures, "\n  "))
		return c
	}

	c.notef("all ch9 checks pass")
	return c
}
