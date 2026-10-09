"""Bind committed interpreted support to unchanged real historical Go builds.

No builds or credentials. The original build maps/associations are retained;
compiled/runtime/module changes refuse before creating the output.
"""
import argparse
import copy
import json
from pathlib import Path
from identity import SUPPORT_PREFIX, PREFIX, source_paths, historical, digest, preflight


def combine(parent, repo, revision):
    value = copy.deepcopy(parent)
    # First adoption preserves the complete original source manifest forever.
    value['build_sources'] = value.get('build_sources', copy.deepcopy(parent['sources']))
    value['support_revision'] = revision
    support = {p:digest(historical(repo,revision,p)) for p in source_paths(repo,revision) if p.startswith(SUPPORT_PREFIX)}
    value['sources'] = {p:h for p,h in value['build_sources'].items() if not p.startswith(SUPPORT_PREFIX)} | support
    value['support'] = support
    value['modules'] = {p:h for p,h in value['sources'].items() if Path(p).name in ('go.mod','go.sum')}
    value['catalog'] = {p:h for p,h in support.items() if '/catalog/' in p or p.endswith('/bindings.json')}
    value['schedule_sha256'] = support[SUPPORT_PREFIX+'schedule.json']
    preflight(value,repo)
    return value


if __name__ == '__main__':
    p=argparse.ArgumentParser();p.add_argument('--parent',required=True,type=Path);p.add_argument('--repo',required=True,type=Path);p.add_argument('--revision',required=True);p.add_argument('--output',required=True,type=Path)
    a=p.parse_args(); b=combine(json.loads(a.parent.read_text()),a.repo.resolve(),a.revision)
    with a.output.open('x') as f: json.dump(b,f,indent=2); f.write('\n')
    print('runtime/build revision '+b['source_revision']+'; interpreted support revision '+b['support_revision'])
