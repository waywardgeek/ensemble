'use strict';
const assert=require('node:assert/strict');
const fixture=require('./client_round1_test.cjs');
const f=fixture();f.replay();
function receive(u){f.run('socket.onmessage({data:'+JSON.stringify(JSON.stringify({type:'usage',agent_id:'agent',...u}))+'})');}
const meter={requests:1,last:{input:11,output:13,cache_read:17,cache_write:19},session:{input:11,output:13,cache_read:17,cache_write:19},cost:{usd:.1234,status:'known'},buckets:[{vendor:'anthropic',surface:'messages',model:'served-model',requests:1,usage:{input:11,output:13,cache_read:17,cache_write:19},cost:{usd:.1234,status:'known'}}],cache:{marker_count:0,marker_reason:'implicit',prediction:{},sections:[]}};
const u={input:11,output:13,cache_read:17,cache_write:19,total:{input:11,output:13,cache_read:17,cache_write:19},hit_rate:17/47,cost_usd:.1234,priced:true,meter};
receive(u);assert.equal(f.$('cache-meter').textContent,'$0.1234 · In 47 · Cache 17 · Hit 36.2%');
assert.match(f.$('cache-meter').title,/served-model/);assert.match(f.$('cache-meter').title,/output 13/);
// Public latest/total counts, not auxiliary meter snapshots, drive the tooltip.
receive({...u,meter:{...meter,last:{input:901,output:902,cache_read:903,cache_write:904},session:{input:911,output:912,cache_read:913,cache_write:914}}});
assert.match(f.$('cache-meter').title,/Last response: uncached input 11, output 13, cache write 19, cache read 17/);
assert.match(f.$('cache-meter').title,/Session: uncached input 11, output 13, cache write 19, cache read 17/);

receive({...u,input:-1,cache_read:-1,total:{...u.total,input:-1,cache_read:-1},cost_usd:null,priced:false,meter:{...meter,cost:{usd:.1234,status:'partial'}}});
assert.match(f.$('cache-meter').textContent,/partial \(known \$0.1234\).*In unknown/);
assert.match(f.$('cache-meter').title,/Last response: uncached input unknown/);
assert.match(f.$('cache-meter').title,/Session: uncached input unknown, output 13, cache write 19, cache read unknown/);
const empty={input:0,output:0,cache_read:0,cache_write:0,total:{input:0,output:0,cache_read:0,cache_write:0},hit_rate:0,cost_usd:0,priced:true,meter:{...meter,requests:0,last:{input:-1,output:-1,cache_read:-1,cache_write:-1},session:{input:0,output:0,cache_read:0,cache_write:0},cost:{usd:0,status:'known'},buckets:[]}};
f.run('replay({client:"owner",session:"fresh",agents:[{id:"agent",state:'+JSON.stringify(f.state())+',config:{vendor:"anthropic",model:"fixture",surface:"messages"}}],events:['+JSON.stringify({type:'usage',agent_id:'agent',...empty})+'],partials:[]})');
assert.equal(f.$('cache-meter').textContent,'$0.0000 · In 0 · Cache 0 · Hit 0.0%');
console.log('Chapter 15 R1 production usage consumer passed');
if(process.argv[2]){
 const emitted=JSON.parse(require('node:fs').readFileSync(process.argv[2],'utf8'));
 const actual=fixture();
 const replay=frame=>actual.replay(actual.state(),[{...frame,agent_id:'agent'}]);
 replay(emitted.initial);assert.equal(actual.$('cache-meter').textContent,'$0.0000 · In 0 · Cache 0 · Hit 0.0%');
 actual.run('socket.onmessage({data:'+JSON.stringify(JSON.stringify({...emitted.live,agent_id:'agent'}))+'})');
 assert.equal(actual.$('cache-meter').textContent,'$0.0003 · In 47 · Cache 17 · Hit 36.2%');
 assert.match(actual.$('cache-meter').title,/claude-sonnet-4-6/);assert.match(actual.$('cache-meter').title,/Session: uncached input 11, output 13, cache write 19, cache read 17/);
 actual.run('socket.onmessage({data:'+JSON.stringify(JSON.stringify({...emitted.partial,agent_id:'agent'}))+'})');
 assert.equal(actual.$('cache-meter').textContent,'partial (known $0.0003) · In 21 · Cache 7 · Hit 35.3%');
 assert.match(actual.$('cache-meter').title,/unknown-served-model/);
 replay(emitted.reconnected);assert.equal(actual.$('cache-meter').textContent,'partial (known $0.0003) · In 21 · Cache 7 · Hit 35.3%');
 replay(emitted.initial);assert.equal(actual.$('cache-meter').textContent,'$0.0000 · In 0 · Cache 0 · Hit 0.0%');
 console.log('Actual initial/live/reconnect/reset WebSocket frames consumed by production JS');
}
