#!/usr/bin/env python3
"""Bind the historical comparison's inspected sources; no runtime acceptance."""
import hashlib
import json
from pathlib import Path
import subprocess
import sys

ROOT = Path(__file__).resolve().parents[2]
FREEZE = "786ff239ddf337a5ebb39a12f6552d047f52c1ab"
RUNTIME = "75a72554bbb13f5e86c8b806cff780b5fcd79c22"
STANDARD = "04c6e0255e4a1a24149cc6ad72c5e494ff187227"


def blob(revision, path):
    return subprocess.check_output(["git", "show", f"{revision}:{path}"], cwd=ROOT)


def digest(data):
    return hashlib.sha256(data).hexdigest()


def main():
    binding_path = "solutions/edition-2/main/evidence/ch09/stream-binding.json"
    binding_bytes = blob(FREEZE, binding_path)
    binding = json.loads(binding_bytes)
    assert binding["source_revision"] == RUNTIME
    sources = binding["sources"]
    assert len(sources) == 134
    for path, expected in sources.items():
        assert path.startswith("solutions/edition-2/main/") and ".." not in Path(path).parts
        assert digest(blob(RUNTIME, path)) == expected, path
        assert digest(blob(FREEZE, path)) == expected, path
        assert digest((ROOT / path).read_bytes()) == expected, path
    historical = {}
    for path in ["book/chapter-10.md", "solutions/ch10/agent.go",
                 "solutions/ch10/internal/common/skill.go",
                 "solutions/ch10/internal/common/skill_registry.go",
                 "solutions/ch10/internal/common/vars.go",
                 "solutions/ch10/internal/tools/tools.go"]:
        historical[path] = digest(blob(STANDARD, path))
    corrections = {
        "98878ebb21609e716444595faf930706dc2d32f6": ["agent/internal/skills/registry.go"],
        "81a9fcd58a86d6ec2786d4a518994a01dbb553b9": ["internal/grade/ch10_harness.go"],
    }
    corrected = {rev: {path: digest(blob(rev, path)) for path in paths}
                 for rev, paths in corrections.items()}
    readings = {path: digest((ROOT / path).read_bytes()) for path in [
        "book/edition-2/skills/ensemble-coding/SKILL.md",
        "book/edition-2/architecture.md", "book/voice.md",
        "book/chapter-writing-procedure.md", "book/edition-2/chapter-09.md",
        "book/chapter-15.md", "book/chapter-21.md"]}
    receipt = {
        "scope": "Source identity for manual historical code comparison; no new runtime test or performance measurement",
        "initial_freeze": FREEZE, "runtime": RUNTIME, "standard": STANDARD,
        "binding_sha256": digest(binding_bytes), "sources": sources,
        "historical_sources": historical, "later_corrections": corrected,
        "readings_at_review": readings, "reviewer_script_sha256": digest(Path(__file__).read_bytes()),
        "verified_runtime_files": len(sources),
        "working_runtime_matches_initial_freeze": True,
        "provider_calls": 0, "runtime_builds": 0,
    }
    encoded = json.dumps(receipt, indent=2) + "\n"
    # Complete original identities are validated before this derived write.
    if len(sys.argv) == 2:
        Path(sys.argv[1]).write_text(encoded)
    else:
        print(encoded, end="")


if __name__ == "__main__":
    main()
