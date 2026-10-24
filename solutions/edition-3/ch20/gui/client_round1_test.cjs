'use strict';
// Controlled DOM/native-callback regression: does not launch a browser or audio.
const assert=require('node:assert/strict'),fs=require('node:fs'),vm=require('node:vm');
class Element {
 constructor(tag){this.tagName=tag;this.children=[];this._text='';this.value='';this.checked=false;this.hidden=false;this.open=false;this.dataset={};this.style={setProperty(){}};}
 addEventListener(){}
 set textContent(v){this._text=String(v);this.children=[];}get textContent(){return this._text+this.children.map(c=>c.textContent).join('');}
 set innerHTML(_){throw Error('Runtime HTML injection is forbidden');}
 get parentElement(){return this.parent||null;}
 append(...children){for(const child of children){child.remove();child.parent=this;this.children.push(child);}}
 replaceChildren(...children){this._text='';this.children=[];this.append(...children);}
 remove(){if(this.parent)this.parent.children=this.parent.children.filter(c=>c!==this);this.parent=null;}
 setAttribute(name,value){this[name]=value;}
 querySelector(selector){return this.querySelectorAll(selector)[0]||null;}
 querySelectorAll(selector){const matches=e=>selector[0]==='.'?(e.className||'').split(' ').includes(selector.slice(1)):e.tagName===selector;return this.children.flatMap(c=>[...(matches(c)?[c]:[]),...c.querySelectorAll(selector)]);}
}
// This synchronous fixture supports exactly the shipped three-file graph.
// Evaluate real source in one DOM realm, with separate renderer/component scopes;
// unfamiliar imports/exports must fail rather than be discarded by an adapter.
function productionSource(file, expectedImport, replacement, expectedExports){
 let source=fs.readFileSync(__dirname+'/web/gui/'+file,'utf8');
 assert.deepEqual(source.match(/^import[^\n]*/gm)||[],expectedImport?[expectedImport]:[],file+' import graph changed');
 assert.deepEqual(source.match(/^export (?:function|class|const) [A-Za-z_$][\w$]*/gm)||[],expectedExports,file+' exports changed');
 if(expectedImport)source=source.replace(expectedImport,replacement);
 for(const form of expectedExports)source=source.replace(new RegExp('^'+form+'\\b','m'),form.slice(7));
 assert(!/^\s*(?:import|export)\b/m.test(source)&&! /\bimport\s*[.(]/.test(source),file+' has unsupported module syntax');
 return source;
}
function loadProduction(context){
 const renderers=productionSource('renderers.js',null,'',['export function renderToolResult','export function renderPart','export const renderers']);
 const graph={renderers:vm.runInContext('(()=>{\n"use strict";\n'+renderers+'\nreturn {renderToolResult,renderPart,renderers};\n})()',context,{filename:'web/gui/renderers.js'})};
 context.__fixtureModules=graph;
 const artifact=productionSource('artifact-scroll.js',"import {renderers as defaultRenderers} from './renderers.js';",'const defaultRenderers=__fixtureModules.renderers.renderers;',['export class ArtifactScroll','export function createArtifactScroll']);
 graph.artifact=vm.runInContext('(()=>{\n"use strict";\n'+artifact+'\nreturn {ArtifactScroll,createArtifactScroll};\n})()',context,{filename:'web/gui/artifact-scroll.js'});
 const app=productionSource('app.js',"import {ArtifactScroll} from './artifact-scroll.js';",'const {ArtifactScroll}=__fixtureModules.artifact;',[]);
 vm.runInContext('"use strict";\n'+app,context,{filename:'web/gui/app.js'});
 delete context.__fixtureModules;
}
function fixture(){
 const nodes=new Map(),sent=[],utterances=[],timers=[];let cancelled=0;
 class WebSocket {static OPEN=1;constructor(){this.readyState=1;}send(raw){sent.push(JSON.parse(raw));}close(){this.readyState=3;}}
 class SpeechSynthesisUtterance{constructor(text){this.text=text;}}
 const speechSynthesis={speak:u=>utterances.push(u),cancel:()=>cancelled++};
 const context={console,URL,URLSearchParams,WebSocket,SpeechSynthesisUtterance,speechSynthesis,document:{getElementById:id=>{if(id==='tool-cards')return null;if(!nodes.has(id))nodes.set(id,new Element('div'));return nodes.get(id);},createElement:tag=>new Element(tag),documentElement:{scrollHeight:0,dataset:{},style:{setProperty(){}}}},location:{protocol:'http:',host:'local.test'},innerHeight:1000,scrollY:0,requestAnimationFrame:fn=>fn(),scrollTo:()=>{},addEventListener:()=>{},setTimeout:(fn,ms)=>{const t={fn,ms,cancelled:false};timers.push(t);return t;},clearTimeout:t=>{if(t)t.cancelled=true;}};
 // This fixture's existing card assertions use a single conversation pane;
 // the optional separate tool pane is absent, so the real component falls back.
 context.window=context;vm.createContext(context);loadProduction(context);
 const api=vm.runInContext('({apply,replay,$,maybeResume,enqueueAudio,cancelAudio,channel:window.speechChannel})',context);
 const state=(epoch=1,continued=0,speaking=true,paused=true)=>({turn:'idle',paused,speaking,epoch,continue_epoch:continued});
 function replay(s=state(),events=[]){api.replay({client:'owner',session:'test',agents:[{id:'agent',config:{vendor:'local',model:'test',surface:'chat'},state:s}],events,partials:[]});}
 function emit(type,data,extra={}){api.apply(type==='state_changed'?{type,agent_id:'agent',from:data.state.turn,to:data.state.turn,...data,...extra}:{type,agent_id:'agent',data,...extra});}
 function speak(id,text,epoch=1){emit('speech_requested',{id,epoch});emit('speech_claimed',{id,epoch,owner:'owner',reason:'automatic',text});}
 replay();return{...api,flush:()=>{const t=timers.find(t=>!t.cancelled&&t.ms===0);if(t){t.cancelled=true;t.fn();return true;}return false;},run:code=>vm.runInContext(code,context),emit,speak,state,replay,sent,utterances,nodes,get cancelled(){return cancelled;}};
}
module.exports=fixture;
if(require.main===module&&process.argv[2]!=='chapter12'){
{
 const f=fixture(),unsafe='<img src=x onerror=attack()>',frames=[{type:'part_final',agent_id:'agent',part_id:1,replay:true,part:{type:'thinking',text:unsafe+' complete thinking'}},{type:'part_final',agent_id:'agent',part_id:2,replay:true,part:{type:'tool_call',name:'edit_file',text:'edit full',args:{path:'demo.go',old_text:'old <script>',new_text:'new <img>'}}}];
 f.replay(f.state(),frames);assert.equal(f.utterances.length,0,'replay must be silent');assert.equal(f.sent.length,0,'replay must not claim speech');
 const cards=f.$('conversation').children;assert.equal(cards.length,2);for(const card of cards)assert.equal(card.querySelector('details').open,false,'final thinking/call compact');
 assert(cards[0].querySelector('details').textContent.includes(unsafe));assert(cards[1].querySelector('.deletion').textContent.includes('old <script>'));assert(cards[1].querySelector('.insertion').textContent.includes('new <img>'));assert.equal(cards[1].querySelectorAll('img').length,0);
}
{
 const f=fixture();f.$('input').value='draft';f.$('input').oninput({isTrusted:true});f.speak('cancelled','words before continue');const old=f.utterances[0];
 f.$('continue').onclick();f.emit('state_changed',{state:f.state(2,1,true,false)});
 const pauses=f.sent.filter(x=>x.type==='pause').length,releases=f.sent.filter(x=>x.type==='unpause').length,afterContinue=f.sent.length;
 old.onend();f.emit('speech_claimed',{id:'cancelled',owner:'owner',epoch:1,text:'stale must not run'});
 assert.equal(f.sent.length,afterContinue,'Continue cannot be vetoed by an old claim or native callback');assert.equal(f.utterances.length,1,'stale work must not revive audio');
 // No new human input: this queue must own new edges, not reuse Continue's release.
 f.speak('next','next fresh queue',2);assert.equal(f.sent.filter(x=>x.type==='pause').length,pauses+1,'new queue acquires its own pause after Continue');
 f.utterances[1].onend();assert.equal(f.sent.filter(x=>x.type==='unpause').length,releases,'native done alone must not release the server queue');
 f.emit('speech_idle',{owner:'owner',epoch:2});assert.equal(f.sent.filter(x=>x.type==='unpause').length,releases+1,'next queue releases its own pause after Continue');
}
{
 const f=fixture();f.$('input').value='old draft';f.$('input').oninput({isTrusted:true});f.speak('cancelled-typing','old speech');const old=f.utterances[0];
 f.$('continue').onclick();const pauses=f.sent.filter(x=>x.type==='pause').length,releases=f.sent.filter(x=>x.type==='unpause').length;
 // This is new human intent, arriving before the delayed Continue acknowledgement.
 f.$('input').value='new trusted draft';f.$('input').oninput({isTrusted:true});assert.equal(f.sent.filter(x=>x.type==='pause').length,pauses+1,'new trusted typing after Continue acquires a pause');
 f.emit('state_changed',{state:f.state(2,1,true,false)});old.onend();f.emit('speech_claimed',{id:'cancelled-typing',owner:'owner',epoch:1,text:'stale must not run'});
 assert.equal(f.utterances.length,1,'delayed acknowledgement and stale work must not revive audio');assert.equal(f.sent.filter(x=>x.type==='pause').length,pauses+1,'stale work must not create another pause');assert.equal(f.sent.filter(x=>x.type==='unpause').length,releases,'acknowledgement must not clear newer human typing');
 f.speak('typing-held','new speech while typing',2);f.utterances[1].onend();f.emit('speech_idle',{owner:'owner',epoch:2});
 assert.equal(f.sent.filter(x=>x.type==='pause').length,pauses+1,'speech shares the genuine human pause');assert.equal(f.sent.filter(x=>x.type==='unpause').length,releases,'speech idle must retain newer human typing');
 f.$('input').value='';f.$('input').oninput({isTrusted:true});assert.equal(f.sent.filter(x=>x.type==='unpause').length,releases+1,'human clearing releases the retained pause');
}
{
 const f=fixture();f.speak('owned','read this');const old=f.utterances[0];f.emit('state_changed',{state:f.state(2,0,false,true)});f.emit('speech_cancelled',{id:'owned',owner:'owner',release:true,epoch:2});
 assert.equal(f.cancelled,1);assert.equal(f.sent.filter(x=>x.type==='unpause').length,1,'other-tab Speaking off releases empty owner queue');const before=f.sent.length;old.onend();assert.equal(f.sent.length,before,'cancelled native callback inert');
}
{
 const f=fixture();f.$('input').value='legitimate typing';f.$('input').oninput({isTrusted:true});f.speak('owned','read');f.emit('state_changed',{state:f.state(2,0,false,true)});f.emit('speech_cancelled',{id:'owned',owner:'owner',release:true,epoch:2});assert.equal(f.sent.filter(x=>x.type==='unpause').length,0,'speech release must respect current typing');f.$('input').value='';f.$('input').oninput({isTrusted:true});assert.equal(f.sent.filter(x=>x.type==='unpause').length,1);
}
{
 const f=fixture(),text='complete '.repeat(450)+'🙂 final';f.speak('long',text);let i=0;while(i<f.utterances.length)f.utterances[i++].onend();assert.equal(f.utterances.map(u=>u.text).join(' '),text,'full readable content must reach native callbacks');assert.equal(f.sent.filter(x=>x.type==='speech_done').length,1);assert.equal(f.sent.filter(x=>x.type==='unpause').length,0,'no release between server-queued tasks');
 f.speak('follow-on','more accepted content');f.utterances[i].onend();assert.equal(f.sent.filter(x=>x.type==='pause').length,1,'same owner queue has one pause edge');f.emit('speech_idle',{owner:'owner',epoch:1});assert.equal(f.sent.filter(x=>x.type==='unpause').length,1);
}
{
 const f=fixture();f.$('input').value='offline draft';f.$('input').oninput({isTrusted:true});f.replay(f.state(2,1,true,false));f.speak('after-reconnect','fresh speech',2);f.utterances[0].onend();f.emit('speech_idle',{owner:'owner',epoch:2});assert(f.sent.some(x=>x.type==='unpause'),'replay applies missed Continue authority without speaking old content');
}
{
 const f=fixture();f.emit('part_delta',null,{part_id:41,kind:'text',chunk:'hel'});f.emit('part_delta',null,{part_id:41,kind:'text',chunk:'lo'});
 assert.equal(f.$('conversation').children[0].querySelector('.body').textContent,'hello','public chunks accumulate incrementally');
 f.emit('part_final',null,{part_id:41,part:{type:'text',text:'hello'}});f.emit('part_delta',null,{part_id:41,kind:'text',chunk:'stale'});
 assert.equal(f.$('conversation').children.length,1);assert.equal(f.$('conversation').children[0].querySelector('.body').textContent,'hello','public final fences later chunks');
}
console.log('Chapter 7 round 1+2 controlled DOM/callback regressions passed (no browser/audio).');

}
if(require.main===module){
 const f=fixture(),input=f.$('input');input.value='programmatic';input.oninput({isTrusted:false});
 assert.equal(f.sent.filter(x=>x.type==='pause').length,0,'synthetic input must not claim human pause');
 f.speak('synthetic-draft','speech with a synthetic draft');f.utterances[0].onend();f.emit('speech_idle',{owner:'owner',epoch:1});
 assert(f.sent.some(x=>x.type==='unpause'),'speech completion must not reclassify a synthetic draft as typing');
 input.value=' ';input.oninput({isTrusted:true});
 const releases=f.sent.filter(x=>x.type==='unpause').length;
 f.speak('human-draft','speech with human whitespace');f.utterances[1].onend();f.emit('speech_idle',{owner:'owner',epoch:1});
 assert.equal(f.sent.filter(x=>x.type==='unpause').length,releases,'human whitespace still owns its pause');
 input.value='';input.oninput({isTrusted:true});
 assert.equal(f.sent.filter(x=>x.type==='unpause').length,releases+1,'clearing human input releases typing pause');
}
