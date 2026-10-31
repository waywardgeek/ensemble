#!/usr/bin/env node
// Comparative review: real DOM, controlled socket/synthesis, no provider calls.
const fs=require('fs'),path=require('path'),http=require('http'),crypto=require('crypto');
const {chromium}=require(path.join(process.env.CH07_BROWSER_TOOLS||'/Users/bill/projects/ensemble-edition-2-revisions/browser-tools','node_modules/playwright'));
const source=path.resolve(process.argv[2]),assets=path.join(source,'gui/web/gui');
const html='<!doctype html><main><p data-status></p><p data-pause></p><section data-artifacts></section><textarea data-input></textarea><button data-prompt>Send</button><button data-hint>Hint</button><button data-interrupt>Interrupt</button><button data-latest>Latest</button><button data-auto-speech>Auto</button><button data-cancel-speech>Cancel</button><p data-notice></p></main>';
const server=http.createServer((req,res)=>{if(req.url==='/'){res.setHeader('Content-Type','text/html');res.end(html);return;}const p=path.join(assets,req.url);if(!p.startsWith(assets+path.sep)||!fs.existsSync(p)){res.writeHead(404);res.end();return;}res.setHeader('Content-Type','text/javascript');res.end(fs.readFileSync(p));});
(async()=>{
 await new Promise(r=>server.listen(0,'127.0.0.1',r));const origin='http://127.0.0.1:'+server.address().port;
 const browser=await chromium.launch({channel:'chrome',headless:true}),page=await browser.newPage();await page.route('**/*',r=>r.request().url().startsWith(origin+'/')?r.continue():r.abort());await page.goto(origin);
 const result=await page.evaluate(async()=>{
  const sockets=[],commands=[],tick=()=>new Promise(r=>setTimeout(r,0));
  class Socket{static OPEN=1;constructor(){this.readyState=1;this.ordinal=sockets.length;sockets.push(this);}frame(m){this.onmessage?.({data:JSON.stringify(m)});}send(raw){const m=JSON.parse(raw);commands.push({...m,socket:this.ordinal});queueMicrotask(()=>{if(m.type==='subscribe'){this.frame({type:'snapshot_begin',id:m.id,agent_id:'a',generation:'g'+this.ordinal,watermark:0,omitted:0,state:{paused:false,typing_clients:0,speaking_clients:0}});this.frame({type:'snapshot_end',generation:'g'+this.ordinal,watermark:0});}if(m.type==='pause')this.frame({type:'ack',id:m.id,paused:m.typing||m.speaking,typing_clients:+m.typing,speaking_clients:+m.speaking});if(m.type==='prompt')this.frame({type:'accepted',id:m.id,request_id:'r'+this.ordinal});});}close(){if(this.readyState!==1)return;this.readyState=3;this.onclose?.();}}
  window.WebSocket=Socket;Object.defineProperty(window,'speechSynthesis',{configurable:true,value:{speak(){},cancel(){}}});window.SpeechSynthesisUtterance=class{};
  const {Page}=await import('/page.js'),root=document.querySelector('main');
  const tracked=new Map(), add=EventTarget.prototype.addEventListener, remove=EventTarget.prototype.removeEventListener; EventTarget.prototype.addEventListener=function(type,fn,options){if(this===root||root.contains(this)){let rows=tracked.get(this);if(!rows)tracked.set(this,rows=[]);if(!rows.some(r=>r.type===type&&r.fn===fn))rows.push({type,fn});}return add.call(this,type,fn,options);}; EventTarget.prototype.removeEventListener=function(type,fn,options){const rows=tracked.get(this);if(rows)tracked.set(this,rows.filter(r=>r.type!==type||r.fn!==fn));return remove.call(this,type,fn,options);};
  const app=await fetch('/application.js').then(r=>r.ok)?new (await import('/application.js')).BrowserApplication():null; const create=()=>app?app.createPage(root,'ws://fixture'):new Page(root,'ws://fixture');
  const first=create();sockets.at(-1).onopen();await tick();first.input.value='POSITIVE';root.querySelector('[data-prompt]').click();await tick();
  const positive=commands.filter(x=>x.type==='prompt').map(x=>x.text);
  first.close();await tick();const listenersAfterClose=[...tracked.values()].reduce((n,rows)=>n+rows.length,0);const second=create();sockets.at(-1).onopen();await tick();second.input.value='REPLACEMENT';root.querySelector('[data-prompt]').click();await tick();
  const replacement=commands.filter(x=>x.type==='prompt'&&x.socket===1).map(x=>x.text),notice=second.notice.textContent;second.close();
  return {positive,replacement,notice,listenersAfterClose};
 });
 const passed=JSON.stringify(result.positive)==='["POSITIVE"]'&&JSON.stringify(result.replacement)==='["REPLACEMENT"]'&&result.listenersAfterClose===0;
 await browser.close();await new Promise(r=>server.close(r));const hashes={};for(const name of fs.readdirSync(assets)){const p=path.join(assets,name);if(fs.statSync(p).isFile())hashes[name]=crypto.createHash('sha256').update(fs.readFileSync(p)).digest('hex');}
 console.log(JSON.stringify({scope:'Public Page close/recreate on retained DOM. Controlled socket/speech, no model or real-audio claim.',source,assets:hashes,checker_sha256:crypto.createHash('sha256').update(fs.readFileSync(__filename)).digest('hex'),passed,result},null,2));process.exitCode=passed?0:1;
})().catch(e=>{console.error(e);server.close();process.exitCode=1;});
