#!/usr/bin/env python3
"""POSIX, offline-dependency runner for the Chapter7 copy components (buffered by default)."""
import argparse
import contextlib
import http.server
import json
import os
from pathlib import Path
import signal
import socket
import subprocess
import threading
import time
import urllib.error
import urllib.request

from fixtures import buffered_fixture, streaming_fixture

HERE = Path(__file__).resolve().parent
LIMIT = 1024 * 1024
BROWSER_CHECKS = [
    'actual-provider-markdown-visible',
    'only-closed-assistant-fences-copyable',
    'real-human-fence-has-no-copy-control',
    'clipboard-write-requires-click',
    'long-copy-starts-collapsed',
    'exact-logical-payloads-including-empty-and-full-long-code',
    'table-and-literal-safe-rendering-without-image-fetch',
    'denied-write-shows-exact-readonly-selected-fallback',
    'missing-api-shows-exact-fallback',
    'copy-preserves-draft-and-does-not-read-or-submit',
    'no-page-errors-or-off-origin-fetches',
]

STREAM_CHECKS = [
    'unfinished-fence-visible-without-copy-before-release',
    'closed-stream-fence-copies-exact-payload',
    'stream-update-retains-original-focused-copy-control',
    'stream-update-retains-expansion-and-feedback',
    'manual-fallback-established-before-final-delta',
    'canonical-final-and-completed-turn-applied-before-selection-check',
    'stream-update-retains-original-manual-selection',
    'stream-copy-does-not-read-or-resubmit',
    'stream-page-has-no-errors-or-off-origin-fetches',
]


def valid_browser(report, case='buffered'):
    checks = BROWSER_CHECKS if case == 'buffered' else STREAM_CHECKS
    contract = 'chapter7-copy-browser-component-v1' if case == 'buffered' else 'chapter7-stream-browser-component-v1'
    return (report.get('contract') == contract
            and report.get('passed') is True
            and 'failure' in report and report['failure'] is None
            and report.get('checks') == checks
            and report.get('promptFrames') == 1
            and report.get('errors') == [] and report.get('blocked') == [])


def existing(value, directory=False, executable=False):
    path = Path(value)
    if not path.is_absolute() or not (path.is_dir() if directory else path.is_file()):
        raise argparse.ArgumentTypeError('Expected an absolute existing ' + ('directory' if directory else 'file'))
    if executable and not os.access(path, os.X_OK):
        raise argparse.ArgumentTypeError('Expected an executable file')
    return path.resolve()


def write_json(path, value):
    path.write_text(json.dumps(value, ensure_ascii=False, indent=2) + '\n')


def read_json(path):
    with path.open('rb') as stream:
        data = stream.read(LIMIT + 1)
    if len(data) > LIMIT:
        raise ValueError('receipt-exceeds-limit')
    return json.loads(data)


def stop_process(process, seconds):
    if process is None:
        return dict(started=False, joined=True, forced=False, exit=None)
    forced = False
    if process.poll() is None:
        with contextlib.suppress(ProcessLookupError):
            os.killpg(process.pid, signal.SIGTERM)
        try:
            process.wait(timeout=seconds)
        except subprocess.TimeoutExpired:
            forced = True
            with contextlib.suppress(ProcessLookupError):
                os.killpg(process.pid, signal.SIGKILL)
    process.wait()
    # A launcher may have exited while leaving descendants. Signal its owned
    # process group too; never target a listener or a process found by name.
    with contextlib.suppress(ProcessLookupError):
        os.killpg(process.pid, signal.SIGKILL)
    return dict(started=True, joined=True, forced=forced, exit=process.returncode)


