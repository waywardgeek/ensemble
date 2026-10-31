#!/usr/bin/env python3
"""External-consumer Chapter 4 checks; run only after the initial source freeze.

The adapter is built against the public library, outside its module. No internal
package imports or private field names are used. Prepared against accepted Ch3
declarations; method spelling may be adapted after freeze without changing the
behavioral assertions. Local fake providers; no paid model use.
"""

import argparse
import hashlib
import json
import os
import pathlib
import subprocess
import tempfile
import threading
import time

from accept_ch04 import (Session, TOOLS, VENDORS, MODEL, NARRATION, call, response,
                         require, result_map, result_text, lifecycle_errors,
                         fixture_command, artifact, process_alive)


class Consumer:
    def __init__(self, binary, root):
        self.proc = subprocess.Popen([str(binary)], cwd=root, stdin=subprocess.PIPE,
                                     stdout=subprocess.PIPE, stderr=subprocess.PIPE,
                                     text=True)
        self.condition = threading.Condition()
        self.messages, self.diagnostics = [], []
        self.counter = 0

        def read():
            for line in self.proc.stdout:
                with self.condition:
                    self.messages.append(json.loads(line))
                    self.condition.notify_all()
        self.reader = threading.Thread(target=read, daemon=True)
        self.reader.start()
        self.err_reader = threading.Thread(target=lambda: self.diagnostics.extend(self.proc.stderr), daemon=True)
        self.err_reader.start()

    def await_message(self, predicate, seconds=8):
        end = time.monotonic() + seconds
        with self.condition:
            while True:
                for item in self.messages:
                    if predicate(item):
                        return item
                remaining = end - time.monotonic()
                if remaining <= 0:
                    raise TimeoutError("public consumer boundary timed out")
                self.condition.wait(remaining)

    def request(self, operation, **values):
        self.counter += 1
        identifier = str(self.counter)
        self.proc.stdin.write(json.dumps(dict(values, ID=identifier, Operation=operation)) + "\n")
        self.proc.stdin.flush()
        return self.await_message(lambda m: m.get("id") == identifier)

    def close(self):
        self.proc.stdin.close()
        try:
            self.proc.wait(timeout=5)
        except subprocess.TimeoutExpired:
            self.proc.kill()
            self.proc.wait()
        self.reader.join(timeout=1)
        self.err_reader.join(timeout=1)
        self.proc.stdout.close()
        self.proc.stderr.close()


def agent_config(session, root, builtins):
    config = dict(APIKey="LOCAL-FIXTURE-KEY", Model=MODEL, ResolvedModel=MODEL,
                  Vendor=session.vendor, BaseURL=f"http://127.0.0.1:{session.server.server_port}",
                  Workspace=str(root), LogPath=str(session.log))
    if builtins is not None:
        config["Builtins"] = builtins
    return config


