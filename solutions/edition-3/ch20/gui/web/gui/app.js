import {ArtifactScroll} from './artifact-scroll.js';
'use strict';
const $ = id => document.getElementById(id);
let socket, clientID = '', selected = '', retry = 500, session = '';
const skillsByAgent=new Map();
const mcpAgents=new Set(), mcpReady=new Set(), speechRecent=[];
const nodeIDs=new WeakMap();let nextNodeID=0;
const agents = new Map();
let audio = [], speakingNow = null;
const typing = new Set(), delayed = new Map(), requests = new Map(), voiceAgents = new Set();
let inputOccupied=false,pageTurnRevision=0,pageTurnsEnded=0,inputRevision=0;
const observedTurns=new Map(),speechStartedAt=Date.now();
const speechTranscript=[],pauseRequested=new Set(),admittedSpeech=new Set();
let speechMode=new URLSearchParams(location.search).get('speech')==='transcript'?'transcript':'native';
// Full speech-channel records, separate from clipped GUI-debug diagnostics.
window.speechChannel={
 status(agent=selected){const state=agents.get(agent)?.state,turn=observedTurns.get(agent);return {ready:socket?.readyState===1&&settingsReady&&!!state,agent_id:agent,selected_agent_id:selected,turn:state?.turn,turn_id:turn?.id||0,outcome:turn?.terminal?turn.outcome:'pending',prompt:turn?.prompt||'',paused:state?.paused||false,busy:!!agent&&busy(agent),revision:pageTurnRevision,ended:pageTurnsEnded,input_revision:inputRevision,draft:agent===selected?$('input').value:'',human_typing:typing.has(agent)};},
 releaseDraft(agent,revision,text){if(agent!==selected||revision!==inputRevision||$('input').value!==text)return false;$('input').value='';inputRevision++;inputOccupied=false;typing.delete(agent);updatePause(agent);return true;},
 interruptOwned(agent,turn){if(observedTurns.get(agent)?.id!==turn||observedTurns.get(agent)?.terminal)return false;cancelAudio(agent,false);return send('interrupt',{agent_id:agent});},
 read(after=0){
  if(!Number.isSafeInteger(after)||after<0||after>speechTranscript.length)throw Error('invalid speech cursor');
  return {cursor:speechTranscript.length,mode:speechMode,events:speechTranscript.slice(after)};
 },
 setMode(mode){
  if(!['transcript','native'].includes(mode))throw Error('speech mode must be transcript or native');
  if(speakingNow||audio.length||delayed.size||requests.size)throw Error('speech is busy');
  speechMode=mode;return {mode};
 }
};
function channelRecord(task,event,extra={}){
 speechTranscript.push({ms:Date.now()-speechStartedAt,kind:event==='complete'?'status':event,seq:speechTranscript.length+1,causal_sequence:speechTranscript.length+1,at:Date.now(),event,mode:speechMode,agent_id:task.agent,speech_id:task.id,turn_id:task.turn_id||0,...extra});
}
function speechChunks(task){
 const out=[];
 for(const part of task.segments||[{text:task.text,source:'system'}]){
  for(const paragraph of part.text.split(/\n\s*\n/)){
   let rest=paragraph.trim();
   while(Array.from(rest).length>1200){
    const prefix=Array.from(rest).slice(0,1200).join('');
    let cut=prefix.lastIndexOf(' ');if(cut<600)cut=prefix.length;
    out.push({...part,text:rest.slice(0,cut)});rest=rest.slice(cut).trimStart();
   }
   if(rest)out.push({...part,text:rest});
  }
 }
 return out;
}
let settings={theme:'dark',tts_enabled:true,pause_while_speaking:true,tts_speed:1,font_size:16,context_target:0,max_tool_rounds:0,log_mode:'full',diagnostic_max_records:200,safe_mode:false,enable_web_search:true},settingsReady=false;
// A target draft is local; accepted shared settings stay authoritative.
let contextTargetDraft=null,contextTargetPending=null;
const preferenceFields={'theme':'theme','tts-enabled':'tts_enabled','pause-while-speaking':'pause_while_speaking','tts-speed':'tts_speed','font-size':'font_size','context-target':'context_target','max-tool-rounds':'max_tool_rounds','log-mode':'log_mode','diagnostic-max-records':'diagnostic_max_records','safe-mode':'safe_mode','enable-web-search':'enable_web_search'};
const systemTheme=window.matchMedia?.('(prefers-color-scheme: dark)');
function applyTheme(){document.documentElement.dataset.theme=settings.theme==='system'?(systemTheme?.matches?'dark':'light'):settings.theme;}
if(systemTheme?.addEventListener)systemTheme.addEventListener('change',()=>{if(settings.theme==='system')applyTheme();});
function preferenceControls(){
 for(const [id,key] of Object.entries(preferenceFields)){const control=$(id);if(key==='context_target'&&contextTargetDraft!==null){control.value=contextTargetDraft;continue;}if(control.type==='checkbox')control.checked=settings[key];else control.value=String(key==='context_target'&&settings[key]===0?400000:settings[key]);}
 $('font-size-value').textContent=String(settings.font_size);$('tts-speed-value').textContent=`${settings.tts_speed}×`;$('preferences').disabled=!settingsReady;
}
function automaticSpeech(task){return !['on demand','user'].includes(task.reason);}
function cancelAutomaticAudio(){
 const removed=audio.filter(automaticSpeech);audio=audio.filter(task=>!automaticSpeech(task));
 for(const [id,item] of delayed){if(automaticSpeech(item.task)){clearTimeout(item.timer);delayed.delete(id);removed.push(item.task);}}
 if(speakingNow&&automaticSpeech(speakingNow)){removed.push(speakingNow);speakingNow=null;window.speechSynthesis?.cancel();}
 for(const task of removed){speechRecord(task,'disabled');send('speech_done',{agent_id:task.agent,speech_id:task.id,status:'disabled'});}
 speakNext();
}
function applySettings(value){
 if(!value)return;
 settings=value;settingsReady=true;
 if(contextTargetPending&&settings.context_target===contextTargetPending.value){
  contextTargetDraft=null;contextTargetPending=null;
 }
 applyTheme();document.documentElement.style.setProperty('--font-size',`${settings.font_size}px`);preferenceControls();
 // Speech and HUMAN input own independent holds. Disabling speech does not
 // grant permission to resume through a draft the human is still composing.
 if(!settings.tts_enabled)cancelAutomaticAudio();
 for(const agent of agents.keys())updatePause(agent);
 $('settings-status').textContent=contextTargetDraft===null?'Saved · shared across your open tabs':contextTargetPending?'Saving context target…':'Context target edit not saved';
}
$('context-target').oninput=()=>{
 contextTargetDraft=$('context-target').value;
 $('settings-status').textContent='Context target edit not saved';
 // New typing invalidates the earlier submission, even if its text is identical.
 contextTargetPending=null;
};
for(const [id,key] of Object.entries(preferenceFields))$(id).onchange=()=>{
 if(key==='context_target'){contextTargetDraft=$(id).value;contextTargetPending=null;}
 const control=$(id);let value=control.type==='checkbox'?control.checked:control.value;
 if(['font_size','tts_speed','context_target','max_tool_rounds','diagnostic_max_records'].includes(key))value=Number(value);
 if((typeof value==='number'&&!Number.isFinite(value))||(['context_target','max_tool_rounds','diagnostic_max_records'].includes(key)&&!Number.isSafeInteger(value))){$('settings-status').textContent='Enter a finite whole-number budget.';preferenceControls();return;}
 // The target draft stays visible while accepted settings remain unchanged.
 preferenceControls();
 if(send('update_settings',{settings:{[key]:value}})){
  // A negative draft can only be refused; it must not converge to the floor.
  if(key==='context_target'&&value>=0)contextTargetPending={value:value===0?0:Math.max(20000,value)};
  $('settings-status').textContent='Saving…';
 }
};
$('font-size').oninput=()=>{$('font-size-value').textContent=$('font-size').value;};
$('tts-speed').oninput=()=>{$('tts-speed-value').textContent=`${$('tts-speed').value}×`;};
const element = (tag, cls, text) => { const e = document.createElement(tag); if(cls) e.className=cls; if(text !== undefined) e.textContent=String(text); return e; };
function send(type, data={}) { if(socket?.readyState !== WebSocket.OPEN) { $('errors').textContent='Disconnected; wait for session replay before sending.'; return false; } socket.send(JSON.stringify({type,agent_id:selected,...data})); return true; }
const artifactScroll=new ArtifactScroll($('conversation'),{toolContainer:$('tool-cards'),followButton:$('follow'),onSpeak:(agent_id,part_id)=>send('speak',{agent_id,part_id})});

