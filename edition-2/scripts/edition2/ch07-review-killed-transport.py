#!/usr/bin/env python3
"""Actual Chrome page reload over real WebSocket using frozen real job events, with no paid calls."""
import argparse,hashlib,importlib.util,json,subprocess,tempfile
from pathlib import Path
from accept_ch05 import command
HERE=Path(__file__).resolve().parent
spec=importlib.util.spec_from_file_location('retained',HERE/'ch07-review-retained.py');helper=importlib.util.module_from_spec(spec);spec.loader.exec_module(helper)
p=argparse.ArgumentParser(description=__doc__);p.add_argument('revision');a=p.parse_args();log=HERE.parents[1]/'solutions/edition-2/main/evidence/ch07/browser-anthropic-r1/session.log'
with tempfile.TemporaryDirectory(prefix='ch07-killed-browser-') as tmp:
 root=Path(tmp)/'source';revision,hashes=helper.extract(a.revision,root);target=root/'gui/cmd/review/main.go';target.parent.mkdir(parents=True);target.write_bytes((HERE/'ch07-review-killed-server.go.txt').read_bytes());binary=Path(tmp)/'server';workspace=Path(tmp)/'work';workspace.mkdir()
 build=command(['go','build','-o',binary,'./cmd/review'],root/'gui',120);result={}
 if build['exit']==0:
  process=subprocess.Popen([str(binary),str(log),str(workspace)],stdout=subprocess.PIPE,stderr=subprocess.PIPE,text=True)
  try:
   launch=json.loads(process.stdout.readline());result=command(['node',HERE/'ch07-review-killed-browser.cjs',launch['url'],launch['target_key']],HERE.parents[1],120)
  finally:
   process.terminate();process.wait(timeout=5);stderr=process.stderr.read();process.stdout.close();process.stderr.close()
 receipt=dict(scope=__doc__,source_revision=revision,source_files=hashes,original_log_sha256=hashlib.sha256(log.read_bytes()).hexdigest(),checker_files={x.name:hashlib.sha256(x.read_bytes()).hexdigest() for x in [Path(__file__),HERE/'ch07-review-killed-server.go.txt',HERE/'ch07-review-killed-browser.cjs']},build=build,command=result,passed=build['exit']==0 and result.get('exit')==0)
 print(json.dumps(receipt,indent=2));raise SystemExit(0 if receipt['passed'] else 1)
