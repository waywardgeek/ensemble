import pathlib,shutil,tempfile,subprocess,os,json
root=pathlib.Path('/Users/bill/projects/ensemble/solutions/edition-2'); results=[]
cases=[
 ('click effect','mcp.js','element.click()','void element'),
 ('input change event','mcp.js','element.dispatchEvent(new Event("change", { bubbles: true }));','// change notification removed'),
 ('byte cap','mcp.js','4096 - 128','8192 - 128'),
 ('truncation disclosure','mcp.js','${elided} characters elided; ${omitted} artifacts omitted','0 characters elided; 0 artifacts omitted'),
 ('control state','mcp.js','"value", "checked", "disabled"','"value", "disabled"'),
 ('actual TTS queue','app.js','return active.concat(this.queue.map','return [].concat([].map'),
]
with tempfile.TemporaryDirectory(prefix='ensemble-ch12-browser-mutants-') as directory:
 web=pathlib.Path(directory)/'web';shutil.copytree(root/'main/gui/web/gui',web)
 for claim,path,before,after in cases:
  file=web/path;original=file.read_text();assert before in original,claim;file.write_text(original.replace(before,after,1))
  capture=pathlib.Path(directory)/'result.json'
  run=subprocess.run(['node',str(root/'evidence/ch12/browser-check.mjs')],env={**os.environ,'CH12_WEB':str(web),'CH12_BROWSER_OUT':str(capture)},capture_output=True,text=True,timeout=40)
  result=json.loads(capture.read_text());results.append({'claim':claim,'mutation':before+' -> '+after,'exit':run.returncode,'result':result['result'],'error':result.get('error'),'checks':result['checks']})
  file.write_text(original);print(claim,run.returncode,result.get('error'),flush=True)
(root/'evidence/ch12/browser-mutations.json').write_text(json.dumps(results,indent=2)+'\n')
