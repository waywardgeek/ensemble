package agent

import (
	"ensemble/internal/common"
	"testing"
)

func TestNextTurnBudgetPolicyDoesNotChangeCapturedLimit(t *testing.T) {
	// The running turn has its own captured boundary. Updating the next-turn
	// policy must not rewrite it or the immutable constructor configuration.
	a := &agent{config: common.Config{MaxToolRounds: 7}, active: &turn{limit: 7}}
	a.maxToolRounds.Store(7)
	for _, rounds := range []int{1, 12345, 0} {
		if err := a.SetMaxToolRounds(rounds); err != nil {
			t.Fatal(err)
		}
		if a.Config().MaxToolRounds != rounds || a.active.limit != 7 || a.config.MaxToolRounds != 7 {
			t.Fatalf("policy %d changed the captured boundary", rounds)
		}
	}
	if err := a.SetMaxToolRounds(-1); err == nil {
		t.Fatal("negative policy accepted")
	}
	if a.Config().MaxToolRounds != 0 {
		t.Fatal("rejected update changed the policy")
	}
}
