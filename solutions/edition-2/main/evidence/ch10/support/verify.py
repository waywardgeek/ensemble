"""Seal originals and verify identity before writing derived feature findings.

Findings are bounded observations, never full chapter/live acceptance. No replay
is charged and no provider access exists here. Run only on a finished launch.
"""
import argparse
import json
from pathlib import Path
from identity import preflight, check_launch, file_hash, digest, canonical, require


def read_json(path):return json.loads(Path(path).read_text())


def originals(run):
    result={}
    for p in sorted(run.rglob('*')):
        require(not p.is_symlink(),'symlink in originals')
        if p.is_file() and p.name not in ('originals.json',) and 'derived' not in p.relative_to(run).parts:
            result[str(p.relative_to(run))]=file_hash(p)
    require(result,'empty originals')
    return result


def inspect_browser(rows):
    findings={'saved_ack_after_applied':False,'measured_speech_boundaries':[], 'policy_ack':False,'preference_ack':False}
    for row in rows:
        if row['kind'] in ('resume_speech','reconnect_speech','page_closed_speech','page_reopened_speech','reload_speech'):
            findings['measured_speech_boundaries'].append(row['kind'])
        if row['kind']=='dom':
            applied=None
            for item in row.get('applied',[]):
                if item['kind']=='applied':applied=item
                if item['kind']=='saved':
                    require(applied is not None,'saved before applied session')
                    require(str(applied['revision'])==str(item['watch_revision']),'saved revision mismatch')
                    require(str(applied['session']['checkpoint_seq'])==str(item['as_of']),'saved anchor mismatch')
                    findings['saved_ack_after_applied']=True
        if row['kind']=='received':
            msg=json.loads(row['payload'])
            if msg.get('type')=='policy_ack':findings['policy_ack']=True
            if msg.get('type')=='preferences_ack':findings['preference_ack']=True
    return findings


def check_budget(path):
    rows=[json.loads(line) for line in path.read_text().splitlines()]
    # Imported only after binding preflight in the CLI path.
    from relay import SCHEDULE
    generation=[r for r in rows if r['kind']=='generation']
    require(len(generation)<=22,'generation ceiling exceeded')
    require(sum(r['kind']=='discovery' for r in rows)<=1,'discovery ceiling exceeded')
    for step,spec in SCHEDULE['steps'].items():require(sum(r['step']==step for r in generation)<=spec['cap'],'step ceiling exceeded')
    for row,cap in SCHEDULE['rows'].items():require(sum(r['row']==row for r in generation)<=cap,'row ceiling exceeded')
    return {'generation_attempts':len(generation),'discovery_attempts':len(rows)-len(generation)}


def run(args):
    b=read_json(args.binding);preflight(b,args.repo)
    launch=read_json(args.run/'launch.json');check_launch(b,launch)
    require((args.run/'exit.json').is_file(),'launch has not finished')
    if args.seal:
        with (args.run/'originals.json').open('x') as f:json.dump(originals(args.run),f,indent=2)
        return
    expected=read_json(args.run/'originals.json')
    require(expected==originals(args.run),'original receipt identity mismatch')
    findings={'source_revision':b['source_revision'],'binding_sha256':digest(canonical(b)), 'mode':launch['mode'],'chapter_accepted':False,'identity_verified':True}
    if (args.run/'browser-original.jsonl').is_file():findings['browser']=inspect_browser([json.loads(s) for s in (args.run/'browser-original.jsonl').read_text().splitlines()])
    if (args.run/'attempts-final.jsonl').is_file():findings['budget']=check_budget(args.run/'attempts-final.jsonl')
    findings['limitations']=['Originals must be reviewed against all A/B/C/D observables. Identity and these limited checks are not full feature acceptance.','No human hearing, provider success, or chapter acceptance inferred from a local fixture.']
    # All checks above precede creating derived directory/output.
    destination=args.run/'derived';destination.mkdir(exist_ok=True)
    with (destination/'verified.json').open('x') as f:json.dump(findings,f,indent=2)


if __name__=='__main__':
    p=argparse.ArgumentParser();p.add_argument('--binding',required=True,type=Path);p.add_argument('--repo',required=True,type=Path);p.add_argument('--run',required=True,type=Path);p.add_argument('--seal',action='store_true');run(p.parse_args())
