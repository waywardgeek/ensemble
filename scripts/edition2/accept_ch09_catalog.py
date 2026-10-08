#!/usr/bin/env python3
"""Contract-first Chapter 9 catalog/format bounds, actual CLI; no model calls.

This is a focused supplement, not full Chapter 9 acceptance. Public scalar,
graph mutation and runtime deletion coverage live in separate checks.
"""
import argparse
import copy
import json
from pathlib import Path
import subprocess
import tempfile

from accept_ch09 import ROOT, definition, environment, sha

LIMIT = 65536
TOTAL = 8388608
TOOLS = '- load_skill\n- read_file\n- unload_skill'


def primary(body=b'Identity.\n'):
    return definition('base','primary','',tools='read_file') + body


def sized(source, length):
    assert len(source) <= length
    return source + b'x' * (length-len(source))


def aggregate(length):
    # 129 individually valid files let the total exceed 8 MiB by exactly one
    # without inadvertently hitting the independent 64 KiB source guard first.
    files={'base':primary()}
    files.update({f's{i:03}':definition(f's{i:03}','loadable','') for i in range(128)})
    remaining=length-sum(map(len,files.values()))
    for name in files:
        added=min(LIMIT-len(files[name]),remaining)
        files[name]+=b'x'*added;remaining-=added
    assert remaining==0 and sum(map(len,files.values()))==length
    assert all(len(x)<=LIMIT for x in files.values())
    return files


