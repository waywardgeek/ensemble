package ensemble_test

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"ensemble"
	"ensemble/internal/common"
)

func TestRound3UsageDependencies(t *testing.T) {
	type fixture struct {
		name    string
		vendor  common.Vendor
		usage   string
		want    common.Usage
		missing []string
	}
	var cases []fixture
	// Each invalid operand leaves valid independent measured buckets intact.
	for _, bad := range []struct{ name, value string }{{"malformed", `"bad"`}, {"null", `null`}, {"negative", `-1`}} {
		cases = append(cases,
			fixture{"openai/read-" + bad.name, common.OpenAI, `"usage":{"prompt_tokens":15,"completion_tokens":4,"prompt_tokens_details":{"cached_tokens":` + bad.value + `,"cache_write_tokens":2}}`, common.Usage{Input: -1, Output: 4, CacheRead: -1}, []string{"cache_read", "input"}},
			fixture{"openai/write-" + bad.name, common.OpenAI, `"usage":{"prompt_tokens":15,"completion_tokens":4,"prompt_tokens_details":{"cached_tokens":3,"cache_write_tokens":` + bad.value + `}}`, common.Usage{Input: 12, Output: 4, CacheRead: 3}, nil},
			fixture{"openai/details-" + bad.name, common.OpenAI, `"usage":{"prompt_tokens":15,"completion_tokens":4,"prompt_tokens_details":` + bad.value + `}`, common.Usage{Input: -1, Output: 4, CacheRead: -1}, []string{"cache_read", "input"}},
			fixture{"gemini/read-" + bad.name, common.Gemini, `"usageMetadata":{"promptTokenCount":15,"candidatesTokenCount":2,"thoughtsTokenCount":2,"cachedContentTokenCount":` + bad.value + `}`, common.Usage{Input: -1, Output: 4, CacheRead: -1}, []string{"cache_read", "input"}},
		)
	}
	// Omitted optional cache counts remain zero (Chapter 2); invalid facts do not.
	// A generic OpenAI cache_write_tokens field is not a Chat Completions fact.
	cases = append(cases,
		fixture{"openai/omitted-details", common.OpenAI, `"usage":{"prompt_tokens":15,"completion_tokens":4}`, common.Usage{Input: 15, Output: 4}, nil},
		fixture{"openai/omitted-read", common.OpenAI, `"usage":{"prompt_tokens":15,"completion_tokens":4,"prompt_tokens_details":{"cache_write_tokens":2}}`, common.Usage{Input: 15, Output: 4}, nil},
		fixture{"openai/omitted-write", common.OpenAI, `"usage":{"prompt_tokens":15,"completion_tokens":4,"prompt_tokens_details":{"cached_tokens":3}}`, common.Usage{Input: 12, Output: 4, CacheRead: 3}, nil},
		fixture{"openai/measured-cache", common.OpenAI, `"usage":{"prompt_tokens":15,"completion_tokens":4,"prompt_tokens_details":{"cached_tokens":3,"cache_write_tokens":2}}`, common.Usage{Input: 12, Output: 4, CacheRead: 3}, nil},
		fixture{"gemini/omitted-cache", common.Gemini, `"usageMetadata":{"promptTokenCount":15,"candidatesTokenCount":2,"thoughtsTokenCount":2}`, common.Usage{Input: 15, Output: 4}, nil},
		fixture{"gemini/measured-cache", common.Gemini, `"usageMetadata":{"promptTokenCount":15,"candidatesTokenCount":2,"thoughtsTokenCount":2,"cachedContentTokenCount":3}`, common.Usage{Input: 12, Output: 4, CacheRead: 3}, nil},
	)
	for _, f := range cases {
		t.Run(f.name, func(t *testing.T) {
			content := `"choices":[{"message":{"content":"observed answer"}}]`
			if f.vendor == common.Gemini {
				content = `"candidates":[{"content":{"parts":[{"text":"observed answer"}]}}]`
			}
			body := []byte("{" + content + "," + f.usage + "}")
			calls := 0
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { calls++; _, _ = w.Write(body) }))
			defer server.Close()
			a := r2Agent(t, ensemble.Config{Vendor: f.vendor, BaseURL: server.URL, APIKey: "local-fixture"})
			wantError := len(f.missing) > 0 && !strings.Contains(f.name, "omitted")
			check := func(events []common.Event, err error) {
				t.Helper()
				if (err != nil) != wantError {
					t.Fatalf("error=%v missing=%v", err, f.missing)
				}
				var response *common.ResponseData
				for _, event := range events {
					if event.Type == common.ResponseEnded {
						response = event.Response
					}
				}
				if response == nil || response.Usage != f.want {
					t.Fatalf("response=%+v want measured %+v", response, f.want)
				}
				got := map[string]bool{}
				for _, bucket := range response.UsageMissing {
					got[bucket] = true
				}
				if len(got) != len(f.missing) {
					t.Fatalf("missing=%v want=%v", response.UsageMissing, f.missing)
				}
				for _, bucket := range f.missing {
					if !got[bucket] {
						t.Fatalf("unmeasured %s not marked missing", bucket)
					}
				}
			}
			events, parseErr := a.ParseResponse(200, body)
			check(events, parseErr)
			if calls != 0 || a.Usage() != (common.Usage{}) || a.Log().Context().Usage != (common.Usage{}) {
				t.Fatal("offline Parse charged usage or mutated History")
			}
			_, exchangeErr := a.Ask(context.Background(), "question")
			check(a.Log().Events(), exchangeErr)
			if calls != 1 || a.Usage() != f.want || a.Log().Context().Usage != f.want {
				t.Fatalf("calls=%d live=%+v historical=%+v want=%+v", calls, a.Usage(), a.Log().Context().Usage, f.want)
			}
			if wantError && len(a.History()) != 0 {
				t.Fatal("invalid accounting became a completed answer")
			}
			replay := r2Agent(t, ensemble.Config{Vendor: f.vendor})
			if err := replay.Log().Load(bytes.NewReader(r2Dump(t, a))); err != nil {
				t.Fatal(err)
			}
			if replay.Usage() != (common.Usage{}) || replay.Log().Context().Usage != f.want {
				t.Fatal("replay invented a charge or lost independent measured counts")
			}
			check(replay.Log().Events(), exchangeErr)
		})
	}
}
