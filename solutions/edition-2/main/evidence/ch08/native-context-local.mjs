// Local actual-native control: uses already-rendered real-provider history;
// no prompts, model requests, or synthesis replacements.
import {chromium} from '/Users/bill/projects/ensemble-edition-2-revisions/browser-tools/node_modules/playwright/index.mjs';
import {writeFileSync,readFileSync} from 'node:fs';
import {createHash} from 'node:crypto';
const root=new URL('./',import.meta.url),binding=JSON.parse(readFileSync(new URL('initial-binding.json',root)));
const browser=await chromium.launch({executablePath:binding.executables.chrome.path,headless:false});
const context=await browser.newContext(),events=[],pages=[];
try{
 for(let i=0;i<2;i++){
  const page=await context.newPage();pages.push(page);
  await page.exposeFunction('recordNative',detail=>events.push({page:i,at:Date.now(),...detail}));
  await page.addInitScript(()=>document.addEventListener('ensemble-speech',e=>window.recordNative(e.detail),true));
  await page.goto(process.argv[2]);await page.getByText(/Connected —/).waitFor();
 }
 await pages[0].locator('.artifact').filter({hasText:'Tool result:'}).first().getByRole('button',{name:'Speak full card'}).click();
 await pages[0].waitForTimeout(1000);
 await pages[1].locator('.artifact').filter({hasText:'Tool result:'}).first().getByRole('button',{name:'Speak full card'}).click();
 await pages[0].waitForTimeout(7000);
 const before=events.slice();
 await pages[1].locator('[data-cancel-speech]').click();
 await pages[0].waitForTimeout(1000);
 const text=await Promise.all(pages.map(p=>p.locator('body').innerText()));
 const service=readFileSync(new URL('../../gui/web/gui/speech-service.js',root));
 const report={at:new Date().toISOString(),source:binding.source_revision,service_sha256:createHash('sha256').update(service).digest('hex'),same_context:true,before,events,text};
 const filename='native-context-'+Date.now()+'.json';writeFileSync(new URL(filename,root),JSON.stringify(report,null,2)+'\n');console.log(JSON.stringify({filename,events}));
}finally{for(const p of pages)await p.locator('[data-cancel-speech]').click();await browser.close()}
