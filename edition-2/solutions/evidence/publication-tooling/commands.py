import pathlib,subprocess,time,json,os
root=pathlib.Path.cwd();out=root/'solutions/evidence/publication-tooling';rows=[]
def run(name,cmd,cwd=root,input=None):
 t=time.monotonic();p=subprocess.run(cmd,cwd=cwd,input=input,text=True,capture_output=True,timeout=90);(out/(name+'.txt')).write_text(p.stdout+p.stderr);row=dict(name=name,command=cmd,cwd=str(cwd.relative_to(root)) or '.',exit=p.returncode,seconds=round(time.monotonic()-t,3));rows.append(row);print(row,flush=True);return p
run('current-ch01',['go','run','./cmd/grade-edition2','-ch','1','-json'])
run('current-ch26',['go','run','./cmd/grade-edition2','-ch','26','-json'])
run('archive-ch01',['go','run','./cmd/grade','-ch','1','-json','./solutions/ch01'],root/'edition-1')
run('tooling-vet',['go','vet','./cmd/grade-edition2','./cmd/fakevendor','./internal/gradeedition2'])
run('current-ch01-regression',['go','test','./internal/gradeedition2','-run','^TestChapter1VersionAndPreservedChecks$','-count=1'])
run('tooling-build',['go','build','./cmd/grade-edition2','./cmd/fakevendor'])
run('exporter-tests',['python3','-m','unittest','discover','-s','scripts/edition2','-p','test_export_snapshot.py'])
run('fakevendor-ch03',['go','run','./cmd/fakevendor','-ch','3','chat'],input='Show the local tool loop.\n')
(out/'commands.json').write_text(json.dumps(rows,indent=2)+'\n')
assert all(x['exit']==0 for x in rows)
