import {chromium} from 'playwright';
import {WebSocket} from 'ws';
import fs from 'node:fs';
import http from 'node:http';
import {spawn} from 'node:child_process';
const base='/Users/bill/projects/ensemble/solutions/edition-2/';
const web=process.env.CH09_WEB || base+'main/gui/web/gui';
const evidence=process.env.CH09_EVIDENCE || base+'evidence/ch09/speed-r1.json';
const scratch=fs.mkdtempSync('/tmp/ch09-speed-');
const fake=http.createServer(async(req,res)=>{
 for await(const chunk of req){}
 res.writeHead(200,{'Content-Type':'text/event-stream'});
 res.write('data: '+JSON.stringify({model:'gpt-4o-mini',choices:[{delta:{content:'Speech speed check.'}}]})+'\n\n');
 res.end('data: '+JSON.stringify({choices:[{delta:{},finish_reason:'stop'}],usage:{prompt_tokens:1,completion_tokens:1}})+'\n\ndata: [DONE]\n\n');
});
await new Promise(r=>fake.listen(0,'127.0.0.1',r));
let child,peer;
async function start(){
 child=spawn('/tmp/ensemble-ch09-browser/ensemble-gui',['--port','19393','--gui-dir',web],{cwd:scratch,env:{...process.env,LLM_VENDOR:'openai',LLM_MODEL:'gpt-4o-mini',LLM_API_KEY:'fake',LLM_BASE_URL:'http://127.0.0.1:'+fake.address().port},stdio:['ignore','ignore','ignore']});
 for(let i=0;i<100;i++){try{if((await fetch('http://localhost:19393')).ok)break}catch{}await new Promise(r=>setTimeout(r,30))}
 peer=new WebSocket('ws://localhost:19393/ws');await new Promise((resolve,reject)=>{peer.once('open',resolve);peer.once('error',reject)});
 return await message('current_settings',{type:'subscribe'});
}
function message(type,command){return new Promise((resolve,reject)=>{
 const timeout=setTimeout(()=>reject(Error('missing '+type)),5000);
 const receive=bytes=>{const value=JSON.parse(bytes);if(value.type===type){clearTimeout(timeout);peer.off('message',receive);resolve(value.settings)}};
 peer.on('message',receive);peer.send(JSON.stringify(command));
})}
async function stop(){peer?.close();if(child?.exitCode===null){const done=new Promise(r=>child.once('exit',r));child.kill('SIGTERM');await done}}
const browser=await chromium.launch({executablePath:'/Applications/Google Chrome.app/Contents/MacOS/Google Chrome',headless:false,args:['--autoplay-policy=no-user-gesture-required']});
let result={observations:[]};
try{
 await start();
 const saved=await message('settings_changed',{type:'update_settings',settings:{tts_speed:1.75}});
 await stop();const restored=await start();
 result.saved=saved.tts_speed;result.restored=restored.tts_speed;
 const page=await browser.newPage();
 await page.addInitScript(()=>{
  globalThis.speechEvidence=[];const speak=speechSynthesis.speak.bind(speechSynthesis);
  speechSynthesis.speak=u=>{speechEvidence.push({type:'queued',text:u.text,rate:u.rate});for(const type of ['start','end','error'])u.addEventListener(type,e=>speechEvidence.push({type,text:u.text,error:e.error}));return speak(u)};
 });
 await page.goto('http://localhost:19393');await page.getByRole('tab',{name:'Settings',exact:true}).click();
 for(const speed of [1.75,1.2345]){
  if(speed!==1.75)await message('settings_changed',{type:'update_settings',settings:{tts_speed:speed}});
  await page.waitForFunction(speed=>Number(document.querySelector('#speed-value').value)===speed,speed);
  const controls=await page.evaluate(()=>({slider:Number(document.querySelector('#tts-speed').value),label:Number(document.querySelector('#speed-value').value)}));
  await page.bringToFront();await page.evaluate(()=>{speechEvidence.length=0});
  await page.locator('#input').fill('Speak at the saved speed');await page.locator('#input').press('Enter');
  await page.waitForFunction(()=>speechEvidence.some(e=>e.type==='queued'&&e.text==='Speech speed check.'));
  await page.waitForFunction(()=>speechEvidence.some(e=>['end','error'].includes(e.type)&&e.text==='Speech speed check.'),null,{timeout:15000});
  const speech=await page.evaluate(()=>speechEvidence);
  result.observations.push({speed,...controls,speech});
 }
 // Check after capture: the restored defect must retain both its rounded DOM
 // value and actual native rate, rather than fail before speech is exercised.
 result.result=result.saved===1.75&&result.restored===1.75&&result.observations.every(o=>o.slider===o.speed&&o.label===o.speed&&o.speech.some(e=>e.type==='queued'&&e.text==='Speech speed check.'&&Math.abs(e.rate-o.speed)<0.00001))?'PASS':'FAIL';
 if(result.result!=='PASS')process.exitCode=1;
}catch(error){result.result='FAIL';result.error=error.message;process.exitCode=1}
finally{await browser.close();await stop();fake.close();fs.rmSync(scratch,{recursive:true,force:true});fs.writeFileSync(evidence,JSON.stringify(result,null,2));console.log(JSON.stringify(result))}
