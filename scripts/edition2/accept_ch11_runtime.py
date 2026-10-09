#!/usr/bin/env python3
"""Partial Chapter 11 CLI runtime checks. Local model fixture; no paid calls.

See chapter-11-grader-review.md for the published build-binding contract and
the complete remaining §11.10 map. This command never claims full acceptance.
"""
import argparse
import copy
import hashlib
import json
import os
from pathlib import Path
import signal
import subprocess
import sys
import tempfile
import threading
import time

from accept_ch09 import Fixture, call, declarations
from accept_ch11 import request

ROOT = Path(__file__).resolve().parents[2]
PREFIX = "solutions/edition-2/main/"
PEER = Path(__file__).with_name("ch11_effect_peer.py")
TEXT = "checked\nline"
EXPECTED = b'{"content":[{"text":"checked\\nline","type":"text"}],"isError":false,"structuredContent":{"written":12}}\n'
REMAINING = {
    "transport_independence": "memory/custom public adapters, ownership/copies and same-suite transport proof",
    "protocol_bounds": "pagination, malformed/unknown/stale/reverse messages and complete exact/+1 resource matrix",
    "authority": "two-Agent grants, Skills unload/admitted lifetime and durable/artifact failure zero-send",
    "schema_results": "complete schema profile/precision/budgets, isError and full result boundaries",
    "lifecycle": "permits, cancellation/races, stalled I/O, generations, uint64 seams and joined shared shutdown",
    "persistence": "v1/v2 identities, pre-prepare validation, snapshot/tail and replay without effects",
    "existing_behavior": "retained chapter gates, report limits, hints/interrupt and all three continuations",
    "public_usability": "human CLI, GUI/watch, custom headless real-model runs and immutable snapshots",
    "quality": "structure review, runtime deletion controls, race/delivered-module checks and historical comparison",
}


def sha(data):
    return hashlib.sha256(data).hexdigest()


def git(*args):
    return subprocess.check_output(["git", "-C", str(ROOT), *args])


def source_file(path):
    return Path(path).suffix in {".go", ".js", ".html", ".css"} or Path(path).name in {"go.mod", "go.sum"}


def bind(cli, source, revision, association_path):
    """Reuse the Chapter 10 build-association/build-inputs receipt format.

    A receipt is an auditable build assertion, not cryptographic proof that a
    binary came from source. Verify its complete map, pinned bytes and binary.
    """
    revision = git("rev-parse", "--verify", revision + "^{commit}").decode().strip()
    association = json.loads(association_path.read_bytes())
    if association.get("source_revision") != revision or association.get("before_after_source_equal") is not True:
        raise ValueError("build association source revision/stability")
    if not association.get("build_commands"):
        raise ValueError("build association lacks actual build command references")
    inputs_name = association.get("inputs_receipt", "")
    if not inputs_name or Path(inputs_name).name != inputs_name:
        raise ValueError("build inputs must be a sibling receipt basename")
    inputs_path = association_path.with_name(inputs_name)
    inputs = json.loads(inputs_path.read_bytes())
    mapping = inputs.get("source_sha256")
    if inputs.get("source_revision") != revision or not isinstance(mapping, dict) or not mapping:
        raise ValueError("empty/incompatible build source map")
    tracked = git("ls-tree", "-r", "--name-only", revision, "--", PREFIX).decode().splitlines()
    required = {p for p in tracked if source_file(p) and "evidence" not in Path(p[len(PREFIX):]).parts}
    if not required or not required <= mapping.keys():
        raise ValueError("incomplete build source map")
    for name, digest in mapping.items():
        relative = name.removeprefix(PREFIX)
        if not name.startswith(PREFIX) or ".." in Path(relative).parts or Path(relative).is_absolute():
            raise ValueError("build source map path")
        expected = sha(git("show", revision + ":" + name))
        if expected != digest:
            raise ValueError("build source map differs from pinned revision: " + name)
        path = source / relative
        if path.is_symlink() or sha(path.read_bytes()) != expected:
            raise ValueError("local source differs from pinned revision: " + relative)
    local = {PREFIX + str(p.relative_to(source)) for p in source.rglob("*")
             if p.is_file() and source_file(p) and "evidence" not in p.relative_to(source).parts}
    if local != required:
        raise ValueError("local source set differs from pinned revision")
    binary = sha(cli.read_bytes())
    if association.get("binaries", {}).get("cli", {}).get("sha256") != binary:
        raise ValueError("CLI binary differs from build association")
    return {"source_revision": revision, "source_directory": str(source), "source_sha256": mapping,
            "cli": str(cli), "cli_sha256": binary, "build_binding": str(association_path),
            "build_binding_sha256": sha(association_path.read_bytes()),
            "build_inputs_sha256": sha(inputs_path.read_bytes())}


def identities():
    paths = [Path(__file__), PEER, Path(__file__).with_name("accept_ch09.py"),
             Path(__file__).with_name("accept_ch11.py"), ROOT / "book/edition-2/chapter-11.md"]
    return {str(p.relative_to(ROOT)): sha(p.read_bytes()) for p in paths}


