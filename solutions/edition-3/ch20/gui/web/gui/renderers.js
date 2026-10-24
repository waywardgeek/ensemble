// The full result is retained; only its initial preview is shortened.
export function renderToolResult(part, open=false) {
 const node=document.createElement('div'), text=part.text||'';
 const preview=document.createElement('div'); preview.className='body preview';
 preview.textContent=text.length>400?text.slice(0,400)+'…':text;
 const details=document.createElement('details');details.open=!!open;
 const summary=document.createElement('summary');summary.textContent=`Full result · ${text.length.toLocaleString()} characters`;
 const full=document.createElement('pre');full.className='body';full.textContent=text;
 details.append(summary,full);node.append(preview,details);return node;
}

// Render one public Wire frame into a caller-owned node. No application globals.
const element = (tag, cls, text) => { const e = document.createElement(tag); if(cls) e.className=cls; if(text !== undefined) e.textContent=String(text); return e; };
function fullDetails(title,child,open){const details=element('details');details.open=!!open;details.append(element('summary','',title),child);return details;}
function toolContent(c){
 const body=element('div');
 if(c.name==='edit_file'&&c.args&&typeof c.args.old_text==='string'&&typeof c.args.new_text==='string'){
  body.append(element('div','muted',c.args.path||''));
  const diff=element('pre','diff');for(const line of c.args.old_text.split('\n'))diff.append(element('div','deletion','− '+line));for(const line of c.args.new_text.split('\n'))diff.append(element('div','insertion','+ '+line));body.append(diff);
 }
 body.append(element('pre','code',c.partial?c.text:JSON.stringify(c.args??{},null,2)));return body;
}
function jobInfo(job){const box=element('div','job-state');let text=`Job ${job.handle} · ${job.status} · ${job.bytes} bytes`;if(job.exit_code!==undefined)text+=` · exit ${job.exit_code}`;if(job.reason)text+=` · ${job.reason}`;box.append(element('div','',text));if(job.output?.locator)box.append(element('div','muted',job.output.locator));return box;}
export function renderPart(record, {onSpeak} = {}) {
 const {frame, node} = record, c=frame.part, wasOpen=node.querySelector('details')?.open;
 node.dataset.agentId=frame.agent_id;node.dataset.partId=String(frame.part_id);node.id=`part-${frame.agent_id}-${frame.part_id}`;
 node.replaceChildren();node.className=`card ${c.type} ${c.role==='human'?'human':''} ${c.partial?'partial':''} ${c.is_error?'error':''}`;
 const heading=element('div','card-title');
 const names={text:c.role==='human'?'You':c.role==='system'?'System':'Assistant',thinking:'Thinking',reasoning_summary:'Reasoning summary',continuation:'Private continuation',tool_call:`Tool request · ${c.name||'pending'}`,tool_result:`Tool result · ${c.name||c.call_id||''}`,image:'Image',audio:'Audio',video:'Video',file:'File'};
 heading.append(element('span','',names[c.type]||c.type));
 if(c.partial) heading.append(element('span','','Streaming…'));
 if(c.type==='tool_call'&&frame.type!=='job_update') heading.append(element('span',c.dispatched?'badge':'',c.dispatched?'Dispatched':'Requested · not dispatched'));
 if(c.is_error) heading.append(element('span','','Error'));
 if(c.from) heading.append(element('span','muted',`${c.from.vendor} / ${c.from.model}`));
 if(onSpeak && !c.partial && c.text && ['text','thinking','reasoning_summary','tool_call','tool_result'].includes(c.type) && !c.text.startsWith('[opaque')) {
  const speak=element('button','','Speak');speak.title='Speak the complete readable part';speak.onclick=()=>onSpeak(frame.agent_id,frame.part_id);heading.append(speak);
 }
 node.append(heading);
 if(c.type==='tool_call') {if(c.partial)node.append(toolContent(c));else {node.append(element('div','preview',`${c.name} · ${JSON.stringify(c.args??{}).slice(0,160)}`));node.append(fullDetails('Full request / edit diff',toolContent(c),wasOpen));}if(c.call_id)node.append(element('div','muted',`Provider call: ${c.call_id}`));}
 else if(['thinking','reasoning_summary'].includes(c.type)&&!c.partial){node.append(element('div','preview',(c.text||'').slice(0,180)));node.append(fullDetails(c.type==='reasoning_summary'?'Full reasoning summary':'Full thinking',element('div','body',c.text||''),wasOpen));}
 else if(c.type==='tool_result') {
  node.append(renderToolResult(c,wasOpen));
 } else if(['image','audio','video','file'].includes(c.type)) {
  node.append(element('div','media',c.text||`${c.type} attachment · ${c.mime||'unknown type'}`));
  if(c.locator) {let safe=false;try { const u=new URL(c.locator);safe=['http:','https:'].includes(u.protocol); }catch{}
   if(safe) {const a=element('a','',c.locator);a.href=c.locator;a.target='_blank';a.rel='noopener noreferrer';node.append(a);}else node.append(element('div','muted',c.locator)); }
 } else node.append(element('div','body',c.text||''));
 if(c.job)node.append(jobInfo(c.job));
 return node;
}

// Override entries directly or pass a renderer map to ArtifactScroll. Each entry
// owns rendering, updating, and removing its node for a complete public frame.
function defaultRenderer(){return {
 renderPart(frame,options){return renderPart({frame,node:document.createElement('article')},options);},
 updatePart(node,frame,options){renderPart({frame,node},options);},
 removePart(node){node.remove();}
};}
export const renderers=Object.fromEntries(['text','thinking','reasoning_summary','continuation','tool_call','tool_result','image','audio','video','file'].map(kind=>[kind,defaultRenderer()]));
