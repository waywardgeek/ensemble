"""Summarize retained runs after identity-first verification; no API requests."""
import json
import re
from evidence import HERE, preflight
binding=json.loads((HERE/'binding.json').read_text());preflight(binding,{})
summary=[]
for run in sorted(p.parent for p in HERE.glob('*/launch.json')):
    launch=json.loads((run/'launch.json').read_text());assert launch['source_revision']==binding['source_revision'] and launch['exit_code']==0
    terminal=(run/'terminal.txt').read_text();records=[]
    for log in sorted(run.glob('*.log')):
        events=[json.loads(x) for x in log.read_text().splitlines()][1:]
        assert [e['seq'] for e in events]==list(range(1,len(events)+1));records+=events
    responses=[e['response'] for e in records if e['type']=='response_ended']
    usage={k:sum(r['usage'].get(k,0) for r in responses) for k in ['input','cache_write','cache_read','output']}
    entry={'run':run.name,'source':launch['source_revision'],'vendor':launch['vendor'],'requested_model':launch['requested_model'],'returned_models':sorted({r['from']['model'] for r in responses}),'delivery':launch['delivery'],'requests':launch['requests'],'outcomes':[e['turn']['outcome'] for e in records if e['type']=='turn_ended'],'accepted_usage':usage,'stop_reasons':[r['stop_reason'] for r in responses]}
    assert len(list((run/'requests').glob('*.json')))==launch['requests']
    if launch['mode']=='chat':
        totals=re.search(r'Final usage: input=(\d+), cache write=(\d+), cache read=(\d+), output=(\d+)',terminal)
        assert totals and [int(v) for v in totals.groups()]==list(usage.values())
        artifact=run/'workspace/cr/io/1';assert artifact.read_text()=='CHAPTER-SIX-FILE-MARKER\nport=8080\n'
        tool=[e['tool'] for e in records if e['type']=='tool_returned'];assert len(tool)==1 and tool[0]['parts'][0]['text']==artifact.read_text()
        assert 'CHAPTER-SIX-FILE-MARKER' in terminal and '8080' in terminal
        if launch['delivery']=='stream':
            assert entry['outcomes']==['success','interrupted','success']
            assert 'Interrupt: request=r2 interrupted=true.' in terminal and 'Incomplete display: interrupted' in terminal
            assert '[r2 Assistant part 1]' in terminal and '[r3 Assistant part 1] RECOVERED-SIX-REVISION' in terminal
            start=next(i for i,e in enumerate(records) if e['type']=='turn_started' and e['turn']['request_id']=='r2')
            end=next(i for i,e in enumerate(records) if e['type']=='turn_ended' and e['turn']['request_id']=='r2')
            assert not any(e['type']=='response_ended' for e in records[start:end]);entry['interrupted_operation_has_no_accepted_response']=True
        else:
            assert 'proposed tool' not in terminal and 'Assistant part' not in terminal and terminal.count('Assistant:')==1
        entry['artifact']=str(artifact.relative_to(HERE))
    else:
        data=next(json.loads(line) for line in terminal.splitlines() if line.startswith('{'))
        assert data['completion_before_slow_release'] and len(data['completions'])==2
        identities=set();thinking=0
        for i,completion in enumerate(data['completions']):
            observations=data['observations'][i];finals=[v for v in observations if v['kind']=='part_final']
            assert finals==data['finals_only'][i] and [v['part'] for v in finals]==completion['parts']
            assert completion['outcome']=='success' and all(v['agent_id']==completion['agent_id'] for v in observations)
            for v in observations:
                if v['kind']=='part_delta':
                    identities.add((v['agent_id'],v['request_id'],v['operation_id'],v['part_id']));thinking+=v['channel']=='thinking'
            assert sum(v['kind']=='model_end' and v['accepted'] for v in observations)==1
        entry.update(completion_before_slow_release=True,full_identity_groups=len(identities),slow_subscription_status=data['slow_subscription_status'],thinking_delta_count=thinking,finals_match_completions=True)
    summary.append(entry)
assert sum(r['requests'] for r in summary)==20
(HERE/'live-summary.json').write_text(json.dumps(summary,indent=2)+'\n')
print(json.dumps(summary,indent=2))
