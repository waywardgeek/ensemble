#!/usr/bin/env python3
"""Independent Chapter 4 CLI/job acceptance from the new published contract.

Local scripted APIs only. No student implementation is imported or inspected.
This is scoped evidence: owner chains, cross-Agent isolation, injected storage
faults, goroutine containment, terminal races, and real model use need separate
checks. Report wording is not fixed; authoritative job facts and artifact bytes
carry the strongest assertions. Tests run serially and clean up fixture children.
"""

import argparse
import copy
import hashlib
import http.server
import json
import os
import pathlib
import re
import shlex
import shutil
import signal
import subprocess
import sys
import tempfile
import threading
import time

from accept_ch03 import MODEL, VENDORS, FINAL, NARRATION, USAGE, call, response, require, result_map, result_text, wire_facts
from accept_ch02 import Harness as OfflineHarness
from accept_chat import Terminal, usage_errors


SUPERVISION = {"wait_for_job", "send_input", "kill_job", "tool_limits"}
TOOLS = sorted(SUPERVISION | {"read_file", "write_file", "edit_file", "list_directory", "search_files", "run_command"})
FIXTURE = r'''import os, pathlib, sys, time, termios
mode = sys.argv[1]
pathlib.Path("fixture-parent.pid").write_text(str(os.getpid()))
if mode == "interactive":
    print("TTY=" + str(all(os.isatty(i) for i in (0,1,2))), flush=True)
    print("READY", flush=True)
    for line in sys.stdin:
        time.sleep(.2)
        print("RECEIVED:" + (line.rstrip("\n") or "<blank>"), flush=True)
        print("READY", flush=True)
elif mode == "late":
    print("EARLY-MARKER", flush=True)
    time.sleep(.35)
    print("LATE-MARKER", flush=True)
elif mode == "blocked":
    child = os.fork()
    if child == 0:
        pathlib.Path("fixture-child.pid").write_text(str(os.getpid()))
        while True: time.sleep(1)
    while not pathlib.Path("fixture-child.pid").exists(): time.sleep(.005)
    print("CHILD-OUTPUT: done exit_code:7", flush=True)
    print("READY", flush=True)
    while True: time.sleep(1)
elif mode == "normalization":
    attrs = termios.tcgetattr(1)
    attrs[1] &= ~termios.ONLCR
    termios.tcsetattr(1, termios.TCSANOW, attrs)
    print("TTY=" + str(all(os.isatty(i) for i in (0,1,2))), flush=True)
    print("TERM=" + os.environ.get("TERM", ""), flush=True)
    size = os.get_terminal_size(1)
    print("SIZE=%dx%d" % (size.lines, size.columns), flush=True)
    os.write(1, b"SPLIT\r"); time.sleep(.05); os.write(1, b"\nLONE\rRETURN\n")
    os.write(2, b"MERGED-STDERR\n")
elif mode == "flood":
    os.write(1, b"BEGIN" + b"~"*100000 + b"END\n")
elif mode == "gated":
    print("READY", flush=True)
    while not pathlib.Path("release-job").exists(): time.sleep(.005)
    print("COMPLETED-DURING-HTTP", flush=True)
elif mode == "cursor":
    os.write(1, b"ABCDEFGHIJ")
    while not pathlib.Path("append-output").exists(): time.sleep(.005)
    os.write(1, b"kl")
    while True: time.sleep(1)
'''


def rows(path):
    data = [json.loads(line) for line in path.read_text().splitlines() if line.strip()]
    if not data or data[0].get("log_version") != 1:
        raise ValueError("missing log header")
    return data[1:]


def fixture_command(root, mode):
    (root / "fixture.py").write_text(FIXTURE)
    return "exec " + shlex.quote(sys.executable) + " " + shlex.quote(str(root / "fixture.py")) + " " + shlex.quote(mode)


def process_alive(pid):
    try:
        os.kill(pid, 0)
    except ProcessLookupError:
        return False
    # A reparented killed child can briefly be a zombie: it cannot run or hold
    # the PTY open. Distinguish that from a surviving process without waiting
    # for an unrelated system reaper's scheduling.
    state = subprocess.run(["ps", "-o", "stat=", "-p", str(pid)], capture_output=True, text=True).stdout.strip()
    return bool(state) and not state.startswith("Z")


def cleanup_fixture(root):
    for path in root.glob("fixture-*.pid"):
        try:
            pid = int(path.read_text())
            if process_alive(pid):
                if os.getpgid(pid) == pid:
                    os.killpg(pid, signal.SIGKILL)
                else:
                    os.kill(pid, signal.SIGKILL)
        except (ValueError, ProcessLookupError, FileNotFoundError):
            pass


