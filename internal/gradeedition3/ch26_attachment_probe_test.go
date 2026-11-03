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
	config.Fault = "acknowledgements"
	if err := RunAttachmentPickerProbe(config); err != nil {
		t.Fatalf("acknowledgements: %v", err)
	}
	config.Fault = "duplicate-wire"
	if err := RunAttachmentPickerProbe(config); err != nil {
		t.Fatalf("duplicate wire: %v", err)
	}
	config.Fault = ""
	for _, negative := range []struct{ mutation, refusal string }{
		{"lost-caption", "Caption changed in transport"},
		{"preview-only", "Preview-only: no image envelope was sent"},
		{"misdirected", "Image envelope targets the wrong Agent"},
		{"duplicate-send", "Duplicate or missing image admission request"},
	} {
		config.BrowserMutation = negative.mutation
		config.Recipient = negative.mutation == "misdirected"
		err = RunAttachmentPickerProbe(config)
		if err != nil && (strings.Contains(err.Error(), "shutdown") || strings.Contains(err.Error(), "cleanup")) {
			t.Fatalf("negative control cleanup failed: %v", err)
		}
		if err == nil || !strings.Contains(err.Error(), negative.refusal) {
			t.Fatalf("%s must fail its intended gate, got %v", negative.mutation, err)
		}
	}
}

func TestAttachmentSetupRefusesBeforeCandidate(t *testing.T) {
	root := t.TempDir()
	candidate := filepath.Join(root, "candidate")
	node := filepath.Join(root, "node")
	marker := filepath.Join(root, "launched")
	if err := os.WriteFile(candidate, []byte("#!/bin/sh\ntouch '"+marker+"'\n"), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(node, []byte("#!/bin/sh\necho synthetic-browser-unavailable >&2\nexit 3\n"), 0700); err != nil {
		t.Fatal(err)
	}
	report, err := Chapter26Attachments(AttachmentProbe{Candidate: candidate, Node: node, PlaywrightModule: root})
	if err == nil || !strings.Contains(err.Error(), "not a candidate grade") {
		t.Fatalf("expected setup refusal, got %v", err)
	}
	if report.Passed {
		t.Fatal("setup refusal produced a passing candidate grade")
	}
	if _, err := os.Stat(marker); !os.IsNotExist(err) {
		t.Fatalf("candidate launched during setup refusal: %v", err)
	}
}
