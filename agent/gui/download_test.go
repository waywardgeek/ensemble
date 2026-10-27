package gui

import (
	"bytes"
	"strings"
	"testing"
)

func TestDownloadCapabilityIsNotWrittenToTrace(t *testing.T) {
	var log bytes.Buffer
	server := &Server{log: &log}
	token := strings.Repeat("a", 64)
	server.trace("<", []byte(`{"type":"download_offered","token":"`+token+`","filename":"report.pdf"}`))
	if strings.Contains(log.String(), token) || !strings.Contains(log.String(), "[redacted]") {
		t.Fatalf("download capability was not redacted: %s", log.String())
	}
}
