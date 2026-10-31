package gradeedition2

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/waywardgeek/ensemble/internal/fakevendor"
	"github.com/waywardgeek/ensemble/internal/grade"
)

// Chapter23 preserves original17's ten weights without adding recursive parity.
// core identifies the submitted library; dataRoot selects only the grader's
// corpus location under a scratch cwd, not a new application launch option.
func Chapter23(bin, core, gui, dataRoot string) (grade.Report, error) {
	r := recallChecks(bin, dataRoot, true)
	r.set("recall-is-a-spoke", recallTopology(core, gui))
	checks := grade.Ch17Checks(grade.Ch17Result{Errs: r.Errs, Fatal: r.Fatal})
	return grade.NewTitledReport("Edition 2, Chapter 23 — Auto-Recall", checks, ""), nil
}

const recallHeader = "[Auto-recalled memories]"
const recallQuery = "Please explain the quartz zephyr migration decision and the rejected deployment approach from our archived project discussion."
const recallFact = "The quartz zephyr migration selected copper ferry because the silver bridge lost ordering."

// Scratch corpus paths are explicit. Accepted GUI owns cr; the historical GUI
// owns cwd. A missing source must not be hidden by writing duplicate fixtures.
type recallSuite struct {
	bin, dir, skills, gui, data string
	next                        int
	launch                      func(recallOpts) recallOut
}

func recallChecks(bin, dataRoot string, timeout bool) contextResult {
	r := contextResult{Errs: map[string]string{}}
	for _, c := range grade.Ch17Checks(grade.Ch17Result{}) {
		r.set(c.ID, "")
	}
	dir, err := os.MkdirTemp("", "recall-grade-")
	if err != nil {
		r.Fatal = err.Error()
		return r
	}
	defer os.RemoveAll(dir)
	s := recallSuite{bin: bin, dir: dir, skills: filepath.Join(dir, "skills"), gui: filepath.Join(dir, "web"), data: dataRoot}
	if err := recallWrite(s.skills, map[string]string{"base/SKILL.md": "---\nname: base\ndescription: Recall fixture\ntype: primary\ntools: read_file micro_handoff\n---\nYou are a helpful agent.\n"}); err != nil {
		r.Fatal = err.Error()
		return r
	}
	os.MkdirAll(s.gui, 0755)
	recallIndexing(&s, &r)
	recallCompaction(&s, &r)
	recallRanking(&s, &r)
	recallChunking(&s, &r)
	recallQuotas(&s, &r)
	recallJudging(&s, &r)
	recallFallback(&s, &r, timeout)
	recallLifetime(&s, &r)
	recallCap(&s, &r)
	return r
}
func (s *recallSuite) run(files map[string]string, o recallOpts) recallOut {
	if o.dir == "" {
		s.next++
		o.dir = filepath.Join(s.dir, fmt.Sprint(s.next))
	}
	if err := os.MkdirAll(o.dir, 0755); err != nil {
		return recallOut{fatal: err.Error()}
	}
	if err := recallWrite(filepath.Join(o.dir, s.data), files); err != nil {
		return recallOut{fatal: err.Error()}
	}
	if o.model == "" {
		o.model = "claude-sonnet-5"
	}
	if o.prompts == nil {
		o.prompts = []string{recallQuery}
	}
	if o.judge == "" {
		o.judge = "unreadable judgment"
	}
	if s.launch != nil {
		return s.launch(o)
	}
	return recallLaunch(s.bin, s.skills, s.gui, o)
}
func recallWrite(dir string, files map[string]string) error {
	for name, text := range files {
		path := filepath.Join(dir, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			return err
		}
		if err := os.WriteFile(path, []byte(text), 0644); err != nil {
			return err
		}
	}
	return nil
}
func recallOK(r *contextResult, id string, out recallOut) bool {
	if out.fatal != "" {
		r.fail(id, "session: %s", out.fatal)
		return false
	}
	if len(recallTurns(out.reqs)) == 0 {
		r.fail(id, "no actual main request")
		return false
	}
	return true
}
func recallLast(out recallOut) string {
	turns := recallTurns(out.reqs)
	if len(turns) == 0 {
		return ""
	}
	return recallWire(turns[len(turns)-1].Body)
}
func recallPadding() string {
	var b strings.Builder
	for i := 0; i < 25; i++ {
		fmt.Fprintf(&b, "## Garden %d\nTomato seedlings prefer warm soil and steady watering.\n\n", i)
	}
	return b.String()
}
func recallDocs(bodies ...string) map[string]string {
	files := map[string]string{"docs/padding.md": recallPadding()}
	for i, b := range bodies {
		files[fmt.Sprintf("docs/%02d.md", i)] = "## Record\n" + b
	}
	return files
}
func recallLongQuery(terms string) string {
	return terms + strings.Repeat(" the and we were with our ", 5)
}

