// Bounded human-GUI evidence launcher. The local preflight replaces only HTTP.
// Raw application logs stay private in scratch; receipts use explicit allowlists.
import assert from 'node:assert/strict';
import {createHash} from 'node:crypto';
import {execFileSync, spawn} from 'node:child_process';
import {createRequire} from 'node:module';
import {mkdir, mkdtemp, readFile, writeFile} from 'node:fs/promises';
import {join, resolve} from 'node:path';
import {tmpdir, homedir} from 'node:os';
import http from 'node:http';
import net from 'node:net';

const repo=resolve(process.argv[2]||'.');
const mode=process.argv[3]||'preflight';
assert.ok(['preflight','live'].includes(mode));
const require=createRequire(join(repo,'agent/gui/package.json'));
const {chromium}=require('playwright');
const env=Object.fromEntries(Object.entries(process.env).filter(([k])=>! /^(LLM_|OPENAI_|ANTHROPIC_|GEMINI_|EN_|CODEX_)/.test(k)));
const git=(...args)=>execFileSync('git',args,{cwd:repo,env,encoding:'utf8'}).trim();
const sha=bytes=>createHash('sha256').update(bytes).digest('hex');
const delay=ms=>new Promise(resolve=>setTimeout(resolve,ms));
const scratch=await mkdtemp(join(tmpdir(),'ensemble-background-live-'));
const started=performance.now(), stamp=()=>+(performance.now()-started).toFixed(3);
const receipt={mode,source_commit:git('rev-parse','HEAD'),agent_tree:git('rev-parse','HEAD:agent'),scratch,started_utc:new Date().toISOString(),observations:[],limits:[]};
receipt.launcher_sha256=sha(await readFile(new URL(import.meta.url)));
assert.equal(git('status','--porcelain','--untracked-files=all','--','agent'),'','Agent must be committed with no nonignored untracked inputs before building');
const binary=join(scratch,'ensemble-gui');
execFileSync('go',['build','-o',binary,'./cmd'],{cwd:join(repo,'agent/gui'),env,stdio:'pipe'});
assert.equal(git('rev-parse','HEAD:agent'),receipt.agent_tree,'Agent source changed during build');
assert.equal(git('status','--porcelain','--untracked-files=all','--','agent'),'','Agent inputs changed during build');
receipt.binary_sha256=sha(await readFile(binary));
const listen=server=>new Promise(resolve=>server.listen(0,'127.0.0.1',()=>resolve(server.address().port)));
const portProbe=net.createServer(),port=await listen(portProbe);
await new Promise(resolve=>portProbe.close(resolve));

