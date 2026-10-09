#!/usr/bin/env python3
"""Transport-neutral Chapter 11 fixture scripts, not runtime acceptance.

The future public consumer feeds these same complete messages through stdio,
the delivered memory adapter and a separately implemented public transport.
No student API spelling is encoded here. Raw message strings preserve number
tokens; comparison of client requests permits only semantic JSON equivalence.
"""
import argparse
import copy
import hashlib
import json
from pathlib import Path
import sys

from accept_ch11 import MAX_MESSAGE, Peer, Refusal, canonical, compact, framed, parse, request, cancellation

ROOT = Path(__file__).resolve().parents[2]


def wire(value):
    return compact(value).decode()


def response(identity, result):
    return wire({"jsonrpc": "2.0", "id": identity, "result": result})


def step(expected, replies=()):
    return {"request": wire(expected), "responses": list(replies)}


def call(number, text):
    return request(number, "tools/call", name="notes.append", arguments={"text": text})


def result(text):
    return {"content": [{"type": "text", "text": text}], "structuredContent": {"written": len(text)}}


def discovery():
    peer = Peer()
    return [step(request(n, method), [peer.accept(compact(request(n, method))).decode()])
            for n, method in [(1, "server/discover"), (2, "tools/list")]]


def scenarios():
    """Seven finite scripts; future runtime tests must inspect actual Job effects."""
    base = discovery()
    first, second = call(3, "first"), call(4, "second")
    a, b = response("rpc-3", result("first")), response("rpc-4", result("second"))
    return {
        "ordinary": {"steps": base + [step(first, [a])],
                     "expectation": "One ordinary successful Job/result; first text retained."},
        "out-of-order": {"steps": base + [step(first), step(second, [b, a])],
                         "expectation": "Two admitted calls; each Job gets only its matching result despite reverse reply order."},
        "cancel-late-sibling": {"steps": base + [step(first), step(second), step(cancellation("rpc-3"), [a, b])],
                               "expectation": "Kill first only after both requests observed; first remains killed, stale reply cannot replace it, second succeeds on same generation."},
        "remote-error": {"steps": base + [step(first, [wire({"jsonrpc": "2.0", "id": "rpc-3", "error": {"code": -32001, "message": "CH11_REMOTE_PRIVATE", "data": {"private": "CH11_REMOTE_PRIVATE"}}})]), step(second, [b])],
                         "expectation": "First done/error with mcp_remote_error and decimal -32001 but no private data; subsequent call succeeds without reopening."},
        "future-id": {"steps": base + [step(first, [response("rpc-4", result("future"))])], "optional_tail": [wire(cancellation("rpc-3"))],
                      "expectation": "Connection faults; pending call fails safely; no result or automatic retry for unissued rpc-4."},
        "stale-duplicate": {"steps": base + [step(first, [a]), step(second, [a, b])],
                            "expectation": "Settled rpc-3 duplicate discarded; next pending Job receives second exactly once."},
        "reverse-request": {"steps": base + [step(first, [wire({"jsonrpc": "2.0", "id": "peer-1", "method": "tools/call", "params": {"name": "write_file", "arguments": {"path": "must-not-exist", "content": "forbidden"}}})])], "optional_tail": [wire(cancellation("rpc-3"))],
                            "expectation": "Connection faults without dispatching a local handler, writing the requested file or sending an Agent response."},
    }


def schema_vectors():
    """Paired known examples; this file does not claim to implement JSON Schema."""
    return [
        {"id": "local-pointer-sibling", "schema": '{"type":"object","$defs":{"a/b":{"type":"integer"}},"properties":{"n":{"$ref":"#/$defs/a~1b","minimum":2}},"required":["n"],"additionalProperties":false}',
         "accept": ['{"n":2}', '{"n":2.0}'], "reject": ['{"n":1}', '{"n":2,"extra":0}']},
        {"id": "precision", "schema": '{"type":"object","properties":{"n":{"type":"integer","minimum":9007199254740993,"maximum":9007199254740993}},"required":["n"]}',
         "accept": ['{"n":9007199254740993}', '{"n":90071992547409930e-1}'], "reject": ['{"n":9007199254740992}', '{"n":9007199254740994}']},
        {"id": "scalar-array-count", "schema": '{"type":"object","properties":{"v":{"type":"array","minItems":1,"maxItems":1,"items":{"type":"string","minLength":1,"maxLength":1}}},"required":["v"],"minProperties":1,"maxProperties":1}',
         "accept": ['{"v":["😀"]}'], "reject": ['{"v":[]}', '{"v":["😀x"]}', '{"v":["x"],"extra":true}']},
        {"id": "composition-enum", "schema": '{"type":"object","properties":{"n":{"allOf":[{"type":["integer","null"]},{"anyOf":[{"const":null},{"enum":[1,2,3]}]},{"not":{"const":2}},{"oneOf":[{"type":"null"},{"minimum":1,"type":"integer"}]}]}},"required":["n"]}',
         "accept": ['{"n":null}', '{"n":1}', '{"n":3.0}'], "reject": ['{"n":2}', '{"n":4}', '{"n":"1"}']},
        {"id": "numeric-open-interval", "schema": '{"type":"object","properties":{"n":{"exclusiveMinimum":1,"exclusiveMaximum":2}},"required":["n"]}',
         "accept": ['{"n":1.5}'], "reject": ['{"n":1}', '{"n":2}']},
        {"id": "annotations-do-not-insert", "schema": '{"type":"object","$schema":"https://json-schema.org/draft/2020-12/schema","title":"Example","description":"Data only","properties":{"n":{"type":"integer","default":3,"examples":[3]}},"required":["n"]}',
         "accept": ['{"n":3}'], "reject": ['{}']},
    ]


