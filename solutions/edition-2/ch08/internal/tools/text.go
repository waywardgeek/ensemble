package tools

import (
	"example.com/ensemble/internal/common"
	"unicode/utf8"
)

// Finalize only the retained prefix. Stream writes may split a complete rune,
// so capture buffers bytes until the process finishes before applying this rule.
// The parent keeps diagnostics reachable without duplicating logging services.
func textPrefix(parent common.Registry, text string, limit int) string {
	if len(text) > limit {
		text = text[:limit]
	}
	if len(text) == 0 {
		return text
	}
	start := len(text) - 1
	for start > 0 && !utf8.RuneStart(text[start]) {
		start--
	}
	if !utf8.FullRuneInString(text[start:]) {
		text = text[:start]
	}
	return text
}