// Isolate discovery from competition: three sources, three available slots.
// Memory/newest and skill discovery get separate controls with their own roots.
func recallIndexing(s *recallSuite, r *contextResult) {
	const id = "bm25-indexes"
	files := map[string]string{}
	for i, source := range []string{"docs", "handoffs", "learnings"} {
		files[source+"/padding.md"] = recallPadding()
		files[source+"/target.md"] = fmt.Sprintf("## Decision\n%s indexmark%d", recallFact, i)
	}
	out := s.run(files, recallOpts{})
	if recallOK(r, id, out) {
		text := recallLast(out)
		for i := 0; i < 3; i++ {
			if !strings.Contains(text, fmt.Sprintf("indexmark%d", i)) {
				r.fail(id, "source %d missing from actual recall", i)
			}
		}
	}
	// Archive names avoid the date-number pattern loaded into live session bands.
	// Their appearance must come from recall in both applications, not startup.
	files = map[string]string{"memory/memory/zy-older.md": "## Memory\n" + recallFact + " oldestmark", "memory/memory/zz-newest.md": "## Memory\n" + recallFact + " newestmark", "memory/memory/padding.md": recallPadding()}
	out = s.run(files, recallOpts{})
	if recallOK(r, id, out) && (!strings.Contains(recallLast(out), "newestmark") || !strings.Contains(recallLast(out), "oldestmark")) {
		r.fail(id, "all-memory corpus omitted a newest/older matching file")
	}
	path := filepath.Join(s.skills, "discovery", "SKILL.md")
	if err := recallWrite(s.skills, map[string]string{"discovery/SKILL.md": "---\nname: discovery\ndescription: Searchable archive example\ntype: loadable\n---\n" + recallFact + " skillmark"}); err != nil {
		r.fail(id, "skill fixture: %v", err)
		return
	}
	out = s.run(nil, recallOpts{})
	os.Remove(path)
	if recallOK(r, id, out) && !strings.Contains(recallLast(out), "skillmark") {
		r.fail(id, "available skill document was not indexed")
	}
	// Edit after a completed turn in the same process; restarting would hide a
	// stale in-memory index. The replacement changes only the controlled marker.
	dir := filepath.Join(s.dir, "refresh")
	files = recallDocs(recallFact + " beforeedit")
	out = s.run(files, recallOpts{dir: dir, prompts: []string{recallQuery, recallQuery}, after: func(i int) error {
		if i == 0 {
			return recallWrite(filepath.Join(dir, s.data), recallDocs(recallFact+" afteredit"))
		}
		return nil
	}})
	if recallOK(r, id, out) && !strings.Contains(recallLast(out), "afteredit") {
		r.fail(id, "same-process corpus edit did not reach later recall")
	}
}

