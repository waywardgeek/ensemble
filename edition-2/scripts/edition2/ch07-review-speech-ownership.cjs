#!/usr/bin/env node
// Actual Page/queue/application components in Chrome; controlled native API and
// socket acknowledgements. This establishes ownership, not audible synthesis.
const fs=require('fs'),path=require('path'),http=require('http'),crypto=require('crypto');
const {chromium}=require(path.join(process.env.CH07_BROWSER_TOOLS||'/Users/bill/projects/ensemble-edition-2-revisions/browser-tools','node_modules/playwright'));
const source=path.resolve(process.argv[2]),assets=path.join(source,'gui/web/gui');
const template='<p data-status></p><p data-pause></p><section data-artifacts></section><textarea data-input></textarea><button data-prompt>Send</button><button data-hint>Hint</button><button data-interrupt>Interrupt</button><button data-latest>Latest</button><button data-auto-speech>Auto</button><button data-cancel-speech>Cancel</button><p data-notice></p>';
const server=http.createServer((req,res)=>{if(req.url==='/'){res.setHeader('Content-Type','text/html');res.end('<!doctype html><main id="a">'+template+'</main><main id="b">'+template+'</main>');return;}const p=path.join(assets,req.url);if(!p.startsWith(assets+path.sep)||!fs.existsSync(p)){res.writeHead(404);res.end();return;}res.setHeader('Content-Type','text/javascript');res.end(fs.readFileSync(p));});
(async()=>{
 await new Promise(r=>server.listen(0,'127.0.0.1',r));const origin='http://127.0.0.1:'+server.address().port;
 const browser=await chromium.launch({channel:'chrome',headless:true}),checks=[];
 for(const scenario of ['idle-page-reset','queued-page-cancel','active-page-cancel','fifo-ready-heads','closed-owner-and-stale-callbacks','whole-application-close']){
  const page=await browser.newPage(),errors=[];page.on('pageerror',e=>errors.push(e.message));await page.route('**/*',r=>r.request().url().startsWith(origin+'/')?r.continue():r.abort());await page.goto(origin);
  try{
   const result=await page.evaluate(async scenario=>{
    const sockets=[],spoken=[],commands=[],tick=()=>new Promise(r=>setTimeout(r,0));let active=null,cancels=0,overlap=false;
    class WS{static OPEN=1;constructor(){this.readyState=1;this.ordinal=sockets.length;sockets.push(this);}frame(m){this.onmessage?.({data:JSON.stringify(m)});}send(raw){const m=JSON.parse(raw);commands.push({...m,socket:this.ordinal});queueMicrotask(()=>{if(m.type==='subscribe'){this.frame({type:'snapshot_begin',id:m.id,agent_id:'a'+this.ordinal,generation:'g'+this.ordinal,watermark:0,omitted:0,state:{paused:false,typing_clients:0,speaking_clients:0}});this.frame({type:'snapshot_end',generation:'g'+this.ordinal,watermark:0});}if(m.type==='pause')this.frame({type:'ack',id:m.id,paused:m.typing||m.speaking,typing_clients:+m.typing,speaking_clients:+m.speaking});});}close(){if(this.readyState!==1)return;this.readyState=3;this.onclose?.();}}
    const synthesis={speak(u){if(active)overlap=true;active=u;spoken.push(u);u.onstart?.();},cancel(){cancels++;const old=active;active=null;old?.onerror?.({error:'interrupted'});}};
    class Utterance{constructor(text){this.text=text;}}
    window.WebSocket=WS;Object.defineProperty(window,'speechSynthesis',{configurable:true,value:synthesis});window.SpeechSynthesisUtterance=Utterance;
    const {Page}=await import('/page.js'),app=await fetch('/application.js').then(r=>r.ok)?new (await import('/application.js')).BrowserApplication(synthesis,Utterance):null;
    const create=async id=>{const p=app?app.createPage(document.getElementById(id),'ws://fixture'):new Page(document.getElementById(id),'ws://fixture');sockets.at(-1).onopen();await tick();return p;};
    const expect=(condition,message)=>{if(!condition)throw Error(message);};
    const a=await create('a');a.input.value='typing remains independent';a.input.dispatchEvent(new Event('input'));await tick();
    a.speech.enqueue('A1','A1');await tick();expect(active?.text==='A1','positive first Page did not reach native speech');
    let b,c;
    if(scenario==='idle-page-reset'){
     const prior=cancels;b=await create('b');b.speech.reset();b.close();await tick();expect(cancels===prior&&active?.text==='A1','idle Page canceled another Page native utterance');
    }else{
     // Starting the second Page before speech is the ordinary positive setup;
     // the isolated idle-reset case above protects construction during speech.
     const held=active;active=null;held.onend();await tick();b=await create('b');a.speech.enqueue('A2','A2');await tick();
     b.speech.enqueue('B1','B1');await tick();expect(!overlap&&active?.text==='A2'&&!spoken.some(u=>u.text==='B1'),'second Page bypassed shared one-utterance admission');
     if(scenario==='queued-page-cancel'){
      const prior=cancels;b.speech.cancel();await tick();expect(cancels===prior&&active?.text==='A2','pending Page cancellation canceled active peer');const u=active;active=null;u.onend();await tick();expect(!active&&!b.speech.busy(),'canceled pending Page later spoke');
     }else if(scenario==='active-page-cancel'){
      const stale=active,prior=cancels;a.speech.cancel();await tick();expect(cancels===prior+1&&active?.text==='B1','active owner cancellation did not advance waiting peer');stale.onend();stale.onerror({error:'late'});await tick();expect(active?.text==='B1'&&b.speech.busy(),'stale canceled callback settled peer');
     }else if(scenario==='whole-application-close'){
      const before=spoken.length;const stale=active;expect(app,'shared application owner is required for whole application shutdown');app.close();await tick();stale.onend();stale.onerror({error:'late'});await tick();expect(spoken.length===before&&!active&&!a.speech.busy()&&!b.speech.busy(),'whole application shutdown admitted pending native speech');
     }else if(scenario==='fifo-ready-heads'){
      const root=document.getElementById('b').cloneNode(true);root.id='c';document.body.append(root);c=await create('c');c.speech.enqueue('C1','C1');await tick();a.speech.enqueue('A3','A3');await tick();const first=active;active=null;first.onend();await tick();expect(active?.text==='B1','ready peer lost FIFO place to same Page successor');const second=active;active=null;second.onend();await tick();expect(active?.text==='C1','third Page lost its FIFO place');const third=active;active=null;third.onend();await tick();expect(active?.text==='A3','same Page successor disappeared');
     }else{
      const stale=active;a.close();await tick();expect(active?.text==='B1','closing active Page did not advance peer');const before=spoken.length;stale.onend();stale.onerror({error:'late'});a.speech.enqueue('closed','CLOSED');await tick();expect(spoken.length===before&&active?.text==='B1','closed Page or stale callback revived speech');const last=active;active=null;last.onend();await tick();expect(!active&&!spoken.some(u=>u.text==='CLOSED'),'closed Page retained native work behind peer');
     }
    }
    const typing=commands.filter(m=>m.socket===0&&m.type==='pause').at(-1);if(scenario!=='closed-owner-and-stale-callbacks')expect(typing?.typing,'speech transition cleared independent typing');
    const result={spoken:spoken.map(u=>u.text),cancels,overlap,lastTyping:typing?.typing};a.close();b?.close();c?.close();app?.close();return result;
   },scenario);
   if(errors.length)throw Error(errors.join('; '));checks.push({id:scenario,passed:true,result});
  }catch(e){checks.push({id:scenario,passed:false,error:e.message,page_errors:errors});}finally{await page.close();}
 }
 const hashes={};for(const name of fs.readdirSync(assets)){const p=path.join(assets,name);if(fs.statSync(p).isFile())hashes[name]=crypto.createHash('sha256').update(fs.readFileSync(p)).digest('hex');}
 await browser.close();await new Promise(r=>server.close(r));const passed=checks.every(r=>r.passed);console.log(JSON.stringify({scope:'Real Chrome Pages and shared document speech ownership; controlled native speech/socket, no audio/provider claim.',source,assets:hashes,checker_sha256:crypto.createHash('sha256').update(fs.readFileSync(__filename)).digest('hex'),passed,checks},null,2));process.exitCode=passed?0:1;
})().catch(e=>{console.error(e.stack);server.close();process.exitCode=1;});
