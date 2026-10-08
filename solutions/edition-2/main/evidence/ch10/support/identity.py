"""Identity-first evidence checks, adapted from permitted Chapter 9 support.

No credentials, discovery, builds, or derived writes during preflight. A binding
is only produced after separately recorded real builds; no draft binary claim.
"""
import hashlib
import json
from pathlib import Path
import re
import subprocess
import sys

HERE = Path(__file__).resolve().parent
PREFIX = 'solutions/edition-2/main/'
SUPPORT_PREFIX = PREFIX + 'evidence/ch10/support/'
ROLES = {'cli', 'gui', 'consumer', 'python', 'node', 'chrome', 'recorder'}
BUILD_ROLES = {'cli': ('', './cmd'), 'gui': ('gui', './cmd/ensemble-gui'),
               'consumer': ('evidence/ch10/support/consumer', '.')}


def require(value, message):
    if not value:
        raise ValueError(message)


def digest(data):
    return hashlib.sha256(data).hexdigest()


def file_hash(path):
    with Path(path).open('rb') as f:
        return hashlib.file_digest(f, 'sha256').hexdigest()


def canonical(value):
    return json.dumps(value, sort_keys=True, separators=(',', ':'), ensure_ascii=False).encode()


def git(repo, *args):
    return subprocess.check_output(['git', '-C', str(repo), *args])


def source_paths(repo, revision):
    require(re.fullmatch(r'[0-9a-f]{40}', revision), 'immutable source revision required')
    paths = git(repo, 'ls-tree', '-r', '--name-only', revision, '--', PREFIX).decode().splitlines()
    # Include all tracked runtime/docs/assets, all nested module manifests, and
    # this complete support tree. Prior evidence and binary artifacts stay out.
    return sorted(p for p in paths if '/evidence/' not in p or p.startswith(SUPPORT_PREFIX))


def historical(repo, revision, path):
    require(path.startswith(PREFIX) and '..' not in Path(path).parts, 'source path boundary')
    return git(repo, 'show', revision + ':' + path)


def dependency_paths(root):
    root = Path(root).resolve()
    require((root / 'package-lock.json').is_file(), 'browser dependency lock required')
    return sorted(str(p.relative_to(root)) for p in root.rglob('*') if p.is_file())


def check_sources(binding, repo):
    revision = binding['source_revision']
    expected = source_paths(repo, revision)
    require(expected and set(binding['sources']) == set(expected), 'incomplete source map')
    actual = set()
    for p in (Path(repo)/PREFIX).rglob('*'):
        relative = p.relative_to(repo).as_posix()
        if '__pycache__' in p.parts or '.git' in p.parts: continue
        if '/evidence/' in relative and not relative.startswith(SUPPORT_PREFIX): continue
        require(not p.is_symlink(), 'source symlink requires explicit build review')
        if p.is_file(): actual.add(relative)
    require(actual == set(expected), 'working source set mismatch')
    for path in expected:
        require(digest(historical(repo, revision, path)) == binding['sources'][path], 'source identity mismatch')
        require(file_hash(Path(repo) / path) == binding['sources'][path], 'working source identity mismatch')
    return expected


