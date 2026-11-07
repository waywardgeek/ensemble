'use strict';
const assert = require('node:assert/strict');
const { chromium } = require('../../edition-2/agent/gui/node_modules/playwright');
const base = process.env.GUI_FIXTURE_URL;
assert.equal(new URL(base).hostname, '127.0.0.1');
(async () => {
 const browser = await chromium.launch({executablePath:'/Applications/Google Chrome.app/Contents/MacOS/Google Chrome',headless:true});
 try {
  for (const width of [1280,390]) {
   const page = await browser.newPage({viewport:{width,height:844}});
   const tail = 'x'.repeat(65520)+'\nLAST OUTPUT END';
   assert.equal(Buffer.byteLength(tail),65536);
   await page.route('**/jobs/window?*', route => route.fulfill({json:{handle:7,status:'done',start:100,end:65636,rolled_out:true,encoding:'base64',data:Buffer.from(tail).toString('base64')}}));
   await page.goto(base);
   await page.waitForFunction(() => settings !== null && !selectionPending && window.EnsembleArtifactRenderers?.list_directory);
   if(width===390) await page.locator('.pane-tab[data-pane="tools"]').click();
   await page.evaluate(() => {
    handle({type:'tool_dispatched',part_id:'compact-call',name:'list_directory',input:{path:'.'}});
    handle({type:'tool_finished',part_id:'compact-result',tool_part_id:'compact-call',result:'short model observation',job:{handle:7,status:'done'},recorded_at:'2026-11-07T22:00:22Z',observation_ms:211});
   });
   await page.waitForFunction(() => cards.get('compact-result')?.node.querySelector('.preview')?.textContent.endsWith('LAST OUTPUT END'));
   const result=page.locator('[data-part-id="compact-result"]');
   assert.equal(await result.locator('.card-heading strong').textContent(),'Job 7 completed');
   assert.equal(await result.locator('details').isVisible(),false);
   assert.equal(await result.locator('.tool-state').count(),0);
   assert((await result.locator('.preview').textContent())===tail, 'visible tail differs from retained bytes');
   assert.match(await result.locator('.output-window-notice').textContent(),/64 KiB/);
   const labels=await result.locator('.card-heading button').evaluateAll(nodes=>nodes.map(n=>n.getAttribute('aria-label')));
   assert.equal(labels[0],'Download complete output'); assert.equal(labels[1],'View call'); assert.match(labels[2],/^Speak/);
   assert.equal(await page.evaluate(()=>cards.get('compact-result').text),'short model observation');
   await result.locator('.preview').focus();
   assert(await page.evaluate(()=>{const c=cards.get('compact-result'), n=c.node.querySelector('.preview').firstChild; render(c);return document.activeElement===c.node.querySelector('.preview')&&n===c.node.querySelector('.preview').firstChild;}));
   const stamp=await result.locator('.recorded-time').textContent();
   assert(!stamp.includes('Recorded')&&!stamp.includes('2026'));
   assert.equal(await result.locator('.observation-duration').textContent(),' · 0.211 s');
   const boxes=await result.locator('.recorded-time,.observation-duration').evaluateAll(nodes=>nodes.map(n=>n.getBoundingClientRect().top));
   assert(Math.abs(boxes[0]-boxes[1])<2);
   await result.locator('[data-action="view-pair"]').click();
   assert.equal(await page.evaluate(()=>document.activeElement.dataset.partId),'compact-call');
   await page.evaluate(() => {
    handle({type:'tool_dispatched',part_id:'partial-call',name:'list_directory',input:{path:'.'}});
    handle({type:'tool_finished',part_id:'partial-result',tool_part_id:'partial-call',result:'model text',job:{handle:7,status:'done'},directory_listing:{total_entries:3,entries:[{name:'captured-name',kind:'unknown'}]}});
   });
   await page.waitForFunction(() => cards.get('partial-result')?.node.querySelector('.preview')?.textContent.endsWith('LAST OUTPUT END'));
   assert.equal(await page.locator('[data-part-id="partial-result"] .directory-preview li').count(),1);
   assert.match(await page.locator('[data-part-id="partial-result"] .preview').textContent(),/Retained output tail/);
   let held;
   await page.route('**/jobs/window?*', async route => {
    if(new URL(route.request().url()).searchParams.get('handle')==='8') held=route;
    else await route.fallback();
   });
   await page.evaluate(() => handle({type:'tool_finished',part_id:'stale-window',result:'old',job:{handle:8,status:'done'}}));
   for(let i=0;!held&&i<100;i++) await new Promise(r=>setTimeout(r,20));
   assert(held,'delayed fetch was not admitted');
   await page.evaluate(() => {
    window.oldWindowCard=cards.get('stale-window');
    oldWindowCard.node.remove(); cards.delete('stale-window');
    handle({type:'tool_finished',part_id:'stale-window',result:'replacement'});
   });
   await held.fulfill({json:{handle:8,status:'done',start:0,end:4,rolled_out:false,encoding:'base64',data:Buffer.from('late').toString('base64')}});
   await page.waitForFunction(() => oldWindowCard.windowController===null);
   assert.equal(await page.evaluate(()=>cards.get('stale-window').windowText),undefined);
   assert.equal(await page.locator('[data-part-id="stale-window"] .preview').textContent(),'replacement');
   await page.close(); console.log('PASS compact controls, faithful 64KiB tail, model text unchanged, focus/pairing/timing width '+width);
  }
 } finally { await browser.close(); }
})().catch(e=>{console.error(e);process.exitCode=1;});
