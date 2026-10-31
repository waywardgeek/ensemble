package gradeedition2

import "testing"

// This integration exercises the separate Go module, its real registration,
// framework loop, handler and logger; only the HTTP provider is a fake.
func TestChapter4ExternalConsumer(t *testing.T) {
	const core = "../../solutions/ch04"
	example, err := ReadConsumerCase("")
	if err != nil {
		t.Fatal(err)
	}
	checks := append(exerciseConsumer(core+"/ch05", example), consumerTopology(core, core+"/ch05"))
	for _, check := range checks {
		if !check.Passed {
			t.Errorf("%s: %v", check.ID, check.Details)
		}
	}
}
