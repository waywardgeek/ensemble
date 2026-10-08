#!/usr/bin/env python3
"""Delete selected checker predicates, not student code; require exact refusals."""
import argparse
import hashlib
import json
from pathlib import Path
import types


def load(source, path):
    module = types.ModuleType("ch11_oracle_audit")
    module.__file__ = str(path)
    exec(compile(source, str(path), "exec"), module.__dict__)
    return module


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--receipt", type=Path)
    args = parser.parse_args()
    path = Path(__file__).resolve().with_name("accept_ch11.py")
    source = path.read_text()
    baseline = load(source, path)
    rows = baseline.self_test()
    mutations = [
        ('require(key not in result, "duplicate-key")', "duplicate-key"),
        ('require(len(raw) <= MAX_MESSAGE, "message-bytes")', "message-plus-one"),
        ('require(count <= 100000, "json-nodes")', "nodes-plus-one"),
        ('require(depth <= 64, "json-depth")', "depth-plus-one"),
        ('require(value <= watermark, "unknown-id")', "unknown-rpc-5"),
        ('require(self.pages <= 64, "discovery-pages")', "pages-plus-one"),
        ('require(tool["name"] not in self.candidate, "duplicate-tool")', "duplicate-across-pages"),
        ('require(cursor not in self.cursors, "cursor-repeat")', "cursor-repeat"),
        ('require(isinstance(params, dict) and params.get("_meta") == META, "request-metadata")', "missing-request-metadata"),
    ]
    controls = []
    for anchor, check in mutations:
        if source.count(anchor) != 1:
            raise RuntimeError("mutation anchor not unique: " + check)
        mutant = load(source.replace(anchor, "pass", 1), path)
        expected = "negative-accepted:" + check
        try:
            mutant.self_test()
        except mutant.Refusal as error:
            if error.code != expected:
                raise RuntimeError("wrong failing check: " + check + ": " + error.code) from error
        else:
            raise RuntimeError("predicate deletion survived: " + check)
        controls.append({"removed_predicate": anchor, "expected_failure": expected, "passed": True})
    if path.read_text() != source:
        raise RuntimeError("checker changed during audit")
    result = {
        "scope": "checker-oracle predicate deletions; no student runtime mutated or tested",
        "runtime_acceptance": False,
        "checker_sha256": hashlib.sha256(source.encode()).hexdigest(),
        "audit_sha256": hashlib.sha256(Path(__file__).read_bytes()).hexdigest(),
        "positive_checks": len(rows),
        "deletion_controls": controls,
        "passed": len(controls),
        "total": len(mutations),
    }
    if args.receipt:
        args.receipt.write_text(json.dumps(result, indent=2) + "\n")
    print(json.dumps(result, indent=2))


if __name__ == "__main__":
    main()
