import pathlib,shutil,subprocess,tempfile,json
base=pathlib.Path('/Users/bill/projects/ensemble/solutions/edition-2');source=base/'main';root=pathlib.Path(tempfile.mkdtemp(prefix='ensemble-ch14-mutants-'));tree=root/'main'
shutil.copytree(source,tree,ignore=shutil.ignore_patterns('node_modules','*.log','cr','solutions'))
(tree/'gui/node_modules').symlink_to(source/'gui/node_modules',target_is_directory=True)
changes=[
 ('persona-type','gui/virtual-user.mjs','await this.page.locator("#input").fill(args.text);','/* no input */'),
 ('persona-wait','gui/virtual-user.mjs','(turn) => window.ensembleTurn >= turn,','(turn) => true,'),
 ('persona-hear-cursor','gui/virtual-user.mjs','this.cursor += entries.length;','/* repeats transcript */'),
 ('persona-client-extra-tools','gui/cmd/virtual-user/main.go','root := ensemble.New(os.Stderr)','cfg.BuiltinTools = true\n\troot := ensemble.New(os.Stderr)'),
]

results=[]
try:
 for name,relative,old,new in changes:
  p=tree/relative;original=p.read_text();assert old in original,name;p.write_text(original.replace(old,new))
  run=subprocess.run(['node','--test','speech.test.mjs'],cwd=tree/'gui',capture_output=True,text=True,timeout=65)
  p.write_text(original);record={'mutation':name,'applied':True,'exit':run.returncode,'output':run.stdout+run.stderr};results.append(record)
  (base/'evidence/ch14/persona-mutations.json').write_text(json.dumps(results,indent=2));print(name,run.returncode,flush=True)
finally:shutil.rmtree(root)
