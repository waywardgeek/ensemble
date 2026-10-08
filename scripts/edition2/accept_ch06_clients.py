#!/usr/bin/env python3
"""Initial Chapter 6 CLI barrier checks; not the complete chapter acceptance gate.

Local fixtures only. The server cannot send its terminal frames until the test
has observed the first fragment through an actual PTY or protocol --observe.
No student implementation is imported or inspected.
"""
import argparse
import http.server
import json
import os
import pathlib
import pty
import select
import signal
import subprocess
import tempfile
import termios
import threading
import time

VENDORS = ("anthropic", "openai", "gemini")
MODEL = "fixture-ch06"


def frame(value, event=None):
    prefix = "event: " + event + "\n" if event else ""
    return (prefix + "data: " + json.dumps(value) + "\n\n").encode()


def fixture(vendor):
    if vendor == "anthropic":
        def f(kind, **kw):
            return frame(dict(type=kind, **kw), kind)
        first = f("message_start", message=dict(model=MODEL, content=[], usage=dict(input_tokens=10, output_tokens=0)))
        first += f("content_block_start", index=0, content_block=dict(type="text", text=""))
        first += f("content_block_delta", index=0, delta=dict(type="text_delta", text="Hel"))
        tail = f("content_block_delta", index=0, delta=dict(type="text_delta", text="lo."))
        tail += f("content_block_stop", index=0)
        tail += f("message_delta", delta=dict(stop_reason="end_turn"), usage=dict(output_tokens=2))
        return first, tail + f("message_stop")
    if vendor == "openai":
        def f(delta, reason=None):
            return frame(dict(model=MODEL, choices=[dict(index=0, delta=delta, finish_reason=reason)]))
        first = f(dict(role="assistant", content="Hel"))
        tail = f(dict(content="lo."), "stop")
        tail += frame(dict(model=MODEL, choices=[], usage=dict(prompt_tokens=10, completion_tokens=2)))
        return first, tail + b"data: [DONE]\n\n"
    first = frame(dict(modelVersion=MODEL, candidates=[dict(index=0, content=dict(role="model", parts=[dict(text="Hel")]))]))
    tail = frame(dict(modelVersion=MODEL, candidates=[dict(index=0, content=dict(role="model", parts=[dict(text="lo."), dict(text="", thoughtSignature="fixture-signature")]), finishReason="STOP")], usageMetadata=dict(promptTokenCount=10, candidatesTokenCount=2)))
    return first, tail


class Server:
    def __init__(self, vendor):
        self.release = threading.Event()
        self.requested = threading.Event()
        self.requests = []
        self.errors = []
        first, tail = fixture(vendor)
        owner = self

        class Handler(http.server.BaseHTTPRequestHandler):
            def log_message(self, *args):
                pass

            def do_POST(self):
                try:
                    body = json.loads(self.rfile.read(int(self.headers["Content-Length"])))
                    owner.requests.append((self.path, body))
                    owner.requested.set()
                    self.send_response(200)
                    self.send_header("Content-Type", "text/event-stream; charset=utf-8")
                    self.end_headers()
                    self.wfile.write(first)
                    self.wfile.flush()
                    owner.release.wait(15)
                    self.wfile.write(tail)
                    self.wfile.flush()
                except (BrokenPipeError, ConnectionResetError):
                    pass
                except Exception as exc:
                    owner.errors.append(str(exc))

        self.http = http.server.ThreadingHTTPServer(("127.0.0.1", 0), Handler)
        self.http.daemon_threads = True
        self.thread = threading.Thread(target=self.http.serve_forever, daemon=True)
        self.thread.start()

    def close(self):
        self.release.set()
        self.http.shutdown()
        self.http.server_close()
        self.thread.join()


class Client:
    def __init__(self, binary, root, server, vendor, human):
        env = {k: v for k, v in os.environ.items() if not k.startswith(("LLM_", "CH02_", "COURSE_", "ANTHROPIC_", "OPENAI_", "GEMINI_", "EN_DISABLE_STREAMING"))}
        env.update(LLM_VENDOR=vendor, LLM_MODEL=MODEL, LLM_RESOLVED_MODEL=MODEL,
                   LLM_API_KEY="LOCAL-ONLY", LLM_BASE_URL=f"http://127.0.0.1:{server.http.server_port}",
                   CH02_LOG=str(root / "events.jsonl"))
        self.human = human
        self.output = b""
        self.err = tempfile.TemporaryFile()
        if human:
            self.fd, slave = pty.openpty()
            attrs = termios.tcgetattr(slave)
            attrs[3] &= ~termios.ECHO
            termios.tcsetattr(slave, termios.TCSANOW, attrs)
            self.process = subprocess.Popen([str(binary), "chat"], stdin=slave, stdout=slave, stderr=self.err, cwd=root, env=env, start_new_session=True)
            os.close(slave)
        else:
            self.process = subprocess.Popen([str(binary), "protocol", "--observe"], stdin=subprocess.PIPE, stdout=subprocess.PIPE, stderr=self.err, cwd=root, env=env, start_new_session=True)
            self.fd = self.process.stdout.fileno()

    def send(self, text):
        if self.human:
            os.write(self.fd, text.encode())
        else:
            self.process.stdin.write(text.encode())
            self.process.stdin.flush()

    def records(self):
        return [json.loads(line) for line in self.output.split(b"\n")[:-1] if line]

    def wait(self, predicate, failure, timeout):
        deadline = time.monotonic() + timeout
        while not predicate():
            if time.monotonic() >= deadline:
                raise AssertionError(failure)
            if select.select([self.fd], [], [], min(.05, max(0, deadline-time.monotonic())))[0]:
                try:
                    data = os.read(self.fd, 65536)
                except OSError:
                    data = b""
                if not data:
                    raise AssertionError(failure + " (client exited)")
                self.output += data

    def close(self):
        if self.process.poll() is None:
            os.killpg(self.process.pid, signal.SIGKILL)
        self.process.wait(timeout=5)
        if self.human:
            os.close(self.fd)
        else:
            self.process.stdin.close()
            self.process.stdout.close()
        self.err.close()


