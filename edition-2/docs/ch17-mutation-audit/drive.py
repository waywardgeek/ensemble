#!/usr/bin/env python3
"""Serial ch17 mutation audit driver.

One mutant at a time. Each deletes exactly ONE behaviour, must still compile,
and is reverted with `git checkout` before the next one. Never run two graders
at once: a concurrent sweep produced 11/100 for a grader that scores 100/100
alone, so every number here is measured with nothing else running.
"""
import subprocess, sys, os, pathlib

ROOT = pathlib.Path("/Users/bill/projects/ensemble")
REC = ROOT / "agent/internal/recall"
EVID = ROOT / "docs/ch17-mutation-audit"
IDS = ["bm25-indexes","bm25-scores","bm25-stop-words","chunking-splits",
       "per-source-quota","judge-filters","judge-fallback","recall-is-a-spoke",
       "recall-is-own-kind","injection-capped"]

# name -> (file, old, new, target check)
MUTANTS = [
  ("M1-drop-handoffs-source", "sources.go",
   '{Name: SourceHandoffs, Dir: filepath.Join(workspace, "handoffs")},', '',
   "bm25-indexes"),
  ("M2-no-length-norm", "bm25.go", "b  = 0.75", "b  = 0.0", "bm25-scores"),
  ("M3-no-stop-words", "bm25.go", "!ix.stop[t]", "true", "bm25-stop-words"),
  ("M4-no-section-split", "chunk.go", 'sectionPrefix    = "## "',
   'sectionPrefix    = "\\x00nosplit "', "chunking-splits"),
  ("M5-no-source-reservation", "recall.go", "MemoryQuota:    0.5",
   "MemoryQuota:    1.0", "per-source-quota"),
  ("M7-no-byte-cap", "recall.go", "MaxRecallBytes: 6144",
   "MaxRecallBytes: 1 << 30", "injection-capped"),
  ("M8-recall-imports-llm", "recall.go", 'import (',
   'import (\n\t_ "github.com/waywardgeek/ensemble/internal/llm"', "recall-is-a-spoke"),
]

def sh(cmd, cwd=ROOT):
    return subprocess.run(cmd, shell=True, cwd=cwd, capture_output=True, text=True)

def clean():
    sh("git checkout -- agent/internal/recall")
    d = sh("git --no-pager diff --stat")
    return d.stdout.strip() == ""

results = {}
for name, fname, old, new, target in MUTANTS:
    if not clean():
        print(f"!! tree dirty before {name}; aborting"); break
    p = REC / fname
    src = p.read_text()
    if old not in src:
        print(f"=== {name}: PATCH DID NOT APPLY (missing {old!r} in {fname})")
        results[name] = ("PATCH-FAILED", target, [])
        continue
    p.write_text(src.replace(old, new, 1))

    b = sh("go build ./...", cwd=ROOT / "agent")
    if b.returncode != 0:
        print(f"=== {name}: BUILD FAILED -- invalid mutant, proves nothing")
        print(b.stderr[:600])
        results[name] = ("BUILD-FAILED", target, [])
        clean(); continue

    r = sh(f"go run ./cmd/grade -ch 17 ./agent")
    out = r.stdout + r.stderr
    (EVID / f"{name}.txt").write_text(out)
    marks = [("FAIL" if "[FAIL]" in ln else "pass")
             for ln in out.splitlines() if "[PASS]" in ln or "[FAIL]" in ln]
    if len(marks) != 10:
        print(f"=== {name}: SCENARIO PRODUCED {len(marks)} CHECK LINES (expected 10) -- suspect")
    failed = [IDS[i] for i, m in enumerate(marks) if m == "FAIL"]
    score = [l for l in out.splitlines() if "/100" in l]
    verdict = ("KILLED-EXACTLY" if failed == [target]
               else "SURVIVED" if not failed
               else "KILLED-BROADLY")
    print(f"=== {name} -> target {target}: {verdict}")
    print(f"    failed: {failed if failed else 'NONE'}   {score[-1].strip() if score else ''}")
    results[name] = (verdict, target, failed)
    clean()

print("\n===== MATRIX =====")
for n, (v, t, f) in results.items():
    print(f"{n:28} target={t:20} {v:16} failed={','.join(f) if f else 'none'}")
print("tree clean at end:", clean())
