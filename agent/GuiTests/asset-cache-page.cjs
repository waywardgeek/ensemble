'use strict';
const assert = require('node:assert/strict');
const { chromium } = require('../../edition-2/agent/gui/node_modules/playwright');
const base=process.env.GUI_FIXTURE_URL || 'http://127.0.0.1:18202/assistant/';
assert.equal(new URL(base).hostname,'127.0.0.1');
assert(!['8088','8089'].includes(new URL(base).port));
(async()=>{
 const browser=await chromium.launch({executablePath:'/Applications/Google Chrome.app/Contents/MacOS/Google Chrome',headless:true});
 try {
  const page=await browser.newPage();const assets=[],errors=[];
  page.on('pageerror',error=>errors.push(error.message));
  page.on('response',response=>{const url=new URL(response.url());if(/\.(js|css)$/.test(url.pathname))assets.push({path:url.pathname,status:response.status()});});
  const response=await page.goto(base);
  assert.equal(response.headers()['cache-control'],'no-store');
  await page.waitForFunction(()=>settings!==null&&!selectionPending&&pageCapture!==null&&voiceControl!==null);
  for(const name of ['app.js','markdown.js','style.css','page-capture.js','capture-consent.js','voice-control.js','voice-session.js'])
   assert(assets.some(asset=>asset.path.endsWith('/'+name)&&asset.status===200),name+' loaded successfully');
  const versions=new Set();
  for(const asset of assets){const match=asset.path.match(/\/_assets\/([a-f0-9]{64})\//);assert(match,'Unversioned request: '+asset.path);versions.add(match[1]);}
  assert.equal(versions.size,1);assert.deepEqual(errors,[]);
  await page.reload();await page.waitForFunction(()=>settings!==null&&!selectionPending&&pageCapture!==null&&voiceControl!==null);
  assert.deepEqual(errors,[]);
  console.log('PASS no-store HTML; all static/dynamic/nested modules use one content-version path; reload reconnects');
 } finally {await browser.close();}
})().catch(error=>{console.error(error);process.exitCode=1;});
