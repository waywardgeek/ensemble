"""Bounded probes of the remaining current claims; replay incurs no API calls."""
import hashlib,json,pathlib,shutil,subprocess,tempfile
repo=pathlib.Path.cwd(); source=repo/'solutions/edition-2/main'; evidence=repo/'solutions/edition-2/evidence/ch19'
paths=subprocess.check_output(['git','ls-files','solutions/edition-2/main'],text=True).splitlines()+['solutions/edition-2/main/gui/responses.test.mjs']
cases=[
 ('pkce','internal/auth/oauth.go','"code_verifier": {verifier}','"code_verifier": {"wrong"}','TestOAuthLoginRefreshAndLogout','PKCE/code/redirect not bound'),
 ('resource','internal/auth/oauth.go','"resource": {"https://api.openai.com/v1"}','"resource": {"https://wrong.test"}','TestOAuthLoginRefreshAndLogout','token exchange lost issued registration/resource'),
 ('transient_preservation','internal/auth/oauth.go','errors.As(err, &rejection) && terminal(string(rejection))','errors.As(err, &rejection) || err != nil','TestOAuthTransientRecovery','transient state/retry'),
 ('unauthorized_renewal','internal/engine/engine.go','if err == nil && resp.StatusCode == http.StatusUnauthorized','if false && err == nil && resp.StatusCode == http.StatusUnauthorized','TestOAuthInferenceBillingAndIsolation','secret echo reached answer'),
 ('foreign_replay','internal/engine/responses.go','if p.From == e.Target() {','if true {','TestResponsesOpaqueProvenance','foreign reasoning replayed'),
 ('failed_completion','internal/engine/responses_stream.go','case "response.completed":','case "response.completed", "response.failed", "response.incomplete":','TestResponsesRequiresCompletion','incomplete stream accepted'),
]
results=[]
with tempfile.TemporaryDirectory(prefix='ensemble-ch19-supplemental-') as temporary:
 root=pathlib.Path(temporary)/'main'
 for name in paths:
  src=repo/name; dst=root/src.relative_to(source); dst.parent.mkdir(parents=True,exist_ok=True);shutil.copy2(src,dst)
 baseline=subprocess.run(['go','test','./ensemble','-run','OAuth|Responses','-count=1'],cwd=root,text=True,capture_output=True,timeout=45)
 (evidence/'supplemental-baseline.txt').write_text(baseline.stdout+baseline.stderr)
 if baseline.returncode:raise SystemExit('baseline failed')
 for label,name,old,new,test,needle in cases:
  path=root/name; original=path.read_text();assert old in original
  changed=original.replace(old,new)
  if label=='failed_completion':changed=changed.replace('case "response.failed", "response.incomplete", "error":','case "error":')
  path.write_text(changed)
  run=subprocess.run(['go','test','./ensemble','-run','^'+test+'$','-count=1'],cwd=root,text=True,capture_output=True,timeout=20)
  output=run.stdout+run.stderr
  results.append({'case':label,'exit':run.returncode,'status':'killed' if run.returncode and needle in output and '[build failed]' not in output else 'unproven','output':output})
  path.write_text(original)
 # Retain actual browser baseline and then remove only incremental summary speech.
 (root/'gui/node_modules').symlink_to(source/'gui/node_modules',target_is_directory=True)
 baseline=subprocess.run(['node','--test','responses.test.mjs'],cwd=root/'gui',text=True,capture_output=True,timeout=30)
 (evidence/'gui-responses-baseline.txt').write_text(baseline.stdout+baseline.stderr)
 if baseline.returncode:raise SystemExit('GUI baseline failed')
 path=root/'gui/web/gui/app.js';original=path.read_text();old='["text", "thinking", "reasoning_summary"].includes(message.kind)';assert old in original
 path.write_text(original.replace(old,'["text", "thinking"].includes(message.kind)'))
 run=subprocess.run(['node','--test','responses.test.mjs'],cwd=root/'gui',text=True,capture_output=True,timeout=30)
 output=run.stdout+run.stderr
 results.append({'case':'incremental_gui_speech','exit':run.returncode,'status':'killed' if run.returncode and 'waitForFunction: Timeout' in output else 'unproven','output':output})
 path.write_text(original)
 # Reuse real recorded request bytes through the fixed diagnostic. No model call.
 requests=[json.loads(x)['data'] for x in (evidence/'live-openai-api.log').read_text().splitlines() if json.loads(x)['event']=='request']
 (root/'requests.json').write_text(json.dumps(requests))
 (root/'ensemble/live_lens_test.go').write_text('''package ensemble_test
import("os";"encoding/json";"strings";"testing";"ensemble/ensemble")
func TestLiveLensReplay(t *testing.T){
 raw,err:=os.ReadFile("../requests.json");if err!=nil{t.Fatal(err)}
 var requests []string;if err=json.Unmarshal(raw,&requests);err!=nil{t.Fatal(err)}
 var log strings.Builder
 a,err:=ensemble.New(&log).NewAgent(ensemble.Config{Vendor:ensemble.OpenAI,Model:"gpt-6.1-sol",DataDir:t.TempDir()});if err!=nil{t.Fatal(err)};defer a.Shutdown()
 for i,body:=range requests{p:=a.Engine().CacheLens().ObserveRequest("openai/gpt-6.1-sol",[]byte(body));if i>0&&(p<=0||p>=1){t.Fatalf("invalid prediction %f",p)}}
 if strings.Contains(log.String(),"cache prefix changed"){t.Fatal(log.String())}
 t.Log(log.String())
}
''')
 run=subprocess.run(['go','test','./ensemble','-run','TestLiveLensReplay','-v','-count=1'],cwd=root,text=True,capture_output=True,timeout=20)
 (evidence/'live-lens-replay.txt').write_text(run.stdout+run.stderr)
 if run.returncode:raise SystemExit('live lens replay failed')
(evidence/'supplemental-mutations.json').write_text(json.dumps(results,indent=2)+'\n')
print(json.dumps([{k:v for k,v in r.items() if k!='output'} for r in results],indent=2))