def read_records(path):
    if not path.exists():
        return []
    raw = path.read_bytes()
    try:
        return [json.loads(line) for line in raw.splitlines()]
    except (ValueError, UnicodeError) as error:
        return [{"capture_error": str(error), "raw_hex": raw.hex()}]


def cleanup_peer(rows, work):
    """Never signal an arbitrary PID: require our exact peer and unique workspace.

    Forced cleanup is retained and fails normal-close coverage; it is not a pass.
    """
    result = []
    for row in rows:
        if not row.get("startup"):
            continue
        pid = row["pid"]
        observed = subprocess.run(["ps", "-p", str(pid), "-o", "command="], capture_output=True, text=True)
        if observed.returncode != 0:
            continue
        command = observed.stdout.strip()
        result.append({"pid": pid, "command": command, "forced": False})
        if str(PEER) in command and str(work) in command:
            os.kill(pid, signal.SIGKILL)
            result[-1]["forced"] = True
    return result


def capture(cli, case):
    arguments = {"text": 7} if case == "invalid-arguments" else {"text": TEXT}
    server = Fixture("anthropic", [[call("remote-1", "append_note", **arguments)], []])
    thread = threading.Thread(target=lambda: server.serve_forever(poll_interval=.01))
    thread.start()
    try:
        with tempfile.TemporaryDirectory(prefix="ch11-cli-") as directory:
            work = Path(directory)
            config_dir = work / "config"
            config_dir.mkdir()
            # An actual relative executable path; neither PATH lookup nor shell.
            (config_dir / "python").symlink_to(Path(sys.executable).resolve())
            peer_args = [str(PEER), str(work)] + (["--mixed"] if case == "mixed-result" else [])
            connection = {"key": "catalog", "transport": "stdio", "command": "./python", "args": peer_args,
                          "cwd": ".", "env_allowlist": ["CH11_ALLOWED"], "call_timeout_seconds": 5}
            unused = dict(connection, key="unused", command="./must-not-launch")
            config = {"version": 1, "connections": [connection, unused],
                      "bindings": [{"alias": "append_note", "connection": "catalog", "remote_name": "notes.append"}]}
            if case == "missing-reference":
                config["bindings"][0]["connection"] = "absent"
            config_path = config_dir / "mcp.json"
            config_path.write_text(json.dumps(config))
            # Do not inherit credentials, runtime flags or host user config.
            env = {"PATH": "/usr/bin:/bin", "HOME": str(work), "TMPDIR": str(work),
                   "LLM_VENDOR": "anthropic", "LLM_MODEL": "claude-sonnet-4-6", "LLM_API_KEY": "fixture-only",
                   "LLM_BASE_URL": f"http://127.0.0.1:{server.server_port}", "EN_DISABLE_STREAMING": "1",
                   "CH02_LOG": str(work / "session.jsonl"), "CH11_ALLOWED": "allowlisted-value",
                   "CH11_UNLISTED": "must-not-inherit"}
            argv = [str(cli), "protocol", "--mcp-config", str(config_path)]
            started = time.monotonic()
            child = subprocess.Popen(argv, stdin=subprocess.PIPE, stdout=subprocess.PIPE, stderr=subprocess.PIPE,
                                     cwd=work, env=env, text=True, start_new_session=True)
            timeout = False
            try:
                stdout, stderr = child.communicate('{"user":"Append the literal note through append_note."}\n', timeout=40)
            except subprocess.TimeoutExpired:
                timeout = True
                os.killpg(child.pid, signal.SIGKILL)
                stdout, stderr = child.communicate(timeout=5)
            peers = read_records(work / "peer.jsonl")
            cleanup = cleanup_peer(peers, work)
            files = {str(p.relative_to(work)): p.read_bytes().hex() for p in work.glob("cr/io/*") if p.is_file()}
            notebook = work / "notebook.txt"
            result = {"case": case, "argv": argv, "cwd": str(work), "config": config,
                      "stdin": '{"user":"Append the literal note through append_note."}\n',
                      "exit": child.returncode, "timeout": timeout, "elapsed": time.monotonic() - started,
                      "stdout": stdout, "stderr": stderr, "requests": copy.deepcopy(server.requests),
                      "peer": peers, "peer_cleanup": cleanup, "artifacts_hex": files,
                      "notebook_hex": notebook.read_bytes().hex() if notebook.exists() else None,
                      "events": read_records(work / "session.jsonl"), "workspace_removed_after_capture": True}
            return result
    finally:
        server.shutdown()
        server.server_close()
        thread.join()


