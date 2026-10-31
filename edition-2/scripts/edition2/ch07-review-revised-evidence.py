#!/usr/bin/env python3
"""Independent revised Chapter 7 receipt audit. No network or model calls."""
import argparse
import copy
import hashlib
import importlib.util
import json
from pathlib import Path
import re
import shutil
import subprocess
import sys
import tempfile

REPO = Path(__file__).resolve().parents[2]
PREFIX = 'solutions/edition-2/main/evidence/ch07/revision-1/'
HERE = REPO / PREFIX


def read(path):
    return json.loads(path.read_text())


def sha(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()


def records(path):
    return [json.loads(line) for line in path.read_text().splitlines()]


def amplitude(path, start, end):
    result = subprocess.run(
        ['ffmpeg', '-nostdin', '-hide_banner', '-i', str(path), '-af',
         f'atrim=start={start}:end={end},volumedetect', '-f', 'null', '-'],
        capture_output=True, text=True, check=True)
    return {name: float(re.findall(name + r'_volume: ([-\d.]+) dB', result.stderr)[-1])
            for name in ('mean', 'max')}


def audit(revision):
    revision = subprocess.check_output(['git', 'rev-parse', revision + '^{commit}'],
                                       cwd=REPO, text=True).strip()
    binding = read(HERE / 'initial-binding.json')
    archived = read(REPO / 'book/edition-2/checkpoint-evidence/ch07-revision-executable-archive.json')
    assert archived['source_revision'] == binding['source_revision']
    paths = {row['role']: row['archive_path'] for row in archived['files']}
    for row in archived['files']:
        assert sha(Path(row['archive_path'])) == row['sha256'] == binding['executables'][row['role']]['sha256']
    sys.path.insert(0, str(HERE))
    spec = importlib.util.spec_from_file_location('revised_ch07_verifier', HERE / 'verify-receipts.py')
    verifier = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(verifier)
    runs = [HERE / f'consumer-{vendor}-r1' for vendor in ('anthropic', 'openai', 'gemini')]
    originals = {str(p.relative_to(HERE)): sha(p) for run in runs for p in run.rglob('*') if p.is_file()}
    assert originals
    # Verify all originals against committed blobs before exercising any replay.
    for path, expected in originals.items():
        raw = subprocess.check_output(['git', 'show', revision + ':' + PREFIX + path], cwd=REPO)
        assert hashlib.sha256(raw).hexdigest() == expected, 'original differs from checkpoint: ' + path
    controls = []
    with tempfile.TemporaryDirectory(prefix='ch07-revised-review-') as temporary:
        tmp = Path(temporary)
        positive = verifier.verify(binding, runs, paths, tmp / 'original-positive')
        assert [row['requests'] for row in positive] == [2, 2, 2]
        controls.append({'id': 'six-exact-request-replays', 'passed': True})
        copies = []
        for run in runs:
            dest = tmp / run.name
            shutil.copytree(run, dest)
            copies.append(dest)
        verifier.verify(binding, copies, paths, tmp / 'copied-positive')
        controls.append({'id': 'valid-copy-paths', 'passed': True})
        last = copies[-1] / 'launch.json'
        saved_launch = last.read_bytes()
        for name, expected in (
                ('source-hash', 'historical source mismatch'),
                ('source-set', 'incomplete historical source set'),
                ('last-launch-source', 'launch source mismatch'),
                ('last-launch-executable', 'launch executable mismatch')):
            candidate = copy.deepcopy(binding)
            launch = json.loads(saved_launch)
            if name == 'source-hash':
                candidate['sources'][next(iter(candidate['sources']))] = '0' * 64
            elif name == 'source-set':
                candidate['sources'].pop(next(iter(candidate['sources'])))
            elif name == 'last-launch-source':
                launch['source_revision'] = 'intended-review-mismatch'
            else:
                launch['executables']['consumer'] = '0' * 64
            last.write_text(json.dumps(launch))
            try:
                verifier.verify(candidate, copies, paths, tmp / name)
                raise RuntimeError('mutation accepted: ' + name)
            except AssertionError as error:
                assert expected in str(error), (name, str(error))
                assert not (tmp / name).exists(), 'derived write before refusal'
                controls.append({'id': name, 'passed': True, 'refusal': str(error)})
            finally:
                last.write_bytes(saved_launch)
        body = sorted((copies[-1] / 'requests').glob('*.json'))[-1]
        value = read(body)
        value['model'] = 'intended-review-body-mismatch'
        body.write_text(json.dumps(value))
        try:
            verifier.verify(binding, copies, paths, tmp / 'late-body-mismatch')
            raise RuntimeError('body mismatch accepted')
        except AssertionError as error:
            assert 'request reconstruction mismatch' in str(error)
            assert not (tmp / 'late-body-mismatch').exists()
            controls.append({'id': 'late-body-mismatch', 'passed': True, 'refusal': str(error)})
    providers = []
    for run in runs:
        launch = read(run / 'launch.json')
        assert launch['exit_code'] == 0 and launch['requests'] == 2
        assert len(list((run / 'requests').glob('*.json'))) == len(list((run / 'responses').glob('*.body'))) == 2
        logs = sorted((run / 'workspace').glob('*.jsonl'))
        assert len(logs) == 2
        answers = []
        for log in logs:
            events = records(log)[1:]
            assert [event['seq'] for event in events] == list(range(1, len(events) + 1))
            assert sum(e['type'] == 'turn_started' for e in events) == 1
            response = [e['response'] for e in events if e['type'] == 'response_ended']
            assert len(response) == 1
            answers.append({'log': log.name, 'stop_reason': response[0]['stop_reason'],
                            'model': response[0]['from']['model'], 'usage': response[0]['usage']})
        browser = records(run / 'browser-original.jsonl')
        assert not [row for row in browser if row['kind'] in ('pageerror', 'action_failed')]
        browser_launch = next(row for row in browser if row['kind'] == 'launch')
        assert browser_launch['source_revision'] == binding['source_revision']
        assert browser_launch['executables'] == launch['executables']
        assert sum(row['kind'] == 'idle_view_closed' for row in browser) == 2
        assert sum(row['kind'] == 'idle_view_reopened' for row in browser) == 2
        actions = [r['action']['action'] for r in browser if r['kind'] == 'action']
        assert actions.count('prompt') == 2 and {'type', 'cancel-speech', 'audio'} <= set(actions)
        action_rows = {r['sequence']: r for r in browser if r['kind'] == 'action'}
        first = {'anthropic': 12, 'openai': 8, 'gemini': 9}[launch['vendor']]
        sequence = [action_rows[first + offset] for offset in range(6)]
        assert [(r['action']['action'], r['action']['panel']) for r in sequence] == [
            ('speak', 0), ('speak', 1), ('cancel-speech', 1),
            ('speak', 1), ('cancel-speech', 0), ('cancel-speech', 1)]
        starts = [r for r in browser if r['kind'] == 'speech' and r['detail']['type'] == 'start']
        before = [r for r in starts if sequence[0]['at'] < r['at'] < sequence[4]['at']]
        after = [r for r in starts if sequence[4]['at'] < r['at'] < sequence[5]['at']]
        assert len(before) == len(after) == 1
        assert '/agent-1/' in before[0]['detail']['key'] and '/agent-2/' in after[0]['detail']['key']
        page = next(r['text'] for r in browser if r['kind'] == 'page_receipt' and r['sequence'] == first + 4)
        first_page, second_page = page.split('second Agent', 1)
        assert 'typing: 1' in first_page and 'speaking: 1' in second_page
        # Text capture may precede the asynchronous pause acknowledgement;
        # the screenshot/receipt timestamp is later. Use the applied wire fact.
        applied = []
        for row in browser:
            if row['kind'] == 'browser_received' and sequence[4]['at'] < row['at'] < sequence[5]['at']:
                observation = json.loads(row['payload']).get('observation', {})
                if observation.get('kind') == 'pause_changed':
                    applied.append(observation)
        assert any(r['agent_id'] == 'agent-1' and r['typing_clients'] == 1
                   and r['speaking_clients'] == 0 for r in applied)
        controls.append({'id': launch['vendor'] + '-queued-cancel-and-active-handoff',
                         'passed': True, 'action_sequences': [r['sequence'] for r in sequence]})
        audio = []
        for row in browser:
            if row['kind'] != 'audio':
                continue
            wav = run / Path(row['path']).name
            assert row['exit'] == 0 and row['browserPID'] == browser_launch['browser_pid']
            assert f"pid={row['browserPID']} " in row['output']
            action = action_rows[row['sequence']]
            closed = next(r for r in browser if r['kind'] == 'idle_view_closed' and r['sequence'] == row['sequence'])
            reopened = next(r for r in browser if r['kind'] == 'idle_view_reopened' and r['sequence'] == row['sequence'])
            active = [r for r in starts if action['at'] < r['at'] < closed['at']]
            assert len(active) == 1 and closed['at'] < reopened['at'] < row['at']
            assert closed['panel'] == reopened['panel'] == action['action']['cycle_idle_panel']
            assert action['action']['panel'] != closed['panel']
            assert f"/agent-{action['action']['panel'] + 1}/" in active[0]['detail']['key']
            assert not [r for r in browser if r['kind'] == 'speech' and r['detail']['type'] != 'start'
                        and active[0]['at'] < r['at'] < reopened['at']]
            silence, speech = amplitude(wav, 0, 1), amplitude(wav, 2, 6)
            assert silence['max'] <= -90, 'silent PCM control missing'
            assert speech['mean'] > -40, 'speech interval lacks signal'
            audio.append({'path': wav.name, 'sha256': sha(wav), 'browser_pid': row['browserPID'],
                          'silent_pcm_0_to_1': silence, 'speech_pcm_2_to_6': speech,
                          'speech_start': active[0]['at'], 'idle_close': closed['at'], 'idle_reopen': reopened['at']})
        assert len(audio) == 2
        second_prompt = next(r for r in action_rows.values() if r['action']['action'] == 'prompt' and r['action']['panel'] == 1)
        first_remount = next(r for r in browser if r['kind'] == 'idle_view_reopened' and r['panel'] == 1)
        assert first_remount['at'] < second_prompt['at'], 'second prompt did not exercise replacement view'
        providers.append({'provider': launch['vendor'], 'requests': 2, 'prompts': 2,
                          'answers': answers, 'audio': audio, 'browser_errors': 0})
    settings = read(Path.home() / '.cr/settings.json')
    secrets = [settings[key].encode() for key in ('directClaudeAPIKey', 'directOpenAIAPIKey', 'directGeminiAPIKey') if settings.get(key)]
    for path in originals:
        assert not any(value in (HERE / path).read_bytes() for value in secrets), 'credential in original receipt'
    assert originals == {path: sha(HERE / path) for path in originals}, 'original changed during audit'
    return {'passed': True, 'evidence_commit': revision, 'source_revision': binding['source_revision'],
            'controls': controls, 'providers': providers, 'original_files': originals,
            'credential_scan_files': len(originals), 'review_script_sha256': sha(Path(__file__)),
            'limits': 'Revised public two-Agent browser only. No paid calls. Signal/PID evidence is not transcription or human listening. Initial unaffected demonstrations retain their original identity.'}


if __name__ == '__main__':
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('evidence_commit')
    print(json.dumps(audit(parser.parse_args().evidence_commit), indent=2))
