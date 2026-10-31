import pathlib,shutil,subprocess,tempfile,json
base=pathlib.Path('/Users/bill/projects/ensemble/solutions/edition-2');source=base/'main';root=pathlib.Path(tempfile.mkdtemp(prefix='ensemble-ch14-mutants-'));tree=root/'main'
shutil.copytree(source,tree,ignore=shutil.ignore_patterns('node_modules','*.log','cr','solutions'))
(tree/'gui/node_modules').symlink_to(source/'gui/node_modules',target_is_directory=True)
changes=[
 ('fragments','gui/web/gui/speech.js','this.drain(id, false);','this.flush(id);'),
 ('wrapped-newline','gui/web/gui/speech.js','if (c === "\\n" &&','if (c === "\\n") return i + 1;\n    if (c === "\\n" &&'),
 ('final-tail','gui/web/gui/speech.js','this.speech.enqueue(text, part.source);','/* omit unfinished tail */'),
 ('fence-before-split','gui/web/gui/speech.js','part.text.replace(/```[^\\n]*\\n[\\s\\S]*?```/g, " code block ")','part.text'),
 ('emphasis-inline','gui/web/gui/speech.js','.replace(/[*_`]/g, "")',''),
 ('link-filter','gui/web/gui/speech.js','.replace(/!?\\[([^\\]]*)\\]\\([^)]*\\)/g, "$1")',''),
 ('identifiers','gui/web/gui/speech.js','.replace(/([a-z\\d])([A-Z])/g, "$1 $2")',''),
 ('source','gui/web/gui/speech.js','part.source','"text"'),
 ('audio-bypass','gui/web/gui/app.js','if (!this.audio || !voice.checked) return;','if (!voice.checked) return;'),
 ('whole-part','gui/web/gui/app.js','card.streamed ? "" : card.text','""'),
 ('stream-duplicate','gui/web/gui/app.js','card.streamed ? "" : card.text','card.text'),
 ('error-speech','gui/web/gui/app.js','speech.enqueue(message.error, "error");','/* omit spoken error */'),
 ('silent-tool-result','gui/web/gui/app.js','// Tool output stays silent automatically','speech.enqueue(message.result, "tool_result");\n      // Tool output stays silent automatically'),
 ('escape-with-input','gui/web/gui/app.js','if (event.key === "Escape") speech.cancel();','if (event.key === "Escape" && input.value === "") speech.cancel();'),
 ('blind-registry','gui/virtual-user.mjs','const tools = [','const tools = [{ name: "gui_snapshot", inputSchema: {type: "object", properties: {}} },'),
 ('blind-dispatch','gui/virtual-user.mjs','if (!tools.some(tool => tool.name === name))','if (name === "gui_snapshot") return "DOM contents";\n    if (!tools.some(tool => tool.name === name))'),
]
results=[]
try:
 for name,relative,old,new in changes:
  p=tree/relative;original=p.read_text();assert old in original,name;p.write_text(original.replace(old,new))
  run=subprocess.run(['node','--test','speech.test.mjs'],cwd=tree/'gui',capture_output=True,text=True,timeout=65)
  p.write_text(original);record={'mutation':name,'applied':True,'exit':run.returncode,'output':run.stdout+run.stderr};results.append(record)
  (base/'evidence/ch14/mutations.json').write_text(json.dumps(results,indent=2));print(name,run.returncode,flush=True)
finally:shutil.rmtree(root)