function skillView(){
 const box=$('skills');if(!box)return;box.replaceChildren();const state=skillsByAgent.get(selected);if(!state){box.textContent='No skills configured.';return;}
 const loadable=new Set((state.loadable_skills||[]).map(s=>s.name));
 for(const s of state.skills||[]){
  const row=document.createElement('div'),label=document.createElement('span'),button=document.createElement('button');
  label.textContent=`${s.name} (${s.type}) — ${s.enabled?(s.pending_unload?(s.mcp_servers?.length?'external tools stopped; instructions retained':'pending unload; still active'):'enabled'):(loadable.has(s.name)?'loadable':'discovered')} ${s.description||''} `;
  row.append(label);if(s.enabled&&!s.pending_unload){button.textContent='Mark pending unload';button.onclick=()=>send('unload_skill',{name:s.name});row.append(button);}else if(loadable.has(s.name)||s.pending_unload&&s.type==='loadable'){button.textContent=s.pending_unload?(s.mcp_servers?.length?'Reload':'Keep loaded'):'Load';button.onclick=()=>send('load_skill',{name:s.name});row.append(button);}box.append(row);
 }
 if(!(state.skills||[]).length)box.textContent='No skills configured.';
}
// Public GUI/History observations, scoped to each Agent and turn.
function observeTurn(frame){
 if(frame.stale||!frame.turn_id)return;
 let turn=observedTurns.get(frame.agent_id);
 if(turn&&frame.turn_id<turn.id)return;
 if(!turn||frame.turn_id>turn.id){turn={id:frame.turn_id,outcome:'success',terminal:false,prompt:''};observedTurns.set(frame.agent_id,turn);}
 if(frame.part?.role==='human')turn.prompt=frame.part.text||'';
 if(frame.type==='error'&&turn.outcome!=='interrupted')turn.outcome='error';
 if(frame.type==='turn_interrupted'){turn.outcome='interrupted';turn.terminal=true;}
 if(frame.type==='turn_ended')turn.terminal=true;
}
function apply(frame) {
 if(frame.type==='human_question') { const q=frame.question;const answer=window.prompt(q.Text);if(answer!==null)send('answer_human',{agent_id:q.Sender,request_id:q.RequestID,text:answer});return; }
 if(frame.type==='usage'){const a=agents.get(frame.agent_id);if(a){a.usage=frame;a.meter=frame.meter;}if(frame.agent_id===selected)renderMeter();return;}
 observeTurn(frame);
 artifactScroll.receive(frame);
 if(frame.type==='state'||frame.type==='state_changed') {
  if(!frame.replay)pageTurnRevision++;
  const a=agents.get(frame.agent_id), next=frame.type==='state'?frame.data:frame.state, previous=a?.state;if(a) a.state=next;
  if(frame.agent_id===selected) controls();
  // A state acknowledgement may follow newer input. Only explicit Continue
  // authority (below), not an automatic continuation epoch, clears typing.
  if(!frame.replay&&next.epoch!==previous?.epoch)cancelAudio(frame.agent_id,false);
 } else if(frame.type==='turn_ended') {if(!frame.replay)pageTurnsEnded++;}
 else if(frame.type==='continued'&&!frame.replay){
  if(frame.data.client!==clientID&&frame.data.epoch===agents.get(frame.agent_id)?.state.epoch)explicitContinue(frame.agent_id);
 }else if(frame.type==='skill_state'){skillsByAgent.set(frame.agent_id,frame);skillView();
 if(frame.browser_mcp&&!mcpAgents.has(frame.agent_id)){mcpAgents.add(frame.agent_id);send('mcp_open',{agent_id:frame.agent_id});}
 }else if(frame.type==='current_settings'||frame.type==='settings_changed') applySettings(frame.settings);
 else if(frame.type==='settings_error') {contextTargetPending=null;$('settings-status').textContent=String(frame.data);preferenceControls();}
 else if(frame.type==='error') $('errors').textContent=String(frame.data);
 else if(frame.type==='speech_requested'&&!frame.replay){
  if(frame.data.epoch!==agents.get(frame.agent_id)?.state.epoch)return;
  requests.set(frame.data.id,{agent:frame.agent_id,epoch:frame.data.epoch});
  $('audio-status').textContent='Browser speech requested';send('claim_speech',{agent_id:frame.agent_id,speech_id:frame.data.id});
 }else if(frame.type==='speech_claimed'&&!frame.replay){
  const request=requests.get(frame.data.id);requests.delete(frame.data.id);
  if(frame.data.owner!==clientID||admittedSpeech.has(frame.data.id)||(request?request.epoch!==frame.data.epoch:!voiceAgents.has(frame.agent_id))||frame.data.epoch!==agents.get(frame.agent_id)?.state.epoch)return;
  admittedSpeech.add(frame.data.id);
  const task={agent:frame.agent_id,...frame.data};voiceAgents.add(task.agent);updatePause(task.agent);
  if(task.reason==='prompt'&&$('mic').checked&&settings.tts_enabled){const timer=setTimeout(()=>{delayed.delete(task.id);enqueueAudio(task);},2000);delayed.set(task.id,{task,timer});}else enqueueAudio(task);
 }else if(frame.type==='speech_cancelled'){
  cancelAudio(frame.agent_id,false);
  if(frame.data.release&&frame.data.owner===clientID)maybeResume(frame.agent_id);
 }else if(frame.type==='speech_idle'){
  if(frame.data.owner===clientID&&frame.data.epoch===agents.get(frame.agent_id)?.state.epoch){voiceAgents.delete(frame.agent_id);maybeResume(frame.agent_id);}
 }else if(frame.type==='speech_done'){requests.delete(frame.data.id);$('audio-status').textContent=`Speech: ${frame.data.status}`;}

}
function selectedModel(a){return a?.meter?.selected_model||a?.config?.model||'';}
function renderAgentTree(a){
 const node=$('agent-root');node.disabled=!a;if(!a)return;
 const state=a.state.paused?'paused':a.state.turn,label=state.replaceAll('_',' ');
 $('agent-name').textContent=`Agent ${a.id}`;$('agent-activity').textContent=label.charAt(0).toUpperCase()+label.slice(1);
 $('agent-cue').dataset.busy=String(state!=='idle');node.dataset.agentId=a.id;node.title=selectedModel(a);node.setAttribute('aria-selected',String(a.id===selected));
}
// All four values are process-local Engine measurements, never replay totals.
function renderMeter(){
 const node=$('cache-meter'),a=agents.get(selected),u=a?.usage,m=u?.meter;
 if(document.activeElement!==$('model-name'))$('model-name').value=selectedModel(a);
 $('model').textContent=selectedModel(a);
 renderAgentTree(a);
 const percent=n=>n==null?'unknown':`${(n*100).toFixed(1)}%`;
 const count=n=>n==null||n<0?'unknown':n.toLocaleString();
 if(!u||!m){node.textContent='$0.0000 · In 0 · Cache 0 · Hit 0.0%';node.title='No current-process response yet; initial zeroes are not provider measurements.';return;}
 const allInput=[u.input,u.cache_read,u.cache_write].every(n=>n!=null&&n>=0)?u.input+u.cache_read+u.cache_write:-1;
 const cost=u.priced&&u.cost_usd!=null?`$${u.cost_usd.toFixed(4)}`:`${m.cost.status} (known $${m.cost.usd.toFixed(4)})`;
 node.textContent=`${cost} · In ${count(allInput)} · Cache ${count(u.cache_read)} · Hit ${percent(u.hit_rate)}`;
 const usage=(label,v)=>`${label}: uncached input ${count(v.input)}, output ${count(v.output)}, cache write ${count(v.cache_write)}, cache read ${count(v.cache_read)}`;
 const c=m.cache,comparison=c.comparison;
 node.title=[
  'Current process only; restored conversation usage is not included.',
  `Selected model: ${m.selected_model||'unknown'}; last producing model: ${m.last_model||'unknown'}.`,
  m.requests===0?'No response yet: initial zeroes are not measured usage.':'In = latest uncached input + cache write + cache read, counted once. Cache = latest read tokens. Hit = session token-weighted ratio.',
  usage('Last response',u),usage('Session',u.total),
  `Cost: ${m.cost.status}; known subtotal $${m.cost.usd.toFixed(6)}. Unknown work is not priced as zero.`,
  ...m.buckets.map(b=>`${b.vendor}/${b.surface}/${b.model||'unknown model'}: ${b.requests} responses; $${b.cost.usd.toFixed(6)} (${b.cost.status}); ${usage('tokens',b.usage)}`),
  `Breakpoints: ${c.marker_count}. ${c.marker_reason||'No sent request yet.'}`,
  `Marked blocks: ${c.current?.segments.filter(s=>s.marked).map(s=>s.path).join(', ')||'none'}`,
  `Local request comparison (not provider cache hits): ${c.shared_bytes||0} shared prompt bytes; first difference: ${c.first_difference||'none'}.`,
  ...(c.sections||[]).map(s=>`${s.name}: ${s.change} (${s.bytes} bytes)`),
  `Non-prompt options: ${c.options_change||'no send yet'}. ${c.prediction?.reason||''}`,
  ...(comparison?[`Response to request ${comparison.prediction.request_id}: predicted byte fraction ${percent(comparison.prediction.fraction)}; measured token cache-hit ratio ${percent(comparison.actual_hit_rate)}. Neither promises a future provider cache grant.`]:[]),
  ...(c.diagnostic_error?[`Diagnostic file error: ${c.diagnostic_error}`]:[])
 ].join('\n');
}
function controls(){skillView();renderMeter();const a=agents.get(selected);renderAgentTree(a);if(!a)return;const s=a.state;$('vendor').textContent=a.config.vendor;$('model').textContent=selectedModel(a);$('surface').textContent=a.config.surface+(a.config.billing_mode?' · '+a.config.billing_mode+' funding':'');$('state').textContent=s.paused?'Paused':s.turn.replaceAll('_',' ');$('reason').textContent=s.reason||'';$('pause').disabled=s.paused;$('continue').disabled=!s.paused;$('audio-status').textContent=s.speaking?'Speech enabled':'Speech disabled';}
function replay(data){
 mcpAgents.clear();mcpReady.clear();
 const previousAgents=new Map(agents);if(session!==data.session)observedTurns.clear();
 cancelAudio(undefined,false);admittedSpeech.clear();clientID=data.client;session=data.session;artifactScroll.reset();agents.clear();skillsByAgent.clear();
 for(const a of data.agents){if(!a.state.paused||a.state.continue_epoch!==previousAgents.get(a.id)?.state.continue_epoch)typing.delete(a.id);agents.set(a.id,a);}
 if(!agents.has(selected))selected=data.agents[0]?.id||'';artifactScroll.selectAgent(selected);
 for(const frame of data.events) apply(frame);for(const frame of data.partials) apply(frame);controls();
}
function connect(){socket=new WebSocket(`${location.protocol==='https:'?'wss':'ws'}://${location.host}/ws`);
 socket.onopen=()=>{$('connection').textContent='Connected · synchronizing';retry=500;};
 socket.onmessage=event=>{let frame;try{frame=JSON.parse(event.data);}catch{return;}if(handleMCPFrame(frame))return;if(frame.type==='replay'){replay(frame.data);$('connection').textContent='Connected';}else apply(frame);};
 socket.onclose=()=>{contextTargetPending=null;$('connection').textContent='Disconnected · reconnecting';settingsReady=false;preferenceControls();cancelAudio(undefined,false);setTimeout(connect,retry);retry=Math.min(retry*2,8000);};
 socket.onerror=()=>socket.close();
}
function busy(agent){return voiceAgents.has(agent)||[...requests.values()].some(r=>r.agent===agent)||speakingNow?.agent===agent||audio.some(t=>t.agent===agent)||[...delayed.values()].some(x=>x.task.agent===agent);}
function updatePause(agent){
 const want=typing.has(agent)||(!!settings.pause_while_speaking&&voiceAgents.has(agent));
 if(want===pauseRequested.has(agent))return;
 if(want)pauseRequested.add(agent);else pauseRequested.delete(agent);
 const cause=typing.has(agent)?'typing':voiceAgents.has(agent)?'speech':'clear';
 speechTranscript.push({ms:Date.now()-speechStartedAt,seq:speechTranscript.length+1,kind:want?'pause':'resume',at:Date.now(),event:want?'pause':'resume',cause,agent_id:agent});
 send(want?'pause':'unpause',{agent_id:agent,automatic:true});
}
function maybeResume(agent){updatePause(agent);}
function explicitContinue(agent){pauseRequested.delete(agent);typing.delete(agent);if(agent===selected)inputOccupied=false;cancelAudio(agent,false);}
function speechRecord(task,state,extra={}){if(['done','cancelled','disabled','unsupported','failed'].includes(state))channelRecord(task,'complete',{status:state});const record={speech_id:task.id,agent_id:task.agent,state,at:Date.now(),text:task.text,reason:task.reason,...extra};speechRecent.push(record);if(speechRecent.length>40)speechRecent.shift();task.speechState=state;}
function enqueueAudio(task){speechRecord(task,'queued');audio.push(task);speakNext();}
function cancelAudio(agent,resume=false){
 for(const [id,r] of requests)if(!agent||r.agent===agent)requests.delete(id);
 if(agent)voiceAgents.delete(agent);else voiceAgents.clear();
 const removed=audio.filter(t=>!agent||t.agent===agent);audio=audio.filter(t=>agent&&t.agent!==agent);
 for(const [id,x] of delayed){if(!agent||x.task.agent===agent){clearTimeout(x.timer);delayed.delete(id);removed.push(x.task);}}
 if(speakingNow&&(!agent||speakingNow.agent===agent)){removed.push(speakingNow);speakingNow=null;window.speechSynthesis?.cancel();}
 for(const t of removed){speechRecord(t,'cancelled');send('speech_done',{agent_id:t.agent,speech_id:t.id,status:'cancelled'});}
 if(resume)for(const id of new Set(removed.map(t=>t.agent)))maybeResume(id);
 speakNext();
}
function speakNext(){
 if(speakingNow||!audio.length)return;
 const task=audio.shift();speakingNow=task;
 const complete=status=>{if(speakingNow!==task)return;speechRecord(task,status);speakingNow=null;send('speech_done',{agent_id:task.agent,speech_id:task.id,status});speakNext();};
 const chunks=speechChunks(task);let index=0,activeUtterance=null;
 const next=()=>{
  if(speakingNow!==task)return;
  if(index>=chunks.length){complete('done');return;}
  const part=chunks[index++],token={};activeUtterance=token;
  const text=(part.source==='thinking'?'Thinking: ':part.source==='error'?'Error: ':'')+part.text;
  // Admit exactly this segment before deciding whether native playback is enabled.
  // A queued or cancelled future segment is never reported as heard.
  channelRecord(task,'utterance',{text,utterance:text,audio:true,source:part.source,part_id:part.part_id||0,turn_id:part.turn_id||task.turn_id||0,evidence:'speech-channel'});
  const valid=()=>speakingNow===task&&activeUtterance===token;
  const ended=()=>{if(!valid())return;activeUtterance=null;speechRecord(task,'chunk-ended',{chunk:index-1});next();};
  if(speechMode==='transcript'||(automaticSpeech(task)&&!settings.tts_enabled)||!window.speechSynthesis||!window.SpeechSynthesisUtterance){setTimeout(ended,0);return;}
  const utterance=new SpeechSynthesisUtterance(text);
  utterance.rate=Math.max(0.5,Math.min(3,settings.tts_speed));
  utterance.onstart=()=>{if(valid()){speechRecord(task,'native-start',{chunk:index-1});speechRecord(task,'speaking',{chunk:index-1});}};
  utterance.onend=ended;
  utterance.onerror=event=>{if(!valid())return;activeUtterance=null;speechRecord(task,'error',{error:event.error});complete('failed');};
  try{speechSynthesis.speak(utterance);}catch(error){if(valid()){activeUtterance=null;speechRecord(task,'error',{error:String(error)});complete('failed');}}
 };next();
}
$('agent-root').onclick=()=>{selected=$('agent-root').dataset.agentId;artifactScroll.selectAgent(selected);controls();};
function showSidebar(open){$('sidebar').hidden=!open;$('sidebar-resizer').hidden=!open;$('menu-toggle').setAttribute('aria-expanded',String(open));}
showSidebar(!window.matchMedia?.('(max-width: 1050px)')?.matches);
$('menu-toggle').onclick=()=>showSidebar($('sidebar').hidden);
function selectTab(name){for(const tab of ['settings','activity']){const active=tab===name;$(`${tab}-tab`).setAttribute('aria-selected',String(active));$(`${tab}-tab`).tabIndex=active?0:-1;$(`${tab}-panel`).hidden=!active;}}
for(const name of ['settings','activity']){
 $(`${name}-tab`).onclick=()=>selectTab(name);
 $(`${name}-tab`).onkeydown=event=>{if(['ArrowLeft','ArrowRight','Home','End'].includes(event.key)){event.preventDefault();const next=event.key==='Home'?'settings':event.key==='End'?'activity':name==='settings'?'activity':'settings';selectTab(next);$(`${next}-tab`).focus();}};
}
// Pane widths are transient UI state. Bounds leave useful room for conversation.
for(const [id,pane,property,minimum,maximum,direction] of [['sidebar-resizer','sidebar','--sidebar-width',220,420,1],['tools-resizer','tools-pane','--tools-width',260,700,-1]]){
 const handle=$(id);
 const resize=width=>{const other=$(pane==='sidebar'?'tools-pane':'sidebar');const available=$('workspace').clientWidth-(other.hidden?0:other.getBoundingClientRect().width)-334;const value=Math.max(minimum,Math.min(maximum,available,width));document.documentElement.style.setProperty(property,`${value}px`);handle.setAttribute('aria-valuenow',String(Math.round($(pane).getBoundingClientRect().width)));};
 handle.onkeydown=event=>{if(!['ArrowLeft','ArrowRight','Home','End'].includes(event.key))return;event.preventDefault();resize(event.key==='Home'?minimum:event.key==='End'?maximum:$(pane).getBoundingClientRect().width+(event.key==='ArrowRight'?20:-20)*direction);};
 handle.onpointerdown=event=>{if(event.button!==0)return;event.preventDefault();const start=event.clientX,width=$(pane).getBoundingClientRect().width;handle.setPointerCapture(event.pointerId);document.body.classList.add('resizing');handle.onpointermove=move=>resize(width+(move.clientX-start)*direction);};
 const stop=()=>{handle.onpointermove=null;document.body.classList.remove('resizing');};handle.onpointerup=stop;handle.onpointercancel=stop;handle.onlostpointercapture=stop;
}
$('pause').onclick=()=>send('pause');$('continue').onclick=()=>{explicitContinue(selected);send('unpause');};$('interrupt').onclick=()=>{cancelAudio(selected,false);send('interrupt');};
$('cancel-audio').onclick=()=>cancelAudio(selected,true);
addEventListener('keydown',event=>{if(event.key==='Escape')cancelAudio(selected,true);});
function submit(type){
 const text=$('input').value;if(!text.trim())return;
 // Preserve the draft and its HUMAN hold when sending is unavailable.
 if(socket?.readyState!==WebSocket.OPEN){send(type,{text});return;}
 $('input').value='';inputRevision++;inputOccupied=false;typing.delete(selected);
 // Automatic release advances the Agent epoch. Send it before new input can
 // create prompt speech; an independent speech hold still prevents release.
 maybeResume(selected);
 if(send(type,{text}))$('errors').textContent='';
}
// Keep the cause: synthetic MCP events must not become human typing during speech.
function draftChanged(event){
 inputRevision++;
 // A human space counts as input. Scripted edits neither acquire nor release
 // HUMAN intent. Submit, clearing the draft, or Continue can release it.
 if(!event.isTrusted)return;
 inputOccupied=$('input').value.length>0;
 if(inputOccupied)typing.add(selected);else typing.delete(selected);
 updatePause(selected);
}
$('input').oninput=draftChanged;
$('composer').onsubmit=e=>{e.preventDefault();submit('input');};$('hint').onclick=()=>submit('hint');$('input').onkeydown=e=>{if(e.key==='Enter'&&(e.ctrlKey||e.metaKey)){e.preventDefault();submit('input');}};
addEventListener('pagehide',()=>cancelAudio(undefined,false));connect();

