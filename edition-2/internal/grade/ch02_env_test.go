package grade

import (
	"github.com/waywardgeek/ensemble/internal/fakevendor"
	"strings"
	"testing"
)

func TestCh2ReplayTargetEnvironment(t *testing.T) {
	for _, vendor := range Ch2Vendors {
		for _, live := range []bool{false, true} {
			endpoint := ""
			want := ch2RequestedModel(vendor)
			if live {
				endpoint = "http://127.0.0.1:1"
				want = fakevendor.Models[vendor]
			}
			got := ""
			for _, e := range vendorEnv(vendor, endpoint, t.TempDir()) {
				if strings.HasPrefix(e, "LLM_RESOLVED_MODEL=") {
					got = strings.TrimPrefix(e, "LLM_RESOLVED_MODEL=")
				}
			}
			if got != want {
				t.Fatalf("%s live=%v target=%q want=%q", vendor, live, got, want)
			}
		}
	}
}
