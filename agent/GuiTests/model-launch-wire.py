"""Public GUI route/policy round trip using an owned provider; no live services.
Usage: python3 model-launch-wire.py ABS_DOTNET ABS_GUI_DLL NEW_DIRECTORY
Buffered replies deliberately do not certify SSE parsing.
"""
import base64
import hashlib
import http.server
import json
import os
from pathlib import Path
import signal
import socket
import struct
import subprocess
import sys
import threading
import time


def require(value, reason):
    if not value:
        raise AssertionError(reason)


class Client:
    def __init__(self, port):
        self.socket = socket.create_connection(("127.0.0.1", port), timeout=5)
        self.socket.settimeout(5)
        self.stream = self.socket.makefile("rb")
        try:
            key = base64.b64encode(os.urandom(16)).decode()
            self.socket.sendall((f"GET /ws HTTP/1.1\r\nHost: 127.0.0.1:{port}\r\n"
                f"Origin: http://127.0.0.1:{port}\r\nUpgrade: websocket\r\nConnection: Upgrade\r\n"
                f"Sec-WebSocket-Key: {key}\r\nSec-WebSocket-Version: 13\r\n\r\n").encode())
            require(b" 101 " in self.stream.readline(4096), "websocket upgrade")
            headers = {}
            for _ in range(64):
                line = self.stream.readline(4096)
                if line == b"\r\n":
                    break
                name, value = line.decode().split(":", 1)
                headers[name.lower()] = value.strip()
            expected = base64.b64encode(hashlib.sha1((key +
                "258EAFA5-E914-47DA-95CA-C5AB0DC85B11").encode()).digest()).decode()
            require(headers.get("sec-websocket-accept") == expected, "websocket acceptance")
        except BaseException:
            self.close()
            raise

    def exact(self, count):
        result = bytearray()
        while len(result) < count:
            chunk = self.stream.read(count - len(result))
            if not chunk:
                raise EOFError("websocket closed")
            result.extend(chunk)
        return bytes(result)

    def send(self, value, opcode=1):
        payload = json.dumps(value).encode() if opcode == 1 else value
        mask = os.urandom(4)
        size = len(payload)
        header = bytes([128 | opcode, 128 | size]) if size < 126 else bytes([128 | opcode, 254]) + struct.pack("!H", size)
        self.socket.sendall(header + mask + bytes(v ^ mask[i % 4] for i, v in enumerate(payload)))

    def until(self, predicate):
        deadline = time.monotonic() + 20
        message = bytearray()
        while time.monotonic() < deadline:
            first, second = self.exact(2)
            require(not second & 128, "unmasked server frame")
            length = second & 127
            if length == 126:
                length = struct.unpack("!H", self.exact(2))[0]
            elif length == 127:
                length = struct.unpack("!Q", self.exact(8))[0]
            require(length <= 1024 * 1024, "bounded frame")
            payload = self.exact(length)
            opcode = first & 15
            if opcode == 9:
                self.send(payload, 10)
                continue
            require(opcode in (0, 1), "text frame")
            message.extend(payload)
            require(len(message) <= 1024 * 1024, "bounded message")
            if first & 128:
                frame = json.loads(message)
                message.clear()
                require(frame.get("type") != "error", "GUI refusal: " + str(frame))
                if predicate(frame):
                    return frame
        raise TimeoutError("GUI acknowledgment deadline")

    def close(self):
        self.stream.close()
        self.socket.close()


