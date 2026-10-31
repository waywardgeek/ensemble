// One bounded human-GUI exercise of the first-green binary from a fresh cwd.
// Only the Ensemble child receives the approved profile location. No credentials
// are read by this launcher, passed to Chrome, or placed in command arguments.
import fs from 'node:fs';
import os from 'node:os';
import path from 'node:path';
import http from 'node:http';
import { once } from 'node:events';
import { spawn, execFileSync } from 'node:child_process';
import { createHash } from 'node:crypto';
import { pathToFileURL } from 'node:url';

const main = path.resolve(process.argv[2]);
const evidence = path.dirname(new URL(import.meta.url).pathname);
const { chromium } = await import(pathToFileURL(path.join(main,'gui/node_modules/playwright/index.mjs')));
const scratch = fs.mkdtempSync(path.join(os.tmpdir(),'ensemble-ch24-live-'));
const binary = path.join(scratch,'ensemble-gui');
execFileSync('go',['build','-o',binary,'./cmd'],{cwd:path.join(main,'gui')});
const baseEnv = Object.fromEntries(['PATH','HOME','TMPDIR','LANG'].filter(k=>process.env[k]).map(k=>[k,process.env[k]]));
const env = {...baseEnv, EN_HOME:'/Users/bill/ensemble-ch19-signin-01v0rmp3/auth',
  OPENAI_BILLING_MODE:'chatgpt', OPENAI_OIDC_ISSUER:'https://auth.openai.com',
  LLM_VENDOR:'openai', LLM_MODEL:'gpt-6-astra', OPENAI_REASONING_EFFORT:'high',
  LLM_BASE_URL:'https://api.openai.com', EN_RECALL_DISABLED:'1'};
const probe = http.createServer(); probe.listen(0,'127.0.0.1'); await once(probe,'listening');
const port = probe.address().port; await new Promise(r=>probe.close(r));
const result = { source:'35df549210130f824fdc2a4eb632ea878b2c529b',
  main_tree:'5b36f28cbc8bac320febe1a9afa810efd27452e3', scratch,
  binary:{path:binary,sha256:createHash('sha256').update(fs.readFileSync(binary)).digest('hex'),
    build_command:'go build -o <scratch>/ensemble-gui ./cmd',build_cwd:path.join(main,'gui')},
  model:'gpt-6-astra',reasoning:'high',billing:'chatgpt',assets:[],frames:[],page_errors:[] };
const prompt = 'Use the file tools to create deployment-demo.txt containing exactly "Embedded GUI works from a fresh directory." Then read it back and report the observed text in one short sentence. Do not run shell commands or make additional changes.';
result.prompt = prompt;
const start = Date.now();
const child = spawn(binary,['--port',String(port)],{cwd:scratch,env,stdio:['pipe','pipe','pipe']});
const exited = once(child,'exit');
let stdout='',stderr=''; child.stdout.on('data',b=>stdout+=b);child.stderr.on('data',b=>stderr+=b);
const watchdog = setTimeout(()=>{result.timeout=true;child.kill('SIGTERM')},120000);
let browser;
try {
  await Promise.race([new Promise(resolve=>child.stderr.on('data',b=>{if(b.toString().includes('Open http'))resolve()})),exited.then(()=>{throw Error('GUI exited before startup')})]);
  browser = await chromium.launch({headless:true,env:baseEnv,executablePath:'/Applications/Google Chrome.app/Contents/MacOS/Google Chrome'});
  const page = await browser.newPage();
  page.on('pageerror',error=>result.page_errors.push(error.message));
  page.on('response',response=>result.assets.push({path:new URL(response.url()).pathname,status:response.status()}));
  page.on('websocket',socket=>socket.on('framereceived',frame=>{
    try {result.frames.push(JSON.parse(frame.payload))} catch {}
  }));
  await page.goto(`http://127.0.0.1:${port}/?speech=transcript`);
  await page.waitForFunction(()=>document.querySelector('#state').textContent!=='connecting');
  await page.locator('#input').fill(prompt);
  await page.locator('#input').press('Enter');
  await page.waitForFunction(()=>window.ensembleTurn>=1,null,{timeout:110000});
  result.dom = await page.locator('body').innerText();
  result.file = fs.readFileSync(path.join(scratch,'deployment-demo.txt'),'utf8');
  result.task_success = result.file === 'Embedded GUI works from a fresh directory.' && result.dom.includes(result.file);
  result.source_assets_in_cwd = fs.existsSync(path.join(scratch,'web'));
} catch(error) {result.error=String(error)} finally {
  if(browser) await browser.close();
  child.stdin.end();
  const [code,signal] = await exited;
  clearTimeout(watchdog);
  result.exit=code;result.signal=signal;result.seconds=(Date.now()-start)/1000;
  fs.writeFileSync(path.join(evidence,'live-stdout.txt'),stdout);
  fs.writeFileSync(path.join(evidence,'live-stderr.txt'),stderr);
  for(const name of ['api.log','gui.log','ch02.jsonl','save.json']) {
    const file=path.join(scratch,name);
    if(fs.existsSync(file)) fs.copyFileSync(file,path.join(evidence,'live-'+name));
  }
  fs.writeFileSync(path.join(evidence,'live-results.json'),JSON.stringify(result,null,2)+'\n');
  console.log(JSON.stringify({task_success:result.task_success,exit:result.exit,seconds:result.seconds,error:result.error}));
}
