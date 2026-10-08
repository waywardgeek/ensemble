#!/usr/bin/env node
// Real Chrome, actual delivered HTML/CSS and public browser components. Socket
// and native speech are controlled; actual audio/provider acceptance is separate.
const fs=require('fs'),path=require('path'),http=require('http'),os=require('os'),crypto=require('crypto');
const {chromium}=require(path.join(process.env.CH07_BROWSER_TOOLS||'/Users/bill/projects/ensemble-edition-2-revisions/browser-tools','node_modules/playwright'));
const source=path.resolve(process.argv[2]),copy=fs.mkdtempSync(path.join(os.tmpdir(),'ch08-browser-')),assets=path.join(copy,'assets');
const only=process.argv[3]||'';
fs.cpSync(path.join(source,'gui/web/gui'),assets,{recursive:true});
const html=fs.readFileSync(path.join(assets,'index.html'),'utf8').replace(/<script\b[^>]*>[\s\S]*?<\/script>/gi,'');
const server=http.createServer((req,res)=>{
 if(req.url==='/'){res.setHeader('Content-Type','text/html');res.end(html);return;}
 const file=req.url==='/__review.js'?path.join(__dirname,'ch08-browser-fixture.js'):path.join(assets,decodeURIComponent(req.url.split('?')[0]));
 if((!file.startsWith(assets+path.sep)&&req.url!=='/__review.js')||!fs.existsSync(file)){res.writeHead(404);res.end();return;}
 res.setHeader('Content-Type',file.endsWith('.css')?'text/css':'text/javascript');res.end(fs.readFileSync(file));
});
const checks=[],expect=(value,message)=>{if(!value)throw Error(message);};
(async()=>{
 await new Promise(r=>server.listen(0,'127.0.0.1',r));const origin='http://127.0.0.1:'+server.address().port;
 const browser=await chromium.launch({channel:'chrome',headless:true});
 async function run(id,fn,options={}){if(only&&only!==id)return;const page=await browser.newPage({viewport:{width:1500,height:1000}}),errors=[];page.setDefaultTimeout(2500);page.on('pageerror',e=>errors.push(e.message));await page.route('**/*',r=>r.request().url().startsWith(origin+'/')?r.continue():r.abort());
  try{await page.goto(origin);await page.evaluate(async options=>{window.review=await(await import('/__review.js')).setup(options)},options);const result=await fn(page);expect(errors.length===0,'uncaught browser errors: '+errors.join('; '));checks.push({id,passed:true,result});}
  catch(e){checks.push({id,passed:false,error:e.stack,page_errors:errors});}finally{await page.close();}
 }
 await run('remote-style-preserves-cards-focus-and-system-theme',async page=>{
  await page.evaluate(async()=>{
   review.a.observation({kind:'message_received',agent_id:'a1',event:{seq:1,type:'message_received',message:{actor:'human',purpose:'dialogue',parts:[{type:'text',text:'RETAINED CARD'}]}}});
   window.card=document.querySelector('#review-a article');if(!card)throw Error('positive retained card absent');review.a.input.focus();await review.remote({theme:'light',font_size:22,sidebar_width:300,actions_width:420});
  });
  const result=await page.evaluate(()=>({same:card===document.querySelector('#review-a article'),focused:document.activeElement===review.a.input,values:[review.a,review.b].map(p=>({theme:p.root.dataset.theme,font:getComputedStyle(p.root.querySelector('article')||p.input).fontSize,sidebar:p.root.querySelector('[data-divider="sidebar_width"]').getAttribute('aria-valuenow'),actions:p.root.querySelector('[data-divider="actions_width"]').getAttribute('aria-valuenow')}))}));
  expect(result.same&&result.focused,'restyling rebuilt retained cards or moved focus');expect(result.values.every(x=>x.theme==='light'&&x.font==='22px'&&x.sidebar==='300'&&x.actions==='420'),'remote preferences lack a visible consumer in both Pages');
  await page.evaluate(()=>review.remote({theme:'system'}));await page.emulateMedia({colorScheme:'dark'});await page.waitForFunction(()=>review.a.root.dataset.theme==='dark');await page.emulateMedia({colorScheme:'light'});await page.waitForFunction(()=>review.b.root.dataset.theme==='light');return result;
 });
 await run('semantic-keyboard-pointer-and-unsaved-value',async page=>{
  const root=page.locator('#review-a');await root.getByRole('tab',{name:'Settings',exact:true}).click();expect(await root.getByRole('tab',{name:'Settings',exact:true}).getAttribute('aria-selected')==='true','sidebar selection not inspectable');
  const divider=root.getByRole('separator',{name:'Sidebar width',exact:true});await divider.focus();await page.keyboard.press('ArrowRight');await page.waitForFunction(()=>review.snapshot().preferences.sidebar_width===270);
  const before=await page.evaluate(()=>review.commands.filter(m=>m.type==='preferences_update').length),box=await divider.boundingBox();await page.mouse.move(box.x+box.width/2,box.y+box.height/2);await page.mouse.down();await page.mouse.move(box.x+box.width/2+40,box.y+box.height/2,{steps:4});
  expect(await page.evaluate(()=>review.commands.filter(m=>m.type==='preferences_update').length)===before,'pointer motion persisted every intermediate width');await page.mouse.up();await page.waitForFunction(()=>review.snapshot().preferences.sidebar_width===310);
  expect(await page.evaluate(()=>review.commands.filter(m=>m.type==='preferences_update').length)===before+1,'completed drag did not send exactly one sparse patch');
  const font=root.getByRole('spinbutton',{name:'Font size',exact:true});await font.fill('99');await font.blur();await page.waitForFunction(()=>/not saved/.test(review.a.notice.textContent));expect(await page.evaluate(()=>review.snapshot().preferences.font_size)===16,'invalid editor value became applied');
  expect((await root.locator('[data-preferences-applied]').textContent()).includes('16px'),'applied summary relabeled invalid draft');
  const autoplay=root.getByRole('checkbox',{name:'Autoplay new answers',exact:true});await autoplay.check();await page.waitForFunction(()=>review.snapshot().preferences.autoplay===true);await autoplay.uncheck();await page.waitForFunction(()=>review.snapshot().preferences.autoplay===false);expect(await page.locator('#review-b [data-preference="autoplay"]').isChecked()===false,'remote checkbox lost explicit false');
  const policy=root.getByRole('spinbutton',{name:'Maximum model requests per turn',exact:true});await policy.fill('2');await root.getByRole('button',{name:'Apply Agent policy',exact:true}).click();await page.waitForFunction(()=>review.b.settings.executionPolicy.max_model_requests===2);await policy.fill('0');await root.getByRole('button',{name:'Apply Agent policy',exact:true}).click();await page.waitForFunction(()=>review.b.settings.executionPolicy.max_model_requests===0);expect((await root.locator('[data-policy-applied]').textContent()).includes('Default (16)'),'zero limit lacks effective default label');
  const sent=await page.evaluate(()=>review.commands.length);await page.setViewportSize({width:600,height:900});await page.waitForTimeout(50);expect(await page.evaluate(()=>review.snapshot().preferences.sidebar_width)===310,'viewport rewrote desired width');expect(await page.evaluate(()=>review.commands.length)===sent,'resize sent an unsolicited settings patch');expect(await root.locator('[data-input]').evaluate(e=>e.getBoundingClientRect().width)>150,'narrow viewport made central input unusable');await page.setViewportSize({width:1500,height:1000});return {sidebar:310};
 });
 await run('mixed-parts-safe-expansion-and-restyle-identity',async page=>{
  await page.evaluate(()=>{const p=review.a,id={agent_id:'a1',request_id:'r1',operation_id:'m1'};p.observation({...id,kind:'part_delta',part_id:0,channel:'text',text:'ANSWER'});p.observation({...id,kind:'part_final',part_id:0,response_seq:5,part_index:0,part:{type:'text',text:'ANSWER'}});p.observation({...id,kind:'part_final',part_id:1,response_seq:5,part_index:1,part:{type:'tool_call',call_id:'c1',name:'read_file',args:{path:'notes.txt'}}});p.observation({agent_id:'a1',kind:'response_ended',event:{seq:5,type:'response_ended',response:{parts:[{type:'text',text:'ANSWER'},{type:'tool_call',call_id:'c1',name:'read_file',args:{path:'notes.txt'}}]}}});p.observation({agent_id:'a1',kind:'tool_returned',event:{seq:6,type:'tool_returned',tool:{call_id:'c1',parts:[{type:'text',text:'<img src=x onerror="window.executed=true"> '+('large '.repeat(1000))+' END-RETAINED'}]}}});});
  const chat=page.locator('#review-a [data-artifacts]'),actions=page.locator('#review-a [data-actions]');expect((await chat.textContent()).includes('ANSWER')&&!(await actions.textContent()).includes('ANSWER'),'mixed response text routed to wrong pane');expect((await actions.textContent()).includes('read_file'),'mixed call missing from actions');expect(await chat.locator('article').count()===1,'final answer duplicated provisional card');
  await actions.getByRole('button',{name:'Expand full retained text'}).first().click();expect((await actions.textContent()).includes('END-RETAINED'),'expanded result lost full retained bytes');expect(await actions.locator('img,script,iframe').count()===0&&!await page.evaluate(()=>window.executed),'result markup executed');
  await page.evaluate(()=>{window.cards=[...review.a.root.querySelectorAll('article')];return review.remote({font_size:20,theme:'light'})});expect(await page.evaluate(()=>cards.every((c,i)=>c===[...review.a.root.querySelectorAll('article')][i])),'setting update rebuilt mixed-part cards');
 });
 await run('speech-enqueue-revision-rate-off-buffer-and-local-cancel',async page=>{
  const result=await page.evaluate(async()=>{
   const r=review,tick=r.tick,expect=(x,m)=>{if(!x)throw Error(m)};r.a.speech.activate();r.b.speech.activate();r.b.input.value='Independent typing';r.b.input.dispatchEvent(new Event('input'));await tick();await r.remote({autoplay:true,speech_rate:1.25});
   const id={agent_id:'a1',request_id:'r1',operation_id:'m1',part_id:0},delta=(p,text,part=0)=>p.observation({...id,part_id:part,kind:'part_delta',channel:'text',text});
   delta(r.a,'A first. ');await r.until(()=>r.active?.text==='A first. ','positive first native admission absent');delta(r.a,'A queued. ');delta(r.b,'B queued. ');await tick();await r.remote({speech_rate:1.75});delta(r.a,'A later. ');delta(r.a,'BUFFERED',1);delta(r.a,'ONLY BUFFER',2);await tick();
   const before=r.cancels;await r.remote({autoplay:false});delta(r.a,' OFF TEXT.',1);delta(r.a,'OFF ONLY.',3);expect(r.cancels===before&&r.active?.text==='A first. ','shared disable canceled already admitted speech');await r.remote({autoplay:true});for(const [part,text] of [[1,'BUFFERED OFF TEXT.'],[2,'ONLY BUFFER'],[3,'OFF ONLY.']])r.a.observation({...id,part_id:part,kind:'part_final',part:{type:'text',text}});await tick();
   for(let i=0;i<4;i++)await r.end();
   const first=r.spoken.map(u=>({text:u.text,rate:u.rate})),starts=r.speechEvents.filter(e=>e.type==='start');
   expect(JSON.stringify(first)===JSON.stringify([{text:'A first. ',rate:1.25},{text:'B queued. ',rate:1.25},{text:'A queued. ',rate:1.25},{text:'A later. ',rate:1.75}]),'queued rate changed, off buffer replayed, or shared FIFO lost work');
   expect(starts.map(e=>e.revision).join(',')==='1,1,1,2','queued preference revision not captured');
   r.a.speak('manual-A','MANUAL A');await r.until(()=>r.active?.text==='MANUAL A','positive manual native admission absent');r.b.speak('manual-B','MANUAL B');await tick();const stale=r.active;r.a.speech.cancel();await r.until(()=>r.active?.text==='MANUAL B','peer native admission absent after local cancel');expect(r.active?.text==='MANUAL B','local cancel cleared peer queue');stale.onend?.();await tick();expect(r.active?.text==='MANUAL B','stale callback settled peer');r.b.speech.cancel();await tick();expect(r.commands.filter(m=>m.socket===1&&m.type==='pause').at(-1)?.typing===true,'speech change cleared independent typing');return {first,starts};
  });return result;
 });
 await run('old-socket-preferences-fenced-before-generation',async page=>{
  return page.evaluate(async()=>{const r=review,old=r.sockets[0],stale=old.onmessage;await r.remote({theme:'light'});r.a.connector.connect();const fresh=r.sockets.at(-1),applied=r.a.root.dataset.theme;
   const late=()=>stale({data:JSON.stringify({type:'preferences_snapshot',revision:999,preferences:{...r.snapshot().preferences,theme:'dark'}})});late();if(r.a.root.dataset.theme!==applied||r.a.settings.applied.revision===999)throw Error('abandoned socket overwrote preferences before new generation');fresh.onopen();await r.tick();late();if(r.a.root.dataset.theme!==applied||r.a.settings.applied.revision===999)throw Error('abandoned socket overwrote completed new settings');return {applied,revision:r.a.settings.applied.revision};});
 });
 await run('settings-wait-for-complete-initial-domains',async page=>{
  return page.evaluate(async()=>{const r=review;if(r.a.connector.settingsReady||r.b.connector.settingsReady)throw Error('settings ready before Agent snapshot end');const before=r.commands.length;await r.a.settings.change({theme:'light'});if(r.commands.length!==before)throw Error('settings control wrote before both initial domains known');r.sockets[0].finish();await r.tick();if(!r.a.connector.settingsReady||r.b.connector.settingsReady)throw Error('settings readiness leaked between connections');r.sockets[1].finish();await r.tick();await r.a.settings.change({theme:'light'});if(r.a.root.dataset.theme!=='light'||r.b.root.dataset.theme!=='light')throw Error('positive complete handoff did not enable actual controls');return {revision:r.snapshot().revision};});
 },{deferEnd:true});
 await run('unavailable-playback-holds-no-speaking-pause',async page=>{
  return page.evaluate(async()=>{const r=review;await r.remote({autoplay:true});r.app.synthesis=null;r.a.speech.activate();r.a.observation({kind:'part_delta',agent_id:'a1',request_id:'r1',operation_id:'m1',part_id:0,channel:'text',text:'Cannot play this. '});await r.tick();if(r.a.speech.busy()||r.commands.filter(m=>m.socket===0&&m.type==='pause').at(-1)?.speaking)throw Error('unavailable playback retained speaking work');if(!/unavailable/i.test(r.a.notice.textContent))throw Error('unavailable playback lacked visible status');return {notice:r.a.notice.textContent};});
 });
 await run('pending-settings-disposal-and-same-dom-remount',async page=>{
  return page.evaluate(async()=>{
   const r=review;let old=r.a;
   for(const domain of ['preferences','policy']){
    const control=old.root.querySelector(domain==='preferences'?'[data-preference="theme"]':'[data-policy-save]'),input=old.root.querySelector(domain==='preferences'?'[data-preference="theme"]':'[data-policy]'),socket=old.connector.socket,send=socket.send.bind(socket);
    let held=false;socket.send=raw=>{const m=JSON.parse(raw);if(m.type===domain+'_update'){held=true;return;}send(raw);};
    if(domain==='policy')input.value='8';
    const pending=domain==='preferences'?old.settings.change({theme:'light'},control):old.settings.policyChange();if(!held||!control.disabled)throw Error('positive pending settings control did not enter disabled state');
    old.close();const next=r.app.createPage(old.root,'ws://controlled');const draft=domain==='preferences'?'system':'9';input.value=draft;await pending;await r.tick();
    if(input.value!==draft)throw Error('late old settings continuation overwrote replacement draft');
    r.sockets.at(-1).onopen();await r.tick();await r.tick();
    if(control.disabled||next.root.querySelector('[data-policy-save]').disabled)throw Error('remounted settings retained old owner disabled control');
    old=next;
   }
   await old.settings.change({theme:'light'},old.root.querySelector('[data-preference="theme"]'));if(old.root.dataset.theme!=='light')throw Error('replacement settings control did not recover');return {theme:old.root.dataset.theme};
  });
 });
 await browser.close();await new Promise(r=>server.close(r));const hashes={};for(const name of fs.readdirSync(assets).sort()){const file=path.join(assets,name);if(fs.statSync(file).isFile())hashes[name]=crypto.createHash('sha256').update(fs.readFileSync(file)).digest('hex');}
 fs.rmSync(copy,{recursive:true,force:true});const passed=checks.length===(only?1:8)&&checks.every(r=>r.passed);console.log(JSON.stringify({scope:'Actual Chrome HTML/CSS/public browser components, controlled socket/native API; no audio/live-model claim.',selected_case:only||null,passed,checks,source,assets:hashes,checkers:Object.fromEntries(['accept_ch08_browser.cjs','ch08-browser-fixture.js'].map(name=>[name,crypto.createHash('sha256').update(fs.readFileSync(path.join(__dirname,name))).digest('hex')]))},null,2));process.exitCode=passed?0:1;
})().catch(e=>{console.error(e.stack);server.close();fs.rmSync(copy,{recursive:true,force:true});process.exitCode=1;});
