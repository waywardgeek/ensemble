#!/usr/bin/env python3
"""Independent Chapter 6 wire cases through the public CLI; partial acceptance.

Published fixtures, local HTTP only. No student parser or grader implementation
is imported. Lifecycle/public-owner and exhaustive split checks remain separate.
"""
import argparse
import hashlib
import http.server
import json
import os
import pathlib
import subprocess
import tempfile
import threading

from accept_ch06_clients import MODEL, VENDORS, fixture, frame

LIMIT = 1_048_576


def plain_fixture(vendor):
    if vendor == "anthropic":
        return dict(model=MODEL, content=[dict(type="text", text="Hello.")],
                    stop_reason="end_turn", usage=dict(input_tokens=10, output_tokens=2))
    if vendor == "openai":
        return dict(model=MODEL, choices=[dict(index=0, message=dict(role="assistant", content="Hello."), finish_reason="stop")],
                    usage=dict(prompt_tokens=10, completion_tokens=2))
    return dict(modelVersion=MODEL, candidates=[dict(index=0, content=dict(role="model", parts=[dict(text="Hello."),
                dict(text="", thoughtSignature="fixture-signature")]), finishReason="STOP")],
                usageMetadata=dict(promptTokenCount=10, candidatesTokenCount=2))


def events(body):
    """Fixture helper: canonical fixtures use one JSON data line per LF event."""
    return [block + b"\n\n" for block in body.split(b"\n\n") if block]


def one_byte_over_data_frame(exact):
    """Keep a valid data field/JSON value; exceed only the physical byte limit."""
    assert len(exact) == LIMIT and exact.count(b"data: ") == 1
    return exact.replace(b"data: ", b"data:  ", 1)


def without_usage(vendor, body):
    output = []
    for block in events(body):
        lines = block.decode().splitlines()
        for i, line in enumerate(lines):
            if not line.startswith("data: ") or line == "data: [DONE]":
                continue
            value = json.loads(line[6:])
            value.pop("usage", None)
            value.pop("usageMetadata", None)
            if isinstance(value.get("message"), dict):
                value["message"].pop("usage", None)
            lines[i] = "data: " + json.dumps(value)
        output.append(("\n".join(lines) + "\n\n").encode())
    return b"".join(output)


def incomplete_call(vendor):
    # Complete-looking arguments must never authorize an effect before the
    # API terminal. The independent witness is the absence of this file.
    args = dict(path="forbidden.txt", content="unexpected write")
    if vendor == "anthropic":
        return (frame(dict(type="message_start", message=dict(model=MODEL, usage=dict(input_tokens=10, output_tokens=0))))
                + frame(dict(type="content_block_start", index=0, content_block=dict(type="tool_use", id="write-a", name="write_file", input={})))
                + frame(dict(type="content_block_delta", index=0, delta=dict(type="input_json_delta", partial_json=json.dumps(args))))
                + frame(dict(type="content_block_stop", index=0))
                + frame(dict(type="message_delta", delta=dict(stop_reason="tool_use"), usage=dict(output_tokens=2))))
    if vendor == "openai":
        return (frame(dict(model=MODEL, choices=[dict(index=0, delta=dict(tool_calls=[dict(index=0, id="write-a", type="function",
                    function=dict(name="write_file", arguments=json.dumps(args)))]), finish_reason="tool_calls")]))
                + frame(dict(model=MODEL, choices=[], usage=dict(prompt_tokens=10, completion_tokens=2))))
    return frame(dict(modelVersion=MODEL, candidates=[dict(index=0, content=dict(role="model", parts=[
                dict(functionCall=dict(id="write-a", name="write_file", args=args))]))],
                usageMetadata=dict(promptTokenCount=10, candidatesTokenCount=2)))


class Exchange:
    def __init__(self, body, content_type, chunk_size):
        self.requests = []
        self.errors = []
        owner = self

        class Handler(http.server.BaseHTTPRequestHandler):
            def log_message(self, *args):
                pass

            def do_POST(self):
                try:
                    owner.requests.append((self.path, json.loads(self.rfile.read(int(self.headers["Content-Length"])))))
                    self.send_response(200)
                    self.send_header("Content-Type", content_type)
                    self.end_headers()
                    for start in range(0, len(body), chunk_size):
                        self.wfile.write(body[start:start + chunk_size])
                        self.wfile.flush()
                except (BrokenPipeError, ConnectionResetError):
                    pass  # A client rejecting this deliberate bad stream may close.
                except Exception as error:
                    owner.errors.append(type(error).__name__ + ": " + str(error))

        self.server = http.server.ThreadingHTTPServer(("127.0.0.1", 0), Handler)
        self.server.daemon_threads = True
        self.thread = threading.Thread(target=lambda: self.server.serve_forever(poll_interval=.01), daemon=True)
        self.thread.start()

    def close(self):
        self.server.shutdown()
        self.server.server_close()
        self.thread.join()


