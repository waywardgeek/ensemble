#!/usr/bin/env python3
"""Partial Ch10 exact-bound checks against a pinned new-source extraction.

Actual Agent-owned codec controls cover decimal/scalar canonicalization,
nesting 128/129, handler count 1,024/+1 and canonical total 16 MiB/+1.
Public import/turn controls cover the burned request cursor through uint64 max.
No claim of aggregate semantic-count, 256/512 MiB, 1 GiB, allocation-memory,
activation/job whole-group, browser or complete chapter acceptance is made.
Binding covers every runtime Go/module/API-format file, excluding evidence
helpers. Complete delivered-tree discovery remains a separate root-owned gate.
"""
import argparse
import hashlib
import json
from pathlib import Path
import shutil
import subprocess
import tempfile

HERE = Path(__file__).resolve().parent
ROOT = HERE.parents[1]
PREFIX = "solutions/edition-2/main/"
FIXTURES = ("ch10-session-public_test.go", "ch10-session-roundtrip_test.go",
            "ch10-storage-bounds_test.go")
MUTATIONS = (
    ("canonical-through-float", "internal/persistence/json.go", "number = c.number(number)",
     "rounded, _ := strconv.ParseFloat(number, 64); number = c.number(strconv.FormatFloat(rounded, 'g', -1, 64))", "Canonical"),
    ("nesting-one-extra", "internal/persistence/json.go", "depth >= 128", "depth >= 129", "Nesting"),
    ("handler-count-one-extra", "internal/persistence/codec.go", "len(id.Handlers) > 1024", "len(id.Handlers) > 1025", "HandlerCount"),
    ("handler-bytes-one-extra", "internal/persistence/codec.go", "len(canonical) > 16<<20", "len(canonical) > (16<<20)+1", "HandlerBytes"),
    ("request-wrap", "internal/llm/actor.go", "if a.next == ^uint64(0) {", "if false {", "RequestUint64"),
)


def sha(data):
    return hashlib.sha256(data).hexdigest()


def git(*args):
    return subprocess.check_output(["git", "-C", str(ROOT), *args])


def binding(source, revision):
    commit = git("rev-parse", "--verify", revision + "^{commit}").decode().strip()
    names = git("ls-tree", "-r", "--name-only", commit, "--", PREFIX).decode().splitlines()
    names = [p for p in names if not p.startswith(PREFIX + "evidence/") and
             (p.endswith(".go") or Path(p).name in
              ("go.mod", "go.sum", "session-api.md", "persistence-format.md"))]
    if not names or PREFIX + "go.mod" not in names:
        raise ValueError("empty/incomplete immutable runtime source map")
    result = {}
    for name in names:
        relative = name.removeprefix(PREFIX)
        expected = sha(git("show", commit + ":" + name))
        path = source / relative
        if not path.is_file() or sha(path.read_bytes()) != expected:
            raise ValueError("source identity mismatch: " + relative)
        result[relative] = expected
    # Additional Go sources could affect a module even when all expected files
    # match. An extraction with injected source is not the pinned source.
    actual = {str(p.relative_to(source)) for p in source.rglob("*.go")
              if p.relative_to(source).parts[0] != "evidence"}
    if actual != {p for p in result if p.endswith(".go")}:
        raise ValueError("source contains missing/additional Go files")
    return commit, result


def evaluate(args):
    source = args.source_directory.resolve(strict=True)
    commit, source_files = binding(source, args.source_commit)
    checker_paths = [Path(__file__).resolve(), *(HERE / name for name in FIXTURES)]
    checkers = {p.name: sha(p.read_bytes()) for p in checker_paths}
    result = dict(source_commit=commit, source_directory=str(source), source_files=source_files,
                  checker_files=checkers, scope=__doc__, full_chapter_acceptance=False,
                  passed=False, complete=False, checks=[])

    def save():
        if args.receipt:
            args.receipt.write_text(json.dumps(result, indent=2) + "\n")

    def run(command, work, *, intended_failure=False, group=None):
        row = dict(command=command, intended_failure=intended_failure, group=group, started=True)
        result["checks"].append(row)
        save()  # Retain the started command even if a compiler exhausts disk.
        try:
            proc = subprocess.run(command, cwd=work, text=True, capture_output=True, timeout=180)
            row.update(exit=proc.returncode, stdout=proc.stdout, stderr=proc.stderr)
            if intended_failure:
                row["passed"] = proc.returncode != 0 and ("--- FAIL: TestCh10Storage" + group) in proc.stdout
            else:
                row["passed"] = proc.returncode == 0 and (command[0] != "gofmt" or not proc.stdout)
        except subprocess.TimeoutExpired as exc:
            row.update(passed=False, timeout=True, stdout=(exc.stdout or b"").decode(errors="replace")
                       if isinstance(exc.stdout, bytes) else exc.stdout,
                       stderr=(exc.stderr or b"").decode(errors="replace")
                       if isinstance(exc.stderr, bytes) else exc.stderr)
        save()
        return row["passed"]

    save()
    with tempfile.TemporaryDirectory(prefix="ch10-storage-") as directory:
        work = Path(directory)
        (work / "go.mod").write_text("module example.com/ch10-storage-consumer\n\ngo 1.25\n\n"
                                    "require example.com/ensemble v0.0.0\nreplace example.com/ensemble => "
                                    + json.dumps(str(source)) + "\n")
        for name in FIXTURES:
            shutil.copyfile(HERE / name, work / name)
        commands = (["go", "mod", "tidy"], ["gofmt", "-l", *FIXTURES], ["go", "vet", "./..."],
                    ["go", "test", "-race", "-count=1", "-timeout=90s", "-run", "^TestCh10Storage", "-v", "./..."])
        good = all(run(command, work) for command in commands)
        if good and args.audit:
            for name, relative, old, new, group in MUTATIONS:
                original = (source / relative).read_text()
                if original.count(old) != 1:
                    raise ValueError("mutation adapter anchor is not unique: " + name)
                changed = original.replace(old, new, 1)
                replacement = work / (name + ".txt")
                replacement.write_text(changed)
                overlay = work / (name + ".json")
                overlay.write_text(json.dumps({"Replace": {str(source / relative): str(replacement)}}))
                result.setdefault("mutations", []).append(dict(name=name, file=relative, old=old, new=new,
                                                                 original_sha256=sha(original.encode()),
                                                                 changed_sha256=sha(changed.encode())))
                good = run(["go", "test", "-overlay", str(overlay), "-count=1", "-timeout=90s", "-run",
                            "^TestCh10Storage" + group + "$", "-v", "./..."], work,
                           intended_failure=True, group=group) and good
    if binding(source, commit)[1] != source_files or checkers != {p.name: sha(p.read_bytes()) for p in checker_paths}:
        raise ValueError("source/checker changed during checks")
    result.update(passed=good, complete=True)
    save()
    return result


if __name__ == "__main__":
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("source_directory", type=Path)
    parser.add_argument("--source-commit", required=True)
    parser.add_argument("--receipt", type=Path, required=True)
    parser.add_argument("--audit", action="store_true", help="run five intended behavior deletions after real positives")
    args = parser.parse_args()
    report = evaluate(args)
    print(json.dumps({"passed": report["passed"], "checks": len(report["checks"]),
                      "source_commit": report["source_commit"], "receipt": str(args.receipt)}))
    raise SystemExit(0 if report["passed"] else 1)