def cases():
    result=[]
    def add(name,source=None,accepted=True,body=None,files=None,filesystem=None,environment_override=None):
        result.append({'id':name,'files':files if files is not None else {'base':source if source is not None else primary()},
                       'accepted':accepted,'body':body,'filesystem':filesystem,'environment':environment_override or {}})
    base=primary()
    add('body-leading-blank-horizontal-rule',primary(b'\nA\n---\nB\n'),body='\nA\n---\nB\n')
    add('body-crlf-exact',primary(b'A\n---\nB\n').replace(b'\n',b'\r\n'),body='A\r\n---\r\nB\r\n')
    add('body-no-final-newline',primary(b'last'),body='last')
    add('source-exact-65536',sized(base,LIMIT))
    add('source-one-over',sized(base,LIMIT+1),False)
    for n in (LIMIT,LIMIT+1):
        repeats,tail=divmod(n,len(TOOLS.encode()))
        source=primary(b'$TOOLS'*repeats+b'x'*tail)
        assert len(source)<=LIMIT
        add('render-'+('exact-65536' if n==LIMIT else 'one-over'),source,n==LIMIT,TOOLS*repeats+'x'*tail if n==LIMIT else None)
    for n in (256,257):
        fs={'base':base,**{f's{i:03}':definition(f's{i:03}','loadable','') for i in range(n-1)}}
        add('catalog-count-'+str(n),accepted=n==256,files=fs)
    add('catalog-bytes-exact',files=aggregate(TOTAL))
    add('catalog-bytes-one-over',accepted=False,files=aggregate(TOTAL+1))
    for n in (256,257):
        add('description-ascii-'+str(n),base.replace(b'Base scratch files',b'D'*n),n==256)
    for n,valid in [(127,True),(128,False)]:
        # One ASCII leading letter plus two-byte characters: 255/257 bytes.
        add('description-utf8-'+str(1+2*n),base.replace(b'Base scratch files',b'A'+('é'*n).encode()),valid)
    for name,description,valid in [('quoted-numeric',b'"19"',True),('quoted-reserved',b"'true'",True),
                                  ('single-apostrophe',b"'Reader''s guide'",True),('quoted-hash',b'"A # B"',True),
                                  ('unquoted-colon',b'A: B',True),('unquoted-true',b'TrUe',False),('unquoted-false',b'false',False),
                                  ('unquoted-null',b'NULL',False),('unquoted-comment',b'A # B',False),
                                  ('quoted-trailing-comment',b'"A" # B',False),('quoted-trailing-token',b'"A" junk',False),
                                  ('decoded-newline',b'"A\\nB"',False),('decoded-nul',b'"A\\u0000B"',False)]:
        add('description-'+name,base.replace(b'Base scratch files',description),valid)
    for name,list_value,valid in [('quoted-scalar',b'"read_file unload_skill"',True),('single-scalar',b"'read_file unload_skill'",True),
                                 ('quoted-block',b'\n  - "read_file"\n  - unload_skill',True),('empty-sequence',b'[]',True),
                                 ('comma',b'read_file, unload_skill',False),('tab',b'read_file\tunload_skill',False),
                                 ('duplicate-block',b'\n  - read_file\n  - read_file',False),('one-space',b'\n - read_file',False),
                                 ('three-spaces',b'\n   - read_file',False),('empty-item',b'\n  - ',False),
                                 ('mixed',b'read_file\n  - unload_skill',False),('flow-nonempty',b'[read_file]',False)]:
        add('tools-'+name,base.replace(b'tools: read_file',b'tools: '+list_value),valid)
    for name,line in [('comment',b'# comment\n'),('anchor',b'x: &anchor text\n'),('nested',b'  nested: value\n'),
                      ('block-scalar',b'description: |\n'),('tag',b'description: !!str hello\n')]:
        source=base.replace(b'description: Base scratch files\n',line) if name in ('block-scalar','tag') else base.replace(b'name: base\n',b'name: base\n'+line)
        add('unsupported-'+name,source,False)
    for token in (b'$',b'$5',b'${}',b'${TOOLS',b'$tools',b'$TOOLS_SUFFIX',b'$_PRIVATE',b'${TOOLS}tail$'):
        add('variable-invalid-'+str(len(result)),primary(token),False)
    add('dollar-escape',primary(b'$$5 $$TOOLS ${TOOLS}\n'),body='$5 $TOOLS '+TOOLS+'\n')
    add('ordinary-root-file-ignored',filesystem='root-file')
    add('missing-skill-in-child',accepted=False,filesystem='missing-skill')
    add('symlink-child',accepted=False,filesystem='symlink-child')
    add('symlink-file',accepted=False,filesystem='symlink-file')
    add('unreachable-bad-edge-accepted',files={'base':base,'hidden':definition('hidden','loadable','',extra='depends: absent\n')})
    add('unreachable-uninstalled-tool-accepted',files={'base':base,'hidden':definition('hidden','loadable','',tools='fictional_handler')})
    for name,changes in [('only-dir',{'LLM_PRIMARY_SKILL':None}),('only-primary',{'LLM_SKILLS_DIR':None}),
                         ('blank-dir',{'LLM_SKILLS_DIR':''}),('blank-primary',{'LLM_PRIMARY_SKILL':''}),('unknown-primary',{'LLM_PRIMARY_SKILL':'absent'})]:
        add('environment-'+name,accepted=False,environment_override=changes)
    return result


def evaluate(case, exit_code, stderr, events):
    initialized=[e for e in events if e.get('type')=='skills_initialized']
    if not case['accepted']:
        return exit_code!=0 and bool(stderr.strip()) and not initialized
    if exit_code!=0 or len(initialized)!=1: return False
    if case['body'] is not None:
        material=[x for x in initialized[0]['skills']['activated'] if x['name']=='base']
        return len(material)==1 and material[0]['body']==case['body'] and material[0]['sha256']==sha(case['body'].encode())
    return True


