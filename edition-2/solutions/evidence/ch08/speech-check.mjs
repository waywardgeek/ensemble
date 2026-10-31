import {chromium} from 'playwright';import http from 'node:http';import fs from 'node:fs';import {spawn} from 'node:child_process';
const base='/Users/bill/projects/ensemble/solutions/edition-2/';
const web=process.env.CH08_WEB||base+'main/gui/web/gui';
const evidence=process.env.CH08_EVIDENCE||base+'evidence/ch08/speech.json';
const fake=http.createServer(async(req,res)=>{for await(const p of req){};res.writeHead(200,{'Content-Type':'text/event-stream'});const frame=x=>res.write('data: '+JSON.stringify(x)+'\n\n');frame({model:'gpt-4o-mini',choices:[{delta:{content:'A short sentence is spoken. A second sentence follows.'}}]});frame({choices:[{delta:{},finish_reason:'stop'}],usage:{prompt_tokens:1,completion_tokens:1}});res.end('data: [DONE]\n\n')});await new Promise(r=>fake.listen(0,'127.0.0.1',r));
const scratch=fs.mkdtempSync('/tmp/ch08-speech-');const child=spawn('/tmp/ensemble-ch08-browser/ensemble-gui',['--port','19390','--gui-dir',web],{cwd:scratch,env:{...process.env,LLM_VENDOR:'openai',LLM_MODEL:'gpt-4o-mini',LLM_API_KEY:'fake',LLM_BASE_URL:'http://127.0.0.1:'+fake.address().port},stdio:['pipe','ignore','ignore']});
const browser=await chromium.launch({executablePath:'/Applications/Google Chrome.app/Contents/MacOS/Google Chrome',headless:false,args:['--autoplay-policy=no-user-gesture-required']});let output={};
try{
 const page=await browser.newPage();await page.addInitScript(()=>{globalThis.events=[];const speak=speechSynthesis.speak.bind(speechSynthesis);speechSynthesis.speak=u=>{events.push({type:'queued',text:u.text});for(const kind of ['start','end','error'])u.addEventListener(kind,e=>events.push({type:kind,text:u.text,error:e.error}));return speak(u)}});
 for(let i=0;i<100;i++){try{if((await fetch('http://localhost:19390')).ok)break}catch{}await new Promise(r=>setTimeout(r,50))}
 await page.goto('http://localhost:19390');await page.bringToFront();await page.locator('#input').fill('Speak two short sentences');await page.locator('#input').press('Enter');
 await page.waitForFunction(()=>events.some(e=>e.type==='start'),null,{timeout:15000});
 // Interrupt the real external speech service. This does not replace it or
 // invoke the application's own cancel path, so its native onerror must drain.
 await page.evaluate(()=>speechSynthesis.cancel());
 await page.waitForFunction(()=>events.some(e=>e.type==='error'),null,{timeout:10000});
 await page.waitForFunction(()=>document.querySelector('#pause-state').textContent==='',null,{timeout:15000});
 const events=await page.evaluate(()=>globalThis.events);
 if(events.filter(e=>e.type==='queued'&&e.text.includes('second')).length!==1)throw Error('speech error did not continue queue exactly once');
 output={result:'PASS',events};console.log(JSON.stringify(output));
}catch(e){output={result:'FAIL',error:e.message.split('\n')[0],events:await browser.contexts()[0]?.pages()[0]?.evaluate(()=>globalThis.events)};console.log(JSON.stringify(output));process.exitCode=1;}
finally{fs.writeFileSync(evidence,JSON.stringify(output,null,2));await browser.close();child.kill('SIGTERM');await new Promise(r=>child.once('exit',r));fake.close();fs.rmSync(scratch,{recursive:true,force:true})}
