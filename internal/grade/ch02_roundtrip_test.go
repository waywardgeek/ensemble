package grade

import (
	"strings"
	"testing"
)

func TestCh2RoundTripFixturePreservesAndCompletes(t *testing.T) {
	const dump = `{"log_version":1}
{"seq":10,"type":"response_ended","response":{"parts":[{"type":"tool_call","call_id":"one"},{"type":"tool_call","call_id":"two"}]}}
{"seq":20,"type":"tool_returned","tool":{"call_id":"one","parts":[{"type":"text","text":"actual result"}]}}
`
	lines, parseErr := parseLogLines(dump)
	got, err := ch2RoundTripFixture(&VendorSession{DumpOut: dump, Log: lines, LogErr: parseErr})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(string(got), dump) {
		t.Fatal("fixture must preserve every original dump byte")
	}
	completed, parseErr := parseLogLines(string(got))
	if parseErr != "" || len(completed) != len(lines)+1 {
		t.Fatalf("expected exactly one added result: %s", parseErr)
	}
	last := completed[len(completed)-1]
	if last.Seq != 21 || last.Type != normName("tool_returned") || getStr(getMap(last.Data, "tool"), "call_id") != "two" {
		t.Fatalf("wrong fixture completion: %+v", last)
	}
	// Completed histories must not acquire another synthetic result on replay.
	again, err := ch2RoundTripFixture(&VendorSession{DumpOut: string(got), Log: completed})
	if err != nil || string(again) != string(got) {
		t.Fatal("complete history changed")
	}
}

func TestCh2RoundTripFixtureDoesNotRepairMalformedDump(t *testing.T) {
	if _, err := ch2RoundTripFixture(&VendorSession{DumpOut: "broken", LogErr: "line 1 invalid"}); err == nil {
		t.Fatal("malformed input was hidden by fixture completion")
	}
}
