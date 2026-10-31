#!/usr/bin/env python3
"""Partial independent Chapter 9 CLI acceptance; local endpoints, no credentials.

The oracle is derived from edition-2/chapter-09.md, not a student implementation.
Use --self-test for canned protocol positives and distinguishing oracle deletions.
Those controls test this checker, never certify a runtime implementation.
"""
import argparse
import copy
import hashlib
import http.server
import json
import os
from pathlib import Path
import subprocess
import tempfile
import threading

ROOT = Path(__file__).resolve().parents[2]
MODELS = {"anthropic": "claude-sonnet-4-6", "openai": "gpt-4.1-mini-2025-04-14", "gemini": "models/gemini-3.8-flash"}
BASE_TOOLS = ["load_skill", "read_file", "unload_skill"]
EDIT_TOOLS = ["load_skill", "read_file", "unload_skill", "write_file"]
PRIMARY = "Identity.\n- load_skill\n- read_file\n- unload_skill\n- edit: Edit scratch files\n"
MANUAL = "Use write_file.\n---\n- load_skill\n- read_file\n- unload_skill\n- write_file\n"
ENVELOPE = "[skill edit activation 2]\n" + MANUAL + "\n[/skill]"
PENDING = ["complete source/render/catalog/custom-scalar boundaries", "public owned copies, creation-only configuration and two-Agent isolation", "complete graph/discovery/diamond/shared-unload behavior", "frozen catalog lifetime after file edits/deletion", "tool_limits consumption and management argument strictness", "pause/interruption/held HTTP/admitted job concurrency", "append failure and result-after-commit failure", "strict forged replay/public-append transition refusals and result redaction", "GUI safe cards, speech, snapshot window and reusable components", "complete inherited behavior/architecture/deletion audit", "actual CLI/browser/public real-model demonstrations and historical comparison"]


def sha(data):
    return hashlib.sha256(data).hexdigest()


def definition(name, kind, body, tools="", offers="", extra=""):
    return (f"---\nname: {name}\ndescription: {name.title()} scratch files\ntype: {kind}\ntools: {tools or '[]'}\nloadable-skills: {offers or '[]'}\n{extra}---\n" + body).encode()


def catalog():
    return {"base": definition("base", "primary", "Identity.\n$TOOLS\n$SKILLS\n", "read_file", "edit"),
            "edit": definition("edit", "loadable", "Use write_file.\n---\n$TOOLS\n", "write_file"),
            "hidden": definition("hidden", "loadable", "HIDDEN-UNACTIVATED-BODY", "write_file")}


def environment(work, vendor, endpoint):
    env = {k: v for k, v in os.environ.items() if not k.startswith(("LLM_", "ANTHROPIC_", "OPENAI_", "GEMINI_", "CH02_", "EN_DISABLE_"))}
    env.update(LLM_VENDOR=vendor, LLM_MODEL=MODELS[vendor], LLM_RESOLVED_MODEL=MODELS[vendor].removeprefix("models/"),
               LLM_API_KEY="local-fixture-only", LLM_BASE_URL=endpoint, EN_DISABLE_STREAMING="1", CH02_LOG=str(work / "session.jsonl"))
    return env


def call(identity, tool_name, **arguments):
    return {"id": identity, "name": tool_name, "arguments": arguments}


def response(vendor, calls):
    model = MODELS[vendor].removeprefix("models/")
    if vendor == "anthropic":
        content = [{"type": "tool_use", "id": c["id"], "name": c["name"], "input": c["arguments"]} for c in calls] if calls else [{"type": "text", "text": "done"}]
        return {"model": model, "content": content, "stop_reason": "tool_use" if calls else "end_turn", "usage": {"input_tokens": 1, "output_tokens": 1}}
    if vendor == "openai":
        message = {"role": "assistant", "content": "" if calls else "done"}
        if calls:
            message["tool_calls"] = [{"id": c["id"], "type": "function", "function": {"name": c["name"], "arguments": json.dumps(c["arguments"])}} for c in calls]
        return {"model": model, "choices": [{"index": 0, "message": message, "finish_reason": "tool_calls" if calls else "stop"}], "usage": {"prompt_tokens": 1, "completion_tokens": 1}}
    parts = [{"functionCall": {"id": c["id"], "name": c["name"], "args": c["arguments"]}} for c in calls] if calls else [{"text": "done"}]
    return {"modelVersion": model, "candidates": [{"content": {"role": "model", "parts": parts}, "finishReason": "STOP"}], "usageMetadata": {"promptTokenCount": 1, "candidatesTokenCount": 1}}


