"""Actual retained-build parent; synthetic launch for identity-only controls.

No Go/client/provider invocation and no relabeling of an old actual launch. Each
intended negative begins with a valid parent and precedes derived writes.
"""
import copy
import json
from pathlib import Path
import sys

here=Path(__file__).resolve().parent
sys.path.insert(0,str(here/'support'))
from identity import preflight, verify_run, file_hash, digest, canonical, SUPPORT_PREFIX
from driver import launch_record
repo=here.parents[4]
binding=json.loads(Path(sys.argv[1]).read_text());out=Path(sys.argv[2]);out.mkdir()
preflight(binding,repo)
run=out/'synthetic-launch';run.mkdir()
launch=launch_record(binding,'openai','fixture','cli',[binding['binaries']['cli']['path'],'chat'],{'LLM_VENDOR':'openai','LLM_MODEL':'fixture'},run,run)
(run/'launch.json').write_text(json.dumps(launch,indent=2)+'\n')
(run/'synthetic.txt').write_text('Identity-only launch fixture. No process or provider was invoked. Actual Go builds retain their original associations.\n')
originals={p.name:file_hash(p) for p in run.iterdir()}
verify_run(binding,launch,originals,repo,run,out/'valid-parent.json')
rows=[]
def negative(name,expected,mutate):
    preflight(binding,repo)
    b=copy.deepcopy(binding);l=copy.deepcopy(launch);mutate(b,l);destination=out/(name+'.json')
    before={str(p.relative_to(out)):file_hash(p) for p in out.rglob('*') if p.is_file()}
    try:verify_run(b,l,originals,repo,run,destination)
    except ValueError as e:
        assert expected in str(e),(name,str(e));assert not destination.exists()
        after={str(p.relative_to(out)):file_hash(p) for p in out.rglob('*') if p.is_file()}
        assert before==after
        rows.append({'case':name,'refusal':str(e),'no_derived_write':True})
    else:raise AssertionError('accepted mutation '+name)
source=next(p for p in binding['sources'] if not p.startswith(SUPPORT_PREFIX))
support=SUPPORT_PREFIX+'provider.py'
negative('source','source identity',lambda b,l:b['sources'].__setitem__(source,'0'*64))
negative('support-source','source identity',lambda b,l:b['sources'].__setitem__(support,'0'*64))
negative('empty-sources','incomplete source',lambda b,l:b.__setitem__('sources',{}))
negative('support-revision','incomplete source',lambda b,l:b.__setitem__('support_revision',b['source_revision']))
negative('historical-build','historical build source',lambda b,l:b['build_sources'].__setitem__(source,'0'*64))
negative('missing-build','incomplete historical build',lambda b,l:b['build_sources'].pop(source))
for role in binding['binaries']:
    negative('binary-'+role,role+' binary identity',lambda b,l,role=role:b['binaries'][role].__setitem__('sha256','0'*64))
negative('support-map','incomplete support',lambda b,l:b['support'].pop(support))
negative('module','module dependency',lambda b,l:b['modules'].__setitem__(next(iter(b['modules'])),'0'*64))
negative('catalog','catalog/binding',lambda b,l:b['catalog'].__setitem__(next(iter(b['catalog'])),'0'*64))
negative('dependency','browser dependency identity',lambda b,l:b['browser_dependencies']['files'].__setitem__(next(iter(b['browser_dependencies']['files'])),'0'*64))
negative('build-association','build association',lambda b,l:b['builds']['consumer'].__setitem__('source_revision','0'*40))
negative('launch-runtime','launch source',lambda b,l:l.__setitem__('source_revision','0'*40))
negative('launch-support','launch support revision',lambda b,l:l.__setitem__('support_revision','0'*40))
negative('launch-binary','launch executable path',lambda b,l:l['command'].__setitem__(0,'/invalid/cli'))
# Changing actual support bytes uses a tiny alternate input set, never mutates main.
support_root=out/'alternate-support';support_root.mkdir()
for name in binding['support']:
    target=support_root/name.removeprefix(SUPPORT_PREFIX);target.parent.mkdir(parents=True,exist_ok=True);target.write_bytes((repo/name).read_bytes())
preflight(binding,repo,support_root=support_root)
probe=support_root/'provider.py';probe.write_bytes(probe.read_bytes()+b'\n')
destination=out/'support-file.json'
try:
    preflight(binding,repo,support_root=support_root)
    verify_run(binding,launch,originals,repo,run,destination)
except ValueError as error:
    assert 'support identity mismatch' in str(error);assert not destination.exists()
    rows.append({'case':'actual-support-bytes','refusal':str(error),'no_derived_write':True})
else:raise AssertionError('accepted support mutation')
# Retain test inputs without adding apparent development modules to main.
retained=[]
for p in sorted(support_root.rglob('*')):
    if p.is_file() and (p.suffix=='.go' or p.name in ('go.mod','go.sum')):
        sha=file_hash(p);old=str(p.relative_to(out));target=p.with_name(p.name+'.fixture');p.rename(target)
        retained.append({'tested_path':old,'retained_path':str(target.relative_to(out)),'sha256':sha})
(out/'fixture-retention.json').write_text(json.dumps(retained,indent=2)+'\n')
preflight(binding,repo)
(out/'results.json').write_text(json.dumps({'runtime_revision':binding['source_revision'],'support_revision':binding['support_revision'],'actual_build_parent':True,'launch_is_synthetic_not_executed':True,'negative_controls':rows,'restored_parent_passed':True},indent=2)+'\n')
print(len(rows),'intended refusals from actual build parent; synthetic launch only; zero client/provider calls')
