"""Verify a complete logical run set while preserving historical run identities.

Each run refers to its own immutable source/executable binding. All bindings
and launch identities are checked before replay or derived-file creation.
Original receipts are only read; derivatives go to a separate destination.
"""
import argparse
from collections import Counter
import json
import os
from pathlib import Path
import subprocess

from evidence import HERE, ROOT, PREFIX, digest, preflight

EXPECTED = {mode + '-' + vendor for mode in ('controls', 'eof', 'workflow', 'collection')
            for vendor in ('anthropic', 'openai', 'gemini')}
EXPECTED |= {'input-anthropic', 'input-openai'}


def owned_path(name):
    path = (HERE / name).resolve()
    assert path.is_relative_to(HERE), 'receipt path escapes evidence directory'
    return path


def validate_binding(binding, paths):
    preflight(binding, paths)
    support = binding.get('addendum_support', {})
    if support:
        assert set(support) == {'terminal-run-addendum.py', 'verify-addendum.py'}, 'incomplete addendum support identities'
    for name, identity in support.items():
        assert digest((HERE / name).read_bytes()) == identity, 'addendum support mismatch'
        historical = subprocess.check_output(['git', 'show', binding['source_revision'] + ':' + PREFIX + 'evidence/ch05/' + name], cwd=ROOT)
        assert digest(historical) == identity, 'historical addendum support mismatch'


def validate_plan(plan, overrides):
    assert plan['verifier_sha256'] == digest(Path(__file__).read_bytes()), 'addendum verifier mismatch'
    assert set(plan['runs']) == EXPECTED, 'incomplete logical run set'
    assert plan['bindings'], 'empty binding set'
    bindings = {}
    for name, entry in plan['bindings'].items():
        data = owned_path(entry['path']).read_bytes()
        assert digest(data) == entry['sha256'], 'binding identity mismatch'
        binding = json.loads(data)
        paths = {n: d['path'] for n, d in binding['executables'].items()}
        paths.update(overrides.get(name, {}))
        validate_binding(binding, paths)
        bindings[name] = (binding, paths)
    launches = {}
    directories = set()
    for logical, entry in plan['runs'].items():
        run = owned_path(entry['directory'])
        assert run not in directories, 'one raw run cannot satisfy two logical runs'
        directories.add(run)
        binding, paths = bindings[entry['binding']]
        launch = json.loads((run / 'launch.json').read_text())
        assert launch['source_revision'] == binding['source_revision'], 'launch source mismatch'
        assert launch['executables'] == {n: d['sha256'] for n, d in binding['executables'].items()}, 'launch executable mismatch'
        assert launch['support'] == binding['support'], 'launch support mismatch'
        assert launch.get('addendum_support', {}) == binding.get('addendum_support', {}), 'launch addendum support mismatch'
        assert set(launch['executable_paths']) == set(binding['executables']), 'incomplete launch executable paths'
        mode, vendor = logical.split('-')
        expected_mode = 'chat' if mode in ('controls', 'eof', 'input') else mode
        executable = 'cli' if expected_mode == 'chat' else 'workflow'
        assert launch['mode'] == expected_mode and launch['vendor'] == vendor, 'launch provider/mode mismatch'
        assert launch['launched_executable'] == executable, 'launched executable mismatch'
        assert launch['command'] == [launch['executable_paths'][executable], expected_mode], 'launch command mismatch'
        if vendor == 'gemini':
            assert launch['requested_model'] == 'models/gemini-3.8-flash', 'Gemini model outside current validation target'
        assert launch['exit_code'] == 0, 'unsuccessful launch remains unvalidated'
        launches[logical] = (run, launch, binding, paths)
    assert set(plan['bindings']) == {r['binding'] for r in plan['runs'].values()}, 'unused binding in plan'
    return launches