def owner_isolation(binary, vendor):
    with tempfile.TemporaryDirectory(prefix="ensemble-ch04-public-") as tmp:
        root = pathlib.Path(tmp)
        (root / "notes").write_text("ABCDEFGHIJ")
        gated = fixture_command(root, "gated")
        command_a = "echo $$ > a.pid; printf AREADY; while :; do sleep 1; done"
        command_b = "echo $$ > b.pid; printf BREADY; while :; do sleep 1; done"
        batches_a = [
            [call("a-start", "run_command", command=gated, ai_callback_pattern="READY", ai_callback_delay=2),
             call("a-set", "tool_limits", max_output_bytes=4)],
            [call("a-read", "read_file", path="notes"), call("a-poll", "wait_for_job", handle=1, ai_callback_delay=0)],
            [call("a-running", "run_command", command=command_a, ai_callback_pattern="AREADY", ai_callback_delay=2)],
        ]
        batches_b = [
            [call("foreign-wait", "wait_for_job", handle=1),
             call("foreign-send", "send_input", handle=1, input="BAD"),
             call("foreign-kill", "kill_job", handle=1), call("b-read", "read_file", path="notes")],
            [call("b-running", "run_command", command=command_b, ai_callback_pattern="BREADY", ai_callback_delay=2)],
        ]
        def replies(batches):
            return [item for batch in batches for item in (response(vendor, batch, NARRATION), response(vendor, []))]
        consumer = Consumer(binary, root)
        errors, observed, agent_ids = [], {}, {}

        def parked(session, index):
            if index != 3:
                return
            (root / "release-job").touch()
            item = consumer.await_message(lambda m: m.get("observation", {}).get("agent_id") == agent_ids["a"]
                                          and m["observation"].get("kind") == "job_ended"
                                          and m["observation"]["event"].get("job", {}).get("handle") == 1)
            observed["terminal_while_http_parked"] = item["observation"]["seq"]

        sessions = [Session(binary, root, vendor, replies(batches_a), before_reply=parked),
                    Session(binary, root, vendor, replies(batches_b)),
                    Session(binary, root, vendor, [response(vendor, [call("disabled", "read_file", path="notes")], NARRATION), response(vendor, [])])]
        try:
            for label, session in zip(("a", "b", "c"), sessions):
                session.log = root / (label + ".jsonl")
                created = consumer.request("new", Agent=label, Config=agent_config(session, root, TOOLS if label != "c" else None))
                require("error" not in created, "public Agent creation failed: " + str(created.get("error")), errors)
                agent_ids[label] = created["agent_id"]
            for label in ("a", "b", "a", "c"):
                reply = consumer.request("prompt", Agent=label, Prompt="Run the next fixture step.")
                require("error" not in reply, "public request failed: " + str(reply.get("error")), errors)
            events_a = consumer.request("events", Agent="a")["events"]
            events_b = consumer.request("events", Agent="b")["events"]
            events_c = consumer.request("events", Agent="c")["events"]
            a, b, c = map(result_map, (events_a, events_b, events_c))
            require(a["a-start"]["job"]["handle"] == 1 and b["b-read"]["job"]["handle"] == 2
                    and a["a-read"]["job"]["handle"] == 3, "handles not application-wide and monotone", errors)
            for name in ("foreign-wait", "foreign-send", "foreign-kill"):
                require(b[name].get("is_error") is True, "foreign handle accepted by " + name, errors)
            require("ABCDEFGHIJ" in result_text(b, "b-read") and "tool_limits" not in result_text(b, "b-read"),
                    "pending settings leaked between Agents", errors)
            limited = result_text(a, "a-read")
            require("tool_limits" in limited and "AB" in limited and "IJ" in limited and "CDEFGH" not in limited,
                    "Agent's pending settings lost or improperly consumed", errors)
            require(c["disabled"].get("is_error") is True and not any(e.get("tool", {}).get("job") for e in events_c),
                    "omitted public tool selection enabled execution", errors)
            require(bool(observed), "public terminal observation waited behind HTTP", errors)
            require(b"COMPLETED-DURING-HTTP" in artifact(root, 1), "foreign operations disturbed owning job", errors)
            require(artifact(root, 2) == artifact(root, 3) == b"ABCDEFGHIJ", "shared-workspace artifacts collided", errors)
            for label in ("a", "b"):
                reply = consumer.request("prompt", Agent=label, Prompt="Start the managed blocker.")
                require("error" not in reply, "blocker request failed", errors)
            require("error" not in consumer.request("close-agent", Agent="a"), "Agent close failed", errors)
            require(not process_alive(int((root / "a.pid").read_text())), "Agent close left its process alive", errors)
            require(process_alive(int((root / "b.pid").read_text())), "Agent close killed another Agent's process", errors)
            require("error" not in consumer.request("close-agent", Agent="a"), "Agent close not idempotent", errors)
            refused = consumer.request("prompt", Agent="a", Prompt="Must be refused after close.")
            require("error" in refused, "closed Agent accepted another operation", errors)
            require("error" not in consumer.request("close-all"), "application close failed", errors)
            require(not process_alive(int((root / "b.pid").read_text())), "application close left another Agent alive", errors)
            require("error" not in consumer.request("close-all"), "application close not idempotent", errors)
            for label in ("a", "b", "c"):
                events = consumer.request("events", Agent=label)["events"]
                errors.extend(lifecycle_errors(events, root))
                observations = [m["observation"] for m in consumer.messages
                                if m.get("observation", {}).get("agent_id") == agent_ids[label]]
                require([o["event"] for o in observations] == events, "public observations missing, reordered or changed for " + label, errors)
            for session in sessions:
                errors.extend(session.errors)
        except (OSError, ValueError, KeyError, TypeError, AttributeError, TimeoutError) as err:
            errors.append(f"fixture failed: {type(err).__name__}: {err}")
        finally:
            consumer.close()
            for session in sessions:
                session.close()
            for name in ("a.pid", "b.pid"):
                path = root / name
                if path.exists():
                    pid = int(path.read_text())
                    if process_alive(pid):
                        os.killpg(pid, 9)
        return {"id": vendor + "/public-owners-isolation-observation", "passed": not errors,
                "details": errors, "boundaries": observed}


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("consumer", type=pathlib.Path, help="compiled external public-consumer adapter")
    parser.add_argument("--vendor", choices=VENDORS, action="append")
    args = parser.parse_args()
    binary = args.consumer.resolve(strict=True)
    checks = [owner_isolation(binary, vendor) for vendor in args.vendor or VENDORS]
    print(json.dumps({"scope": __doc__, "consumer_sha256": hashlib.sha256(binary.read_bytes()).hexdigest(),
                      "checker_sha256": hashlib.sha256(pathlib.Path(__file__).read_bytes()).hexdigest(),
                      "passed": all(c["passed"] for c in checks), "checks": checks}, indent=2))
    return 0 if all(c["passed"] for c in checks) else 1


if __name__ == "__main__":
    raise SystemExit(main())
