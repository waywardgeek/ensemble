#!/usr/bin/env python3
"""Independent local review of the initial Chapter 6 live receipts. No HTTP."""
import copy
import hashlib
import importlib.util
import json
from pathlib import Path
import shutil
import subprocess
import sys
import tempfile

REPO = Path(__file__).resolve().parents[2]
HERE = REPO / "solutions/edition-2/main/evidence/ch06"


def sha(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()


def read(path):
    return json.loads(path.read_text())


def run():
    sys.path.insert(0, str(HERE))
    spec = importlib.util.spec_from_file_location("reviewed_verifier", HERE / "verify-receipts.py")
    verifier = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(verifier)
    binding = read(HERE / "initial-binding.json")
    archived = read(REPO / "book/edition-2/checkpoint-evidence/executable-archive.json")["files"]
    paths = {name: next(x["archive_path"] for x in archived if x["sha256"] == binding["executables"][name]["sha256"])
             for name in ("cli", "consumer")}
    runs = sorted(p.parent for p in HERE.glob("*/launch.json"))
    assert len(runs) == 9
    # Bind originals and assert that replay/control exercises leave them intact.
    originals = {str(p.relative_to(HERE)): sha(p) for root in runs for p in root.rglob("*") if p.is_file()}
    controls, reviewed = [], []
    with tempfile.TemporaryDirectory(prefix="ch06-receipt-review-") as directory:
        temporary = Path(directory)
        results = verifier.verify(binding, runs, paths, temporary / "positive")
        assert sum(r["requests"] for r in results) == 33
        controls.append({"id": "all-nine-valid-original-runs", "passed": True, "requests": 33})
        # Valid copies retain all original logs and bodies. Mutate only the
        # intended identity; late-batch failures must precede any output write.
        copies = []
        for source in runs:
            dest = temporary / source.name
            shutil.copytree(source, dest)
            copies.append(dest)
        verifier.verify(binding, copies, paths, temporary / "valid-copy-control")
        controls.append({"id": "all-nine-valid-copy-paths-before-mutation", "passed": True, "requests": 33})
        mutations = [
            ("source-hash", "historical source mismatch"),
            ("incomplete-source", "incomplete historical source set"),
            ("empty-source", "empty source set"),
            ("cli-hash", "executable mismatch"),
            ("consumer-hash", "executable mismatch"),
            ("interpreter-hash", "interpreter mismatch"),
            ("recorder-hash", "executable mismatch"),
            ("support-hash", "support identity mismatch"),
            ("missing-support", "incomplete support identities"),
            ("last-launch-source", "launch source mismatch"),
            ("last-launch-binary", "launch executable mismatch"),
            ("last-launch-support", "launch support mismatch"),
            ("last-launch-role", "unknown launch executable"),
        ]
        last = copies[-1] / "launch.json"
        pristine = last.read_bytes()
        for name, reason in mutations:
            candidate = copy.deepcopy(binding)
            launch = json.loads(pristine)
            if name == "source-hash":
                candidate["sources"][next(iter(candidate["sources"]))] = "0" * 64
            elif name == "incomplete-source":
                candidate["sources"].pop(next(iter(candidate["sources"])))
            elif name == "empty-source":
                candidate["sources"] = {}
            elif name in ("cli-hash", "consumer-hash", "interpreter-hash", "recorder-hash"):
                candidate["executables"][name.removesuffix("-hash")]["sha256"] = "0" * 64
            elif name == "support-hash":
                candidate["support"]["evidence.py"] = "0" * 64
            elif name == "missing-support":
                candidate["support"].pop("evidence.py")
            elif name == "last-launch-source":
                launch["source_revision"] = "wrong"
            elif name == "last-launch-binary":
                launch["executables"]["consumer"] = "0" * 64
            elif name == "last-launch-support":
                launch["support"]["terminal-run.py"] = "0" * 64
            elif name == "last-launch-role":
                launch["launched_executable"] = "unknown"
            last.write_text(json.dumps(launch))
            output = temporary / name
            try:
                verifier.verify(candidate, copies, paths, output)
                raise RuntimeError("identity mutation accepted: " + name)
            except AssertionError as error:
                assert reason in str(error), (name, str(error))
                assert not output.exists(), name + " wrote derived files before refusal"
                controls.append({"id": name, "passed": True, "refusal": str(error)})
            finally:
                last.write_bytes(pristine)
        body_path = sorted((copies[-1] / "requests").glob("*.json"))[-1]
        original_body = body_path.read_bytes()
        changed_body = json.loads(original_body)
        changed_body["model"] = "review-intended-body-mismatch"
        body_path.write_text(json.dumps(changed_body))
        output = temporary / "last-request-body-mismatch"
        try:
            verifier.verify(binding, copies, paths, output)
            raise RuntimeError("late request body mismatch accepted")
        except AssertionError as error:
            assert "request reconstruction mismatch" in str(error)
            assert not output.exists()
            controls.append({"id": "last-request-body-mismatch", "passed": True, "refusal": str(error)})
        finally:
            body_path.write_bytes(original_body)
        for root in runs:
            launch = read(root / "launch.json")
            assert launch["exit_code"] == 0
            logs = sorted(root.glob("*.log"))
            expected_logs = 2 if launch["mode"] == "consumer" else 1
            assert len(logs) == expected_logs
            bodies = [read(p) for p in sorted((root / "requests").glob("*.json"))]
            assert len(bodies) == launch["requests"]
            assert len(list((root / "responses").glob("*.body"))) == len(bodies)
            events = [[json.loads(line) for line in log.read_text().splitlines()][1:] for log in logs]
            for history in events:
                assert [e["seq"] for e in history] == list(range(1, len(history) + 1))
                assert all(e["request"]["delivery"] == launch["delivery"] for e in history if e["type"] == "request_sent")
            usage = {key: sum(e["response"]["usage"][key] for history in events for e in history if e["type"] == "response_ended")
                     for key in ("input", "cache_write", "cache_read", "output")}
            record = {"run": root.name, "requests": len(bodies), "logs": len(logs), "accepted_usage": usage}
            terminal = (root / "terminal.txt").read_text()
            if launch["mode"] == "chat":
                history = events[0]
                requests = [e for e in history if e["type"] == "request_sent"]
                # The student's verifier compares multisets even for one Agent.
                # Here chronological replay must match captured wire order too.
                rendered = [json.loads(subprocess.check_output([paths["cli"], "replay", str(logs[0]), str(e["seq"])], cwd=root)) for e in requests]
                assert rendered == bodies, root.name + " chronological replay mismatch"
                hints = [e for e in history if e["type"] == "hint_received"]
                hint_receipts = []
                for hint in hints:
                    text, seq = hint["hint"]["text"], hint["seq"]
                    consumes = [i for i, e in enumerate(requests) if seq in e["request"]["hints"]]
                    assert len(consumes) == 1
                    index = consumes[0]
                    assert requests[index]["seq"] > seq
                    assert all(text not in json.dumps(body) for body in bodies[:index])
                    messages = bodies[index].get("messages", bodies[index].get("contents"))
                    matching = [m for m in messages if text in json.dumps(m)]
                    assert len(matching) == 1 and matching[0]["role"] == "user"
                    if launch["vendor"] == "gemini":
                        assert {"text": text} in matching[0]["parts"]
                        assert text not in json.dumps(bodies[index].get("systemInstruction"))
                    assert all(text not in json.dumps(body) for body in bodies[index + 1:])
                    hint_receipts.append({"received_seq": seq, "consumed_seq": requests[index]["seq"], "wire_request": index + 1})
                called = [e for e in history if e["type"] == "tool_called"]
                returned = [e for e in history if e["type"] == "tool_returned"]
                assert len(called) == len(returned) == 1
                assert called[0]["tool"]["name"] == "read_file"
                assert returned[0]["tool"]["call_id"] == called[0]["tool"]["call_id"]
                content = (root / "workspace/notes.txt").read_bytes()
                artifact = root / "workspace" / returned[0]["tool"]["job"]["output"]["locator"]
                assert artifact.read_bytes() == content
                assert returned[0]["tool"]["parts"] == [{"type": "text", "text": content.decode()}]
                assert "CHAPTER-SIX-FILE-MARKER" in terminal and "8080" in terminal
                ending = f"Final usage: input={usage['input']}, cache write={usage['cache_write']}, cache read={usage['cache_read']}, output={usage['output']}"
                assert ending in terminal
                if launch["delivery"] == "stream":
                    assert len(hints) == 1
                    assert "interrupted=true" in terminal and "Incomplete display:" in terminal
                    assert "proposed tool" in terminal and "RECOVERED-SIX" in terminal
                    interrupted = [e for e in history if e["type"] == "turn_ended" and e["turn"]["outcome"] == "interrupted"]
                    assert len(interrupted) == 1
                    end = interrupted[0]
                    start = next(e for e in history if e["type"] == "turn_started" and e["turn"]["request_id"] == end["turn"]["request_id"])
                    assert not any(e["type"] in ("response_ended", "tool_called") for e in history if start["seq"] < e["seq"] < end["seq"])
                    assert "thinking part" not in terminal and "observation overflow" not in terminal
                    record["interrupted_request"] = end["turn"]["request_id"]
                else:
                    assert not hints and "Assistant:\nPLAIN-SIX\n" in terminal
                    assert "proposed tool" not in terminal and "(stream)" not in terminal
                record.update(chronological_replay=True, hints=hint_receipts, read_artifact_sha256=sha(artifact))
            else:
                public = json.loads(terminal)
                assert public["completion_before_slow_release"] is True
                assert public["slow_subscription_status"] == ["", ""]
                identities = set()
                for completion, observed, finals, history in zip(public["completions"], public["observations"], public["finals_only"], events, strict=True):
                    assert completion["outcome"] == "success"
                    assert finals == [o for o in observed if o["kind"] == "part_final"]
                    assert [o["part"] for o in finals] == completion["parts"]
                    assert all(o["agent_id"] == completion["agent_id"] for o in observed)
                    assert all(o["request_id"] == completion["request_id"] for o in finals)
                    assert not any(o.get("channel") == "thinking" for o in observed)
                    for o in finals:
                        identities.add((o["agent_id"], o["request_id"], o["operation_id"], o["part_id"]))
                        accepted = next(e for e in history if e["seq"] == o["response_seq"])
                        assert accepted["response"]["parts"][o["part_index"]] == o["part"]
                    assert completion["usage"] == next(e["response"]["usage"] for e in history if e["type"] == "response_ended")
                assert len(public["completions"]) == 2
                assert len({c["agent_id"] for c in public["completions"]}) == 2
                record.update(distinct_full_part_identities=len(identities), thinking_deltas=0, live_overflow=False)
            reviewed.append(record)
    assert all(sha(HERE / name) == identity for name, identity in originals.items())
    # Compare secret bytes only in memory. Neither the settings nor values enter
    # the output, subprocess arguments, exceptions or retained evidence.
    settings = read(Path.home() / ".cr/settings.json")
    keys = [settings[name].encode() for name in ("directClaudeAPIKey", "directOpenAIAPIKey", "directGeminiAPIKey")]
    files = [p for p in HERE.rglob("*") if p.is_file() and "__pycache__" not in p.parts]
    assert not any(key and key in p.read_bytes() for p in files for key in keys), "credential match in evidence"
    return {"passed": True, "source_revision": binding["source_revision"], "historical_source_files": len(binding["sources"]),
            "archived_executables": paths, "checker_sha256": sha(Path(__file__)), "controls": controls,
            "runs": reviewed, "original_receipt_hashes": originals, "originals_unchanged": True,
            "credential_scan_files": len(files), "credential_matches": 0,
            "limits": "Local replay/control and receipt consistency audit; no new model calls. Public callback-release ordering additionally inspected in archived consumer source. No live thinking or overflow claimed."}


if __name__ == "__main__":
    print(json.dumps(run(), indent=2))
