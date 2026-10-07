"""Supplemental all-provider public-library jobs observations; no chat claim."""
import datetime
import hashlib
import json
import os
from pathlib import Path
import subprocess
import sys

here=Path(__file__).resolve().parent
vendor,model=sys.argv[1:]
run=here/('public-live-'+vendor)
run.mkdir()
env=os.environ.copy()
for key in list(env):
    if key.startswith(('LLM_','ANTHROPIC_','OPENAI_','GEMINI_')):del env[key]
field={'anthropic':'directClaudeAPIKey','openai':'directOpenAIAPIKey','gemini':'directGeminiAPIKey'}[vendor]
key=json.loads((Path.home()/'.cr/settings.json').read_text())[field]
env.update(LLM_VENDOR=vendor,LLM_MODEL=model,LLM_RESOLVED_MODEL=model.removeprefix('models/'),LLM_API_KEY=key,ENSEMBLE_LIVE_WORKSPACE=str(run/'workspace'))
binary=Path('/tmp/ensemble-ed2-ch04-public')
command=[str(binary),'--live']
receipt={'start':datetime.datetime.now(datetime.timezone.utc).isoformat(),'command':command,'vendor':vendor,'requested_model':model,'source_revision':subprocess.check_output(['git','rev-parse','HEAD'],text=True).strip(),'binary_sha256':hashlib.sha256(binary.read_bytes()).hexdigest(),'kind':'supplemental real-model public library; main human PTY receipts remain separate'}
with (run/'stdout.json').open('w') as stdout,(run/'stderr.txt').open('w') as stderr:
    result=subprocess.run(command,env=env,stdout=stdout,stderr=stderr,timeout=120)
receipt.update(exit_code=result.returncode,end=datetime.datetime.now(datetime.timezone.utc).isoformat())
(run/'launch.json').write_text(json.dumps(receipt,indent=2)+'\n')
print(vendor,'exit',result.returncode)
print((run/'stdout.json').read_text())
if result.returncode:print((run/'stderr.txt').read_text())
raise SystemExit(result.returncode)
