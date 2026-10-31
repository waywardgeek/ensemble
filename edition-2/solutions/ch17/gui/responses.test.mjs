import test from 'node:test';
import assert from 'node:assert/strict';
import { createServer } from 'node:http';
import { spawn, execFileSync } from 'node:child_process';
import { mkdtempSync, rmSync } from 'node:fs';
import { tmpdir } from 'node:os';
import { join, resolve } from 'node:path';
import { once } from 'node:events';
import { chromium } from 'playwright';

// The provider cannot finish until the actual browser speech pipeline has seen
// the summary. This protects incremental delivery, not merely a final DOM card.
test('Responses summaries reach speech before completion without duplication', async () => {
 const scratch=mkdtempSync(join(tmpdir(),'ensemble-responses-ui-'));
 const binary=join(scratch,'gui');
 execFileSync('go',['build','-o',binary,'./cmd']);
 let release;
 const gate=new Promise(resolve=>{release=resolve});
 const provider=createServer(async (req,res)=>{
  req.resume();
  res.setHeader('content-type','text/event-stream');
  const frame=(type,value)=>res.write(`event: ${type}\ndata: ${JSON.stringify(value)}\n\n`);
  frame('response.reasoning_summary_text.delta',{output_index:0,delta:'Considering carefully. '});
  await gate;
  frame('response.output_text.delta',{output_index:1,content_index:0,delta:'Final answer.'});
  frame('response.completed',{response:{status:'completed',model:'gpt-ch19-course',output:[
   {type:'reasoning',summary:[{type:'summary_text',text:'Considering carefully. '}],encrypted_content:'opaque'},
   {type:'message',content:[{type:'output_text',text:'Final answer.'}]}
  ],usage:{input_tokens:20,output_tokens:10}}});
  res.end();
 });
 provider.listen(0,'127.0.0.1');await once(provider,'listening');
 const probe=createServer();probe.listen(0,'127.0.0.1');await once(probe,'listening');const port=probe.address().port;await new Promise(r=>probe.close(r));
 const child=spawn(binary,['--port',String(port),'--gui-dir',resolve('web/gui')],{cwd:scratch,env:{...process.env,EN_HOME:join(scratch,'auth'),LLM_VENDOR:'openai',LLM_MODEL:'gpt-ch19-course',LLM_API_KEY:'fixture',LLM_BASE_URL:`http://127.0.0.1:${provider.address().port}`,OPENAI_BILLING_MODE:'api',EN_SKILLS_DIR:'',EN_PRIMARY_SKILL:''},stdio:['pipe','ignore','pipe']});
 const exited=once(child,'exit');
 const ready=new Promise((resolve,reject)=>{child.stderr.on('data',b=>{if(b.toString().includes('Open http'))resolve()});child.on('exit',()=>reject(Error('GUI exited before ready')))});
 const browser=await chromium.launch({headless:true,executablePath:process.env.CHROME_BIN||'/Applications/Google Chrome.app/Contents/MacOS/Google Chrome'});
 try {
  await ready;
  const page=await browser.newPage();
  await page.goto(`http://127.0.0.1:${port}/?speech=transcript`);
  await page.waitForFunction(()=>document.querySelector('#state').textContent!=='connecting');
  await page.locator('#input').fill('Explain briefly');await page.locator('#input').press('Enter');
  await page.waitForFunction(()=>window.ensembleSpeech.transcript.some(x=>x.source==='reasoning_summary'),null,{timeout:5000});
  const early=await page.evaluate(()=>window.ensembleSpeech.transcript);
  assert.equal(early.some(x=>x.text.includes('Final answer')),false);
  release();
  await page.waitForFunction(()=>window.ensembleTurn>=1);
  const transcript=await page.evaluate(()=>window.ensembleSpeech.transcript);
  assert.equal(transcript.filter(x=>x.source==='reasoning_summary').map(x=>x.text).join(' '),'Considering carefully.');
  assert.equal(transcript.filter(x=>x.source==='text').map(x=>x.text).join(' '),'Final answer.');
 } finally {release();await browser.close();child.kill('SIGINT');await exited;await new Promise(r=>provider.close(r));rmSync(scratch,{recursive:true,force:true});}
});
