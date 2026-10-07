#!/usr/bin/env python3
"""Chapter 3 public CLI acceptance, independently derived from the new contract.

No student source is inspected. All HTTP is local and scripted. This does not
prove public API capability isolation, owner chains, persistence fault handling,
or live usability; those remain separate gates. Ordinary tool output has no
fixed grammar, so textual diagnostics use broad semantic markers. A disputed
presentation needs reviewer inspection, not a new unpublished output format.
"""

import argparse
import hashlib
import http.server
import json
import os
import pathlib
import re
import signal
import subprocess
import tempfile
import threading


VENDORS = ("anthropic", "openai", "gemini")
MODEL = "fixture-ch03"
TOOLS = sorted(("read_file", "list_directory", "search_files", "write_file",
                "edit_file", "run_command"))
FINAL = "FINAL-ANSWER-937"
NARRATION = "INTERMEDIATE-NARRATION-418"
USAGE = {"input": 5, "cache_write": 0, "cache_read": 0, "output": 2}


def call(identifier, name, **args):
    return {"id": identifier, "name": name, "args": args}


def response(vendor, calls, text=FINAL):
    """Literal wire fixtures; no student serialization or model tables."""
    if vendor == "anthropic":
        parts = [{"type": "text", "text": text}]
        parts += [{"type": "tool_use", "id": c["id"], "name": c["name"],
                   "input": c["args"]} for c in calls]
        return {"id": "message-fixture", "type": "message", "role": "assistant",
                "model": MODEL, "content": parts,
                "stop_reason": "tool_use" if calls else "end_turn",
                "usage": {"input_tokens": 5, "output_tokens": 2}}
    if vendor == "openai":
        message = {"role": "assistant", "content": text}
        if calls:
            message["tool_calls"] = [
                {"id": c["id"], "type": "function", "function": {
                    "name": c["name"], "arguments": json.dumps(c["args"])}}
                for c in calls]
        return {"id": "chat-fixture", "object": "chat.completion", "model": MODEL,
                "choices": [{"index": 0, "message": message,
                             "finish_reason": "tool_calls" if calls else "stop"}],
                "usage": {"prompt_tokens": 5, "completion_tokens": 2, "total_tokens": 7}}
    parts = [{"text": text}]
    parts += [{"functionCall": {"id": c["id"], "name": c["name"], "args": c["args"]},
               "thoughtSignature": "fixture-signature-" + c["id"]} for c in calls]
    return {"modelVersion": MODEL, "candidates": [{"content": {"role": "model", "parts": parts},
                                                   "finishReason": "STOP"}],
            "usageMetadata": {"promptTokenCount": 5, "candidatesTokenCount": 2,
                              "totalTokenCount": 7}}


def text_content(value):
    if isinstance(value, str):
        return value
    if isinstance(value, list):
        if any(not isinstance(p, dict) or p.get("type") != "text" or
               not isinstance(p.get("text"), str) for p in value):
            raise ValueError("expected textual result parts")
        return "".join(p["text"] for p in value)
    raise ValueError("expected result text")