def run(binary):
    results=[]
    for case in cases():
        with tempfile.TemporaryDirectory(prefix='ch09-catalog-') as directory:
            work=Path(directory)
            for identity,data in case['files'].items():
                path=work/'catalog'/identity/'SKILL.md';path.parent.mkdir(parents=True);path.write_bytes(data)
            kind=case['filesystem']
            if kind=='root-file': (work/'catalog/readme.txt').write_text('Ignored ordinary file\n')
            if kind=='missing-skill': (work/'catalog/incomplete').mkdir()
            if kind=='symlink-child':
                (work/'outside').mkdir();(work/'outside/SKILL.md').write_bytes(definition('linked','loadable',''))
                (work/'catalog/linked').symlink_to(work/'outside',target_is_directory=True)
            if kind=='symlink-file':
                (work/'outside.md').write_bytes(case['files']['base']);path=work/'catalog/base/SKILL.md';path.unlink();path.symlink_to(work/'outside.md')
            env=environment(work,'anthropic','http://127.0.0.1:1')
            env.update(LLM_SKILLS_DIR='catalog',LLM_PRIMARY_SKILL='base')
            for name,value in case['environment'].items():
                if value is None: env.pop(name,None)
                else: env[name]=value
            child=subprocess.run([str(binary),'protocol'],input='',text=True,capture_output=True,cwd=work,env=env,timeout=20)
            log=work/'session.jsonl';events=[json.loads(x) for x in log.read_text().splitlines()] if log.exists() else []
            assert not any(e.get('type')=='request_sent' for e in events), 'startup unexpectedly requested a model'
            results.append({'id':case['id'],'passed':evaluate(case,child.returncode,child.stderr,events),'expected_accept':case['accepted'],
                            'exit':child.returncode,'stderr':child.stderr,'source_count':len(case['files']),
                            'source_bytes':sum(map(len,case['files'].values())),'initializers':sum(e.get('type')=='skills_initialized' for e in events)})
    return results


def self_test():
    inputs=cases();ids=[x['id'] for x in inputs];assert len(ids)==len(set(ids))
    for c in inputs:
        assert c['files'] and all(isinstance(v,bytes) for v in c['files'].values())
    exact=next(x for x in inputs if x['id']=='catalog-bytes-exact');over=next(x for x in inputs if x['id']=='catalog-bytes-one-over')
    assert sum(map(len,exact['files'].values()))==TOTAL and sum(map(len,over['files'].values()))==TOTAL+1
    c=next(x for x in inputs if x['id']=='dollar-escape')
    good=[{'type':'skills_initialized','skills':{'activated':[{'name':'base','body':c['body'],'sha256':sha(c['body'].encode())}]}}]
    assert evaluate(c,0,'',good)
    for mutate in ('body','sha256'):
        bad=copy.deepcopy(good);bad[0]['skills']['activated'][0][mutate]='wrong';assert not evaluate(c,0,'',bad)
    invalid=next(x for x in inputs if not x['accepted']);assert evaluate(invalid,1,'safe error',[])
    assert not evaluate(invalid,1,'safe error',good)
    return {'label':'fixture sizes and canned acceptance predicates only; no runtime positive','cases':len(inputs),'predicate_negative_controls':3,'passed':True}


def main():
    parser=argparse.ArgumentParser(description=__doc__);parser.add_argument('binary',nargs='?',type=Path);parser.add_argument('--self-test',action='store_true');parser.add_argument('--receipt',type=Path);args=parser.parse_args()
    paths=[Path(__file__),ROOT/'scripts/edition2/accept_ch09.py',ROOT/'book/edition-2/chapter-09.md']
    if args.binary: args.binary=args.binary.resolve();paths.append(args.binary)
    hashes={str(p):sha(p.read_bytes()) for p in paths}
    if args.self_test: result=self_test();code=0
    else:
        if not args.binary:parser.error('CLI_BINARY or --self-test required')
        results=run(args.binary);result={'checks':results,'passed':sum(x['passed'] for x in results),'total':len(results)};code=int(result['passed']!=result['total'])
    assert hashes=={str(p):sha(p.read_bytes()) for p in paths},'checker, contract or executable changed during run'
    result['input_sha256']=hashes
    if args.receipt:args.receipt.write_text(json.dumps(result,indent=2)+'\n')
    print(json.dumps(result,indent=2));raise SystemExit(code)


if __name__=='__main__':main()
