package gui

import (
	"regexp"
	"strings"
	"unicode"
)

// speechText holds raw offsets, never offsets in filtered output. An incomplete
// fence, link, identifier or sentence remains with its own part until stable.
type speechText struct{ consumed int }

func (s *speechText) update(text string, final bool) string {
	if s.consumed > len(text) {
		return ""
	}
	end := len(text)
	if !final {
		end = stableSpeech(text, s.consumed)
	}
	if end <= s.consumed {
		return ""
	}
	out := cleanSpeech(text[s.consumed:end])
	s.consumed = end
	return out
}

// Resolve multiline syntax before selecting phrase boundaries. In particular,
// a newline inside a fenced block cannot cause its delimiter to leak to speech.
func stableSpeech(text string, start int) int {
	end, fence, inline, brackets, parens := start, "", false, 0, 0
	lineStart := start == 0 || text[start-1] == '\n'
	for i := start; i < len(text); i++ {
		c := text[i]
		if lineStart {
			lineEnd := strings.IndexByte(text[i:], '\n')
			if lineEnd < 0 {
				lineEnd = len(text) - i
			}
			line := strings.TrimSpace(text[i : i+lineEnd])
			if strings.HasPrefix(line, "```") || strings.HasPrefix(line, "~~~") {
				marker := line[:3]
				if fence == "" {
					fence = marker
				} else if fence == marker {
					fence = ""
				}
				i += lineEnd
				lineStart = true
				continue
			}
		}
		lineStart = c == '\n'
		if fence != "" {
			continue
		}
		if c == '`' {
			inline = !inline
			continue
		}
		if inline {
			continue
		}
		switch c {
		case '[':
			brackets++
		case ']':
			if brackets > 0 {
				brackets--
			}
		case '(':
			parens++
		case ')':
			if parens > 0 {
				parens--
			}
		}
		if brackets > 0 || parens > 0 {
			continue
		}
		if c == '\n' && i > start && text[i-1] == '\n' {
			end = i + 1
		}
		if strings.ContainsRune(".!?", rune(c)) && i+1 < len(text) && unicode.IsSpace(rune(text[i+1])) {
			end = i + 1
		}
	}
	return end
}

var (
	speechLink = regexp.MustCompile(`!?\[([^\]]*)\]\([^)]*\)`)
	speechURL  = regexp.MustCompile(`https?://[^\s<>]+`)
	speechTag  = regexp.MustCompile(`<[^>\n]+>`)
	// Prefixes and rules are line-local: \s would also consume paragraph newlines.
	speechMarker   = regexp.MustCompile(`(?m)^[ \t]{0,3}(?:#{1,6}[ \t]+|>[ \t]*|[-+*][ \t]+|[0-9]+[.)][ \t]+)`)
	speechRule     = regexp.MustCompile(`(?m)^[ \t]*(?:[-*_][ \t]*){3,}$`)
	speechAcronym  = regexp.MustCompile(`([A-Z]+)([A-Z][a-z])`)
	speechCamel    = regexp.MustCompile(`([a-z0-9])([A-Z])`)
	speechSentence = regexp.MustCompile(`([.!?])\s+`)
)

func cleanSpeech(text string) string {
	text = strings.ReplaceAll(text, "\r\n", "\n")
	var lines []string
	fence := ""
	for _, line := range strings.Split(text, "\n") {
		trimmed := strings.TrimSpace(line)
		if fence != "" {
			if strings.HasPrefix(trimmed, fence) {
				fence = ""
			}
			continue
		}
		if strings.HasPrefix(trimmed, "```") || strings.HasPrefix(trimmed, "~~~") {
			fence = trimmed[:3]
			lines = append(lines, "code block")
			continue
		}
		lines = append(lines, line)
	}
	text = strings.Join(lines, "\n")
	text = speechLink.ReplaceAllString(text, "$1")
	text = speechURL.ReplaceAllStringFunc(text, func(url string) string {
		// Preserve sentence punctuation after a bare URL, but not the URL itself.
		return url[len(strings.TrimRight(url, ".,;:!?")):]
	})
	text = speechTag.ReplaceAllString(text, "")
	text = speechRule.ReplaceAllString(text, "")
	text = speechMarker.ReplaceAllString(text, "")
	text = strings.NewReplacer("`", "", "*", "", "_", " ", "~~", "", "&amp;", "&", "&lt;", "<", "&gt;", ">").Replace(text)
	text = speechAcronym.ReplaceAllString(text, "$1 $2")
	text = speechCamel.ReplaceAllString(text, "$1 $2")
	// Sentence and paragraph boundaries survive; a soft wrap is just whitespace.
	var phrases []string
	for _, paragraph := range strings.Split(text, "\n\n") {
		paragraph = strings.Join(strings.Fields(paragraph), " ")
		paragraph = speechSentence.ReplaceAllString(paragraph, "$1\n\n")
		for _, phrase := range strings.Split(paragraph, "\n\n") {
			if strings.ContainsFunc(phrase, func(r rune) bool { return unicode.IsLetter(r) || unicode.IsDigit(r) }) {
				phrases = append(phrases, phrase)
			}
		}
	}
	return strings.Join(phrases, "\n\n")
}

func speechSource(c Card) string {
	if c.IsError || c.Kind == "error" {
		return "error"
	}
	if c.Kind == "tool_call" {
		return "tool"
	}
	return c.Kind
}