def wire_facts(body, vendor):
    """Decode actual semantic positions; a marker elsewhere cannot pass."""
    calls, results, signatures = [], [], {}
    if vendor == "openai":
        for m in body["messages"]:
            for c in m.get("tool_calls", []):
                if m["role"] != "assistant" or c["type"] != "function":
                    raise ValueError("wrong call role/type")
                calls.append((c["id"], c["function"]["name"], json.loads(c["function"]["arguments"])))
            if m["role"] == "tool":
                results.append((m["tool_call_id"], text_content(m["content"]), None))
    else:
        for m in body["messages" if vendor == "anthropic" else "contents"]:
            for p in m["content" if vendor == "anthropic" else "parts"]:
                if vendor == "anthropic" and p.get("type") == "tool_use":
                    if m["role"] != "assistant":
                        raise ValueError("wrong call role")
                    calls.append((p["id"], p["name"], p["input"]))
                elif vendor == "anthropic" and p.get("type") == "tool_result":
                    if m["role"] != "user":
                        raise ValueError("wrong result role")
                    results.append((p["tool_use_id"], text_content(p["content"]), p.get("is_error", False)))
                elif vendor == "gemini" and "functionCall" in p:
                    if m["role"] != "model":
                        raise ValueError("wrong call role")
                    c = p["functionCall"]
                    calls.append((c["id"], c["name"], c["args"]))
                    signatures[c["id"]] = p.get("thoughtSignature")
                elif vendor == "gemini" and "functionResponse" in p:
                    if m["role"] != "user":
                        raise ValueError("wrong result role")
                    r = p["functionResponse"]
                    if (r["id"], r["name"]) not in [(c[0], c[1]) for c in calls]:
                        raise ValueError("result function name does not match its call")
                    failure = "error" in r["response"]
                    value = r["response"]["error" if failure else "result"]
                    if not isinstance(value, str):
                        raise ValueError("expected fixture result text")
                    results.append((r["id"], value, failure))
    return calls, results, signatures


def declaration_errors(body, vendor):
    try:
        if vendor == "openai":
            declarations = [(t["function"]["name"], t["function"]["parameters"])
                            for t in body["tools"] if t["type"] == "function"]
        elif vendor == "anthropic":
            declarations = [(t["name"], t["input_schema"]) for t in body["tools"]]
        else:
            declarations = [(t["name"], t["parametersJsonSchema"])
                            for group in body["tools"] for t in group["functionDeclarations"]]
        if [name for name, _ in declarations] != TOOLS:
            return ["six declarations absent, duplicated, or not in stable name order"]
        if any(not isinstance(schema, dict) or schema.get("type") != "object" or schema.get("additionalProperties") is not False
               for _, schema in declarations):
            return ["declarations lost object/unknown-field constraints"]
    except (KeyError, TypeError, ValueError):
        return ["invalid declaration wire structure"]
    return []


class Session:
    def __init__(self, binary, root, vendor, replies):
        self.binary, self.root, self.vendor, self.replies = binary, root, vendor, replies
        self.requests, self.server_errors = [], []
        self.log = root / "events.jsonl"
        session = self

        class Handler(http.server.BaseHTTPRequestHandler):
            def log_message(self, *_):
                pass

            def do_POST(self):
                try:
                    body = json.loads(self.rfile.read(int(self.headers.get("Content-Length", "0"))))
                    session.requests.append(body)
                    index = len(session.requests) - 1
                    reply = session.replies[index] if index < len(session.replies) else None
                    status = 503 if reply is None else 200
                    payload = {"error": {"message": "FAKE-FAILURE-PRIVATE-182"}} if reply is None else reply
                except (ValueError, TypeError) as err:
                    session.server_errors.append(type(err).__name__)
                    status, payload = 400, {"error": "bad fixture request"}
                encoded = json.dumps(payload).encode()
                self.send_response(status)
                self.send_header("Content-Type", "application/json")
                self.send_header("Content-Length", str(len(encoded)))
                self.end_headers()
                self.wfile.write(encoded)

        self.server = http.server.ThreadingHTTPServer(("127.0.0.1", 0), Handler)
        self.thread = threading.Thread(target=lambda: self.server.serve_forever(poll_interval=0.01), daemon=True)
        self.thread.start()

    def close(self):
        self.server.shutdown()
        self.server.server_close()
        self.thread.join()

    def invoke(self, args=(), input_text=None, offline=False):
        env = {k: v for k, v in os.environ.items()
               if not k.startswith(("LLM_", "ANTHROPIC_", "OPENAI_", "GEMINI_", "CH02_", "COURSE_"))}
        env.update(LLM_VENDOR=self.vendor, LLM_MODEL=MODEL, LLM_RESOLVED_MODEL=MODEL,
                   LLM_BASE_URL=f"http://127.0.0.1:{self.server.server_port}", CH02_LOG=str(self.log))
        if not offline:
            env["LLM_API_KEY"] = "LOCAL-FIXTURE-KEY"
        # Kill the child process group on timeout, including a defective tool's shell.
        p = subprocess.Popen([str(self.binary), *args], cwd=self.root, env=env,
                             stdin=subprocess.PIPE, stdout=subprocess.PIPE,
                             stderr=subprocess.PIPE, text=True, start_new_session=True)
        try:
            stdout, stderr = p.communicate(input_text, timeout=15)
        except subprocess.TimeoutExpired:
            os.killpg(p.pid, signal.SIGKILL)
            stdout, stderr = p.communicate()
            raise TimeoutError("CLI exceeded bounded fixture time; process group terminated")
        return p.returncode, stdout, stderr


