#!/usr/bin/env node
// Actual Chrome/public components, controlled socket/native service; no provider
// calls, audible output or valid full-history exhaustion claim.
const fs=require('fs'),path=require('path'),os=require('os'),http=require('http'),crypto=require('crypto');
const {chromium}=require(path.join(process.env.CH07_BROWSER_TOOLS||'/Users/bill/projects/ensemble-edition-2-revisions/browser-tools','node_modules/playwright'));
const source=path.resolve(process.argv[2]),temporary=fs.mkdtempSync(path.join(os.tmpdir(),'ch09-browser-')),assets=path.join(temporary,'assets');
const hash=p=>crypto.createHash('sha256').update(fs.readFileSync(p)).digest('hex');
function files(root,prefix=''){return Object.fromEntries(fs.readdirSync(root).sort().flatMap(name=>{const p=path.join(root,name),relative=path.join(prefix,name);return fs.statSync(p).isDirectory()?Object.entries(files(p,relative)):[[relative,hash(p)]]}));}
const before=files(path.join(source,'gui/web/gui'));fs.cpSync(path.join(source,'gui/web/gui'),assets,{recursive:true});
const html=fs.readFileSync(path.join(assets,'index.html'),'utf8').replace(/<script\b[^>]*>[\s\S]*?<\/script>/gi,'');
const server=http.createServer((req,res)=>{
 if(req.url==='/'){res.setHeader('Content-Type','text/html');res.end(html);return;}
 const helper={'/__review.js':'ch09-browser-fixture.js','/__base.js':'ch08-browser-fixture.js'}[req.url];const file=helper?path.join(__dirname,helper):path.resolve(assets,'.'+decodeURIComponent(req.url.split('?')[0]));
 if((!helper&&!file.startsWith(assets+path.sep))||!fs.existsSync(file)){res.writeHead(404);res.end();return;}res.setHeader('Content-Type',file.endsWith('.css')?'text/css':'text/javascript');let bytes=fs.readFileSync(file);if(helper==='ch08-browser-fixture.js'){const text=bytes.toString(),anchor='state:{lifecycle:';if(text.split(anchor).length!==2)throw Error('unique predecessor snapshot anchor changed');bytes=Buffer.from(text.replace(anchor,'state:{skills:null,lifecycle:'));}res.end(bytes);
});
const checks=[],expect=(v,m)=>{if(!v)throw Error(m)};
(async()=>{await new Promise(r=>server.listen(0,'127.0.0.1',r));const origin='http://127.0.0.1:'+server.address().port,browser=await chromium.launch({channel:'chrome',headless:true});
 async function run(id,fn){const page=await browser.newPage({viewport:{width:1500,height:1000}}),errors=[];page.on('pageerror',e=>errors.push(e.message));page.setDefaultTimeout(2000);await page.route('**/*',r=>r.request().url().startsWith(origin+'/')?r.continue():r.abort());try{await page.goto(origin);await page.evaluate(async()=>window.review=await(await import('/__review.js')).setup());const result=await fn(page);expect(errors.length===0,'uncaught errors: '+errors.join('; '));checks.push({id,passed:true,result});}catch(e){checks.push({id,passed:false,error:e.stack,page_errors:errors});}finally{await page.close();}}
 await run('inherited-card-and-native-service-positive',async page=>{
  await page.evaluate(()=>review.a.observation({kind:'message_received',agent_id:'a1',event:{seq:1,type:'message_received',message:{actor:'human',purpose:'dialogue',parts:[{type:'text',text:'BASELINE_TEXT'}]}}}));
  const card=page.locator('#review-a [data-artifacts] article').filter({hasText:'BASELINE_TEXT'});expect(await card.count()===1,'inherited positive card absent');await card.getByRole('button',{name:'Speak full card'}).click();await page.waitForFunction(()=>review.spoken.length===1);expect(await page.evaluate(()=>review.spoken[0].text==='BASELINE_TEXT'),'inherited explicit speech positive absent');await page.evaluate(()=>review.end());return {spoken:1};
 });
 await run('safe-manual-cards-silent-replay-expansion-and-explicit-speech',async page=>{
  await page.evaluate(async()=>{await review.remote({autoplay:true});review.a.speech.activate();review.snapshot();await review.tick();});
  const chat=page.locator('#review-a [data-artifacts]'),alpha=chat.locator('article').filter({hasText:'ALPHA_MARK'});
  expect(await alpha.count()===1&&await chat.locator('article').filter({hasText:'BETA_MARK'}).count()===1,'distinct material cards absent');
  expect(await page.evaluate(()=>review.spoken.length)===0,'historical manual auto-spoke');await alpha.getByRole('button',{name:'Expand full retained text'}).click();expect((await alpha.textContent()).includes('END_ALPHA'),'expansion lost retained tail');expect(await alpha.locator('img,script,iframe').count()===0&&!await page.evaluate(()=>window.ch09Executed),'manual HTML executed');
  await alpha.getByRole('button',{name:'Speak full card'}).click();await page.waitForFunction(()=>review.spoken.length===1);expect(await page.evaluate(()=>review.spoken[0].text===review.body),'speaker did not use full retained manual');await page.evaluate(()=>review.end());
  await page.evaluate(()=>review.observe({kind:'skills_changed',agent_id:'a1',skills:review.state(false,true)}));expect(await alpha.count()===1&&/retired/i.test(await alpha.textContent()),'retirement duplicated or failed to label card');
  await page.evaluate(()=>review.observe({kind:'tool_called',agent_id:'a1',event:{seq:30,type:'tool_called',tool:{call_id:'management',name:'load_skill',args:{name:'alpha'}}}},12));expect((await page.locator('#review-a [data-actions]').textContent()).includes('load_skill'),'management action lost actions pane');
  await page.evaluate(()=>review.snapshot());expect(await alpha.count()===1&&await page.evaluate(()=>review.spoken.length)===1,'reconnect duplicated or auto-spoke manuals');return {cards:await chat.locator('article').count()};
 });
 await run('current-state-beyond-window-and-independent-page',async page=>{
  await page.evaluate(()=>{review.snapshot(review.sockets[0],false,[],review.state(false,true));review.snapshot(review.sockets[1],false,[],null);});
  const a=await page.locator('#review-a aside[aria-label="Sidebar"]').textContent(),b=await page.locator('#review-b aside[aria-label="Sidebar"]').textContent();expect(a.includes('base')&&a.includes('beta')&&a.includes('read_file'),'current Skills unavailable when transition is outside window');expect(!b.includes('beta'),'other Page inherited skill authority');expect(await page.locator('#review-a [data-artifacts] article').count()===0,'current state fabricated out-of-window body cards');return {a,b};
 });
 await run('uint64-snapshot-material-retired-identity-and-opaque-boundary',async page=>{
  return page.evaluate(async()=>{const r=review,expect=(v,m)=>{if(!v)throw Error(m)};r.snapshot(r.sockets[0],true);const snapshot=r.snapshots.at(-1),s=snapshot.state.skills;
   expect(String(s.revision)==='18446744073709551615','skill revision rounded');expect(s.active.filter(x=>x.name!=='base').map(x=>String(x.activation)).join(',')==='9007199254740992,9007199254740993','active IDs rounded');
   const records=snapshot.events[0].skills.activated;expect(records.map(x=>String(x.activation)).join(',')==='9007199254740992,9007199254740993'&&String(records[1].dependencies[0])==='9007199254740992','material/dependency identity rounded');
   const cards=[...r.a.root.querySelectorAll('[data-artifacts] article')];expect(cards.length===2&&cards.some(c=>c.textContent.includes('ALPHA_MARK'))&&cards.some(c=>c.textContent.includes('BETA_MARK')),'adjacent unsafe activations share card identity');
   r.observe({kind:'skills_changed',agent_id:'a1',skills:r.state(true,true)});expect(String(r.observations.at(-1).skills.retired[0].activation)==='9007199254740992','retired identity rounded');expect(cards.some(c=>c.textContent.includes('ALPHA_MARK')&&/retired/i.test(c.textContent)),'exact retired identity did not update matching card');
   r.observe({kind:'tool_called',agent_id:'a1',event:{seq:30,type:'tool_called',tool:{call_id:'opaque',name:'read_file',args:{revision:9007199254740993n,activation:9007199254740993n,path:'manual revision 9007199254740993'}}}},12);
   const args=r.observations.at(-1).event.tool.args;expect(typeof args.revision==='number'&&typeof args.activation==='number'&&args.path==='manual revision 9007199254740993','protocol conversion rewrote arbitrary tool arguments/text');return {revision:String(s.revision),keys:cards.map(c=>c.dataset.key)};
  });
 });
 await run('unsafe-counter-without-source-context-refuses-before-acceptance',async page=>{
  return page.evaluate(()=>{const r=review,parse=JSON.parse;JSON.parse=function(raw,reviver){return parse(raw,reviver?function(k,v){return reviver.call(this,k,v)}:undefined)};try{r.snapshot(r.sockets[0],false,[],r.state(false));if(r.snapshots.length!==1)throw Error('safe-integer positive failed without source context');const before=r.snapshots.length,notice=r.a.notice.textContent;r.snapshot(r.sockets[0],true);if(r.snapshots.length!==before)throw Error('unsafe snapshot accepted after rounded counter');if(!r.a.notice.textContent||r.a.notice.textContent===notice)throw Error('unsafe refusal lacked visible diagnostic');return {notice:r.a.notice.textContent};}finally{JSON.parse=parse}});
 });
 await browser.close();await new Promise(r=>server.close(r));expect(JSON.stringify(before)===JSON.stringify(files(path.join(source,'gui/web/gui'))),'input assets changed during run');
 const passed=checks.length===5&&checks.every(x=>x.passed);console.log(JSON.stringify({passed,scope:'Actual Chrome, controlled public socket/native callbacks. Counter fixtures are projection inputs, not valid complete exhausted logs. No provider or audible claim.',checks,source,assets:before,checkers:Object.fromEntries(['accept_ch09_browser.cjs','ch09-browser-fixture.js','ch08-browser-fixture.js'].map(name=>[name,hash(path.join(__dirname,name))]))},null,2));fs.rmSync(temporary,{recursive:true,force:true});process.exitCode=passed?0:1;
})().catch(e=>{console.error(e.stack);server.close();fs.rmSync(temporary,{recursive:true,force:true});process.exitCode=1});
