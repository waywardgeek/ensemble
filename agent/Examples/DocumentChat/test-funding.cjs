// Public-host funding checks; nonexistent credentials, no provider requests.
const assert=require('node:assert/strict'),fs=require('node:fs'),os=require('node:os'),path=require('node:path');
const {spawn}=require('node:child_process');const net=require('node:net');
const {chromium}=require(process.env.PLAYWRIGHT_MODULE||'playwright');
const dll=path.resolve(process.argv[2]),sdk=process.env.DOTNET;
const dir=fs.mkdtempSync(path.join(os.tmpdir(),'document-funding-'));
fs.writeFileSync(path.join(dir,'doc.txt'),'[PDF PAGE 1]\nSynthetic reference.');
const env={PATH:'/usr/bin:/bin',HOME:dir,TMPDIR:dir,DOTNET_CLI_HOME:dir,DOCUMENT_CHAT_API_KEY:'must-not-be-used'};
function start(args){const p=spawn(sdk,[dll,'--document',path.join(dir,'doc.txt'),...args],{cwd:dir,env});p.stdout.resume();p.stderr.resume();p.done=new Promise(r=>p.on('exit',(code,signal)=>r({code,signal})));return p;}
async function bounded(p){let timer;try{return await Promise.race([p.done,new Promise((_,r)=>timer=setTimeout(()=>r(Error('host timeout')),10000))]);}finally{clearTimeout(timer);}}
(async()=>{let host,browser;try{
 for(const args of [ ['--billing','unknown'],['--billing','codex'],['--billing','codex','--codex-auth-file',path.join(dir,'missing'),'--base-url','http://127.0.0.1:1'],['--codex-auth-file',path.join(dir,'missing')]]){
  const p=start(args);try{assert.notEqual((await bounded(p)).code,0);}finally{if(p.exitCode===null&&p.signalCode===null){p.kill('SIGTERM');await p.done;}}
 }
 const reservation=net.createServer();await new Promise(r=>reservation.listen(0,'127.0.0.1',r));const port=reservation.address().port;await new Promise(r=>reservation.close(r));
 host=start(['--billing','codex','--codex-auth-file',path.join(dir,'missing'),'--port',String(port)]);
 const base=`http://127.0.0.1:${port}`;let about;
 for(let i=0;i<100;i++){try{const r=await fetch(base+'/about');if(r.ok){about=await r.json();break;}}catch{}await new Promise(r=>setTimeout(r,100));}
 assert.equal(about?.billing,'codex');assert(!JSON.stringify(about).includes(dir));assert(!JSON.stringify(about).includes(env.DOCUMENT_CHAT_API_KEY));
 browser=await chromium.launch({headless:true,channel:'chrome'});const page=await browser.newPage();await page.goto(base);
 await page.waitForFunction(()=>!document.querySelector('#question').disabled);
 await page.fill('#question','Synthetic funding refusal');await page.click('#send');
 await page.waitForFunction(()=>document.querySelector('#status').textContent.includes('request could not be completed'));
 assert.equal(await page.locator('#messages').textContent(),'');
 assert(!(await page.locator('body').textContent()).includes(dir));
 console.log('PASS: four invalid configurations; Codex missing-auth refusal with ambient API key ignored; public metadata sanitized.');
 }finally{await browser?.close();if(host){host.kill('SIGTERM');assert.equal((await bounded(host)).code,0);}fs.rmSync(dir,{recursive:true,force:true});}})().catch(e=>{console.error(e);process.exitCode=1;});
