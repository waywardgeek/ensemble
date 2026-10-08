#!/usr/bin/env python3
"""Chapter 11 protocol-fixture foundation; NOT an Ensemble runtime acceptance gate.

--self-test exercises independent oracle vectors and a real subprocess fixture
peer. --peer is the bounded stdio peer, and --emit DIR exports literal fixtures
for a future public-adapter consumer. No student API is guessed by this module.
"""
import argparse
import copy
from decimal import Decimal
import hashlib
import json
import re
from pathlib import Path
import subprocess
import sys

ROOT = Path(__file__).resolve().parents[2]
VERSION = "2026-07-28"
MAX_MESSAGE = 8 * 1024 * 1024
META = {"io.modelcontextprotocol/protocolVersion": VERSION,
        "io.modelcontextprotocol/clientCapabilities": {},
        "io.modelcontextprotocol/clientInfo": {"name": "Ensemble", "version": "edition-2-ch11"}}
UINT64 = 18446744073709551615


class Refusal(Exception):
    def __init__(self, code):
        super().__init__(code)
        self.code = code


def require(condition, code):
    if not condition:
        raise Refusal(code)


def compact(value):
    return json.dumps(value, ensure_ascii=False, separators=(",", ":")).encode()


def pairs(items):
    result = {}
    for key, value in items:
        require(key not in result, "duplicate-key")
        result[key] = value
    return result


def parse(raw):
    """Fixture oracle for published complete-message limits, not a client parser."""
    require(len(raw) <= MAX_MESSAGE, "message-bytes")
    try:
        value = json.loads(raw.decode("utf-8"), object_pairs_hook=pairs,
                           parse_float=Decimal, parse_int=Decimal,
                           parse_constant=lambda _: (_ for _ in ()).throw(Refusal("json")))
    except (UnicodeError, ValueError, ArithmeticError, RecursionError) as error:
        raise Refusal("json") from error
    require(isinstance(value, dict), "object-envelope")
    count = 0
    stack = [(value, 0)]
    while stack:
        item, depth = stack.pop()
        count += 1
        require(count <= 100000, "json-nodes")
        if isinstance(item, (dict, list)):
            depth += 1
            require(depth <= 64, "json-depth")
            children = item.values() if isinstance(item, dict) else item
            stack.extend((child, depth) for child in children)
        strings = list(item) if isinstance(item, dict) else [item] if isinstance(item, str) else []
        for text in strings:
            try:
                text.encode("utf-8")
            except UnicodeError as error:
                raise Refusal("surrogate") from error
    return value


def ordinal(identity):
    require(isinstance(identity, str) and re.fullmatch(r"rpc-[1-9][0-9]{0,19}", identity), "unknown-id")
    value = int(identity[4:])
    require(value <= UINT64, "unknown-id")
    return value


def correlation(identity, watermark, pending):
    value = ordinal(identity)
    require(value <= watermark, "unknown-id")
    return "pending" if identity in pending else "stale"


def envelope(raw):
    value = parse(raw)
    require(value.get("jsonrpc") == "2.0", "jsonrpc")
    if "method" in value:
        require(set(value) <= {"jsonrpc", "method", "params", "id"} and isinstance(value["method"], str), "envelope-fields")
        return "reverse-request" if "id" in value else "notification", value
    require(set(value) in ({"jsonrpc", "id", "result"}, {"jsonrpc", "id", "error"}), "envelope-fields")
    ordinal(value["id"])
    if "error" in value:
        error = value["error"]
        require(isinstance(error, dict) and {"code", "message"} <= set(error) <= {"code", "message", "data"}, "error-shape")
        code = error["code"]
        require(isinstance(code, Decimal) and Decimal(-2147483648) <= code <= Decimal(2147483647) and code == code.to_integral_value(), "error-code")
        require(isinstance(error["message"], str), "error-shape")
        return "remote-error", value
    result = value["result"]
    require(isinstance(result, dict), "result-object")
    kind = result.get("resultType", "complete")
    require(kind in ("complete", "input_required"), "result-type")
    return "unsupported-result" if kind == "input_required" else "complete", value


