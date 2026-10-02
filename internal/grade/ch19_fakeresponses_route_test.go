package grade

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"testing"
)

// The fake picks the route from the credential, as the vendor does. The plan
// route's two differences from the documented metered route were measured live
// (2026-10-02); each broke every plan turn of an agent built from the docs, and
// the grader could not see either while its fake sent the documented shape.
func TestCh19RespRouteFollowsTheCredential(t *testing.T) {
	cases := []struct {
		name        string
		auth        string
		wantCT      bool // Content-Type header present at all
		wantOutputs bool // response.completed.output non-empty
	}{
		{"plan token", ch19RespTestAuth, false, false},
		{"metered key", "Bearer " + ch19APIKey, true, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			s := ch19NewFakeResponses(ch19RespOptions{Scenarios: []string{"text"}})
			defer s.Close()

			raw, err := json.Marshal(ch19RespGoodBody())
			if err != nil {
				t.Fatal(err)
			}
			req, err := http.NewRequest(http.MethodPost, s.URL()+"/v1/responses", bytes.NewReader(raw))
			if err != nil {
				t.Fatal(err)
			}
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", tc.auth)
			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				t.Fatal(err)
			}
			defer resp.Body.Close()
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				t.Fatal(err)
			}

			_, hasCT := resp.Header["Content-Type"]
			if hasCT != tc.wantCT {
				t.Errorf("Content-Type present = %v (%q), want %v", hasCT, resp.Header.Get("Content-Type"), tc.wantCT)
			}

			var completedOutputs, doneItems int
			for _, ev := range ch19RespParseSSE(t, body) {
				switch ev.Name {
				case "response.output_item.done":
					doneItems++
				case "response.completed":
					r, _ := ev.Data["response"].(map[string]any)
					out, _ := r["output"].([]any)
					completedOutputs = len(out)
				}
			}
			if doneItems == 0 {
				t.Errorf("no response.output_item.done: on either route the items must arrive somewhere")
			}
			if (completedOutputs > 0) != tc.wantOutputs {
				t.Errorf("response.completed.output has %d items, want non-empty=%v", completedOutputs, tc.wantOutputs)
			}
		})
	}
}
