#!/usr/bin/env python3
"""Independent fixed graph scenarios through the public CLI, all three adapters.

Expected transitions are concrete contract examples, not calls to a student
resolver. No real model endpoints or credentials are used.
"""
import argparse
import copy
import json
from pathlib import Path
import subprocess
import tempfile
import threading

from accept_ch09 import ROOT, MODELS, Fixture, call, definition, environment, sha, dialogue, declarations


def graph():
    return {
        'base':definition('base','primary','Primary remains fixed.\n','read_file','edit review broken cycle unavailable variable'),
        'read':definition('read','dependency','Read dependency.\n$TOOLS\n$SKILLS\n','list_directory','search'),
        'edit':definition('edit','loadable','Edit material.\n$TOOLS\n$SKILLS\n','write_file',extra='depends: read\n'),
        'review':definition('review','loadable','Review material.\n',extra='depends: read\n'),
        'search':definition('search','loadable','Search material.\n','search_files'),
        'hidden':definition('hidden','loadable','Never visible.'),
        'broken':definition('broken','loadable','Never partially active.',extra='depends: first second\n'),
        'first':definition('first','dependency','Valid first branch.','write_file'),
        'second':definition('second','dependency','Invalid second branch.',extra='depends: missing\n'),
        'cycle':definition('cycle','loadable','Cycle root.',extra='depends: left\n'),
        'left':definition('left','dependency','Left.',extra='depends: right\n'),
        'right':definition('right','dependency','Right.',extra='depends: left\n'),
        'unavailable':definition('unavailable','loadable','Uninstalled handler.','fictional_handler'),
        'variable':definition('variable','loadable','Unknown variable $UNSUPPLIED.','write_file'),
    }


BATCHES = [
    [call('edit1','load_skill',name='edit')],
    [call('review','load_skill',name='review')],
    [call('search','load_skill',name='search')],
    [call('edit-off1','unload_skill',name='edit'),call('write-refused','write_file',path='forbidden.txt',content='wrong')],
    [call('edit2','load_skill',name='edit')],
    [call('review-off','unload_skill',name='review')],
    [call('edit-off2','unload_skill',name='edit'),call('already-search','load_skill',name='search')],
    [call('search-off','unload_skill',name='search'),call('noop','unload_skill',name='search')],
    [call('broken','load_skill',name='broken'),call('cycle','load_skill',name='cycle'),call('unavailable','load_skill',name='unavailable'),call('variable','load_skill',name='variable')],
    [call('hidden','load_skill',name='hidden'),call('primary','load_skill',name='base'),call('dependency','load_skill',name='read')],
    [],
]
# Each tuple is one changed state: roots, active name/ID pairs, retired IDs.
STATES = [
    (['edit'],{'base':1,'read':2,'edit':3},[]),
    (['edit','review'],{'base':1,'read':2,'edit':3,'review':4},[]),
    (['edit','review','search'],{'base':1,'read':2,'edit':3,'review':4,'search':5},[]),
    (['review','search'],{'base':1,'read':2,'review':4,'search':5},[3]),
    (['edit','review','search'],{'base':1,'read':2,'edit':6,'review':4,'search':5},[3]),
    (['edit','search'],{'base':1,'read':2,'edit':6,'search':5},[3,4]),
    (['search'],{'base':1,'search':5},[2,3,4,6]),
    ([],{'base':1},[2,3,4,5,6]),
]