def evaluate(run):
    case = run["case"]
    sent = [json.loads(row["request"]) for row in run["peer"] if "request" in row]
    returns = [e["tool"] for e in run["events"] if e.get("type") == "tool_returned" and e.get("tool", {}).get("call_id") == "remote-1"]
    common = run["exit"] == 0 and not run["timeout"] and not run["peer_cleanup"] and not any("capture_error" in r for r in run["events"] + run["peer"])
    if case == "missing-reference":
        return {"missing-reference-before-launch": run["exit"] != 0 and not run["timeout"] and not run["peer"] and not run["requests"] and not run["artifacts_hex"]}
    discovery = [request(1, "server/discover"), request(2, "tools/list")]
    if case == "invalid-arguments":
        return {"invalid-arguments-zero-send": common and sent == discovery and run["notebook_hex"] is None and
                len(returns) == 1 and returns[0].get("is_error") is True and not run["artifacts_hex"]}
    expected_call = request(3, "tools/call", name="notes.append", arguments={"text": TEXT})
    checks = {"literal-discover-list-call": common and sent == discovery + [expected_call],
              "actual-notebook-effect-once": run["notebook_hex"] == TEXT.encode().hex()}
    if case == "mixed-result":
        files = [bytes.fromhex(v) for v in run["artifacts_hex"].values()]
        want = b'{"error":"mcp_unsupported_content","message":"Remote result contains unsupported content."}\n'
        checks["mixed-result-whole-refusal"] = common and files == [want] and len(returns) == 1 and returns[0].get("is_error") is True
        return checks
    startup = [r for r in run["peer"] if r.get("startup")]
    checks["relative-path-unused-entry-allowlist"] = len(startup) == 1 and startup[0]["cwd"] == str(Path(run["cwd"]) / "config") and startup[0]["allowed"] == "allowlisted-value" and not startup[0]["credential_present"] and not startup[0]["unlisted_present"]
    checks["canonical-artifact-and-job"] = len(returns) == 1 and not returns[0].get("is_error", False) and returns[0].get("job", {}).get("status") == "done" and run["artifacts_hex"] == {returns[0].get("job", {}).get("output", {}).get("locator", ""): EXPECTED.hex()}
    requests = run["requests"]
    continuations = [part for body in requests[1:] for message in body.get("messages", []) for part in message.get("content", []) if isinstance(part, dict) and part.get("type") == "tool_result" and part.get("tool_use_id") == "remote-1"]
    def result_text(part):
        content = part.get("content", [])
        return content if isinstance(content, str) else "".join(p.get("text", "") for p in content if isinstance(p, dict))
    checks["declaration-and-recorded-continuation"] = len(requests) == 2 and "append_note" in declarations("anthropic", requests[0]) and len(continuations) == 1 and EXPECTED.decode() in result_text(continuations[0])
    checks["metadata-excluded"] = common and bool(returns) and any("CH11_IGNORED_METADATA" in r.get("response", "") for r in run["peer"]) and "CH11_IGNORED_METADATA" not in json.dumps([requests, run["events"], run["stdout"], run["stderr"]])
    return checks


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--cli", required=True, type=Path)
    parser.add_argument("--source", required=True, type=Path)
    parser.add_argument("--source-commit", required=True)
    parser.add_argument("--build-binding", required=True, type=Path)
    parser.add_argument("--receipt", required=True, type=Path)
    args = parser.parse_args()
    if args.receipt.exists():
        parser.error("receipt must not exist; preserve earlier attempts")
    identity = bind(args.cli.resolve(), args.source.resolve(), args.source_commit, args.build_binding.resolve())
    checker = identities()
    runs, rows = [], []
    parent = capture(args.cli.resolve(), "valid")
    runs.append(parent)
    parent_checks = evaluate(parent)
    rows.extend({"id": "valid/" + name, "status": "pass" if passed else "fail"} for name, passed in parent_checks.items())
    for case in ("invalid-arguments", "mixed-result", "missing-reference"):
        if not all(parent_checks.values()):
            rows.append({"id": case, "status": "blocked", "blocked_by": "valid parent"})
            continue
        captured = capture(args.cli.resolve(), case)
        runs.append(captured)
        rows.extend({"id": case + "/" + name, "status": "pass" if passed else "fail"} for name, passed in evaluate(captured).items())
    if bind(args.cli.resolve(), args.source.resolve(), args.source_commit, args.build_binding.resolve()) != identity or identities() != checker:
        raise ValueError("identity changed during run; no acceptance receipt written")
    receipt = {"runtime_acceptance": False, "scope": "partial CLI/stdout stdio client runtime subset, Anthropic local model fixture",
               "binding": identity, "checker_sha256": checker, "rows": rows, "runs": runs, "remaining": REMAINING}
    with args.receipt.open("x") as output:
        json.dump(receipt, output, indent=2)
        output.write("\n")
    counts = {s: sum(r["status"] == s for r in rows) for s in ("pass", "fail", "blocked")}
    print(json.dumps({"receipt": str(args.receipt), "counts": counts, "full_acceptance": False}))
    return int(any(r["status"] != "pass" for r in rows))


if __name__ == "__main__":
    sys.exit(main())