// Compare competing hits, not exact scores. Query terms occur in bodies so
// either choice about indexing heading text remains valid. No ties are graded.
func recallRanking(s *recallSuite, r *contextResult) {
	cases := []struct{ name, query, padding, want, other, a, b string }{
		{"rarity", "rarezed commonword", "commonword ordinary", "rarebest", "commonother", "commonother " + strings.Repeat("commonword ", 10), "rarebest rarezed"},
		{"saturation", "alpha beta gamma", strings.Repeat("garden ", 100), "diversebest", "repeatother", "repeatother " + strings.Repeat("alpha ", 100), "diversebest alpha beta gamma " + strings.Repeat("garden ", 97)},
		{"length", "alpha", strings.Repeat("garden ", 30), "focusedbest", "longother", "longother alpha " + strings.Repeat("garden ", 1000), "focusedbest alpha decision"},
		{"document-stop-words", "alpha", strings.Repeat("garden ", 30), "shortbest", "stopother", "stopother alpha " + strings.Repeat("the and we ", 300), "shortbest alpha decision"},
	}
	for _, tc := range cases {
		id := "bm25-scores"
		if tc.name == "document-stop-words" {
			id = "bm25-stop-words"
		}
		files := recallDocs(tc.a, tc.b)
		var pad strings.Builder
		for i := 0; i < 25; i++ {
			fmt.Fprintf(&pad, "## Padding %d\n%s\n\n", i, tc.padding)
		}
		files["docs/padding.md"] = pad.String()
		out := s.run(files, recallOpts{prompts: []string{recallLongQuery(tc.query)}})
		if !recallOK(r, id, out) {
			continue
		}
		text := recallLast(out)
		winner, other := strings.Index(text, tc.want), strings.Index(text, tc.other)
		if winner < 0 || other >= 0 && other < winner {
			r.fail(id, "%s ranking selected wrong competing hit: %s", tc.name, text)
		}
	}
	files := recallDocs("stoptrap "+strings.Repeat("the and we were with our ", 8), recallFact+" contentmatch")
	out := s.run(files, recallOpts{})
	if recallOK(r, "bm25-stop-words", out) {
		text := recallLast(out)
		if !strings.Contains(text, "contentmatch") || strings.Contains(text, "stoptrap") {
			r.fail("bm25-stop-words", "query stop-word trap recalled or real match absent: %s", text)
		}
	}
}

func recallChunking(s *recallSuite, r *contextResult) {
	const id = "chunking-splits"
	garden := strings.Repeat("Tomato soil watering garden. ", 90)
	files := recallDocs("## First\nwrongsection " + garden + "\n## Parent\n" + garden + "\n### Relevant\n" + recallFact + " rightsection\n\n" + garden + "\n\nwrongparagraph " + garden + "\n### Other\nwrongchild " + garden)
	out := s.run(files, recallOpts{})
	if !recallOK(r, id, out) {
		return
	}
	text := recallLast(out)
	if !strings.Contains(text, "rightsection") || strings.Contains(text, "wrongsection") || strings.Contains(text, "wrongchild") || strings.Contains(text, "wrongparagraph") || !strings.Contains(text, "Parent > Relevant") {
		r.fail(id, "section/subsection separation or breadcrumb missing: %s", text)
	}
}

// A fallback's final three slots must include low-scoring memories even when
// many documents precede them in the judge's bounded candidate list.
func recallQuotas(s *recallSuite, r *contextResult) {
	const id = "per-source-quota"
	files := map[string]string{"docs/padding.md": recallPadding(), "memory/memory/padding.md": recallPadding()}
	for i := 0; i < 24; i++ {
		files[fmt.Sprintf("docs/match%02d.md", i)] = fmt.Sprintf("## Decision\n%s docquota%d", strings.Repeat("quartz zephyr migration ", 8), i)
	}
	for i := 0; i < 4; i++ {
		files[fmt.Sprintf("memory/memory/archive%02d.md", i+1)] = fmt.Sprintf("## Decision\nmemquota%d quartz zephyr migration %s", i, strings.Repeat("garden soil water ", 80))
	}
	out := s.run(files, recallOpts{})
	if recallOK(r, id, out) {
		text := recallLast(out)
		if strings.Count(text, "memquota") != 2 || strings.Count(text, "docquota") != 1 {
			r.fail(id, "final fallback quota did not reserve two memories and one document: %s", text)
		}
	}
	out = s.run(files, recallOpts{judge: "[0,1,2]"})
	if recallOK(r, id, out) {
		text := recallLast(out)
		if strings.Count(text, "docquota") != 3 || strings.Contains(text, "memquota") {
			r.fail(id, "successful judge choices were overridden by quota: %s", text)
		}
	}
	out = s.run(recallDocs(recallFact+" sparemark0", recallFact+" sparemark1", recallFact+" sparemark2"), recallOpts{})
	if recallOK(r, id, out) && strings.Count(recallLast(out), "sparemark") != 3 {
		r.fail(id, "unused memory slots did not flow to matching documents")
	}

}