def run(binary, vendor, body, *, plain=False, content_type="text/event-stream", chunk_size=65536):
    exchange = Exchange(body, content_type, chunk_size)
    try:
        with tempfile.TemporaryDirectory(prefix="ch06-wire-") as directory:
            root = pathlib.Path(directory)
            env = {k: v for k, v in os.environ.items() if not k.startswith(
                ("LLM_", "CH02_", "COURSE_", "ANTHROPIC_", "OPENAI_", "GEMINI_", "EN_DISABLE_STREAMING"))}
            env.update(LLM_VENDOR=vendor, LLM_MODEL=MODEL, LLM_RESOLVED_MODEL=MODEL,
                       LLM_API_KEY="LOCAL-ONLY", LLM_BASE_URL=f"http://127.0.0.1:{exchange.server.server_port}",
                       CH02_LOG=str(root / "history.jsonl"), EN_DISABLE_STREAMING="1" if plain else "0")
            result = subprocess.run([str(binary), "protocol", "--observe"], cwd=root, env=env,
                                    input=json.dumps(dict(kind="prompt", text="Exercise this local fixture.")) + "\n",
                                    text=True, stdout=subprocess.PIPE, stderr=subprocess.PIPE, timeout=15)
            records = [json.loads(line) for line in result.stdout.splitlines() if line]
            history = [json.loads(line) for line in (root / "history.jsonl").read_text().splitlines()]
            assert not (root / "forbidden.txt").exists(), "incomplete proposed call changed the filesystem"
            assert not exchange.errors, "local fixture failed: " + repr(exchange.errors)
            assert len(exchange.requests) == 1, "unexpected retry or tool continuation"
            path, request = exchange.requests[0]
            if vendor == "gemini":
                assert (":streamGenerateContent" in path) == (not plain), "wrong delivery endpoint"
            else:
                assert (request.get("stream") is True) == (not plain), "wrong delivery flag"
            sent = [row["request"] for row in history if row.get("type") == "request_sent"]
            assert len(sent) == 1 and sent[0].get("delivery") == ("plain" if plain else "stream"), "recorded delivery differs"
            return dict(records=records, history=history, exit_code=result.returncode)
    finally:
        exchange.close()


def check_result(result, success=True):
    records, history = result["records"], result["history"]
    completions = [row["completion"] for row in records if "completion" in row]
    assert len(completions) == 1, "request did not settle exactly once"
    observations = [row["observation"] for row in records if "observation" in row]
    begins = [row for row in observations if row["kind"] == "model_begin"]
    ends = [row for row in observations if row["kind"] == "model_end"]
    assert len(begins) == len(ends) == 1, "operation did not begin/end exactly once"
    assert ends[0]["accepted"] is success, "incorrect operation acceptance"
    responses = [row for row in history if row.get("type") == "response_ended"]
    finals = [row for row in observations if row["kind"] == "part_final"]
    if not success:
        assert completions[0]["outcome"] == "error", "malformed operation did not fail its request"
        assert not responses and not finals, "failed operation published accepted content"
        assert not any(row.get("type") == "tool_called" for row in history), "failed operation dispatched a tool"
        return None
    assert result["exit_code"] == 0 and completions[0]["outcome"] == "success", "valid operation failed"
    assert len(responses) == 1, "accepted response cardinality"
    response = responses[0]["response"]
    assert response["usage"] == dict(input=10, cache_write=0, cache_read=0, output=2), "normalized usage differs"
    assert [row["part"] for row in finals] == response["parts"], "finals differ from persisted typed parts"
    assert all(row["response_seq"] == responses[0]["seq"] for row in finals), "final mapping differs from actual sequence"
    assert [row["part_index"] for row in finals] == list(range(len(finals))), "final positions differ"
    if begins[0]["delivery"] == "plain":
        assert not any(row["kind"] == "part_delta" for row in observations), "plain delivery emitted fragments"
    return response


