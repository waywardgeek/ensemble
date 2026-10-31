import {chromium} from 'playwright';
import fs from 'node:fs';
import http from 'node:http';
import {spawn} from 'node:child_process';
const base='/Users/bill/projects/ensemble/solutions/edition-2/';
const evidence=process.env.CH09_EVIDENCE || base+'evidence/ch09/browser.json';
const web=process.env.CH09_WEB || base+'main/gui/web/gui';
const scratch=fs.mkdtempSync('/tmp/ensemble-ch09-');
let release;
const gate=new Promise(r=>release=r);
const fake=http.createServer(async(req,res)=>{
 let body='';for await(const chunk of req)body+=chunk;
 const request=JSON.parse(body);
 const last=request.messages.at(-1);
 res.writeHead(200,{'Content-Type':'text/event-stream'});
 const frame=delta=>res.write('data: '+JSON.stringify({model:'gpt-4o-mini',choices:[{delta}]})+'\n\n');
 if(last.role==='tool') frame({content:'The tool completed.'});
 else if(JSON.stringify(last).includes('TOOL')) {
  frame({content:'Working now. '}); await gate;
  frame({tool_calls:[{index:0,id:'one',type:'function',function:{name:'run_command',arguments:'{"command":"printf PANE_RESULT","yield_time_ms":1000}'}}]});
 } else frame({content:'Native speech uses the saved speed.'});
 res.end('data: '+JSON.stringify({choices:[{delta:{},finish_reason:'stop'}],usage:{prompt_tokens:1,completion_tokens:1}})+'\n\ndata: [DONE]\n\n');
});
await new Promise(r=>fake.listen(0,'127.0.0.1',r));
let child;
async function start(){
 child=spawn('/tmp/ensemble-ch09-browser/ensemble-gui',['--port','19391','--gui-dir',web],{cwd:scratch,env:{...process.env,LLM_VENDOR:'openai',LLM_MODEL:'gpt-4o-mini',LLM_API_KEY:'fake',LLM_BASE_URL:'http://127.0.0.1:'+fake.address().port},stdio:['pipe','ignore','ignore']});
 for(let i=0;i<100;i++){try{if((await fetch('http://localhost:19391')).ok)return}catch{}await new Promise(r=>setTimeout(r,30))}throw Error('server startup timeout');
}
async function stop(){if(child?.exitCode===null){const ended=new Promise(r=>child.once('exit',r));child.kill('SIGTERM');await ended}}
const browser=await chromium.launch({executablePath:'/Applications/Google Chrome.app/Contents/MacOS/Google Chrome',headless:false,args:['--autoplay-policy=no-user-gesture-required']});
let output={checks:[]};
const check=(value,name)=>{if(!value)throw Error(name);output.checks.push(name)};
try{
 await start();
 const context=await browser.newContext({viewport:{width:1440,height:900}});
 await context.addInitScript(()=>{
  globalThis.speechEvidence=[];
  const speak=speechSynthesis.speak.bind(speechSynthesis);
  speechSynthesis.speak=utterance=>{
   speechEvidence.push({type:'queued',text:utterance.text,rate:utterance.rate});
   for(const type of ['start','end','error'])utterance.addEventListener(type,event=>speechEvidence.push({type,text:utterance.text,error:event.error}));
   return speak(utterance);
  };
 });
 const page=await context.newPage(), second=await context.newPage();
 await page.goto('http://localhost:19391');await second.goto('http://localhost:19391');
 await page.getByRole('tab',{name:'Settings',exact:true}).click();
 check(await page.getByRole('tab',{name:'Settings',exact:true}).getAttribute('aria-selected')==='true','selected tab exposed');
 await page.locator('#voice').uncheck();await page.locator('#theme').selectOption('light');
 await second.waitForFunction(()=>document.body.dataset.theme==='light'&&!document.querySelector('#voice').checked);
 check(true,'theme and false TTS broadcast to second tab');
 await page.locator('#font-size').selectOption('20');
 await second.waitForFunction(()=>getComputedStyle(document.body).fontSize==='20px');
 check(true,'font setting changes effective layout across tabs');
 await page.locator('#theme').selectOption('system');await page.emulateMedia({colorScheme:'dark'});
 await page.waitForFunction(()=>document.body.dataset.theme==='dark');await page.emulateMedia({colorScheme:'light'});
 await page.waitForFunction(()=>document.body.dataset.theme==='light');check(true,'system theme follows OS changes');
 for(const [bar,pane,dx] of [['#sidebar-drag','.sidebar',45],['#actions-drag','.actions',-45]]){
  const before=await page.locator(pane).boundingBox(),box=await page.locator(bar).boundingBox();
  await page.mouse.move(box.x+3,box.y+200);await page.mouse.down();await page.mouse.move(box.x+3+dx,box.y+200,{steps:5});await page.mouse.up();
  const after=await page.locator(pane).boundingBox();check(after.width>before.width+30,bar+' resizes pane');
 }
 await page.locator('#input').fill('TOOL');await page.locator('#input').press('Enter');
 await page.waitForFunction(()=>document.querySelector('#agent-tree li').dataset.state==='in_flight');
 check(await page.locator('#agent-tree > li').count()===1 && await page.locator('.agent-indicator').evaluate(node=>getComputedStyle(node).animationName)==='pulse','single agent shows active state');
 release();await page.locator('#actions').getByText(/PANE_RESULT/).first().waitFor();await page.locator('#artifacts').getByText('The tool completed.',{exact:true}).waitFor();
 check((await page.locator('#actions').innerText()).includes('run_command')&&!(await page.locator('#artifacts').innerText()).includes('PANE_RESULT'),'tool artifacts route to actions');
 check(!(await page.locator('#actions').innerText()).includes('The tool completed.'),'assistant text stays in chat');
 check(!(await page.evaluate(()=>speechEvidence.some(e=>e.type==='queued'&&e.text.trim()))),'disabled TTS suppresses speech');
 output.phase='speed control';
 await page.locator('#tts-speed').focus();await page.locator('#tts-speed').press('Home');
 await second.waitForFunction(()=>document.querySelector('#tts-speed').value==='0.5');
 await page.locator('#tts-speed').press('End');
 await second.waitForFunction(()=>document.querySelector('#tts-speed').value==='2');
 output.phase='native speech';
 await page.locator('#voice').check();await second.waitForFunction(()=>document.querySelector('#voice').checked);await page.bringToFront();await page.locator('#input').fill('Speak');await page.locator('#input').press('Enter');
 await page.waitForFunction(()=>speechEvidence.some(e=>e.type==='queued'&&e.text==='Native speech uses the saved speed.'&&Math.abs(e.rate-2)<0.001));
 await page.waitForFunction(()=>speechEvidence.some(e=>['end','error'].includes(e.type)&&e.text==='Native speech uses the saved speed.'),null,{timeout:30000});
 check(true,'native speech consumes persisted speed');
 output.speech=await page.evaluate(()=>speechEvidence);
 await page.locator('#voice').uncheck();
 await second.waitForFunction(()=>!document.querySelector('#voice').checked);
 output.phase='restart';
 await stop();await start();await page.reload();await page.getByRole('tab',{name:'Settings',exact:true}).click();
 await page.waitForFunction(()=>document.querySelector('#theme').value==='system'&&document.querySelector('#font-size').value==='20'&&document.querySelector('#tts-speed').value==='2'&&!document.querySelector('#voice').checked);
 check(true,'visible controls restore all four values after process restart');
 await page.screenshot({path:evidence.replace('.json','.png')});
 output.result='PASS';output.browser=await browser.version();
}catch(error){output.result='FAIL';output.error=error.message;output.stack=error.stack;output.speech=await browser.contexts()[0]?.pages()[0]?.evaluate(()=>speechEvidence);process.exitCode=1}
finally{release();await browser.close();await stop();fake.close();fs.rmSync(scratch,{recursive:true,force:true});fs.writeFileSync(evidence,JSON.stringify(output,null,2));console.log(JSON.stringify(output))}