class Session:
    def __init__(self, binary, root, vendor, replies, before_reply=None):
        self.binary, self.root, self.vendor = binary, root, vendor
        self.log = root / "events.jsonl"
        self.requests, self.errors = [], []
        session = self

        class Handler(http.server.BaseHTTPRequestHandler):
            def log_message(self, *_):
                pass

            def do_POST(self):
                try:
                    body = json.loads(self.rfile.read(int(self.headers.get("Content-Length", "0"))))
                    session.requests.append(body)
                    index = len(session.requests) - 1
                    if before_reply:
                        before_reply(session, index)
                    payload = replies[index] if index < len(replies) else None
                    status = 503 if payload is None else 200
                    encoded = json.dumps(payload if payload is not None else {"error": {"message": "LOCAL-FAILURE"}}).encode()
                except Exception as err:
                    session.errors.append("fake provider: " + repr(err))
                    status, encoded = 500, b'{}'
                self.send_response(status)
                self.send_header("Content-Type", "application/json")
                self.send_header("Content-Length", str(len(encoded)))
                self.end_headers()
                self.wfile.write(encoded)

        self.server = http.server.ThreadingHTTPServer(("127.0.0.1", 0), Handler)
        self.thread = threading.Thread(target=lambda: self.server.serve_forever(poll_interval=.01), daemon=True)
        self.thread.start()

    def close(self):
        self.server.shutdown()
        self.server.server_close()
        self.thread.join()
        cleanup_fixture(self.root)

    def run(self):
        env = {k: v for k, v in os.environ.items()
               if not k.startswith(("LLM_", "ANTHROPIC_", "OPENAI_", "GEMINI_", "CH02_", "CH03_", "CH04_", "COURSE_"))}
        env.update(LLM_VENDOR=self.vendor, LLM_MODEL=MODEL, LLM_RESOLVED_MODEL=MODEL,
                   LLM_API_KEY="LOCAL-FIXTURE-KEY", LLM_BASE_URL=f"http://127.0.0.1:{self.server.server_port}",
                   CH02_LOG=str(self.log))
        p = subprocess.Popen([str(self.binary)], cwd=self.root, env=env, stdin=subprocess.PIPE,
                             stdout=subprocess.PIPE, stderr=subprocess.PIPE, text=True, start_new_session=True)
        try:
            out, err = p.communicate('{"user":"Run the fixture plan."}\n', timeout=25)
        except subprocess.TimeoutExpired:
            os.killpg(p.pid, signal.SIGKILL)
            p.communicate()
            raise TimeoutError("CLI exceeded25s; fixture process group killed")
        return p.returncode, out, err


def declarations(body, vendor):
    if vendor == "anthropic":
        return [t["name"] for t in body["tools"]]
    if vendor == "openai":
        return [t["function"]["name"] for t in body["tools"]]
    return [t["name"] for group in body["tools"] for t in group["functionDeclarations"]]


def lifecycle_errors(events, root, check_files=True):
    """Pure structural validator also exercised with independent literal controls."""
    errors, jobs, terminal, previous = [], {}, set(), 0
    for e in events:
        seq = e.get("seq", 0)
        require(type(seq) is int and seq > previous, "event sequence not strictly increasing", errors)
        previous = seq
        kind = e.get("type")
        snapshot = e.get("job") if kind in ("job_ended", "job_killed") else e.get("tool", {}).get("job")
        if not snapshot:
            continue
        handle = snapshot.get("handle")
        require(type(handle) is int and handle > 0, "invalid job handle", errors)
        require(snapshot.get("output") == {"kind": 3, "locator": f"cr/io/{handle}"}, "wrong job output reference", errors)
        size = snapshot.get("bytes")
        require(type(size) is int and size >= 0, "invalid job byte count", errors)
        status = snapshot.get("status")
        require(status in ("running", "done", "killed"), "invalid job status", errors)
        if status != "done":
            require("exit_code" not in snapshot, "non-done job has exit code", errors)
        if "exit_code" in snapshot:
            require(type(snapshot["exit_code"]) is int, "invalid exit code", errors)
        if kind == "tool_called":
            require(handle not in jobs and status == "running", "job creation reused handle or starts terminal", errors)
            require(e["tool"]["name"] not in SUPERVISION, "supervision allocated its own job", errors)
            require(size == 0, "job produced bytes before dispatch was recorded", errors)
            jobs[handle] = {"call_id": e["tool"]["call_id"], "name": e["tool"]["name"],
                            "status": "running", "bytes": size}
        elif handle not in jobs:
            errors.append("snapshot or terminal event for unknown job")
        else:
            known = jobs[handle]
            require(size >= known["bytes"], "job byte count decreased", errors)
            if handle in terminal:
                require(size == known["bytes"], "output changed after terminal publication", errors)
            if known["name"] != "run_command":
                require("exit_code" not in snapshot, "local function invented process exit code", errors)
            elif status == "done" and not snapshot.get("is_error", False):
                require("exit_code" in snapshot, "completed process lost exit code", errors)
            if kind in ("job_ended", "job_killed"):
                require(handle not in terminal and known["status"] == "running", "duplicate or reversing terminal event", errors)
                require(status == ("done" if kind == "job_ended" else "killed"), "terminal event/status mismatch", errors)
                if kind == "job_killed":
                    require(snapshot.get("reason") in ("kill_job", "shutdown"), "missing kill reason", errors)
                terminal.add(handle)
                known["status"] = status
            elif kind == "tool_returned":
                require(e["tool"]["call_id"] == known["call_id"], "job snapshot attached to different call", errors)
                require(status == known["status"], "report reversed or invented terminal state", errors)
            known["bytes"] = size
    for handle, known in jobs.items():
        require(handle in terminal, "CLI exit left job without terminal record", errors)
        if check_files:
            path = root / f"cr/io/{handle}"
            require(path.exists(), "job artifact absent", errors)
            if path.exists():
                require(len(path.read_bytes()) == known["bytes"], "final snapshot byte count disagrees with artifact", errors)
    return errors


