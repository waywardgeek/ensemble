"""Actual GUI host startup regression; existing SDK/runtime only, no inference.
Usage: python3 model-host.py ABS_DOTNET ABS_GUI_DLL NEW_EVIDENCE_DIRECTORY
Uses fresh homes, loopback ports, synthetic credentials and joined child processes.
"""
import base64
import hashlib
import json
import os
from pathlib import Path
import signal
import socket
import struct
import subprocess
import sys
import time


def require(condition, message):
    if not condition:
        raise AssertionError(message)



def require_no_inference(log):
    records = [json.loads(line) for line in log.read_text().splitlines() if line.strip()]
    if records and "log_version" in records[0]:
        require(records[0] == {"log_version": 1}, "unsupported event log header")
        records = records[1:]
    require(bool(records), "event log has no events")
    require(all(isinstance(event.get("type"), str) for event in records), "event type missing")
    require(all(event["type"] != "request_sent" for event in records), "startup initiated inference")


def receive_exact(stream, count):
    data = bytearray()
    while len(data) < count:
        chunk = stream.read(count - len(data))
        if not chunk:
            raise EOFError("GUI websocket closed")
        data.extend(chunk)
    return bytes(data)


def usage_frame(port):
    # Public read-only websocket consumer, not application-frame injection.
    with socket.create_connection(("127.0.0.1", port), timeout=5) as connection:
        connection.settimeout(5)
        stream = connection.makefile("rb")
        try:
            key = base64.b64encode(os.urandom(16)).decode()
            origin = f"http://127.0.0.1:{port}"
            connection.sendall((f"GET /ws HTTP/1.1\r\nHost: 127.0.0.1:{port}\r\n"
                f"Origin: {origin}\r\nUpgrade: websocket\r\nConnection: Upgrade\r\n"
                f"Sec-WebSocket-Key: {key}\r\nSec-WebSocket-Version: 13\r\n\r\n").encode())
            require(b" 101 " in stream.readline(4096), "websocket upgrade")
            headers = {}
            for _ in range(64):
                line = stream.readline(4096)
                if line == b"\r\n":
                    break
                name, value = line.decode().split(":", 1)
                headers[name.lower()] = value.strip()
            expected = base64.b64encode(hashlib.sha1((key +
                "258EAFA5-E914-47DA-95CA-C5AB0DC85B11").encode()).digest()).decode()
            require(headers.get("sec-websocket-accept") == expected, "websocket acceptance")
            message = bytearray()
            for _ in range(100):
                first, second = receive_exact(stream, 2)
                require(not second & 128, "server frame must not be masked")
                length = second & 127
                if length == 126:
                    length = struct.unpack("!H", receive_exact(stream, 2))[0]
                elif length == 127:
                    length = struct.unpack("!Q", receive_exact(stream, 8))[0]
                require(length <= 1024 * 1024, "bounded server frame")
                payload = receive_exact(stream, length)
                opcode = first & 15
                require(opcode in (0, 1), "expected text websocket message")
                message.extend(payload)
                require(len(message) <= 1024 * 1024, "bounded server message")
                if first & 128:
                    frame = json.loads(message)
                    message.clear()
                    if frame.get("type") == "usage":
                        return frame
            raise AssertionError("missing usage frame")
        finally:
            stream.close()


def document(binding="none"):
    return {"version": 1, "vendors": [{"id": "local", "label": "Local", "endpoints": [{
        "id": "test", "address": "http://127.0.0.1:1", "surface": "chat_completions",
        "authentication": binding, "models": [{"id": "small", "name": "same-wire",
        "label": "Small", "features": {"stream": True, "context_window": 262144},
        "serving_context_tokens": 4096}]}]}]}