def run(dotnet, gui, directory):
    directory.mkdir()
    for part in ("home", "tmp", "cli", "skills/primary"):
        (directory / part).mkdir(parents=True)
    (directory / "skills/primary/SKILL.md").write_text(
        "---\nname: primary\ndescription: Wire test\ntype: primary\ntools:\n---\nAnswer briefly.\n")
    requests = []
    errors = []
    key = "synthetic-launch-wire-key"

    class Provider(http.server.BaseHTTPRequestHandler):
        def log_message(self, *args):
            pass

        def do_POST(self):
            try:
                length = int(self.headers.get("Content-Length", "0"))
                require(0 < length <= 1024 * 1024, "bounded request")
                body = json.loads(self.rfile.read(length))
                requests.append({"path": self.path, "authorization": self.headers.get("Authorization"), "body": body})
                number = len(requests)
                require(number <= 3, "no extra provider requests")
                answer = f"wire-answer-{number}"
                if self.path.endswith("/responses"):
                    reply = {"status": "completed", "model": body["model"], "output": [{
                        "type": "message", "role": "assistant", "status": "completed", "content": [
                            {"type": "output_text", "text": answer}]}],
                        "usage": {"input_tokens": 30, "output_tokens": 5}}
                else:
                    reply = {"model": body["model"], "choices": [{"index": 0, "finish_reason": "stop",
                        "message": {"role": "assistant", "content": answer}}],
                        "usage": {"prompt_tokens": 30, "completion_tokens": 5}}
                encoded = json.dumps(reply).encode()
                self.send_response(200)
                self.send_header("Content-Type", "application/json")
                self.send_header("Content-Length", str(len(encoded)))
                self.end_headers()
                self.wfile.write(encoded)
            except Exception as error:
                errors.append(str(error))
                self.send_error(500)

    class BoundedServer(http.server.HTTPServer):
        def get_request(self):
            connection, address = super().get_request()
            connection.settimeout(5)
            return connection, address

    server = BoundedServer(("127.0.0.1", 0), Provider)
    thread = threading.Thread(target=server.serve_forever)
    process = None
    client = None
    forced = False
    failure = None
    checks = []
    thread.start()
    try:
        base = f"http://127.0.0.1:{server.server_port}"
        offering = {"vendor": "local", "endpoint": "test", "model": "small"}
        launch = {"vendor": "host-launch", "endpoint": "launch", "model": "launch"}
        document = {"version": 1, "vendors": [{"id": "local", "label": "Local", "endpoints": [{
            "id": "test", "address": base + "/local", "surface": "chat_completions", "authentication": "none",
            "models": [{"id": "small", "name": "local-wire", "label": "Small", "features": {"stream": True},
                        "serving_context_tokens": 4096}]}]}]}
        (directory / "models.json").write_text(json.dumps(document))
        with socket.socket() as reservation:
            reservation.bind(("127.0.0.1", 0))
            port = reservation.getsockname()[1]
        env = {"PATH": "/usr/bin:/bin", "HOME": str(directory / "home"), "TMPDIR": str(directory / "tmp"),
            "DOTNET_CLI_HOME": str(directory / "cli"), "DOTNET_SKIP_FIRST_TIME_EXPERIENCE": "1",
            "DOTNET_CLI_TELEMETRY_OPTOUT": "1", "DOTNET_GENERATE_ASPNET_CERTIFICATE": "false", "LLM_API_KEY": key}
        command = [dotnet, gui, "--workspace", str(directory), "--data-dir", str(directory / "data"),
            "--agent-name", "wire", "--port", str(port), "--vendor", "openai", "--model", "gpt-6-astra",
            "--base-url", base + "/launch", "--surface", "responses", "--billing", "api", "--no-stream",
            "--reasoning-effort", "high", "--reasoning-summary", "off", "--prompt-cache-key", "launch-cache",
            "--no-save", "--model-config", str(directory / "models.json"), "--skills-dir", str(directory / "skills"),
            "--primary-skill", "primary"]
        with (directory / "stdout.log").open("wb") as stdout, (directory / "stderr.log").open("wb") as stderr:
            process = subprocess.Popen(command, cwd=directory, env=env, stdout=stdout, stderr=stderr, start_new_session=True)
            deadline = time.monotonic() + 15
            while True:
                require(process.poll() is None, "host startup")
                try:
                    with socket.create_connection(("127.0.0.1", port), timeout=.2):
                        break
                except OSError:
                    require(time.monotonic() < deadline, "startup deadline")
                    time.sleep(.025)
            client = Client(port)
            initial = client.until(lambda f: f.get("type") == "usage")
            require(initial.get("selected_offering") is None, "initial legacy selection unchanged")
            checks.append("initial-legacy-selection")
            for number, target in enumerate((None, offering, launch), 1):
                if target is not None:
                    client.send({"type": "set_model", "offering": target})
                    client.until(lambda f: f.get("type") == "usage" and f.get("selected_offering") == target)
                client.send({"type": "prompt", "text": f"wire-prompt-{number}"})
                completed = client.until(lambda f: f.get("type") == "turn_ended")
                require(completed.get("reason") == "completed" and completed.get("text") == f"wire-answer-{number}",
                        "completed correlated turn " + str(number))
                checks.append("completed-turn-" + str(number))
        require(len(requests) == 3 and not errors, "exact three successful provider exchanges")
        first, local, returned = requests
        for request in (first, returned):
            body = request["body"]
            require(request["path"] == "/launch/v1/responses", "launch endpoint restored")
            require(request["authorization"] == "Bearer " + key, "launch authentication restored")
            require(body["model"] == "gpt-6-astra" and not body.get("stream", False), "launch model and buffered policy")
            require(body.get("reasoning") == {"effort": "high"}, "explicit high reasoning and summary off")
            require(body.get("prompt_cache_key") == "launch-cache", "launch cache key restored")
        checks.append("launch-route-auth-and-policy-restored")
        require(local["path"] == "/local/v1/chat/completions" and local["authorization"] is None, "local unauthenticated route")
        require(local["body"]["model"] == "local-wire" and local["body"].get("stream") is True, "local streaming policy")
        require("reasoning" not in local["body"] and "prompt_cache_key" not in local["body"], "old policy absent locally")
        checks.append("local-route-without-old-policy-or-credentials")
        for index, request in enumerate(requests):
            wire = json.dumps(request["body"])
            for earlier in range(1, index + 1):
                require(f"wire-prompt-{earlier}" in wire and f"wire-answer-{earlier}" in wire, "conversation retained")
        checks.append("conversation-retained-across-routes")
    except Exception as error:
        failure = str(error)
    finally:
        if client is not None:
            client.close()
        if process is not None and process.poll() is None:
            os.killpg(process.pid, signal.SIGTERM)
            try:
                process.wait(timeout=10)
            except subprocess.TimeoutExpired:
                forced = True
                os.killpg(process.pid, signal.SIGKILL)
                process.wait()
        server.shutdown()
        server.server_close()
        thread.join()
    if forced or process is None or process.returncode != 0:
        failure = failure or "host did not exit cleanly"
    result = {"passed": failure is None, "checks": checks, "failure": failure,
              "requests": requests, "provider_errors": errors, "host_exit": process.returncode if process else None,
              "forced": forced, "server_joined": not thread.is_alive()}
    (directory / "result.json").write_text(json.dumps(result, indent=2))
    print(json.dumps({k: v for k, v in result.items() if k != "requests"}))
    return 0 if failure is None else 1


if __name__ == "__main__":
    sys.exit(run(sys.argv[1], sys.argv[2], Path(sys.argv[3]).resolve()))
