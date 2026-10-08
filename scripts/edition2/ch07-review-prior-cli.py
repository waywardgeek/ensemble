#!/usr/bin/env python3
"""Chapter 7 layout adapter for unchanged Chapter 6 CLI bridge assertions.

Supports both the original cmd package and the extracted public cli package.
The historical checker/probe remain unchanged; only placement is adapted.

Exercises the real CLI loops with a deterministic blocked writer and actual
HTTP/Agent/Observer path. Complements the separate real PTY delivery checks;
this injected-writer fixture is not itself a PTY or a live-model demonstration.
"""
import argparse
import hashlib
import io
import json
from pathlib import Path
import subprocess
import tarfile
import tempfile


HERE = Path(__file__).resolve().parent
REPO = HERE.parents[1]
PREFIX = "solutions/edition-2/main/"
PROBE = HERE / "ch06-review-cli-recovery_test.go.txt"
MUTANTS = {
    "suppress-gap": (
        "cmd/progress.go", "return gap()", "return nil",
        {"TestReviewCLIRecovery/human=false/fragments=300",
         "TestReviewCLIRecovery/human=true/fragments=300"},
        "missing single"),
    "suppress-human-recovery": (
        "cmd/chat.go", "progress.reported || !progress.streamed[c.RequestID]",
        "!progress.streamed[c.RequestID]",
        {"TestReviewCLIRecovery/human=true/fragments=300"},
        "missing labeled complete recovered human answer"),
}


def run(cmd, cwd):
    completed = subprocess.run(cmd, cwd=cwd, capture_output=True, text=True, timeout=120)
    return {"command": cmd, "exit": completed.returncode,
            "stdout": completed.stdout, "stderr": completed.stderr}


def evaluate(revision):
    revision = subprocess.check_output(
        ["git", "rev-parse", revision + "^{commit}"], cwd=REPO, text=True).strip()
    archive = subprocess.check_output(["git", "archive", revision, PREFIX], cwd=REPO)
    checks = []
    hashes = {}
    fixtures = {}
    for name in ["positive", *MUTANTS]:
        with tempfile.TemporaryDirectory(prefix="ch07-prior-cli-") as directory:
            temporary = Path(directory)
            # Preserve committed source assets needed by embeds as well as Go files.
            # Raw evidence is excluded; original checkers are never edited.
            with tarfile.open(fileobj=io.BytesIO(archive)) as source:
                for member in source:
                    if not member.isfile():
                        continue
                    path = Path(member.name)
                    relative = path.relative_to(PREFIX)
                    is_source = path.suffix == ".go" or path.name in ("go.mod", "go.sum")
                    if "evidence" in relative.parts or "node_modules" in relative.parts:
                        continue
                    if ".." in relative.parts:
                        raise ValueError("unsafe archive path")
                    data = source.extractfile(member).read()
                    dest = temporary / relative
                    dest.parent.mkdir(parents=True, exist_ok=True)
                    dest.write_bytes(data)
                    if name == "positive":
                        (hashes if is_source else fixtures)[str(relative)] = hashlib.sha256(data).hexdigest()
            client_dir = "cli" if (temporary / "cli/session.go").exists() else "cmd"
            probe = temporary / client_dir / "review_cli_recovery_test.go"
            body = PROBE.read_text()
            if client_dir == "cli":
                if not body.startswith("package main\n"):
                    raise ValueError("historical probe package anchor changed")
                body = body.replace("package main\n", "package cli\n", 1)
            probe.write_text(body)
            mutation = None
            if name != "positive":
                file, old, new, expected, reason = MUTANTS[name]
                file = file.replace("cmd/", client_dir + "/", 1)
                path = temporary / file
                text = path.read_text()
                if text.count(old) != 1:
                    raise ValueError(f"{name}: mutation anchor not unique")
                path.write_text(text.replace(old, new))
                mutation = {"file": file, "old": old, "new": new}
            format_check = run(["gofmt", "-l", str(probe)], temporary)
            if format_check["exit"] or format_check["stdout"]:
                raise ValueError("unformatted review probe: " + json.dumps(format_check))
            vet = run(["go", "vet", "./..." if name == "positive" else "./" + client_dir], temporary)
            if vet["exit"]:
                raise ValueError("probe or mutation failed vet: " + json.dumps(vet))
            result = run(["go", "test", "-race", "-json", "-count=1", "-timeout=60s",
                          "-run", "^TestReviewCLIRecovery$", "./" + client_dir], temporary)
            records = [json.loads(line) for line in result["stdout"].splitlines()]
            failures = {r["Test"] for r in records if r.get("Action") == "fail" and "Test" in r}
            leaf_failures = {test for test in failures if test.count("/") == 2}
            ran = {r["Test"] for r in records if r.get("Action") == "run" and r.get("Test", "").count("/") == 2}
            expected_runs = {f"TestReviewCLIRecovery/human={human}/fragments={count}"
                             for human in ("false", "true") for count in (2, 300)}
            passed = ran == expected_runs
            if name == "positive":
                passed = passed and result["exit"] == 0 and not failures
            else:
                passed = passed and result["exit"] != 0 and leaf_failures == expected and reason in result["stdout"]
                passed = passed and "DATA RACE" not in result["stdout"] + result["stderr"]
            module_tests = None
            if name == "positive":
                module_tests = run(["go", "test", "./...", "-count=1"], temporary)
                passed = passed and module_tests["exit"] == 0
            checks.append({"id": name, "passed": passed, "mutation": mutation,
                           "ran": sorted(ran), "leaf_failures": sorted(leaf_failures),
                           "vet": vet, "module_tests": module_tests, "result": result})
    return {"scope": __doc__, "source_revision": revision, "source_files": hashes, "testdata": fixtures,
            "checkers": {str(p.relative_to(REPO)): hashlib.sha256(p.read_bytes()).hexdigest()
                         for p in (Path(__file__), PROBE)},
            "passed": all(check["passed"] for check in checks), "checks": checks}


if __name__ == "__main__":
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("revision", help="immutable student source commit")
    args = parser.parse_args()
    receipt = evaluate(args.revision)
    print(json.dumps(receipt, indent=2))
    raise SystemExit(0 if receipt["passed"] else 1)
