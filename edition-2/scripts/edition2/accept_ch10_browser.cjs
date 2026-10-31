#!/usr/bin/env node
// Actual Chrome/public components with controlled wire/native callbacks.
// No GUI-server persistence, valid exhausted history, provider or audio claim.
const fs=require('fs'),path=require('path'),os=require('os'),http=require('http'),crypto=require('crypto');
const {chromium}=require(path.join(process.env.CH07_BROWSER_TOOLS||'/Users/bill/projects/ensemble-edition-2-revisions/browser-tools','node_modules/playwright'));
const source=path.resolve(process.argv[2]),temporary=fs.mkdtempSync(path.join(os.tmpdir(),'ch10-browser-')),assets=path.join(temporary,'assets');
const hash=p=>crypto.createHash('sha256').update(fs.readFileSync(p)).digest('hex');
function files(root,prefix=''){return Object.fromEntries(fs.readdirSync(root).sort().flatMap(name=>{const p=path.join(root,name),relative=path.join(prefix,name);return fs.statSync(p).isDirectory()?Object.entries(files(p,relative)):[[relative,hash(p)]]}));}
const before=files(path.join(source,'gui/web/gui'));fs.cpSync(path.join(source,'gui/web/gui'),assets,{recursive:true});
const html=fs.readFileSync(path.join(assets,'index.html'),'utf8').replace(/<script\b[^>]*>[\s\S]*?<\/script>/gi,'');
const helpers={'/__review.js':'ch10-browser-fixture.js','/__base.js':'ch08-browser-fixture.js'};
const checkerFiles=['accept_ch10_browser.cjs',...Object.values(helpers)],checkerBefore=Object.fromEntries(checkerFiles.map(n=>[n,hash(path.join(__dirname,n))]));
const server=http.createServer((req,res)=>{
 if(req.url==='/'){res.setHeader('Content-Type','text/html');res.end(html);return;}
 const helper=helpers[req.url],file=helper?path.join(__dirname,helper):path.resolve(assets,'.'+decodeURIComponent(req.url.split('?')[0]));
 if((!helper&&!file.startsWith(assets+path.sep))||!fs.existsSync(file)){res.writeHead(404);res.end();return;}
 res.setHeader('Content-Type',file.endsWith('.css')?'text/css':'text/javascript');res.end(fs.readFileSync(file));
});
const checks=[],expect=(value,message)=>{if(!value)throw Error(message)};
(async()=>{
 await new Promise(resolve=>server.listen(0,'127.0.0.1',resolve));
 const origin='http://127.0.0.1:'+server.address().port,browser=await chromium.launch({channel:'chrome',headless:true});
 async function run(id,action){
  const page=await browser.newPage({viewport:{width:1500,height:1000}}),errors=[];page.on('pageerror',e=>errors.push(e.message));page.setDefaultTimeout(2500);
  await page.route('**/*',r=>r.request().url().startsWith(origin+'/')?r.continue():r.abort());
  try{await page.goto(origin);await page.evaluate(async()=>window.review=await(await import('/__review.js')).setup());const result=await action(page);expect(!errors.length,'uncaught errors: '+errors.join('; '));checks.push({id,passed:true,result});}
  catch(error){const diagnostic=await page.evaluate(()=>({label:document.querySelector('#review-a [data-session]')?.textContent,notice:document.querySelector('#review-a [data-notice]')?.textContent,cards:[...document.querySelectorAll('#review-a article')].map(c=>({key:c.dataset.key,text:c.textContent.slice(0,200)})),commands:window.review?.commands})).catch(e=>({error:String(e)}));checks.push({id,passed:false,error:error.stack,page_errors:errors,diagnostic});}
  finally{await page.close();}
 }
 await run('standalone-and-new-session-checkpoint-affordance',async page=>{
  await page.evaluate(()=>review.snapshot(null));expect(await page.locator('#review-a [data-checkpoint]').isDisabled(),'standalone checkpoint must be unavailable');expect(/standalone/i.test(await page.locator('#review-a [data-session]').textContent()),'standalone state mislabeled');
  await page.evaluate(()=>review.snapshot(review.session(null,false)));expect(await page.locator('#review-a [data-checkpoint]').isEnabled(),'new session checkpoint inaccessible');const label=await page.locator('#review-a [data-session]').textContent();expect(label.includes('a'.repeat(32))&&/new/i.test(label),'new session durable identity/provenance absent');return {label};
 });
 await run('checkpoint-control-applied-anchor-and-correlated-refusal',async page=>{
  await page.evaluate(()=>review.snapshot(review.session(7)));await page.locator('#review-a [data-checkpoint]').click();await page.waitForFunction(()=>review.pending.length===1);
  expect(await page.evaluate(()=>{const v=review.pending[0].value;return Object.keys(v).sort().join(',')==='id,type'&&v.type==='checkpoint'&&typeof v.id==='string'&&v.id.length>0}),'Checkpoint did not issue exact correlated command');
  expect((await page.locator('#review-a [data-session]').textContent()).includes('7'),'click fabricated a later committed anchor');
  await page.evaluate(()=>{review.change(13);review.ack(0,13)});await page.waitForFunction(()=>/saved/i.test(review.a.notice.textContent)&&review.a.notice.textContent.includes('13'));expect((await page.locator('#review-a [data-session]').textContent()).includes('13'),'applied session change did not update label');
  await page.locator('#review-a [data-checkpoint]').click();await page.waitForFunction(()=>review.pending.length===2);await page.evaluate(()=>review.refuse(1));await page.waitForFunction(()=>/busy/i.test(review.a.notice.textContent));expect((await page.locator('#review-a [data-session]').textContent()).includes('13'),'refusal altered committed anchor');return {commands:await page.evaluate(()=>review.pending.length)};
 });
 await run('historical-status-and-live-owner-are-separate',async page=>{
  await page.evaluate(()=>review.snapshot(review.session(25),['running','done','killed'].map((s,i)=>review.job(i+1,s,20+i)),[1,2,3].map(handle=>({handle,live:false}))));
  const cards=page.locator('#review-a [data-actions] article');expect(await cards.count()===3,'historical reports lost or bookkeeping became cards');
  for(let i=0;i<3;i++){const card=cards.filter({hasText:'JOB_'+(i+1)}),text=await card.textContent();expect(text.includes(['running','done','killed'][i])&&/historical; no live owner/i.test(text),'historical availability replaced/omitted actual status');expect(await card.getByRole('button',{name:/wait|send input|kill/i}).evaluateAll(nodes=>nodes.every(n=>n.disabled||n.hidden||n.getClientRects().length===0)),'historical supervision control enabled');}
  await page.evaluate(()=>review.snapshot(review.session(25),[review.job(1,'running')],[{handle:1,live:true}]));expect(!/historical; no live owner/i.test(await cards.textContent()),'live positive incorrectly marked historical');return {historical:3,live_positive:1};
 });
 await run('restart-replaces-transient-view-with-silent-retained-history',async page=>{
  await page.evaluate(async()=>{await review.remote({autoplay:true});review.a.speech.activate();review.snapshot(review.session(7));review.observe({kind:'part_delta',agent_id:'session-agent',request_id:'old-request',operation_id:'old-operation',part_id:'old-part',channel:'text',text:'OLD_PROVISIONAL'})});
  expect((await page.locator('#review-a [data-artifacts]').textContent()).includes('OLD_PROVISIONAL'),'provisional positive absent');
  await page.evaluate(()=>{review.snapshot(review.session(20),[{seq:15,type:'response_ended',response:{parts:[{type:'text',text:'RETAINED_ACCEPTED'}]}}],[],{generation:'after-process-restart',agent_id:'fresh-runtime',omitted:104});});
  const text=await page.locator('#review-a [data-artifacts]').textContent();expect(text.includes('RETAINED_ACCEPTED')&&!text.includes('OLD_PROVISIONAL')&&text.includes('104'),'restart lost retained window or kept dead provisional state');
  expect(await page.evaluate(()=>!review.a.speech.busy()&&!review.commands.some(c=>['prompt','hint'].includes(c.type))),'restart restored speech or resent input');
  const count=await page.evaluate(()=>review.spoken.length);await page.waitForTimeout(50);expect(await page.evaluate(()=>review.spoken.length)===count,'replayed history auto-spoke');return {retained:1,omitted:104};
 });
 await run('session-anchor-and-historical-handle-full-uint64',async page=>{
  await page.evaluate(()=>review.snapshot(review.session(18446744073709551615n),[review.job(9007199254740992n,'running',20),review.job(9007199254740993n,'done',21)],[{handle:9007199254740992n,live:false},{handle:9007199254740993n,live:false}]));
  expect((await page.locator('#review-a [data-session]').textContent()).includes('18446744073709551615'),'checkpoint sequence rounded');
  const decoded=await page.evaluate(()=>review.snapshots.at(-1).state.job_access.map(j=>String(j.handle)));expect(decoded.join(',')==='9007199254740992,9007199254740993','adjacent unsafe historical handles rounded');
  const cards=page.locator('#review-a [data-actions] article');expect(await cards.count()===2&&/historical; no live owner/i.test(await cards.nth(0).textContent())&&/historical; no live owner/i.test(await cards.nth(1).textContent()),'exact handle projection lost card ownership');return {handles:decoded};
 });
 await run('checkpoint-ack-anchor-full-uint64',async page=>{
  await page.evaluate(()=>review.snapshot(review.session(9007199254740992n)));await page.locator('#review-a [data-checkpoint]').click();await page.waitForFunction(()=>review.pending.length===1);
  await page.evaluate(()=>{review.change(9007199254740993n);review.ack(0,9007199254740993n)});await page.waitForFunction(()=>/saved/i.test(review.a.notice.textContent));const notice=await page.locator('#review-a [data-notice]').textContent();expect(notice.includes('9007199254740993')&&!notice.includes('9007199254740992'),'acknowledged anchor rounded');return {notice};
 });
 await browser.close();await new Promise(resolve=>server.close(resolve));expect(JSON.stringify(before)===JSON.stringify(files(path.join(source,'gui/web/gui'))),'source assets changed');expect(JSON.stringify(checkerBefore)===JSON.stringify(Object.fromEntries(checkerFiles.map(n=>[n,hash(path.join(__dirname,n))]))),'checker changed');
 const passed=checks.length===6&&checks.every(x=>x.passed);console.log(JSON.stringify({passed,scope:'Actual Chrome/public components; controlled projection/socket/native fixtures, not valid exhausted logs or GUI persistence/live/audio evidence.',checks,source,assets:before,checkers:checkerBefore},null,2));fs.rmSync(temporary,{recursive:true,force:true});process.exitCode=passed?0:1;
})().catch(error=>{console.error(error.stack);server.close();fs.rmSync(temporary,{recursive:true,force:true});process.exitCode=1});
