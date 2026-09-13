#!/usr/bin/env python3
"""Property-level audit of the chapter 1 grader (course-policy P9).

Method: delete ONE behaviour the chapter promises from a COPY of the reference
solution, score it, and record the failing check set. A row that reads 100 -> 100
is a finding: the chapter promised something the grader never collects on.

Every mutation asserts its anchor matched. A silently-unapplied mutation would
score 100 and manufacture a fake finding, which is the failure mode this audit
exists to catch.
"""
import json
import os
import re
import shutil
import subprocess
import sys

REPO = subprocess.run(["git", "rev-parse", "--show-toplevel"],
                      capture_output=True, text=True, check=True).stdout.strip()
SRC = os.path.join(REPO, "solutions", "ch01")
WORK = "/tmp/ch01audit/work"

GOMOD = "module mutant\n\ngo 1.21\n"


def sub1(pattern, repl, flags=0):
    """Return a mutator that requires EXACTLY ONE match."""
    def f(src):
        rx = re.compile(pattern, flags)
        n = len(rx.findall(src))
        if n != 1:
            raise AssertionError(f"anchor matched {n} times, want 1: {pattern!r}")
        return rx.sub(repl, src, count=1)
    return f


def chain(*fs):
    def f(src):
        for g in fs:
            src = g(src)
        return src
    return f


BLOCK_WALK = (
    r'[ \t]*for _, block := range parsed\.Content \{\n'
    r'[ \t]*if block\.Type == "text" \{\n'
    r'[ \t]*text\.WriteString\(block\.Text\)\n'
    r'[ \t]*\}\n'
    r'[ \t]*\}\n'
)

MUTANTS = [
    # name, promise (chapter), mutation
    ("none",
     "control: unmodified reference solution",
     lambda s: s),

    ("blocks-first-only",
     "L116 'always a list of typed blocks. Walk it and concatenate the text blocks'",
     sub1(BLOCK_WALK,
          '\tif len(parsed.Content) > 0 {\n\t\ttext.WriteString(parsed.Content[0].Text)\n\t}\n')),

    ("blocks-no-type-filter",
     "L116 typed blocks: only the 'text' blocks may be concatenated",
     sub1(BLOCK_WALK,
          '\tfor _, block := range parsed.Content {\n\t\ttext.WriteString(block.Text)\n\t}\n')),

    ("assistant-not-appended",
     "the conversation is resent: the model's own replies must be appended",
     sub1(r'[ \t]*\*?conv = append\([^\n]*"assistant"[^\n]*\n', '')),

    ("usage-last-only",
     "cumulative usage from the first request",
     chain(sub1(r'c\.InputTokens \+=', 'c.InputTokens ='),
           sub1(r'c\.OutputTokens \+=', 'c.OutputTokens ='))),

    ("usage-line-missing",
     "on EOF print {\"usage\":{...}} and exit 0",
     sub1(r'return out\.Encode\(u\)\n', 'return nil\n')),

    ("hdr-version-dropped",
     "send the anthropic-version header",
     sub1(r'[ \t]*req\.Header\.Set\("anthropic-version"[^\n]*\n', '')),

    ("hdr-content-type-dropped",
     "send content-type: application/json",
     sub1(r'[ \t]*req\.Header\.Set\("[Cc]ontent-[Tt]ype"[^\n]*\n', '')),

    ("hdr-apikey-dropped",
     "send ANTHROPIC_API_KEY as the x-api-key header",
     sub1(r'[ \t]*req\.Header\.Set\("x-api-key"[^\n]*\n', '')),

    ("apikey-hardcoded",
     "L296 'read all three, hardcode none' (ANTHROPIC_API_KEY)",
     sub1(r'os\.Getenv\("ANTHROPIC_API_KEY"\)', '"sk-ant-hardcoded-not-from-env"')),

    ("model-hardcoded",
     "L296 'read all three, hardcode none' (ANTHROPIC_MODEL)",
     sub1(r'os\.Getenv\("ANTHROPIC_MODEL"\)', '"claude-hardcoded-not-from-env"')),

    ("baseurl-hardcoded",
     "L296 'read all three, hardcode none' (ANTHROPIC_BASE_URL)",
     sub1(r'os\.Getenv\("ANTHROPIC_BASE_URL"\)', '"http://127.0.0.1:9"')),

    ("maxtokens-zero",
     "max_tokens is required on every request",
     sub1(r'MaxTokens:(\s+)[A-Za-z0-9_.]+,', r'MaxTokens:\g<1>0,')),

    ("stdout-noise",
     "stdout carries the protocol only; diagnostics go to stderr",
     sub1(r'([ \t]*)(if err := out\.Encode\(graderOut\{Assistant: reply\}\))',
          r'\1fmt.Println("[debug] got a reply")\n\1\2')),

    ("exit-nonzero",
     "after the usage line, exit 0",
     sub1(r'return out\.Encode\(u\)\n',
          'if err := out.Encode(u); err != nil {\n\t\treturn err\n\t}\n\tos.Exit(3)\n\treturn nil\n')),
]


def build_mutant(name, mutate):
    d = os.path.join(WORK, name)
    shutil.rmtree(d, ignore_errors=True)
    os.makedirs(d)
    changed = False
    for fn in sorted(os.listdir(SRC)):
        if not fn.endswith(".go"):
            continue
        src = open(os.path.join(SRC, fn)).read()
        try:
            out = mutate(src)
        except AssertionError:
            out = src
        if out != src:
            changed = True
        open(os.path.join(d, fn), "w").write(out)
    with open(os.path.join(d, "go.mod"), "w") as f:
        f.write(GOMOD)
    if name != "none":
        # re-run to surface the real anchor error if nothing changed anywhere
        if not changed:
            src = open(os.path.join(SRC, "main.go")).read()
            mutate(src)  # raises with the failing anchor
            raise AssertionError(f"{name}: mutation was a no-op")
    return d


def score(d):
    p = subprocess.run(
        ["go", "run", "./cmd/grade", "-ch", "1", "-json", d],
        cwd=REPO, capture_output=True, text=True, timeout=300)
    txt = p.stdout.strip()
    start = txt.find("{")
    if start < 0:
        return None, [], (p.stderr or txt)[-400:]
    rep = json.loads(txt[start:])
    failing = sorted(c["id"] for c in rep["checks"] if not c["passed"])
    return rep["score"], failing, ""


def main():
    os.makedirs(WORK, exist_ok=True)
    rows = []
    for name, promise, mutate in MUTANTS:
        try:
            d = build_mutant(name, mutate)
        except AssertionError as e:
            rows.append((name, promise, "ANCHOR-FAIL", [], str(e)))
            print(f"!! {name}: {e}", file=sys.stderr)
            continue
        s, failing, err = score(d)
        rows.append((name, promise, s, failing, err))
        print(f"{name:26} score={s!s:>4}  failing={','.join(failing) or '-'}  {err}",
              flush=True)
    with open("/tmp/ch01audit/results.json", "w") as f:
        json.dump([{"name": n, "promise": p, "score": s, "failing": fl, "err": e}
                   for n, p, s, fl, e in rows], f, indent=2)


if __name__ == "__main__":
    main()
