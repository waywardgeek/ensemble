#!/usr/bin/env python3
"""Independent Chapter 2/3 human CLI checks, including actual PTY interaction.

Local fake APIs establish deterministic interface behavior, not paid/live model
usability. The latter requires separate coder receipts. No student source is
read. Large lines use explicit piped chat to avoid the terminal driver's own
canonical-line limit. Chapter 3 adds actual tools and successful human redaction;
architecture and public-library isolation remain separate checks.
"""

import argparse
import errno
import hashlib
import json
import os
import pathlib
import pty
import re
import select
import signal
import subprocess
import tempfile
import termios
import time

from accept_ch03 import (Session, VENDORS, MODEL, response, records, text_content,
                         call, declaration_errors, wire_facts, result_map, result_text)


PROMPT = b"You> "
LIMIT = 1024 * 1024
KEY = "LOCAL-CHAT-SECRET-481"
ANSWER = "DISPLAY-FIRST-381\nDISPLAY-SECOND-629"
PRIVATE_RESULT = "FILE-RESULT-TO-REDACT-826"
INTERMEDIATE = "MODEL-INTERMEDIATE-NOT-FINAL-742"


def environment(session):
    env = {k: v for k, v in os.environ.items()
           if not k.startswith(("LLM_", "ANTHROPIC_", "OPENAI_", "GEMINI_", "CH02_", "COURSE_"))}
    env.update(LLM_VENDOR=session.vendor, LLM_MODEL=MODEL, LLM_RESOLVED_MODEL=MODEL,
               LLM_API_KEY=KEY, CH02_LOG=str(session.log),
               LLM_BASE_URL=f"http://127.0.0.1:{session.server.server_port}")
    return env


def user_texts(body, vendor):
    if vendor == "openai":
        return [text_content(m["content"]) for m in body["messages"] if m["role"] == "user"]
    values = []
    for m in body["messages" if vendor == "anthropic" else "contents"]:
        if m["role"] == "user":
            values.append("".join(p.get("text", "") for p in m["content" if vendor == "anthropic" else "parts"]))
    return values


def usage_errors(text, count):
    errors = []
    for label, number in (("input", 5 * count), ("cache[ _-]*write", 0),
                          ("cache[ _-]*read", 0), ("output", 2 * count)):
        # Punctuation and spacing are presentation choices; labels and values
        # are public facts. Require adjacency so an unrelated number cannot pass.
        if not re.search(r"\b" + label + r"(?:\s+tokens)?\s*[:=]?\s*" + str(number) + r"\b", text, re.I):
            errors.append("missing readable usage value: " + label)
    return errors


class Terminal:
    def __init__(self, session, args):
        self.master, slave = pty.openpty()
        attrs = termios.tcgetattr(slave)
        attrs[3] &= ~termios.ECHO
        termios.tcsetattr(slave, termios.TCSANOW, attrs)
        self.stderr = tempfile.TemporaryFile()
        self.process = subprocess.Popen([str(session.binary), *args], cwd=session.root,
                                        env=environment(session), stdin=slave, stdout=slave,
                                        stderr=self.stderr, start_new_session=True)
        os.close(slave)
        self.output = bytearray()
        self.cursor = 0

    def read(self, seconds):
        if select.select([self.master], [], [], seconds)[0]:
            try:
                chunk = os.read(self.master, 65536)
            except OSError as exc:
                if exc.errno == errno.EIO:
                    return False
                raise
            self.output.extend(chunk)
            return bool(chunk)
        return True

    def prompt(self, data=None):
        if data is not None:
            os.write(self.master, data)
        deadline = time.monotonic() + 6
        while True:
            index = self.output.find(PROMPT, self.cursor)
            if index >= 0:
                end = index + len(PROMPT)
                value = bytes(self.output[self.cursor:end]).decode("utf-8")
                self.cursor = end
                return value.replace("\r\n", "\n")
            if time.monotonic() >= deadline or not self.read(.05):
                raise AssertionError("missing flushed next prompt in terminal")

    def finish(self, data):
        os.write(self.master, data)
        deadline = time.monotonic() + 6
        while self.process.poll() is None:
            if time.monotonic() >= deadline:
                raise AssertionError("terminal did not finish after EOF/quit")
            self.read(.05)
        while self.read(0):
            if not select.select([self.master], [], [], 0)[0]:
                break
        self.stderr.seek(0)
        return (self.process.returncode,
                bytes(self.output[self.cursor:]).decode("utf-8").replace("\r\n", "\n"),
                self.stderr.read().decode("utf-8"))

    def close(self):
        if self.process.poll() is None:
            os.killpg(self.process.pid, signal.SIGKILL)
        self.process.wait()
        os.close(self.master)
        self.stderr.close()