const caseText='Our project is Cedar Lantern, release code CL-721. We rejected automatic deletion of imported notes because the support team needs the originals when customers report mistakes. The replacement design keeps source notes in an archive and stores the accepted summary separately. Morgan owns the archive review, scheduled for October 18. Acceptance requires preserving a source identifier with each summary. The unresolved question is whether attachments belong in the same archive or a separate one; do not claim that has been decided. The default retention is thirty days. A failed summarization must leave the original note available. The initial approach coupled cleanup to request arrival, so an unrelated request could delete unreviewed work. We moved cleanup to an explicit completed-review boundary instead.';
const prompts=[
 'This is a bounded memory exercise in an empty scratch workspace. Do not use filesystem or command tools. Analyze this case, explain the decision briefly, then call micro_handoff exactly once with the important facts and reasons. After the tool returns, finish this turn without another tool call. '+caseText,
 'Without tools or a checkpoint, what is 17 plus 25? Answer briefly.',
 'Without tools or a checkpoint, recall the project name, release code, owner, date, rejected approaches and reasons, source-identifier requirement, retention, and unresolved attachment question from the supplied case. Do not invent a resolved answer.'
];
receipt.prompts=prompts;
const memory='I analyzed Cedar Lantern (CL-721). '+caseText;
const textItem=text=>({type:'message',role:'assistant',content:[{type:'output_text',text}]});
const callItem=(name,args)=>({type:'function_call',call_id:name+'-one',name,arguments:JSON.stringify(args)});
let releaseWorker=()=>{},fake,mainCalls=0,workerCalls=0;
if(mode==='preflight'){
 const held=new Promise(resolve=>{releaseWorker=resolve;});
 fake=http.createServer(async(request,response)=>{
  let raw='';for await(const bytes of request)raw+=bytes;
  const body=JSON.parse(raw);
  const tools=(body.tools||[]).flatMap(t=>t.type==='namespace'?t.tools:[t]);
  const worker=tools.length===1&&tools[0].name==='submit'&&tools[0].parameters?.properties?.memory;
  let output;
  if(worker){workerCalls++;await held;output=[textItem('I will preserve the owner’s decision and reasons in the requested memory.'),callItem('submit',{memory})];}
  else {mainCalls++;output=mainCalls===1?[textItem('I will checkpoint the completed analysis, including why originals must remain available.'),callItem('micro_handoff',{text:caseText})]:[textItem(mainCalls===2?'Analysis checkpointed.':mainCalls===3?'42.':memory)];}
  response.setHeader('content-type','application/json');
  response.end(JSON.stringify({status:'completed',model:'gpt-6-astra',output,usage:{input_tokens:worker?31:101,output_tokens:worker?17:7,input_tokens_details:{cached_tokens:worker?11:41}}}));
 });
}
const baseURL=fake?`http://127.0.0.1:${await listen(fake)}/v1`:'https://chatgpt.com/backend-api/codex';
await mkdir(join(scratch,'skills/base'),{recursive:true});
await writeFile(join(scratch,'skills/base/SKILL.md'),'---\nname: base\ntype: primary\ndescription: Analyze supplied cases and preserve completed-work reasons\ntools: read_file write_file list_directory\n---\nPreserve decisions, reasons, failed approaches, constraints and uncertainty in checkpoint notes. Do only the current task. Explain your action in ordinary visible text before every tool call. Finish main turns in chat.\n');
if(fake)await writeFile(join(scratch,'fixture-key'),'unused-local-key',{mode:0o600});
const config={version:1,provider:{vendor:'openai',model:'gpt-6-astra',billing_mode:fake?'api':'codex',base_url:baseURL,credential_file:fake?join(scratch,'fixture-key'):join(homedir(),'.codex/auth.json'),reasoning_effort:'high',reasoning_summary:'detailed'},workspace:scratch,skills_dir:'skills',primary_skill:'base',data_dir:'cr',save_path:'save.json',event_log:'event.log',auth_directory:'auth',gui_port:port,enable_web_search:false,recall:{disabled:true},gui:{tts_enabled:false,max_tool_rounds:3,memory:{conversation:{budget:600},session:{budget:100000,high_watermark:200000,low_watermark:100000},'8x':{disabled:true},'64x':{disabled:true},memory:{disabled:true},soul:{disabled:true}}}};
await writeFile(join(scratch,'settings.json'),JSON.stringify(config),{mode:0o600});
receipt.configuration={model:config.provider.model,effort:'high',route:config.provider.billing_mode,endpoint:baseURL,endpoint_evidence:'explicit launch configuration and source, not independently captured network destination',memory:config.gui.memory,memory_evidence:'Launch configuration. Omitted conversation high/low watermarks resolve to 1200/600 bytes in internal/memory.Settings (2*budget/budget); these effective values are source-derived, not captured GUI settings.',recall_disabled:true};

