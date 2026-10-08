#!/usr/bin/env python3
"""Chapter 10 physical read boundaries through actual public CLI operations.

Creates an empty genuine session, then changes only physical JSON whitespace for
size controls. Files are streamed one at a time and removed between cases. This
does not prove bounded allocations, write admission, or complete acceptance.
No credentials or provider requests are needed. --self-test checks the fixture
writer and result classifier only, never a student runtime.
"""
import argparse
import hashlib
import json
from pathlib import Path
import subprocess
import tempfile

from accept_ch09 import environment

CAP = 64 * 1024 * 1024
OLD_CAP = 16 * 1024 * 1024
HEADER = b'{"log_version":1}'
HERE = Path(__file__).resolve().parent


def digest(path):
    h = hashlib.sha256()
    with path.open('rb') as stream:
        for block in iter(lambda: stream.read(65536), b''):
            h.update(block)
    return h.hexdigest()


def source_identity(root):
    return {str(p.relative_to(root)): digest(p) for p in sorted(root.rglob('*'))
            if p.is_file() and 'evidence' not in p.relative_to(root).parts and
            (p.suffix == '.go' or p.name in ('go.mod', 'go.sum'))}


def write_record(stream, raw, size=None, lf=True):
    size = len(raw) + int(lf) if size is None else size
    remaining = size - len(raw) - int(lf)
    if remaining < 0:
        raise ValueError('fixture record cannot fit requested physical size')
    stream.write(raw)
    block = b' ' * 65536
    while remaining:
        n = min(remaining, len(block))
        stream.write(block[:n])
        remaining -= n
    if lf:
        stream.write(b'\n')


def message(sequence):
    return json.dumps(dict(seq=sequence, time='2026-01-01T00:00:00Z',
                          type='message_received', message=dict(actor='system',
                          purpose='instruction', parts=[dict(type='text',
                          text='CH10-PHYSICAL-BOUNDARY')])),
                      separators=(',', ':')).encode()


def invoke(binary, work, command, stdin=''):
    env = environment(work, 'openai', 'http://127.0.0.1:1')
    env.pop('CH02_LOG', None)
    env.pop('LLM_API_KEY', None)
    # Construction receives a dummy local-only credential, never a user key.
    if command[-1:] == ['protocol']:
        env['LLM_API_KEY'] = 'CH10_EMPTY_SESSION_LOCAL_ONLY'
    argv = [str(binary), *map(str, command)]
    try:
        p = subprocess.run(argv, input=stdin, cwd=work, env=env, text=True,
                           capture_output=True, timeout=75)
        return dict(argv=argv, exit=p.returncode, stdout=p.stdout[:4000],
                    stderr=p.stderr[:4000], timeout=False)
    except subprocess.TimeoutExpired as e:
        def safe(value):
            return (value.decode(errors='replace') if isinstance(value, bytes)
                    else value or '')[:4000]
        return dict(argv=argv, exit=None, stdout=safe(e.stdout),
                    stderr=safe(e.stderr), timeout=True)


def outcome(run, want, explicit):
    if run['timeout']:
        return False
    if want:
        if run['exit'] != 0:
            return False
        try:
            return isinstance(json.loads(run['stdout']), dict)
        except ValueError:
            return False
    if run['exit'] in (None, 0) or run['stdout'].strip():
        return False
    return ('session_corrupt' if explicit else 'line') in run['stderr']


def cases(initial, small):
    """Each tuple: ID, physical records, expected pass, explicit-session check."""
    ordinary = message(2)
    basic = [(HEADER, None, True), (initial, None, True)]
    yield 'session-initial-parent', basic, True, True
    yield 'session-ordinary-parent', basic + [(ordinary, None, True)], True, True
    yield 'standalone-parent', [(HEADER, None, True), (message(1), None, True)], True, False
    if small:
        return
    for lf in (True, False):
        ending = 'LF' if lf else 'EOF'
        for extra in (0, 1):
            label = 'exact' if extra == 0 else 'one-over'
            yield f'session-initial-{label}-{ending}', [
                (HEADER, None, True), (initial, CAP + extra, lf)], extra == 0, True
            yield f'session-ordinary-{label}-{ending}', basic + [
                (ordinary, CAP + extra, lf)], extra == 0, True
    # The inherited Scanner-style standalone boundary requires spare capacity
    # to discover EOF: exact LF passes, exact unterminated ordinary record fails.
    for label, size, lf, want in [('exact-LF', OLD_CAP, True, True),
                                 ('below-EOF', OLD_CAP - 1, False, True),
                                 ('exact-EOF', OLD_CAP, False, False),
                                 ('one-over-LF', OLD_CAP + 1, True, False)]:
        yield 'standalone-' + label, [(HEADER, None, True),
                                      (message(1), size, lf)], want, False
    for extra in (0, 1):
        yield 'session-header-' + ('exact' if extra == 0 else 'one-over'), [
            (HEADER, OLD_CAP + extra, True), (initial, None, True)], extra == 0, True
    # A syntactically shaped initializer with invalid identity cannot authorize
    # larger standalone admission. Verify small semantic refusal as well.
    invalid = json.loads(initial)
    invalid['session']['session_id'] = 'not-a-session-id'
    invalid = json.dumps(invalid, separators=(',', ':')).encode()
    interior = json.loads(initial)
    interior['seq'] = 2
    interior = json.dumps(interior, separators=(',', ':')).encode()
    for label, raw, prefix in [('invalid-initial', invalid, [(HEADER, None, True)]),
                                ('interior-initial', interior, [(HEADER, None, True),
                                                               (message(1), None, True)])]:
        for size in (None, OLD_CAP + 1):
            yield label + ('-small' if size is None else '-oversize'), prefix + [
                (raw, size, True)], False, True


