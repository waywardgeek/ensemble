#!/usr/bin/env python3
"""Exercise the real macOS launcher. Supply absolute dotnet and launcher DLL paths.
No provider credentials or external service is used; all HTTP endpoints are owned
by this test. A real candidate process and its descendant attempt denied effects.
"""
import http.server
import json
import os
import signal
from pathlib import Path
import subprocess
import sys
import tempfile
import threading

DOTNET = Path(sys.argv[1]).resolve(strict=True)
LAUNCHER = Path(sys.argv[2]).resolve(strict=True)

class Handler(http.server.BaseHTTPRequestHandler):
    def do_GET(self):
        self.send_response(200)
        self.end_headers()
        self.wfile.write(b"owned-fake")

    def log_message(self, *args):
        pass

with tempfile.TemporaryDirectory(prefix="confined-grading-") as temporary:
    base = Path(temporary).resolve()
    root = base / "workspace"
    root.mkdir()
    sentinel = base / "outside.txt"
    sentinel.write_text("private-sentinel")
    imported = base / "readonly-import"
    imported.mkdir()
    imported.joinpath("sentinel").write_text("read-only-sentinel")
    root.joinpath("escape").symlink_to(sentinel)
    servers = [http.server.ThreadingHTTPServer(("127.0.0.1", 0), Handler) for _ in range(2)]
    threads = []
    for server in servers:
        thread = threading.Thread(target=server.serve_forever, daemon=True)
        thread.start()
        threads.append(thread)
    try:
        allowed, forbidden = [server.server_port for server in servers]
        candidate = root / "candidate"
        candidate.write_text('''#!/usr/bin/python3
import errno, json, os, socket, subprocess, sys, urllib.request
from pathlib import Path

def denied(action):
    try:
        action()
    except OSError as error:
        return error.errno in (errno.EPERM, errno.EACCES)
    return False

def connect(port, host='127.0.0.1'):
    with socket.socket() as client:
        client.settimeout(1)
        client.connect((host, port))

def listen():
    with socket.socket() as server:
        server.bind(('127.0.0.1', 0))
        server.listen(1)

def udp():
    with socket.socket(socket.AF_INET, socket.SOCK_DGRAM) as client:
        client.sendto(b'probe', ('127.0.0.1', ALLOWED))

checks = {
    'outside_read': denied(lambda: Path(OUTSIDE).read_text()),
    'outside_write': denied(lambda: Path(OUTSIDE).write_text('changed')),
    'symlink': denied(lambda: Path('escape').read_text()),
    'other_port': denied(lambda: connect(FORBIDDEN)),
    'nonlocal': denied(lambda: connect(ALLOWED, '192.0.2.1')),
    'import_read': Path(IMPORTED).read_text() == 'read-only-sentinel',
    'import_write': denied(lambda: Path(IMPORTED).write_text('changed')),
    'listen': denied(listen),
    'udp': denied(udp),
    'secret_absent': 'UNUSUAL_GRADING_SECRET' not in os.environ,
    'home_confined': Path(os.environ['HOME']).resolve() == Path.cwd(),
    'fake': urllib.request.urlopen(os.environ['ANTHROPIC_BASE_URL'], timeout=2).read() == b'owned-fake',
}
if '--descendant' in sys.argv:
    print(json.dumps(checks), flush=True)
    sys.exit(0 if all(checks.values()) else 1)
result = subprocess.run(['/usr/bin/python3', __file__, '--descendant'], capture_output=True, text=True)
checks['descendant'] = result.returncode == 0 and all(json.loads(result.stdout).values())
checks['stdin'] = sys.stdin.readline() == 'stdin-ok\\n'
Path('candidate-output').write_text('workspace-write')
Path('candidate-pid').write_text(str(os.getpid()))
print(json.dumps(checks, sort_keys=True), flush=True)
sys.exit(17)
'''.replace("ALLOWED", str(allowed)).replace("FORBIDDEN", str(forbidden)).replace("OUTSIDE", repr(str(sentinel))).replace("IMPORTED", repr(str(imported / "sentinel"))))
        candidate.chmod(0o700)
        environment = dict(os.environ)
        environment.update(ANTHROPIC_BASE_URL=f"http://127.0.0.1:{allowed}",
                           ANTHROPIC_API_KEY="sk-ant-course-grader-fake",
                           ANTHROPIC_MODEL="claude-fake-course-1",
                           UNUSUAL_GRADING_SECRET="not-for-candidate")
        direct_command = [str(DOTNET), str(LAUNCHER), str(root), str(candidate), str(imported)]
        wrapper = base / "trusted-wrapper"
        helper = Path(__file__).resolve().with_name("prepare-wrapper.py")
        prepare = [sys.executable, str(helper), *direct_command, str(wrapper)]
        prepared = subprocess.run(prepare, capture_output=True, text=True, timeout=15)
        assert prepared.returncode == 0, prepared.stderr
        original_wrapper = wrapper.read_bytes()
        duplicate = subprocess.run(prepare, capture_output=True, text=True, timeout=15)
        assert duplicate.returncode != 0 and wrapper.read_bytes() == original_wrapper
        unsafe = prepare.copy()
        unsafe[-1] = str(root / "unsafe-wrapper")
        refused = subprocess.run(unsafe, capture_output=True, text=True, timeout=15)
        assert refused.returncode != 0 and not (root / "unsafe-wrapper").exists()
        command = [str(wrapper)]
        process = subprocess.Popen(command, env=environment, stdin=subprocess.PIPE,
                                   stdout=subprocess.PIPE, stderr=subprocess.PIPE, text=True,
                                   start_new_session=True)
        try:
            stdout, stderr = process.communicate("stdin-ok\n", timeout=30)
        except subprocess.TimeoutExpired:
            # The candidate owns a descendant; terminate the whole test group,
            # then reap the direct process before disposing its workspace.
            os.killpg(process.pid, signal.SIGKILL)
            process.communicate()
            raise
        assert process.returncode == 17, (process.returncode, stdout, stderr)
        assert int(root.joinpath("candidate-pid").read_text()) == process.pid
        checks = json.loads(stdout)
        assert all(checks.values()), checks
        assert sentinel.read_text() == "private-sentinel"
        assert root.joinpath("candidate-output").read_text() == "workspace-write"
        print("PASS real candidate and descendant confinement, exact stdio and exit17", flush=True)

        # Refuse before executing a candidate, even when all other inputs are valid.
        for address in ["http://localhost:1234", "http://192.0.2.1:1234", "https://127.0.0.1:1234",
                        "http://user@127.0.0.1:1234", "http://127.0.0.1:1234/path",
                        "http://127.0.0.1:1234/?query", "http://127.0.0.1:1234/#fragment"]:
            changed = dict(environment, ANTHROPIC_BASE_URL=address)
            result = subprocess.run(command, env=changed, capture_output=True, text=True, timeout=15)
            assert result.returncode == 126 and not result.stdout, (address, result.returncode, result.stdout)
        changed = dict(environment, ANTHROPIC_API_KEY="synthetic-not-the-grader-key")
        result = subprocess.run(command, env=changed, capture_output=True, text=True, timeout=15)
        assert result.returncode == 126 and not result.stdout
        outside_command = direct_command.copy()
        outside_command[-2] = str(sentinel)
        result = subprocess.run(outside_command, env=environment, capture_output=True, text=True, timeout=15)
        assert result.returncode == 126 and not result.stdout
        print("PASS endpoint, fake-key and candidate-path refusals", flush=True)
    finally:
        for server in servers:
            server.shutdown()
            server.server_close()
        for thread in threads:
            thread.join()
