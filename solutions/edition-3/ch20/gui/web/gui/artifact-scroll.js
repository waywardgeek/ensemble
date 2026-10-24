import {renderers as defaultRenderers} from './renderers.js';

// Reusable public-frame consumer. Renderers receive the complete Wire frame so
// Agent/part identity and metadata are available without application globals.
export class ArtifactScroll {
 constructor(container,{renderers=defaultRenderers,toolContainer,followButton,threshold=100,onSpeak}={}) {
  this.container=container;this.toolContainer=toolContainer||container;this.renderers=renderers;
  this.followButton=followButton;this.threshold=threshold;this.onSpeak=onSpeak;
  this.records=new Map();this.ended=new Map();this.selected='';
  this.onScroll=()=>{if(this.followButton)this.followButton.hidden=this.nearBottom();};
  this.onFollow=()=>{this.container.scrollTo({top:this.container.scrollHeight,behavior:'smooth'});if(this.followButton)this.followButton.hidden=true;};
  container.addEventListener('scroll',this.onScroll,{passive:true});followButton?.addEventListener('click',this.onFollow);
 }
 nearBottom(container=this.container){return container.scrollHeight-container.clientHeight-container.scrollTop<this.threshold;}
 follow(wasNear,container=this.container){if(wasNear)requestAnimationFrame(()=>{container.scrollTop=container.scrollHeight;});else if(container===this.container&&this.followButton)this.followButton.hidden=false;}
 selectAgent(id=''){this.selected=id;for(const r of this.records.values())r.node.hidden=!!id&&r.frame.agent_id!==id;}
 remove(key){const r=this.records.get(key);if(r){r.renderer.removePart(r.node);this.records.delete(key);}}
 reset(){for(const key of this.records.keys())this.remove(key);this.ended.clear();}
 destroy(){this.reset();this.container.removeEventListener('scroll',this.onScroll);this.followButton?.removeEventListener('click',this.onFollow);}
 receive(frame){
  if(!frame||frame.stale)return false;
  if(frame.type==='replay') {this.reset();for(const f of frame.data?.events||[])this.receive(f);for(const f of frame.data?.partials||[])this.receive(f);return true;}
  const key=`${frame.agent_id}/${frame.part_id}`,prior=this.records.get(key);
  if(frame.type==='turn_ended'||frame.type==='turn_interrupted') {
   if(frame.turn_id)this.ended.set(frame.agent_id,Math.max(this.ended.get(frame.agent_id)||0,frame.turn_id));
   for(const [id,r] of this.records)if(r.frame.agent_id===frame.agent_id&&r.frame.part.partial&&(!frame.turn_id||r.frame.turn_id===frame.turn_id))this.remove(id);
   return true;
  }
  if(frame.type==='part_removed'){if(prior?.frame.part.partial)this.remove(key);return true;}
  if(!['part_delta','part_final','tool_dispatched','tool_finished','job_update'].includes(frame.type))return false;
  let part=frame.part?{...frame.part}:undefined;
  if(frame.type==='part_delta') {
   if(prior&&!prior.frame.part.partial)return false;
   if(frame.turn_id&&frame.turn_id<=(this.ended.get(frame.agent_id)||0))return false;
   part={type:frame.kind,name:frame.name,role:'agent',partial:true,text:(prior?.frame.part.text||'')+(frame.chunk||'')};
  } else if(frame.type==='tool_dispatched') {
   part={...prior?.frame.part,type:'tool_call',name:frame.name,call_id:frame.call_id,args:frame.input,...part,partial:false,dispatched:true};
  } else if(frame.type==='tool_finished') {
   part={type:'tool_result',name:frame.name,call_id:frame.call_id,text:frame.result||'',is_error:frame.is_error,...part,partial:false};
  } else if(frame.type==='job_update') {
   part={type:'tool_result',text:'',...prior?.frame.part,...part};
  }
  if(!part)return false;
  part.dispatched=!!(part.dispatched||prior?.frame.part.dispatched);
  const normalized={...prior?.frame,...frame,part};
  const renderer=this.renderers[part.type]||this.renderers.text;
  if(!renderer)throw Error(`No renderer for ${part.type}`);
  const target=['tool_call','tool_result'].includes(part.type)?this.toolContainer:this.container;
  const near=this.nearBottom(target),options={onSpeak:this.onSpeak};
  let record=prior;
  if(record&&record.renderer!==renderer){this.remove(key);record=undefined;}
  if(!record){record={frame:normalized,renderer,node:renderer.renderPart(normalized,options)};this.records.set(key,record);}
  else {record.frame=normalized;renderer.updatePart(record.node,normalized,options);}
  record.node.hidden=!!this.selected&&frame.agent_id!==this.selected;
  if(record.node.parentElement!==target){target.querySelector('.empty-state')?.remove();target.append(record.node);}
  this.follow(near,target);return true;
 }
}

// Retained factory spelling; it now returns the complete frame consumer.
export function createArtifactScroll(container,options){return new ArtifactScroll(container,options);}
