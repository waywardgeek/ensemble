#!/usr/bin/env node
// Actual delivered UI and public Connector; exact wire tokens are recorded raw.
const fs=require('fs'),path=require('path');
const {chromium}=require(path.join(process.env.CH07_BROWSER_TOOLS||'/Users/bill/projects/ensemble-edition-2-revisions/browser-tools','node_modules/playwright'));
const [url,override]=process.argv.slice(2);
const evidence={};
const until=async(test,reason)=>{const end=Date.now()+4000;while(!test()){if(Date.now()>end)throw Error(reason);await new Promise(r=>setTimeout(r,10));}};
(async()=>{
 const browser=await chromium.launch({channel:'chrome',headless:true});
 try{
  const page=await browser.newPage(),outbound=[],inbound=[],errors=[],steps=[];Object.assign(evidence,{outbound,inbound,errors,steps});
  if(override)await page.route('**/connector.js',r=>r.fulfill({contentType:'text/javascript',body:fs.readFileSync(override,'utf8')}));
  page.on('pageerror',e=>errors.push(e.message));
  page.on('websocket',ws=>{ws.on('framesent',e=>outbound.push(e.payload));ws.on('framereceived',e=>inbound.push(e.payload));});
  const ready=async()=>{await page.waitForFunction(()=>document.querySelector('[data-preferences-applied]')?.textContent.includes('Applied revision'));await page.locator('[data-tab="settings"]').click();};
  await page.goto(url);await ready();
  async function action(domain,value,label){
   const start=outbound.length,received=inbound.length;
   if(domain==='preferences')await page.locator('[data-preference="theme"]').selectOption(value);
   else {await page.locator('[data-policy]').fill(String(value));await page.locator('[data-policy-save]').click();}
   let command;
   await until(()=>{command=outbound.slice(start).map(JSON.parse).find(m=>m.type===domain+'_update');return command;},label+' command absent');
   let reply;
   await until(()=>{reply=inbound.slice(received).map(JSON.parse).find(m=>m.id===command.id);return reply;},label+' reply absent');
   await page.waitForFunction(domain=>!document.querySelector(domain==='preferences'?'[data-preference="theme"]':'[data-policy-save]').disabled,domain);
   steps.push({label,command:outbound.slice(start).find(raw=>JSON.parse(raw).id===command.id),reply:inbound.slice(received).find(raw=>JSON.parse(raw).id===command.id),notice:await page.locator('[data-notice]').textContent(),applied:await page.locator(domain==='preferences'?'[data-preferences-applied]':'[data-policy-applied]').textContent()});
  }
  if(override){
   await page.locator('[data-preference="theme"]').selectOption('dark');
   await page.waitForFunction(()=>/base revision.*exact/i.test(document.querySelector('[data-notice]').textContent));
   const mutation_refusal=await page.locator('[data-notice]').textContent();
   if(outbound.map(JSON.parse).some(m=>m.type==='preferences_update'))throw Error('decoder deletion sent a rounded command');
   console.log(JSON.stringify({mutation_refusal,outbound,inbound,errors}));return;
  }
  await action('preferences','dark','preferences-no-change');
  await action('policy',0,'policy-no-change');
  await action('preferences','light','preferences-first');await action('policy',2,'policy-first');
  await action('preferences','system','preferences-second');await action('policy',3,'policy-second');
  await page.reload();await ready();
  const reloaded=await page.evaluate(()=>({preferences:document.querySelector('[data-preferences-applied]').textContent,policy:document.querySelector('[data-policy-applied]').textContent}));
  const retries=await page.evaluate(async()=>{
   const {Connector}=await import('/connector.js');let snapshot,preferences,resolve;
   const ready=new Promise(r=>resolve=r),owner={snapshot(s){snapshot=s;resolve();},preferences(s){preferences=s;},observation(){},reply(){},connection(){},diagnostic(){}};
   const c=new Connector(owner,location.origin.replace('http','ws')+'/ws');c.connect();await ready;
   const result=[];
   try{for(const domain of ['preferences','policy']){
    const current=domain==='preferences'?preferences:snapshot.state.execution_policy;
    const patch=domain==='preferences'?{theme:current.preferences.theme}:{max_model_requests:current.max_model_requests};
    let failure;
    try{await c.send(domain+'_update',{base_revision:BigInt(current.revision)-1n,patch});throw Error('stale update unexpectedly accepted');}catch(e){failure=e;}
    if(failure.code!=='revision_conflict'||!failure.current)throw Error('intended conflict/current absent');
    const reply=await c.send(domain+'_update',{base_revision:failure.current.revision,patch});
    result.push({domain,before:String(current.revision),conflict:String(failure.current.revision),ack:String(reply.revision)});
   }}finally{c.close();}
   return result;
  });
  const protocol=await page.evaluate(async()=>{
   const {Connector}=await import('/connector.js'),NativeSocket=globalThis.WebSocket,nativeParse=JSON.parse;
   const sockets=[],diagnostics=[],snapshots=[],observations=[];
   class ControlledSocket {static OPEN=1;constructor(){this.readyState=1;this.sent=[];sockets.push(this);}send(raw){this.sent.push(raw);}close(){this.didClose=true;}}
   globalThis.WebSocket=ControlledSocket;
   const owner={snapshot(s){snapshots.push(s);},preferences(s){this.preference=s;},observation(o){observations.push(o);},reply(){},connection(){},diagnostic(m){diagnostics.push(m);}};
   const c=new Connector(owner,'ws://controlled'),d=new Connector(owner,'ws://controlled');
   try{
    c.connect();const socket=sockets.at(-1),emit=raw=>socket.onmessage({data:raw});
    emit('{"type":"preferences_snapshot","revision":9007199254740993,"preferences":{}}');
    emit('{"type":"snapshot_begin","id":"s","generation":"g","watermark":9007199254740993,"state":{"execution_policy":{"revision":9007199254740993}}}');
    const rawText='literal "base_revision":9007199254740993 and \\ backslash';
    emit(JSON.stringify({type:'snapshot_event',generation:'g',event:{type:'response_ended',response:{parts:[{type:'tool_call',args:{revision:7,label:rawText}}]}}}));
    emit('{"type":"snapshot_end","generation":"g","watermark":9007199254740993}');
    emit('{"type":"observation","generation":"g","revision":9007199254740994,"observation":{"kind":"policy_changed","execution_policy":{"revision":9007199254740995}}}');
    if(snapshots.length!==1||observations.length!==1||String(snapshots[0].state.execution_policy.revision)!=='9007199254740993'||String(observations[0].execution_policy.revision)!=='9007199254740995')throw Error('counter normalization or exact watch successor failed');
    const args=snapshots[0].events[0].response.parts[0].args;
    if(args.revision!==7||args.label!==rawText||JSON.stringify(args)!==JSON.stringify({revision:7,label:rawText}))throw Error('unrelated argument or literal text changed');
    const pending=c.send('preferences_update',{base_revision:9007199254740993n,patch:{unrecognized:rawText}}).catch(()=>{});
    const encoded=socket.sent.at(-1);
    if(!encoded.includes('"base_revision":9007199254740993')||nativeParse(encoded).patch.unrecognized!==rawText)throw Error('exact base encoding corrupted user strings');
    c.close();await pending;
    // An older JSON.parse implementation must either retain the value by an
    // independent path, or explicitly refuse it before claiming readiness.
    JSON.parse=(raw,reviver)=>nativeParse(raw,reviver?function(k,v){return reviver.call(this,k,v);}:undefined);
    owner.preference=null;d.connect();const old=sockets.at(-1);old.onmessage({data:'{"type":"preferences_snapshot","revision":9007199254740993,"preferences":{}}'});
    if(String(owner.preference?.revision)!=='9007199254740993'&&!(old.didClose&&!d.settingsReady&&diagnostics.length))throw Error('unsupported parser silently rounded counter');
    return {exact_state:String(snapshots[0].state.execution_policy.revision),exact_observation:String(observations[0].execution_policy.revision),encoded,unrelated_text:args.label,unsupported_refused:old.didClose,diagnostics};
   }finally{JSON.parse=nativeParse;c.close();d.close();globalThis.WebSocket=NativeSocket;}
  });
  if(errors.length)throw Error(errors.join('; '));
  console.log(JSON.stringify({steps,reloaded,retries,protocol,outbound,inbound,errors}));
 }finally{await browser.close();}
})().catch(e=>{console.log(JSON.stringify({...evidence,error:e.stack}));console.error(e.stack);process.exit(1);});
