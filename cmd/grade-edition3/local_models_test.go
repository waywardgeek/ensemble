package main

import (
	"os"
	"path/filepath"
	"strconv"
	"testing"
)

func TestLocalModelFlagsRefuseBeforeTools(t *testing.T) {
	root := t.TempDir()
	marker := filepath.Join(root, "ran")
	tool := filepath.Join(root, "tool")
	if err := os.WriteFile(tool, []byte("#!/bin/sh\necho ran > "+strconv.Quote(marker)+"\nexit 1\n"), 0700); err != nil {
		t.Fatal(err)
	}
	base := []string{"--scope", "local-models", "-executable-only", "-dotnet", tool}
	cases := [][]string{{"--scope", "local-models", root}, append(append([]string{}, base...), root)}
	for _, extra := range [][]string{{"-ch", "27"}, {"-browser-node", tool}, {"-playwright-module", root}, {"-media-scope", "voice"}, {"-core", root}} {
		args := append(append([]string{}, base...), extra...)
		cases = append(cases, append(args, root))
	}
	for _, args := range cases {
		if code := run(args); code != 2 {
			t.Fatalf("%q exit %d", args, code)
		}
		if _, err := os.Stat(marker); !os.IsNotExist(err) {
			t.Fatal("invalid input launched tool")
		}
	}
}
