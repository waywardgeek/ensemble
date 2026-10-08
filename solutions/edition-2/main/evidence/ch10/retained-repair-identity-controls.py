"""Actual built parent identity controls; no model transport or derived rewrites."""
from pathlib import Path
import copy,json,sys
repo=Path(__file__).resolve().parents[5]
ev=Path(__file__).resolve().parent
sys.path.insert(0,str(ev/'support'))
from identity import preflight,verify_run,file_hash
b=json.loads((ev/'retained-repair-build/binding.json').read_text())
run=ev/'retained-repair-local/anthropic-A1';launch=json.loads((run/'launch.json').read_text())
originals={str(p.relative_to(run)):file_hash(p) for p in run.rglob('*') if p.is_file()}
out=ev/'retained-repair-identity-controls';out.mkdir()
verify_run(b,launch,originals,repo,run,out/'valid-parent.json')
rows=[]
def check(name,expected,mutate):
 altered=copy.deepcopy(b);l=copy.deepcopy(launch);mutate(altered,l);dest=out/(name+'.json')
 try:verify_run(altered,l,originals,repo,run,dest)
 except ValueError as error:
  assert expected in str(error),(name,str(error));assert not dest.exists();rows.append({'case':name,'refusal':str(error),'no_derived_write':True})
 else:raise AssertionError('accepted mutation '+name)
source=next(iter(b['sources']))
check('source','source identity mismatch',lambda v,l:v['sources'].__setitem__(source,'0'*64))
for role in b['binaries']:
 check('binary-'+role,role+' binary identity mismatch',lambda v,l,role=role:v['binaries'][role].__setitem__('sha256','0'*64))
check('support-map','incomplete support map',lambda v,l:v['support'].pop(next(iter(v['support']))))
check('module','module dependency identity mismatch',lambda v,l:v['modules'].__setitem__(next(iter(v['modules'])),'0'*64))
check('dependency','browser dependency identity mismatch',lambda v,l:v['browser_dependencies']['files'].__setitem__(next(iter(v['browser_dependencies']['files'])),'0'*64))
check('association','build association mismatch',lambda v,l:v['builds']['consumer'].__setitem__('source_revision','0'*40))
check('launch-source','launch source mismatch',lambda v,l:l.__setitem__('source_revision','0'*40))
check('launch-binary','launch executable path mismatch',lambda v,l:l['command'].__setitem__(0,'/invalid/cli'))
# Mutate one real working support byte only; restore in finally, never build it.
p=ev/'support/speech-probe.mjs';raw=p.read_bytes()
try:
 p.write_bytes(raw+b'\n')
 check('working-support','working source identity mismatch',lambda v,l:None)
finally:p.write_bytes(raw)
preflight(b,repo)
(out/'results.json').write_text(json.dumps({'actual_source_revision':b['source_revision'],'valid_actual_parent':True,'negative_controls':rows,'restored_parent_passed':True},indent=2)+'\n')
print(len(rows),'intended negative controls passed from real build/launch parent')