def records(path):
    rows = [json.loads(line) for line in path.read_text().splitlines() if line.strip()]
    if not rows or rows[0].get("log_version") != 1:
        raise ValueError("missing log header")
    return rows[1:]


def result_map(rows):
    return {row["tool"]["call_id"]: row["tool"] for row in rows if row.get("type") == "tool_returned"}


def result_text(results, identifier):
    return "\n".join(p["text"] for p in results[identifier]["parts"] if p.get("type") == "text")


def truncation(text):
    return bool(re.search(r"truncat|omitt|not shown|cut off", text, re.I))


def require(condition, message, errors):
    if not condition:
        errors.append(message)


def check_turn(session, batches, success=True, final=FINAL):
    rc, stdout, stderr = session.invoke(input_text='{"user":"Execute the fixture plan."}\n')
    errors = list(session.server_errors)
    require((rc == 0) == success, "unexpected CLI success/failure status", errors)
    rows = records(session.log)
    output = [json.loads(line) for line in stdout.splitlines() if line.strip()]
    answers = [r["assistant"] for r in output if "assistant" in r]
    usage = [r["usage"] for r in output if "usage" in r]
    require(answers == ([final] if success else []), "CLI must emit only the completed final answer", errors)
    require(len(usage) == (1 if success else 0), "wrong success-usage line count", errors)
    accepted = len(batches) + (1 if success else 0)
    if success and usage:
        require(usage[0] == {k: v * accepted for k, v in USAGE.items()},
                "usage did not count every accepted response exactly once", errors)
    if not success:
        require(bool(stderr.strip()), "failed turn needs a diagnostic", errors)
    expected_requests = accepted + (1 if not success and len(batches) < 16 else 0)
    require(len(session.requests) == expected_requests, "wrong request count or automatic retry", errors)
    responses = [r for r in rows if r.get("type") == "response_ended"]
    require(sum(r.get("type") == "message_received" and r.get("message", {}).get("actor") == "human"
                for r in rows) == 1, "continuation invented another human input", errors)
    require(len(responses) == accepted, "accepted responses were lost or duplicated", errors)
    for r in responses:
        require(r["response"].get("usage") == USAGE, "persisted response usage changed", errors)
    for r in responses[:len(batches)]:
        require(any(p.get("type") == "text" and p.get("text") == NARRATION
                    for p in r["response"].get("parts", [])), "intermediate narration lost from history", errors)
    calls = [c for batch in batches for c in batch]
    events = [r for r in rows if r.get("type") in ("tool_called", "tool_returned")]
    expected_events = [(kind, c["id"]) for c in calls for kind in ("tool_called", "tool_returned")]
    require([(r["type"], r["tool"]["call_id"]) for r in events] == expected_events,
            "calls/results lost, duplicated, or dispatched out of order", errors)
    results = result_map(rows)
    for i, body in enumerate(session.requests):
        errors += declaration_errors(body, session.vendor)
        previous = [c for batch in batches[:i] for c in batch]
        got_calls, got_results, signatures = wire_facts(body, session.vendor)
        require(got_calls == [(c["id"], c["name"], c["args"]) for c in previous],
                f"request {i + 1} changed previous call identities/arguments/order", errors)
        require([r[0] for r in got_results] == [c["id"] for c in previous],
                f"request {i + 1} omitted/reordered result IDs", errors)
        for identifier, text, failed in got_results:
            expected_failure = results[identifier].get("is_error", False)
            recorded_text = result_text(results, identifier)
            require(text == recorded_text or (expected_failure and recorded_text in text),
                    "wire result lost or changed recorded content", errors)
            if session.vendor != "openai":
                require(failed == expected_failure, "wire error status disagrees with recorded result", errors)
            elif expected_failure:
                require(bool(re.search(r"error|fail|refus|unavailable|unknown|invalid", text, re.I)),
                        "OpenAI error result lacks an explicit failure description", errors)
            if session.vendor == "gemini":
                require(signatures.get(identifier) == "fixture-signature-" + identifier,
                        "Gemini bound signature was lost or moved", errors)
    require("FAKE-FAILURE-PRIVATE-182" not in stdout + stderr, "provider error body leaked", errors)
    return rows, results, errors


