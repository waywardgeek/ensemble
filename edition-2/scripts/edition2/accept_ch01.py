#!/usr/bin/env python3
"""Independent black-box checks for second-edition Chapter 1's CLI contract.

Runs only against a loopback fake, never a live provider. Architecture and
external-library review remain separate evidence; this report does not claim
to prove them or replace the inherited grader. No student source is imported.
"""

import argparse
import contextlib
import http.server
import json
import os
import pathlib
import socket
import subprocess
import tempfile
import threading
import time


KEY = "edition-two-local-only-credential"
MODEL = "edition-two-local-only-model"
GOOD = {"content": [{"type": "text", "text": "alpha"},
                    {"type": "text", "text": " beta"}],
        "usage": {"input_tokens": 11, "output_tokens": 3}}


@contextlib.contextmanager
def server(mode):
    requests = []
    release = threading.Event()

    class Handler(http.server.BaseHTTPRequestHandler):
        def log_message(self, *_):
            pass

        def do_POST(self):
            size = int(self.headers.get("Content-Length", "0"))
            raw = self.rfile.read(size)
            requests.append({"path": self.path, "body": json.loads(raw),
                             "key": self.headers.get("x-api-key"),
                             "version": self.headers.get("anthropic-version")})
            if mode == "transport":
                self.connection.shutdown(socket.SHUT_RDWR)
                self.connection.close()
                return
            if mode == "timeout":
                release.wait(95)
                return
            body = json.loads(json.dumps(GOOD))
            status = 200
            if mode == "http":
                status = 503
            elif mode == "empty-answer":
                body["content"] = []
            elif mode == "missing-usage":
                del body["usage"]
            elif mode == "missing-input":
                del body["usage"]["input_tokens"]
            elif mode == "missing-output":
                del body["usage"]["output_tokens"]
            elif mode in {"negative", "fractional", "null", "string", "boolean"}:
                body["usage"]["input_tokens"] = {
                    "negative": -1, "fractional": 1.5, "null": None,
                    "string": "11", "boolean": True}[mode]
            elif mode == "nontext":
                body["content"].insert(0, {"type": "thinking", "thinking": "ignore",
                                          "signature": "local-fixture"})
            elif mode == "zero":
                body["usage"] = {"input_tokens": 0, "output_tokens": 0}
            elif mode == "late-failure" and len(requests) > 1:
                status = 503
            raw = b"{" if mode == "malformed-response" else json.dumps(body).encode()
            self.send_response(status)
            self.send_header("Content-Type", "application/json")
            self.send_header("Content-Length", str(len(raw)))
            self.end_headers()
            try:
                self.wfile.write(raw)
            except (BrokenPipeError, ConnectionResetError):
                pass

    httpd = http.server.ThreadingHTTPServer(("127.0.0.1", 0), Handler)
    thread = threading.Thread(target=lambda: httpd.serve_forever(poll_interval=0.01), daemon=True)
    thread.start()
    try:
        yield f"http://127.0.0.1:{httpd.server_port}///", requests
    finally:
        release.set()
        httpd.shutdown()
        httpd.server_close()
        thread.join()


