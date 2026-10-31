"""Two targeted browser-handler defects; only disposable static assets change."""
import json,os,pathlib,shutil,subprocess,tempfile
root=pathlib.Path(__file__).resolve().parents[4];out=pathlib.Path(__file__).resolve().parent
results=[]
for name,old,new in [('browser-toggle','field === "disabled" ? !input.checked','field === "disabled" ? input.checked'),('browser-budget',': Number(input.value)',': 0')]:
 scratch=pathlib.Path(tempfile.mkdtemp(prefix='ch16-assets-'))
 try:
  shutil.copytree(root/'solutions/edition-2/main/gui/web/gui',scratch,dirs_exist_ok=True)
  p=scratch/'app.js';s=p.read_text();assert old in s;p.write_text(s.replace(old,new,1))
  env=os.environ.copy();env['CH16_GUI_DIR']=str(scratch)
  run=subprocess.run(['node',str(out/'browser_settings.mjs')],cwd=root,env=env,text=True,capture_output=True,timeout=45)
  text=run.stdout+run.stderr;(out/('mutation-'+name+'.txt')).write_text(text)
  results.append({'claim':name,'old':old,'new':new,'exit':run.returncode,'intended_failure':run.returncode!=0 and 'browser changes not persisted' in text})
 finally:shutil.rmtree(scratch)
(out/'browser-mutations.json').write_text(json.dumps(results,indent=2))
print(json.dumps(results))
