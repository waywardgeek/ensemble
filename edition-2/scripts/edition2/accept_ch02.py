#!/usr/bin/env python3
"""Independent offline CLI checks for the published second-edition Chapter 2.

Fixtures come from the contract, not student internals. This checks no paid
provider, public Go API ownership, or GUI behavior; those need separate evidence.
"""

import argparse
import copy
import hashlib
import http.server
import json
import os
import pathlib
import re
import subprocess
import tempfile
import threading


PROVENANCE = {"vendor": "anthropic", "model": "fixture-messages", "surface": "messages"}
USAGE = {"input": 10, "cache_write": 0, "cache_read": 2, "output": 3}
PRIVATE_MARKER = "DO_NOT_ECHO_MALFORMED_RECORD_271"


def event(seq, kind, payload, value):
    return {"seq": seq, "type": kind, "time": "2026-01-01T00:00:00Z", payload: value}


def human(seq, text):
    return event(seq, "message_received", "message",
                 {"actor": "human", "purpose": "dialogue", "parts": [{"type": "text", "text": text}]})


def history():
    return [
        {"log_version": 1}, human(1, "inspect config"),
        event(2, "response_ended", "response", {
            "from": copy.deepcopy(PROVENANCE), "usage": dict(USAGE),
            "parts": [{"type": "text", "text": "Inspecting."},
                      {"type": "tool_call", "call_id": "call-1", "name": "read_file",
                       "args": {"path": "config"}, "from": copy.deepcopy(PROVENANCE)}]}),
        event(3, "tool_called", "tool", {"call_id": "call-1", "name": "read_file", "args": {"path": "config"}}),
        event(4, "tool_returned", "tool", {
            "call_id": "call-1", "parts": [{"type": "text", "text": "OFFLINE_RESULT_MARKER"},
                                            {"type": "blob", "mime": "text/plain",
                                             "ref": {"kind": 2, "locator": "https://example.invalid/result"}}]}),
        human(5, "continue"),
    ]


def projection(output, vendor):
    """Read the fixture's meaningful wire positions, never arbitrary substrings."""
    body = json.loads(output)
    entries = []

    def text_content(value, separator=""):
        if isinstance(value, str):
            return value
        if not isinstance(value, list):
            raise ValueError("expected text content")
        if any(p.get("type") != "text" or not isinstance(p.get("text"), str) for p in value):
            raise ValueError("expected text-only fixture content")
        return separator.join(p["text"] for p in value)

    if vendor == "openai":
        for message in body["messages"]:
            role = message["role"]
            if role == "system":
                continue
            if role == "tool":
                entries.append(("result", message["tool_call_id"], text_content(message["content"])))
                continue
            if message.get("content"):
                entries.append(("text", role, text_content(message["content"])))
            for call in message.get("tool_calls", []):
                assert role == "assistant" and call["type"] == "function"
                fn = call["function"]
                entries.append(("call", call["id"], fn["name"], json.loads(fn["arguments"])))
    else:
        for message in body["messages" if vendor == "anthropic" else "contents"]:
            role = message["role"]
            for part in message["content" if vendor == "anthropic" else "parts"]:
                if "text" in part:
                    if vendor == "anthropic":
                        assert part["type"] == "text"
                    entries.append(("text", "assistant" if role == "model" else role, part["text"]))
                elif vendor == "anthropic" and part.get("type") == "tool_use":
                    assert role == "assistant"
                    entries.append(("call", part["id"], part["name"], part["input"]))
                elif vendor == "anthropic" and part.get("type") == "tool_result":
                    assert role == "user"
                    content = part["content"]
                    if isinstance(content, list):
                        content = "\n".join(p["text"] for p in content)
                    entries.append(("result", part["tool_use_id"], content))
                elif vendor == "gemini" and "functionCall" in part:
                    assert role == "model"
                    call = part["functionCall"]
                    entries.append(("call", call["id"], call["name"], call["args"]))
                elif vendor == "gemini" and "functionResponse" in part:
                    assert role == "user"
                    result = part["functionResponse"]
                    assert result["name"] == "read_file"
                    entries.append(("result", result["id"], result["response"]["result"]))
                else:
                    raise ValueError("unexpected fixture wire part")
    return entries


def check_projection(output, vendor, redacted=False, final_human="continue"):
    expected = [
        ("text", "user", "inspect config"), ("text", "assistant", "Inspecting."),
        ("call", "call-1", "read_file", {"path": "config"}),
        ("result", "call-1", ("[redacted]\n[redacted]" if redacted else
                               "OFFLINE_RESULT_MARKER\n[text/plain]") + " https://example.invalid/result"),
        ("text", "user", final_human),
    ]
    try:
        if projection(output, vendor) == expected:
            return []
    except (ValueError, KeyError, TypeError, AttributeError, AssertionError):
        pass
    return ["wire projection changed content, call/result linkage, roles, or ordering"]


