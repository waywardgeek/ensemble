'use strict';
// Real embedded page; Markdown frames injected into its production handler.
// Existing speech-after-turn covers the separate real provider/WebSocket path.
const assert = require('node:assert/strict');
const { chromium } = require('../../edition-2/agent/gui/node_modules/playwright');
const base = process.env.GUI_FIXTURE_URL || 'http://127.0.0.1:18202/assistant/';
assert.equal(new URL(base).hostname, '127.0.0.1');
assert(!['8088','8089'].includes(new URL(base).port));
(async () => {
 const browser = await chromium.launch({executablePath:'/Applications/Google Chrome.app/Contents/MacOS/Google Chrome',headless:true});
 try {
  const page = await browser.newPage({viewport:{width:390,height:844}});
  const errors = [], remote = [];
  page.on('pageerror', error => errors.push(error.message));
  page.on('request', request => { if (!request.url().startsWith(new URL(base).origin + '/')) remote.push(request.url()); });
  await page.goto(base);
  await page.waitForFunction(() => settings !== null && !selectionPending);
  for (const kind of ['thinking','reasoning_summary']) {
   await page.evaluate(kind=>handle({type:'part_delta',part_id:'empty-'+kind,kind,chunk:' \n'}),kind);
   assert.equal(await page.locator(`[data-part-id="empty-${kind}"]`).isVisible(),false);
   await page.evaluate(kind=>handle({type:'part_delta',part_id:'empty-'+kind,kind,chunk:'**Visible thought**'}),kind);
   await page.waitForFunction(kind=>!cards.get('empty-'+kind).node.hidden,kind);
   assert.equal(await page.locator(`[data-part-id="empty-${kind}"] strong`).last().textContent(),'Visible thought');
   await page.evaluate(kind=>handle({type:'part_final',part_id:'empty-'+kind,part:{type:kind,text:''}}),kind);
   assert.equal(await page.locator(`[data-part-id="empty-${kind}"]`).isVisible(),false);
   assert.equal(await page.evaluate(kind=>cards.get('empty-'+kind).text,kind),'');
  }
  console.log('PASS empty thinking/summary hidden, streamed content revealed, empty final hidden');
  const sample='## Heading\n\n**bold with *nested emphasis***\n\n| Name | Value |\n|---|---:|\n| One | 1 |\n\n1. First\n   - Nested\n\n> Quote\n\n```js\nconst value = "<script>";\n```\n\n[Safe](https://example.com)';
  await page.evaluate(sample => {
   window.speechBefore = _ttsLog.length;
   settings={...settings,tts_enabled:false};
   for (const chunk of sample) handle({type:'part_delta',part_id:'markdown-test',kind:'text',chunk});
   window.savedButton=cards.get('markdown-test').node.querySelector('[data-action="speak"]');
   handle({type:'part_final',part_id:'markdown-test',part:{type:'text',text:sample}});
  },sample);
  const card=page.locator('[data-part-id="markdown-test"]');
  assert.equal(await card.locator('table').count(),1);assert.equal(await card.locator('strong em').count(),1);assert.equal(await card.locator('ol ul').count(),1);
  assert.equal(await card.locator('th[scope="col"]').count(),2);
  assert.equal(await page.evaluate(()=>cards.get('markdown-test').text),sample);
  assert(await page.evaluate(()=>savedButton===cards.get('markdown-test').node.querySelector('[data-action="speak"]')));
  assert.equal(await page.evaluate(()=>_ttsLog.length),await page.evaluate(()=>speechBefore));
  console.log('PASS streamed Markdown final, semantic table/list/inline, original text and stable playback');
  await page.evaluate(()=>{
   savedButton.focus();render(cards.get('markdown-test'));
   window.focusKept=document.activeElement===savedButton;
   window.parseCount=0;window.originalMarkdown=window.EnsembleArtifactRenderers.prose;
   window.EnsembleArtifactRenderers.prose=(...args)=>{parseCount++;return originalMarkdown(...args);};
   for(let i=0;i<100;i++)handle({type:'part_delta',part_id:'batch-test',kind:'text',chunk:'x'});
   window.immediateParseCount=parseCount;
  });
  assert(await page.evaluate(()=>focusKept));
  assert.equal(await page.evaluate(()=>immediateParseCount),1);
  await page.waitForFunction(()=>pendingCardRenders.size===0);
  assert.equal(await page.evaluate(()=>parseCount),2);
  await page.evaluate(()=>{window.EnsembleArtifactRenderers.prose=originalMarkdown;});
  console.log('PASS playback keyboard focus and 100 deltas coalesced into two paints');
  const hostile='<script>window.pwned=1</script>\n\n<img src=x onerror="window.pwned=1">\n\n![tracker](https://example.com/tracker.png)\n\n[bad](javascript:alert(1)) [data](data:text/html,evil) [entity](jav&#x61;script:alert(1))';
  await page.evaluate(text=>handle({type:'part_final',part_id:'hostile-test',part:{type:'text',text}}),hostile);
  assert.equal(await page.locator('[data-part-id="hostile-test"] img,[data-part-id="hostile-test"] script,[data-part-id="hostile-test"] a').count(),0);
  assert.equal(await page.evaluate(()=>window.pwned),undefined);assert.deepEqual(remote,[]);
  console.log('PASS unsafe HTML/links inert; no remote-image requests');
  await page.evaluate(()=>{
   const c=cards.get('markdown-test');const r=document.createRange();r.selectNodeContents(c.node.querySelector('h2'));
   getSelection().removeAllRanges();getSelection().addRange(r);
   c.text='Deferred final text';render(c);
  });
  assert.equal(await card.locator('h2').innerText(),'Heading');
  await page.evaluate(()=>getSelection().removeAllRanges());
  await page.waitForFunction(()=>cards.get('markdown-test').node.textContent.includes('Deferred final text'));
  console.log('PASS selection survives update and deferred final paints after release');
  const long='```js\n'+('const x = 42;\n').repeat(40)+'```';
  await page.evaluate(text=>handle({type:'part_final',part_id:'code-test',part:{type:'text',text}}),long);
  await page.locator('[data-part-id="code-test"] details summary').click();
  await page.evaluate(()=>render(cards.get('code-test')));
  assert.equal(await page.locator('[data-part-id="code-test"] details').getAttribute('open'),'');
  assert.equal(await page.locator('[data-part-id="code-test"] details code').textContent(),('const x = 42;\n').repeat(40));
  console.log('PASS complete long code retained and expansion survives render');
  await page.evaluate(sample=>{
   const c=cardFor('prefix-test');
   for(let i=1;i<=sample.length;i++){c.text=sample.slice(0,i);render(c);}
   for(let i=0;i<40;i++) handle({type:'part_final',part_id:'scroll-'+i,part:{type:'text',text:'Older paragraph'}});
   transcript.scrollTop=100;window.oldScroll=transcript.scrollTop;
   c.text+='\n\nAdditional paragraph';render(c);
  },sample);
  assert(await page.evaluate(()=>Math.abs(transcript.scrollTop-oldScroll)<2));
  assert(await page.evaluate(()=>transcript.scrollWidth<=transcript.clientWidth+1));
  console.log('PASS incomplete prefixes, old scroll position and narrow-pane containment');
  await page.evaluate(()=>{
   const r=document.createRange();
   r.setStart(cards.get('scroll-0').node.querySelector('.message p').firstChild,0);
   r.setEnd(cards.get('scroll-2').node.querySelector('.message p').firstChild,3);
   getSelection().removeAllRanges();getSelection().addRange(r);
   const middle=cards.get('scroll-1');middle.text='Updated middle card';render(middle);
  });
  assert.equal(await page.locator('[data-part-id="scroll-1"] .message').innerText(),'Older paragraph');
  await page.evaluate(()=>getSelection().removeAllRanges());
  await page.waitForFunction(()=>cards.get('scroll-1').node.textContent.includes('Updated middle card'));
  console.log('PASS cross-message selection protects intermediate card');
  await page.evaluate(()=>{
   const renderer=window.EnsembleArtifactRenderers.prose;
   window.EnsembleArtifactRenderers.prose=()=>{throw new Error('deliberate refusal');};
   handle({type:'part_final',part_id:'fallback-prose',part:{type:'text',text:'<b>literal recorded text</b>'}});
   window.EnsembleArtifactRenderers.prose=renderer;
  });
  assert.equal(await page.locator('[data-part-id="fallback-prose"] .message').textContent(),'<b>literal recorded text</b>');
  assert.equal(await page.locator('[data-part-id="fallback-prose"] .message b').count(),0);
  assert.equal(await page.evaluate(()=>cards.get('fallback-prose').text),'<b>literal recorded text</b>');
  assert.equal(await page.locator('[data-part-id="fallback-prose"] [data-action="speak"]').count(),1);
  console.log('PASS prose renderer refusal preserves literal text and GUI controls');
  await page.evaluate(()=>{
   handle({type:'part_delta',part_id:'pending-test',kind:'text',chunk:'start'});
   handle({type:'part_delta',part_id:'pending-test',kind:'text',chunk:'tail'});
   handle({type:'replay',agent_id:agentID,data:{events:[],partials:[],agents:[]}});
  });
  await page.waitForTimeout(100);
  assert.equal(await page.locator('[data-part-id="pending-test"]').count(),0);
  assert.deepEqual(errors,[]);console.log('PASS replay cancels pending render; no browser exceptions');
 } finally {await browser.close();}
})().catch(error=>{console.error(error);process.exitCode=1;});
