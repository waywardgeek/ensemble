// Real Chrome, actual GUI command, skill activation and Agent loop. Only the
// HTTP vendor is fake in the deterministic input/pause regression.
import {chromium} from '/Users/bill/projects/ensemble-edition-2-revisions/browser-tools/node_modules/playwright/index.mjs';
import fs from 'node:fs';
import http from 'node:http';
import {spawn} from 'node:child_process';
const base='/Users/bill/projects/ensemble/solutions/edition-2/';
const out=process.env.CH13_OUT || base+'evidence/ch13/browser.json';
const web=process.env.CH13_WEB || base+'main/gui/web/gui';
const vendor=process.env.CH13_PROVIDER;
const providers={anthropic:['directClaudeAPIKey','claude-haiku-5-5'],openai:['directOpenAIAPIKey','gpt-4o-mini'],gemini:['directGeminiAPIKey','gemini-3.8-flash']};
const scratch=fs.mkdtempSync('/tmp/ensemble-ch13-browser-');
const requests=[];
let fake,key='',model='fake';
const env={PATH:process.env.PATH,HOME:process.env.HOME,LLM_VENDOR:vendor||'anthropic',LLM_MODEL:model,LLM_API_KEY:'fake'};
if(vendor){
 const settings=JSON.parse(fs.readFileSync(process.env.HOME+'/.cr/settings.json','utf8'));
 model=providers[vendor][1];key=settings[providers[vendor][0]];
 env.LLM_MODEL=model;env.LLM_API_KEY=key;
}else{
 fake=http.createServer(async(req,res)=>{
  let body='';for await(const chunk of req)body+=chunk;
  requests.push(JSON.parse(body));
  const call=(name,input)=>({type:'tool_use',id:'call-'+requests.length,name,input});
  const content=requests.length===1?[call('gui_input',{selector:'#input',text:'MCP submitted'})]:requests.length===2?[call('gui_click',{selector:'#input-form button'})]:[{type:'text',text:'Input and submit completed.'}];
  res.setHeader('Content-Type','application/json');res.end(JSON.stringify({content,usage:{input_tokens:1,output_tokens:1}}));
 });
 await new Promise(r=>fake.listen(0,'127.0.0.1',r));env.LLM_BASE_URL='http://127.0.0.1:'+fake.address().port;
}
const child=spawn(process.env.CH13_GUI||'/tmp/ensemble-ch13-gui',['--port','19413','--gui-dir',web,'--gui-debug','--skills-dir',base+'main/skills','--save',scratch+'/save.json',...(vendor?[]:['--no-stream'])],{cwd:scratch,env,stdio:['pipe','ignore','pipe']});
let stderr='';child.stderr.on('data',d=>stderr+=d);
const browser=await chromium.launch({executablePath:'/Applications/Google Chrome.app/Contents/MacOS/Google Chrome',headless:false,args:['--autoplay-policy=no-user-gesture-required']});
const result={vendor:vendor||'external HTTP fake',model,checks:[]};
const check=(ok,name)=>{if(!ok)throw Error(name);result.checks.push(name)};
const safe=s=>key?s.replaceAll(key,'[REDACTED]'):s;
try{
 for(let i=0;i<100;i++){try{if((await fetch('http://localhost:19413')).ok)break}catch{}await new Promise(r=>setTimeout(r,30))}
 const page=await browser.newPage({viewport:{width:1440,height:900}});
 const errors=[];page.on('pageerror',e=>errors.push(e.message));
 await page.goto('http://localhost:19413');
 await page.waitForFunction(()=>document.querySelector('#state').textContent!=='connecting');
 for(let i=0;i<100;i++){if(fs.readFileSync(scratch+'/gui.log','utf8').includes('gui_snapshot'))break;await new Promise(r=>setTimeout(r,30))}
 await page.getByRole('tab',{name:'Settings',exact:true}).click();await page.locator('#voice').uncheck();await page.getByRole('tab',{name:'Agents',exact:true}).click();
 // A real keystroke still pauses; submission releases that local cause.
 await page.locator('#input').pressSequentially('x');
 await page.waitForFunction(()=>document.querySelector('#pause-state').textContent==='paused');
 check(true,'human typing still pauses Agent');
 await page.locator('#input').fill('');
 await page.waitForFunction(()=>document.querySelector('#pause-state').textContent==='');
 const prompt=vendor?'Use only browser tools. Click the Settings tab using gui_click, set font size to 20 using gui_input, then inspect the next automatic snapshot and report the observed font size. Unload gui-debug using unload_skill, then reply CH13_BROWSER_DONE. Do not read files or run commands.':'Exercise input then submit through browser tools.';
 const directedPrompt=vendor==='openai'?prompt+' The observed font-size control is #font-size (a select, not a number input). Wait for the Settings snapshot before editing, and repair any tool error before unloading.':prompt;
 result.prompt=directedPrompt;
 await page.locator('#input').fill(directedPrompt);await page.locator('#input').press('Enter');
 if(vendor){
  await page.waitForFunction(()=>[...document.querySelectorAll('#artifacts .artifact--text')].some(e=>e.innerText.includes('CH13_BROWSER_DONE')),{},{timeout:60000});
  check(await page.locator('#settings-tab').getAttribute('aria-selected')==='true','live model clicked actual Settings tab');
  check(await page.locator('#font-size').inputValue()==='20'&&await page.evaluate(()=>getComputedStyle(document.body).fontSize)==='20px','live model changed visible/effective font size');
 }else{
  await page.getByText('Input and submit completed.',{exact:true}).first().waitFor({timeout:6000});
  check(await page.locator('#input').inputValue()==='','MCP input followed by submit clears actual field');
  await page.waitForFunction(()=>document.querySelector('#pause-state').textContent==='',{},{timeout:6000});
  check(true,'MCP input did not strand Agent pause');
  check(requests.length>=3,'MCP submit remained reachable through actual dispatch');
  const observed=requests[1].messages.flatMap(m=>m.content||[]).filter(p=>p.type==='text').map(p=>p.text).join('\n');
  check(observed.includes('\"value\":\"MCP submitted\"'),'next snapshot observed actual programmatic input value');
 }
 check(errors.length===0,'no browser runtime error');
 result.artifacts=await page.locator('#artifacts').innerText();
 result.trace=safe(fs.readFileSync(scratch+'/gui.log','utf8'));
 result.result='PASS';result.browser=await browser.version();
}catch(error){result.result='FAIL';result.error=error.message;result.stderr=safe(stderr);process.exitCode=1}
finally{
 await browser.close();
 if(child.exitCode===null){const done=new Promise(r=>child.once('exit',r));child.kill('SIGTERM');await done}
 if(fake)fake.close();
 result.requests=vendor?fs.readFileSync(scratch+'/api.log','utf8').split('\n').filter(Boolean).map(x=>JSON.parse(safe(x))).filter(x=>x.event==='request').map(x=>JSON.parse(x.data)):requests;
 if(fs.existsSync(scratch+'/event.log'))result.events=safe(fs.readFileSync(scratch+'/event.log','utf8'));
 fs.writeFileSync(out,JSON.stringify(result,null,2));fs.rmSync(scratch,{recursive:true,force:true});
 console.log(JSON.stringify({vendor:result.vendor,result:result.result,checks:result.checks,error:result.error}));
}
