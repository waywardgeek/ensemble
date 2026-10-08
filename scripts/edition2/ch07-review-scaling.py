#!/usr/bin/env python3
"""Fixed-fragment projection allocation control, including deliberate repeated copying.

Cumulative allocation only; no peak-memory, browser-latency or provider claim.
"""
import argparse
import hashlib
import json
from pathlib import Path
import re
import shutil
import statistics
import tempfile
from accept_ch05 import command
from accept_ch07_components import install,HERE


def measure(root):
    result=command(['go','test','-run','^$','-bench','^BenchmarkC7ProjectionScaling$','-benchtime=1x','-count=3','./internal/llm'],root,timeout=120)
    values={65536:[],131072:[]}
    for line in result['stdout'].splitlines():
        match=re.match(r'BenchmarkC7ProjectionScaling/bytes=(65536|131072)-\d+\s+',line)
        if match:
            fields=line.split();values[int(match[1])].append(int(fields[fields.index('B/op')-1]))
    complete=result['exit']==0 and all(len(v)==3 for v in values.values())
    ratio=statistics.median(values[131072])/statistics.median(values[65536]) if complete else None
    return dict(command=result,samples=values,ratio=ratio,complete=complete)


def evaluate(source):
    with tempfile.TemporaryDirectory(prefix='ch07-projection-scaling-') as directory:
        root=Path(directory)/'source';shutil.copytree(source,root,ignore=shutil.ignore_patterns('.git','evidence','node_modules','__pycache__'))
        hashes={str(p.relative_to(root)):hashlib.sha256(p.read_bytes()).hexdigest() for p in sorted(root.rglob('*')) if p.is_file()}
        install(root);positive=measure(root)
        target=root/'internal/llm/watch.go';text=target.read_text()
        old='append(a.partials[o.PartID][o.Channel], o.Text...)';new='append(append([]byte(nil), a.partials[o.PartID][o.Channel]...), o.Text...)'
        if text.count(old)!=1:raise ValueError('projection append mutation anchor not unique')
        target.write_text(text.replace(old,new));negative=measure(root)
        passed=(positive['complete'] and positive['ratio']<2.8 and negative['complete'] and negative['ratio']>3.2)
        return dict(scope=__doc__,source=str(source),source_files=hashes,
                    checker_files={p.name:hashlib.sha256(p.read_bytes()).hexdigest() for p in [Path(__file__),HERE/'ch07_internal_test.go.txt',HERE/'ch06_internal_test.go.txt']},
                    passed=passed,positive=positive,negative=negative,
                    control={'old':old,'new':new,'positive_ratio_max':2.8,'negative_ratio_min':3.2,'fragment_bytes':128,'sample_sizes':[65536,131072]})


if __name__=='__main__':
    p=argparse.ArgumentParser(description=__doc__);p.add_argument('source',type=Path);a=p.parse_args();r=evaluate(a.source.resolve(strict=True));print(json.dumps(r,indent=2));raise SystemExit(0 if r['passed'] else 1)
