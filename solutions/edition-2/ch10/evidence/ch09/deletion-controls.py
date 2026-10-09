#!/usr/bin/env python3
"""Student-owned Go overlay mutations; never edits delivered source or a grader."""
import argparse
import hashlib
import json
import pathlib
import re
import subprocess
import tempfile

parser = argparse.ArgumentParser()
parser.add_argument('--source', required=True, type=pathlib.Path)
args = parser.parse_args()
root = args.source.resolve()
checks = []

def run(test, package='.', overlay=None):
    command = ['go', 'test']
    if overlay:
        command += ['-overlay', str(overlay)]
    command += [package, '-run', test, '-count=1', '-timeout=30s']
    result = subprocess.run(command, cwd=root, text=True, capture_output=True, timeout=45)
    return result.returncode, result.stdout + result.stderr

for test, package in [('^TestSkillAdmissionEffectBoundary$', '.'), ('^TestSkillPublishedHintMaterialPromptProjection$', '.'), ('^TestDiamondSecondBranchFailureLeavesWholeLedger$', './internal/skills')]:
    code, output = run(test, package)
    assert code == 0, output
    checks.append({'positive': test, 'passed': True})

mutations = []
path = root / 'internal/tools/registry.go'
source = path.read_text()
start = source.index('func (r *Registry) Kind(')
end = source.index('// ExecuteJob', start)
part = source[start:end]
matches = list(re.finditer(r'\tif available \{\n.*?\n\t\}\n', part, re.S))
assert len(matches) == 1
old = matches[0].group()
assert source.count(old) == 1
mutations.append((path, source.replace(old, ''), '^TestSkillAdmissionEffectBoundary$', '.', 'disabled tool created forbidden-before'))

path = root / 'internal/llm/events.go'
source = path.read_text()
start = source.index('func Apply(')
anchor = '\tcase "redacted":\n'
assert source[start:].count(anchor) == 1
pos = source.index(anchor, start) + len(anchor)
mutated = source[:pos] + '\t\tfor i:=range c.Entries {if c.Entries[i].Purpose=="skill" {c.Entries[i].Parts=nil}}\n' + source[pos:]
mutations.append((path, mutated, '^TestSkillPublishedHintMaterialPromptProjection$', '.', 'consumption/retention failed'))

path = root / 'internal/skills/skills.go'
source = path.read_text()
anchor = '\t\tvisits[name] = 2\n'
assert source.count(anchor) == 1
mutated = source.replace(anchor, '\t\tif s.committed!=nil {s.committed.state.Tools=append(s.committed.state.Tools,def.Tools...)}\n' + anchor)
mutations.append((path, mutated, '^TestDiamondSecondBranchFailureLeavesWholeLedger$', './internal/skills', 'half diamond published'))

with tempfile.TemporaryDirectory(prefix='ch09-student-overlays-') as temp:
    temp = pathlib.Path(temp)
    for index, (path, mutated, test, package, expected) in enumerate(mutations):
        original = path.read_bytes()
        changed = temp / f'mutated-{index}.go'
        changed.write_text(mutated)
        overlay = temp / f'overlay-{index}.json'
        overlay.write_text(json.dumps({'Replace': {str(path): str(changed)}}))
        code, output = run(test, package, overlay)
        assert code != 0 and expected in output, output
        assert path.read_bytes() == original
        checks.append({'mutation': test, 'passed': True, 'specific_failure': expected, 'source_sha256': hashlib.sha256(original).hexdigest()})
print(json.dumps({'scope': 'local student controls with temporary Go overlays; no runtime source edits', 'checks': checks, 'passed': True}, indent=2))
