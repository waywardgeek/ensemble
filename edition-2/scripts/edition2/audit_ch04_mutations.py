#!/usr/bin/env python3
"""Targeted deletion audit against disposable Chapter 4 source copies.

Anchors describe the reviewed initial implementation and must occur exactly
once. A compile failure is an invalid mutation, never a detected behavior.
No input source, historical reference, live evidence, or repository is edited.
"""
import argparse
import hashlib
import json
import os
import pathlib
import shutil
import subprocess
import tempfile
import time

import accept_ch04 as cli
import accept_ch04_public as public
import prepare_ch04_probes as probes


def replace(root, relative, before, after):
    path = root / relative
    text = path.read_text()
    if text.count(before) != 1:
        raise ValueError(f"mutation anchor not unique: {relative}: {before!r}")
    path.write_text(text.replace(before, after))


def command(args, cwd, timeout=60):
    result = subprocess.run(args, cwd=cwd, capture_output=True, text=True, timeout=timeout)
    return {"args": args, "exit": result.returncode, "stdout": result.stdout, "stderr": result.stderr}


def scenario(binary, name):
    setup, verify = {
        "group-kill": (cli.setup_kill, lambda *args: cli.verify_stopped(*args, "kill_job")),
        "eof-shutdown": (cli.setup_shutdown, lambda *args: cli.verify_stopped(*args, "shutdown")),
    }[name]
    with tempfile.TemporaryDirectory(prefix="ch04-mutation-run-") as tmp:
        root = pathlib.Path(tmp)
        batches = setup(root)
        replies = [cli.response("anthropic", b, cli.NARRATION) for b in batches] + [cli.response("anthropic", [])]
        session = cli.Session(binary, root, "anthropic", replies)
        try:
            events, results, errors = cli.evaluate(session, batches)
            verify(root, events, results, errors)
        except Exception as err:
            errors = [f"fixture failed: {type(err).__name__}: {err}"]
        finally:
            session.close()
    return {"id": name, "passed": not errors, "details": errors}


MUTATIONS = [
    ("truncate-occupied-artifact", "occupied", [("internal/jobs/jobs.go", "os.O_CREATE|os.O_EXCL|os.O_RDWR", "os.O_CREATE|os.O_TRUNC|os.O_RDWR")]),
    ("repeat-consumed-output", "cursor", [("internal/jobs/report.go", "j.cursor = snapshot.Bytes", "j.cursor = 0")]),
    ("predict-missing-id", "identity", [("ensemble.go", 'fmt.Sprintf("call-%d-%d", owned.Seq, index)', 'fmt.Sprintf("call-%d-%d", owned.Seq-1, index)')]),
    ("omit-background-observation", "public", [("ensemble.go", "if notify {", 'if notify && owned.Type != "job_ended" {')]),
    ("skip-durable-dispatch", "TestIndependentDurablePreDispatch", [("internal/llm/turn.go", 'if err = owner.RecordTurn(common.Event{Type: "tool_called", Tool: &called}); err != nil {', 'if err = error(nil); err != nil {')]),
    ("consume-before-report-commit", "TestIndependentFailedReportRetainsCursor", [("internal/jobs/report.go", 'if err = s.parent.RecordJob(common.Event{Type: "tool_returned", Tool: &result}); err != nil {', 'j.cursor = snapshot.Bytes\n\tif err = s.parent.RecordJob(common.Event{Type: "tool_returned", Tool: &result}); err != nil {')]),
    ("ignore-spool-write-failure", "TestIndependentSpoolFailureWakesWaiter", [("internal/jobs/jobs.go", 's.fail(fmt.Errorf("job %d output write failed; complete output unavailable", j.snapshot.Handle))', '// mutation: silently ignore spool write failure')]),
    ("killed-as-done", "group-kill", [("internal/jobs/process.go", 's.finish(j, "killed", j.snapshot.Reason)', 's.finish(j, "done", "")')]),
    ("kill-only-shell", "group-kill", [("internal/jobs/jobs.go", 'if err := syscall.Kill(-j.pid, syscall.SIGKILL);', 'if err := syscall.Kill(j.pid, syscall.SIGKILL);')]),
    ("skip-agent-shutdown", "eof-shutdown", [("ensemble.go", "err := a.jobs.Close()", "var err error")]),
    ("source-notice-in-artifact", "source-cap", [("internal/jobs/jobs.go", "s.write(j, []byte(result.Text))", "s.write(j, []byte(result.Text + result.Note))")]),
]


