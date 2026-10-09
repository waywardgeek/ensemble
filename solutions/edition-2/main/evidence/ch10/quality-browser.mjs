// Local captured-data component regression, never a new provider/live receipt.
import assert from 'node:assert/strict';
import {createHash} from 'node:crypto';
import {execFileSync} from 'node:child_process';
import {readFileSync, writeFileSync, mkdirSync} from 'node:fs';
import {createServer} from 'node:http';
import {join, resolve, dirname, relative} from 'node:path';
import {fileURLToPath, pathToFileURL} from 'node:url';

const here=dirname(fileURLToPath(import.meta.url));
const [repoArg, revision, outputArg]=process.argv.slice(2), repo=resolve(repoArg), output=resolve(outputArg);
assert.match(revision,/^[0-9a-f]{40}$/);
const originalRevision='44d762789ea65a1e11fd7982ec9b748d4a8112c1';
const sha=bytes=>createHash('sha256').update(bytes).digest('hex');
const git=(rev,path)=>execFileSync('git',['-C',repo,'show',`${rev}:${path}`],{maxBuffer:32*1024*1024});
const ownPrefix='solutions/edition-2/main/', assetPrefix=ownPrefix+'gui/web/gui/';
const helperPath=relative(repo,fileURLToPath(import.meta.url));
assert.equal(sha(readFileSync(fileURLToPath(import.meta.url))),sha(git(revision,helperPath)),'helper source identity');
// Enumerate the complete served asset set, not a caller-supplied partial map.
const paths=execFileSync('git',['-C',repo,'ls-tree','-r','--name-only',revision,'--',assetPrefix],{encoding:'utf8'}).trim().split('\n');
assert(paths.length>0 && ['artifacts.js','speech.js','connector.js','style.css'].every(n=>paths.includes(assetPrefix+n)));
const sources={}, assets=new Map();
for(const path of paths){assert(path.startsWith(assetPrefix));const bytes=git(revision,path);assert.equal(sha(readFileSync(join(repo,path))),sha(bytes),`source identity ${path}`);sources[path]=sha(bytes);assets.set('/'+path.slice(assetPrefix.length),bytes);}
const bindingPath=ownPrefix+'evidence/ch10/provider-prep-binding-final.json';
const binding=JSON.parse(git(originalRevision,bindingPath));
const inputs={};
function original(path){const bytes=readFileSync(join(repo,path));assert.equal(sha(bytes),sha(git(originalRevision,path)),`original identity ${path}`);inputs[path]=sha(bytes);return bytes;}
const captures={};
for(const name of ['gemini-B','gemini-B-restart']){
  const prefix=ownPrefix+'evidence/ch10/live-20261008/'+name+'/';
  const manifest=JSON.parse(original(prefix+'originals.json'));
  assert(Object.keys(manifest).length>0 && manifest['browser-original.jsonl'] && manifest['launch.json']);
  for(const [name,hash] of Object.entries(manifest)) {assert(!name.includes('..')&&!name.startsWith('/'));assert.equal(sha(original(prefix+name)),hash,'sealed original');}
  captures[name]=readFileSync(join(repo,prefix+'browser-original.jsonl'),'utf8').trim().split('\n').map(JSON.parse);
}
for(const role of ['node','chrome'])assert.equal(sha(readFileSync(binding.binaries[role].path)),binding.binaries[role].sha256,role+' binary');
assert.equal(sha(readFileSync(process.execPath)),binding.binaries.node.sha256,'actual node');
const dependencies=binding.browser_dependencies;
assert(Object.keys(dependencies.files).length>0 && dependencies.files['package-lock.json']);
for(const [name,hash] of Object.entries(dependencies.files)){assert(!name.includes('..')&&!name.startsWith('/'));assert.equal(sha(readFileSync(join(dependencies.root,name))),hash,'browser dependency');}
// Every source/original/binary check precedes creating any derived output.
mkdirSync(output);
const identity={source_revision:revision,original_revision:originalRevision,original_runtime:binding.source_revision,
  original_support:binding.support_revision,sources,inputs,binaries:{node:binding.binaries.node,chrome:binding.binaries.chrome},
  browser_dependencies_sha256:sha(Buffer.from(JSON.stringify(dependencies))),provider_calls:0,scope:'local captured-data Artifact/SpeechQueue component test; no live rerun'};
