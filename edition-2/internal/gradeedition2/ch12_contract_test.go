package gradeedition2

import (
	"encoding/json"
	"testing"

	"github.com/waywardgeek/ensemble/internal/fakevendor"
	"github.com/waywardgeek/ensemble/internal/grade"
)

func guiDebugFixtureRequests() []map[string]any {
	ids := []string{"debug-snapshot-before", "debug-queue-before", "debug-click", "debug-input", "debug-snapshot-after", "debug-queue-after", "debug-unload", "debug-guessed", "debug-guessed-snapshot"}
	names := []string{"gui_snapshot", "tts_queue", "gui_click", "gui_input", "gui_snapshot", "tts_queue", "unload_skill", "gui_click", "gui_snapshot"}
	results := []string{"SNAPSHOT_ONLY_1 button=enabled input=<empty>", `[{"text":"TTS_ONLY_1","status":"pending"}]`, "CLICK_RESULT_ONLY", "INPUT_RESULT_ONLY:INPUT_ARGUMENT_ONLY", "SNAPSHOT_ONLY_2 button=disabled input=INPUT_ARGUMENT_ONLY", `[{"text":"TTS_ONLY_2","status":"pending"}]`, "unloaded; instructions pending", "tool unavailable", "tool unavailable"}
	var requests []map[string]any
	for i := 0; i < 10; i++ {
		tools := []any{map[string]any{"name": "load_skill"}, map[string]any{"name": "unload_skill"}}
		if i < 7 {
			for _, n := range []string{"gui_snapshot", "tts_queue", "gui_click", "gui_input"} {
				tools = append(tools, map[string]any{"name": n})
			}
		}
		messages := []any{map[string]any{"role": "user", "content": "Exercise the browser tools."}}
		for j := 0; j < i; j++ {
			messages = append(messages, map[string]any{"role": "assistant", "content": []any{map[string]any{"type": "tool_use", "id": ids[j], "name": names[j], "input": map[string]any{}}}}, map[string]any{"role": "user", "content": []any{map[string]any{"type": "tool_result", "tool_use_id": ids[j], "content": results[j], "is_error": j >= 7}}})
		}
		requests = append(requests, map[string]any{"system": guiDebugPrimaryManual + "\n" + guiDebugStartupManual, "tools": tools, "messages": messages})
	}
	return requests
}

func TestChapter12ExplicitObservations(t *testing.T) {
	cases := []struct {
		name   string
		change func([]map[string]any)
		pass   bool
	}{
		{"explicit results retained after unload", func([]map[string]any) {}, true},
		{"lost startup instructions", func(r []map[string]any) { r[0]["system"] = guiDebugPrimaryManual }, false},
		{"unsolicited snapshot", func(r []map[string]any) {
			r[0]["messages"] = append(r[0]["messages"].([]any), map[string]any{"role": "user", "content": "SNAPSHOT_ONLY_1 button=enabled"})
		}, false},
		{"unsolicited queue in system", func(r []map[string]any) {
			r[3]["system"] = guiDebugPrimaryManual + " " + guiDebugStartupManual + " TTS_ONLY_1"
		}, false},
		{"automatic observation not callable", func(r []map[string]any) { r[0]["tools"] = r[0]["tools"].([]any)[:2] }, false},
		{"declarations survive unload", func(r []map[string]any) { r[7]["tools"] = r[0]["tools"] }, false},
		{"uncorrelated snapshot", func(r []map[string]any) { resultBlock(r[1], 2)["tool_use_id"] = "wrong" }, false},
		{"stale after snapshot", func(r []map[string]any) {
			resultBlock(r[5], 10)["content"] = "SNAPSHOT_ONLY_1 button=enabled input=<empty>"
		}, false},
		{"queue lacks status", func(r []map[string]any) { resultBlock(r[2], 4)["content"] = `[{"text":"TTS_ONLY_1"}]` }, false},
		{"missing click result", func(r []map[string]any) { resultBlock(r[3], 6)["content"] = "ok" }, false},
		{"wrong input effect", func(r []map[string]any) { resultBlock(r[4], 8)["content"] = "INPUT_RESULT_ONLY:wrong" }, false},
		{"guessed click accepted", func(r []map[string]any) { resultBlock(r[8], 16)["is_error"] = false }, false},
		{"guessed snapshot accepted", func(r []map[string]any) { resultBlock(r[9], 18)["is_error"] = false }, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			r := guiDebugFixtureRequests()
			tc.change(r)
			var requests []fakevendor.Recorded
			for _, body := range r {
				b, err := json.Marshal(body)
				if err != nil {
					t.Fatal(err)
				}
				requests = append(requests, fakevendor.Recorded{Body: b})
			}
			checks := grade.Ch13Checks(grade.Ch13Result{SkillLoadsOK: true, EphemeralInjectedOK: true, SnapshotUpdatesOK: true, TTSVisibilityOK: true, GUIInteractionOK: true, SkillUnloadOK: true})[:6]
			guiDebugRequests(checks, "fixture", requests)
			pass := true
			for _, c := range checks {
				if !c.Passed {
					pass = false
				}
			}
			if pass != tc.pass {
				t.Fatalf("pass=%v want%v checks=%+v", pass, tc.pass, checks)
			}
		})
	}
}

func resultBlock(r map[string]any, message int) map[string]any {
	return r["messages"].([]any)[message].(map[string]any)["content"].([]any)[0].(map[string]any)
}
