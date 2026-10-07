#!/usr/bin/env python3
"""Bounded post-run mutation controls for the independent human CLI checker.

Each disposable edit must compile and produce its exact expected case failures.
No paid API calls, source changes, or comprehensive mutation coverage claim.
"""

import argparse
import hashlib
import json
import pathlib
import shutil
import subprocess
import sys
import tempfile

from accept_chat import VENDORS


MUTATIONS = [
    ("terminal-default-disabled", "cmd/terminal.go", "return ok && terminalFD(",
     "return false && ok && terminalFD(", ["pty-default"]),
    ("slash-escape-retained", "cmd/chat.go", "line = line[1:]", "line = line[0:]", ["pty-chat", "pty-default"]),
    ("invalid-utf8-accepted", "cmd/chat.go", "if !utf8.Valid(line) {", "if false && !utf8.Valid(line) {", ["invalid-utf8"]),
    ("answer-json-escaped", "cmd/chat.go", '"Assistant:\\n%s\\n", text', '"Assistant:\\n%q\\n", text', ["pty-chat", "pty-default"]),
    ("prompt-not-flushed", "cmd/chat.go",
     '// Flush before reading so a person can see that the client is ready.\n\t\tif err := writer.Flush(); err != nil {',
     '// Deliberate defect: leave the prompt buffered while waiting for input.\n\t\tif err := error(nil); err != nil {',
     ["pty-chat", "pty-default"]),
]


def digest(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()


def mutations(chapter):
    result = []
    for name, file, old, new, failures in MUTATIONS:
        failures = list(failures)
        if chapter == 3 and name in ("answer-json-escaped", "prompt-not-flushed"):
            failures.append("tool-turn-redaction")
        if chapter == 3 and name == "prompt-not-flushed":
            failures.append("tool-continuation-failure")
        result.append((name, file, old, new, failures))
    if chapter == 3:
        result.extend([
            ("history-call-id-hidden", "cmd/chat.go", "if event.Tool != nil {",
             "if false && event.Tool != nil {", ["tool-turn-redaction"]),
            ("redaction-keeps-original", "internal/llm/events.go",
             'p.Parts[k] = common.Part{Type: "redacted", Stub: "[redacted]", Ref: old.Ref}',
             "p.Parts[k] = old", ["tool-turn-redaction"]),
        ])
    return result


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("solution", type=pathlib.Path)
    parser.add_argument("--chapter", type=int, choices=(2, 3), default=2)
    args = parser.parse_args()
    source = args.solution.resolve(strict=True)
    checker = pathlib.Path(__file__).with_name("accept_chat.py").resolve()
    paths = [p for p in source.rglob("*.go") if p.relative_to(source).parts[0]
             not in ("evidence", "gui", "examples", "tmp", ".git")]
    paths += [p for p in (source / "go.mod", source / "go.sum") if p.exists()]
    hashes = {str(p.relative_to(source)): digest(p) for p in sorted(paths)}
    results = []
    with tempfile.TemporaryDirectory(prefix="ensemble-chat-audit-") as tmp:
        root = pathlib.Path(tmp)
        baseline = root / "baseline"
        for p in paths:
            target = baseline / p.relative_to(source)
            target.parent.mkdir(parents=True, exist_ok=True)
            shutil.copyfile(p, target)
        if any(digest(source / p) != h or digest(baseline / p) != h for p, h in hashes.items()):
            raise RuntimeError("source changed while freezing audit")
        for name, file, old, new, failures in [("control", None, None, None, []), *mutations(args.chapter)]:
            work = root / "work"
            shutil.copytree(baseline, work)
            try:
                mutation = None
                if file:
                    path = work / file
                    before = path.read_text()
                    if before.count(old) != 1 or old == new:
                        raise RuntimeError(name + ": mutation must change one unique anchor")
                    path.write_text(before.replace(old, new, 1))
                    mutation = {"file": file, "old": old, "new": new,
                                "original_sha256": hashes[file], "mutated_sha256": digest(path)}
                binary = work / "fixture-bin"
                build = subprocess.run(["go", "build", "-o", str(binary), "./cmd"], cwd=work,
                                       capture_output=True, text=True, timeout=120)
                row = {"id": name, "mutation": mutation, "build_exit": build.returncode,
                       "expected_failures": sorted(v + "/" + f for v in VENDORS for f in failures)}
                if build.returncode:
                    row.update(passed=False, build_stderr=build.stderr)
                else:
                    run = subprocess.run([sys.executable, "-B", str(checker), str(binary),
                                          "--chapter", str(args.chapter)],
                                         capture_output=True, text=True, timeout=120)
                    receipt = json.loads(run.stdout)
                    actual = sorted(c["id"] for c in receipt["checks"] if not c["passed"])
                    row.update(actual_failures=actual, acceptance_exit=run.returncode, receipt=receipt,
                               passed=actual == row["expected_failures"] and run.returncode == int(bool(failures)))
                results.append(row)
                print(f"{name}: {'PASS' if row['passed'] else 'FAIL'}", file=sys.stderr, flush=True)
            finally:
                shutil.rmtree(work)
    unchanged = all(digest(source / p) == h for p, h in hashes.items())
    passed = unchanged and all(r["passed"] for r in results)
    print(json.dumps({"scope": __doc__, "chapter": args.chapter,
                      "source_files": hashes, "source_unchanged": unchanged,
                      "checker_sha256": digest(checker), "audit_sha256": digest(pathlib.Path(__file__)),
                      "passed": passed, "results": results}, indent=2))
    return 0 if passed else 1


if __name__ == "__main__":
    raise SystemExit(main())
