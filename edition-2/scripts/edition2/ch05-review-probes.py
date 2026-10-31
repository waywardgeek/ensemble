#!/usr/bin/env python3
"""Run narrow Chapter 5 comparative-review probes in a disposable source copy."""
import argparse
import hashlib
import json
import os
import pathlib
import shutil
import subprocess
import tempfile

p = argparse.ArgumentParser(description=__doc__)
p.add_argument("source", type=pathlib.Path)
p.add_argument("--small-input", action="store_true")
p.add_argument("--run", default="^TestReview")
args = p.parse_args()
source = args.source.resolve(strict=True)
here = pathlib.Path(__file__).resolve().parent
with tempfile.TemporaryDirectory(prefix="ensemble-ch05-review-") as directory:
    target = pathlib.Path(directory) / "source"
    shutil.copytree(source, target, ignore=shutil.ignore_patterns("evidence", ".git"))
    hashes = {str(f.relative_to(source)): hashlib.sha256(f.read_bytes()).hexdigest()
              for f in sorted(source.rglob("*.go")) if "evidence" not in f.relative_to(source).parts}
    for name in ("send", "subscriptions"):
        fixture = target / ("review_" + name + "_test.go")
        fixture.write_bytes((here / ("ch05-review-" + name + "_test.go.txt")).read_bytes())
        subprocess.run(["gofmt", "-w", str(fixture)], check=True)
    env = os.environ.copy()
    if args.small_input:
        env["ENSEMBLE_REVIEW_SMALL_INPUT"] = "1"
    command = ["go", "test", "-race", "-run", args.run, "-count=1", "-timeout=20s", "."]
    result = subprocess.run(command, cwd=target, env=env, text=True, capture_output=True, timeout=60)
    fixtures = {f.name: hashlib.sha256(f.read_bytes()).hexdigest()
                for f in (here / "ch05-review-send_test.go.txt", here / "ch05-review-subscriptions_test.go.txt")}
    print(json.dumps(dict(source=str(source), source_files=hashes, fixtures=fixtures, small_input=args.small_input,
                         command=command, exit=result.returncode, stdout=result.stdout, stderr=result.stderr), indent=2))
    raise SystemExit(result.returncode)
