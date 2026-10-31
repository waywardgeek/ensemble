package gradeedition2

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/waywardgeek/ensemble/internal/fakevendor"
)

// Create the corpus through an actual successful or failed graduation, then
// restart with recall enabled. Handwritten final files alone cannot establish
// the interaction between retained outcomes, live memory and archived originals.
func recallCompaction(s *recallSuite, r *contextResult) {
	const indexID, lifetimeID = "bm25-indexes", "recall-is-own-kind"
	const replacement = "I retained the quartz zephyr migration decision because ordering matters. FINALIZED_CORPUS_16"
	for _, failed := range []bool{false, true} {
		dir := filepath.Join(s.dir, fmt.Sprintf("compaction-%t", failed))
		files := map[string]string{
			"memory/padding.md.retired": recallPadding(),
			"docs/padding.md":           recallPadding(),
			"docs/decision.md":          "## Decision\n" + recallFact + " CORPUS_DOC_16",
		}
		for i := 1; i <= 3; i++ {
			files[fmt.Sprintf("memory/2026-10-01-%d.md", i)] = fmt.Sprintf("## Decision\n%s CORPUS_ORIGINAL_%d %s", recallFact, i, strings.Repeat("garden soil ", 8))
		}
		if err := recallWrite(filepath.Join(dir, s.data), files); err != nil {
			r.fail(indexID, "compaction corpus: %v", err)
			r.fail(lifetimeID, "compaction corpus: %v", err)
			continue
		}
		budget := len(files["memory/2026-10-01-3.md"]) + 10
		settings := memorySettings(budget)
		object(settings["memory"])["conversation"] = map[string]any{"budget": 100000}
		prepared := memoryLaunch(s.bin, s.skills, s.gui, memoryOpts{
			dir: dir, model: "claude-sonnet-5", settings: settings,
			prompts: []string{"Checkpoint now."}, expect: []int{2}, replies: memoryTalk(1),
			route: func(body []byte) *fakevendor.Reply {
				if !memoryCompressor(body) {
					return nil
				}
				text := replacement
				if failed {
					text = " "
				}
				return &fakevendor.Reply{ToolName: "submit", ToolID: "recall-corpus-memory", ToolArgs: memoryJSON(text), Usage: fakevendor.Canonical{Input: 100, Output: 20}}
			},
		})
		if prepared.fatal != "" || len(prepared.reqs) != 3 {
			r.fail(indexID, "real graduation failed=%t: %s (%d requests)", failed, prepared.fatal, len(prepared.reqs))
			r.fail(lifetimeID, "real graduation preparation did not complete")
			continue
		}
		// The third source is an active-memory exclusion control. Only the first
		// two retire after success; an empty candidate must retire neither.
		valid := true
		for i := 1; i <= 3; i++ {
			name := fmt.Sprintf("memory/2026-10-01-%d.md", i)
			actual := name
			if !failed && i < 3 {
				actual += ".retired"
			}
			data, err := os.ReadFile(filepath.Join(dir, s.data, actual))
			if err != nil || string(data) != files[name] {
				r.fail(indexID, "graduation failed=%t did not preserve expected source %s: %v", failed, actual, err)
				valid = false
			}
		}
		if !valid {
			r.fail(lifetimeID, "graduation source preparation failed")
			continue
		}
		out := s.run(nil, recallOpts{dir: dir, prompts: []string{recallQuery, recallQuery}})
		if !recallOK(r, indexID, out) {
			r.fail(lifetimeID, "post-compaction recall did not complete")
			continue
		}
		turns := recallTurns(out.reqs)
		if len(turns) != 2 || len(out.reqs) != 3 {
			r.fail(indexID, "post-compaction requests must be one judge and two main, got %d total/%d main", len(out.reqs), len(turns))
			r.fail(lifetimeID, "post-compaction attachment sequence did not complete")
			if len(turns) != 2 {
				continue
			}
		}
		recalled := recallLast(out)
		if !strings.Contains(recalled, "CORPUS_DOC_16") || strings.Contains(recalled, "CORPUS_ORIGINAL_3") || strings.Contains(recalled, "FINALIZED_CORPUS_16") {
			r.fail(indexID, "post-compaction recall lost archive control or duplicated active memory")
		}
		for i := 1; i <= 2; i++ {
			marker := fmt.Sprintf("CORPUS_ORIGINAL_%d", i)
			if strings.Contains(recalled, marker) == failed {
				r.fail(indexID, "failed=%t retired-source recall inclusion is wrong for %s", failed, marker)
			}
		}
		for _, turn := range turns {
			body := memoryConversation(turn.Body)
			if strings.Contains(body, replacement) == failed {
				r.fail(indexID, "failed=%t restart has wrong finalized replacement", failed)
			}
			for i := 1; i <= 3; i++ {
				if n := strings.Count(body, fmt.Sprintf("CORPUS_ORIGINAL_%d", i)); n != 1 {
					r.fail(indexID, "failed=%t source %d appears %d times in actual context", failed, i, n)
				}
			}
		}
		attachments := recallAttachments(out)
		if len(attachments) != 1 || attachments[0] == "" || strings.Contains(attachments[0], recallQuery) {
			r.fail(lifetimeID, "post-compaction recall did not retain one separate recorded attachment")
			continue
		}
		// Remove the recalled files, not their recorded attachment. A short
		// restored prompt must render those exact bytes without new retrieval.
		removed := true
		paths := []string{"docs/decision.md"}
		if !failed {
			for i := 1; i <= 2; i++ {
				paths = append(paths, fmt.Sprintf("memory/2026-10-01-%d.md.retired", i))
			}
		}
		for _, path := range paths {
			if err := os.Remove(filepath.Join(dir, s.data, path)); err != nil {
				r.fail(lifetimeID, "remove recalled source before replay: %v", err)
				removed = false
			}
		}
		if !removed {
			continue
		}
		restored := s.run(nil, recallOpts{dir: dir, prompts: []string{"Continue."}})
		if recallOK(r, lifetimeID, restored) && (len(restored.reqs) != 1 || !strings.Contains(memoryConversation(restored.reqs[0].Body), attachments[0])) {
			r.fail(lifetimeID, "post-compaction recall bytes did not survive restart without their source files")
		}
	}
}