func recallJudging(s *recallSuite, r *contextResult) {
	const id = "judge-filters"
	files := recallDocs("rankfirst "+strings.Repeat("quartz zephyr migration ", 6), "ranksecond quartz zephyr migration "+strings.Repeat("garden ", 100))
	for _, reply := range []string{"[1]", "Here: [\"1\"] done", "[\"1\",1,999]", "[]"} {
		out := s.run(files, recallOpts{judge: reply})
		if !recallOK(r, id, out) {
			continue
		}
		judges := 0
		for _, req := range out.reqs {
			if recallJudge(req.Body) {
				judges++
				text := memoryConversation(req.Body)
				if !strings.Contains(text, "rankfirst") || !strings.Contains(text, "ranksecond") {
					r.fail(id, "judge never received both competing candidates")
				}
			}
		}
		if judges != 1 {
			r.fail(id, "expected one actual stateless judge request, got %d", judges)
		}
		text := recallLast(out)
		if reply == "[]" {
			if text != "" {
				r.fail(id, "valid empty judgment attached recall")
			}
		} else if !strings.Contains(text, "ranksecond") || strings.Contains(text, "rankfirst") || strings.Count(text, "ranksecond") != 1 {
			r.fail(id, "judge selection %s not reflected in actual recall: %s", reply, text)
		}
	}
}

func recallFallback(s *recallSuite, r *contextResult, timeout bool) {
	const id = "judge-fallback"
	for _, mode := range []string{"malformed", "HTTP error", "timeout"} {
		if mode == "timeout" && !timeout {
			continue
		}
		start := time.Now()
		out := s.run(recallDocs(recallFact+" fallbackmark"), recallOpts{judgeError: mode == "HTTP error", judgeHang: mode == "timeout"})
		if recallOK(r, id, out) && !strings.Contains(recallLast(out), "fallbackmark") {
			r.fail(id, "%s lost mechanical fallback", mode)
		}
		// The HTTP handler remains held until process cleanup. Main completion is
		// therefore actual timeout fallback, not a late successful judge response.
		if mode == "timeout" && time.Since(start) > 22*time.Second {
			r.fail(id, "held judge did not yield within its 15-second budget plus launch allowance")
		}
	}
	out := s.run(nil, recallOpts{})
	if recallOK(r, id, out) && recallLast(out) != "" {
		r.fail(id, "empty archive attached an empty recall block")
	}
	out = s.run(recallDocs(recallFact+" shortquerymark"), recallOpts{prompts: []string{"quartz zephyr migration"}})
	if recallOK(r, id, out) && (recallLast(out) != "" || len(out.reqs) != 1) {
		r.fail(id, "short matching query did not bypass retrieval")
	}
}

