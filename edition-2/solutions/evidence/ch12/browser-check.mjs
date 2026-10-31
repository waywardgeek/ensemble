// Actual browser, GUI tunnel and Agent loop; HTTP is the external provider fake.
import {chromium} from '/Users/bill/projects/ensemble-edition-2-revisions/browser-tools/node_modules/playwright/index.mjs';
import fs from 'node:fs';
import http from 'node:http';
import {spawn} from 'node:child_process';
const base='/Users/bill/projects/ensemble/solutions/edition-2/';
const evidence=process.env.CH12_BROWSER_OUT || base+'evidence/ch12/browser.json';
const web=process.env.CH12_WEB || base+'main/gui/web/gui';
const scratch=fs.mkdtempSync('/tmp/ensemble-ch12-browser-');
const requests=[];
const fake=http.createServer(async(req,res)=>{
 let body=''; for await(const chunk of req) body+=chunk;
 requests.push(JSON.parse(body));
 const call=(name,input)=>({type:'tool_use',id:'call-'+requests.length,name,input});
 const content=requests.length===1?[call('gui_click',{selector:'#settings-tab'})]:requests.length===2?[call('gui_input',{selector:'#font-size',text:'20'})]:[{type:'text',text:'Browser tools completed.'}];
 res.setHeader('Content-Type','application/json');res.end(JSON.stringify({content,usage:{input_tokens:1,output_tokens:1}}));
});
await new Promise(r=>fake.listen(0,'127.0.0.1',r));
let child=spawn(process.env.CH12_GUI || '/tmp/ensemble-ch12-gui',['--port','19412','--gui-dir',web,'--no-stream','--save',scratch+'/save.json'],{cwd:scratch,env:{PATH:process.env.PATH,HOME:process.env.HOME,LLM_VENDOR:'anthropic',LLM_MODEL:'fake',LLM_API_KEY:'fake',LLM_BASE_URL:'http://127.0.0.1:'+fake.address().port},stdio:['pipe','ignore','pipe']});
let stderr='';child.stderr.on('data',d=>stderr+=d);
const browser=await chromium.launch({executablePath:'/Applications/Google Chrome.app/Contents/MacOS/Google Chrome',headless:false,args:['--autoplay-policy=no-user-gesture-required']});
const result={checks:[],requests};
const check=(ok,name)=>{if(!ok)throw Error(name);result.checks.push(name)};
try {
 for(let i=0;i<100;i++){try{if((await fetch('http://localhost:19412')).ok)break}catch{}await new Promise(r=>setTimeout(r,30))}
 const page=await browser.newPage({viewport:{width:1440,height:900}});
 await page.addInitScript(()=>{globalThis.nativeSpeech=[];const speak=speechSynthesis.speak.bind(speechSynthesis);speechSynthesis.speak=u=>{nativeSpeech.push({text:u.text,queued:Date.now()});u.addEventListener('start',()=>nativeSpeech.push({text:u.text,started:Date.now()}));return speak(u)}});
 const errors=[];page.on('pageerror',e=>errors.push(e.message));
 await page.goto('http://localhost:19412');
 // The trace confirms completed discovery before the first human message.
 await page.waitForFunction(()=>document.querySelector('#state').textContent!=='connecting');
 for(let i=0;i<100;i++){if(fs.readFileSync(scratch+'/gui.log','utf8').includes('gui_snapshot'))break;await new Promise(r=>setTimeout(r,30))}
 await page.getByRole('tab',{name:'Settings',exact:true}).click();await page.locator('#voice').uncheck();await page.getByRole('tab',{name:'Agents',exact:true}).click();
 await page.locator('#input').fill('Exercise browser tools');await page.locator('#input').press('Enter');
 await page.getByText('Browser tools completed.',{exact:true}).waitFor({timeout:10000});
 check(await page.locator('#settings-tab').getAttribute('aria-selected')==='true','MCP gui_click changed actual selected tab');
 check(await page.locator('#font-size').inputValue()==='20'&&await page.evaluate(()=>getComputedStyle(document.body).fontSize)==='20px','MCP gui_input changed applied and effective setting');
 check(requests.length===3,'actual tool loop made three provider requests');
 for(const [i,request] of requests.entries()) {
  const text=JSON.stringify(request.messages);
  check(text.includes('Visible GUI')&&text.includes('tts_queue (auto-updated)'),`round ${i+1} received browser ephemera`);
  check(!request.tools.some(t=>['gui_snapshot','tts_queue'].includes(t.name)),`round ${i+1} hides automatic tools`);
 }
 check(JSON.stringify(requests[1]).includes('aria-selected\\\":\\\"true'),'second snapshot exposes selected state');
 // Native speech is observed through the actual app's automatic tts_queue.
 await page.locator('#artifacts .artifact').first().getByRole('button',{name:'Read artifact aloud'}).click();
 child.stdin.write('Observe speech now\n');
 for(let i=0;i<100&&requests.length<4;i++)await new Promise(r=>setTimeout(r,20));
 const ttsText=JSON.stringify(requests[3]?.messages);
 check(ttsText.includes('queued_at')&&ttsText.includes('started_at')&&(ttsText.includes('speaking')||ttsText.includes('starting')),'actual native utterance appears with TTS state and timing');
 result.nativeSpeech=await page.evaluate(()=>nativeSpeech);
 // Exercise the browser server directly with real DOM/native speech as well,
 // using its same public wire dispatcher. No stubbed DOM or speech engine.
 result.browserDetails=await page.evaluate(async()=>{
  const {browserMCP}=await import('./mcp.js');
  const replies=[];
  const speech={pending:()=>[]};
  const dispatch=browserMCP(message=>replies.push(message.payload),speech);
  for(let i=0;i<12;i++){const a=document.createElement('article');a.className='artifact';a.id='flood-'+i;a.textContent='visible artifact '+i+' '+ 'z'.repeat(3000);document.querySelector('#artifacts').append(a)}
  dispatch({jsonrpc:'2.0',id:1,method:'tools/call',params:{name:'gui_snapshot',arguments:{}}});
  dispatch({jsonrpc:'2.0',id:2,method:'tools/call',params:{name:'gui_click',arguments:{selector:'#absent'}}});
  return replies;
 });
 const snapshot=result.browserDetails[0].result.content[0].text;
 check(Buffer.byteLength(snapshot)<=4096&&/[1-9]\d* characters elided; [1-9]\d* artifacts omitted/.test(snapshot),'snapshot byte cap discloses elision and omitted artifacts');
 check(snapshot.includes('"checked":false')&&snapshot.includes('"value":"20"'),'snapshot exposes native control state');
 check(result.browserDetails[1].result.isError,'missing selector gives tool error');
 check(errors.length===0,'no browser runtime errors');
 result.result='PASS';result.browser=await browser.version();
 result.trace=fs.readFileSync(scratch+'/gui.log','utf8');
} catch(error) {result.result='FAIL';result.error=error.message;result.stderr=stderr;process.exitCode=1}
finally {await browser.close(); if(child.exitCode===null){const done=new Promise(r=>child.once('exit',r));child.kill('SIGTERM');await done}fake.close();fs.writeFileSync(evidence,JSON.stringify(result,null,2));fs.rmSync(scratch,{recursive:true,force:true});console.log(JSON.stringify({result:result.result,checks:result.checks,error:result.error}))}