def cases(binary, vendor):
    first, tail = fixture(vendor)
    body = first + tail

    def expected_parts(response):
        parts = [dict(type="text", text="Hello.")]
        if vendor == "gemini":
            parts.append(dict(type="text", text="", opaque="fixture-signature",
                              **{"from": dict(vendor="gemini", model=MODEL, surface="generate_content")}))
        assert response["parts"] == parts, "accepted parts differ from the published fixture"

    def equivalence():
        streamed = check_result(run(binary, vendor, body))
        plain = check_result(run(binary, vendor, json.dumps(plain_fixture(vendor)).encode(),
                                       plain=True, content_type="application/json"))
        expected_parts(plain)
        assert streamed == plain, "plain/stream ordered typed response, provenance, stop reason or usage differ"

    yield "plain-stream-equivalence", equivalence
    transformations = {
        "crlf": body.replace(b"\n", b"\r\n"),
        "lone-cr": body.replace(b"\n", b"\r"),
        "leading-bom": b"\xef\xbb\xbf" + body,
        "ignored-fields": b": comment\nid: ignored\nretry: 3\n\n" + body,
        "multiline-data": body.replace(b"data: {", b"data: {\ndata: "),
        "exact-comment-frame": b":" + b"x" * (LIMIT - 3) + b"\n\n" + body,
        "two-exact-comment-frames": (b":" + b"x" * (LIMIT - 3) + b"\n\n") * 2 + body,
    }
    initial, rest = body.split(b"\n\n", 1)
    event = initial + b"\n\n"
    exact = b":" + b"x" * (LIMIT - len(event) - 2) + b"\n" + event
    transformations["exact-data-frame"] = exact + rest
    transformations["exact-data-frame-with-bom"] = b"\xef\xbb\xbf" + exact + rest
    for name, wire in transformations.items():
        def valid(wire=wire):
            value = check_result(run(binary, vendor, wire))
            expected_parts(value)
        yield name, valid
    yield "one-byte-server-writes", lambda: check_result(run(binary, vendor, body, chunk_size=1))
    if vendor == "anthropic":
        cumulative = body.replace(b'"usage": {"output_tokens": 2}',
                                  b'"usage": {"input_tokens": 10, "output_tokens": 2}')
        assert cumulative != body, "cumulative fixture transformation did not apply"
        yield "cumulative-usage-replaces", lambda: check_result(run(binary, vendor, cumulative))
    elif vendor == "gemini":
        lines = first.decode().splitlines()
        value = json.loads(next(line[6:] for line in lines if line.startswith("data: ")))
        value["usageMetadata"] = dict(promptTokenCount=10, candidatesTokenCount=1)
        cumulative = frame(value) + tail
        yield "cumulative-usage-replaces", lambda: check_result(run(binary, vendor, cumulative))
    invalid = {
        "unfinished-final-frame": body[:-1],
        "required-usage-absent": without_usage(vendor, body),
        "comment-frame-overflow": b":" + b"x" * (LIMIT - 2) + b"\n\n" + body,
        "data-frame-overflow": one_byte_over_data_frame(exact) + rest,
        "invalid-utf8-data": body.replace(b"Hel", b"\xff", 1),
        "complete-proposed-call-without-terminal": incomplete_call(vendor),
    }
    if vendor == "gemini":
        invalid["finish-reason-absent"] = body.replace(b'"finishReason": "STOP"', b'"finishReason": ""')
    else:
        invalid["terminal-absent"] = b"".join(events(body)[:-1])
    if vendor in ("openai", "gemini"):
        invalid["returned-model-conflict"] = body.replace(MODEL.encode(), b"fixture-other", 1)
    for name, wire in invalid.items():
        yield name, lambda wire=wire: check_result(run(binary, vendor, wire), False)
    yield "non-sse-success", lambda: check_result(run(binary, vendor, json.dumps(plain_fixture(vendor)).encode(),
                                                     content_type="application/json"), False)


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("binary", type=pathlib.Path)
    parser.add_argument("--vendor", choices=VENDORS, action="append")
    args = parser.parse_args()
    binary = args.binary.resolve(strict=True)
    rows = []
    for vendor in args.vendor or VENDORS:
        for name, action in cases(binary, vendor):
            try:
                action()
                errors = []
            except Exception as error:
                errors = [type(error).__name__ + ": " + str(error)]
            rows.append(dict(id=vendor + "/" + name, passed=not errors, details=errors))
    result = dict(scope="partial wire/typed-equivalence CLI checks; full Chapter 6 acceptance pending",
                  binary_sha256=hashlib.sha256(binary.read_bytes()).hexdigest(),
                  checker_sha256=hashlib.sha256(pathlib.Path(__file__).read_bytes()).hexdigest(),
                  passed=all(row["passed"] for row in rows), checks=rows)
    print(json.dumps(result, indent=2))
    return 0 if result["passed"] else 1


if __name__ == "__main__":
    raise SystemExit(main())
