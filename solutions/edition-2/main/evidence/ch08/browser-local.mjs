// Student-owned deterministic browser checks. Native speech callbacks here are
// controlled fixtures; these checks make no audible or real-model claim.
import {chromium} from '/Users/bill/projects/ensemble-edition-2-revisions/browser-tools/node_modules/playwright/index.mjs';
import {spawn} from 'node:child_process';
import {createServer} from 'node:http';
import {mkdtemp,writeFile,readFile,rm} from 'node:fs/promises';
import {tmpdir} from 'node:os';
import {join,dirname} from 'node:path';
import {fileURLToPath} from 'node:url';
import assert from 'node:assert/strict';
const frames=[];
const runID=Date.now(),here=dirname(fileURLToPath(import.meta.url)),dir=await mkdtemp(join(tmpdir(),'ensemble-ch08-ui-')),checks=[],failures=[];
let requests=0,gui,browser;
const fixture=createServer(async(req,res)=>{let body='';for await(const data of req)body+=data;requests++;res.setHeader('content-type','application/json');res.end(JSON.stringify({content:requests===1?[{type:'text',text:'I will inspect the file.'},{type:'tool_use',id:'read-one',name:'read_file',input:{path:'notes.txt'}}]:[{type:'text',text:'**Completed.**\n[unsafe](javascript:alert(1))'}],usage:{input_tokens:1,output_tokens:1}}));});
await new Promise(r=>fixture.listen(0,'127.0.0.1',r));
async function start(log){
 gui=spawn(process.argv[2]||'/tmp/ensemble-ch08-gui',['--port','0'],{cwd:dir,env:{...process.env,LLM_VENDOR:'anthropic',LLM_MODEL:'fixture',LLM_API_KEY:'fixture',LLM_BASE_URL:`http://127.0.0.1:${fixture.address().port}`,EN_DISABLE_STREAMING:'1',CH02_LOG:join(dir,log)},stdio:['ignore','pipe','pipe']});
 let stderr='';gui.stderr.on('data',d=>stderr+=d);return await new Promise((resolve,reject)=>{gui.stdout.once('data',d=>resolve(d.toString().trim()));gui.once('exit',code=>reject(new Error(`GUI exited ${code}: ${stderr}`)));});
}
async function stop(){if(gui&&!gui.killed){const done=new Promise(r=>gui.once('exit',r));gui.kill('SIGTERM');await done;}}
async function settled(page){await page.waitForFunction(()=>![...document.querySelectorAll('[data-preference], [data-policy-save], [data-divider]')].some(control=>control.disabled));}
async function check(name,fn){try{await fn();checks.push(name);}catch(e){failures.push({name,error:e.stack});throw e;}}
try{
 await writeFile(join(dir,'notes.txt'),'<img src=x onerror="globalThis.injected=true">\n'+('retained-safe-content '.repeat(130))+'END-MARKER');
 let url=await start('first.log');browser=await chromium.launch({headless:true,channel:'chrome'});
 const a=await browser.newPage({viewport:{width:1500,height:1000}}),b=await browser.newPage({viewport:{width:1500,height:1000}});
 for(const p of [a,b]){p.on('websocket',socket=>socket.on('framereceived',e=>frames.push({page:p===a?'a':'b',message:JSON.parse(e.payload.toString())})));p.on('pageerror',e=>failures.push({name:'pageerror',error:e.message}));await p.goto(url);await p.locator('[data-status]').filter({hasText:'Connected'}).waitFor();await p.getByRole('tab',{name:'Settings'}).click();}
 await check('initial-snapshot-remains-contiguous-with-fast-control',async()=>{
  const frames=await a.evaluate(()=>new Promise((resolve,reject)=>{const socket=new WebSocket(location.origin.replace(/^http/,'ws')+'/ws'),frames=[];socket.onopen=()=>{socket.send(JSON.stringify({type:'subscribe',id:'s'}));socket.send(JSON.stringify({type:'pause',id:'p',typing:false,speaking:false}));};socket.onmessage=e=>{const m=JSON.parse(e.data);frames.push(m.type);if(m.id==='p'){socket.close();resolve(frames);}};socket.onerror=()=>reject(new Error('fixture socket failed'));}));assert.equal(frames[0],'preferences_snapshot');assert.equal(frames.at(-1),'ack');assert.equal(frames.at(-2),'snapshot_end');
 });
 await check('shared-theme-font-and-keyboard-width',async()=>{
  await a.locator('[data-preference=theme]').selectOption('light');await b.waitForFunction(()=>document.querySelector('main').dataset.theme==='light');await settled(a);
  await a.locator('[data-preference=font_size]').fill('20');await a.locator('[data-preference=font_size]').press('Tab');await b.waitForFunction(()=>getComputedStyle(document.querySelector('main')).fontSize==='20px');await settled(a);
  await a.getByRole('separator',{name:'Sidebar width',exact:true}).press('ArrowRight');await b.waitForFunction(()=>document.querySelector('[data-divider=sidebar_width]').getAttribute('aria-valuenow')==='270');await settled(a);
 });
 await check('pointer-width-and-responsive-desired-width',async()=>{
  const d=a.locator('[data-divider=actions_width]'),rect=await d.boundingBox();await a.mouse.move(rect.x+5,rect.y+40);await a.mouse.down();await a.mouse.move(rect.x-25,rect.y+40);await a.mouse.up();await b.waitForFunction(()=>document.querySelector('[data-divider=actions_width]').getAttribute('aria-valuenow')==='410');await settled(a);
  await a.setViewportSize({width:500,height:850});assert(await a.evaluate(()=>document.documentElement.scrollWidth<=innerWidth));assert.equal(await a.locator('[data-divider=actions_width]').getAttribute('aria-valuenow'),'410');await a.setViewportSize({width:1500,height:1000});
 });
 await check('invalid-draft-distinct-from-applied-and-recovery',async()=>{
  await a.locator('[data-preference=font_size]').fill('40');await a.locator('[data-preference=font_size]').press('Tab');await a.locator('[data-notice]').filter({hasText:'not saved'}).waitFor();assert.equal(await a.locator('[data-preference=font_size]').inputValue(),'20');
 });
 await check('false-zero-and-policy-control',async()=>{
  await a.locator('[data-preference=autoplay]').check();await b.locator('[data-preference=autoplay]').waitFor();await b.waitForFunction(()=>document.querySelector('[data-preference=autoplay]').checked);await settled(a);
  await a.locator('[data-preference=autoplay]').uncheck();await b.waitForFunction(()=>!document.querySelector('[data-preference=autoplay]').checked);await settled(a);
  await a.locator('[data-policy]').fill('2');await a.locator('[data-policy-save]').click();await b.locator('[data-policy-applied]').filter({hasText:'Next turn: 2 '}).waitFor();await settled(a);
  await a.locator('[data-policy]').fill('0');await a.locator('[data-policy-save]').click();await b.locator('[data-policy-applied]').filter({hasText:'Default (16)'}).waitFor();await settled(a);assert.equal(requests,0);
 });
 await check('mixed-response-safe-two-pane-full-retained-content',async()=>{
  await a.locator('[data-input]').fill('Read notes.txt and report.');await a.locator('[data-prompt]').click();await a.locator('[data-notice]').filter({hasText:'success'}).waitFor();
  assert.match(await a.locator('[data-artifacts]').innerText(),/I will inspect the file/);assert.match(await a.locator('[data-actions]').innerText(),/read_file/);
  const result=a.locator('[data-actions] .artifact').filter({hasText:'Tool result: read-one'});await result.getByRole('button',{name:'Expand full retained text'}).click();assert.match(await result.innerText(),/END-MARKER/);assert.equal(await a.locator('img,iframe').count(),0);assert.equal(await a.evaluate(()=>globalThis.injected),undefined);assert.equal(requests,2);
  await a.locator('[data-input]').fill('unfinished correction');await a.locator('[data-input]').focus();await b.locator('[data-preference=theme]').selectOption('dark');await a.waitForFunction(()=>document.querySelector('main').dataset.theme==='dark');assert.equal(await a.locator('[data-input]').inputValue(),'unfinished correction');assert.equal(await result.getByRole('button',{name:'Show preview'}).getAttribute('aria-expanded'),'true');assert(await a.locator('[data-input]').evaluate(el=>document.activeElement===el));await a.locator('[data-input]').fill('');
 });
 await check('system-theme-and-browser-restart-persistence',async()=>{
  await a.locator('[data-preference=theme]').selectOption('system');await b.waitForFunction(()=>document.querySelector('[data-preference=theme]').value==='system');await settled(a);await a.emulateMedia({colorScheme:'light'});await a.waitForFunction(()=>document.querySelector('main').dataset.theme==='light');await a.emulateMedia({colorScheme:'dark'});await a.waitForFunction(()=>document.querySelector('main').dataset.theme==='dark');
  await a.screenshot({path:join(here,`local-browser-${runID}.png`),fullPage:true});await a.close();await b.close();await stop();url=await start('restart.log');const p=await browser.newPage();await p.goto(url);await p.locator('[data-status]').filter({hasText:'Connected'}).waitFor();assert.equal(await p.locator('[data-preference=font_size]').inputValue(),'20');assert.equal(await p.locator('[data-divider=sidebar_width]').getAttribute('aria-valuenow'),'270');assert.equal(await p.locator('[data-preference=autoplay]').isChecked(),false);assert.match(await p.locator('[data-policy-applied]').textContent(),/Default \(16\)/);assert.equal(await p.locator('.artifact').count(),0);await p.close();
 });
 const p=await browser.newPage();await p.goto(url);await p.locator('[data-status]').filter({hasText:'Connected'}).waitFor();
 await check('controlled-speech-revision-rate-and-page-cancellation',async()=>{
  const result=await p.evaluate(async()=>{
   const {SpeechQueue}=await import('/speech.js'),{BrowserApplication}=await import('/application.js');const spoken=[],ends=[],causes=[],errors=[];let cancels=0;
   class U{constructor(text){this.text=text;}}
   const synthesis={speak(u){spoken.push(u);u.onstart?.()},cancel(){cancels++}};
   const app=new BrowserApplication(synthesis,U);
   const waitSpoken=async count=>{const deadline=Date.now()+5000;while(spoken.length<count){if(Date.now()>deadline)throw new Error('native lease was not granted');await new Promise(r=>setTimeout(r,5));}};
   const owner=name=>({name,typing:false,application:()=>app,diagnostic:m=>errors.push(m),speechEvent:e=>ends.push({name,...e}),reconcile(){causes.push({name,typing:this.typing,speaking:this.queue.busy()});return Promise.resolve()}});
   const a=owner('a'),b=owner('b');a.queue=new SpeechQueue(a);b.queue=new SpeechQueue(b);b.typing=true;
   const prefs=(revision,autoplay,rate)=>({revision,preferences:{autoplay,speech_rate:rate}});
   a.queue.preferences(prefs(4,true,1.5));b.queue.preferences(prefs(4,true,1.5));a.queue.activate();b.queue.activate();
   a.queue.enqueue('old-a','A old sentence');b.queue.enqueue('old-b','B old sentence');await waitSpoken(1);
   const old=spoken[0];a.queue.preferences(prefs(5,false,0.7));b.queue.preferences(prefs(5,false,0.7));
   const before={a:a.queue.busy(),b:b.queue.busy(),rate:old.rate,revision:a.queue.current.revision};
   a.queue.cancel();await waitSpoken(2);old.onend();old.onerror({error:'stale'});
   const after={b:b.queue.busy(),rate:spoken[1]?.rate,revision:b.queue.current?.revision,cancels};
   const o={agent_id:'a',request_id:'r',operation_id:'m',part_id:1};
   b.queue.observe({...o,kind:'part_delta',channel:'text',text:'Off backlog.'});b.queue.preferences(prefs(6,true,0.8));b.queue.observe({...o,kind:'part_final',part:{type:'text',text:'Off backlog.'}});
   const backlog=b.queue.queue.length;
   b.queue.observe({...o,part_id:2,kind:'part_final',part:{type:'text',text:'New sentence.'}});const future={rate:b.queue.queue[0].rate,revision:b.queue.queue[0].revision};
   spoken[1].onend();await waitSpoken(3);spoken[2].onend();await Promise.resolve();await Promise.resolve();const final=causes.filter(v=>v.name==='b').at(-1);
   a.queue.close();b.queue.close();app.close();return {before,after,backlog,future,final,errors};
  });
  assert.deepEqual(result.before,{a:true,b:true,rate:1.5,revision:4});assert.deepEqual(result.after,{b:true,rate:1.5,revision:4,cancels:1});assert.equal(result.backlog,0);assert.deepEqual(result.future,{rate:.8,revision:6});assert.deepEqual(result.final,{name:'b',typing:true,speaking:false});
 });
 assert.equal(failures.length,0,JSON.stringify(failures));
} catch(error){if(!failures.length)failures.push({name:'setup',error:error.stack});}
finally{await browser?.close();await stop();await new Promise(r=>fixture.close(r));await rm(dir,{recursive:true,force:true});}
const result={scope:'Local fake-model browser and controlled speech callbacks, no live or audible claim',passed:failures.length===0,checks,failures,frames,model_requests:requests};await writeFile(join(here,`browser-local-${runID}.json`),JSON.stringify(result,null,2)+'\n');console.log(JSON.stringify(result,null,2));if(failures.length)process.exitCode=1;
