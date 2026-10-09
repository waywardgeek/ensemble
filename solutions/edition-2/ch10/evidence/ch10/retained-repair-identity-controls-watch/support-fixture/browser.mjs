// Deliberate one-action-at-a-time browser recorder, adapted from Chapter 9.
// This module is not run during support preparation; --check only uses Python.
import {spawnSync} from 'node:child_process';
import {readFileSync,appendFileSync,writeFileSync} from 'node:fs';
import {dirname,join,resolve} from 'node:path';
import {fileURLToPath,pathToFileURL} from 'node:url';
import readline from 'node:readline';
import {speechState,assertNoRestoredSpeech} from './speech-probe.mjs';
const here=dirname(fileURLToPath(import.meta.url));
const [bindingPath,repo,runPath,url]=process.argv.slice(2);
if (!url) throw new Error('usage: browser.mjs BINDING REPO RUN LOOPBACK_GUI_URL');
const run=resolve(runPath),binding=JSON.parse(readFileSync(bindingPath));
const launch=JSON.parse(readFileSync(join(run,'browser-launch.json')));
const pre=spawnSync(binding.binaries.python.path,[join(here,'identity.py'),bindingPath,'--repo',repo,'--launch',join(run,'browser-launch.json')],{encoding:'utf8'});
if(pre.status!==0)throw new Error(pre.stderr||pre.stdout);
if(launch.role!=='gui'||launch.gui_url!==url||new URL(url).hostname!=='127.0.0.1')throw new Error('unbound GUI origin');
const {chromium}=await import(pathToFileURL(join(binding.browser_dependencies.root,'node_modules/playwright/index.mjs')));
const browser=await chromium.launch({executablePath:binding.binaries.chrome.path,headless:false});
const context=await browser.newContext(); let sequence=0,connections=0;
const record=(kind,data)=>appendFileSync(join(run,'browser-original.jsonl'),JSON.stringify({at:new Date().toISOString(),kind,...data})+'\n');
const sourceResult=spawnSync('git',['-C',repo,'show',binding.source_revision+':solutions/edition-2/main/gui/web/gui/app.js'],{encoding:'utf8'});
if(sourceResult.status!==0)throw new Error('historical browser bootstrap missing');
const source=sourceResult.stdout;
// Expose the actual application for read-only owner probes; record this shim as
// support, while the underlying served runtime source must match its binding.
await context.route('**/app.js',route=>route.fulfill({contentType:'application/javascript',body:source+`
 globalThis.__ch10App=application; globalThis.__ch10Applied=[];
 const submit=application.service.submit.bind(application.service);
 application.__ch10ServiceAdmissions=0;
 application.service.submit=(...args)=>{application.__ch10ServiceAdmissions++;return submit(...args);};
 function probe(view){
  const observe=view.observation.bind(view),reply=view.reply.bind(view);
  view.observation=o=>{const r=observe(o);if(o.kind==='session_changed')globalThis.__ch10Applied.push({kind:'applied',session:o.session,revision:view.connector.revision});return r;};
  view.reply=m=>{const r=reply(m);if(m.type==='command_ack'&&m.status==='saved')globalThis.__ch10Applied.push({kind:'saved',id:m.id,as_of:m.as_of,watch_revision:m.watch_revision});return r;};
  return view;
 }
 for(const view of application.pages)probe(view);
 const create=application.createPage.bind(application);application.createPage=(...args)=>probe(create(...args));
 `}));