// Merge only complete journal lines with the latest snapshot, keyed by Seq.
async function events(save=join(scratch,'save.json')){
 const found=new Map();
 try {for(const e of JSON.parse(await readFile(save,'utf8')).log||[])found.set(e.seq,e);}catch(e){if(e.code!=='ENOENT')throw e;}
 try {const lines=(await readFile(save+'.journal','utf8')).split('\n');lines.pop();for(const line of lines)if(line.trim()){const e=JSON.parse(line);found.set(e.seq,e);}}catch(e){if(e.code!=='ENOENT')throw e;}
 return [...found.values()].sort((a,b)=>a.seq-b.seq);
}
async function requests(){
 try {const lines=(await readFile(join(scratch,'api.log'),'utf8')).split('\n');lines.pop();return lines.filter(Boolean).map(JSON.parse).filter(x=>x.event==='request').map(x=>JSON.parse(x.data));}catch(e){if(e.code==='ENOENT')return[];throw e;}
}
const messageText=body=>(body.input||[]).filter(x=>x.type==='message').map(x=>(x.content||[]).filter(p=>['input_text','output_text'].includes(p.type)).map(p=>p.text).join('\n')).join('\n');
const toolsOf=body=>(body.tools||[]).flatMap(t=>t.type==='namespace'?t.tools:[t]);
const isWorker=body=>{const tools=toolsOf(body);return tools.length===1&&tools[0].name==='submit'&&tools[0].parameters?.properties?.memory;};
const deadline=performance.now()+(fake?90000:600000);
let proc,browser,pages=[],exited,exitResult,observing=true;
const seen=new Set();
const observe=(async()=>{
 while(observing){
  for(const e of await events())if(!seen.has(e.seq)){
   seen.add(e.seq);
   if(['compactor_launched','compaction_outcome','request_sent','response_ended'].includes(e.type)||e.type==='message_received'&&e.message?.actor==='human')receipt.observations.push({event:e.type,seq:e.seq,recorded_utc:e.time,observed_ms:stamp(),task_id:e.compactor?.task_id||e.compaction_outcome?.task_id});
  }
  await delay(25);
 }
})();
async function waitFor(predicate,label){
 while(performance.now()<deadline){const value=await predicate();if(value)return value;if(exitResult)throw Error('GUI exited while waiting for '+label);await delay(40);}
 throw Error('Observation budget exhausted waiting for '+label);
}
function launch(){
 exitResult=undefined;
 proc=spawn(binary,['--config',join(scratch,'settings.json')],{cwd:scratch,env,stdio:['pipe','pipe','pipe']});
 // Do not mirror raw provider diagnostics or terminal/private parts into evidence.
 proc.stdout.resume();proc.stderr.resume();
 exited=new Promise(resolve=>proc.on('exit',(code,signal)=>{exitResult={code,signal};resolve(exitResult);}));
}
async function stop(){
 if(!proc||exitResult)return exitResult;
 proc.stdin.end();
 let timer=setTimeout(()=>proc.kill('SIGTERM'),5000),killer=setTimeout(()=>proc.kill('SIGKILL'),10000);
 try{return await exited;}finally{clearTimeout(timer);clearTimeout(killer);}
}
const frames=[];
async function connect(label){
 const page=await browser.newPage();pages.push(page);
 page.on('websocket',socket=>socket.on('framereceived',frame=>{
  const value=JSON.parse(frame.payload);
  // Exclude part_delta and artifact payloads, which may contain provider replay.
  if(['usage','turn_ended','agent_selected','agent_tree'].includes(value.type))frames.push({at_ms:stamp(),page:label,...value});
 }));
 await page.goto(`http://127.0.0.1:${port}/?speech=transcript`);
 await page.locator('[data-agent-key="main"]').waitFor();
 return page;
}
async function prompt(page,text){
 receipt.observations.push({event:'input_sent',at_ms:stamp(),text});
 await page.locator('#input').fill(text);await page.locator('#input-form button').click();
}
const mainEnds=()=>frames.filter(f=>f.type==='turn_ended'&&f.agent_key==='main'&&f.page==='main');
try{
 launch();
 await waitFor(async()=>{try{return(await fetch(`http://127.0.0.1:${port}`)).ok;}catch{return false;}},'isolated listener');
 browser=await chromium.launch({headless:true,executablePath:'/Applications/Google Chrome.app/Contents/MacOS/Google Chrome'});
 const page=await connect('main');
 await prompt(page,prompts[0]);
 await waitFor(()=>mainEnds().length>=1,'first main completion');
 assert.equal(mainEnds()[0].error,'');
 const launchEvent=await waitFor(async()=>(await events()).find(e=>e.compactor?.task_id),'worker launch');
 receipt.launch=launchEvent.compactor;receipt.launch_observed_ms=stamp();
 const childKey='main/'+receipt.launch.task_id;
 // Keep the main tab on its owner, so a selected child cannot receive the hint.
 const childPage=await connect('child-observer');
 const child=childPage.locator(`[data-agent-key="${childKey}"]`);
 if(await child.count()){
  try{
   await child.click({timeout:2000});
   await childPage.waitForFunction(key=>document.querySelector('#conversation-title').textContent===key,childKey,{timeout:2000});
   receipt.child_selected_ms=stamp();
   receipt.child_view={title:await childPage.locator('#conversation-title').textContent(),tree:await childPage.locator('#agent-tree').textContent()};
   await childPage.locator('#agent-tree').screenshot({path:join(scratch,'running-tree.png')});
  }catch(error){
   if(!(await events()).some(e=>e.compaction_outcome?.task_id===receipt.launch.task_id))throw error;
   receipt.limits.push('Worker finished during selection; selected-child fallback not established.');
  }
 }else receipt.limits.push('Worker completed before browser selection; no selected-child fallback observed.');
 await prompt(page,prompts[1]);
 await waitFor(()=>mainEnds().length>=2,'unrelated main completion');
 assert.equal(mainEnds()[1].error,'');
 receipt.usage_before_completion=frames.filter(f=>f.type==='usage'&&f.agent_key==='main'&&f.page==='main').at(-1);
 receipt.outcome_already_present=(await events()).some(e=>e.compaction_outcome?.task_id===receipt.launch.task_id);
 releaseWorker();
 const outcome=await waitFor(async()=>(await events()).find(e=>e.compaction_outcome?.task_id===receipt.launch.task_id),'terminal memory outcome');
 receipt.outcome=outcome.compaction_outcome;receipt.outcome_observed_ms=stamp();
 assert.equal(receipt.outcome.state,'committed');
 const ownerEvents=await events(),secondInput=ownerEvents.filter(e=>e.type==='message_received'&&e.message?.actor==='human')[1];
 receipt.main_progress_during_worker={launch_seq:launchEvent.seq,second_input_seq:secondInput?.seq,outcome_seq:outcome.seq,answered_before_outcome:ownerEvents.some(e=>e.type==='response_ended'&&e.seq>secondInput?.seq&&e.seq<outcome.seq)};
 if(!receipt.main_progress_during_worker.answered_before_outcome)receipt.limits.push('No unrelated second-turn answer before worker outcome in this run; deterministic held-provider proof remains separate.');
 await waitFor(async()=>!(await page.locator(`[data-agent-key="${childKey}"]`).count()),'automatic tree dismissal');
 if(receipt.child_selected_ms){await childPage.waitForFunction(()=>document.querySelector('#conversation-title').textContent==='main');receipt.fallback_observed_ms=stamp();}
 // Fresh subscription observes the owner's current meter without model work.
 const meterPage=await connect('meter');
 await waitFor(()=>frames.some(f=>f.type==='usage'&&f.page==='meter'),'post-completion meter');
 receipt.usage_after_completion=frames.filter(f=>f.type==='usage'&&f.page==='meter').at(-1);
 await meterPage.close();
 const sourceRequests=await requests(),workerRequests=sourceRequests.filter(isWorker);
 assert.ok(workerRequests.length>0);
 receipt.worker_requests=workerRequests.map(body=>({model:body.model,effort:body.reasoning?.effort,schema:toolsOf(body)[0].parameters,visible_policy:messageText(body).includes('visible'),readable_input:messageText(body)}));
 const workerSave=join(scratch,'cr',receipt.launch.trace_path,'save.json');
 const childEvents=await events(workerSave);
 receipt.worker_visible=childEvents.filter(e=>e.response).flatMap(e=>e.response.parts||[]).filter(p=>p.type==='text').map(p=>p.text);
 receipt.worker_response_order=childEvents.filter(e=>e.response).map(e=>({seq:e.seq,parts:e.response.parts.filter(p=>['text','tool_call'].includes(p.type)).map(p=>p.type==='text'?{type:p.type,text:p.text}:{type:p.type,name:p.name,args:p.args})}));
 receipt.worker_submissions=childEvents.filter(e=>e.response).flatMap(e=>e.response.parts||[]).filter(p=>p.type==='tool_call'&&p.name==='submit').map(p=>p.args);
 const population=(await events()).find(e=>e.populated?.task_id===receipt.launch.task_id)?.populated;
 assert.ok(population?.text);receipt.committed_memory=population.text;
 const path=join(scratch,'cr/memory',`${population.file.date}-${population.file.num}.md`);
 receipt.committed_file=await readFile(path,'utf8');
 receipt.sizes={selected_source_utf8:Buffer.byteLength(messageText(workerRequests[0]).split('<source-data>\n')[1]?.split('\n</source-data>')[0]||''),submitted_memory_utf8:Buffer.byteLength(population.text),committed_file_utf8:Buffer.byteLength(receipt.committed_file)};
 await prompt(page,prompts[2]);await waitFor(()=>mainEnds().length>=3,'recall main completion');
 assert.equal(mainEnds()[2].error,'');receipt.recall_reply=mainEnds()[2].text;
 const recallRequest=(await requests()).filter(x=>!isWorker(x)).at(-1);
 receipt.recall_received_exact_memory=messageText(recallRequest).includes(population.text);
 assert.ok(receipt.recall_received_exact_memory);
 receipt.request_metadata=(await requests()).map(x=>({model:x.model,effort:x.reasoning?.effort,worker:!!isWorker(x)}));
 receipt.main_turns=mainEnds();receipt.browser_selections=frames.filter(f=>f.type==='agent_selected');
 receipt.last_tree=frames.filter(f=>f.type==='agent_tree').at(-1);
 if(fake){assert.equal(workerCalls,1);assert.equal(mainCalls,4);assert.equal(receipt.outcome_already_present,false);assert.equal(receipt.main_progress_during_worker.answered_before_outcome,true);}
 receipt.first_exit=await stop();assert.equal(receipt.first_exit.code,0);
 for(const p of pages)if(!p.isClosed())await p.close();pages=[];
 // The GUI recreates api.log at restart. Preserve the sanitized first-run receipt now.
 await writeFile(join(scratch,'receipt.json'),JSON.stringify(receipt,null,2));
 launch();await waitFor(async()=>{try{return(await fetch(`http://127.0.0.1:${port}`)).ok;}catch{return false;}},'restart listener');
 await connect('restart');await delay(300);
 receipt.restart_requests=(await requests()).length;assert.equal(receipt.restart_requests,0);
 assert.ok((await events()).some(e=>e.compaction_outcome?.task_id===receipt.launch.task_id&&e.compaction_outcome.state==='committed'));
 receipt.restart_exit=await stop();assert.equal(receipt.restart_exit.code,0);
 await checkReceipt(receipt);
 if(fake){
  // A receipt verifier must reject an incomplete identity before trusting results.
  const controls=[['empty',()=>({})],['missing source',r=>({...r,source_commit:''})],['wrong source tree',r=>({...r,agent_tree:'0'.repeat(40)})],['wrong binary',r=>({...r,binary_sha256:'0'.repeat(64)})],['empty worker request',r=>({...r,worker_requests:[]})],['wrong committed bytes',r=>({...r,committed_file:'unrelated'})],['missing restart',r=>({...r,restart_exit:undefined})]];
  receipt.verifier_negative_controls=[];
  for(const [name,change] of controls){await assert.rejects(()=>checkReceipt(change(structuredClone(receipt))));receipt.verifier_negative_controls.push(name);}
 }
 receipt.status='complete';
}catch(error){receipt.status='failed';receipt.error=error.message;process.exitCode=1;}
finally{
 releaseWorker();await stop();if(browser)await browser.close();
 observing=false;await observe;
 if(fake){fake.closeAllConnections();await new Promise(resolve=>fake.close(resolve));}
 receipt.finished_utc=new Date().toISOString();await writeFile(join(scratch,'receipt.json'),JSON.stringify(receipt,null,2));
 console.log(JSON.stringify({mode,status:receipt.status,error:receipt.error,receipt:join(scratch,'receipt.json')}));
}

