const test=require('node:test'),assert=require('node:assert/strict'),fs=require('node:fs'),vm=require('node:vm');
const source=fs.readFileSync(__dirname+'/web/gui/app.js','utf8');
// Run the actual browser-tool implementation; this is a controllable DOM, not
// a browser/native speech test and not a replacement production observer.
function fixture(nodes){
 const sent=[];
 // The evaluated tail also registers page controls; their behavior is covered
 // by full-page tests. Keep them separate from the observed MCP snapshot DOM.
 const pageControls={'set-model':{addEventListener(){}},'agent-status':{addEventListener(){}}};
 const document={body:{querySelectorAll:()=>nodes},querySelectorAll:selector=>nodes.filter(n=>selector===`[data-ensemble-snapshot-id="${n.attrs['data-ensemble-snapshot-id']}"]`)};
 const context=vm.createContext({document,$:id=>pageControls[id],TextEncoder,Date,console,selected:'fixture-agent',nodeIDs:new WeakMap(),nextNodeID:0,mcpAgents:new Set(),mcpReady:new Set(),send:(type,data)=>sent.push({type,...data}),getComputedStyle:n=>({display:n.hidden?'none':'block',visibility:'visible',opacity:'1'}),Event:class{constructor(type){this.type=type;}}});
 vm.runInContext(source.slice(source.indexOf('// Historical browser MCP tools:')),context);
 return {context,sent,run:code=>vm.runInContext(code,context),snapshot:()=>vm.runInContext('guiSnapshot()',context),document};
}
function node(tag,text,extra={}){
 return {tagName:tag.toUpperCase(),innerText:text,attrs:{},dataset:{},events:[],clicks:0,...extra,
  setAttribute(k,v){this.attrs[k]=v;},getAttribute(k){return this.attrs[k]??null;},hasAttribute(k){return Object.hasOwn(this.attrs,k);},
  getBoundingClientRect(){return {x:0,y:0,width:100,height:this.hidden?0:20};},
  closest(selector){if(selector.includes('[hidden]'))return this.hidden?this:null;if(selector==='[data-part-id]')return this.dataset.partId?this:null;if(selector==='[data-agent-id]')return this.dataset.agentId?this:null;return null;},
  matches(selector){if(selector===':disabled')return !!this.disabled;if(selector==='input,textarea,select')return ['INPUT','TEXTAREA','SELECT'].includes(this.tagName);if(selector.includes('[type=password]')&&!selector.includes(':not'))return this.type==='password'||this.type==='file';if(selector.includes('input:not'))return this.tagName==='TEXTAREA'||this.tagName==='INPUT'&&!['password','file'].includes(this.type);return false;},
  click(){this.clicks++;},dispatchEvent(e){this.events.push(e.type);}
 };
}
test('snapshot selectors operate anonymous dynamic controls and retain actual state',()=>{
 const button=node('button','Dynamic action'),input=node('input','',{value:'before',readOnly:false,checked:false}),disabled=node('input','',{type:'checkbox',checked:true,disabled:true,value:'yes'}),password=node('input','',{type:'password',value:'do-not-leak'});
 const f=fixture([button,input,disabled,password]);const first=f.snapshot();
 assert.match(first,/^# GUI snapshot/);assert.match(first,/checked=true/);assert.match(first,/disabled=true/);assert.match(first,/value=/);assert.ok(!first.includes('do-not-leak'));
 const selector=n=>`[data-ensemble-snapshot-id="${n.attrs['data-ensemble-snapshot-id']}"]`;
 const stable=selector(button);assert.ok(first.includes(stable));assert.equal(f.document.querySelectorAll(stable)[0],button);
 f.run(`browserCall('gui_click',{selector:${JSON.stringify(stable)}})`);assert.equal(button.clicks,1);
 f.run(`browserCall('gui_input',{selector:${JSON.stringify(selector(input))},text:'after 😀'})`);assert.equal(input.value,'after 😀');assert.deepEqual(input.events,['input','change']);
 assert.ok(f.snapshot().includes(stable));assert.ok(f.snapshot().includes('after 😀'));
 assert.throws(()=>f.run(`browserCall('gui_click',{selector:${JSON.stringify(selector(disabled))}})`),/disabled/);
});
test('Markdown is UTF-8 bounded and accounts for every clipped preview/artifact',()=>{
 const nodes=Array.from({length:80},(_,i)=>node('article',`${i}: `+'界😀é'.repeat(350),{dataset:{partId:String(i+1),agentId:'a'}}));nodes.push(node('article','hidden text',{hidden:true}));
 const f=fixture(nodes),markdown=f.snapshot();assert.ok(Buffer.byteLength(markdown,'utf8')<=4096);assert.ok(!markdown.includes('\uFFFD'));
 const previews=[...markdown.matchAll(/; preview: ("(?:\\.|[^"\\])*")/g)].map(m=>JSON.parse(m[1]));
 const total=nodes.filter(n=>!n.hidden).reduce((n,x)=>n+[...x.innerText].length,0),shown=previews.reduce((n,x)=>n+[...x].length,0);
 const counts=markdown.match(/Elided visible preview characters: (\d+); omitted artifacts: (\d+)\/(\d+) \((\d+) hidden/);assert.ok(counts);
 assert.equal(+counts[1],total-shown);assert.equal(+counts[2],nodes.length-previews.length);assert.equal(+counts[3],nodes.length);assert.equal(+counts[4],1);assert.ok(+counts[1]>0);assert.ok(+counts[2]>0);
});
test('same page clears closed readiness and speaks jsonrpc/payload after reconnect',()=>{
 const f=fixture([node('button','Click')]);
 const receive=payload=>{f.context.payload=payload;f.run("handleMCPFrame({type:'jsonrpc',agent_id:'a',payload})");};
 receive({jsonrpc:'2.0',id:1,method:'initialize'});assert.equal(f.sent[0].type,'jsonrpc');assert.equal(f.sent[0].payload.id,1);
 receive({jsonrpc:'2.0',method:'notifications/initialized'});assert.equal(f.run("mcpReady.has('a')"),true);
 receive({jsonrpc:'2.0',id:2,method:'tools/call',params:{name:'gui_snapshot',arguments:{}}});assert.match(f.sent.at(-1).payload.result.content[0].text,/^# GUI snapshot/);
 f.run("handleMCPFrame({type:'mcp_closed',agent_id:'a'})");assert.equal(f.run("mcpAgents.has('a')||mcpReady.has('a')"),false);
 receive({jsonrpc:'2.0',id:3,method:'initialize'});receive({jsonrpc:'2.0',method:'notifications/initialized'});assert.equal(f.run("mcpAgents.has('a')&&mcpReady.has('a')"),true);
 receive({jsonrpc:'2.0',id:4,method:'tools/list'});assert.ok(f.sent.at(-1).payload.result.tools.some(t=>t.name==='gui_snapshot'));
});