def evaluate(session, batches, success=True):
    rc, stdout, stderr = session.run()
    errors = list(session.errors)
    require((rc == 0) == success, "unexpected CLI status", errors)
    output = [json.loads(line) for line in stdout.splitlines() if line.strip()]
    require([x["assistant"] for x in output if "assistant" in x] == ([FINAL] if success else []), "wrong final-only answer protocol", errors)
    usage = [x["usage"] for x in output if "usage" in x]
    require(len(usage) == int(success), "wrong final usage protocol", errors)
    accepted = len(batches) + int(success)
    if success and usage:
        require(usage[0] == {k: v * accepted for k, v in USAGE.items()}, "usage lost intermediate responses", errors)
    events = rows(session.log)
    errors += lifecycle_errors(events, session.root)
    results = result_map(events)
    expected = [c for batch in batches for c in batch]
    tool_events = [e for e in events if e.get("type") in ("tool_called", "tool_returned")]
    require([(e["type"], e["tool"]["call_id"]) for e in tool_events] ==
            [(kind, c["id"]) for c in expected for kind in ("tool_called", "tool_returned")],
            "tool call/result pairing or dispatch order changed", errors)
    require(len(session.requests) == len(batches) + 1, "wrong request count/retry", errors)
    for i, request in enumerate(session.requests):
        require(declarations(request, session.vendor) == TOOLS, "all ten tools not declared in stable order", errors)
        calls, returned, signatures = wire_facts(request, session.vendor)
        previous = [c for batch in batches[:i] for c in batch]
        require(calls == [(c["id"], c["name"], c["args"]) for c in previous], "continuation lost call facts", errors)
        require([r[0] for r in returned] == [c["id"] for c in previous], "continuation lost result identities/order", errors)
        if session.vendor == "gemini":
            require(all(signatures.get(c["id"]) == "fixture-signature-" + c["id"] for c in previous), "signed replay lost signature", errors)
    return events, results, errors


def successful(results, identifier, errors):
    require(identifier in results, "missing result " + identifier, errors)
    if identifier not in results:
        return ""
    require(not results[identifier].get("is_error", False), "unexpected tool failure " + identifier, errors)
    return result_text(results, identifier)


def job_of(results, identifier):
    return results[identifier]["job"]


def artifact(root, handle):
    return (root / f"cr/io/{handle}").read_bytes()


def await_record(path, predicate, seconds=3):
    """Wait on a durable fixture boundary, tolerating an unfinished last write."""
    deadline = time.monotonic() + seconds
    while time.monotonic() < deadline:
        if path.exists():
            records = []
            for line in path.read_bytes().splitlines(keepends=True):
                if not line.endswith(b"\n"):
                    continue
                records.append(json.loads(line))
            for record in records:
                if predicate(record):
                    return record
        time.sleep(.005)
    raise TimeoutError("durable fixture boundary did not occur before provider release")


def completion_during_http(binary, vendor):
    """HTTP stays parked until the earlier job publishes its terminal event."""
    with tempfile.TemporaryDirectory(prefix="ensemble-ch04-inflight-") as tmp:
        root = pathlib.Path(tmp)
        (root / "notes").write_text("AFTER-COMPLETION")
        batches = [[call("gated-start", "run_command", command=fixture_command(root, "gated"),
                         ai_callback_delay=2, ai_callback_pattern="READY")],
                   [call("not-yet-assigned", "read_file", path="notes"),
                    call("supplied-id-survives", "read_file", path="notes")]]
        replies = [response(vendor, batch, NARRATION) for batch in batches] + [response(vendor, [])]
        boundaries = {}

        def parked(session, index):
            if index != 1:
                return
            sent = [e for e in rows(session.log) if e.get("type") == "request_sent"][-1]
            boundaries["request_seq"] = sent["seq"]
            (root / "release-job").touch()
            terminal = await_record(session.log, lambda e: e.get("type") == "job_ended"
                                    and e.get("job", {}).get("handle") == 1)
            boundaries["terminal_seq"] = terminal["seq"]
            expected_seq = rows(session.log)[-1]["seq"] + 1
            if vendor == "gemini":
                batches[1][0]["id"] = f"call-{expected_seq}-1"
                replies[1] = response(vendor, batches[1], NARRATION)
                del replies[1]["candidates"][0]["content"]["parts"][1]["functionCall"]["id"]

        session = Session(binary, root, vendor, replies, before_reply=parked)
        try:
            events, results, errors = evaluate(session, batches)
            responses = [e for e in events if e.get("type") == "response_ended"]
            require(boundaries.get("request_seq", 0) < boundaries.get("terminal_seq", 0)
                    < responses[1]["seq"], "terminal event not published while HTTP was parked", errors)
            require("COMPLETED-DURING-HTTP" in artifact(root, 1).decode(),
                    "gated process did not actually finish", errors)
            for item in batches[1]:
                require("AFTER-COMPLETION" in successful(results, item["id"], errors),
                        "post-completion call lost its result", errors)
            if vendor == "gemini":
                part = responses[1]["response"]["parts"][1]
                require(part["call_id"] == f"call-{responses[1]['seq']}-1",
                        "missing call ID used a sequence predicted before HTTP", errors)
        except (OSError, ValueError, KeyError, TypeError, AttributeError, IndexError, TimeoutError) as err:
            errors = [f"fixture failed: {type(err).__name__}: {err}"]
        finally:
            session.close()
        return {"id": vendor + "/completion-during-http", "passed": not errors,
                "details": errors, "observed_boundaries": boundaries}