def expect_result(results, identifier, failed, errors, includes=(), excludes=()):
    value = results.get(identifier)
    require(value is not None, f"missing result {identifier}", errors)
    if value is None:
        return ""
    require(value.get("is_error", False) == failed, f"wrong error status for {identifier}", errors)
    text = result_text(results, identifier)
    for marker in includes:
        require(marker in text, f"{identifier} missing {marker!r}", errors)
    for marker in excludes:
        require(marker not in text, f"{identifier} contains excluded {marker!r}", errors)
    return text


def edit_case(root):
    for name in ("ambiguous", "missing", "empty-anchor", "unique", "delete", "noop"):
        (root / name).write_text("first anchor\nsecond anchor\n")
    (root / "guard").write_text("KEEP-ORIGINAL")
    (root / "replace").write_text("OLD")
    (root / "append").write_text("HEAD")
    return [call("ambiguous", "edit_file", path="ambiguous", old_text="anchor", new_text="BAD"),
            call("missing", "edit_file", path="missing", old_text="absent", new_text="BAD"),
            call("empty-anchor", "edit_file", path="empty-anchor", old_text="", new_text="BAD"),
            call("unique", "edit_file", path="unique", old_text="first anchor", new_text="first edit"),
            call("delete", "edit_file", path="delete", old_text="first anchor\n", new_text=""),
            call("noop", "edit_file", path="noop", old_text="first anchor", new_text="first anchor"),
            call("guard", "write_file", path="guard", content="BAD"),
            call("replace", "write_file", path="replace", content="NEW", overwrite=True),
            call("append", "write_file", path="append", content="TAIL", append=True),
            call("new", "write_file", path="parents/new", content="CREATED"),
            call("empty", "write_file", path="empty", content=""),
            call("append-new", "write_file", path="append-new", content="FIRST", append=True)]


def check_edits(root, results, errors):
    for identifier in ("ambiguous", "missing", "empty-anchor", "guard"):
        expect_result(results, identifier, True, errors)
    for identifier in ("unique", "delete", "noop", "replace", "append", "new", "empty", "append-new"):
        expect_result(results, identifier, False, errors)
    for name, number in (("ambiguous", 2), ("missing", 0)):
        require(bool(re.search(rf"\b{number}\b", result_text(results, name))),
                f"{name} refusal did not report observed match count", errors)
    require(bool(re.search(r"no.?op|unchanged|identical|no change|same", result_text(results, "noop"), re.I)),
            "successful no-op was not reported as such", errors)
    guard = result_text(results, "guard")
    require("guard" in guard and bool(re.search(r"\b13\b", guard)),
            "overwrite refusal omitted target or existing size", errors)
    for name, expected in {"ambiguous": "first anchor\nsecond anchor\n", "missing": "first anchor\nsecond anchor\n",
                           "empty-anchor": "first anchor\nsecond anchor\n", "unique": "first edit\nsecond anchor\n",
                           "delete": "second anchor\n", "noop": "first anchor\nsecond anchor\n", "guard": "KEEP-ORIGINAL",
                           "replace": "NEW", "append": "HEADTAIL", "parents/new": "CREATED", "empty": "",
                           "append-new": "FIRST"}.items():
        require((root / name).exists() and (root / name).read_text() == expected,
                f"incorrect actual disk bytes for {name}", errors)


