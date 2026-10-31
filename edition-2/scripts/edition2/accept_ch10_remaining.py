#!/usr/bin/env python3
"""Additional Ch10 semantic aggregate and streaming reader controls.

Component coverage: actual million entries/parts/request identities and
1 GiB streaming input. Public checks: forged construction facts and store leaves.
Does not claim physical log, checkpoint-file, canonical
state, activation/call, whole-group overflow, or full chapter coverage.
"""
import argparse
import json
from pathlib import Path
import subprocess
import tempfile

from accept_ch10_lifecycle import bind, digest, save_receipt

HERE = Path(__file__).resolve().parent
FIXTURES = ("ch10-remaining-aggregate_test.go", "ch10-remaining-stream_test.go",
            "ch10-remaining-public_test.go")
MUTATIONS = {
    "aggregate-entries": ("AggregateEntries", {
        "internal/llm/semantic.go": [("c.Index.Entries > 1000000 || ", "", 1),
                                     ("len(entries) > 1000000 || ", "", 1)],
        "internal/llm/session_limits.go": [("d.Entries > 1000000 || ", "", 1)],
    }),
    "aggregate-parts": ("AggregateParts", {
        "internal/llm/semantic.go": [("c.Index.Parts > 1000000", "false", 1),
                                     ("parts > 1000000", "false", 1)],
        "internal/llm/session_limits.go": [("d.Parts > 1000000 || ", "", 1)],
    }),
    "seen-requests": ("SeenRequests", {
        "internal/llm/semantic.go": [(" || len(c.TurnIDs) > 1000000", "", 1)],
        "internal/llm/session_limits.go": [(" || turns > 1000000", "", 1),
                                          ("calls, turns := len(c.Calls), len(c.TurnIDs)", "calls := len(c.Calls)", 1),
                                          ('if e.Type == "turn_started" {\n\t\tturns++\n\t}', "", 1)],
    }),
    "streaming-log-one-extra": ("StreamingLogBytes", {
        "internal/eventlog/log.go": [("size > 1<<30", "size > (1<<30)+1", 2)],
    }),
    "public-initialization": ("PublicConstructionFacts", {
        "ensemble.go": [('if err == nil && owned.Session != nil && persist && a.id != "" {', "if false {", 1)],
    }),
    "store-leaf-kind": ("StoreLeaves", {
        "internal/persistence/store.go": [("if !st.Mode().IsRegular() {", "if st.Mode().IsRegular() && false {", 1)],
    }),
}


def evaluate(args):
    source = args.source_directory.resolve(strict=True)
    commit, source_files = bind(source, args.source_commit)
    paths = [Path(__file__).resolve(), HERE / "accept_ch10_lifecycle.py",
             *(HERE / name for name in FIXTURES)]
    checkers = {p.name: digest(p) for p in paths}
    result = dict(source_commit=commit, source_directory=str(source), source_files=source_files,
                  checker_files=checkers, complete=False, passed=False, checks=[],
                  scope=__doc__, selected=args.run, full_chapter_acceptance=False)

    def save():
        save_receipt(args.receipt, result)

    def run(command, mutation=None, group=None):
        row = dict(command=command, mutation=mutation, started=True, complete=False)
        result["checks"].append(row)
        save()
        try:
            p = subprocess.run(command, cwd=source, capture_output=True, text=True, timeout=600)
            row.update(exit=p.returncode, stdout=p.stdout, stderr=p.stderr, complete=True)
            if mutation:
                row["passed"] = p.returncode != 0 and ("--- FAIL: TestCh10Remaining" + group) in p.stdout
            else:
                row["passed"] = p.returncode == 0 and (command[0] != "gofmt" or not p.stdout)
        except subprocess.TimeoutExpired as exc:
            def text(x):
                return x.decode(errors="replace") if isinstance(x, bytes) else x
            row.update(passed=False, timeout=True, stdout=text(exc.stdout), stderr=text(exc.stderr))
        save()
        return row["passed"]

    save()
    with tempfile.TemporaryDirectory(prefix="ch10-remaining-") as d:
        work = Path(d)
        base = {}
        for fixture in FIXTURES:
            destination = source / fixture.replace("-", "_")
            if destination.exists():
                raise ValueError("overlay fixture collision: " + str(destination))
            target = work / fixture
            target.write_bytes((HERE / fixture).read_bytes())
            base[str(destination)] = str(target)
        overlay = work / "positive.json"
        overlay.write_text(json.dumps({"Replace": base}))
        commands = [
            ["gofmt", "-l", *base.values()],
            ["go", "vet", "-overlay", str(overlay), "./..."],
            ["go", "test", "-overlay", str(overlay), "-count=1", "-timeout=8m", "-run", args.run, "-v", "."],
        ]
        good = all(run(c) for c in commands)
        if good and args.audit:
            if args.run != "^TestCh10Remaining":
                raise ValueError("audit requires the complete positive subset first")
            for name, (group, edits) in MUTATIONS.items():
                replacements = dict(base)
                mutation = dict(name=name, edits=[])
                for relative, changes in edits.items():
                    original = (source / relative).read_text()
                    changed = original
                    for old, new, count in changes:
                        if changed.count(old) != count:
                            raise ValueError("mutation anchor count changed: " + name + ": " + relative)
                        changed = changed.replace(old, new)
                    path = work / (name + "-" + Path(relative).name)
                    path.write_text(changed)
                    replacements[str(source / relative)] = str(path)
                    mutation["edits"].append(dict(path=relative, original_sha256=digest(source / relative),
                                                   changed_sha256=digest(path), replacements=changes))
                result.setdefault("mutations", []).append(mutation)
                mutant = work / (name + ".json")
                mutant.write_text(json.dumps({"Replace": replacements}))
                good = run(["go", "test", "-overlay", str(mutant), "-count=1", "-timeout=8m", "-run",
                            "^TestCh10Remaining" + group + "$", "-v", "."], name, group) and good
    if bind(source, commit)[1] != source_files or checkers != {p.name: digest(p) for p in paths}:
        raise ValueError("source or checker identity changed")
    result.update(complete=True, passed=good)
    save()
    return good


if __name__ == "__main__":
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("source_directory", type=Path)
    parser.add_argument("--source-commit", required=True)
    parser.add_argument("--receipt", type=Path, required=True)
    parser.add_argument("--run", default="^TestCh10Remaining")
    parser.add_argument("--audit", action="store_true")
    args = parser.parse_args()
    raise SystemExit(0 if evaluate(args) else 1)
