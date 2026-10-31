import {chromium} from 'playwright';
import fs from 'node:fs';
import http from 'node:http';
import {spawn} from 'node:child_process';
const evidence=process.env.CH08_EVIDENCE_DIR||'/Users/bill/projects/ensemble/solutions/edition-2/evidence/ch08/';
const web=process.env.CH08_WEB||'/Users/bill/projects/ensemble/solutions/edition-2/main/gui/web/gui';
const scratch=fs.mkdtempSync('/tmp/ensemble-ch08-browser/run-');
let release;
const gate=new Promise(r=>release=r);
const fake=http.createServer(async(req,res)=>{
 let body='';for await(const piece of req)body+=piece;
 const request=JSON.parse(body);
 res.writeHead(200,{'Content-Type':'text/event-stream'});
 const frame=d=>res.write('data: '+JSON.stringify({model:'gpt-4o-mini',choices:[{delta:d}]})+'\n\n');
 const prompt=JSON.stringify(request.messages.at(-1));
 if(prompt.includes('STREAM')){frame({content:'Visible partial '});await gate;frame({content:'is complete.\n**Bold words** and `code`.\n<img src=x onerror="globalThis.injected=true">'});}
 else if(prompt.includes('LONG'))frame({content:'Large '.repeat(950)+'END_OF_FULL_ARTIFACT'});
 else frame({content:'A calm voice reads this sentence.'});
 res.end('data: '+JSON.stringify({choices:[{delta:{},finish_reason:'stop'}],usage:{prompt_tokens:2,completion_tokens:2}})+'\n\ndata: [DONE]\n\n');
});
await new Promise(r=>fake.listen(0,'127.0.0.1',r));
const port=19388;
const child=spawn('/tmp/ensemble-ch08-browser/ensemble-gui',['--port',String(port),'--gui-dir',web],{cwd:scratch,env:{...process.env,LLM_VENDOR:'openai',LLM_MODEL:'gpt-4o-mini',LLM_API_KEY:'fake',LLM_BASE_URL:'http://127.0.0.1:'+fake.address().port},stdio:['pipe','pipe','pipe']});
let stderr='';child.stderr.on('data',d=>stderr+=d);child.stdout.resume();
const browser=await chromium.launch({executablePath:'/Applications/Google Chrome.app/Contents/MacOS/Google Chrome',headless:false,args:['--autoplay-policy=no-user-gesture-required']});
try{
 for(let i=0;i<100;i++){try{if((await fetch('http://localhost:'+port)).ok)break;}catch{}await new Promise(r=>setTimeout(r,50));}
 const context=await browser.newContext();
 await context.addInitScript(()=>{
  globalThis.speechEvidence=[];
  const speak=speechSynthesis.speak.bind(speechSynthesis);
  speechSynthesis.speak=u=>{globalThis.speechEvidence.push({type:'queued',text:u.text,volume:u.volume});for(const kind of ['start','end','error'])u.addEventListener(kind,e=>globalThis.speechEvidence.push({type:kind,error:e.error,text:u.text}));return speak(u)};
 });
 const first=await context.newPage();await first.goto('http://localhost:'+port);await first.locator('#voice').uncheck();
 await first.locator('#input').fill('STREAM');await first.locator('#input').press('Enter');
 await first.getByText('Visible partial ',{exact:true}).waitFor();
 const second=await context.newPage();await second.goto('http://localhost:'+port);await second.locator('#voice').uncheck();
 await second.getByText('Visible partial ',{exact:true}).waitFor();
 release();await first.getByText('Bold words',{exact:true}).waitFor();await second.getByText('Bold words',{exact:true}).waitFor();
 if(await first.evaluate(()=>!!globalThis.injected)||await first.locator('.content img').count())throw Error('untrusted markup executed');
 await first.locator('#input').fill('typing');await second.getByText('paused',{exact:true}).waitFor();
 await second.locator('#resume').click();await first.waitForFunction(()=>document.querySelector('#pause-state').textContent==='');
 await first.locator('#input').fill('');await second.reload();await second.getByText('Bold words',{exact:true}).waitFor();await second.locator('#voice').uncheck();
 await first.locator('#input').fill('LONG');await first.locator('#input').press('Enter');
 await first.locator('summary').getByText('Full artifact',{exact:true}).waitFor();
 const details=first.locator('details').last();await details.locator('summary').click();if(!(await details.innerText()).includes('END_OF_FULL_ARTIFACT'))throw Error('truncation lost tail');
 await first.bringToFront();await first.locator('#voice').check();await first.locator('#input').fill('SPEAK');await first.locator('#input').press('Enter');
 await first.waitForFunction(()=>speechEvidence.some(e=>e.type==='queued'&&e.text==='A calm voice reads this sentence.'));
 await first.waitForFunction(()=>speechEvidence.some(e=>(e.type==='end'||e.type==='error')&&e.text==='A calm voice reads this sentence.'),{timeout:20000});
 await second.bringToFront();await first.bringToFront();
 const speech=await first.evaluate(()=>({voices:speechSynthesis.getVoices().length,events:speechEvidence}));
 await first.locator('.artifact').last().getByRole('button',{name:'Read artifact aloud'}).click();
 await first.keyboard.press('Escape');
 await first.waitForFunction(()=>document.querySelector('#pause-state').textContent==='');
 await first.screenshot({path:evidence+'browser.png',fullPage:false});
 fs.writeFileSync(evidence+'browser.json',JSON.stringify({result:'PASS',checks:['two tabs stream same Agent','in-flight reconnect','completed replay','safe Markdown','full long artifact reachable','typing pauses with voice off','other-tab explicit unpause','native auto-speech callback','speaker and Escape cancellation','tab visibility wake'],speech,stderr},null,2));
 console.log(JSON.stringify({result:'PASS',speech}));
}finally{release();await browser.close();child.kill('SIGTERM');await new Promise(r=>child.on('exit',r));fake.close();fs.copyFileSync(scratch+'/gui.log',evidence+'browser-gui.log');fs.rmSync(scratch,{recursive:true,force:true});}