def run(args):
    work = args.work
    if not work.is_absolute():
        raise ValueError('Work directory must be absolute')
    work.mkdir()  # Refuse existing evidence rather than overwrite it.
    for name in ('home', 'tmp', 'cli'):
        (work / name).mkdir()
    fixture = buffered_fixture() if args.case == 'buffered' else streaming_fixture()
    write_json(work / 'fixture.json', fixture)
    env = dict(PATH='/usr/bin:/bin', HOME=str(work / 'home'), TMPDIR=str(work / 'tmp'),
               DOTNET_CLI_HOME=str(work / 'cli'), DOTNET_SKIP_FIRST_TIME_EXPERIENCE='1',
               DOTNET_CLI_TELEMETRY_OPTOUT='1', DOTNET_GENERATE_ASPNET_CERTIFICATE='false',
               LLM_API_KEY='synthetic-unused', CH06_API_LOG=str(work / 'api.log'),
               CH07_GUI_LOG=str(work / 'gui.log'))
    requests, provider_errors = [], []
    gates = [threading.Event() for _ in range(3)]
    releases = []
    provider_lock = threading.Lock()
    host = browser = server = thread = None
    browser_report = None
    failure = None
    cleanup_errors = []
    interrupted = False

    class Handler(http.server.BaseHTTPRequestHandler):
        def log_message(self, *_):
            pass

        def send_event(self, name, value):
            self.wfile.write(('event: ' + name + '\ndata: ' + json.dumps(value) + '\n\n').encode())
            self.wfile.flush()

        def stream_response(self, request):
            if request.get('stream') is not True:
                raise AssertionError('actual-stream-request-required')
            self.send_response(200)
            self.send_header('Content-Type', 'text/event-stream')
            self.send_header('Connection', 'close')
            self.end_headers()
            self.send_event('message_start', dict(type='message_start', message=dict(
                id='stream-fixture', type='message', role='assistant',
                model='claude-3-haiku-20240307', content=[], usage=dict(input_tokens=10,
                output_tokens=0, cache_creation_input_tokens=0, cache_read_input_tokens=0))))
            self.send_event('content_block_start', dict(type='content_block_start', index=7,
                            content_block=dict(type='text', text='')))
            for index, text in enumerate(fixture['phases']):
                if index and not gates[index - 1].wait(20):
                    raise AssertionError('stream-gate-timeout-' + str(index))
                self.send_event('content_block_delta', dict(type='content_block_delta', index=7,
                                delta=dict(type='text_delta', text=text)))
            self.send_event('content_block_stop', dict(type='content_block_stop', index=7))
            self.send_event('message_delta', dict(type='message_delta',
                            delta=dict(stop_reason='end_turn'), usage=dict(output_tokens=7)))
            self.send_event('message_stop', dict(type='message_stop'))

        def do_POST(self):
            self.close_connection = True
            try:
                self.connection.settimeout(5)
                if args.case == 'streaming' and self.path.startswith('/fixture/step/'):
                    step = int(self.path.rsplit('/', 1)[1])
                    with provider_lock:
                        if step != len(releases) + 1 or step not in (1, 2, 3):
                            raise AssertionError('out-of-order-fixture-step')
                        releases.append(step)
                        gates[step - 1].set()
                    self.send_response(200)
                    self.send_header('Content-Length', '0')
                    self.end_headers()
                    return
                count = int(self.headers.get('Content-Length', '0'))
                if not 0 < count <= LIMIT:
                    raise ValueError('provider-request-bound')
                data = self.rfile.read(count)
                if len(data) != count:
                    raise ValueError('short-provider-request')
                request = json.loads(data)
                with provider_lock:
                    if len(requests) >= 4:
                        raise ValueError('provider-request-bound')
                    requests.append(request)
                if args.case == 'streaming':
                    self.stream_response(request)
                    return
                body = json.dumps(dict(
                    id='synthetic-message', type='message', role='assistant',
                    model='claude-3-haiku-20240307',
                    content=[dict(type='text', text=fixture['markdown'])],
                    stop_reason='end_turn', usage=dict(input_tokens=10, output_tokens=7,
                    cache_creation_input_tokens=0, cache_read_input_tokens=0))).encode()
                self.send_response(200)
                self.send_header('Content-Type', 'application/json')
                self.send_header('Content-Length', str(len(body)))
                self.end_headers()
                self.wfile.write(body)
            except Exception as error:
                with provider_lock:
                    provider_errors.append(type(error).__name__ + ': ' + str(error))
                self.close_connection = True

    class Server(http.server.ThreadingHTTPServer):
        daemon_threads = False
        block_on_close = True
        def get_request(self):
            connection, address = super().get_request()
            connection.settimeout(5)
            return connection, address

    def interrupt(_number, _frame):
        nonlocal interrupted
        if not interrupted:
            interrupted = True
            raise InterruptedError('runner-interrupted')

    old_handlers = {number: signal.signal(number, interrupt)
                    for number in (signal.SIGTERM, signal.SIGINT)}
    cleanup = {}
    try:
        server = Server(('127.0.0.1', 0), Handler)
        thread = threading.Thread(target=server.serve_forever, kwargs={'poll_interval': .05})
        thread.start()
        with socket.socket() as reservation:
            reservation.bind(('127.0.0.1', 0))
            port = reservation.getsockname()[1]
        if port in (8088, 8089, 8090):
            raise ValueError('protected-port-selected')
        url = f'http://127.0.0.1:{port}/'
        command = [str(args.launcher), '--port', str(port), '--workspace', str(work),
                   '--vendor', 'anthropic', '--model', 'claude-3-haiku-20240307',
                   '--base-url', f'http://127.0.0.1:{server.server_port}',
                   '--event-log', str(work / 'events.jsonl'),
                   '--gui-dir', str(args.assets)]
        if args.case == 'buffered':
            command.append('--no-stream')
        with (work / 'host.stdout').open('wb') as out, (work / 'host.stderr').open('wb') as err:
            host = subprocess.Popen(command, env=env, cwd=work, stdout=out, stderr=err,
                                    stdin=subprocess.DEVNULL, start_new_session=True)
        expected = (args.assets / 'index.html').read_bytes()
        opener = urllib.request.build_opener(urllib.request.ProxyHandler({}))
        deadline = time.monotonic() + 15
        while True:
            if host.poll() is not None or time.monotonic() >= deadline:
                raise RuntimeError('gui-not-ready')
            try:
                with opener.open(url, timeout=.5) as response:
                    page = response.read(LIMIT + 1)
                if page != expected:
                    raise ValueError('unexpected-page-at-owned-port')
                break
            except (urllib.error.URLError, TimeoutError, ConnectionError):
                time.sleep(.05)
        component = 'buffered.cjs' if args.case == 'buffered' else 'streaming.cjs'
        command = [str(args.node), str(HERE / component), url, str(work / 'fixture.json')]
        if args.case == 'streaming':
            command.append(f'http://127.0.0.1:{server.server_port}')
        command.extend([str(args.playwright), str(args.bindings)])
        with (work / 'browser.stdout').open('wb') as out, (work / 'browser.stderr').open('wb') as err:
            browser = subprocess.Popen(command, env=env, cwd=work, stdout=out, stderr=err,
                                       stdin=subprocess.DEVNULL, start_new_session=True)
        browser.wait(timeout=45)
        browser_report = read_json(work / 'browser.stdout')
        if browser.returncode != 0 or not valid_browser(browser_report, args.case):
            raise AssertionError(args.case + '-browser-check-failed')
    except Exception as error:
        failure = type(error).__name__ + ': ' + str(error)
    finally:
        # Ignore repeated signals while draining our already-owned children.
        for number in old_handlers:
            signal.signal(number, signal.SIG_IGN)
        for gate in gates:
            gate.set()  # Release owned provider work even when a browser assertion fails.
        for name, child, grace in [('browser', browser, 5), ('host', host, 10)]:
            try:
                cleanup[name] = stop_process(child, grace)
            except Exception as error:
                cleanup_errors.append(name + ': ' + str(error))
        if server is not None:
            if thread is not None and thread.ident is not None:
                server.shutdown()
            server.server_close()
        if thread is not None and thread.ident is not None:
            thread.join()
        cleanup['server_joined'] = thread is None or not thread.is_alive()
        for number, handler in old_handlers.items():
            signal.signal(number, handler)

    try:
        if cleanup_errors or not cleanup['server_joined']:
            raise AssertionError('cleanup-failed')
        if cleanup.get('host', {}).get('exit') != 0 or any(
                cleanup.get(name, {}).get('forced', False) for name in ('browser', 'host')):
            raise AssertionError('host-or-browser-did-not-stop-cleanly')
        if args.case == 'streaming' and releases != [1, 2, 3]:
            raise AssertionError('stream-gates-not-released-in-order')
        if provider_errors:
            raise AssertionError('provider-fixture-error')
        if len(requests) != 1:
            raise AssertionError('exactly-one-provider-request')
        def text(content):
            if isinstance(content, str):
                return content
            return ''.join(p.get('text', '') for p in content if p.get('type') == 'text')
        inputs = [text(m.get('content', [])) for m in requests[0].get('messages', [])
                  if m.get('role') == 'user']
        if inputs != [fixture['prompt']]:
            raise AssertionError('provider-human-payload-changed')
        with (work / 'events.jsonl').open('rb') as log:
            data = log.read(LIMIT + 1)
        if len(data) > LIMIT:
            raise AssertionError('canonical-log-exceeds-limit')
        facts = [json.loads(line) for line in data.splitlines() if line.strip()]
        human = [p.get('text') for e in facts
                 if e.get('type') in ('message_received', 'hint_received')
                 and e.get('message', {}).get('actor') == 'human'
                 for p in e['message'].get('parts', []) if p.get('type') == 'text']
        answers = [p.get('text') for e in facts if e.get('type') == 'response_ended'
                   for p in e['response'].get('parts', []) if p.get('type') == 'text']
        if human != [fixture['prompt']] or answers != [fixture['markdown']]:
            raise AssertionError('canonical-original-payload-changed')
        if any(sum(e.get('type') == kind for e in facts) != 1
               for kind in ('request_sent', 'response_ended')):
            raise AssertionError('canonical-request-response-counts')
        if any(e.get('type') in ('tool_called', 'error_occurred') for e in facts):
            raise AssertionError('unexpected-tool-or-error-fact')
    except Exception as error:
        failure = failure or type(error).__name__ + ': ' + str(error)
    result = dict(contract='chapter7-' + args.case + '-runner-v1', passed=failure is None,
                  case=args.case, releases=releases,
                  failure=failure, interrupted=interrupted, browser=browser_report,
                  requests=len(requests), provider_errors=provider_errors,
                  cleanup=cleanup, cleanup_errors=cleanup_errors)
    write_json(work / 'requests.json', requests)
    write_json(work / 'result.json', result)
    print(json.dumps(result))
    return 2 if interrupted or cleanup_errors else 0 if result['passed'] else 1


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--case', choices=['buffered', 'streaming'], default='buffered')
    parser.add_argument('--launcher', required=True, type=lambda v: existing(v, executable=True))
    parser.add_argument('--node', required=True, type=lambda v: existing(v, executable=True))
    parser.add_argument('--playwright', required=True, type=lambda v: existing(v, directory=True))
    parser.add_argument('--assets', required=True, type=lambda v: existing(v, directory=True))
    parser.add_argument('--bindings', type=existing, default=HERE / 'bindings.json')
    parser.add_argument('--work', required=True, type=Path)
    args = parser.parse_args()
    if os.name != 'posix':
        parser.error('POSIX process groups are required')
    try:
        return run(args)
    except Exception as error:
        print('Runner setup failed: ' + str(error), file=__import__('sys').stderr)
        return 2


if __name__ == '__main__':
    raise SystemExit(main())
