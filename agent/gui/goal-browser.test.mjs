import test from 'node:test';
import assert from 'node:assert/strict';
import { createServer } from 'node:http';
import { spawn, execFileSync } from 'node:child_process';
import { mkdtemp, mkdir, writeFile, readFile, rm } from 'node:fs/promises';
import { tmpdir } from 'node:os';
import { join, resolve } from 'node:path';
import { once } from 'node:events';
import { chromium } from 'playwright';

test('isolated real GUI projects configured goals across reconnect and selection', {timeout:90000}, async()=>{
 const dir=await mkdtemp(join(tmpdir(),'ensemble-goals-browser-'));
 const binary=process.env.GOALS_TEST_GUI || join(dir,'gui');
 let browser,child,exited,stderr='';
 const requests=[];
 const provider=createServer(async(req,res)=>{
  let body='';for await(const chunk of req)body+=chunk;
  requests.push(JSON.parse(body));
  const content=requests.length===1
   ?[{type:'tool_use',id:'load',name:'load_skill',input:{name:'goals'}}]
   :requests.length===2
    ?[{type:'tool_use',id:'configure',name:'goal_control',input:{file:join(dir,'GOAL.md'),enabled:true,idle_seconds:60}}]
    :[{type:'text',text:'Configured the isolated goal.'}];
  res.setHeader('content-type','application/json');res.end(JSON.stringify({content,usage:{input_tokens:1,output_tokens:1}}));
 });
 try {
  if(!process.env.GOALS_TEST_GUI)execFileSync('go',['build','-o',binary,'./cmd']);
  await mkdir(join(dir,'cr'));await writeFile(join(dir,'cr/settings.json'),JSON.stringify({tts_enabled:false}));
  await mkdir(join(dir,'skills/base'),{recursive:true});await mkdir(join(dir,'skills/goals'));
  await writeFile(join(dir,'skills/base/SKILL.md'),'---\nname: base\ntype: primary\ndescription: Isolated goal test.\nloadable-skills: goals\ntools: load_skill unload_skill\n---\nUse only the requested goal controls.\n');
  await writeFile(join(dir,'skills/goals/SKILL.md'),await readFile('../skills/goals/SKILL.md'));
  await writeFile(join(dir,'GOAL.md'),'ISOLATED_BROWSER_GOAL: inspect current work; do not repeat actions.');
  provider.listen(0,'127.0.0.1');await once(provider,'listening');
  const probe=createServer();probe.listen(0,'127.0.0.1');await once(probe,'listening');const port=probe.address().port;await new Promise(r=>probe.close(r));
  child=spawn(binary,['--port',String(port),'--gui-dir',resolve('web/gui'),'--no-stream'],{cwd:dir,env:{...process.env,LLM_VENDOR:'anthropic',LLM_MODEL:'fake',LLM_API_KEY:'dummy',LLM_BASE_URL:`http://127.0.0.1:${provider.address().port}`,EN_SKILLS_DIR:join(dir,'skills'),EN_PRIMARY_SKILL:'base'},stdio:['pipe','ignore','pipe']});
  exited=new Promise(resolve=>child.on('exit',(code,signal)=>resolve({code,signal})));
  child.stderr.on('data',b=>{stderr+=b.toString();});
  await new Promise((resolve,reject)=>{const timer=setTimeout(()=>reject(Error('GUI startup timeout: '+stderr)),15000);child.stderr.on('data',()=>{if(stderr.includes('Open http')){clearTimeout(timer);resolve();}});child.on('exit',()=>{clearTimeout(timer);reject(Error('GUI exited: '+stderr));});});
  browser=await chromium.launch({headless:true,executablePath:process.env.CHROME_BIN||'/Applications/Google Chrome.app/Contents/MacOS/Google Chrome'});
  const page=await browser.newPage({viewport:{width:1100,height:800}}),errors=[];
  page.on('pageerror',e=>errors.push(e.message));
  await page.addInitScript(()=>{const Native=window.WebSocket;window.WebSocket=class extends Native{constructor(...args){super(...args);window.goalTestSocket=this;}};});
  await page.goto(`http://127.0.0.1:${port}/?speech=transcript`);
  const indicator=page.locator('#goal-status');
  await page.waitForFunction(()=>document.querySelector('#goal-status').textContent==='No goal');
  await page.locator('#input').fill('Load goals and configure the isolated goal.');await page.locator('#input-form button').click();
  await page.waitForFunction(()=>document.querySelector('#goal-status').textContent==='Goal: waiting');
  assert.equal(requests.length,3);
  assert.ok(JSON.stringify(requests[2].system).includes('ISOLATED_BROWSER_GOAL'));
  const title=await indicator.getAttribute('title');
  assert.ok(title.includes('Next reminder:'));assert.ok(title.includes('Idle delay: 60 seconds'));
  assert.equal(await indicator.getAttribute('aria-label'),title);
  await page.reload();await page.waitForFunction(()=>document.querySelector('#goal-status').textContent==='Goal: waiting');
  assert.equal(await indicator.getAttribute('title'),title,'reconnect must not restart core timer');
  await page.locator('#agent-name').fill('peer');await page.locator('#agent-model').fill('fake');await page.locator('#create-agent button').click();
  await page.locator('[data-agent-key="peer"]').click();
  await page.waitForFunction(()=>document.querySelector('#conversation-title').textContent==='peer' && document.querySelector('#goal-status').textContent==='No goal');
  await page.locator('[data-agent-key="main"]').click();await page.waitForFunction(()=>document.querySelector('#goal-status').textContent==='Goal: waiting');
  await page.locator('#interrupt').click();await page.waitForFunction(()=>document.querySelector('#goal-status').textContent==='Goal: reminders off');
  const layout=await indicator.evaluate(e=>{const r=e.getBoundingClientRect(),s=getComputedStyle(e);return {text:e.textContent,width:r.width,height:r.height,x:r.x,y:r.y,font:s.fontSize,color:s.color,background:s.backgroundColor};});
  assert.ok(layout.width>0&&layout.height>0&&layout.x>=0&&layout.y>=0);
  assert.equal(await page.locator('#artifacts').count(),1);assert.equal(await page.locator('#actions').count(),1);
  await page.evaluate(()=>window.goalTestSocket.dispatchEvent(new MessageEvent('message',{data:JSON.stringify({type:'goal_status',agent_key:'peer',goal:{file:'WRONG',phase:'error'}})})));
  assert.equal(await indicator.textContent(),'Goal: reminders off','foreign status must be ignored');
  await page.evaluate(()=>window.goalTestSocket.dispatchEvent(new MessageEvent('message',{data:JSON.stringify({type:'agent_selected',key:'peer'})})));
  assert.equal(await indicator.textContent(),'Goal: loading','selection clears stale status before replay');
  await page.locator('[data-agent-key="main"]').click();await page.waitForFunction(()=>document.querySelector('#goal-status').textContent==='Goal: reminders off');
  child.kill('SIGTERM');await exited;child=undefined;
  await page.waitForFunction(()=>document.querySelector('#goal-status').textContent==='Goal: disconnected');
  assert.deepEqual(errors,[]);
  console.log(JSON.stringify({requests:requests.length,layout,reconnect:'same due time',selection:'isolated',interrupt:'reminders off',speech:'transcript only',disconnect:'observed after owned server termination'}));
 } finally {
  if(browser)await browser.close();
  if(child){child.stdin.end();const timer=setTimeout(()=>child.kill('SIGTERM'),3000);await exited;clearTimeout(timer);}
  await new Promise(r=>provider.close(r));
  await rm(dir,{recursive:true,force:true});
 }
});