def occupied_artifacts(binary, vendor):
    """Old leaves survive two applications; old handles confer no authority."""
    with tempfile.TemporaryDirectory(prefix="ensemble-ch04-occupied-") as tmp:
        root = pathlib.Path(tmp)
        spool = root / "cr/io"
        spool.mkdir(parents=True)
        (spool / "1").write_bytes(b"OLD-ARTIFACT\x00\xff")
        (spool / "2").mkdir()
        (spool / "2/retained").write_text("DIRECTORY-CONTENT")
        (root / "target").write_text("SYMLINK-TARGET")
        (spool / "3").symlink_to(root / "target")
        (root / "notes").write_text("NEW-RESULT")
        errors, observations = [], []
        for attempt, handle in enumerate((4, 5)):
            batches = [[call("old-handle", "wait_for_job", handle=1 if attempt == 0 else 4),
                        call("new-read", "read_file", path="notes")]]
            replies = [response(vendor, batches[0], NARRATION), response(vendor, [])]
            session = Session(binary, root, vendor, replies)
            session.log = root / f"events-{attempt}.jsonl"
            try:
                events, results, found = evaluate(session, batches)
                errors.extend(found)
                require(results["old-handle"].get("is_error") is True,
                        "new application accepted an old job handle", errors)
                require(job_of(results, "new-read")["handle"] == handle,
                        "allocation did not skip occupied candidates monotonically", errors)
                require("NEW-RESULT" in successful(results, "new-read", errors),
                        "new work failed in reused workspace", errors)
                created = [e["tool"]["job"]["handle"] for e in events
                           if e.get("type") == "tool_called" and e["tool"].get("job")]
                require(created == [handle], "skipped artifact created a job record", errors)
                observations.append(created)
            except (OSError, ValueError, KeyError, TypeError, AttributeError, TimeoutError) as err:
                errors.append(f"fixture failed: {type(err).__name__}: {err}")
            finally:
                session.close()
        try:
            require((spool / "1").read_bytes() == b"OLD-ARTIFACT\x00\xff", "old artifact changed", errors)
            require((spool / "2/retained").read_text() == "DIRECTORY-CONTENT", "occupied directory changed", errors)
            require((spool / "3").is_symlink() and os.readlink(spool / "3") == str(root / "target"),
                    "occupied symlink changed", errors)
            require((root / "target").read_text() == "SYMLINK-TARGET", "symlink target changed", errors)
            require(artifact(root, 4) == artifact(root, 5) == b"NEW-RESULT",
                    "restart changed previous produced output", errors)
        except OSError as err:
            errors.append(f"preserved artifact check failed: {err}")
        return {"id": vendor + "/occupied-artifacts-restart", "passed": not errors,
                "details": errors, "created_handles": observations}


def allocation_failure(binary, vendor):
    """A non-directory parent is an I/O error, not an occupied leaf to skip."""
    with tempfile.TemporaryDirectory(prefix="ensemble-ch04-allocation-") as tmp:
        root = pathlib.Path(tmp)
        (root / "cr").mkdir()
        (root / "cr/io").write_text("RETAIN-PARENT")
        batches = [[call("refused", "write_file", path="must-not-exist", content="SIDE-EFFECT")]]
        replies = [response(vendor, batches[0], NARRATION), response(vendor, [])]
        session = Session(binary, root, vendor, replies)
        try:
            events, results, errors = evaluate(session, batches)
            require(results["refused"].get("is_error") is True, "allocation error did not refuse operation", errors)
            require(not (root / "must-not-exist").exists(), "handler ran after artifact creation failure", errors)
            require((root / "cr/io").read_text() == "RETAIN-PARENT", "allocation replaced invalid parent", errors)
            require(not any(e.get("job") or e.get("tool", {}).get("job") for e in events),
                    "failed allocation published an admitted job", errors)
        except (OSError, ValueError, KeyError, TypeError, AttributeError, TimeoutError) as err:
            errors = [f"fixture failed: {type(err).__name__}: {err}"]
        finally:
            session.close()
        return {"id": vendor + "/allocation-error-no-execution", "passed": not errors, "details": errors}


def cursor_after_omission(binary, vendor):
    with tempfile.TemporaryDirectory(prefix="ensemble-ch04-cursor-") as tmp:
        root = pathlib.Path(tmp)
        batches = [[call("start", "run_command", command=fixture_command(root, "cursor"),
                         ai_callback_delay=2, ai_callback_pattern="ABCDEFGHIJ", max_output_bytes=4)],
                   [call("empty-poll", "wait_for_job", handle=1, ai_callback_delay=0)],
                   [call("new-output", "wait_for_job", handle=1, ai_callback_delay=2, ai_callback_pattern="kl")],
                   [call("kill", "kill_job", handle=1)]]
        replies = [response(vendor, batch, NARRATION) for batch in batches] + [response(vendor, [])]

        def release(session, index):
            if index == 2:
                (root / "append-output").touch()

        session = Session(binary, root, vendor, replies, before_reply=release)
        try:
            events, results, errors = evaluate(session, batches)
            first = successful(results, "start", errors)
            require("AB" in first and "IJ" in first and "CDEFGH" not in first,
                    "running capped interval did not retain head/tail", errors)
            require(job_of(results, "start")["status"] == "running", "blocked cursor job was prematurely terminal", errors)
            empty = successful(results, "empty-poll", errors)
            later = successful(results, "new-output", errors)
            require(all(x not in empty for x in ("AB", "IJ", "CDEFGH")), "poll repeated consumed capped interval", errors)
            require("kl" in later and all(x not in later for x in ("AB", "IJ", "CDEFGH")),
                    "cursor lost appended bytes or repeated the omitted interval", errors)
            require(artifact(root, 1) == b"ABCDEFGHIJkl", "cursor reporting changed complete artifact bytes", errors)
        except (OSError, ValueError, KeyError, TypeError, AttributeError, TimeoutError) as err:
            errors = [f"fixture failed: {type(err).__name__}: {err}"]
        finally:
            session.close()
        return {"id": vendor + "/cursor-after-omission", "passed": not errors, "details": errors}


