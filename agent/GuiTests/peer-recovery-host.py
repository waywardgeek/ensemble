"""Installed-style host regression: switched root, legacy peers, GUI recovery.
Uses synthetic credentials, isolated homes and no inference. Arguments: dotnet Gui.dll output.
"""
import importlib.util
import json
import os
from pathlib import Path
import signal
import socket
import subprocess
import sys
import time
import urllib.request

spec = importlib.util.spec_from_file_location('model_host', Path(__file__).with_name('model-host.py'))
helper = importlib.util.module_from_spec(spec)
spec.loader.exec_module(helper)
require = helper.require


def main():
    dotnet, gui, output = sys.argv[1:]
    root = Path(output).resolve()
    root.mkdir()
    for name in ('home/.en/oauth', 'tmp', 'cli', 'skills/primary', 'data/agents/Cadence'):
        (root / name).mkdir(parents=True, exist_ok=True)
    (root / 'skills/primary/SKILL.md').write_text('---\nname: primary\ndescription: Offline test\ntype: primary\n---\nNo inference.\n')
    global_path = root / 'home/.en/settings.json'
    global_path.write_text(json.dumps({'openai_api_key': 'fake-openai', 'anthropic_api_key': 'fake-anthropic'}))
    codex = root / 'codex.json'
    codex.write_text(json.dumps({'auth_mode': 'chatgpt', 'tokens': {
        'access_token': 'fake-codex', 'account_id': 'fake-account'}}))
    oauth = root / 'home/.en/oauth'
    oauth.chmod(0o700)
    credentials = oauth / 'credentials.json'
    credentials.write_text(json.dumps({'version': 2, 'host_id': 'fixture', 'selected_plan_account': 'test', 'accounts': [{
        'label': 'test', 'issuer': 'https://auth.openai.com', 'subject': 'test-subject', 'client_id': 'fixture-client',
        'email': None, 'access_token': 'fake-access', 'refresh_token': 'fake-refresh', 'id_token': 'fake-id',
        'expires_at': '2099-01-01T00:00:00.0000000+00:00', 'scope': 'chatgpt.tokens.use.direct'}]}))
    credentials.chmod(0o600)
    original_credentials = credentials.read_bytes()
    memory = {name: {'budget': 1000, 'disabled': True} for name in ('conversation', 'soul', 'memory', 'archive64', 'archive8', 'session')}
    (root / 'data/agents/Cadence/settings.json').write_text(json.dumps({'version': 1, 'memory': memory,
        'selected_offering': {'vendor': 'anthropic', 'endpoint': 'api', 'model': 'claude-opus-5'}}))
    registry = root / 'data/ensemble.json'
    registry.write_text(json.dumps([{'name': name, 'model': 'gpt-6-astra', 'skill': 'primary',
                                   'created': '2026-01-01T00:00:00Z'} for name in ('Hobson', 'Resume')]))
    original_registry = registry.read_bytes()
    with socket.socket() as reservation:
        reservation.bind(('127.0.0.1', 0))
        port = reservation.getsockname()[1]
    origin = f'http://127.0.0.1:{port}'
    env = {'PATH': '/usr/bin:/bin', 'HOME': str(root / 'home'), 'TMPDIR': str(root / 'tmp'),
           'DOTNET_CLI_HOME': str(root / 'cli'), 'DOTNET_GENERATE_ASPNET_CERTIFICATE': 'false',
           'DOTNET_CLI_TELEMETRY_OPTOUT': '1'}
    command = [dotnet, gui, '--workspace', str(root), '--data-dir', str(root / 'data'),
               '--agent-name', 'Cadence', '--port', str(port), '--billing', 'codex', '--model', 'gpt-6-astra',
               '--codex-auth-file', str(codex), '--credentials-dir', str(oauth), '--skills-dir', str(root / 'skills'),
               '--primary-skill', 'primary']

    def request(path, body=None):
        payload = None if body is None else json.dumps(body).encode()
        req = urllib.request.Request(origin + path, payload, headers={
            'X-Ensemble-Peers': '1', 'Origin': origin, 'Content-Type': 'application/json'})
        with urllib.request.urlopen(req, timeout=10) as response:
            return json.load(response)

    for attempt in range(3):
        with (root / f'host-{attempt}.log').open('wb') as log:
            process = subprocess.Popen(command, cwd=root, env=env, stdout=log, stderr=log, start_new_session=True)
            try:
                deadline = time.monotonic() + 15
                while True:
                    require(process.poll() is None, 'host exited before recovery UI')
                    try:
                        choices = request('/peers/choices')
                        break
                    except OSError:
                        require(time.monotonic() < deadline, 'host startup deadline')
                        time.sleep(.03)
                frame = helper.usage_frame(port)
                require(frame['selected_offering']['vendor'] == 'anthropic', 'root lost saved Opus selection')
                funding = {row['label'].split(' · ')[0] for row in choices['funding_choices'] if row['vendor_id'] == 'openai'}
                require(funding == {'API key', 'Codex OAuth', 'ChatGPT OAuth'}, 'missing independent OpenAI funding routes')
                require(len([row for row in choices['preferred_models'] if row['preference'] ==
                             {'vendor': 'openai', 'model': 'gpt-6-astra'}]) == 1, 'Astra duplicated by funding')
                if attempt < 2:
                    require(len(request('/peers/saved')) == 2, 'failed fleet is not recoverable')
                    if attempt == 0:
                        selected = next(row for row in choices['funding_choices'] if row['label'] == 'Codex OAuth')
                        request('/models/funding', {'offering': selected['offering']})
                        request('/peers/restore', {'name': 'Hobson', 'preference': {'vendor': 'openai', 'model': 'gpt-6-astra'}})
                    else:
                        request('/peers/restore', {'name': 'Hobson'})
                        request('/peers/restore', {'name': 'Resume', 'preference': {'vendor': 'openai', 'model': 'gpt-6-astra'}})
                    for name in ('Hobson',) if attempt == 0 else ('Hobson', 'Resume'):
                        saved = json.loads((root / f'data/agents/{name}/settings.json').read_text())
                        require(saved['selected_offering'] == {'vendor': 'openai', 'endpoint': 'codex', 'model': 'gpt-6-astra'},
                                'peer changed to root/API funding')
                else:
                    require(request('/peers/saved') == [], 'repaired fleet did not restore automatically')
                require(registry.read_bytes() == original_registry, 'recovery rewrote old registry')
                require(credentials.read_bytes() == original_credentials, 'startup refreshed/rewrote OAuth grant')
            finally:
                if process.poll() is None:
                    os.killpg(process.pid, signal.SIGTERM)
                process.wait(timeout=15)
                require(process.returncode == 0, 'host did not join cleanly')
    logs = list((root / 'data').rglob('events.jsonl'))
    require(bool(logs), 'recovery event logs missing')
    for log in logs:
        helper.require_no_inference(log)
    print('PASS actual host: GUI survives failed fleet; three funding sources; explicit per-peer recovery; third restart succeeds; no inference')


if __name__ == '__main__':
    main()
