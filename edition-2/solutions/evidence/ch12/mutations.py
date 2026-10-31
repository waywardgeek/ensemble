"""One-chapter defect probes in an isolated source copy; original grader untouched."""
import json,pathlib,shutil,subprocess,tempfile,os,signal
root=pathlib.Path('/Users/bill/projects/ensemble/solutions/edition-2')
cases=[
 ('initialize notification','internal/mcp/client.go','Method: "notifications/initialized"','Method: "notifications/omitted"','TestMCPLoop'),
 ('request correlation','internal/mcp/client.go','reply := c.pending[id]','reply := c.pending[id]; if id > 2 { reply = c.pending[id+1] }','TestMCPCorrelation'),
 ('remote arguments','internal/mcp/client.go','"arguments": args','"arguments": json.RawMessage(`{}`)','TestMCPLoop'),
 ('hidden declarations','internal/agent/agent.go','tool.Ephemeral != "" || !a.toolEnabled(tool.Name)','!a.toolEnabled(tool.Name)','TestMCPLoop'),
 ('hidden dispatch','internal/agent/agent.go','(automatic || tool.Ephemeral == "")','(automatic || true)','TestMCPLoop'),
 ('fresh round snapshot','internal/agent/mcp.go','tool.Ephemeral == "round" ||','(tool.Ephemeral == "round" && !t.ephemeralTurn) ||','TestMCPLoop'),
 ('turn frequency','internal/agent/mcp.go','tool.Ephemeral == "turn" && !t.ephemeralTurn','tool.Ephemeral == "turn"','TestMCPLoop'),
 ('ephemeral error rejection','internal/mcp/client.go','if result.IsError {','if false && result.IsError {','TestMCPLoop'),
 ('reverse dispatch','internal/mcp/client.go','Invoke(context.Background(), params.Name, params.Arguments)','Invoke(context.Background(), "missing_tool", params.Arguments)','TestMCPLoop'),
 ('kill cancellation','internal/jobs/process.go','j.data.Status = "killed"\n\tj.cancel()','j.data.Status = "killed"','TestMCPCorrelation'),
 ('callback is not cancellation','internal/jobs/jobs.go','func (j *job) Wait(limits common.Limits) (string, error) {','func (j *job) Wait(limits common.Limits) (string, error) {\n defer j.cancel()','TestMCPCorrelation'),
 ('disconnected visibility','internal/mcp/mcp.go','case <-c.done:\n\t\t\tcontinue','case <-c.done:\n\t\t\t// mutant retains disconnected tools','TestMCPCorrelation'),
 ('skill admission','internal/mcp/mcp.go','allowed := entry.allowed == nil','allowed := true','TestMCPSkill'),
 ('unload visibility','internal/mcp/mcp.go','if entry.skill == skill {','if false && entry.skill == skill {','TestMCPSkill'),
 ('unload process lifetime','internal/mcp/mcp.go','failures = append(failures, c.close())','_ = c','TestMCPSkill'),
 ('captured empty args','internal/engine/anthropic_stream.go','if args := arguments[frame.Index]; args != "" {','if args, ok := arguments[frame.Index]; ok {','TestMCPAnthropic'),
]
results=[]
with tempfile.TemporaryDirectory(prefix='ensemble-ch12-mutants-') as directory:
 scratch=pathlib.Path(directory)/'main';shutil.copytree(root/'main',scratch)
 baseline=subprocess.run(['go','test','./ensemble','-run','TestMCP','-count=1','-timeout','15s'],cwd=scratch,text=True,capture_output=True)
 if baseline.returncode:raise RuntimeError('baseline failed '+baseline.stdout+baseline.stderr)
 for claim,path,before,after,test in cases:
  if os.getenv("CH12_MUTATIONS") and claim not in os.getenv("CH12_MUTATIONS").split("|"):continue
  file=scratch/path;original=file.read_text()
  if before not in original:raise RuntimeError('mutation missing: '+claim)
  file.write_text(original.replace(before,after,1))
  try:
   run=subprocess.run(['go','test','./ensemble','-run',test,'-count=1','-timeout','12s'],cwd=scratch,text=True,capture_output=True,timeout=20)
   output=run.stdout+run.stderr;code=run.returncode
  except subprocess.TimeoutExpired as e:output='outer timeout '+str(e);code=124
  finally:
   file.write_text(original)
   # Only the deliberately leaked external peer from this mutation is stopped.
   for line in subprocess.check_output(['ps','-axo','pid,command'],text=True).splitlines()[1:]:
    fields=line.strip().split(None,1)
    if len(fields)==2 and 'python' in fields[1].split(' ',1)[0] and "open('peer.pid','w')" in fields[1]:
     try:os.kill(int(fields[0]),signal.SIGKILL)
     except ProcessLookupError:pass
  results.append({'claim':claim,'file':path,'mutation':before+' -> '+after,'exit':code,'output':output,'status':'failure to inspect' if code else 'SURVIVED'})
  print(claim,'exit',code,flush=True)
 final=subprocess.run(['go','test','./ensemble','-run','TestMCP','-count=1','-timeout','15s'],cwd=scratch,text=True,capture_output=True)
 results.append({'restored_checks_exit':final.returncode,'output':final.stdout+final.stderr})
(root/('evidence/ch12/mutations-corrections.json' if os.getenv('CH12_MUTATIONS') else 'evidence/ch12/mutations.json')).write_text(json.dumps(results,indent=2)+'\n')
