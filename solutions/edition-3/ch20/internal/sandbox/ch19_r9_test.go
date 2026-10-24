package sandbox

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"ensemble/internal/common"
)

func TestCh19R9LiteralRootRead(t *testing.T) {
	profile := Profile("/private/tmp/owned-work", "/private/var/folders/fixture/T")
	if !strings.Contains(profile, `(allow file-read* (literal "/"))`) {
		t.Fatal("missing required literal / directory-entry read")
	}
	for _, broad := range []string{`(subpath "/")`, `(allow file-read-metadata)`, `(subpath "/private/tmp")`, `(subpath "/tmp")`, `(literal "/private/var/select/sh")`} {
		if strings.Contains(profile, broad) {
			t.Fatalf("unexpected broadened grant: %s", broad)
		}
	}
	if !strings.Contains(profile, `xcrun_db[^/]*$`) {
		t.Fatal("bounded immediate cache grant lost")
	}
}

func TestCh19R9PathErrorSpelling(t *testing.T) {
	top := t.TempDir()
	root := filepath.Join(top, "workspace")
	outside := filepath.Join(top, "outside")
	for _, dir := range []string{root, outside, filepath.Join(root, "sub")} {
		if err := os.MkdirAll(dir, 0700); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(root, "inside"), []byte("INSIDE"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, filepath.Join(root, "out-link")); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink("loop", filepath.Join(root, "loop")); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink("inside", filepath.Join(root, "in-link")); err != nil {
		t.Fatal(err)
	}
	files, err := New(&ch19Parent{cfg: common.Config{BaseDir: root, SandboxRoot: root}})
	if err != nil {
		t.Fatal(err)
	}
	defer files.Close()
	for _, path := range []string{"../outside/file", "out-link/file", "./loop", filepath.Join(outside, "file")} {
		if _, err := files.ValidatePath(path); err == nil || !strings.Contains(err.Error(), path) {
			t.Errorf("refusal must identify supplied %q: %v", path, err)
		} else if !filepath.IsAbs(path) && strings.Contains(err.Error(), top) {
			t.Errorf("synthesized host context: %v", err)
		}
	}
	if _, err := files.ReadFile("./missing"); !os.IsNotExist(err) || !strings.Contains(err.Error(), "./missing") || strings.Contains(err.Error(), top) {
		t.Fatalf("filesystem error lost spelling/cause: %v", err)
	}
	for _, path := range []string{"inside", "sub/../inside", "in-link", filepath.Join(root, "inside")} {
		data, err := files.ReadFile(path)
		if err != nil || string(data) != "INSIDE" {
			t.Fatalf("allowed %q: %q %v", path, data, err)
		}
	}
}
