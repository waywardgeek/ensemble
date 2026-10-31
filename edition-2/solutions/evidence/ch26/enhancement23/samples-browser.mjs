// Real Chrome runs the published GUI; only the remote model/auth endpoints are fakes.
import { chromium } from '/Users/bill/projects/ensemble/solutions/edition-2/main/gui/node_modules/playwright/index.mjs';
import fs from 'node:fs';
const [url,work,out,mode]=process.argv.slice(2);
const browser=await chromium.launch({executablePath:'/Applications/Google Chrome.app/Contents/MacOS/Google Chrome',headless:true});
const context=await browser.newContext();
const frames=[],sent=[],errors=[];
async function page(){const p=await context.newPage();p.on('pageerror',e=>errors.push(String(e)));p.on('websocket',ws=>{ws.on('framereceived',e=>{try{frames.push(JSON.parse(String(e.payload)))}catch{}});ws.on('framesent',e=>{try{sent.push(JSON.parse(String(e.payload)))}catch{}})});await p.goto(url+'?speech=transcript');await p.locator('[data-agent-key="main"]').waitFor();return p;}
async function create(p,name){await p.locator('#agent-name').fill(name);await p.locator('#agent-model').fill('gpt-6-astra');await p.locator('#agent-skill').fill('ensemble');await p.locator('#create-agent button').click();await p.locator(`[data-agent-key="${name}"]`).waitFor();}
async function selected(p,key){await p.locator(`[data-agent-key="${key}"]`).click();await p.waitForFunction(key=>document.querySelector('#conversation-title').textContent.includes(key),key);}
try{
 const p=await page();
 if(mode==='first'){
  const second=await page();
  await create(p,'console');await selected(p,'console');
  await p.locator('#input').fill(fs.readFileSync(work+'/prompt.txt','utf8'));await p.locator('#input-form button').click();
  await p.locator('[data-agent-key="console/helper"]').waitFor();await selected(p,'console/helper');
  if(!(await second.locator('[data-agent-key="main"]').getAttribute('aria-pressed')==='true'))throw Error('second tab selection moved');
  fs.writeFileSync(work+'/sample-release','');await selected(p,'console');
  await p.waitForFunction(()=>document.querySelector('#state').textContent==='idle'&&document.body.textContent.includes('CH23_FLEET_COMPLETE'),null,{timeout:20000});
  await create(p,'spare');await p.getByRole('button',{name:'Delete spare',exact:true}).click();await p.locator('[data-agent-key="spare"]').waitFor({state:'detached'});
  await p.screenshot({path:out+'/samples-browser.png',fullPage:true});
 }else{
  await selected(p,'console');await p.waitForFunction(()=>document.querySelector('#state').textContent==='idle'&&document.body.textContent.includes('CH23_FLEET_COMPLETE'));
  if(await p.locator('[data-agent-key="spare"],[data-agent-key="console/helper"]').count())throw Error('deleted/ephemeral peer returned');
 }
 if(errors.length)throw Error(JSON.stringify(errors));
 fs.writeFileSync(out+`/samples-browser-${mode}.json`,JSON.stringify({result:'PASS',sent,frames,errors},null,2)+'\n');
}finally{await browser.close();}
