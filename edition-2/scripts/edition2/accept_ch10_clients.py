#!/usr/bin/env python3
"""Prepared Chapter 10 CLI/WebSocket lifetime checks, using local HTTP only.

Requires coherent CLI/GUI binaries. Public snapshot semantics, browser DOM,
full process-job restoration, storage fault injection and full limits are separate.
Source hashes describe the supplied tree, not proof that binaries were built
from it; immutable gate/build provenance must establish that association.
"""
import argparse
import contextlib
import hashlib
import http.server
import json
import os
from pathlib import Path
import queue
import re
import signal
import subprocess
import tempfile
import threading

from accept_ch07 import Socket
from accept_ch09 import environment, response
from accept_ch10 import CANARY, envelope, inspect_created, uint
from accept_ch10_public import identities

HERE = Path(__file__).resolve().parent


def digest(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()


def source_identities(source):
    values = identities(source)
    for path in source.rglob('*'):
        relative = path.relative_to(source)
        if path.is_file() and 'evidence' not in relative.parts and path.suffix in ('.js', '.mjs', '.html', '.css'):
            values[str(relative)] = digest(path)
    return values


def session(value, resumed=None):
    assert isinstance(value, dict) and set(value) == {'id', 'resumed', 'checkpoint_seq'}, 'unsafe/incomplete session projection'
    assert isinstance(value['id'], str) and re.fullmatch('[0-9a-f]{32}', value['id']), 'invalid SessionID'
    assert type(value['resumed']) is bool, 'resumed is not Boolean'
    if resumed is not None:
        assert value['resumed'] is resumed, 'wrong mount provenance'
    seq = value['checkpoint_seq']
    assert seq is None or type(seq) is int and 0 < seq <= 2**64 - 1, 'checkpoint counter'
    return value


def saved(records, command_id, previous, agent_id):
    replies = [x for x in records if x.get('id') == command_id and x.get('type') in ('command_ack', 'command_error')]
    assert len(replies) == 1, 'checkpoint needs one terminal command reply'
    ack = replies[0]
    assert ack.get('type') == 'command_ack' and ack.get('status') == 'saved', 'checkpoint refused: ' + str(ack.get('code'))
    assert type(ack.get('as_of')) is int and ack['as_of'] > 0, 'missing captured anchor'
    assert type(ack.get('watch_revision')) is int and ack['watch_revision'] > 0, 'missing applied watch revision'
    changes = [x for x in records[:records.index(ack)] if x.get('type') == 'observation' and x.get('revision') == ack['watch_revision']]
    assert len(changes) == 1, 'matching applied observation must precede checkpoint acknowledgement'
    observation = changes[0]['observation']
    assert observation.get('kind') == 'session_changed' and observation.get('agent_id') == agent_id, 'wrong session observation owner/kind'
    current = session(observation.get('session'))
    assert current == dict(previous, checkpoint_seq=ack['as_of']), 'save changed SessionID/mount provenance or acknowledged a different anchor'
    return ack, current


def subscribe(client, identifier='subscribe'):
    start = len(client.records)
    client.send(dict(type='subscribe', id=identifier))
    end = client.until(lambda x: x.get('type') == 'snapshot_end')
    records = client.records[start:]
    assert records[0].get('type') == 'preferences_snapshot', 'preferences/snapshot order'
    begin = next(x for x in records if x.get('type') == 'snapshot_begin')
    assert begin['id'] == identifier and begin['generation'] == end['generation'] and begin['watermark'] == end['watermark'], 'snapshot boundary mismatch'
    assert 'session' in begin['state'] and 'job_access' in begin['state'], 'new safe state absent'
    accesses = begin['state']['job_access']
    assert isinstance(accesses, list) and all(set(x) == {'handle', 'live'} and type(x['handle']) is int and x['handle'] > 0 and type(x['live']) is bool for x in accesses), 'job availability shape'
    assert [x['handle'] for x in accesses] == sorted({x['handle'] for x in accesses}), 'job availability order/duplicates'
    assert CANARY not in json.dumps(records), 'configuration secret leaked'
    return begin


def checkpoint(client, current, agent_id, identifier):
    start = len(client.records)
    client.send(dict(type='checkpoint', id=identifier))
    client.until(lambda x: x.get('id') == identifier and x.get('type') in ('command_ack', 'command_error'))
    return saved(client.records[start:], identifier, current, agent_id)


class Backend(http.server.ThreadingHTTPServer):
    def __init__(self):
        super().__init__(('127.0.0.1', 0), Handler)
        self.requests = []
        self.admitted = threading.Event()
        self.release = threading.Event()
        self.release.set()
        self.batches = []


class Handler(http.server.BaseHTTPRequestHandler):
    def log_message(self, *_):
        pass

    def do_POST(self):
        self.server.requests.append(json.loads(self.rfile.read(int(self.headers['Content-Length']))))
        self.server.admitted.set()
        if not self.server.release.wait(8):
            self.send_error(503, 'local barrier not released')
            return
        calls = self.server.batches.pop(0) if self.server.batches else []
        raw = json.dumps(response('openai', calls)).encode()
        self.send_response(200)
        self.send_header('Content-Type', 'application/json')
        self.send_header('Content-Length', str(len(raw)))
        self.end_headers()
        try:
            self.wfile.write(raw)
        except (BrokenPipeError, ConnectionResetError):
            pass


class Program:
    def __init__(self, binary, work, backend, arguments, standalone=False):
        env = environment(work, 'openai', f'http://127.0.0.1:{backend.server_port}')
        env.pop('CH02_LOG', None)
        env['LLM_API_KEY'] = CANARY
        if standalone:
            env['CH02_LOG'] = str(work / 'standalone.log')
        self.lines = queue.Queue()
        self.output = []
        self.process = subprocess.Popen([str(binary), *arguments], cwd=work, env=env,
                                        stdin=subprocess.PIPE, stdout=subprocess.PIPE,
                                        stderr=subprocess.STDOUT, text=True, start_new_session=True)
        self.reader = threading.Thread(target=self.read, daemon=True)
        self.reader.start()

    def read(self):
        for line in self.process.stdout:
            self.output.append(line)
            self.lines.put(line)
        self.lines.put(None)

    def listen(self):
        for _ in range(100):
            line = self.lines.get(timeout=6)
            assert line is not None, 'GUI exited before listening: ' + ''.join(self.output)[-1500:]
            match = re.search(r'http://127\.0\.0\.1:\d+', line)
            if match:
                return match.group()
        raise AssertionError('listening URL absent')

    def stop(self, sig=signal.SIGTERM):
        if self.process.poll() is None:
            os.killpg(self.process.pid, sig)
        code = self.process.wait(timeout=8)
        self.reader.join(timeout=2)
        if self.process.stdin and not self.process.stdin.closed:
            self.process.stdin.close()
        return code

    def close(self):
        try:
            self.stop()
        except subprocess.TimeoutExpired:
            os.killpg(self.process.pid, signal.SIGKILL)
            self.process.wait()
        self.process.stdout.close()


def evaluate(cli, gui, source, argument_order):
    binaries = {str(p): digest(p) for p in (cli, gui)}
    source_before = source_identities(source)
    checker_files = [Path(__file__), HERE / 'accept_ch07.py', HERE / 'accept_ch09.py', HERE / 'accept_ch10.py', HERE / 'accept_ch10_public.py']
    checkers = {p.name: digest(p) for p in checker_files}
    rows = []
    backend = Backend()
    worker = threading.Thread(target=backend.serve_forever, daemon=True)
    worker.start()
    with tempfile.TemporaryDirectory(prefix='ch10-clients-') as temporary:
        root = Path(temporary)

        def run(name, action):
            before = len(backend.requests)
            try:
                details = action(root / name)
                rows.append(dict(id=name, passed=True, details=details, http_requests=len(backend.requests) - before))
            except Exception as error:
                rows.append(dict(id=name, passed=False, details=type(error).__name__ + ': ' + str(error), http_requests=len(backend.requests) - before))
            finally:
                backend.release.set()

        def cli_local(work):
            work.mkdir()
            store = work / 'selected'
            flags = ['--session-dir', str(store)]
            command = [str(cli), *(flags + ['chat'] if argument_order == 'flags-first' else ['chat'] + flags)]
            env = environment(work, 'openai', f'http://127.0.0.1:{backend.server_port}')
            env.pop('CH02_LOG', None)
            before = len(backend.requests)
            receipts = []
            for commands in ('/session\n/checkpoint\n/quit\n', '/session\n'):
                result = subprocess.run(command, input=commands, text=True, capture_output=True, cwd=work, env=env, timeout=10)
                assert result.returncode == 0, result.stderr[-1500:]
                value, _ = inspect_created(store)
                assert value['session_id'] in result.stdout, '/session omitted durable identity'
                receipts.append(dict(exit=result.returncode, stdout=result.stdout, stderr=result.stderr, checkpoint_anchor=uint(value['as_of'])))
            assert len(backend.requests) == before, 'local session/save/EOF commands sent HTTP'
            return receipts

        def wire(work, standalone=False, busy=False, terminal=False):
            work.mkdir()
            store = work / 'selected'
            args = ['--port', '0'] + ([] if standalone else ['--session-dir', str(store)]) + (['--terminal'] if terminal else [])
            before = len(backend.requests)
            with contextlib.closing(Program(gui, work, backend, args, standalone)) as program:
                url = program.listen()
                with contextlib.closing(Socket(url)) as client:
                    begin = subscribe(client)
                    current = begin['state']['session']
                    if standalone:
                        assert current is None, 'standalone created a session'
                        client.send(dict(type='checkpoint', id='standalone-save'))
                        error = client.until(lambda x: x.get('id') == 'standalone-save')
                        assert error.get('type') == 'command_error' and error.get('code') == 'session_conflict', 'standalone checkpoint must refuse, not create a store'
                        assert not (work / '.ensemble' / 'session').exists()
                    else:
                        session(current, False)
                        checkpoint_path = store / 'checkpoint.json'
                        initial_checkpoint = checkpoint_path.read_bytes() if checkpoint_path.exists() else None
                        expected_anchor = uint(envelope(initial_checkpoint)['as_of']) if initial_checkpoint is not None else None
                        assert current['checkpoint_seq'] == expected_anchor, 'snapshot claimed an uncommitted checkpoint'
                        assert not begin['state']['job_access'], 'empty session invented jobs'
                        if terminal:
                            program.process.stdin.close()
                        if busy:
                            backend.admitted.clear()
                            backend.release.clear()
                            client.send(dict(type='prompt', id='active', text='one local controlled answer'))
                            assert backend.admitted.wait(4), 'prompt did not reach held HTTP barrier'
                            client.send(dict(type='checkpoint', id='busy-save'))
                            refusal = client.until(lambda x: x.get('id') == 'busy-save')
                            assert refusal.get('type') == 'command_error' and refusal.get('code') == 'session_busy', 'active capture waited or did not return busy'
                            after_busy = checkpoint_path.read_bytes() if checkpoint_path.exists() else None
                            assert after_busy == initial_checkpoint, 'busy capture changed checkpoint'
                            backend.release.set()
                            client.until(lambda x: x.get('type') == 'completion')
                        ack, current = checkpoint(client, current, begin['agent_id'], 'save')
                        value = envelope((store / 'checkpoint.json').read_bytes())
                        assert uint(value['as_of']) == ack['as_of'], 'file/ack anchor differs'
                        with contextlib.closing(Socket(url)) as second:
                            view = subscribe(second)
                            assert view['state']['session'] == current, 'new watcher missed committed session'
                        client.close()
                        with contextlib.closing(Socket(url)) as replacement:
                            view = subscribe(replacement)
                            assert view['state']['session'] == current, 'client close disposed session'
                            checkpoint(replacement, current, view['agent_id'], 'save-reconnected')
                    assert program.process.poll() is None, 'terminal EOF or socket close stopped GUI'
                assert program.stop() >= 0, 'GUI terminated by signal instead of closing'
            assert len(backend.requests) - before == int(busy), 'unexpected model request count'
            if not standalone:
                inspect_created(store)
            return dict(standalone=standalone, held_http=busy, terminal_eof=terminal)

        def lifetime(work, sig):
            work.mkdir()
            store = work / '.ensemble' / 'session'
            before = len(backend.requests)
            with contextlib.closing(Program(gui, work, backend, ['--port', '0'])) as first:
                url = first.listen()
                with contextlib.closing(Socket(url)) as client:
                    old = subscribe(client)
                    original = session(old['state']['session'], False)
                    inode = (store / 'owner.lock').stat().st_ino
                    initial_checkpoint = (store / 'checkpoint.json').read_bytes() if (store / 'checkpoint.json').exists() else None
                # Only an actual mounted positive can exercise lock death/release.
                with contextlib.closing(Program(gui, work, backend, ['--port', '0'])) as contender:
                    assert contender.process.wait(timeout=4) != 0, 'second writer entered mounted store'
                    contender.reader.join(timeout=2)
                    assert 'session_in_use' in ''.join(contender.output), 'wrong second-writer refusal'
                exit_code = first.stop(sig)
                assert exit_code == -signal.SIGKILL if sig == signal.SIGKILL else exit_code >= 0, 'signal did not follow required lifetime'
            if sig != signal.SIGKILL:
                inspect_created(store)
            else:
                after_kill = (store / 'checkpoint.json').read_bytes() if (store / 'checkpoint.json').exists() else None
                assert after_kill == initial_checkpoint, 'forced-death fixture changed its settled checkpoint before SIGKILL'
            with contextlib.closing(Program(gui, work, backend, ['--port', '0'])) as resumed:
                url = resumed.listen()
                with contextlib.closing(Socket(url)) as client:
                    view = subscribe(client)
                    current = session(view['state']['session'], True)
                    assert current['id'] == original['id'], 'restart changed SessionID'
                    assert (store / 'owner.lock').stat().st_ino == inode, 'close replaced lock inode'
                    assert not view['state']['job_access'], 'empty restart invented live jobs'
                    checkpoint(client, current, view['agent_id'], 'restart-save')
                assert resumed.stop() >= 0
            assert len(backend.requests) == before, 'restart/signal invoked model'
            return dict(signal=sig.name, first_exit=exit_code, same_lock_inode=True)

        def historical_job(work):
            work.mkdir()
            store = work / 'selected'
            args = ['--port', '0', '--session-dir', str(store)]
            before = len(backend.requests)
            backend.batches = [[dict(id='ch10-job-call', name='run_command', arguments=dict(command='printf CH10_JOB_BOUNDARY'))], []]
            with contextlib.closing(Program(gui, work, backend, args)) as first:
                url = first.listen()
                with contextlib.closing(Socket(url)) as client:
                    subscribe(client)
                    client.send(dict(type='prompt', id='job', text='perform the controlled local job'))
                    completion = client.until(lambda x: x.get('type') == 'completion')
                    assert completion.get('outcome') == 'success', 'local tool-positive turn failed'
                with contextlib.closing(Socket(url)) as observer:
                    live = subscribe(observer)
                    access = live['state']['job_access']
                    assert len(access) == 1 and access[0]['live'] is True, 'actual owned Job missing from positive snapshot'
                    handle = access[0]['handle']
                assert first.stop() >= 0
            _, events = inspect_created(store)
            calls = [x for x in events if x['type'] == 'tool_called']
            assert len(calls) == 1 and calls[0]['tool']['name'] == 'run_command', 'seed did not admit exactly one real tool'
            accepted = (store / 'events.log').read_bytes()
            assert b'CH10_JOB_BOUNDARY' in accepted, 'actual tool output absent'
            before_resume = len(backend.requests)
            with contextlib.closing(Program(gui, work, backend, args)) as second:
                url = second.listen()
                with contextlib.closing(Socket(url)) as client:
                    view = subscribe(client)
                    assert view['state']['job_access'] == [dict(handle=handle, live=False)], 'historical Job reacquired live ownership or disappeared'
                    assert view['state']['session']['id'] == live['state']['session']['id']
                    assert view['omitted'] == live['omitted'], 'bookkeeping consumed renderable slots'
                    assert (store / 'events.log').read_bytes() == accepted, 'mount invented terminal/job facts'
                    checkpoint(client, view['state']['session'], view['agent_id'], 'historical-save')
                assert second.stop() >= 0
            assert len(backend.requests) == before_resume == before + 2, 'resume replayed work or tool-positive used wrong request count'
            assert (store / 'events.log').read_bytes() == accepted, 'historical close invented job_killed'
            return dict(handle=handle, original_log_sha256=hashlib.sha256(accepted).hexdigest(), requests=2)

        try:
            run('human-local-checkpoint-quit-and-eof', cli_local)
            run('gui-checkpoint-reconnect', wire)
            run('gui-standalone-refusal', lambda p: wire(p, standalone=True))
            run('gui-held-http-busy-then-save', lambda p: wire(p, busy=True))
            run('gui-terminal-eof-detaches', lambda p: wire(p, terminal=True))
            for sig in (signal.SIGINT, signal.SIGTERM, signal.SIGKILL):
                run('gui-' + sig.name.lower() + '-lock-lifetime', lambda p, s=sig: lifetime(p, s))
            run('gui-historical-job-ownership-after-restart', historical_job)
        finally:
            backend.release.set()
            backend.shutdown()
            backend.server_close()
            worker.join(timeout=2)
    assert binaries == {str(p): digest(p) for p in (cli, gui)}, 'binary changed during checks'
    assert source_before == source_identities(source), 'supplied source changed during checks'
    assert checkers == {p.name: digest(p) for p in checker_files}, 'checker changed during checks'
    return dict(passed=len(rows) == 9 and all(x['passed'] for x in rows), checks=rows, binary_files=binaries,
                source_files=source_before, checker_files=checkers, scope=__doc__, full_chapter_acceptance=False)


if __name__ == '__main__':
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('cli_binary', type=Path)
    parser.add_argument('gui_binary', type=Path)
    parser.add_argument('--source-directory', type=Path, required=True)
    parser.add_argument('--argument-order', choices=('flags-first', 'mode-first'), default='flags-first')
    parser.add_argument('--receipt', type=Path)
    args = parser.parse_args()
    result = evaluate(args.cli_binary.resolve(strict=True), args.gui_binary.resolve(strict=True), args.source_directory.resolve(strict=True), args.argument_order)
    if args.receipt:
        args.receipt.write_text(json.dumps(result, indent=2) + '\n')
    print(json.dumps(result, indent=2))
    raise SystemExit(0 if result['passed'] else 1)
