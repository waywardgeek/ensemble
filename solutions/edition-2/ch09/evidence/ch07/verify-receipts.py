"""Identity-first replay: raw wire receipts remain separate from derived renders."""
import argparse
import json
from pathlib import Path
import subprocess
from evidence import HERE, preflight

def verify(binding, runs, paths, output):
    preflight(binding, paths)
    launches=[]
    # Preflight every launch before making the output directory or any derived file.
    for run in runs:
        launch=json.loads((run/'launch.json').read_text())
        assert launch['source_revision']==binding['source_revision'], 'launch source mismatch; no evidence modified'
        assert launch['executables']=={n:v['sha256'] for n,v in binding['executables'].items()}, 'launch executable mismatch; no evidence modified'
        assert launch['support']==binding['support'], 'launch support mismatch; no evidence modified'
        assert launch['browser_tools']==binding['browser_tools'], 'launch browser tools mismatch; no evidence modified'
        assert launch['launched_executable'] in ('cli','gui','consumer'), 'unknown launch executable'
        launches.append((run,launch))
    results=[]
    reconstructed=[]
    for run,launch in launches:
        requests=[json.loads(p.read_text()) for p in sorted((run/'requests').glob('*.json'))]
        renders=[]
        for log in sorted(list(run.glob('*.log'))+list((run/'workspace').glob('*.jsonl'))):
            events=[json.loads(line) for line in log.read_text().splitlines()][1:]
            for event in events:
                if event.get('type')!='request_sent':continue
                command=[paths.get('cli',binding['executables']['cli']['path']),'replay',str(log),str(event['seq'])]
                body=subprocess.check_output(command,cwd=run)
                value=json.loads(body);renders.append(value)
                reconstructed.append((f'{run.name}-{log.stem}-{event["seq"]}.json',body))
        # Multi-Agent wire order is not deterministic; exact multiset equality is.
        canonical=lambda values:sorted(json.dumps(v,sort_keys=True,separators=(',',':')) for v in values)
        assert canonical(requests)==canonical(renders), f'request reconstruction mismatch: {run.name}'
        results.append({'run':run.name,'requests':len(requests),'reconstructed':len(renders),'delivery':launch['delivery']})
    output.mkdir()
    for name,body in reconstructed:(output/name).write_bytes(body)
    (output/'receipts.json').write_text(json.dumps(results,indent=2)+'\n')
    return results

if __name__=='__main__':
    p=argparse.ArgumentParser();p.add_argument('runs',nargs='+');p.add_argument('--output',required=True);p.add_argument('--cli');p.add_argument('--consumer');p.add_argument('--gui');args=p.parse_args()
    binding=json.loads((HERE/'initial-binding.json').read_text())
    paths={k:getattr(args,k) for k in ('cli','gui','consumer') if getattr(args,k)}
    print(json.dumps(verify(binding,[HERE/name for name in args.runs],paths,HERE/args.output),indent=2))
