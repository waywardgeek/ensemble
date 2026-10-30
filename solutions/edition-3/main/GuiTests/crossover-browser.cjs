// Run from solutions/edition-3/main after starting --crossover-host. This is a
// real-browser consumer of the mounted, embedded GUI, not a fake WebSocket UI.
const { chromium } = require('/Users/bill/projects/ensemble/agent/gui/node_modules/playwright');
const assert = require('node:assert/strict');
const fs = require('node:fs');
const crypto = require('node:crypto');
const path = require('node:path');
const root = process.cwd();
const evidence = path.join(root, '.crossover/evidence');
const base = 'http://127.0.0.1:18202/assistant/';
const check = (condition, message) => { assert.ok(condition, message); console.log('PASS '+message); };
(async () => {
 const context = await chromium.launchPersistentContext(path.join(root,'.crossover/profile-browser'), {
  executablePath:'/Applications/Google Chrome.app/Contents/MacOS/Google Chrome', headless:true,
  viewport:{width:1600,height:1000},
  args:['--no-first-run','--disable-background-networking']
 });
 const page = await context.newPage();
 const fetchOutput = (url, options={}) => context.request.get(url, {
  ...options, headers: { 'X-Ensemble-Output':'1', ...options.headers }
 });
 const errors=[]; const external=[]; const requests=[]; const frames=[];
 const controls=[];
 // Check actual rendered colors for representative text and input boundaries.
 // This is scoped contrast evidence, not an automated accessibility verdict.
 const checkContrast=async(theme)=>{
  const report=await page.evaluate(()=>{
   const rgb=s=>s.match(/[\d.]+/g).slice(0,3).map(Number);
   const luminance=c=>c.map(v=>{v/=255;return v<=.04045?v/12.92:((v+.055)/1.055)**2.4}).reduce((sum,v,i)=>sum+v*[.2126,.7152,.0722][i],0);
   const ratio=(a,b)=>{const x=[luminance(rgb(a)),luminance(rgb(b))].sort((a,b)=>a-b);return(x[1]+.05)/(x[0]+.05)};
   const background=e=>{while(e){const c=getComputedStyle(e).backgroundColor;if(c!=='rgba(0, 0, 0, 0)'&&c!=='transparent')return c;e=e.parentElement;}throw Error('No opaque background');};
   const text=['#conversation-agent','#connection','#state','#agents-tab','#settings-tab','#agent-tree .tree-state','#prompt','#send','.card-heading strong','.card-heading .btn'].map(selector=>{
    const e=document.querySelector(selector);return{selector,ratio:ratio(getComputedStyle(e).color,background(e))};
   });
   const prompt=document.querySelector('#prompt'),style=getComputedStyle(prompt);
   const toggle=document.querySelector('.tree-expand').getBoundingClientRect();
   return{text,inputBorder:ratio(style.borderTopColor,background(prompt)),target:{width:toggle.width,height:toggle.height}};
  });
  for(const item of report.text)check(item.ratio>=4.5,`${theme} text contrast ${item.selector}: ${item.ratio.toFixed(2)}:1`);
  check(report.inputBorder>=3,`${theme} input boundary contrast: ${report.inputBorder.toFixed(2)}:1`);
  check(report.target.width>=44&&report.target.height>=44,`${theme} tree expand target at least 44 pixels`);
 };

 const modelRequests=[];
 // Observe real outgoing controls, without replacing the socket or the actor.
 page.on('websocket', ws=>ws.on('framesent', frame=>{
  try { const value=JSON.parse(frame.payload); if(value.type==='set_model')modelRequests.push(value); if(['pause','unpause','interrupt'].includes(value.type)) controls.push(value.type); } catch {}
 }));
 page.on('pageerror', e=>errors.push(e.stack));
 page.on('request', r=>{ requests.push(r.url()); if (!/^(http:\/\/127\.0\.0\.1:18202\/|data:)/.test(r.url())) external.push(r.url()); });
 page.on('websocket', ws=>ws.on('framereceived', f=>{try { frames.push(JSON.parse(f.payload)); } catch {} }));
 try {
  await page.goto(base);
  await page.waitForFunction(()=>document.querySelector('#connection').textContent==='Connected' && !document.querySelector('#send').disabled);
  check(await page.locator('html').getAttribute('data-bs-theme')==='dark','fresh settings default to dark');
  await page.locator('#settings-tab').click();
  await page.locator('#human-name').fill('Taylor <friend>');
  await page.locator('#human-description').fill('Writes $TOOLS <examples>');
  await page.locator('#user-settings button').click();
  await page.waitForFunction(()=>[...document.querySelectorAll('article[data-role="human"] .card-heading strong')].some(el=>el.textContent==='Taylor <friend>'));
  await page.reload();
  await page.waitForFunction(()=>document.querySelector('#connection').textContent==='Connected' && !document.querySelector('#send').disabled);
  await page.waitForFunction(()=>document.querySelector('#human-name').value==='Taylor <friend>');
  check(await page.locator('#human-description').inputValue()==='Writes $TOOLS <examples>','User description survives reload without interpretation');
  await page.waitForFunction(()=>document.querySelector('article[data-role="assistant"] .card-heading strong')?.textContent==='captain');
  const names=await page.evaluate(()=>[...document.querySelectorAll('#transcript article[data-kind="text"]')].map(el=>({
   name:el.querySelector('.card-heading strong').textContent, title:el.querySelector('.card-heading strong').title,
   speak:el.querySelector('[data-action="speak"]').getAttribute('aria-label'),
   above:el.querySelector('.card-heading').getBoundingClientRect().bottom<=el.querySelector('.message').getBoundingClientRect().top+1
  })));
  check(names.length>0&&names.every(x=>x.above&&!/\b(Assistant|You)\b/.test(x.name+x.title+x.speak)),'names appear above messages, including accessible labels');
  check(await page.locator('.card-heading friend').count()===0,'user name is plain text, not HTML');
  await page.locator('#agents-tab').click();

  check(!(await page.locator('#narrate').isChecked()),'speech starts opt-in');
  await page.locator('.brand-mark').evaluate(image=>image.decode());
  const artwork=await page.locator('.brand-mark').evaluate(image=>({width:image.naturalWidth,height:image.naturalHeight,ratio:image.getBoundingClientRect().width/image.getBoundingClientRect().height}));
  check(artwork.width===128&&artwork.height===120&&Math.abs(artwork.ratio-128/120)<.01,'original proportional artwork loaded without crop');
  const favicon=await context.request.get(new URL('artwork/ensemble-32.png',base).href);
  check(favicon.ok()&&favicon.headers()['content-type']==='image/png','embedded PNG favicon has correct media type');
  check(await page.locator('#theme-toggle use').getAttribute('href')==='#icon-moon','dark theme uses moon with explicit switch action');
  check(await page.locator('#theme-toggle').getAttribute('aria-label')==='Switch to light theme','theme icon has accessible action');
  check(await page.locator('[data-action="speak"] svg').count()>0,'per-message Speak actions use locally defined icons');
  check(await page.locator('#stop-speech').getAttribute('aria-label')==='Stop speech','Stop speech remains explicitly named');
  await page.locator('#narrate').focus(); await page.keyboard.press('Space');
  await page.waitForFunction(()=>document.querySelector('#narrate').checked&&document.querySelector('.speech-toggle').title.startsWith('Speech on'));
  await page.keyboard.press('Space');
  await page.waitForFunction(()=>!document.querySelector('#narrate').checked&&document.querySelector('.speech-toggle').title.startsWith('Speech off'));
  check(true,'native speech icon toggles by keyboard and acknowledges saved on/off state');
  await page.locator('#send img').evaluate(image=>image.decode());
  const sendImage=await context.request.get(new URL('artwork/paper-airplane-send.svg',base).href);
  check(sendImage.ok()&&sendImage.headers()['content-type']==='image/svg+xml','supplied Send SVG loads with image media type');
  const composer=await page.locator('#compose').evaluate(el=>{
   const children=[...el.children].filter(x=>x.getBoundingClientRect().height>0);
   const input=el.querySelector('textarea'),button=el.querySelector('button');
   const a=input.getBoundingClientRect(),b=button.getBoundingClientRect();
   return {children:children.map(x=>x.id),sameRow:Math.min(a.bottom,b.bottom)>Math.max(a.top,b.top),placeholder:input.placeholder};
  });
  assert.deepEqual(composer.children,['prompt','send']);
  check(composer.sameRow,'composer is exactly input and Send in one row');
  check(composer.placeholder===`Ask ${await page.locator('#conversation-agent').textContent()} anything…`,'placeholder names the selected Agent');
  const measureComposer=()=>page.locator('#prompt').evaluate(el=>{const c=getComputedStyle(el);return{height:el.getBoundingClientRect().height,line:parseFloat(c.lineHeight),padding:parseFloat(c.paddingTop)+parseFloat(c.paddingBottom)+parseFloat(c.borderTopWidth)+parseFloat(c.borderBottomWidth),scroll:el.scrollHeight,client:el.clientHeight,overflow:c.overflowY};});
  const one=await measureComposer();
  check(Math.abs(one.height-one.line-one.padding)<2,'composer starts at one line');
  await page.locator('#prompt').fill('first\nsecond');const two=await measureComposer();
  check(Math.abs(two.height-one.height)<2,'multiline input keeps a one-line composer');
  await page.locator('#prompt').fill('a wrapped message '.repeat(100));const many=await measureComposer();
  check(Math.abs(many.height-one.height)<2&&many.scroll>many.client&&many.overflow==='auto','long wrapped input scrolls inside one-line area');
  await page.locator('#prompt').fill('');const empty=await measureComposer();
  check(Math.abs(empty.height-one.height)<2,'clearing input preserves one-line area');
  await page.waitForFunction(()=>document.querySelector('#state').textContent==='inputpending');

  // Structural presentation checks do not certify pixel-level beauty. These
  // protect the selected design's distinction between navigation and actions.
  await page.waitForFunction(()=>document.querySelector('#conversation-agent').textContent!=='Waiting');
  const design=await page.evaluate(()=>{
   const tab=document.querySelector('#agents-tab'), tree=document.querySelector('.tree-select');
   const t=getComputedStyle(tab), row=getComputedStyle(tree);
   return {title:document.querySelector('#conversation-agent').textContent,
    selected:document.querySelector('#agent-choice').value,
    tabButtonClass:tab.classList.contains('btn'),tabBottom:t.borderBottomColor,
    tabBackground:t.backgroundColor,tabBottomRadius:t.borderBottomLeftRadius,
    treeBorder:row.borderTopWidth,treeBackground:row.backgroundColor,
    sendPrimary:document.querySelector('#send').classList.contains('btn-primary'),
    controlNames:['pause','resume','interrupt'].map(id=>document.getElementById(id).getAttribute('aria-label'))};
  });
  check(design.title===design.selected,'conversation title reflects selected Agent');
  check(!design.tabButtonClass && design.tabBottom===design.tabBackground && design.tabBottomRadius==='0px','tabs connect to their pane instead of appearing as action buttons');
  check(design.treeBorder==='0px' && design.treeBackground==='rgba(0, 0, 0, 0)','tree rows use quiet navigation styling');
  check(design.sendPrimary && design.controlNames.join(',')==='Pause,Resume,Interrupt','primary Send and named compact Agent controls');
  await checkContrast('dark');
  await page.screenshot({path:path.join(evidence,'design-dark-initial.png')});

  const markdownShape=await page.evaluate(()=>{
   const target=document.createElement('div');
   target.className='part';document.body.append(target);
   markdown(target,'- first\n- second\n\n1. one\n2. two\n\nA wrapped\nparagraph with **bold** and *emphasis*.');
   const shape={whiteSpace:getComputedStyle(target.querySelector('p')).whiteSpace,ul:target.querySelectorAll('ul').length,ol:target.querySelectorAll('ol').length,
    li:target.querySelectorAll('li').length,p:target.querySelectorAll('p').length,
    text:target.querySelector('p')?.textContent.replace(/\s+/g,' '),strong:target.querySelectorAll('strong').length,
    em:target.querySelectorAll('em').length};
   target.remove();return shape;
  });
  assert.deepEqual(markdownShape,{whiteSpace:'normal',ul:1,ol:1,li:4,p:1,text:'A wrapped paragraph with bold and emphasis.',strong:1,em:1});
  console.log('PASS semantic Markdown lists, wrapped paragraphs and inline emphasis');
  // Geometry-only stress: this does not test Engine accounting. Preserve and
  // restore the real values; later assertions test actual acknowledged usage.
  const geometry=await page.evaluate(()=>{
   const ids=['model-choice','narrate','pause','resume','interrupt','theme-toggle'];
   const measure=()=>Object.fromEntries(ids.map(id=>{
    const r=document.getElementById(id).getBoundingClientRect();return[id,[r.x,r.y,r.width,r.height]];
   }));
   const targets={'conversation-agent':'A very long selected agent name that must not move controls',
    state:'a-long-changing-state-value',connection:'Disconnected — reconnecting',
    'usage-cache':'123,456,789','usage-hit':'99.99%',
    'usage-all':'999,999,999','usage-cost':'$1234.56789'};
   const saved=Object.fromEntries(Object.keys(targets).map(id=>[id,document.getElementById(id).textContent]));
   const before=measure();
   for(const [id,text] of Object.entries(targets))document.getElementById(id).textContent=text;
   const after=measure();
   for(const [id,text] of Object.entries(saved))document.getElementById(id).textContent=text;
   return {before,after};
  });
  for(const id of Object.keys(geometry.before)) for(let i=0;i<4;i++)
   assert.ok(Math.abs(geometry.before[id][i]-geometry.after[id][i])<1,`${id} moved during status-only changes`);
  console.log('PASS status changes leave control rectangles stable');
  check(await page.locator('#set-model, #funding-source, #selected-name').count()===0,'header removes Apply, repeated Agent name and funding status');
  check(await page.locator('.model-caption').count()===0,'caption removed from header');
  const header=await page.evaluate(()=>({
   height:document.querySelector('#status-bar').getBoundingClientRect().height,
   groups:['.run-state','#model-controls','#usage-meter','.status-actions'].map(selector=>{const r=document.querySelector(selector).getBoundingClientRect();return {left:r.left,right:r.right,centerX:(r.left+r.right)/2,centerY:(r.top+r.bottom)/2};}),
   slots:[...document.querySelectorAll('#usage-meter output')].map(e=>({width:e.getBoundingClientRect().width,align:getComputedStyle(e).textAlign,gap:parseFloat(getComputedStyle(e.parentElement).columnGap)}))
  }));
  check(header.height<=55,'desktop header is at most 55px at the unchanged 18px font');
  check(header.groups.every(g=>Math.abs(g.centerY-header.groups[0].centerY)<1),'all header groups share one centered row');
  check(header.groups.every((g,i)=>!i||g.left>=header.groups[i-1].right),'status, model, usage and actions appear in order without overlap');
  check(Math.abs(header.groups[1].centerX-(page.viewportSize().width/2))<1,'model is centered in the entire header');
  check(Math.abs(header.groups[2].centerX-(header.groups[1].right+header.groups[3].left)/2)<1,'usage is centered between model and actions');
  check(header.slots.every(s=>s.width>0&&s.align==='left'&&s.gap<5),'usage figures occupy fixed left-aligned slots beside labels');
  const usageFits=await page.evaluate(()=>{
   const values={'usage-cost':'$1234.56','usage-all':'1,234,567','usage-cache':'1,234,567','usage-hit':'100.0%'};
   return Object.entries(values).every(([id,value])=>{const e=document.getElementById(id),saved=e.textContent;e.textContent=value;const fits=e.scrollWidth<=e.clientWidth+1;e.textContent=saved;return fits;});
  });
  check(usageFits,'tight numeric slots retain seven-digit counts, cost and full hit rate');
  const originalModel=await page.locator('#model-choice').inputValue();
  const otherModel=await page.locator('#model-choice option').evaluateAll((options,current)=>options.find(o=>o.value!==current)?.value,originalModel);
  check(!!otherModel,'fixture supplies another same-provider model');
  const startRequests=modelRequests.length;
  const startFrames=frames.length;
  await page.locator('#model-choice').selectOption(otherModel);
  await page.waitForFunction(model=>!document.querySelector('#model-choice').disabled&&document.querySelector('#model-choice').value===model,otherModel);
  check(modelRequests.length===startRequests+1&&modelRequests.at(-1).text===otherModel,'one selection immediately sends one model change');
  check(frames.slice(startFrames).some(f=>f.type==='usage'&&f.selected_model===otherModel),'server acknowledges changed model');
  await page.locator('#model-choice').evaluate(el=>{const o=new Option('Invalid empty model',' ');el.add(o);});
  await page.locator('#model-choice').selectOption(' ');
  await page.waitForFunction(model=>!document.querySelector('#model-choice').disabled&&document.querySelector('#model-choice').value===model,otherModel);
  check((await page.locator('#notice').textContent()).includes('Model selection refused'),'refused model restores acknowledged choice');
  console.log('PASS immediate model selection, acknowledgment and refusal recovery');

  check(await page.locator('#tool-transcript .part').count()===6,'tool call, results and media appear only in Tools');
  check(await page.locator('#transcript [data-kind="tool_call"],#transcript [data-kind="tool_result"]').count()===0,'Chat excludes tools');
  check(await page.locator('#tool-transcript img').evaluate(e=>e.complete && e.naturalWidth>0),'embedded image renders');
  const originalIDs=await page.locator('[data-part-id]').evaluateAll(nodes=>nodes.map(n=>n.dataset.partId));
  check(new Set(originalIDs).size===originalIDs.length,'replay IDs are unique across both panes');
  // Bootstrap owns keyboard tab selection; hidden panels must leave focus order.
  await page.locator('#agents-tab').focus(); await page.keyboard.press('ArrowRight');
  check(await page.locator('#settings-tab').getAttribute('aria-selected')==='true','tab ArrowRight selects Settings');
  check(await page.locator('#agents-panel').isHidden(),'inactive Agents panel is hidden');
  await page.keyboard.press('ArrowLeft');
  check(await page.locator('#agents-tab').getAttribute('aria-selected')==='true','tab ArrowLeft returns to Agents');
  await page.locator('[data-agent="captain"]').focus(); await page.keyboard.press('ArrowRight');
  check(await page.evaluate(()=>document.activeElement.dataset.agent)==='captain/worker','tree ArrowRight moves to child');
  await page.keyboard.press('ArrowRight');
  check(await page.evaluate(()=>document.activeElement.dataset.agent)==='captain/worker/nested','tree reaches grandchild');
  await page.keyboard.press('ArrowLeft');
  check(await page.evaluate(()=>document.activeElement.dataset.agent)==='captain/worker','tree ArrowLeft returns to parent');
  await page.locator('#left-splitter').focus(); const before=await page.locator('#left-splitter').getAttribute('aria-valuenow');
  await page.keyboard.press('ArrowRight');
  check(await page.locator('#left-splitter').getAttribute('aria-valuenow')!==before,'separator supports keyboard resizing');
  const savedWidths=await page.evaluate(()=>['left','right'].map(side=>document.documentElement.style.getPropertyValue(`--${side}-width`)));
  await page.locator('#left-splitter').press('Home');await page.locator('#right-splitter').press('End');
  const paneGeometry=()=>page.evaluate(()=>Object.fromEntries(['workspace','left-pane','chat-pane','tools-pane'].map(id=>[id,document.getElementById(id).getBoundingClientRect().width])));
  let panes=await paneGeometry();
  check(panes['tools-pane']>panes.workspace*.7&&panes['chat-pane']>=159&&panes['chat-pane']<170,'Actions can occupy most of the workspace while chat is narrow');
  await page.setViewportSize({width:1200,height:1000});
  await page.waitForFunction(()=>document.querySelector('#chat-pane').getBoundingClientRect().width>=159);
  panes=await paneGeometry();check(['left-pane','chat-pane','tools-pane'].every(id=>panes[id]>=159),'window resizing preserves reachable minimum panes');
  await page.setViewportSize({width:1600,height:1000});
  const handle=await page.locator('#right-splitter').boundingBox();
  await page.mouse.move(handle.x+handle.width/2,handle.y+20);await page.mouse.down();await page.mouse.move(1595,handle.y+20,{steps:8});await page.mouse.up();
  panes=await paneGeometry();check(panes['chat-pane']>panes.workspace*.7,'mouse dragging can give the conversation most of the workspace');
  check(await page.locator('#transcript article[data-kind="text"]').first().evaluate(el=>el.getBoundingClientRect().width)>900,'messages use the available wide conversation instead of a fixed content cap');
  await page.evaluate(values=>['left','right'].forEach((side,i)=>{const style=document.documentElement.style;values[i]?style.setProperty(`--${side}-width`,values[i]):style.removeProperty(`--${side}-width`);}),savedWidths);

  // Full output is owned by an Agent and Job, not by the browser's displayed path.
  const fixture=JSON.parse(fs.readFileSync('.crossover/runtime/fixture.json'));
  const output=await fetchOutput(base+`jobs/output?agent=captain&handle=${fixture.handle}`);
  const complete=await output.text();
  check(output.status()===200 && complete.includes('fixture-line-1999') && complete.length>60000,'complete output extends far beyond bounded preview');
  check(!output.headers()['content-type'].includes('html') && output.headers()['x-content-type-options']==='nosniff','output is passive text');
  const downloaded=await fetchOutput(base+`jobs/output?agent=captain&handle=${fixture.handle}&download=1`);
  check(downloaded.headers()['content-disposition'].includes('attachment') && await downloaded.text()===complete,'download returns complete bytes');
  for (const [query,status] of [
   [`agent=peer&handle=${fixture.handle}`,404], [`agent=captain/worker&handle=${fixture.handle}`,404],
   [`agent=removed&handle=${fixture.removedHandle}`,404], [`agent=captain&handle=999999`,404],
   [`agent=captain&handle=${fixture.handle}&path=/etc/passwd`,400], [`agent=captain&handle=${fixture.handle}&agent=peer`,400]
  ]) check((await fetchOutput(base+'jobs/output?'+query)).status()===status,'output refusal '+query);
  check((await fetchOutput(base+`jobs/output?agent=captain&handle=${fixture.handle}`,{headers:{Origin:'https://foreign.invalid'}})).status()===403,'foreign Origin refused');
  check((await fetchOutput(base+`jobs/output?agent=captain&handle=${fixture.handle}`,{headers:{'Sec-Fetch-Site':'cross-site'}})).status()===403,'cross-site navigation refused');
  const running=await fetchOutput(base+`jobs/output?agent=captain&handle=${fixture.runningHandle}`);
  check(running.status()===409,'running output explicitly refuses as incomplete');
  const nested=await fetchOutput(base+`jobs/output?agent=captain/worker&handle=${fixture.nestedHandle}`);
  check(nested.status()===200 && await nested.text()==='nested-owned-output\nexit_code: 0\n','nested owner address resolves its own output');
  check((await fetchOutput(base+`jobs/output?agent=captain&handle=${fixture.nestedHandle}`)).status()===404,'parent cannot claim child Job by handle');
  check((await context.request.get(base+`jobs/output?agent=captain&handle=${fixture.handle}`)).status()===403,'missing custom header refuses');
  for(const [agent,handle,label] of [['files',1,'unreadable'],['files',2,'FIFO'],['files',3,'missing file'],['confined',1,'rooted symlink escape']]) {
   const refusal=await fetchOutput(base+`jobs/output?agent=${agent}&handle=${handle}`, {timeout:3000});
   check(refusal.status()===404 && !(await refusal.text()).includes('must-not-cross-root'),label+' backing file refuses without blocking or revealing content');
  }

  check((await fetchOutput(base+`jobs/output?agent=captain&handle=${fixture.oversizedHandle}`)).status()===413,'oversized output explicitly refuses');
  check(crypto.createHash('sha256').update(await output.body()).digest('hex').toUpperCase()===fixture.outputHash,'complete response exact bytes match owned output');
  const downloadEvent=page.waitForEvent('download');
  await page.getByRole('button',{name:'Download complete output',exact:true}).first().click();
  const browserDownload=await downloadEvent;
  await browserDownload.saveAs(path.join(evidence,'downloaded-output.txt'));
  check(crypto.createHash('sha256').update(fs.readFileSync(path.join(evidence,'downloaded-output.txt'))).digest('hex').toUpperCase()===fixture.outputHash,'guarded browser download exact bytes');
  const cursorOutput=await fetchOutput(base+`jobs/output?agent=captain&handle=${fixture.cursorHandle}`);
  check(cursorOutput.status()===200 && (await cursorOutput.text()).includes('unobserved-suffix'),'download completed two-stage output before ordinary observation');
  const observed=await context.request.get('http://127.0.0.1:18202/fixture/observe');
  check((await observed.text()).includes('unobserved-suffix'),'later ordinary Job observer still receives unconsumed tail');
  check(await page.getByRole('button',{name:/^Copy(?: code| text)?$/}).count()===0,'no Copy buttons in messages, code or output');
  check((await page.locator('details pre code').allTextContents()).some(text=>text.includes('value65')),'complete code remains available for expansion and selection');
  await page.locator('#transcript details').first().locator('summary').click();
  check(await page.locator('#transcript .code-keyword').count()>0,'expanded code has safe token highlighting');
  await page.screenshot({path:path.join(evidence,'desktop-dark.png')});
  await page.locator('[data-agent="peer"]').click();
  await page.waitForFunction(()=>document.querySelector('#transcript').textContent.includes('PEER-ONLY'));
  check(!(await page.locator('#transcript').innerText()).includes('Offline crossover'),'selection replaces Chat replay');
  check(await page.locator('#tool-transcript .part').count()===0,'selection clears old tool display');
  check((await page.locator('#relay-agent').textContent())==='captain','view selection does not retarget relay authority');
  await page.locator('[data-agent="captain"]').click();
  await page.waitForFunction(()=>document.querySelector('#transcript').textContent.includes('Offline crossover'));
  await page.locator('#prompt').fill('Stream the fixture response'); await page.locator('#send').click();
  await page.waitForFunction(()=>document.querySelector('#transcript').textContent.includes('STREAM-IDENTITY first'));
  check(await page.locator('#model-choice').isDisabled(),'model mutation disabled while running');
  check(await page.locator('#send').isEnabled(),'ordinary input remains available during a turn');
  check(await page.locator('#interrupt').isEnabled(),'interrupt enabled while running');
  const streamed=await page.locator('#transcript .part').filter({hasText:'STREAM-IDENTITY first'}).getAttribute('data-part-id');
  await page.waitForFunction(()=>document.querySelector('#transcript').textContent.includes('and final.'));
  await page.waitForFunction(()=>!document.querySelector('#model-choice').disabled);
  check(await page.locator('#transcript .part').filter({hasText:'STREAM-IDENTITY first and final.'}).count()===1,'stream final updates one card');
  check(await page.locator('#transcript .part').filter({hasText:'STREAM-IDENTITY first and final.'}).getAttribute('data-part-id')===streamed,'stream/final part identity stable');
  check((await page.locator('#usage-all').textContent())==='31','meter uses acknowledged Engine input');
  const paused=()=>page.waitForFunction(()=>document.querySelector('#state').textContent==='paused');
  const resumed=()=>page.waitForFunction(()=>['idle','inputpending'].includes(document.querySelector('#state').textContent) && !document.querySelector('#model-choice').disabled);
  const expectNoUnpause=async (action,label)=>{
   const start=controls.length; await action();
   // A settings acknowledgment is ordered after controls on the same socket.
   // It gives the real actor a processing boundary without a quiet-time guess.
   const theme=await page.locator('html').getAttribute('data-bs-theme');
   await page.locator('#theme-toggle').click();
   await page.waitForFunction(old=>document.documentElement.dataset.bsTheme!==old,theme);
   check(!controls.slice(start).includes('unpause'),label+' sends no automatic unpause');
   check(await page.locator('#state').textContent()==='paused',label+' leaves real actor paused');
   await page.locator('#theme-toggle').click();
   await page.waitForFunction(old=>document.documentElement.dataset.bsTheme===old,theme);
  };
  await page.locator('#pause').click(); await paused();
  check(await page.locator('#model-choice').isDisabled(),'paused model mutation stays disabled');
  await expectNoUnpause(async()=>{
   await page.locator('#prompt').fill('manual pause then typing');await page.locator('#prompt').fill('');
  },'manual Pause then type/clear');
  await expectNoUnpause(async()=>{
   await page.locator('#prompt').fill('admitted while manually paused');await page.locator('#send').click();
  },'manual Pause then type/Send');
  await page.locator('#resume').click();
  await resumed();
  await page.locator('#prompt').fill('automatic typing pause');await paused();
  await page.locator('#pause').click();
  await expectNoUnpause(async()=>{await page.locator('#prompt').fill('');},'typing then manual Pause then clear');
  await expectNoUnpause(async()=>{
   await page.locator('#prompt').fill('second manually paused turn');await page.locator('#send').click();
  },'typing then manual Pause then Send');
  await page.locator('#resume').click();await resumed();
  // A purely automatic pause still releases; explicit Resume with an unchanged
  // draft must also work, without immediately re-pausing the actor.
  await page.locator('#prompt').fill('automatic-only');await paused();
  await page.locator('#prompt').fill('');await resumed();
  await page.locator('#prompt').fill('resume despite draft');await paused();
  await page.locator('#resume').click();await resumed();
  const afterExplicitResume=controls.length;
  await page.locator('#prompt').fill('');
  check(!controls.slice(afterExplicitResume).includes('unpause'),'clearing explicitly resumed draft does not release another pause');
  // Another tab's existing pause is not owned by this tab's later draft.
  const other=await context.newPage();await other.goto(base);
  await other.waitForFunction(()=>!document.querySelector('#pause').disabled);
  await other.locator('#pause').click();await paused();
  await expectNoUnpause(async()=>{
   await page.locator('#prompt').fill('already paused elsewhere');await page.locator('#prompt').fill('');
  },'other tab Pause then local typing');
  await other.close();
  // Selection must not release a manual pause on the old Agent or carry local
  // pause ownership into the newly selected conversation.
  await page.locator('[data-agent="peer"]').click();
  await page.waitForFunction(()=>document.querySelector('#conversation-agent').textContent==='peer' && !document.querySelector('#send').disabled);
  check(await page.locator('#prompt').getAttribute('placeholder')==='Ask peer anything…','placeholder follows Agent selection');
  await page.locator('[data-agent="captain"]').click();await paused();
  await expectNoUnpause(async()=>{
   await page.locator('#prompt').fill('after selection');await page.locator('#prompt').fill('');
  },'selection preserves old manual pause');
  await page.reload();await page.waitForFunction(()=>!document.querySelector('#resume').disabled);
  await expectNoUnpause(async()=>{
   await page.locator('#prompt').fill('after reconnect');await page.locator('#prompt').fill('');
  },'reconnect does not claim existing pause');
  await page.locator('#resume').click();await resumed();
  await page.locator('#prompt').fill('automatic pause before interrupt');await paused();
  await page.locator('#interrupt').click();
  await expectNoUnpause(async()=>{await page.locator('#prompt').fill('');},'interrupt clears automatic release intent');
  await page.locator('#resume').click();await resumed();
  const beforeReconnect=await page.locator('#transcript .part').evaluateAll(nodes=>nodes.map(n=>n.dataset.partId));
  // Disconnect/reconnect must rebuild the same canonical cards, not append them.
  await context.setOffline(true); await page.waitForTimeout(300); await context.setOffline(false);
  await page.waitForFunction(()=>document.querySelector('#connection').textContent==='Connected' && !document.querySelector('#send').disabled);
  await page.reload(); await page.waitForFunction(()=>!document.querySelector('#send').disabled);
  assert.deepEqual(await page.locator('#transcript .part').evaluateAll(nodes=>nodes.map(n=>n.dataset.partId)),beforeReconnect);
  console.log('PASS reconnect/reload preserves canonical card identities');
  // Acknowledged theme/font are durable, then deliberately tested together.
  await page.locator('#theme-toggle').click(); await page.waitForFunction(()=>document.documentElement.dataset.bsTheme==='light');
  check(await page.locator('#theme-toggle use').getAttribute('href')==='#icon-sun','light theme uses sun');
  await checkContrast('light');
  await page.screenshot({path:path.join(evidence,'desktop-light.png')});
  await page.locator('#settings-tab').click(); await page.locator('#font-size').fill('32'); await page.locator('#font-size').dispatchEvent('change');
  await page.waitForFunction(()=>getComputedStyle(document.documentElement).fontSize==='32px');
  check(JSON.parse(fs.readFileSync('.crossover/runtime/appearance.json')).font_size===32,'font size acknowledged after durable save');
  await page.setViewportSize({width:390,height:844});
  await page.waitForFunction(()=>document.body.classList.contains('compact'));
  check(await page.locator('#pane-navigation').isVisible(),'pane navigation remains reachable at narrow largest text');
  for (const pane of ['agents','chat','tools']) {
   await page.locator(`button[data-pane="${pane}"]`).click();
   check(await page.locator('#workspace').getAttribute('data-pane')===pane,'narrow pane switch '+pane);
   check(await page.evaluate(()=>document.documentElement.scrollWidth<=window.innerWidth+1),'no page horizontal overflow in '+pane);
   if (pane==='chat') {
    const bounds=await page.locator('#compose').boundingBox();
    check(bounds.height>=44 && bounds.y+bounds.height<=844,'composer remains on-screen at largest text/narrow width');
    await page.locator('#prompt').fill('narrow-window text');
    check(await page.locator('#send').isVisible(),'Send remains reachable');
   }
   await page.screenshot({path:path.join(evidence,`narrow-32-${pane}.png`)});
  }
  await page.reload(); await page.waitForFunction(()=>!document.querySelector('#theme-toggle').disabled);
  check(await page.locator('html').getAttribute('data-bs-theme')==='light','light preference survives reload');
  check(await page.evaluate(()=>getComputedStyle(document.documentElement).fontSize)==='32px','largest font survives reload');
  await page.locator('button[data-pane="agents"]').click(); await page.locator('#settings-tab').click();
  await page.locator('#font-size').fill('18'); await page.locator('#font-size').dispatchEvent('change');
  await page.waitForFunction(()=>getComputedStyle(document.documentElement).fontSize==='18px');
  await page.locator('#theme-toggle').click(); await page.waitForFunction(()=>document.documentElement.dataset.bsTheme==='dark');
  // Cross enlarged text with narrow/intermediate windows. Manual textarea
  // sizing must not push Send outside the viewport. Scrollable header groups
  // must permit reaching each control without page-wide horizontal overflow.
  for(const size of [18,32]) {
   await page.locator('button[data-pane="agents"]').click();await page.locator('#settings-tab').click();
   await page.locator('#font-size').fill(String(size));await page.locator('#font-size').dispatchEvent('change');
   await page.waitForFunction(value=>getComputedStyle(document.documentElement).fontSize===`${value}px`,size);
   for(const [width,height] of [[390,844],[390,1000],[768,1000],[1536,1000]]) {
    await page.setViewportSize({width,height});
    await page.waitForFunction(()=>document.body.classList.contains('compact')===(innerWidth<62*parseFloat(getComputedStyle(document.documentElement).fontSize)));
    if(await page.locator('#pane-navigation').isVisible())await page.locator('button[data-pane="chat"]').click();
    await page.locator('#prompt').evaluate(el=>el.style.height='1000px');
    const layout=await page.evaluate(()=>{
     const ids=['model-choice','narrate','pause','resume','interrupt','theme-toggle'];
     const measure=()=>ids.map(id=>{
      const r=document.getElementById(id).getBoundingClientRect();return[r.x,r.y,r.width,r.height];
     });
     const saved=['conversation-agent','state','usage-all','usage-cost'].map(id=>[id,document.getElementById(id).textContent]);
     const before=measure();
     for(const [id] of saved)document.getElementById(id).textContent='Long changing status: 123,456,789.12345';
     const after=measure();
     for(const [id,text] of saved)document.getElementById(id).textContent=text;
     const send=document.querySelector('#send').getBoundingClientRect();
     return {before,after,overflow:document.documentElement.scrollWidth>innerWidth+1,send:{top:send.top,bottom:send.bottom,right:send.right}};
    });
    assert.equal(layout.overflow,false,`page overflow ${width}/${size}`);
    assert.ok(layout.send.top>=0 && layout.send.bottom<=height && layout.send.right<=width,`Send unreachable ${width}x${height}/${size}: ${JSON.stringify(layout.send)}`);
    for(let i=0;i<layout.before.length;i++)for(let j=0;j<4;j++)assert.ok(Math.abs(layout.before[i][j]-layout.after[i][j])<1,`status moved control ${i} at ${width}/${size}`);
    for(const id of ['model-choice','narrate','pause','resume','interrupt','theme-toggle']) {
     await page.locator('#'+id).scrollIntoViewIfNeeded();
     assert.ok(await page.locator('#'+id).evaluate(el=>{
      const r=el.getBoundingClientRect(),bar=document.querySelector('#status-bar').getBoundingClientRect();
      return r.left>=0 && r.right<=innerWidth+1 && r.top>=bar.top-1 && r.bottom<=bar.bottom+1;
     }),`status control ${id} clipped at ${width}/${size}`);
    }
    console.log(`PASS stable header, reachable controls/composer and no page overflow at ${width}x${height}px / font ${size}`);
    await page.locator('#prompt').evaluate(el=>el.style.height='');
   }
   await page.setViewportSize({width:390,height:844});
   await page.waitForFunction(()=>document.body.classList.contains('compact'));
  }
  check(external.length===0,'no external page requests'); check(errors.length===0,'no browser script errors');
  fs.writeFileSync(path.join(evidence,'browser-network.json'),JSON.stringify({requests,external,errors,frameTypes:[...new Set(frames.map(f=>f.type))]},null,2));
  console.log('PASS crossover browser checks complete');
 } catch(error) {
  await page.screenshot({path:path.join(evidence,'browser-failure.png')}).catch(()=>{});
  fs.writeFileSync(path.join(evidence,'browser-failure.json'),JSON.stringify({error:String(error),errors,external,controls,
   ui:await page.evaluate(()=>({state:document.querySelector('#state').textContent,
    connection:document.querySelector('#connection').textContent,
    pause:document.querySelector('#pause').disabled,resume:document.querySelector('#resume').disabled})).catch(()=>null),
   frames:frames.slice(-30)},null,2));
  throw error;
 } finally { await context.close(); }
})().catch(e=>{console.error(e);process.exitCode=1;});
