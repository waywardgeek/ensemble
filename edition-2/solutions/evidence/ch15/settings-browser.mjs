// Actual browser and shipped GUI; settings need no provider calls or credentials.
import { chromium } from '/Users/bill/projects/ensemble/solutions/edition-2/main/gui/node_modules/playwright/index.mjs';
import { spawn } from 'node:child_process';
import { mkdtemp, mkdir, writeFile, readFile, rm } from 'node:fs/promises';
import { tmpdir } from 'node:os';
import path from 'node:path';
const base='/Users/bill/projects/ensemble/solutions/edition-2';
const scratch=await mkdtemp(path.join(tmpdir(),'ensemble-ch15-settings-'));
await mkdir(path.join(scratch,'cr'));
await writeFile(path.join(scratch,'settings.json'),JSON.stringify({context_target:123456}));
await writeFile(path.join(scratch,'cr/settings.json'),JSON.stringify({context_target:45000,log_retention:0,tts_enabled:false}));
const child=spawn('/tmp/ensemble-ch15-gui',['--port','19417','--gui-dir',(process.env.CH15_GUI_ASSETS||base+'/main/gui/web/gui')],{cwd:scratch,env:{PATH:process.env.PATH,HOME:process.env.HOME,LLM_VENDOR:'anthropic',LLM_MODEL:'claude-sonnet-5-course',LLM_API_KEY:'fake'},stdio:['pipe','pipe','pipe']});
let stderr='';child.stderr.on('data',chunk=>stderr+=chunk);child.stdout.resume();
const result={};let browser;
try {
 for(let i=0;i<100;i++){try{const r=await fetch('http://127.0.0.1:19417');if(r.ok)break;}catch{}await new Promise(r=>setTimeout(r,30));}
 browser=await chromium.launch({executablePath:'/Applications/Google Chrome.app/Contents/MacOS/Google Chrome',headless:true});
 const page=await browser.newPage();await page.goto('http://127.0.0.1:19417');
 await page.waitForFunction(()=>document.querySelector('#context-target').value==='45000');
 await page.getByRole('tab',{name:'Context Management'}).click();
 if(await page.getByRole('tab',{name:'Context Management'}).getAttribute('aria-selected')!=='true')throw Error('tab not selected');
 await page.locator('#context-target').fill('1');await page.locator('#log-retention').click();
 await page.waitForFunction(()=>document.querySelector('#context-target').value==='20000');
 await page.locator('#log-retention').fill('3');await page.locator('#context-target').click();
 await page.waitForTimeout(100);
 result.saved=JSON.parse(await readFile(path.join(scratch,'cr/settings.json'),'utf8'));
 if(result.saved.context_target!==20000||result.saved.log_retention!==3)throw Error('browser patch not persisted');
 await page.reload();await page.waitForFunction(()=>document.querySelector('#context-target').value==='20000'&&document.querySelector('#log-retention').value==='3');
 await page.getByRole('tab',{name:'Context Management'}).click();
 // Another real tab supplies the acknowledgment while the first retains an
 // unsent edit. This fixes the interleaving instead of relying on network timing.
 const second=await browser.newPage();await second.goto('http://127.0.0.1:19417');
 await second.waitForFunction(()=>document.querySelector('#context-target').value==='20000');
 await second.getByRole('tab',{name:'Context Management'}).click();
 await page.locator('#log-retention').fill('0');
 await second.locator('#context-target').fill('0');await second.locator('#log-retention').click();
 await page.waitForFunction(()=>document.querySelector('#context-target').value==='0');
 if(await page.locator('#log-retention').inputValue()!=='0')throw Error('broadcast erased unsent context edit');
 await page.locator('#context-target').click();
 await second.waitForFunction(()=>document.querySelector('#log-retention').value==='0');
 result.zero=JSON.parse(await readFile(path.join(scratch,'cr/settings.json'),'utf8'));
 if(result.zero.context_target!==0||result.zero.log_retention!==0)throw Error('explicit zero lost');
 result.cwd_fixture=JSON.parse(await readFile(path.join(scratch,'settings.json'),'utf8'));
 result.pass=true;
}catch(error){result.error=String(error);process.exitCode=1;}finally{
 if(browser)await browser.close();child.stdin.end();await new Promise(resolve=>child.once('exit',resolve));
 result.stderr=stderr;await writeFile(base+'/evidence/ch15/'+(process.env.CH15_RESULT||'settings-browser-corrected.json'),JSON.stringify(result,null,2));await rm(scratch,{recursive:true,force:true});
}
console.log(JSON.stringify(result));
