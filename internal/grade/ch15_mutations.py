#!/usr/bin/env python3
"""P9 mutation audit for the chapter 15 grader.

Each mutant deletes exactly one protected behaviour from a scratch copy
of a solution snapshot (never the snapshot itself), grades the copy, and
prints the exact set of failing checks. A mutant whose patch anchor is
missing is reported as ANCHOR-MISSING (the reference moved; fix the
mutant); one that does not compile is BUILD-FAIL and is not scored.

Usage: python3 internal/grade/ch15_mutations.py [grade-binary] [mutant ...]
The grade binary defaults to /tmp/grade15 (go build -o /tmp/grade15 ./cmd/grade).
"""
import os, re, shutil, subprocess, sys, tempfile

ROOT = os.path.dirname(os.path.dirname(os.path.dirname(os.path.abspath(__file__))))

# (file, old, new, nth): nth is 1-based among occurrences of old; 0 means
# old must occur exactly once. ("re", pattern, repl, count) is a regex sub
# that must make exactly count replacements.
SCRATCH = '''
var scratchCuts []common.RedactData

// mutant: the ladder's cuts are never recorded; they are recomputed from
// the uncut context and the current settings on every render.
func (e *Engine) scratch(ev common.Event) error {
	scratchCuts = append(scratchCuts, *ev.Redact)
	return nil
}

func (e *Engine) view() *common.Context {
	b, _ := json.Marshal(e.Ctx)
	c := common.NewContext()
	json.Unmarshal(b, c)
	for i := range scratchCuts {
		c.Apply(common.Event{Seq: 1 << 40, Type: common.Redacted, Redact: &scratchCuts[i]})
	}
	scratchCuts = nil
	return c
}
'''

MUTANTS = {
    "M1-skill-body-in-result": ("ch15", "skill body left in the load_skill tool result", [
        ("internal/tools/tools.go", 'result := fmt.Sprintf("Loaded skill %q.',
         'result := body + fmt.Sprintf("Loaded skill %q.', 0)]),
    "M2-span-eats-survivors": ("ch15", "span walk does not skip survivors", [
        ("internal/common/context.go", "return e.Kind == KindDialogue && e.Seq >= r.From && e.Seq <= r.To",
         "return e.Seq >= r.From && e.Seq <= r.To", 0)]),
    # M2 is equivalent under the levels ch15 emits: RedactResult and
    # RedactTool touch only tool parts, which rule 3 keeps out of
    # survivors. M2b is compound on purpose (two deletions) to show the
    # grader does see a lost survivor once a level can reach its text.
    "M2b-span-eats-survivor-text": ("ch15", "COMPOUND: M2 plus RedactTool also dropping text", [
        ("internal/common/context.go", "return e.Kind == KindDialogue && e.Seq >= r.From && e.Seq <= r.To",
         "return e.Seq >= r.From && e.Seq <= r.To", 0),
        ("internal/common/context.go", "case ToolCallPart, ToolResultPart:", "case ToolCallPart, ToolResultPart, TextPart:", 1)]),
    "M3-handoff-orphans-results": ("ch15", "handoff clears calls but leaves their results", [
        ("internal/common/context.go", "case ToolCallPart, ToolResultPart:", "case ToolCallPart:", 2)]),
    "M4-stub-ignores-keep": ("ch15", "per-round-trip stubbing ignores keep_tool_results", [
        ("internal/llm/policy.go", "ok && c.Name == common.KeepToolResults {", 'ok && c.Name == "never-kept" {', 0)]),
    "M5-stub-without-feature": ("ch15", "stubbing on a model whose row lacks the column", [
        ("internal/llm/policy.go", "if feats.StubsToolResults {", "if feats.StubsToolResults || true {", 0)]),
    "M6-ladder-recomputed": ("ch15", "ladder cuts recomputed from current settings, never recorded", [
        ("re", r"e\.Record\((common\.Event\{Type: common\.Redacted, Redact: &common\.RedactData\{\s*From: from, To: d\[cut\]\.Seq, Level: common\.Redact(?:Result|Tool), Reason: reasonLadder)",
         r"e.scratch(\1", 2, "internal/llm/policy.go"),
        ("internal/llm/policy.go", "import (", 'import (\n\t"encoding/json"', 0),
        ("internal/llm/engine.go", "renderer.Render(e.Ctx, e.Cfg)", "renderer.Render(e.view(), e.Cfg)", 0),
        ("append", "internal/llm/policy.go", SCRATCH)]),
    "M7-prefix-redeclares": ("ch15", "startup tools array re-declared on skill load", [
        ("internal/llm/claude.go", "if !features.InlineTools {", "if true {", 2)]),
    "M8-log-only-at-exit": ("ch15", "events reach disk only at shutdown", [
        ("internal/common/journal.go", "if _, err := j.f.Write(b); err != nil {",
         "if _, err := len(b), error(nil); err != nil {", 0)]),
    "M9a-reducer-aborts": ("ch15", "load aborts on an unappliable event", [
        ("internal/common/save.go", 'diag(fmt.Errorf("restore: skipped event %d: %w", e.Seq, err))', "panic(err)", 0)]),
    "M9b-reducer-stops": ("ch15", "load stops at the first unappliable event, keeping what came before", [
        ("internal/common/save.go", 'diag(fmt.Errorf("restore: skipped event %d: %w", e.Seq, err))',
         'diag(fmt.Errorf("restore: skipped event %d: %w", e.Seq, err))\n\t\t\tbreak', 0)]),
    "M11-tools-never-in-dialog": ("ch15", "a skill load emits no ToolsChanged", [
        ("internal/tools/tools.go", "if ev, ok := r.toolsChanged(before); ok {",
         "if ev, ok := r.toolsChanged(before); ok && false {", 1)]),
    "M10-ch10-no-var-render": ("ch10", "ch10: $VAR tokens not rendered in skill bodies", [
        ("internal/common/skill_registry.go", "entry.Props.Body = vars.Render(entry.Props.RawBody)",
         "entry.Props.Body = entry.Props.RawBody", 0)]),
}