class Fixture(http.server.ThreadingHTTPServer):
    def __init__(self, vendor, batches):
        super().__init__(("127.0.0.1", 0), Handler)
        self.vendor, self.batches, self.requests = vendor, batches, []


class Handler(http.server.BaseHTTPRequestHandler):
    def log_message(self, *_):
        pass

    def do_POST(self):
        owner = self.server
        body = json.loads(self.rfile.read(int(self.headers["Content-Length"])))
        owner.requests.append(body)
        index = len(owner.requests) - 1
        if index >= len(owner.batches):
            self.send_error(429, "local fixture request bound")
            return
        data = json.dumps(response(owner.vendor, owner.batches[index])).encode()
        self.send_response(200)
        self.send_header("Content-Type", "application/json")
        self.send_header("Content-Length", str(len(data)))
        self.end_headers()
        self.wfile.write(data)


def declarations(vendor, body):
    if vendor == "anthropic":
        return sorted(t["name"] for t in body.get("tools", []))
    if vendor == "openai":
        return sorted(t["function"]["name"] for t in body.get("tools", []))
    return sorted(d["name"] for t in body.get("tools", []) for d in t.get("functionDeclarations", []))


def system_text(vendor, body):
    if vendor == "anthropic":
        value = body.get("system", "")
        return value if isinstance(value, str) else "\n".join(p["text"] for p in value)
    if vendor == "openai":
        return "\n".join(m["content"] for m in body.get("messages", []) if m["role"] == "system")
    return "\n".join(p["text"] for p in body.get("systemInstruction", {}).get("parts", []))


def dialogue(vendor, body):
    """Normalize only text and tool-result positions needed by this contract."""
    result = []
    for message in body.get("contents" if vendor == "gemini" else "messages", []):
        if vendor == "openai":
            if message["role"] == "tool":
                result.append(("result", message["tool_call_id"]))
            elif message["role"] != "system" and isinstance(message.get("content"), str):
                result.append(("text", message["content"]))
            continue
        for part in message.get("parts" if vendor == "gemini" else "content", []):
            if "text" in part:
                result.append(("text", part["text"]))
            if part.get("type") == "tool_result":
                result.append(("result", part["tool_use_id"]))
            if "functionResponse" in part:
                result.append(("result", part["functionResponse"]["id"]))
    return result


def material_checks(vendor, requests):
    if len(requests) != 5:
        return {"declarations": False, "primary-fixed": False, "manual-after-batch": False}
    items = dialogue(vendor, requests[2])
    text = [(i, v) for i, (k, v) in enumerate(items) if k == "text" and ENVELOPE in v]
    result_positions = {v: i for i, (k, v) in enumerate(items) if k == "result"}
    return {
        "declarations": [declarations(vendor, r) for r in requests] == [BASE_TOOLS, BASE_TOOLS, EDIT_TOOLS, BASE_TOOLS, BASE_TOOLS],
        "primary-fixed": all(system_text(vendor, r) == PRIMARY for r in requests),
        "manual-after-batch": len(text) == 1 and text[0][1].count(ENVELOPE) == 1 and all(i in result_positions and result_positions[i] < text[0][0] for i in ["load", "enabled"]) and
            all(sum(v.count(ENVELOPE) for k, v in dialogue(vendor, r) if k == "text") == 1 for r in requests[2:]),
    }


def evaluate(vendor, evidence):
    events, requests, files = evidence["events"], evidence["requests"], evidence["files"]
    returns = {e["tool"]["call_id"]: e["tool"] for e in events if e.get("type") == "tool_returned"}
    changes = [e["skills"] for e in events if e.get("type") == "skills_changed"]
    initial = [e["skills"] for e in events if e.get("type") == "skills_initialized"]
    results = material_checks(vendor, requests)
    results["forced-admission"] = files == {"enabled.txt": "enabled"} and all(returns.get(i, {}).get("is_error") is True for i in ["disabled", "revoked"]) and "enabled" in returns and not returns["enabled"].get("is_error", False)
    expected = {"load": {"status": "loaded", "name": "edit", "revision": 1, "changed": True}, "unload": {"status": "unloaded", "name": "edit", "revision": 2, "changed": True}, "hidden": {"error": "skill_unavailable", "name": "hidden", "revision": 2}, "noop": {"status": "unchanged", "name": "edit", "revision": 2, "changed": False}}
    acknowledgements = True
    for identity, want in expected.items():
        record = returns.get(identity, {})
        text = "".join(p.get("text", "") for p in record.get("parts", []))
        try:
            decoded = json.loads(text)
            acknowledgements &= decoded == want and text == json.dumps(decoded, separators=(",", ":")) and bool(record.get("is_error", False)) == (identity == "hidden")
        except (ValueError, TypeError):
            acknowledgements = False
    results["management-acks"] = acknowledgements
    management = [e["tool"] for e in events if e.get("type") in ("tool_called", "tool_returned") and e.get("tool", {}).get("call_id") in expected]
    results["management-no-jobs"] = len(management) == 8 and all(not x.get("job") for x in management)
    results["transition-count-noops"] = len(initial) == 1 and [x.get("action") for x in changes] == ["load", "unload"] and [x.get("state", {}).get("revision") for x in changes] == [1, 2]
    results["material-identities"] = False
    if len(initial) == 1 and len(changes) == 2:
        first, loaded, unloaded = initial[0], changes[0], changes[1]
        records = first.get("activated", []) + loaded.get("activated", [])
        results["material-identities"] = first.get("action") == "initialize" and first.get("state", {}).get("revision") == 0 and len(records) == 2 and [r.get("activation") for r in records] == [1, 2] and [r.get("body") for r in records] == [PRIMARY, MANUAL] and all(r.get("sha256") == sha(r["body"].encode()) for r in records) and unloaded.get("activated") == [] and unloaded.get("state", {}).get("retired") == [{"name": "edit", "activation": 2}]
    return results