await context.addInitScript(()=>{
  globalThis.__ch10Native=0;
  const synth=globalThis.speechSynthesis;
  if(!synth)throw new Error('native speech API absent; cannot measure admission');
  const original=synth.speak.bind(synth);
  synth.speak=(u)=>{globalThis.__ch10Native++;return original(u);};
});
const pages=[];
async function state(page){return page.evaluate(`(${speechState.toString()})(globalThis.__ch10App,globalThis.__ch10Native)`);}
async function ready(page){await page.locator('[data-status]').filter({hasText:'Connected'}).waitFor({timeout:10000});}
async function open(){
 const p=await context.newPage(),index=pages.length;pages.push(p);
 p.on('websocket',ws=>{const connection=++connections;ws.on('framesent',e=>record('sent',{page:index,connection,payload:e.payload.toString()}));ws.on('framereceived',e=>record('received',{page:index,connection,payload:e.payload.toString()}));});
 await p.goto(url);await ready(p);const initial=await state(p);
 assertNoRestoredSpeech({nativeAdmissions:0},initial);record('resume_speech',{page:index,state:initial});return p;
}
try{
 await open();console.log('ready: send one JSON action, inspect its result before next action');
 for await(const line of readline.createInterface({input:process.stdin,crlfDelay:Infinity})){
  if(!line.trim())continue;const a=JSON.parse(line);if(a.action==='quit')break;
  const p=pages[a.page??0];sequence++;record('action',{sequence,action:a});
  // Failure ends this run; no implicit retry or second model submission.
  switch(a.action){
   case 'open':await open();break;
   case 'settings':await p.getByRole('tab',{name:'Settings',exact:true}).click();break;
   case 'font':await p.locator('[data-preference=font_size]').fill('18');await p.locator('[data-preference=font_size]').press('Tab');break;
   case 'policy':await p.locator('[data-policy]').fill('1');await p.locator('[data-policy-save]').click();break;
   case 'checkpoint':await p.locator('[data-checkpoint]').click();break;
   case 'prompt':{
    const schedule=JSON.parse(readFileSync(join(here,'schedule.json')));
    const text=schedule.steps.B.prompt.replaceAll('-P','-'+launch.vendor);
    // Exclusive create prevents a second submitted prompt across driver restarts.
    writeFileSync(join(launch.provider_root,'B-prompt-admitted.json'),JSON.stringify({text,at:new Date().toISOString()}),{flag:'wx'});
    await p.locator('[data-input]').fill(text);await p.locator('[data-prompt]').click();break;
   }
   case 'reconnect':{
    const before=await state(p);
    await p.evaluate(()=>[...globalThis.__ch10App.pages][0].connector.socket.close());
    await p.waitForTimeout(750);await ready(p);const after=await state(p);
    assertNoRestoredSpeech(before,after);record('reconnect_speech',{before,after});break;
   }
   case 'page-close-reopen':{
    const before=await state(p);
    await p.evaluate(()=>{const app=globalThis.__ch10App;for(const view of [...app.pages])view.close();});
    const closed=await state(p);assertNoRestoredSpeech(before,closed);record('page_closed_speech',{before,closed});
    await p.evaluate(()=>globalThis.__ch10App.createPage(document.querySelector('main')));await ready(p);
    const after=await state(p);assertNoRestoredSpeech(closed,after);record('page_reopened_speech',{after});break;
   }
   case 'reload':await p.reload();await ready(p);{const after=await state(p);assertNoRestoredSpeech({nativeAdmissions:0},after);record('reload_speech',{after});}break;
   case 'close-tab':record('before_tab_close',{state:await state(p)});await p.close();break;
   case 'wait':await p.getByText(a.text,{exact:false}).first().waitFor({timeout:30000});break;
   case 'inspect':record('owned_state',{state:await state(p)});break;
   default:throw new Error('unknown action');
  }
  if(!p.isClosed()){
   const text=await p.locator('body').innerText();writeFileSync(join(run,`browser-${sequence}.txt`),text,{flag:'wx'});
   await p.screenshot({path:join(run,`browser-${sequence}.png`),fullPage:true});
   record('dom',{sequence,text,applied:await p.evaluate(()=>JSON.parse(JSON.stringify(globalThis.__ch10Applied,(_,v)=>typeof v==='bigint'?v.toString():v)))});
  }
  console.log(JSON.stringify({sequence,ok:true}));
 }
}finally{await context.close();await browser.close();record('driver_closed',{});}
