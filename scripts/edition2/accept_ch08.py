#!/usr/bin/env python3
"""Partial Chapter 8 settings wire/persistence checks from the published contract.

Usage: python3 scripts/edition2/accept_ch08.py GUI_BINARY
No paid backend, browser rendering, speech or turn-limit execution is tested.
Those and persistence/concurrency fault controls remain separate required gates.
"""
import argparse
import contextlib
import hashlib
import http.server
import json
import os
from pathlib import Path
import queue
import re
import signal
import subprocess
import tempfile
import threading

from accept_ch07 import Socket


def defaults():
    return dict(theme="dark", font_size=16, sidebar_width=260,
                actions_width=380, autoplay=False, speech_rate=1)


def check_preferences(value, revision, expected):
    assert type(value.get("revision")) is int and value["revision"] == revision, "preference revision"
    assert value.get("preferences") == expected, "complete applied preferences differ"
    assert type(value["preferences"]["autoplay"]) is bool, "false must remain a Boolean"


def check_policy(value, revision, maximum):
    assert set(value) == {"revision", "persistent", "max_model_requests", "effective_max_model_requests"}, "unsafe/incomplete policy snapshot"
    assert type(value["revision"]) is int and value["revision"] == revision
    assert value["persistent"] is True
    assert type(value["max_model_requests"]) is int and value["max_model_requests"] == maximum
    assert value["effective_max_model_requests"] == (maximum or 16), "effective policy differs"


def subscribe(client, preferences, preference_revision, maximum, policy_revision):
    client.send(dict(type="subscribe", id="subscribe"))
    first = client.next()
    assert first.get("type") == "preferences_snapshot", "preferences must precede Agent snapshot"
    check_preferences(first, preference_revision, preferences)
    begin = client.next()
    assert begin.get("type") == "snapshot_begin" and begin.get("id") == "subscribe"
    state = begin["state"]
    check_policy(state["execution_policy"], policy_revision, maximum)
    assert state["active_max_model_requests"] is None and state["active_request_id"] is None
    end = client.until(lambda x: x.get("type") == "snapshot_end")
    assert end["generation"] == begin["generation"] and end["watermark"] == begin["watermark"]
    assert "LOCAL-CH08-NOT-A-SECRET" not in json.dumps(client.records), "unsafe credential projection"


def update(client, domain, command_id, base, patch, revision, expected):
    start = len(client.records)
    client.send(dict(type=domain + "_update", id=command_id, base_revision=base, patch=patch))
    reply = client.until(lambda x: x.get("id") == command_id and x.get("type") in (domain + "_ack", "error"))
    assert reply.get("type") == domain + "_ack", "valid patch refused: " + str(reply.get("code"))
    assert type(reply.get("revision")) is int and reply["revision"] == revision, "acknowledgement revision"
    records = client.records[start:]
    if revision == base:
        kind = domain + "_changed"
        assert not any(x.get("type") == kind or x.get("observation", {}).get("kind") == kind for x in records), "no-change patch broadcast a change"
        return
    if domain == "preferences":
        change = next((x for x in records if x.get("type") == "preferences_changed" and x.get("revision") == revision), None)
        assert change is not None, "sender change must precede preferences acknowledgement"
        check_preferences(change, revision, expected)
    else:
        # Public actor publication precedes its reply. Wire delivery can arrive
        # later through the watch, so do not invent a same-socket ordering rule.
        match = lambda x: x.get("type") == "observation" and x.get("observation", {}).get("kind") == "policy_changed" and x["observation"]["execution_policy"].get("revision") == revision
        change = next((x for x in records if match(x)), None)
        if change is None:
            change = client.until(match)
        assert change["revision"] == reply.get("watch_revision"), "policy/watch acknowledgement cut differs"
        check_policy(change["observation"]["execution_policy"], revision, expected)


class Backend(http.server.BaseHTTPRequestHandler):
    def do_POST(self):
        self.server.requests.append(self.path)
        self.send_error(500, "unexpected local request")

    def log_message(self, *_):
        pass


class Program:
    def __init__(self, binary, directory, endpoint, number):
        env = os.environ.copy()
        env.update(LLM_VENDOR="openai", LLM_MODEL="fixture-ch08-independent",
                   LLM_RESOLVED_MODEL="fixture-ch08-independent", LLM_API_KEY="LOCAL-CH08-NOT-A-SECRET",
                   LLM_BASE_URL=endpoint, CH02_LOG=str(directory / f"events-{number}.log"))
        self.lines = queue.Queue()
        self.process = subprocess.Popen([str(binary), "--port", "0", "--preferences", str(directory / "preferences.json"),
                                         "--policy", str(directory / "policy.json")], cwd=directory, env=env,
                                        stdout=subprocess.PIPE, stderr=subprocess.STDOUT, text=True, start_new_session=True)
        self.reader = threading.Thread(target=self.read, daemon=True)
        self.reader.start()

    def read(self):
        for line in self.process.stdout:
            self.lines.put(line)
        self.lines.put(None)

    def listen(self):
        for _ in range(100):
            line = self.lines.get(timeout=5)
            assert line is not None, "program exited before listening"
            match = re.search(r"http://127\.0\.0\.1:\d+", line)
            if match:
                return match.group()
        raise AssertionError("listening URL absent")

    def close(self):
        if self.process.poll() is None:
            os.killpg(self.process.pid, signal.SIGTERM)
        try:
            self.process.wait(timeout=5)
        except subprocess.TimeoutExpired:
            os.killpg(self.process.pid, signal.SIGKILL)
            self.process.wait()
        self.process.stdout.close()
        self.reader.join(timeout=2)