def read_case(root):
    (root / "notes.md").write_text("alpha line one\nbeta line two\ngamma line three\ndelta line four\nepsilon line five\n")
    (root / "empty").write_text("")
    (root / "cap").write_text("~" * 32)
    return [call("range", "read_file", path="notes.md", start_line=3, end_line=4),
            call("clamp", "read_file", path="notes.md", start_line=5, end_line=999),
            call("empty", "read_file", path="empty"), call("empty-zero", "read_file", path="empty", end_line=0),
            call("empty-range", "read_file", path="empty", start_line=1),
            call("beyond", "read_file", path="notes.md", start_line=6),
            call("exact", "read_file", path="cap", max_bytes=32),
            call("omit", "read_file", path="cap", max_bytes=16)]


def check_reads(root, results, errors):
    expect_result(results, "range", False, errors, ("gamma line three", "delta line four"), ("alpha line one", "epsilon line five"))
    expect_result(results, "clamp", False, errors, ("epsilon line five",), ("alpha line one",))
    expect_result(results, "empty", False, errors)
    expect_result(results, "empty-zero", False, errors)
    expect_result(results, "empty-range", True, errors)
    expect_result(results, "beyond", True, errors)
    for name, count, cut in (("exact", 32, False), ("omit", 16, True)):
        text = expect_result(results, name, False, errors)
        require(text.count("~") == count, f"{name} read retained wrong content byte count", errors)
        require(truncation(text) == cut, f"{name} read has incorrect truncation notice", errors)


def search_case(root):
    read_case(root)  # Plant independently of any write tool.
    (root / "tree/.git").mkdir(parents=True)
    (root / "tree/a.txt").write_text("NEEDLE-VISIBLE\n")
    (root / "tree/b.txt").write_bytes(b"\x00NEEDLE-BINARY\n")
    (root / "tree/c.md").write_text("NEEDLE-FILTERED\n")
    (root / "tree/.git/hidden.txt").write_text("NEEDLE-GIT\n")
    (root / "outside").mkdir()
    (root / "outside/linked.txt").write_text("NEEDLE-LINK\n")
    (root / "search-cap.txt").write_text("%" * 100 + "\n")
    (root / "tree/link").symlink_to(root / "outside", target_is_directory=True)
    return [call("context", "search_files", path="notes.md", pattern="gamma", context_lines=1),
            call("merge", "search_files", path="notes.md", pattern="beta|delta", context_lines=1),
            call("groups", "search_files", path="notes.md", pattern="alpha|epsilon", context_lines=1),
            call("exact", "search_files", path="notes.md", pattern="alpha|epsilon", max_matches=2),
            call("omit", "search_files", path="notes.md", pattern="alpha|epsilon", max_matches=1),
            call("none", "search_files", path="notes.md", pattern="NO-SUCH-MATCH"),
            call("bad-regex", "search_files", path="notes.md", pattern="["),
            call("bad-glob", "search_files", path="notes.md", pattern="alpha", file_pattern="["),
            call("tree", "search_files", path="tree", pattern="NEEDLE", file_pattern="*.txt"),
            call("byte-cap", "search_files", path="search-cap.txt", pattern="%", max_bytes=40)]


