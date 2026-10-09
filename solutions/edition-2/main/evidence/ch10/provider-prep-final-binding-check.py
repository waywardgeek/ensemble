"""Final split binding controls after deadline-only support refinement; no Go/API."""
import copy,json,sys
from pathlib import Path
here=Path(__file__).resolve().parent;sys.path.insert(0,str(here/'support'))
from identity import preflight,check_launch,SUPPORT_PREFIX
from driver import launch_record
repo=here.parents[4];b=json.loads((here/'provider-prep-binding-final.json').read_text())
preflight(b,repo)
launch=launch_record(b,'openai','fixture','cli',[b['binaries']['cli']['path'],'chat'],{'LLM_VENDOR':'openai','LLM_MODEL':'fixture'},here,here)
checks=[]
for case in ('support-bytes','launch-support'):
    altered=copy.deepcopy(b);l=copy.deepcopy(launch);destination=here/('provider-prep-forbidden-'+case+'.json')
    assert not destination.exists()
    try:
        if case=='support-bytes':altered['sources'][SUPPORT_PREFIX+'provider.py']='0'*64
        else:l['support_revision']='0'*40
        preflight(altered,repo);check_launch(altered,l);destination.write_text('must not write')
    except ValueError as e:
        expected='source identity mismatch' if case=='support-bytes' else 'launch support revision mismatch'
        assert expected in str(e);assert not destination.exists();checks.append({'case':case,'refusal':str(e),'no_derived_write':True})
    else:raise AssertionError('accepted mutation')
(here/'provider-prep-final-binding-controls.json').write_text(json.dumps({'runtime_revision':b['source_revision'],'support_revision':b['support_revision'],'valid_actual_build_parent':True,'launch_is_synthetic_not_executed':True,'checks':checks},indent=2)+'\n')
print('final actual-build preflight and two changed-support/launch refusal controls passed; no client/provider execution')
