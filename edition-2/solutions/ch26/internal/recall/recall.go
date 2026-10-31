package recall

import (
	"context"
	"crypto/sha256"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"unicode/utf8"

	"ensemble/internal/common"
)

// Queries can overlap briefly after interruption. The lock protects index state
// only; a slow judge never holds it or blocks a newer turn's corpus scan.
type recaller struct {
	parent  common.Agent
	mu      sync.Mutex
	digest  [32]byte
	indexes map[string]index
}

// New creates one Agent-owned corpus. Indexing waits until the first query so
// callers can load History and finish staged memory recovery before any scan.
func New(parent common.Agent) common.Recaller {
	if parent == nil {
		panic("Recall requires Agent")
	}
	return &recaller{parent: parent}
}

// Agent exposes the required owner of configuration, files and model services.
func (r *recaller) Agent() common.Agent { return r.parent }

// Validate rejects negative limits; zero retains the documented default.
func Validate(s common.RecallSettings) error {
	if s.MinScore < 0 || s.MaxSnippets < 0 || s.MinQueryLength < 0 || s.Candidates < 0 || s.ContextMsgs < 0 || s.MaxRecallBytes < 0 {
		return fmt.Errorf("recall limits must be nonnegative")
	}
	return nil
}

// Defaults are applied to a local value, leaving explicit Agent configuration
// available to clients without copying mutable policy into the child.
func settings(s common.RecallSettings) common.RecallSettings {
	if s.MinScore == 0 {
		s.MinScore = 3
	}
	if s.MaxSnippets == 0 {
		s.MaxSnippets = 3
	}
	if s.MinQueryLength == 0 {
		s.MinQueryLength = 80
	}
	if s.Candidates == 0 {
		s.Candidates = 20
	}
	if s.ContextMsgs == 0 {
		s.ContextMsgs = 3
	}
	if s.MaxRecallBytes == 0 {
		s.MaxRecallBytes = 6144
	}
	return s
}