def evaluate(events,requests,files,vendor):
    changes=[e['skills'] for e in events if e.get('type')=='skills_changed']
    initial=[e['skills'] for e in events if e.get('type')=='skills_initialized']
    results={}
    states=[]
    for index,want in enumerate(STATES):
        if index>=len(changes):states.append(False);continue
        state=changes[index]['state'];roots,active,retired=want
        states.append(state['revision']==index+1 and state['primary']=='base' and state['roots']==roots and
                      [(x['name'],x['activation']) for x in state['active']]==sorted(active.items()) and
                      [x['activation'] for x in state['retired']]==retired)
    results['exact-closure-roots-retirement']=all(states) and len(changes)==8
    results['fresh-material-identities']=False
    if len(initial)==1 and len(changes)==8:
        materials=initial[0]['activated']+sum((x['activated'] for x in changes),[])
        expected=[('base',1),('read',2),('edit',3),('review',4),('search',5),('edit',6)]
        results['fresh-material-identities']=[(x['name'],x['activation']) for x in materials]==expected and all(x['sha256']==sha(x['body'].encode()) for x in materials)
    returns={e['tool']['call_id']:e['tool'] for e in events if e.get('type')=='tool_returned'}
    expected_errors={'broken':'skill_dependency','cycle':'skill_dependency','unavailable':'skill_tool_unavailable','variable':'skill_variable','hidden':'skill_unavailable','primary':'skill_unavailable','dependency':'skill_unavailable'}
    atomic=True
    for name,code in expected_errors.items():
        value=returns.get(name,{});text=''.join(x.get('text','') for x in value.get('parts',[]))
        try:ack=json.loads(text)
        except ValueError:ack={}
        atomic &= value.get('is_error') is True and ack=={'error':code,'name':{'primary':'base','dependency':'read'}.get(name,name),'revision':8}
    results['failed-graph-variable-candidate-atomic']=atomic and len(changes)==8
    results['revocation-real-effect']=not files and returns.get('write-refused',{}).get('is_error') is True
    noops=True
    for name,revision in [('already-search',7),('noop',8)]:
        value=returns.get(name,{});text=''.join(x.get('text','') for x in value.get('parts',[]))
        try:ack=json.loads(text)
        except ValueError:ack={}
        noops &= not value.get('is_error',False) and ack=={'status':'unchanged','name':'search','revision':revision,'changed':False}
    results['explicit-child-noop-after-advertiser-loss']=noops and len(changes)==8
    grants=[]
    for roots,active,_ in STATES:
        names={'load_skill','unload_skill','read_file'}
        if 'read' in active:names.add('list_directory')
        if 'edit' in active:names.add('write_file')
        if 'search' in active:names.add('search_files')
        grants.append(sorted(names))
    results['shared-and-primary-grants']=len(requests)==11 and [declarations(vendor,x) for x in requests[1:9]]==grants and all(declarations(vendor,x)==grants[-1] for x in requests[9:])
    results['dependency-offers']=len(changes)==8 and [x['name'] for x in changes[0]['state']['available']]==['broken','cycle','review','search','unavailable','variable']
    results['candidate-variable-freeze']=False
    if len(initial)==1 and len(changes)==8:
        first=changes[0]['activated'];again=changes[4]['activated']
        want_tools='\n'.join('- '+x for x in grants[0])
        want_offers='\n'.join('- '+x+': '+x.title()+' scratch files' for x in ['broken','cycle','review','search','unavailable','variable'])
        results['candidate-variable-freeze']=len(first)==2 and first[0]['body']=='Read dependency.\n'+want_tools+'\n'+want_offers+'\n' and first[1]['body']=='Edit material.\n'+want_tools+'\n'+want_offers+'\n' and len(again)==1 and again[0]['activation']==6
    results['retired-manual-once']=False
    if requests and len(initial)==1 and len(changes)==8:
        texts=[v for kind,v in dialogue(vendor,requests[-1]) if kind=='text']
        results['retired-manual-once']=all(sum(v.count(f'[skill {name} activation {identity}]') for v in texts)==1 for name,identity in [('read',2),('edit',3),('review',4),('search',5),('edit',6)])
    return results


def capture(binary,vendor):
    server=Fixture(vendor,BATCHES);thread=threading.Thread(target=server.serve_forever);thread.start()
    try:
        with tempfile.TemporaryDirectory(prefix='ch09-graph-') as directory:
            work=Path(directory)
            for name,data in graph().items():
                path=work/'catalog'/name/'SKILL.md';path.parent.mkdir(parents=True);path.write_bytes(data)
            env=environment(work,vendor,f'http://127.0.0.1:{server.server_port}');env.update(LLM_SKILLS_DIR='catalog',LLM_PRIMARY_SKILL='base')
            result=subprocess.run([str(binary),'protocol'],input='{"user":"Exercise the local graph fixture."}\n',text=True,capture_output=True,cwd=work,env=env,timeout=40)
            path=work/'session.jsonl';events=[json.loads(x) for x in path.read_text().splitlines()][1:] if path.exists() else []
            files={x.name:x.read_text() for x in work.glob('*.txt')}
            return {'exit':result.returncode,'stderr':result.stderr,'events':events,'requests':server.requests,'files':files,
                    'checks':evaluate(events,server.requests,files,vendor)}
    finally:server.shutdown();server.server_close();thread.join()


def main():
    p=argparse.ArgumentParser(description=__doc__);p.add_argument('binary',type=Path);p.add_argument('--receipt',type=Path);args=p.parse_args();binary=args.binary.resolve()
    inputs=[Path(__file__),ROOT/'scripts/edition2/accept_ch09.py',ROOT/'book/edition-2/chapter-09.md',binary]
    identities={str(x):sha(x.read_bytes()) for x in inputs}
    runs={vendor:capture(binary,vendor) for vendor in MODELS}
    checks=[{'id':vendor+'/'+name,'passed':passed} for vendor,r in runs.items() for name,passed in r['checks'].items()]
    result={'passed':sum(x['passed'] for x in checks),'total':len(checks),'checks':checks,'runs':runs,'input_sha256':identities,
            'limits':'Focused graph/material scenarios; public concurrency, parser limits, forged replay and browser are separate.'}
    assert identities=={str(x):sha(x.read_bytes()) for x in inputs},'input identity changed during run'
    if args.receipt:args.receipt.write_text(json.dumps(result,indent=2)+'\n')
    print(json.dumps({k:v for k,v in result.items() if k!='runs'},indent=2));raise SystemExit(int(result['passed']!=result['total']))


if __name__=='__main__':main()