def capture(binary, vendor, skills=True, control=False):
    batches = [[call("ordinary", "write_file", path="enabled.txt", content="enabled")], []] if control else [
        [call("disabled", "write_file", path="disabled.txt", content="unauthorized")],
        [call("load", "load_skill", name="edit"), call("enabled", "write_file", path="enabled.txt", content="enabled")],
        [call("unload", "unload_skill", name="edit"), call("revoked", "write_file", path="revoked.txt", content="unauthorized")],
        [call("hidden", "load_skill", name="hidden"), call("noop", "unload_skill", name="edit")], []]
    server = Fixture(vendor, batches)
    thread = threading.Thread(target=server.serve_forever)
    thread.start()
    try:
        with tempfile.TemporaryDirectory(prefix="ch09-wire-") as directory:
            work = Path(directory)
            env = environment(work, vendor, f"http://127.0.0.1:{server.server_port}")
            if skills:
                for name, source in catalog().items():
                    target = work / "catalog" / name / "SKILL.md"
                    target.parent.mkdir(parents=True)
                    target.write_bytes(source)
                env.update(LLM_SKILLS_DIR="catalog", LLM_PRIMARY_SKILL="base")
            child = subprocess.run([str(binary), "protocol"], input='{"user":"Local fixture: follow the proposed operations."}\n', text=True, capture_output=True, cwd=work, env=env, timeout=25)
            path = work / "session.jsonl"
            events = [json.loads(s) for s in path.read_text().splitlines()][1:] if path.exists() else []
            result = {"exit": child.returncode, "stdout": child.stdout, "stderr": child.stderr, "requests": server.requests,
                      "events": events, "files": {p.name: p.read_text() for p in work.glob("*.txt")}}
            if not control and child.returncode == 0 and any(e.get("type") == "skills_initialized" for e in events):
                # Remove every catalog file before offline reconstruction. This
                # distinguishes recorded facts from another live catalog load.
                for source in (work / "catalog").rglob("SKILL.md"):
                    source.unlink()
                replay_env = environment(work, vendor, "http://127.0.0.1:1")
                reconstructed = []
                for event in events:
                    if event.get("type") != "request_sent":
                        continue
                    render = subprocess.run([str(binary), "replay", str(path), str(event["seq"])], cwd=work, env=replay_env, capture_output=True, text=True, timeout=10)
                    reconstructed.append(json.loads(render.stdout) if render.returncode == 0 else {"replay_error": render.stderr})
                result["offline-exact"] = reconstructed == server.requests
            return result
    finally:
        server.shutdown()
        server.server_close()
        thread.join()


