import {chromium} from 'playwright';
import fs from 'node:fs';
import os from 'node:os';
import {spawn} from 'node:child_process';
const base='/Users/bill/projects/ensemble/solutions/edition-2/';
const privateSettings=JSON.parse(fs.readFileSync(os.homedir()+'/.cr/settings.json','utf8'));
const key=privateSettings.directOpenAIAPIKey;
if(!key)throw Error('OpenAI credential missing');
// Discovery and inference use the same provider credential, never argv/log output.
const response=await fetch('https://api.openai.com/v1/models',{headers:{Authorization:'Bearer '+key}});
if(!response.ok)throw Error('OpenAI discovery status '+response.status);
const models=(await response.json()).data.map(model=>model.id);
const model=models.find(id=>id==='gpt-4o-mini');
if(!model)throw Error('Supported gpt-4o-mini absent from live discovery');
const scratch=fs.mkdtempSync('/tmp/ensemble-ch09-live-');
let child;
async function start(){
 child=spawn('/tmp/ensemble-ch09-browser/ensemble-gui',['--port','19392','--gui-dir',base+'main/gui/web/gui'],{cwd:scratch,env:{...process.env,LLM_VENDOR:'openai',LLM_MODEL:model,LLM_API_KEY:key},stdio:['pipe','pipe','pipe']});
 child.stdout.resume();child.stderr.resume();
 for(let i=0;i<100;i++){try{if((await fetch('http://localhost:19392')).ok)return}catch{}await new Promise(r=>setTimeout(r,30))}throw Error('server startup timeout');
}
async function stop(){if(child?.exitCode===null){const ended=new Promise(r=>child.once('exit',r));child.kill('SIGTERM');await ended}}
const browser=await chromium.launch({executablePath:'/Applications/Google Chrome.app/Contents/MacOS/Google Chrome',headless:false,args:['--autoplay-policy=no-user-gesture-required']});
let output={model,discovery:'model present in authenticated model list'};
try{
 await start();
 const page=await browser.newPage({viewport:{width:1440,height:900}});
 await page.addInitScript(()=>{
  globalThis.speechEvents=[];const speak=speechSynthesis.speak.bind(speechSynthesis);
  speechSynthesis.speak=u=>{speechEvents.push({type:'queued',text:u.text,rate:u.rate});for(const type of ['start','end','error'])u.addEventListener(type,e=>speechEvents.push({type,text:u.text,error:e.error}));return speak(u)};
 });
 await page.goto('http://localhost:19392');await page.getByRole('tab',{name:'Settings',exact:true}).click();
 await page.locator('#voice').uncheck();await page.locator('#theme').selectOption('light');await page.locator('#font-size').selectOption('18');
 await page.waitForFunction(()=>document.body.dataset.theme==='light'&&getComputedStyle(document.body).fontSize==='18px'&&!document.querySelector('#voice').checked);
 await stop();await start();await page.reload();await page.getByRole('tab',{name:'Settings',exact:true}).click();
 await page.waitForFunction(()=>document.body.dataset.theme==='light'&&getComputedStyle(document.body).fontSize==='18px'&&!document.querySelector('#voice').checked);
 await page.locator('#input').fill('Run run_command with command "sleep 1; printf CH09_LIVE_OK" and yield_time_ms 1500. Then briefly report the exact output. Do not edit files.');
 await page.locator('#input').press('Enter');
 await page.locator('#actions').getByText(/run_command/).first().waitFor({timeout:60000});
 child.stdin.write('Keep the final reply to one sentence.\n');
 await page.locator('#artifacts').getByText('Keep the final reply to one sentence.',{exact:true}).waitFor();
 await page.waitForFunction(()=>document.querySelector('#state').textContent==='idle'&&document.querySelector('#actions').textContent.includes('exit_code: 0'),null,{timeout:60000});
 // Wait for the final post-tool reply, not the transient idle reduction between parts.
 await page.waitForFunction(()=>[...document.querySelectorAll('#artifacts .artifact--text .content')].some(n=>n.textContent.includes('CH09_LIVE_OK')),null,{timeout:30000});
 output.chat=await page.locator('#artifacts').innerText();output.actions=await page.locator('#actions').innerText();
 output.savedControls=await page.evaluate(()=>({theme:document.querySelector('#theme').value,font:document.querySelector('#font-size').value,tts:document.querySelector('#voice').checked}));
 output.terminalHintSeen=output.chat.includes('Keep the final reply to one sentence.');
 await page.bringToFront();await page.locator('#artifacts .artifact--text').last().getByRole('button',{name:'Read artifact aloud'}).click();
 await page.waitForFunction(()=>speechEvents.some(e=>e.type==='start'),null,{timeout:15000});
 await page.waitForFunction(()=>speechEvents.some(e=>['end','error'].includes(e.type)),null,{timeout:30000});
 output.speech=await page.evaluate(()=>speechEvents);
 output.result=output.actions.includes('CH09_LIVE_OK')&&output.terminalHintSeen?'PASS':'FAIL';
 await page.screenshot({path:base+'evidence/ch09/live.png'});
}catch(error){output.result='FAIL';output.error=error.message;process.exitCode=1}
finally{
 await browser.close();await stop();
 // Only selected non-secret artifacts are retained. Wire logs never include
 // authorization headers, but the source checks also precede any live launch.
 fs.writeFileSync(base+'evidence/ch09/live.json',JSON.stringify(output,null,2));
 fs.rmSync(scratch,{recursive:true,force:true});console.log(JSON.stringify(output));
}
