import test from 'node:test';
import assert from 'node:assert/strict';
import { createServer } from 'node:http';
import { spawn, execFileSync } from 'node:child_process';
import { mkdtempSync, rmSync, writeFileSync, mkdirSync } from 'node:fs';
import { tmpdir } from 'node:os';
import { join, resolve } from 'node:path';
import { once } from 'node:events';
import { chromium } from 'playwright';

// The browser and application stay real. Only the paid HTTP boundary is fake;
// known/unknown model runs protect visible cost, current counts and tooltip totals.
test('session meter renders measured usage and unknown price', async () => {
 const scratch=mkdtempSync(join(tmpdir(),'ensemble-cache-ui-'));
 const binary=join(scratch,'gui');
 execFileSync('go',['build','-o',binary,'./cmd']);
 const provider=createServer((req,res)=>{
  req.resume();
  res.setHeader('content-type','application/json');
  res.end(JSON.stringify({content:[{type:'text',text:'meter ready'}],usage:{input_tokens:20,cache_read_input_tokens:80,cache_creation_input_tokens:0,output_tokens:2}}));
 });
 provider.listen(0,'127.0.0.1');await once(provider,'listening');
 const browser=await chromium.launch({headless:true,executablePath:process.env.CHROME_BIN||'/Applications/Google Chrome.app/Contents/MacOS/Google Chrome'});
 try {
  for(const model of ['claude-sonnet-5','unknown-model']) {
   const cwd=join(scratch,model);mkdirSync(join(cwd,'cr'),{recursive:true});writeFileSync(join(cwd,'cr/settings.json'),JSON.stringify({tts_enabled:false}));
   const probe=createServer();probe.listen(0,'127.0.0.1');await once(probe,'listening');const port=probe.address().port;await new Promise(r=>probe.close(r));
   const child=spawn(binary,['--port',String(port),'--gui-dir',resolve('web/gui'),'--no-stream'],{cwd,env:{...process.env,LLM_VENDOR:'anthropic',LLM_MODEL:model,LLM_API_KEY:'fixture',LLM_BASE_URL:`http://127.0.0.1:${provider.address().port}`,EN_SKILLS_DIR:'',EN_PRIMARY_SKILL:''},stdio:['pipe','ignore','pipe']});
   const exited=once(child,'exit');
   const page=await browser.newPage();
   try {
    await new Promise((resolve,reject)=>{child.stderr.on('data',b=>{if(b.toString().includes('Open http'))resolve()});child.on('exit',()=>reject(Error('GUI exited before ready')))});
    await page.goto(`http://127.0.0.1:${port}/?speech=transcript`);
    await page.waitForFunction(()=>document.querySelector('#state').textContent!=='connecting');
    for(let turn=1;turn<=2;turn++) {
     await page.locator('#input').fill('show meter');await page.locator('#input').press('Enter');
     await page.waitForFunction(n=>window.ensembleTurn>=n,turn);
     const text=await page.locator('#usage').innerText();
     assert.equal(text,`Cost ${model==='unknown-model'?'—':`$${(.000076*turn).toFixed(4)}`} · Input 100 · Cache 80 · Hit 80.0%`);
     assert.equal(await page.locator('#usage').getAttribute('title'),`Session tokens: input ${20*turn}, cache write 0, cache read ${80*turn}, output ${2*turn}`);
    }
   } finally {await page.close();child.kill('SIGINT');await exited;}
  }
 } finally {await browser.close();await new Promise(r=>provider.close(r));rmSync(scratch,{recursive:true,force:true});}
});
