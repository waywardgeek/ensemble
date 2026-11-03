package gradeedition3

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAttachmentPickerCandidate(t *testing.T) {
	candidate := os.Getenv("CH26_GUI_CANDIDATE")
	if candidate == "" {
		t.Skip("explicit disposable GUI candidate required")
	}
	driver, err := filepath.Abs("testdata/ch26-attachments.cjs")
	if err != nil {
		t.Fatal(err)
	}
	config := AttachmentProbe{
		Candidate:        candidate,
		Node:             os.Getenv("CH26_NODE"),
		BrowserDriver:    driver,
		PlaywrightModule: os.Getenv("CH26_PLAYWRIGHT"),
	}
	err = RunAttachmentPickerProbe(config)
	if err != nil {
		t.Fatal(err)
	}
	config.BrowserMutation = "lost-caption"
	err = RunAttachmentPickerProbe(config)
	if err == nil || !strings.Contains(err.Error(), "Caption changed in transport") {
		t.Fatalf("lost-caption mutation must fail the caption gate, got %v", err)
	}
}
