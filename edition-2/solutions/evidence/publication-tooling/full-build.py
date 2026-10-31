import pathlib,subprocess,time,json,hashlib,tempfile
from PIL import Image
root=pathlib.Path.cwd();out=root/'solutions/evidence/publication-tooling';rows=[]
inputs=list((root/'book').glob('chapter-[0-9][0-9].md'))+[root/'book'/x for x in ['preface.md','epilogue.md','note-from-bill.md','front-matter.md','front-matter-print.md','print-header.html','epub.css','cover.png']]
binding={str(p.relative_to(root)):hashlib.sha256(p.read_bytes()).hexdigest() for p in inputs}
def run(name,cmd):
 t=time.monotonic();p=subprocess.run(cmd,capture_output=True,text=True,timeout=600);(out/(name+'.txt')).write_text(p.stdout+p.stderr);row=dict(name=name,command=cmd,exit=p.returncode,seconds=round(time.monotonic()-t,3));rows.append(row);(out/'full-build.json').write_text(json.dumps(dict(inputs=binding,commands=rows),indent=2)+'\n');print(row,flush=True);assert p.returncode==0,p.stdout+p.stderr
run('full-markdown',['make','book'])
run('full-epub',['scripts/build-ebook.sh'])
run('full-print',['python3','book/build-print.py'])
with tempfile.TemporaryDirectory(prefix='ensemble-cover-') as tmp:
 cover=pathlib.Path(tmp)/'cover.pdf';Image.open(root/'book/cover.png').convert('RGB').resize((1800,2700),Image.Resampling.LANCZOS).save(cover,'PDF',resolution=300.0)
 run('full-cover-pdf',['pdfunite',str(cover),'book/The Art of Building AI Coding Agents - Print Interior.pdf','book/The Art of Building AI Coding Agents.pdf'])
assert binding=={str(p.relative_to(root)):hashlib.sha256(p.read_bytes()).hexdigest() for p in inputs},'publication inputs changed during build'
