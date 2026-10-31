#!/usr/bin/env python3
"""Delete real Chapter 1 behaviors in disposable copies; check exact failures.

This audit targets the new reference snapshot, not arbitrary student naming.
The black-box acceptor itself remains implementation-independent. Anchor
changes require deliberate maintenance; never silently skip a missing mutant.
"""

import argparse
import json
import os
import pathlib
import shutil
import subprocess
import sys
import tempfile


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("solution", type=pathlib.Path)
    args = parser.parse_args()
    acceptor = pathlib.Path(__file__).with_name("accept_ch01.py").resolve()
    cases = [
        ("control", None, None, None, []),
        ("accept-missing-key", "ensemble.go", 'if config.APIKey == "" {',
         "if false {", ["missing-key"]),
        ("accept-missing-model", "ensemble.go", 'if config.Model == "" {',
         "if false {", ["missing-model"]),
        ("forget-history", "ensemble.go", "a.history = append(a.history, user, assistant)",
         "/* deleted history retention */", ["exact-growth"]),
        ("forget-input-usage", "internal/llm/engine.go", "e.usage.Input += usage.Input",
         "/* deleted input accounting */", ["blank-lines", "exact-growth", "nontext-block"]),
        ("forget-output-usage", "internal/llm/engine.go", "e.usage.Output += usage.Output",
         "/* deleted output accounting */", ["blank-lines", "exact-growth", "nontext-block"]),
        ("first-text-only", "internal/llm/engine.go", "answer.WriteString(block.Text)",
         "answer.WriteString(block.Text); break",
         ["blank-lines", "exact-growth", "late-failure", "nontext-block", "zero-usage"]),
        ("allow-negative-usage", "internal/llm/engine.go",
         " || *response.Usage.Input < 0 || *response.Usage.Output < 0", "", ["negative"]),
        ("accept-http-error", "internal/llm/engine.go",
         "if response.StatusCode != http.StatusOK {", "if false {", ["http", "late-failure"]),
        ("accept-empty-answer", "internal/llm/engine.go", "if answer.Len() == 0 {",
         "if false {", ["empty-answer"]),
        ("reject-blank-lines", "cmd/main.go", 'if strings.TrimSpace(string(line)) == "" {',
         "if false {", ["blank-lines"]),
        ("ignore-malformed-input", "cmd/main.go", 'return fmt.Errorf("malformed input JSON")',
         "continue", ["bad-input"]),
    ]
    results = []
    for name, filename, old, new, expected in cases:
        with tempfile.TemporaryDirectory(prefix="edition2-ch01-mutation-") as tmp:
            root = pathlib.Path(tmp) / "solution"
            shutil.copytree(args.solution, root,
                            ignore=shutil.ignore_patterns(".git", "evidence", "__pycache__"))
            if filename:
                path = root / filename
                source = path.read_text()
                if source.count(old) != 1:
                    raise RuntimeError(f"{name}: mutation anchor is not unique; update deliberately")
                path.write_text(source.replace(old, new, 1))
            binary = pathlib.Path(tmp) / "student"
            env = dict(os.environ, GOCACHE="/tmp/ensemble-edition2-go-cache")
            build = subprocess.run(["go", "build", "-o", str(binary), "./cmd"], cwd=root,
                                   env=env, capture_output=True, text=True, timeout=120)
            if build.returncode:
                raise RuntimeError(f"{name}: mutant does not build; no behavior was tested\n{build.stderr}")
            run = subprocess.run([sys.executable, str(acceptor), str(binary), "--skip-timeout"],
                                 capture_output=True, text=True, timeout=120)
            report = json.loads(run.stdout)
            failed = sorted(c["id"] for c in report["checks"] if not c["passed"])
            passed = failed == sorted(expected) and run.returncode == (1 if expected else 0)
            results.append({"mutation": name, "passed": passed,
                            "expected": sorted(expected), "actual": failed})
    print(json.dumps({"scope": "reference deletion audit; timeout separately tested",
                      "passed": all(r["passed"] for r in results), "mutations": results}, indent=2))
    return 0 if all(r["passed"] for r in results) else 1


if __name__ == "__main__":
    raise SystemExit(main())
