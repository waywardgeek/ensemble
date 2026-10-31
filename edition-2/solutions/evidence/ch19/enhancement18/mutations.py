import pathlib,subprocess,tempfile,shutil,json,time,difflib
root=pathlib.Path.cwd();out=root/'solutions/edition-2/evidence/ch19/enhancement18';src=root/'solutions/edition-2/ch18'
def case(name,file,old,new,check,positive=False):return dict(name=name,file=file,old=old,new=new,check=check,positive=positive)
cases=[
case('wrong-bearer','internal/engine/engine.go','"Bearer "+cfg.APIKey','"Bearer wrong-credential"','credential-provider'),
case('logged-bearer','internal/engine/engine.go','secrets = append(secrets, cfg.APIKey)','secrets = append(secrets, cfg.APIKey)\n e.parent.Ensemble().Logf("credential: %s", cfg.APIKey)','credential-provider'),
case('pkce-plain','internal/auth/oauth.go','"code_challenge_method": {"S256"}','"code_challenge_method": {"plain"}','oauth-flow'),
case('unchecked-signature','internal/auth/identity.go','if !valid {','if !valid && false {','oauth-flow'),
case('unchecked-nonce','internal/auth/identity.go','(nonce != "" && claims.Nonce != nonce)','false','oauth-flow'),
case('missing-grant-allowed','internal/auth/oauth.go','if !status(value).PlanEnabled {','if false {','oauth-flow'),
case('expired-only-renewal','internal/auth/oauth.go','p.ExpiresAt.Add(-5 * time.Minute)','p.ExpiresAt','token-refresh'),
case('rotated-old-bearer','internal/auth/oauth.go','replace(p, token)','oldAccess := p.AccessToken\n replace(p, token)\n p.AccessToken = oldAccess','token-refresh'),
case('store-true','internal/engine/responses.go','"store": false','"store": true','responses-format'),
case('no-encrypted-include','internal/engine/responses.go','[]string{"reasoning.encrypted_content"}','[]string{}','responses-format'),
case('drop-opaque','internal/engine/responses.go','if p.From == e.Target() {','if false && p.From == e.Target() {','responses-format'),
case('wrong-result-id','internal/engine/responses.go','"call_id": p.CallID, "output"','"call_id": "wrong-result", "output"','responses-format'),
case('empty-success-read','internal/tools/tools.go','// SplitAfter retains real newline bytes','data = nil\n // SplitAfter retains real newline bytes','responses-format'),
case('plan-explicit-cache','internal/engine/capabilities.go','omitExplicitCache: true','omitExplicitCache: false','cache-breakpoints'),
case('missing-api-marker','internal/engine/responses.go','last["prompt_cache_breakpoint"] = map[string]string{"mode": "explicit"}','// marker accidentally omitted','cache-breakpoints'),
case('misplaced-api-marker','internal/engine/responses.go','last["prompt_cache_breakpoint"] = map[string]string{"mode": "explicit"}','input[i]["prompt_cache_breakpoint"] = map[string]string{"mode": "explicit"}','cache-breakpoints'),
case('summary-as-text','internal/engine/responses_stream.go','emit(key, common.DeltaReasoningSummary, frame.Delta)','emit(key, common.DeltaText, frame.Delta)','reasoning-summaries'),
case('summary-only-at-final','internal/engine/responses_stream.go','emit(key, common.DeltaReasoningSummary, frame.Delta)','// wait for the completed reasoning item','reasoning-summaries'),
case('summary-truncated','internal/engine/responses_stream.go','emit(key, common.DeltaReasoningSummary, frame.Delta)','emit(key, common.DeltaReasoningSummary, strings.TrimSpace(frame.Delta))','reasoning-summaries'),
case('lost-refusal-error','gui/gui.go','"error": o.Error','"error": ""','billing-mode'),
case('nonzero-success-exit','gui/cmd/main.go','func main() {','func main() {\n if len(os.Args)<2 || os.Args[1]!="auth" { defer os.Exit(17) }','credential-provider'),
case('split-summary-control','internal/engine/responses_stream.go','cb.OnDelta(id, kind, value)','if kind == common.DeltaReasoningSummary { for _, r := range value { cb.OnDelta(id, kind, string(r)) } } else { cb.OnDelta(id, kind, value) }','reasoning-summaries',True),
]
subprocess.run(['go','build','-o','/tmp/edition2-grade18','./cmd/grade-edition2'],check=True)
with tempfile.TemporaryDirectory(prefix='ch18-mutants-') as tmp:
 work=pathlib.Path(tmp)/'source';shutil.copytree(src,work);rows=[]
 for c in cases:
  f=work/c['file'];old=f.read_text();assert c['old'] in old,c['name'];new=old.replace(c['old'],c['new']);f.write_text(new)
  patch=''.join(difflib.unified_diff(old.splitlines(True),new.splitlines(True),fromfile='a/'+c['file'],tofile='b/'+c['file']));(out/(c['name']+'.patch')).write_text(patch)
  t=time.monotonic();r=subprocess.run(['/tmp/edition2-grade18','-ch','18','-json',str(work)],capture_output=True,text=True,timeout=65);secs=round(time.monotonic()-t,3)
  (out/(c['name']+'.json')).write_text(r.stdout);(out/(c['name']+'.stderr')).write_text(r.stderr)
  try:
   report=json.loads(r.stdout);failed={x['id']:x.get('details',[]) for x in report['checks'] if not x['passed']};score=report['score'];caught=c['check'] in failed
  except Exception:failed={};score=None;caught=False
  row={k:c[k] for k in ['name','file','check','positive']};row.update(exit=r.returncode,seconds=secs,score=score,failed=failed,intended=(r.returncode==0 if c['positive'] else caught));rows.append(row)
  (out/'mutations.json').write_text(json.dumps(rows,indent=2)+'\n');print(c['name'],score,row['intended'],flush=True);f.write_text(old)
 r=subprocess.run(['/tmp/edition2-grade18','-ch','18','-json',str(work)],capture_output=True,text=True);(out/'restored.json').write_text(r.stdout);(out/'restored.stderr').write_text(r.stderr);print('restored',r.returncode,flush=True)