func recallLifetime(s *recallSuite, r *contextResult) {
	const id = "recall-is-own-kind"
	dir := filepath.Join(s.dir, "lifetime")
	out := s.run(recallDocs(recallFact+" lifetimemark"), recallOpts{dir: dir, prompts: []string{recallQuery, recallQuery}})
	if !recallOK(r, id, out) {
		return
	}
	attachments := recallAttachments(out)
	turns := recallTurns(out.reqs)
	if len(attachments) != 1 || len(turns) != 2 {
		r.fail(id, "two turns did not retain exactly one recorded attachment")
		return
	}
	text := attachments[0]
	if text == "" || strings.Contains(text, recallQuery) || !strings.Contains(text, "lifetimemark") {
		r.fail(id, "recall event lost bytes or fused human input")
	}
	for _, turn := range turns {
		wire := memoryConversation(turn.Body)
		if !strings.Contains(wire, text) || strings.Count(wire, "lifetimemark") != 1 || !strings.Contains(wire, recallQuery) {
			r.fail(id, "recorded recall/human words not retained distinctly on actual wire")
		}
	}
	// Startup with no prompts must not ask a judge. A later short prompt reads
	// retained bytes even after the source file disappears, proving persistence.
	os.RemoveAll(filepath.Join(dir, s.data, "docs"))
	restored := s.run(nil, recallOpts{dir: dir, prompts: []string{}})
	if restored.fatal != "" || len(restored.reqs) != 0 {
		r.fail(id, "startup contacted provider or failed: %s", restored.fatal)
	}
	restored = s.run(nil, recallOpts{dir: dir, prompts: []string{"Continue."}})
	if recallOK(r, id, restored) && (!strings.Contains(recallLast(restored), text) || len(restored.reqs) != 1) {
		r.fail(id, "recorded bytes absent after restart without source")
	}
	// A real checkpoint removes recall from its next request, without inventing
	// a special recall-removal API or rewriting the human message.
	cut := s.run(nil, recallOpts{dir: dir, prompts: []string{"Checkpoint now."}, expect: []int{2}, replies: []fakevendor.Reply{{ToolName: "micro_handoff", ToolID: "recall-cut", ToolArgs: `{"text":"Continue after checkpoint."}`}, {Text: "Checkpoint complete."}}})
	if recallOK(r, id, cut) && strings.Contains(recallLast(cut), "lifetimemark") {
		r.fail(id, "checkpoint retained recalled attachment")
	}
}

func recallCap(s *recallSuite, r *contextResult) {
	const id = "injection-capped"
	// An indivisible paragraph may exceed chunk size. The formatted attachment,
	// including attribution, still has a hard cap and must not repeat next turn.
	files := recallDocs("capmark " + strings.Repeat("quartz zephyr migration archive fact. ", 350))
	out := s.run(files, recallOpts{judge: "[0]", prompts: []string{recallQuery, recallQuery}})
	if !recallOK(r, id, out) {
		return
	}
	attachments := recallAttachments(out)
	if len(attachments) != 1 {
		r.fail(id, "capped source attached %d times across two turns", len(attachments))
		return
	}
	text := attachments[0]
	if len(text) > 6144 || !strings.Contains(text, "capmark") {
		r.fail(id, "formatted cap/content invalid: %d bytes", len(text))
	}
	turns := recallTurns(out.reqs)
	if len(turns) != 2 || strings.Count(recallWire(turns[1].Body), "capmark") != 1 {
		r.fail(id, "actual second request duplicated or lost capped text")
	}
}

// Resolve the actual built GUI dependency graph. A healthy unused package next
// to the executable cannot satisfy the presence check; composition roots outside
// internal may intentionally construct spokes. Parent semantics remain review.
func recallTopology(core, gui string) string {
	base := grade.DiscoverBase(core)
	if base == "" {
		return "core module is missing"
	}
	packages, err := listPackages(gui, "-deps", "./cmd")
	if err != nil {
		return err.Error()
	}
	all := map[string]listedPackage{}
	for _, p := range packages {
		all[p.ImportPath] = p
	}
	target, hub := base+"/internal/recall", base+"/internal/common"
	recall, ok := all[target]
	if !ok {
		return "built GUI does not depend on submitted recall spoke"
	}
	wanted, err := filepath.Abs(core)
	if err != nil {
		return err.Error()
	}
	wanted, _ = filepath.EvalSymlinks(wanted)
	if recall.Module == nil {
		return "recall has no resolved module"
	}
	actual, _ := filepath.EvalSymlinks(recall.Module.Dir)
	if actual != wanted {
		return "GUI resolves recall from another core tree"
	}
	for _, imp := range recall.Imports {
		if imp != hub && !all[imp].Standard {
			return fmt.Sprintf("recall imports non-hub dependency %s", imp)
		}
	}
	for _, p := range packages {
		if p.ImportPath == target || !strings.HasPrefix(p.ImportPath, base+"/internal/") {
			continue
		}
		for _, imp := range p.Imports {
			if imp == target {
				return fmt.Sprintf("implementation package %s imports recall directly", p.ImportPath)
			}
		}
	}
	return ""
}