def check_search(root, results, errors):
    context = expect_result(results, "context", False, errors,
                            ("notes.md-2-beta line two", "notes.md:3:gamma line three", "notes.md-4-delta line four"),
                            ("alpha line one", "epsilon line five"))
    require(context.find("beta line two") < context.find("gamma line three") < context.find("delta line four"),
            "search context reordered lines", errors)
    merged = expect_result(results, "merge", False, errors, ("notes.md:2:beta line two", "notes.md:4:delta line four"))
    require(not re.search(r"(?m)^--\s*$", merged), "adjacent windows were not merged", errors)
    require(merged.count("gamma line three") == 1, "merged context duplicated or lost shared line", errors)
    groups = expect_result(results, "groups", False, errors, ("alpha line one", "epsilon line five"), ("gamma line three",))
    require(bool(re.search(r"(?m)^--\s*$", groups)), "separated groups lack separator", errors)
    exact = expect_result(results, "exact", False, errors, ("alpha line one", "epsilon line five"))
    omitted = expect_result(results, "omit", False, errors, ("alpha line one",), ("epsilon line five",))
    require(not truncation(exact) and truncation(omitted), "match cap confuses exact completion with omission", errors)
    expect_result(results, "none", False, errors)
    expect_result(results, "bad-regex", True, errors)
    expect_result(results, "bad-glob", True, errors)
    expect_result(results, "tree", False, errors, ("NEEDLE-VISIBLE",),
                  ("NEEDLE-BINARY", "NEEDLE-FILTERED", "NEEDLE-GIT", "NEEDLE-LINK"))
    capped = expect_result(results, "byte-cap", False, errors)
    require(capped.count("%") <= 40 and truncation(capped), "search byte cap absent or unreported", errors)


def command_case(root):
    return [call("exit7", "run_command", command="exit 7"),
            call("streams", "run_command", command="printf OUT-MARKER; printf ERR-MARKER >&2"),
            call("exact", "run_command", command="printf '~~~~~~~~~~~~~~~~'", max_output_bytes=16),
            call("drain", "run_command", command="awk 'BEGIN {for(i=0;i<200000;i++) printf \"~\"}'; "
                 "awk 'BEGIN {for(i=0;i<200000;i++) printf \"^\"}' >&2; printf FINISHED > drained", max_output_bytes=16)]


def check_commands(root, results, errors):
    exit7 = expect_result(results, "exit7", False, errors)
    require(bool(re.search(r"(?:exit|status)[^\n0-9]{0,30}7\b", exit7, re.I)), "silent command did not report exit status 7", errors)
    streams = expect_result(results, "streams", False, errors, ("OUT-MARKER", "ERR-MARKER"))
    require(bool(re.search("stdout", streams, re.I)) and bool(re.search("stderr", streams, re.I)),
            "command output does not identify the two streams", errors)
    exact = expect_result(results, "exact", False, errors)
    require(exact.count("~") == 16 and not truncation(exact), "exact command cap falsely truncates or loses bytes", errors)
    drain = expect_result(results, "drain", False, errors)
    require(drain.count("~") == 16 and drain.count("^") == 16 and truncation(drain),
            "command cap is not applied separately to both streams", errors)
    require((root / "drained").exists() and (root / "drained").read_text() == "FINISHED",
            "capped process did not drain and finish", errors)


def unicode_case(root):
    # Added after the initial run exposed JSON's replacement of a sliced rune;
    # the author published this explicit boundary before the student revision.
    (root / "utf8.txt").write_text("éX", encoding="utf-8")
    return [call(f"read-{n}", "read_file", path="utf8.txt", max_bytes=n) for n in (1, 2, 3)] + [
        call(f"command-{n}", "run_command", command="printf 'éX'", max_output_bytes=n) for n in (1, 2, 3)]


def check_unicode(root, results, errors):
    for kind in ("read", "command"):
        for n in (1, 2, 3):
            text = expect_result(results, f"{kind}-{n}", False, errors)
            require("\ufffd" not in text, f"{kind}-{n} manufactured a replacement character", errors)
            require(("é" in text) == (n >= 2) and ("X" in text) == (n == 3),
                    f"{kind}-{n} retained the wrong complete UTF-8 prefix", errors)
            require(truncation(text) == (n < 3), f"{kind}-{n} has wrong omission notice", errors)


def invalid_case(root):
    (root / "valid").write_text("AFTER-ERROR-READ")
    return [call("unknown", "unavailable_tool", path="never-created"),
            call("type", "read_file", path=42), call("unknown-field", "read_file", path="valid", invented=True),
            call("zero", "read_file", path="valid", max_bytes=0),
            call("fraction", "read_file", path="valid", start_line=1.5),
            call("negative", "read_file", path="valid", end_line=-1),
            call("after", "read_file", path="valid")]