def check(binary, vendor, human, timeout=5):
    with tempfile.TemporaryDirectory(prefix="ch06-barrier-") as tmp:
        server = Server(vendor)
        c = Client(binary, pathlib.Path(tmp), server, vendor, human)
        try:
            if human:
                c.wait(lambda: b"You>" in c.output, "human prompt unavailable", timeout)
                c.send("Explain this local fixture.\n")
            else:
                c.send(json.dumps(dict(kind="prompt", text="Explain this local fixture.")) + "\n")
            assert server.requested.wait(timeout), "provider request not received"
            path, request = server.requests[0]
            if vendor == "gemini":
                assert ":streamGenerateContent" in path and "alt=sse" in path, "stream endpoint/options absent"
            else:
                assert request.get("stream") is True, "stream option absent"
                if vendor == "openai":
                    assert request.get("stream_options", {}).get("include_usage") is True, "stream usage option absent"
            def early():
                if human:
                    return b"Hel" in c.output
                return any(x.get("observation", {}).get("kind") == "part_delta" and x["observation"].get("channel") == "text" and x["observation"].get("text") == "Hel" for x in c.records())
            c.wait(early, "first fragment absent while terminal frames withheld", timeout)
            assert not server.release.is_set()
            if not human:
                before = c.records()
                assert not any("completion" in x for x in before), "completion preceded API terminal"
            server.release.set()
            if human:
                c.wait(lambda: b"lo." in c.output and b"success" in c.output.lower(), "human final text/status absent", timeout)
                assert c.output.count(b"Hel") == 1 and c.output.count(b"lo.") == 1, "normal answer duplicated"
                assert b"fixture-signature" not in c.output, "opaque signature printed"
                c.send("/quit\n")
            else:
                c.wait(lambda: any("completion" in x for x in c.records()), "completion absent", timeout)
                records = c.records()
                observations = [x["observation"] for x in records if "observation" in x]
                begins = [x for x in observations if x["kind"] == "model_begin"]
                ends = [x for x in observations if x["kind"] == "model_end"]
                assert len(begins) == len(ends) == 1, "operation begin/end cardinality"
                begin, end = begins[0], ends[0]
                assert begin["delivery"] == "stream" and end["accepted"] is True
                identity = tuple(begin[k] for k in ("agent_id", "request_id", "operation_id"))
                assert all(isinstance(v, str) and v for v in identity), "empty operation identity"
                assert all(tuple(x[k] for k in ("agent_id", "request_id", "operation_id")) == identity for x in observations), "operation identity changed"
                deltas = [x for x in observations if x["kind"] == "part_delta"]
                finals = [x for x in observations if x["kind"] == "part_final"]
                assert "".join(x["text"] for x in deltas if x["channel"] == "text") == "Hello."
                assert all(type(x["part_id"]) is int and x["part_id"] > 0 for x in deltas + finals)
                assert len(finals) == (2 if vendor == "gemini" else 1), "final typed part count"
                assert finals[0]["part"] == {"type": "text", "text": "Hello."}
                assert all(x["part_id"] == finals[0]["part_id"] for x in deltas), "text part identity changed"
                assert [x["part_index"] for x in finals] == list(range(len(finals)))
                assert end["response_seq"] > 0 and all(x["response_seq"] == end["response_seq"] for x in finals)
                assert observations[0] == begin and observations[-1] == end
                assert len({x["part_id"] for x in finals}) == len(finals), "final part IDs reused"
                accepted = next(i for i, x in enumerate(records) if x.get("accepted") == "prompt")
                bi = next(i for i, x in enumerate(records) if x.get("observation") == begin)
                ei = next(i for i, x in enumerate(records) if x.get("observation") == end)
                ci = next(i for i, x in enumerate(records) if "completion" in x)
                assert accepted < bi < ei < ci, "acceptance/observation/completion order"
                done = records[ci]["completion"]
                assert done["request_id"] == identity[1] and done["outcome"] == "success" and done["text"] == "Hello."
            assert len(server.requests) == 1 and not server.errors, "unexpected retry or fixture failure"
        finally:
            c.close()
            server.close()


def main():
    p = argparse.ArgumentParser(description=__doc__)
    p.add_argument("binary", type=pathlib.Path)
    p.add_argument("--vendor", action="append", choices=VENDORS)
    p.add_argument("--mode", action="append", choices=("human-pty", "observe"))
    args = p.parse_args()
    rows = []
    for vendor in args.vendor or VENDORS:
        for mode in args.mode or ("human-pty", "observe"):
            try:
                check(args.binary.resolve(strict=True), vendor, mode == "human-pty")
                errors = []
            except Exception as exc:
                errors = [type(exc).__name__ + ": " + str(exc)]
            rows.append(dict(id=vendor + "/" + mode, passed=not errors, details=errors))
    result = dict(scope="initial CLI barriers; complete Chapter 6 gate pending", passed=all(x["passed"] for x in rows), checks=rows)
    print(json.dumps(result, indent=2))
    return 0 if result["passed"] else 1


if __name__ == "__main__":
    raise SystemExit(main())
