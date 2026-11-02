// Isolated public-interface test; no live credentials, provider, or user document.
const assert = require('node:assert/strict');
const fs = require('node:fs'); const os = require('node:os'); const path = require('node:path');
const http = require('node:http'); const net = require('node:net'); const { spawn } = require('node:child_process');
const { chromium } = require(process.env.PLAYWRIGHT_MODULE || 'playwright');
const dll = process.argv[2]; if (!dll) throw Error('Pass the built DocumentChat.dll path.');
const sdk = process.env.DOTNET || 'dotnet';
const scratch = fs.mkdtempSync(path.join(os.tmpdir(), 'document-chat-test-'));
const document = Array.from({length:266}, (_,i)=>`[PDF PAGE ${i+1}]\nSynthetic reference ${i+1}. DOCUMENT_MARKER_${i+1}\n`).join('\n');
const doc = path.join(scratch,'document.txt'); fs.writeFileSync(doc,document);
fs.writeFileSync(path.join(scratch,'SOUL.md'),'POISON_PERSONAL_MEMORY');
fs.writeFileSync(path.join(scratch,'MEMORY.md'),'POISON_PERSONAL_MEMORY');
fs.mkdirSync(path.join(scratch,'memory')); fs.writeFileSync(path.join(scratch,'memory','old.md'),'POISON_PERSONAL_MEMORY');
const requests=[]; let host, browser, closed; let stdout='', stderr='';
const provider=http.createServer(async(req,res)=>{
 try {
  let text=''; for await(const chunk of req) {text+=chunk; if(text.length>4_000_000) throw Error('oversized test request');}
  const body=JSON.parse(text); requests.push(body);
  if (JSON.stringify(body.input).includes('SYNTHETIC_PROVIDER_FAILURE')) throw Error('synthetic provider refusal');
  const questions=JSON.stringify(body.input).match(/QUESTION_[A-Z0-9]+/g)||[];
  const question=questions.at(-1); if(!question) throw Error('missing test question');
  res.setHeader('content-type','application/json');
  res.end(JSON.stringify({id:'resp_fixture_'+requests.length,object:'response',status:'completed',model:'gpt-6-astra',
   output:[{id:'msg_fixture',type:'message',role:'assistant',status:'completed',content:[{type:'output_text',text:'ANSWER_'+question+' <img src=x onerror=alert(1)>',annotations:[]}]}],
   usage:{input_tokens:100,output_tokens:10,total_tokens:110}}));
 } catch(e) {res.statusCode=500; res.end(JSON.stringify({error:{message:String(e)}}));}
});
const listen = server => new Promise(resolve=>server.listen(0,'127.0.0.1',()=>resolve(server.address().port)));
const sleep=ms=>new Promise(resolve=>setTimeout(resolve,ms));
(async()=>{
 const providerPort=await listen(provider); const reservation=net.createServer(); const port=await listen(reservation);
 await new Promise(resolve=>reservation.close(resolve)); const base=`http://127.0.0.1:${port}`;
 try {
  host=spawn(sdk,[path.resolve(dll),'--document',doc,'--port',String(port),'--base-url',`http://127.0.0.1:${providerPort}`],
   {cwd:scratch,env:{PATH:'/usr/bin:/bin',HOME:scratch,TMPDIR:scratch,DOCUMENT_CHAT_API_KEY:'synthetic-only',DOTNET_CLI_HOME:scratch,DOTNET_SKIP_FIRST_TIME_EXPERIENCE:'1',DOTNET_GENERATE_ASPNET_CERTIFICATE:'false'}});
  closed=new Promise(resolve=>host.on('exit',(code,signal)=>resolve({code,signal})));
  host.stdout.on('data',b=>{stdout=(stdout+b).slice(-8192)});host.stderr.on('data',b=>{stderr=(stderr+b).slice(-8192)});
  let about; const end=Date.now()+15000;
  while(Date.now()<end) { if(host.exitCode!==null) throw Error('host exited: '+stderr); try {const response=await fetch(base+'/about'); if(response.ok){about=await response.json();break}}catch{} await sleep(100); }
  assert(about,'host readiness'); assert.equal(about.model,'gpt-6-astra'); assert.equal(about.documentBytes,Buffer.byteLength(document));
  browser=await chromium.launch({headless:true,channel:'chrome'});
  const a=await browser.newContext(); const b=await browser.newContext(); const pa=await a.newPage(); const pb=await b.newPage();
  const errors=[]; pa.on('pageerror',e=>errors.push(String(e)));pb.on('pageerror',e=>errors.push(String(e)));
  async function ready(page){await page.waitForFunction(()=>!document.querySelector('#question').disabled,{},{timeout:15000});}
  await pa.goto(base); await pb.goto(base); await ready(pa);await ready(pb);assert.equal(requests.length,0,'loading a document must not call the provider');
  assert.equal(await pa.locator('#instructions').textContent(),about.instructions);
  async function ask(page,question){await ready(page);await page.fill('#question',question);await page.click('#send');await page.waitForFunction(q=>document.querySelector('#messages').textContent.includes('ANSWER_'+q),question,{timeout:15000});}
  await ask(pa,'QUESTION_A1'); await ask(pb,'QUESTION_B1'); fs.unlinkSync(doc); await ask(pa,'QUESTION_A2');
  assert(!JSON.stringify(requests[1]).includes('QUESTION_A1'),'visitor B saw A');
  assert(JSON.stringify(requests[2]).includes('ANSWER_QUESTION_A1'),'follow-up lost its own answer');
  assert(!JSON.stringify(requests[2]).includes('QUESTION_B1'),'visitor A saw B');
  await pa.reload();await ready(pa);assert.equal(await pa.locator('#messages').textContent(),'');await ask(pa,'QUESTION_A3');
  assert(!JSON.stringify(requests[3]).includes('QUESTION_A1'),'refresh restored old session');assert(!JSON.stringify(requests[3]).includes('QUESTION_A2'));
  await pb.click('#new-session');await ready(pb);await ask(pb,'QUESTION_B2');assert(!JSON.stringify(requests[4]).includes('QUESTION_B1'),'new conversation restored old session');
  assert.equal(requests.length,5);
  for(const request of requests){
   assert.equal(request.instructions,about.instructions);assert.equal(request.store,false);assert.equal((request.tools||[]).length,0);
   const wire=JSON.stringify(request);assert(!wire.includes('POISON_PERSONAL_MEMORY'));
   const texts=[]; function walk(v){if(typeof v==='string')texts.push(v);else if(Array.isArray(v))v.forEach(walk);else if(v&&typeof v==='object')Object.values(v).forEach(walk);}walk(request.input);
   assert.equal(texts.filter(t=>t.includes(document)).length,1,'full immutable reference missing or repeated');
  }
  assert.equal(await pa.locator('#messages img').count(),0,'model HTML was interpreted');
  await pa.setViewportSize({width:390,height:844});assert(await pa.evaluate(()=>document.documentElement.scrollWidth<=innerWidth),'phone overflow');
  await ready(pa); await pa.fill('#question','SYNTHETIC_PROVIDER_FAILURE'); await pa.click('#send');
  await pa.waitForFunction(()=>document.querySelector('#messages').textContent === '' && document.querySelector('#question').disabled);
  assert.match(await pa.locator('#status').textContent(), /The request could not be completed/, 'close hid sanitized failure reason');
  assert.equal(requests.length,6);
  assert.deepEqual(errors,[]);console.log('PASS: five successful real Agent requests plus one refused fake-provider request; two-visitor isolation, follow-up, refresh/new-session isolation, full reference retention after deletion, no memory/tools, store:false, exact visible instructions, escaped output, phone geometry, retained sanitized failure with discarded chat.');
 } finally {
  if(browser)await browser.close();
  if(host&&host.exitCode===null){host.kill('SIGTERM');const timeout=new Promise(resolve=>{const t=setTimeout(()=>resolve(null),20000);t.unref()});const result=await Promise.race([closed,timeout]);if(!result){host.kill('SIGKILL');await closed;throw Error('host failed graceful joined teardown')}assert.equal(result.code,0,'host shutdown: '+stderr);}
  await new Promise(resolve=>provider.close(resolve));fs.rmSync(scratch,{recursive:true,force:true});
 }
})().catch(e=>{console.error(e);console.error('host:',stderr);process.exitCode=1});