def titles():
    """Map each check's printed title back to its id, read from the grader
    source so the table cannot drift from the weights."""
    m = {}
    for ch in ("ch15_checks.go", "ch10_checks.go"):
        p = os.path.join(ROOT, "internal", "grade", ch)
        if os.path.exists(p):
            for cid, title in re.findall(r'\{"([a-z0-9-]+)", "([^"]+)"', open(p).read()):
                m[title] = cid
    return m


TITLES = titles()


def apply(tree, edits):
    for ed in edits:
        if ed[0] == "re":
            _, pat, repl, count, rel = ed
            p = os.path.join(tree, rel)
            s = open(p).read()
            s2, n = re.subn(pat, repl, s)
            if n != count:
                return f"{rel}: regex matched {n}, want {count}"
            open(p, "w").write(s2)
            continue
        if ed[0] == "append":
            _, rel, text = ed
            with open(os.path.join(tree, rel), "a") as f:
                f.write(text)
            continue
        rel, old, new, nth = ed
        p = os.path.join(tree, rel)
        s = open(p).read()
        n = s.count(old)
        if nth == 0:
            if n != 1:
                return f"{rel}: anchor occurs {n} times, want 1: {old!r}"
            s = s.replace(old, new)
        else:
            if n < nth:
                return f"{rel}: anchor occurs {n} times, want >= {nth}: {old!r}"
            i = -1
            for _ in range(nth):
                i = s.index(old, i + 1)
            s = s[:i] + new + s[i + len(old):]
        open(p, "w").write(s)
    return None


def main():
    args = sys.argv[1:]
    grader = "/tmp/grade15"
    if args and os.path.sep in args[0]:
        grader = args.pop(0)
    names = args or list(MUTANTS)
    print(f"{'mutant':30} {'score':>6}  failing checks")
    for name in names:
        ch, what, edits = MUTANTS[name]
        tmp = tempfile.mkdtemp(prefix="ch15mut-")
        tree = os.path.join(tmp, "tree")
        shutil.copytree(os.path.join(ROOT, "solutions", ch), tree)
        err = apply(tree, edits)
        if err:
            print(f"{name:30} {'ANCHOR-MISSING':>6}  {err}")
            shutil.rmtree(tmp)
            continue
        build = subprocess.run(["go", "build", "-o", os.devnull, "."], cwd=tree,
                               capture_output=True, text=True)
        if build.returncode != 0:
            print(f"{name:30} {'BUILD-FAIL':>6}  {build.stderr.strip().splitlines()[:3]}")
            shutil.rmtree(tmp)
            continue
        chnum = ch.replace("ch", "").lstrip("0")
        out = subprocess.run([grader, "-ch", chnum, tree], capture_output=True, text=True, cwd=ROOT)
        score = re.search(r"(\d+)\s*/\s*100", out.stdout)
        failing = [TITLES.get(t.strip(), t.strip()[:40]) for t in re.findall(r"\[FAIL\]\s+(.*?)\s+\(\d+/\d+\)", out.stdout)]
        print(f"{name:30} {score.group(1) if score else '?':>6}  {','.join(failing) or '(none: SURVIVED)'}   [{what}]")
        shutil.rmtree(tmp)


if __name__ == "__main__":
    main()
