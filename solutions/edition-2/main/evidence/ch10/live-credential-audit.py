"""Safe phase6 evidence/owned Git-diff audit. No network or secret output."""
import json
from pathlib import Path
import subprocess
import sys

HERE = Path(__file__).resolve().parent
sys.path.insert(0, str(HERE/'support'))
from identity import preflight
from provider import load_key, Redactor

def main():
    binding=json.loads((HERE/'provider-prep-binding-final.json').read_text())
    repo=HERE.parents[4]
    preflight(binding,repo)
    # Same reviewed loader, selected fields only. Neither keys nor derivatives
    # leave this process; report only aggregate pass/fail and scope counts.
    tokens=Redactor([load_key(v) for v in ('anthropic','openai','gemini')]).tokens
    files=sorted(p for p in (HERE/'live-20261008').rglob('*') if p.is_file())
    owned=[HERE/n for n in ('live-command.py','live-observations.py',
                            'live-credential-audit.py','live-handback.md',
                            'student-review.md','implementation-status.md')]
    files.extend(p for p in owned if p.is_file())
    matched=False
    for p in files:
        if p.is_symlink():raise ValueError('symlink')
        with p.open('rb') as f:
            tail=b''
            while True:
                chunk=f.read(1024*1024)
                if not chunk:break
                data=tail+chunk
                matched |= any(t in data for t in tokens)
                tail=data[-max(map(len,tokens)):]
    paths=[str(p.relative_to(repo)) for p in owned]+[str((HERE/'live-20261008').relative_to(repo))]
    patch=subprocess.check_output(['git','diff','--cached','--binary','--',*paths],cwd=repo)
    matched |= any(t in patch for t in tokens)
    receipt={'source_revision':binding['source_revision'],'support_revision':binding['support_revision'],
             'passed':not matched,'files_scanned':len(files),'owned_staged_patch_bytes':len(patch),
             'scope':'phase6 retained files, named own prose/helpers, scoped staged diff',
             'forms':'literal, URL quoted, full percent upper/lower, JSON unicode escapes',
             'network_calls':0,'secret_output':False}
    name=sys.argv[1]
    if name not in ('credential-audit-files.json','credential-audit-staged.json'):raise ValueError('receipt')
    with (HERE/'live-20261008'/name).open('x') as f:json.dump(receipt,f,indent=2);f.write('\n')
    print('Credential audit: '+('PASS' if not matched else 'FAIL'))
    return 1 if matched else 0

if __name__=='__main__':
    try:code=main()
    except Exception:raise SystemExit('Credential audit refused safely; no secret details emitted') from None
    raise SystemExit(code)