def startup_cases():
    base = catalog()["base"]
    valid = {
        "scalar": base,
        "block": base.replace(b"tools: read_file", b"tools:\n  - 'read_file'").replace(b"loadable-skills: edit", b"loadable-skills:\n  - edit"),
        "crlf": base.replace(b"\n", b"\r\n"),
        "quoted": base.replace(b"name: base", b'name: "base"').replace(b"description: Base scratch files", b"description: 'Base: reader''s files'"),
    }
    invalid = {
        "duplicate-key": base.replace(b"name: base", b"name: base\nname: base"),
        "unknown-integration": base.replace(b"type: primary", b"type: primary\nmcp_servers: local"),
        "flow-list": base.replace(b"tools: read_file", b"tools: [read_file]"),
        "duplicate-item": base.replace(b"tools: read_file", b"tools: read_file read_file"),
        "invalid-id": base.replace(b"name: base", b"name: ../base"),
        "name-mismatch": base.replace(b"name: base", b"name: another"),
        "invalid-utf8": base + b"\xff",
        "bom": b"\xef\xbb\xbf" + base,
        "nul": base + b"\0",
        "bare-cr": base + b"\rX",
        "empty-primary": base.split(b"---\n", 2)[0] + base[:base.rfind(b"---\n") + 4],
        "numeric-description": base.replace(b"description: Base scratch files", b"description: 19"),
        "unknown-variable": base + b"$UNKNOWN",
        "malformed-dollar": base + b"$5",
        "source-over-limit": base + b"x" * 65536,
        "missing-dependency": base.replace(b"type: primary", b"type: primary\ndepends: absent"),
        "unknown-handler": base.replace(b"tools: read_file", b"tools: fictional_handler"),
    }
    return [(name, data, True) for name, data in valid.items()] + [(name, data, False) for name, data in invalid.items()]


def startup(binary):
    outcomes = []
    for name, source, accepted in startup_cases():
        with tempfile.TemporaryDirectory(prefix="ch09-format-") as directory:
            work = Path(directory)
            for identity, data in {**catalog(), "base": source}.items():
                target = work / "catalog" / identity / "SKILL.md"
                target.parent.mkdir(parents=True)
                target.write_bytes(data)
            env = environment(work, "anthropic", "http://127.0.0.1:1")
            env.update(LLM_SKILLS_DIR="catalog", LLM_PRIMARY_SKILL="base")
            child = subprocess.run([str(binary), "protocol"], input="", text=True, capture_output=True, cwd=work, env=env, timeout=10)
            lines = (work / "session.jsonl").read_text().splitlines() if (work / "session.jsonl").exists() else []
            initialized = sum(json.loads(s).get("type") == "skills_initialized" for s in lines)
            passed = (child.returncode == 0 and initialized == 1) if accepted else (child.returncode != 0 and bool(child.stderr.strip()) and initialized == 0)
            outcomes.append({"id": "format/" + name, "passed": passed, "expected_accept": accepted, "exit": child.returncode, "initializers": initialized, "diagnostic": child.stderr})
    return outcomes


def oracle_fixture(vendor):
    """Canned protocol facts only: deliberately not a Skills implementation."""
    def wire(names, items):
        if vendor == "anthropic":
            return {"system": PRIMARY, "tools": [{"name": n} for n in names], "messages": [{"role": "user", "content": [{"type": "text", "text": v} if k == "text" else {"type": "tool_result", "tool_use_id": v, "content": "ok"} for k, v in items]}]}
        if vendor == "openai":
            return {"tools": [{"function": {"name": n}} for n in names], "messages": [{"role": "system", "content": PRIMARY}] + [{"role": "user", "content": v} if k == "text" else {"role": "tool", "tool_call_id": v, "content": "ok"} for k, v in items]}
        return {"systemInstruction": {"parts": [{"text": PRIMARY}]}, "tools": [{"functionDeclarations": [{"name": n} for n in names]}], "contents": [{"role": "user", "parts": [{"text": v} if k == "text" else {"functionResponse": {"id": v, "name": "fixture", "response": {"result": "ok"}}} for k, v in items]}]}
    material = [("result", "load"), ("result", "enabled"), ("text", ENVELOPE)]
    requests = [wire(BASE_TOOLS, []), wire(BASE_TOOLS, []), wire(EDIT_TOOLS, material), wire(BASE_TOOLS, material), wire(BASE_TOOLS, material)]
    record = lambda n, body: {"activation": n, "body": body, "sha256": sha(body.encode())}
    events = [{"type": "skills_initialized", "skills": {"action": "initialize", "state": {"revision": 0}, "activated": [record(1, PRIMARY)]}},
              {"type": "skills_changed", "skills": {"action": "load", "state": {"revision": 1}, "activated": [record(2, MANUAL)]}},
              {"type": "skills_changed", "skills": {"action": "unload", "state": {"revision": 2, "retired": [{"name": "edit", "activation": 2}]}, "activated": []}}]
    values = {"load": {"status": "loaded", "name": "edit", "revision": 1, "changed": True}, "unload": {"status": "unloaded", "name": "edit", "revision": 2, "changed": True}, "hidden": {"error": "skill_unavailable", "name": "hidden", "revision": 2}, "noop": {"status": "unchanged", "name": "edit", "revision": 2, "changed": False}}
    for identity, value in values.items():
        events += [{"type": "tool_called", "tool": {"call_id": identity}}, {"type": "tool_returned", "tool": {"call_id": identity, "parts": [{"type": "text", "text": json.dumps(value, separators=(",", ":"))}], "is_error": identity == "hidden"}}]
    events += [{"type": "tool_returned", "tool": {"call_id": i, "is_error": i != "enabled"}} for i in ["disabled", "enabled", "revoked"]]
    return {"requests": requests, "events": events, "files": {"enabled.txt": "enabled"}}