def run_case(root, dotnet, gui, name, content, binding="none", billing="api", key="synthetic-host-key",
             expected=True, via_environment=False, override=False, omitted=False,
             global_content=None, global_override=False):
    directory = root / name
    directory.mkdir()
    for part in ("home", "tmp", "cli", "skills/primary"):
        (directory / part).mkdir(parents=True)
    (directory / "skills/primary/SKILL.md").write_text(
        "---\nname: primary\ndescription: Startup test\ntype: primary\ntools:\n---\nNo inference.\n")
    path = directory / "models.json"
    if content is not None:
        path.write_bytes(content)
    global_path = directory / "home/.en/settings.json"
    if global_override:
        global_path = directory / "explicit-global.json"
    if global_content is not None:
        global_path.parent.mkdir(parents=True, exist_ok=True)
        global_path.write_bytes(global_content)
    with socket.socket() as reservation:
        reservation.bind(("127.0.0.1", 0))
        port = reservation.getsockname()[1]
    env = {"PATH": "/usr/bin:/bin", "HOME": str(directory / "home"),
        "TMPDIR": str(directory / "tmp"), "DOTNET_CLI_HOME": str(directory / "cli"),
        "DOTNET_SKIP_FIRST_TIME_EXPERIENCE": "1", "DOTNET_CLI_TELEMETRY_OPTOUT": "1",
        "DOTNET_GENERATE_ASPNET_CERTIFICATE": "false", "LLM_API_KEY": key}
    command = [dotnet, gui, "--workspace", str(directory), "--data-dir", str(directory / "data"),
        "--agent-name", "startup", "--port", str(port), "--vendor", "openai", "--model", "launch-wire",
        "--base-url", "http://127.0.0.1:1", "--billing", billing, "--no-save",
        "--skills-dir", str(directory / "skills"), "--primary-skill", "primary"]
    if via_environment or override:
        env["EN_MODEL_CONFIG"] = str(directory / "missing.json") if override else str(path)
    if not via_environment and not omitted:
        command += ["--model-config", str(path)]
    if global_override:
        env["EN_GLOBAL_SETTINGS"] = str(directory / "invalid-global.json")
        (directory / "invalid-global.json").write_text("null")
        command += ["--global-settings", str(global_path)]
    process = None
    forced = False
    frame = None
    with (directory / "stdout.log").open("wb") as stdout, (directory / "stderr.log").open("wb") as stderr:
        try:
            process = subprocess.Popen(command, cwd=directory, env=env, stdout=stdout, stderr=stderr,
                                       start_new_session=True)
            if expected:
                deadline = time.monotonic() + 15
                while True:
                    require(process.poll() is None, "host exited before serving")
                    try:
                        with socket.create_connection(("127.0.0.1", port), timeout=.2):
                            break
                    except OSError:
                        require(time.monotonic() < deadline, "host startup deadline")
                        time.sleep(.025)
                frame = usage_frame(port)
                (directory / "usage.json").write_text(json.dumps(frame, indent=2))
                require(frame["selected_model"] == "launch-wire" and frame["selected_offering"] is None,
                        "loading must preserve legacy launch selection")
                choices = frame["configured_models"]
                launch_funding = ("Unauthenticated endpoint; no credential or cloud fallback."
                                  if billing == "none" else "API key (metered usage).")
                require(frame["funding"] == "Model: launch-wire. Funding: " + launch_funding,
                        "initial launch funding preserved")
                if global_content is not None:
                    api_choices = [item for item in choices
                                   if item["offering"]["vendor"] in ("openai", "anthropic", "gemini")]
                    require(len(api_choices) == 7 and all(item["enabled"] for item in api_choices),
                            "all curated global API offerings enabled")
                    require({item["offering"]["vendor"] for item in api_choices} ==
                            {"openai", "anthropic", "gemini"}, "separate configured vendors")
                    require(all(item["offering"]["endpoint"] == "api" for item in api_choices),
                            "stable API endpoint identity")
                    for secret in json.loads(global_content).values():
                        if isinstance(secret, str):
                            require(secret not in json.dumps(frame), "global secret absent from status")
                    choices = [item for item in choices if item not in api_choices]
                if omitted:
                    require(choices == [], "no implicit configured graph")
                else:
                    has_launch = billing == "none" or bool(key)
                    require(len(choices) == (2 if has_launch else 1), "document and available launch offerings")
                    choice = next(item for item in choices if item["offering"]["vendor"] == "local")
                    if has_launch:
                        launch = next(item for item in choices if item["offering"]["vendor"] == "host-launch")
                        require(launch["offering"] == {"vendor": "host-launch", "endpoint": "launch", "model": "launch"}
                                and launch["enabled"] and launch["funding"] == frame["funding"],
                                "explicit launch return identity and funding")
                    require(choice["offering"] == {"vendor": "local", "endpoint": "test", "model": "small"},
                            "qualified configured identity")
                    require(choice["enabled"] and choice["serving_context_tokens"] == 4096
                            and choice["advertised_context_tokens"] == 262144, "loaded serving/capacity")
                    funding = ("API key (metered usage)." if binding == "launch" and billing == "api"
                               else "Unauthenticated endpoint; no credential or cloud fallback.")
                    require(choice["funding"] == "Model: same-wire. Funding: " + funding,
                            "explicit host binding funding")
                require(key not in json.dumps(frame) if key else True, "no secret in status")
            else:
                require(process.wait(timeout=15) == 1, "invalid startup must refuse")
                require(not (directory / "data/agents/startup").exists(), "refusal before Agent creation")
        finally:
            if process is not None and process.poll() is None:
                os.killpg(process.pid, signal.SIGTERM)
                try:
                    process.wait(timeout=10)
                except subprocess.TimeoutExpired:
                    forced = True
                    os.killpg(process.pid, signal.SIGKILL)
                    process.wait()
    require(not forced, "host must join without forced kill")
    require(process.returncode == (0 if expected else 1), "joined host exit")
    logs = list((directory / "data").rglob("events.jsonl"))
    require(bool(logs) == expected, "event log exists only after valid admission")
    for log in logs:
        require_no_inference(log)
    if global_content is not None:
        require(global_path.read_bytes() == global_content, "startup never rewrites global settings")
    return {"case": name, "passed": True, "exit": process.returncode, "joined": True, "forced": forced}


