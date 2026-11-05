package main

import (
	"os"
	"path/filepath"
	"strconv"
	"testing"
)

func TestArtifactFlagsRefuseBeforeTools(t *testing.T) {
	root := t.TempDir()
	marker := filepath.Join(root, "ran")
	tool := filepath.Join(root, "tool")
	if err := os.WriteFile(tool, []byte("#!/bin/sh\necho ran > "+strconv.Quote(marker)+"\nexit 1\n"), 0700); err != nil {
		t.Fatal(err)
	}
	base := []string{"--scope", "artifacts", "-executable-only", "-dotnet", tool, "-browser-node", tool, "-playwright-module", root}
	cases := [][]string{{"--scope", "unknown", root}, {"--scope", "", root}, {"--scope", "artifacts", root}}
	for _, flags := range [][]string{{"-ch", "1"}, {"-ch", "26"}, {"-media-scope", "image-files"}, {"-allow-native-host-screen"}, {"-core", root}, {"-unknown"}} {
		args := append(append([]string{}, base...), flags...)
		cases = append(cases, append(args, root))
	}
	cases = append(cases, append(append([]string{}, base...), ""), append(append([]string{}, base...), " "))
	cases = append(cases, append(append([]string{}, base...), root)) // Missing candidate assemblies.
	for _, args := range cases {
		if got := run(args); got != 2 {
			t.Fatalf("%q returned %d", args, got)
		}
		if _, err := os.Stat(marker); !os.IsNotExist(err) {
			t.Fatal("invalid input launched tool")
		}
	}
}
