#!/usr/bin/env python3
"""Contract-derived offline skill transition checks; no catalog or HTTP required.

Forgery refusals count only after their valid parent fixture renders successfully.
This is an offline reader check, not proof of live append authorization.
"""
import argparse
import copy
import json
from pathlib import Path
import subprocess
import tempfile

from accept_ch09 import ROOT, MODELS, environment, sha, system_text, dialogue


def material(identity, name, kind, body, tools=(), offers=()):
    return dict(activation=identity, name=name, type=kind, body=body,
                sha256=sha(body.encode()), tools=sorted(tools), dependencies=[],
                offers=[dict(name=n, description=d) for n, d in offers])


def state(revision, edit=None, retired=()):
    active=[dict(name='base',type='primary',activation=1)]
    if edit is not None:active.append(dict(name='edit',type='loadable',activation=edit))
    available=[dict(name='side',description='Side manual')]
    if edit is None:available.insert(0,dict(name='edit',description='Edit manual'))
    return dict(revision=revision,primary='base',roots=['edit'] if edit else [],active=active,
                available=available,tools=sorted(['load_skill','read_file','unload_skill']+(['write_file'] if edit else [])),
                retired=[dict(name='edit',activation=n) for n in retired])


def fixture():
    base=material(1,'base','primary','Replay identity.', ['read_file'], [('edit','Edit manual'),('side','Side manual')])
    edit=material(2,'edit','loadable','First frozen manual.\n',['write_file'])
    second=material(3,'edit','loadable','First frozen manual.\n',['write_file'])
    values=[('skills_initialized',dict(action='initialize',name='base',ceiling=['load_skill','read_file','unload_skill','write_file'],state=state(0),activated=[base])),
            ('message_received',dict(actor='human',purpose='dialogue',parts=[dict(type='text',text='Before editing.')])),
            ('response_ended',{'from':dict(vendor='openai',model='gpt-4.1-mini-2025-04-14',surface='chat_completions'),
                               'parts':[dict(type='text',text='Accepted before editing.')],
                               'usage':dict(input=1,cache_write=0,cache_read=0,output=1)}),
            ('skills_changed',dict(action='load',name='edit',state=state(1,2),activated=[edit])),
            ('skills_changed',dict(action='unload',name='edit',state=state(2,retired=[2]),activated=[])),
            ('skills_changed',dict(action='load',name='edit',state=state(3,3,[2]),activated=[second])),
            ('message_received',dict(actor='human',purpose='dialogue',parts=[dict(type='text',text='After editing.')]))]
    return [dict(seq=i,time='2026-01-01T00:00:00Z',type=kind,**{'skills' if kind.startswith('skills_') else 'response' if kind=='response_ended' else 'message':value}) for i,(kind,value) in enumerate(values,1)]


def forgeries():
    out={}
    def changed(name, edit):
        events=fixture();edit(events);out[name]=events
    def set_at(index,key,value):return lambda e:e[index]['skills'].__setitem__(key,value)
    changed('later-initializer',lambda e:e.insert(3,copy.deepcopy(e[0])))
    changed('initial-revision',lambda e:e[0]['skills']['state'].__setitem__('revision',1))
    changed('initial-retired',lambda e:e[0]['skills']['state'].__setitem__('retired',[dict(name='edit',activation=2)]))
    changed('initial-dynamic-root',lambda e:e[0]['skills']['state'].__setitem__('roots',['edit']))
    changed('initial-primary-mismatch',lambda e:e[0]['skills']['state'].__setitem__('primary','other'))
    changed('initial-id-not-one',lambda e:(e[0]['skills']['activated'][0].__setitem__('activation',2),e[0]['skills']['state']['active'][0].__setitem__('activation',2)))
    changed('revision-skip',lambda e:e[3]['skills']['state'].__setitem__('revision',2))
    changed('hidden-root',set_at(3,'name','hidden'))
    changed('unrelated-requested-root',set_at(3,'name','side'))
    changed('changed-ceiling',set_at(3,'ceiling',['load_skill','unload_skill','write_file']))
    changed('changed-primary',lambda e:e[3]['skills']['state'].__setitem__('primary','edit'))
    changed('fresh-id-gap',lambda e:(e[3]['skills']['activated'][0].__setitem__('activation',3),e[3]['skills']['state']['active'][1].__setitem__('activation',3)))
    changed('body-hash-mismatch',lambda e:e[3]['skills']['activated'][0].__setitem__('body','Altered body.'))
    changed('uppercase-digest',lambda e:e[3]['skills']['activated'][0].__setitem__('sha256',e[3]['skills']['activated'][0]['sha256'].upper()))
    changed('grant-outside-ceiling',lambda e:(e[3]['skills']['activated'][0].__setitem__('tools',['fictional']),e[3]['skills']['state'].__setitem__('tools',['fictional','load_skill','read_file','unload_skill'])))
    changed('missing-management-grant',lambda e:e[3]['skills']['state']['tools'].remove('unload_skill'))
    changed('unearned-grant',lambda e:e[0]['skills']['state']['tools'].append('write_file'))
    changed('missing-active-grant',lambda e:e[3]['skills']['state']['tools'].remove('write_file'))
    changed('false-discoverable',lambda e:e[3]['skills']['state']['available'].append(dict(name='hidden',description='Hidden')))
    changed('self-dependency',lambda e:e[3]['skills']['activated'][0].__setitem__('dependencies',[2]))
    changed('primary-as-dependency',lambda e:e[3]['skills']['activated'][0].__setitem__('dependencies',[1]))
    changed('extra-active-record',lambda e:e[3]['skills']['state']['active'].append(dict(name='side',type='loadable',activation=99)))
    changed('altered-retained-material',lambda e:e[3]['skills']['activated'].insert(0,material(1,'base','primary','Altered primary.',['read_file'])))
    changed('unload-new-material',lambda e:e[4]['skills']['activated'].append(material(3,'side','loadable','Unexpected.')))
    changed('unload-wrong-root',set_at(4,'name','side'))
    changed('missing-retirement',lambda e:e[4]['skills']['state'].__setitem__('retired',[]))
    changed('primary-retirement',lambda e:e[4]['skills']['state']['retired'].append(dict(name='base',activation=1)))
    changed('reused-retired-id',lambda e:(e[5]['skills']['activated'][0].__setitem__('activation',2),e[5]['skills']['state']['active'][1].__setitem__('activation',2)))
    changed('forgotten-old-retirement',lambda e:e[5]['skills']['state'].__setitem__('retired',[]))
    changed('duplicate-retirement',lambda e:e[5]['skills']['state']['retired'].append(dict(name='edit',activation=2)))
    changed('zero-activation',lambda e:e[3]['skills']['activated'][0].__setitem__('activation',0))
    changed('uint64-overflow',lambda e:e[3]['skills']['activated'][0].__setitem__('activation',18446744073709551616))
    changed('unknown-material-field',lambda e:e[3]['skills']['activated'][0].__setitem__('unexpected',True))
    changed('missing-array',lambda e:e[3]['skills']['activated'][0].pop('offers'))
    # Envelope sequence remains valid: target the skill reducer, not log ordering.
    for events in out.values():
        for index,event in enumerate(events,1):event['seq']=index
    return out


