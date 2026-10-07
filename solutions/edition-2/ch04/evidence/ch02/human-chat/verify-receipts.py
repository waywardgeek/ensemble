"""Verify saved terminal runs offline; no provider requests or credential output."""
import hashlib
import json
import os
from pathlib import Path
import subprocess
import argparse

HERE = Path(__file__).resolve().parent
MODULE = HERE.parents[2]
parser = argparse.ArgumentParser(description=__doc__)
parser.add_argument("--binary", default="/tmp/ensemble-ed2-ch02-human-chat", help="initial live-tested executable, checked against every launch receipt")
parser.add_argument("--source-revision", default="56dacfad01f71f2a1b20d39bca15846edef41ddd", help="preserved initial human-client source checkpoint")
options = parser.parse_args()
BINARY = options.binary
binary_hash = hashlib.sha256(Path(BINARY).read_bytes()).hexdigest()


def run_cli(args, vendor, model, log):
    env = {k: v for k, v in os.environ.items() if not k.startswith(("LLM_", "ANTHROPIC_", "OPENAI_", "GEMINI_"))}
    env.update(LLM_VENDOR=vendor, LLM_MODEL=model, CH02_LOG=str(log))
    result = subprocess.run([BINARY, *args], env=env, capture_output=True, check=True)
    assert not result.stderr
    return result.stdout


summary = []
for directory in sorted(p for p in HERE.iterdir() if p.is_dir()):
    launch = json.loads((directory / "launch.json").read_text())
    transcript = (directory / "terminal.txt").read_text()
    log = (directory / "session.log").read_bytes()
    rows = [json.loads(line) for line in log.splitlines()]
    events = rows[1:]
    row = {"run": directory.name, "exit_code": launch["exit_code"], "vendor": launch["vendor"], "selected_model": launch["selected_model"], "mode": launch["mode"], "requests": sum(e["type"] == "request_sent" for e in events)}
    assert "You> " in transcript and "Type /help" in transcript
    assert binary_hash == launch["binary_sha256"], "executable does not match initial live-tested binary"
    for path, digest in launch["source_sha256"].items():
        source = subprocess.check_output(["git", "show", options.source_revision + ":" + path], cwd=MODULE)
        assert hashlib.sha256(source).hexdigest() == digest, "source revision does not match launch receipt"
    if directory.name.endswith("-main"):
        assert launch["exit_code"] == 0 and row["requests"] == 3
        responses = [e["response"] for e in events if e["type"] == "response_ended"]
        answers = ["".join(p.get("text", "") for p in r["parts"] if p["type"] == "text") for r in responses]
        assert answers[1].splitlines() == [answers[0], answers[0][::-1]]
        assert answers[2] == "SLASH-OK"
        assert "Assistant:\n" + answers[1] + "\nYou> " in transcript
        for expected in ["Command error", "No tool_returned events", "Recorded ephemeral.", "Final usage:", "/redact FROM TO REASON"]:
            assert expected in transcript
        usage = {k: sum(r["usage"][k] for r in responses) for k in ["input", "cache_write", "cache_read", "output"]}
        assert "Final usage: input={input}, cache write={cache_write}, cache read={cache_read}, output={output}".format(**usage) in transcript
        row.update(answers=answers, usage=usage, identities=[{"requested": r["requested"], "returned": r["from"], "model_reported": r["model_reported"]} for r in responses])
        assert len(events) == 10
        human = [e["message"]["parts"][0]["text"] for e in events if e["type"] == "message_received" and e["message"]["actor"] == "human"]
        assert human[-1].startswith("/help") and not human[-1].startswith("//")
        ephemera = []
        for event in events:
            if event["type"] != "request_sent":
                continue
            prefix = directory / ("prefix-before-%d.log" % event["seq"])
            prefix.write_text("\n".join(json.dumps(r, separators=(",", ":")) for r in rows if "seq" not in r or r["seq"] < event["seq"]) + "\n")
            body = run_cli(["render", str(prefix)], launch["vendor"], launch["selected_model"], prefix)
            (directory / ("request-before-%d.json" % event["seq"])).write_bytes(body)
            ephemera.append({"request_seq": event["seq"], "consumed_sequences": event["request"]["ephemera"], "marker_in_reconstructed_request": b"LILAC-614" in body})
        assert [x["marker_in_reconstructed_request"] for x in ephemera] == [False, True, False]
        assert [x["consumed_sequences"] for x in ephemera] == [[], [4], []]
        row["ephemera_offline_reconstruction"] = ephemera
        dump = run_cli(["dump"], launch["vendor"], launch["selected_model"], directory / "session.log")
        assert [json.loads(x) for x in dump.splitlines()] == rows
        one = run_cli(["render", str(directory / "session.log")], launch["vendor"], launch["selected_model"], directory / "session.log")
        two = run_cli(["render", str(directory / "session.log")], launch["vendor"], launch["selected_model"], directory / "session.log")
        assert one == two
        (directory / "final-render.json").write_bytes(one)
        row["offline_dump_and_pure_render"] = True
    elif directory.name.endswith("-eof"):
        assert launch["exit_code"] == 0 and not events
        assert "Final usage: input=0, cache write=0, cache read=0, output=0" in transcript
        row["kind"] = "actual terminal clean EOF; no paid request"
    else:
        assert launch["exit_code"] == 1 and not events
        assert "Final usage" not in transcript
        row["kind"] = "local invalid redaction target control; no paid request"
    summary.append(row)

settings = json.loads((Path.home() / ".cr/settings.json").read_text())
keys = [settings[x].encode() for x in ["directClaudeAPIKey", "directOpenAIAPIKey", "directGeminiAPIKey"]]
assert all(keys)
for path in HERE.rglob("*"):
    if path.is_file():
        content = path.read_bytes()
        assert not any(key in content for key in keys), "credential found in evidence"
(HERE / "receipts.json").write_text(json.dumps({"runs": summary, "credential_scan": "all evidence scanned against actual authorized keys; no matches", "ephemera_note": "Requests reconstructed offline from actual durable log prefixes; not intercepted live HTTP bodies."}, indent=2) + "\n")
print(json.dumps(summary, indent=2))
