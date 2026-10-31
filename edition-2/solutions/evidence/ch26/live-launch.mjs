import assert from 'node:assert/strict';
import {chromium} from '/Users/bill/projects/ensemble/solutions/edition-2/main/gui/node_modules/playwright/index.mjs';
import {spawn} from 'node:child_process';
import {mkdtemp,writeFile,readFile,cp,readdir} from 'node:fs/promises';
import net from 'node:net';
const evidence='/Users/bill/projects/ensemble/solutions/edition-2/evidence/ch26';
const workspace=await mkdtemp('/tmp/ch26-live-');
const env={...process.env,EN_HOME:'/Users/bill/ensemble-ch19-signin-01v0rmp3/auth',OPENAI_BILLING_MODE:'chatgpt',OPENAI_OIDC_ISSUER:'https://auth.openai.com',LLM_BASE_URL:'https://api.openai.com',LLM_VENDOR:'openai',LLM_MODEL:'gpt-6-astra',OPENAI_REASONING_EFFORT:'high',OPENAI_REASONING_SUMMARY:'auto',EN_RECALL_DISABLED:'1',EN_SKILLS_DIR:'/Users/bill/projects/ensemble/solutions/edition-2/main/skills',EN_PRIMARY_SKILL:'ensemble',ENSEMBLE_WORKSPACE:workspace};
for(const name of ['LLM_API_KEY','OPENAI_API_KEY','ANTHROPIC_API_KEY','GEMINI_API_KEY','EN_DISABLE_STREAMING','EN_SAFE_MODE','EN_SANDBOX_ROOT','LLM_RESOLVED_MODEL'])delete env[name];
const frames=[],errors=[],outcomes=[];let app,browser,page,stdout='',stderr='',exitPromise,port;
async function start(){
 const probe=net.createServer();await new Promise(r=>probe.listen(0,'127.0.0.1',r));port=probe.address().port;await new Promise(r=>probe.close(r));
 app=spawn('/tmp/ch26-gui-live',['--port',String(port)],{cwd:workspace,env,stdio:['pipe','pipe','pipe']});
 app.stdout.on('data',b=>stdout+=b);app.stderr.on('data',b=>stderr+=b);
 exitPromise=new Promise(r=>app.on('exit',(code,signal)=>r({code,signal})));
 for(let i=0;i<120;i++){try{if((await fetch(`http://127.0.0.1:${port}`)).ok)return;}catch{}await new Promise(r=>setTimeout(r,50));}
 throw Error('GUI startup timed out');
}
async function stop(){app.stdin.end();const t=setTimeout(()=>app.kill('SIGTERM'),5000);let result=await exitPromise;clearTimeout(t);outcomes.push(result);}
async function open(){page=await browser.newPage();page.setDefaultTimeout(180000);page.on('pageerror',e=>errors.push(e.message));page.on('websocket',ws=>{ws.on('framereceived',f=>frames.push({direction:'in',frame:JSON.parse(f.payload)}));ws.on('framesent',f=>frames.push({direction:'out',frame:JSON.parse(f.payload)}));});await page.goto(`http://127.0.0.1:${port}/?speech=transcript`);await page.locator('[data-agent-key="main"]').waitFor();}
const prompt=`Work only in ${workspace}. This is one small bounded coding task. Create exactly one child with agent_id helper and delegate implementing slugify.py with slugify(text) that lowercases and joins whitespace-separated words using hyphens (empty input returns empty string), plus a unittest file covering mixed case, repeated whitespace and empty input. The child must run python3 -m unittest -v and submit its actual result. Do not create further children. As parent use get_sub_agent_status and wait_for_agent_change to monitor helper, then join_agents and get_submission. Independently rerun the tests, write fleet-result.txt containing your measured test outcome, and submit a concise final result beginning CH26_LIVE_COMPLETE. Use the inherited model. Do not browse or install packages.`;
let success=false;
try{
 await start();browser=await chromium.launch({headless:true,executablePath:'/Applications/Google Chrome.app/Contents/MacOS/Google Chrome'});await open();
 await page.locator('#agent-name').fill('console');await page.locator('#agent-model').fill('gpt-6-astra');await page.locator('#agent-skill').fill('ensemble');await page.locator('#create-agent button').click();
 await page.locator('[data-agent-key="console"]').click();
 await page.locator('#input').fill(prompt);await page.locator('#input-form button').click();
 await page.locator('[data-agent-key="console/helper"]').waitFor();
 await page.locator('[data-agent-key="console/helper"]').click();
 await page.screenshot({path:evidence+'/live-child.png',fullPage:true});
 await writeFile(evidence+'/live-child-dom.html',await page.content());
 await page.locator('[data-agent-key="console"]').click();
 // Inspect actual submission, not a tool-call argument that merely names the marker.
 await page.waitForFunction(()=>document.querySelector('#actions').textContent.includes('submitted: CH26_LIVE_COMPLETE'));
 assert.ok((await readFile(workspace+'/fleet-result.txt','utf8')).length>0);
 await page.screenshot({path:evidence+'/live-complete.png',fullPage:true});await writeFile(evidence+'/live-complete-dom.html',await page.content());
 await page.locator('#agent-name').fill('spare');await page.locator('#create-agent button').click();await page.locator('[data-agent-key="spare"]').waitFor();await page.getByRole('button',{name:'Delete spare',exact:true}).click();await page.waitForFunction(()=>!document.querySelector('[data-agent-key="spare"]'));
 await page.close();await stop();
 await cp(workspace+'/api.log',evidence+'/live-api.log');await cp(workspace+'/gui.log',evidence+'/live-gui.log');
 await start();await open();await page.locator('[data-agent-key="console"]').click();
 await page.waitForFunction(()=>document.querySelector('#actions').textContent.includes('submitted: CH26_LIVE_COMPLETE'));
 assert.equal(await page.locator('[data-agent-key="console/helper"]').count(),0,'ephemeral child restored as persistent');
 assert.equal(await page.locator('[data-agent-key="spare"]').count(),0);
 await page.screenshot({path:evidence+'/live-restarted.png',fullPage:true});await writeFile(evidence+'/live-restarted-dom.html',await page.content());
 await page.close();await stop();success=true;
}catch(e){errors.push(e.message);console.log(JSON.stringify({stage:'live failure',error:e.message,workspace}));}
finally{
 if(app&&app.exitCode===null&&app.signalCode===null){app.kill('SIGTERM');await exitPromise;}
 if(browser)await browser.close();
 // These are this isolated application's outputs, never its external auth store.
 for(const name of await readdir(workspace)){if(name==='api.log'||name==='gui.log'){try{await cp(workspace+'/'+name,evidence+'/live-final-'+name);}catch{}}else if(['agents','cr','ensemble.json','save.json','slugify.py','test_slugify.py','fleet-result.txt','event.log','debug.log'].includes(name))await cp(workspace+'/'+name,evidence+'/live-artifacts/'+name,{recursive:true});}
 await writeFile(evidence+'/live-frames.json',JSON.stringify(frames,null,2));await writeFile(evidence+'/live-stdout.txt',stdout);await writeFile(evidence+'/live-stderr.txt',stderr);
 await writeFile(evidence+'/live-result.json',JSON.stringify({success,workspace,prompt,errors,outcomes,frames:frames.length},null,2));
 console.log(JSON.stringify({success,workspace,errors,outcomes,frames:frames.length}));process.exitCode=success?0:1;
}