def rendered_ok(result,vendor):
    if result['exit']!=0:return False
    try:body=json.loads(result['stdout'])
    except ValueError:return False
    texts=[text for kind,text in dialogue(vendor,body) if kind=='text']
    required=['Before editing.','[skill edit activation 2]\nFirst frozen manual.\n\n[/skill]',
              '[skill edit activation 3]\nFirst frozen manual.\n\n[/skill]','After editing.']
    joined='\n'.join(texts)
    return system_text(vendor,body)=='Replay identity.' and all(joined.count(x)==1 for x in required) and [joined.index(x) for x in required]==sorted(joined.index(x) for x in required)


def capture(binary,vendor,events):
    with tempfile.TemporaryDirectory(prefix='ch09-replay-') as directory:
        work=Path(directory);path=work/'offline.jsonl'
        path.write_text('\n'.join(json.dumps(x,separators=(',',':')) for x in [dict(log_version=1)]+events)+'\n')
        original=path.read_bytes();env=environment(work,vendor,'http://127.0.0.1:1')
        # Deliberately absent live catalog; offline authority comes from the log.
        process=subprocess.run([str(binary),'render',str(path)],cwd=work,env=env,text=True,capture_output=True,timeout=10)
        return dict(exit=process.returncode,stdout=process.stdout,stderr=process.stderr,unchanged=path.read_bytes()==original)


def main():
    p=argparse.ArgumentParser(description=__doc__);p.add_argument('binary',type=Path);p.add_argument('--receipt',type=Path);args=p.parse_args();binary=args.binary.resolve()
    identities={str(path):sha(path.read_bytes()) for path in [Path(__file__),ROOT/'scripts/edition2/accept_ch09.py',ROOT/'book/edition-2/chapter-09.md',binary]}
    runs={};checks=[]
    for vendor in MODELS:
        positive=capture(binary,vendor,fixture());valid=rendered_ok(positive,vendor) and positive['unchanged']
        runs[vendor]={'positive':positive,'negative':{}};checks.append(dict(id=vendor+'/valid-offline-chronology',passed=valid,diagnostic=None if valid else dict(exit=positive['exit'],stderr=positive['stderr'][:2000],stdout=positive['stdout'][:2000])))
        for name,events in forgeries().items():
            result=capture(binary,vendor,events);runs[vendor]['negative'][name]=result
            checks.append(dict(id=vendor+'/'+name,passed=valid and result['exit']!=0 and bool(result['stderr'].strip()) and not result['stdout'].strip() and result['unchanged'],positive_parent=valid))
    assert identities=={path:sha(Path(path).read_bytes()) for path in identities},'input changed'
    result=dict(passed=sum(x['passed'] for x in checks),total=len(checks),checks=checks,runs=runs,input_sha256=identities,
                limits='Offline transition validation only. Invalid cases receive no credit without a passing valid log; live append/candidate authorization and unresolved tool batches remain separate.')
    if args.receipt:args.receipt.write_text(json.dumps(result,indent=2)+'\n')
    print(json.dumps({k:v for k,v in result.items() if k!='runs'},indent=2));return int(result['passed']!=result['total'])


if __name__=='__main__':raise SystemExit(main())
