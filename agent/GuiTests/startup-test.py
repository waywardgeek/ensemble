#!/usr/bin/env python3
"""Run the actual foreground launcher using only generated/private inputs."""
import json
import os
from pathlib import Path
import signal
import socket
import subprocess
import time
import urllib.request

root = Path.cwd()
folder = root / '.crossover/startup'
folder.mkdir(parents=True, exist_ok=True)
evidence = root / '.crossover/evidence'
for child in ['home', 'tmp', 'cli', 'workspace', 'data']:
    (folder / child).mkdir(exist_ok=True)
sdk = '/Users/bill/.local/share/ensemble-dotnet10'
env = {'PATH':'/usr/bin:/bin', 'HOME':str(folder/'home'), 'TMPDIR':str(folder/'tmp'),
       'DOTNET_ROOT':sdk, 'DOTNET_CLI_HOME':str(folder/'cli'),
       'DOTNET_CLI_TELEMETRY_OPTOUT':'1', 'DOTNET_SKIP_FIRST_TIME_EXPERIENCE':'1',
       'DOTNET_GENERATE_ASPNET_CERTIFICATE':'false', 'LLM_API_KEY':'generated-startup-fixture-key'}
base = [str(root/'start'), '--workspace', str(folder/'workspace'), '--data-dir', str(folder/'data'),
        '--vendor','anthropic','--model','claude-3-haiku-20240307','--base-url','http://127.0.0.1:18239', '--no-save']
# This cwd is intentionally not the source/build directory.
cwd = str(folder)

def failure(name, args, expected):
    result = subprocess.run(base+args, cwd=cwd, env=env, capture_output=True, text=True, timeout=20)
    text = result.stdout + result.stderr
    (evidence/(name+'.log')).write_text(text)
    assert result.returncode == 1, (name, result.returncode)
    assert expected in text, (name, text)
    assert env['LLM_API_KEY'] not in text and 'secret-canary' not in text
    print('PASS startup '+name, flush=True)

failure('malformed-port',['--port','secret-canary'], 'Invalid launch configuration')
failure('malformed-flag',['--secret-canary'], 'Invalid launch configuration')
listener = socket.socket()
listener.bind(('127.0.0.1',18230))
listener.listen()
try:
    failure('occupied-port',['--port','18230'], 'Loopback port is already in use')
    assert listener.getsockname()[1] == 18230
finally:
    listener.close()
# A bad explicit diagnostic path tests partial setup cleanup without disclosing
# its private filesystem value or the fixture provider key.
bad = folder/'not-a-directory'
bad.write_text('fixture')
failure('partial-startup',['--data-dir',str(bad/'secret-canary'),'--port','18231'], 'GUI failed')
with (evidence/'foreground-start.log').open('w') as log:
    process = subprocess.Popen(base+['--port','18231'],cwd=cwd,env=env,stdout=log,stderr=log,start_new_session=True)
    try:
        deadline = time.monotonic()+15
        while time.monotonic()<deadline:
            if process.poll() is not None:
                raise AssertionError('foreground host exited before ready')
            try:
                with urllib.request.urlopen('http://127.0.0.1:18231/', timeout=.5) as response:
                    html = response.read().decode()
                assert 'vendor/bootstrap-5.3.8.min.css' in html
                with urllib.request.urlopen('http://127.0.0.1:18231/vendor/bootstrap-5.3.8.min.css',timeout=.5) as response:
                    assert response.read() == (root/'gui/wwwroot/vendor/bootstrap-5.3.8.min.css').read_bytes()
                break
            except OSError:
                time.sleep(.1)
        else:
            raise AssertionError('foreground host never ready')
        process.send_signal(signal.SIGINT)
        assert process.wait(timeout=10) == 0
        print('PASS startup foreground from private cwd; local embedded vendor; SIGINT exit 0',flush=True)
    finally:
        if process.poll() is None:
            os.killpg(process.pid,signal.SIGKILL)
            process.wait(timeout=3)