def preflight(binding, repo, binaries=None, support_root=HERE):
    revision=binding['source_revision']
    check_sources(binding, repo)
    support = {p: h for p, h in binding['sources'].items() if p.startswith(SUPPORT_PREFIX)}
    require(support and binding['support'] == support, 'incomplete support map')
    for path, sha in support.items():
        require(file_hash(Path(support_root) / path.removeprefix(SUPPORT_PREFIX)) == sha, 'support identity mismatch')
    modules = {p: h for p, h in binding['sources'].items() if Path(p).name in ('go.mod', 'go.sum')}
    require(modules and binding['modules'] == modules, 'module dependency identity mismatch')
    catalog = {p: h for p, h in support.items() if '/catalog/' in p or p.endswith('/bindings.json')}
    require(catalog and binding['catalog'] == catalog, 'catalog/binding identity mismatch')
    require(set(binding['binaries']) == ROLES, 'incomplete binary identities')
    for role, item in binding['binaries'].items():
        path = (binaries or {}).get(role, item['path'])
        require(file_hash(path) == item['sha256'], role + ' binary identity mismatch')
    require(file_hash(sys.executable) == binding['binaries']['python']['sha256'], 'running interpreter mismatch')
    require(set(binding['builds']) == set(BUILD_ROLES), 'incomplete build associations')
    for role, (directory, target) in BUILD_ROLES.items():
        build = binding['builds'][role]
        require(build['source_revision'] == revision and build['binary_sha256'] == binding['binaries'][role]['sha256'], 'build association mismatch')
        require(build['module'] == PREFIX + directory and build['target'] == target and build['exit'] == 0, 'build command mismatch')
        require(build['modules'] == modules and build['go_version'] and build['module_graph'] and build['binary_build_info'], 'incomplete build dependency evidence')
    dep = binding['browser_dependencies']
    require(set(dep['files']) == set(dependency_paths(dep['root'])), 'incomplete browser dependencies')
    for path, sha in dep['files'].items():
        require(file_hash(Path(dep['root']) / path) == sha, 'browser dependency identity mismatch')
    require(binding['schedule_sha256'] == support[SUPPORT_PREFIX + 'schedule.json'], 'schedule identity mismatch')
    return binding


def check_launch(binding, launch):
    require(launch['binding_sha256'] == digest(canonical(binding)), 'launch binding mismatch')
    require(launch['source_revision'] == binding['source_revision'], 'launch source mismatch')
    require(launch['vendor'] in ('anthropic', 'openai', 'gemini'), 'launch vendor mismatch')
    require(launch['mode'] in ('local', 'live'), 'launch mode mismatch')
    require(set(launch['environment']) <= {'LLM_VENDOR', 'LLM_MODEL', 'LLM_RESOLVED_MODEL', 'LLM_BASE_URL', 'LLM_SKILLS_DIR', 'LLM_PRIMARY_SKILL', 'EN_DISABLE_STREAMING'}, 'unsafe launch environment')
    require(launch['environment'].get('LLM_VENDOR') == launch['vendor'], 'launch vendor/environment mismatch')
    require(launch['environment'].get('LLM_MODEL') == launch['model'], 'launch model mismatch')
    require(launch['role'] in BUILD_ROLES, 'launch executable role mismatch')
    require(launch['command'][0] == binding['binaries'][launch['role']]['path'], 'launch executable path mismatch')
    require(launch['command_sha256'] == digest(canonical(launch['command'])), 'launch command mismatch')
    require(launch['identity_sha256'] == digest(canonical({k:v for k,v in launch.items() if k != 'identity_sha256'})), 'launch identity mismatch')


def verify_run(binding, launch, originals, repo, run, destination):
    preflight(binding, repo)
    check_launch(binding, launch)
    require(originals, 'empty originals map')
    run = Path(run).resolve()
    for name, sha in originals.items():
        path = (run / name).resolve()
        require(path.is_relative_to(run), 'original path boundary')
        require(file_hash(path) == sha, 'original receipt identity mismatch')
    # This is an identity receipt, NOT a feature-acceptance score.
    with Path(destination).open('x') as f:
        json.dump({'identity_verified': True, 'chapter_accepted': False,
                   'source_revision': binding['source_revision'], 'originals': originals}, f, indent=2)
        f.write('\n')


if __name__ == '__main__':
    import argparse
    p = argparse.ArgumentParser()
    p.add_argument('binding', type=Path); p.add_argument('--repo', required=True, type=Path)
    p.add_argument('--launch', type=Path)
    a = p.parse_args(); b = json.loads(a.binding.read_text())
    preflight(b, a.repo)
    if a.launch: check_launch(b, json.loads(a.launch.read_text()))
    print('identity preflight passed; no evidence modified')
