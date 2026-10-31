package fakevendor

import (
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestJSONBodyFixtureOverride(t *testing.T) {
	legacy := Reply{Text: "legacy", Usage: Canonical{Input: 11, CacheWrite: 22, CacheRead: 33, Output: 44}}
	for _, tc := range []struct {
		name   string
		reply  Reply
		status int
		body   string
	}{
		{"default", legacy, 200, openAIBody(legacy)},
		{"explicit", Reply{JSONBody: `{"model":"gpt-5.6","fixture":true}`}, 200, `{"model":"gpt-5.6","fixture":true}`},
		{"error-precedence", Reply{JSONBody: `{"unused":true}`, Status: 503, ErrBody: `{"error":"held"}`}, 503, `{"error":"held"}`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			s := New([]Reply{tc.reply})
			defer s.Close()
			r, err := http.Post(s.URL()+"/v1/chat/completions", "application/json", strings.NewReader(`{"stream":false}`))
			if err != nil {
				t.Fatal(err)
			}
			defer r.Body.Close()
			b, err := io.ReadAll(r.Body)
			if err != nil {
				t.Fatal(err)
			}
			if r.StatusCode != tc.status || string(b) != tc.body || r.Header.Get("Content-Type") != "application/json" {
				t.Fatalf("status=%d content-type=%q body=%s", r.StatusCode, r.Header.Get("Content-Type"), b)
			}
			if len(s.Requests()) != 1 {
				t.Fatal("request recording changed")
			}
		})
	}
}
