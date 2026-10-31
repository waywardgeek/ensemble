package gradeedition2

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRecallExecutableScope(t *testing.T) {
	checks := recallExecutableChecks()
	total := 0
	for _, c := range checks {
		total += c.Points
		if c.ID == "bm25-indexes" || c.ID == "recall-is-own-kind" || c.ID == "recall-is-a-spoke" {
			t.Fatalf("excluded row: %s", c.ID)
		}
	}
	if len(checks) != 7 || total != 70 {
		t.Fatalf("scope %d/%d", len(checks), total)
	}
	raw, err := recallExecutableSettings()
	if err != nil {
		t.Fatal(err)
	}
	var doc map[string]any
	if err = json.Unmarshal(raw, &doc); err != nil {
		t.Fatal(err)
	}
	if doc["version"] != float64(1) {
		t.Fatal("version")
	}
	memory := object(doc["memory"])
	if len(memory) != 6 {
		t.Fatal("memory policies")
	}
	for name, band := range memory {
		if object(band)["disabled"] != true {
			t.Fatalf("live memory %s enabled", name)
		}
	}
	recall := object(doc["recall"])
	if len(recall) != 9 || recall["disabled"] != false || recall["disable_judge"] != false || recall["max_recall_bytes"] != float64(6144) {
		t.Fatalf("recall: %s", raw)
	}
}

func TestRecallExecutableEvents(t *testing.T) {
	path := filepath.Join(t.TempDir(), "events.jsonl")
	valid := "{\"type\":\"recall_added\",\"parts\":[{\"type\":\"recall\",\"text\":\"retained fact\",\"hashes\":[]}]}\n{\"type\":\"agent_shutdown\"}\n"
	if err := os.WriteFile(path, []byte(valid), 0600); err != nil {
		t.Fatal(err)
	}
	events, err := recallExecutableEvents(path)
	if err != nil {
		t.Fatal(err)
	}
	texts := recallAttachments(recallOut{events: events})
	if len(texts) != 1 || texts[0] != "retained fact" {
		t.Fatalf("attachments %q", texts)
	}
	for _, bad := range []string{"", strings.TrimSuffix(valid, "\n"), "{}\n", "broken\n"} {
		if err = os.WriteFile(path, []byte(bad), 0600); err != nil {
			t.Fatal(err)
		}
		if _, err = recallExecutableEvents(path); err == nil {
			t.Fatalf("accepted %q", bad)
		}
	}
}
