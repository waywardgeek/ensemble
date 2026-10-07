#!/usr/bin/env python3
"""Serial, disposable Chapter 3 mutation controls for selected new guarantees.

Run only after preserving the student's first attempt. Each source edit must
land once, compile, and lose exactly its advertised independent checks.
This is scoped sensitivity evidence, not comprehensive mutation coverage.
"""

import argparse
import hashlib
import json
import pathlib
import shutil
import subprocess
import sys
import tempfile


MUTATIONS = [
    ("read-byte-slice", "internal/tools/files.go",
     "prefix := textPrefix(r, text, limit)",
     "prefix := text\n\tif len(prefix) > limit { prefix = prefix[:limit] }",
     ["utf8-prefix-caps"]),
    ("command-byte-slice", "internal/tools/command.go",
     "text := textPrefix(c.parent, raw, c.limit)", "text := raw",
     ["utf8-prefix-caps"]),
    ("stop-batch-after-error", "internal/llm/turn.go",
     '\t\tresult := owner.Registry().Execute(part)',
     '\t\tresult := owner.Registry().Execute(part)', []),
    ("fifteen-request-bound", "internal/llm/turn.go",
     "round <= 16", "round < 16", ["last-batch-limit", "sixteenth-final-success"]),
    ("skip-sixteenth-batch", "internal/llm/turn.go",
     "\t\t\tcalls++", '\t\t\tif round == 16 { return fail("round_limit", failure(e, "fixture early bound")) }\n\t\t\tcalls++',
     ["last-batch-limit"]),
]

# Insert only after the matching result has been recorded. The defect leaves
# later calls unanswered rather than destroying the result of the first error.
MUTATIONS[2] = (
    "stop-batch-after-error", "internal/llm/turn.go",
    '\t\t\t\treturn common.ClientResult{}, fmt.Errorf("tool %s completion could not be recorded; inspect its actual effect before any new attempt: %w", part.Name, err)\n\t\t\t}',
    '\t\t\t\treturn common.ClientResult{}, fmt.Errorf("tool %s completion could not be recorded; inspect its actual effect before any new attempt: %w", part.Name, err)\n\t\t\t}\n\t\t\tif result.IsError { break }',
    ["edits-writes", "read-ranges-caps", "search-projection", "ordinary-errors-complete-batch", "listing-order-caps"],
)


def digest(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("solution", type=pathlib.Path)
    args = parser.parse_args()
    source = args.solution.resolve(strict=True)
    checker = pathlib.Path(__file__).with_name("accept_ch03.py").resolve()
    source_files = [p for p in source.rglob("*.go")
                    if p.relative_to(source).parts[0] not in ("evidence", "gui", "examples", ".git")]
    source_files += [p for p in (source / "go.mod", source / "go.sum") if p.exists()]
    hashes = {str(p.relative_to(source)): digest(p) for p in sorted(source_files)}
    results = []
    with tempfile.TemporaryDirectory(prefix="ensemble-ch03-audit-") as tmp:
        root = pathlib.Path(tmp)
        baseline = root / "baseline"
        for p in source_files:
            target = baseline / p.relative_to(source)
            target.parent.mkdir(parents=True, exist_ok=True)
            shutil.copyfile(p, target)
        if any(digest(source / name) != value or digest(baseline / name) != value for name, value in hashes.items()):
            raise RuntimeError("source changed while freezing audit baseline")
        for name, file, old, new, expected in [("control", None, None, None, []), *MUTATIONS]:
            work = root / "work"
            shutil.copytree(baseline, work)
            try:
                mutation = None
                if file:
                    path = work / file
                    before = path.read_text()
                    if before.count(old) != 1 or old == new:
                        raise RuntimeError(f"{name}: mutation anchor is not unique or is a no-op")
                    path.write_text(before.replace(old, new, 1))
                    mutation = {"file": file, "original_sha256": hashes[file], "mutated_sha256": digest(path),
                                "old": old, "new": new}
                binary = work / "fixture-bin"
                build = subprocess.run(["go", "build", "-o", str(binary), "./cmd"], cwd=work,
                                       capture_output=True, text=True, timeout=120)
                row = {"id": name, "mutation": mutation, "build_exit": build.returncode,
                       "expected_failures": sorted("anthropic/" + x for x in expected)}
                if build.returncode:
                    row.update(passed=False, build_stderr=build.stderr)
                else:
                    run = subprocess.run([sys.executable, "-B", str(checker), str(binary), "--vendor", "anthropic"],
                                         capture_output=True, text=True, timeout=240)
                    receipt = json.loads(run.stdout)
                    actual = sorted(c["id"] for c in receipt["checks"] if not c["passed"])
                    row.update(acceptance_exit=run.returncode, actual_failures=actual, receipt=receipt,
                               passed=actual == row["expected_failures"] and run.returncode == (1 if expected else 0))
                results.append(row)
                # Progress is stderr so stdout remains one durable JSON artifact.
                print(f"{name}: {'PASS' if row['passed'] else 'FAIL'}", file=sys.stderr, flush=True)
            finally:
                shutil.rmtree(work)
    unchanged = all(digest(source / name) == value for name, value in hashes.items())
    passed = unchanged and all(r["passed"] for r in results)
    print(json.dumps({"scope": "one passing control; UTF-8 read/command, batch-error continuation, request bound and final-batch mutations only",
                      "source_files": hashes, "source_unchanged": unchanged,
                      "checker_sha256": digest(checker), "audit_sha256": digest(pathlib.Path(__file__)),
                      "passed": passed, "results": results}, indent=2))
    return 0 if passed else 1


if __name__ == "__main__":
    raise SystemExit(main())