def debugger_terminal(binary, vendor):
    """Actual human CLI and actual Delve; scripted local API, not a live model."""
    debugger = shutil.which("dlv")
    if not debugger:
        return {"id": vendor + "/human-debugger", "passed": False,
                "details": ["required Delve prerequisite missing from PATH"]}
    with tempfile.TemporaryDirectory(prefix="ensemble-ch04-debugger-") as tmp:
        root = pathlib.Path(tmp)
        source = 'package main\nimport "fmt"\nfunc main() {\n answer := 37\n answer += 5\n fmt.Println(answer)\n}\n'
        (root / "debug.go").write_text(source)
        built = subprocess.run(["go", "build", "-gcflags=all=-N -l", "-o", "debug-target", "debug.go"],
                               cwd=root, capture_output=True, text=True, timeout=60)
        if built.returncode:
            return {"id": vendor + "/human-debugger", "passed": False,
                    "details": ["debug target failed to build: " + built.stderr]}
        version = subprocess.check_output([debugger, "version"], text=True).strip()
        command = "echo $$ > fixture-parent.pid; exec " + shlex.quote(debugger) + " exec ./debug-target"
        start = call("debug-start", "run_command", command=command, ai_callback_delay=3, ai_callback_pattern=r"\(dlv\) ")
        commands = [("debug-break", "break inspected debug.go:6"), ("debug-continue", "continue"),
                    ("debug-print", "print answer"), ("debug-finish", "continue"), ("debug-exit", "exit")]
        calls = [call(identifier, "send_input", handle=1, input=text, ai_callback_delay=3,
                      **({"ai_callback_pattern": r"\(dlv\) "} if identifier != "debug-exit" else {}))
                 for identifier, text in commands]
        replies = [response(vendor, [start], NARRATION), response(vendor, [], "DEBUGGER-READY"),
                   *[response(vendor, [item], NARRATION) for item in calls[:3]],
                   response(vendor, [], "VALUE-OBSERVED"),
                   *[response(vendor, [item], NARRATION) for item in calls[3:]],
                   response(vendor, [], "DEBUGGER-CLOSED"), response(vendor, [], "REDACTION-VERIFIED")]
        session = Session(binary, root, vendor, replies)
        terminal = None
        errors = []
        try:
            terminal = Terminal(session, ["chat"])
            require("/help" in terminal.prompt(), "human debugger session lacked visible prompt/help", errors)
            for prompt, answer in [("Start the debugger and wait for its prompt.", "DEBUGGER-READY"),
                                   ("Set the breakpoint, continue, and print the variable.", "VALUE-OBSERVED"),
                                   ("Finish execution and exit the debugger.", "DEBUGGER-CLOSED")]:
                display = terminal.prompt((prompt + "\n").encode())
                require("Assistant:" in display and answer in display and NARRATION not in display,
                        "debugger did not complete through readable human chat", errors)
            events = rows(session.log)
            results = result_map(events)
            require("(dlv)" in successful(results, "debug-start", errors), "Delve prompt absent", errors)
            require("Breakpoint" in successful(results, "debug-break", errors), "Delve did not set breakpoint", errors)
            stopped = successful(results, "debug-continue", errors)
            require("debug.go:6" in stopped and "main.main" in stopped, "Delve did not reach selected source breakpoint", errors)
            printed = successful(results, "debug-print", errors)
            require(re.search(r"(?m)^\s*42\s*$", printed) is not None,
                    "actual debugger print result does not contain known value42", errors)
            require("42" in artifact(root, 1).decode(), "debugger artifact lost inspected value", errors)
            target = next(e["seq"] for e in events if e.get("type") == "tool_returned"
                          and e["tool"]["call_id"] == "debug-print")
            history = terminal.prompt(b"/history\n")
            require(f"{target} tool_returned call_id=debug-print" in history,
                    "human history did not expose debugger result target", errors)
            terminal.prompt(f"/redact {target} {target} debugger demonstration\n".encode())
            require("REDACTION-VERIFIED" in terminal.prompt(b"Acknowledge the redacted debugger result.\n"),
                    "human conversation failed after result redaction", errors)
            _, returned, _ = wire_facts(session.requests[-1], vendor)
            text = next(value for identifier, value, _ in returned if identifier == "debug-print")
            require("[redacted]" in text and "42" not in text, "debugger result redaction lost its wire projection", errors)
            errors += usage_errors(terminal.prompt(b"/usage\n"), len(replies))
            rc, final, diagnostics = terminal.finish(b"/quit\n")
            require(rc == 0, "human debugger session failed to close", errors)
            errors += usage_errors(final, len(replies))
            errors += lifecycle_errors(rows(session.log), root)
            require(len(session.requests) == len(replies), "unexpected debugger model request count", errors)
            errors += session.errors
        except (OSError, ValueError, KeyError, TypeError, AttributeError, AssertionError, StopIteration, TimeoutError) as err:
            errors.append(f"fixture failed: {type(err).__name__}: {err}")
        finally:
            if terminal:
                terminal.close()
            session.close()
        return {"id": vendor + "/human-debugger", "passed": not errors, "details": errors,
                "debugger_version": version, "scope": "actual PTY CLI + Delve with local scripted API; no paid model"}


