// Adapt the reviewer's native-speech reproduction into a bounded regression.
// The provider alone is fake; Chrome speech, GUI, MCP, actor and jobs are real.
import {chromium} from '/Users/bill/projects/ensemble-edition-2-revisions/browser-tools/node_modules/playwright/index.mjs';
import fs from 'node:fs';
import http from 'node:http';
import {spawn} from 'node:child_process';
const base='/Users/bill/projects/ensemble/solutions/edition-2/';
const web=process.env.CH14_WEB||base+'main/gui/web/gui';
const out=process.env.CH14_OUT||base+'evidence/ch14/native-speech.json';
const scratch=fs.mkdtempSync('/tmp/ensemble-ch13-speech-');
let release1,release2;
const gate1=new Promise(r=>release1=r),gate2=new Promise(r=>release2=r),requests=[];
const fake=http.createServer(async(req,res)=>{
 let body='';for await(const c of req)body+=c;requests.push(JSON.parse(body));
 const n=requests.length;if(n===1)await gate1;if(n===2)await gate2;
 const content=n===1?[{type:'tool_use',id:'input-1',name:'gui_input',input:{selector:'#input',text:'MCP submitted'}}]:n===2?[{type:'tool_use',id:'click-2',name:'gui_click',input:{selector:'#input-form button'}}]:[{type:'text',text:'Speech and submit complete.'}];
 res.setHeader('Content-Type','application/json');res.end(JSON.stringify({content,usage:{input_tokens:1,output_tokens:1}}));
});
await new Promise(r=>fake.listen(0,'127.0.0.1',r));
const child=spawn('/tmp/ensemble-ch14-gui',['--port','19413','--gui-dir',web,'--gui-debug','--skills-dir',base+'main/skills','--save',scratch+'/save.json','--no-stream'],{cwd:scratch,env:{PATH:process.env.PATH,HOME:process.env.HOME,LLM_VENDOR:'anthropic',LLM_MODEL:'fake',LLM_API_KEY:'fake',LLM_BASE_URL:'http://127.0.0.1:'+fake.address().port},stdio:['pipe','ignore','pipe']});
let stderr='';child.stderr.on('data',d=>stderr+=d);
const browser=await chromium.launch({executablePath:'/Applications/Google Chrome.app/Contents/MacOS/Google Chrome',headless:false,args:['--autoplay-policy=no-user-gesture-required']});
const result={checks:[]},check=(ok,name)=>{if(!ok)throw Error(name);result.checks.push(name)};
let page;
try{
 for(let i=0;i<100;i++){try{if((await fetch('http://localhost:19413')).ok)break}catch{}await new Promise(r=>setTimeout(r,30));}
 page=await browser.newPage();page.setDefaultTimeout(5000);
 await page.addInitScript(()=>{window.nativeEvents=[];const speak=speechSynthesis.speak.bind(speechSynthesis);speechSynthesis.speak=u=>{for(const event of ['start','end','error'])u.addEventListener(event,e=>window.nativeEvents.push({event,text:u.text,error:e.error||null}));return speak(u);};});
 await page.goto('http://localhost:19413');await page.waitForFunction(()=>document.querySelector('#state').textContent!=='connecting');
 await page.locator('#input').fill('Go.');await page.locator('#input').press('Enter');
 const speaker=page.getByRole('button',{name:'Read artifact aloud',exact:true}).first();
 await speaker.click();
 await page.waitForFunction(()=>window.nativeEvents.some(e=>e.event==='end'&&e.text==='Go.'),{},{timeout:12000});
 await page.waitForFunction(()=>document.querySelector('#pause-state').textContent==='');
 check(true,'native speech completion unpauses with empty input');
 // Whitespace is still human input under the existing nonempty-field rule.
 await page.locator('#input').pressSequentially(' ');
 await page.waitForFunction(()=>document.querySelector('#pause-state').textContent==='paused');
 await speaker.click();
 await page.waitForFunction(()=>window.nativeEvents.filter(e=>e.event==='end'&&e.text==='Go.').length===2,{},{timeout:12000});
 check(await page.locator('#pause-state').textContent()==='paused','native speech completion preserves human whitespace typing pause');
 const other=await browser.newPage();await other.goto('http://localhost:19413');
 await other.waitForFunction(()=>document.querySelector('#state').textContent!=='connecting');
 await other.locator('#resume').click();
 await page.waitForFunction(()=>document.querySelector('#pause-state').textContent==='');
 check(await page.locator('#input').inputValue()===' ','other tab explicitly unpauses Agent while human text remains');
 await page.locator('#input').fill('');
 await other.close(); // Avoid two simultaneous native speech queues in this probe.
 await page.bringToFront();
 release1();
 await page.waitForFunction(()=>document.querySelector('#input').value==='MCP submitted');
 // The manual utterance queues behind any actual tool announcement, so its end
 // proves the native speech queue has passed through the problematic callback.
 await speaker.click();
 await page.waitForFunction(()=>window.nativeEvents.filter(e=>e.event==='end'&&e.text==='Go.').length===3,{},{timeout:15000});
 await page.waitForFunction(()=>document.querySelector('#pause-state').textContent==='',{},{timeout:2000});
 check(await page.locator('#input').inputValue()==='MCP submitted','speech completion releases pause while programmatic text remains');
 release2();
 await page.getByText('Speech and submit complete.',{exact:true}).first().waitFor({timeout:15000});
 check(await page.locator('#input').inputValue()==='','actual next MCP submit clears field after speech');
 // Check the native gate itself, independently of the text in request arguments.
 check((await page.evaluate(()=>window.nativeEvents)).some(e=>e.event==='start'&&e.text === 'gui input'),'actual gui_input announcement used native speech');
 await page.waitForFunction(()=>window.nativeEvents.some(e=>e.event==='end'&&e.text==='Speech and submit complete.'),{},{timeout:12000});
 await page.waitForFunction(()=>document.querySelector('#pause-state').textContent==='');
 check(true,'final speech completes without retaining typing pause');
 result.result='PASS';
}catch(error){result.result='FAIL';result.error=error.message;process.exitCode=1;}
finally{
 if(page){result.nativeEvents=await page.evaluate(()=>window.nativeEvents).catch(()=>[]);result.finalPause=await page.locator('#pause-state').textContent().catch(()=>null);result.finalInput=await page.locator('#input').inputValue().catch(()=>null);}
 release1();release2();await browser.close();
 if(child.exitCode===null){const done=new Promise(r=>child.once('exit',r));child.kill('SIGTERM');await done;}
 fake.close();result.requests=requests;result.stderr=stderr;
 if(fs.existsSync(scratch+'/gui.log'))result.trace=fs.readFileSync(scratch+'/gui.log','utf8');
 fs.writeFileSync(out,JSON.stringify(result,null,2));fs.rmSync(scratch,{recursive:true,force:true});
 console.log(JSON.stringify({result:result.result,checks:result.checks,error:result.error,finalPause:result.finalPause,finalInput:result.finalInput}));
}
