import {chromium} from '/Users/bill/projects/ensemble-edition-2-revisions/browser-tools/node_modules/playwright/index.mjs';
import http from 'node:http';
import {spawn} from 'node:child_process';
import {mkdtemp,writeFile} from 'node:fs/promises';
import {tmpdir} from 'node:os';
import {join} from 'node:path';
import assert from 'node:assert/strict';
const workspace=await mkdtemp(join(tmpdir(),'ensemble-ch07-browser-'));
let calls=0;
const model=http.createServer((req,res)=>{calls++;req.resume();res.writeHead(200,{'Content-Type':'text/event-stream'});
const send=(event,data)=>res.write(`event: ${event}\ndata: ${JSON.stringify(data)}\n\n`);
send('message_start',{type:'message_start',message:{id:'fixture',model:'fixture',usage:{input_tokens:1,output_tokens:0},content:[]}});
send('content_block_start',{type:'content_block_start',index:0,content_block:{type:'text',text:''}});
send('content_block_delta',{type:'content_block_delta',index:0,delta:{type:'text_delta',text:'Hel'}});
setTimeout(()=>{send('content_block_delta',{type:'content_block_delta',index:0,delta:{type:'text_delta',text:'lo.'}});send('content_block_stop',{type:'content_block_stop',index:0});send('message_delta',{type:'message_delta',delta:{stop_reason:'end_turn'},usage:{output_tokens:2}});send('message_stop',{type:'message_stop'});res.end()},300);
});await new Promise(r=>model.listen(0,'127.0.0.1',r));
const child=spawn(process.argv[2]||'/tmp/ensemble-ed2-ch07-gui',['--port','0','--gui-log',join(workspace,'gui.jsonl')],{cwd:workspace,env:{...process.env,LLM_VENDOR:'anthropic',LLM_MODEL:'fixture',LLM_API_KEY:'fixture-only',LLM_BASE_URL:`http://127.0.0.1:${model.address().port}`,CH02_LOG:join(workspace,'events.jsonl'),EN_DISABLE_STREAMING:'0'}});
let stderr='';child.stderr.on('data',d=>stderr+=d);
const url=await new Promise((resolve,reject)=>{let output='';child.stdout.on('data',d=>{output+=d;const found=output.match(/http:\/\/127\.0\.0\.1:\d+/);if(found)resolve(found[0]);});child.on('exit',code=>reject(new Error('GUI exit '+code+' '+stderr)));});
const browserServer=await chromium.launchServer({channel:'chrome',headless:process.env.HEADFUL !== '1'});const browser=await chromium.connect(browserServer.wsEndpoint());const browserPID=browserServer.process().pid;
const receipt={kind:'deterministic browser controls; no paid calls',workspace,url,browser:browser.version(),checks:[],errors:[]};
try{
 const page=await browser.newPage();page.on('pageerror',e=>receipt.errors.push(e.message));await page.goto(url);await page.locator('[data-status]').filter({hasText:'Connected'}).waitFor();
 await page.locator('[data-input]').fill('local fixture');await page.getByRole('button',{name:'Send prompt',exact:true}).click();
 await page.getByText('Hello.',{exact:true}).waitFor();await page.getByText('Outcome: success',{exact:true}).waitFor();assert.equal(await page.getByText('Hello.',{exact:true}).count(),1);receipt.checks.push('stream-final-one-card');
 assert.equal(await page.locator('[data-input]').inputValue(),'');await page.locator('[data-pause]').filter({hasText:'typing: 0'}).waitFor();receipt.checks.push('submit-clears-pause');
 await page.reload();await page.getByText('Hello.',{exact:true}).waitFor();assert.equal(await page.getByText('Hello.',{exact:true}).count(),1);receipt.checks.push('reconnect-final-one-card');
 const result=await page.evaluate(async()=>{
  const {ArtifactScroll}=await import('/artifacts.js');const {SpeechQueue}=await import('/speech.js');
  const host=document.createElement('section');document.body.append(host);const spoken=[],notes=[];const owner={speak:(key,text)=>spoken.push(text),diagnostic:m=>notes.push(m)};const scroll=new ArtifactScroll(owner,host);
  const payload='<img src=x onerror=alert(1)> [open](javascript:alert(1))';
  scroll.reset({agent_id:'a',omitted:1,events:[{seq:101,type:'tool_returned',tool:{call_id:'older-call',parts:[{type:'text',text:payload+'x'.repeat(1400)}],is_error:false}}],partials:[]});
  const before=host.textContent;host.querySelector('button').click();const after=host.textContent;
  const safe=!host.querySelector('img,iframe,a,script')&&after.includes(payload)&&after.includes('Missing call context')&&before.includes('characters omitted');
  const starts=[],events=[];const synthesis={speak:u=>starts.push(u),cancel:()=>{}};class Utterance{constructor(text){this.text=text}}
  let typing=true;const causes=[];const speechOwner={reconcile:async()=>{causes.push({typing,speaking:queue.busy()})},diagnostic:m=>notes.push(m),speechEvent:e=>events.push(e)};const queue=new SpeechQueue(speechOwner,synthesis,Utterance);
  queue.enqueue('A','A');queue.enqueue('B','B');await Promise.resolve();await Promise.resolve();const a=starts[0];queue.cancel();queue.enqueue('C','C');await Promise.resolve();await Promise.resolve();a.onend();a.onerror({error:'canceled'});await Promise.resolve();
  const stale=starts.map(u=>u.text).join(',')==='A,C'&&queue.current?.text==='C'&&causes.every(c=>c.typing);
  starts.at(-1).onend();queue.enable(true);const o={agent_id:'a',request_id:'r',operation_id:'m',part_id:1};queue.observe({...o,kind:'part_delta',channel:'text',text:'Words.'});await Promise.resolve();await Promise.resolve();queue.observe({...o,kind:'part_final',part:{type:'text',text:'Words.'}});const once=starts.filter(u=>u.text==='Words.').length===1&&queue.queue.length===0;
  return {safe,stale,once,starts:starts.map(u=>u.text),causes};
 });assert.ok(result.safe);assert.ok(result.stale);assert.ok(result.once);receipt.checks.push('safe-content-expansion-earlier-call','speech-stale-callback','speech-final-not-repeated');receipt.speechFixture=result;
 if (process.env.CHECK_REAL_SPEECH === '1') {
  await page.evaluate(() => { window.realSpeechEvents=[]; document.querySelector('main').addEventListener('ensemble-speech', e=>window.realSpeechEvents.push({...e.detail, at:Date.now()})); });
  const voices=await page.evaluate(()=>speechSynthesis.getVoices().map(v=>({name:v.name,lang:v.lang,localService:v.localService})));
  let capture, captureDone, captureOutput='';
  if (process.env.AUDIO_CAPTURE_BINARY) {
    capture=spawn(process.env.AUDIO_CAPTURE_BINARY,[join(workspace,'browser-speech.wav'),String(browserPID)]);
    captureDone=new Promise(resolve=>capture.on('exit',code=>resolve(code)));
    capture.stdout.on('data',d=>captureOutput+=d);capture.stderr.on('data',d=>captureOutput+=d);
    await Promise.race([new Promise(resolve=>capture.stdout.on('data',()=>{if(captureOutput.includes('capture_ready'))resolve()})),captureDone]);
  }

  if(capture) await new Promise(resolve=>setTimeout(resolve,2000));
  await page.locator('main .artifact').filter({has:page.getByRole('heading',{name:'Answer',exact:true})}).getByRole('button',{name:'Speak full card'}).click();
  await page.waitForFunction(()=>window.realSpeechEvents.some(e=>e.type==='end'||e.type==='error'),{},{timeout:15000}).catch(()=>{});
  receipt.realSpeech={voices,events:await page.evaluate(()=>window.realSpeechEvents),status:await page.locator('[data-notice]').textContent(),audioCapture:'not yet captured; callbacks alone are not audio evidence'};
  await page.getByRole('button',{name:'Cancel speech',exact:true}).click();
  if(capture){receipt.realSpeech.captureExit=await captureDone;receipt.realSpeech.captureOutput=captureOutput;receipt.realSpeech.audioPath=join(workspace,'browser-speech.wav');}

 }
 await page.screenshot({path:join(workspace,'local.png'),fullPage:true});assert.deepEqual(receipt.errors,[]);assert.equal(calls,1);receipt.passed=true;
}catch(e){receipt.passed=false;receipt.failure=e.stack;process.exitCode=1}finally{await browser.close();await browserServer.close();child.kill('SIGTERM');model.close();receipt.modelCalls=calls;receipt.stderr=stderr;await writeFile(join(workspace,'receipt.json'),JSON.stringify(receipt,null,2));console.log(JSON.stringify(receipt,null,2));}
