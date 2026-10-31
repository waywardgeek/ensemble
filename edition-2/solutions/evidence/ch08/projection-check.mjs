// Exercise the production browser against a small external WebSocket peer.
// Speech instrumentation observes and calls the native service; it never supplies callbacks.
import {chromium} from 'playwright';
import {WebSocketServer} from 'ws';
import http from 'node:http';
import fs from 'node:fs';
import path from 'node:path';
const base='/Users/bill/projects/ensemble/solutions/edition-2/';
const web=process.env.CH08_WEB||base+'main/gui/web/gui';
const outputFile=process.env.CH08_EVIDENCE||base+'evidence/ch08/projection.json';
const retained=JSON.parse(fs.readFileSync(base+'evidence/ch08/gemini-live.json'));
const empty=retained.wire.find(m=>m.type==='part_final'&&m.part_id===2);
if(empty?.part.type!=='text'||empty.part.text)throw Error('retained empty Gemini artifact missing');
const cases=[
 {name:'retained Gemini empty text',part:empty.part,text:''},
 {name:'redacted thinking',part:{type:'opaque',data:{type:'redacted_thinking',data:'OPAQUE_REPLAY_BYTES'}},text:''},
 {name:'signature only',part:{type:'opaque',data:{thoughtSignature:'SIGNATURE_ONLY'}},text:''},
 {name:'Anthropic readable thinking',part:{type:'opaque',data:{type:'thinking',thinking:'A readable thought.',signature:'PRIVATE_SIGNATURE'}},text:'A readable thought.'},
 {name:'Gemini readable thinking',part:{type:'opaque',data:{thought:true,text:'Another readable thought.',thoughtSignature:'PRIVATE_SIGNATURE'}},text:'Another readable thought.'}
];
// Run audible positive controls before deliberately broken values can provoke
// native cancellation. Keep all display and speech-input failures in the result.
cases.unshift(...cases.splice(3));
let index=0;
const server=http.createServer((req,res)=>{
 const file=path.join(web,req.url==='/'?'index.html':req.url);
 try{res.setHeader('Content-Type',file.endsWith('.js')?'text/javascript':file.endsWith('.css')?'text/css':'text/html');res.end(fs.readFileSync(file))}catch{res.writeHead(404).end()}
});
const wss=new WebSocketServer({server,path:'/ws'});
wss.on('connection',ws=>ws.on('message',bytes=>{
 const m=JSON.parse(bytes);
 if(m.type==='subscribe')ws.send(JSON.stringify({type:'state_changed',from:'idle',to:'idle',replay:true}));
 if(m.type==='prompt'){
  ws.send(JSON.stringify({type:'part_final',part_id:++index,part:cases[index-1].part}));
  ws.send(JSON.stringify({type:'turn_ended'}));
 }
}));
await new Promise(r=>server.listen(0,'127.0.0.1',r));
const browser=await chromium.launch({executablePath:'/Applications/Google Chrome.app/Contents/MacOS/Google Chrome',headless:false});
const output={result:'FAIL',source:'retained gemini-live.json artifact 2 plus opaque peer fixtures',checks:[]};
try{
 const page=await browser.newPage();
 await page.addInitScript(()=>{
  globalThis.events=[];
  const native=speechSynthesis.speak.bind(speechSynthesis);
  speechSynthesis.speak=u=>{
   events.push({type:'queued',text:u.text});
   for(const kind of ['start','end','error'])u.addEventListener(kind,e=>events.push({type:kind,text:u.text,error:e.error}));
   return native(u);
  };
 });
 await page.goto('http://127.0.0.1:'+server.address().port);
 await page.bringToFront();
 await page.waitForFunction(()=>document.querySelector('#state').textContent==='idle');
 for(let i=0;i<cases.length;i++){
  const item=cases[i];
  await page.evaluate(()=>events.length=0);
  await page.locator('#input').fill(item.name);
  await page.locator('#input').press('Enter');
  const card=page.locator(`article[data-id="${i+1}"]`);
  await card.waitFor();
  const display=(await card.locator('.content p').allTextContents()).join('\n').trim();
  if(item.text)await page.waitForFunction(()=>events.some(e=>e.type==='end'),null,{timeout:15000});
  const automatic=await page.evaluate(()=>events.filter(e=>e.text));
  await page.keyboard.press('Escape');
  await page.evaluate(()=>events.length=0);
  await card.locator('button').click();
  if(item.text)await page.waitForFunction(()=>events.some(e=>e.type==='end'),null,{timeout:15000});
  const speaker=await page.evaluate(()=>events.filter(e=>e.text));
  await page.keyboard.press('Escape');
  const queued=events=>events.filter(e=>e.type==='queued').map(e=>e.text.trim());
  const expected=item.text?[item.text]:[];
  output.checks.push({name:item.name,display,displayPass:display===item.text,automatic,speaker,
   autoPass:JSON.stringify(queued(automatic))===JSON.stringify(expected),
   speakerPass:JSON.stringify(queued(speaker))===JSON.stringify(expected)});
 }
 output.result=output.checks.every(c=>c.displayPass&&c.autoPass&&c.speakerPass)?'PASS':'FAIL';
}catch(e){output.error=e.message.split('\n')[0]}
finally{
 fs.writeFileSync(outputFile,JSON.stringify(output,null,2));
 console.log(JSON.stringify(output));
 await browser.close();for(const c of wss.clients)c.terminate();wss.close();server.close();
 if(output.result!=='PASS')process.exitCode=1;
}