// This checks the actual bounded observation, not model faithfulness. A human
// still compares retained reasons/uncertainty against the supplied public case.
async function checkReceipt(r){
 assert.match(r.source_commit||'',/^[a-f0-9]{40}$/);
 assert.equal(git('rev-parse',r.source_commit+':agent'),r.agent_tree);
 assert.equal(r.scratch,scratch);assert.equal(r.binary_sha256,sha(await readFile(binary)));
 assert.equal(r.launcher_sha256,sha(await readFile(new URL(import.meta.url))));
 assert.ok(r.worker_requests?.length);assert.ok(r.sizes?.selected_source_utf8>0);
 const schema={type:'object',properties:{memory:{type:'string'}},required:['memory'],additionalProperties:false};
 for(const request of r.worker_requests){assert.equal(request.model,'gpt-6-astra');assert.equal(request.effort,'high');assert.deepEqual(request.schema,schema);}
 assert.equal(r.outcome?.task_id,r.launch?.task_id);assert.equal(r.outcome?.state,'committed');
 assert.equal(r.committed_file,r.committed_memory);
 assert.ok(r.worker_submissions.some(p=>p.memory===r.committed_memory));
 const accepted=r.worker_response_order.find(e=>e.parts.some(p=>p.type==='tool_call'&&p.args?.memory===r.committed_memory));
 const toolIndex=accepted?.parts.findIndex(p=>p.type==='tool_call');
 assert.ok(toolIndex>0&&accepted.parts.slice(0,toolIndex).some(p=>p.type==='text'&&p.text.trim()));
 assert.equal(r.recall_received_exact_memory,true);assert.equal(r.main_turns?.length,3);
 assert.equal(r.first_exit?.code,0);assert.equal(r.restart_exit?.code,0);assert.equal(r.restart_requests,0);
 if(r.child_selected_ms){assert.equal(r.child_view?.title,'main/'+r.launch.task_id);assert.ok(r.fallback_observed_ms>r.child_selected_ms);}
 if(!r.outcome_already_present){
  const before=r.usage_before_completion,after=r.usage_after_completion;
  for(const field of ['input','cache_write','cache_read','output']){
   assert.equal(after[field],before[field],'auxiliary completion must not replace latest main usage');
   const child=r.outcome.usage.reduce((sum,b)=>sum+(b.usage[field]||0),0);
   assert.equal(after.total[field]-before.total[field],child,'auxiliary usage must enter totals once');
  }
 }
}
