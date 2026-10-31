package gradeedition2

import (
	"encoding/json"
	"path/filepath"
	"testing"

	"github.com/waywardgeek/ensemble/internal/grade"
)

func TestChapter20ExecutableScope(t *testing.T) {
	total := 0
	for _, check := range grade.Ch25Checks(grade.Ch25Result{}) {
		total += check.Points
	}
	if total != 100 {
		t.Fatalf("fleet scope drifted: %d", total)
	}
	for _, skills := range []string{"", "relative", filepath.Join(t.TempDir(), "missing")} {
		if _, err := Chapter20Executable("must-not-launch", skills); err == nil {
			t.Fatalf("invalid skill scope accepted: %q", skills)
		}
	}
}

func TestChapter20ExecutableObservations(t *testing.T) {
	body := func(rows ...map[string]any) []byte {
		messages := []any{}
		for _, row := range rows {
			messages = append(messages, map[string]any{"role": "user", "content": []any{row}})
		}
		b, _ := json.Marshal(map[string]any{"messages": messages})
		return b
	}
	result := func(id, text string, bad bool) map[string]any {
		return map[string]any{"type": "tool_result", "tool_use_id": id, "content": text, "is_error": bad}
	}
	running := result("join", "job 7, status: running, bytes: 0, truncated: false, output_path: /owned/7\n", false)
	reply, _, bad := c20Await(body(running), "join", "join_agents", nil)
	if bad || reply == nil || reply.ToolName != "wait_for_job" {
		t.Fatalf("running join did not continue its Job: %#v", reply)
	}
	var args map[string]any
	if json.Unmarshal([]byte(reply.ToolArgs), &args) != nil || args["handle"] != float64(7) {
		t.Fatalf("wrong wait handle: %s", reply.ToolArgs)
	}
	done := result("join-wait", "job 7, status: done, bytes: 40\n{\"submissions\":{\"child\":{\"result\":\"ok\"}}}\n", false)
	stale := result("join-wait", "job 7, status: running, bytes: 0\n", false)
	reply, text, bad := c20Await(body(running, stale, done), "join", "join_agents", nil)
	if reply != nil || bad || c20JSON(text)["submissions"] == nil {
		t.Fatal("completed latest observation lost")
	}
	refused := result("join-wait", "child stopped without submission", true)
	_, _, bad = c20Await(body(running, refused), "join", "join_agents", nil)
	if !bad {
		t.Fatal("failed child was treated as successful completion")
	}
	if c20JSON("ordinary prose") != nil {
		t.Fatal("prose parsed as submission")
	}

	hint := "Hint received before dispatch; not executed."
	reply, _, bad = c20Await(body(result("join", hint, true)), "join", "join_agents", map[string]any{"agent_ids": []string{"child"}})
	if bad || reply == nil || reply.ToolName != "join_agents" {
		t.Fatal("undispatched join was not retried")
	}
	reply, _, bad = c20Await(body(running, result("join-wait", hint, true)), "join", "join_agents", nil)
	if bad || reply == nil || reply.ToolName != "wait_for_job" {
		t.Fatal("undispatched wait lost its original Job")
	}
	reply, _, bad = c20Await(body(result("join", "different failure", true)), "join", "join_agents", nil)
	if !bad || reply != nil {
		t.Fatal("real failure was retried")
	}
}