def setup_local(root):
    (root / "selected").write_text("ABCDEFGHIJ")
    (root / "exact").write_text("WXYZ")
    (root / "unicode").write_text("éABCΩ")
    return [[call("limits", "tool_limits", max_output_bytes=4), call("read", "read_file", path="selected")],
            [call("again", "wait_for_job", handle=1, ai_callback_delay=0),
             call("limits2", "tool_limits", max_output_bytes=4), call("exact", "read_file", path="exact")],
            [call("missing", "read_file", path="absent")],
            [call("failed-wait", "wait_for_job", handle=3, ai_callback_delay=0),
             call("finished-input", "send_input", handle=1, input="x"),
             call("done-kill", "kill_job", handle=1)],
            [call("unicode-limit", "tool_limits", max_output_bytes=4),
             call("unicode", "read_file", path="unicode")],
            [call("odd-limit", "tool_limits", max_output_bytes=3),
             call("odd-unicode", "read_file", path="unicode")]]


def verify_local(root, events, results, errors):
    first = successful(results, "read", errors)
    require(artifact(root, 1) == b"ABCDEFGHIJ", "source selection was truncated in artifact", errors)
    require("AB" in first and "IJ" in first and "CDEFGH" not in first, "head/tail report lost selected boundaries", errors)
    require(re.search(r"\b6\b", first) and re.search(r"\b10\b", first) and "cr/io/1" in first, "omission/total/locator absent", errors)
    again = successful(results, "again", errors)
    require("AB" not in again and "IJ" not in again and "CDEFGH" not in again, "cursor repeated consumed bytes", errors)
    exact = successful(results, "exact", errors)
    require("WXYZ" in exact and not re.search("truncat|omitt", exact, re.I), "exact cap falsely truncates", errors)
    require(results["missing"].get("is_error") is True and job_of(results, "missing").get("is_error") is True,
            "failed local handler lost result/job error distinction", errors)
    text = successful(results, "failed-wait", errors)
    require(bool(re.search("error|fail|absent", text, re.I)), "failed-job observation hides original failure", errors)
    require(results["finished-input"].get("is_error") is True, "finished job accepted input", errors)
    successful(results, "done-kill", errors)
    require(not any(e.get("type") == "job_killed" for e in events), "kill of completed job created terminal transition", errors)
    unicode = successful(results, "unicode", errors)
    require("é" in unicode and "Ω" in unicode and "ABC" not in unicode and "�" not in unicode,
            "head/tail cut corrupted complete UTF-8 boundaries", errors)
    require(re.search(r"\b3\b", unicode) and re.search(r"\b7\b", unicode),
            "UTF-8 omission count does not count omitted bytes", errors)
    odd = successful(results, "odd-unicode", errors)
    require("é" in odd and "Ω" not in odd and "�" not in odd and re.search(r"\b5\b", odd),
            "odd UTF-8 budget retained a partial tail or miscounted omissions", errors)


def setup_late(root):
    return [[call("start", "run_command", command=fixture_command(root, "late"), ai_callback_delay=.02)],
            [call("wait", "wait_for_job", handle=1, ai_callback_delay=2)],
            [call("again", "wait_for_job", handle=1, ai_callback_delay=0)]]


def verify_late(root, events, results, errors):
    require(job_of(results, "start")["status"] == "running", "short wake waited for process completion", errors)
    text = successful(results, "wait", errors)
    require("LATE-MARKER" in text, "late output lost after dispatch returned", errors)
    spool = artifact(root, 1)
    require(b"EARLY-MARKER" in spool and b"LATE-MARKER" in spool and spool.endswith(b"exit_code: 0\n"), "incomplete process artifact", errors)
    again = successful(results, "again", errors)
    require("LATE-MARKER" not in again and "EARLY-MARKER" not in again, "completed wait repeated output", errors)


def setup_interactive(root):
    return [[call("start", "run_command", command=fixture_command(root, "interactive"), ai_callback_delay=2, ai_callback_pattern="READY")],
            [call("partial", "send_input", handle=1, input="ab", append_newline=False, ai_callback_delay=.05, ai_callback_pattern="READY")],
            [call("line", "send_input", handle=1, input="c", ai_callback_delay=2, ai_callback_pattern="READY")],
            [call("poll", "send_input", handle=1, input="", append_newline=False, ai_callback_delay=0)],
            [call("blank", "send_input", handle=1, input="", ai_callback_delay=2, ai_callback_pattern="READY")],
            [call("kill", "kill_job", handle=1)], [call("after", "wait_for_job", handle=1, ai_callback_delay=2)]]


def verify_interactive(root, events, results, errors):
    require("TTY=True" in successful(results, "start", errors), "interactive child did not have a PTY on all streams", errors)
    partial = successful(results, "partial", errors)
    require("RECEIVED:" not in partial, "exact input silently appended newline", errors)
    line = successful(results, "line", errors)
    require("RECEIVED:abc" in line and "READY" in line, "old prompt satisfied new-input wait or newline default failed", errors)
    blank = successful(results, "blank", errors)
    require("RECEIVED:<blank>" in blank, "empty default input did not press Enter", errors)
    spool = artifact(root, 1)
    require(b"abc\n" in spool, "PTY input echo missing", errors)
    successful(results, "poll", errors)
    successful(results, "after", errors)


