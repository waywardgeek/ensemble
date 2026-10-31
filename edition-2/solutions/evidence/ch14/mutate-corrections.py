import pathlib,shutil,subprocess,tempfile,json
base=pathlib.Path('/Users/bill/projects/ensemble/solutions/edition-2');source=base/'main';root=pathlib.Path(tempfile.mkdtemp(prefix='ensemble-ch14-mutants-'));tree=root/'main'
shutil.copytree(source,tree,ignore=shutil.ignore_patterns('node_modules','*.log','cr','solutions'))
(tree/'gui/node_modules').symlink_to(source/'gui/node_modules',target_is_directory=True)
changes=[
 ('whole-part','gui/web/gui/app.js','card.streamed ? "" : card.text','""'),
 ('stream-duplicate','gui/web/gui/app.js','card.streamed ? "" : card.text','card.text'),
 ('blind-registry','gui/virtual-user.mjs','const tools = [','const tools = [{ name: "gui_snapshot", inputSchema: {type: "object", properties: {}} },'),
 ('blind-dispatch','gui/virtual-user.mjs','if (!tools.some((tool) => tool.name === name))','if (name === "gui_snapshot") return "DOM contents";\n    if (!tools.some((tool) => tool.name === name))'),
]

results=[]
try:
 for name,relative,old,new in changes:
  p=tree/relative;original=p.read_text();assert old in original,name;p.write_text(original.replace(old,new))
  run=subprocess.run(['node','--test','speech.test.mjs'],cwd=tree/'gui',capture_output=True,text=True,timeout=65)
  p.write_text(original);record={'mutation':name,'applied':True,'exit':run.returncode,'output':run.stdout+run.stderr};results.append(record)
  (base/'evidence/ch14/mutation-corrections.json').write_text(json.dumps(results,indent=2));print(name,run.returncode,flush=True)
finally:shutil.rmtree(root)
