// Controlled public wire projection inputs, not claimed valid complete logs.
import {setup as predecessor} from '/__base.js';
export function encode(value) {
 if(typeof value==='bigint')return value.toString();
 if(Array.isArray(value))return '['+value.map(encode).join(',')+']';
 if(value&&typeof value==='object')return '{'+Object.entries(value).map(([k,v])=>JSON.stringify(k)+':'+encode(v)).join(',')+'}';
 return JSON.stringify(value);
}
export async function setup() {
 const r=await predecessor();r.frames=[];r.snapshots=[];r.observations=[];r.pending=[];
 for(const page of [r.a,r.b]){
  const snapshot=page.snapshot.bind(page),observation=page.observation.bind(page);
  page.snapshot=s=>{r.snapshots.push(s);snapshot(s)};
  page.observation=o=>{r.observations.push(o);observation(o)};
 }
 r.deliver=(socket,value)=>{const raw=encode(value);r.frames.push(raw);socket.onmessage?.({data:raw});};
 for(const socket of r.sockets){
  const send=socket.send.bind(socket);
  socket.send=raw=>{const value=JSON.parse(raw);if(value.type==='checkpoint'){r.commands.push({...value,socket:socket.ordinal});r.pending.push({socket,value});}else send(raw)};
 }
 r.session=(seq=null,resumed=true)=>({id:'a'.repeat(32),resumed,checkpoint_seq:seq});
 r.snapshot=(session=r.session(),events=[],access=[],options={},socket=r.sockets[0])=>{
  const begin={...socket.begin,generation:'session-'+socket.ordinal,agent_id:'session-agent',watermark:10,omitted:0,first_seq:events.length?events[0].seq:null,last_seq:events.length?events.at(-1).seq:null,log_seq:30,...options,state:{...socket.begin.state,skills:null,session,job_access:access}};
  socket.reviewBegin=begin;
  r.deliver(socket,begin);for(const event of events)r.deliver(socket,{type:'snapshot_event',generation:begin.generation,event});r.deliver(socket,{type:'snapshot_end',generation:begin.generation,watermark:begin.watermark});
 };
 r.observe=(value,revision=11,socket=r.sockets[0])=>r.deliver(socket,{type:'observation',generation:socket.reviewBegin.generation,revision,observation:value});
 r.change=(seq,revision=11,socket=r.sockets[0])=>r.observe({kind:'session_changed',agent_id:socket.reviewBegin.agent_id,session:r.session(seq)},revision,socket);
 r.ack=(index,seq,revision=11)=>{const {socket,value}=r.pending[index];r.deliver(socket,{type:'command_ack',id:value.id,status:'saved',as_of:seq,watch_revision:revision});};
 r.refuse=index=>{const {socket,value}=r.pending[index];r.deliver(socket,{type:'command_error',id:value.id,code:'session_busy',message:'Checkpoint is busy'});};
 r.job=(handle,status,seq=20,call='job-'+String(handle))=>({seq,type:'tool_returned',tool:{call_id:call,name:'run_command',parts:[{type:'text',text:'JOB_'+String(handle)}],is_error:false,job:{handle,status,output:{kind:3,locator:'cr/io/'+String(handle)},bytes:5,exit_code:status==='done'?0:null,is_error:false,reason:status==='killed'?'shutdown':'',cwd:''}}});
 return r;
}