class Harness:
    def __init__(self, binary, root):
        self.binary, self.root = binary, root
        self.requests = 0
        harness = self

        class Trap(http.server.BaseHTTPRequestHandler):
            def log_message(self, *_):
                pass

            def do_POST(self):
                harness.requests += 1
                self.send_response(500)
                self.end_headers()

            do_GET = do_POST

        self.server = http.server.ThreadingHTTPServer(("127.0.0.1", 0), Trap)
        self.thread = threading.Thread(target=lambda: self.server.serve_forever(poll_interval=0.01), daemon=True)
        self.thread.start()

    def close(self):
        self.server.shutdown()
        self.server.server_close()
        self.thread.join()

    def run(self, lines, command="dump", vendor="anthropic", model="fixture-messages", success=True):
        path = self.root / "fixture.log"
        path.write_text("\n".join(json.dumps(line) for line in lines) + "\n")
        env = {k: v for k, v in os.environ.items()
               if not k.startswith(("LLM_", "ANTHROPIC_", "OPENAI_", "GEMINI_", "CH02_", "COURSE_"))}
        env.update(CH02_LOG=str(path), LLM_VENDOR=vendor,
                   LLM_BASE_URL=f"http://127.0.0.1:{self.server.server_port}")
        if command == "render":
            env["LLM_MODEL"] = model
        args = [str(self.binary), command] + ([str(path)] if command == "render" else [])
        before, original = self.requests, path.read_bytes()
        p = subprocess.run(args, cwd=self.root, env=env, capture_output=True, text=True, timeout=10)
        errors = []
        if self.requests != before:
            errors.append("offline command attempted HTTP")
        if path.read_bytes() != original:
            errors.append("offline command changed its source log")
        if success:
            if p.returncode != 0:
                errors.append("valid fixture was rejected")
            if not p.stdout.strip():
                errors.append("valid command produced no output")
            try:
                if command == "render":
                    if not isinstance(json.loads(p.stdout), dict):
                        errors.append("render must produce one JSON request object")
                else:
                    [json.loads(line) for line in p.stdout.splitlines()]
            except ValueError:
                errors.append("offline output is not JSON")
        else:
            if p.returncode == 0 or p.stdout.strip() or not p.stderr.strip():
                errors.append("invalid fixture must fail with diagnostic and empty stdout")
            if command == "dump" and not re.search(r"\bline\s+\d+\b", p.stderr, re.I):
                errors.append("invalid log diagnostic omitted line location")
            if PRIVATE_MARKER in p.stderr or PRIVATE_MARKER in p.stdout:
                errors.append("invalid log diagnostic echoed private record content")
        return p.stdout, errors


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("binary", type=pathlib.Path)
    args = parser.parse_args()
    checks = []

    def record(name, errors):
        checks.append({"id": name, "passed": not errors, "details": errors})

    with tempfile.TemporaryDirectory(prefix="ensemble-ch02-offline-") as tmp:
        h = Harness(args.binary.resolve(strict=True), pathlib.Path(tmp))
        try:
            original = history()
            out, errors = h.run(original)
            if not errors and [json.loads(line) for line in out.splitlines()] != original:
                errors.append("dump changed the fully specified fixture facts")
            record("dump-without-model-or-key", errors)

            positives = []

            def allow(name, transform):
                rows = history()
                transform(rows)
                positives.append((name, rows))

            allow("present-empty-text", lambda r: r[1]["message"]["parts"][0].update(text=""))
            allow("zero-usage", lambda r: r[2]["response"].update(usage={k: 0 for k in USAGE}))
            allow("increasing-sequence-gaps", lambda r: [e.update(seq=e["seq"] * 10) for e in r[1:]])
            allow("unknown-metadata", lambda r: [e.update(annotation="permitted") for e in r])
            allow("legacy-missing-purpose", lambda r: r[1]["message"].pop("purpose"))
            allow("legacy-surface-alias", lambda r: [p.update(vendor="gemini", surface="generatecontent") for p in
                                                     (r[2]["response"]["from"], r[2]["response"]["parts"][1]["from"])])
            for name, rows in positives:
                _, errors = h.run(rows)
                record(name, errors)
            for kind in (1, 3):
                rows = history()
                ref = {"kind": kind, "locator": f"unresolved-reference-{kind}"}
                rows[4]["tool"]["parts"][1]["ref"] = ref
                out, errors = h.run(rows)
                if not errors and json.loads(out.splitlines()[4])["tool"]["parts"][1]["ref"] != ref:
                    errors.append("dump did not preserve reference kind and locator")
                record(f"reference-kind-{kind}-dump", errors)
                _, errors = h.run(rows, "render", success=False)
                record(f"reference-kind-{kind}-unsupported-render", errors)

            bad = []

            def change(name, transform):
                rows = history()
                transform(rows)
                bad.append((name, rows))

            change("unsupported-version", lambda r: r[0].update(log_version=999))
            change("missing-header", lambda r: r.pop(0))
            change("duplicate-sequence", lambda r: r[2].update(seq=1))
            change("zero-sequence", lambda r: r[1].update(seq=0))
            change("bad-timestamp", lambda r: r[1].update(time="yesterday"))
            change("unknown-event", lambda r: r[1].update(type="future_event"))
            change("extra-known-payload", lambda r: r[1].update(tool={"call_id": "x"}))
            change("null-payload", lambda r: r[1].update(message=None))
            change("missing-text", lambda r: r[1]["message"]["parts"][0].pop("text"))
            change("unknown-part", lambda r: r[1]["message"]["parts"][0].update(type="future_part"))
            change("agent-cannot-bypass-response-event", lambda r: r[1]["message"].update(actor="agent"))
            change("unknown-purpose", lambda r: r[1]["message"].update(purpose="recall"))
            change("zero-reference-kind", lambda r: r[4]["tool"]["parts"][1]["ref"].update(kind=0))
            change("unknown-reference-kind", lambda r: r[4]["tool"]["parts"][1]["ref"].update(kind=99))
            change("legacy-blob-path", lambda r: r[4]["tool"]["parts"].__setitem__(1, {"type": "blob", "mime": "text/plain", "path": "old"}))
            change("missing-normalized-usage", lambda r: r[2]["response"].pop("usage"))
            change("negative-usage", lambda r: r[2]["response"]["usage"].update(input=-1))
            change("fractional-usage", lambda r: r[2]["response"]["usage"].update(output=1.5))
            change("unknown-result-call", lambda r: r[4]["tool"].update(call_id="not-issued"))
            change("duplicate-call-id", lambda r: r[2]["response"]["parts"].append(copy.deepcopy(r[2]["response"]["parts"][1])))
            change("wrong-dispatch-args", lambda r: r[3]["tool"].update(args={"path": "other"}))
            bad.append(("unsolicited-response", [original[0], dict(original[2], seq=1)]))
            bad.append(("second-pending-human", [original[0], human(1, "one"), human(2, "two")]))
            bad.append(("duplicate-result", original + [event(6, "tool_returned", "tool", copy.deepcopy(original[4]["tool"]))]))
            for name, rows in bad:
                rows = copy.deepcopy(rows)
                # Unknown metadata is permitted; this also checks that errors
                # name a location without dumping the offending record.
                for row in rows:
                    row["diagnostic_private"] = PRIVATE_MARKER
                _, errors = h.run(rows, success=False)
                record(name, errors)

            for vendor, model in [("anthropic", "fixture-messages"), ("openai", "fixture-openai"), ("gemini", "fixture-gemini")]:
                out, errors = h.run(original, "render", vendor, model)
                twice, second = h.run(original, "render", vendor, model)
                errors += second
                if out != twice:
                    errors.append("repeated render changed bytes")
                errors += check_projection(out, vendor)
                record(f"{vendor}-deterministic-render", errors)

                redacted = original + [event(6, "redacted", "redact", {"from": 4, "to": 4, "level": "redact_result", "reason": "fixture"})]
                out, errors = h.run(redacted, "render", vendor, model)
                errors += check_projection(out, vendor, redacted=True)
                record(f"{vendor}-redaction", errors)

            queued = copy.deepcopy(original[:4]) + [human(4, "deferred"), dict(copy.deepcopy(original[4]), seq=5)]
            _, errors = h.run(queued)
            out, rendered = h.run(queued, "render")
            errors += rendered
            errors += check_projection(out, "anthropic", final_human="deferred")
            record("deferred-human-after-call", errors)
            _, errors = h.run(queued[:5])
            record("unanswered-call-remains-loadable", errors)
            _, errors = h.run(queued[:5], "render", success=False)
            record("unanswered-call-refuses-render", errors)
        finally:
            h.close()

    passed = all(c["passed"] for c in checks)
    print(json.dumps({"scope": "Chapter 2 offline CLI only; no live/API-ownership/GUI claim",
                      "binary_sha256": hashlib.sha256(args.binary.read_bytes()).hexdigest(),
                      "passed": passed, "checks": checks}, indent=2))
    return 0 if passed else 1


if __name__ == "__main__":
    raise SystemExit(main())