def setup_terminal(root):
    (root / "subdir").mkdir()
    return [[call("terminal", "run_command", command=fixture_command(root, "normalization"))],
            [call("override", "run_command", command="pwd; exit 7", cwd="subdir"),
             call("default", "run_command", command="pwd"),
             call("badcwd", "run_command", command="printf BAD > should-not-exist", cwd="missing-dir")]]


def verify_terminal(root, events, results, errors):
    text = successful(results, "terminal", errors)
    for marker in ("TTY=True", "TERM=dumb", "SIZE=50x200", "MERGED-STDERR"):
        require(marker in text, "terminal setup missing " + marker, errors)
    spool = artifact(root, 1)
    require(b"SPLIT\nLONE\rRETURN\n" in spool, "split CRLF/lone CR normalization wrong", errors)
    require(b"\r\n" not in spool, "artifact retains CRLF", errors)
    override = job_of(results, "override")
    require(override.get("cwd") == str((root / "subdir").resolve()) and override.get("exit_code") == 7,
            "cwd override/exit7 not in job snapshot", errors)
    require("cwd" not in job_of(results, "default"), "default command retained sticky cwd", errors)
    require(str(root.resolve()) in successful(results, "default", errors), "default command ran outside Agent workspace", errors)
    require(results["badcwd"].get("is_error") is True and not (root / "should-not-exist").exists(), "invalid cwd executed with fallback", errors)


def setup_flood(root):
    return [[call("flood", "run_command", command=fixture_command(root, "flood"), max_output_bytes=64, ai_callback_delay=2)],
            [call("again", "wait_for_job", handle=1, ai_callback_delay=2),
             call("recover", "read_file", path="cr/io/1", max_bytes=100100)]]


def verify_flood(root, events, results, errors):
    data = artifact(root, 1)
    require(data.startswith(b"BEGIN" + b"~" * 100000 + b"END\n") and data.endswith(b"exit_code: 0\n"), "spool did not retain full output", errors)
    text = successful(results, "flood", errors)
    require(text.count("~") <= 64 and "BEGIN" in text and "END" in text, "bounded report missing head/tail", errors)
    require(str(len(data)) in text and "cr/io/1" in text, "report omitted total/locator", errors)
    require("BEGIN" not in successful(results, "again", errors), "cursor repeated flooded interval", errors)
    # Recovery's own inline report is capped, but its artifact must contain the
    # complete explicitly requested selected source text.
    require(artifact(root, 2) == data, "read_file cannot recover complete artifact bytes", errors)


def setup_kill(root):
    return [[call("start", "run_command", command=fixture_command(root, "blocked"), ai_callback_delay=2, ai_callback_pattern="READY")],
            [call("kill", "kill_job", handle=1)], [call("wait", "wait_for_job", handle=1, ai_callback_delay=2), call("again", "kill_job", handle=1)]]


def setup_shutdown(root):
    return [[call("start", "run_command", command=fixture_command(root, "blocked"), ai_callback_delay=2, ai_callback_pattern="READY")]]


def verify_stopped(root, events, results, errors, reason):
    ended = [e["job"] for e in events if e.get("type") == "job_killed"]
    require(len(ended) == 1 and ended[0].get("reason") == reason, "missing or duplicated terminal kill reason", errors)
    for name in ("fixture-parent.pid", "fixture-child.pid"):
        path = root / name
        require(path.exists(), "fixture process did not reach readiness", errors)
        if path.exists():
            deadline = time.monotonic() + 1
            while process_alive(int(path.read_text())) and time.monotonic() < deadline:
                time.sleep(.01)
            require(not process_alive(int(path.read_text())), "managed process/group child survived CLI cleanup", errors)
    require(b"CHILD-OUTPUT: done exit_code:7" in artifact(root, 1), "kill discarded child text resembling status metadata", errors)
    if reason == "kill_job":
        successful(results, "wait", errors)
        successful(results, "again", errors)


def setup_limits(root):
    (root / "source").write_text("ABCDEFGHIJ")
    return [[call("set", "tool_limits", max_output_bytes=4), call("burn", "unavailable", ignored=True),
             call("unlimited", "read_file", path="source")],
            [call("set2", "tool_limits", max_output_bytes=4), call("set3", "tool_limits", max_output_bytes=6),
             call("six", "read_file", path="source")],
            [call("set4", "tool_limits", max_output_bytes=4), call("badsetter", "tool_limits", ai_callback_delay=-1),
             call("afterbad", "read_file", path="source")],
            [call("set5", "tool_limits", ai_callback_delay=0, max_output_bytes=2),
             call("explicit", "run_command", command="printf TENLETTERS", ai_callback_delay=2, max_output_bytes=100)]]


def verify_limits(root, events, results, errors):
    require(results["burn"].get("is_error") is True and results["badsetter"].get("is_error") is True, "invalid attempted calls unexpectedly succeeded", errors)
    for identifier in ("burn", "set3", "badsetter", "explicit", "six"):
        first = result_text(results, identifier).splitlines()[0]
        require("tool_limits" in first, "one-shot consumption not visible on " + identifier, errors)
    for identifier in ("unlimited", "afterbad"):
        text = successful(results, identifier, errors)
        require("ABCDEFGHIJ" in text and "tool_limits" not in text, "one-shot limits leaked past attempted call", errors)
    six = successful(results, "six", errors)
    require("ABC" in six and "HIJ" in six and "DEFG" not in six, "second setter did not replace overrides", errors)
    require("TENLETTERS" in successful(results, "explicit", errors), "explicit call arguments failed to override pending limits", errors)


