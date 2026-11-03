package gradeedition3

import (
	"os"
	"testing"
)

func TestVoiceExternalCandidate(t *testing.T) {
	runtime := os.Getenv("CH26_VOICE_RUNTIME")
	if runtime == "" {
		t.Skip("explicit candidate runtime and tools required")
	}
	report, err := Chapter26Voice(VoiceGradeConfig{
		RuntimeDirectory: runtime, Dotnet: os.Getenv("CH26_DOTNET"),
		Node: os.Getenv("CH26_NODE"), PlaywrightModule: os.Getenv("CH26_PLAYWRIGHT"),
	})
	if err != nil {
		t.Fatal(err)
	}
	if !report.Passed {
		t.Fatalf("voice scope failed: %+v", report)
	}
}
