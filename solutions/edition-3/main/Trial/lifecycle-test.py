#!/usr/bin/env python3
"""Real disposable listeners prove controller lifecycle, not live GUI recovery."""
import http.server
import json
import os
from pathlib import Path
import signal
import subprocess
import sys
import threading
import time
import urllib.request


def server():
    port, label, mode = int(sys.argv[2]), sys.argv[3], sys.argv[4]
    class Handler(http.server.BaseHTTPRequestHandler):
        def do_GET(self):
            self.send_response(200)
            self.end_headers()
            self.wfile.write(label.encode())
        def log_message(self, *_):
            pass
    listener = http.server.ThreadingHTTPServer(('127.0.0.1', port), Handler)
    if mode == 'ignore':
        signal.signal(signal.SIGTERM, signal.SIG_IGN)
    if mode == 'slow':
        def delayed(*_):
            time.sleep(.7)
            os._exit(0)
        signal.signal(signal.SIGTERM, delayed)
    if mode == 'crash':
        threading.Timer(.4, lambda: os._exit(7)).start()
    if mode == 'release':
        def release():
            listener.shutdown()
            listener.server_close()
        threading.Timer(.5, release).start()
    listener.serve_forever(poll_interval=.05)
    while True:
        time.sleep(1)


def main():
    root = Path.cwd()
    evidence = root / '.crossover/evidence/recovery'
    evidence.mkdir(parents=True, exist_ok=True)
    script = str(Path(__file__).resolve())
    cases = ['timer', 'early', 'crash', 'ignore', 'changed-owner', 'cancel', 'slow', 'origin-exit', 'start-failure']
    for i, case in enumerate(cases):
        if len(sys.argv) > 1 and sys.argv[1] != case:
            continue
        port = 18220 + i
        folder = evidence / (case + '-' + str(time.time_ns()))
        folder.mkdir(mode=0o700)
        state = folder / 'state'
        processes = []
        owned = set()
        original_mode = 'slow' if case == 'slow' else 'normal'
        original_argv = [sys.executable, script, '--server', str(port), 'original', original_mode]
        candidate_mode = {'changed-owner': 'release'}.get(case, case if case in ['crash', 'ignore'] else 'normal')
        candidate_argv = [sys.executable, script, '--server', str(port), 'candidate', candidate_mode]
        if case == 'start-failure':
            candidate_argv = [str(folder / 'nonexistent-fixture-candidate')]
        def start(argv):
            output = (folder / ('process-' + str(len(processes)) + '.log')).open('wb')
            proc = subprocess.Popen(argv, stdout=output, stderr=output, start_new_session=True)
            output.close()
            processes.append(proc)
            owned.add(proc.pid)
            return proc
        def body():
            try:
                with urllib.request.urlopen(f'http://127.0.0.1:{port}/', timeout=.3) as response:
                    return response.read().decode()
            except OSError:
                return ''
        def current():
            for proc in processes:
                proc.poll()  # reap original exits while controller verifies identity
            if not (state / 'status.json').exists():
                return {}
            value = json.loads((state / 'status.json').read_text())
            owned.update(value.get('owned', {}).values())
            if value.get('controller_pid'):
                owned.add(value['controller_pid'])
            return value
        def until(predicate, message, timeout=12):
            deadline = time.monotonic() + timeout
            while time.monotonic() < deadline:
                if predicate():
                    return
                time.sleep(.05)
            raise AssertionError(message + ': ' + str(current()))
        try:
            original = start(original_argv)
            until(lambda: body() == 'original', 'original listener')
            plan = {'port': port, 'original_pid': original.pid, 'original_argv': original_argv,
                    'candidate_argv': candidate_argv, 'cwd': str(root), 'state': str(state),
                    'seconds': .4 if case == 'slow' else 2, 'startup_seconds': 2, 'shutdown_seconds': .3}
            plan_file = folder / 'plan.json'
            plan_file.write_text(json.dumps(plan))
            controller_argv = [sys.executable, str(root / 'Trial/recover.py'), str(plan_file)]
            if case == 'origin-exit':
                # The launching process exits before arm/switch. The controller
                # must not depend on its parent, pipes, or original GUI group.
                helper = 'import subprocess,sys; subprocess.Popen(sys.argv[1:], stdin=subprocess.DEVNULL, start_new_session=True)'
                controller = start([sys.executable, '-c', helper] + controller_argv)
                assert controller.wait(timeout=3) == 0
            else:
                controller = start(controller_argv)
            until(lambda: current().get('state') == 'armed', 'observed armed')
            assert original.poll() is None and body() == 'original'
            if case == 'cancel':
                (state / 'cancel').touch()
                until(lambda: current().get('state') == 'cancelled', 'cancel acknowledged')
                assert original.poll() is None and body() == 'original'
            else:
                (state / 'switch').touch()
                if case == 'early':
                    until(lambda: current().get('state') == 'candidate', 'candidate listener')
                    (state / 'restore').touch()
                if case == 'changed-owner':
                    until(lambda: current().get('state') == 'candidate', 'candidate listener')
                    until(lambda: body() == '', 'candidate releases port')
                    rogue = start([sys.executable, script, '--server', str(port), 'unexpected', 'normal'])
                    until(lambda: body() == 'unexpected', 'unexpected owner installed by fixture')
                    until(lambda: current().get('state') == 'refused', 'changed-owner refusal')
                    assert rogue.poll() is None and body() == 'unexpected'
                else:
                    until(lambda: current().get('state') == 'restored', 'observed restoration')
                    controller.wait(timeout=3)
                    time.sleep(.3)
                    assert body() == 'original', 'restored listener survives controller exit'
            controller.wait(timeout=3)
            assert controller.returncode == 0
            print('PASS recovery ' + case, flush=True)
        finally:
            current()
            for pid in owned:
                try:
                    os.killpg(pid, signal.SIGKILL)
                except ProcessLookupError:
                    pass
            for process in processes:
                process.wait(timeout=3)


if __name__ == '__main__':
    if len(sys.argv) > 1 and sys.argv[1] == '--server':
        server()
    else:
        main()
