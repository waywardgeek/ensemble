// Independent local wire/native fixtures for actual browser components. These
// are not a GUI-server implementation or audible/live-provider evidence.
export async function setup({deferEnd=false}={}) {
 const sockets=[],commands=[],spoken=[],speechEvents=[];
 let revision=0,preferences={theme:'dark',font_size:16,sidebar_width:260,actions_width:380,autoplay:false,speech_rate:1},active=null,cancels=0;
 let watchRevision=0,policy={revision:0,persistent:true,max_model_requests:0,effective_max_model_requests:16};
 const tick=()=>new Promise(resolve=>setTimeout(resolve,0));
 const until=async(predicate,reason)=>{const end=performance.now()+2000;while(performance.now()<end){if(predicate())return;await tick();}throw Error(reason);};
 const snapshot=()=>({revision,preferences:{...preferences}});
 class WS {
  static OPEN=1;
  constructor(){this.readyState=1;this.ordinal=sockets.length;this.causes={typing:false,speaking:false};sockets.push(this);}
  frame(value){this.onmessage?.({data:JSON.stringify(value)});}
  send(raw){const m=JSON.parse(raw);commands.push({...m,socket:this.ordinal});queueMicrotask(()=>{
   if(this.readyState!==1)return;
   if(m.type==='subscribe'){
    this.subscribed=true;this.frame({type:'preferences_snapshot',...snapshot()});
    this.begin={type:'snapshot_begin',id:m.id,generation:'g'+this.ordinal,agent_id:'a1',watermark:watchRevision,omitted:0,first_seq:null,last_seq:null,log_seq:0,state:{lifecycle:'idle',active_request_id:null,active_operation:null,queued_request_ids:[],usage:[],paused:false,typing_clients:0,speaking_clients:0,active_max_model_requests:null,execution_policy:{...policy}}};
    this.frame(this.begin);if(!deferEnd)this.finish();return;
   }
   if(m.type==='pause'){
    this.causes={typing:m.typing,speaking:m.speaking};const peers=sockets.filter(s=>s.readyState===1),typing=peers.filter(s=>s.causes.typing).length,speaking=peers.filter(s=>s.causes.speaking).length;
    this.frame({type:'ack',id:m.id,revision:0,paused:!!(typing||speaking),typing_clients:typing,speaking_clients:speaking});return;
   }
   if(m.type==='preferences_update'){
    if(m.base_revision!==revision){this.frame({type:'error',id:m.id,code:'revision_conflict',message:'preferences changed',domain:'preferences',current:snapshot()});return;}
    const proposed={...preferences,...m.patch};
    if(!Number.isInteger(proposed.font_size)||proposed.font_size<12||proposed.font_size>28){this.frame({type:'error',id:m.id,code:'invalid_preferences',message:'font_size must be an integer between 12 and 28'});return;}
    if(JSON.stringify(proposed)!==JSON.stringify(preferences)){preferences=proposed;revision++;for(const s of sockets)if(s.readyState===1&&s.subscribed)s.frame({type:'preferences_changed',...snapshot()});}
    this.frame({type:'preferences_ack',id:m.id,revision});return;
   }
   if(m.type==='policy_update'){
    if(m.base_revision!==policy.revision){this.frame({type:'error',id:m.id,code:'revision_conflict',message:'policy changed',domain:'policy',current:{...policy}});return;}
    const value=m.patch.max_model_requests;
    if(!Number.isInteger(value)||value<0||value>256){this.frame({type:'error',id:m.id,code:'invalid_policy',message:'invalid request limit'});return;}
    if(value!==policy.max_model_requests){policy={...policy,revision:policy.revision+1,max_model_requests:value,effective_max_model_requests:value||16};watchRevision++;for(const s of sockets)if(s.readyState===1&&s.subscribed)s.frame({type:'observation',generation:s.begin.generation,revision:watchRevision,observation:{kind:'policy_changed',agent_id:'a1',execution_policy:{...policy}}});}
    this.frame({type:'policy_ack',id:m.id,revision:policy.revision,watch_revision:watchRevision});return;
   }
   if(m.type==='prompt')this.frame({type:'accepted',id:m.id,request_id:'r-local'});
  });}
  finish(){this.frame({type:'snapshot_end',generation:this.begin.generation,watermark:this.begin.watermark});}
  close(){if(this.readyState!==1)return;this.readyState=3;this.causes={typing:false,speaking:false};this.onclose?.();}
 }
 window.WebSocket=WS;
 const native={speak(u){if(active)throw Error('native overlap');active=u;spoken.push(u);u.onstart?.();},cancel(){cancels++;const old=active;active=null;old?.onerror?.({error:'interrupted'});}};
 class Utterance{constructor(text){this.text=text;this.rate=1;}}
 const {BrowserApplication}=await import('/application.js');
 const app=new BrowserApplication(native,Utterance);
 const rootA=document.querySelector('main'),rootB=rootA.cloneNode(true);rootA.id='review-a';rootB.id='review-b';
 for(const el of rootB.querySelectorAll('[id]'))el.id='second-'+el.id;
 for(const el of rootB.querySelectorAll('[for]'))el.htmlFor='second-'+el.htmlFor;
 for(const el of rootB.querySelectorAll('[aria-controls]'))el.setAttribute('aria-controls','second-'+el.getAttribute('aria-controls'));
 document.body.append(rootB);
 for(const [root,label] of [[rootA,'A'],[rootB,'B']])root.addEventListener('ensemble-speech',e=>speechEvents.push({page:label,...e.detail}));
 const a=app.createPage(rootA,'ws://controlled'),b=app.createPage(rootB,'ws://controlled');
 sockets.forEach(s=>s.onopen());await tick();await tick();
 const remote=async patch=>{preferences={...preferences,...patch};revision++;for(const s of sockets)if(s.readyState===1&&s.subscribed)s.frame({type:'preferences_changed',...snapshot()});await tick();};
 const end=async()=>{const old=active;if(!old)throw Error('positive native utterance missing before controlled end');active=null;old.onend?.();await until(()=>active||(!a.speech.busy()&&!b.speech.busy()),'next native admission or queue settlement absent');};
 return {a,b,app,sockets,commands,spoken,speechEvents,tick,until,remote,end,snapshot,get active(){return active},get cancels(){return cancels}};
}