writeFileSync(join(output,'identity.json'),JSON.stringify(identity,null,2),{flag:'wx'});
const frames=rows=>rows.filter(r=>r.kind==='received').map(r=>({connection:r.connection,message:JSON.parse(r.payload)}));
function snapshot(frames){let current;for(const {message:m} of frames){if(m.type==='snapshot_begin')current={...m,events:[],partials:[]};else if(current&&m.type==='snapshot_event')current.events.push(m.event);else if(current&&m.type==='snapshot_partial')current.partials.push(m.partial);else if(current&&m.type==='snapshot_end')return current;}throw Error('missing complete snapshot');}
const initialFrames=frames(captures['gemini-B']), restartFrames=frames(captures['gemini-B-restart']);
const firstConnection=initialFrames[0].connection;
const observations=initialFrames.filter(r=>r.connection===firstConnection&&r.message.type==='observation').map(r=>r.message.observation);
const initial=snapshot(initialFrames), restored=snapshot(restartFrames);
const emptyFinal=observations.find(o=>o.kind==='part_final'&&o.response_seq===20&&o.part_index===1);
assert.deepEqual(emptyFinal.part,{parts:[],text:'',type:'text'});
const rawLog=readFileSync(join(repo,ownPrefix+'evidence/ch10/live-20261008/gemini-B/closed-store/events.log'),'utf8').trim().split('\n').map(JSON.parse);
const rawEmpty=rawLog.filter(e=>e.type==='response_ended').map(e=>({seq:e.seq,part:e.response.parts[1]}));
assert.deepEqual(rawEmpty.map(e=>e.seq),[5,10,15,20]);
assert(rawEmpty.every(e=>e.part.type==='text'&&e.part.text===''));
assert.equal(rawEmpty.filter(e=>e.part.opaque).length,2);
const server=createServer((req,res)=>{const asset=assets.get(req.url);res.setHeader('Content-Type',req.url?.endsWith('.css')?'text/css':'application/javascript');if(asset)res.end(asset);else if(req.url==='/'){res.setHeader('Content-Type','text/html');res.end('<!doctype html><link rel="stylesheet" href="/style.css"><main><h1>Local captured-data quality check</h1><section id="cards"></section></main>');}else{res.statusCode=404;res.end();}});
let browser,context,checks;const errors=[],blocked=[];
try{
  await new Promise(r=>server.listen(0,'127.0.0.1',r));const origin=`http://127.0.0.1:${server.address().port}`;
  const {chromium}=await import(pathToFileURL(join(dependencies.root,'node_modules/playwright/index.mjs')));
  browser=await chromium.launch({executablePath:binding.binaries.chrome.path,headless:false});context=await browser.newContext();
  await context.route('**/*',route=>{if(new URL(route.request().url()).origin===origin)return route.continue();blocked.push(route.request().url());return route.abort();});
  const page=await context.newPage();page.on('pageerror',e=>errors.push(e.message));await page.goto(origin);
  checks=await page.evaluate(async({initial,restored,observations,emptyFinal})=>{
    const {ArtifactScroll,partKey}=await import('/artifacts.js'),{SpeechQueue}=await import('/speech.js');
    const checks=[];const check=(condition,name)=>{if(!condition)throw Error(name);checks.push(name);};
    const admitted=[],manual=[];
    const service={available:()=>true,submit(owner,text,callbacks){admitted.push(text);callbacks.end();},cancel(){}};
    const owner={application(){return {speech:()=>service};},diagnostic(){},async reconcile(){},speechEvent(){},speak(key,text){manual.push({key,text});queue.enqueue(key,text);}};
    const queue=new SpeechQueue(owner);queue.enabled=true;queue.playbackReady=true;
    const scroll=new ArtifactScroll(owner,document.querySelector('#cards'),'all');
    const shape=()=>[...scroll.cards.values()].map(c=>({key:c.key,title:c.title,text:c.text,empty:!!c.emptyResponse}));
    const empties=()=>[...scroll.cards.values()].filter(c=>c.emptyResponse);
    const keys=()=>[...scroll.cards.keys()];
    const checkEmpty=label=>{
      check(JSON.stringify(empties().map(c=>c.key))===JSON.stringify([5,10,15,20].map(s=>`d/${restored.agent_id}/${s}/1`)),label+' four durable positions');
      for(const card of empties()){
        check(card.heading.textContent==='Empty response text'&&!card.element.innerText.includes('Accepted'),label+' truthful compact text '+card.key);
        check(card.speaker.hidden&&card.speaker.disabled&&card.expand.hidden&&card.expand.disabled,label+' no controls '+card.key);
        card.speaker.dispatchEvent(new MouseEvent('click'));card.expand.dispatchEvent(new MouseEvent('click'));
        check(!card.expanded,label+' no hidden expansion '+card.key);
      }
    };
    scroll.reset(initial);
    // The recorded B live-final path remains distinct from restart snapshot.
    for(const observation of observations)scroll.observation(observation);
    checkEmpty('captured live-final');const live=shape();
    scroll.reset(restored);checkEmpty('captured restart');check(JSON.stringify(shape())===JSON.stringify(live),'live-final and replay same keys/order/content');
    const stable=shape();
    for(let i=0;i<2;i++){scroll.reset(restored);check(JSON.stringify(shape())===JSON.stringify(stable),`reset ${i+1} same order/content`);check(document.querySelectorAll('article').length===scroll.cards.size,`reset ${i+1} no duplicates`);}
    for(const event of restored.events.filter(e=>e.type==='response_ended'))for(let i=0;i<event.response.parts.length;i++){
      const part=event.response.parts[i];if(part.type==='text'&&part.text==='')queue.observe({...emptyFinal,response_seq:event.seq,part_id:i+1,operation_id:'speech-'+event.seq,part});
    }
    await new Promise(r=>setTimeout(r,0));check(!queue.busy()&&admitted.length===0&&manual.length===0,'empty automatic/manual speech absent');
    const originalKey=`d/${emptyFinal.agent_id}/${emptyFinal.response_seq}/${emptyFinal.part_index}`;
    const placeholder=scroll.cards.get(originalKey);const long='Nonempty replacement. '.repeat(80);
    scroll.part(originalKey,{type:'text',text:long},restored.agent_id);
    check(scroll.cards.get(originalKey)===placeholder&&!placeholder.emptyResponse&&!placeholder.speaker.hidden&&!placeholder.speaker.disabled&&!placeholder.expand.hidden,'nonempty replacement restores same card and controls');
    placeholder.expand.click();check(placeholder.expanded,'nonempty expansion positive');placeholder.speaker.click();
    await new Promise(r=>setTimeout(r,0));check(manual.length===1&&admitted[0]===long,'nonempty manual speech positive');
    queue.observe({...emptyFinal,operation_id:'positive',part:{type:'text',text:'Automatic positive.'}});
    await new Promise(r=>setTimeout(r,0));check(admitted.includes('Automatic positive.'),'nonempty automatic speech positive');
    const synthetic={...emptyFinal,operation_id:'provisional-control',part_id:77,response_seq:999,part_index:0};
    const provisional=scroll.partial(synthetic,{text:''});scroll.final(synthetic);
    check(scroll.cards.get(`d/${synthetic.agent_id}/999/0`)===provisional&&!scroll.cards.has('p/'+partKey(synthetic)),'provisional-to-durable identity retained');
    for(const [name,part] of [['absent',{type:'text'}],['whitespace',{type:'text',text:' \n\t '}],['opaque',{type:'opaque',placeholder:true}]]){
      const card=scroll.part(name,part,restored.agent_id);check(!card.emptyResponse&&!card.speaker.hidden,name+' retains ordinary presentation');if(name==='whitespace')check(card.text===part.text,'nonempty whitespace not trimmed');if(name==='opaque')check(card.text.includes('unavailable for display'),'opaque placeholder retained');
    }
    scroll.event({type:'tool_returned',seq:1000,tool:{call_id:'local-empty-report',parts:[{type:'text',text:''}],job:{handle:999,status:'done'}}},false,restored.agent_id);
    const report=scroll.cards.get(`e/${restored.agent_id}/1000`);check(report&&!report.emptyResponse&&report.text===''&&report.meta.textContent.includes('Report delivered')&&report.meta.textContent.includes('done'),'empty tool-result and lifecycle retained');
    // Leave the screenshot on the actual captured restart, not synthetic controls.
    scroll.reset(restored);queue.close();return {checks,keys:keys(),empty_keys:empties().map(c=>c.key),manual_admissions:manual.length,service_admissions:admitted.length};
  },{initial,restored,observations,emptyFinal});
  assert.deepEqual(errors,[]);assert.deepEqual(blocked,[]);
  writeFileSync(join(output,'dom.txt'),await page.locator('body').innerText(),{flag:'wx'});
  await page.screenshot({path:join(output,'captured-restart.png'),fullPage:true});
  writeFileSync(join(output,'result.json'),JSON.stringify({passed:true,...checks,raw_empty_parts:rawEmpty.map(e=>({response_seq:e.seq,part_index:1,signed:!!e.part.opaque})),errors,provider_calls:0},null,2),{flag:'wx'});
  console.log(JSON.stringify({passed:true,checks:checks.checks.length,provider_calls:0}));
}catch(error){writeFileSync(join(output,'failure.json'),JSON.stringify({passed:false,error:String(error),errors,provider_calls:0},null,2),{flag:'wx'});throw error;}
finally{await context?.close();await browser?.close();await new Promise(r=>server.close(r));}
