import subprocess,pathlib,tempfile,shutil,json,difflib,time
root=pathlib.Path.cwd();out=root/'solutions/edition-2/evidence/ch15/enhancement14';source=root/'solutions/edition-2/ch14'
cases=[
 ('wrong-effective-target','internal/agent/context.go','target := r.owner.Config().ContextTarget','target := 400000','skill-survives-the-ladder'),
 ('read-empty-success','internal/tools/tools.go','return result, nil','return "", nil','skill-survives-the-ladder'),
 ('handoff-lands-mid-batch','internal/history/reduce.go','if !toolsComplete(c.Dialogue) {\n\t\tc.Held','if false {\n\t\tc.Held','micro-handoff-shape'),
 ('inline-before-ephemera','internal/engine/anthropic.go','declarations = append(declarations, blocks...)','appendMessage("system", blocks)','frozen-prefix'),
 ('addition-ordinary-text','internal/engine/anthropic.go','"type": "tool_addition"','"type": "text"','frozen-prefix'),
 ('removal-ordinary-text','internal/engine/anthropic.go','"type": "tool_removal"','"type": "text"','frozen-prefix'),
 ('mcp-empty-result','internal/mcp/client.go','return joined, nil','if name == "gui_input" { return "", nil }; return joined, nil','frozen-prefix'),
 ('mcp-wrong-arguments','internal/mcp/mcp.go','return c.call(call.Job().Context(), info.Name, args)','if info.Name == "gui_input" { args = json.RawMessage(`{"selector":"#search-box","text":"WRONG"}`) }; return c.call(call.Job().Context(), info.Name, args)','frozen-prefix'),
 ('automatic-content-empty','internal/mcp/client.go','return joined, nil','if name == "gui_snapshot" || name == "tts_queue" { return "", nil }; return joined, nil','frozen-prefix'),
 ('ephemera-retained','internal/history/reduce.go','c.Ephemera = nil // consumption','// c.Ephemera = nil // consumption','frozen-prefix'),
 ('startup-prefix-rewritten','internal/engine/engine.go','return *c.StartupTools','return c.Tools','frozen-prefix'),
 ('nonzero-after-valid-exit','gui/cmd/main.go','func main() {','func main() {\n defer os.Exit(17)','skill-survives-the-ladder'),
 ('split-system-positive','internal/engine/anthropic.go','return json.Marshal(struct {','var split []anthropicMessage\n for _,m := range messages { if m.Role == "system" { for _,b := range m.Content { split=append(split,anthropicMessage{Role:"system",Content:[]json.RawMessage{b}}) } } else { split=append(split,m) } }; messages=split\n return json.Marshal(struct {',None),
]
results=[]
with tempfile.TemporaryDirectory(prefix='ch14-mutants-') as tmp:
 tree=pathlib.Path(tmp)/'source';shutil.copytree(source,tree)
 for label,file,old,new,intended in cases:
  path=tree/file;original=path.read_text();assert original.count(old)==1,(label,original.count(old));mutated=original.replace(old,new)
  path.write_text(mutated);out.joinpath(label+'.patch').write_text(''.join(difflib.unified_diff(original.splitlines(True),mutated.splitlines(True),fromfile=file,tofile=file)))
  command=['/tmp/grade-edition2-context','-ch','14','-json','-gui',str(tree/'gui'),str(tree)]
  start=time.monotonic();r=subprocess.run(command,capture_output=True,text=True,timeout=115)
  out.joinpath(label+'.json').write_text(r.stdout or '{}\n');out.joinpath(label+'.stderr').write_text(r.stderr)
  try:report=json.loads(r.stdout)
  except:report={}
  failures=[c for c in report.get('checks',[]) if not c['passed']]
  result={'case':label,'exit':r.returncode,'seconds':round(time.monotonic()-start,3),'intended_check':intended,'intended_failure':any(c['id']==intended for c in failures) if intended else None,'score':report.get('score'),'failures':failures}
  results.append(result);out.joinpath('mutations.json').write_text(json.dumps(results,indent=2)+'\n');print(label,r.returncode,report.get('score'),result['intended_failure'],flush=True)
  path.write_text(original)
 # Restore exact runtime source, then rerun the affected suite.
 r=subprocess.run(['/tmp/grade-edition2-context','-ch','14','-json','-gui',str(tree/'gui'),str(tree)],capture_output=True,text=True,timeout=115)
 out.joinpath('restored-full.json').write_text(r.stdout);out.joinpath('restored-full.stderr').write_text(r.stderr);assert r.returncode==0
