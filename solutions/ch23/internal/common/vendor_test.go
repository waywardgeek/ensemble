package common

import "testing"

// A model's vendor is a fact about the model, so every real row must declare
// one. The zero value is invalid, which makes a forgotten row detectable
// rather than silently Anthropic. Models driven against several vendors by env
// prefix are the one legitimate exception and are named here, so adding
// another is a deliberate act rather than an oversight.
func TestEveryModelDeclaresItsVendor(t *testing.T) {
	multiVendor := map[string]bool{"fake-model": true}

	for id := range models() {
		v, ok := VendorFor(id)
		if multiVendor[id] {
			if ok {
				t.Errorf("%s is driven by env prefix and must not claim vendor %v", id, v)
			}
			continue
		}
		if !ok {
			t.Errorf("%s declares no vendor; add one to its row in models()", id)
		}
	}
}

// The GUI label and the API actually dialed must come from the same mapping.
func TestVendorLabelMatchesDialedVendor(t *testing.T) {
	cases := map[string]struct {
		want  Vendor
		label string
	}{
		"claude-sonnet-5": {VendorAnthropic, "Anthropic"},
		"gpt-5.6-sol":     {VendorOpenAI, "OpenAI"},
	}
	for id, want := range cases {
		got, ok := VendorFor(id)
		if !ok || got != want.want {
			t.Errorf("VendorFor(%s) = %v,%v, want %v", id, got, ok, want.want)
		}
		if label := vendor(id); label != want.label {
			t.Errorf("vendor(%s) = %q, want %q", id, label, want.label)
		}
	}

	// An unknown model must refuse rather than guess: a prefix table would
	// happily return Anthropic here.
	if v, ok := VendorFor("claude-not-a-real-model"); ok {
		t.Errorf("unknown model resolved to %v, want refusal", v)
	}
}
