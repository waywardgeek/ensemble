#!/usr/bin/env python3
"""Explicit disposable behavior deletions after a frozen-source passing parent.

Never mutates the source tree or weakens production identity verification. The
Go overlay's changed file bytes and their relationship to the verified parent
are recorded separately. Compile/setup failure cannot count as a detected mutant.
"""
import argparse
import hashlib
import importlib.util
import json
import subprocess
import tempfile
from pathlib import Path

HERE = Path(__file__).resolve().parent
SPEC = importlib.util.spec_from_file_location('ch10_lifecycle', HERE / 'accept_ch10_lifecycle.py')
LIFE = importlib.util.module_from_spec(SPEC)
SPEC.loader.exec_module(LIFE)
FIXTURES = {**LIFE.FILES, 'ch10_independent_admission_test.go': 'ch10-session-admission_test.go'}


def mutation(name, path, old, new, test, message, scope=None, failed=None):
    return dict(name=name, path=path, old=old, new=new, scope=scope, test=test,
                expected_message=message, expected_failed_tests=failed or [test])


STORE = 'internal/persistence/store.go'
REPLACE = 'func (s *Store) ReplaceCheckpoint(raw []byte) error {'
PARENT = 'TestCh10ReviewCheckpointFaults'


def write_mutation(name, old, new, leaf):
    return mutation(name, STORE, old, new, PARENT + '/' + leaf,
                    'wanted session_io, got <nil>', REPLACE, [PARENT, PARENT + '/' + leaf])


MUTATIONS = [
    write_mutation('ignore_short_checkpoint_write', 'if err == nil && n != len(raw) {',
                   'if err == nil && n < 0 {', 'short'),
    write_mutation('ignore_checkpoint_sync_error', 'err = f.Sync()', '_ = f.Sync()', 'sync'),
    write_mutation('ignore_checkpoint_close_error', 'err = closeErr', '_ = closeErr', 'close'),
    write_mutation('ignore_checkpoint_replace_error', 'err = s.io.Rename(name, filepath.Join(s.path, "checkpoint.json"))',
                   '_ = s.io.Rename(name, filepath.Join(s.path, "checkpoint.json"))', 'replace'),
    mutation('hide_terminal_append_fact', 'ensemble.go', 'return a.appendFailure', 'return nil',
             'TestCh10ReviewAppendFailurePreservesCheckpoint', 'faulted close attempted final checkpoint',
             'func (a turnAgent) AppendFailure() error {'),
    mutation('omit_public_append_terminal_propagation', 'internal/llm/actor.go',
             '\t\t\t\tif terminal := a.parent.AppendFailure(); terminal != nil {\n\t\t\t\t\ta.persistence(terminal)\n\t\t\t\t}',
             '\t\t\t\t// Deleted public-append terminal propagation.',
             'TestCh10ReviewAppendFaultStopsAdmission', 'terminal append failure left request admission open'),
    mutation('release_store_lock_during_worker', STORE, '\ts.busy = true\n',
             '\ts.busy = true\n\t_ = s.lock.Close() // deliberately release owned lock too soon\n',
             'TestCh10ReviewCloseJoinsWriterBeforeUnlock', 'wanted session_in_use, got <nil>',
             'func (s *Store) Begin(cp common.Checkpoint, save bool) (<-chan WriteResult, error) {'),
    mutation('omit_checkpoint_worker_join', 'internal/llm/actor.go', '\t\ta.workers.Wait()\n',
             '\t\t// Deleted worker join.\n', 'TestCh10ReviewCloseJoinsWriterBeforeUnlock',
             ['close returned before held writer', 'barrier did not complete: first save completion', 'session_busy:']),
    mutation('drop_post_checkpoint_job_tail', 'session.go', '\tif err := a.append(event, false, false); err != nil {',
             '\tif r.Checkpoint == nil && event.Type == "job_ended" { return nil }\n\tif err := a.append(event, false, false); err != nil {',
             'TestCh10ReviewIndependentJobTail', 'job tail was lost or startup sent HTTP',
             'func (a *Agent) AcceptSessionRecord(event common.Event, line int, r *common.SessionReadState) error {'),
]


def changed_bytes(original, spec):
    text = original.decode()
    if spec['scope']:
        if text.count(spec['scope']) != 1:
            raise ValueError('mutation scope is not unique: ' + spec['name'])
        start = text.index(spec['scope'])
        end = text.find('\nfunc ', start + len(spec['scope']))
        end = len(text) if end < 0 else end
    else:
        start, end = 0, len(text)
    body = text[start:end]
    if body.count(spec['old']) != 1:
        raise ValueError('mutation anchor is not unique: ' + spec['name'])
    return (text[:start] + body.replace(spec['old'], spec['new']) + text[end:]).encode()


