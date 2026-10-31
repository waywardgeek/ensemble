#!/usr/bin/env python3
"""Grader-owned stdio fixture; real bounded file effect, no model/backend calls."""
import argparse
import json
import os
from pathlib import Path
import sys

from accept_ch11 import MAX_MESSAGE, Peer, Refusal, compact, framed


def serve(directory, mixed=False):
    peer = Peer()
    with (directory / "peer.jsonl").open("x") as receipt:
        def record(value):
            receipt.write(json.dumps(value, separators=(",", ":")) + "\n")
            receipt.flush()

        record({"startup": True, "pid": os.getpid(), "cwd": os.getcwd(),
                "allowed": os.environ.get("CH11_ALLOWED"),
                "unlisted_present": "CH11_UNLISTED" in os.environ,
                "credential_present": "LLM_API_KEY" in os.environ})
        try:
            while True:
                line = sys.stdin.buffer.readline(MAX_MESSAGE + 2)
                if not line:
                    record({"eof": True})
                    return 0
                raw = framed(line)
                record({"request": raw.decode("utf-8")})
                answer = peer.accept(raw)
                message = json.loads(raw)
                if message.get("method") == "tools/call":
                    args = message["params"]["arguments"]
                    if set(args) != {"text"} or len(args["text"]) > 2000:
                        raise Refusal("fixture-arguments")
                    data = args["text"].encode("utf-8")
                    with (directory / "notebook.txt").open("ab") as notebook:
                        notebook.write(data)
                    value = json.loads(answer)
                    value["result"]["_meta"] = {"secret": "CH11_IGNORED_METADATA"}
                    value["result"]["content"][0]["annotations"] = {"audience": ["user"]}
                    if mixed:
                        value["result"]["content"].append({"type": "image", "data": "AA==", "mimeType": "image/png"})
                    answer = compact(value)
                    record({"written_hex": data.hex()})
                if answer is not None:
                    record({"response": answer.decode("utf-8")})
                    sys.stdout.buffer.write(answer + b"\n")
                    sys.stdout.buffer.flush()
        except (Refusal, ValueError) as error:
            record({"failure": getattr(error, "code", str(error))})
            return 2


if __name__ == "__main__":
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("directory", type=Path)
    parser.add_argument("--mixed", action="store_true")
    options = parser.parse_args()
    sys.exit(serve(options.directory, options.mixed))