def preparation_vectors():
    positive = {"type": "object", "properties": {"n": {"type": "integer"}}}
    return [{"id": name, "positive": wire(positive), "reject": wire(schema)} for name, schema in [
        ("unsupported-keyword", dict(positive, pattern="x")),
        ("external-reference", dict(positive, **{"$ref": "https://127.0.0.1:1/never-fetch"})),
        ("unresolved-local-reference", dict(positive, **{"$ref": "#/$defs/missing"})),
        ("containment-reference-cycle", {"type": "object", "properties": {"n": {"$ref": "#"}}}),
        ("semantic-enum-duplicate", {"type": "object", "properties": {"n": {"enum": [1, 1.0]}}}),
    ]]


class Script:
    def __init__(self, spec):
        self.steps = copy.deepcopy(spec["steps"])
        self.position = 0
        self.optional_tail = list(spec.get("optional_tail", []))

    def accept(self, raw):
        if self.position == len(self.steps):
            # Fault teardown may abandon the outstanding call before closure.
            # Neither require that cleanup notice nor reject its valid delivery.
            if raw.decode("utf-8") in self.optional_tail:
                self.optional_tail.remove(raw.decode("utf-8"))
                return []
            raise Refusal("unexpected-extra-message")
        current = self.steps[self.position]
        expected = current["request"].encode()
        actual = parse(raw)
        if canonical(actual) != canonical(parse(expected)):
            raise Refusal("script-request-mismatch")
        if actual.get("method") == "notifications/cancelled" and raw != expected:
            raise Refusal("cancellation-literal-bytes")
        self.position += 1
        return [r.encode() for r in current["responses"]]

    def finish(self):
        if self.position != len(self.steps):
            raise Refusal("script-incomplete")


def serve(name, receipt):
    # This peer's receipt is separate from stdout; its records prove peer inputs,
    # not client outcomes. Runtime acceptance must additionally inspect the client.
    script = Script(scenarios()[name])
    with receipt.open("x") as output:
        def record(value):
            output.write(json.dumps(value, separators=(",", ":")) + "\n")
            output.flush()
        try:
            while True:
                line = sys.stdin.buffer.readline(MAX_MESSAGE + 2)
                if not line:
                    script.finish()
                    record({"fixture_complete": True, "runtime_acceptance": False})
                    return 0
                raw = framed(line)
                record({"incoming_hex": raw.hex()})
                for answer in script.accept(raw):
                    record({"outgoing_hex": answer.hex()})
                    sys.stdout.buffer.write(answer + b"\n")
                    sys.stdout.buffer.flush()
        except Refusal as error:
            record({"fixture_failure": error.code})
            return 2


def bundle():
    return {"runtime_acceptance": False, "purpose": "fixture inputs for later independent public integration",
            "plan_revision": "55c9e7a44364123dd28ef7967a0a3663d74de925",
            "scenarios": scenarios(), "argument_vectors": schema_vectors(),
            "preparation_vectors": preparation_vectors(),
            "remaining": "Full chapter-11-grader-review §11.10 map remains; these vectors do not prove implementation coverage."}


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    modes = parser.add_mutually_exclusive_group(required=True)
    modes.add_argument("--emit", type=Path)
    modes.add_argument("--peer", choices=sorted(scenarios()))
    parser.add_argument("--receipt", type=Path)
    args = parser.parse_args()
    if args.peer:
        if args.receipt is None:
            parser.error("--peer requires an absent --receipt")
        return serve(args.peer, args.receipt)
    value = bundle()
    value["fixture_sha256"] = hashlib.sha256(Path(__file__).read_bytes()).hexdigest()
    value["contract_sha256"] = hashlib.sha256((ROOT / "book/edition-2/chapter-11.md").read_bytes()).hexdigest()
    with args.emit.open("x") as output:
        json.dump(value, output, indent=2, ensure_ascii=False)
        output.write("\n")
    return 0


if __name__ == "__main__":
    sys.exit(main())
