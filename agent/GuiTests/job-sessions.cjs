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
   let count=0, release=null, held=false, unavailable=false;
   let windowStart=100;
   let replyStatus='running', replyText='\x1b[1;31mSAFE <img>\x1b[0m\nnext';
   await page.route('**/jobs/window?*',async route=>{
    count++;
    if(held) await new Promise(resolve=>release=resolve);
    if (unavailable) { await route.fulfill({status:503,body:'unavailable'}).catch(()=>{}); return; }
    const data=Buffer.from(replyText);
    await route.fulfill({json:{handle:77,status:replyStatus,start:windowStart,end:windowStart+data.length,rolled_out:windowStart>0,encoding:'base64',data:data.toString('base64'),origin_call_id:'origin',origin_tool:'run_command'}}).catch(()=>{});
   });
   await page.goto(base);
   await page.waitForFunction(()=>settings!==null&&!selectionPending&&jobSessions!==null&&typeof window.EnsembleArtifactRenderers?.command === "function");
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
   await page.waitForSelector('[data-part-id="first"] .session-output-link');
   assert.equal(count,1,'Session must be the only initial output fetch owner');
   assert.equal(await page.locator('.job-session strong').textContent(),'Job 77 running');
   assert(await page.locator('.job-session button').isHidden());
   assert.equal(await page.locator('[data-part-id="first"] .preview').textContent(),'View Job 77 output');
   await page.evaluate(()=>{window.backgroundSession=document.querySelector('.job-session');const marker=document.createElement('div');marker.id='autonomous-marker';toolTranscript.append(marker);});
   replyText += '\nbackground progress';
   await page.waitForFunction(()=>document.querySelector('.terminal-output')?.textContent.includes('background progress'));
   assert(await page.evaluate(()=>toolTranscript.lastElementChild.id==='autonomous-marker'&&document.querySelector('.job-session')===backgroundSession),'Background polling must update without foreground movement');
   replyStatus='done'; replyText='\x1b[1;31mSAFE <img>\x1b[0m\nnext';

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
   await page.waitForSelector('[data-part-id="wait"] .session-output-link');
   assert.equal(await page.locator('[data-part-id="wait"] .preview').textContent(),'View Job 77 output');
   assert.equal(await page.locator('[data-part-id="wait"] details').getAttribute('open'),null);
   assert.equal(await page.locator('[data-part-id="wait"] details summary').textContent(),'Recorded observation');
   await page.locator('[data-part-id="wait"] details summary').click();
   assert(await page.locator('[data-part-id="wait"] details pre').isVisible(),'Recorded output remains inspectable');
   await page.locator('[data-part-id="wait"] details summary').click();
   await page.evaluate(()=>{document.querySelector('.terminal-output').focus();addJobObservation('input','send_input');});
   await page.waitForSelector('[data-part-id="input"] .session-output-link');
   assert.equal(await page.locator('[data-part-id="input"] .preview').textContent(),'View Job 77 output');
   assert(await page.evaluate(()=>{
    const card=cards.get('wait'), agent=card.agent, job=card.job;
    card.agent='different-agent'; render(card);
    const otherAgent=!card.node.querySelector('.session-output-link');
    card.agent=agent; card.job={...job,origin_call_id:'different-origin'}; render(card);
    const otherOrigin=!card.node.querySelector('.session-output-link');
    card.job=job; render(card);
    return otherAgent&&otherOrigin&&!!card.node.querySelector('.session-output-link');
   }),'Session association must retain Agent and origin identity');
   await page.evaluate(()=>{
    handle({type:'tool_dispatched',part_id:'error-call',name:'send_input',input:{}});
    handle({type:'tool_finished',part_id:'error-result',tool_part_id:'error-call',result:'Input refused',is_error:true,
      job:{handle:77,status:'done',origin_call_id:'origin',origin_tool:'run_command'}});
   });
   assert.equal(await page.locator('[data-part-id="error-result"] .preview').textContent(),'Input refused');
   assert.equal(await page.locator('[data-part-id="error-result"] .session-output-link').count(),0);


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
   await page.evaluate(()=>{
    window.originalCommandRenderer=window.EnsembleArtifactRenderers.command;
    window.EnsembleArtifactRenderers.command=(parent)=>{parent.textContent='CUSTOM';};
    document.dispatchEvent(new Event('ensemble-renderers-changed'));
   });
   assert(await page.evaluate(()=>document.querySelector('.terminal-output span')===firstSpan),'renderer change disrupted selection');
   await page.evaluate(()=>window.getSelection().removeAllRanges());
   await page.waitForFunction(()=>document.querySelector('.terminal-output').textContent==='CUSTOM');
   await page.evaluate(()=>{
    window.EnsembleArtifactRenderers.command=(_parent,bytes)=>{bytes.fill(88);throw new Error('renderer refusal');};
    document.dispatchEvent(new Event('ensemble-renderers-changed'));
   });
   assert.match(await page.locator('.terminal-output').textContent(),/SAFE <img>/);
   assert.equal(await page.locator('.terminal-output img').count(),0);
   await page.evaluate(()=>{
    window.EnsembleArtifactRenderers.command=window.originalCommandRenderer;
    document.dispatchEvent(new Event('ensemble-renderers-changed'));
   });
   assert.equal(await page.locator('.terminal-bold.terminal-fg-1').textContent(),'SAFE <img>');
   assert(await page.evaluate(()=>document.querySelector('.job-session')===sessionNode));

   await page.evaluate(()=>{const marker=document.createElement('div');marker.id='background-marker';toolTranscript.append(marker);});
   await page.evaluate(()=>jobSessions.observe(cards.get('first')));
   await page.waitForTimeout(100);
   assert(await page.evaluate(()=>toolTranscript.lastElementChild.id==='background-marker'));
   unavailable=true;
   await page.evaluate(()=>jobSessions.observe(cards.get('first')));
   await page.waitForFunction(()=>document.querySelector('[data-part-id="wait"] .preview')?.textContent==='immutable wait');
   assert.equal(await page.locator('[data-part-id="wait"] .session-output-link').count(),0,'Failed session fetch restores recorded observation');
   assert(await page.locator('.job-session button').isVisible());
   unavailable=false;
   await page.locator('.job-session button').click();
   await page.waitForSelector('[data-part-id="wait"] .session-output-link');
   assert.equal(await page.locator('.job-session strong').textContent(),'Job 77 completed');
   windowStart=0;
   await page.evaluate(()=>jobSessions.observe(cards.get('first')));
   await page.waitForFunction(()=>document.querySelector('.job-session p').hidden);
   assert(await page.locator('.job-session button').isHidden());
   assert.match(await page.locator('.terminal-output').getAttribute('title'), /Terminal log/);
   assert.match(await page.locator('.terminal-output').getAttribute('aria-label'), /Output bytes 0/);

   held=true;
   await page.evaluate(()=>jobSessions.observe(cards.get('first')));
   await page.waitForFunction(()=>true);
   for(let i=0;i<100&&!release;i++) await new Promise(r=>setTimeout(r,10));
   assert(release,'held request admitted');
   await page.evaluate(()=>jobSessions.reset());
   release(); held=false;
   await page.waitForTimeout(100);
   assert.equal(await page.locator('.job-session').count(),0);
   await page.waitForFunction(()=>document.querySelector('[data-part-id="wait"] .preview')?.textContent==='immutable wait');

   assert(count>=3);
   assert.deepEqual(errors,[]);
   await page.close();
   console.log(`PASS continuing command ${width}: stable node, foreground movement, background stability, raw history, safe ANSI and stale response reset`);
  }
 } finally { await browser.close(); }
})().catch(e=>{console.error(e);process.exitCode=1;});
