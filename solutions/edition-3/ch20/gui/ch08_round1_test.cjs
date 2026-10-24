// Focused actual-page regressions. This controlled DOM models range value
// sanitization and captures native constructor inputs, not real voice playback.
const assert=require('node:assert/strict'),fs=require('node:fs'),path=require('node:path'),vm=require('node:vm');
const html=fs.readFileSync(path.join(__dirname,'web/gui/index.html'),'utf8');
const source=fs.readFileSync(path.join(__dirname,'web/gui/app.js'),'utf8');
const stylesheet=fs.readFileSync(path.join(__dirname,'web/gui/style.css'),'utf8');
function page(){
 class Element{
  constructor(tag='div',attrs={}){this.tagName=tag.toUpperCase();this.attrs=attrs;this.type=attrs.type||'';this.children=[];this.hidden=false;this.disabled=false;this.checked=false;this._value='';this.textContent='';this.dataset={};this.listeners={};this.style={setProperty(){}};this.classList={add(){},remove(){}};this.scrollHeight=100;this.clientHeight=100;this.scrollTop=0;this.clientWidth=1400;if(attrs.value)this.value=attrs.value;}
  set value(v){v=String(v);if(this.type==='range'){let n=Math.min(+this.attrs.max,Math.max(+this.attrs.min,Number(v)));if(this.attrs.step!=='any'){const step=+(this.attrs.step||1),min=+this.attrs.min;n=min+Math.round((n-min)/step)*step;}v=String(n);}this._value=v;}
  get value(){return this._value;}
  append(...nodes){for(const node of nodes){node.parentElement=this;this.children.push(node);}}
  replaceChildren(...nodes){this.children=[];this.append(...nodes);}
  setAttribute(k,v){this.attrs[k]=String(v);} getAttribute(k){return this.attrs[k]??null;}
  querySelector(){return null;} addEventListener(type,listener){(this.listeners[type]??=[]).push(listener);} dispatchEvent(event){for(const listener of this.listeners[event.type]||[])listener(event);} focus(){} remove(){if(this.parentElement)this.parentElement.children=this.parentElement.children.filter(x=>x!==this);}
  scrollTo(options){this.lastScroll=options;this.scrollTop=this.scrollHeight;}getBoundingClientRect(){return {width:300};}setPointerCapture(){}
 }
 const ids=new Map();for(const match of html.matchAll(/<([a-z][\w-]*)\b([^>]*\bid="[^"]+"[^>]*)>/gi)){const attrs={};for(const a of match[2].matchAll(/([\w-]+)="([^"]*)"/g))attrs[a[1]]=a[2];ids.set(attrs.id,new Element(match[1],attrs));}
 const sent=[],utterances=[],sockets=[],timers=new Map();let timerID=0,cancels=0;
 class WebSocket{static OPEN=1;constructor(){this.readyState=1;sockets.push(this);}send(raw){sent.push(JSON.parse(raw));}close(){this.readyState=3;this.onclose?.();}}
 class SpeechSynthesisUtterance{constructor(text){this.text=text;}}
 const speechSynthesis={speak(u){utterances.push(u);},cancel(){cancels++;}};
 const matchMedia=()=>({matches:false,addEventListener(){}});
 const document={getElementById:id=>ids.get(id)||null,createElement:t=>new Element(t),documentElement:new Element(),body:new Element()};
 const context=vm.createContext({console,URLSearchParams,document,WebSocket,SpeechSynthesisUtterance,speechSynthesis,window:{SpeechSynthesisUtterance,speechSynthesis,matchMedia},location:{protocol:'http:',host:'fixture.invalid'},addEventListener(){},requestAnimationFrame:f=>f(),setTimeout:(f,ms)=>{timers.set(++timerID,{f,ms});return timerID;},clearTimeout:id=>timers.delete(id)});
 vm.runInContext(source,context,{filename:'actual-app.js'});
 const frame=value=>sockets.at(-1).onmessage({data:JSON.stringify(value)});
 const state=(extra={})=>({turn:'idle',paused:false,speaking:true,epoch:1,continue_epoch:0,...extra});
 const preferences={theme:'dark',tts_enabled:true,pause_while_speaking:true,tts_speed:1,font_size:16,max_tool_rounds:0};
 function settings(patch={},type='settings_changed'){Object.assign(preferences,patch);frame({type,settings:{...preferences}});}
 function replay(){frame({type:'replay',data:{session:'session',client:'owner',agents:[{id:'root-7',config:{model:'fixture-model',vendor:'anthropic',surface:'messages'},state:state()}],events:[],partials:[]}});settings({},'current_settings');}
 replay();
 function claim(id,text='hello',reason='automatic'){
  frame({type:'speech_requested',agent_id:'root-7',data:{id,text,reason,turn_id:7,epoch:1}});
  frame({type:'speech_claimed',agent_id:'root-7',data:{id,text,reason,owner:'owner',turn_id:7,epoch:1}});
 }
 return {ids,sent,utterances,frame,state,settings,replay,claim,channel:context.window.speechChannel,flushQuiet(){for(const [id,t] of timers){if(t.ms===0){timers.delete(id);t.f();return true;}}return false;},cancels:()=>cancels};
}
const tests={
 'R3 follow control reserves space and preserves Chat follow behavior'(){
  // Structural CSS/markup regression, not a browser geometry/visibility proof.
  const rules=[...stylesheet.matchAll(/#follow\s*\{([^}]*)\}/g)];assert.equal(rules.length,1,'one shared rule must apply at all responsive breakpoints');
  const rule=Object.fromEntries(rules[0][1].split(';').filter(Boolean).map(d=>d.split(':').map(v=>v.trim())));
  assert.equal(rule.position,'static','follow control must reserve its own row, not overlay readable content');
  assert.equal(rule.flex,'0 0 auto','follow control must not shrink into neighboring content');
  for(const offset of ['top','right','bottom','left','transform'])assert.equal(rule[offset],undefined,'no positioned offset: '+offset);
  assert.match(html,/<section id="conversation"[\s\S]*?<\/section>\s*<button id="follow" hidden>[^<]*<\/button>\s*<footer>/,'follow stays outside the transcript, before the composer');
  assert.match(stylesheet,/\.pane\{[^}]*display:flex;flex-direction:column/);
  assert.match(stylesheet,/\.card-scroll\{[^}]*flex:1;min-height:0;overflow:auto/);
  assert.match(stylesheet,/\[hidden\]\{display:none!important\}/);
  const p=page(),chat=p.ids.get('conversation'),tools=p.ids.get('tool-cards'),button=p.ids.get('follow');
  chat.scrollHeight=2400;chat.clientHeight=600;chat.scrollTop=300;
  tools.scrollHeight=1800;tools.clientHeight=500;tools.scrollTop=100;
  chat.dispatchEvent({type:'scroll'});assert.equal(button.hidden,false,'away from bottom keeps the control visible');
  p.frame({type:'state_changed',agent_id:'root-7',state:p.state()});
  assert.equal(chat.scrollTop,300,'away-from-bottom updates preserve the reader position');assert.equal(tools.scrollTop,100);
  button.onclick();assert.equal(chat.lastScroll.top,2400);assert.equal(chat.lastScroll.behavior,'smooth');assert.equal(button.hidden,true);assert.equal(tools.scrollTop,100,'jumping Chat must not move Actions');
  chat.scrollTop=1790;chat.dispatchEvent({type:'scroll'});assert.equal(button.hidden,true,'near-bottom scroll hides the control');
  p.frame({type:'state_changed',agent_id:'root-7',state:p.state()});
  assert.equal(chat.scrollTop,2400,'near-bottom updates still follow Chat');assert.equal(tools.scrollTop,100,'independent Actions reader position is retained');
 },
 'fractional controls and reload'(){const p=page();p.settings({tts_speed:1.2345,font_size:18});assert.equal(p.ids.get('tts-speed').value,'1.2345');assert.equal(p.ids.get('font-size').value,'18');p.ids.get('tts-speed').onchange();assert.equal(p.sent.at(-1).settings.tts_speed,1.2345);p.replay();assert.equal(p.ids.get('tts-speed').value,'1.2345');assert.equal(p.ids.get('font-size').value,'18');assert.equal(p.ids.get('tts-speed-value').textContent,'1.2345×');},
 'native rate clamps each chunk but not saved preference'(){const p=page();p.settings({tts_speed:2.5});p.claim('long','x'.repeat(2500));assert.equal(p.utterances[0].rate,2.5);p.settings({tts_speed:5});p.utterances[0].onend();assert.equal(p.utterances[1].rate,3);assert.equal(p.ids.get('tts-speed').value,'5');p.settings({tts_speed:.5});p.utterances[1].onend();assert.equal(p.utterances[2].rate,.5);p.utterances[2].onend();assert.equal(p.sent.filter(x=>x.type==='speech_done'&&x.speech_id==='long').length,1);p.settings({tts_enabled:false,tts_speed:1.2345});p.claim('explicit','user selected speech','on demand');assert.equal(p.utterances.at(-1).text,'user selected speech');assert.equal(p.utterances.at(-1).rate,1.2345);},
 'Read Aloud preserves HUMAN hold until the draft is cleared'(){const p=page(),input=p.ids.get('input');input.value='preserve this draft';input.oninput({isTrusted:true});p.frame({type:'state_changed',agent_id:'root-7',state:p.state({paused:true})});const releases=p.sent.filter(x=>x.type==='unpause').length,pauses=p.sent.filter(x=>x.type==='pause').length;p.settings({tts_enabled:false});p.settings({tts_enabled:true});assert.equal(p.sent.filter(x=>x.type==='unpause').length,releases);assert.equal(input.value,'preserve this draft');input.value+=' edited';input.oninput({isTrusted:true});assert.equal(p.sent.filter(x=>x.type==='pause').length,pauses);input.value='';input.oninput({isTrusted:true});assert.equal(p.sent.filter(x=>x.type==='unpause').length,releases+1);assert.equal(p.cancels(),0);},
 'automatic cancellation rejects stale callbacks'(){const p=page();p.claim('old');const old=p.utterances.at(-1);p.frame({type:'state_changed',agent_id:'root-7',state:p.state({paused:true})});p.settings({tts_enabled:false});p.frame({type:'speech_idle',agent_id:'root-7',data:{owner:'owner',epoch:1}});p.settings({tts_enabled:true});p.frame({type:'state_changed',agent_id:'root-7',state:p.state({paused:true,epoch:2})});const commands=p.sent.length;old.onend();old.onerror();p.frame({type:'speech_idle',agent_id:'root-7',data:{owner:'owner',epoch:1}});assert.equal(p.sent.length,commands,'stale callbacks and idle cannot release a new epoch');assert.equal(p.utterances.length,1);assert.equal(p.sent.filter(x=>x.type==='speech_done'&&x.speech_id==='old').length,1);assert.ok(p.sent.some(x=>x.type==='unpause'));},
 'R2 delayed automatic acknowledgement preserves fresh typing'(){
  const p=page(),input=p.ids.get('input');input.value='old draft';input.oninput({isTrusted:true});
  p.frame({type:'state_changed',agent_id:'root-7',state:p.state({paused:true})});
  p.settings({tts_enabled:false});const releases=p.sent.filter(x=>x.type==='unpause').length;
  assert.ok(p.sent.filter(x=>x.type==='unpause').every(x=>x.automatic===true));
  input.value='fresh edit before acknowledgement';input.oninput({isTrusted:true});
  p.frame({type:'state_changed',agent_id:'root-7',state:p.state({paused:false,epoch:2,continue_epoch:1})});
  p.frame({type:'state_changed',agent_id:'root-7',state:p.state({paused:true,epoch:2,continue_epoch:1})});
  p.frame({type:'speech_idle',agent_id:'root-7',data:{owner:'owner',epoch:2}});
  assert.equal(p.sent.filter(x=>x.type==='unpause').length,releases,'older automatic ack and later idle must not release fresh typing');
  assert.equal(input.value,'fresh edit before acknowledgement');
  input.value='';input.oninput({isTrusted:true});assert.equal(p.sent.filter(x=>x.type==='unpause').length,releases+1,'clearing the fresh draft still releases its pause');
  const pauses=p.sent.filter(x=>x.type==='pause').length;input.value='future draft';input.oninput({isTrusted:true});assert.equal(p.sent.filter(x=>x.type==='pause').length,pauses+1);
 },
 'R2 explicit Continue authority precedes its acknowledgement'(){
  const p=page(),input=p.ids.get('input');input.value='draft kept';input.oninput({isTrusted:true});p.claim('speaking');const old=p.utterances[0];
  p.frame({type:'state_changed',agent_id:'root-7',state:p.state({paused:true})});
  p.frame({type:'continued',agent_id:'root-7',data:{client:'another-tab',epoch:1}});
  assert.equal(p.cancels(),1,'another tab Continue immediately cancels owned speech');assert.equal(input.value,'draft kept');
  let pauses=p.sent.filter(x=>x.type==='pause').length;input.value+=' fresh';input.oninput({isTrusted:true});assert.equal(p.sent.filter(x=>x.type==='pause').length,pauses+1);
  p.frame({type:'state_changed',agent_id:'root-7',state:p.state({paused:false,epoch:2,continue_epoch:1})});
  p.frame({type:'state_changed',agent_id:'root-7',state:p.state({paused:true,epoch:2,continue_epoch:1})});
  let releases=p.sent.filter(x=>x.type==='unpause').length;old.onend();old.onerror();p.frame({type:'speech_idle',agent_id:'root-7',data:{owner:'owner',epoch:2}});
  assert.equal(p.sent.filter(x=>x.type==='unpause').length,releases);assert.equal(p.sent.filter(x=>x.type==='speech_done'&&x.speech_id==='speaking').length,1);
  p.ids.get('continue').onclick();assert.notEqual(p.sent.at(-1).automatic,true,'explicit Continue must not be tagged automatic');
  pauses=p.sent.filter(x=>x.type==='pause').length;input.value+=' after own Continue';input.oninput({isTrusted:true});assert.equal(p.sent.filter(x=>x.type==='pause').length,pauses+1);
  p.frame({type:'continued',agent_id:'root-7',data:{client:'owner',epoch:2}});
  p.frame({type:'state_changed',agent_id:'root-7',state:p.state({paused:true,epoch:3,continue_epoch:2})});
  p.frame({type:'continued',agent_id:'root-7',data:{client:'another-tab',epoch:1}});
  p.frame({type:'continued',agent_id:'another-agent',data:{client:'another-tab',epoch:3}});
  p.frame({type:'continued',agent_id:'root-7',replay:true,data:{client:'another-tab',epoch:3}});
  releases=p.sent.filter(x=>x.type==='unpause').length;p.frame({type:'speech_idle',agent_id:'root-7',data:{owner:'owner',epoch:3}});
  assert.equal(p.sent.filter(x=>x.type==='unpause').length,releases,'own echo, old epoch, other Agent, replay and delayed ack must not erase fresh input');
 },
 'R2 production explicit Speak works with Read Aloud off'(){
  const p=page();p.settings({tts_enabled:false});p.claim('explicit-off','selected complete text','on demand');
  assert.equal(p.utterances.length,1,'production on demand must reach the native constructor');assert.equal(p.utterances[0].text,'selected complete text');
  p.claim('automatic-off','must not play','automatic');p.utterances[0].onend();
  assert.equal(p.utterances.length,1);assert.ok(p.channel.read().events.some(x=>x.event==='utterance'&&x.speech_id==='automatic-off'));assert.ok(p.flushQuiet());assert.ok(p.sent.some(x=>x.type==='speech_done'&&x.speech_id==='automatic-off'&&x.status==='done'));
  p.settings({tts_enabled:true});p.claim('automatic-on','cancel this','automatic');const old=p.utterances.at(-1);p.settings({tts_enabled:false});old.onend();old.onerror();
  assert.equal(p.sent.filter(x=>x.type==='speech_done'&&x.speech_id==='automatic-on').length,1);assert.equal(p.utterances.length,2);
 },
 'single root tree reflects replay and live activity'(){const p=page();const tree=p.ids.get('agent-tree'),node=p.ids.get('agent-root');assert.ok(tree,'actual page needs an Agent tree');assert.equal(tree.getAttribute('role'),'tree');assert.equal(node.getAttribute('role'),'treeitem');assert.equal((html.match(/role="treeitem"/g)||[]).length,1);assert.equal(p.ids.get('agent-name').textContent,'Agent root-7');assert.equal(p.ids.get('agent-activity').textContent,'Idle');p.frame({type:'state_changed',agent_id:'root-7',state:p.state({turn:'waiting_for_model'})});assert.equal(p.ids.get('agent-activity').textContent,'Waiting for model');assert.equal(p.ids.get('agent-cue').dataset.busy,'true');p.frame({type:'state_changed',agent_id:'root-7',state:p.state({turn:'waiting_for_tools',paused:true})});assert.equal(p.ids.get('agent-activity').textContent,'Paused');node.onclick();assert.equal(node.getAttribute('aria-selected'),'true');p.replay();assert.equal(p.ids.get('agent-name').textContent,'Agent root-7');assert.equal(p.ids.get('agent-activity').textContent,'Idle');p.ids.get('activity-tab').onclick();assert.equal(p.ids.get('activity-panel').hidden,false);p.ids.get('settings-tab').onclick();assert.equal(p.ids.get('settings-panel').hidden,false);}
};
let failed=0,ran=0;for(const [name,test] of Object.entries(tests)){if(process.argv[2]&&!name.startsWith(process.argv[2]))continue;ran++;try{test();console.log(`PASS ${name}`);}catch(err){failed++;console.error(`FAIL ${name}: ${err.stack}`);}}assert.ok(ran,'test filter must select checks');process.exitCode=failed?1:0;
