// Local actual-native control: uses already-rendered real-provider history;
// no prompts, model requests, or synthesis replacements.
import {chromium} from '/Users/bill/projects/ensemble-edition-2-revisions/browser-tools/node_modules/playwright/index.mjs';
import {writeFileSync,readFileSync} from 'node:fs';
import {createHash} from 'node:crypto';
const root=new URL('./',import.meta.url),binding=JSON.parse(readFileSync(new URL('binding-bd5c05a.json',root)));
const browser=await chromium.launch({executablePath:binding.executables.chrome.path,headless:false});
const context=await browser.newContext(),events=[],pages=[];
try{
 for(let i=0;i<2;i++){
  const page=await context.newPage();pages.push(page);
  await page.exposeFunction('recordNative',detail=>events.push({page:i,at:Date.now(),...detail}));
  await page.addInitScript(()=>document.addEventListener('ensemble-speech',e=>window.recordNative(e.detail),true));
  await page.goto(process.argv[2]);await page.getByText(/Connected —/).waitFor();
  if(process.argv[3]==='component')await page.evaluate(async()=>{
   const {BrowserApplication}=await import('/application.js'),app=new BrowserApplication(),owner={};
   const start=document.createElement('button'),cancel=document.createElement('button');start.id='native-start';cancel.id='native-cancel';start.textContent='Start native control';cancel.textContent='Cancel native control';document.body.append(start,cancel);
   start.onclick=()=>app.speech().submit(owner,'This actual native speech belongs to its originating page. '.repeat(20),{start:()=>window.recordNative({type:'start'}),end:error=>window.recordNative({type:error?'error':'end',error})});
   cancel.onclick=()=>app.speech().cancel(owner);window.addEventListener('pagehide',()=>app.close(),{once:true});
  });
 }
 const start=p=>process.argv[3]==='component'?p.locator('#native-start'):p.locator('.artifact').filter({hasText:'Tool result:'}).first().getByRole('button',{name:'Speak full card'});
 const cancel=p=>p.locator(process.argv[3]==='component'?'#native-cancel':'[data-cancel-speech]');
 await start(pages[0]).click();
 await pages[0].waitForTimeout(1000);
 await start(pages[1]).click();
 await pages[0].waitForTimeout(7000);
 const before=events.slice();
 await cancel(pages[1]).click();
 await pages[0].waitForTimeout(1000);
 const text=await Promise.all(pages.map(p=>p.locator('body').innerText()));
 const service=readFileSync(new URL('../../gui/web/gui/speech-service.js',root));
 const report={at:new Date().toISOString(),source:'working source identified by service_sha256',service_sha256:createHash('sha256').update(service).digest('hex'),same_context:true,component:process.argv[3]==='component',before,events,text};
 const filename='native-context-'+Date.now()+'.json';writeFileSync(new URL(filename,root),JSON.stringify(report,null,2)+'\n');console.log(JSON.stringify({filename,events}));
}finally{for(const p of pages)await p.locator(process.argv[3]==='component'?'#native-cancel':'[data-cancel-speech]').click();await browser.close()}
