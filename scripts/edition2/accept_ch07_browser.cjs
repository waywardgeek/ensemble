#!/usr/bin/env node
/* Independent real-Chrome DOM/component checks. Transport and speech are controlled
   doubles here; actual synthesis/audio and paid model browser paths remain separate. */
const fs = require('fs'), path = require('path'), http = require('http'), crypto = require('crypto');
const source = path.resolve(process.argv[2] || 'solutions/edition-2/main');
const toolsRoot = process.env.CH07_BROWSER_TOOLS || '/Users/bill/projects/ensemble-edition-2-revisions/browser-tools';
const {chromium} = require(path.join(toolsRoot, 'node_modules/playwright'));
const assetCopy = fs.mkdtempSync(path.join(require('os').tmpdir(), 'ch07-browser-assets-'));
const assets = path.join(assetCopy, 'assets');
fs.cpSync(path.join(source, 'gui/web/gui'), assets, {recursive:true});
const html = `<!doctype html><meta charset="utf-8"><main><p data-status></p><p data-pause></p><section data-artifacts tabindex="0" style="height:200px;overflow:auto"></section><textarea data-input aria-label="Prompt or correction"></textarea><button data-prompt>Send prompt</button><button data-hint>Send hint</button><button data-interrupt>Interrupt</button><button data-latest>Return to latest</button><button data-auto-speech>Auto speech</button><button data-cancel-speech>Cancel speech</button><p data-notice></p></main>`;
const server = http.createServer((req,res)=>{
 if(req.url==='/review.html'){res.setHeader('Content-Type','text/html');res.end(html);return;}
 const target=path.join(assets,decodeURIComponent(req.url.split('?')[0]));
 if(!target.startsWith(assets+path.sep)||!fs.existsSync(target)||!fs.statSync(target).isFile()){res.writeHead(404);res.end();return;}
 res.setHeader('Content-Type',target.endsWith('.js')?'text/javascript':'text/plain');res.end(fs.readFileSync(target));
});
const checks=[];
const assert=(value,message)=>{if(!value)throw new Error(message);};
(async()=>{
 await new Promise(resolve=>server.listen(0,'127.0.0.1',resolve));
 const origin=`http://127.0.0.1:${server.address().port}`;
 const browser=await chromium.launch({channel:'chrome',headless:true});
 const context=await browser.newContext();await context.route('**/*',route=>route.request().url().startsWith(origin+'/')?route.continue():route.abort());let page;
 async function run(id,action){
  page=await context.newPage();const errors=[],requests=[],dialogs=[];
  page.on('pageerror',e=>errors.push(e.message));page.on('dialog',async d=>{dialogs.push(d.message());await d.dismiss();});
  page.on('request',r=>{if(!r.url().startsWith(origin+'/'))requests.push(r.url());});
  try{await page.goto(origin+'/review.html');await action(page);assert(!dialogs.length,'rendered payload executed a dialog');assert(!requests.length,'rendered payload fetched external bytes');assert(!errors.length,'uncaught browser exception: '+errors.join('; '));checks.push({id,passed:true,page_errors:errors});}
  catch(e){checks.push({id,passed:false,details:e.stack,page_errors:errors,dialogs,external_requests:requests});}
  finally{await page.close();}
 }
 await run('cards-identity-partial-final-window',async page=>{
  const result=await page.evaluate(async()=>{
   const {ArtifactScroll}=await import('/artifacts.js');const box=document.querySelector('[data-artifacts]');
   const owner={speak(){},diagnostic(){}};const cards=new ArtifactScroll(owner,box);const id={agent_id:'a1',request_id:'r1',operation_id:'m1',part_id:1};
   cards.reset({agent_id:'a1',omitted:0,events:[],partials:[{...id,channels:{text:'Hel'}}]});
   cards.observation({...id,kind:'part_delta',channel:'text',text:'lo.'});cards.observation({...id,kind:'part_final',response_seq:8,part_index:0,part:{type:'text',text:'Hello.'}});
   cards.observation({kind:'response_ended',agent_id:'a1',event:{seq:8,type:'response_ended',response:{parts:[{type:'text',text:'Hello.'}]}}});
   const one=[...box.querySelectorAll('article')].map(x=>x.textContent);
   cards.observation({...id,agent_id:'a2',kind:'part_delta',channel:'text',text:'SECOND-AGENT'});
   cards.observation({...id,operation_id:'m2',kind:'part_delta',channel:'text',text:'SECOND-OPERATION'});
   const separate=box.querySelectorAll('article').length;
   cards.observation({...id,operation_id:'m2',kind:'model_end',accepted:false});const incomplete=box.textContent.includes('Incomplete');
   const snapshot={agent_id:'a1',omitted:19,events:[{seq:8,type:'response_ended',response:{parts:[{type:'text',text:'Hello.'}]}},{seq:10,type:'tool_returned',tool:{call_id:'older-call',is_error:false,parts:[{type:'text',text:'EARLIER-RESULT'}]}}],partials:[]};
   cards.reset(snapshot);cards.reset(snapshot);
   return {one,separate,incomplete,after:box.textContent,count:box.querySelectorAll('article').length};
  });
  assert(result.one.length===1&&result.one[0].includes('Hello.'),'partial/final accepted answer duplicated or lost');
  assert(result.separate===3,'full Agent/operation identity collapsed');assert(result.incomplete,'rejected provisional card not marked incomplete');
  assert(result.count===3&&result.after.includes('19')&&result.after.includes('older-call')&&/missing.*context/i.test(result.after)&&result.after.includes('EARLIER-RESULT'),'window reset/earlier-call disclosure lost');
 });
 await run('hostile-content-expansion-keyboard-focus-scroll',async page=>{
  await page.evaluate(async()=>{
   const {ArtifactScroll}=await import('/artifacts.js');window.cards=new ArtifactScroll({speak:(key,text)=>{window.spoken=text;},diagnostic(){}},document.querySelector('[data-artifacts]'));
   const hostile='<img src="https://invalid.example/pixel" onerror="alert(1)"> [open](javascript:alert(1)) \x1b[31mRED\x1b[0m';
   const long='START '+hostile+' '+('retained '.repeat(400))+' FULL-END-MARKER';window.expected=long;
   cards.reset({agent_id:'a',omitted:0,partials:[],events:[{seq:1,type:'tool_called',tool:{call_id:'evil',name:hostile,args:{path:hostile}}},{seq:2,type:'tool_returned',tool:{call_id:'evil',is_error:false,parts:[{type:'text',text:long}] }},{seq:3,type:'response_ended',response:{parts:[{type:'text',text:'Paragraph *emphasis*\n\n- list item\n\n```js\n<img src=x onerror=alert(2)>\n```\n'+hostile}]}}]});
   document.querySelector('[data-input]').focus();
  });
  assert(await page.locator('article').filter({hasText:'FULL-END-MARKER'}).count()===0,'long-card positive did not truncate preview');
  assert(await page.locator('[data-artifacts]').textContent().then(t=>t.includes('<img')&&t.includes('javascript:alert(1)')&&t.includes('characters omitted')),'hostile fixture never reached renderer');
  const expand=page.getByRole('button',{name:'Expand full retained text'}).first();await expand.focus();await page.keyboard.press('Enter');
  assert(await page.getByRole('button',{name:'Show preview'}).first().getAttribute('aria-expanded')==='true','expander inaccessible state absent');
  assert((await page.locator('[data-artifacts]').textContent()).includes('FULL-END-MARKER'),'keyboard expansion failed to expose full retained text');
  assert(await page.locator('[data-artifacts] img, [data-artifacts] iframe, [data-artifacts] script, [data-artifacts] a[href^="javascript:"]').count()===0,'hostile markup became executable DOM');
  assert(await page.locator('[data-artifacts] em').count()>0&&await page.locator('[data-artifacts] li').count()>0&&await page.locator('[data-artifacts] pre code').count()>0,'safe Markdown subset not rendered');
  await page.evaluate(()=>{document.querySelector('[data-input]').focus();for(let i=0;i<20;i++)cards.card('extra-'+i,'card', 'retained text '.repeat(50));const box=document.querySelector('[data-artifacts]');box.scrollTop=0;box.dispatchEvent(new Event('scroll'));cards.card('last','new output','NEW');});
  assert(await page.locator('[data-input]').evaluate(x=>document.activeElement===x),'incoming cards stole input focus');
  assert(await page.locator('[data-artifacts]').evaluate(x=>x.scrollTop)===0,'incoming card pulled reader away from earlier result');
  await page.evaluate(()=>cards.latest());assert(await page.locator('[data-artifacts]').evaluate(x=>x.scrollTop)>0,'explicit latest action did not restore following');
 });
 await run('running-report-and-job-terminal',async page=>{
  const result=await page.evaluate(async()=>{
   const {ArtifactScroll}=await import('/artifacts.js');const box=document.querySelector('[data-artifacts]'),cards=new ArtifactScroll({speak(){},diagnostic(){}},box);
   cards.reset({agent_id:'a',omitted:0,partials:[],events:[{seq:1,type:'tool_called',tool:{call_id:'job-a',name:'run_command',args:{command:'controlled job'}}},{seq:2,type:'tool_returned',tool:{call_id:'job-a',parts:[{type:'text',text:'job still producing'}],job:{handle:11,status:'running'}}}]});
   const before=box.textContent;cards.observation({agent_id:'a',kind:'job_ended',event:{seq:3,type:'job_ended',job:{handle:11,status:'completed'}}});return {before,after:box.textContent};
  });
  assert(result.before.includes('running')&&!result.before.includes('Job completed'),'running report treated as terminal process');assert(result.after.includes('Job completed'),'later job event did not update its retained card');
 });
 await run('speech-cancel-stale-callback-and-final-dedup',async page=>{
  const result=await page.evaluate(async()=>{
   const {SpeechQueue}=await import('/speech.js');let q;const spoken=[],causes=[],diagnostics=[],events=[];const tick=()=>new Promise(r=>setTimeout(r,0));
   const synthesis={speak:u=>spoken.push(u),cancel(){}};class Utterance{constructor(text){this.text=text;}}
   const owner={reconcile:async()=>{causes.push({typing:true,speaking:q.busy()});},diagnostic:s=>diagnostics.push(s),speechEvent:e=>events.push(e)};
   q=new SpeechQueue(owner,synthesis,Utterance);q.enqueue('A','A.');q.enqueue('B','B.');await tick();const a=spoken[0];q.cancel();q.enqueue('C','C.');await tick();a.onend();a.onerror({error:'interrupted'});await tick();const afterCancel=spoken.map(u=>u.text);const busy=q.busy();spoken.at(-1).onerror({error:'network'});await tick();const released=!q.busy();
   q.reset();q.enable(true);const id={agent_id:'a',request_id:'r1',operation_id:'m1',part_id:1};q.seed([{...id,channels:{text:'OLD HISTORY. '}}]);
   q.observe({...id,kind:'part_delta',channel:'text',text:'New sentence.'});await tick();q.observe({...id,kind:'part_final',part:{type:'text',text:'OLD HISTORY. New sentence.'}});q.observe({...id,kind:'model_end',accepted:true});await tick();const finals=spoken.map(u=>u.text);
   q.observe({...id,operation_id:'m2',kind:'part_final',part:{type:'text',text:'New operation.'}});q.observe({...id,agent_id:'b',kind:'part_final',part:{type:'text',text:'Other Agent.'}});
   for(let drain=0;q.busy()&&drain<20;drain++){spoken.at(-1).onend();await tick();}if(q.busy())throw new Error('speech queue did not drain');
   const all=spoken.map(u=>u.text);q.observe({...id,kind:'tool_returned',event:{type:'tool_returned',tool:{parts:[{type:'text',text:'SILENT RESULT'}]}}});await tick();
   return {afterCancel,busy,released,finals,all,causes,diagnostics,resultSpoken:spoken.some(u=>u.text.includes('SILENT RESULT'))};
  });
  assert(JSON.stringify(result.afterCancel)===JSON.stringify(['A.','C.']),'stale canceled callbacks restarted abandoned speech');assert(result.busy&&result.released,'speech error failed to settle queue');
  assert(result.finals.filter(x=>x==='New sentence.').length===1&&!result.finals.some(x=>x.includes('OLD HISTORY')),'replay or finalization spoke text twice');assert(result.all.includes('New operation.')&&result.all.includes('Other Agent.'),'speech cursor collapsed full identity');
  assert(result.causes.every(x=>x.typing)&&result.causes.some(x=>x.speaking)&&result.causes.some(x=>!x.speaking),'speech transitions lost independent typing or speaking cause');assert(!result.resultSpoken,'tool result auto-spoken');
 });
 await run('connector-atomic-reset-and-uncertain-prompt',async page=>{
  const result=await page.evaluate(async()=>{
   const {Connector}=await import('/connector.js');const sockets=[];class WS{static OPEN=1;constructor(){this.readyState=1;this.sent=[];sockets.push(this);}send(s){this.sent.push(JSON.parse(s));}close(){this.readyState=3;this.onclose?.();}}
   window.WebSocket=WS;const snapshots=[],observations=[],messages=[];const owner={snapshot:s=>snapshots.push(s),observation:o=>observations.push(o),reply(){},connection(){},diagnostic:m=>messages.push(m)};
   const c=new Connector(owner,'ws://local');c.connect();const s=sockets[0];s.onopen();
   const frame=m=>s.onmessage({data:JSON.stringify(m)});frame({type:'snapshot_begin',id:'c1',generation:'g1',watermark:4,agent_id:'a',state:{},omitted:0});frame({type:'snapshot_event',generation:'g1',event:{seq:1,type:'message_received'}});
   const staged=snapshots.length===0;frame({type:'snapshot_end',generation:'g1',watermark:4});
   frame({type:'snapshot_begin',id:'unused',generation:'g2',watermark:9,agent_id:'a',state:{},omitted:0});frame({type:'snapshot_event',generation:'g2',event:{seq:2,type:'message_received'}});
   const uncertain=c.send('prompt',{text:'UNCERTAIN-PROMPT'}).catch(e=>e.message);s.close();const error=await uncertain;
   c.connect();const next=sockets.at(-1);next.onopen();const sentAgain=next.sent.filter(x=>x.type==='prompt').length;
   s.onmessage({data:JSON.stringify({type:'snapshot_end',generation:'g2',watermark:9})});c.close();
   return {staged,snapshots:snapshots.length,first:snapshots[0].events.length,error,sentAgain,observations};
  });
  assert(result.staged&&result.snapshots===1&&result.first===1,'partial snapshot merged into visible history');assert(/acceptance unknown/i.test(result.error)&&result.sentAgain===0,'uncertain prompt resent or missing explanation');
 });
 await run('page-input-submit-and-lost-acceptance',async page=>{
  const result=await page.evaluate(async()=>{
   const sockets=[],commands=[];let lose=false;const tick=()=>new Promise(r=>setTimeout(r,0));
   class WS{static OPEN=1;constructor(){this.readyState=1;sockets.push(this);}frame(m){this.onmessage?.({data:JSON.stringify(m)});}send(s){const m=JSON.parse(s);commands.push(m);queueMicrotask(()=>{
    if(m.type==='subscribe'){this.frame({type:'snapshot_begin',id:m.id,generation:'g1',watermark:0,agent_id:'a',omitted:0,state:{paused:false,typing_clients:0,speaking_clients:0}});this.frame({type:'snapshot_end',generation:'g1',watermark:0});}
    if(lose&&m.type==='prompt'){this.close();return;}
    if(!lose&&m.type==='prompt')this.frame({type:'accepted',id:m.id,request_id:'r1'});
    if(!lose&&m.type==='pause')this.frame({type:'ack',id:m.id,revision:0,paused:m.typing||m.speaking,typing_clients:+m.typing,speaking_clients:+m.speaking});
   });}close(){if(this.readyState!==1)return;this.readyState=3;this.onclose?.();}}
   const synthesized=[];Object.defineProperty(window,'speechSynthesis',{configurable:true,value:{speak:u=>synthesized.push(u),cancel(){}}});window.SpeechSynthesisUtterance=class{constructor(text){this.text=text;}};
   window.WebSocket=WS;const {Page}=await import('/page.js');const p=new Page(document.querySelector('main'),'ws://local');sockets[0].onopen();await tick();
   p.input.value='FIRST';p.input.dispatchEvent(new Event('input'));await tick();await p.submit('prompt');
   const clear=commands.some(m=>m.type==='pause'&&!m.typing&&!m.speaking),value=p.input.value;
   p.input.value='UNCERTAIN';p.input.dispatchEvent(new Event('input'));await tick();p.speech.enqueue('manual-test','A queued sentence.');await tick();const duringSpeech=commands.filter(m=>m.type==='pause').at(-1);synthesized[0].onerror({error:'network'});await tick();const afterSpeech=commands.filter(m=>m.type==='pause').at(-1);lose=true;await p.submit('prompt');await tick();const notice=p.notice.textContent;p.close();
   return {clear,value,notice,duringSpeech,afterSpeech,prompts:commands.filter(m=>m.type==='prompt').length};
  });
  assert(result.clear&&result.value==='','explicit input clearing failed to reconcile pause');
  assert(result.duringSpeech.typing&&result.duringSpeech.speaking&&result.afterSpeech.typing&&!result.afterSpeech.speaking,'actual page speech error cleared its independent typing cause');
  assert(/acceptance unknown/i.test(result.notice),'page hid uncertain prompt acceptance behind a control failure');
  assert(result.prompts===2,'page automatically resent an uncertain prompt');
 });
 await browser.close();await new Promise(resolve=>server.close(resolve));
 const files={};for(const name of fs.readdirSync(assets).sort()){const p=path.join(assets,name);if(fs.statSync(p).isFile())files[name]=crypto.createHash('sha256').update(fs.readFileSync(p)).digest('hex');}
 const result={scope:'Real Chrome public-component DOM and controlled socket/speech tests; not actual speech/audio or live model evidence',browser:browser.version(),source,asset_sha256:files,checker_sha256:crypto.createHash('sha256').update(fs.readFileSync(__filename)).digest('hex'),passed:checks.every(x=>x.passed),checks};
 fs.rmSync(assetCopy,{recursive:true,force:true});console.log(JSON.stringify(result,null,2));process.exitCode=result.passed?0:1;
})().catch(e=>{console.error(e.stack);server.close();process.exitCode=1;});
