#!/usr/bin/env python3
"""Retain the Chapter 8 adaptations and extend one old test double's API.

No assertion or production code changes. Historical fixture sources stay intact.
"""
import hashlib
from pathlib import Path
from ch08_retained_adapters import prepare as previous

def prepare(destination,source=None):
    result=previous(destination,source)
    name='ch05_stale_test.go.txt';path=destination/name;text=path.read_text()
    old='func (auditLateRegistry) Declarations() []common.ToolDefinition { return nil }'
    new=old+'\nfunc (auditLateRegistry) Management(string) bool { return false }'
    assert text.count(old)==1,'retained non-management fixture anchor changed'
    path.write_text(text.replace(old,new))
    result['changes'].append(dict(file=name,old=old,new=new,count=1))
    result['adapted_files'][name]=hashlib.sha256(path.read_bytes()).hexdigest()
    result['chapter9_adapter_sha256']=hashlib.sha256(Path(__file__).read_bytes()).hexdigest()
    return result
