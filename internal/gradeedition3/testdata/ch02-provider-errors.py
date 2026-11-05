#!/usr/bin/env python3
"""Public Chapter 2 supplement: real CLI failure facts, no candidate imports."""
import argparse
import http.server
import json
import os
from pathlib import Path
import signal
import subprocess
import tempfile
import threading
import uuid


def run_case(executable, workspace, vendor, malformed):
    name = vendor + ("-empty-candidates" if malformed else "-valid-response")
    work = workspace / name
    work.mkdir()
    for directory in ("home", "tmp", "cli"):
        (work / directory).mkdir()
    marker = "synthetic-" + uuid.uuid4().hex
    if vendor == "openai":
        body = {
            "model": "review-model",
            "choices": [] if malformed else [{"message": {"content": marker}}],
            "usage": {"prompt_tokens": 3, "completion_tokens": 2},
        }
    else:
        body = {
            "modelVersion": "review-model",
            "candidates": [] if malformed else [
                {"content": {"parts": [{"text": marker}]}}
            ],
            "usageMetadata": {"promptTokenCount": 3, "candidatesTokenCount": 2},
        }
    # A diagnostic must not echo the raw malformed response.
    if malformed:
        body["untrusted_diagnostic"] = marker
    encoded = json.dumps(body).encode("utf-8")
    requests = []

    class Handler(http.server.BaseHTTPRequestHandler):
        def do_POST(self):
            length = int(self.headers.get("Content-Length", "0"))
            if length < 0 or length > 1024 * 1024:
                self.send_error(413)
                return
            self.rfile.read(length)
            requests.append(self.path)
            self.send_response(200)
            self.send_header("Content-Type", "application/json")
            self.send_header("Content-Length", str(len(encoded)))
            self.end_headers()
            self.wfile.write(encoded)

        def log_message(self, *_):
            pass

    # Bound socket I/O as well as the candidate lifetime, so joined teardown
    # cannot wait indefinitely for a candidate that sends an incomplete request.
    class BoundedServer(http.server.HTTPServer):
        def get_request(self):
            connection, address = super().get_request()
            connection.settimeout(5)
            return connection, address

    server = BoundedServer(("127.0.0.1", 0), Handler)
    worker = threading.Thread(target=server.serve_forever)
    worker.start()
    environment = {
        "PATH": "/usr/bin:/bin:/opt/homebrew/bin",
        "HOME": str(work / "home"),
        "TMPDIR": str(work / "tmp"),
        "DOTNET_CLI_HOME": str(work / "cli"),
        "DOTNET_SKIP_FIRST_TIME_EXPERIENCE": "1",
        "DOTNET_CLI_TELEMETRY_OPTOUT": "1",
        "DOTNET_NOLOGO": "1",
        "LLM_VENDOR": vendor,
        "LLM_MODEL": "review-model",
        "LLM_API_KEY": "synthetic-unused",
        "LLM_BASE_URL": "http://127.0.0.1:" + str(server.server_port),
        "CH02_LOG": str(work / "events.jsonl"),
    }
    process = None
    timed_out = False
    try:
        with (work / "stdout.txt").open("wb") as output, (work / "stderr.txt").open("wb") as errors:
            process = subprocess.Popen(
                [str(executable)], cwd=work, env=environment,
                stdin=subprocess.PIPE, stdout=output, stderr=errors,
                start_new_session=True,
            )
            try:
                process.communicate(b'{"user":"synthetic provider boundary"}\n', timeout=20)
            except subprocess.TimeoutExpired:
                timed_out = True
                os.killpg(process.pid, signal.SIGKILL)
                process.communicate()
    finally:
        # The process group belongs to this probe. Clean descendants even when
        # the leader exits before them; never infer cleanup from listener closure.
        if process is not None:
            try:
                os.killpg(process.pid, signal.SIGKILL)
            except ProcessLookupError:
                pass
            process.wait()
        server.shutdown()
        worker.join()
        server.server_close()

    failure = None
    checks = []

    def require(condition, check):
        if not condition:
            raise ValueError(check)
        checks.append(check)

    try:
        require(not timed_out, "bounded-process-exit")
        require(len(requests) == 1, "exactly-one-request-no-retry")
        require((work / "stdout.txt").stat().st_size <= 65536
                and (work / "stderr.txt").stat().st_size <= 65536,
                "bounded-diagnostic-output")
        stdout = (work / "stdout.txt").read_text()
        stderr = (work / "stderr.txt").read_text()
        records = [json.loads(line) for line in stdout.splitlines() if line.strip()]
        events = [json.loads(line) for line in (work / "events.jsonl").read_text().splitlines() if line.strip()]
        kinds = [event.get("type") for event in events if "type" in event]
        require(kinds.count("message_received") == 1 and kinds.count("request_sent") == 1,
                "human-and-request-recorded")
        if malformed:
            require(process.returncode == 1, "controlled-failure-exit-not-crash")
            require(not any("assistant" in record for record in records), "no-fabricated-answer")
            require(kinds.count("error_occurred") == 1 and "response_ended" not in kinds,
                    "one-error-fact-no-success-fact")
            require(kinds.index("request_sent") < kinds.index("error_occurred"),
                    "error-recorded-after-request")
            require(marker not in stdout + stderr + json.dumps(events), "raw-response-not-echoed")
        else:
            require(process.returncode == 0, "successful-response-exit")
            require([r.get("assistant") for r in records if "assistant" in r] == [marker],
                    "real-response-answer-preserved")
            require(kinds.count("response_ended") == 1 and "error_occurred" not in kinds,
                    "success-recorded-without-error")
    except Exception as error:
        failure = str(error)
    return {"name": name, "passed": failure is None, "checks": checks,
            "failure": failure, "candidate_exit": process.returncode,
            "workspace": str(work)}


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("executable", type=Path)
    parser.add_argument("--work-root", required=True, type=Path)
    args = parser.parse_args()
    if os.name != "posix" or not args.executable.is_absolute() or not args.executable.is_file():
        parser.error("requires a POSIX host and an absolute candidate executable")
    if not args.work_root.is_absolute() or not args.work_root.is_dir():
        parser.error("--work-root must be an existing absolute owned directory")
    workspace = Path(tempfile.mkdtemp(prefix="ch2-provider-errors-", dir=args.work_root))
    cases = [run_case(args.executable, workspace, vendor, malformed)
             for vendor in ("openai", "gemini") for malformed in (False, True)]
    report = {"contract": "chapter2-provider-errors-v1", "passed": all(c["passed"] for c in cases),
              "cases": cases, "scope": "Offline CLI supplement; no chapter points or live-provider certification."}
    (workspace / "result.json").write_text(json.dumps(report, indent=2) + "\n")
    print(json.dumps(report))
    return 0 if report["passed"] else 1


if __name__ == "__main__":
    raise SystemExit(main())
