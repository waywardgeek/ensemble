#!/usr/bin/env python3
"""Independent initial Chapter7 evidence audit; replay/local analysis only, no HTTP."""
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
HERE = REPO / 'solutions/edition-2/main/evidence/ch07'
INITIAL = '8cc87f2'
VENDORS = ('anthropic', 'openai', 'gemini')


def read(path):
    return json.loads(path.read_text())


def sha(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()


def records(path):
    return [json.loads(line) for line in path.read_text().splitlines()]


def audit():
    index = read(HERE / 'live-receipts.json')
    binding = read(HERE / 'initial-binding.json')
    archived = read(REPO / 'book/edition-2/checkpoint-evidence/ch07-executable-archive.json')
    paths = {r['role']: r['archive_path'] for r in archived['files']}
    sys.path.insert(0, str(HERE))
    spec = importlib.util.spec_from_file_location('reviewed_ch07_verifier', HERE / 'verify-receipts.py')
    verifier = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(verifier)
    runs = [HERE / f'{mode}-{vendor}-r1' for vendor in VENDORS for mode in ('browser', 'plain', 'consumer')]
    originals = {str(p.relative_to(HERE)): sha(p) for run in runs for p in run.rglob('*') if p.is_file()}
    initial = subprocess.check_output(['git', 'rev-parse', INITIAL], cwd=REPO, text=True).strip()
    # Bind every original to the actual initial Git checkpoint, not a mutable index.
    prefix = 'solutions/edition-2/main/evidence/ch07/'
    for path, expected in originals.items():
        raw = subprocess.check_output(['git', 'show', initial + ':' + prefix + path], cwd=REPO)
        assert hashlib.sha256(raw).hexdigest() == expected, 'original differs from checkpoint: ' + path
    controls = []
    with tempfile.TemporaryDirectory(prefix='ch07-independent-receipts-') as temporary:
        tmp = Path(temporary)
        positive = verifier.verify(binding, runs, paths, tmp / 'original-positive')
        assert sum(r['requests'] for r in positive) == 44
        controls.append({'id': 'valid-original-nine-runs', 'passed': True, 'requests': 44})
        copies = []
        for source in runs:
            dest = tmp / source.name
            shutil.copytree(source, dest)
            copies.append(dest)
        verifier.verify(binding, copies, paths, tmp / 'valid-copy-paths')
        controls.append({'id': 'valid-copied-paths-before-mutations', 'passed': True})
        last = copies[-1] / 'launch.json'
        original_launch = last.read_bytes()
        mutations = [
            ('source-hash', 'historical source mismatch'),
            ('source-set', 'incomplete historical source set'),
            ('empty-source', 'empty source set'),
            ('missing-executable', 'incomplete executable identities'),
            *[(name + '-hash', 'interpreter mismatch' if name == 'interpreter' else 'executable mismatch')
              for name in ('cli', 'gui', 'consumer', 'capture', 'chrome', 'node', 'interpreter', 'recorder')],
            ('support-hash', 'historical support mismatch'),
            ('support-set', 'incomplete support identities'),
            ('browser-hash', 'browser dependency mismatch'),
            ('browser-set', 'incomplete browser dependency set'),
            ('last-launch-source', 'launch source mismatch'),
            ('last-launch-binary', 'launch executable mismatch'),
            ('last-launch-support', 'launch support mismatch'),
            ('last-launch-browser', 'launch browser tools mismatch'),
            ('last-launch-role', 'unknown launch executable'),
        ]
        for name, reason in mutations:
            candidate = copy.deepcopy(binding)
            launch = json.loads(original_launch)
            if name == 'source-hash': candidate['sources'][next(iter(candidate['sources']))] = '0' * 64
            elif name == 'source-set': candidate['sources'].pop(next(iter(candidate['sources'])))
            elif name == 'empty-source': candidate['sources'] = {}
            elif name == 'missing-executable': candidate['executables'].pop('gui')
            elif name.endswith('-hash') and name[:-5] in candidate['executables']:
                candidate['executables'][name[:-5]]['sha256'] = '0' * 64
            elif name == 'support-hash': candidate['support']['evidence.py'] = '0' * 64
            elif name == 'support-set': candidate['support'].pop('evidence.py')
            elif name == 'browser-hash': candidate['browser_tools'][next(iter(candidate['browser_tools']))] = '0' * 64
            elif name == 'browser-set': candidate['browser_tools'].pop(next(iter(candidate['browser_tools'])))
            elif name == 'last-launch-source': launch['source_revision'] = 'wrong'
            elif name == 'last-launch-binary': launch['executables']['gui'] = '0' * 64
            elif name == 'last-launch-support': launch['support']['evidence.py'] = '0' * 64
            elif name == 'last-launch-browser': launch['browser_tools'].pop(next(iter(launch['browser_tools'])))
            elif name == 'last-launch-role': launch['launched_executable'] = 'unknown'
            else: raise RuntimeError(name)
            last.write_text(json.dumps(launch))
            output = tmp / name
            try:
                verifier.verify(candidate, copies, paths, output)
                raise RuntimeError('mutation accepted: ' + name)
            except AssertionError as error:
                assert reason in str(error), (name, str(error))
                assert not output.exists(), 'wrote before refusing ' + name
                controls.append({'id': name, 'passed': True, 'refusal': str(error)})
            finally:
                last.write_bytes(original_launch)
        body = sorted((copies[-1] / 'requests').glob('*.json'))[-1]
        original_body = body.read_bytes()
        altered = read(body); altered['model'] = 'intended-review-body-mismatch'
        body.write_text(json.dumps(altered))
        try:
            verifier.verify(binding, copies, paths, tmp / 'late-body-mismatch')
            raise RuntimeError('request mismatch accepted')
        except AssertionError as error:
            assert 'request reconstruction mismatch' in str(error)
            assert not (tmp / 'late-body-mismatch').exists()
            controls.append({'id': 'late-request-body-mismatch', 'passed': True, 'refusal': str(error)})
        finally:
            body.write_bytes(original_body)
    providers = []
    for vendor in VENDORS:
        expected = index['providers'][vendor]
        requests = prompts = 0
        for mode in ('browser', 'plain', 'consumer'):
            run = HERE / f'{mode}-{vendor}-r1'
            launch = read(run / 'launch.json')
            assert launch['exit_code'] == 0
            bodies = list((run / 'requests').glob('*.json'))
            assert len(bodies) == launch['requests'] == len(list((run / 'responses').glob('*.body')))
            requests += len(bodies)
            logs = list(run.glob('*.log')) + list((run / 'workspace').glob('*.jsonl'))
            assert len(logs) == (2 if mode == 'consumer' else 1)
            for log in logs:
                events = records(log)[1:]
                assert [e['seq'] for e in events] == list(range(1, len(events) + 1))
                prompts += sum(e['type'] == 'turn_started' for e in events)
            row = next(row for row in expected['launches'] if row['run'] == run.name)
            assert sha(run / 'terminal.txt') == row['terminal_sha256']
        assert requests == expected['forwarded_http_requests'] <= 32
        assert prompts == expected['admitted_prompts'] <= 12
        run = HERE / f'browser-{vendor}-r1'
        browser = records(run / 'browser-original.jsonl')
        assert not [r for r in browser if r['kind'] in ('pageerror', 'action_failed')]
        actions = [r['action']['action'] for r in browser if r['kind'] == 'action']
        assert {'prompt', 'hint', 'type', 'open', 'reload', 'close', 'auto-speech', 'cancel-speech', 'audio'} <= set(actions)
        assert (run / 'workspace/browser-note.txt').read_bytes() == b'BROWSER-CH07\n'
        events = records(run / 'session.log')[1:]
        killed = [e for e in events if e['type'] == 'job_killed']
        assert killed and killed[-1]['job']['status'] == 'killed'
        hints = [e['hint']['text'] for e in events if e['type'] == 'hint_received']
        assert hints and any(hints[0] in p.read_text() for p in (run / 'requests').glob('*.json'))
        assert any(e['type'] == 'turn_ended' and e['turn'].get('outcome') == 'interrupted' for e in events)
        assert b'BROWSER-STILL-LIVE' in (run / 'session.log').read_bytes()
        audio = []
        launch_browser = next(r for r in browser if r['kind'] == 'launch')
        for row in read(run / 'audio-audit.json')['audio']:
            wav = run / row['raw_audio']
            assert sha(wav) == row['sha256']
            assert row['capture']['browserPID'] == launch_browser['browser_pid']
            assert f"pid={launch_browser['browser_pid']} " in row['capture']['output']
            assert row['capture']['exit'] == 0
            measures = {}
            for name, interval in row['metrics'].items():
                cmd = ['ffmpeg', '-nostdin', '-hide_banner', '-i', str(wav), '-af',
                       f"atrim=start={interval['start_seconds']}:end={interval['end_seconds']},volumedetect", '-f', 'null', '-']
                result = subprocess.run(cmd, capture_output=True, text=True, check=True)
                mean = float(re.findall(r'mean_volume: ([-\d.]+) dB', result.stderr)[-1])
                peak = float(re.findall(r'max_volume: ([-\d.]+) dB', result.stderr)[-1])
                assert mean == interval['mean_db'] and peak == interval['peak_db']
                measures[name] = {'mean_db': mean, 'peak_db': peak}
            assert measures['silent_prefix']['peak_db'] <= -90
            assert measures['speech']['mean_db'] > -40
            audio.append({'path': row['raw_audio'], 'sha256': row['sha256'], 'metrics': measures})
        providers.append({'provider': vendor, 'requests': requests, 'prompts': prompts, 'audio': audio,
                          'job_killed_events': len(killed), 'hint_delivery': True, 'browser_errors': 0})
    # Only configured credential values are read; never serialize their values.
    settings = read(Path.home() / '.cr/settings.json')
    secrets = [settings[field].encode() for field in ('directClaudeAPIKey', 'directOpenAIAPIKey', 'directGeminiAPIKey') if settings.get(field)]
    scanned = 0
    for path in originals:
        data = (HERE / path).read_bytes()
        assert not any(key in data for key in secrets), 'credential found in retained original'
        scanned += 1
    assert originals == {path: sha(HERE / path) for path in originals}, 'original receipts changed during audit'
    return {'passed': True, 'initial_checkpoint': initial, 'source_revision': binding['source_revision'],
            'runtime_source_revision': index['runtime_source_revision'], 'controls': controls,
            'providers': providers, 'original_files': originals, 'credential_scan_files': scanned,
            'review_script_sha256': sha(Path(__file__)),
            'limits': 'Initial runtime only. No HTTP or paid calls. Audio amplitude and PID evidence, not transcription or human listening. Historical comparison and later revisions remain separate.'}


if __name__ == '__main__':
    print(json.dumps(audit(), indent=2))
