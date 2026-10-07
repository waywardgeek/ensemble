package recall

// Chunking: turning a markdown file into pieces small enough to quote.
//
// The unit of recall is not the file. A 20KB design document is useless as a
// recalled snippet — it would blow the byte cap on its own and bury the one
// paragraph that mattered. The unit is a section, because markdown authors
// already did the work of deciding where one idea stops and the next starts,
// and a `## ` header is that decision written down.

import (
	"fmt"
	"strings"

	"github.com/waywardgeek/ensemble/agent/internal/common"
)

const (
	// maxChunkBytes is the size above which a section is split further.
	// Chosen so a single chunk can never dominate the 6KB recall budget.
	maxChunkBytes = 4096
	// sectionPrefix and subSectionPrefix are the markdown headers chunking
	// splits on, in that order.
	sectionPrefix    = "## "
	subSectionPrefix = "### "
	// breadcrumbSep joins a parent header to a sub-header so a reader of a
	// sub-chunk can still tell where in the document it came from.
	breadcrumbSep = " > "
)

// ChunkMarkdown splits one markdown file into chunks.
//
// Three passes, each a fallback for the one before: `## ` sections, then
// `### ` sub-sections for any section still over the limit, then paragraph
// boundaries for anything that is still too big (a long section with no
// sub-headers, which is common in notes). A file with no headers at all
// comes back as a single chunk, or as paragraphs if it is large.
func ChunkMarkdown(filename, text string) []common.Chunk {
	var out []common.Chunk
	for _, sec := range splitOn(text, sectionPrefix) {
		if len(sec.body) <= maxChunkBytes {
			out = append(out, emit(filename, sec.header, sec.body)...)
			continue
		}
		subs := splitOn(sec.body, subSectionPrefix)
		for _, sub := range subs {
			header := sec.header
			if sub.header != "" && sub.header != sec.header {
				header = breadcrumb(sec.header, sub.header)
			}
			if len(sub.body) <= maxChunkBytes {
				out = append(out, emit(filename, header, sub.body)...)
				continue
			}
			for _, para := range splitParagraphs(sub.body) {
				out = append(out, emit(filename, header, para)...)
			}
		}
	}
	return out
}

// breadcrumb formats a sub-chunk's header as "Parent > Sub".
func breadcrumb(parent, sub string) string {
	if parent == "" {
		return sub
	}
	return fmt.Sprintf("%s%s%s", parent, breadcrumbSep, sub)
}

// section is one header and the text under it.
type section struct {
	header string
	body   string
}

// splitOn cuts text at lines beginning with prefix.
//
// Text before the first such line becomes a leading section with an empty
// header — preamble is still worth indexing, and a file whose only header is
// a `# ` title would otherwise vanish entirely.
func splitOn(text, prefix string) []section {
	lines := strings.Split(text, "\n")
	var out []section
	cur := section{}
	var buf []string
	flush := func() {
		body := strings.TrimSpace(strings.Join(buf, "\n"))
		if body != "" {
			cur.body = body
			out = append(out, cur)
		}
		buf = nil
	}
	for _, ln := range lines {
		if strings.HasPrefix(ln, prefix) {
			flush()
			cur = section{header: strings.TrimSpace(strings.TrimPrefix(ln, prefix))}
			continue
		}
		buf = append(buf, ln)
	}
	flush()
	return out
}

// splitParagraphs is the last resort: blank-line boundaries, then a hard cut
// for any single paragraph that is still over the limit. The hard cut is
// ugly and rare — a minified blob or a pasted log with no blank lines — but
// the alternative is one chunk that eats the whole recall budget.
func splitParagraphs(text string) []string {
	var out []string
	cur := ""
	for _, p := range strings.Split(text, "\n\n") {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		if cur != "" && len(cur)+len(p)+2 > maxChunkBytes {
			out = append(out, cur)
			cur = ""
		}
		if cur == "" {
			cur = p
		} else {
			cur = cur + "\n\n" + p
		}
		for len(cur) > maxChunkBytes {
			out = append(out, cur[:maxChunkBytes])
			cur = cur[maxChunkBytes:]
		}
	}
	if strings.TrimSpace(cur) != "" {
		out = append(out, cur)
	}
	return out
}

// emit builds a chunk, dropping anything that is only whitespace.
func emit(filename, header, body string) []common.Chunk {
	body = strings.TrimSpace(body)
	if body == "" {
		return nil
	}
	return []common.Chunk{{Filename: filename, Header: header, Content: body}}
}