def terminal_case(session, automatic=False):
    errors = []
    terminal = Terminal(session, [] if automatic else ["chat"])
    try:
        banner = terminal.prompt()
        if MODEL not in banner or "/help" not in banner:
            errors.append("banner omitted selected model or help")
        original = session.log.read_bytes()
        for command in (b"   \n", b"/help\n", b"/usage\n", b"/history\n", b"/unknown\n", b"/ephemeral\n", b"/redact nope\n"):
            out = terminal.prompt(command)
            if command == b"/help\n":
                if any(name not in out for name in ("/help", "/usage", "/history", "/ephemeral", "/redact", "/quit")):
                    errors.append("help omitted command")
                if not re.search(r"1\s*MiB|1048576|1,048,576|1024\s*KiB", out, re.I):
                    errors.append("help omitted byte limit")
            if command == b"/usage\n":
                errors.extend(usage_errors(out, 0))
            if session.requests or session.log.read_bytes() != original:
                errors.append("blank/local inspection/malformed command contacted API or mutated log")
        prompts = ['  preserve spaces  ', '//help', '{"user":"literal JSON"}']
        for index, value in enumerate(prompts):
            if index == 1:
                terminal.prompt(b"/ephemeral TRANSIENT-726\n")
            out = terminal.prompt(value.encode() + b"\n")
            if "Assistant:" not in out or ANSWER not in out or "\\nDISPLAY-SECOND" in out:
                errors.append("answer missing or JSON-escaped in human presentation")
        if len(session.requests) != 3:
            errors.append("ordinary human lines did not produce exactly three requests")
        else:
            for index, expected in enumerate((prompts[0], "/help", prompts[2])):
                if not any(expected in value for value in user_texts(session.requests[index], session.vendor)):
                    errors.append("human input altered, slash escape broken, or JSON interpreted")
            present = ["TRANSIENT-726" in json.dumps(r) for r in session.requests]
            if present != [False, True, False]:
                errors.append("human ephemeral not restricted to next actual request")
        events = records(session.log)
        human = [e["message"] for e in events if e["type"] == "message_received"
                 and e["message"]["actor"] == "human" and e["message"]["purpose"] == "dialogue"]
        if ["".join(p.get("text", "") for p in m["parts"]) for m in human] != [prompts[0], "/help", prompts[2]]:
            errors.append("recorded human text altered or local command logged as conversation")
        out = terminal.prompt(b"/history\n")
        for event in events:
            if not any(str(event["seq"]) in line and event["type"] in line for line in out.splitlines()):
                errors.append("history omitted recorded sequence/kind")
                break
        errors.extend(usage_errors(terminal.prompt(b"/usage\n"), 3))
        code, out, err = terminal.finish(b"\x04" if automatic else b"/quit\n")
        if code != 0:
            errors.append("clean terminal finish failed")
        errors.extend(usage_errors(out, 3))
        if KEY in bytes(terminal.output).decode() + err:
            errors.append("credential leaked")
        if len(session.requests) != 3 or records(session.log) != events:
            errors.append("inspection or clean exit mutated conversation")
    finally:
        terminal.close()
    return errors