$("save").onclick=()=>send("save");

// Historical browser MCP tools: on-demand observations, never automatic context.
function nodeIdentity(node){
 if(!nodeIDs.has(node))nodeIDs.set(node,`dom-${++nextNodeID}`);
 const id=nodeIDs.get(node);node.setAttribute('data-ensemble-snapshot-id',id);return id;
}
function visibleNode(node){const style=getComputedStyle(node),rect=node.getBoundingClientRect();return !node.closest('[hidden],[aria-hidden="true"]')&&style.display!=='none'&&style.visibility!=='hidden'&&style.opacity!=='0'&&rect.width>0&&rect.height>0;}
function clipping(value,limit){const text=String(value??'');return {text:text.slice(0,limit),total_characters:text.length,truncated:text.length>limit};}
function guiSnapshot(){
 const all=[...document.body.querySelectorAll('button,input,textarea,select,summary,[role],article,.card,header,h1,h2,h3,pre,.preview,.media')];
 const visible=all.filter(visibleNode),hidden=all.length-visible.length;
 const records=visible.map(node=>{
  const id=nodeIdentity(node),selector=`[data-ensemble-snapshot-id="${id}"]`,tag=node.tagName.toLowerCase();
  const state=[`visible=true`,`disabled=${node.matches(':disabled')||!!node.disabled}`];
  if('checked'in node)state.push(`checked=${!!node.checked}`);
  if('readOnly'in node)state.push(`readonly=${!!node.readOnly}`);
  if('open'in node)state.push(`open=${!!node.open}`);
  for(const key of ['aria-expanded','aria-selected','aria-pressed','aria-checked'])if(node.hasAttribute(key))state.push(`${key}=${JSON.stringify(node.getAttribute(key))}`);
  const part=node.closest('[data-part-id]')?.dataset.partId;if(part)state.push(`part=${JSON.stringify(part)}`);
  const agent=node.closest('[data-agent-id]')?.dataset.agentId;if(agent)state.push(`agent=${JSON.stringify(agent)}`);
  // Read rendered text, not hidden DOM, URLs, blobs, or private replay data.
  let text=String(node.innerText??node.getAttribute('aria-label')??'');
  if(text.startsWith('[opaque'))text='[opaque content withheld]';
  if(node.matches('input,textarea,select'))text=`value=${node.matches('input[type=password],input[type=file]')?'[withheld]':JSON.stringify(node.value)}; text=${text}`;
  return {prefix:`- \`${selector}\` ${tag} ${state.join(' ')}; preview: `,points:[...text]};
 });
 const encoder=new TextEncoder(),bytes=text=>encoder.encode(text).length;
 const total=records.reduce((n,r)=>n+r.points.length,0);let shown=0,characters=0;
 let markdown=`# GUI snapshot\nObserved: ${new Date().toISOString()} · selected agent: ${JSON.stringify(selected)}\n`;
 const footer=(count,chars)=>`\n_Elided visible preview characters: ${total-chars}; omitted artifacts: ${all.length-count}/${all.length} (${hidden} hidden, not read)._\n`;
 for(const r of records){
  // A preview cap keeps a long card from hiding every later control. Both this
  // clipping and whole-record omissions are counted in the same byte budget.
  let low=0,high=Math.min(r.points.length,240),fit=-1,line='';
  while(low<=high){const n=Math.floor((low+high)/2),candidate=r.prefix+JSON.stringify(r.points.slice(0,n).join(''))+(n<r.points.length?' …':'')+'\n';
   if(bytes(markdown+candidate+footer(shown+1,characters+n))<=4096){fit=n;line=candidate;low=n+1;}else high=n-1;
  }
  if(fit<0)continue;markdown+=line;shown++;characters+=fit;
 }
 return markdown+footer(shown,characters);
}
function ttsQueue(){
 const describe=task=>task?{speech_id:task.id,agent_id:task.agent,reason:task.reason,state:task.speechState||'pending',text:clipping(task.text,2000)}:null;
 return {observed_at:Date.now(),client_id:clientID,current:describe(speakingNow),pending:audio.slice(0,80).map(describe),delayed:[...delayed.values()].slice(0,80).map(x=>describe(x.task)),requests:[...requests.entries()].slice(0,80).map(([id,r])=>({id,agent_id:r.agent})),recent:speechRecent.map(item=>({...item,text:clipping(item.text,2000)})),native:{speaking:window.speechSynthesis?.speaking??null,pending:window.speechSynthesis?.pending??null,paused:window.speechSynthesis?.paused??null},omitted:{pending:Math.max(0,audio.length-80),delayed:Math.max(0,delayed.size-80),requests:Math.max(0,requests.size-80)},limits:{recent_events:40,text_characters:2000},note:'Browser-observed queue and callbacks only; native pending is aggregate, not proof of playback.'};
}
const browserTools=[
 {name:'gui_snapshot',description:'Inspect current visible browser DOM, stable node/part identity, visibility and omission limits.',inputSchema:{type:'object',properties:{},additionalProperties:false}},
 {name:'tts_queue',description:'Inspect actual browser speech queue and recent native speech callbacks on demand.',inputSchema:{type:'object',properties:{},additionalProperties:false}},
 {name:'gui_click',description:'Click a currently visible element by CSS selector; inspect again to verify effects.',inputSchema:{type:'object',properties:{selector:{type:'string'}},required:['selector'],additionalProperties:false}},
 {name:'gui_input',description:'Set a visible text input by CSS selector and dispatch input/change events.',inputSchema:{type:'object',properties:{selector:{type:'string'},text:{type:'string'}},required:['selector','text'],additionalProperties:false}}
];
function browserCall(name,args){
 if(!args||typeof args!=='object'||Array.isArray(args))throw new Error('arguments must be an object');
 const definition=browserTools.find(tool=>tool.name===name);if(!definition)throw new Error('unknown browser tool');
 if(Object.keys(args).some(key=>!Object.hasOwn(definition.inputSchema.properties,key)))throw new Error('unknown tool argument');
 if(name==='gui_snapshot')return guiSnapshot();if(name==='tts_queue')return ttsQueue();
 if(typeof args.selector!=='string'||!args.selector)throw new Error('selector is required');
 const matches=[...document.querySelectorAll(args.selector)].filter(visibleNode);if(matches.length!==1)throw new Error(`selector must match exactly one visible element; found ${matches.length}`);
 const node=matches[0];if(node.disabled)throw new Error('element is disabled');
 if(name==='gui_click')node.click();else{if(typeof args.text!=='string'||!node.matches('textarea,input:not([type=password]):not([type=file])')||node.readOnly)throw new Error('editable text input and text are required');node.value=args.text;node.dispatchEvent(new Event('input',{bubbles:true}));node.dispatchEvent(new Event('change',{bubbles:true}));}
 return {node_id:nodeIdentity(node),action:name,status:'dispatched',note:'Inspect a fresh snapshot to verify the result.'};
}
function handleMCP(agent,message){
 if(!message||message.jsonrpc!=='2.0'||typeof message.method!=='string')return;
 const reply=result=>{if(Object.hasOwn(message,'id'))send('jsonrpc',{agent_id:agent,payload:{jsonrpc:'2.0',id:message.id,result}});};
 const error=(code,text)=>{if(Object.hasOwn(message,'id'))send('jsonrpc',{agent_id:agent,payload:{jsonrpc:'2.0',id:message.id,error:{code,message:text}}});};
 if(message.method==='initialize'){mcpAgents.add(agent);mcpReady.delete(agent);reply({protocolVersion:'2024-11-05',capabilities:{tools:{}},serverInfo:{name:'browser-debug',version:'edition-3-ch11'}});return;}
 if(message.method==='notifications/initialized'){mcpReady.add(agent);return;}
 if(!mcpReady.has(agent)){error(-32002,'not initialized');return;}
 if(message.method==='tools/list'){reply({tools:browserTools});return;}
 if(message.method==='ping'){reply({});return;}
 if(message.method!=='tools/call'){error(-32601,'method not found');return;}
 try{const result=browserCall(message.params?.name,message.params?.arguments??{});reply({content:[{type:'text',text:typeof result==='string'?result:JSON.stringify(result)}]});}catch(err){reply({content:[{type:'text',text:err.message}],isError:true});}
}

// The page and socket survive an external-skill unload. A later load opens a
// new protocol session; initialization, not a stale declaration, marks it ready.
function handleMCPFrame(frame){
 if(frame.type==='jsonrpc'){handleMCP(frame.agent_id,frame.payload);return true;}
 if(frame.type==='mcp_closed'){mcpAgents.delete(frame.agent_id);mcpReady.delete(frame.agent_id);return true;}
 return false;
}

$('set-model').addEventListener('click',()=>send('set_model',{text:$('model-name').value.trim()}));
$('agent-status').addEventListener('click',()=>send('agent_status'));
