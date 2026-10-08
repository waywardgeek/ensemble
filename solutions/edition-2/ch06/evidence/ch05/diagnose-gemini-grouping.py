"""Bounded provider investigation; no returned tool call is ever executed."""
import datetime
import json
from pathlib import Path
import urllib.error
import urllib.request

from evidence import HERE, digest, preflight
from importlib.util import module_from_spec, spec_from_file_location

binding = json.loads((HERE / 'initial-binding.json').read_text())
paths = {n: d['path'] for n, d in binding['executables'].items()}
preflight(binding, paths)
original = HERE / 'controls-gemini/requests/002.json'
launch = json.loads((original.parent.parent / 'launch.json').read_text())
assert launch['source_revision'] == binding['source_revision']
assert launch['executables'] == {n: d['sha256'] for n, d in binding['executables'].items()}
assert launch['support'] == binding['support']
assert launch['command'] == [launch['executable_paths']['cli'], 'chat']
baseline = original.read_bytes()
body = json.loads(baseline)
parts = body['contents'][-1]['parts']
assert body['contents'][-1]['role'] == 'user'
assert len(parts) == 2 and 'functionResponse' in parts[0] and 'text' in parts[1]
body['contents'][-1]['parts'] = parts[:1]
body['contents'].append({'role': 'user', 'parts': parts[1:]})
split = json.dumps(body, separators=(',', ':'), sort_keys=True).encode()
run = HERE / 'gemini-grouping-diagnostic'
run.mkdir()
spec = spec_from_file_location('safe_launcher', HERE / 'terminal-run.py')
launcher = module_from_spec(spec)
spec.loader.exec_module(launcher)
key = launcher.credential('gemini')
model = launch['requested_model'].removeprefix('models/')
receipt = {'source_revision': binding['source_revision'], 'source_request': str(original.relative_to(HERE)),
           'original_request_sha256': digest(baseline), 'model': model,
           'executables': binding['executables'], 'support': binding['support'],
           'diagnostic_script_sha256': digest(Path(__file__).read_bytes()),
           'at': datetime.datetime.now(datetime.timezone.utc).isoformat(),
           'scope': 'Two calls only. Same request/model/config, last user content split at functionResponse/text boundary. No returned tools executed. Not live-interface acceptance.', 'results': []}
for name, data in [('merged', baseline), ('split', split)]:
    assert key.encode() not in data
    (run / (name + '-request.json')).write_bytes(data)
    request = urllib.request.Request('https://generativelanguage.googleapis.com/v1beta/models/' + model + ':generateContent',
                                     data=data, headers={**launcher.headers('gemini', key), 'content-type': 'application/json'}, method='POST')
    try:
        with urllib.request.urlopen(request, timeout=60) as response:
            status, raw = response.status, response.read()
    except urllib.error.HTTPError as error:
        status, raw = error.code, error.read()
    safe = raw.replace(key.encode(), b'[redacted credential]')
    (run / (name + '-response.json')).write_bytes(safe)
    response = json.loads(safe)
    row = {'variant': name, 'status': status, 'request_sha256': digest(data), 'response_sha256': digest(safe),
           'candidates': response.get('candidates'), 'usage': response.get('usageMetadata')}
    receipt['results'].append(row)
    (run / 'receipt.json').write_text(json.dumps(receipt, indent=2) + '\n')
    print(json.dumps(row))
