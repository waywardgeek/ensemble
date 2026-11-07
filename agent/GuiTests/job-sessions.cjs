'use strict';
// Actual page with synthetic observations/window replies; endpoint tested in C#.
const assert = require('node:assert/strict');
const { chromium } = require('../../edition-2/agent/gui/node_modules/playwright');
const base = process.env.GUI_FIXTURE_URL;
assert.equal(new URL(base).hostname, '127.0.0.1');
assert(!['8088','8089','8090'].includes(new URL(base).port));
(async () => {
 const browser = await chromium.launch({ executablePath:'/Applications/Google Chrome.app/Contents/MacOS/Google Chrome',headless:true });
 try {
  for (const width of [1280,390]) {
   const page = await browser.newPage({viewport:{width,height:844}});
   const errors=[]; page.on('pageerror',e=>errors.push(e.message));
   let count=0, release=null, held=false;
   await page.route('**/jobs/window?*',async route=>{
    count++;
    if(held) await new Promise(resolve=>release=resolve);
    const data=Buffer.from('\x1b[1;31mSAFE <img>\x1b[0m\nnext');
    await route.fulfill({json:{handle:77,status:'done',start:100,end:100+data.length,rolled_out:true,encoding:'base64',data:data.toString('base64'),origin_call_id:'origin',origin_tool:'run_command'}}).catch(()=>{});
   });
   await page.goto(base);
   await page.waitForFunction(()=>settings!==null&&!selectionPending&&jobSessions!==null);
   if(width===390) await page.locator('.pane-tab[data-pane="tools"]').click();
   await page.evaluate(()=>{
    settings={...settings,tts_enabled:false};
    window.addJobObservation=(id,name)=>{
     handle({type:'tool_dispatched',part_id:id+'-call',name,input:{}});
     handle({type:'tool_finished',part_id:id,tool_part_id:id+'-call',result:'immutable '+id,job:{handle:77,status:'done',origin_call_id:'origin',origin_tool:'run_command'}});
    };
    addJobObservation('first','run_command');
   });
   await page.waitForFunction(()=>document.querySelector('.terminal-output')?.textContent.includes('SAFE'));
   assert.equal(await page.locator('.job-session').count(),1);
   assert.equal(await page.locator('.terminal-output img').count(),0);
   assert.match(await page.locator('.job-session p').textContent(),/rolled out/);
   assert.equal(await page.locator('.terminal-bold.terminal-fg-1').textContent(),'SAFE <img>');
   await page.evaluate(()=>{
    window.sessionNode=document.querySelector('.job-session');
    const marker=document.createElement('div');marker.id='after-session';toolTranscript.append(marker);
    document.querySelector('.terminal-output').focus();
    addJobObservation('wait','wait_for_job');
   });
   assert(await page.evaluate(()=>toolTranscript.lastElementChild===sessionNode&&document.querySelector('.job-session')===sessionNode));
   assert.equal(await page.locator('[data-part-id="first"] details pre').textContent(),'immutable first');
   assert.equal(await page.locator('[data-part-id="wait"] details pre').textContent(),'immutable wait');
   assert.equal(await page.locator('.job-session').count(),1);
   assert(await page.evaluate(()=>document.activeElement===document.querySelector('.terminal-output')),'focus survives foreground move');
   await page.evaluate(()=>{
    window.firstSpan=document.querySelector('.terminal-output span');
    const range=document.createRange();range.selectNodeContents(firstSpan);
    const selection=window.getSelection();selection.removeAllRanges();selection.addRange(range);
   });
   await page.evaluate(()=>document.querySelector('.job-session button').click());
   await page.waitForTimeout(100);
   assert(await page.evaluate(()=>document.querySelector('.terminal-output span')===firstSpan&&window.getSelection().toString()==='SAFE <img>'));
   await page.evaluate(()=>window.getSelection().removeAllRanges());
   await page.evaluate(()=>{const marker=document.createElement('div');marker.id='background-marker';toolTranscript.append(marker);});
   await page.locator('.job-session button').click();
   await page.waitForTimeout(100);
   assert(await page.evaluate(()=>toolTranscript.lastElementChild.id==='background-marker'));
   held=true;
   await page.locator('.job-session button').click();
   await page.waitForFunction(()=>true);
   for(let i=0;i<100&&!release;i++) await new Promise(r=>setTimeout(r,10));
   assert(release,'held request admitted');
   await page.evaluate(()=>jobSessions.reset());
   release(); held=false;
   await page.waitForTimeout(100);
   assert.equal(await page.locator('.job-session').count(),0);
   assert(count>=3);
   assert.deepEqual(errors,[]);
   await page.close();
   console.log(`PASS continuing command ${width}: stable node, foreground movement, background stability, raw history, safe ANSI and stale response reset`);
  }
 } finally { await browser.close(); }
})().catch(e=>{console.error(e);process.exitCode=1;});
