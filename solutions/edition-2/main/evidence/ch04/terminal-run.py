"""Credential-safe launcher. The coder observes and supplies each PTY turn."""
import datetime
import hashlib
import json
import os
from pathlib import Path
import subprocess
import sys
import urllib.request

HERE = Path(__file__).resolve().parent
MODULE = HERE.parents[1]
FIELDS = {'anthropic': 'directClaudeAPIKey', 'openai': 'directOpenAIAPIKey', 'gemini': 'directGeminiAPIKey'}

def credential(vendor):
    return json.loads((Path.home() / '.cr/settings.json').read_text())[FIELDS[vendor]]

def discover():
    urls = {'anthropic':'https://api.anthropic.com/v1/models?limit=100','openai':'https://api.openai.com/v1/models','gemini':'https://generativelanguage.googleapis.com/v1beta/models?pageSize=1000'}
    result = {'at':datetime.datetime.now(datetime.timezone.utc).isoformat(),'providers':{}}
    for vendor,url in urls.items():
        key = credential(vendor)
        headers = {'anthropic':{'x-api-key':key,'anthropic-version':'2023-06-01'},'openai':{'Authorization':'Bearer '+key},'gemini':{'x-goog-api-key':key}}[vendor]
        with urllib.request.urlopen(urllib.request.Request(url,headers=headers),timeout=30) as response:
            page = json.load(response)
        result['providers'][vendor] = [m.get('id',m.get('name')) for m in page.get('data',page.get('models',[])) if vendor!='gemini' or 'generateContent' in m.get('supportedGenerationMethods',[])]
    (HERE/'discovery.json').write_text(json.dumps(result,indent=2)+'\n')
    print(json.dumps(result,indent=2))

def launch(vendor,model,name=None,binary_path="/tmp/ensemble-ed2-ch04-live"):
    binary = Path(binary_path)
    run = HERE / (name or ('live-'+vendor))
    run.mkdir()
    workspace = run/'workspace'
    workspace.mkdir()
    (workspace/'subdir').mkdir()
    (workspace/'notes.txt').write_text('alpha one\nbeta two\ngamma three\ndelta four\n')
    (workspace/'interactive.py').write_text('import time\nprint("READY>",flush=True)\nwhile True:\n    try: value=input()\n    except EOFError: break\n    time.sleep(0.2)\n    print("REPLY:"+value+" READY>",flush=True)\n')
    (workspace/'debuggee.go').write_text('package main\nimport "fmt"\nfunc main() {\n value := 42\n fmt.Println(value)\n}\n')
    env = os.environ.copy()
    env["PATH"] = str(Path.home()/"go/bin") + ":/usr/local/go/bin:" + env.get("PATH", "")
    for key in list(env):
        if key.startswith(('LLM_','ANTHROPIC_','OPENAI_','GEMINI_')): del env[key]
    env.update(LLM_VENDOR=vendor,LLM_MODEL=model,LLM_RESOLVED_MODEL=model.removeprefix('models/'),LLM_API_KEY=credential(vendor),CH02_LOG=str(run/'session.log'))
    command = [str(binary),'chat']
    receipt = {'start':datetime.datetime.now(datetime.timezone.utc).isoformat(),'actor':'Codex student coder, not Bill','vendor':vendor,'requested_model':model,'command':command,'workspace':str(workspace),'source_revision':subprocess.check_output(['git','rev-parse','HEAD'],cwd=MODULE,text=True).strip(),'binary_sha256':hashlib.sha256(binary.read_bytes()).hexdigest(),'dlv_version':subprocess.check_output([str(Path.home()/'go/bin/dlv'),'version'],text=True),'transport':'actual execution-tool PTY and macOS script terminal'}
    (run/'launch.json').write_text(json.dumps(receipt,indent=2)+'\n')
    code = subprocess.call(['/usr/bin/script','-q',str(run/'terminal.txt'),*command],cwd=workspace,env=env)
    receipt.update(exit_code=code,end=datetime.datetime.now(datetime.timezone.utc).isoformat())
    (run/'launch.json').write_text(json.dumps(receipt,indent=2)+'\n')
    raise SystemExit(code)

if sys.argv[1:]==['discover']: discover()
else: launch(*sys.argv[1:])