def imported_fixture():
    provenance = {"vendor": "anthropic", "model": MODEL, "surface": "messages"}
    job = {"handle": 1, "status": "running", "output": {"kind": 3, "locator": "cr/io/1"}, "bytes": 0}
    def e(seq, kind, key, value):
        return {"seq": seq, "type": kind, "time": "2026-01-01T00:00:00Z", key: value}
    return [{"log_version": 1},
            e(1, "message_received", "message", {"actor": "human", "purpose": "dialogue", "parts": [{"type": "text", "text": "read"}]}),
            e(2, "response_ended", "response", {"from": provenance, "usage": USAGE, "parts": [{"type": "tool_call", "call_id": "read-1", "name": "read_file", "args": {"path": "missing"}, "from": provenance}]}),
            e(3, "tool_called", "tool", {"call_id": "read-1", "name": "read_file", "args": {"path": "missing"}, "job": job}),
            e(4, "job_ended", "job", dict(job, status="done", bytes=3)),
            e(5, "tool_returned", "tool", {"call_id": "read-1", "parts": [{"type": "text", "text": "ok\n"}], "job": dict(job, status="done", bytes=3)})]


def offline_checks(binary, checks):
    with tempfile.TemporaryDirectory(prefix="ensemble-ch04-replay-") as tmp:
        h = OfflineHarness(binary, pathlib.Path(tmp))
        try:
            original = imported_fixture()
            for command in ("dump", "render"):
                out, errors = h.run(original, command, model=MODEL)
                if command == "dump" and not errors:
                    require([json.loads(x) for x in out.splitlines()] == original, "dump changed imported job facts", errors)
                checks.append({"id": "offline/" + command + "-without-artifact", "passed": not errors, "details": errors})
            mutations = [
                ("unknown-terminal", lambda r: r[4]["job"].update(handle=2)),
                ("changed-locator", lambda r: r[4]["job"]["output"].update(locator="elsewhere")),
                ("running-exit-code", lambda r: r[3]["tool"]["job"].update(exit_code=0)),
                ("decreasing-bytes", lambda r: r[5]["tool"]["job"].update(bytes=2)),
                ("reversing-result", lambda r: r[5]["tool"]["job"].update(status="running")),
                ("duplicate-terminal", lambda r: r.append(dict(copy.deepcopy(r[4]), seq=6))),
            ]
            for name, change in mutations:
                fixture = copy.deepcopy(original)
                # Break shared literal-map aliases so only the intended record changes.
                fixture = json.loads(json.dumps(fixture))
                change(fixture)
                _, errors = h.run(fixture, success=False)
                checks.append({"id": "offline/" + name, "passed": not errors, "details": errors})
        finally:
            h.close()


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("binary", type=pathlib.Path)
    parser.add_argument("--vendor", choices=VENDORS, action="append")
    args = parser.parse_args()
    binary = args.binary.resolve(strict=True)
    checks = []
    scenarios = [("local-jobs-cursor", setup_local, verify_local), ("late-output", setup_late, verify_late),
                 ("interactive-new-output", setup_interactive, verify_interactive), ("pty-cwd", setup_terminal, verify_terminal),
                 ("full-spool", setup_flood, verify_flood), ("one-shot-limits", setup_limits, verify_limits),
                 ("group-kill", setup_kill, lambda *a: verify_stopped(*a, "kill_job")),
                 ("eof-shutdown", setup_shutdown, lambda *a: verify_stopped(*a, "shutdown")),
                 ("error-shutdown", setup_shutdown, lambda *a: verify_stopped(*a, "shutdown"))]
    for vendor in args.vendor or VENDORS:
        for name, setup, verify in scenarios:
            with tempfile.TemporaryDirectory(prefix="ensemble-ch04-") as tmp:
                root = pathlib.Path(tmp)
                batches = setup(root)
                success = name != "error-shutdown"
                replies = [response(vendor, batch, NARRATION) for batch in batches] + [response(vendor, []) if success else None]
                s = Session(binary, root, vendor, replies)
                try:
                    events, results, errors = evaluate(s, batches, success)
                    verify(root, events, results, errors)
                except (OSError, ValueError, KeyError, TypeError, AttributeError, TimeoutError) as err:
                    errors = [f"fixture failed: {type(err).__name__}: {err}"]
                finally:
                    s.close()
                checks.append({"id": vendor + "/" + name, "passed": not errors, "details": errors})
                print(f"{vendor}/{name}: {'PASS' if not errors else 'FAIL'}", file=sys.stderr, flush=True)
        for extra in (completion_during_http, cursor_after_omission, occupied_artifacts,
                      allocation_failure, debugger_terminal):
            check = extra(binary, vendor)
            checks.append(check)
            print(f"{check['id']}: {'PASS' if check['passed'] else 'FAIL'}", file=sys.stderr, flush=True)
    offline_checks(binary, checks)
    passed = all(c["passed"] for c in checks)
    print(json.dumps({"scope": __doc__, "binary_sha256": hashlib.sha256(binary.read_bytes()).hexdigest(),
                      "checker_sha256": hashlib.sha256(pathlib.Path(__file__).read_bytes()).hexdigest(),
                      "passed": passed, "checks": checks}, indent=2))
    return 0 if passed else 1


if __name__ == "__main__":
    raise SystemExit(main())
