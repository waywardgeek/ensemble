package sandbox

import (
	"strconv"
	"strings"
	"testing"
)

func TestCh19R10DeveloperSelectionMetadata(t *testing.T) {
	profile := Profile("/private/tmp/r10-work", "/private/var/folders/fixture/T")
	for _, path := range []string{"/var", "/private/var/select", "/private/var/select/developer_dir"} {
		quoted := strconv.Quote(path)
		want := "(allow file-read-metadata (literal " + quoted + "))"
		if strings.Count(profile, want) != 1 {
			t.Errorf("need one metadata literal for %s", path)
		}
		for _, line := range strings.Split(profile, "\n") {
			if strings.Contains(line, quoted) && line != want {
				t.Errorf("selection path received more than literal metadata: %s", line)
			}
		}
	}
	for _, path := range []string{"/var", "/private/var", "/var/select", "/private/var/select"} {
		if strings.Contains(profile, "(subpath "+strconv.Quote(path)+")") {
			t.Errorf("selection subtree grant: %s", path)
		}
	}
	for _, path := range []string{"/var/select/sh", "/private/var/select/sh", "/etc/gitattributes", "/private/etc/gitattributes"} {
		if strings.Contains(profile, strconv.Quote(path)) {
			t.Errorf("unrelated warning-path grant: %s", path)
		}
	}
	if strings.Contains(profile, "(allow file-read-metadata)") || strings.Contains(profile, `(subpath "/")`) {
		t.Fatal("global traversal/read grant")
	}
	if !strings.Contains(profile, `(allow file-read* (literal "/"))`) || !strings.Contains(profile, `(deny network*)`) || !strings.Contains(profile, `xcrun_db[^/]*$`) {
		t.Fatal("existing root-entry/network/immediate-cache boundary lost")
	}
}
