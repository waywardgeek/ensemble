package ensemble

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDiagnosticRetention(t *testing.T) {
	path := filepath.Join(t.TempDir(), "api.log")
	limit := 2
	d, err := OpenDiagnosticLog(path, func() int { return limit })
	if err != nil {
		t.Fatal(err)
	}
	for _, record := range []string{"first\n", strings.Repeat("x", 100000) + "\n", "last\n"} {
		if _, err = d.Write([]byte(record)); err != nil {
			t.Fatal(err)
		}
	}
	if err = d.Retain(); err != nil {
		t.Fatal(err)
	}
	b, _ := os.ReadFile(path)
	if strings.HasPrefix(string(b), "first") || strings.Count(string(b), "\n") != 2 {
		t.Fatal("wrong retention")
	}
	limit = 0
	if _, err = d.Write([]byte("unlimited\n")); err != nil {
		t.Fatal(err)
	}
	b, _ = os.ReadFile(path)
	if strings.Count(string(b), "\n") != 3 {
		t.Fatal("zero did not preserve records")
	}
}
