package gradeedition2

import (
	"encoding/json"
	"github.com/waywardgeek/ensemble/internal/fakevendor"
	"testing"
)

func TestChapter15ExecutableMeterContract(t *testing.T) {
	decode := func(s string) map[string]any {
		t.Helper()
		var m map[string]any
		if err := json.Unmarshal([]byte(s), &m); err != nil {
			t.Fatal(err)
		}
		return m
	}
	empty := decode(`{"type":"usage","usage":{"input_tokens":null,"cache_write_tokens":null,"cache_read_tokens":null,"output_tokens":null,"all_input":null,"totals":{"input_tokens":0,"cache_write_tokens":0,"cache_read_tokens":0,"output_tokens":0}},"total_cost_usd":0,"hit_rate":null}`)
	if !invoiceEmptySession(empty) {
		t.Fatal("published empty session refused")
	}
	empty["hit_rate"] = float64(0)
	if invoiceEmptySession(empty) {
		t.Fatal("zero invented as an observed hit rate")
	}
	empty["hit_rate"] = nil
	u := empty["usage"].(map[string]any)
	delete(u, "all_input")
	if invoiceEmptySession(empty) {
		t.Fatal("missing latest field accepted as explicit unknown")
	}
	m := decode(`{"usage":{"input_tokens":11,"cache_write_tokens":22,"cache_read_tokens":33,"output_tokens":44,"totals":{"input_tokens":110,"cache_write_tokens":220,"cache_read_tokens":330,"output_tokens":440}}}`)
	if !invoiceCounts(m, "", fakevendor.Canonical{Input: 11, CacheWrite: 22, CacheRead: 33, Output: 44}) || !invoiceCounts(m, "total", fakevendor.Canonical{Input: 110, CacheWrite: 220, CacheRead: 330, Output: 440}) {
		t.Fatal("latest/total scopes mixed")
	}
	m["usage"].(map[string]any)["input_tokens"] = nil
	if invoiceCounts(m, "", fakevendor.Canonical{Input: 11, CacheWrite: 22, CacheRead: 33, Output: 44}) {
		t.Fatal("unknown input credited as known")
	}
}

func TestChapter15ExecutableOpenAIContract(t *testing.T) {
	u := fakevendor.Canonical{Input: 100, CacheRead: 1400, Output: 20}
	var body map[string]any
	if err := json.Unmarshal([]byte(invoiceExecutableOpenAI("gpt-4o-mini-2024-07-18", "default", "ok", u)), &body); err != nil {
		t.Fatal(err)
	}
	if body["model"] != "gpt-4o-mini-2024-07-18" || body["service_tier"] != "default" {
		t.Fatalf("identity lost: %v", body)
	}
	usage := body["usage"].(map[string]any)
	if usage["prompt_tokens"] != float64(1500) || usage["completion_tokens"] != float64(20) {
		t.Fatalf("counts: %v", usage)
	}
	details := usage["prompt_tokens_details"].(map[string]any)
	if details["cached_tokens"] != float64(1400) || len(details) != 1 {
		t.Fatalf("cached read contract: %v", details)
	}
	if _, exists := usage["cache_write_input_tokens"]; exists {
		t.Fatal("invented explicit-write field")
	}
}