def canonical(value):
    """Exact Chapter 10 encoding for the bounded fixture corpus used here."""
    if isinstance(value, int) and not isinstance(value, bool):
        value = Decimal(value)
    if isinstance(value, Decimal):
        sign, digits, exponent = value.as_tuple()
        coefficient = ''.join(map(str, digits)).lstrip('0')
        if not coefficient:
            return b"0"
        removed = len(coefficient) - len(coefficient.rstrip('0'))
        coefficient = coefficient.rstrip('0')
        exponent += removed
        return (("-" if sign else "") + coefficient + ("e" + str(exponent) if exponent else "")).encode()
    if isinstance(value, dict):
        return b"{" + b",".join(compact(k) + b":" + canonical(value[k]) for k in sorted(value, key=lambda k: k.encode())) + b"}"
    if isinstance(value, list):
        return b"[" + b",".join(map(canonical, value)) + b"]"
    return compact(value)


def normalize(descriptor):
    require(isinstance(descriptor, dict) and {"name", "inputSchema"} <= set(descriptor) <= {"name", "description", "inputSchema", "outputSchema", "title", "icons", "annotations", "_meta"}, "descriptor-fields")
    require(isinstance(descriptor["name"], str) and re.fullmatch(r"[A-Za-z0-9_.-]{1,128}", descriptor["name"]), "remote-name")
    description = descriptor.get("description", "")
    require(isinstance(description, str) and len(description.encode()) <= 65536, "description-bytes")
    schema = descriptor["inputSchema"]
    require(isinstance(schema, dict) and schema.get("type") == "object", "input-root")
    for key in ["inputSchema", "outputSchema"]:
        if key in descriptor:
            require(isinstance(descriptor[key], (dict, bool)), "schema-shape")
            require(len(canonical(descriptor[key])) <= 256 * 1024, "schema-bytes")
    return {"name": descriptor["name"], "description": description, "inputSchema": schema, "outputSchema": descriptor.get("outputSchema")}


class Discovery:
    """Review oracle owns only its candidate; no ready state before final page."""
    def __init__(self):
        self.candidate, self.cursors = {}, set()
        self.pages = self.bytes = 0
        self.published = None

    def page(self, result):
        self.pages += 1
        require(self.pages <= 64, "discovery-pages")
        require(isinstance(result.get("tools"), list), "tools-array")
        for raw in result["tools"]:
            tool = normalize(raw)
            require(tool["name"] not in self.candidate, "duplicate-tool")
            self.candidate[tool["name"]] = tool
            self.bytes += len(canonical(tool))
            require(len(self.candidate) <= 1024, "discovery-tools")
            require(self.bytes <= 16 * 1024 * 1024, "discovery-bytes")
        if "nextCursor" in result:
            cursor = result["nextCursor"]
            require(isinstance(cursor, str) and 0 < len(cursor.encode()) <= 4096, "cursor")
            require(cursor not in self.cursors, "cursor-repeat")
            self.cursors.add(cursor)
            return cursor
        self.published = copy.deepcopy(self.candidate)
        return None


def selected_compatible(frozen, replacement):
    return all(name in replacement and canonical(value) == canonical(replacement[name]) for name, value in frozen.items())


def request(n, method, **params):
    return {"jsonrpc": "2.0", "id": "rpc-" + str(n), "method": method, "params": {**params, "_meta": copy.deepcopy(META)}}


def cancellation(identity):
    return {"jsonrpc": "2.0", "method": "notifications/cancelled", "params": {"requestId": identity, "reason": "Cancelled by client"}}


def descriptor():
    return {"name": "notes.append", "description": "Append a note.", "inputSchema": {"type": "object", "properties": {"text": {"type": "string", "maxLength": 2000}}, "required": ["text"], "additionalProperties": False}}


class Peer:
    """Only a fixture peer. Client authority and lifecycle are not implemented here."""
    def __init__(self):
        self.methods, self.cancelled = [], []

    def accept(self, raw):
        message = parse(raw)
        require(message.get("jsonrpc") == "2.0", "jsonrpc")
        if "id" not in message:
            require(message == cancellation(message.get("params", {}).get("requestId")), "cancellation-bytes")
            self.cancelled.append(message["params"]["requestId"])
            return None
        identity, method = message["id"], message.get("method")
        ordinal(identity)
        params = message.get("params")
        require(isinstance(params, dict) and params.get("_meta") == META, "request-metadata")
        self.methods.append(method)
        if method == "server/discover":
            result = {"resultType": "complete", "supportedVersions": [VERSION], "capabilities": {"tools": {}}}
        elif method == "tools/list":
            result = {"resultType": "complete", "tools": [descriptor()]}
        elif method == "tools/call":
            require(params.get("name") == "notes.append" and isinstance(params.get("arguments", {}).get("text"), str), "fixture-call")
            text = params["arguments"]["text"]
            result = {"resultType": "complete", "content": [{"type": "text", "text": text}], "structuredContent": {"written": len(text)}, "isError": False}
        else:
            raise Refusal("fixture-method")
        return compact({"jsonrpc": "2.0", "id": identity, "result": result})