def pipe_case(session, kind):
    args, data, success, count = ["chat"], b"", True, 0
    env = environment(session)
    if kind == "empty":
        pass
    elif kind == "exact-limit-crlf":
        data, count = ("é" * (LIMIT // 2)).encode() + b"\r\n", 1
    elif kind == "over-limit":
        data, success = b"x" * (LIMIT + 1) + b"\n", False
    elif kind == "invalid-utf8":
        data, success = b"secret-invalid-\xff\n", False
    elif kind == "operation-refusal":
        data, success = b"/redact 1 1 no result exists\n", False
    elif kind == "provider-failure":
        data, success, count = b"hello\n", False, 1
    elif kind == "missing-model":
        env.pop("LLM_MODEL")
        success = False
    elif kind == "unknown-subcommand":
        args, success = ["not-a-command"], False
    elif kind in ("protocol", "redirected-default"):
        args = ["protocol"] if kind == "protocol" else []
        data, count = b'{"user":"/help"}\n', 1
    elif kind in ("empty-response", "tool-only"):
        data, count = b"hello\n", 1
    else:
        raise ValueError(kind)
    p = subprocess.run([str(session.binary), *args], input=data, capture_output=True,
                       cwd=session.root, env=env, timeout=12)
    out, err = p.stdout.decode("utf-8"), p.stderr.decode("utf-8")
    errors = []
    if (p.returncode == 0) != success or len(session.requests) != count:
        errors.append("incorrect exit status or HTTP request count")
    if KEY in out + err:
        errors.append("credential leaked")
    if success:
        if kind in ("protocol", "redirected-default"):
            try:
                lines = [json.loads(line) for line in out.splitlines()]
                if lines != [{"assistant": ANSWER}, {"usage": {"input": 5, "cache_write": 0, "cache_read": 0, "output": 2}}]:
                    errors.append("machine mode changed exact output")
            except ValueError:
                errors.append("machine mode included chat presentation")
        else:
            errors.extend(usage_errors(out, count))
        if kind == "exact-limit-crlf" and session.requests:
            if user_texts(session.requests[0], session.vendor)[-1] != "é" * (LIMIT // 2):
                errors.append("exact UTF-8 limit rejected, truncated, or CRLF retained")
        if kind == "empty-response" and "[No text returned]" not in out:
            errors.append("empty answer notice missing")
        if kind == "tool-only" and "[Tool calls returned; execution is not available in this chapter]" not in out:
            errors.append("tool-only notice missing")
    else:
        if not err.strip():
            errors.append("failure lacked safe stderr diagnostic")
        if not usage_errors(out, 0):
            errors.append("failed session printed successful usage summary")
        if kind in ("missing-model", "unknown-subcommand") and (PROMPT.decode() in out or "Assistant:" in out):
            errors.append("configuration/subcommand error printed conversation UI")
    return errors


def fixture(vendor, kind):
    if kind == "provider-failure":
        return [None]
    if kind == "empty-response":
        return [response(vendor, [], "")]
    if kind == "tool-only":
        body = response(vendor, [{"id": "literal-call", "name": "inspect", "args": {}}], "")
        if vendor == "anthropic":
            body["content"] = [p for p in body["content"] if p["type"] != "text"]
        elif vendor == "openai":
            body["choices"][0]["message"]["content"] = None
        else:
            body["candidates"][0]["content"]["parts"] = [p for p in body["candidates"][0]["content"]["parts"] if "text" not in p]
        return [body]
    return [response(vendor, [], ANSWER) for _ in range(3)]


def tool_plan(root):
    (root / "notes.txt").write_text(PRIVATE_RESULT + "\n")
    return [call("missing", "read_file", path="absent.txt"),
            call("read-visible", "read_file", path="notes.txt"),
            call("write", "write_file", path="made.txt", content="BEFORE\n"),
            call("edit", "edit_file", path="made.txt", old_text="BEFORE", new_text="AFTER"),
            call("list", "list_directory", path="."),
            call("search", "search_files", path="made.txt", pattern="AFTER"),
            call("command", "run_command", command="printf 'OUT-482'; printf 'ERR-371' >&2; exit 7")]


def tool_projection_errors(body, vendor, calls, redacted):
    errors = declaration_errors(body, vendor)
    actual, results, signatures = wire_facts(body, vendor)
    if actual != [(c["id"], c["name"], c["args"]) for c in calls]:
        errors.append("continuation changed call identities/arguments/order")
    if [r[0] for r in results] != [c["id"] for c in calls]:
        errors.append("continuation lost result pairing/order")
    values = {identifier: (text, failed) for identifier, text, failed in results}
    if "read-visible" not in values:
        errors.append("missing independent read result")
    else:
        text = values["read-visible"][0]
        if redacted:
            if PRIVATE_RESULT in text or "[redacted]" not in text:
                errors.append("selected result was not replaced by a redaction stub")
        elif PRIVATE_RESULT not in text:
            errors.append("positive result content absent before redaction")
    if vendor == "gemini" and any(signatures.get(c["id"]) != "fixture-signature-" + c["id"] for c in calls):
        errors.append("human continuation lost bound replay signatures")
    return errors


def tool_terminal_case(session, calls, failure=False):
    errors = []
    terminal = Terminal(session, ["chat"])
    try:
        terminal.prompt()
        if failure:
            code, out, err = terminal.finish(b"Exercise the tools.\n")
            if code == 0 or not err.strip() or "Assistant:" in out or "usage" in out.lower():
                errors.append("continuation failure produced success UI or lacked failure status")
        else:
            out = terminal.prompt(b"Exercise the tools.\n")
            if out.count("Assistant:") != 1 or ANSWER not in out or INTERMEDIATE in out or PRIVATE_RESULT in out:
                errors.append("tool turn displayed intermediate facts instead of one final readable answer")
            if "execution is not available" in out:
                errors.append("Chapter 2 tool-only notice survived into tool execution")
        captured = records(session.log)
        results = result_map(captured)
        if list(results) != [c["id"] for c in calls]:
            errors.append("tool batch not fully paired after ordinary error")
        else:
            if not results["missing"].get("is_error", False):
                errors.append("missing file did not produce ordinary tool error")
            if any(results[c["id"]].get("is_error", False) for c in calls[1:]):
                errors.append("valid tool or nonzero command exit became tool failure")
            for identifier, markers in (("read-visible", [PRIVATE_RESULT]), ("list", ["made.txt"]),
                                        ("search", ["AFTER"]), ("command", ["OUT-482", "ERR-371", "7"])):
                if any(marker not in result_text(results, identifier) for marker in markers):
                    errors.append("tool result missing observed content: " + identifier)
        if (session.root / "made.txt").read_bytes() != b"AFTER\n":
            errors.append("write/edit result disagreed with disk")
        if len(session.requests) != 2:
            errors.append("first full tool turn did not use exactly two requests")
        else:
            errors.extend(tool_projection_errors(session.requests[1], session.vendor, calls, False))
        for request in session.requests:
            errors.extend(declaration_errors(request, session.vendor))
        if failure:
            if not any(e["type"] == "error_occurred" for e in captured):
                errors.append("failed continuation lost safe durable error")
            return errors
        before_history = session.log.read_bytes()
        history = terminal.prompt(b"/history\n")
        target = next(e for e in captured if e["type"] == "tool_returned" and e["tool"]["call_id"] == "read-visible")
        # Discover the number in the displayed row, then cross-check the log.
        # No hardcoded event sequence is assumed.
        candidate = [line for line in history.splitlines() if "tool_returned" in line and "read-visible" in line]
        numbers = re.findall(r"\b\d+\b", candidate[0]) if len(candidate) == 1 else []
        if str(target["seq"]) not in numbers:
            errors.append("history did not expose usable result sequence and call identity")
        if session.log.read_bytes() != before_history or len(session.requests) != 2:
            errors.append("history inspection mutated or contacted API")
        sequence = target["seq"]
        acknowledgement = terminal.prompt(f"/redact {sequence} {sequence} controlled terminal redaction\n".encode())
        if "redact" not in acknowledgement.lower() or len(session.requests) != 2:
            errors.append("redaction lacked local acknowledgement or made HTTP request")
        after_redaction = records(session.log)
        if after_redaction[:-1] != captured or after_redaction[-1]["type"] != "redacted":
            errors.append("redaction rewrote recorded facts or omitted directive")
        out = terminal.prompt(b"Continue after the deliberate redaction.\n")
        if out.count("Assistant:") != 1 or ANSWER not in out:
            errors.append("redacted follow-up did not display final answer")
        if len(session.requests) != 3:
            errors.append("redacted follow-up made wrong request count")
        else:
            errors.extend(tool_projection_errors(session.requests[2], session.vendor, calls, True))
        errors.extend(usage_errors(terminal.prompt(b"/usage\n"), 3))
        code, out, err = terminal.finish(b"/quit\n")
        if code != 0:
            errors.append("full tool/redaction chat failed clean exit")
        errors.extend(usage_errors(out, 3))
        if KEY in bytes(terminal.output).decode() + err:
            errors.append("credential leaked")
    finally:
        terminal.close()
    return errors


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("binary", type=pathlib.Path)
    parser.add_argument("--chapter", type=int, choices=(2, 3), default=2)
    args = parser.parse_args()
    binary = args.binary.resolve(strict=True)
    checks = []
    kinds = ("pty-chat", "pty-default", "empty", "exact-limit-crlf", "over-limit", "invalid-utf8",
             "operation-refusal", "provider-failure", "missing-model", "unknown-subcommand",
             "protocol", "redirected-default", "empty-response", "tool-only")
    if args.chapter == 3:
        kinds = tuple(k for k in kinds if k != "tool-only") + ("tool-turn-redaction", "tool-continuation-failure")
    for vendor in VENDORS:
        for kind in kinds:
            with tempfile.TemporaryDirectory(prefix="ensemble-chat-check-") as tmp:
                root = pathlib.Path(tmp)
                calls = tool_plan(root) if kind.startswith("tool-") and args.chapter == 3 else None
                replies = fixture(vendor, kind)
                if calls:
                    replies = [response(vendor, calls, INTERMEDIATE),
                               None if kind == "tool-continuation-failure" else response(vendor, [], ANSWER),
                               response(vendor, [], ANSWER)]
                session = Session(binary, root, vendor, replies)
                try:
                    if calls:
                        errors = tool_terminal_case(session, calls, kind == "tool-continuation-failure")
                    else:
                        errors = (terminal_case(session, kind == "pty-default") if kind.startswith("pty-")
                                  else pipe_case(session, kind))
                    errors.extend(session.server_errors)
                except (AssertionError, ValueError, KeyError, TypeError, OSError, subprocess.TimeoutExpired) as exc:
                    errors = [type(exc).__name__ + ": " + str(exc)]
                finally:
                    session.close()
                checks.append({"id": vendor + "/" + kind, "passed": not errors, "details": errors})
    print(json.dumps({"chapter": args.chapter, "binary_sha256": hashlib.sha256(binary.read_bytes()).hexdigest(),
                      "checker_sha256": hashlib.sha256(pathlib.Path(__file__).read_bytes()).hexdigest(),
                      "checks": checks}, indent=2))
    return 0 if all(c["passed"] for c in checks) else 1


if __name__ == "__main__":
    raise SystemExit(main())
