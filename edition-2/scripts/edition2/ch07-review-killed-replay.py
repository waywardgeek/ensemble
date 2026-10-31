#!/usr/bin/env python3
"""Public snapshot check using preserved real lifecycle events, without executing tools or models."""
import argparse,hashlib,json,os
from pathlib import Path
import tempfile
import importlib.util
HERE=Path(__file__).resolve().parent
spec=importlib.util.spec_from_file_location('retained',HERE/'ch07-review-retained.py');helper=importlib.util.module_from_spec(spec);spec.loader.exec_module(helper)
from accept_ch05 import command
p=argparse.ArgumentParser(description=__doc__);p.add_argument('revision');p.add_argument('--snapshot',required=True,type=Path);a=p.parse_args()
log=HERE.parents[1]/'solutions/edition-2/main/evidence/ch07/browser-anthropic-r1/session.log'
with tempfile.TemporaryDirectory(prefix='ch07-killed-review-') as tmp:
 root=Path(tmp)/'source';revision,hashes=helper.extract(a.revision,root)
 (root/'independent_review_killed_test.go').write_bytes((HERE/'ch07-review-killed-test.go.txt').read_bytes())
 os.environ['C7_REAL_LOG']=str(log);os.environ['C7_SNAPSHOT']=str(a.snapshot.resolve())
 result=command(['go','test','-race','-count=1','-run','^TestC7ReviewKilledWindow$','.'],root,120)
 receipt=dict(scope=__doc__,source_revision=revision,source_files=hashes,original_log=str(log),original_log_sha256=hashlib.sha256(log.read_bytes()).hexdigest(),checker_files={x.name:hashlib.sha256(x.read_bytes()).hexdigest() for x in [Path(__file__),HERE/'ch07-review-killed-test.go.txt']},passed=result['exit']==0,command=result)
 print(json.dumps(receipt,indent=2));raise SystemExit(0 if receipt['passed'] else 1)