def verify(plan, overrides, destination):
    launches = validate_plan(plan, overrides)
    assert not destination.exists(), 'derived destination already exists'
    env = {k: v for k, v in os.environ.items() if not k.startswith(('LLM_', 'ANTHROPIC_', 'OPENAI_', 'GEMINI_'))}
    rows, derivatives = {}, {}
    for logical, (run, launch, binding, paths) in launches.items():
        raw = {p.name: p.read_bytes() for p in sorted((run / 'requests').glob('*.json'))}
        assert raw and len(raw) == launch['requests'], 'incomplete raw request set'
        remaining = Counter(raw.values())
        requests, terminal = [], {}
        calls = results = 0
        usage = {k: 0 for k in ('input', 'cache_write', 'cache_read', 'output')}
        logs = sorted(run.glob('*.log'))
        assert logs, 'missing durable logs'
        hint_texts = {}
        input_calls = set()
        input_reports = set()
        for log in logs:
            records = [json.loads(line) for line in log.read_text().splitlines()]
            assert records[0] == {'log_version': 1}
            events = records[1:]
            assert all(e['seq'] == i + 1 for i, e in enumerate(events))
            called = [e['tool']['call_id'] for e in events if e['type'] == 'tool_called']
            returned = [e['tool']['call_id'] for e in events if e['type'] == 'tool_returned']
            assert len(called) == len(set(called)) and len(returned) == len(set(returned)) and set(called) == set(returned)
            calls += len(called)
            results += len(returned)
            ended = [e['job']['handle'] for e in events if e['type'] in ('job_ended', 'job_killed')]
            assert len(ended) == len(set(ended)), 'duplicate durable job terminal event'
            for e in events:
                if e['type'] == 'tool_called' and e['tool']['name'] == 'send_input':
                    input_calls.add(e['tool']['call_id'])
                if e['type'] == 'tool_returned' and any('send_input accepted ' in part.get('text', '') for part in e['tool'].get('parts', [])):
                    input_reports.add(e['tool']['call_id'])
                if e['type'] == 'hint_received':
                    hint_texts[e['seq']] = e['hint']['text']
                if e['type'] == 'response_ended':
                    for k in usage:
                        usage[k] += e['response']['usage'][k]
                if e['type'] in ('job_ended', 'job_killed'):
                    job = e['job']
                    data = (run / 'workspace' / job['output']['locator']).read_bytes()
                    assert len(data) == job['bytes']
                    terminal[str(job['handle'])] = {'status': job['status'], 'reason': job.get('reason'), 'sha256': digest(data), 'bytes': len(data)}
                if e['type'] != 'request_sent':
                    continue
                output = subprocess.check_output([paths['cli'], 'replay', str(log), str(e['seq'])], cwd=run / 'workspace', env=env).removesuffix(b'\n')
                assert remaining[output] > 0, f'{logical}/{log.name}:{e["seq"]} differs from raw wire request'
                remaining[output] -= 1
                hints = e['request']['hints']
                if logical == 'controls-gemini':
                    body = json.loads(output)
                    assert 'hints' not in body
                    parts = body['contents'][-1]['parts']
                    for seq in hints:
                        assert {'text': hint_texts[seq]} in parts
                    if hints:
                        positions = [i for i, part in enumerate(parts) if part == {'text': hint_texts[hints[0]]}]
                        result_positions = [i for i, part in enumerate(parts) if 'functionResponse' in part]
                        assert result_positions and max(result_positions) < positions[0], 'hint does not follow result group'
                    else:
                        assert all({'text': text} not in parts for text in hint_texts.values()), 'consumed hint leaked into following request'
                requests.append({'log': log.name, 'seq': e['seq'], 'sha256': digest(output), 'hints': hints})
                derivatives[logical + '/' + log.stem + '-' + str(e['seq']) + '.json'] = output
        assert not any(remaining.values()), 'wire request lacks exact historical prefix'
        transcript = (run / 'terminal.txt').read_bytes()
        assert transcript
        if launch['mode'] == 'chat':
            assert b'Final usage:' in transcript
        if logical.startswith('controls-'):
            for marker in (b'Hint received', b'interrupted=true', b'Command refused: Agent busy', b'/history', b'/usage'):
                assert marker in transcript, marker
            assert any(r['hints'] for r in requests)
            assert any(j['reason'] == 'shutdown' for j in terminal.values())
            assert any(j['reason'] == 'kill_job' for j in terminal.values())
        if logical.startswith('workflow-'):
            assert all((run / (role + '-completion.json')).is_file() for role in ('author', 'editor', 'reviewer'))
            draft = (run / 'workspace/draft.txt').read_text().lower()
            assert all(text in draft for text in ('saturday', '10', 'bring a broken lamp', 'repairs are free'))
        if logical.startswith('collection-'):
            c = json.loads((run / 'collection-completion.json').read_text())
            assert len(c['results']) == len(c['independent_results']) == 2 and c['canceled']['outcome'] == 'canceled' and c['exhausted']
        if logical.startswith('input-') or logical == 'controls-gemini':
            assert input_calls and input_calls <= input_reports, 'missing revised send_input receipt'
            assert any(j['reason'] == 'kill_job' for j in terminal.values())
        rows[logical] = {'directory': str(run.relative_to(HERE)), 'source_revision': binding['source_revision'],
                         'requested_model': launch['requested_model'], 'requests': requests, 'usage': usage,
                         'raw_requests': {n: digest(data) for n, data in raw.items()}, 'terminal_sha256': digest(transcript),
                         'calls': calls, 'results': results, 'jobs': terminal}
    # No derived files or directories are created until the entire plan passes.
    destination.mkdir()
    for name, data in derivatives.items():
        target = destination / name
        target.parent.mkdir(exist_ok=True)
        target.write_bytes(data)
    (destination / 'receipts.json').write_text(json.dumps({'plan': plan, 'runs': rows}, indent=2) + '\n')
    return rows


if __name__ == '__main__':
    parser = argparse.ArgumentParser()
    parser.add_argument('plan', type=Path)
    parser.add_argument('destination', type=Path)
    parser.add_argument('--executable', action='append', default=[], help='BINDING.NAME=PATH; identity remains hash-bound')
    args = parser.parse_args()
    overrides = {}
    for value in args.executable:
        identity, path = value.split('=', 1)
        binding, name = identity.split('.', 1)
        overrides.setdefault(binding, {})[name] = str(Path(path).resolve())
    rows = verify(json.loads(args.plan.read_text()), overrides, args.destination.resolve())
    print(json.dumps({name: {'requests': len(row['requests']), 'usage': row['usage']} for name, row in rows.items()}, indent=2))
