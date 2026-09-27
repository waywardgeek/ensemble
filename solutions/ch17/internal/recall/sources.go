package recall

// Where the corpus comes from.
//
// Five sources, each its own directory of markdown and each its own BM25
// index. They are separate rather than merged because the quota policy in
// recall.go needs to hand out slots per source, and because they have
// genuinely different value: a daily log is what this agent did, a design
// doc is what somebody wrote down, and a SKILL.md is a capability the agent
// does not yet know it has.

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Source names. Exported because the quota policy names one of them and the
// formatter attributes by it.
const (
	SourceMemory    = "memory"
	SourceDocs      = "docs"
	SourceHandoffs  = "handoffs"
	SourceLearnings = "learnings"
	SourceSkills    = "skills"
)

// SkipNewestMemories is how many of the most recent daily logs are left out
// of the index.
//
// Two, because they are already in the system prompt verbatim. Recalling
// them would spend the snippet budget telling the agent something it is
// already looking at, which is the opposite of the job. This is the only
// deduplication the system does across sources; see §17.7.
const SkipNewestMemories = 2

// SourceSpec describes one directory to index.
type SourceSpec struct {
	Name string
	Dir  string
	// Recursive walks subdirectories. Skill docs need it — a SKILL.md lives
	// in a directory named for its skill — and the flat sources do not.
	Recursive bool
	// SkipNewest drops the N newest files by name. Daily logs are named
	// YYYY-MM-DD-N.md, so lexical order is chronological order.
	SkipNewest int
}

// DefaultSources is the standard layout, rooted at the agent's workspace.
//
// A function rather than a package-level slice: a slice at package scope
// would be a mutable global, and any caller could append a source to it and
// change every agent in the process.
func DefaultSources(workspace, skillsDir string) []SourceSpec {
	return []SourceSpec{
		// Chapter 16's session memories. Its store is rooted at
		// <workspace>/memory and keeps daily logs one level further down.
		{Name: SourceMemory, Dir: filepath.Join(workspace, "memory", "memory"), SkipNewest: SkipNewestMemories},
		{Name: SourceDocs, Dir: filepath.Join(workspace, "docs"), Recursive: true},
		{Name: SourceHandoffs, Dir: filepath.Join(workspace, "handoffs")},
		{Name: SourceLearnings, Dir: filepath.Join(workspace, "learnings")},
		{Name: SourceSkills, Dir: skillsDir, Recursive: true},
	}
}

// source is one indexed directory.
type source struct {
	name  string
	index *Index
}

// loadSource reads every markdown file under a spec and indexes it.
//
// A missing directory is not an error. Most agents have no handoffs and many
// have no docs, and recall must degrade to "fewer sources" rather than to a
// startup failure — the whole system is optional by design.
func loadSource(spec SourceSpec) *source {
	files := markdownFiles(spec.Dir, spec.Recursive)
	if spec.SkipNewest > 0 {
		// Sort newest first so the skip takes the most recent.
		sort.Sort(sort.Reverse(sort.StringSlice(files)))
		if len(files) <= spec.SkipNewest {
			files = nil
		} else {
			files = files[spec.SkipNewest:]
		}
	}
	sort.Strings(files)

	ix := newIndex()
	for _, path := range files {
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		name := displayName(spec.Dir, path)
		for _, c := range ChunkMarkdown(name, string(data)) {
			ix.Add(c)
		}
	}
	return &source{name: spec.Name, index: ix}
}

// markdownFiles lists .md files in a directory.
func markdownFiles(dir string, recursive bool) []string {
	if dir == "" {
		return nil
	}
	var out []string
	if !recursive {
		ents, err := os.ReadDir(dir)
		if err != nil {
			return nil
		}
		for _, e := range ents {
			if e.IsDir() || !strings.EqualFold(filepath.Ext(e.Name()), ".md") {
				continue
			}
			out = append(out, filepath.Join(dir, e.Name()))
		}
		return out
	}
	filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		if strings.EqualFold(filepath.Ext(d.Name()), ".md") {
			out = append(out, path)
		}
		return nil
	})
	return out
}

// displayName is what a recalled snippet is attributed to: the path relative
// to its source directory, so "2026-09-27-1.md" rather than an absolute path
// nobody can read.
func displayName(dir, path string) string {
	if rel, err := filepath.Rel(dir, path); err == nil && !strings.HasPrefix(rel, "..") {
		return filepath.ToSlash(rel)
	}
	return filepath.Base(path)
}

// chunkCount reports how many chunks a source indexed.
func (s *source) chunkCount() int { return len(s.index.Chunks) }