// Corpus files belong to the Agent, not to cwd or another application's home.
// Hash the bytes already read so human edits (even equal-size edits) invalidate
// indexes. Only changed corpora incur tokenization; there is no watcher lifetime.
func (r *recaller) search(query []string, floor float64) ([]hit, error) {
	if pending := r.parent.History().Context().Compression; pending != nil && pending.Committed != nil {
		return nil, fmt.Errorf("memory commit requires disk recovery before recall")
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	root := r.parent.Config().DataDir
	// Retired files remain searchable after graduation removed their original
	// bytes from active memory. Bands and archives share the memory quota.
	sources := []struct{ name, path string }{
		{"memory", filepath.Join(root, "memory")},
		{"memory", filepath.Join(root, "bucket-0")},
		{"memory", filepath.Join(root, "bucket-1")},
		{"memory", filepath.Join(root, "MEMORY.md")},
		{"memory", filepath.Join(root, "SOUL.md")},
		{"docs", filepath.Join(root, "docs")},
		{"handoffs", filepath.Join(root, "handoffs")},
		{"learnings", filepath.Join(root, "learnings")},
	}
	// Reach skill identity through its owner, rather than saving another copy
	// of the application's directory setting in recall. Discovery grants no tools.
	if skills := r.parent.Skills(); skills != nil {
		sources = append(sources, struct{ name, path string }{"skills", skills.Directory()})
	}
	corpus := map[string][]chunk{}
	hash := sha256.New()
	// Keep the file texts until we know whether rebuilding is necessary.
	type document struct{ source, file, text string }
	var documents []document
	for _, source := range sources {
		walk := r.parent.Sandbox().WalkDir
		read := r.parent.Sandbox().ReadFile
		// Application-installed skills may intentionally live outside the workspace.
		// That trusted corpus is distinct from model-writable memory and handoffs.
		if source.name == "skills" {
			if r.parent.Spec().SandboxRoot == "" {
				walk, read = filepath.WalkDir, os.ReadFile
			} else if _, err := r.parent.Sandbox().Resolve(source.path); err != nil {
				// An explicitly selected external skill directory is trusted input, but
				// even there a skill file may not redirect recall outside its own corpus.
				root, err := os.OpenRoot(source.path)
				if err != nil {
					return nil, err
				}
				defer root.Close()
				read = func(path string) ([]byte, error) {
					rel, err := filepath.Rel(source.path, path)
					if err != nil {
						return nil, err
					}
					return root.ReadFile(rel)
				}
				walk = func(path string, visit fs.WalkDirFunc) error {
					return fs.WalkDir(root.FS(), ".", func(name string, d fs.DirEntry, err error) error { return visit(filepath.Join(path, name), d, err) })
				}
			}
		}
		err := walk(source.path, func(path string, entry fs.DirEntry, err error) error {
			// Optional corpus directories need not exist on a fresh Agent.
			// Other I/O failures remain diagnostic errors, not empty successes.
			if os.IsNotExist(err) {
				return nil
			}
			if err != nil {
				return err
			}
			if entry.IsDir() {
				return nil
			}
			if filepath.Ext(path) != ".md" && !(source.name == "memory" && strings.HasSuffix(path, ".md.retired")) || source.name == "skills" && entry.Name() != "SKILL.md" {
				return nil
			}
			if !entry.Type().IsRegular() {
				return nil
			}
			if source.name == "skills" && !r.parent.Spec().EnableWebSearch && filepath.Base(filepath.Dir(path)) == "web-search" {
				return nil
			}
			body, err := read(path)
			if err != nil {
				return err
			}
			// Skill identity can be absolute while the Agent's data root is relative.
			// Compute attribution against the source's actual owner before Rel.
			base := root
			if source.name == "skills" {
				base = source.path
			}
			file, err := filepath.Rel(base, path)
			if err != nil {
				return err
			}
			if source.name == "skills" {
				file = filepath.Join("skills", file)
			}
			// Length-delimited content and paths distinguish renames/deletions
			// as well as edits; mtimes alone can miss equal-size human changes.
			fmt.Fprintf(hash, "%s\x00%d\x00%s\x00", file, len(body), body)
			documents = append(documents, document{source.name, file, string(body)})
			return nil
		})
		if err != nil {
			return nil, err
		}
	}
	var digest [32]byte
	copy(digest[:], hash.Sum(nil))
	// Publish a complete replacement under the lock. No search observes half
	// an index, and removed files cannot linger in the next request's results.
	if r.indexes == nil || digest != r.digest {
		for _, doc := range documents {
			corpus[doc.source] = append(corpus[doc.source], chunks(doc.file, doc.text)...)
		}
		r.indexes = map[string]index{}
		for source, docs := range corpus {
			r.indexes[source] = build(docs)
		}
		r.digest = digest
	}
	var hits []hit
	// Fixed order makes equal scores deterministic without filesystem/map lottery.
	for _, source := range []string{"memory", "docs", "handoffs", "learnings", "skills"} {
		hits = append(hits, r.indexes[source].search(source, query, floor)...)
	}
	rank(hits)
	return hits, nil
}

// Recall returns immutable attachment bytes. The caller passes a snapshot so
// the stateless judge can run off the actor without borrowing live conversation.
func (r *recaller) Recall(ctx context.Context, query string, conversation *common.Context) (common.RecallData, error) {
	s := settings(r.parent.Config().Recall)
	if s.Disabled || utf8.RuneCountInString(query) < s.MinQueryLength {
		return common.RecallData{}, nil
	}
	// The judge gets a broad candidate net; plain selection needs a stricter
	// keyword floor because there is no semantic filter to remove false hits.
	floor := s.MinScore
	if !s.DisableJudge {
		floor = 0.5
	}
	hits, err := r.search(queryTokens(query), floor)
	if err != nil {
		return common.RecallData{}, err
	}
	// Include text nested in tool results and blockquoted prior attachments.
	// Normalizing quote prefixes permits comparison with the original snippet.
	var contextText strings.Builder
	contextText.WriteString(r.parent.Engine().SystemPrompt())
	var collect func([]common.Part)
	collect = func(parts []common.Part) {
		for _, p := range parts {
			contextText.WriteByte('\n')
			contextText.WriteString(p.Text)
			collect(p.Parts)
		}
	}
	for _, e := range conversation.Dialogue {
		collect(e.Parts)
	}
	collect(conversation.Ephemera)
	present := unquote(contextText.String())
	// A byte-capped attachment may contain only the prefix of a source snippet.
	// Retain its identity with that entry instead of requiring the unseen suffix
	// to occur in context. Hashing the complete content still permits human edits.
	seen := map[string]bool{}
	for _, entry := range conversation.Dialogue {
		if entry.Kind == common.KindRecall {
			for _, hash := range entry.RecallHashes {
				seen[hash] = true
			}
		}
	}
	available := hits[:0]
	for _, h := range hits {
		if !seen[snippetHash(h)] && !strings.Contains(present, unquote(h.content)) {
			available = append(available, h)
		}
	}
	hits = available
	if len(hits) == 0 {
		return common.RecallData{}, nil
	}
	// Keep the full hit list for mechanical fallback. Capping candidates must
	// not discard low-scoring memories before the final quota can protect them.
	var chosen []hit
	judged := false
	if !s.DisableJudge {
		candidates := hits[:min(len(hits), s.Candidates)]
		chosen, err = r.judge(ctx, query, conversation, candidates, s)
		// Empty is a valid verdict. Only failure re-enters mechanical selection;
		// otherwise an intentional [] would inject the very hits the judge rejected.
		judged = err == nil
		if err != nil {
			r.parent.Ensemble().Logf("recall judge failed; using source quotas: %v", err)
		}
	}
	if !judged {
		chosen = quotas(hits, s.MaxSnippets)
	}
	text, hashes := format(chosen, s.MaxRecallBytes)
	if text == "" {
		return common.RecallData{}, nil
	}
	return common.RecallData{Parts: []common.Part{{Type: "text", Text: text}}, Hashes: hashes}, nil
}
func unquote(s string) string {
	lines := strings.Split(s, "\n")
	for i, line := range lines {
		lines[i] = strings.TrimPrefix(line, "> ")
	}
	return strings.TrimSpace(strings.Join(lines, "\n"))
}

// Quotas apply to the FINAL mechanical selection. A successful judge never
// comes through here, including a valid empty selection. Spare slots flow to
// the best remaining scores, so a sparse source does not waste the budget.
func quotas(hits []hit, limit int) []hit {
	quota := map[string]int{"memory": (limit + 1) / 2}
	var sources []string
	for _, name := range []string{"docs", "handoffs", "learnings", "skills"} {
		for _, h := range hits {
			if h.source == name {
				sources = append(sources, name)
				break
			}
		}
	}
	// Divide only among secondary sources with actual hits. Fixed source order
	// gives the remainder a deterministic home when slots do not divide evenly.
	remaining := limit - quota["memory"]
	for i, name := range sources {
		quota[name] = remaining / len(sources)
		if i < remaining%len(sources) {
			quota[name]++
		}
	}
	// Two passes reserve source shares first, then spend every spare slot on
	// remaining score-ranked hits. Track positions so no hit is selected twice.
	selected := make([]bool, len(hits))
	var out []hit
	for i, h := range hits {
		if quota[h.source] > 0 {
			out = append(out, h)
			selected[i] = true
			quota[h.source]--
		}
	}
	for i, h := range hits {
		if len(out) < limit && !selected[i] {
			out = append(out, h)
		}
	}
	return out
}

// Count the complete wire-visible text, including heading, attribution and
// quote prefixes. Cutting source content first would undercount that overhead.
func format(hits []hit, limit int) (string, []string) {
	var hashes []string
	const title = "[Auto-recalled memories]\n"
	out := title
	for _, h := range hits {
		prefix := fmt.Sprintf("\nFrom %s — %s:\n> ", h.file, h.header)
		room := limit - len(out) - len(prefix)
		if room <= 0 {
			break
		}
		body := strings.ReplaceAll(h.content, "\n", "\n> ")
		if len(body) > room {
			body = body[:room]
			// The cap is bytes, but its boundary must not split a UTF-8 character.
			for !utf8.ValidString(body) {
				body = body[:len(body)-1]
			}
		}
		out += prefix + body
		// A selected snippet that did not fit has no identity in live context.
		hashes = append(hashes, snippetHash(h))
	}
	if out == title {
		return "", nil
	}
	return out, hashes
}

func snippetHash(h hit) string { return fmt.Sprintf("%x", sha256.Sum256([]byte(h.content))) }
