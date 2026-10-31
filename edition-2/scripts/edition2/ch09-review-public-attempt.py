#!/usr/bin/env python3
"""Audit the preserved Anthropic P4 partial attempt; never certify run success.

Read-only replay after full source/launch/original validation. No provider call,
original modification, or relaxation of the student's success-only verifier.
"""
import argparse
import hashlib
import json
import os
from pathlib import Path
import subprocess
import sys

ROOT = Path(__file__).resolve().parents[2]
SUPPORT = ROOT / 'solutions/edition-2/main/evidence/ch09'
sys.path.insert(0, str(SUPPORT))
from evidence import preflight, check_launch, file_digest


def audit(run, binding_path):
    binding = json.loads(binding_path.read_text())
    preflight(binding)
    launch_path = run / 'launch.json'
    launch = json.loads(launch_path.read_text())
    check_launch(binding, launch)
    assert launch['scenario'] == 'P' and launch['vendor'] == 'anthropic'
    assert launch['exit_code'] == 1 and launch['requests'] == launch['http_cap'] == 4
    logs = ['workspace/alpha/events.jsonl', 'workspace/beta/events.jsonl']
    assert launch['logs'] == logs
    required = {'terminal.txt', *logs}
    required |= {f'{kind}/{i:03}.{suffix}' for i in range(1, 5)
                 for kind, suffix in [('requests', 'json'), ('responses', 'body')]}
    required |= {str(p.relative_to(run)) for p in (run / 'workspace').rglob('*') if p.is_file()}
    required |= {str(p.relative_to(run)) for p in (run / 'inputs').rglob('*') if p.is_file()}
    required |= {str(p.relative_to(run)) for p in (run / 'catalog').rglob('*') if p.is_file()}
    assert required == set(launch['originals']), 'incomplete original set'
    originals = {**launch['originals'], 'launch.json': file_digest(launch_path)}
    for path, expected in originals.items():
        actual = (run / path).resolve()
        assert actual.is_relative_to(run) and file_digest(actual) == expected, path
    before_binding = file_digest(binding_path)
    findings = []
    capture = 1
    for name, log in zip(('alpha', 'beta'), logs):
        events = [json.loads(line) for line in (run / log).read_text().splitlines()][1:]
        requests = [e for e in events if e['type'] == 'request_sent']
        responses = [e for e in events if e['type'] == 'response_ended']
        assert len(requests) == len(responses) == 2
        changes = [e for e in events if e['type'].startswith('skills_')]
        initial, latest = changes[0]['skills'], changes[-1]['skills']['state']
        ceiling = ['load_skill', 'read_file', 'unload_skill'] + (['write_file'] if name == 'alpha' else [])
        assert initial['ceiling'] == ceiling
        for change in changes:
            for material in change['skills']['activated']:
                assert hashlib.sha256(material['body'].encode()).hexdigest() == material['sha256']
                assert name + '-$TOOLS' in material['body']
        matched = []
        for request, response in zip(requests, responses):
            command = [binding['executables']['cli']['path'], 'replay', str(run / log), str(request['seq'])]
            body = subprocess.check_output(command, cwd=run,
                                           env={'PATH': os.environ['PATH']}, timeout=30).removesuffix(b'\n')
            raw = (run / f'requests/{capture:03}.json').read_bytes()
            assert body == raw, 'chronological request reconstruction mismatch'
            wire = json.loads(raw)
            assert [tool['name'] for tool in wire['tools']] == ceiling
            manual = '[skill ' + ('edit' if name == 'alpha' else 'inspect') + ' activation 2]'
            texts = [p.get('text', '') for m in wire['messages'] for p in m['content']]
            assert sum(manual in text for text in texts) == 1
            assert any(name + '-$TOOLS' in text for text in texts)
            assert ('beta' if name == 'alpha' else 'alpha') + '-$TOOLS' not in raw.decode()
            sse = [json.loads(line[6:]) for line in (run / f'responses/{capture:03}.body').read_text().splitlines() if line.startswith('data: ')]
            terminal = [e for e in sse if e.get('type') == 'message_delta'][-1]
            usage = response['response']['usage']
            assert usage['input'] == terminal['usage']['input_tokens']
            assert usage['output'] == terminal['usage']['output_tokens']
            matched.append(dict(capture=capture, request_seq=request['seq'], usage=usage))
            capture += 1
        calls = [e['tool']['name'] for e in events if e['type'] == 'tool_called']
        assert calls == (['read_file'] if name == 'alpha' else ['load_skill', 'read_file', 'unload_skill'])
        ended = [e['turn']['outcome'] for e in events if e['type'] == 'turn_ended']
        assert ended == (['success'] if name == 'alpha' else ['round_limit'])
        read_results = [e['tool'] for e in events if e['type'] == 'tool_returned' and e['tool'].get('job')]
        assert len(read_results) == 1
        result = read_results[0]
        artifact = run / 'workspace' / name / result['job']['output']['locator']
        note = (run / 'inputs' / name / 'notes.txt').read_bytes()
        assert artifact.read_bytes() == note == result['parts'][0]['text'].encode()
        assert (run / 'workspace' / name / 'notes.txt').read_bytes() == note
        assert latest['revision'] == (1 if name == 'alpha' else 2)
        assert latest['roots'] == (['edit'] if name == 'alpha' else [])
        findings.append(dict(agent=name, ceiling=ceiling, final_state=latest,
                             calls=calls, turn_outcome=ended[0], requests=matched,
                             note_and_artifact_sha256=hashlib.sha256(note).hexdigest()))
    assert capture == 5
    assert all(file_digest(run / p) == h for p, h in originals.items())
    assert file_digest(binding_path) == before_binding
    assert file_digest(binding['executables']['cli']['path']) == binding['executables']['cli']['sha256']
    return dict(scope=__doc__, audit_passed=True, attempt_success=False,
                exit_code=1, requests=4, http_cap=4, source_revision=binding['source_revision'],
                binding_sha256=before_binding, original_files=originals, originals_unchanged=True,
                source_files_verified=len(binding['sources']), executable_identities_verified=len(binding['executables']),
                findings=findings, checker_sha256=file_digest(Path(__file__)),
                limits='One failed/partial public attempt only. All four requests replay exactly in observed sequential order. Beta final prose answer absent. No authorization to retry or claim overall chapter acceptance.')


if __name__ == '__main__':
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('run', type=Path)
    parser.add_argument('--binding', type=Path, required=True)
    parser.add_argument('--receipt', type=Path, required=True)
    args = parser.parse_args()
    result = audit(args.run.resolve(strict=True), args.binding.resolve(strict=True))
    args.receipt.write_text(json.dumps(result, indent=2) + '\n')
    print(json.dumps({k: result[k] for k in ('audit_passed', 'attempt_success', 'requests', 'exit_code')}))
