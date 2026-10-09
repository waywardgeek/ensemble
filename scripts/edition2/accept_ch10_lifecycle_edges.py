#!/usr/bin/env python3
"""Ch10 job identity/floor, partial topology, consumed-limit fault and surviving-child lock controls.

An additional deterministic subset. No real provider, large physical file,
uint64 group atomicity, or complete chapter acceptance claim.
"""
import argparse
from datetime import datetime, timezone
import json
import os
from pathlib import Path
import re
import signal
import shutil
import subprocess
import tempfile

from accept_ch10_lifecycle import bind, digest, save_receipt

HERE = Path(__file__).resolve().parent
FILES = {
    "ch10_lifecycle_edges_test.go": "ch10-lifecycle-edges_test.go",
    "ch10_lock_child_edges_test.go": "ch10-lock-child-edges_test.go",
    "ch10_job_uint64_edges_test.go": "ch10-job-uint64-edges_test.go",
    "ch10_remaining_aggregate_test.go": "ch10-remaining-aggregate_test.go",
    "internal/eventlog/ch10_edge_writer_seam.go": "ch10-fault-log-seam.go",
}
MUTATIONS = (
    ("root-floor-not-raised", "session.go", "if floor > e.nextHandle {", "if false {", "RootJobFloor"),
    # Leave fresh floor zero alone so the genuine job101 parent survives; the
    # defect lowers an already higher root only when restoring a real job floor.
    ("root-floor-lowered", "session.go", "if floor > e.nextHandle {", "if floor != 0 {", "RootJobFloor"),
    ("lock-inherited-by-exec", "internal/persistence/lock_unix.go",
     "syscall.CloseOnExec(int(f.Fd()))", "syscall.Syscall(syscall.SYS_FCNTL, f.Fd(), syscall.F_SETFD, 0)", "SurvivingChildLock"),
    ("root-job-wrap", "ensemble.go", "if e.nextHandle == ^uint64(0) {", "if false {", "JobUint64AndCollision"),
)
EXPECTED_FAILURES = {
    "root-floor-not-raised": "root allocator floor: got",
    "root-floor-lowered": "root allocator floor: got",
    "lock-inherited-by-exec": "surviving tool child retained store lock:",
    "root-job-wrap": "exhausted root wrapped or reused a handle",
}


def mutation_detected(process, mutation, group):
    """Only the intended assertion earns credit after a genuine positive."""
    return (process.returncode != 0
            and ("--- FAIL: TestCh10Edges" + group) in process.stdout
            and EXPECTED_FAILURES[mutation] in process.stdout
            and "[build failed]" not in process.stdout
            and "panic: test timed out" not in process.stdout)


def cleanup_owned(ledger):
    """Bound timeout cleanup to the exact scratch child/holder identities."""
    if not ledger.exists():
        return []
    value = json.loads(ledger.read_text())
    marker = value["marker"]
    if not re.fullmatch(r"ch10-edge-owned-\d+", marker):
        raise ValueError("invalid owned-child ledger marker")
    workspace = Path(value["workspace"])
    observations = []
    try:
        child = int((workspace / "edge-child.pid").read_text().strip())
    except (OSError, ValueError):
        child = 0
    candidates = [(child, marker, "tool child"),
                  (value["holder_pid"], value["binary"] + " -test.run=^TestCh10EdgesLockHolder$", "holder")]
    for pid, identity, role in candidates:
        if not isinstance(pid, int) or pid <= 1 or pid == os.getpid():
            continue
        p = subprocess.run(["ps", "-p", str(pid), "-o", "command="], text=True, capture_output=True, timeout=5)
        matches = p.returncode == 0 and identity in p.stdout
        row = dict(role=role, pid=pid, identity_matched=matches, killed=False)
        if matches:
            try:
                os.kill(pid, signal.SIGKILL)
                row["killed"] = True
            except ProcessLookupError:
                row["already_exited"] = True
        observations.append(row)
    ledger.unlink()
    return observations


