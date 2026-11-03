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
	config.InputMode = "paste"
	if err := RunAttachmentPickerProbe(config); err != nil {
		t.Fatalf("paste/removal: %v", err)
	}
	config.Recipient = true
	if err := RunAttachmentPickerProbe(config); err != nil {
		t.Fatalf("recipient: %v", err)
	}
	config.Fault = "late-reader"
	if err := RunAttachmentPickerProbe(config); err != nil {
		t.Fatalf("late reader: %v", err)
	}
	config.Recipient = false
	config.Fault = "refusal"
	if err := RunAttachmentPickerProbe(config); err != nil {
		t.Fatalf("refusal recovery: %v", err)
	}
	config.Fault = "uncertain"
	if err := RunAttachmentPickerProbe(config); err != nil {
		t.Fatalf("uncertain reconnect: %v", err)
	}
	config.Fault = ""
	config.BrowserMutation = "lost-caption"
	err = RunAttachmentPickerProbe(config)
	if err == nil || !strings.Contains(err.Error(), "Caption changed in transport") {
		t.Fatalf("lost-caption mutation must fail the caption gate, got %v", err)
	}
}