def main():
    dotnet, gui, output = sys.argv[1:]
    require(Path(dotnet).is_absolute() and Path(gui).is_absolute(), "absolute existing runtime paths")
    root = Path(output).resolve()
    root.mkdir()
    global_keys = json.dumps({"openai_api_key": "fixture-openai-global",
                              "anthropic_api_key": "fixture-anthropic-global",
                              "gemini_api_key": "fixture-gemini-global"}).encode()
    cases = [
        ("global-default-home", None, {"omitted": True, "global_content": global_keys}),
        ("global-explicit-precedence", None,
            {"omitted": True, "global_content": global_keys, "global_override": True}),
        ("global-malformed", None,
            {"omitted": True, "global_content": b"null", "expected": False}),
        ("global-reserved-vendor", json.dumps(document()).replace('"id": "local"', '"id": "openai"').encode(),
            {"global_content": global_keys, "expected": False}),
        ("legacy-no-option", None, {"omitted": True}),
        ("none-binding", json.dumps(document()).encode(), {}),
        ("launch-binding", json.dumps(document("launch")).encode(), {"binding": "launch"}),
        ("none-launch-env", json.dumps(document("launch")).encode(),
            {"binding": "launch", "billing": "none", "via_environment": True}),
        ("flag-precedence", json.dumps(document()).encode(), {"override": True}),
        ("unknown-binding", json.dumps(document("unknown")).encode(), {"expected": False}),
        ("keyless-local-only", json.dumps(document()).encode(), {"key": ""}),
        ("reserved-launch-vendor", json.dumps(document()).replace('"id": "local"', '"id": "host-launch"').encode(),
            {"expected": False}),
        ("missing-launch-key", json.dumps(document("launch")).encode(), {"key": "", "expected": False}),
        ("invalid-utf8", b"\xff", {"expected": False}),
        ("oversize", b" " * (1024 * 1024 + 1), {"expected": False}),
        ("missing-file", None, {"expected": False}),
    ]
    results = []
    try:
        for name, content, options in cases:
            results.append(run_case(root, dotnet, gui, name, content, **options))
        report = {"passed": True, "cases": results}
    except Exception as error:
        report = {"passed": False, "cases": results, "failure": str(error)}
    (root / "result.json").write_text(json.dumps(report, indent=2))
    print(json.dumps(report, indent=2))
    return 0 if report["passed"] else 1


if __name__ == "__main__":
    sys.exit(main())
