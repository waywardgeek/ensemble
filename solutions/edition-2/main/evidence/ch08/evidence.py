"""Identity-first Chapter 8 launch/replay support, adapted from validated Chapter 7."""
import hashlib
import json
from pathlib import Path
import subprocess
import sys
HERE=Path(__file__).resolve().parent
ROOT=HERE.parents[4]
PREFIX='solutions/edition-2/main/'
SUPPORT=['evidence.py','terminal-run.py','verify-receipts.py','browser-live.mjs','capture-audio.swift','test-evidence.py']
EXECUTABLES={'cli','gui','consumer','headless','interpreter','node','recorder','capture','chrome'}
def digest(data):return hashlib.sha256(data).hexdigest()
def historical(revision,path):return subprocess.check_output(['git','show',revision+':'+path],cwd=ROOT)
def source_set(revision):
    paths=subprocess.check_output(['git','ls-tree','-r','--name-only',revision,'--',PREFIX],cwd=ROOT,text=True).splitlines()
    return sorted(p for p in paths if '/evidence/' not in p and (p.endswith('.go') or Path(p).name in ('go.mod','go.sum') or '/gui/web/gui/' in p))
def browser_set(root):
    root=Path(root).resolve()
    assert (root/'package-lock.json').is_file(), 'browser dependency root missing lockfile; no evidence modified'
    return sorted(str(p.resolve()) for p in root.rglob('*') if p.is_file() and (p.suffix in ('.js','.mjs','.json') or p.name=='package-lock.json'))
def preflight(binding,paths):
    revision=binding['source_revision'];expected=source_set(revision)
    assert expected and binding['sources'],'empty source set; no evidence modified'
    assert sorted(binding['sources'])==expected,'incomplete historical source set; no evidence modified'
    for p in expected:assert digest(historical(revision,p))==binding['sources'][p],'historical source mismatch; no evidence modified'
    assert set(binding['executables'])==EXECUTABLES,'incomplete executable identities; no evidence modified'
    assert digest(Path(sys.executable).read_bytes())==binding['executables']['interpreter']['sha256'],'interpreter mismatch; no evidence modified'
    for name,identity in binding['executables'].items():assert digest(Path(paths.get(name,identity['path'])).read_bytes())==identity['sha256'],'executable mismatch; no evidence modified'
    assert set(binding['support'])==set(SUPPORT),'incomplete support identities; no evidence modified'
    for name in SUPPORT:
        path=PREFIX+'evidence/ch08/'+name
        assert digest(historical(revision,path))==binding['support'][name],'historical support mismatch; no evidence modified'
        assert digest((HERE/name).read_bytes())==binding['support'][name],'support identity mismatch; no evidence modified'
    assert binding['browser_tools'],'missing browser dependency identities; no evidence modified'
    assert sorted(binding['browser_tools'])==browser_set(binding['browser_root']),'incomplete browser dependency set; no evidence modified'
    for path,expected_hash in binding['browser_tools'].items():assert digest(Path(path).read_bytes())==expected_hash,'browser dependency mismatch; no evidence modified'
def make_binding(revision,paths,browser_root):
    paths={**paths,'interpreter':sys.executable,'recorder':'/usr/bin/script'}
    browser_root=Path(browser_root)
    dependencies=[Path(p) for p in browser_set(browser_root)]
    b={'browser_root':str(browser_root.resolve()),'source_revision':revision,'sources':{p:digest(historical(revision,p)) for p in source_set(revision)},'executables':{n:{'path':str(Path(p).resolve()),'sha256':digest(Path(p).read_bytes())} for n,p in paths.items()},'support':{n:digest((HERE/n).read_bytes()) for n in SUPPORT},'browser_tools':{str(p.resolve()):digest(p.read_bytes()) for p in dependencies}}
    preflight(b,paths);return b
