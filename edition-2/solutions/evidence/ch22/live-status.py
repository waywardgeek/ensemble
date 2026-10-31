"""Bounded two-model browser status demonstration using the dedicated profile."""
import hashlib
import json
import os
import pathlib
import socket
import subprocess
import tempfile
import time
import urllib.request

repo = pathlib.Path.cwd()
evidence = repo / 'solutions/edition-2/evidence/ch22'
build = json.loads((evidence / 'build.json').read_text())
auth = pathlib.Path('/Users/bill/ensemble-ch19-signin-01v0rmp3/auth')
private = []

def remember_private():
    # Read only this authorized profile; retain values solely for redaction.
    value = json.loads((auth / 'auth.json').read_text())
    for key in ('host_id', 'active', 'pending'):
        if value.get(key): private.append(value[key].encode())
    for profile in value['profiles'].values():
        for key in ('access_token', 'refresh_token', 'id_token', 'email', 'subject', 'client_id'):
            if profile.get(key): private.append(profile[key].encode())

def save(name, raw):
    for secret in set(private): raw = raw.replace(secret, b'[REDACTED]')
    assert not any(secret in raw for secret in private)
    (evidence / name).write_bytes(raw)

# Browser and build processes receive no provider credentials. Only the actual
# Ensemble child receives access to the dedicated profile; no API fallback key.
base_env = {k: os.environ[k] for k in ('PATH','HOME','USER','TMPDIR','LANG') if k in os.environ}
env = dict(base_env, EN_HOME=str(auth), LLM_VENDOR='openai', LLM_MODEL='gpt-6-astra',
           LLM_BASE_URL='https://api.openai.com', OPENAI_BILLING_MODE='chatgpt',
           OPENAI_REASONING_EFFORT='high', OPENAI_REASONING_SUMMARY='detailed',
           EN_RECALL_DISABLED='1', EN_PRIMARY_SKILL='ensemble',
           EN_SKILLS_DIR=str(repo/'solutions/edition-2/main/skills'))
remember_private()
for kind in ('cli','gui'):
    assert hashlib.sha256(pathlib.Path(build[kind]['binary']).read_bytes()).hexdigest() == build[kind]['binary_sha256']
# Discover account-available identifiers through Ensemble before choosing both.
catalog = subprocess.run([build['cli']['binary'],'auth','models'],env=env,capture_output=True,timeout=30)
remember_private()
save('model-catalog.txt',catalog.stdout+catalog.stderr)
assert catalog.returncode == 0
assert all(model in catalog.stdout.decode() for model in ('gpt-6-astra','gpt-6.1-sol'))
work = pathlib.Path(tempfile.mkdtemp(prefix='ensemble-ch22-status-'))
with socket.socket() as port_socket:
    port_socket.bind(('127.0.0.1',0)); port=port_socket.getsockname()[1]
start = time.monotonic()
with (work/'stdout.txt').open('wb') as out, (work/'stderr.txt').open('wb') as err:
    child = subprocess.Popen([build['gui']['binary'],'--port',str(port),'--gui-dir',str(repo/'solutions/edition-2/main/gui/web/gui')],cwd=work,env=env,stdin=subprocess.PIPE,stdout=out,stderr=err)
    browser = None
    try:
        ready = False
        for _ in range(100):
            try:
                with urllib.request.urlopen(f'http://127.0.0.1:{port}',timeout=.2) as response: ready=response.status == 200
            except OSError: pass
            if ready or child.poll() is not None: break
            time.sleep(.05)
        assert ready, 'GUI failed readiness'
        browser = subprocess.run(['node',str(evidence/'live-browser.mjs'),f'http://127.0.0.1:{port}/'],env=base_env,capture_output=True,timeout=210)
    finally:
        child.stdin.close()
        try: child.wait(timeout=10)
        except subprocess.TimeoutExpired: child.kill();child.wait()
remember_private()
for name in ('stdout.txt','stderr.txt','api.log','event.log','save.json','debug.log','gui.log'):
    if (work/name).exists(): save('live-'+name,(work/name).read_bytes())
if browser:
    save('live-browser.json',browser.stdout)
    save('live-browser-errors.txt',browser.stderr)
result=dict(source_commit=build['source_commit'], main_tree=build['main_tree'],binary=build['gui'],
            workspace=str(work),browser_exit=browser.returncode if browser else None,gui_exit=child.returncode,
            elapsed_seconds=round(time.monotonic()-start,2), models=['gpt-6-astra','gpt-6.1-sol'],reasoning_effort='high',
            billing='explicit chatgpt',model_api_key_configured=False,profile_retained=True,logout=False,
            endpoint_from_config_and_source='https://api.openai.com/v1/responses',endpoint_independently_captured=False)
(evidence/'live-result.json').write_text(json.dumps(result,indent=2)+'\n')
print(json.dumps(result,indent=2))