def evaluate(binary):
    rows = []
    backend = http.server.ThreadingHTTPServer(("127.0.0.1", 0), Backend)
    backend.requests = []
    thread = threading.Thread(target=backend.serve_forever, daemon=True)
    thread.start()
    try:
        with tempfile.TemporaryDirectory(prefix="ch08-settings-") as temporary:
            directory = Path(temporary)
            endpoint = f"http://127.0.0.1:{backend.server_port}"
            stage = "initial-full-snapshots"
            try:
                with contextlib.closing(Program(binary, directory, endpoint, 1)) as program:
                    url = program.listen()
                    with contextlib.closing(Socket(url)) as a, contextlib.closing(Socket(url)) as b:
                        subscribe(a, defaults(), 0, 0, 0)
                        subscribe(b, defaults(), 0, 0, 0)
                        rows.append(dict(id=stage, passed=True))
                        stage = "sparse-false-broadcast-and-no-change"
                        expected = defaults() | dict(theme="light", autoplay=True, speech_rate=1.5)
                        update(a, "preferences", "enable", 0, dict(theme="light", autoplay=True, speech_rate=1.5), 1, expected)
                        check_preferences(b.until(lambda x: x.get("type") == "preferences_changed"), 1, expected)
                        expected["autoplay"] = False
                        update(a, "preferences", "disable", 1, dict(autoplay=False), 2, expected)
                        check_preferences(b.until(lambda x: x.get("type") == "preferences_changed"), 2, expected)
                        saved = (directory / "preferences.json").read_bytes()
                        stat = (directory / "preferences.json").stat()
                        update(a, "preferences", "same", 2, dict(autoplay=False), 2, expected)
                        assert (directory / "preferences.json").read_bytes() == saved, "no-change write changed file"
                        after = (directory / "preferences.json").stat()
                        assert (after.st_ino, after.st_mtime_ns) == (stat.st_ino, stat.st_mtime_ns), "no-change patch rewrote file"
                        assert json.loads(saved) == dict(version=1, revision=2, preferences=expected)
                        rows.append(dict(id=stage, passed=True))
                        stage = "policy-positive-zero-and-watch"
                        update(a, "policy", "two", 0, dict(max_model_requests=2), 1, 2)
                        change = b.until(lambda x: x.get("observation", {}).get("kind") == "policy_changed")
                        check_policy(change["observation"]["execution_policy"], 1, 2)
                        update(a, "policy", "default", 1, dict(max_model_requests=0), 2, 0)
                        change = b.until(lambda x: x.get("observation", {}).get("kind") == "policy_changed")
                        check_policy(change["observation"]["execution_policy"], 2, 0)
                        assert json.loads((directory / "policy.json").read_bytes()) == dict(version=1, revision=2, policy=dict(max_model_requests=0))
                        rows.append(dict(id=stage, passed=True))
                        stage = "invalid-and-conflict-remain-usable"
                        for index, (domain, patch, code, base) in enumerate([
                            ("preferences", dict(autoplay=None), "invalid_preferences", 2),
                            ("preferences", dict(font_size=11), "invalid_preferences", 2),
                            ("preferences", dict(theme="light", surprise=True), "invalid_preferences", 2),
                            ("policy", dict(max_model_requests=True), "invalid_policy", 2),
                            ("policy", dict(max_model_requests=257), "invalid_policy", 2),
                            ("preferences", dict(theme="dark"), "revision_conflict", 1),
                        ]):
                            command_id = f"invalid-{index}"
                            a.send(dict(type=domain + "_update", id=command_id, base_revision=base, patch=patch))
                            error = a.until(lambda x: x.get("id") == command_id)
                            assert error.get("type") == "error" and error.get("code") == code, "wrong intended refusal"
                            if code == "revision_conflict":
                                assert error.get("domain") == domain
                                check_preferences(error["current"], 2, expected)
                        update(a, "preferences", "after-errors", 2, dict(font_size=22), 3, expected | dict(font_size=22))
                        expected["font_size"] = 22
                        rows.append(dict(id=stage, passed=True))
                stage = "fresh-process-persistence"
                with contextlib.closing(Program(binary, directory, endpoint, 2)) as program:
                    with contextlib.closing(Socket(program.listen())) as client:
                        subscribe(client, expected, 3, 0, 2)
                rows.append(dict(id=stage, passed=True))
            except Exception as error:
                rows.append(dict(id=stage, passed=False, details=[type(error).__name__ + ": " + str(error)]))
    finally:
        backend.shutdown()
        backend.server_close()
        thread.join(timeout=2)
    rows.append(dict(id="no-model-http", passed=not backend.requests, details=backend.requests))
    return dict(scope=__doc__, binary_sha256=hashlib.sha256(binary.read_bytes()).hexdigest(),
                checker_sha256=hashlib.sha256(Path(__file__).read_bytes()).hexdigest(),
                transport_helper_sha256=hashlib.sha256(Path(__file__).with_name("accept_ch07.py").read_bytes()).hexdigest(),
                passed=len(rows) == 6 and all(x["passed"] for x in rows), checks=rows)


if __name__ == "__main__":
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("binary", type=Path)
    arguments = parser.parse_args()
    result = evaluate(arguments.binary.resolve(strict=True))
    print(json.dumps(result, indent=2))
    raise SystemExit(0 if result["passed"] else 1)
