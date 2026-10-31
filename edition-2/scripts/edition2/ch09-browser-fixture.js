// Browser-only wire projection fixtures. Large identities are counter fixtures,
// not fabricated complete histories purporting to allocate 2^53 activations.
import {setup as predecessor} from '/__base.js';
export function encode(value) {
 if(typeof value==='bigint')return value.toString();
 if(Array.isArray(value))return '['+value.map(encode).join(',')+']';
 if(value&&typeof value==='object')return '{'+Object.entries(value).map(([k,v])=>JSON.stringify(k)+':'+encode(v)).join(',')+'}';
 return JSON.stringify(value);
}
export async function setup() {
 const r=await predecessor();r.frames=[];r.snapshots=[];r.observations=[];
 for(const p of [r.a,r.b]){
  const snapshot=p.snapshot.bind(p),observation=p.observation.bind(p);
  p.snapshot=s=>{r.snapshots.push(s);snapshot(s)};
  p.observation=o=>{r.observations.push(o);observation(o)};
 }
 r.deliver=(socket,value)=>{const raw=encode(value);r.frames.push(raw);socket.onmessage?.({data:raw});};
 r.identity=(large=false)=>large?[9007199254740992n,9007199254740993n]:[2,3];
 r.body='<img src=x onerror="window.ch09Executed=true"> ALPHA_MARK '+('retained '.repeat(200))+' END_ALPHA';
 const digest=async text=>[...new Uint8Array(await crypto.subtle.digest('SHA-256',new TextEncoder().encode(text)))].map(v=>v.toString(16).padStart(2,'0')).join('');
 const alphaHash=await digest(r.body),betaHash=await digest('BETA_MARK');
 r.state=(large=false,retired=false)=>{
  const [a,b]=r.identity(large),revision=large?18446744073709551615n:3;
  return {revision,primary:'base',roots:retired?['beta']:['alpha','beta'],active:[...(retired?[]:[{name:'alpha',type:'loadable',activation:a}]),{name:'base',type:'primary',activation:1},{name:'beta',type:'loadable',activation:b}],available:retired?[{name:'alpha',description:'Alpha'}]:[],tools:['load_skill','read_file','unload_skill'],retired:retired?[{name:'alpha',activation:a}]:[]};
 };
 r.event=(large=false)=>{const [a,b]=r.identity(large);return {seq:25,type:'skills_changed',skills:{action:'load',name:'beta',state:r.state(large),activated:[
  {activation:a,name:'alpha',type:'loadable',body:r.body,sha256:alphaHash,tools:[],dependencies:[],offers:[]},
  {activation:b,name:'beta',type:'loadable',body:'BETA_MARK',sha256:betaHash,tools:[],dependencies:[a],offers:[]} ]}};};
 // Safe projection tests do not ask the browser to validate full reducer history.
 r.snapshot=(socket=r.sockets[0],large=false,events=[r.event(large)],state=r.state(large))=>{
  const begin={...socket.begin,generation:'skill-'+socket.ordinal,watermark:10,omitted:125,first_seq:events.length?events[0].seq:null,last_seq:events.length?events.at(-1).seq:null,log_seq:125,state:{...socket.begin.state,skills:state}};
  r.deliver(socket,begin);for(const event of events)r.deliver(socket,{type:'snapshot_event',generation:begin.generation,event});r.deliver(socket,{type:'snapshot_end',generation:begin.generation,watermark:10});
 };
 r.observe=(value,revision=11,socket=r.sockets[0])=>r.deliver(socket,{type:'observation',generation:'skill-'+socket.ordinal,revision,observation:value});
 return r;
}