def evaluate(binary, source, small):
    before = source_identity(source)
    if not before:
        raise ValueError('empty source identity')
    bound = {str(p): digest(p) for p in (binary, Path(__file__), HERE / 'accept_ch09.py')}
    rows = []
    with tempfile.TemporaryDirectory(prefix='ch10-physical-') as temp:
        work = Path(temp)
        original = work / 'original'
        create = invoke(binary, work, ['--session-dir', original, 'protocol'])
        inspect = invoke(binary, work, ['session', 'inspect', original])
        valid = create['exit'] == 0 and outcome(inspect, True, True)
        rows.append(dict(id='genuine-empty-session', passed=valid, create=create, inspect=inspect))
        initial = None
        if valid:
            lines = (original / 'events.log').read_bytes().splitlines()
            if len(lines) == 2 and lines[0] == HEADER:
                candidate = json.loads(lines[1])
                if candidate.get('seq') == 1 and candidate.get('type') == 'session_initialized':
                    initial = lines[1]
        if initial is None:
            rows.append(dict(id='empty-session-initializer', passed=False,
                             blocked='genuine one-event empty session required'))
        else:
            target = work / 'candidate'
            target.mkdir()
            log = target / 'events.log'
            parents = {}
            parent_output = {}
            for name, records, want, explicit in cases(initial, small):
                with log.open('wb') as stream:
                    for raw, size, lf in records:
                        write_record(stream, raw, size, lf)
                original_hash = digest(log)
                physical_bytes = log.stat().st_size
                for surface in ('generic', 'explicit') if explicit else ('generic',):
                    run = invoke(binary, work, ['session', 'inspect', target]
                                 if surface == 'explicit' else ['render', log])
                    passed = outcome(run, want, surface == 'explicit')
                    if name.endswith('-parent'):
                        parents[(name, surface)] = passed
                        if passed:
                            parent_output[(name, surface)] = json.loads(run['stdout'])
                    parent_name = ('standalone-parent' if name.startswith('standalone-')
                                   else 'session-ordinary-parent' if name.startswith('session-ordinary-')
                                   else 'session-initial-parent')
                    parent = parents.get((parent_name, surface), False)
                    same_projection = (not want or not passed or
                                       json.loads(run['stdout']) == parent_output.get((parent_name, surface)))
                    unchanged = digest(log) == original_hash
                    rows.append(dict(id=surface + '/' + name,
                                     passed=passed and parent and unchanged and same_projection,
                                     expected_accept=want, positive_parent=parent,
                                     same_projection=same_projection,
                                     physical_file_bytes=physical_bytes,
                                     records=[dict(bytes=len(raw) + int(lf) if size is None else size,
                                                   framing_lf=lf) for raw, size, lf in records],
                                     input_sha256=original_hash, unchanged=unchanged, run=run))
                # Only this fixture's scratch file is removed. No candidate is
                # mounted live, so the reader may leave at most its lock file.
                log.unlink()
    unchanged = before == source_identity(source)
    binding_unchanged = bound == {p: digest(Path(p)) for p in bound}
    return dict(passed=all(r['passed'] for r in rows) and unchanged and binding_unchanged,
                checks=rows, source_directory=str(source), source_files=before,
                source_unchanged=unchanged, checker_and_binary=bound,
                binding_unchanged=binding_unchanged, small_only=small,
                scope=__doc__, full_chapter_acceptance=False,
                limitations='File-backed read-side controls only. No allocation instrumentation, '
                            'write bounds, skill 64MiB envelope boundary, nesting, total store bound, '
                            'paid use or complete inherited acceptance. Binary hash is retained; '
                            'source map records observed inputs, not an attested build relationship.')


def self_test():
    import io
    rows = []
    for lf in (True, False):
        stream = io.BytesIO()
        write_record(stream, b'{"x":1}', 27, lf)
        raw = stream.getvalue()
        rows.append(len(raw) == 27 and raw.endswith(b'\n') == lf and json.loads(raw) == {'x': 1})
    good = dict(exit=0, stdout='{}', stderr='', timeout=False)
    bad = dict(exit=1, stdout='', stderr='session_corrupt: line 2 bound', timeout=False)
    wrong = dict(exit=1, stdout='', stderr='missing executable', timeout=False)
    rows += [outcome(good, True, True), outcome(bad, False, True),
             outcome(bad, False, False), not outcome(wrong, False, True),
             not outcome(wrong, False, False), not outcome(bad, True, True),
             not outcome(dict(good, timeout=True), True, True)]
    try:
        write_record(io.BytesIO(), b'{}', 1)
        rows.append(False)
    except ValueError:
        rows.append(True)
    return dict(passed=all(rows), oracle_controls=len(rows), runtime_run=False)


if __name__ == '__main__':
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('binary', nargs='?', type=Path)
    parser.add_argument('--source-directory', type=Path)
    parser.add_argument('--small-only', action='store_true')
    parser.add_argument('--self-test', action='store_true')
    parser.add_argument('--receipt', type=Path)
    args = parser.parse_args()
    if args.self_test:
        result = self_test()
    else:
        if args.binary is None or args.source_directory is None:
            parser.error('binary and --source-directory are required')
        result = evaluate(args.binary.resolve(strict=True),
                          args.source_directory.resolve(strict=True), args.small_only)
    if args.receipt:
        args.receipt.write_text(json.dumps(result, indent=2) + '\n')
    print(json.dumps(result, indent=2))
    raise SystemExit(0 if result['passed'] else 1)