def self_test():
    rows = []
    for vendor in MODELS:
        good = oracle_fixture(vendor)
        assert all(evaluate(vendor, good).values()), evaluate(vendor, good)
        for mutation, intended in [("effect", "forced-admission"), ("manual", "manual-after-batch"), ("job", "management-no-jobs"), ("hash", "material-identities"), ("ack", "management-acks")]:
            bad = copy.deepcopy(good)
            if mutation == "effect":
                bad["files"]["disabled.txt"] = "unauthorized"
            if mutation == "manual":
                bad["requests"][2] = json.loads(json.dumps(bad["requests"][2]).replace(ENVELOPE.replace("\n", "\\n"), "removed"))
            if mutation == "job":
                next(e["tool"] for e in bad["events"] if e.get("type") == "tool_called")["job"] = {"handle": 1}
            if mutation == "hash":
                bad["events"][0]["skills"]["activated"][0]["sha256"] = "0" * 64
            if mutation == "ack":
                next(e["tool"] for e in bad["events"] if e.get("type") == "tool_returned" and e["tool"]["call_id"] == "load")["parts"][0]["text"] = '{}'
            failures = sorted(k for k, v in evaluate(vendor, bad).items() if not v)
            assert failures == [intended], (vendor, mutation, failures)
            rows.append({"vendor": vendor, "mutation": mutation, "expected_failure": intended, "passed": True})
    return {"label": "checker canned-protocol controls, not runtime acceptance", "positive_vendors": list(MODELS), "deletions": rows}


def main():
    checker_digest = sha(Path(__file__).read_bytes())
    contract_digest = sha((ROOT / "book/edition-2/chapter-09.md").read_bytes())
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("binary", nargs="?", type=Path)
    parser.add_argument("--self-test", action="store_true")
    parser.add_argument("--receipt", type=Path)
    args = parser.parse_args()
    if args.self_test:
        result = self_test()
        code = 0
    else:
        if args.binary is None:
            parser.error("supply CLI_BINARY or --self-test")
        binary = args.binary.resolve()
        binary_digest = sha(binary.read_bytes())
        rows, runs = [], {}
        for vendor in MODELS:
            control = capture(binary, vendor, skills=False, control=True)
            runs[vendor + "-baseline"] = control
            control_ok = control["exit"] == 0 and control["files"] == {"enabled.txt": "enabled"} and len(control["requests"]) == 2 and '"assistant":"done"' in control["stdout"]
            rows.append({"id": vendor + "/baseline-local-transport", "passed": control_ok})
            if not control_ok:
                continue
            evidence = capture(binary, vendor)
            runs[vendor] = evidence
            rows += [{"id": vendor + "/" + k, "passed": v} for k, v in evaluate(vendor, evidence).items()]
            rows.append({"id": vendor + "/offline-replay", "passed": evidence.get("offline-exact", False)})
        rows += startup(binary)
        assert sha(binary.read_bytes()) == binary_digest, "CLI executable changed during acceptance"
        code = 0 if all(r["passed"] for r in rows) else 1
        result = {"scope": "initial partial Chapter 9 CLI contract acceptance", "binary": str(binary), "binary_sha256": binary_digest,
                  "checker_sha256": sha(Path(__file__).read_bytes()), "contract_sha256": sha((ROOT / "book/edition-2/chapter-09.md").read_bytes()),
                  "passed": sum(r["passed"] for r in rows), "total": len(rows), "checks": rows, "pending_material_promises": PENDING, "runs": runs}
    assert sha(Path(__file__).read_bytes()) == checker_digest, "Checker changed during acceptance"
    assert sha((ROOT / "book/edition-2/chapter-09.md").read_bytes()) == contract_digest, "Chapter contract changed during acceptance"
    result.update(checker_sha256=checker_digest, contract_sha256=contract_digest)
    if args.receipt:
        args.receipt.write_text(json.dumps(result, indent=2) + "\n")
    print(json.dumps({k: v for k, v in result.items() if k != "runs"}, indent=2))
    return code


if __name__ == "__main__":
    raise SystemExit(main())