def check_invalid(root, results, errors):
    for identifier in ("unknown", "type", "unknown-field", "zero", "fraction", "negative"):
        expect_result(results, identifier, True, errors)
    for identifier, field in (("type", "path"), ("unknown-field", "invented"),
                              ("zero", "max_bytes"), ("fraction", "start_line"), ("negative", "end_line")):
        text = result_text(results, identifier)
        require("read_file" in text and field in text, f"{identifier} diagnostic omitted tool/field", errors)
    expect_result(results, "after", False, errors, ("AFTER-ERROR-READ",))
    require(not (root / "never-created").exists(), "unknown tool caused a side effect", errors)


def list_case(root):
    (root / "listing/subdir").mkdir(parents=True)
    (root / "listing/a-entry").write_text("x")
    (root / "listing/b-entry").write_text("x")
    (root / "listing/subdir/MUST-NOT-RECURSE").write_text("x")
    (root / "two").mkdir()
    for name in ("a-entry", "b-entry"):
        (root / "two" / name).write_text("x")
    (root / "long-list").mkdir()
    (root / "long-list" / ("%" * 100)).write_text("x")
    return [call("list", "list_directory", path="listing"),
            call("exact", "list_directory", path="two", max_entries=2),
            call("omit", "list_directory", path="two", max_entries=1),
            call("not-directory", "list_directory", path="two/a-entry"),
            call("missing", "list_directory", path="absent"),
            call("byte-cap", "list_directory", path="long-list", max_bytes=40)]


def check_list(root, results, errors):
    text = expect_result(results, "list", False, errors, ("a-entry", "b-entry", "subdir"), ("MUST-NOT-RECURSE",))
    require(text.find("a-entry") < text.find("b-entry") < text.find("subdir"), "listing not in name order", errors)
    exact = expect_result(results, "exact", False, errors, ("a-entry", "b-entry"))
    omitted = expect_result(results, "omit", False, errors, ("a-entry",), ("b-entry",))
    require(not truncation(exact) and truncation(omitted), "entry cap confuses exact completion with omission", errors)
    expect_result(results, "not-directory", True, errors)
    expect_result(results, "missing", True, errors)
    capped = expect_result(results, "byte-cap", False, errors)
    require(capped.count("%") <= 40 and truncation(capped), "listing byte cap absent or unreported", errors)


