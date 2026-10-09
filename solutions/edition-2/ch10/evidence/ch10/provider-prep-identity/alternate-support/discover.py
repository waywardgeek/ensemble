"""One explicitly released live discovery request; no automatic model choice."""
import argparse
import json
from pathlib import Path
from identity import preflight, digest, canonical
from provider import Provider, load_key, FIELDS
from relay import Budget


def run(args):
    binding=json.loads(args.binding.read_text());preflight(binding,args.repo)
    if not args.live: raise ValueError('explicit live selection required')
    root=args.root.resolve()
    if args.output.exists() or args.output.with_suffix('.body').exists(): raise ValueError('discovery output exists')
    # No key access before complete binding and input checks.
    provider=Provider(args.vendor,load_key(args.vendor))
    root.mkdir(exist_ok=True)
    result=provider.discover(Budget(root/'attempts.jsonl',args.vendor),args.output,digest(canonical(binding)))
    print('Discovery '+result['outcome']+'; inspect receipt for returned models; no pagination or retry.')
    return 0 if result['outcome']=='completed' else 1


if __name__=='__main__':
    p=argparse.ArgumentParser();p.add_argument('--binding',type=Path,required=True);p.add_argument('--repo',type=Path,required=True)
    p.add_argument('--vendor',choices=FIELDS,required=True);p.add_argument('--root',type=Path,required=True);p.add_argument('--output',type=Path,required=True);p.add_argument('--live',action='store_true')
    try: raise SystemExit(run(p.parse_args()))
    except (OSError,ValueError,TypeError,KeyError): raise SystemExit('bounded discovery refused; inspect nonsecret receipt if created') from None