def evaluate(args):
    source = args.source_directory.resolve(strict=True)
    commit, source_files = bind(source, args.source_commit)
    paths = [Path(__file__).resolve(), HERE / "accept_ch10_lifecycle.py", *(HERE / f for f in FILES.values())]
    checkers = {p.name: digest(p) for p in paths}
    result = dict(source_commit=commit, source_files=source_files, source_directory=str(source),
                  checker_files=checkers, scope=__doc__, full_chapter_acceptance=False,
                  passed=False, complete=False, checks=[], selected=args.run)

    def save():
        save_receipt(args.receipt, result)

    def run(command, mutation=None, group=None):
        row = dict(command=command, mutation=mutation, started=True, complete=False,
                   started_utc=datetime.now(timezone.utc).isoformat(),
                   disk_free_before=shutil.disk_usage(source).free)
        result["checks"].append(row)
        save()
        try:
            environment = dict(os.environ, ENSEMBLE_CH10_EDGE_LEDGER=str(work / "owned-child.json"))
            p = subprocess.run(command, cwd=source, env=environment, text=True, capture_output=True, timeout=240)
            row.update(exit=p.returncode, stdout=p.stdout, stderr=p.stderr, complete=True)
            row["passed"] = mutation_detected(p, mutation, group) if mutation else (
                p.returncode == 0 and (command[0] != "gofmt" or not p.stdout))
        except subprocess.TimeoutExpired as exc:
            def text(v):
                return v.decode(errors="replace") if isinstance(v, bytes) else v
            row.update(passed=False, timeout=True, stdout=text(exc.stdout), stderr=text(exc.stderr))
        finally:
            row["owned_process_cleanup"] = cleanup_owned(work / "owned-child.json")
            row["finished_utc"] = datetime.now(timezone.utc).isoformat()
            row["disk_free_after"] = shutil.disk_usage(source).free
            save()
        save()
        return row["passed"]

    save()
    with tempfile.TemporaryDirectory(prefix="ch10-lifecycle-edges-") as d:
        work = Path(d)
        base = {}
        for target, fixture in FILES.items():
            virtual = source / target
            if virtual.exists():
                raise ValueError("overlay collision: " + target)
            local = work / fixture
            local.write_bytes((HERE / fixture).read_bytes())
            base[str(virtual)] = str(local)
        overlay = work / "positive.json"
        overlay.write_text(json.dumps({"Replace": base}))
        good = all(run(c) for c in [
            ["gofmt", "-l", *base.values()],
            ["go", "vet", "-overlay", str(overlay), "./..."],
            ["go", "test", "-overlay", str(overlay), "-race", "-count=1", "-timeout=120s", "-run", args.run, "-v", "."],
            ["go", "test", "./...", "-count=1"],
        ])
        if good and args.audit:
            if args.run != "^TestCh10Edges":
                raise ValueError("audit requires complete positive subset")
            for name, relative, old, new, group in MUTATIONS:
                original = (source / relative).read_text()
                if original.count(old) != 1:
                    raise ValueError("mutation anchor not unique: " + name)
                local = work / (name + ".go")
                local.write_text(original.replace(old, new, 1))
                replacements = dict(base)
                replacements[str(source / relative)] = str(local)
                mutant = work / (name + ".json")
                mutant.write_text(json.dumps({"Replace": replacements}))
                result.setdefault("mutations", []).append(dict(name=name, file=relative, old=old, new=new,
                    original_sha256=digest(source / relative), changed_sha256=digest(local)))
                good = run(["go", "test", "-overlay", str(mutant), "-count=1", "-timeout=120s", "-run",
                            "^TestCh10Edges" + group + "$", "-v", "."], name, group) and good
    if bind(source, commit)[1] != source_files or checkers != {p.name: digest(p) for p in paths}:
        raise ValueError("source/checker changed during checks")
    result.update(complete=True, passed=good)
    save()
    return good


if __name__ == "__main__":
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("source_directory", type=Path)
    parser.add_argument("--source-commit", required=True)
    parser.add_argument("--receipt", type=Path, required=True)
    parser.add_argument("--run", default="^TestCh10Edges")
    parser.add_argument("--audit", action="store_true")
    args = parser.parse_args()
    raise SystemExit(0 if evaluate(args) else 1)
