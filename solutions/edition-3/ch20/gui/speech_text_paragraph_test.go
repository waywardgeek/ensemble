package gui

import (
	"strings"
	"testing"
)

func TestSpeechMarkupParagraphBoundaries(t *testing.T) {
	for _, tt := range []struct {
		name, input, want string
	}{
		{"heading", "The harbor grows quiet\n\n## Evening notes\n\nLanterns glow", "The harbor grows quiet\n\nEvening notes\n\nLanterns glow"},
		{"quote and bullet", "The trail bends west\n\n> Rest beneath the cedar\n\n- Refill the flask", "The trail bends west\n\nRest beneath the cedar\n\nRefill the flask"},
		{"numbered markers", "Lanterns are packed\n\n1. Carry the map\n\n2) Check the latch", "Lanterns are packed\n\nCarry the map\n\nCheck the latch"},
		{"rule", "The kettle warms\n\n---\n\nPour tea slowly", "The kettle warms\n\nPour tea slowly"},
		{"indented markup", "Morning frost\n\n  ### Orchard notes\n\n\t* * *\n\nBirds return", "Morning frost\n\nOrchard notes\n\nBirds return"},
		{"soft wraps", "The river runs\nunder the old bridge\npast the mill", "The river runs under the old bridge past the mill"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			if got := cleanSpeech(tt.input); got != tt.want {
				t.Errorf("full formatter = %q; want %q", got, tt.want)
			}
			var filter speechText
			if got := filter.update(tt.input, true); got != tt.want {
				t.Errorf("final-only formatter = %q; want %q", got, tt.want)
			}
			if got := filter.update(tt.input, true); got != "" {
				t.Errorf("repeated final emitted %q", got)
			}
		})
	}
}

func TestSpeechParagraphGrouping(t *testing.T) {
	parts := []string{
		"Clouds drift past the ridge\n\n",
		"## Trail log\n\n",
		"The bridge boards\nfeel firm\n\n",
		"> Listen for the brook\n\n",
		"---\n\n",
		"We reach the orchard",
	}
	text := strings.Join(parts, "")
	// Independent expectations: comparing streaming only with cleanSpeech(text)
	// could let both paths agree on the same lost paragraph boundary.
	want := "Clouds drift past the ridge\n\nTrail log\n\nThe bridge boards feel firm\n\nListen for the brook\n\nWe reach the orchard"
	for _, tt := range []struct {
		name   string
		groups []int
	}{
		{"one coalesced update", []int{6}},
		{"mixed coalesced updates", []int{2, 3, 1}},
		{"paragraph updates", []int{1, 1, 1, 1, 1, 1}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			var filter speechText
			var output []string
			n := 0
			for _, size := range tt.groups {
				n += size
				if phrase := filter.update(strings.Join(parts[:n], ""), false); phrase != "" {
					output = append(output, phrase)
				}
			}
			if phrase := filter.update(text, true); phrase != "" {
				output = append(output, phrase)
			}
			if got := strings.Join(output, "\n\n"); got != want {
				t.Errorf("grouped formatter = %q; want %q", got, want)
			}
			if got := filter.update(text, true); got != "" {
				t.Errorf("repeated final emitted %q", got)
			}
		})
	}
}