def check_replay(session, rows, errors):
    # The live command appends once. Reset its effect, then run each public
    # offline command: an accidental redispatch cannot hide behind idempotence.
    sentinel = session.root / "replay-effect"
    sentinel.write_text("BASE")
    before, original = len(session.requests), session.log.read_bytes()
    rendered = []
    for args in (("dump",), ("render", str(session.log)), ("render", str(session.log))):
        rc, out, _ = session.invoke(args, offline=True)
        require(rc == 0 and bool(out.strip()), "valid offline replay failed", errors)
        if args[0] == "dump":
            decoded = [json.loads(line) for line in out.splitlines() if line.strip()]
            require(decoded[1:] == rows, "dump changed recorded events", errors)
        else:
            require(isinstance(json.loads(out), dict), "render is not a whole request object", errors)
            rendered.append(out)
        require(sentinel.read_text() == "BASE", "offline replay executed a recorded tool", errors)
        require(session.log.read_bytes() == original, "offline replay mutated original log", errors)
    require(len(session.requests) == before, "offline replay made HTTP request", errors)
    require(len(rendered) == 2 and rendered[0] == rendered[1], "repeated offline render changed bytes", errors)


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("binary", type=pathlib.Path)
    parser.add_argument("--vendor", choices=VENDORS, action="append", help="default: all three")
    args = parser.parse_args()
    binary = args.binary.resolve(strict=True)
    checks = []
    scenarios = [("edits-writes", edit_case, check_edits), ("read-ranges-caps", read_case, check_reads),
                 ("search-projection", search_case, check_search), ("command-streams-drain", command_case, check_commands),
                 ("ordinary-errors-complete-batch", invalid_case, check_invalid),
                 ("listing-order-caps", list_case, check_list),
                 ("utf8-prefix-caps", unicode_case, check_unicode)]
    for vendor in args.vendor or VENDORS:
        for name, setup, verify in scenarios:
            with tempfile.TemporaryDirectory(prefix="ensemble-ch03-") as tmp:
                root = pathlib.Path(tmp)
                batch = setup(root)
                s = Session(binary, root, vendor, [response(vendor, batch, NARRATION), response(vendor, [])])
                try:
                    _, results, errors = check_turn(s, [batch])
                    verify(root, results, errors)
                except (OSError, ValueError, KeyError, TypeError, AttributeError, TimeoutError) as err:
                    errors = [f"fixture failed: {type(err).__name__}: {err}"]
                finally:
                    s.close()
                checks.append({"id": vendor + "/" + name, "passed": not errors, "details": errors})
        for name in ("last-batch-limit", "sixteenth-final-success", "continuation-http-failure",
                     "continuation-parse-failure", "replay-no-effects", "empty-final"):
            with tempfile.TemporaryDirectory(prefix="ensemble-ch03-") as tmp:
                root = pathlib.Path(tmp)
                count = 16 if name == "last-batch-limit" else 15 if name == "sixteenth-final-success" else 1
                if name == "empty-final":
                    count = 0
                batches = [[call(f"write-{i}", "write_file", path=f"effect-{i}", content=f"EFFECT-{i}"),
                            call(f"tail-{i}", "write_file", path=f"tail-{i}", content="TAIL")]
                           for i in range(count)]
                if name == "replay-no-effects":
                    batches = [[call("replay", "run_command", command="printf EXECUTED >> replay-effect")]]
                success = name in ("sixteenth-final-success", "replay-no-effects", "empty-final")
                final = "" if name == "empty-final" else FINAL
                replies = [response(vendor, batch, NARRATION) for batch in batches]
                if success:
                    replies.append(response(vendor, [], final))
                elif name == "continuation-http-failure":
                    replies.append(None)
                elif name == "continuation-parse-failure":
                    replies.append({"malformed": "FAKE-FAILURE-PRIVATE-182"})
                s = Session(binary, root, vendor, replies)
                try:
                    rows, results, errors = check_turn(s, batches, success=success, final=final)
                    if name == "replay-no-effects":
                        require((root / "replay-effect").read_text() == "EXECUTED", "live replay control did not execute once", errors)
                        check_replay(s, rows, errors)
                    else:
                        for i in range(count):
                            require((root / f"effect-{i}").exists() and (root / f"effect-{i}").read_text() == f"EFFECT-{i}",
                                    f"accepted side effect {i} lost", errors)
                            require((root / f"tail-{i}").exists(), f"last call in batch {i} did not execute", errors)
                        failures = [r for r in rows if r.get("type") == "error_occurred"]
                        if not success:
                            require(bool(failures), "failed continuation/limit has no error event", errors)
                        if name == "last-batch-limit":
                            require(any(r["error"]["code"] == "round_limit" for r in failures), "missing explicit round_limit event", errors)
                except (OSError, ValueError, KeyError, TypeError, AttributeError, TimeoutError) as err:
                    errors = [f"fixture failed: {type(err).__name__}: {err}"]
                finally:
                    s.close()
                checks.append({"id": vendor + "/" + name, "passed": not errors, "details": errors})
    passed = all(c["passed"] for c in checks)
    print(json.dumps({"scope": "Chapter 3 CLI/fake-wire/recorded-effects only; no live, public ownership, capability isolation, or persistence-fault claim. Command stream markers/labels and separate caps are checked; full label-to-stream association remains in inherited grading/review. Search/list byte caps bound planted content markers, not unspecified presentation metadata.",
                      "binary_sha256": hashlib.sha256(binary.read_bytes()).hexdigest(),
                      "checker_sha256": hashlib.sha256(pathlib.Path(__file__).read_bytes()).hexdigest(),
                      "passed": passed, "checks": checks}, indent=2))
    return 0 if passed else 1


if __name__ == "__main__":
    raise SystemExit(main())