def framed(line):
    require(line.endswith(b"\n"), "partial-eof")
    raw = line[:-1]
    require(bool(raw), "empty-line")
    parse(raw)
    return raw


def peer_main():
    peer = Peer()
    try:
        while True:
            line = sys.stdin.buffer.readline(MAX_MESSAGE + 2)
            if not line:
                return 0
            raw = framed(line)
            response = peer.accept(raw)
            if response is not None:
                sys.stdout.buffer.write(response + b"\n")
                sys.stdout.buffer.flush()
    except Refusal as error:
        print(error.code, file=sys.stderr)
        return 2


def fixtures():
    return [request(1, "server/discover"), request(2, "tools/list"), request(3, "tools/call", name="notes.append", arguments={"text": "checked\nline"}), cancellation("rpc-3")]


def self_test():
    rows = []
    def positive(name, operation):
        operation()
        rows.append({"id": name, "passed": True})
    def negative(name, expected, operation):
        try:
            operation()
        except Refusal as error:
            require(error.code == expected, "wrong-negative:" + name + ":" + error.code)
        else:
            raise Refusal("negative-accepted:" + name)
        rows.append({"id": name, "expected_refusal": expected, "passed": True})
    good = b'{"jsonrpc":"2.0","id":"rpc-3","result":{}}'
    positive("response-positive", lambda: require(envelope(good)[0] == "complete", "control"))
    for name, raw, code in [
        ("duplicate-key", b'{"jsonrpc":"2.0","id":"rpc-3","id":"rpc-3","result":{}}', "duplicate-key"),
        ("trailing", good + b'{}', "json"), ("invalid-utf8", good + b'\xff', "json"),
        ("unpaired-surrogate", b'{"x":"\\ud800"}', "surrogate"),
        ("unknown-member", good[:-1] + b',"extra":0}', "envelope-fields"),
        ("result-and-error", good[:-1] + b',"error":{"code":0,"message":"x"}}', "envelope-fields"),
        ("result-scalar", b'{"jsonrpc":"2.0","id":"rpc-3","result":1}', "result-object"),
        ("unknown-result-type", b'{"jsonrpc":"2.0","id":"rpc-3","result":{"resultType":"other"}}', "result-type"),
    ]:
        negative(name, code, lambda raw=raw: envelope(raw))
    positive("supported-unsupported-result", lambda: require(envelope(b'{"jsonrpc":"2.0","id":"rpc-3","result":{"resultType":"input_required"}}')[0] == "unsupported-result", "control"))
    for token in ["-2147483648", "2147483647", "2.0", "20e-1"]:
        positive("error-code-" + token, lambda token=token: envelope(('{"jsonrpc":"2.0","id":"rpc-3","error":{"code":' + token + ',"message":"x","data":{"ignored":true}}}').encode()))
    for token in ["-2147483649", "2147483648", "1.1", "1e1000000", '"1"', "true"]:
        negative("bad-error-code-" + token, "error-code", lambda token=token: envelope(('{"jsonrpc":"2.0","id":"rpc-3","error":{"code":' + token + ',"message":"x"}}').encode()))
    positive("message-exact", lambda: parse(good + b' ' * (MAX_MESSAGE - len(good))))
    negative("message-plus-one", "message-bytes", lambda: parse(good + b' ' * (MAX_MESSAGE + 1 - len(good))))
    positive("depth-exact", lambda: parse(b'{"x":' + b'[' * 63 + b'0' + b']' * 63 + b'}'))
    negative("depth-plus-one", "json-depth", lambda: parse(b'{"x":' + b'[' * 64 + b'0' + b']' * 64 + b'}'))
    positive("nodes-exact", lambda: parse(b'{"x":[' + b','.join([b'0'] * 99998) + b']}'))
    negative("nodes-plus-one", "json-nodes", lambda: parse(b'{"x":[' + b','.join([b'0'] * 99999) + b']}'))
    positive("framing-lf-excluded", lambda: framed(good + b' ' * (MAX_MESSAGE - len(good)) + b'\n'))
    negative("framing-partial-eof", "partial-eof", lambda: framed(good))
    negative("framing-empty", "empty-line", lambda: framed(b'\n'))
    positive("correlation-out-of-order", lambda: require([correlation(i, 4, {"rpc-3", "rpc-4"}) for i in ["rpc-4", "rpc-3"]] == ["pending", "pending"], "control"))
    positive("correlation-stale", lambda: require(correlation("rpc-3", 4, {"rpc-4"}) == "stale", "control"))
    positive("correlation-max", lambda: correlation("rpc-" + str(UINT64), UINT64, set()))
    for identity in ["rpc-5", "rpc-03", "rpc-0", "rpc-+3", "other-3", 3, "rpc-18446744073709551616"]:
        negative("unknown-" + str(identity), "unknown-id", lambda identity=identity: correlation(identity, 4, {"rpc-3"}))
    positive("reverse-request-distinct", lambda: require(envelope(b'{"jsonrpc":"2.0","id":"server-id","method":"tools/call","params":{}}')[0] == "reverse-request", "control"))
    positive("notification-distinct", lambda: require(envelope(b'{"jsonrpc":"2.0","method":"progress","params":{}}')[0] == "notification", "control"))
    first = {"tools": [descriptor()], "nextCursor": "second"}
    d = Discovery()
    positive("page-one-unpublished", lambda: (d.page(first), require(d.published is None, "control")))
    positive("page-two-publishes", lambda: (d.page({"tools": []}), require(list(d.published) == ["notes.append"], "control")))
    for name, second, code in [("duplicate-across-pages", {"tools": [descriptor()]}, "duplicate-tool"), ("cursor-repeat", {"tools": [], "nextCursor": "second"}, "cursor-repeat"), ("empty-cursor", {"tools": [], "nextCursor": ""}, "cursor")]:
        candidate = Discovery(); candidate.page(first)
        negative(name, code, lambda candidate=candidate, second=second: candidate.page(second))
        require(candidate.published is None, "partial-publication")
    positive("cursor-exact", lambda: Discovery().page({"tools": [], "nextCursor": "x" * 4096}))
    negative("cursor-plus-one", "cursor", lambda: Discovery().page({"tools": [], "nextCursor": "x" * 4097}))
    pages = Discovery()
    for index in range(63):
        pages.page({"tools": [], "nextCursor": str(index)})
    positive("pages-exact", lambda: (pages.page({"tools": []}), require(pages.published == {}, "control")))
    pages = Discovery()
    for index in range(64):
        pages.page({"tools": [], "nextCursor": str(index)})
    negative("pages-plus-one", "discovery-pages", lambda: pages.page({"tools": []}))
    require(pages.published is None, "partial-publication")
    many = [{**descriptor(), "name": "tool-" + str(index)} for index in range(1024)]
    tools = Discovery()
    tools.page({"tools": many})
    positive("tools-exact", lambda: require(len(tools.published) == 1024, "control"))
    negative("tools-plus-one", "discovery-tools", lambda: Discovery().page({"tools": many + [{**descriptor(), "name": "tool-extra"}]}))
    positive("description-exact", lambda: normalize({**descriptor(), "description": "x" * 65536}))
    negative("description-plus-one", "description-bytes", lambda: normalize({**descriptor(), "description": "x" * 65537}))
    baseline_schema = {"type": "object", "description": ""}
    padding = 256 * 1024 - len(canonical(baseline_schema))
    positive("schema-bytes-exact", lambda: normalize({**descriptor(), "inputSchema": {**baseline_schema, "description": "x" * padding}}))
    negative("schema-bytes-plus-one", "schema-bytes", lambda: normalize({**descriptor(), "inputSchema": {**baseline_schema, "description": "x" * (padding + 1)}}))
    normalized = normalize(parse(compact(descriptor())))
    frozen = {"notes.append": normalized}
    same = copy.deepcopy(normalized); same["inputSchema"]["properties"]["text"]["maxLength"] = Decimal("2000.00")
    positive("frozen-semantic-number", lambda: require(selected_compatible(frozen, {"notes.append": same, "unselected": {}}), "control"))
    for field in ["description", "inputSchema", "outputSchema"]:
        changed = copy.deepcopy(normalized)
        changed[field] = "changed" if field == "description" else {"type": "boolean"}
        positive("frozen-refuses-" + field, lambda changed=changed: require(not selected_compatible(frozen, {"notes.append": changed}), "control"))
    positive("frozen-precision", lambda: require(canonical(parse(b'{"n":9007199254740992}')) != canonical(parse(b'{"n":9007199254740993}')), "control"))
    positive("canonical-fixture-int-equals-wire-number", lambda: require(canonical({"n": 2000}) == canonical(parse(b'{"n":2000.00}')) == b'{"n":2e3}', "control"))
    positive("canonical-huge-exponent", lambda: require(canonical(parse(b'{"n":1e1000000}')) == b'{"n":1e1000000}', "control"))
    messages = fixtures()
    expected_cancel = b'{"jsonrpc":"2.0","method":"notifications/cancelled","params":{"requestId":"rpc-3","reason":"Cancelled by client"}}'
    positive("cancellation-literal", lambda: require(compact(messages[-1]) == expected_cancel, "control"))
    memory_peer = Peer()
    memory = [reply for message in messages if (reply := memory_peer.accept(compact(message))) is not None]
    positive("memory-fixture-sequence", lambda: require(memory_peer.methods == ["server/discover", "tools/list", "tools/call"] and memory_peer.cancelled == ["rpc-3"], "control"))
    bad = copy.deepcopy(messages[0]); bad["params"]["_meta"].pop("io.modelcontextprotocol/clientInfo")
    negative("missing-request-metadata", "request-metadata", lambda: Peer().accept(compact(bad)))
    child = subprocess.run([sys.executable, str(Path(__file__).resolve()), "--peer"], input=b''.join(compact(m) + b'\n' for m in messages), capture_output=True, timeout=5)
    positive("stdio-peer-positive-and-eof", lambda: require(child.returncode == 0 and child.stdout.splitlines() == memory and child.stderr == b'', "control"))
    for name, raw, reason in [("stdio-peer-partial", compact(messages[0]), "partial-eof"), ("stdio-peer-empty", b'\n', "empty-line")]:
        child = subprocess.run([sys.executable, str(Path(__file__).resolve()), "--peer"], input=raw, capture_output=True, timeout=5)
        positive(name, lambda child=child, reason=reason: require(child.returncode == 2 and child.stdout == b'' and child.stderr.strip() == reason.encode(), "control"))
    return rows


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    actions = parser.add_mutually_exclusive_group(required=True)
    actions.add_argument("--self-test", action="store_true")
    actions.add_argument("--peer", action="store_true")
    actions.add_argument("--emit", type=Path)
    parser.add_argument("--receipt", type=Path)
    args = parser.parse_args()
    if args.peer:
        return peer_main()
    script_hash = hashlib.sha256(Path(__file__).read_bytes()).hexdigest()
    contract = ROOT / "book/edition-2/chapter-11.md"
    contract_hash = hashlib.sha256(contract.read_bytes()).hexdigest()
    if args.emit:
        args.emit.mkdir()
        peer = Peer()
        for n, message in enumerate(fixtures(), 1):
            raw = compact(message)
            (args.emit / f"{n:02}-request.json").write_bytes(raw)
            response = peer.accept(raw)
            if response is not None:
                (args.emit / f"{n:02}-response.json").write_bytes(response)
        result = {"scope": "literal protocol fixture export", "runtime_acceptance": False, "directory": str(args.emit),
                  "files": {p.name: hashlib.sha256(p.read_bytes()).hexdigest() for p in sorted(args.emit.iterdir())}}
    else:
        rows = self_test()
        result = {"scope": "protocol oracle and fixture-peer controls only", "runtime_acceptance": False, "checks": rows, "passed": len(rows), "total": len(rows)}
    require(script_hash == hashlib.sha256(Path(__file__).read_bytes()).hexdigest(), "checker-changed")
    require(contract_hash == hashlib.sha256(contract.read_bytes()).hexdigest(), "contract-changed")
    result.update(checker_sha256=script_hash, contract_sha256=contract_hash)
    if args.receipt:
        args.receipt.write_text(json.dumps(result, indent=2) + "\n")
    print(json.dumps(result, indent=2))
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
