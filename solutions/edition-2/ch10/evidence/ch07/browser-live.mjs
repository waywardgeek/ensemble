// Interactive actual-browser driver. Each line performs one deliberate human UI
// action; model work is never retried or automatically resubmitted.
import {chromium} from '/Users/bill/projects/ensemble-edition-2-revisions/browser-tools/node_modules/playwright/index.mjs';
import {spawnSync,spawn} from 'node:child_process';
import {readFileSync,appendFileSync,writeFileSync} from 'node:fs';
import {dirname,join,resolve} from 'node:path';
import {fileURLToPath} from 'node:url';
import readline from 'node:readline';
const here=dirname(fileURLToPath(import.meta.url)),run=resolve(here,process.argv[2]),url=process.argv[3];
const check=spawnSync('python3',['-c',`import json,sys; from pathlib import Path; from evidence import HERE,preflight; b=json.loads((HERE/'initial-binding.json').read_text()); preflight(b,{}); l=json.loads((Path(sys.argv[1])/'launch.json').read_text()); assert l['source_revision']==b['source_revision'],'launch source mismatch'; assert l['executables']=={n:v['sha256'] for n,v in b['executables'].items()},'launch executable mismatch'; assert l['support']==b['support'],'launch support mismatch'; assert l['browser_tools']==b['browser_tools'],'launch browser tools mismatch'`,run],{cwd:here,encoding:'utf8'});
if(check.status!==0)throw new Error(check.stderr||check.stdout);
const binding=JSON.parse(readFileSync(join(here,'initial-binding.json')));
const browserServer=await chromium.launchServer({channel:'chrome',headless:false});
const browser=await chromium.connect(browserServer.wsEndpoint());
const browserPID=browserServer.process().pid,pages=[];let sequence=0;
const record=(kind,value)=>appendFileSync(join(run,'browser-original.jsonl'),JSON.stringify({at:new Date().toISOString(),kind,...value})+'\n');
record('launch',{source_revision:binding.source_revision,browser:browser.version(),browser_pid:browserPID,executables:Object.fromEntries(Object.entries(binding.executables).map(([k,v])=>[k,v.sha256])),url});
async function open(){const page=await browser.newPage();const index=pages.length;pages.push(page);
 page.on('pageerror',e=>record('pageerror',{page:index,error:e.message}));
 page.on('websocket',socket=>{socket.on('framesent',e=>record('browser_sent',{page:index,payload:e.payload.toString()}));socket.on('framereceived',e=>record('browser_received',{page:index,payload:e.payload.toString()}));});
 await page.exposeFunction('recordSpeech',detail=>record('speech',{page:index,detail}));
 await page.addInitScript(()=>document.addEventListener('ensemble-speech',e=>window.recordSpeech({...e.detail,unix_ms:Date.now()}),true));
 await page.goto(url);await page.locator('[data-status]').first().filter({hasText:'Connected'}).waitFor({timeout:10000});return page;
}
await open();console.log(JSON.stringify({ready:true,browser:browser.version(),browserPID,url}));
const lines=readline.createInterface({input:process.stdin,crlfDelay:Infinity});
try{for await(const line of lines){if(!line.trim())continue;const action=JSON.parse(line);if(action.action==='quit')break;
 sequence++;const page=pages[action.page||0];const root=action.panel===undefined?page.locator('main'):page.locator('main').nth(action.panel);record('action',{sequence,action});
 try{
  switch(action.action){
   case 'open':await open();break;
   case 'type':await root.locator('[data-input]').fill(action.text);break;
   case 'prompt':case 'hint':await root.locator('[data-input]').fill(action.text);await root.locator(action.action==='prompt'?'[data-prompt]':'[data-hint]').click();break;
   case 'interrupt':await root.locator('[data-interrupt]').click();break;
   case 'reload':await page.reload();await root.locator('[data-status]').filter({hasText:'Connected'}).waitFor();break;
   case 'close':await page.close();break;
   case 'auto-speech':await root.locator('[data-auto-speech]').click();break;
   case 'cancel-speech':await root.locator('[data-cancel-speech]').click();break;
   case 'escape':await root.locator('[data-input]').press('Escape');break;
   case 'wait':await root.getByText(action.text,{exact:!!action.exact}).first().waitFor({timeout:action.timeout||60000});break;
   case 'speak':await root.locator('.artifact').filter({hasText:action.text}).first().getByRole('button',{name:'Speak full card'}).click();break;
   case 'audio':{
    const path=join(run,`audio-${sequence}.wav`),capture=spawn(binding.executables.capture.path,[path,String(browserPID)]);let output='';const ended=new Promise(resolve=>capture.on('exit',code=>resolve(code)));
    capture.stdout.on('data',d=>output+=d);capture.stderr.on('data',d=>output+=d);
    await Promise.race([new Promise(resolve=>capture.stdout.on('data',()=>{if(output.includes('capture_ready'))resolve()})),ended]);
    if(!output.includes('capture_ready'))throw new Error('audio capture unavailable: '+output);
    await new Promise(resolve=>setTimeout(resolve,2000));
    await root.locator('.artifact').filter({hasText:action.text}).first().getByRole('button',{name:'Speak full card'}).click();
    const exit=await ended;record('audio',{sequence,path,exit,output,browserPID,silent_control_seconds:2});if(exit!==0)throw new Error('capture failed');break;
   }
   case 'inspect':break;
   default:throw new Error('unknown action');
  }
  if(!page.isClosed()){const text=await page.locator('body').innerText();const screenshot=join(run,`browser-${sequence}.png`);await page.screenshot({path:screenshot,fullPage:true});writeFileSync(join(run,`browser-${sequence}.txt`),text);record('page_receipt',{sequence,text,screenshot});console.log(JSON.stringify({sequence,ok:true,text}));}
  else console.log(JSON.stringify({sequence,ok:true,closed:true}));
 }catch(error){record('action_failed',{sequence,error:error.message});console.log(JSON.stringify({sequence,ok:false,error:error.message}));}
}}
finally{await browser.close();await browserServer.close();record('closed',{});}
