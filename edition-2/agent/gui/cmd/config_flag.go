package main

import (
	"fmt"
	"strings"
)

// Locate the explicit config before reading environment defaults. In file mode
// even an invalid inherited LLM_VENDOR must be irrelevant. flag.Parse still
// owns the full CLI syntax; duplicate configuration selectors are rejected.
func configPathFromArgs(args []string) (string, error) {
	var path string
	seen := false
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if arg == "--" {
			break
		}
		var value string
		switch {
		case arg == "--config" || arg == "-config":
			i++
			if i == len(args) {
				return "", fmt.Errorf("--config requires a path")
			}
			value = args[i]
		case strings.HasPrefix(arg, "--config=") || strings.HasPrefix(arg, "-config="):
			value = strings.SplitN(arg, "=", 2)[1]
		default:
			continue
		}
		if seen || strings.TrimSpace(value) == "" || strings.HasPrefix(value, "--") {
			return "", fmt.Errorf("provide exactly one nonempty --config path")
		}
		path, seen = value, true
	}
	return path, nil
}
