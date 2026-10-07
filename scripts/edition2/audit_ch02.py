#!/usr/bin/env python3
"""Scoped Chapter 2 deletion audit in disposable reference-solution copies.

These source anchors deliberately target this reference snapshot. The CLI
acceptor does not depend on student internals. Build failures never count as
detected behavioral defects; changed anchors require explicit maintenance.
"""

import argparse
import hashlib
import json
import pathlib
import shutil
import subprocess
import sys
import tempfile


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("solution", type=pathlib.Path)
    args = parser.parse_args()
    acceptor = pathlib.Path(__file__).with_name("accept_ch02.py").resolve()
    redactions = [f"{v}-redaction" for v in ("anthropic", "openai", "gemini")]
    cases = [
        ("control", None, None, None, []),
        ("ignore-version", "internal/eventlog/log.go", "h.Version != 1",
         "false", ["unsupported-version"]),
        ("ignore-sequence-order", "internal/llm/events.go", "e.Seq <= c.LastSeq",
         "false", ["duplicate-sequence"]),
        ("allow-zero-sequence", "internal/llm/events.go", "e.Seq == 0 || e.Seq <= c.LastSeq",
         "(e.Seq != 0 && e.Seq <= c.LastSeq)", ["zero-sequence"]),
        ("ignore-timestamp-error", "internal/llm/events.go", "err != nil || offset != 0",
         "(err != nil && false) || offset != 0", ["bad-timestamp"]),
        ("render-unanswered-call", "internal/llm/render.go", "if unresolved(owner, c) {",
         "if false {", ["unanswered-call-refuses-render"]),
        ("retain-redacted-content", "internal/llm/events.go",
         'p.Parts[k] = common.Part{Type: "redacted", Stub: "[redacted]", Ref: old.Ref}',
         "p.Parts[k] = old", redactions),
        ("omit-redaction-stub", "internal/llm/events.go", 'Stub: "[redacted]"',
         'Stub: ""', redactions),
        ("break-messages-result-pairing", "internal/llm/render.go", '"tool_use_id": p.CallID',
         '"tool_use_id": "wrong"', ["anthropic-deterministic-render", "anthropic-redaction", "deferred-human-after-call"]),
        ("break-chat-result-pairing", "internal/llm/render.go", '"tool_call_id": part.CallID',
         '"tool_call_id": "wrong"', ["openai-deterministic-render", "openai-redaction"]),
        ("break-gemini-result-pairing", "internal/llm/render.go", '"name": call.Part.Name',
         '"name": call.Part.Name + "wrong"', ["gemini-deterministic-render", "gemini-redaction"]),
    ]
    results = []
    # Freeze one source copy so a concurrent student revision cannot change
    # the control halfway through the audit.
    with tempfile.TemporaryDirectory(prefix="edition2-ch02-audit-") as tmp:
        baseline = pathlib.Path(tmp) / "baseline"
        shutil.copytree(args.solution, baseline,
                        ignore=shutil.ignore_patterns(".git", "evidence", "__pycache__"))
        hashes = {str(p.relative_to(baseline)): hashlib.sha256(p.read_bytes()).hexdigest()
                  for p in sorted(baseline.rglob("*"))
                  if p.is_file() and (p.suffix == ".go" or p.name in ("go.mod", "go.sum"))}
        for name, filename, old, new, expected in cases:
            root = pathlib.Path(tmp) / name
            shutil.copytree(baseline, root)
            if filename:
                path = root / filename
                source = path.read_text()
                if source.count(old) != 1:
                    raise RuntimeError(f"{name}: anchor is not unique; update deliberately")
                path.write_text(source.replace(old, new, 1))
            binary = root / "student"
            build = subprocess.run(["go", "build", "-o", str(binary), "./cmd"],
                                   cwd=root, capture_output=True, text=True, timeout=120)
            if build.returncode:
                raise RuntimeError(f"{name}: build failed; no behavior tested\n{build.stderr}")
            run = subprocess.run([sys.executable, str(acceptor), str(binary)],
                                 capture_output=True, text=True, timeout=120)
            report = json.loads(run.stdout)
            actual = sorted(c["id"] for c in report["checks"] if not c["passed"])
            passed = actual == sorted(expected) and run.returncode == (1 if expected else 0)
            results.append({"mutation": name, "passed": passed,
                            "expected": sorted(expected), "actual": actual})
    passed = all(r["passed"] for r in results)
    print(json.dumps({"scope": "partial reference deletion audit of offline CLI",
                      "source_sha256": hashes,
                      "passed": passed, "mutations": results}, indent=2))
    return 0 if passed else 1


if __name__ == "__main__":
    raise SystemExit(main())
