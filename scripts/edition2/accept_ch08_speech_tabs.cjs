#!/usr/bin/env node
// Actual Chrome tabs and Web Locks; native speech callbacks are controlled.
// This proves cooperative lease behavior, not audible or cross-profile isolation.
const fs=require('fs'),path=require('path'),http=require('http'),os=require('os'),crypto=require('crypto');
const {chromium}=require(path.join(process.env.CH07_BROWSER_TOOLS||'/Users/bill/projects/ensemble-edition-2-revisions/browser-tools','node_modules/playwright'));
const source=path.resolve(process.argv[2]),copy=fs.mkdtempSync(path.join(os.tmpdir(),'ch08-native-tabs-')),assets=path.join(copy,'assets');
fs.cpSync(path.join(source,'gui/web/gui'),assets,{recursive:true});
const html=fs.readFileSync(path.join(assets,'index.html'),'utf8').replace(/<script\b[^>]*>[\s\S]*?<\/script>/gi,'');
const server=http.createServer((req,res)=>{
 if(req.url==='/'){res.setHeader('Content-Type','text/html');res.end(html);return;}
 const file=req.url==='/__review.js'?path.join(__dirname,'ch08-browser-fixture.js'):path.join(assets,decodeURIComponent(req.url.split('?')[0]));
 if((!file.startsWith(assets+path.sep)&&req.url!=='/__review.js')||!fs.existsSync(file)){res.writeHead(404);res.end();return;}
 res.setHeader('Content-Type',file.endsWith('.css')?'text/css':'text/javascript');res.end(fs.readFileSync(file));
});
const checks=[],assert=(x,m)=>{if(!x)throw Error(m)};
(async()=>{
 await new Promise(r=>server.listen(0,'127.0.0.1',r));const origin='http://127.0.0.1:'+server.address().port;
 const browser=await chromium.launch({channel:'chrome',headless:true});
 async function run(id,fn){const context=await browser.newContext();const pages=[],events=[],errors=[];let active=null;
  await context.exposeBinding('__nativeEvent',(source,event)=>{const tab=pages.indexOf(source.page);events.push({tab,...event});if(event.type==='speak'){if(active!==null)errors.push('native overlap '+active+' -> '+tab);active=tab;}else if(event.type==='cancel'){if(active!==tab)errors.push('nonholder native cancel '+tab+' while '+active);active=null;}else if(event.type==='end'&&active===tab)active=null;});
  async function open(){const p=await context.newPage();p.setDefaultTimeout(6500);pages.push(p);p.on('pageerror',e=>errors.push(e.message));await p.goto(origin);await p.evaluate(async()=>{
   window.review=await(await import('/__review.js')).setup();
   const native=review.app.synthesis,originalSpeak=native.speak,originalCancel=native.cancel;
   native.speak=u=>{window.__nativeEvent({type:'speak',text:u.text});originalSpeak(u);};
   native.cancel=()=>{window.__nativeEvent({type:'cancel'});originalCancel();};
   window.enqueue=text=>review.a.speech.enqueue(text,text);
  });return p;}
  try{await fn(open,events);assert(!errors.length,errors.join('; '));checks.push({id,passed:true,events});}
  catch(e){checks.push({id,passed:false,error:e.stack,events,errors});}
  finally{await context.close();}
 }
 await run('same-context-held-lease-waiter-cancel-owner-close',async(open,events)=>{
  const a=await open(),b=await open();await a.evaluate(()=>enqueue('A held.'));await a.waitForFunction(()=>review.spoken.length===1);
  await b.evaluate(()=>enqueue('B waiting.'));await b.waitForFunction(()=>review.a.speech.busy());
  await b.waitForTimeout(5300);
  assert(await b.evaluate(()=>review.spoken.length===0&&review.cancels===0&&review.a.speech.busy()),'waiter used native service or timed out before lease');
  await b.evaluate(()=>review.a.speech.cancel());assert(await b.evaluate(()=>!review.a.speech.busy()&&review.cancels===0),'canceled waiter invoked native cancel');
  await b.evaluate(()=>enqueue('B next.'));await b.waitForFunction(()=>review.a.speech.busy());
  await a.evaluate(()=>{window.oldNative=review.spoken[0];review.app.close();});await b.waitForFunction(()=>review.spoken.length===1);
  await a.evaluate(()=>{oldNative.onend();oldNative.onerror({error:'late'});});
  assert(await b.evaluate(()=>review.active?.text==='B next.'&&review.a.speech.busy()),'late old-owner callbacks settled next owner');
  await b.evaluate(()=>{window.__nativeEvent({type:'end'});return review.end();});
  await b.waitForFunction(()=>!review.a.speech.busy());
  assert(events.filter(x=>x.type==='speak').map(x=>x.text).join('|')==='A held.|B next.','canceled waiting utterance later started');
 });
 await run('waiting-application-close-does-not-start-native-work',async(open,events)=>{
  const a=await open(),b=await open();await a.evaluate(()=>enqueue('A holds.'));await a.waitForFunction(()=>review.spoken.length===1);
  await b.evaluate(()=>{enqueue('B abandoned.');review.app.close();});
  await a.evaluate(()=>{window.__nativeEvent({type:'end'});return review.end();});
  await a.waitForFunction(()=>!review.a.speech.busy());
  // Query the actual shared browser lock manager after a new exclusive barrier.
  await a.evaluate(()=>navigator.locks.request('ensemble-native-speech',()=>{}));
  assert(await b.evaluate(()=>review.spoken.length===0&&review.cancels===0&&!review.a.speech.busy()),'closed waiter acquired or canceled native work');
 });
 await run('coordination-unavailable-releases-local-pause',async(open)=>{
  const a=await open();await a.evaluate(()=>{review.app.locks=null;enqueue('Unavailable.');});
  await a.waitForFunction(()=>!review.a.speech.busy());
  assert(await a.evaluate(()=>review.spoken.length===0&&review.cancels===0&&!review.sockets[0].causes.speaking),'unsupported coordination retained speech or speaking pause');
  assert(await a.evaluate(()=>/unavailable/i.test(review.a.notice.textContent)),'unsupported coordination has no visible status');
 });
 await browser.close();await new Promise(r=>server.close(r));
 const assetsHash={};for(const name of fs.readdirSync(assets).sort()){const p=path.join(assets,name);if(fs.statSync(p).isFile())assetsHash[name]=crypto.createHash('sha256').update(fs.readFileSync(p)).digest('hex');}
 const result={scope:'Actual same-context Chrome tabs and browser-native Web Locks, controlled native speech. No audio or unrelated-origin/profile claim.',passed:checks.length===3&&checks.every(x=>x.passed),checks,browser:browser.version(),asset_sha256:assetsHash,checker_sha256:crypto.createHash('sha256').update(fs.readFileSync(__filename)).digest('hex')};
 fs.rmSync(copy,{recursive:true,force:true});console.log(JSON.stringify(result,null,2));process.exitCode=result.passed?0:1;
})().catch(e=>{console.error(e.stack);server.close();process.exitCode=1;});