def check(binary, name, mode, stdin, missing=None, success=False, calls=1):
    with server(mode) as (url, requests):
        env = {k: v for k, v in os.environ.items()
               if not k.startswith(("ANTHROPIC_", "LLM_", "COURSE_"))}
        env.update(ANTHROPIC_API_KEY=KEY, ANTHROPIC_MODEL=MODEL,
                   ANTHROPIC_BASE_URL=url)
        if missing:
            env.pop(missing)
        started = time.monotonic()
        with tempfile.TemporaryDirectory(prefix="edition2-ch01-run-") as cwd:
            try:
                p = subprocess.run([str(binary)], input=stdin, text=True,
                                   capture_output=True, env=env, cwd=cwd, timeout=90)
            except subprocess.TimeoutExpired:
                return {"id": name, "passed": False,
                        "details": ["Harness reached 90-second infrastructure limit; no bounded completion observed."]}
        errors = []
        if len(requests) != calls:
            errors.append(f"expected {calls} HTTP requests; observed {len(requests)}")
        if KEY in p.stdout or KEY in p.stderr:
            errors.append("diagnostics exposed the local test credential")
        if success:
            if p.returncode != 0:
                errors.append("successful input did not exit zero")
            try:
                lines = [json.loads(line) for line in p.stdout.splitlines()]
            except ValueError:
                lines = []
                errors.append("stdout is not JSON lines")
            usage = {"input": 11 * calls, "output": 3 * calls}
            if mode == "zero":
                usage = {"input": 0, "output": 0}
            expected = [{"assistant": "alpha beta"}] * calls + [{"usage": usage}]
            if lines != expected:
                errors.append("answers or final usage differ from the fake's actual data")
            questions = [json.loads(line)["user"] for line in stdin.splitlines() if line.strip()]
            history = []
            for i, request in enumerate(requests):
                if i >= len(questions):
                    break  # the request-count assertion already diagnoses retries
                history.append({"role": "user", "content": questions[i]})
                msgs = request["body"].get("messages")
                normalized = []
                for msg in msgs or []:
                    text = msg.get("content")
                    if isinstance(text, list):
                        text = "".join(b.get("text", "") for b in text if b.get("type") == "text")
                    normalized.append({"role": msg.get("role"), "content": text})
                if normalized != history:
                    errors.append(f"request {i + 1} did not preserve exact conversation growth")
                if (request["path"] != "/v1/messages" or request["key"] != KEY
                        or request["version"] != "2023-06-01"
                        or request["body"].get("model") != MODEL):
                    errors.append(f"request {i + 1} ignored supplied transport configuration")
                history.append({"role": "assistant", "content": "alpha beta"})
        else:
            if p.returncode == 0:
                errors.append("rejected input/response exited zero")
            if not p.stderr.strip():
                errors.append("failure produced no stderr diagnostic")
            expected_stdout = '{"assistant":"alpha beta"}\n' if mode == "late-failure" else ""
            # JSON spacing is not a protocol requirement.
            try:
                actual = [json.loads(line) for line in p.stdout.splitlines()] if p.stdout.strip() else []
            except ValueError:
                actual = ["invalid stdout"]
            expected = [json.loads(expected_stdout)] if expected_stdout else []
            if actual != expected:
                errors.append("failure fabricated an answer or emitted success usage")
            if missing and missing not in p.stderr:
                errors.append("missing configuration diagnostic did not name the variable")
            if missing == "ANTHROPIC_MODEL" and "/v1/models" not in p.stderr:
                errors.append("missing model diagnostic did not point to discovery")
        return {"id": name, "passed": not errors, "details": errors,
                "http_requests": len(requests),
                "elapsed_seconds": round(time.monotonic() - started, 3)}


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("binary", type=pathlib.Path)
    parser.add_argument("--only", help="one check ID, useful for a focused mutation")
    parser.add_argument("--skip-timeout", action="store_true",
                        help="omit the slow timeout probe in unrelated mutation runs")
    args = parser.parse_args()
    binary = args.binary.resolve(strict=True)
    one = '{"user":"one"}\n'
    cases = [
        ("missing-key", "good", one, "ANTHROPIC_API_KEY", False, 0),
        ("missing-model", "good", one, "ANTHROPIC_MODEL", False, 0),
        ("empty-input", "good", "", None, True, 0),
        ("blank-lines", "good", "\n  \n" + one, None, True, 1),
        ("bad-input", "good", "{\n", None, False, 0),
        ("empty-question", "good", '{"user":""}\n', None, False, 0),
        ("exact-growth", "good", one + '{"user":"two"}\n', None, True, 2),
        ("nontext-block", "nontext", one, None, True, 1),
        ("zero-usage", "zero", one, None, True, 1),
        ("late-failure", "late-failure", one + '{"user":"two"}\n', None, False, 2),
    ]
    for mode in ["http", "transport", "malformed-response", "empty-answer",
                 "missing-usage", "missing-input", "missing-output", "negative",
                 "fractional", "null", "string", "boolean", "timeout"]:
        cases.append((mode, mode, one, None, False, 1))
    if args.only and args.only not in {c[0] for c in cases}:
        parser.error("unknown check ID")
    results = [check(binary, *case) for case in cases
               if (not args.only or case[0] == args.only)
               and not (args.skip_timeout and case[0] == "timeout")]
    print(json.dumps({"scope": "second-edition Chapter 1 CLI acceptance; no live or architecture claim",
                      "passed": all(r["passed"] for r in results), "checks": results}, indent=2))
    return 0 if all(r["passed"] for r in results) else 1


if __name__ == "__main__":
    raise SystemExit(main())
