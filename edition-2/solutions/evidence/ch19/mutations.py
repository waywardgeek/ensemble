"""Concrete Chapter 19 defects in disposable copies; no original grader edits."""
import json,pathlib,shutil,subprocess,tempfile
repo=pathlib.Path.cwd();source=repo/'solutions/edition-2/main';evidence=repo/'solutions/edition-2/evidence/ch19'
cases=[
 ('signature','internal/auth/identity.go','valid := false','valid := true','TestOAuthRejectsInvalidIdentity/signature','invalid identity accepted'),
 ('issuer','internal/auth/identity.go','claims.Issuer != config.Issuer','false && claims.Issuer != config.Issuer','TestOAuthRejectsInvalidIdentity/issuer','invalid identity accepted'),
 ('audience','internal/auth/identity.go','if !matched ||','if false && !matched ||','TestOAuthRejectsInvalidIdentity/audience','invalid identity accepted'),
 ('nonce','internal/auth/identity.go','nonce != "" && claims.Nonce != nonce','false && nonce != "" && claims.Nonce != nonce','TestOAuthRejectsInvalidIdentity/nonce','invalid identity accepted'),
 ('expiry','internal/auth/identity.go','claims.ExpiresAt <= time.Now().Unix()','claims.ExpiresAt <= time.Now().Unix()-100000000000','TestOAuthRejectsInvalidIdentity/expiry','invalid identity accepted'),
 ('subject','internal/auth/identity.go','claims.Subject == ""','false && claims.Subject == ""','TestOAuthRejectsInvalidIdentity/subject','invalid identity accepted'),
 ('state','internal/auth/oauth.go','if values.Get("state") != state {','if false && values.Get("state") != state {','TestOAuthRejectsInvalidIdentity/state','invalid identity accepted'),
 ('host','internal/auth/oauth.go','if value.HostID == "" {','if true {','TestOAuthLoginRefreshAndLogout','returning registration'),
 ('scope','internal/auth/oauth.go','if !status(value).PlanEnabled {','if false && !status(value).PlanEnabled {','TestOAuthMissingScopeAndTerminalRefresh','missing grant permitted inference'),
 ('proactive_refresh','internal/auth/oauth.go','if !time.Now().Before(p.ExpiresAt.Add(-5 * time.Minute)) {','if false && !time.Now().Before(p.ExpiresAt.Add(-5 * time.Minute)) {','TestOAuthLoginRefreshAndLogout','wrong bearer'),
 ('rotation_lock','internal/auth/store.go','err = syscall.Flock(int(f.Fd()), syscall.LOCK_EX|syscall.LOCK_NB)','err = nil','TestOAuthLoginRefreshAndLogout','stale rotating refresh token'),
 ('terminal_cleanup','internal/auth/oauth.go','clear(p)','_ = p','TestOAuthMissingScopeAndTerminalRefresh','terminal cleanup'),
 ('transient_retry','internal/auth/oauth.go','attempt < 3','attempt < 1','TestOAuthTransientRecovery','transient state/retry'),
 ('logout_clear','internal/auth/oauth.go','clear(&p)','_ = p','TestOAuthLoginRefreshAndLogout','logout changed identity or billing'),
 ('catalog_visibility','internal/auth/oauth.go','m.Visibility == "list" && m.Slug != ""','m.Slug != ""','TestOAuthLoginRefreshAndLogout','catalog:'),
 ('owner_only','internal/auth/store.go','path := f.Name()','f.Chmod(0644)\n path := f.Name()','TestOAuthLoginRefreshAndLogout','credentials are not owner-only'),
 ('billing_selection','ensemble/ensemble.go','cfg.BillingMode = status.Billing','cfg.BillingMode = common.CredentialAPIKey; _ = status','TestOAuthInferenceBillingAndIsolation','incorrect selected bearer'),
 ('bearer_redaction','internal/engine/engine.go','data = bytes.ReplaceAll(data, []byte(secret), []byte("[REDACTED]"))','data = bytes.ReplaceAll(data, []byte(secret), []byte(secret))','TestOAuthInferenceBillingAndIsolation','secret echo reached answer'),
 ('plan_namespace','internal/engine/capabilities.go','namespace: "ensemble"','namespace: ""','TestOAuthInferenceBillingAndIsolation','plan functions missing namespace'),
 ('stream_flag','internal/engine/responses.go','"stream": true','"stream": false','TestResponsesSummaryToolLoop','Responses delivery flags missing'),
 ('developer','internal/engine/responses.go','message("developer", "input_text", e.SystemPrompt())','message("system", "input_text", e.SystemPrompt())','TestResponsesSummaryToolLoop','developer cache block missing'),
 ('cache_marker','internal/engine/responses.go','last["prompt_cache_breakpoint"] =','last["ignored_marker"] =','TestResponsesSummaryToolLoop','developer cache block missing'),
 ('tool_result','internal/engine/responses.go','"output": text(p.Parts)','"output": "lost"','TestResponsesSummaryToolLoop','result/reasoning replay absent'),
 ('opaque_replay','internal/engine/responses.go','if p.From == e.Target() {','if false && p.From == e.Target() {','TestResponsesSummaryToolLoop','result/reasoning replay absent'),
 ('summary_channel','internal/engine/responses_stream.go','emit(key, common.DeltaReasoningSummary, frame.Delta)','emit(key, common.DeltaText, frame.Delta)','TestResponsesSummaryToolLoop','context deadline exceeded'),
 ('input_usage','internal/engine/responses_stream.go','Input: *u.Input - u.Details.Read - u.Details.Write','Input: *u.Input','TestResponsesSummaryToolLoop','disjoint usage'),
 ('model_boundary','internal/engine/media.go','ContextWindow: 128000, Media:','Responses:true, ContextWindow: 128000, Media:','TestResponsesModelBoundary','model transport'),
 ('input_lens','internal/cachelens/lens.go','"messages", "contents", "input"','"messages", "contents"','TestResponsesLensIncludesInput','Responses content omitted'),
]
paths=subprocess.check_output(['git','ls-files','solutions/edition-2/main'],text=True).splitlines()
results=[]
with tempfile.TemporaryDirectory(prefix='ensemble-ch19-mutants-') as temporary:
 root=pathlib.Path(temporary)/'main'
 for name in paths:
  src=repo/name;dst=root/src.relative_to(source);dst.parent.mkdir(parents=True,exist_ok=True);shutil.copy2(src,dst)
 baseline=subprocess.run(['go','test','./ensemble','-run','OAuth|Responses','-count=1'],cwd=root,text=True,capture_output=True,timeout=45)
 (evidence/'mutation-baseline.txt').write_text(baseline.stdout+baseline.stderr)
 if baseline.returncode:raise SystemExit('baseline failed')
 for label,name,old,new,test,needle in cases:
  path=root/name;original=path.read_text()
  if old not in original:results.append({'case':label,'status':'setup mismatch'});continue
  path.write_text(original.replace(old,new))
  try:
   run=subprocess.run(['go','test','./ensemble','-run','^'+test+'$','-count=1'],cwd=root,text=True,capture_output=True,timeout=20)
   output=run.stdout+run.stderr
   status='killed' if run.returncode and needle in output and '[build failed]' not in output else ('survived' if not run.returncode else 'wrong failure')
   results.append({'case':label,'status':status,'exit':run.returncode,'expected_failure':needle,'output':output})
  except subprocess.TimeoutExpired as failure:results.append({'case':label,'status':'harness timeout'})
  finally:path.write_text(original)
  (evidence/'mutations.json').write_text(json.dumps(results,indent=2)+'\n')
 print(json.dumps({'cases':len(results),'results':[{k:v for k,v in r.items() if k!='output'} for r in results]},indent=2))