def run_check(root, case, work):
    if case.startswith("Test"):
        probes.install(root)
        result = command(["go", "test", "./...", "-run", "^" + case + "$", "-count=1", "-timeout=12s"], root)
        result.update(id=case, passed=result["exit"] == 0)
        return result
    if case == "public":
        consumer = work / "consumer"
        consumer.mkdir()
        shutil.copyfile(pathlib.Path(__file__).with_name("ch04_public_consumer.go.txt"), consumer / "main.go")
        (consumer / "go.mod").write_text("module independent.example/mutation\n\ngo 1.25\n\nrequire example.com/ensemble v0.0.0\nreplace example.com/ensemble => " + str(root) + "\n")
        tidy = command(["go", "mod", "tidy"], consumer)
        if tidy["exit"]: return dict(tidy, invalid=True)
        binary = work / "public-bin"
        built = command(["go", "build", "-o", str(binary), "."], consumer)
        if built["exit"]: return dict(built, invalid=True)
        return public.owner_isolation(binary, "anthropic")
    binary = work / "cli"
    built = command(["go", "build", "-o", str(binary), "./cmd"], root)
    if built["exit"]: return dict(built, invalid=True)
    if case == "occupied": return cli.occupied_artifacts(binary, "anthropic")
    if case == "cursor": return cli.cursor_after_omission(binary, "anthropic")
    if case == "identity": return cli.completion_during_http(binary, "gemini")
    if case == "source-cap": return cli.source_cap_artifacts(binary, "anthropic")
    return scenario(binary, case)


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("source", type=pathlib.Path)
    parser.add_argument("--only", action="append")
    args = parser.parse_args()
    source = args.source.resolve(strict=True)
    records = []
    selected = [m for m in MUTATIONS if not args.only or m[0] in args.only]
    for name, case, edits in selected:
        print(name, flush=True, file=__import__("sys").stderr)
        with tempfile.TemporaryDirectory(prefix="ch04-deletion-") as tmp:
            work = pathlib.Path(tmp)
            root = work / "source"
            shutil.copytree(source, root, ignore=shutil.ignore_patterns(".git", "evidence"))
            control = run_check(root, case, work)
        with tempfile.TemporaryDirectory(prefix="ch04-deletion-") as tmp:
            work = pathlib.Path(tmp)
            root = work / "source"
            shutil.copytree(source, root, ignore=shutil.ignore_patterns(".git", "evidence"))
            for relative, before, after in edits:
                replace(root, relative, before, after)
            built = command(["go", "build", "./..."], root)
            observed = run_check(root, case, work) if built["exit"] == 0 else dict(built, invalid=True)
        valid_failure = not observed.get("invalid") and not observed.get("passed", True)
        if case.startswith("Test"):
            valid_failure = valid_failure and ("--- FAIL: " + case) in observed.get("stdout", "")
        records.append({"mutation": name, "expected_failing_check": case, "control": control,
                        "mutated": observed, "passed": bool(control.get("passed") and valid_failure)})
    print(json.dumps({"scope": __doc__, "checker_sha256": hashlib.sha256(pathlib.Path(__file__).read_bytes()).hexdigest(),
                      "passed": all(r["passed"] for r in records), "audits": records}, indent=2))
    return 0 if all(r["passed"] for r in records) else 1


if __name__ == "__main__":
    raise SystemExit(main())