def command_result(command, source):
    try:
        p = subprocess.run(command, cwd=source, text=True, capture_output=True, timeout=100)
        row = dict(command=command, exit=p.returncode, stdout=p.stdout, stderr=p.stderr)
    except subprocess.TimeoutExpired as e:
        def text(v):
            return v.decode(errors='replace') if isinstance(v, bytes) else (v or '')
        row = dict(command=command, exit=None, stdout=text(e.stdout), stderr=text(e.stderr), timeout_seconds=100)
    events = []
    for line in row['stdout'].splitlines():
        try:
            event = json.loads(line)
            if isinstance(event, dict):
                events.append(event)
        except ValueError:
            pass
    row['failed_tests'] = sorted({e['Test'] for e in events if e.get('Action') == 'fail' and e.get('Test')})
    row['passed_tests'] = sorted({e['Test'] for e in events if e.get('Action') == 'pass' and e.get('Test')})
    row['test_output'] = ''.join(e.get('Output', '') for e in events)
    return row


def evaluate(source, revision, receipt, selected):
    commit, identities = LIFE.bind(source, revision)
    checker_paths = [Path(__file__), HERE / 'accept_ch10_lifecycle.py', *(HERE / p for p in FIXTURES.values())]
    checkers = {p.name: LIFE.digest(p) for p in checker_paths}
    report = dict(passed=False, complete_run=False, source_commit=commit, source_files=identities,
                  source_directory=str(source), checker_files=checkers, commands=[], mutations=[], scope=__doc__)
    LIFE.save_receipt(receipt, report)
    def run(command):
        report['active_command'] = command
        LIFE.save_receipt(receipt, report)
        row = command_result(command, source)
        report['commands'].append(row)
        report['active_command'] = None
        LIFE.save_receipt(receipt, report)
        return row
    with tempfile.TemporaryDirectory(prefix='ch10-fault-deletions-') as directory:
        work = Path(directory)
        overlay_files = {}
        for virtual, name in FIXTURES.items():
            if (source / virtual).exists():
                raise ValueError('overlay collision: ' + virtual)
            local = work / name
            local.write_bytes((HERE / name).read_bytes())
            overlay_files[str(source / virtual)] = str(local)
        overlay = work / 'overlay.json'
        overlay.write_text(json.dumps({'Replace': overlay_files}))
        fmt = run(['gofmt', '-l', *overlay_files.values()])
        if fmt['exit'] != 0 or fmt['stdout']:
            return report
        vet = run(['go', 'vet', '-overlay', str(overlay), '.'])
        if vet['exit'] != 0:
            return report
        positive = run(['go', 'test', '-json', '-overlay', str(overlay), '-race', '-count=1', '-timeout=45s', '-run', '^TestCh10Review', '.'])
        required = {m['test'] for m in MUTATIONS if not selected or m['name'] in selected}
        if positive['exit'] != 0 or not required.issubset(set(positive['passed_tests'])):
            return report
        report['positive_parent_passed'] = True
        for spec in MUTATIONS:
            if selected and spec['name'] not in selected:
                continue
            original = (source / spec['path']).read_bytes()
            modified = changed_bytes(original, spec)
            mutation_file = work / (spec['name'] + '.go')
            mutation_file.write_bytes(modified)
            replacements = dict(overlay_files, **{str(source / spec['path']): str(mutation_file)})
            overlay.write_text(json.dumps({'Replace': replacements}))
            selector = '/'.join('^' + part + '$' for part in spec['test'].split('/'))
            result = run(['go', 'test', '-json', '-overlay', str(overlay), '-race', '-count=1', '-timeout=20s', '-run', selector, '.'])
            messages = spec['expected_message']
            messages = messages if isinstance(messages, list) else [messages]
            detected = (result['exit'] == 1 and result['failed_tests'] == sorted(spec['expected_failed_tests'])
                        and any(message in result['test_output'] for message in messages)
                        and '[build failed]' not in result['test_output'])
            report['mutations'].append(dict(**spec, original_sha256=hashlib.sha256(original).hexdigest(),
                                           modified_sha256=hashlib.sha256(modified).hexdigest(),
                                           command_index=len(report['commands'])-1, detected=detected))
            LIFE.save_receipt(receipt, report)
            if not detected:
                break
    report['source_unchanged'] = identities == LIFE.files(source)
    report['checker_unchanged'] = checkers == {p.name: LIFE.digest(p) for p in checker_paths}
    expected = len(selected) if selected else len(MUTATIONS)
    report.update(complete_run=True, passed=len(report['mutations']) == expected and all(m['detected'] for m in report['mutations'])
                  and report['source_unchanged'] and report['checker_unchanged'])
    LIFE.save_receipt(receipt, report)
    return report


if __name__ == '__main__':
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('source_directory', type=Path)
    parser.add_argument('--source-commit', required=True)
    parser.add_argument('--receipt', required=True, type=Path)
    parser.add_argument('--select', action='append', choices=[m['name'] for m in MUTATIONS])
    args = parser.parse_args()
    result = evaluate(args.source_directory.resolve(strict=True), args.source_commit, args.receipt, set(args.select or []))
    print(json.dumps({'passed': result['passed'], 'complete_run': result['complete_run'], 'receipt': str(args.receipt)}))
    raise SystemExit(0 if result['passed'] else 1)
