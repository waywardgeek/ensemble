package grade

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/waywardgeek/ensemble/internal/fakevendor"
)

// A control does not allocate work, but its result can describe existing work.
// Keep legacy absent metadata compatible; this is not the new preservation gate.
func TestCh4ControlObservedJob(t *testing.T) {
	cases := []struct {
		name   string
		change func(map[string]any, map[string]any)
		want   bool
	}{
		{"existing-job-observation", func(c, r map[string]any) {}, true},
		{"legacy-no-observation", func(c, r map[string]any) { delete(r, "job") }, true},
		{"wrapper-on-call", func(c, r map[string]any) { c["job"] = r["job"] }, false},
		{"wrapper-without-return-snapshot", func(c, r map[string]any) { c["job"] = r["job"]; delete(r, "job") }, false},
		{"wrong-handle", func(c, r map[string]any) { r["job"].(map[string]any)["handle"] = float64(4) }, false},
		{"wrong-status", func(c, r map[string]any) { r["job"].(map[string]any)["status"] = "running" }, false},
		{"wrong-output", func(c, r map[string]any) { r["job"].(map[string]any)["output"] = map[string]any{"locator": "cr/io/4"} }, false},
		{"wrong-bytes", func(c, r map[string]any) { r["job"].(map[string]any)["bytes"] = float64(999) }, false},
		{"invented-exit", func(c, r map[string]any) { r["job"].(map[string]any)["exit_code"] = float64(0) }, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			fixture, called, returned := ch4ControlFixture(t)
			tc.change(called, returned)
			got := checkJobModel(fixture)
			if got.Passed != tc.want {
				t.Fatalf("passed=%v want=%v: %+v", got.Passed, tc.want, got)
			}
			if !tc.want {
				wantDiagnostic := "observation does not match"
				if strings.HasPrefix(tc.name, "wrapper-") {
					wantDiagnostic = "allocated a wrapper Job"
				}
				if len(got.Details) != 1 || !strings.Contains(got.Details[0], wantDiagnostic) {
					t.Fatalf("wrong rejection: %+v", got.Details)
				}
			}
			if tc.want && got.Earned != 25 {
				t.Fatalf("earned=%d want25", got.Earned)
			}
		})
	}
}

func ch4ControlFixture(t *testing.T) (*Ch4Result, map[string]any, map[string]any) {
	t.Helper()
	s := &Ch4Session{Ch3Session: &Ch3Session{Answers: []string{"finished"}}, IO: map[string][]byte{}}
	var blocks []any
	addWire := func(id, text string, failed bool) {
		blocks = append(blocks, map[string]any{"type": "tool_result", "tool_use_id": id, "content": text, "is_error": failed})
	}
	addLog := func(kind, id string, job map[string]any) map[string]any {
		tool := map[string]any{"callid": id}
		if job != nil {
			tool["job"] = job
		}
		s.Log = append(s.Log, Ch2LogLine{Type: kind, Data: map[string]any{"tool": tool}})
		return tool
	}
	job := func(handle int, text string) map[string]any {
		return map[string]any{"handle": float64(handle), "status": "done", "bytes": float64(len(text)), "output": map[string]any{"locator": fmt.Sprintf("cr/io/%d", handle)}}
	}
	for i, id := range []string{"toolu_jm_read", "toolu_jm_ls", "toolu_jm_sh"} {
		text := fmt.Sprintf("result-%d", i+1)
		s.IO[fmt.Sprint(i+1)] = []byte(text)
		addLog("tool_called", id, job(i+1, text))
		addLog("tool_returned", id, job(i+1, text))
		addWire(id, text, false)
	}
	called := addLog("tool_called", "toolu_jm_wait", nil)
	returned := addLog("tool_returned", "toolu_jm_wait", job(1, "result-1"))
	addWire("toolu_jm_wait", "finished", false)
	cwd := "/owned/testdata/dbg"
	j := job(4, cwd+"\n")
	j["cwd"] = cwd
	addLog("tool_returned", "toolu_jm_cwd", j)
	addWire("toolu_jm_cwd", cwd+"\n", false)
	addLog("tool_returned", "toolu_jm_cwd_after", job(5, "/owned\n"))
	addWire("toolu_jm_cwd_after", "/owned\n", false)
	addWire("toolu_jm_cwd_missing", "unavailable", true)
	body, err := json.Marshal(map[string]any{"messages": []any{map[string]any{"role": "user", "content": blocks}}})
	if err != nil {
		t.Fatal(err)
	}
	s.Requests = []fakevendor.Recorded{{Vendor: "anthropic", Body: body}}
	return &Ch4Result{JobModel: s}, called, returned
}
